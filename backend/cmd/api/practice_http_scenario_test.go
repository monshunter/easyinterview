package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient"
	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient/observability"
	"github.com/monshunter/easyinterview/backend/internal/ai/registry"
	api "github.com/monshunter/easyinterview/backend/internal/api/generated"
	apipractice "github.com/monshunter/easyinterview/backend/internal/api/practice"
	"github.com/monshunter/easyinterview/backend/internal/auth"
	"github.com/monshunter/easyinterview/backend/internal/middleware/idempotency"
	"github.com/monshunter/easyinterview/backend/internal/platform/config"
	domainpractice "github.com/monshunter/easyinterview/backend/internal/practice"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
	sharedevents "github.com/monshunter/easyinterview/backend/internal/shared/events"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
	storepractice "github.com/monshunter/easyinterview/backend/internal/store/practice"
	"github.com/monshunter/easyinterview/backend/internal/targetjob"
)

const (
	practiceHTTPScenarioUserAID = "01918fa0-0000-7000-8000-0000000000a1"
	practiceHTTPScenarioUserBID = "01918fa0-0000-7000-8000-0000000000b1"
)

func TestE2EP0022PracticePlanBaselineCreateAndRead(t *testing.T) {
	h := newPracticeHTTPScenarioHarness(t)
	body := api.CreatePracticePlanRequest{
		TargetJobId:          "target-job-p0-022-a",
		ResumeAssetId:        strPtr("resume-asset-p0-022-a"),
		Goal:                 sharedtypes.PracticeGoalBaseline,
		Mode:                 sharedtypes.PracticeModeAssisted,
		InterviewerPersona:   sharedtypes.InterviewerRoleHiringManager,
		Difficulty:           "standard",
		Language:             "zh-CN",
		TimeBudgetMinutes:    30,
		QuestionBudget:       6,
		FocusCompetencyCodes: []string{"communication", "design-systems"},
	}

	raw := h.doJSON(t, practiceHTTPScenarioUserAID, http.MethodPost, "/api/v1/practice/plans", "e2e-p0-022-create-plan", body, http.StatusCreated)
	var created api.PracticePlan
	decodeJSON(t, raw, &created)
	if created.Id == "" || created.Status != "ready" || created.Goal != sharedtypes.PracticeGoalBaseline || created.TargetJobId != body.TargetJobId {
		t.Fatalf("unexpected createPracticePlan response: %+v", created)
	}
	if h.store.planCount() != 1 || h.store.auditCount() != 1 {
		t.Fatalf("createPracticePlan should write one plan and one audit row, plans=%d audits=%d", h.store.planCount(), h.store.auditCount())
	}

	duplicateRaw := h.doJSON(t, practiceHTTPScenarioUserAID, http.MethodPost, "/api/v1/practice/plans", "e2e-p0-022-create-plan", body, http.StatusCreated)
	var duplicate api.PracticePlan
	decodeJSON(t, duplicateRaw, &duplicate)
	if duplicate.Id != created.Id || h.store.planCount() != 1 || h.store.auditCount() != 1 {
		t.Fatalf("idempotency replay should not duplicate side effects: duplicate=%+v plans=%d audits=%d", duplicate, h.store.planCount(), h.store.auditCount())
	}

	detailRaw := h.doJSON(t, practiceHTTPScenarioUserAID, http.MethodGet, "/api/v1/practice/plans/"+created.Id, "", nil, http.StatusOK)
	var detail api.PracticePlan
	decodeJSON(t, detailRaw, &detail)
	if detail.Id != created.Id || detail.Status != "ready" || detail.QuestionBudget != body.QuestionBudget {
		t.Fatalf("getPracticePlan did not return the created plan: %+v", detail)
	}

	crossUserRaw := h.doJSON(t, practiceHTTPScenarioUserBID, http.MethodGet, "/api/v1/practice/plans/"+created.Id, "", nil, http.StatusNotFound)
	var crossUser api.ApiErrorResponse
	decodeJSON(t, crossUserRaw, &crossUser)
	if crossUser.Error.Code != sharederrors.CodePracticePlanNotFound || crossUser.Error.Retryable {
		t.Fatalf("cross-user getPracticePlan should hide existence with PRACTICE_PLAN_NOT_FOUND: %+v", crossUser.Error)
	}

	assertNoEvidenceLeak(t, h.store.auditPayloads(), "question_text", "answer_text", "hint_text", "prompt body", "response body", "legacy debrief replay value")
}

func TestE2EP0023PracticeSessionStartAndFirstQuestion(t *testing.T) {
	h := newPracticeHTTPScenarioHarness(t)
	h.store.seedReadyPlan(domainpractice.CreatePlanStoreInput{
		PlanID:               "practice-plan-p0-023",
		AuditEventID:         "audit-p0-023",
		UserID:               practiceHTTPScenarioUserAID,
		TargetJobID:          "target-job-p0-023-a",
		ResumeAssetID:        "resume-asset-p0-023-a",
		Goal:                 sharedtypes.PracticeGoalBaseline,
		Mode:                 sharedtypes.PracticeModeAssisted,
		InterviewerPersona:   sharedtypes.InterviewerRoleHiringManager,
		Difficulty:           "standard",
		Language:             "zh-CN",
		TimeBudgetMinutes:    30,
		QuestionBudget:       6,
		FocusCompetencyCodes: []string{"system-design"},
		Now:                  fixedScenarioNow(),
	})

	raw := h.doJSON(t, practiceHTTPScenarioUserAID, http.MethodPost, "/api/v1/practice/sessions", "e2e-p0-023-start-session", api.StartPracticeSessionRequest{
		PlanId:       "practice-plan-p0-023",
		HintsEnabled: practiceBoolPtr(true),
	}, http.StatusCreated)
	var started api.PracticeSession
	decodeJSON(t, raw, &started)
	if started.Id == "" || started.Status != sharedtypes.SessionStatusRunning || started.CurrentTurn == nil {
		t.Fatalf("unexpected startPracticeSession response: %+v", started)
	}
	if started.CurrentTurn.TurnIndex != 1 || started.CurrentTurn.Status != "asked" || strings.TrimSpace(started.CurrentTurn.QuestionText) == "" {
		t.Fatalf("first turn was not synchronously returned: %+v", started.CurrentTurn)
	}
	if started.CurrentTurn.QuestionIntent == nil || *started.CurrentTurn.QuestionIntent != "behavioral.leadership.design_system" {
		t.Fatalf("first turn intent mismatch: %+v", started.CurrentTurn)
	}

	detailRaw := h.doJSON(t, practiceHTTPScenarioUserAID, http.MethodGet, "/api/v1/practice/sessions/"+started.Id, "", nil, http.StatusOK)
	var detail api.PracticeSession
	decodeJSON(t, detailRaw, &detail)
	if detail.Id != started.Id || detail.CurrentTurn == nil || detail.CurrentTurn.Id != started.CurrentTurn.Id {
		t.Fatalf("getPracticeSession did not return the running session with current turn: %+v", detail)
	}
	if h.store.turnCount(started.Id) != 1 || h.store.sessionEventCount(started.Id) != 1 || h.store.outboxCount() != 1 {
		t.Fatalf("session start should persist one turn, one event, and one outbox row: turns=%d events=%d outbox=%d", h.store.turnCount(started.Id), h.store.sessionEventCount(started.Id), h.store.outboxCount())
	}
	if h.store.aiCalledInsideTransaction {
		t.Fatalf("AI call must happen outside the repository transaction window")
	}
	status := h.store.idempotencyStatus(practiceHTTPScenarioUserAID, "practice", "startPracticeSession", "e2e-p0-023-start-session")
	if status != idempotency.StatusSucceeded {
		t.Fatalf("startPracticeSession idempotency status = %q, want %q", status, idempotency.StatusSucceeded)
	}

	var startedPayload sharedevents.PracticeSessionStartedPayload
	decodeJSON(t, h.store.outboxPayloads()[0], &startedPayload)
	if startedPayload.SessionID != started.Id || startedPayload.PlanID != "practice-plan-p0-023" || startedPayload.TargetJobID != "target-job-p0-023-a" {
		t.Fatalf("practice.session.started payload mismatch: %+v", startedPayload)
	}
	assertNoEvidenceLeak(t, h.store.outboxPayloads(), started.CurrentTurn.QuestionText, "question_text", "answer_text", "hint_text", "prompt body", "response body", "provider secret")
}

