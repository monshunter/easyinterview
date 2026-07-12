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
	domain "github.com/monshunter/easyinterview/backend/internal/practice"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

type conversationHandlerService struct {
	planService
	createInput  domain.CreatePlanRequest
	messageInput domain.SendPracticeMessageRequest
	messageErr   error
}

func (s *conversationHandlerService) CreatePracticePlan(_ context.Context, in domain.CreatePlanRequest) (domain.PlanRecord, error) {
	s.createInput = in
	return domain.PlanRecord{ID: "plan-1", TargetJobID: in.TargetJobID, ResumeID: in.ResumeID, Goal: in.Goal,
		InterviewerPersona: in.InterviewerPersona, Difficulty: in.Difficulty, Language: in.Language,
		TimeBudgetMinutes: in.TimeBudgetMinutes, Status: "ready", CreatedAt: time.Unix(1, 0).UTC()}, nil
}
func (s *conversationHandlerService) SendPracticeMessage(_ context.Context, in domain.SendPracticeMessageRequest) (domain.SendPracticeMessageResult, error) {
	s.messageInput = in
	if s.messageErr != nil {
		return domain.SendPracticeMessageResult{}, s.messageErr
	}
	now := time.Unix(2, 0).UTC()
	user := domain.MessageRecord{ID: "m-user", Role: "user", Content: in.Text, SeqNo: 2, CreatedAt: now}
	assistant := domain.MessageRecord{ID: "m-assistant", Role: "assistant", Content: "继续说说你的取舍。", SeqNo: 3, CreatedAt: now}
	return domain.SendPracticeMessageResult{Acknowledged: true, UserMessage: user, AssistantMessage: assistant,
		Session: domain.SessionRecord{ID: in.SessionID, PlanID: "plan-1", TargetJobID: "target-1", Status: sharedtypes.SessionStatusRunning,
			Language: "zh-CN", Messages: []domain.MessageRecord{user, assistant}, CreatedAt: now, UpdatedAt: now}}, nil
}

func TestSendPracticeMessageMapsConflictAndIsolationErrors(t *testing.T) {
	for _, tc := range []struct {
		name   string
		code   string
		status int
	}{
		{name: "client replay mismatch", code: "PRACTICE_SESSION_CONFLICT", status: http.StatusConflict},
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
	body := api.CreatePracticePlanRequest{TargetJobId: "target-1", ResumeId: "resume-1", Goal: sharedtypes.PracticeGoalBaseline,
		InterviewerPersona: sharedtypes.InterviewerRoleHiringManager, Difficulty: "standard", Language: "zh-CN", TimeBudgetMinutes: 30}
	raw, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/practice/plans", bytes.NewReader(raw))
	rec := httptest.NewRecorder()
	newConversationHandler(service).CreatePracticePlan(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if service.createInput.TimeBudgetMinutes != 30 || service.createInput.ResumeID != "resume-1" {
		t.Fatalf("unexpected input: %+v", service.createInput)
	}
	for _, stale := range []string{"mode", "questionBudget", "hintsEnabled"} {
		if strings.Contains(rec.Body.String(), stale) {
			t.Fatalf("response contains stale field %s: %s", stale, rec.Body.String())
		}
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
}

func TestCreatePracticeVoiceTurnFailsClosed(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/practice/sessions/session-1/voice-turns", strings.NewReader(`{}`))
	rec := httptest.NewRecorder()
	newConversationHandler(&conversationHandlerService{}).CreatePracticeVoiceTurn(rec, req, "session-1")
	if rec.Code != http.StatusUnprocessableEntity || !strings.Contains(rec.Body.String(), "AI_UNSUPPORTED_CAPABILITY") {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}
