package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"testing"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient"
	api "github.com/monshunter/easyinterview/backend/internal/api/generated"
	"github.com/monshunter/easyinterview/backend/internal/auth"
	"github.com/monshunter/easyinterview/backend/internal/middleware/idempotency"
	"github.com/monshunter/easyinterview/backend/internal/platform/config"
	"github.com/monshunter/easyinterview/backend/internal/runner"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
	"github.com/monshunter/easyinterview/backend/internal/targetjob"

	_ "github.com/lib/pq"
)

const (
	fullFunnelSeedEmail      = "full-funnel-seed@example.com"
	fullFunnelJourneyEmail   = "full-funnel-journey@example.com"
	fullFunnelAuthPepper     = "full-funnel-test-pepper"
	fullFunnelSessionSecret  = "full-funnel-test-session-secret"
	fullFunnelSeedResumeText = "Full funnel seed resume text with Go, React, and async job ownership."
	fullFunnelJourneyJDText  = "Full funnel private JD text for a backend platform role. This raw text must stay out of scenario evidence."
	fullFunnelAnswerText     = "I split migration risk by dependency owner and shipped the first backend slice behind a measurable rollout gate."
)

func TestE2EP0098FullFunnelImportToNextRound(t *testing.T) {
	h := newFullFunnelJourneyHarness(t)
	if h.handler == nil || h.kernel == nil || h.cookie == nil || h.userID == "" {
		t.Fatalf("full-funnel harness is incomplete: handler=%v kernel=%v cookie=%v userID=%q", h.handler != nil, h.kernel != nil, h.cookie != nil, h.userID)
	}

	seed := h.seedReadyResume(t)
	imported := h.importTargetJob(t)
	h.runKernelOnce(t, "target_import")
	target := h.getTargetJob(t, imported.TargetJobId)
	if target.AnalysisStatus != sharedtypes.TargetJobParseStatusReady {
		t.Fatalf("target analysisStatus=%q, want ready: %+v", target.AnalysisStatus, target)
	}
	if target.Title == "" || len(target.Requirements) == 0 || target.Summary == nil || target.FitSummary == nil {
		t.Fatalf("target parse result is incomplete: %+v", target)
	}
	h.assertTargetImportPersisted(t, imported.TargetJobId, imported.Job.Id)

	plan := h.createBaselinePracticePlan(t, imported.TargetJobId, seed.ResumeID)
	h.assertBaselinePracticePlanPersisted(t, plan.Id, imported.TargetJobId, seed.ResumeID)

	session := h.startPracticeSession(t, plan.Id)
	sessionReplay := h.startPracticeSessionWithKey(t, plan.Id, "e2e-p0-098-start-session")
	h.assertStartSessionReplayNoDuplicate(t, session, sessionReplay, plan.Id)
	event := h.appendAnswerEvent(t, session)
	h.assertPracticeSessionEventLoopPersisted(t, session.Id, plan.Id, imported.TargetJobId, event)

	reportWithJob := h.completePracticeSession(t, session.Id)
	reportReplay := h.completePracticeSessionWithKey(t, session.Id, "e2e-p0-098-complete-session")
	h.assertCompleteSessionReplayNoDuplicate(t, reportWithJob, reportReplay, session.Id)
	h.runKernelOnce(t, "report_generate")
	report := h.getFeedbackReport(t, reportWithJob.ReportId)
	h.assertFeedbackReportPersisted(t, report, reportWithJob.Job.Id, session.Id, imported.TargetJobId)

	nextRoundPlan := h.createNextRoundPracticePlan(t, imported.TargetJobId, seed.ResumeID, report.Id)
	nextRoundReplay := h.createNextRoundPracticePlanWithKey(t, imported.TargetJobId, seed.ResumeID, report.Id, "e2e-p0-098-create-next-round-plan")
	if nextRoundReplay.Id != nextRoundPlan.Id {
		t.Fatalf("next_round createPracticePlan replay id=%q, want %q", nextRoundReplay.Id, nextRoundPlan.Id)
	}
	h.assertNextRoundPracticePlanPersisted(t, nextRoundPlan.Id, plan.Id, imported.TargetJobId, seed.ResumeID, report.Id)
	h.assertPrivacyRedlines(t)
}

func TestE2EP0098CreatePracticePlanAcceptsEmptyFocusCodes(t *testing.T) {
	h := newFullFunnelJourneyHarness(t)
	seed := h.seedReadyResume(t)
	imported := h.importTargetJob(t)
	h.runKernelOnce(t, "target_import")

	raw := h.doJSON(t, http.MethodPost, "/api/v1/practice/plans", "e2e-p0-098-create-empty-focus-plan", api.CreatePracticePlanRequest{
		TargetJobId:          imported.TargetJobId,
		ResumeId:             seed.ResumeID,
		Goal:                 sharedtypes.PracticeGoalBaseline,
		Mode:                 sharedtypes.PracticeModeAssisted,
		InterviewerPersona:   sharedtypes.InterviewerRoleHiringManager,
		Difficulty:           "standard",
		Language:             "en",
		TimeBudgetMinutes:    30,
		QuestionBudget:       6,
		FocusCompetencyCodes: []string{},
	}, http.StatusCreated)
	var plan api.PracticePlan
	decodeJSON(t, raw, &plan)
	if plan.Id == "" || plan.TargetJobId != imported.TargetJobId || plan.Goal != sharedtypes.PracticeGoalBaseline {
		t.Fatalf("empty focus createPracticePlan response mismatch: %+v", plan)
	}

	var focusCount int
	if err := h.db.QueryRowContext(h.ctx, `
select cardinality(focus_competency_codes)
from practice_plans
where id = $1
  and user_id = $2`, plan.Id, h.userID).Scan(&focusCount); err != nil {
		t.Fatalf("select empty focus practice plan: %v", err)
	}
	if focusCount != 0 {
		t.Fatalf("focus_competency_codes cardinality = %d, want 0", focusCount)
	}
}