func TestE2EP0024PracticeSessionAIFailureRetry(t *testing.T) {
	ai := &scenarioPracticeAIClient{
		failures: []error{sharederrors.Wrap(sharederrors.CodeAiProviderTimeout, "prompt body and response body timed out", true)},
	}
	h := newPracticeHTTPScenarioHarness(t, practiceHTTPScenarioOptions{ai: ai})
	h.seedReadyScenarioPlan("practice-plan-p0-024", "target-job-p0-024-a", "resume-asset-p0-024-a", practiceHTTPScenarioUserAID)

	body := api.StartPracticeSessionRequest{PlanId: "practice-plan-p0-024", HintsEnabled: practiceBoolPtr(true)}
	failedRaw := h.doJSON(t, practiceHTTPScenarioUserAID, http.MethodPost, "/api/v1/practice/sessions", "e2e-p0-024-start-session", body, http.StatusBadGateway)
	var failed api.ApiErrorResponse
	decodeJSON(t, failedRaw, &failed)
	if failed.Error.Code != sharederrors.CodeAiProviderTimeout || !failed.Error.Retryable {
		t.Fatalf("first failure should map to retryable AI_PROVIDER_TIMEOUT: %+v", failed.Error)
	}
	assertNoEvidenceLeak(t, [][]byte{failedRaw}, "prompt body", "response body")
	if h.store.idempotencyStatus(practiceHTTPScenarioUserAID, "practice", "startPracticeSession", "e2e-p0-024-start-session") != idempotency.StatusFailedRetry {
		t.Fatalf("first failure should leave idempotency record failed_retryable")
	}
	if h.store.failedSessionCount("practice-plan-p0-024") != 1 || h.store.outboxCount() != 0 {
		t.Fatalf("first failure should mark one failed session and emit no outbox, failed=%d outbox=%d", h.store.failedSessionCount("practice-plan-p0-024"), h.store.outboxCount())
	}

	retryRaw := h.doJSON(t, practiceHTTPScenarioUserAID, http.MethodPost, "/api/v1/practice/sessions", "e2e-p0-024-start-session", body, http.StatusCreated)
	var retry api.PracticeSession
	decodeJSON(t, retryRaw, &retry)
	if retry.Status != sharedtypes.SessionStatusRunning || retry.CurrentTurn == nil || retry.CurrentTurn.TurnIndex != 1 {
		t.Fatalf("retry did not start a running session with first turn: %+v", retry)
	}
	if h.store.idempotencyStatus(practiceHTTPScenarioUserAID, "practice", "startPracticeSession", "e2e-p0-024-start-session") != idempotency.StatusSucceeded {
		t.Fatalf("retry success should mark idempotency succeeded")
	}
	if h.store.outboxCount() != 1 || ai.calls != 2 {
		t.Fatalf("retry should call AI twice total and emit one outbox, calls=%d outbox=%d", ai.calls, h.store.outboxCount())
	}
}

func TestE2EP0025PracticeIdempotencyAndIsolationMatrix(t *testing.T) {
	h := newPracticeHTTPScenarioHarness(t)
	planA1 := h.seedReadyScenarioPlan("practice-plan-p0-025-a1", "target-job-p0-025-a1", "resume-asset-p0-025-a1", practiceHTTPScenarioUserAID)
	planA2 := h.seedReadyScenarioPlan("practice-plan-p0-025-a2", "target-job-p0-025-a2", "resume-asset-p0-025-a2", practiceHTTPScenarioUserAID)
	planB := h.seedReadyScenarioPlan("practice-plan-p0-025-b1", "target-job-p0-025-b1", "resume-asset-p0-025-b1", practiceHTTPScenarioUserBID)

	bodyA1 := api.StartPracticeSessionRequest{PlanId: planA1.ID, HintsEnabled: practiceBoolPtr(true)}
	firstRaw := h.doJSON(t, practiceHTTPScenarioUserAID, http.MethodPost, "/api/v1/practice/sessions", "e2e-p0-025-shared-key", bodyA1, http.StatusCreated)
	var first api.PracticeSession
	decodeJSON(t, firstRaw, &first)
	h.store.forceSessionReplayDrift(first.Id)
	replayRaw := h.doJSON(t, practiceHTTPScenarioUserAID, http.MethodPost, "/api/v1/practice/sessions", "e2e-p0-025-shared-key", bodyA1, http.StatusCreated)
	var replay api.PracticeSession
	decodeJSON(t, replayRaw, &replay)
	if replay.Id != first.Id ||
		replay.Status != first.Status ||
		replay.TurnCount != first.TurnCount ||
		replay.CurrentTurn == nil ||
		first.CurrentTurn == nil ||
		replay.CurrentTurn.QuestionText != first.CurrentTurn.QuestionText ||
		h.store.outboxCount() != 1 {
		t.Fatalf("same user/key/fingerprint should replay without duplicate outbox: first=%+v replay=%+v outbox=%d", first, replay, h.store.outboxCount())
	}

	mismatchRaw := h.doJSON(t, practiceHTTPScenarioUserAID, http.MethodPost, "/api/v1/practice/sessions", "e2e-p0-025-shared-key", api.StartPracticeSessionRequest{
		PlanId: planA1.ID, HintsEnabled: practiceBoolPtr(false),
	}, http.StatusConflict)
	var mismatch api.ApiErrorResponse
	decodeJSON(t, mismatchRaw, &mismatch)
	if mismatch.Error.Code != sharederrors.CodePracticeSessionConflict || strings.Contains(string(mismatchRaw), first.Id) {
		t.Fatalf("fingerprint mismatch should return conflict without first resource leak: %s", string(mismatchRaw))
	}

	bodyB := api.StartPracticeSessionRequest{PlanId: planB.ID, HintsEnabled: practiceBoolPtr(true)}
	userBRaw := h.doJSON(t, practiceHTTPScenarioUserBID, http.MethodPost, "/api/v1/practice/sessions", "e2e-p0-025-shared-key", bodyB, http.StatusCreated)
	var userB api.PracticeSession
	decodeJSON(t, userBRaw, &userB)
	if userB.Id == first.Id || h.store.outboxCount() != 2 {
		t.Fatalf("cross-user same key should be isolated: userA=%s userB=%s outbox=%d", first.Id, userB.Id, h.store.outboxCount())
	}

	bodyA2 := api.StartPracticeSessionRequest{PlanId: planA2.ID, HintsEnabled: practiceBoolPtr(true)}
	activeRaw := h.doJSON(t, practiceHTTPScenarioUserAID, http.MethodPost, "/api/v1/practice/sessions", "e2e-p0-025-active-1", bodyA2, http.StatusCreated)
	var active api.PracticeSession
	decodeJSON(t, activeRaw, &active)
	conflictRaw := h.doJSON(t, practiceHTTPScenarioUserAID, http.MethodPost, "/api/v1/practice/sessions", "e2e-p0-025-active-2", bodyA2, http.StatusConflict)
	var conflict api.ApiErrorResponse
	decodeJSON(t, conflictRaw, &conflict)
	if conflict.Error.Code != sharederrors.CodePracticeSessionConflict || strings.Contains(string(conflictRaw), active.Id) {
		t.Fatalf("same plan multi-key conflict should not leak active session: %s", string(conflictRaw))
	}

	planCrossRaw := h.doJSON(t, practiceHTTPScenarioUserBID, http.MethodGet, "/api/v1/practice/plans/"+planA1.ID, "", nil, http.StatusNotFound)
	var planCross api.ApiErrorResponse
	decodeJSON(t, planCrossRaw, &planCross)
	sessionCrossRaw := h.doJSON(t, practiceHTTPScenarioUserBID, http.MethodGet, "/api/v1/practice/sessions/"+first.Id, "", nil, http.StatusNotFound)
	var sessionCross api.ApiErrorResponse
	decodeJSON(t, sessionCrossRaw, &sessionCross)
	if planCross.Error.Code != sharederrors.CodePracticePlanNotFound || sessionCross.Error.Code != sharederrors.CodePracticeSessionNotFound {
		t.Fatalf("cross-user GET should hide plan/session existence: plan=%+v session=%+v", planCross.Error, sessionCross.Error)
	}
}

