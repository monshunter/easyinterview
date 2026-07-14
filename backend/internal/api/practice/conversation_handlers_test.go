package practice

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	api "github.com/monshunter/easyinterview/backend/internal/api/generated"
	"github.com/monshunter/easyinterview/backend/internal/middleware/idempotency"
	domain "github.com/monshunter/easyinterview/backend/internal/practice"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

type conversationHandlerService struct {
	planService
	createInput    domain.CreatePlanRequest
	createCalls    int
	messageInput   domain.SendPracticeMessageRequest
	messageErr     error
	completeErr    error
	completeResult domain.CompleteSessionResult
	completeInput  domain.CompletePracticeSessionRequest
}

func (s *conversationHandlerService) CreatePracticePlan(_ context.Context, in domain.CreatePlanRequest) (domain.PlanRecord, error) {
	s.createCalls++
	s.createInput = in
	return domain.PlanRecord{ID: "plan-1", TargetJobID: in.TargetJobID, ResumeID: in.ResumeID, SourceReportID: in.SourceReportID, Goal: in.Goal,
		InterviewerPersona: in.InterviewerPersona, Difficulty: in.Difficulty, Language: in.Language,
		TimeBudgetMinutes: in.TimeBudgetMinutes, RoundID: in.RoundID, RoundSequence: 2,
		Status: "ready", CreatedAt: time.Unix(1, 0).UTC()}, nil
}
func (s *conversationHandlerService) SendPracticeMessage(_ context.Context, in domain.SendPracticeMessageRequest) (domain.SendPracticeMessageResult, error) {
	s.messageInput = in
	if s.messageErr != nil {
		return domain.SendPracticeMessageResult{}, s.messageErr
	}
	now := time.Unix(2, 0).UTC()
	user := domain.MessageRecord{
		ID: "m-user", Role: "user", Content: in.Text, SeqNo: 2,
		ClientMessageID: in.ClientMessageID, ReplyStatus: domain.PracticeReplyStatusComplete,
		CreatedAt: now,
	}
	assistant := domain.MessageRecord{ID: "m-assistant", Role: "assistant", Content: "继续说说你的取舍。", SeqNo: 3, CreatedAt: now}
	return domain.SendPracticeMessageResult{Acknowledged: true, UserMessage: user, AssistantMessage: assistant,
		Session: domain.SessionRecord{ID: in.SessionID, PlanID: "plan-1", TargetJobID: "target-1", Status: sharedtypes.SessionStatusRunning,
			Language: "zh-CN", Messages: []domain.MessageRecord{user, assistant}, CreatedAt: now, UpdatedAt: now}}, nil
}

func (s *conversationHandlerService) CompletePracticeSession(_ context.Context, in domain.CompletePracticeSessionRequest) (domain.CompleteSessionResult, error) {
	s.completeInput = in
	if s.completeErr != nil {
		return domain.CompleteSessionResult{}, s.completeErr
	}
	if s.completeResult.ReportID != "" {
		return s.completeResult, nil
	}
	return domain.CompleteSessionResult{ReportID: "report-1"}, nil
}

