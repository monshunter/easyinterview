package practice

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"
	"time"

	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

func TestSessionEventServiceRouteCoversAllKinds(t *testing.T) {
	service := SessionEventService{}
	session := sessionEventTestSession(1)
	turn := sessionEventTestTurn(1)
	plan := sessionEventTestPlan(3, sharedtypes.PracticeGoalBaseline)
	now := time.Date(2026, 4, 28, 13, 45, 12, 0, time.UTC)

	cases := []struct {
		kind       string
		wantAction string
		wantStatus sharedtypes.SessionStatus
		wantError  string
	}{
		{kind: "answer_submitted", wantAction: "ask_follow_up", wantStatus: sharedtypes.SessionStatusRunning},
		{kind: "hint_requested", wantError: sharederrors.CodePracticeSessionConflict, wantStatus: sharedtypes.SessionStatusRunning},
		{kind: "turn_skipped", wantAction: "ask_question", wantStatus: sharedtypes.SessionStatusRunning},
		{kind: "session_paused", wantAction: "session_wait", wantStatus: sharedtypes.SessionStatusWaitingUserInput},
		{kind: "session_resumed", wantAction: "session_wait", wantStatus: sharedtypes.SessionStatusRunning},
	}

	for _, tc := range cases {
		t.Run(tc.kind, func(t *testing.T) {
			out, err := service.Route(context.Background(), SessionEventInput{
				SessionID:     session.ID,
				ClientEventID: "client-event-1",
				Kind:          tc.kind,
				OccurredAt:    now,
				Payload: map[string]any{
					"turnId":     turn.ID,
					"answerText": "answer",
				},
			}, session, turn, plan)
			if err != nil {
				t.Fatalf("Route returned error: %v", err)
			}
			if out.NextSessionStatus != tc.wantStatus {
				t.Fatalf("NextSessionStatus = %s, want %s", out.NextSessionStatus, tc.wantStatus)
			}
			if tc.wantError != "" {
				if out.Error == nil || out.Error.Code != tc.wantError {
					t.Fatalf("Error = %+v, want code %s", out.Error, tc.wantError)
				}
				return
			}
			if out.Error != nil {
				t.Fatalf("unexpected outcome error: %+v", out.Error)
			}
			if out.AssistantAction.Type != tc.wantAction {
				t.Fatalf("AssistantAction.Type = %s, want %s", out.AssistantAction.Type, tc.wantAction)
			}
		})
	}
}

func TestHandleAnswerSubmittedDecisionBranches(t *testing.T) {
	service := SessionEventService{}
	now := time.Date(2026, 4, 28, 13, 45, 12, 0, time.UTC)

	cases := []struct {
		name           string
		turnCount      int32
		questionBudget int32
		followUpCount  int
		wantAction     string
		wantStatus     sharedtypes.SessionStatus
		wantTurnStatus TurnStatus
	}{
		{
			name:           "ask follow up before first follow up",
			turnCount:      1,
			questionBudget: 3,
			followUpCount:  0,
			wantAction:     "ask_follow_up",
			wantStatus:     sharedtypes.SessionStatusRunning,
			wantTurnStatus: TurnStatusFollowUpRequested,
		},
		{
			name:           "ask next question after one follow up",
			turnCount:      1,
			questionBudget: 3,
			followUpCount:  1,
			wantAction:     "ask_question",
			wantStatus:     sharedtypes.SessionStatusRunning,
			wantTurnStatus: TurnStatusAnswered,
		},
		{
			name:           "complete at question budget",
			turnCount:      3,
			questionBudget: 3,
			followUpCount:  1,
			wantAction:     "session_completed",
			wantStatus:     sharedtypes.SessionStatusCompleted,
			wantTurnStatus: TurnStatusAssessed,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			session := sessionEventTestSession(tc.turnCount)
			turn := sessionEventTestTurn(tc.turnCount)
			out, err := service.Route(context.Background(), SessionEventInput{
				SessionID:     session.ID,
				ClientEventID: "client-event-1",
				Kind:          "answer_submitted",
				OccurredAt:    now,
				Payload: map[string]any{
					"turnId":        turn.ID,
					"answerText":    "answer",
					"followUpCount": tc.followUpCount,
				},
			}, session, turn, sessionEventTestPlan(tc.questionBudget, sharedtypes.PracticeGoalBaseline))
			if err != nil {
				t.Fatalf("Route returned error: %v", err)
			}
			if out.Error != nil {
				t.Fatalf("unexpected outcome error: %+v", out.Error)
			}
			if out.AssistantAction.Type != tc.wantAction {
				t.Fatalf("AssistantAction.Type = %s, want %s", out.AssistantAction.Type, tc.wantAction)
			}
			if out.NextSessionStatus != tc.wantStatus {
				t.Fatalf("NextSessionStatus = %s, want %s", out.NextSessionStatus, tc.wantStatus)
			}
			if out.NextTurn == nil || TurnStatus(out.NextTurn.Status) != tc.wantTurnStatus {
				t.Fatalf("NextTurn.Status = %+v, want %s", out.NextTurn, tc.wantTurnStatus)
			}
		})
	}
}

