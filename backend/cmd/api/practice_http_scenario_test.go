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
	scenarioIdempotencyPepper   = "scenario-challenge-pepper"
)

func TestE2EP0007PracticeVoiceTurnHTTPRoute(t *testing.T) {
	ai := &scenarioPracticeAIClient{
		transcription:   "我主导了设计系统迁移，先把 12 个团队按风险分组。",
		responseText:    "你如何处理最高风险团队的迁移窗口？",
		responseIntent:  "voice.follow_up",
		synthesisAudio:  []byte("voice-tts-audio"),
		synthesisMillis: 1640,
	}
	h := newPracticeHTTPScenarioHarness(t, practiceHTTPScenarioOptions{ai: ai})
	plan := h.seedReadyScenarioPlan("practice-plan-p0-007", "target-job-p0-007-a", "resume-asset-p0-007-a", practiceHTTPScenarioUserAID)
	started := h.startScenarioSession(t, plan.ID, "e2e-p0-007-start-session")
	body := api.CreatePracticeVoiceTurnRequest{
		ClientVoiceTurnId: "client-voice-turn-p0-007",
		TurnId:            started.CurrentTurn.Id,
		Audio: api.PracticeVoiceAudioInput{
			ContentBase64: "T2dnUw==",
			ContentType:   "audio/webm",
			DurationMs:    4320,
		},
		Language:     "zh-CN",
		PracticeMode: sharedtypes.PracticeModeAssisted,
	}

	raw := h.doJSON(t, practiceHTTPScenarioUserAID, http.MethodPost, "/api/v1/practice/sessions/"+started.Id+"/voice-turns", "e2e-p0-007-voice-turn", body, http.StatusOK)
	var out api.PracticeVoiceTurnResult
	decodeJSON(t, raw, &out)
	if out.VoiceTurnId == "" ||
		out.UserTranscriptFinal != ai.transcription ||
		out.AssistantTextDraft != ai.responseText ||
		len(out.TtsChunks) != 1 ||
		out.TtsChunks[0].ContentType != "audio/mpeg" ||
		out.TtsChunks[0].ByteLength != int32(len(ai.synthesisAudio)) ||
		out.ProviderMetaSummary.SttProfile != "practice.voice.stt.default" ||
		out.ProviderMetaSummary.ChatProfile != "practice.followup.default" ||
		out.ProviderMetaSummary.TtsProfile != "practice.voice.tts.default" ||
		out.Session.Id != started.Id ||
		out.Session.CurrentTurn == nil ||
		out.Session.CurrentTurn.Status != string(domainpractice.TurnStatusFollowUpRequested) {
		t.Fatalf("voice turn HTTP response drift: %+v", out)
	}
	replayRaw := h.doJSON(t, practiceHTTPScenarioUserAID, http.MethodPost, "/api/v1/practice/sessions/"+started.Id+"/voice-turns", "e2e-p0-007-voice-turn", body, http.StatusOK)
	var replay api.PracticeVoiceTurnResult
	decodeJSON(t, replayRaw, &replay)
	if replay.VoiceTurnId != out.VoiceTurnId || h.store.sessionEventCount(started.Id) != 2 {
		t.Fatalf("voice turn replay should preserve payload and avoid duplicate events: first=%+v replay=%+v events=%d", out, replay, h.store.sessionEventCount(started.Id))
	}
	keyHash := idempotency.HashKey("e2e-p0-007-voice-turn", scenarioIdempotencyPepper)
	rec := h.store.idempotencyRecords[h.store.idempotencyRecordKey(practiceHTTPScenarioUserAID, "practice", "createPracticeVoiceTurn", keyHash)]
	if rec.Status != idempotency.StatusSucceeded || rec.ResourceType != "practice_voice_turn" || rec.ResourceID != out.VoiceTurnId {
		t.Fatalf("voice turn idempotency resource drift: %+v", rec)
	}
	payloads := append(h.store.sessionEventPayloads(started.Id), raw)
	assertNoEvidenceLeak(t, payloads, "T2dnUw==", string(ai.synthesisAudio), "audio/webm;")
}

