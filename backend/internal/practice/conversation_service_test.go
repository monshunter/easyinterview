package practice

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient"
	"github.com/monshunter/easyinterview/backend/internal/ai/registry"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

type conversationTestStore struct {
	Store
	planInput          CreatePlanStoreInput
	reservation        SessionReservation
	startInput         CommitSessionStartInput
	messageReservation PracticeMessageReservation
	messageInput       CommitPracticeMessageInput
	failedStart        FailSessionStartInput
}

func (s *conversationTestStore) CreatePlan(_ context.Context, in CreatePlanStoreInput) (PlanRecord, error) {
	s.planInput = in
	return PlanRecord{ID: in.PlanID, TargetJobID: in.TargetJobID, ResumeID: in.ResumeID, Goal: in.Goal,
		InterviewerPersona: in.InterviewerPersona, Difficulty: in.Difficulty, Language: in.Language,
		TimeBudgetMinutes: in.TimeBudgetMinutes, Status: "ready", CreatedAt: in.Now}, nil
}
func (s *conversationTestStore) ReserveSessionStart(_ context.Context, _ StartSessionReservationInput) (SessionReservation, error) {
	return s.reservation, nil
}
func (s *conversationTestStore) CommitSessionStart(_ context.Context, in CommitSessionStartInput) (SessionRecord, error) {
	s.startInput = in
	return SessionRecord{ID: in.SessionID, PlanID: in.PlanID, TargetJobID: in.TargetJobID, Status: sharedtypes.SessionStatusRunning,
		Language: in.Language, Messages: []MessageRecord{{ID: in.MessageID, Role: "assistant", Content: in.MessageText, SeqNo: 1, CreatedAt: in.StartedAt}}}, nil
}
func (s *conversationTestStore) FailSessionStart(_ context.Context, in FailSessionStartInput) error {
	s.failedStart = in
	return nil
}
func (s *conversationTestStore) ReservePracticeMessage(_ context.Context, _ ReservePracticeMessageInput) (PracticeMessageReservation, error) {
	return s.messageReservation, nil
}
func (s *conversationTestStore) CommitPracticeMessage(_ context.Context, in CommitPracticeMessageInput) (SendPracticeMessageResult, error) {
	s.messageInput = in
	return SendPracticeMessageResult{Acknowledged: true, UserMessage: s.messageReservation.UserMessage,
		AssistantMessage: MessageRecord{ID: in.AssistantMessageID, Role: "assistant", Content: in.AssistantText, SeqNo: 3, CreatedAt: in.Now}}, nil
}

type conversationTestRegistry struct{}

func (conversationTestRegistry) ResolveActive(context.Context, string, string) (registry.PromptResolution, error) {
	return registry.PromptResolution{FeatureKey: practiceChatFeatureKey, PromptVersion: "v0.1.0", RubricVersion: "v0.1.0",
		ModelProfileName: "practice.chat.default", UserMessageTemplate: "{{language}} {{target_job_context}} {{resume_context}} {{interview_round}} {{practice_goal}} {{focus_competencies}} {{conversation_history}}"}, nil
}

type conversationTestAI struct {
	aiclient.AIClient
	payloads  []aiclient.CompletePayload
	responses []string
	errs      []error
}

func (a *conversationTestAI) Complete(_ context.Context, _ string, payload aiclient.CompletePayload) (aiclient.CompleteResponse, aiclient.AICallMeta, error) {
	a.payloads = append(a.payloads, payload)
	if len(a.errs) > 0 {
		err := a.errs[0]
		a.errs = a.errs[1:]
		if err != nil {
			return aiclient.CompleteResponse{}, aiclient.AICallMeta{}, err
		}
	}
	response := a.responses[0]
	a.responses = a.responses[1:]
	return aiclient.CompleteResponse{Content: response}, aiclient.AICallMeta{}, nil
}

func TestCreateDerivedPracticePlanPassesReportSourceAndCompetencyFocus(t *testing.T) {
	store := &conversationTestStore{}
	service := NewService(ServiceOptions{Store: store, Now: func() time.Time { return time.Unix(1, 0).UTC() }, NewID: func() string { return "id-1" }})
	_, err := service.CreatePracticePlan(context.Background(), CreatePlanRequest{UserID: "user-1", TargetJobID: "target-1", ResumeID: "resume-1", SourceReportID: "report-1",
		Goal: sharedtypes.PracticeGoalRetryCurrentRound, InterviewerPersona: sharedtypes.InterviewerRoleHiringManager,
		Difficulty: "standard", Language: "zh-CN", TimeBudgetMinutes: 30, FocusCompetencyCodes: []string{"technical_depth"}})
	if err != nil {
		t.Fatalf("CreatePracticePlan: %v", err)
	}
	if store.planInput.SourceReportID != "report-1" || !reflect.DeepEqual(store.planInput.FocusCompetencyCodes, []string{"technical_depth"}) {
		t.Fatalf("derived plan input = %+v", store.planInput)
	}
}

