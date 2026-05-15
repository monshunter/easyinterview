package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient"
	"github.com/monshunter/easyinterview/backend/internal/ai/registry"
	apireports "github.com/monshunter/easyinterview/backend/internal/api/reports"
	"github.com/monshunter/easyinterview/backend/internal/auth"
	"github.com/monshunter/easyinterview/backend/internal/platform/config"
	reviewdomain "github.com/monshunter/easyinterview/backend/internal/review"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
	"github.com/monshunter/easyinterview/backend/internal/targetjob"
)

const (
	reportScenarioUserAID = "0197d120-0000-7000-8000-0000000009a1"
	reportScenarioUserBID = "0197d120-0000-7000-8000-0000000009b1"
)

func TestE2EP0052ReportGenerationHappyPath(t *testing.T) {
	now := time.Date(2026, 5, 15, 22, 0, 0, 0, time.UTC)
	reportID := "0197d120-0000-7000-8000-000000000a52"
	repo := &reportScenarioRepository{contexts: map[string]reviewdomain.ReportContext{
		reportID: reportScenarioContext(reportID, 3),
	}}
	ai := &reportScenarioAI{}
	service := reviewdomain.NewService(reviewdomain.ServiceOptions{
		Registry:   reportScenarioRegistry{},
		AI:         ai,
		AITaskRuns: &reportScenarioTaskRuns{},
		Repository: repo,
		Now:        func() time.Time { return now },
		NewID:      reportScenarioIDs("outbox-p0-052", "audit-p0-052"),
	})
	store := &reportScenarioRunnerStore{
		job: reviewdomain.AsyncJob{
			JobID:       "0197d120-0000-7000-8000-000000000b52",
			JobType:     reviewdomain.ReportGenerateJobType,
			ResourceID:  reportID,
			Attempts:    1,
			MaxAttempts: 5,
			AvailableAt: now,
		},
		ok: true,
	}
	runner := reviewdomain.NewRunner(reviewdomain.RunnerOptions{
		Store:   store,
		Service: service,
		Now:     func() time.Time { return now },
	})

	processed, err := runner.RunOnce(context.Background())
	if err != nil || !processed {
		t.Fatalf("RunOnce processed=%v err=%v", processed, err)
	}
	if store.statusUpdate.From != sharedtypes.ReportStatusQueued || store.statusUpdate.To != sharedtypes.ReportStatusGenerating {
		t.Fatalf("status update = %+v", store.statusUpdate)
	}
	if store.succeededJobID != "" || store.failed.JobID != "" {
		t.Fatalf("runner updated async job after service finalization: succeeded=%q failed=%+v", store.succeededJobID, store.failed)
	}
	if repo.persisted.ReportID != reportID || repo.persisted.AsyncJobID != store.job.JobID || repo.persisted.PreparednessLevel == "" || len(repo.persisted.Assessments) != 3 {
		t.Fatalf("persisted result = %+v", repo.persisted)
	}
	if len(ai.payloads) != 4 ||
		ai.payloads[0].Metadata.TaskRun.Capability != aiclient.AITaskRunTaskReportGenerate ||
		ai.payloads[1].Metadata.TaskRun.Capability != aiclient.AITaskRunTaskReportAssessment {
		t.Fatalf("AI payloads = %+v", ai.payloads)
	}
	assertNoReportScenarioLeak(t, mustMarshal(t, repo.persisted.Content), "question_text", "answer_text", "hint_text", "prompt body", "response body")
}