func TestE2EP0026PracticeObservabilityAndPrivacyRedlines(t *testing.T) {
	ai := &scenarioPracticeAIClient{
		responseText:   "response body provider secret sk-test answer_text hint_text",
		responseIntent: "prompt body response body",
	}
	h := newPracticeHTTPScenarioHarness(t, practiceHTTPScenarioOptions{ai: ai, observedAI: true})
	plan := h.seedReadyScenarioPlan(
		"01918fa0-0000-7000-8000-000000004026",
		"01918fa0-0000-7000-8000-000000002026",
		"resume-asset-p0-026-a",
		practiceHTTPScenarioUserAID,
	)

	raw := h.doJSON(t, practiceHTTPScenarioUserAID, http.MethodPost, "/api/v1/practice/sessions", "e2e-p0-026-start-session", api.StartPracticeSessionRequest{
		PlanId:       plan.ID,
		HintsEnabled: practiceBoolPtr(true),
	}, http.StatusCreated)
	var started api.PracticeSession
	decodeJSON(t, raw, &started)
	if started.Status != sharedtypes.SessionStatusRunning || started.CurrentTurn == nil {
		t.Fatalf("unexpected observed startPracticeSession response: %+v", started)
	}

	rows := h.aiTaskRuns.Rows()
	if len(rows) != 1 {
		t.Fatalf("expected one observed ai_task_runs row, got %+v", rows)
	}
	row := rows[0]
	if row.FeatureKey != "practice.session.first_question" ||
		row.ModelProfileName != "practice.first_question.default" ||
		row.ModelFamily != "stub" ||
		len(row.FallbackChain) != 1 ||
		row.ValidationStatus != aiclient.ValidationStatusOK ||
		row.Route != "practice.session.first_question" ||
		row.FeatureFlag != "none" ||
		row.DataSourceVersion != "registry.v1" ||
		row.UserID != practiceHTTPScenarioUserAID ||
		row.Capability != aiclient.AITaskRunTaskQuestionGenerate ||
		row.ResourceType != aiclient.AITaskRunResourceTargetJob ||
		row.ResourceID != plan.TargetJobID {
		t.Fatalf("observed ai_task_runs row incomplete: %+v", row)
	}
	if row.Metadata.PromptHash == "" || row.Metadata.ResponseHash == "" {
		t.Fatalf("observed ai_task_runs row should keep hash summaries only: %+v", row.Metadata)
	}
	if containsString(observability.StandardLabelKeys, "feature_key") ||
		containsString(observability.StandardLabelKeys, "prompt_version") ||
		containsString(observability.StandardLabelKeys, "rubric_version") {
		t.Fatalf("standard metric labels contain high-cardinality provenance keys: %v", observability.StandardLabelKeys)
	}
	labels := h.metrics.CounterLabelValues(observability.MetricRunsTotal)
	if len(labels) != 1 || len(labels[0]) != len(observability.StandardLabelKeys) {
		t.Fatalf("metric label tuple drifted: labels=%v keys=%v", labels, observability.StandardLabelKeys)
	}

	serialized, err := json.Marshal(map[string]any{
		"ai_logs":       h.aiLogs.Entries(),
		"ai_task_runs":  rows,
		"ai_audit":      h.aiAudit.Rows(),
		"metric_labels": labels,
	})
	if err != nil {
		t.Fatalf("marshal observability snapshot: %v", err)
	}
	payloads := append(h.store.auditPayloads(), h.store.outboxPayloads()...)
	payloads = append(payloads, serialized)
	assertNoEvidenceLeak(t, payloads, "question_text", "answer_text", "hint_text", "prompt body", "response body", "provider secret", "sk-test")
}

type practiceHTTPScenarioHarness struct {
	handler    http.Handler
	store      *scenarioPracticeStore
	cookies    map[string]*http.Cookie
	metrics    *observability.InMemoryRegistry
	aiLogs     *observability.MemoryLogger
	aiTaskRuns *scenarioAITaskRunWriter
	aiAudit    *scenarioAIAuditWriter
}

type practiceHTTPScenarioOptions struct {
	ai         *scenarioPracticeAIClient
	observedAI bool
}