func TestE2EP0098FullFunnelLegacyNegativeRoutePattern(t *testing.T) {
	pattern := regexp.MustCompile(fullFunnelLegacyRoutePattern)
	for _, allowed := range []string{
		"startPracticeSession",
		"createPracticePlan",
		"practice_plans",
		"resumeId",
		"resume_assets",
		"/api/v1/practice/sessions/{sessionId}/voice-turns",
	} {
		if pattern.MatchString(allowed) {
			t.Fatalf("legacy route pattern falsely matched canonical token %q", allowed)
		}
	}
	root := scenarioRepoRoot(t)
	for _, rel := range fullFunnelLegacyScanPaths() {
		path := filepath.Join(root, rel)
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("stat legacy scan path %s: %v", rel, err)
		}
		if !info.IsDir() {
			assertFullFunnelLegacyCleanFile(t, pattern, path, rel)
			continue
		}
		err = filepath.WalkDir(path, func(candidate string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() {
				return nil
			}
			assertFullFunnelLegacyCleanFile(t, pattern, candidate, rel)
			return nil
		})
		if err != nil {
			t.Fatalf("walk legacy scan path %s: %v", rel, err)
		}
	}
}

func TestE2EP0FullFunnelReadyResumeSeedUsesRegisterResumeAndRunner(t *testing.T) {
	h := newFullFunnelResumeSeedHarness(t)

	seed := h.seedReadyResume(t)
	if seed.ResumeID == "" || seed.ParseJobID == "" {
		t.Fatalf("seed did not return resumeId and parse job id: %+v", seed)
	}
	h.assertReadyResume(t, seed)
	h.cleanupSeed(t, seed)
	h.assertSeedCleaned(t, seed)
}

type fullFunnelResumeSeed struct {
	UserID     string
	ResumeID   string
	ParseJobID string
}

type fullFunnelJourneyHarness struct {
	ctx     context.Context
	db      *sql.DB
	handler http.Handler
	kernel  *runner.Runtime
	cookie  *http.Cookie
	userID  string
}

type fullFunnelResumeSeedHarness struct {
	ctx     context.Context
	db      *sql.DB
	handler http.Handler
	kernel  *runner.Runtime
	cookie  *http.Cookie
	userID  string
}

func newFullFunnelResumeSeedHarness(t *testing.T) *fullFunnelResumeSeedHarness {
	t.Helper()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL not set; skipping full-funnel resume seed scenario")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("open postgres: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	t.Cleanup(cancel)
	if err := db.PingContext(ctx); err != nil {
		t.Skipf("postgres ping failed (%v); skipping full-funnel resume seed scenario", err)
	}

	cleanupFullFunnelScenarioEmail(t, db, fullFunnelSeedEmail)
	cookie, userID := loginFullFunnelScenarioUser(t, ctx, db, fullFunnelSeedEmail)
	t.Cleanup(func() { cleanupFullFunnelScenarioUser(t, db, userID) })

	loader := loadFullFunnelResumeSeedConfig(t)
	authService := auth.NewPasswordlessService(auth.PasswordlessServiceOptions{
		Store:               auth.NewSQLStore(db),
		ChallengePepper:     fullFunnelAuthPepper,
		SessionCookieSecret: fullFunnelSessionSecret,
	})
	resumeRuntime, err := buildResumeRuntime(
		loader,
		db,
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		uploadRoutes{},
		&apiNoopAIClient{},
	)
	if err != nil {
		t.Fatalf("buildResumeRuntime: %v", err)
	}
	handler := buildAPIHandlerWithUploadAndHandlers(
		loader,
		apiRuntimeFlags{},
		authService,
		targetjob.NewHandler(),
		practiceRoutes{},
		uploadRoutes{},
		resumeRuntime.Routes(),
	)
	kernel := newTestKernel(runner.NewSQLStore(db), resumeRuntime.Handlers)

	return &fullFunnelResumeSeedHarness{
		ctx:     ctx,
		db:      db,
		handler: handler,
		kernel:  kernel,
		cookie:  cookie,
		userID:  userID,
	}
}

func newFullFunnelJourneyHarness(t *testing.T) *fullFunnelJourneyHarness {
	t.Helper()
	return newFullFunnelJourneyHarnessWithTimeout(t, 30*time.Second)
}

func newFullFunnelJourneyHarnessWithTimeout(t *testing.T, timeout time.Duration) *fullFunnelJourneyHarness {
	t.Helper()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL not set; skipping full-funnel journey scenario")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("open postgres: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	t.Cleanup(cancel)
	if err := db.PingContext(ctx); err != nil {
		t.Skipf("postgres ping failed (%v); skipping full-funnel journey scenario", err)
	}

	cleanupFullFunnelScenarioEmail(t, db, fullFunnelJourneyEmail)
	cookie, userID := loginFullFunnelScenarioUser(t, ctx, db, fullFunnelJourneyEmail)
	t.Cleanup(func() { cleanupFullFunnelScenarioUser(t, db, userID) })

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	loader := loadFullFunnelResumeSeedConfig(t)
	authService := auth.NewPasswordlessService(auth.PasswordlessServiceOptions{
		Store:               auth.NewSQLStore(db),
		ChallengePepper:     fullFunnelAuthPepper,
		SessionCookieSecret: fullFunnelSessionSecret,
	})
	targetRuntime, err := buildTargetJobRuntime(loader, db, logger, nil, &privacyDeleteRuntimeHooks{})
	if err != nil {
		t.Fatalf("buildTargetJobRuntime: %v", err)
	}
	t.Cleanup(targetRuntime.Close)

	ai := &fullFunnelScenarioAIClient{}
	resumeRuntime, err := buildResumeRuntime(loader, db, logger, uploadRoutes{}, ai)
	if err != nil {
		t.Fatalf("buildResumeRuntime: %v", err)
	}
	practice, err := buildPracticeRoutes(loader, db, ai)
	if err != nil {
		t.Fatalf("buildPracticeRoutes: %v", err)
	}
	reports, err := buildReportRuntime(loader, db, logger, ai)
	if err != nil {
		t.Fatalf("buildReportRuntime: %v", err)
	}
	jobs := buildJobsRoutes(db)
	handler := buildAPIHandlerWithUploadReportDebriefJobsAndHandlers(
		loader,
		apiRuntimeFlags{},
		authService,
		targetRuntime.Handler,
		practice,
		uploadRoutes{},
		resumeRuntime.Routes(),
		reports.Routes(),
		debriefRoutes{},
		jobs,
	)
	kernel := newTestKernel(runner.NewSQLStore(db), targetRuntime.Handlers, resumeRuntime.Handlers, reports.Handlers)

	return &fullFunnelJourneyHarness{
		ctx:     ctx,
		db:      db,
		handler: handler,
		kernel:  kernel,
		cookie:  cookie,
		userID:  userID,
	}
}