func TestE2EP0053ReportReadAndListing(t *testing.T) {
	h := newReportHTTPScenarioHarness(t)
	now := time.Date(2026, 5, 15, 22, 30, 0, 0, time.UTC)
	readyID := reportScenarioUUID(1)
	queuedID := reportScenarioUUID(2)
	generatingID := reportScenarioUUID(3)
	failedID := reportScenarioUUID(4)
	h.service.reports[readyID] = reportScenarioRecord(readyID, "target-a", sharedtypes.ReportStatusReady, now)
	h.service.reports[queuedID] = reportScenarioRecord(queuedID, "target-a", sharedtypes.ReportStatusQueued, now.Add(-time.Minute))
	h.service.reports[generatingID] = reportScenarioRecord(generatingID, "target-a", sharedtypes.ReportStatusGenerating, now.Add(-2*time.Minute))
	failed := reportScenarioRecord(failedID, "target-a", sharedtypes.ReportStatusFailed, now.Add(-3*time.Minute))
	code := sharederrors.CodeAiProviderTimeout
	failed.ErrorCode = &code
	failed.PreparednessLevel = nil
	failed.Provenance = nil
	h.service.reports[failed.ID] = failed
	for i := 0; i < 18; i++ {
		id := reportScenarioUUID(10 + i)
		h.service.reports[id] = reportScenarioRecord(id, "target-a", sharedtypes.ReportStatusReady, now.Add(-time.Duration(i+4)*time.Minute))
	}

	var ready map[string]any
	decodeJSON(t, h.doJSON(t, reportScenarioUserAID, http.MethodGet, "/api/v1/reports/"+readyID, http.StatusOK), &ready)
	if ready["status"] != string(sharedtypes.ReportStatusReady) || ready["provenance"] == nil {
		t.Fatalf("ready report = %+v", ready)
	}
	assertStrictProvenanceKeys(t, ready)

	var queued map[string]any
	decodeJSON(t, h.doJSON(t, reportScenarioUserAID, http.MethodGet, "/api/v1/reports/"+queuedID, http.StatusOK), &queued)
	if queued["status"] != string(sharedtypes.ReportStatusQueued) || queued["provenance"] != nil {
		t.Fatalf("queued placeholder = %+v", queued)
	}

	var failedOut map[string]any
	decodeJSON(t, h.doJSON(t, reportScenarioUserAID, http.MethodGet, "/api/v1/reports/"+failedID, http.StatusOK), &failedOut)
	if failedOut["status"] != string(sharedtypes.ReportStatusFailed) || failedOut["errorCode"] != sharederrors.CodeAiProviderTimeout {
		t.Fatalf("failed report = %+v", failedOut)
	}

	var firstPage struct {
		Items    []map[string]any `json:"items"`
		PageInfo struct {
			NextCursor *string `json:"nextCursor"`
			PageSize   int     `json:"pageSize"`
			HasMore    bool    `json:"hasMore"`
		} `json:"pageInfo"`
	}
	decodeJSON(t, h.doJSON(t, reportScenarioUserAID, http.MethodGet, "/api/v1/targets/target-a/reports?pageSize=20", http.StatusOK), &firstPage)
	if len(firstPage.Items) != 20 || firstPage.PageInfo.NextCursor == nil || !firstPage.PageInfo.HasMore {
		t.Fatalf("first page = %+v", firstPage.PageInfo)
	}
	var secondPage struct {
		Items    []map[string]any `json:"items"`
		PageInfo struct {
			NextCursor *string `json:"nextCursor"`
			HasMore    bool    `json:"hasMore"`
		} `json:"pageInfo"`
	}
	decodeJSON(t, h.doJSON(t, reportScenarioUserAID, http.MethodGet, "/api/v1/targets/target-a/reports?pageSize=20&cursor="+*firstPage.PageInfo.NextCursor, http.StatusOK), &secondPage)
	if len(secondPage.Items) != 2 || secondPage.PageInfo.HasMore || secondPage.PageInfo.NextCursor != nil {
		t.Fatalf("second page = items=%d pageInfo=%+v", len(secondPage.Items), secondPage.PageInfo)
	}
	var invalid apiErrorEnvelope
	decodeJSON(t, h.doJSON(t, reportScenarioUserAID, http.MethodGet, "/api/v1/targets/target-a/reports?cursor=bad", http.StatusBadRequest), &invalid)
	if invalid.Error.Code != sharederrors.CodeValidationFailed {
		t.Fatalf("invalid cursor error = %+v", invalid)
	}
	var cross apiErrorEnvelope
	decodeJSON(t, h.doJSON(t, reportScenarioUserBID, http.MethodGet, "/api/v1/reports/"+readyID, http.StatusNotFound), &cross)
	if cross.Error.Code != sharederrors.CodeReportNotFound {
		t.Fatalf("cross-user error = %+v", cross)
	}
}