func newPracticeHTTPScenarioHarness(t *testing.T, options ...practiceHTTPScenarioOptions) *practiceHTTPScenarioHarness {
	t.Helper()
	var opts practiceHTTPScenarioOptions
	if len(options) > 0 {
		opts = options[0]
	}
	dir := t.TempDir()
	writeAPIFile(t, filepath.Join(dir, "config.yaml"), `
runtime:
  appVersion: "1.2.3"
  defaultUiLanguage: zh-CN
auth:
  challengeTokenPepper: "scenario-challenge-pepper"
`)
	loader, err := config.Load(config.Options{AppEnv: "test", ConfigDir: dir})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	store := newScenarioPracticeStore()
	authStore := newPracticeScenarioAuthStore("scenario-session-secret")
	cookies := map[string]*http.Cookie{
		practiceHTTPScenarioUserAID: authStore.addSession(practiceHTTPScenarioUserAID, "candidate-a@example.com", "raw-session-token-a"),
		practiceHTTPScenarioUserBID: authStore.addSession(practiceHTTPScenarioUserBID, "candidate-b@example.com", "raw-session-token-b"),
	}
	authService := auth.NewPasswordlessService(auth.PasswordlessServiceOptions{
		Store:               authStore,
		SessionCookieSecret: "scenario-session-secret",
		Now:                 fixedScenarioNow,
	})
	ai := opts.ai
	if ai == nil {
		ai = &scenarioPracticeAIClient{}
	}
	var aiClient aiclient.AIClient = ai.withStore(store)
	metrics := observability.NewInMemoryRegistry()
	aiLogs := observability.NewMemoryLogger()
	aiTaskRuns := &scenarioAITaskRunWriter{}
	aiAudit := &scenarioAIAuditWriter{}
	if opts.observedAI {
		observed, err := observability.New(aiClient,
			observability.WithRegisterer(metrics),
			observability.WithLogger(aiLogs),
			observability.WithAITaskRunWriter(aiTaskRuns),
			observability.WithAuditEventWriter(aiAudit),
			observability.WithProfileResolver(scenarioPracticeProfileResolver{}),
			observability.WithNow(fixedScenarioNow),
		)
		if err != nil {
			t.Fatalf("observability.New: %v", err)
		}
		aiClient = observed
	}
	service := domainpractice.NewService(domainpractice.ServiceOptions{
		Store:    store,
		Registry: &scenarioPracticeRegistry{},
		AI:       aiClient,
		Now:      fixedScenarioNow,
		NewID:    store.nextID,
	})
	practiceHandler := apipractice.NewHandler(apipractice.HandlerOptions{
		Service: service,
		Session: currentUserFromContext,
	})
	routeIdempotency := idempotency.New(idempotency.MiddlewareOptions{
		Store:     store,
		Now:       fixedScenarioNow,
		NewID:     store.nextID,
		KeyPepper: "scenario-challenge-pepper",
	})

	return &practiceHTTPScenarioHarness{
		handler: buildAPIHandlerWithHandlers(loader, apiRuntimeFlags{}, authService, targetjob.NewHandler(targetjob.HandlerOptions{}), practiceRoutes{
			Handler:     practiceHandler,
			Idempotency: routeIdempotency,
		}),
		store:      store,
		cookies:    cookies,
		metrics:    metrics,
		aiLogs:     aiLogs,
		aiTaskRuns: aiTaskRuns,
		aiAudit:    aiAudit,
	}
}

func (h *practiceHTTPScenarioHarness) seedReadyScenarioPlan(planID, targetJobID, resumeAssetID, userID string) domainpractice.PlanRecord {
	h.store.prerequisiteTargetOwner[targetJobID] = userID
	h.store.prerequisiteResumeOwner[resumeAssetID] = userID
	return h.store.seedReadyPlan(domainpractice.CreatePlanStoreInput{
		PlanID:               planID,
		AuditEventID:         "audit-" + planID,
		UserID:               userID,
		TargetJobID:          targetJobID,
		ResumeAssetID:        resumeAssetID,
		Goal:                 sharedtypes.PracticeGoalBaseline,
		Mode:                 sharedtypes.PracticeModeAssisted,
		InterviewerPersona:   sharedtypes.InterviewerRoleHiringManager,
		Difficulty:           "standard",
		Language:             "zh-CN",
		TimeBudgetMinutes:    30,
		QuestionBudget:       6,
		FocusCompetencyCodes: []string{"system-design"},
		Now:                  fixedScenarioNow(),
	})
}