func (h *fullFunnelJourneyHarness) importTargetJob(t *testing.T) api.TargetJobWithJob {
	t.Helper()
	title := "Backend Platform Engineer"
	company := "Full Funnel Systems"
	raw := h.doJSON(t, http.MethodPost, "/api/v1/targets/import", "e2e-p0-098-import-target", api.ImportTargetJobRequest{
		Source: map[string]any{
			"type":    "manual_text",
			"rawText": fullFunnelJourneyJDText,
		},
		TargetLanguage:  "en",
		TitleHint:       &title,
		CompanyNameHint: &company,
	}, http.StatusAccepted)
	var imported api.TargetJobWithJob
	decodeJSON(t, raw, &imported)
	if imported.TargetJobId == "" || imported.Job.Id == "" {
		t.Fatalf("importTargetJob returned empty ids: %+v", imported)
	}
	if imported.Job.JobType != api.JobTypeTargetImport || imported.Job.Status != sharedtypes.JobStatusQueued {
		t.Fatalf("importTargetJob did not queue target_import: %+v", imported.Job)
	}
	return imported
}

func (h *fullFunnelJourneyHarness) seedReadyResume(t *testing.T) fullFunnelResumeSeed {
	t.Helper()
	raw := h.doJSON(t, http.MethodPost, "/api/v1/resumes", "e2e-p0-098-register-resume", api.RegisterResumeRequest{
		Title:      "Full Funnel Journey Resume",
		Language:   "en",
		SourceType: fullFunnelStringPtr("paste"),
		RawText:    fullFunnelStringPtr(fullFunnelSeedResumeText),
	}, http.StatusAccepted)
	var registered api.ResumeWithJob
	decodeJSON(t, raw, &registered)
	if registered.ResumeId == "" || registered.Job.Id == "" {
		t.Fatalf("registerResume did not return resume/job ids: %+v", registered)
	}
	if registered.Job.JobType != api.JobTypeResumeParse || registered.Job.Status != sharedtypes.JobStatusQueued {
		t.Fatalf("registerResume did not queue resume_parse: %+v", registered.Job)
	}
	h.runKernelOnce(t, "resume_parse")
	seed := fullFunnelResumeSeed{UserID: h.userID, ResumeID: registered.ResumeId, ParseJobID: registered.Job.Id}
	h.assertReadyResume(t, seed)
	return seed
}

func (h *fullFunnelJourneyHarness) createBaselinePracticePlan(t *testing.T, targetJobID, resumeID string) api.PracticePlan {
	t.Helper()
	raw := h.doJSON(t, http.MethodPost, "/api/v1/practice/plans", "e2e-p0-098-create-baseline-plan", api.CreatePracticePlanRequest{
		TargetJobId:          targetJobID,
		ResumeId:             resumeID,
		Goal:                 sharedtypes.PracticeGoalBaseline,
		Mode:                 sharedtypes.PracticeModeAssisted,
		InterviewerPersona:   sharedtypes.InterviewerRoleTechnicalManager,
		Difficulty:           "standard",
		Language:             "en",
		TimeBudgetMinutes:    30,
		QuestionBudget:       3,
		FocusCompetencyCodes: []string{"system_design", "async_ownership"},
	}, http.StatusCreated)
	var plan api.PracticePlan
	decodeJSON(t, raw, &plan)
	if plan.Id == "" || plan.TargetJobId != targetJobID || plan.Goal != sharedtypes.PracticeGoalBaseline || plan.SourceReportId != nil {
		t.Fatalf("baseline practice plan mismatch: %+v", plan)
	}
	return plan
}

func (h *fullFunnelJourneyHarness) createNextRoundPracticePlan(t *testing.T, targetJobID, resumeID, sourceReportID string) api.PracticePlan {
	t.Helper()
	return h.createNextRoundPracticePlanWithKey(t, targetJobID, resumeID, sourceReportID, "e2e-p0-098-create-next-round-plan")
}

func (h *fullFunnelJourneyHarness) createNextRoundPracticePlanWithKey(t *testing.T, targetJobID, resumeID, sourceReportID, idempotencyKey string) api.PracticePlan {
	t.Helper()
	raw := h.doJSON(t, http.MethodPost, "/api/v1/practice/plans", idempotencyKey, api.CreatePracticePlanRequest{
		TargetJobId:          targetJobID,
		ResumeId:             resumeID,
		SourceReportId:       fullFunnelStringPtr(sourceReportID),
		Goal:                 sharedtypes.PracticeGoalNextRound,
		Mode:                 sharedtypes.PracticeModeAssisted,
		InterviewerPersona:   sharedtypes.InterviewerRoleTechnicalManager,
		Difficulty:           "standard",
		Language:             "en",
		TimeBudgetMinutes:    30,
		QuestionBudget:       3,
		FocusCompetencyCodes: []string{"system_design", "async_ownership"},
	}, http.StatusCreated)
	var plan api.PracticePlan
	decodeJSON(t, raw, &plan)
	if plan.Id == "" || plan.TargetJobId != targetJobID || plan.Goal != sharedtypes.PracticeGoalNextRound || plan.SourceReportId == nil || *plan.SourceReportId != sourceReportID {
		t.Fatalf("next_round practice plan mismatch: %+v", plan)
	}
	return plan
}

func (h *fullFunnelJourneyHarness) startPracticeSession(t *testing.T, planID string) api.PracticeSession {
	t.Helper()
	return h.startPracticeSessionWithKey(t, planID, "e2e-p0-098-start-session")
}