func TestDerivedPracticePlanRequiresSourceReport(t *testing.T) {
	service := NewService(ServiceOptions{Store: &conversationTestStore{}})
	_, err := service.CreatePracticePlan(context.Background(), CreatePlanRequest{UserID: "user-1", TargetJobID: "target-1", ResumeID: "resume-1",
		Goal: sharedtypes.PracticeGoalNextRound, InterviewerPersona: sharedtypes.InterviewerRoleHiringManager,
		Difficulty: "standard", Language: "zh-CN", TimeBudgetMinutes: 30})
	var serviceErr *ServiceError
	if !errors.As(err, &serviceErr) || serviceErr.Code != "VALIDATION_FAILED" {
		t.Fatalf("error = %v", err)
	}
}

func TestPracticePlanContractHasNoQuestionModeOrHintFields(t *testing.T) {
	for _, value := range []any{CreatePlanRequest{}, CreatePlanStoreInput{}, PlanRecord{}} {
		typ := reflect.TypeOf(value)
		for _, stale := range []string{"Mode", "QuestionBudget", "HintsEnabled"} {
			if _, ok := typ.FieldByName(stale); ok {
				t.Fatalf("%s retains stale field %s", typ.Name(), stale)
			}
		}
	}
}

func TestCreatePracticePlanPassesOnlyConversationPlanFields(t *testing.T) {
	store := &conversationTestStore{}
	service := NewService(ServiceOptions{Store: store, Now: func() time.Time { return time.Unix(1, 0).UTC() }, NewID: func() string { return "id-1" }})
	_, err := service.CreatePracticePlan(context.Background(), CreatePlanRequest{UserID: "user-1", TargetJobID: "target-1", ResumeID: "resume-1",
		Goal: sharedtypes.PracticeGoalBaseline, InterviewerPersona: sharedtypes.InterviewerRoleHiringManager,
		Difficulty: "standard", Language: "zh-CN", TimeBudgetMinutes: 30})
	if err != nil {
		t.Fatalf("CreatePracticePlan: %v", err)
	}
	if store.planInput.TimeBudgetMinutes != 30 || store.planInput.Language != "zh-CN" {
		t.Fatalf("unexpected store input: %+v", store.planInput)
	}
}

func TestStartPracticeSessionCreatesOpeningAssistantMessage(t *testing.T) {
	store := &conversationTestStore{reservation: SessionReservation{IdempotencyRecordID: "idem-1", SessionID: "session-1", UserID: "user-1",
		PlanID: "plan-1", TargetJobID: "target-1", Goal: sharedtypes.PracticeGoalBaseline,
		InterviewerPersona: sharedtypes.InterviewerRoleHiringManager, Language: "zh-CN", RoleTitle: "后端工程师", ResumeProfile: "分布式系统经验"}}
	ai := &conversationTestAI{responses: []string{`{"messageText":"你好，我们先聊聊你最近负责的系统。"}`}}
	service := NewService(ServiceOptions{Store: store, Registry: conversationTestRegistry{}, AI: ai, NewID: func() string { return "id-1" }})
	result, err := service.StartPracticeSession(context.Background(), StartSessionRequest{UserID: "user-1", PlanID: "plan-1", IdempotencyKeyHash: "hash", RequestFingerprint: "fp"})
	if err != nil {
		t.Fatalf("StartPracticeSession: %v", err)
	}
	if len(result.Messages) != 1 || result.Messages[0].Role != "assistant" {
		t.Fatalf("unexpected session: %+v", result)
	}
	if strings.Contains(ai.payloads[0].Messages[len(ai.payloads[0].Messages)-1].Content, "question") {
		t.Fatalf("prompt must not add question structure")
	}
}