func (h *practiceHTTPScenarioHarness) doJSON(t *testing.T, userID, method, path string, idempotencyKey string, body any, wantStatus int) []byte {
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
	cookie, ok := h.cookies[userID]
	if !ok {
		t.Fatalf("missing scenario cookie for user %q", userID)
	}
	req.AddCookie(cookie)
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

type scenarioPracticeRegistry struct{}

func (r *scenarioPracticeRegistry) ResolveActive(ctx context.Context, featureKey, language string) (registry.PromptResolution, error) {
	return registry.PromptResolution{
		FeatureKey:          featureKey,
		PromptVersion:       "prompt.v1",
		RubricVersion:       "rubric.v1",
		ModelProfileName:    "practice.first_question.default",
		DataSourceVersion:   "registry.v1",
		FeatureFlag:         "none",
		UserMessageTemplate: "ask the first interview question",
	}, nil
}

type scenarioPracticeProfileResolver struct{}

func (r scenarioPracticeProfileResolver) Resolve(name string) (*aiclient.ModelProfile, error) {
	if name != "practice.first_question.default" {
		return nil, fmt.Errorf("missing scenario profile %q", name)
	}
	return &aiclient.ModelProfile{
		Name:       "practice.first_question.default",
		Capability: aiclient.CapabilityChat,
		Status:     aiclient.ProfileStatusActive,
		Default: aiclient.ProviderConfig{
			ProviderRef: "stub",
			Model:       "stub-chat-1",
		},
		Route:     "practice.session.first_question",
		TimeoutMs: 5000,
		Version:   "1.0.0",
	}, nil
}

type scenarioPracticeAIClient struct {
	store          *scenarioPracticeStore
	failures       []error
	responseText   string
	responseIntent string
	calls          int
}

func (c *scenarioPracticeAIClient) withStore(store *scenarioPracticeStore) *scenarioPracticeAIClient {
	if c == nil {
		return &scenarioPracticeAIClient{store: store}
	}
	c.store = store
	return c
}

func (c *scenarioPracticeAIClient) Complete(ctx context.Context, profileName string, payload aiclient.CompletePayload) (aiclient.CompleteResponse, aiclient.AICallMeta, error) {
	c.calls++
	if c.store != nil {
		c.store.recordAIObservation()
	}
	if len(c.failures) > 0 {
		err := c.failures[0]
		c.failures = c.failures[1:]
		return aiclient.CompleteResponse{}, aiclient.AICallMeta{}, err
	}
	if profileName != "practice.first_question.default" {
		return aiclient.CompleteResponse{}, aiclient.AICallMeta{}, fmt.Errorf("unexpected profile %q", profileName)
	}
	if payload.Metadata.FeatureKey != "practice.session.first_question" ||
		payload.Metadata.PromptVersion == "" ||
		payload.Metadata.RubricVersion == "" ||
		payload.Metadata.Language == "" ||
		payload.Metadata.FeatureFlag == "" ||
		payload.Metadata.DataSourceVersion == "" {
		return aiclient.CompleteResponse{}, aiclient.AICallMeta{}, fmt.Errorf("incomplete AI metadata: %+v", payload.Metadata)
	}
	if payload.Metadata.TaskRun.Capability != aiclient.AITaskRunTaskQuestionGenerate ||
		payload.Metadata.TaskRun.ResourceType != aiclient.AITaskRunResourceTargetJob ||
		payload.Metadata.TaskRun.ResourceID == "" {
		return aiclient.CompleteResponse{}, aiclient.AICallMeta{}, fmt.Errorf("incomplete AI task run context: %+v", payload.Metadata.TaskRun)
	}
	questionText := c.responseText
	if questionText == "" {
		questionText = "请用 STAR 描述你主导设计系统迁移的项目，重点说明跨团队协调过程。"
	}
	questionIntent := c.responseIntent
	if questionIntent == "" {
		questionIntent = "behavioral.leadership.design_system"
	}
	content, err := json.Marshal(map[string]string{"questionText": questionText, "questionIntent": questionIntent})
	if err != nil {
		return aiclient.CompleteResponse{}, aiclient.AICallMeta{}, err
	}
	return aiclient.CompleteResponse{Content: string(content)}, aiclient.AICallMeta{
		Provider:         "stub",
		ModelFamily:      "stub",
		ModelID:          "stub-chat-1",
		FallbackChain:    []string{"stub/stub-chat-1"},
		ValidationStatus: aiclient.ValidationStatusOK,
	}, nil
}

func (c *scenarioPracticeAIClient) Transcribe(ctx context.Context, profileName string, input aiclient.TranscriptionInput) (aiclient.TranscriptionResponse, aiclient.AICallMeta, error) {
	return aiclient.TranscriptionResponse{}, aiclient.AICallMeta{}, nil
}

func (c *scenarioPracticeAIClient) Stream(ctx context.Context, profileName string, payload aiclient.CompletePayload) (<-chan aiclient.AIStreamEvent, error) {
	return nil, nil
}

func (c *scenarioPracticeAIClient) Synthesize(ctx context.Context, profileName string, input aiclient.SynthesisInput) (aiclient.SynthesisResponse, aiclient.AICallMeta, error) {
	return aiclient.SynthesisResponse{}, aiclient.AICallMeta{}, nil
}

type scenarioAITaskRunWriter struct {
	rows []aiclient.AITaskRunRow
}

func (w *scenarioAITaskRunWriter) WriteAITaskRun(_ context.Context, row aiclient.AITaskRunRow) error {
	w.rows = append(w.rows, row)
	return nil
}

func (w *scenarioAITaskRunWriter) Rows() []aiclient.AITaskRunRow {
	return append([]aiclient.AITaskRunRow{}, w.rows...)
}

type scenarioAIAuditWriter struct {
	rows []aiclient.AuditEventRow
}

func (w *scenarioAIAuditWriter) WriteAuditEvent(_ context.Context, row aiclient.AuditEventRow) error {
	w.rows = append(w.rows, row)
	return nil
}

func (w *scenarioAIAuditWriter) Rows() []aiclient.AuditEventRow {
	return append([]aiclient.AuditEventRow{}, w.rows...)
}

type scenarioPracticeStore struct {
	seq                       int
	plans                     map[string]scenarioPracticePlan
	sessions                  map[string]scenarioPracticeSession
	turns                     map[string][]domainpractice.TurnRecord
	sessionEvents             map[string][]scenarioPracticeSessionEvent
	idempotencyRecords        map[string]scenarioPracticeIdempotencyRecord
	outbox                    []scenarioOutboxEvent
	audits                    [][]byte
	prerequisiteTargetOwner   map[string]string
	prerequisiteResumeOwner   map[string]string
	inTransaction             bool
	aiCalledInsideTransaction bool
}

type scenarioPracticePlan struct {
	domainpractice.PlanRecord
	UserID        string
	ResumeAssetID string
}

type scenarioPracticeSession struct {
	domainpractice.SessionRecord
	UserID      string
	FailureCode string
}

type scenarioPracticeSessionEvent struct {
	ID        string
	SeqNo     int
	EventType string
	Payload   []byte
}

type scenarioOutboxEvent struct {
	EventName sharedevents.EventName
	Payload   []byte
}

type scenarioPracticeIdempotencyRecord struct {
	RecordID    string
	UserID      string
	Domain      string
	Operation   string
	KeyHash     string
	Fingerprint string
	Status      idempotency.Status
	ExpiresAt   time.Time
	Response    []byte
	HTTPStatus  int
	ResourceID  string
	ErrorCode   string
}

func newScenarioPracticeStore() *scenarioPracticeStore {
	s := &scenarioPracticeStore{
		plans:                   map[string]scenarioPracticePlan{},
		sessions:                map[string]scenarioPracticeSession{},
		turns:                   map[string][]domainpractice.TurnRecord{},
		sessionEvents:           map[string][]scenarioPracticeSessionEvent{},
		idempotencyRecords:      map[string]scenarioPracticeIdempotencyRecord{},
		prerequisiteTargetOwner: map[string]string{},
		prerequisiteResumeOwner: map[string]string{},
	}
	s.prerequisiteTargetOwner["target-job-p0-022-a"] = practiceHTTPScenarioUserAID
	s.prerequisiteResumeOwner["resume-asset-p0-022-a"] = practiceHTTPScenarioUserAID
	s.prerequisiteTargetOwner["target-job-p0-023-a"] = practiceHTTPScenarioUserAID
	s.prerequisiteResumeOwner["resume-asset-p0-023-a"] = practiceHTTPScenarioUserAID
	return s
}

func (s *scenarioPracticeStore) nextID() string {
	s.seq++
	return "practice-scenario-id-" + strconv.Itoa(s.seq)
}

func (s *scenarioPracticeStore) CreatePlan(_ context.Context, in domainpractice.CreatePlanStoreInput) (domainpractice.PlanRecord, error) {
	if s.prerequisiteTargetOwner[in.TargetJobID] != in.UserID || s.prerequisiteResumeOwner[in.ResumeAssetID] != in.UserID {
		return domainpractice.PlanRecord{}, domainpractice.ErrPlanPrerequisiteNotFound
	}
	plan := domainpractice.PlanRecord{
		ID:                 in.PlanID,
		TargetJobID:        in.TargetJobID,
		Goal:               in.Goal,
		Mode:               in.Mode,
		InterviewerPersona: in.InterviewerPersona,
		Difficulty:         in.Difficulty,
		Language:           in.Language,
		TimeBudgetMinutes:  in.TimeBudgetMinutes,
		QuestionBudget:     in.QuestionBudget,
		Status:             "ready",
		CreatedAt:          in.Now,
	}
	s.plans[in.PlanID] = scenarioPracticePlan{PlanRecord: plan, UserID: in.UserID, ResumeAssetID: in.ResumeAssetID}
	audit, err := json.Marshal(map[string]any{
		"plan_id":       in.PlanID,
		"goal":          string(in.Goal),
		"mode":          string(in.Mode),
		"language":      in.Language,
		"target_job_id": in.TargetJobID,
	})
	if err != nil {
		return domainpractice.PlanRecord{}, err
	}
	s.audits = append(s.audits, audit)
	return plan, nil
}

func (s *scenarioPracticeStore) seedReadyPlan(in domainpractice.CreatePlanStoreInput) domainpractice.PlanRecord {
	plan, err := s.CreatePlan(context.Background(), in)
	if err != nil {
		panic(err)
	}
	return plan
}

func (s *scenarioPracticeStore) GetPlan(_ context.Context, userID, planID string) (domainpractice.PlanRecord, error) {
	plan, ok := s.plans[planID]
	if !ok || plan.UserID != userID {
		return domainpractice.PlanRecord{}, domainpractice.ErrPlanNotFound
	}
	return plan.PlanRecord, nil
}

func (s *scenarioPracticeStore) GetSession(_ context.Context, userID, sessionID string) (domainpractice.SessionRecord, error) {
	session, ok := s.sessions[sessionID]
	if !ok || session.UserID != userID {
		return domainpractice.SessionRecord{}, domainpractice.ErrSessionNotFound
	}
	return session.SessionRecord, nil
}

func (s *scenarioPracticeStore) ReserveSessionStart(_ context.Context, in domainpractice.StartSessionReservationInput) (domainpractice.SessionReservation, error) {
	plan, ok := s.plans[in.PlanID]
	if !ok || plan.UserID != in.UserID || plan.Status != "ready" {
		return domainpractice.SessionReservation{}, domainpractice.ErrPlanNotFound
	}
	s.inTransaction = true
	defer func() { s.inTransaction = false }()

	key := s.idempotencyRecordKey(in.UserID, "practice", "startPracticeSession", in.IdempotencyKeyHash)
	recordID := in.IdempotencyRecordID
	if existing, ok := s.idempotencyRecords[key]; ok {
		if existing.Fingerprint != in.RequestFingerprint {
			return domainpractice.SessionReservation{}, domainpractice.ErrSessionConflict
		}
		switch existing.Status {
		case idempotency.StatusPending:
			return domainpractice.SessionReservation{}, domainpractice.ErrSessionConflict
		case idempotency.StatusSucceeded:
			if len(existing.Response) == 0 {
				return domainpractice.SessionReservation{}, domainpractice.ErrSessionConflict
			}
			replay, err := scenarioSessionRecordFromResponseBody(existing.Response)
			if err != nil {
				return domainpractice.SessionReservation{}, err
			}
			if replay.ID != existing.ResourceID {
				return domainpractice.SessionReservation{}, domainpractice.ErrSessionConflict
			}
			return domainpractice.SessionReservation{ReplaySession: &replay}, nil
		case idempotency.StatusFailedRetry:
			recordID = existing.RecordID
			existing.Status = idempotency.StatusPending
			existing.ErrorCode = ""
			existing.ResourceID = ""
			existing.Response = nil
			existing.HTTPStatus = 0
			existing.ExpiresAt = in.ExpiresAt
			s.idempotencyRecords[key] = existing
		default:
			return domainpractice.SessionReservation{}, domainpractice.ErrSessionConflict
		}
	} else {
		s.idempotencyRecords[key] = scenarioPracticeIdempotencyRecord{
			RecordID:    recordID,
			UserID:      in.UserID,
			Domain:      "practice",
			Operation:   "startPracticeSession",
			KeyHash:     in.IdempotencyKeyHash,
			Fingerprint: in.RequestFingerprint,
			Status:      idempotency.StatusPending,
			ExpiresAt:   in.ExpiresAt,
		}
	}
	for _, session := range s.sessions {
		if session.UserID == in.UserID && session.PlanID == in.PlanID && (session.Status == sharedtypes.SessionStatusQueued || session.Status == sharedtypes.SessionStatusRunning) {
			return domainpractice.SessionReservation{}, domainpractice.ErrSessionConflict
		}
	}
	session := domainpractice.SessionRecord{
		ID:           in.SessionID,
		PlanID:       in.PlanID,
		TargetJobID:  plan.TargetJobID,
		Status:       sharedtypes.SessionStatusQueued,
		Language:     plan.Language,
		HintsEnabled: in.HintsEnabled,
		TurnCount:    0,
		CreatedAt:    in.Now,
		UpdatedAt:    in.Now,
	}
	s.sessions[in.SessionID] = scenarioPracticeSession{SessionRecord: session, UserID: in.UserID}
	return domainpractice.SessionReservation{
		IdempotencyRecordID: recordID,
		SessionID:           in.SessionID,
		UserID:              in.UserID,
		PlanID:              in.PlanID,
		TargetJobID:         plan.TargetJobID,
		Goal:                plan.Goal,
		Mode:                plan.Mode,
		InterviewerPersona:  plan.InterviewerPersona,
		Language:            plan.Language,
		HintsEnabled:        in.HintsEnabled,
		CreatedAt:           in.Now,
		UpdatedAt:           in.Now,
	}, nil
}

func (s *scenarioPracticeStore) CommitSessionStart(_ context.Context, in domainpractice.CommitSessionStartInput) (domainpractice.SessionRecord, error) {
	session, ok := s.sessions[in.SessionID]
	if !ok {
		return domainpractice.SessionRecord{}, domainpractice.ErrSessionNotFound
	}
	s.inTransaction = true
	defer func() { s.inTransaction = false }()

	turn := domainpractice.TurnRecord{
		ID:             in.TurnID,
		TurnIndex:      1,
		QuestionText:   in.QuestionText,
		QuestionIntent: in.QuestionIntent,
		Status:         "asked",
		AskedAt:        in.StartedAt,
	}
	s.turns[in.SessionID] = append(s.turns[in.SessionID], turn)
	eventPayload, err := json.Marshal(map[string]any{"sessionId": in.SessionID, "turnId": in.TurnID, "turnIndex": 1})
	if err != nil {
		return domainpractice.SessionRecord{}, err
	}
	s.sessionEvents[in.SessionID] = append(s.sessionEvents[in.SessionID], scenarioPracticeSessionEvent{
		ID:        in.SessionEventID,
		SeqNo:     1,
		EventType: "session_started",
		Payload:   eventPayload,
	})
	startedPayload, err := storepractice.BuildPracticeSessionStartedPayload(storepractice.PracticeSessionStartedInput{
		Goal:        in.Goal,
		Language:    in.Language,
		Mode:        in.Mode,
		PlanID:      in.PlanID,
		SessionID:   in.SessionID,
		TargetJobID: in.TargetJobID,
	})
	if err != nil {
		return domainpractice.SessionRecord{}, err
	}
	outboxPayload, err := json.Marshal(startedPayload)
	if err != nil {
		return domainpractice.SessionRecord{}, err
	}
	s.outbox = append(s.outbox, scenarioOutboxEvent{EventName: sharedevents.EventNamePracticeSessionStarted, Payload: outboxPayload})

	audit, err := json.Marshal(map[string]any{
		"plan_id":       in.PlanID,
		"session_id":    in.SessionID,
		"goal":          string(in.Goal),
		"mode":          string(in.Mode),
		"language":      in.Language,
		"target_job_id": in.TargetJobID,
	})
	if err != nil {
		return domainpractice.SessionRecord{}, err
	}
	s.audits = append(s.audits, audit)

	session.Status = sharedtypes.SessionStatusRunning
	session.TurnCount = 1
	session.CurrentTurn = &turn
	session.UpdatedAt = in.StartedAt
	s.sessions[in.SessionID] = session

	responseBody, err := json.Marshal(api.PracticeSession{
		Id:           session.ID,
		PlanId:       session.PlanID,
		TargetJobId:  session.TargetJobID,
		Status:       session.Status,
		Language:     session.Language,
		HintsEnabled: session.HintsEnabled,
		TurnCount:    session.TurnCount,
		CurrentTurn: &api.PracticeTurn{
			Id:             turn.ID,
			TurnIndex:      turn.TurnIndex,
			QuestionText:   turn.QuestionText,
			QuestionIntent: &turn.QuestionIntent,
			Status:         turn.Status,
			AskedAt:        strPtr(turn.AskedAt.UTC().Format(time.RFC3339)),
		},
		CreatedAt: session.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt: session.UpdatedAt.UTC().Format(time.RFC3339),
	})
	if err != nil {
		return domainpractice.SessionRecord{}, err
	}
	for key, record := range s.idempotencyRecords {
		if record.RecordID == in.IdempotencyRecordID {
			record.Status = idempotency.StatusSucceeded
			record.Response = responseBody
			record.HTTPStatus = http.StatusCreated
			record.ResourceID = in.SessionID
			s.idempotencyRecords[key] = record
			break
		}
	}
	return session.SessionRecord, nil
}

func (s *scenarioPracticeStore) FailSessionStart(_ context.Context, in domainpractice.FailSessionStartInput) error {
	session, ok := s.sessions[in.SessionID]
	if !ok || session.UserID != in.UserID {
		return domainpractice.ErrSessionNotFound
	}
	s.inTransaction = true
	defer func() { s.inTransaction = false }()

	session.Status = sharedtypes.SessionStatusFailed
	session.UpdatedAt = in.FailedAt
	s.sessions[in.SessionID] = session

	status := idempotency.StatusFailedTerminal
	if in.Retryable {
		status = idempotency.StatusFailedRetry
	}
	for key, record := range s.idempotencyRecords {
		if record.RecordID == in.IdempotencyRecordID {
			record.Status = status
			record.HTTPStatus = http.StatusBadGateway
			record.ResourceID = in.SessionID
			record.ErrorCode = in.ErrorCode
			s.idempotencyRecords[key] = record
			return nil
		}
	}
	return idempotency.ErrReservationNotFound
}

func (s *scenarioPracticeStore) Reserve(_ context.Context, in idempotency.ReservationInput) (idempotency.Reservation, error) {
	key := s.idempotencyRecordKey(in.UserID, in.Domain, in.Operation, in.IdempotencyKeyHash)
	rec, ok := s.idempotencyRecords[key]
	if !ok || !in.Now.Before(rec.ExpiresAt) {
		s.idempotencyRecords[key] = scenarioPracticeIdempotencyRecord{
			RecordID:    in.RecordID,
			UserID:      in.UserID,
			Domain:      in.Domain,
			Operation:   in.Operation,
			KeyHash:     in.IdempotencyKeyHash,
			Fingerprint: in.RequestFingerprint,
			Status:      idempotency.StatusPending,
			ExpiresAt:   in.ExpiresAt,
		}
		return idempotency.Reservation{State: idempotency.StateExecute, RecordID: in.RecordID}, nil
	}
	if rec.Fingerprint != in.RequestFingerprint {
		return idempotency.Reservation{}, idempotency.ErrFingerprintMismatch
	}
	switch rec.Status {
	case idempotency.StatusPending:
		return idempotency.Reservation{}, idempotency.ErrPending
	case idempotency.StatusSucceeded:
		return idempotency.Reservation{
			State:          idempotency.StateReplay,
			RecordID:       rec.RecordID,
			ResponseStatus: rec.HTTPStatus,
			ResponseBody:   append([]byte{}, rec.Response...),
		}, nil
	case idempotency.StatusFailedRetry:
		rec.Status = idempotency.StatusPending
		rec.ExpiresAt = in.ExpiresAt
		rec.Fingerprint = in.RequestFingerprint
		s.idempotencyRecords[key] = rec
		return idempotency.Reservation{State: idempotency.StateExecute, RecordID: rec.RecordID}, nil
	default:
		return idempotency.Reservation{}, idempotency.ErrUnexpectedStatus
	}
}

func (s *scenarioPracticeStore) MarkSucceeded(_ context.Context, in idempotency.CompletionInput) error {
	for key, rec := range s.idempotencyRecords {
		if rec.RecordID == in.RecordID && rec.UserID == in.UserID && rec.Domain == in.Domain && rec.Operation == in.Operation {
			rec.Status = idempotency.StatusSucceeded
			rec.Response = append([]byte{}, in.ResponseBody...)
			rec.HTTPStatus = in.ResponseStatus
			s.idempotencyRecords[key] = rec
			return nil
		}
	}
	return idempotency.ErrReservationNotFound
}

func (s *scenarioPracticeStore) recordAIObservation() {
	if s.inTransaction {
		s.aiCalledInsideTransaction = true
	}
}

func (s *scenarioPracticeStore) planCount() int {
	return len(s.plans)
}

func (s *scenarioPracticeStore) auditCount() int {
	return len(s.audits)
}

func (s *scenarioPracticeStore) auditPayloads() [][]byte {
	out := make([][]byte, 0, len(s.audits))
	for _, payload := range s.audits {
		out = append(out, append([]byte{}, payload...))
	}
	return out
}

func (s *scenarioPracticeStore) turnCount(sessionID string) int {
	return len(s.turns[sessionID])
}

func (s *scenarioPracticeStore) sessionEventCount(sessionID string) int {
	return len(s.sessionEvents[sessionID])
}

func (s *scenarioPracticeStore) outboxCount() int {
	return len(s.outbox)
}

func (s *scenarioPracticeStore) outboxPayloads() [][]byte {
	out := make([][]byte, 0, len(s.outbox))
	for _, event := range s.outbox {
		out = append(out, append([]byte{}, event.Payload...))
	}
	return out
}

func (s *scenarioPracticeStore) failedSessionCount(planID string) int {
	count := 0
	for _, session := range s.sessions {
		if session.PlanID == planID && session.Status == sharedtypes.SessionStatusFailed {
			count++
		}
	}
	return count
}

func (s *scenarioPracticeStore) forceSessionReplayDrift(sessionID string) {
	session, ok := s.sessions[sessionID]
	if !ok {
		return
	}
	session.Status = sharedtypes.SessionStatusCompleted
	session.TurnCount = 99
	if session.CurrentTurn != nil {
		turn := *session.CurrentTurn
		turn.QuestionText = "mutated question that must not appear in idempotency replay"
		session.CurrentTurn = &turn
	}
	s.sessions[sessionID] = session
}

func scenarioSessionRecordFromResponseBody(raw []byte) (domainpractice.SessionRecord, error) {
	var decoded api.PracticeSession
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return domainpractice.SessionRecord{}, err
	}
	createdAt, err := time.Parse(time.RFC3339, decoded.CreatedAt)
	if err != nil {
		return domainpractice.SessionRecord{}, err
	}
	updatedAt, err := time.Parse(time.RFC3339, decoded.UpdatedAt)
	if err != nil {
		return domainpractice.SessionRecord{}, err
	}
	session := domainpractice.SessionRecord{
		ID:           decoded.Id,
		PlanID:       decoded.PlanId,
		TargetJobID:  decoded.TargetJobId,
		Status:       decoded.Status,
		Language:     decoded.Language,
		HintsEnabled: decoded.HintsEnabled,
		TurnCount:    decoded.TurnCount,
		CreatedAt:    createdAt,
		UpdatedAt:    updatedAt,
	}
	if decoded.CurrentTurn != nil {
		var askedAt time.Time
		if decoded.CurrentTurn.AskedAt != nil && strings.TrimSpace(*decoded.CurrentTurn.AskedAt) != "" {
			askedAt, err = time.Parse(time.RFC3339, *decoded.CurrentTurn.AskedAt)
			if err != nil {
				return domainpractice.SessionRecord{}, err
			}
		}
		intent := ""
		if decoded.CurrentTurn.QuestionIntent != nil {
			intent = *decoded.CurrentTurn.QuestionIntent
		}
		session.CurrentTurn = &domainpractice.TurnRecord{
			ID:             decoded.CurrentTurn.Id,
			TurnIndex:      decoded.CurrentTurn.TurnIndex,
			QuestionText:   decoded.CurrentTurn.QuestionText,
			QuestionIntent: intent,
			Status:         decoded.CurrentTurn.Status,
			AskedAt:        askedAt,
		}
	}
	return session, nil
}