func TestE2EP0022PracticePlanBaselineCreateAndRead(t *testing.T) {
	h := newPracticeHTTPScenarioHarness(t)
	body := api.CreatePracticePlanRequest{
		TargetJobId:          "target-job-p0-022-a",
		ResumeAssetId:        "resume-asset-p0-022-a",
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
	registry   *scenarioPracticeRegistry
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
	registryClient := opts.registry
	if registryClient == nil {
		registryClient = &scenarioPracticeRegistry{}
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
		Store:      store,
		Registry:   registryClient,
		AI:         aiClient,
		AITaskRuns: aiTaskRuns,
		Now:        fixedScenarioNow,
		NewID:      store.nextID,
	})
	practiceHandler := apipractice.NewHandler(apipractice.HandlerOptions{
		Service:              service,
		Session:              currentUserFromContext,
		IdempotencyKeyPepper: scenarioIdempotencyPepper,
	})
	routeIdempotency := idempotency.New(idempotency.MiddlewareOptions{
		Store:     store,
		Now:       fixedScenarioNow,
		NewID:     store.nextID,
		KeyPepper: scenarioIdempotencyPepper,
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

func TestE2EP0038PracticeEventLoopAnswerFlow(t *testing.T) {
	h := newPracticeHTTPScenarioHarness(t)
	plan := h.seedReadyScenarioPlan("practice-plan-p0-038", "target-job-p0-038-a", "resume-asset-p0-038-a", practiceHTTPScenarioUserAID)
	storedPlan := h.store.plans[plan.ID]
	storedPlan.QuestionBudget = 2
	h.store.plans[plan.ID] = storedPlan
	started := h.startScenarioSession(t, plan.ID, "e2e-p0-038-start-session")

	followUpRaw := h.doJSON(t, practiceHTTPScenarioUserAID, http.MethodPost, "/api/v1/practice/sessions/"+started.Id+"/events", "", api.PracticeSessionEventRequest{
		ClientEventId: "e2e-p0-038-event-1",
		Kind:          "answer_submitted",
		OccurredAt:    "2026-04-28T13:45:12Z",
		Payload: map[string]any{
			"turnId":        started.CurrentTurn.Id,
			"answerText":    "我先按影响面拆分迁移风险，再逐个团队确认窗口。",
			"followUpCount": 99,
		},
	}, http.StatusOK)
	var followUp api.SessionEventResult
	decodeJSON(t, followUpRaw, &followUp)
	if !followUp.Acknowledged ||
		followUp.AssistantAction.Type != "ask_follow_up" ||
		followUp.Session.CurrentTurn == nil ||
		followUp.Session.CurrentTurn.Id != started.CurrentTurn.Id ||
		followUp.Session.CurrentTurn.Status != "follow_up_requested" {
		t.Fatalf("first answer should request a same-turn follow-up from server-owned DB state: %+v", followUp)
	}
	if h.store.outboxCount() != 1 {
		t.Fatalf("follow-up request should not complete the turn or emit turn.completed yet, outbox=%d", h.store.outboxCount())
	}

	nextQuestionRaw := h.doJSON(t, practiceHTTPScenarioUserAID, http.MethodPost, "/api/v1/practice/sessions/"+started.Id+"/events", "", api.PracticeSessionEventRequest{
		ClientEventId: "e2e-p0-038-event-2",
		Kind:          "answer_submitted",
		OccurredAt:    "2026-04-28T13:46:12Z",
		Payload: map[string]any{
			"turnId":           started.CurrentTurn.Id,
			"answerText":       "追问里我会说明和安全团队确认风险接受标准。",
			"nextQuestionText": "请描述一次你在范围变化后重新排优先级的经历。",
		},
	}, http.StatusOK)
	var nextQuestion api.SessionEventResult
	decodeJSON(t, nextQuestionRaw, &nextQuestion)
	if nextQuestion.AssistantAction.Type != "ask_question" ||
		nextQuestion.Session.CurrentTurn == nil ||
		nextQuestion.Session.CurrentTurn.Id == started.CurrentTurn.Id ||
		nextQuestion.Session.CurrentTurn.Status != "asked" {
		t.Fatalf("second answer should complete the first turn and advance to a new question: %+v", nextQuestion)
	}

	completedRaw := h.doJSON(t, practiceHTTPScenarioUserAID, http.MethodPost, "/api/v1/practice/sessions/"+started.Id+"/events", "", api.PracticeSessionEventRequest{
		ClientEventId: "e2e-p0-038-event-3",
		Kind:          "answer_submitted",
		OccurredAt:    "2026-04-28T13:47:12Z",
		Payload: map[string]any{
			"turnId":     nextQuestion.Session.CurrentTurn.Id,
			"answerText": "第二题我会围绕影响面、用户风险和工程依赖解释取舍。",
		},
	}, http.StatusOK)
	var completed api.SessionEventResult
	decodeJSON(t, completedRaw, &completed)
	if completed.AssistantAction.Type != "session_completed" ||
		completed.Session.Status != sharedtypes.SessionStatusCompleted ||
		completed.Session.CurrentTurn == nil ||
		completed.Session.CurrentTurn.Status != "assessed" {
		t.Fatalf("third answer should complete the session at question budget: %+v", completed)
	}
	if h.store.sessionEventCount(started.Id) != 4 {
		t.Fatalf("session event count = %d, want 4", h.store.sessionEventCount(started.Id))
	}
	if h.store.outboxCount() != 3 {
		t.Fatalf("outbox count = %d, want 3 (session_started + two turn.completed)", h.store.outboxCount())
	}
}

func TestE2EP0039PracticeEventIdempotencyKindRouterAndHeaderPolicy(t *testing.T) {
	h := newPracticeHTTPScenarioHarness(t)
	plan := h.seedReadyScenarioPlan("practice-plan-p0-039", "target-job-p0-039-a", "resume-asset-p0-039-a", practiceHTTPScenarioUserAID)
	plan.Mode = sharedtypes.PracticeModeStrict
	h.store.plans[plan.ID] = scenarioPracticePlan{PlanRecord: plan, UserID: practiceHTTPScenarioUserAID, ResumeAssetID: "resume-asset-p0-039-a"}
	started := h.startScenarioSession(t, plan.ID, "e2e-p0-039-start-session")
	path := "/api/v1/practice/sessions/" + started.Id + "/events"

	body := api.PracticeSessionEventRequest{
		ClientEventId: "e2e-p0-039-resume",
		Kind:          "session_resumed",
		OccurredAt:    "2026-04-28T13:45:12Z",
		Payload:       map[string]any{"previousStatus": "waiting_user_input"},
	}
	first := h.doJSON(t, practiceHTTPScenarioUserAID, http.MethodPost, path, "", body, http.StatusOK)
	replay := h.doJSON(t, practiceHTTPScenarioUserAID, http.MethodPost, path, "", body, http.StatusOK)
	assertJSONEqualBytes(t, first, replay)

	body.Payload = map[string]any{"previousStatus": "running"}
	mismatchRaw := h.doJSON(t, practiceHTTPScenarioUserAID, http.MethodPost, path, "", body, http.StatusConflict)
	var mismatch api.ApiErrorResponse
	decodeJSON(t, mismatchRaw, &mismatch)
	if mismatch.Error.Code != sharederrors.CodePracticeSessionConflict || strings.Contains(string(mismatchRaw), "previousStatus") {
		t.Fatalf("mismatch should be conflict without leaking payload: %+v raw=%s", mismatch.Error, mismatchRaw)
	}

	headerRaw := h.doJSON(t, practiceHTTPScenarioUserAID, http.MethodPost, path, "must-not-use", api.PracticeSessionEventRequest{
		ClientEventId: "e2e-p0-039-header",
		Kind:          "turn_skipped",
		OccurredAt:    "2026-04-28T13:46:12Z",
		Payload:       map[string]any{"turnId": started.CurrentTurn.Id},
	}, http.StatusBadRequest)
	if !strings.Contains(string(headerRaw), "use_client_event_id") {
		t.Fatalf("header rejection missing policy: %s", headerRaw)
	}

	pausePlan := h.seedReadyScenarioPlan("practice-plan-p0-039-pause", "target-job-p0-039-pause", "resume-asset-p0-039-pause", practiceHTTPScenarioUserAID)
	pausedSession := h.startScenarioSession(t, pausePlan.ID, "e2e-p0-039-pause-start")
	pausePath := "/api/v1/practice/sessions/" + pausedSession.Id + "/events"
	pauseRaw := h.doJSON(t, practiceHTTPScenarioUserAID, http.MethodPost, pausePath, "", api.PracticeSessionEventRequest{
		ClientEventId: "e2e-p0-039-pause",
		Kind:          "session_paused",
		OccurredAt:    "2026-04-28T13:46:30Z",
	}, http.StatusOK)
	var paused api.SessionEventResult
	decodeJSON(t, pauseRaw, &paused)
	if paused.AssistantAction.Type != "session_wait" || paused.Session.Status != sharedtypes.SessionStatusWaitingUserInput {
		t.Fatalf("session_paused should route to session_wait + waiting_user_input: %+v", paused)
	}

	resumeRaw := h.doJSON(t, practiceHTTPScenarioUserAID, http.MethodPost, pausePath, "", api.PracticeSessionEventRequest{
		ClientEventId: "e2e-p0-039-resume-success",
		Kind:          "session_resumed",
		OccurredAt:    "2026-04-28T13:46:45Z",
	}, http.StatusOK)
	var resumed api.SessionEventResult
	decodeJSON(t, resumeRaw, &resumed)
	if resumed.AssistantAction.Type != "session_wait" || resumed.Session.Status != sharedtypes.SessionStatusRunning {
		t.Fatalf("session_resumed should route back to running session_wait: %+v", resumed)
	}

	skipPlan := h.seedReadyScenarioPlan("practice-plan-p0-039-skip", "target-job-p0-039-skip", "resume-asset-p0-039-skip", practiceHTTPScenarioUserAID)
	skipSession := h.startScenarioSession(t, skipPlan.ID, "e2e-p0-039-skip-start")
	skipRaw := h.doJSON(t, practiceHTTPScenarioUserAID, http.MethodPost, "/api/v1/practice/sessions/"+skipSession.Id+"/events", "", api.PracticeSessionEventRequest{
		ClientEventId: "e2e-p0-039-skip",
		Kind:          "turn_skipped",
		OccurredAt:    "2026-04-28T13:46:55Z",
		Payload:       map[string]any{"turnId": skipSession.CurrentTurn.Id, "nextQuestionText": "下一题请讲一个你推动对齐的案例。"},
	}, http.StatusOK)
	var skipped api.SessionEventResult
	decodeJSON(t, skipRaw, &skipped)
	if skipped.AssistantAction.Type != "ask_question" ||
		skipped.Session.CurrentTurn == nil ||
		skipped.Session.CurrentTurn.Id == skipSession.CurrentTurn.Id ||
		skipped.Session.CurrentTurn.Status != "asked" {
		t.Fatalf("turn_skipped should route to ask_question and advance turn: %+v", skipped)
	}

	answerPlan := h.seedReadyScenarioPlan("practice-plan-p0-039-answer", "target-job-p0-039-answer", "resume-asset-p0-039-answer", practiceHTTPScenarioUserAID)
	answerSession := h.startScenarioSession(t, answerPlan.ID, "e2e-p0-039-answer-start")
	answerRaw := h.doJSON(t, practiceHTTPScenarioUserAID, http.MethodPost, "/api/v1/practice/sessions/"+answerSession.Id+"/events", "", api.PracticeSessionEventRequest{
		ClientEventId: "e2e-p0-039-answer",
		Kind:          "answer_submitted",
		OccurredAt:    "2026-04-28T13:47:00Z",
		Payload:       map[string]any{"turnId": answerSession.CurrentTurn.Id, "answerText": "answer_submitted should route through the event state machine"},
	}, http.StatusOK)
	var answered api.SessionEventResult
	decodeJSON(t, answerRaw, &answered)
	if answered.AssistantAction.Type != "ask_follow_up" || answered.Session.CurrentTurn == nil || answered.Session.CurrentTurn.Status != "follow_up_requested" {
		t.Fatalf("answer_submitted should route to the server-owned follow-up branch: %+v", answered)
	}

	hintRaw := h.doJSON(t, practiceHTTPScenarioUserAID, http.MethodPost, path, "", api.PracticeSessionEventRequest{
		ClientEventId: "e2e-p0-039-hint",
		Kind:          "hint_requested",
		OccurredAt:    "2026-04-28T13:47:12Z",
		Payload:       map[string]any{"turnId": started.CurrentTurn.Id},
	}, http.StatusConflict)
	if !strings.Contains(string(hintRaw), "hint_disabled_in_mode") {
		t.Fatalf("hint strict conflict missing policy: %s", hintRaw)
	}

	crossRaw := h.doJSON(t, practiceHTTPScenarioUserBID, http.MethodPost, path, "", api.PracticeSessionEventRequest{
		ClientEventId: "e2e-p0-039-cross",
		Kind:          "session_paused",
		OccurredAt:    "2026-04-28T13:48:12Z",
	}, http.StatusNotFound)
	var cross api.ApiErrorResponse
	decodeJSON(t, crossRaw, &cross)
	if cross.Error.Code != sharederrors.CodePracticeSessionNotFound {
		t.Fatalf("cross-user should hide session existence: %+v", cross.Error)
	}
}

func TestE2EP0040PracticeEventConcurrentSeqNoStaleTurnConflict(t *testing.T) {
	h := newPracticeHTTPScenarioHarness(t)
	plan := h.seedReadyScenarioPlan("practice-plan-p0-040", "target-job-p0-040-a", "resume-asset-p0-040-a", practiceHTTPScenarioUserAID)
	started := h.startScenarioSession(t, plan.ID, "e2e-p0-040-start-session")
	path := "/api/v1/practice/sessions/" + started.Id + "/events"

	precondition := h.doJSON(t, practiceHTTPScenarioUserAID, http.MethodPost, path, "", api.PracticeSessionEventRequest{
		ClientEventId: "e2e-p0-040-follow-up",
		Kind:          "answer_submitted",
		OccurredAt:    "2026-04-28T13:45:00Z",
		Payload: map[string]any{
			"turnId":     started.CurrentTurn.Id,
			"answerText": "initial answer requests a same-turn follow up",
		},
	}, http.StatusOK)
	var followUp api.SessionEventResult
	decodeJSON(t, precondition, &followUp)
	if followUp.AssistantAction.Type != "ask_follow_up" || followUp.Session.CurrentTurn == nil || followUp.Session.CurrentTurn.Id != started.CurrentTurn.Id {
		t.Fatalf("precondition should keep the same current turn for follow-up: %+v", followUp)
	}

	first := h.doJSON(t, practiceHTTPScenarioUserAID, http.MethodPost, path, "", api.PracticeSessionEventRequest{
		ClientEventId: "e2e-p0-040-a",
		Kind:          "answer_submitted",
		OccurredAt:    "2026-04-28T13:45:12Z",
		Payload: map[string]any{
			"turnId":           started.CurrentTurn.Id,
			"answerText":       "first accepted follow-up answer",
			"nextQuestionText": "Next question after the accepted answer.",
		},
	}, http.StatusOK)
	var accepted api.SessionEventResult
	decodeJSON(t, first, &accepted)
	if accepted.Session.CurrentTurn == nil || accepted.Session.CurrentTurn.Id == started.CurrentTurn.Id {
		t.Fatalf("first accepted event should advance to a new current turn: %+v", accepted.Session.CurrentTurn)
	}

	stale := h.doJSON(t, practiceHTTPScenarioUserAID, http.MethodPost, path, "", api.PracticeSessionEventRequest{
		ClientEventId: "e2e-p0-040-b",
		Kind:          "answer_submitted",
		OccurredAt:    "2026-04-28T13:45:13Z",
		Payload: map[string]any{
			"turnId":     started.CurrentTurn.Id,
			"answerText": "stale competing answer",
		},
	}, http.StatusConflict)
	var conflict api.ApiErrorResponse
	decodeJSON(t, stale, &conflict)
	if conflict.Error.Code != sharederrors.CodePracticeSessionConflict || strings.Contains(string(stale), "stale competing answer") {
		t.Fatalf("stale competing event should conflict without leaking payload: %+v raw=%s", conflict.Error, stale)
	}
	if h.store.sessionEventCount(started.Id) != 3 {
		t.Fatalf("precondition and accepted append should write events only, got %d", h.store.sessionEventCount(started.Id))
	}
}

func TestE2EP0041PracticeSessionCompleteCreatesQueuedReportJob(t *testing.T) {
	h := newPracticeHTTPScenarioHarness(t)
	plan := h.seedReadyScenarioPlan("practice-plan-p0-041", "target-job-p0-041-a", "resume-asset-p0-041-a", practiceHTTPScenarioUserAID)
	started := h.startScenarioSession(t, plan.ID, "e2e-p0-041-start-session")

	raw := h.doJSON(t, practiceHTTPScenarioUserAID, http.MethodPost, "/api/v1/practice/sessions/"+started.Id+"/complete", "e2e-p0-041-complete", api.CompletePracticeSessionRequest{
		ClientCompletedAt: "2026-04-28T13:55:12Z",
	}, http.StatusAccepted)
	var out api.ReportWithJob
	decodeJSON(t, raw, &out)
	if out.ReportId == "" || out.Job.JobType != api.JobTypeReportGenerate || out.Job.Status != sharedtypes.JobStatusQueued || out.Job.ResourceType != api.ResourceTypeFeedbackReport {
		t.Fatalf("unexpected complete response: %+v", out)
	}
	if h.store.sessionEventCount(started.Id) != 2 || h.store.outboxCount() != 2 {
		t.Fatalf("complete should append one event and one outbox row, events=%d outbox=%d", h.store.sessionEventCount(started.Id), h.store.outboxCount())
	}
}

func TestE2EP0042PracticeSessionCompleteIdempotencyMatrix(t *testing.T) {
	h := newPracticeHTTPScenarioHarness(t)
	plan := h.seedReadyScenarioPlan("practice-plan-p0-042", "target-job-p0-042-a", "resume-asset-p0-042-a", practiceHTTPScenarioUserAID)
	started := h.startScenarioSession(t, plan.ID, "e2e-p0-042-start-session")
	path := "/api/v1/practice/sessions/" + started.Id + "/complete"
	body := api.CompletePracticeSessionRequest{ClientCompletedAt: "2026-04-28T13:55:12Z"}

	firstRaw := h.doJSON(t, practiceHTTPScenarioUserAID, http.MethodPost, path, "e2e-p0-042-k1", body, http.StatusAccepted)
	var first api.ReportWithJob
	decodeJSON(t, firstRaw, &first)
	replayRaw := h.doJSON(t, practiceHTTPScenarioUserAID, http.MethodPost, path, "e2e-p0-042-k1", body, http.StatusAccepted)
	assertJSONEqualBytes(t, firstRaw, replayRaw)

	h.doJSON(t, practiceHTTPScenarioUserAID, http.MethodPost, path, "e2e-p0-042-k1", api.CompletePracticeSessionRequest{
		ClientCompletedAt: "2026-04-28T14:00:00Z",
	}, http.StatusConflict)

	secondKeyRaw := h.doJSON(t, practiceHTTPScenarioUserAID, http.MethodPost, path, "e2e-p0-042-k2", api.CompletePracticeSessionRequest{
		ClientCompletedAt: "2026-04-28T14:01:00Z",
	}, http.StatusAccepted)
	var secondKey api.ReportWithJob
	decodeJSON(t, secondKeyRaw, &secondKey)
	if secondKey.ReportId != first.ReportId || secondKey.Job.Id != first.Job.Id {
		t.Fatalf("D-35 second key should replay existing report/job, first=%+v second=%+v", first, secondKey)
	}

	crossRaw := h.doJSON(t, practiceHTTPScenarioUserBID, http.MethodPost, path, "e2e-p0-042-cross", body, http.StatusNotFound)
	var cross api.ApiErrorResponse
	decodeJSON(t, crossRaw, &cross)
	if cross.Error.Code != sharederrors.CodePracticeSessionNotFound {
		t.Fatalf("cross-user complete should hide session existence: %+v", cross.Error)
	}
	blockedPlan := h.seedReadyScenarioPlan("practice-plan-p0-042-blocked", "target-job-p0-042-blocked", "resume-asset-p0-042-blocked", practiceHTTPScenarioUserAID)
	blocked := h.startScenarioSession(t, blockedPlan.ID, "e2e-p0-042-blocked-start")
	h.store.forceSessionStatus(blocked.Id, sharedtypes.SessionStatusFailed)
	blockedRaw := h.doJSON(t, practiceHTTPScenarioUserAID, http.MethodPost, "/api/v1/practice/sessions/"+blocked.Id+"/complete", "e2e-p0-042-blocked-complete", body, http.StatusConflict)
	var blockedErr api.ApiErrorResponse
	decodeJSON(t, blockedRaw, &blockedErr)
	if blockedErr.Error.Code != sharederrors.CodePracticeSessionConflict {
		t.Fatalf("illegal completion status should conflict: %+v", blockedErr.Error)
	}
	if len(h.store.completedReports) != 1 {
		t.Fatalf("complete should create one report, got %d", len(h.store.completedReports))
	}
}

func TestE2EP0043PracticeEventLoopPrivacyAndLegacyNegativeSurface(t *testing.T) {
	ai := &scenarioPracticeAIClient{responseText: "请补充你如何处理反对意见。", responseIntent: "behavioral.depth"}
	h := newPracticeHTTPScenarioHarness(t, practiceHTTPScenarioOptions{ai: ai, observedAI: true})
	plan := h.seedReadyScenarioPlan("01918fa0-0000-7000-8000-000000004043", "01918fa0-0000-7000-8000-000000002043", "resume-asset-p0-043-a", practiceHTTPScenarioUserAID)
	started := h.startScenarioSession(t, plan.ID, "e2e-p0-043-start-session")
	path := "/api/v1/practice/sessions/" + started.Id + "/events"

	h.doJSON(t, practiceHTTPScenarioUserAID, http.MethodPost, path, "", api.PracticeSessionEventRequest{
		ClientEventId: "e2e-p0-043-follow-up",
		Kind:          "answer_submitted",
		OccurredAt:    "2026-04-28T13:45:12Z",
		Payload: map[string]any{
			"turnId":     started.CurrentTurn.Id,
			"answerText": "answer_text prompt body response body provider secret sk-test",
		},
	}, http.StatusOK)
	h.doJSON(t, practiceHTTPScenarioUserAID, http.MethodPost, "/api/v1/practice/sessions/"+started.Id+"/complete", "e2e-p0-043-complete", api.CompletePracticeSessionRequest{
		ClientCompletedAt: "2026-04-28T13:55:12Z",
	}, http.StatusAccepted)

	raw := mustMarshalString(t, map[string]any{
		"outbox":      h.store.outboxPayloads(),
		"audit":       h.store.auditPayloads(),
		"ai_logs":     h.aiLogs.Entries(),
		"ai_task_run": h.aiTaskRuns.rows,
		"metric_runs": h.metrics.CounterLabelValues(observability.MetricRunsTotal),
	})
	for _, forbidden := range []string{"answer_text", "prompt body", "response body", "provider secret", "sk-test", "hint_text", "question_text"} {
		if strings.Contains(raw, forbidden) {
			t.Fatalf("privacy surface leaked forbidden evidence %q: %s", forbidden, raw)
		}
	}
	for _, labels := range h.metrics.CounterLabelValues(observability.MetricRunsTotal) {
		for _, label := range labels {
			if strings.Contains(label, "prompt.v1") || strings.Contains(label, "rubric.v1") {
				t.Fatalf("metric label leaked high-cardinality provenance: %v", labels)
			}
		}
	}
}

func TestE2EP0048PracticeHintAssistedAcrossGoals(t *testing.T) {
	for _, goal := range []sharedtypes.PracticeGoal{sharedtypes.PracticeGoalBaseline, sharedtypes.PracticeGoalRetryCurrentRound, sharedtypes.PracticeGoalNextRound, sharedtypes.PracticeGoalDebrief} {
		t.Run(string(goal), func(t *testing.T) {
			ai := &scenarioPracticeAIClient{}
			h := newPracticeHTTPScenarioHarness(t, practiceHTTPScenarioOptions{ai: ai, observedAI: true})
			targetID := map[sharedtypes.PracticeGoal]string{
				sharedtypes.PracticeGoalBaseline:          "01918fa0-0000-7000-8000-000000002048",
				sharedtypes.PracticeGoalRetryCurrentRound: "01918fa0-0000-7000-8000-000000002148",
				sharedtypes.PracticeGoalNextRound:         "01918fa0-0000-7000-8000-000000002248",
				sharedtypes.PracticeGoalDebrief:           "01918fa0-0000-7000-8000-000000002348",
			}[goal]
			plan := h.seedReadyScenarioPlan("practice-plan-p0-048-"+string(goal), targetID, "resume-asset-p0-048-"+string(goal), practiceHTTPScenarioUserAID)
			plan.Goal = goal
			plan.Mode = sharedtypes.PracticeModeAssisted
			if goal == sharedtypes.PracticeGoalDebrief {
				sourceDebriefID := "debrief-p0-048-source-" + string(goal)
				h.store.debriefs[sourceDebriefID] = scenarioDebrief{
					UserID:      practiceHTTPScenarioUserAID,
					TargetJobID: targetID,
					Status:      sharedtypes.DebriefStatusCompleted,
					Questions: []scenarioDebriefQuestion{{
						Text:   "Debrief source question for assisted hint policy.",
						Intent: "debrief.source_question",
					}},
				}
				plan.SourceDebriefID = sourceDebriefID
			}
			h.store.plans[plan.ID] = scenarioPracticePlan{PlanRecord: plan, UserID: practiceHTTPScenarioUserAID, ResumeAssetID: "resume-asset-p0-048-" + string(goal)}
			started := h.startScenarioSession(t, plan.ID, "e2e-p0-048-start-"+string(goal))

			raw := h.doJSON(t, practiceHTTPScenarioUserAID, http.MethodPost, "/api/v1/practice/sessions/"+started.Id+"/events", "", api.PracticeSessionEventRequest{
				ClientEventId: "e2e-p0-048-hint-" + string(goal),
				Kind:          "hint_requested",
				OccurredAt:    "2026-04-28T13:47:32Z",
				Payload:       map[string]any{"turnId": started.CurrentTurn.Id},
			}, http.StatusOK)
			var out api.SessionEventResult
			decodeJSON(t, raw, &out)
			if out.AssistantAction.Type != "show_hint" || out.AssistantAction.Hint == nil || *out.AssistantAction.Hint == "" {
				t.Fatalf("assisted hint should return show_hint: %+v", out.AssistantAction)
			}
			firstHint := *out.AssistantAction.Hint
			if out.Session.TurnCount != 1 || out.Session.Status != sharedtypes.SessionStatusRunning {
				t.Fatalf("hint should not advance session lifecycle: %+v", out.Session)
			}
			turns := h.store.turns[started.Id]
			if len(turns) != 1 || turns[0].Status != "asked" || turns[0].TurnIndex != 1 {
				t.Fatalf("hint should not mutate turn lifecycle: %+v", turns)
			}
			if got := h.store.hintTextForTurn(started.Id, started.CurrentTurn.Id); got == "" {
				t.Fatalf("assisted hint should persist hint_text for current turn")
			}
			if h.store.sessionEventCount(started.Id) != 2 {
				t.Fatalf("hint should append exactly one session event after start, got %d", h.store.sessionEventCount(started.Id))
			}
			if h.store.outboxCount() != 1 || h.store.auditCount() != 2 {
				t.Fatalf("hint should not add outbox/audit rows, outbox=%d audit=%d", h.store.outboxCount(), h.store.auditCount())
			}
			assertNoEvidenceLeak(t, h.store.sessionEventPayloads(started.Id), *out.AssistantAction.Hint, "hint_text", "prompt body", "response body", "provider secret")
			if len(h.aiTaskRuns.rows) == 0 || h.aiTaskRuns.rows[len(h.aiTaskRuns.rows)-1].Capability != aiclient.AITaskRunTaskHintGenerate ||
				h.aiTaskRuns.rows[len(h.aiTaskRuns.rows)-1].ValidationStatus != aiclient.ValidationStatusOK {
				t.Fatalf("hint ai_task_runs row missing: %+v", h.aiTaskRuns.rows)
			}

			ai.responseText = "Second hint for " + string(goal)
			secondBody := api.PracticeSessionEventRequest{
				ClientEventId: "e2e-p0-048-hint-second-" + string(goal),
				Kind:          "hint_requested",
				OccurredAt:    "2026-04-28T13:48:32Z",
				Payload:       map[string]any{"turnId": started.CurrentTurn.Id},
			}
			secondRaw := h.doJSON(t, practiceHTTPScenarioUserAID, http.MethodPost, "/api/v1/practice/sessions/"+started.Id+"/events", "", secondBody, http.StatusOK)
			var second api.SessionEventResult
			decodeJSON(t, secondRaw, &second)
			if second.AssistantAction.Hint == nil || *second.AssistantAction.Hint == firstHint {
				t.Fatalf("second hint should produce an independent response: first=%q second=%+v", firstHint, second.AssistantAction)
			}
			replayRaw := h.doJSON(t, practiceHTTPScenarioUserAID, http.MethodPost, "/api/v1/practice/sessions/"+started.Id+"/events", "", api.PracticeSessionEventRequest{
				ClientEventId: "e2e-p0-048-hint-" + string(goal),
				Kind:          "hint_requested",
				OccurredAt:    "2026-04-28T13:47:32Z",
				Payload:       map[string]any{"turnId": started.CurrentTurn.Id},
			}, http.StatusOK)
			var replay api.SessionEventResult
			decodeJSON(t, replayRaw, &replay)
			if replay.AssistantAction.Hint == nil || *replay.AssistantAction.Hint != firstHint {
				t.Fatalf("first clientEventId replay should return original hint, first=%q replay=%+v", firstHint, replay.AssistantAction)
			}
			if h.store.sessionEventCount(started.Id) != 3 {
				t.Fatalf("replay should not append a third hint event, got %d events", h.store.sessionEventCount(started.Id))
			}
			assertNoEvidenceLeak(t, h.store.sessionEventPayloads(started.Id), firstHint, *second.AssistantAction.Hint, "hint_text", "prompt body", "response body", "provider secret")
		})
	}
}

func TestE2EP0049PracticeHintStrictRefusalAcrossGoals(t *testing.T) {
	for _, goal := range []sharedtypes.PracticeGoal{sharedtypes.PracticeGoalBaseline, sharedtypes.PracticeGoalRetryCurrentRound, sharedtypes.PracticeGoalNextRound, sharedtypes.PracticeGoalDebrief} {
		t.Run(string(goal), func(t *testing.T) {
			h := newPracticeHTTPScenarioHarness(t)
			plan := h.seedReadyScenarioPlan("practice-plan-p0-049-"+string(goal), "target-job-p0-049-"+string(goal), "resume-asset-p0-049-"+string(goal), practiceHTTPScenarioUserAID)
			plan.Goal = goal
			plan.Mode = sharedtypes.PracticeModeStrict
			if goal == sharedtypes.PracticeGoalDebrief {
				sourceDebriefID := "debrief-p0-049-source-" + string(goal)
				h.store.debriefs[sourceDebriefID] = scenarioDebrief{
					UserID:      practiceHTTPScenarioUserAID,
					TargetJobID: plan.TargetJobID,
					Status:      sharedtypes.DebriefStatusCompleted,
					Questions: []scenarioDebriefQuestion{{
						Text:   "Debrief source question for strict hint policy.",
						Intent: "debrief.source_question",
					}},
				}
				plan.SourceDebriefID = sourceDebriefID
			}
			h.store.plans[plan.ID] = scenarioPracticePlan{PlanRecord: plan, UserID: practiceHTTPScenarioUserAID, ResumeAssetID: "resume-asset-p0-049-" + string(goal)}
			started := h.startScenarioSession(t, plan.ID, "e2e-p0-049-start-"+string(goal))
			body := api.PracticeSessionEventRequest{
				ClientEventId: "e2e-p0-049-hint-" + string(goal),
				Kind:          "hint_requested",
				OccurredAt:    "2026-04-28T13:47:32Z",
				Payload:       map[string]any{"turnId": started.CurrentTurn.Id},
			}
			raw := h.doJSON(t, practiceHTTPScenarioUserAID, http.MethodPost, "/api/v1/practice/sessions/"+started.Id+"/events", "", body, http.StatusConflict)
			replay := h.doJSON(t, practiceHTTPScenarioUserAID, http.MethodPost, "/api/v1/practice/sessions/"+started.Id+"/events", "", body, http.StatusConflict)
			if !strings.Contains(string(raw), "hint_disabled_in_mode") || !strings.Contains(string(replay), "hint_disabled_in_mode") {
				t.Fatalf("strict hint conflict/replay missing policy: first=%s replay=%s", raw, replay)
			}
			if h.store.sessionEventCount(started.Id) != 2 || h.store.pendingSessionEventCount(started.Id) != 0 {
				t.Fatalf("strict hint should finalize exactly one conflict event without pending rows: events=%d pending=%d", h.store.sessionEventCount(started.Id), h.store.pendingSessionEventCount(started.Id))
			}
			if got := h.store.hintTextForTurn(started.Id, started.CurrentTurn.Id); got != "" {
				t.Fatalf("strict hint should not persist hint_text, got %q", got)
			}
			if h.store.sessions[started.Id].Status != sharedtypes.SessionStatusRunning || h.store.outboxCount() != 1 || h.store.auditCount() != 2 || len(h.aiTaskRuns.rows) != 0 || h.store.aiCalledInsideTransaction {
				t.Fatalf("strict hint side effects drifted: session=%+v outbox=%d audit=%d taskRuns=%d aiInTx=%v", h.store.sessions[started.Id], h.store.outboxCount(), h.store.auditCount(), len(h.aiTaskRuns.rows), h.store.aiCalledInsideTransaction)
			}
		})
	}
}

func TestE2EP0050PracticeAssistantActionProvenanceAndTaskRuns(t *testing.T) {
	ai := &scenarioPracticeAIClient{responseText: "Use a concrete metric."}
	h := newPracticeHTTPScenarioHarness(t, practiceHTTPScenarioOptions{ai: ai, observedAI: true})
	plan := h.seedReadyScenarioPlan("practice-plan-p0-050", "01918fa0-0000-7000-8000-000000002050", "resume-asset-p0-050", practiceHTTPScenarioUserAID)
	plan.QuestionBudget = 2
	h.store.plans[plan.ID] = scenarioPracticePlan{PlanRecord: plan, UserID: practiceHTTPScenarioUserAID, ResumeAssetID: "resume-asset-p0-050"}
	started := h.startScenarioSession(t, plan.ID, "e2e-p0-050-start")
	actions := make([]api.AssistantAction, 0, 5)
	followRaw := h.doJSON(t, practiceHTTPScenarioUserAID, http.MethodPost, "/api/v1/practice/sessions/"+started.Id+"/events", "", api.PracticeSessionEventRequest{
		ClientEventId: "e2e-p0-050-follow",
		Kind:          "answer_submitted",
		OccurredAt:    "2026-04-28T13:47:20Z",
		Payload:       map[string]any{"turnId": started.CurrentTurn.Id, "answerText": "initial answer"},
	}, http.StatusOK)
	var follow api.SessionEventResult
	decodeJSON(t, followRaw, &follow)
	actions = append(actions, follow.AssistantAction)

	raw := h.doJSON(t, practiceHTTPScenarioUserAID, http.MethodPost, "/api/v1/practice/sessions/"+started.Id+"/events", "", api.PracticeSessionEventRequest{
		ClientEventId: "e2e-p0-050-hint",
		Kind:          "hint_requested",
		OccurredAt:    "2026-04-28T13:47:32Z",
		Payload:       map[string]any{"turnId": started.CurrentTurn.Id},
	}, http.StatusOK)
	var out api.SessionEventResult
	decodeJSON(t, raw, &out)
	actions = append(actions, out.AssistantAction)

	skipRaw := h.doJSON(t, practiceHTTPScenarioUserAID, http.MethodPost, "/api/v1/practice/sessions/"+started.Id+"/events", "", api.PracticeSessionEventRequest{
		ClientEventId: "e2e-p0-050-skip",
		Kind:          "turn_skipped",
		OccurredAt:    "2026-04-28T13:47:40Z",
		Payload:       map[string]any{"turnId": started.CurrentTurn.Id, "nextQuestionText": "Next question."},
	}, http.StatusOK)
	var skipped api.SessionEventResult
	decodeJSON(t, skipRaw, &skipped)
	actions = append(actions, skipped.AssistantAction)

	pauseRaw := h.doJSON(t, practiceHTTPScenarioUserAID, http.MethodPost, "/api/v1/practice/sessions/"+started.Id+"/events", "", api.PracticeSessionEventRequest{
		ClientEventId: "e2e-p0-050-pause",
		Kind:          "session_paused",
		OccurredAt:    "2026-04-28T13:47:50Z",
	}, http.StatusOK)
	var paused api.SessionEventResult
	decodeJSON(t, pauseRaw, &paused)
	actions = append(actions, paused.AssistantAction)

	completeRaw := h.doJSON(t, practiceHTTPScenarioUserAID, http.MethodPost, "/api/v1/practice/sessions/"+started.Id+"/events", "", api.PracticeSessionEventRequest{
		ClientEventId: "e2e-p0-050-complete-action",
		Kind:          "answer_submitted",
		OccurredAt:    "2026-04-28T13:48:00Z",
		Payload:       map[string]any{"turnId": skipped.Session.CurrentTurn.Id, "answerText": "final answer"},
	}, http.StatusOK)
	var completed api.SessionEventResult
	decodeJSON(t, completeRaw, &completed)
	actions = append(actions, completed.AssistantAction)

	wantTypes := []string{"ask_follow_up", "show_hint", "ask_question", "session_wait", "session_completed"}
	for i, action := range actions {
		if action.Type != wantTypes[i] {
			t.Fatalf("action[%d] type = %q, want %q", i, action.Type, wantTypes[i])
		}
		assertProvenanceWireOnly(t, action.Provenance)
	}
	rows := h.aiTaskRuns.Rows()
	if len(rows) != 3 ||
		rows[1].Capability != aiclient.AITaskRunTaskFollowupGenerate ||
		rows[2].Capability != aiclient.AITaskRunTaskHintGenerate ||
		rows[2].FeatureKey != hintFeatureKeyForScenario {
		t.Fatalf("runtime task run should carry feature key: %+v", h.aiTaskRuns.rows)
	}
}

func TestE2EP0051PracticeHintDegradeAndPrivacy(t *testing.T) {
	cases := []struct {
		name       string
		registry   *scenarioPracticeRegistry
		ai         *scenarioPracticeAIClient
		wantCode   string
		observedAI bool
	}{
		{name: "f3-unsupported", registry: &scenarioPracticeRegistry{errByFeature: map[string]error{hintFeatureKeyForScenario: registry.ErrPromptUnsupported}}, ai: &scenarioPracticeAIClient{}, wantCode: sharederrors.CodeAiProviderConfigInvalid},
		{name: "secret-missing", ai: &scenarioPracticeAIClient{hintFailures: []error{sharederrors.Wrap(sharederrors.CodeAiProviderSecretMissing, "provider secret sk-test", false)}}, wantCode: sharederrors.CodeAiProviderSecretMissing, observedAI: true},
		{name: "timeout", ai: &scenarioPracticeAIClient{hintFailures: []error{sharederrors.Wrap(sharederrors.CodeAiProviderTimeout, "timeout prompt body response body", true)}}, wantCode: sharederrors.CodeAiProviderTimeout, observedAI: true},
		{name: "invalid-output", ai: &scenarioPracticeAIClient{rawContent: `{"hint":""}`}, wantCode: sharederrors.CodeAiOutputInvalid, observedAI: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := newPracticeHTTPScenarioHarness(t, practiceHTTPScenarioOptions{ai: tc.ai, registry: tc.registry, observedAI: tc.observedAI})
			plan := h.seedReadyScenarioPlan("practice-plan-p0-051-"+tc.name, "01918fa0-0000-7000-8000-000000002051", "resume-asset-p0-051-"+tc.name, practiceHTTPScenarioUserAID)
			started := h.startScenarioSession(t, plan.ID, "e2e-p0-051-start-"+tc.name)
			raw := h.doJSON(t, practiceHTTPScenarioUserAID, http.MethodPost, "/api/v1/practice/sessions/"+started.Id+"/events", "", api.PracticeSessionEventRequest{
				ClientEventId: "e2e-p0-051-hint-" + tc.name,
				Kind:          "hint_requested",
				OccurredAt:    "2026-04-28T13:47:32Z",
				Payload:       map[string]any{"turnId": started.CurrentTurn.Id},
			}, http.StatusOK)
			var out api.SessionEventResult
			decodeJSON(t, raw, &out)
			if out.AssistantAction.Type != "session_wait" || out.Session.Status != sharedtypes.SessionStatusRunning || h.store.sessions[started.Id].FailureCode != "" {
				t.Fatalf("hint failure should gracefully degrade without failing session: out=%+v session=%+v", out, h.store.sessions[started.Id])
			}
			if got := h.store.hintTextForTurn(started.Id, started.CurrentTurn.Id); got != "" {
				t.Fatalf("degrade should not persist hint_text, got %q", got)
			}
			rows := h.aiTaskRuns.Rows()
			if len(rows) == 0 || rows[len(rows)-1].Capability != aiclient.AITaskRunTaskHintGenerate ||
				rows[len(rows)-1].ValidationStatus != aiclient.ValidationStatusInvalid ||
				rows[len(rows)-1].ErrorCode != tc.wantCode {
				t.Fatalf("degrade should write failed hint_generate row with %s: %+v", tc.wantCode, rows)
			}
			evidence := mustMarshalString(t, map[string]any{
				"session_events": h.store.sessionEvents[started.Id],
				"ai_task_runs":   rows,
				"audit":          h.store.auditPayloads(),
				"outbox":         h.store.outboxPayloads(),
				"ai_logs":        h.aiLogs.Entries(),
				"metric_labels":  h.metrics.CounterLabelValues(observability.MetricRunsTotal),
			})
			for _, forbidden := range []string{"hint_text", "prompt body", "response body", "provider secret", "sk-test"} {
				if strings.Contains(evidence, forbidden) {
					t.Fatalf("degrade privacy surface leaked %q: %s", forbidden, evidence)
				}
			}
		})
	}
}

func TestE2EP0070PracticeDerivedPlanCreateReadReplay(t *testing.T) {
	h := newPracticeHTTPScenarioHarness(t)
	targetID := "target-job-p0-070-a"
	resumeID := "resume-asset-p0-070-a"
	h.store.prerequisiteTargetOwner[targetID] = practiceHTTPScenarioUserAID
	h.store.prerequisiteResumeOwner[resumeID] = practiceHTTPScenarioUserAID
	reportID := "report-p0-070-ready"
	debriefID := "debrief-p0-070-completed"
	h.store.reports[reportID] = scenarioFeedbackReport{UserID: practiceHTTPScenarioUserAID, TargetJobID: targetID, Status: sharedtypes.ReportStatusReady}
	h.store.debriefs[debriefID] = scenarioDebrief{
		UserID:      practiceHTTPScenarioUserAID,
		TargetJobID: targetID,
		Status:      sharedtypes.DebriefStatusCompleted,
		Questions:   []scenarioDebriefQuestion{{Text: "What did the interviewer ask about scope?", Intent: "debrief.scope"}},
	}

	tests := []struct {
		name              string
		goal              sharedtypes.PracticeGoal
		sourceReportID    *string
		sourceDebriefID   *string
		wantSourceReport  string
		wantSourceDebrief string
		idempotencyKey    string
	}{
		{name: "retry", goal: sharedtypes.PracticeGoalRetryCurrentRound, sourceReportID: strPtr(reportID), wantSourceReport: reportID, idempotencyKey: "e2e-p0-070-retry"},
		{name: "next-round", goal: sharedtypes.PracticeGoalNextRound, sourceReportID: strPtr(reportID), wantSourceReport: reportID, idempotencyKey: "e2e-p0-070-next"},
		{name: "debrief", goal: sharedtypes.PracticeGoalDebrief, sourceDebriefID: strPtr(debriefID), wantSourceDebrief: debriefID, idempotencyKey: "e2e-p0-070-debrief"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			body := api.CreatePracticePlanRequest{
				TargetJobId:          targetID,
				ResumeAssetId:        resumeID,
				SourceReportId:       tc.sourceReportID,
				SourceDebriefId:      tc.sourceDebriefID,
				Goal:                 tc.goal,
				Mode:                 sharedtypes.PracticeModeAssisted,
				InterviewerPersona:   sharedtypes.InterviewerRoleHiringManager,
				Difficulty:           "standard",
				Language:             "zh-CN",
				TimeBudgetMinutes:    30,
				QuestionBudget:       6,
				FocusCompetencyCodes: []string{"system-design"},
			}
			raw := h.doJSON(t, practiceHTTPScenarioUserAID, http.MethodPost, "/api/v1/practice/plans", tc.idempotencyKey, body, http.StatusCreated)
			var created api.PracticePlan
			decodeJSON(t, raw, &created)
			if created.Goal != tc.goal || created.SourceReportId == nil != (tc.wantSourceReport == "") || created.SourceDebriefId == nil != (tc.wantSourceDebrief == "") {
				t.Fatalf("derived plan source shape mismatch: %+v", created)
			}
			if tc.wantSourceReport != "" && *created.SourceReportId != tc.wantSourceReport {
				t.Fatalf("sourceReportId = %v, want %s", created.SourceReportId, tc.wantSourceReport)
			}
			if tc.wantSourceDebrief != "" && *created.SourceDebriefId != tc.wantSourceDebrief {
				t.Fatalf("sourceDebriefId = %v, want %s", created.SourceDebriefId, tc.wantSourceDebrief)
			}

			replayRaw := h.doJSON(t, practiceHTTPScenarioUserAID, http.MethodPost, "/api/v1/practice/plans", tc.idempotencyKey, body, http.StatusCreated)
			var replay api.PracticePlan
			decodeJSON(t, replayRaw, &replay)
			if replay.Id != created.Id || replay.SourceReportId == nil != (created.SourceReportId == nil) || replay.SourceDebriefId == nil != (created.SourceDebriefId == nil) {
				t.Fatalf("idempotency replay changed source response: created=%+v replay=%+v", created, replay)
			}

			detailRaw := h.doJSON(t, practiceHTTPScenarioUserAID, http.MethodGet, "/api/v1/practice/plans/"+created.Id, "", nil, http.StatusOK)
			var detail api.PracticePlan
			decodeJSON(t, detailRaw, &detail)
			if detail.Id != created.Id || detail.SourceReportId == nil != (created.SourceReportId == nil) || detail.SourceDebriefId == nil != (created.SourceDebriefId == nil) {
				t.Fatalf("getPracticePlan lost source fields: created=%+v detail=%+v", created, detail)
			}
		})
	}
}

func TestE2EP0071PracticeDebriefStartUsesSourceQuestion(t *testing.T) {
	ai := &scenarioPracticeAIClient{}
	h := newPracticeHTTPScenarioHarness(t, practiceHTTPScenarioOptions{ai: ai})
	targetID := "target-job-p0-071-a"
	resumeID := "resume-asset-p0-071-a"
	sourceDebriefID := "debrief-p0-071-completed"
	h.store.prerequisiteTargetOwner[targetID] = practiceHTTPScenarioUserAID
	h.store.prerequisiteResumeOwner[resumeID] = practiceHTTPScenarioUserAID
	h.store.debriefs[sourceDebriefID] = scenarioDebrief{
		UserID:      practiceHTTPScenarioUserAID,
		TargetJobID: targetID,
		Status:      sharedtypes.DebriefStatusCompleted,
		Questions: []scenarioDebriefQuestion{{
			Text:   "__DEBRIEF_FIRST_QUESTION__",
			Intent: "debrief.source_question",
		}},
	}
	planRaw := h.doJSON(t, practiceHTTPScenarioUserAID, http.MethodPost, "/api/v1/practice/plans", "e2e-p0-071-create", api.CreatePracticePlanRequest{
		TargetJobId:        targetID,
		ResumeAssetId:      resumeID,
		SourceDebriefId:    strPtr(sourceDebriefID),
		Goal:               sharedtypes.PracticeGoalDebrief,
		Mode:               sharedtypes.PracticeModeStrict,
		InterviewerPersona: sharedtypes.InterviewerRoleHiringManager,
		Difficulty:         "standard",
		Language:           "zh-CN",
		TimeBudgetMinutes:  30,
		QuestionBudget:     6,
	}, http.StatusCreated)
	var plan api.PracticePlan
	decodeJSON(t, planRaw, &plan)

	startRaw := h.doJSON(t, practiceHTTPScenarioUserAID, http.MethodPost, "/api/v1/practice/sessions", "e2e-p0-071-start", api.StartPracticeSessionRequest{PlanId: plan.Id}, http.StatusCreated)
	var started api.PracticeSession
	decodeJSON(t, startRaw, &started)
	if started.Status != sharedtypes.SessionStatusRunning || started.CurrentTurn == nil || started.CurrentTurn.QuestionText != "__DEBRIEF_FIRST_QUESTION__" {
		t.Fatalf("debrief start did not use source question: %+v", started)
	}
	if ai.calls != 0 {
		t.Fatalf("debrief start must not call first_question AI, calls=%d", ai.calls)
	}
	replayRaw := h.doJSON(t, practiceHTTPScenarioUserAID, http.MethodPost, "/api/v1/practice/sessions", "e2e-p0-071-start", api.StartPracticeSessionRequest{PlanId: plan.Id}, http.StatusCreated)
	var replay api.PracticeSession
	decodeJSON(t, replayRaw, &replay)
	if replay.Id != started.Id || replay.CurrentTurn == nil || replay.CurrentTurn.QuestionText != started.CurrentTurn.QuestionText || ai.calls != 0 {
		t.Fatalf("debrief replay changed currentTurn or called AI: started=%+v replay=%+v calls=%d", started, replay, ai.calls)
	}
	assertNoEvidenceLeak(t, h.store.outboxPayloads(), "__DEBRIEF_FIRST_QUESTION__", "question_text", "answer_text", "hint_text")
}

func TestE2EP0072PracticeDerivedSourceValidationIsolationPrivacy(t *testing.T) {
	h := newPracticeHTTPScenarioHarness(t)
	targetID := "target-job-p0-072-a"
	resumeID := "resume-asset-p0-072-a"
	h.store.prerequisiteTargetOwner[targetID] = practiceHTTPScenarioUserAID
	h.store.prerequisiteResumeOwner[resumeID] = practiceHTTPScenarioUserAID
	h.store.reports["report-p0-072-cross-user"] = scenarioFeedbackReport{UserID: practiceHTTPScenarioUserBID, TargetJobID: targetID, Status: sharedtypes.ReportStatusReady}
	h.store.reports["report-p0-072-wrong-target"] = scenarioFeedbackReport{UserID: practiceHTTPScenarioUserAID, TargetJobID: "target-other", Status: sharedtypes.ReportStatusReady}
	h.store.debriefs["debrief-p0-072-draft"] = scenarioDebrief{UserID: practiceHTTPScenarioUserAID, TargetJobID: targetID, Status: sharedtypes.DebriefStatusDraft, Questions: []scenarioDebriefQuestion{{Text: "__PRIVATE_DEBRIEF_TEXT__"}}}
	h.store.debriefs["debrief-p0-072-empty"] = scenarioDebrief{UserID: practiceHTTPScenarioUserAID, TargetJobID: targetID, Status: sharedtypes.DebriefStatusCompleted}

	tests := []struct {
		name            string
		goal            sharedtypes.PracticeGoal
		sourceReportID  *string
		sourceDebriefID *string
		sourceID        string
	}{
		{name: "missing report", goal: sharedtypes.PracticeGoalRetryCurrentRound, sourceReportID: strPtr("report-p0-072-missing"), sourceID: "report-p0-072-missing"},
		{name: "cross user report", goal: sharedtypes.PracticeGoalRetryCurrentRound, sourceReportID: strPtr("report-p0-072-cross-user"), sourceID: "report-p0-072-cross-user"},
		{name: "wrong target report", goal: sharedtypes.PracticeGoalNextRound, sourceReportID: strPtr("report-p0-072-wrong-target"), sourceID: "report-p0-072-wrong-target"},
		{name: "draft debrief", goal: sharedtypes.PracticeGoalDebrief, sourceDebriefID: strPtr("debrief-p0-072-draft"), sourceID: "debrief-p0-072-draft"},
		{name: "empty debrief", goal: sharedtypes.PracticeGoalDebrief, sourceDebriefID: strPtr("debrief-p0-072-empty"), sourceID: "debrief-p0-072-empty"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			raw := h.doJSON(t, practiceHTTPScenarioUserAID, http.MethodPost, "/api/v1/practice/plans", "e2e-p0-072-"+tc.name, api.CreatePracticePlanRequest{
				TargetJobId:        targetID,
				ResumeAssetId:      resumeID,
				SourceReportId:     tc.sourceReportID,
				SourceDebriefId:    tc.sourceDebriefID,
				Goal:               tc.goal,
				Mode:               sharedtypes.PracticeModeAssisted,
				InterviewerPersona: sharedtypes.InterviewerRoleHiringManager,
				Difficulty:         "standard",
				Language:           "zh-CN",
				TimeBudgetMinutes:  30,
				QuestionBudget:     6,
			}, http.StatusUnprocessableEntity)
			var out api.ApiErrorResponse
			decodeJSON(t, raw, &out)
			if out.Error.Code != sharederrors.CodeValidationFailed || strings.Contains(string(raw), tc.sourceID) || strings.Contains(string(raw), "__PRIVATE_DEBRIEF_TEXT__") {
				t.Fatalf("source validation leaked resource evidence: error=%+v body=%s", out.Error, string(raw))
			}
		})
	}
}

