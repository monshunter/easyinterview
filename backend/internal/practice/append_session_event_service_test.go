package practice

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/ai/registry"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

func TestAppendSessionEventFollowUpRunsAIOutsideReservationAndCommits(t *testing.T) {
	now := time.Date(2026, 4, 28, 13, 45, 12, 0, time.UTC)
	store := &recordingPlanStore{
		eventReservation: SessionEventReservation{
			UserID:     "user-1",
			Session:    sessionEventTestSession(1),
			Plan:       sessionEventTestPlan(3, sharedtypes.PracticeGoalBaseline),
			LatestTurn: sessionEventTestTurn(1),
		},
	}
	ai := &fakeAIClient{content: firstQuestionJSON(t, "What was the strongest objection?", "behavioral.depth"), store: store}
	service := NewService(ServiceOptions{
		Store: store,
		Registry: &fakePromptResolver{resolution: registry.PromptResolution{
			PromptVersion:       "followup.prompt.v1",
			RubricVersion:       "followup.rubric.v1",
			ModelProfileName:    "practice.follow_up.default",
			FeatureFlag:         "follow_up_v1",
			DataSourceVersion:   "registry.v1",
			UserMessageTemplate: "ask a follow up",
		}},
		AI:    ai,
		Now:   func() time.Time { return now },
		NewID: sequenceIDs("event-1", "outbox-1"),
	})

	result, err := service.AppendSessionEvent(context.Background(), AppendSessionEventRequest{
		UserID:        "user-1",
		SessionID:     "session-1",
		ClientEventID: "client-event-1",
		Kind:          "answer_submitted",
		OccurredAt:    now,
		Payload: map[string]any{
			"turnId":     "turn-1",
			"answerText": "I aligned 12 teams.",
		},
	})
	if err != nil {
		t.Fatalf("AppendSessionEvent returned error: %v", err)
	}
	if !reflect.DeepEqual(store.steps, []string{"reserve-event", "ai", "append-event"}) {
		t.Fatalf("steps = %v", store.steps)
	}
	if !ai.calledOutsideTransaction {
		t.Fatalf("follow-up AI must run outside repository transaction")
	}
	if result.AssistantAction.Type != assistantActionAskFollowUp ||
		result.AssistantAction.QuestionText != "What was the strongest objection?" ||
		result.AssistantAction.Provenance.PromptVersion != "followup.prompt.v1" {
		t.Fatalf("unexpected follow-up action: %+v", result.AssistantAction)
	}
	if store.eventReservationInput.EventID != "event-1" {
		t.Fatalf("reservation event id = %q, want event-1", store.eventReservationInput.EventID)
	}
	if store.appendEvent.EventID != "event-1" || store.appendEvent.OutboxEventID != "outbox-1" {
		t.Fatalf("append ids not generated: %+v", store.appendEvent)
	}
}

func TestAppendSessionEventReplaySkipsAI(t *testing.T) {
	replay := AppendSessionEventResult{
		Acknowledged: true,
		Session:      sessionEventTestSession(1),
		AssistantAction: AssistantActionRecord{
			Type:          assistantActionSessionWait,
			SessionStatus: sharedtypes.SessionStatusRunning,
			Provenance:    (SessionEventService{}).assistantAction(assistantActionSessionWait, "", "", "", sharedtypes.SessionStatusRunning, "zh-CN", false).Provenance,
		},
	}
	store := &recordingPlanStore{
		eventReservation: SessionEventReservation{ReplayResult: &replay},
	}
	ai := &fakeAIClient{content: firstQuestionJSON(t, "Question?", "behavioral"), store: store}
	service := NewService(ServiceOptions{
		Store:    store,
		Registry: &fakePromptResolver{resolution: registry.PromptResolution{ModelProfileName: "practice.follow_up.default"}},
		AI:       ai,
	})

	result, err := service.AppendSessionEvent(context.Background(), AppendSessionEventRequest{
		UserID:        "user-1",
		SessionID:     "session-1",
		ClientEventID: "client-event-1",
		Kind:          "session_resumed",
	})
	if err != nil {
		t.Fatalf("AppendSessionEvent returned error: %v", err)
	}
	if !result.Replay || ai.profileName != "" || !reflect.DeepEqual(store.steps, []string{"reserve-event"}) {
		t.Fatalf("replay should skip AI and commit, result=%+v steps=%v ai=%q", result, store.steps, ai.profileName)
	}
}