func (s *scenarioPracticeStore) idempotencyStatus(userID, domain, operation, rawKey string) idempotency.Status {
	rec, ok := s.idempotencyRecords[s.idempotencyRecordKey(userID, domain, operation, idempotency.HashKey(rawKey, ""))]
	if !ok {
		return ""
	}
	return rec.Status
}

func (s *scenarioPracticeStore) idempotencyRecordKey(userID, domain, operation, keyHash string) string {
	return strings.Join([]string{userID, domain, operation, keyHash}, "\x00")
}

var (
	_ domainpractice.Store = (*scenarioPracticeStore)(nil)
	_ idempotency.Store    = (*scenarioPracticeStore)(nil)
)

type practiceScenarioAuthStore struct {
	sessionSecret  string
	sessionsByHash map[string]auth.SessionRecord
	sessionsByID   map[string]string
	usersByID      map[string]auth.UserContext
}

func newPracticeScenarioAuthStore(sessionSecret string) *practiceScenarioAuthStore {
	return &practiceScenarioAuthStore{
		sessionSecret:  sessionSecret,
		sessionsByHash: map[string]auth.SessionRecord{},
		sessionsByID:   map[string]string{},
		usersByID:      map[string]auth.UserContext{},
	}
}

func (s *practiceScenarioAuthStore) addSession(userID, email, rawToken string) *http.Cookie {
	sessionHash := practiceScenarioSessionHash(s.sessionSecret, rawToken)
	user := auth.UserContext{ID: userID, Email: email, DisplayName: "Scenario User", UILanguage: "zh-CN", PreferredPracticeLanguage: "zh-CN", AnalyticsOptIn: true}
	s.usersByID[userID] = user
	session := auth.SessionRecord{
		ID:          "session-" + userID,
		UserID:      userID,
		SessionHash: sessionHash,
		Status:      auth.SessionStatusActive,
		ExpiresAt:   fixedScenarioNow().Add(auth.SessionTTL),
		CreatedAt:   fixedScenarioNow(),
		UpdatedAt:   fixedScenarioNow(),
	}
	s.sessionsByHash[sessionHash] = session
	s.sessionsByID[session.ID] = sessionHash
	return &http.Cookie{Name: auth.SessionCookieName, Value: rawToken}
}