func TestE2EP0054ReportAIFailureAndRetry(t *testing.T) {
	now := time.Date(2026, 5, 15, 23, 0, 0, 0, time.UTC)
	cases := []struct {
		name         string
		registryErrs map[string]error
		aiErrs       []error
		aiResponses  []string
		wantCode     string
		wantRuns     int
	}{
		{name: "f3_generate", registryErrs: map[string]error{"report.generate": registry.ErrPromptUnsupported}, wantCode: sharederrors.CodeAiProviderConfigInvalid, wantRuns: 1},
		{name: "f3_assessment", registryErrs: map[string]error{"report.question_assessment": registry.ErrLanguageUnsupported}, aiResponses: []string{reportGenerateJSON()}, wantCode: sharederrors.CodeAiProviderConfigInvalid, wantRuns: 1},
		{name: "secret", aiErrs: []error{sharederrors.Wrap(sharederrors.CodeAiProviderSecretMissing, "provider secret", false)}, wantCode: sharederrors.CodeAiProviderSecretMissing},
		{name: "timeout", aiErrs: []error{sharederrors.Wrap(sharederrors.CodeAiProviderTimeout, "timeout", true)}, wantCode: sharederrors.CodeAiProviderTimeout},
		{name: "invalid", aiErrs: []error{sharederrors.Wrap(sharederrors.CodeAiOutputInvalid, "invalid output", false)}, wantCode: sharederrors.CodeAiOutputInvalid},
		{name: "empty", aiResponses: []string{`{}`}, wantCode: sharederrors.CodeAiOutputInvalid, wantRuns: 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			reportID := "report-" + tc.name
			repo := &reportScenarioRepository{contexts: map[string]reviewdomain.ReportContext{reportID: reportScenarioContext(reportID, 1)}}
			runs := &reportScenarioTaskRuns{}
			ai := &reportScenarioAI{errs: tc.aiErrs, responses: tc.aiResponses}
			service := reviewdomain.NewService(reviewdomain.ServiceOptions{
				Registry:   reportScenarioRegistry{errs: tc.registryErrs},
				AI:         ai,
				AITaskRuns: runs,
				Repository: repo,
				Now:        func() time.Time { return now },
				NewID:      reportScenarioIDs("run-"+tc.name, "outbox-"+tc.name, "audit-"+tc.name),
			})
			outcome := service.GenerateReport(context.Background(), reviewdomain.AsyncJob{JobID: "job-" + tc.name, ResourceID: reportID, Attempts: 1, MaxAttempts: 5})
			if outcome.Succeeded || outcome.ErrorCode != tc.wantCode {
				t.Fatalf("outcome = %+v, want %s", outcome, tc.wantCode)
			}
			if !outcome.AsyncJobFinalized {
				t.Fatalf("outcome.AsyncJobFinalized = false, want true")
			}
			if repo.failed.ErrorCode != tc.wantCode {
				t.Fatalf("persisted failure = %+v", repo.failed)
			}
			if len(runs.rows) != tc.wantRuns {
				t.Fatalf("explicit runs = %d, want %d", len(runs.rows), tc.wantRuns)
			}
			for _, row := range runs.rows {
				if row.Status == "succeeded" || row.Status != aiclient.AITaskRunStatusFailed {
					t.Fatalf("ai_task_run status = %q", row.Status)
				}
			}
		})
	}

	store := &reportScenarioRunnerStore{
		job: reviewdomain.AsyncJob{JobID: "job-permanent", ResourceID: "report-permanent", Attempts: 5, MaxAttempts: 5},
		ok:  true,
	}
	runner := reviewdomain.NewRunner(reviewdomain.RunnerOptions{
		Store:   store,
		Service: reportScenarioOutcomeService{outcome: reviewdomain.ReportOutcome{ErrorCode: sharederrors.CodeAiProviderTimeout, ErrorMessage: "timeout", Retryable: true}},
		Now:     func() time.Time { return now },
	})
	if _, err := runner.RunOnce(context.Background()); err != nil {
		t.Fatalf("permanent RunOnce: %v", err)
	}
	if store.failed.AvailableAt.Sub(now) != 16*time.Minute || store.failed.ErrorCode != sharederrors.CodeAiProviderTimeout {
		t.Fatalf("permanent failed update = %+v", store.failed)
	}
}