func TestAppendSessionEventFollowUpAIFailureFallsBackToAskQuestion(t *testing.T) {
	now := time.Date(2026, 4, 28, 13, 45, 12, 0, time.UTC)
	store := &recordingPlanStore{
		eventReservation: SessionEventReservation{
			UserID:     "user-1",
			Session:    sessionEventTestSession(1),
			Plan:       sessionEventTestPlan(3, sharedtypes.PracticeGoalBaseline),
			LatestTurn: sessionEventTestTurn(1),
		},
	}
	service := NewService(ServiceOptions{
		Store:    store,
		Registry: &fakePromptResolver{resolution: registry.PromptResolution{ModelProfileName: "practice.follow_up.default", PromptVersion: "p", RubricVersion: "r", FeatureFlag: "none", DataSourceVersion: "registry.v1"}},
		AI:       &fakeAIClient{err: errors.New("AI_PROVIDER_TIMEOUT"), store: store},
		Now:      func() time.Time { return now },
		NewID:    sequenceIDs("event-1", "outbox-1"),
	})

	result, err := service.AppendSessionEvent(context.Background(), AppendSessionEventRequest{
		UserID:        "user-1",
		SessionID:     "session-1",
		ClientEventID: "client-event-1",
		Kind:          "answer_submitted",
		OccurredAt:    now,
		Payload: map[string]any{
			"turnId":     "turn-1",
			"answerText": "answer",
		},
	})
	if err != nil {
		t.Fatalf("AppendSessionEvent returned error: %v", err)
	}
	if result.AssistantAction.Type != assistantActionAskQuestion || result.AssistantAction.RequiresAI {
		t.Fatalf("AI failure should degrade to non-blocking ask_question: %+v", result.AssistantAction)
	}
}

func TestAppendSessionEventRejectsStaleTurnID(t *testing.T) {
	now := time.Date(2026, 4, 28, 13, 45, 12, 0, time.UTC)
	store := &recordingPlanStore{
		eventReservation: SessionEventReservation{
			UserID:  "user-1",
			Session: sessionEventTestSession(2),
			Plan:    sessionEventTestPlan(3, sharedtypes.PracticeGoalBaseline),
			LatestTurn: func() TurnRecord {
				turn := sessionEventTestTurn(2)
				turn.ID = "turn-2"
				return turn
			}(),
		},
	}
	service := NewService(ServiceOptions{Store: store, Now: func() time.Time { return now }})

	_, err := service.AppendSessionEvent(context.Background(), AppendSessionEventRequest{
		UserID:        "user-1",
		SessionID:     "session-1",
		ClientEventID: "client-event-stale",
		Kind:          "answer_submitted",
		OccurredAt:    now,
		Payload: map[string]any{
			"turnId":     "turn-1",
			"answerText": "stale answer",
		},
	})
	var svcErr *ServiceError
	if !errors.As(err, &svcErr) || svcErr.Code != sharederrors.CodePracticeSessionConflict {
		t.Fatalf("expected stale turn conflict, got %v", err)
	}
	if !reflect.DeepEqual(store.steps, []string{"reserve-event"}) {
		t.Fatalf("stale turn should stop before append, steps=%v", store.steps)
	}
}

func TestAppendSessionEventRejectsMissingAnswerText(t *testing.T) {
	now := time.Date(2026, 4, 28, 13, 45, 12, 0, time.UTC)
	store := &recordingPlanStore{
		eventReservation: SessionEventReservation{
			UserID:     "user-1",
			Session:    sessionEventTestSession(1),
			Plan:       sessionEventTestPlan(3, sharedtypes.PracticeGoalBaseline),
			LatestTurn: sessionEventTestTurn(1),
		},
	}
	service := NewService(ServiceOptions{Store: store, Now: func() time.Time { return now }})

	_, err := service.AppendSessionEvent(context.Background(), AppendSessionEventRequest{
		UserID:        "user-1",
		SessionID:     "session-1",
		ClientEventID: "client-event-missing-answer",
		Kind:          "answer_submitted",
		OccurredAt:    now,
		Payload: map[string]any{
			"turnId": "turn-1",
		},
	})
	var svcErr *ServiceError
	if !errors.As(err, &svcErr) || svcErr.Code != sharederrors.CodeValidationFailed || svcErr.Details["field"] != "payload.answerText" {
		t.Fatalf("expected missing answerText validation error, got %+v", err)
	}
	if len(store.steps) != 0 {
		t.Fatalf("missing answerText should stop before reservation, steps=%v", store.steps)
	}
}