func (h *fullFunnelJourneyHarness) startPracticeSessionWithKey(t *testing.T, planID, idempotencyKey string) api.PracticeSession {
	t.Helper()
	raw := h.doJSON(t, http.MethodPost, "/api/v1/practice/sessions", idempotencyKey, api.StartPracticeSessionRequest{
		PlanId:       planID,
		HintsEnabled: fullFunnelBoolPtr(true),
	}, http.StatusCreated)
	var session api.PracticeSession
	decodeJSON(t, raw, &session)
	if session.Id == "" || session.PlanId != planID || session.Status != sharedtypes.SessionStatusRunning || session.CurrentTurn == nil {
		t.Fatalf("startPracticeSession response mismatch: %+v", session)
	}
	if session.CurrentTurn.TurnIndex != 1 || session.CurrentTurn.Status != "asked" || session.CurrentTurn.QuestionText == "" {
		t.Fatalf("startPracticeSession did not return first asked turn: %+v", session.CurrentTurn)
	}
	return session
}

func (h *fullFunnelJourneyHarness) appendAnswerEvent(t *testing.T, session api.PracticeSession) api.SessionEventResult {
	t.Helper()
	if session.CurrentTurn == nil {
		t.Fatalf("appendAnswerEvent requires current turn: %+v", session)
	}
	raw := h.doJSON(t, http.MethodPost, "/api/v1/practice/sessions/"+session.Id+"/events", "", api.PracticeSessionEventRequest{
		ClientEventId: "e2e-p0-098-answer-1",
		Kind:          "answer_submitted",
		OccurredAt:    "2026-05-24T13:03:00Z",
		Payload: map[string]any{
			"turnId":     session.CurrentTurn.Id,
			"answerText": fullFunnelAnswerText,
		},
	}, http.StatusOK)
	var result api.SessionEventResult
	decodeJSON(t, raw, &result)
	if !result.Acknowledged || result.Session.Id != session.Id || result.AssistantAction.Type != "ask_follow_up" {
		t.Fatalf("appendSessionEvent first answer mismatch: %+v", result)
	}
	if result.Session.CurrentTurn == nil || result.Session.CurrentTurn.Id != session.CurrentTurn.Id || result.Session.CurrentTurn.Status != "follow_up_requested" {
		t.Fatalf("appendSessionEvent did not keep same turn in follow-up state: %+v", result.Session.CurrentTurn)
	}
	return result
}

func (h *fullFunnelJourneyHarness) completePracticeSession(t *testing.T, sessionID string) api.ReportWithJob {
	t.Helper()
	return h.completePracticeSessionWithKey(t, sessionID, "e2e-p0-098-complete-session")
}

func (h *fullFunnelJourneyHarness) completePracticeSessionWithKey(t *testing.T, sessionID, idempotencyKey string) api.ReportWithJob {
	t.Helper()
	raw := h.doJSON(t, http.MethodPost, "/api/v1/practice/sessions/"+sessionID+"/complete", idempotencyKey, api.CompletePracticeSessionRequest{
		ClientCompletedAt: "2026-05-24T13:05:00Z",
	}, http.StatusAccepted)
	var result api.ReportWithJob
	decodeJSON(t, raw, &result)
	if result.ReportId == "" || result.Job.Id == "" || result.Job.JobType != api.JobTypeReportGenerate || result.Job.Status != sharedtypes.JobStatusQueued {
		t.Fatalf("completePracticeSession did not queue report_generate: %+v", result)
	}
	return result
}

func (h *fullFunnelJourneyHarness) getTargetJob(t *testing.T, targetJobID string) api.TargetJob {
	t.Helper()
	raw := h.doJSON(t, http.MethodGet, "/api/v1/targets/"+targetJobID, "", nil, http.StatusOK)
	var target api.TargetJob
	decodeJSON(t, raw, &target)
	return target
}

func (h *fullFunnelJourneyHarness) getFeedbackReport(t *testing.T, reportID string) api.FeedbackReport {
	t.Helper()
	raw := h.doJSON(t, http.MethodGet, "/api/v1/reports/"+reportID, "", nil, http.StatusOK)
	var report api.FeedbackReport
	decodeJSON(t, raw, &report)
	return report
}

func (h *fullFunnelJourneyHarness) runKernelOnce(t *testing.T, jobType string) {
	t.Helper()
	processed, err := h.kernel.RunOnce(h.ctx)
	if err != nil {
		t.Fatalf("run %s job: %v", jobType, err)
	}
	if !processed {
		t.Fatalf("run %s job processed=false, want true", jobType)
	}
}

func (h *fullFunnelJourneyHarness) assertTargetImportPersisted(t *testing.T, targetJobID, jobID string) {
	t.Helper()
	var (
		status    string
		attempts  int
		completed bool
	)
	if err := h.db.QueryRowContext(h.ctx, `
select status, attempts, completed_at is not null
from async_jobs
where id = $1 and job_type = 'target_import' and resource_id = $2`, jobID, targetJobID).Scan(&status, &attempts, &completed); err != nil {
		t.Fatalf("read target_import async job: %v", err)
	}
	if status != "succeeded" || attempts != 1 || !completed {
		t.Fatalf("target_import job not finalized: status=%q attempts=%d completed=%v", status, attempts, completed)
	}
	assertFullFunnelCount(t, h.db, `
select count(*)
from target_job_requirements
where target_job_id = $1`, targetJobID, 1)
}

func (h *fullFunnelJourneyHarness) assertReadyResume(t *testing.T, seed fullFunnelResumeSeed) {
	t.Helper()
	raw := h.doJSON(t, http.MethodGet, "/api/v1/resumes/"+seed.ResumeID, "", nil, http.StatusOK)
	var detail api.Resume
	decodeJSON(t, raw, &detail)
	if detail.Id != seed.ResumeID || detail.ParseStatus != sharedtypes.TargetJobParseStatusReady {
		t.Fatalf("journey seed resume is not ready: %+v", detail)
	}
	if detail.ParsedSummary == nil || len(*detail.ParsedSummary) == 0 {
		t.Fatalf("journey seed resume missing parsed summary: %+v", detail)
	}
}