func TestE2EP0055ReportPrivacyAndLegacy(t *testing.T) {
	h := newReportHTTPScenarioHarness(t)
	now := time.Date(2026, 5, 15, 23, 30, 0, 0, time.UTC)
	privateID := reportScenarioUUID(55)
	h.service.reports[privateID] = reportScenarioRecord(privateID, "target-private", sharedtypes.ReportStatusReady, now)

	var cross apiErrorEnvelope
	raw := h.doJSON(t, reportScenarioUserBID, http.MethodGet, "/api/v1/reports/"+privateID, http.StatusNotFound)
	decodeJSON(t, raw, &cross)
	if cross.Error.Code != sharederrors.CodeReportNotFound || strings.Contains(string(raw), privateID) {
		t.Fatalf("cross-user response leaked report existence: %s", string(raw))
	}
	assertNoReportScenarioLeak(t, raw, "question_text", "answer_text", "hint_text", "prompt body", "response body", "provider secret")

	cmd := exec.Command("python3", "../../../scripts/lint/backend_review_legacy.py", "--repo-root", "../../..", "--phase", "all")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("backend_review_legacy failed: %v\n%s", err, string(out))
	}
}

type reportHTTPScenarioHarness struct {
	handler http.Handler
	service *reportScenarioHTTPService
	cookies map[string]*http.Cookie
}

func newReportHTTPScenarioHarness(t *testing.T) *reportHTTPScenarioHarness {
	t.Helper()
	dir := t.TempDir()
	writeAPIFile(t, filepath.Join(dir, "config.yaml"), `
runtime:
  appVersion: "1.2.3"
  defaultUiLanguage: zh-CN
`)
	loader, err := config.Load(config.Options{AppEnv: "test", ConfigDir: dir})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	authStore := newPracticeScenarioAuthStore("report-scenario-secret")
	cookies := map[string]*http.Cookie{
		reportScenarioUserAID: authStore.addSession(reportScenarioUserAID, "candidate-a@example.com", "report-token-a"),
		reportScenarioUserBID: authStore.addSession(reportScenarioUserBID, "candidate-b@example.com", "report-token-b"),
	}
	authService := auth.NewPasswordlessService(auth.PasswordlessServiceOptions{
		Store:               authStore,
		SessionCookieSecret: "report-scenario-secret",
		Now:                 fixedScenarioNow,
	})
	service := &reportScenarioHTTPService{reports: map[string]reviewdomain.FeedbackReportRecord{}}
	handler := buildAPIHandlerWithUploadReportAndHandlers(loader, apiRuntimeFlags{}, authService, targetjob.NewHandler(), practiceRoutes{}, uploadRoutes{}, resumeRoutes{}, reportRoutes{
		Handler: apireports.NewHandler(apireports.HandlerOptions{Service: service, Session: currentUserFromContext}),
	})
	return &reportHTTPScenarioHarness{handler: handler, service: service, cookies: cookies}
}

func (h *reportHTTPScenarioHarness) doJSON(t *testing.T, userID, method, path string, wantStatus int) []byte {
	t.Helper()
	req := httptest.NewRequest(method, path, bytes.NewReader(nil))
	cookie, ok := h.cookies[userID]
	if !ok {
		t.Fatalf("missing cookie for %s", userID)
	}
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()
	h.handler.ServeHTTP(rec, req)
	if rec.Code != wantStatus {
		t.Fatalf("%s %s status=%d want=%d body=%s", method, path, rec.Code, wantStatus, rec.Body.String())
	}
	return rec.Body.Bytes()
}

type reportScenarioHTTPService struct {
	reports map[string]reviewdomain.FeedbackReportRecord
}

func (s *reportScenarioHTTPService) GetFeedbackReport(_ context.Context, userID, reportID string) (reviewdomain.FeedbackReportRecord, error) {
	report, ok := s.reports[reportID]
	if !ok || userID != reportScenarioUserAID {
		return reviewdomain.FeedbackReportRecord{}, reviewdomain.ErrReportNotFound
	}
	return report, nil
}