func TestHandleHintRequestedDefaultsToStrictConflict(t *testing.T) {
	service := SessionEventService{}
	now := time.Date(2026, 4, 28, 13, 45, 12, 0, time.UTC)
	modes := []sharedtypes.PracticeMode{sharedtypes.PracticeModeAssisted, sharedtypes.PracticeModeStrict}
	goals := []sharedtypes.PracticeGoal{
		sharedtypes.PracticeGoalBaseline,
		sharedtypes.PracticeGoalRetryCurrentRound,
		sharedtypes.PracticeGoalNextRound,
		sharedtypes.PracticeGoalDebrief,
	}

	for _, mode := range modes {
		for _, goal := range goals {
			t.Run(string(mode)+"/"+string(goal), func(t *testing.T) {
				session := sessionEventTestSession(1)
				plan := sessionEventTestPlan(3, goal)
				plan.Mode = mode
				out, err := service.Route(context.Background(), SessionEventInput{
					SessionID:     session.ID,
					ClientEventID: "client-event-1",
					Kind:          "hint_requested",
					OccurredAt:    now,
					Payload: map[string]any{
						"turnId": "turn-1",
					},
				}, session, sessionEventTestTurn(1), plan)
				if err != nil {
					t.Fatalf("Route returned error: %v", err)
				}
				if out.Error == nil || out.Error.Code != sharederrors.CodePracticeSessionConflict {
					t.Fatalf("Error = %+v, want PRACTICE_SESSION_CONFLICT", out.Error)
				}
				if out.Error.Details["policy"] != "hint_disabled_in_mode" || out.Error.Details["mode"] != string(mode) {
					t.Fatalf("unexpected details: %+v", out.Error.Details)
				}
			})
		}
	}
}

func TestUnknownKindReturnsValidationFailedOutcome(t *testing.T) {
	service := SessionEventService{}
	out, err := service.Route(context.Background(), SessionEventInput{
		SessionID:     "session-1",
		ClientEventID: "client-event-1",
		Kind:          "unknown_kind",
		OccurredAt:    time.Date(2026, 4, 28, 13, 45, 12, 0, time.UTC),
	}, sessionEventTestSession(1), sessionEventTestTurn(1), sessionEventTestPlan(3, sharedtypes.PracticeGoalBaseline))
	if err != nil {
		t.Fatalf("Route returned error: %v", err)
	}
	if out.Error == nil || out.Error.Code != sharederrors.CodeValidationFailed {
		t.Fatalf("Error = %+v, want VALIDATION_FAILED", out.Error)
	}
}

func TestStaticAssistantActionProvenanceUsesB2WireFields(t *testing.T) {
	service := SessionEventService{}
	out, err := service.Route(context.Background(), SessionEventInput{
		SessionID:     "session-1",
		ClientEventID: "client-event-1",
		Kind:          "session_paused",
		OccurredAt:    time.Date(2026, 4, 28, 13, 45, 12, 0, time.UTC),
	}, sessionEventTestSession(1), sessionEventTestTurn(1), sessionEventTestPlan(3, sharedtypes.PracticeGoalBaseline))
	if err != nil {
		t.Fatalf("Route returned error: %v", err)
	}

	raw, err := json.Marshal(out.AssistantAction.Provenance)
	if err != nil {
		t.Fatalf("marshal provenance: %v", err)
	}
	var got map[string]string
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("unmarshal provenance: %v", err)
	}
	wantKeys := []string{"dataSourceVersion", "featureFlag", "language", "modelId", "promptVersion", "rubricVersion"}
	if !reflect.DeepEqual(sortedKeys(got), wantKeys) {
		t.Fatalf("provenance keys = %v, want %v", sortedKeys(got), wantKeys)
	}
	for _, key := range wantKeys {
		if got[key] == "" {
			t.Fatalf("provenance field %s must be populated", key)
		}
	}
}

func sessionEventTestSession(turnCount int32) SessionRecord {
	return SessionRecord{
		ID:           "session-1",
		PlanID:       "plan-1",
		TargetJobID:  "target-1",
		Status:       sharedtypes.SessionStatusRunning,
		Language:     "zh-CN",
		HintsEnabled: true,
		TurnCount:    turnCount,
		CurrentTurn:  ptrTurn(sessionEventTestTurn(turnCount)),
		CreatedAt:    time.Date(2026, 4, 28, 12, 0, 0, 0, time.UTC),
		UpdatedAt:    time.Date(2026, 4, 28, 13, 40, 0, 0, time.UTC),
	}
}

func sessionEventTestTurn(index int32) TurnRecord {
	return TurnRecord{
		ID:             "turn-1",
		TurnIndex:      index,
		QuestionText:   "Tell me about a cross-team migration.",
		QuestionIntent: "behavioral.leadership",
		Status:         string(TurnStatusAsked),
		AskedAt:        time.Date(2026, 4, 28, 13, 40, 0, 0, time.UTC),
	}
}

func sessionEventTestPlan(questionBudget int32, goal sharedtypes.PracticeGoal) PlanRecord {
	return PlanRecord{
		ID:                 "plan-1",
		TargetJobID:        "target-1",
		Goal:               goal,
		Mode:               sharedtypes.PracticeModeAssisted,
		InterviewerPersona: sharedtypes.InterviewerRoleHiringManager,
		Difficulty:         "standard",
		Language:           "zh-CN",
		TimeBudgetMinutes:  30,
		QuestionBudget:     questionBudget,
		Status:             "ready",
		CreatedAt:          time.Date(2026, 4, 28, 12, 0, 0, 0, time.UTC),
	}
}

func ptrTurn(turn TurnRecord) *TurnRecord {
	return &turn
}

func sortedKeys(values map[string]string) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	for i := 1; i < len(keys); i++ {
		for j := i; j > 0 && keys[j-1] > keys[j]; j-- {
			keys[j-1], keys[j] = keys[j], keys[j-1]
		}
	}
	return keys
}