func (h *fullFunnelJourneyHarness) assertBaselinePracticePlanPersisted(t *testing.T, planID, targetJobID, resumeID string) {
	t.Helper()
	var (
		gotTarget string
		gotResume string
		goal      string
		status    string
	)
	if err := h.db.QueryRowContext(h.ctx, `
select target_job_id::text, resume_id::text, goal, status
from practice_plans
where id = $1 and user_id = $2`, planID, h.userID).Scan(&gotTarget, &gotResume, &goal, &status); err != nil {
		t.Fatalf("read baseline practice plan: %v", err)
	}
	if gotTarget != targetJobID || gotResume != resumeID || goal != string(sharedtypes.PracticeGoalBaseline) || status != "ready" {
		t.Fatalf("baseline plan persisted mismatch: target=%q resume=%q goal=%q status=%q", gotTarget, gotResume, goal, status)
	}
}

func (h *fullFunnelJourneyHarness) assertPracticeSessionEventLoopPersisted(t *testing.T, sessionID, planID, targetJobID string, event api.SessionEventResult) {
	t.Helper()
	if event.Session.Id != sessionID {
		t.Fatalf("event result session=%q, want %q", event.Session.Id, sessionID)
	}
	assertFullFunnelCount(t, h.db, `
select count(*)
from practice_sessions
where id = $1 and user_id = $2 and plan_id = $3 and target_job_id = $4 and status = 'running'`, sessionID, h.userID, planID, targetJobID, 1)
	assertFullFunnelCount(t, h.db, `
select count(*)
from practice_turns
where session_id = $1 and turn_index = 1 and status = 'follow_up_requested'`, sessionID, 1)
	assertFullFunnelCount(t, h.db, `
select count(*)
from practice_session_events
where session_id = $1`, sessionID, 2)
	assertFullFunnelCount(t, h.db, `
select count(*)
from outbox_events
where aggregate_id = $1 and event_name = 'practice.session.started'`, sessionID, 1)
	assertFullFunnelCount(t, h.db, `
select count(*)
from outbox_events
where aggregate_id = $1`, sessionID, 1)
}

func (h *fullFunnelJourneyHarness) assertStartSessionReplayNoDuplicate(t *testing.T, first, replay api.PracticeSession, planID string) {
	t.Helper()
	if replay.Id != first.Id || replay.PlanId != first.PlanId || replay.CurrentTurn == nil || first.CurrentTurn == nil || replay.CurrentTurn.Id != first.CurrentTurn.Id {
		t.Fatalf("startPracticeSession replay mismatch: first=%+v replay=%+v", first, replay)
	}
	assertFullFunnelCount(t, h.db, `
select count(*)
from practice_sessions
where user_id = $1 and plan_id = $2 and status = 'running'`, h.userID, planID, 1)
	assertFullFunnelCount(t, h.db, `
select count(*)
from practice_session_events
where session_id = $1 and event_type = 'session_started'`, first.Id, 1)
	assertFullFunnelCount(t, h.db, `
select count(*)
from outbox_events
where aggregate_id = $1 and event_name = 'practice.session.started'`, first.Id, 1)
}

func (h *fullFunnelJourneyHarness) assertCompleteSessionReplayNoDuplicate(t *testing.T, first, replay api.ReportWithJob, sessionID string) {
	t.Helper()
	if replay.ReportId != first.ReportId || replay.Job.Id != first.Job.Id {
		t.Fatalf("completePracticeSession replay mismatch: first=%+v replay=%+v", first, replay)
	}
	assertFullFunnelCount(t, h.db, `
select count(*)
from feedback_reports
where user_id = $1 and session_id = $2`, h.userID, sessionID, 1)
	assertFullFunnelCount(t, h.db, `
select count(*)
from async_jobs
where job_type = 'report_generate' and resource_id = $1 and dedupe_key = $2`, first.ReportId, sessionID, 1)
	assertFullFunnelCount(t, h.db, `
select count(*)
from outbox_events
where aggregate_id = $1 and event_name = 'practice.session.completed'`, sessionID, 1)
}

func (h *fullFunnelJourneyHarness) assertFeedbackReportPersisted(t *testing.T, report api.FeedbackReport, jobID, sessionID, targetJobID string) {
	t.Helper()
	if report.Id == "" || report.SessionId != sessionID || report.TargetJobId != targetJobID || report.Status != sharedtypes.ReportStatusReady {
		t.Fatalf("feedback report mismatch: %+v", report)
	}
	hasNextRound := false
	for _, action := range report.NextActions {
		if action.Type == "next_round" {
			hasNextRound = true
		}
	}
	if !hasNextRound {
		t.Fatalf("feedback report nextActions missing next_round: %+v", report.NextActions)
	}
	if len(report.QuestionAssessments) == 0 || report.Provenance == nil {
		t.Fatalf("feedback report missing assessments/provenance: %+v", report)
	}
	var (
		status    string
		attempts  int
		completed bool
	)
	if err := h.db.QueryRowContext(h.ctx, `
select status, attempts, completed_at is not null
from async_jobs
where id = $1 and job_type = 'report_generate' and resource_id = $2`, jobID, report.Id).Scan(&status, &attempts, &completed); err != nil {
		t.Fatalf("read report_generate async job: %v", err)
	}
	if status != "succeeded" || attempts != 1 || !completed {
		t.Fatalf("report_generate job not finalized: status=%q attempts=%d completed=%v", status, attempts, completed)
	}
	assertFullFunnelCount(t, h.db, `
select count(*)
from feedback_reports
where id = $1 and user_id = $2 and session_id = $3 and target_job_id = $4 and status = 'ready'`, report.Id, h.userID, sessionID, targetJobID, 1)
	assertFullFunnelCount(t, h.db, `
select count(*)
from outbox_events
where aggregate_id = $1 and event_name = 'report.generated'`, report.Id, 1)
}