func (s *reportScenarioHTTPService) ListTargetJobReports(_ context.Context, in reviewdomain.ListTargetJobReportsRequest) (reviewdomain.PaginatedFeedbackReportRecord, error) {
	if in.UserID != reportScenarioUserAID {
		return reviewdomain.PaginatedFeedbackReportRecord{}, reviewdomain.ErrReportNotFound
	}
	pageSize := reviewdomain.EffectiveReportPageSize(in.PageSize)
	items := make([]reviewdomain.FeedbackReportRecord, 0, len(s.reports))
	for _, report := range s.reports {
		if report.TargetJobID == in.TargetJobID {
			items = append(items, report)
		}
	}
	sortReports(items)
	start := 0
	if strings.TrimSpace(in.Cursor) != "" {
		createdAt, id, err := reviewdomain.DecodeCursor(in.Cursor)
		if err != nil {
			return reviewdomain.PaginatedFeedbackReportRecord{}, reviewdomain.ErrInvalidCursor
		}
		for i, item := range items {
			if item.CreatedAt.Before(createdAt) || (item.CreatedAt.Equal(createdAt) && item.ID < id) {
				start = i
				break
			}
		}
	}
	end := start + pageSize
	hasMore := end < len(items)
	if end > len(items) {
		end = len(items)
	}
	page := append([]reviewdomain.FeedbackReportRecord(nil), items[start:end]...)
	nextCursor := ""
	if hasMore && len(page) > 0 {
		last := page[len(page)-1]
		nextCursor = reviewdomain.EncodeCursor(last.CreatedAt, last.ID)
	}
	return reviewdomain.PaginatedFeedbackReportRecord{Items: page, PageInfo: reviewdomain.PageInfo{NextCursor: nextCursor, PageSize: pageSize, HasMore: hasMore}}, nil
}

func sortReports(items []reviewdomain.FeedbackReportRecord) {
	for i := 0; i < len(items); i++ {
		for j := i + 1; j < len(items); j++ {
			if items[j].CreatedAt.After(items[i].CreatedAt) || (items[j].CreatedAt.Equal(items[i].CreatedAt) && items[j].ID > items[i].ID) {
				items[i], items[j] = items[j], items[i]
			}
		}
	}
}

func reportScenarioRecord(id, targetID string, status sharedtypes.ReportStatus, createdAt time.Time) reviewdomain.FeedbackReportRecord {
	tier := sharedtypes.ReadinessTierBasicallyReady
	report := reviewdomain.FeedbackReportRecord{
		ID:                id,
		SessionID:         "session-" + id,
		TargetJobID:       targetID,
		Status:            status,
		PreparednessLevel: &tier,
		Highlights:        []reviewdomain.ReportEvidenceRecord{{Dimension: "depth", Evidence: "clear", Confidence: sharedtypes.ConfidenceHigh}},
		Issues:            []reviewdomain.ReportEvidenceRecord{{Dimension: "depth", Evidence: "add tradeoff", Confidence: sharedtypes.ConfidenceMedium}},
		NextActions:       []reviewdomain.ReportNextActionRecord{{Type: string(reviewdomain.NextActionNextRound), Label: "Next"}},
		QuestionAssessments: []reviewdomain.QuestionAssessmentRecord{{
			TurnID:              "turn-" + id,
			QuestionIntent:      "architecture",
			DimensionResults:    map[string]reviewdomain.DimensionResultRecord{"depth": {Status: sharedtypes.DimensionStatusMeetsBar, Confidence: sharedtypes.ConfidenceHigh}},
			ReviewStatus:        sharedtypes.QuestionReviewStatusOpen,
			IncludedInRetryPlan: false,
		}},
		Provenance: &reviewdomain.GenerationProvenanceRecord{
			PromptVersion: "v0.1.0", RubricVersion: "v0.1.0", ModelID: "model-profile:report.generate.default",
			Language: "en", FeatureFlag: "none", DataSourceVersion: "registry.v1",
		},
		CreatedAt: createdAt,
		UpdatedAt: createdAt,
	}
	if status == sharedtypes.ReportStatusQueued || status == sharedtypes.ReportStatusGenerating || status == sharedtypes.ReportStatusFailed {
		report.PreparednessLevel = nil
		report.Highlights = []reviewdomain.ReportEvidenceRecord{}
		report.Issues = []reviewdomain.ReportEvidenceRecord{}
		report.NextActions = []reviewdomain.ReportNextActionRecord{}
		report.QuestionAssessments = []reviewdomain.QuestionAssessmentRecord{}
		report.Provenance = nil
	}
	return report
}