func (s *practiceScenarioAuthStore) CountRecentChallenges(context.Context, string, string, time.Time) (int, error) {
	return 0, nil
}

func (s *practiceScenarioAuthStore) CreateChallenge(context.Context, auth.ChallengeRecord) error {
	return nil
}

func (s *practiceScenarioAuthStore) ConsumeChallenge(context.Context, string, time.Time) (auth.ChallengeRecord, error) {
	return auth.ChallengeRecord{}, auth.ErrChallengeInvalid
}

func (s *practiceScenarioAuthStore) FindOrCreateUserByEmail(context.Context, string, string, time.Time) (auth.UserContext, error) {
	return auth.UserContext{}, auth.ErrSessionInvalid
}

func (s *practiceScenarioAuthStore) CreateSession(_ context.Context, rec auth.SessionRecord) error {
	s.sessionsByHash[rec.SessionHash] = rec
	s.sessionsByID[rec.ID] = rec.SessionHash
	return nil
}

func (s *practiceScenarioAuthStore) GetSessionByHash(_ context.Context, sessionHash string, _ time.Time) (auth.SessionRecord, error) {
	session, ok := s.sessionsByHash[sessionHash]
	if !ok {
		return auth.SessionRecord{}, auth.ErrSessionInvalid
	}
	return session, nil
}