func (h *fullFunnelJourneyHarness) assertNextRoundPracticePlanPersisted(t *testing.T, nextPlanID, firstPlanID, targetJobID, resumeID, reportID string) {
	t.Helper()
	if nextPlanID == firstPlanID {
		t.Fatalf("next_round plan reused first plan id %q", firstPlanID)
	}
	var (
		gotTarget string
		gotResume string
		gotReport string
		goal      string
		status    string
	)
	if err := h.db.QueryRowContext(h.ctx, `
select target_job_id::text, resume_id::text, source_report_id::text, goal, status
from practice_plans
where id = $1 and user_id = $2`, nextPlanID, h.userID).Scan(&gotTarget, &gotResume, &gotReport, &goal, &status); err != nil {
		t.Fatalf("read next_round practice plan: %v", err)
	}
	if gotTarget != targetJobID || gotResume != resumeID || gotReport != reportID || goal != string(sharedtypes.PracticeGoalNextRound) || status != "ready" {
		t.Fatalf("next_round plan persisted mismatch: target=%q resume=%q report=%q goal=%q status=%q", gotTarget, gotResume, gotReport, goal, status)
	}
	assertFullFunnelCount(t, h.db, `
select count(*)
from practice_plans
where user_id = $1 and source_report_id = $2 and goal = 'next_round'`, h.userID, reportID, 1)
}

func (h *fullFunnelJourneyHarness) assertPrivacyRedlines(t *testing.T) {
	t.Helper()
	payloads := h.collectObservablePayloads(t)
	assertNoEvidenceLeak(t, payloads,
		fullFunnelJourneyJDText,
		fullFunnelAnswerText,
		"add tradeoff",
		"question_text",
		"answer_text",
		"prompt body",
		"response body",
		"provider secret",
	)
}

func (h *fullFunnelJourneyHarness) collectObservablePayloads(t *testing.T) [][]byte {
	t.Helper()
	query := `
with owned_resources as (
  select id from resumes where user_id = $1
  union select id from target_jobs where user_id = $1
  union select id from practice_plans where user_id = $1
  union select id from practice_sessions where user_id = $1
  union select id from feedback_reports where user_id = $1
  union select id from practice_turns where session_id in (select id from practice_sessions where user_id = $1)
)
select payload::text from outbox_events where aggregate_id in (select id from owned_resources)
union all
select metadata::text from audit_events where user_id = $1 or resource_id in (select id from owned_resources)
union all
select payload::text from async_jobs where resource_id in (select id from owned_resources)
union all
select response_body::text from idempotency_records where user_id = $1 and response_body is not null`
	rows, err := h.db.QueryContext(h.ctx, query, h.userID)
	if err != nil {
		t.Fatalf("query observable privacy payloads: %v", err)
	}
	defer rows.Close()
	var payloads [][]byte
	for rows.Next() {
		var raw sql.NullString
		if err := rows.Scan(&raw); err != nil {
			t.Fatalf("scan observable privacy payload: %v", err)
		}
		if raw.Valid {
			payloads = append(payloads, []byte(raw.String))
		}
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate observable privacy payloads: %v", err)
	}
	if len(payloads) == 0 {
		t.Fatal("observable privacy payload scan returned no rows")
	}
	return payloads
}

const fullFunnelLegacyRoutePattern = `(^|[[:space:]'"'/#?&=:-])(welcome|growth|mistakes|drill|followup|experiences|star(_editor)?|onboarding)([[:space:]'"'/#?&=:-]|$)|mode=debrief|name=['"](plan|resume|voice)['"]|route=['"](plan|resume|voice)['"]|#route=(plan|resume|voice)([[:space:]'"'/#?&=:-]|$)`

func fullFunnelLegacyScanPaths() []string {
	return []string{
		"backend/cmd/api/main.go",
		"backend/internal/api/generated",
		"backend/internal/api/practice",
		"backend/internal/api/reports",
		"openapi/fixtures/Jobs/getJob.json",
		"openapi/fixtures/PracticePlans/createPracticePlan.json",
		"openapi/fixtures/PracticeSessions/appendSessionEvent.json",
		"openapi/fixtures/PracticeSessions/completePracticeSession.json",
		"openapi/fixtures/PracticeSessions/startPracticeSession.json",
		"openapi/fixtures/Reports/getFeedbackReport.json",
		"openapi/fixtures/Resumes/registerResume.json",
		"openapi/fixtures/TargetJobs/getTargetJob.json",
		"openapi/fixtures/TargetJobs/importTargetJob.json",
	}
}

func assertFullFunnelLegacyCleanFile(t *testing.T, pattern *regexp.Regexp, path, scope string) {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read legacy scan file %s: %v", path, err)
	}
	if match := pattern.Find(raw); len(match) > 0 {
		t.Fatalf("legacy route token %q found in %s within scope %s", string(match), path, scope)
	}
}

func loadFullFunnelResumeSeedConfig(t *testing.T) *config.Loader {
	t.Helper()
	canonical := loadE2EP0ConfigPreflightConfig(t)
	if canonical.AppEnv() != "test" {
		t.Fatalf("canonical AppEnv=%q, want test", canonical.AppEnv())
	}
	root := scenarioRepoRoot(t)
	dir := t.TempDir()
	writeAPIFile(t, dir+"/config.yaml", `
app:
  env: test
runtime:
  appVersion: "full-funnel-seed-test"
  defaultUiLanguage: zh-CN
auth:
  challengeTokenPepper: "`+fullFunnelAuthPepper+`"
  sessionCookieSecret: "`+fullFunnelSessionSecret+`"
ai:
  providerRegistryPath: "`+root+`/config/ai-providers.yaml"
  modelProfilePath: "`+root+`/config/ai-profiles.yaml"
  promptsDir: "`+root+`/config/prompts"
  rubricsDir: "`+root+`/config/rubrics"
`)
	loader, err := config.Load(config.Options{AppEnv: "test", ConfigDir: dir})
	if err != nil {
		t.Fatalf("load full-funnel seed config: %v", err)
	}
	return loader
}

type fullFunnelScenarioAIClient struct{}

func (c *fullFunnelScenarioAIClient) Complete(ctx context.Context, profileName string, payload aiclient.CompletePayload) (aiclient.CompleteResponse, aiclient.AICallMeta, error) {
	switch payload.Metadata.FeatureKey {
	case "practice.session.first_question", "practice.session.follow_up", hintFeatureKeyForScenario:
		return (&scenarioPracticeAIClient{}).Complete(ctx, profileName, payload)
	case "report.generate":
		return aiclient.CompleteResponse{Content: reportGenerateJSON()}, aiclient.AICallMeta{}, nil
	case "report.question_assessment":
		return aiclient.CompleteResponse{Content: `{"dimension_results":{"depth":{"status":"meets_bar","confidence":0.8,"score":0.8}},"overall_status":"meets_bar","confidence":0.8,"strengths":["clear"],"gaps":[],"recommended_framework":"STAR","review_status":"open","included_in_retry_plan":true}`}, aiclient.AICallMeta{}, nil
	default:
		return aiclient.CompleteResponse{}, aiclient.AICallMeta{}, errors.New("unexpected full-funnel AI feature key: " + payload.Metadata.FeatureKey)
	}
}