func reportScenarioUUID(n int) string {
	return fmt.Sprintf("0197d120-0000-7000-8000-%012d", n)
}

type reportScenarioRepository struct {
	contexts  map[string]reviewdomain.ReportContext
	persisted reviewdomain.ReportResultPersistence
	failed    reviewdomain.ReportFailurePersistence
}

func (r *reportScenarioRepository) LoadReportContext(_ context.Context, reportID string) (reviewdomain.ReportContext, error) {
	ctx, ok := r.contexts[reportID]
	if !ok {
		return reviewdomain.ReportContext{}, errors.New("missing report context")
	}
	return ctx, nil
}

func (r *reportScenarioRepository) PersistReportResult(_ context.Context, in reviewdomain.ReportResultPersistence) error {
	r.persisted = in
	return nil
}

func (r *reportScenarioRepository) PersistReportFailure(_ context.Context, in reviewdomain.ReportFailurePersistence) error {
	r.failed = in
	return nil
}

type reportScenarioRegistry struct {
	errs map[string]error
}

func (r reportScenarioRegistry) ResolveActive(_ context.Context, featureKey, _ string) (registry.PromptResolution, error) {
	if err := r.errs[featureKey]; err != nil {
		return registry.PromptResolution{}, err
	}
	profile := "report.generate.default"
	template := "Session metadata: {{session_metadata}}\nTurn summaries: {{turn_summaries}}\nReturn strict JSON."
	if featureKey == "report.question_assessment" {
		profile = "report.assessment.default"
		template = "Question context: {{question_context}}\nAnswer summary: {{answer_summary}}\nReturn strict JSON."
	}
	return registry.PromptResolution{
		FeatureKey: featureKey, PromptVersion: "v0.1.0", RubricVersion: "v0.1.0", ModelProfileName: profile,
		FeatureFlag: "none", DataSourceVersion: "registry.v1", UserMessageTemplate: template,
	}, nil
}

type reportScenarioAI struct {
	responses []string
	errs      []error
	payloads  []aiclient.CompletePayload
}

func (a *reportScenarioAI) Complete(_ context.Context, _ string, payload aiclient.CompletePayload) (aiclient.CompleteResponse, aiclient.AICallMeta, error) {
	a.payloads = append(a.payloads, payload)
	idx := len(a.payloads) - 1
	if idx < len(a.errs) && a.errs[idx] != nil {
		return aiclient.CompleteResponse{}, aiclient.AICallMeta{}, a.errs[idx]
	}
	if idx < len(a.responses) {
		return aiclient.CompleteResponse{Content: a.responses[idx]}, aiclient.AICallMeta{}, nil
	}
	if payload.Metadata.FeatureKey == "report.generate" {
		return aiclient.CompleteResponse{Content: reportGenerateJSON()}, aiclient.AICallMeta{}, nil
	}
	return aiclient.CompleteResponse{Content: `{"dimension_results":{"depth":{"status":"meets_bar","confidence":0.8,"score":0.8}},"overall_status":"meets_bar","confidence":0.8,"strengths":["clear"],"gaps":[],"recommended_framework":"STAR","review_status":"open"}`}, aiclient.AICallMeta{}, nil
}

type reportScenarioTaskRuns struct {
	rows []aiclient.AITaskRunRow
}

func (w *reportScenarioTaskRuns) WriteAITaskRun(_ context.Context, row aiclient.AITaskRunRow) error {
	w.rows = append(w.rows, row)
	return nil
}

type reportScenarioRunnerStore struct {
	job            reviewdomain.AsyncJob
	ok             bool
	statusUpdate   reviewdomain.ReportStatusUpdate
	succeededJobID string
	failed         reviewdomain.AsyncJobFailure
}