func TestE2EP0073PracticeDebriefAssistedStrictAndLegacyNegative(t *testing.T) {
	ai := &scenarioPracticeAIClient{}
	h := newPracticeHTTPScenarioHarness(t, practiceHTTPScenarioOptions{ai: ai})
	targetID := "target-job-p0-073-a"
	resumeID := "resume-asset-p0-073-a"
	sourceDebriefID := "debrief-p0-073-completed"
	h.store.prerequisiteTargetOwner[targetID] = practiceHTTPScenarioUserAID
	h.store.prerequisiteResumeOwner[resumeID] = practiceHTTPScenarioUserAID
	h.store.debriefs[sourceDebriefID] = scenarioDebrief{
		UserID:      practiceHTTPScenarioUserAID,
		TargetJobID: targetID,
		Status:      sharedtypes.DebriefStatusCompleted,
		Questions:   []scenarioDebriefQuestion{{Text: "Debrief mode regression question?", Intent: "debrief.regression"}},
	}

	for _, mode := range []sharedtypes.PracticeMode{sharedtypes.PracticeModeAssisted, sharedtypes.PracticeModeStrict} {
		t.Run(string(mode), func(t *testing.T) {
			planRaw := h.doJSON(t, practiceHTTPScenarioUserAID, http.MethodPost, "/api/v1/practice/plans", "e2e-p0-073-create-"+string(mode), api.CreatePracticePlanRequest{
				TargetJobId:        targetID,
				ResumeAssetId:      resumeID,
				SourceDebriefId:    strPtr(sourceDebriefID),
				Goal:               sharedtypes.PracticeGoalDebrief,
				Mode:               mode,
				InterviewerPersona: sharedtypes.InterviewerRoleHiringManager,
				Difficulty:         "standard",
				Language:           "zh-CN",
				TimeBudgetMinutes:  30,
				QuestionBudget:     6,
			}, http.StatusCreated)
			var plan api.PracticePlan
			decodeJSON(t, planRaw, &plan)
			startRaw := h.doJSON(t, practiceHTTPScenarioUserAID, http.MethodPost, "/api/v1/practice/sessions", "e2e-p0-073-start-"+string(mode), api.StartPracticeSessionRequest{PlanId: plan.Id}, http.StatusCreated)
			var started api.PracticeSession
			decodeJSON(t, startRaw, &started)
			if started.CurrentTurn == nil || started.CurrentTurn.QuestionText != "Debrief mode regression question?" {
				t.Fatalf("debrief %s start mismatch: %+v", mode, started)
			}
		})
	}
	if ai.calls != 0 {
		t.Fatalf("debrief assisted/strict start must not call first_question AI, calls=%d", ai.calls)
	}
	raw := h.doJSON(t, practiceHTTPScenarioUserAID, http.MethodPost, "/api/v1/practice/plans", "e2e-p0-073-legacy-mode", api.CreatePracticePlanRequest{
		TargetJobId:        targetID,
		ResumeAssetId:      resumeID,
		SourceDebriefId:    strPtr(sourceDebriefID),
		Goal:               sharedtypes.PracticeGoalDebrief,
		Mode:               sharedtypes.PracticeMode("debrief"),
		InterviewerPersona: sharedtypes.InterviewerRoleHiringManager,
		Difficulty:         "standard",
		Language:           "zh-CN",
		TimeBudgetMinutes:  30,
		QuestionBudget:     6,
	}, http.StatusUnprocessableEntity)
	var out api.ApiErrorResponse
	decodeJSON(t, raw, &out)
	if out.Error.Code != sharederrors.CodeValidationFailed {
		t.Fatalf("legacy mode should be rejected with validation error: %+v", out.Error)
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

func (h *practiceHTTPScenarioHarness) startScenarioSession(t *testing.T, planID, idempotencyKey string) api.PracticeSession {
	t.Helper()
	raw := h.doJSON(t, practiceHTTPScenarioUserAID, http.MethodPost, "/api/v1/practice/sessions", idempotencyKey, api.StartPracticeSessionRequest{
		PlanId:       planID,
		HintsEnabled: practiceBoolPtr(true),
	}, http.StatusCreated)
	var started api.PracticeSession
	decodeJSON(t, raw, &started)
	if started.Id == "" || started.CurrentTurn == nil {
		t.Fatalf("startPracticeSession did not return a current turn: %+v", started)
	}
	return started
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

func assertJSONEqualBytes(t *testing.T, want, got []byte) {
	t.Helper()
	var wantValue any
	var gotValue any
	if err := json.Unmarshal(want, &wantValue); err != nil {
		t.Fatalf("decode want: %v", err)
	}
	if err := json.Unmarshal(got, &gotValue); err != nil {
		t.Fatalf("decode got: %v", err)
	}
	wantRaw, _ := json.Marshal(wantValue)
	gotRaw, _ := json.Marshal(gotValue)
	if !bytes.Equal(wantRaw, gotRaw) {
		t.Fatalf("json mismatch\nwant: %s\n got: %s", wantRaw, gotRaw)
	}
}

func mustMarshalString(t *testing.T, value any) string {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal value: %v", err)
	}
	return string(raw)
}

func assertProvenanceWireOnly(t *testing.T, provenance api.GenerationProvenance) {
	t.Helper()
	raw, err := json.Marshal(provenance)
	if err != nil {
		t.Fatalf("marshal provenance: %v", err)
	}
	var keys map[string]any
	if err := json.Unmarshal(raw, &keys); err != nil {
		t.Fatalf("unmarshal provenance: %v", err)
	}
	want := map[string]struct{}{
		"promptVersion":     {},
		"rubricVersion":     {},
		"modelId":           {},
		"language":          {},
		"featureFlag":       {},
		"dataSourceVersion": {},
	}
	if len(keys) != len(want) {
		t.Fatalf("provenance key count drifted: %s", raw)
	}
	for key := range keys {
		if _, ok := want[key]; !ok {
			t.Fatalf("provenance leaked runtime key %q: %s", key, raw)
		}
	}
	for _, forbidden := range []string{"featureKey", "feature_key", "modelProfileName", "model_profile_name", "provider", "cost", "latency", "capability"} {
		if strings.Contains(string(raw), forbidden) {
			t.Fatalf("provenance leaked runtime field %q: %s", forbidden, raw)
		}
	}
}

type scenarioPracticeRegistry struct {
	errByFeature map[string]error
}

func (r *scenarioPracticeRegistry) ResolveActive(ctx context.Context, featureKey, language string) (registry.PromptResolution, error) {
	if r != nil && r.errByFeature != nil {
		if err := r.errByFeature[featureKey]; err != nil {
			return registry.PromptResolution{}, err
		}
	}
	profileName := "practice.first_question.default"
	if featureKey == "practice.session.follow_up" {
		profileName = "practice.followup.default"
	}
	if featureKey == hintFeatureKeyForScenario {
		profileName = "practice.turn_observe.default"
	}
	if featureKey == "practice.voice.stt" {
		profileName = "practice.voice.stt.default"
	}
	if featureKey == "practice.voice.tts" {
		profileName = "practice.voice.tts.default"
	}
	return registry.PromptResolution{
		FeatureKey:          featureKey,
		PromptVersion:       "prompt.v1",
		RubricVersion:       "rubric.v1",
		ModelProfileName:    profileName,
		DataSourceVersion:   "registry.v1",
		FeatureFlag:         "none",
		UserMessageTemplate: "ask the first interview question",
	}, nil
}

type scenarioPracticeProfileResolver struct{}

func (r scenarioPracticeProfileResolver) Resolve(name string) (*aiclient.ModelProfile, error) {
	if name != "practice.first_question.default" &&
		name != "practice.followup.default" &&
		name != "practice.turn_observe.default" &&
		name != "practice.voice.stt.default" &&
		name != "practice.voice.tts.default" {
		return nil, fmt.Errorf("missing scenario profile %q", name)
	}
	route := "practice.session.first_question"
	capability := aiclient.CapabilityChat
	if name == "practice.followup.default" {
		route = "practice.session.follow_up"
	}
	if name == "practice.turn_observe.default" {
		route = hintFeatureKeyForScenario
	}
	if name == "practice.voice.stt.default" {
		route = "practice.voice.stt"
		capability = aiclient.CapabilitySTT
	}
	if name == "practice.voice.tts.default" {
		route = "practice.voice.tts"
		capability = aiclient.CapabilityTts
	}
	return &aiclient.ModelProfile{
		Name:       name,
		Capability: capability,
		Status:     aiclient.ProfileStatusActive,
		Default: aiclient.ProviderConfig{
			ProviderRef: "stub",
			Model:       "stub-chat-1",
		},
		Route:     route,
		TimeoutMs: 5000,
		Version:   "1.0.0",
	}, nil
}

type scenarioPracticeAIClient struct {
	store           *scenarioPracticeStore
	failures        []error
	hintFailures    []error
	rawContent      string
	responseText    string
	responseIntent  string
	transcription   string
	synthesisAudio  []byte
	synthesisMillis int
	sttFailures     []error
	ttsFailures     []error
	calls           int
}

const hintFeatureKeyForScenario = "practice.turn.lightweight_observe"

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
	if payload.Metadata.FeatureKey == hintFeatureKeyForScenario && len(c.hintFailures) > 0 {
		err := c.hintFailures[0]
		c.hintFailures = c.hintFailures[1:]
		return aiclient.CompleteResponse{}, aiclient.AICallMeta{}, err
	}
	if len(c.failures) > 0 {
		err := c.failures[0]
		c.failures = c.failures[1:]
		return aiclient.CompleteResponse{}, aiclient.AICallMeta{}, err
	}
	if profileName != "practice.first_question.default" && profileName != "practice.followup.default" && profileName != "practice.turn_observe.default" {
		return aiclient.CompleteResponse{}, aiclient.AICallMeta{}, fmt.Errorf("unexpected profile %q", profileName)
	}
	if payload.Metadata.FeatureKey != "practice.session.first_question" && payload.Metadata.FeatureKey != "practice.session.follow_up" && payload.Metadata.FeatureKey != hintFeatureKeyForScenario {
		return aiclient.CompleteResponse{}, aiclient.AICallMeta{}, fmt.Errorf("unexpected AI feature key: %+v", payload.Metadata)
	}
	if payload.Metadata.PromptVersion == "" ||
		payload.Metadata.RubricVersion == "" ||
		payload.Metadata.Language == "" ||
		payload.Metadata.FeatureFlag == "" ||
		payload.Metadata.DataSourceVersion == "" {
		return aiclient.CompleteResponse{}, aiclient.AICallMeta{}, fmt.Errorf("incomplete AI metadata: %+v", payload.Metadata)
	}
	if payload.Metadata.TaskRun.Capability != aiclient.AITaskRunTaskQuestionGenerate &&
		payload.Metadata.TaskRun.Capability != aiclient.AITaskRunTaskFollowupGenerate &&
		payload.Metadata.TaskRun.Capability != aiclient.AITaskRunTaskHintGenerate {
		return aiclient.CompleteResponse{}, aiclient.AICallMeta{}, fmt.Errorf("incomplete AI task run context: %+v", payload.Metadata.TaskRun)
	}
	if payload.Metadata.TaskRun.ResourceType != aiclient.AITaskRunResourceTargetJob ||
		payload.Metadata.TaskRun.ResourceID == "" {
		return aiclient.CompleteResponse{}, aiclient.AICallMeta{}, fmt.Errorf("incomplete AI task run context: %+v", payload.Metadata.TaskRun)
	}
	if payload.Metadata.FeatureKey == hintFeatureKeyForScenario {
		if c.rawContent != "" {
			return aiclient.CompleteResponse{Content: c.rawContent}, aiclient.AICallMeta{
				Provider:         "stub",
				ModelFamily:      "stub",
				ModelID:          "stub-chat-1",
				FallbackChain:    []string{"stub/stub-chat-1"},
				ValidationStatus: aiclient.ValidationStatusOK,
			}, nil
		}
		hint := c.responseText
		if hint == "" {
			hint = "Anchor the answer in one measurable coordination decision."
		}
		content, err := json.Marshal(map[string]string{"hint": hint})
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
	if c.store != nil {
		c.store.recordAIObservation()
	}
	if len(c.sttFailures) > 0 {
		err := c.sttFailures[0]
		c.sttFailures = c.sttFailures[1:]
		return aiclient.TranscriptionResponse{}, aiclient.AICallMeta{}, err
	}
	if profileName != "practice.voice.stt.default" {
		return aiclient.TranscriptionResponse{}, aiclient.AICallMeta{}, fmt.Errorf("unexpected STT profile %q", profileName)
	}
	if input.Metadata.FeatureKey != "practice.voice.stt" ||
		input.Metadata.PromptVersion == "" ||
		input.Metadata.RubricVersion == "" ||
		input.Metadata.Language == "" ||
		input.Metadata.FeatureFlag == "" ||
		input.Metadata.DataSourceVersion == "" {
		return aiclient.TranscriptionResponse{}, aiclient.AICallMeta{}, fmt.Errorf("incomplete STT metadata: %+v", input.Metadata)
	}
	if input.Metadata.TaskRun.Capability != aiclient.AITaskRunTaskFollowupGenerate ||
		input.Metadata.TaskRun.ResourceType != aiclient.AITaskRunResourceTargetJob ||
		input.Metadata.TaskRun.ResourceID == "" {
		return aiclient.TranscriptionResponse{}, aiclient.AICallMeta{}, fmt.Errorf("incomplete STT task run context: %+v", input.Metadata.TaskRun)
	}
	if len(input.Audio) == 0 || input.ContentType == "" {
		return aiclient.TranscriptionResponse{}, aiclient.AICallMeta{}, fmt.Errorf("incomplete STT audio input")
	}
	transcription := c.transcription
	if transcription == "" {
		transcription = "我先按风险分组，再为关键团队安排试点迁移。"
	}
	return aiclient.TranscriptionResponse{Text: transcription}, aiclient.AICallMeta{
		Provider:         "fixture-stt",
		ModelFamily:      "fixture-stt",
		ModelID:          "fixture-stt-1",
		ModelProfileName: profileName,
		FallbackChain:    []string{"fixture-stt/fixture-stt-1"},
		ValidationStatus: aiclient.ValidationStatusOK,
		LatencyMs:        120,
	}, nil
}

func (c *scenarioPracticeAIClient) Stream(ctx context.Context, profileName string, payload aiclient.CompletePayload) (<-chan aiclient.AIStreamEvent, error) {
	return nil, nil
}

func (c *scenarioPracticeAIClient) Synthesize(ctx context.Context, profileName string, input aiclient.SynthesisInput) (aiclient.SynthesisResponse, aiclient.AICallMeta, error) {
	if c.store != nil {
		c.store.recordAIObservation()
	}
	if len(c.ttsFailures) > 0 {
		err := c.ttsFailures[0]
		c.ttsFailures = c.ttsFailures[1:]
		return aiclient.SynthesisResponse{}, aiclient.AICallMeta{}, err
	}
	if profileName != "practice.voice.tts.default" {
		return aiclient.SynthesisResponse{}, aiclient.AICallMeta{}, fmt.Errorf("unexpected TTS profile %q", profileName)
	}
	if input.Metadata.FeatureKey != "practice.voice.tts" ||
		input.Metadata.PromptVersion == "" ||
		input.Metadata.RubricVersion == "" ||
		input.Metadata.Language == "" ||
		input.Metadata.FeatureFlag == "" ||
		input.Metadata.DataSourceVersion == "" {
		return aiclient.SynthesisResponse{}, aiclient.AICallMeta{}, fmt.Errorf("incomplete TTS metadata: %+v", input.Metadata)
	}
	if input.Metadata.TaskRun.Capability != aiclient.AITaskRunTaskFollowupGenerate ||
		input.Metadata.TaskRun.ResourceType != aiclient.AITaskRunResourceTargetJob ||
		input.Metadata.TaskRun.ResourceID == "" {
		return aiclient.SynthesisResponse{}, aiclient.AICallMeta{}, fmt.Errorf("incomplete TTS task run context: %+v", input.Metadata.TaskRun)
	}
	if strings.TrimSpace(input.Text) == "" {
		return aiclient.SynthesisResponse{}, aiclient.AICallMeta{}, fmt.Errorf("empty TTS text")
	}
	audio := c.synthesisAudio
	if len(audio) == 0 {
		audio = []byte("scenario-tts-audio")
	}
	duration := c.synthesisMillis
	if duration == 0 {
		duration = 1600
	}
	return aiclient.SynthesisResponse{
			Audio:       append([]byte{}, audio...),
			ContentType: "audio/mpeg",
			DurationMs:  duration,
			CharCount:   len([]rune(input.Text)),
		}, aiclient.AICallMeta{
			Provider:         "fixture-tts",
			ModelFamily:      "fixture-tts",
			ModelID:          "fixture-tts-1",
			ModelProfileName: profileName,
			FallbackChain:    []string{"fixture-tts/fixture-tts-1"},
			ValidationStatus: aiclient.ValidationStatusOK,
			LatencyMs:        180,
		}, nil
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
	hintTexts                 map[string]map[string]string
	sessionEvents             map[string][]scenarioPracticeSessionEvent
	idempotencyRecords        map[string]scenarioPracticeIdempotencyRecord
	completedReports          map[string]domainpractice.CompleteSessionResult
	outbox                    []scenarioOutboxEvent
	audits                    [][]byte
	prerequisiteTargetOwner   map[string]string
	prerequisiteResumeOwner   map[string]string
	reports                   map[string]scenarioFeedbackReport
	debriefs                  map[string]scenarioDebrief
	inTransaction             bool
	aiCalledInsideTransaction bool
}

type scenarioPracticePlan struct {
	domainpractice.PlanRecord
	UserID        string
	ResumeAssetID string
}

type scenarioFeedbackReport struct {
	UserID      string
	TargetJobID string
	Status      sharedtypes.ReportStatus
}

type scenarioDebrief struct {
	UserID      string
	TargetJobID string
	Status      sharedtypes.DebriefStatus
	Questions   []scenarioDebriefQuestion
}

type scenarioDebriefQuestion struct {
	Text   string
	Intent string
}

type scenarioPracticeSession struct {
	domainpractice.SessionRecord
	UserID      string
	FailureCode string
}

type scenarioPracticeSessionEvent struct {
	ID            string
	SeqNo         int
	EventType     string
	ClientEventID string
	Payload       []byte
	replayPayload []byte
}

type scenarioOutboxEvent struct {
	EventName sharedevents.EventName
	Payload   []byte
}

type scenarioPracticeIdempotencyRecord struct {
	RecordID     string
	UserID       string
	Domain       string
	Operation    string
	KeyHash      string
	Fingerprint  string
	Status       idempotency.Status
	ExpiresAt    time.Time
	Response     []byte
	HTTPStatus   int
	ResourceType string
	ResourceID   string
	ErrorCode    string
}

func newScenarioPracticeStore() *scenarioPracticeStore {
	s := &scenarioPracticeStore{
		plans:                   map[string]scenarioPracticePlan{},
		sessions:                map[string]scenarioPracticeSession{},
		turns:                   map[string][]domainpractice.TurnRecord{},
		hintTexts:               map[string]map[string]string{},
		sessionEvents:           map[string][]scenarioPracticeSessionEvent{},
		idempotencyRecords:      map[string]scenarioPracticeIdempotencyRecord{},
		completedReports:        map[string]domainpractice.CompleteSessionResult{},
		prerequisiteTargetOwner: map[string]string{},
		prerequisiteResumeOwner: map[string]string{},
		reports:                 map[string]scenarioFeedbackReport{},
		debriefs:                map[string]scenarioDebrief{},
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
	if !s.sourceAvailableForPlan(in) {
		return domainpractice.PlanRecord{}, domainpractice.ErrPlanPrerequisiteNotFound
	}
	plan := domainpractice.PlanRecord{
		ID:                 in.PlanID,
		TargetJobID:        in.TargetJobID,
		SourceReportID:     in.SourceReportID,
		SourceDebriefID:    in.SourceDebriefID,
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
	auditMetadata := map[string]any{
		"plan_id":       in.PlanID,
		"goal":          string(in.Goal),
		"mode":          string(in.Mode),
		"language":      in.Language,
		"target_job_id": in.TargetJobID,
	}
	if in.SourceReportID != "" {
		auditMetadata["source_report_id"] = in.SourceReportID
	}
	if in.SourceDebriefID != "" {
		auditMetadata["source_debrief_id"] = in.SourceDebriefID
	}
	audit, err := json.Marshal(auditMetadata)
	if err != nil {
		return domainpractice.PlanRecord{}, err
	}
	s.audits = append(s.audits, audit)
	return plan, nil
}

func (s *scenarioPracticeStore) sourceAvailableForPlan(in domainpractice.CreatePlanStoreInput) bool {
	switch in.Goal {
	case sharedtypes.PracticeGoalBaseline:
		return strings.TrimSpace(in.SourceReportID) == "" && strings.TrimSpace(in.SourceDebriefID) == ""
	case sharedtypes.PracticeGoalRetryCurrentRound, sharedtypes.PracticeGoalNextRound:
		report, ok := s.reports[in.SourceReportID]
		return ok &&
			report.UserID == in.UserID &&
			report.TargetJobID == in.TargetJobID &&
			report.Status == sharedtypes.ReportStatusReady &&
			strings.TrimSpace(in.SourceDebriefID) == ""
	case sharedtypes.PracticeGoalDebrief:
		debrief, ok := s.debriefs[in.SourceDebriefID]
		return ok &&
			debrief.UserID == in.UserID &&
			debrief.TargetJobID == in.TargetJobID &&
			debrief.Status == sharedtypes.DebriefStatusCompleted &&
			len(debrief.Questions) > 0 &&
			strings.TrimSpace(debrief.Questions[0].Text) != "" &&
			strings.TrimSpace(in.SourceReportID) == ""
	default:
		return false
	}
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

func (s *scenarioPracticeStore) ReserveSessionEvent(_ context.Context, in domainpractice.SessionEventReservationInput) (domainpractice.SessionEventReservation, error) {
	session, ok := s.sessions[in.SessionID]
	if !ok || session.UserID != in.UserID {
		return domainpractice.SessionEventReservation{}, domainpractice.ErrSessionNotFound
	}
	plan, ok := s.plans[session.PlanID]
	if !ok || plan.UserID != in.UserID {
		return domainpractice.SessionEventReservation{}, domainpractice.ErrPlanNotFound
	}
	turns := s.turns[in.SessionID]
	if len(turns) == 0 {
		return domainpractice.SessionEventReservation{}, domainpractice.ErrSessionConflict
	}
	s.inTransaction = true
	defer func() { s.inTransaction = false }()
	if replay, replayErr, hit, err := s.sessionEventReplay(in.SessionID, in.ClientEventID, in.RequestFingerprint); err != nil {
		return domainpractice.SessionEventReservation{}, err
	} else if hit {
		if replayErr != nil {
			return domainpractice.SessionEventReservation{ReplayError: replayErr}, nil
		}
		return domainpractice.SessionEventReservation{ReplayResult: &replay}, nil
	}
	return domainpractice.SessionEventReservation{
		UserID:     in.UserID,
		Session:    session.SessionRecord,
		Plan:       plan.PlanRecord,
		LatestTurn: turns[len(turns)-1],
	}, nil
}

func (s *scenarioPracticeStore) FinalizeSessionEventError(_ context.Context, in domainpractice.FinalizeSessionEventErrorInput) error {
	session, ok := s.sessions[in.SessionID]
	if !ok || session.UserID != in.UserID {
		return domainpractice.ErrSessionNotFound
	}
	s.inTransaction = true
	defer func() { s.inTransaction = false }()
	payload, err := json.Marshal(scenarioAppendEventPayload{
		RequestFingerprint: in.RequestFingerprint,
		Error:              in.Error,
	})
	if err != nil {
		return err
	}
	s.sessionEvents[in.SessionID] = append(s.sessionEvents[in.SessionID], scenarioPracticeSessionEvent{
		ID:            in.EventID,
		SeqNo:         len(s.sessionEvents[in.SessionID]) + 1,
		EventType:     in.Kind,
		ClientEventID: in.ClientEventID,
		Payload:       payload,
	})
	return nil
}

func (s *scenarioPracticeStore) AppendSessionEvent(_ context.Context, in domainpractice.AppendSessionEventStoreInput) (domainpractice.AppendSessionEventResult, error) {
	session, ok := s.sessions[in.SessionID]
	if !ok || session.UserID != in.UserID {
		return domainpractice.AppendSessionEventResult{}, domainpractice.ErrSessionNotFound
	}
	s.inTransaction = true
	defer func() { s.inTransaction = false }()
	if replay, replayErr, hit, err := s.sessionEventReplay(in.SessionID, in.ClientEventID, in.RequestFingerprint); err != nil {
		return domainpractice.AppendSessionEventResult{}, err
	} else if hit {
		if replayErr != nil {
			return domainpractice.AppendSessionEventResult{}, replayErr
		}
		replay.Replay = true
		return replay, nil
	}
	turns := s.turns[in.SessionID]
	if len(turns) == 0 {
		return domainpractice.AppendSessionEventResult{}, domainpractice.ErrSessionConflict
	}
	latestTurn := turns[len(turns)-1]
	if in.Outcome.NextTurn != nil {
		turns[len(turns)-1] = *in.Outcome.NextTurn
	}
	if in.Outcome.AssistantAction.Type == "show_hint" && strings.TrimSpace(in.Outcome.AssistantAction.Hint) != "" {
		turnID := strings.TrimSpace(in.Outcome.AssistantAction.TurnID)
		if turnID == "" {
			turnID = latestTurn.ID
		}
		if s.hintTexts[in.SessionID] == nil {
			s.hintTexts[in.SessionID] = map[string]string{}
		}
		s.hintTexts[in.SessionID][turnID] = strings.TrimSpace(in.Outcome.AssistantAction.Hint)
	}
	if in.NextQuestion != nil {
		turns = append(turns, *in.NextQuestion)
		session.CurrentTurn = in.NextQuestion
		session.TurnCount = in.NextQuestion.TurnIndex
	} else if in.Outcome.NextTurn != nil {
		next := *in.Outcome.NextTurn
		session.CurrentTurn = &next
	}
	session.Status = in.Outcome.NextSessionStatus
	session.UpdatedAt = in.OccurredAt
	s.turns[in.SessionID] = turns
	s.sessions[in.SessionID] = session
	if in.Outcome.OutboxRecord != nil {
		payload, err := storepractice.BuildPracticeTurnCompletedPayload(storepractice.PracticeTurnCompletedInput{
			SessionID:        in.SessionID,
			TurnID:           latestTurn.ID,
			TurnIndex:        int(latestTurn.TurnIndex),
			QuestionIntent:   latestTurn.QuestionIntent,
			FollowUpCount:    in.Outcome.OutboxRecord.FollowUpCount,
			AnswerCharLength: in.Outcome.OutboxRecord.AnswerCharLength,
		})
		if err != nil {
			return domainpractice.AppendSessionEventResult{}, err
		}
		raw, err := json.Marshal(payload)
		if err != nil {
			return domainpractice.AppendSessionEventResult{}, err
		}
		s.outbox = append(s.outbox, scenarioOutboxEvent{EventName: sharedevents.EventNamePracticeTurnCompleted, Payload: raw})
	}
	result := domainpractice.AppendSessionEventResult{
		Acknowledged:    in.Outcome.Acknowledged,
		Session:         session.SessionRecord,
		AssistantAction: in.Outcome.AssistantAction,
	}
	storedResult := result
	if storedResult.AssistantAction.Type == "show_hint" {
		storedResult.AssistantAction.Hint = ""
	}
	payload, err := json.Marshal(scenarioAppendEventPayload{
		RequestFingerprint: in.RequestFingerprint,
		Result:             storedResult,
	})
	if err != nil {
		return domainpractice.AppendSessionEventResult{}, err
	}
	replayPayload, err := json.Marshal(scenarioAppendEventPayload{
		RequestFingerprint: in.RequestFingerprint,
		Result:             result,
	})
	if err != nil {
		return domainpractice.AppendSessionEventResult{}, err
	}
	s.sessionEvents[in.SessionID] = append(s.sessionEvents[in.SessionID], scenarioPracticeSessionEvent{
		ID:            in.EventID,
		SeqNo:         len(s.sessionEvents[in.SessionID]) + 1,
		EventType:     in.Kind,
		ClientEventID: in.ClientEventID,
		Payload:       payload,
		replayPayload: replayPayload,
	})
	return result, nil
}

func (s *scenarioPracticeStore) RecordPracticeVoiceTurn(_ context.Context, in domainpractice.PracticeVoiceTurnStoreInput) (domainpractice.SessionRecord, error) {
	session, ok := s.sessions[in.SessionID]
	if !ok || session.UserID != in.UserID {
		return domainpractice.SessionRecord{}, domainpractice.ErrSessionNotFound
	}
	s.inTransaction = true
	defer func() { s.inTransaction = false }()
	turns := s.turns[in.SessionID]
	if len(turns) == 0 || strings.TrimSpace(in.TurnID) != turns[len(turns)-1].ID {
		return domainpractice.SessionRecord{}, domainpractice.ErrSessionConflict
	}
	latest := turns[len(turns)-1]
	latest.Status = string(domainpractice.TurnStatusFollowUpRequested)
	latest.FollowUpCount = 1
	turns[len(turns)-1] = latest
	session.Status = sharedtypes.SessionStatusRunning
	session.UpdatedAt = in.OccurredAt
	session.CurrentTurn = &latest
	s.turns[in.SessionID] = turns
	s.sessions[in.SessionID] = session
	payload, err := json.Marshal(map[string]any{
		"voiceTurnId":         in.VoiceTurnID,
		"turnId":              in.TurnID,
		"userTranscriptFinal": in.UserTranscriptFinal,
		"assistantTextDraft":  in.AssistantTextDraft,
		"audioByteLength":     in.AudioByteLength,
		"audioDurationMs":     in.AudioDurationMs,
		"ttsChunks":           in.TTSChunks,
		"ttsError":            in.TTSError,
		"providerMetaSummary": in.ProviderMetaSummary,
	})
	if err != nil {
		return domainpractice.SessionRecord{}, err
	}
	s.sessionEvents[in.SessionID] = append(s.sessionEvents[in.SessionID], scenarioPracticeSessionEvent{
		ID:            in.EventID,
		SeqNo:         len(s.sessionEvents[in.SessionID]) + 1,
		EventType:     "follow_up_generated",
		ClientEventID: in.ClientVoiceTurnID,
		Payload:       payload,
	})
	return session.SessionRecord, nil
}

func (s *scenarioPracticeStore) CompleteSession(_ context.Context, in domainpractice.CompleteSessionStoreInput) (domainpractice.CompleteSessionResult, error) {
	session, ok := s.sessions[in.SessionID]
	if !ok || session.UserID != in.UserID {
		return domainpractice.CompleteSessionResult{}, domainpractice.ErrSessionNotFound
	}
	s.inTransaction = true
	defer func() { s.inTransaction = false }()
	if existing, ok := s.completedReports[in.SessionID]; ok {
		existing.Replay = true
		return existing, nil
	}
	if !canCompleteScenarioSessionStatus(session.Status) {
		return domainpractice.CompleteSessionResult{}, domainpractice.ErrSessionConflict
	}
	session.Status = sharedtypes.SessionStatusCompleting
	session.UpdatedAt = in.Now
	s.sessions[in.SessionID] = session
	eventPayload, err := json.Marshal(map[string]any{"sessionId": in.SessionID, "clientCompletedAt": in.ClientCompletedAt.UTC().Format(time.RFC3339)})
	if err != nil {
		return domainpractice.CompleteSessionResult{}, err
	}
	s.sessionEvents[in.SessionID] = append(s.sessionEvents[in.SessionID], scenarioPracticeSessionEvent{
		ID:        in.SessionEventID,
		SeqNo:     len(s.sessionEvents[in.SessionID]) + 1,
		EventType: "session_completed",
		Payload:   eventPayload,
	})
	outboxPayload, err := storepractice.BuildPracticeSessionCompletedPayload(storepractice.PracticeSessionCompletedInput{
		Language:    session.Language,
		PlanID:      session.PlanID,
		SessionID:   session.ID,
		TargetJobID: session.TargetJobID,
		TurnCount:   int(session.TurnCount),
	})
	if err != nil {
		return domainpractice.CompleteSessionResult{}, err
	}
	outboxRaw, err := json.Marshal(outboxPayload)
	if err != nil {
		return domainpractice.CompleteSessionResult{}, err
	}
	s.outbox = append(s.outbox, scenarioOutboxEvent{EventName: sharedevents.EventNamePracticeSessionCompleted, Payload: outboxRaw})
	audit, err := json.Marshal(map[string]any{
		"session_id":    in.SessionID,
		"report_id":     in.ReportID,
		"job_id":        in.JobID,
		"target_job_id": session.TargetJobID,
		"turn_count":    session.TurnCount,
	})
	if err != nil {
		return domainpractice.CompleteSessionResult{}, err
	}
	s.audits = append(s.audits, audit)
	result := domainpractice.CompleteSessionResult{
		ReportID: in.ReportID,
		Job: domainpractice.JobRecord{
			ID:           in.JobID,
			JobType:      api.JobTypeReportGenerate,
			ResourceType: api.ResourceTypeFeedbackReport,
			ResourceID:   in.ReportID,
			Status:       sharedtypes.JobStatusQueued,
			CreatedAt:    in.Now,
			UpdatedAt:    in.Now,
		},
	}
	s.completedReports[in.SessionID] = result
	return result, nil
}

func canCompleteScenarioSessionStatus(status sharedtypes.SessionStatus) bool {
	switch status {
	case sharedtypes.SessionStatusRunning, sharedtypes.SessionStatusWaitingUserInput, sharedtypes.SessionStatusCompleted:
		return true
	default:
		return false
	}
}

func (s *scenarioPracticeStore) ReserveSessionStart(_ context.Context, in domainpractice.StartSessionReservationInput) (domainpractice.SessionReservation, error) {
	plan, ok := s.plans[in.PlanID]
	if !ok || plan.UserID != in.UserID || plan.Status != "ready" {
		return domainpractice.SessionReservation{}, domainpractice.ErrPlanNotFound
	}
	var debriefFirstQuestion scenarioDebriefQuestion
	if plan.Goal == sharedtypes.PracticeGoalDebrief {
		debrief, ok := s.debriefs[plan.SourceDebriefID]
		if !ok ||
			debrief.UserID != in.UserID ||
			debrief.TargetJobID != plan.TargetJobID ||
			debrief.Status != sharedtypes.DebriefStatusCompleted ||
			len(debrief.Questions) == 0 ||
			strings.TrimSpace(debrief.Questions[0].Text) == "" {
			return domainpractice.SessionReservation{}, domainpractice.ErrPlanNotFound
		}
		debriefFirstQuestion = debrief.Questions[0]
	}
	s.inTransaction = true
	defer func() { s.inTransaction = false }()

	key := s.idempotencyRecordKey(in.UserID, "practice", "startPracticeSession", in.IdempotencyKeyHash)
	recordID := in.IdempotencyRecordID
	if existing, ok := s.idempotencyRecords[key]; ok {
		if !in.Now.Before(existing.ExpiresAt) {
			recordID = existing.RecordID
			existing.Fingerprint = in.RequestFingerprint
			existing.Status = idempotency.StatusPending
			existing.ErrorCode = ""
			existing.ResourceID = ""
			existing.Response = nil
			existing.HTTPStatus = 0
			existing.ExpiresAt = in.ExpiresAt
			s.idempotencyRecords[key] = existing
		} else {
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
		IdempotencyRecordID:        recordID,
		SessionID:                  in.SessionID,
		UserID:                     in.UserID,
		PlanID:                     in.PlanID,
		TargetJobID:                plan.TargetJobID,
		Goal:                       plan.Goal,
		Mode:                       plan.Mode,
		InterviewerPersona:         plan.InterviewerPersona,
		Language:                   plan.Language,
		HintsEnabled:               in.HintsEnabled,
		DebriefFirstQuestionText:   strings.TrimSpace(debriefFirstQuestion.Text),
		DebriefFirstQuestionIntent: strings.TrimSpace(debriefFirstQuestion.Intent),
		CreatedAt:                  in.Now,
		UpdatedAt:                  in.Now,
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

type scenarioAppendEventPayload struct {
	RequestFingerprint string                                  `json:"requestFingerprint"`
	Result             domainpractice.AppendSessionEventResult `json:"result"`
	Error              *domainpractice.ServiceError            `json:"error,omitempty"`
}

func (s *scenarioPracticeStore) sessionEventReplay(sessionID, clientEventID, fingerprint string) (domainpractice.AppendSessionEventResult, *domainpractice.ServiceError, bool, error) {
	if strings.TrimSpace(clientEventID) == "" {
		return domainpractice.AppendSessionEventResult{}, nil, false, nil
	}
	for _, event := range s.sessionEvents[sessionID] {
		if event.ClientEventID != clientEventID {
			continue
		}
		raw := event.Payload
		if len(event.replayPayload) > 0 {
			raw = event.replayPayload
		}
		var stored scenarioAppendEventPayload
		if err := json.Unmarshal(raw, &stored); err != nil {
			return domainpractice.AppendSessionEventResult{}, nil, true, err
		}
		if stored.RequestFingerprint != fingerprint {
			return domainpractice.AppendSessionEventResult{}, nil, true, domainpractice.ErrClientEventMismatch
		}
		if stored.Error != nil {
			return domainpractice.AppendSessionEventResult{}, stored.Error, true, nil
		}
		stored.Result.Replay = true
		return stored.Result, nil, true, nil
	}
	return domainpractice.AppendSessionEventResult{}, nil, false, nil
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
	if rec.Status == idempotency.StatusFailedTerminal {
		rec.Status = idempotency.StatusPending
		rec.ExpiresAt = in.ExpiresAt
		rec.Fingerprint = in.RequestFingerprint
		rec.Response = nil
		rec.HTTPStatus = 0
		rec.ResourceType = ""
		rec.ResourceID = ""
		rec.ErrorCode = ""
		s.idempotencyRecords[key] = rec
		return idempotency.Reservation{State: idempotency.StateExecute, RecordID: rec.RecordID}, nil
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
			rec.ResourceType = in.ResourceType
			rec.ResourceID = in.ResourceID
			s.idempotencyRecords[key] = rec
			return nil
		}
	}
	return idempotency.ErrReservationNotFound
}

func (s *scenarioPracticeStore) MarkFailed(_ context.Context, in idempotency.CompletionInput) error {
	for key, rec := range s.idempotencyRecords {
		if rec.RecordID == in.RecordID && rec.UserID == in.UserID && rec.Domain == in.Domain && rec.Operation == in.Operation {
			rec.Status = idempotency.StatusFailedTerminal
			rec.Response = append([]byte{}, in.ResponseBody...)
			rec.HTTPStatus = in.ResponseStatus
			rec.ResourceType = ""
			rec.ResourceID = ""
			rec.ErrorCode = scenarioErrorCodeFromResponseBody(in.ResponseBody)
			s.idempotencyRecords[key] = rec
			return nil
		}
	}
	return idempotency.ErrReservationNotFound
}

func scenarioErrorCodeFromResponseBody(raw []byte) string {
	var decoded struct {
		Error *struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(raw, &decoded); err != nil || decoded.Error == nil {
		return ""
	}
	return decoded.Error.Code
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

func (s *scenarioPracticeStore) sessionEventPayloads(sessionID string) [][]byte {
	out := make([][]byte, 0, len(s.sessionEvents[sessionID]))
	for _, event := range s.sessionEvents[sessionID] {
		out = append(out, append([]byte{}, event.Payload...))
	}
	return out
}

func (s *scenarioPracticeStore) pendingSessionEventCount(sessionID string) int {
	count := 0
	for _, event := range s.sessionEvents[sessionID] {
		if strings.Contains(string(event.Payload), `"pending":true`) {
			count++
		}
	}
	return count
}

func (s *scenarioPracticeStore) hintTextForTurn(sessionID, turnID string) string {
	if s.hintTexts[sessionID] == nil {
		return ""
	}
	return s.hintTexts[sessionID][turnID]
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

func (s *scenarioPracticeStore) forceSessionStatus(sessionID string, status sharedtypes.SessionStatus) {
	session, ok := s.sessions[sessionID]
	if !ok {
		return
	}
	session.Status = status
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
	rec, ok := s.idempotencyRecords[s.idempotencyRecordKey(userID, domain, operation, idempotency.HashKey(rawKey, scenarioIdempotencyPepper))]
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