func (s *practiceScenarioAuthStore) GetUserContext(_ context.Context, userID string) (auth.UserContext, error) {
	user, ok := s.usersByID[userID]
	if !ok {
		return auth.UserContext{}, auth.ErrSessionInvalid
	}
	return user, nil
}

func (s *practiceScenarioAuthStore) TouchSession(_ context.Context, sessionID string, now time.Time, expiresAt time.Time) error {
	hash, ok := s.sessionsByID[sessionID]
	if !ok {
		return auth.ErrSessionInvalid
	}
	session := s.sessionsByHash[hash]
	session.UpdatedAt = now
	session.ExpiresAt = expiresAt
	s.sessionsByHash[hash] = session
	return nil
}

func (s *practiceScenarioAuthStore) RevokeSession(_ context.Context, sessionID string, now time.Time) error {
	hash, ok := s.sessionsByID[sessionID]
	if !ok {
		return auth.ErrSessionInvalid
	}
	session := s.sessionsByHash[hash]
	session.Status = auth.SessionStatusRevoked
	session.RevokedAt = now
	session.UpdatedAt = now
	s.sessionsByHash[hash] = session
	return nil
}

func (s *practiceScenarioAuthStore) CreatePrivacyDeleteHandoff(context.Context, string, string, string, string, time.Time) (auth.PrivacyDeleteHandoff, error) {
	return auth.PrivacyDeleteHandoff{}, fmt.Errorf("not used")
}

func practiceScenarioSessionHash(secret, rawToken string) string {
	sum := sha256.Sum256([]byte(secret + "\x00" + strings.TrimSpace(rawToken)))
	return hex.EncodeToString(sum[:])
}

func practiceBoolPtr(value bool) *bool {
	return &value
}

func containsString(values []string, needle string) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
}