func (s *reportScenarioRunnerStore) LeaseAsyncJob(_ context.Context, _ string, now time.Time) (reviewdomain.AsyncJob, bool, error) {
	if !s.ok {
		return reviewdomain.AsyncJob{}, false, nil
	}
	locked := now
	s.job.LockedAt = &locked
	return s.job, true, nil
}

func (s *reportScenarioRunnerStore) UpdateFeedbackReportStatus(_ context.Context, update reviewdomain.ReportStatusUpdate) error {
	s.statusUpdate = update
	return nil
}

func (s *reportScenarioRunnerStore) UpdateAsyncJobSucceeded(_ context.Context, jobID string, _ time.Time) error {
	s.succeededJobID = jobID
	return nil
}

func (s *reportScenarioRunnerStore) UpdateAsyncJobFailed(_ context.Context, in reviewdomain.AsyncJobFailure) error {
	s.failed = in
	return nil
}

func (s *reportScenarioRunnerStore) ReclaimExpiredLeases(context.Context, string, time.Time, time.Time) (int64, error) {
	return 0, nil
}

type reportScenarioOutcomeService struct {
	outcome reviewdomain.ReportOutcome
}

func (s reportScenarioOutcomeService) GenerateReport(context.Context, reviewdomain.AsyncJob) reviewdomain.ReportOutcome {
	return s.outcome
}

type apiErrorEnvelope struct {
	Error struct {
		Code string `json:"code"`
	} `json:"error"`
}

func reportScenarioContext(reportID string, turnCount int) reviewdomain.ReportContext {
	turns := make([]reviewdomain.TurnSnapshot, 0, turnCount)
	for i := 1; i <= turnCount; i++ {
		turns = append(turns, reviewdomain.TurnSnapshot{ID: "turn-" + reportID + "-" + string(rune('0'+i)), TurnIndex: i, QuestionIntent: "architecture", QuestionContext: "service boundary", AnswerSummary: "clear tradeoff"})
	}
	return reviewdomain.ReportContext{
		Session: reviewdomain.SessionSnapshot{UserID: reportScenarioUserAID, ReportID: reportID, SessionID: "session-" + reportID, PlanID: "plan-" + reportID, TargetJobID: "target-" + reportID, Language: "en"},
		Plan:    reviewdomain.PracticePlanSnapshot{ID: "plan-" + reportID, Goal: "baseline", Mode: "strict", InterviewerPersona: "technical_manager"},
		Turns:   turns,
		Rubric:  registry.RubricSchema{Dimensions: []registry.RubricDimension{{Name: "depth", Weight: 1}}},
	}
}

func reportGenerateJSON() string {
	return `{"summary":"ok","highlights":[{"dimension":"depth","evidence":"clear","confidence":0.8}],"issues":[{"dimension":"depth","evidence":"add tradeoff","confidence":0.5}],"next_actions":[{"type":"next_round","label":"Next"}]}`
}

func reportScenarioIDs(ids ...string) func() string {
	i := 0
	return func() string {
		if i >= len(ids) {
			return "extra-id"
		}
		id := ids[i]
		i++
		return id
	}
}

func mustMarshal(t *testing.T, v any) []byte {
	t.Helper()
	raw, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return raw
}

func assertStrictProvenanceKeys(t *testing.T, report map[string]any) {
	t.Helper()
	provenance, ok := report["provenance"].(map[string]any)
	if !ok {
		t.Fatalf("missing provenance: %+v", report)
	}
	want := map[string]bool{"promptVersion": true, "rubricVersion": true, "modelId": true, "language": true, "featureFlag": true, "dataSourceVersion": true}
	if len(provenance) != len(want) {
		t.Fatalf("provenance keys = %+v", provenance)
	}
	for key := range provenance {
		if !want[key] {
			t.Fatalf("unexpected provenance key %q in %+v", key, provenance)
		}
	}
}

func assertNoReportScenarioLeak(t *testing.T, raw []byte, forbidden ...string) {
	t.Helper()
	lower := strings.ToLower(string(raw))
	for _, token := range forbidden {
		if strings.Contains(lower, strings.ToLower(token)) {
			t.Fatalf("payload leaked %q: %s", token, string(raw))
		}
	}
}