func TestE2EP0047RejectsZeroAnswerCompletion(t *testing.T) {
	service := &conversationHandlerService{completeErr: &domain.ServiceError{
		Code: "VALIDATION_FAILED", Message: "practice session requires an answered candidate message",
	}}
	raw, _ := json.Marshal(api.CompletePracticeSessionRequest{ClientCompletedAt: time.Unix(10, 0).UTC().Format(time.RFC3339)})
	req := httptest.NewRequest(http.MethodPost, "/practice/sessions/session-1/complete", bytes.NewReader(raw))
	rec := httptest.NewRecorder()
	newConversationHandler(service).CompletePracticeSession(rec, req, "session-1")
	if rec.Code != http.StatusUnprocessableEntity || !strings.Contains(rec.Body.String(), "VALIDATION_FAILED") {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestE2EP0047FreezesReportContext(t *testing.T) {
	service := &conversationHandlerService{completeResult: completionHandlerResult(false)}
	raw, _ := json.Marshal(api.CompletePracticeSessionRequest{ClientCompletedAt: time.Unix(10, 0).UTC().Format(time.RFC3339)})
	req := httptest.NewRequest(http.MethodPost, "/practice/sessions/session-1/complete", bytes.NewReader(raw))
	rec := httptest.NewRecorder()
	newConversationHandler(service).CompletePracticeSession(rec, req, "session-1")
	if rec.Code != http.StatusAccepted || service.completeInput.SessionID != "session-1" {
		t.Fatalf("status=%d input=%+v body=%s", rec.Code, service.completeInput, rec.Body.String())
	}
	for _, internal := range []string{"generationContext", "sourceSnapshot", "complete jd", "complete resume"} {
		if strings.Contains(rec.Body.String(), internal) {
			t.Fatalf("completion handoff leaked frozen report content %q: %s", internal, rec.Body.String())
		}
	}
	var response api.ReportWithJob
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil || response.ReportId != "report-1" || response.Job.Id != "job-1" {
		t.Fatalf("response=%+v err=%v", response, err)
	}
	t.Log("REPORT_CONTEXT_SNAPSHOT_PASS")
}

func TestE2EP0047CompletionReplayPreservesReportContext(t *testing.T) {
	service := &conversationHandlerService{completeResult: completionHandlerResult(true)}
	raw, _ := json.Marshal(api.CompletePracticeSessionRequest{ClientCompletedAt: time.Unix(10, 0).UTC().Format(time.RFC3339)})
	req := httptest.NewRequest(http.MethodPost, "/practice/sessions/session-1/complete", bytes.NewReader(raw))
	rec := httptest.NewRecorder()
	newConversationHandler(service).CompletePracticeSession(rec, req, "session-1")
	if rec.Code != http.StatusAccepted || !strings.Contains(rec.Body.String(), `"reportId":"report-1"`) {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	for _, internal := range []string{"generationContext", "sourceSnapshot", "complete jd", "complete resume"} {
		if strings.Contains(rec.Body.String(), internal) {
			t.Fatalf("completion replay leaked frozen report content %q: %s", internal, rec.Body.String())
		}
	}
	t.Log("REPORT_CONTEXT_REPLAY_PASS")
}

func completionHandlerResult(replay bool) domain.CompleteSessionResult {
	now := time.Unix(10, 0).UTC()
	return domain.CompleteSessionResult{
		ReportID: "report-1", Replay: replay,
		Job: domain.JobRecord{
			ID: "job-1", JobType: api.JobTypeReportGenerate, ResourceType: api.ResourceTypeFeedbackReport,
			ResourceID: "report-1", Status: sharedtypes.JobStatusQueued, CreatedAt: now, UpdatedAt: now,
		},
		GenerationContext: domain.ReportContextSnapshot{
			SchemaVersion: domain.ReportContextSchemaVersion,
			TargetJob:     domain.ReportTargetJobSnapshot{RawJD: "complete jd"},
			Resume:        domain.ReportResumeSnapshot{SourceSnapshot: "complete resume"},
		},
	}
}

func TestSendPracticeMessageMapsConflictAndIsolationErrors(t *testing.T) {
	for _, tc := range []struct {
		name   string
		code   string
		status int
	}{
		{name: "client replay mismatch", code: "IDEMPOTENCY_KEY_MISMATCH", status: http.StatusConflict},
		{name: "cross user session", code: "PRACTICE_SESSION_NOT_FOUND", status: http.StatusNotFound},
	} {
		t.Run(tc.name, func(t *testing.T) {
			service := &conversationHandlerService{messageErr: &domain.ServiceError{Code: tc.code, Message: "bounded error"}}
			raw, _ := json.Marshal(api.SendPracticeMessageRequest{ClientMessageId: "01918fa0-0000-7000-8000-000000000001", Text: "继续"})
			req := httptest.NewRequest(http.MethodPost, "/practice/sessions/session-1/messages", bytes.NewReader(raw))
			rec := httptest.NewRecorder()
			newConversationHandler(service).SendPracticeMessage(rec, req, "session-1")
			if rec.Code != tc.status || !strings.Contains(rec.Body.String(), tc.code) {
				t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
			}
		})
	}
}

func newConversationHandler(service planService) *Handler {
	return NewHandler(HandlerOptions{Service: service, Session: func(context.Context) (string, bool) { return "user-1", true }})
}

func TestCreatePracticePlanMapsOnlyCurrentFields(t *testing.T) {
	service := &conversationHandlerService{}
	roundID := "round-2-technical"
	body := api.CreatePracticePlanRequest{
		TargetJobId: testPointer("target-1"), ResumeId: testPointer("resume-1"), Goal: sharedtypes.PracticeGoalBaseline,
		RoundId: &roundID, InterviewerPersona: testPointer(sharedtypes.InterviewerRoleHiringManager), Difficulty: testPointer("standard"),
		Language: testPointer("zh-CN"), TimeBudgetMinutes: testPointer[int32](30),
	}
	raw, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/practice/plans", bytes.NewReader(raw))
	rec := httptest.NewRecorder()
	newConversationHandler(service).CreatePracticePlan(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if service.createInput.TimeBudgetMinutes != 30 || service.createInput.ResumeID != "resume-1" || service.createInput.RoundID != roundID {
		t.Fatalf("unexpected input: %+v", service.createInput)
	}
	var plan api.PracticePlan
	if err := json.Unmarshal(rec.Body.Bytes(), &plan); err != nil {
		t.Fatal(err)
	}
	if plan.RoundId == nil || *plan.RoundId != roundID || plan.RoundSequence == nil || *plan.RoundSequence != 2 {
		t.Fatalf("round identity missing from response: %+v", plan)
	}
	for _, stale := range []string{"mode", "questionBudget", "hintsEnabled"} {
		if strings.Contains(rec.Body.String(), stale) {
			t.Fatalf("response contains stale field %s: %s", stale, rec.Body.String())
		}
	}
}

func TestCreatePracticePlanDerivedRequestIsClosed(t *testing.T) {
	t.Run("goal and source report only", func(t *testing.T) {
		service := &conversationHandlerService{}
		body := api.CreatePracticePlanRequest{
			Goal: sharedtypes.PracticeGoalRetryCurrentRound, SourceReportId: testPointer("01918fa0-0000-7000-8000-000000000001"),
		}
		raw, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/practice/plans", bytes.NewReader(raw))
		rec := httptest.NewRecorder()
		newConversationHandler(service).CreatePracticePlan(rec, req)
		if rec.Code != http.StatusCreated {
			t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
		}
		if service.createCalls != 1 || service.createInput.SourceReportID != *body.SourceReportId ||
			service.createInput.TargetJobID != "" || service.createInput.ResumeID != "" || service.createInput.RoundID != "" ||
			service.createInput.InterviewerPersona != "" || service.createInput.Difficulty != "" || service.createInput.Language != "" ||
			service.createInput.TimeBudgetMinutes != 0 {
			t.Fatalf("derived request copied client authority: calls=%d input=%+v", service.createCalls, service.createInput)
		}
	})

	for _, tc := range []struct {
		name string
		body string
	}{
		{name: "copied target", body: `{"goal":"retry_current_round","sourceReportId":"01918fa0-0000-7000-8000-000000000001","targetJobId":"01918fa0-0000-7000-8000-000000000002"}`},
		{name: "copied settings", body: `{"goal":"next_round","sourceReportId":"01918fa0-0000-7000-8000-000000000001","difficulty":"stretch"}`},
		{name: "unknown focus", body: `{"goal":"retry_current_round","sourceReportId":"01918fa0-0000-7000-8000-000000000001","focusDimensionCodes":["system_design"]}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			service := &conversationHandlerService{}
			req := httptest.NewRequest(http.MethodPost, "/practice/plans", strings.NewReader(tc.body))
			rec := httptest.NewRecorder()
			newConversationHandler(service).CreatePracticePlan(rec, req)
			if rec.Code < http.StatusBadRequest || rec.Code >= http.StatusInternalServerError || service.createCalls != 0 {
				t.Fatalf("status=%d calls=%d body=%s", rec.Code, service.createCalls, rec.Body.String())
			}
		})
	}
}

type derivedPlanIdempotencyStore struct {
	reserveCalls int
	fingerprints []string
	keyHashes    []string
}

func (s *derivedPlanIdempotencyStore) Reserve(_ context.Context, in idempotency.ReservationInput) (idempotency.Reservation, error) {
	s.reserveCalls++
	s.fingerprints = append(s.fingerprints, in.RequestFingerprint)
	s.keyHashes = append(s.keyHashes, in.IdempotencyKeyHash)
	if s.reserveCalls == 1 {
		return idempotency.Reservation{State: idempotency.StateExecute, RecordID: in.RecordID}, nil
	}
	return idempotency.Reservation{}, idempotency.ErrFingerprintMismatch
}

func (*derivedPlanIdempotencyStore) MarkSucceeded(context.Context, idempotency.CompletionInput) error {
	return nil
}

func (*derivedPlanIdempotencyStore) MarkFailed(context.Context, idempotency.CompletionInput) error {
	return nil
}

func TestCreateDerivedPracticePlanIdempotencyMismatchHasNoSecondInsertOrLeak(t *testing.T) {
	service := &conversationHandlerService{}
	store := &derivedPlanIdempotencyStore{}
	handler := idempotency.New(idempotency.MiddlewareOptions{
		Store: store, KeyPepper: "test-pepper", NewID: func() string { return "idem-1" },
		Now: func() time.Time { return time.Unix(1, 0).UTC() },
	}).Handler("practice", "createPracticePlan", func(*http.Request) (string, bool) { return "user-1", true }, http.HandlerFunc(newConversationHandler(service).CreatePracticePlan))

	request := func(sourceReportID string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/practice/plans", strings.NewReader(`{"goal":"retry_current_round","sourceReportId":"`+sourceReportID+`"}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set(idempotency.HeaderName, "same-key")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		return rec
	}

	first := request("01918fa0-0000-7000-8000-000000000001")
	second := request("01918fa0-0000-7000-8000-000000000002")
	if first.Code != http.StatusCreated || second.Code != http.StatusConflict || service.createCalls != 1 {
		t.Fatalf("first=%d second=%d createCalls=%d secondBody=%s", first.Code, second.Code, service.createCalls, second.Body.String())
	}
	if !strings.Contains(second.Body.String(), "IDEMPOTENCY_KEY_MISMATCH") || strings.Contains(second.Body.String(), "01918fa0") {
		t.Fatalf("mismatch response missing code or leaked source: %s", second.Body.String())
	}
	if len(store.fingerprints) != 2 || store.fingerprints[0] == store.fingerprints[1] || store.keyHashes[0] != store.keyHashes[1] {
		t.Fatalf("idempotency evidence fingerprints=%v keyHashes=%v", store.fingerprints, store.keyHashes)
	}
	t.Log("REPORT_DERIVED_IDEMPOTENCY_PASS")
}

func TestToAPIPracticePlanOmitsPartialLegacyRoundIdentity(t *testing.T) {
	plan := toAPIPracticePlan(domain.PlanRecord{RoundID: "round-1-hr", RoundSequence: 0})
	if plan.RoundId != nil || plan.RoundSequence != nil {
		t.Fatalf("partial legacy identity must be omitted as a pair: %+v", plan)
	}
	plan = toAPIPracticePlan(domain.PlanRecord{RoundSequence: 1})
	if plan.RoundId != nil || plan.RoundSequence != nil {
		t.Fatalf("partial legacy identity must be omitted as a pair: %+v", plan)
	}
}

func TestSendPracticeMessageReturnsConversationMessages(t *testing.T) {
	service := &conversationHandlerService{}
	raw, _ := json.Marshal(api.SendPracticeMessageRequest{ClientMessageId: "01918fa0-0000-7000-8000-000000000001", Text: "我负责了灰度发布。"})
	req := httptest.NewRequest(http.MethodPost, "/practice/sessions/session-1/messages", bytes.NewReader(raw))
	rec := httptest.NewRecorder()
	newConversationHandler(service).SendPracticeMessage(rec, req, "session-1")
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	var result api.SendPracticeMessageResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if !result.Acknowledged || result.UserMessage.Role != "user" || result.AssistantMessage.Role != "assistant" {
		t.Fatalf("unexpected result: %+v", result)
	}
	if result.UserMessage.ClientMessageId != "01918fa0-0000-7000-8000-000000000001" || result.UserMessage.ReplyStatus != api.PracticeReplyStatusComplete {
		t.Fatalf("user recovery projection missing: %+v", result.UserMessage)
	}
	var rawResponse map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &rawResponse); err != nil {
		t.Fatal(err)
	}
	assistant, _ := rawResponse["assistantMessage"].(map[string]any)
	if _, found := assistant["clientMessageId"]; found {
		t.Fatalf("assistant leaked clientMessageId: %s", rec.Body.String())
	}
	if _, found := assistant["replyStatus"]; found {
		t.Fatalf("assistant leaked replyStatus: %s", rec.Body.String())
	}
	for _, internal := range []string{"replyGeneration", "replyLeaseExpiresAt", "reply_generation", "reply_lease_expires_at"} {
		if strings.Contains(rec.Body.String(), internal) {
			t.Fatalf("response leaked internal reply field %s: %s", internal, rec.Body.String())
		}
	}
}

func TestCreatePracticeVoiceTurnFailsClosed(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/practice/sessions/session-1/voice-turns", strings.NewReader(`{}`))
	rec := httptest.NewRecorder()
	newConversationHandler(&conversationHandlerService{}).CreatePracticeVoiceTurn(rec, req, "session-1")
	if rec.Code != http.StatusUnprocessableEntity || !strings.Contains(rec.Body.String(), "AI_UNSUPPORTED_CAPABILITY") {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func testPointer[T any](value T) *T { return &value }