func (c *fullFunnelScenarioAIClient) Transcribe(context.Context, string, aiclient.TranscriptionInput) (aiclient.TranscriptionResponse, aiclient.AICallMeta, error) {
	return aiclient.TranscriptionResponse{}, aiclient.AICallMeta{}, errors.New("unexpected full-funnel Transcribe call")
}

func (c *fullFunnelScenarioAIClient) Stream(context.Context, string, aiclient.CompletePayload) (<-chan aiclient.AIStreamEvent, error) {
	return nil, errors.New("unexpected full-funnel Stream call")
}

func (c *fullFunnelScenarioAIClient) Synthesize(context.Context, string, aiclient.SynthesisInput) (aiclient.SynthesisResponse, aiclient.AICallMeta, error) {
	return aiclient.SynthesisResponse{}, aiclient.AICallMeta{}, errors.New("unexpected full-funnel Synthesize call")
}

func (h *fullFunnelResumeSeedHarness) seedReadyResume(t *testing.T) fullFunnelResumeSeed {
	t.Helper()
	raw := h.doJSON(t, http.MethodPost, "/api/v1/resumes", "e2e-p0-098-register-resume", api.RegisterResumeRequest{
		Title:      "Full Funnel Seed Resume",
		Language:   "en",
		SourceType: fullFunnelStringPtr("paste"),
		RawText:    fullFunnelStringPtr(fullFunnelSeedResumeText),
	}, http.StatusAccepted)
	var registered api.ResumeWithJob
	decodeJSON(t, raw, &registered)
	if registered.ResumeId == "" || registered.Job.Id == "" {
		t.Fatalf("registerResume did not return asset/job ids: %+v", registered)
	}
	if registered.Job.JobType != api.JobTypeResumeParse || registered.Job.Status != sharedtypes.JobStatusQueued {
		t.Fatalf("registerResume did not queue resume_parse: %+v", registered.Job)
	}

	processed, err := h.kernel.RunOnce(h.ctx)
	if err != nil {
		t.Fatalf("resume seed RunOnce: %v", err)
	}
	if !processed {
		t.Fatal("resume seed RunOnce processed=false, want true")
	}
	return fullFunnelResumeSeed{UserID: h.userID, ResumeID: registered.ResumeId, ParseJobID: registered.Job.Id}
}

func (h *fullFunnelResumeSeedHarness) assertReadyResume(t *testing.T, seed fullFunnelResumeSeed) {
	t.Helper()
	raw := h.doJSON(t, http.MethodGet, "/api/v1/resumes/"+seed.ResumeID, "", nil, http.StatusOK)
	var detail api.Resume
	decodeJSON(t, raw, &detail)
	if detail.Id != seed.ResumeID || detail.ParseStatus != sharedtypes.TargetJobParseStatusReady {
		t.Fatalf("seed resume is not ready: %+v", detail)
	}
	if detail.ParsedSummary == nil || len(*detail.ParsedSummary) == 0 {
		t.Fatalf("seed resume missing parsed summary: %+v", detail)
	}
	if detail.ParsedTextSnapshot == nil || *detail.ParsedTextSnapshot != fullFunnelSeedResumeText {
		t.Fatalf("seed resume parsed text snapshot mismatch: %+v", detail.ParsedTextSnapshot)
	}
	var (
		status    string
		attempts  int
		completed bool
	)
	if err := h.db.QueryRowContext(h.ctx, `
select status, attempts, completed_at is not null
from async_jobs
where id = $1`, seed.ParseJobID).Scan(&status, &attempts, &completed); err != nil {
		t.Fatalf("read resume parse job: %v", err)
	}
	if status != "succeeded" || attempts != 1 || !completed {
		t.Fatalf("resume parse job not finalized: status=%q attempts=%d completed=%v", status, attempts, completed)
	}
	assertFullFunnelCount(t, h.db, `
select count(*)
from outbox_events
where aggregate_type = 'resume' and aggregate_id = $1 and event_name = 'resume.parse.completed'`, seed.ResumeID, 1)
}

func (h *fullFunnelResumeSeedHarness) cleanupSeed(t *testing.T, seed fullFunnelResumeSeed) {
	t.Helper()
	cleanupFullFunnelScenarioUser(t, h.db, seed.UserID)
}

func (h *fullFunnelResumeSeedHarness) assertSeedCleaned(t *testing.T, seed fullFunnelResumeSeed) {
	t.Helper()
	assertFullFunnelCount(t, h.db, `select count(*) from users where id = $1`, seed.UserID, 0)
	assertFullFunnelCount(t, h.db, `select count(*) from resumes where id = $1`, seed.ResumeID, 0)
	assertFullFunnelCount(t, h.db, `select count(*) from async_jobs where id = $1 or resource_id = $2`, seed.ParseJobID, seed.ResumeID, 0)
	assertFullFunnelCount(t, h.db, `select count(*) from outbox_events where aggregate_id = $1`, seed.ResumeID, 0)
	assertFullFunnelCount(t, h.db, `select count(*) from idempotency_records where user_id = $1`, seed.UserID, 0)
}

func (h *fullFunnelJourneyHarness) doJSON(t *testing.T, method, path, idempotencyKey string, body any, wantStatus int) []byte {
	t.Helper()
	var reader *bytes.Reader
	if body == nil {
		reader = bytes.NewReader(nil)
	} else {
		raw, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request body: %v", err)
		}
		reader = bytes.NewReader(raw)
	}
	req := httptest.NewRequest(method, path, reader)
	req.AddCookie(h.cookie)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if idempotencyKey != "" {
		req.Header.Set(idempotency.HeaderName, idempotencyKey)
	}
	rec := httptest.NewRecorder()
	h.handler.ServeHTTP(rec, req)
	if rec.Code != wantStatus {
		t.Fatalf("%s %s status=%d want=%d body=%s", method, path, rec.Code, wantStatus, rec.Body.String())
	}
	return rec.Body.Bytes()
}