func TestStartPracticeSessionAIErrorFailsReservationWithoutOpeningMessage(t *testing.T) {
	store := &conversationTestStore{reservation: SessionReservation{IdempotencyRecordID: "idem-1", SessionID: "session-1", UserID: "user-1", PlanID: "plan-1", TargetJobID: "target-1", Goal: sharedtypes.PracticeGoalBaseline, InterviewerPersona: sharedtypes.InterviewerRoleHiringManager, Language: "zh-CN"}}
	ai := &conversationTestAI{errs: []error{errors.New("provider timeout"), errors.New("provider timeout")}, responses: []string{`{}`, `{}`}}
	service := NewService(ServiceOptions{Store: store, Registry: conversationTestRegistry{}, AI: ai, NewID: func() string { return "id-1" }})
	_, err := service.StartPracticeSession(context.Background(), StartSessionRequest{UserID: "user-1", PlanID: "plan-1", IdempotencyKeyHash: "hash", RequestFingerprint: "fp"})
	if err == nil || store.failedStart.SessionID != "session-1" || store.startInput.MessageText != "" {
		t.Fatalf("err=%v failed=%+v committed=%+v", err, store.failedStart, store.startInput)
	}
}

func TestSendPracticeMessageUsesOrdinaryConversationHistory(t *testing.T) {
	store := &conversationTestStore{messageReservation: PracticeMessageReservation{
		Session:     SessionReservation{SessionID: "session-1", UserID: "user-1", TargetJobID: "target-1", Language: "zh-CN", Goal: sharedtypes.PracticeGoalBaseline},
		History:     []MessageRecord{{ID: "m1", Role: "assistant", Content: "你好", SeqNo: 1}},
		UserMessage: MessageRecord{ID: "m2", Role: "user", Content: "我想先要一点帮助", SeqNo: 2},
	}}
	ai := &conversationTestAI{responses: []string{`{"messageText":"可以，先从你承担的具体职责说起。"}`}}
	service := NewService(ServiceOptions{Store: store, Registry: conversationTestRegistry{}, AI: ai, NewID: func() string { return "m3" }})
	result, err := service.SendPracticeMessage(context.Background(), SendPracticeMessageRequest{UserID: "user-1", SessionID: "session-1", ClientMessageID: "client-1", Text: "我想先要一点帮助"})
	if err != nil {
		t.Fatalf("SendPracticeMessage: %v", err)
	}
	if !result.Acknowledged || store.messageInput.AssistantText == "" {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestSendPracticeMessageExactReplayReturnsOriginalResultWithoutAICall(t *testing.T) {
	replay := SendPracticeMessageResult{Acknowledged: true, UserMessage: MessageRecord{ID: "m2", Role: "user", Content: "same", SeqNo: 2}, AssistantMessage: MessageRecord{ID: "m3", Role: "assistant", Content: "original", SeqNo: 3}}
	store := &conversationTestStore{messageReservation: PracticeMessageReservation{Replay: &replay}}
	ai := &conversationTestAI{}
	service := NewService(ServiceOptions{Store: store, Registry: conversationTestRegistry{}, AI: ai})
	result, err := service.SendPracticeMessage(context.Background(), SendPracticeMessageRequest{UserID: "user-1", SessionID: "session-1", ClientMessageID: "client-1", Text: "same"})
	if err != nil || result.AssistantMessage.Content != "original" || len(ai.payloads) != 0 {
		t.Fatalf("result=%+v err=%v aiCalls=%d", result, err, len(ai.payloads))
	}
}

func TestSendPracticeMessageMapsClientMismatchAndCrossUserAccess(t *testing.T) {
	for _, tc := range []struct {
		name string
		err  error
		code string
	}{
		{name: "client mismatch", err: ErrClientEventMismatch, code: "PRACTICE_SESSION_CONFLICT"},
		{name: "cross user", err: ErrSessionNotFound, code: "PRACTICE_SESSION_NOT_FOUND"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			store := &conversationTestStoreWithReserveError{conversationTestStore: conversationTestStore{}, err: tc.err}
			service := NewService(ServiceOptions{Store: store})
			_, err := service.SendPracticeMessage(context.Background(), SendPracticeMessageRequest{UserID: "user-2", SessionID: "session-1", ClientMessageID: "client-1", Text: "same"})
			var serviceErr *ServiceError
			if !errors.As(err, &serviceErr) || serviceErr.Code != tc.code {
				t.Fatalf("error=%v want code=%s", err, tc.code)
			}
		})
	}
}

type conversationTestStoreWithReserveError struct {
	conversationTestStore
	err error
}

func (s *conversationTestStoreWithReserveError) ReservePracticeMessage(context.Context, ReservePracticeMessageInput) (PracticeMessageReservation, error) {
	return PracticeMessageReservation{}, s.err
}