func (h *fullFunnelResumeSeedHarness) doJSON(t *testing.T, method, path, idempotencyKey string, body any, wantStatus int) []byte {
	t.Helper()
	var reader *bytes.Reader
	if body == nil {
		reader = bytes.NewReader(nil)
	} else {
		raw, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request body: %v", err)
		}
		reader = bytes.NewReader(raw)
	}
	req := httptest.NewRequest(method, path, reader)
	req.AddCookie(h.cookie)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if idempotencyKey != "" {
		req.Header.Set(idempotency.HeaderName, idempotencyKey)
	}
	rec := httptest.NewRecorder()
	h.handler.ServeHTTP(rec, req)
	if rec.Code != wantStatus {
		t.Fatalf("%s %s status=%d want=%d body=%s", method, path, rec.Code, wantStatus, rec.Body.String())
	}
	return rec.Body.Bytes()
}

func loginFullFunnelScenarioUser(t *testing.T, ctx context.Context, db *sql.DB, email string) (*http.Cookie, string) {
	t.Helper()
	tokenSuffix := time.Now().UTC().Format("20060102150405.000000000")
	challengeToken := "424242"
	sessionToken := "full-funnel-session-" + tokenSuffix
	sink := auth.NewDevMailSink(auth.DevMailSinkOptions{VerifyBaseURL: "http://api.test/api/v1/auth/email/verify"})
	service := auth.NewPasswordlessService(auth.PasswordlessServiceOptions{
		Store:                 auth.NewSQLStore(db),
		Dispatcher:            auth.NewImmediateMailDispatcher(sink),
		DeliverySecrets:       sink,
		TokenGenerator:        apiFixedTokenGenerator(challengeToken),
		SessionTokenGenerator: apiFixedTokenGenerator(sessionToken),
		ChallengePepper:       fullFunnelAuthPepper,
		SessionCookieSecret:   fullFunnelSessionSecret,
	})
	if _, err := service.StartEmailChallenge(ctx, auth.StartEmailChallengeInput{
		Email:      email,
		RemoteAddr: "127.0.0.1:12345",
		UserAgent:  "full-funnel-seed-test",
	}); err != nil {
		t.Fatalf("start full-funnel auth challenge: %v", err)
	}
	verified, err := service.VerifyEmailChallenge(ctx, auth.VerifyEmailChallengeInput{
		Token:      challengeToken,
		RemoteAddr: "127.0.0.1:12345",
		UserAgent:  "full-funnel-seed-test",
	})
	if err != nil {
		t.Fatalf("verify full-funnel auth challenge: %v", err)
	}
	return &http.Cookie{Name: auth.SessionCookieName, Value: verified.SessionToken}, verified.UserID
}

func cleanupFullFunnelScenarioEmail(t *testing.T, db *sql.DB, email string) {
	t.Helper()
	rows, err := db.Query(`select id::text from users where email = $1`, email)
	if err != nil {
		t.Fatalf("query stale full-funnel user: %v", err)
	}
	defer rows.Close()
	var userIDs []string
	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			t.Fatalf("scan stale full-funnel user: %v", err)
		}
		userIDs = append(userIDs, userID)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate stale full-funnel users: %v", err)
	}
	for _, userID := range userIDs {
		cleanupFullFunnelScenarioUser(t, db, userID)
	}
	_, _ = db.Exec(`delete from auth_challenges where email = $1`, email)
}

func cleanupFullFunnelScenarioUser(t *testing.T, db *sql.DB, userID string) {
	t.Helper()
	if userID == "" {
		return
	}
	cleanupQueries := []string{
		`
with owned_resources as (
  select id from resumes where user_id = $1
  union select id from target_jobs where user_id = $1
  union select id from practice_plans where user_id = $1
  union select id from practice_sessions where user_id = $1
  union select id from feedback_reports where user_id = $1
)
delete from outbox_events
where aggregate_id in (select id from owned_resources)`,
		`
with owned_resources as (
  select id from resumes where user_id = $1
  union select id from target_jobs where user_id = $1
  union select id from practice_plans where user_id = $1
  union select id from practice_sessions where user_id = $1
  union select id from feedback_reports where user_id = $1
)
delete from async_jobs
where resource_id in (select id from owned_resources)`,
		`
with owned_resources as (
  select id from resumes where user_id = $1
  union select id from target_jobs where user_id = $1
  union select id from practice_plans where user_id = $1
  union select id from practice_sessions where user_id = $1
  union select id from feedback_reports where user_id = $1
)
delete from ai_task_runs
where user_id = $1 or resource_id in (select id from owned_resources)`,
		`delete from idempotency_records where user_id = $1`,
		`delete from auth_challenges where user_id = $1`,
		`delete from sessions where user_id = $1`,
		`delete from users where id = $1`,
	}
	for _, query := range cleanupQueries {
		if _, err := db.Exec(query, userID); err != nil {
			t.Fatalf("cleanup full-funnel user %s failed: %v", userID, err)
		}
	}
}

func assertFullFunnelCount(t *testing.T, db *sql.DB, query string, argsAndWant ...any) {
	t.Helper()
	if len(argsAndWant) < 1 {
		t.Fatal("assertFullFunnelCount requires expected count")
	}
	want, ok := argsAndWant[len(argsAndWant)-1].(int)
	if !ok {
		t.Fatalf("last assertFullFunnelCount argument must be int, got %T", argsAndWant[len(argsAndWant)-1])
	}
	var got int
	if err := db.QueryRow(query, argsAndWant[:len(argsAndWant)-1]...).Scan(&got); err != nil {
		t.Fatalf("count query failed: %v", err)
	}
	if got != want {
		t.Fatalf("count query got %d, want %d: %s", got, want, query)
	}
}

func fullFunnelStringPtr(v string) *string {
	return &v
}

func fullFunnelBoolPtr(v bool) *bool {
	return &v
}
