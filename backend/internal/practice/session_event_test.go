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
		{kind: "hint_requested", wantAction: "show_hint", wantStatus: sharedtypes.SessionStatusRunning},
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

func TestSessionEventServiceRoutesVoicePlaybackEvents(t *testing.T) {
	service := SessionEventService{}
	session := sessionEventTestSession(1)
	turn := sessionEventTestTurn(1)
	turn.Status = string(TurnStatusFollowUpRequested)
	plan := sessionEventTestPlan(3, sharedtypes.PracticeGoalBaseline)
	now := time.Date(2026, 5, 17, 8, 51, 0, 0, time.UTC)

	cases := []struct {
		kind    string
		payload map[string]any
	}{
		{
			kind: "tts_chunk_started",
			payload: map[string]any{
				"voiceTurnId":      "voice-turn-1",
				"chunkId":          "chunk-1",
				"playbackOffsetMs": 0,
			},
		},
		{
			kind: "tts_chunk_played",
			payload: map[string]any{
				"voiceTurnId":      "voice-turn-1",
				"chunkId":          "chunk-1",
				"playedTextHash":   "sha256:chunk-1",
				"playedTextLength": 36,
				"playbackOffsetMs": 2840,
			},
		},
		{
			kind: "barge_in_detected",
			payload: map[string]any{
				"voiceTurnId":         "voice-turn-1",
				"chunkId":             "chunk-1",
				"playbackOffsetMs":    1480,
				"userSpeechStartedAt": "2026-05-17T08:51:05Z",
			},
		},
		{
			kind: "assistant_context_committed",
			payload: map[string]any{
				"voiceTurnId":         "voice-turn-1",
				"chunkId":             "chunk-1",
				"committedTextHash":   "sha256:chunk-1",
				"committedTextLength": 36,
				"playbackOffsetMs":    2840,
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.kind, func(t *testing.T) {
			out, err := service.Route(context.Background(), SessionEventInput{
				SessionID:     session.ID,
				ClientEventID: "client-event-voice",
				Kind:          tc.kind,
				OccurredAt:    now,
				Payload:       tc.payload,
			}, session, turn, plan)
			if err != nil {
				t.Fatalf("Route returned error: %v", err)
			}
			if out.Error != nil {
				t.Fatalf("unexpected outcome error: %+v", out.Error)
			}
			if !out.Acknowledged || out.AssistantAction.Type != assistantActionSessionWait || out.AssistantAction.TurnID != turn.ID {
				t.Fatalf("unexpected voice playback outcome: %+v", out)
			}
			if out.NextSessionStatus != sharedtypes.SessionStatusRunning {
				t.Fatalf("NextSessionStatus = %s", out.NextSessionStatus)
			}
			if out.AuditMetadata["event_kind"] != tc.kind ||
				out.AuditMetadata["voice_turn_id"] != "voice-turn-1" ||
				out.AuditMetadata["chunk_id"] != "chunk-1" {
				t.Fatalf("voice audit metadata missing event summary: %+v", out.AuditMetadata)
			}
			if _, leaked := out.AuditMetadata["assistantTextDraft"]; leaked {
				t.Fatalf("voice audit metadata must not store assistant text: %+v", out.AuditMetadata)
			}
		})
	}
}

func TestSessionEventServiceRejectsMalformedVoicePlaybackEvent(t *testing.T) {
	service := SessionEventService{}
	session := sessionEventTestSession(1)
	turn := sessionEventTestTurn(1)
	turn.Status = string(TurnStatusFollowUpRequested)
	out, err := service.Route(context.Background(), SessionEventInput{
		SessionID:     session.ID,
		ClientEventID: "client-event-voice",
		Kind:          "tts_chunk_played",
		OccurredAt:    time.Date(2026, 5, 17, 8, 51, 0, 0, time.UTC),
		Payload: map[string]any{
			"chunkId":          "chunk-1",
			"playedTextHash":   "sha256:chunk-1",
			"playedTextLength": 36,
			"playbackOffsetMs": 2840,
		},
	}, session, turn, sessionEventTestPlan(3, sharedtypes.PracticeGoalBaseline))
	if err != nil {
		t.Fatalf("Route returned error: %v", err)
	}
	if out.Error == nil || out.Error.Code != sharederrors.CodeValidationFailed || out.Error.Details["field"] != "payload.voiceTurnId" {
		t.Fatalf("expected payload.voiceTurnId validation error, got %+v", out.Error)
	}
}

func TestHandleAnswerSubmittedDecisionBranches(t *testing.T) {
	service := SessionEventService{}
	now := time.Date(2026, 4, 28, 13, 45, 12, 0, time.UTC)

	cases := []struct {
		name                 string
		turnCount            int32
		questionBudget       int32
		turnFollowUpCount    int
		payloadFollowUpCount *int
		wantAction           string
		wantStatus           sharedtypes.SessionStatus
		wantTurnStatus       TurnStatus
		wantNextFollowUps    int
		wantOutbox           bool
	}{
		{
			name:                 "ask follow up before first stored follow up and ignore client count",
			turnCount:            1,
			questionBudget:       3,
			turnFollowUpCount:    0,
			payloadFollowUpCount: intPtr(99),
			wantAction:           "ask_follow_up",
			wantStatus:           sharedtypes.SessionStatusRunning,
			wantTurnStatus:       TurnStatusFollowUpRequested,
			wantNextFollowUps:    1,
			wantOutbox:           false,
		},
		{
			name:              "ask next question after stored follow up without client count",
			turnCount:         1,
			questionBudget:    3,
			turnFollowUpCount: 1,
			wantAction:        "ask_question",
			wantStatus:        sharedtypes.SessionStatusRunning,
			wantTurnStatus:    TurnStatusAssessed,
			wantNextFollowUps: 1,
			wantOutbox:        true,
		},
		{
			name:                 "complete at question budget from stored count",
			turnCount:            3,
			questionBudget:       3,
			turnFollowUpCount:    1,
			payloadFollowUpCount: intPtr(0),
			wantAction:           "session_completed",
			wantStatus:           sharedtypes.SessionStatusCompleted,
			wantTurnStatus:       TurnStatusAssessed,
			wantNextFollowUps:    1,
			wantOutbox:           true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			session := sessionEventTestSession(tc.turnCount)
			turn := sessionEventTestTurn(tc.turnCount)
			turn.FollowUpCount = tc.turnFollowUpCount
			payload := map[string]any{
				"turnId":     turn.ID,
				"answerText": "answer",
			}
			if tc.payloadFollowUpCount != nil {
				payload["followUpCount"] = *tc.payloadFollowUpCount
			}
			out, err := service.Route(context.Background(), SessionEventInput{
				SessionID:     session.ID,
				ClientEventID: "client-event-1",
				Kind:          "answer_submitted",
				OccurredAt:    now,
				Payload:       payload,
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
			if out.NextTurn.FollowUpCount != tc.wantNextFollowUps {
				t.Fatalf("NextTurn.FollowUpCount = %d, want %d", out.NextTurn.FollowUpCount, tc.wantNextFollowUps)
			}
			if (out.OutboxRecord != nil) != tc.wantOutbox {
				t.Fatalf("OutboxRecord present = %v, want %v", out.OutboxRecord != nil, tc.wantOutbox)
			}
			if out.OutboxRecord != nil && out.OutboxRecord.FollowUpCount != tc.wantNextFollowUps {
				t.Fatalf("OutboxRecord.FollowUpCount = %d, want %d", out.OutboxRecord.FollowUpCount, tc.wantNextFollowUps)
			}
		})
	}
}

func TestRouteRejectsClosedSessionAndTerminalTurnEvents(t *testing.T) {
	service := SessionEventService{}
	now := time.Date(2026, 4, 28, 13, 45, 12, 0, time.UTC)
	basePayload := map[string]any{
		"turnId":     "turn-1",
		"answerText": "late duplicate answer",
	}

	cases := []struct {
		name    string
		session SessionRecord
		turn    TurnRecord
	}{
		{
			name: "completed session rejects new answer event",
			session: func() SessionRecord {
				session := sessionEventTestSession(3)
				session.Status = sharedtypes.SessionStatusCompleted
				return session
			}(),
			turn: sessionEventTestTurn(3),
		},
		{
			name:    "assessed turn rejects new answer event on running session",
			session: sessionEventTestSession(1),
			turn: func() TurnRecord {
				turn := sessionEventTestTurn(1)
				turn.Status = string(TurnStatusAssessed)
				return turn
			}(),
		},
		{
			name:    "skipped turn rejects new answer event on running session",
			session: sessionEventTestSession(1),
			turn: func() TurnRecord {
				turn := sessionEventTestTurn(1)
				turn.Status = string(TurnStatusSkipped)
				return turn
			}(),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out, err := service.Route(context.Background(), SessionEventInput{
				SessionID:     tc.session.ID,
				ClientEventID: "client-event-late",
				Kind:          "answer_submitted",
				OccurredAt:    now,
				Payload:       basePayload,
			}, tc.session, tc.turn, sessionEventTestPlan(3, sharedtypes.PracticeGoalBaseline))
			if err != nil {
				t.Fatalf("Route returned error: %v", err)
			}
			if out.Error == nil || out.Error.Code != sharederrors.CodePracticeSessionConflict {
				t.Fatalf("Error = %+v, want PRACTICE_SESSION_CONFLICT", out.Error)
			}
		})
	}
}

func TestHandleHintRequestedModeMatrix(t *testing.T) {
	service := SessionEventService{}
	now := time.Date(2026, 4, 28, 13, 45, 12, 0, time.UTC)
	goals := []sharedtypes.PracticeGoal{
		sharedtypes.PracticeGoalBaseline,
		sharedtypes.PracticeGoalRetryCurrentRound,
		sharedtypes.PracticeGoalNextRound,
		sharedtypes.PracticeGoalDebrief,
	}

	for _, goal := range goals {
		t.Run("assisted/"+string(goal), func(t *testing.T) {
			session := sessionEventTestSession(1)
			plan := sessionEventTestPlan(3, goal)
			plan.Mode = sharedtypes.PracticeModeAssisted
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
			if out.Error != nil {
				t.Fatalf("unexpected assisted error: %+v", out.Error)
			}
			if out.AssistantAction.Type != assistantActionShowHint || !out.AssistantAction.RequiresAI {
				t.Fatalf("AssistantAction = %+v, want show_hint requiring AI", out.AssistantAction)
			}
			if out.NextSessionStatus != session.Status || out.OutboxRecord != nil || out.NextTurn != nil {
				t.Fatalf("hint should preserve lifecycle: %+v", out)
			}
		})

		t.Run("strict/"+string(goal), func(t *testing.T) {
			session := sessionEventTestSession(1)
			plan := sessionEventTestPlan(3, goal)
			plan.Mode = sharedtypes.PracticeModeStrict
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
			if out.Error.Details["policy"] != "hint_disabled_in_mode" || out.Error.Details["mode"] != string(sharedtypes.PracticeModeStrict) {
				t.Fatalf("unexpected details: %+v", out.Error.Details)
			}
		})
	}

	for _, mode := range []sharedtypes.PracticeMode{"", "legacy debrief replay value"} {
		t.Run("unknown/"+string(mode), func(t *testing.T) {
			session := sessionEventTestSession(1)
			plan := sessionEventTestPlan(3, sharedtypes.PracticeGoalBaseline)
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
		})
	}
}

func TestHandleHintRequestedTurnLifecycle(t *testing.T) {
	service := SessionEventService{}
	session := sessionEventTestSession(1)
	plan := sessionEventTestPlan(3, sharedtypes.PracticeGoalBaseline)
	plan.Mode = sharedtypes.PracticeModeAssisted
	out, err := service.Route(context.Background(), SessionEventInput{
		SessionID:     session.ID,
		ClientEventID: "client-event-1",
		Kind:          "hint_requested",
		OccurredAt:    time.Date(2026, 4, 28, 13, 45, 12, 0, time.UTC),
		Payload:       map[string]any{"turnId": "turn-1"},
	}, session, sessionEventTestTurn(1), plan)
	if err != nil {
		t.Fatalf("Route returned error: %v", err)
	}
	if out.OutboxRecord != nil || out.NextTurn != nil || out.NextSessionStatus != session.Status {
		t.Fatalf("hint should not advance turn/session lifecycle: %+v", out)
	}
	want := map[string]any{"event_kind": sessionEventKindHintRequested, "mode": "assisted"}
	if !reflect.DeepEqual(out.AuditMetadata, want) {
		t.Fatalf("AuditMetadata = %+v, want %+v", out.AuditMetadata, want)
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

func TestAssistantActionProvenanceFieldsAreWireOnly(t *testing.T) {
	got := reflect.TypeOf(AssistantActionProvenance{})
	fields := make(map[string]string, got.NumField())
	for i := 0; i < got.NumField(); i++ {
		field := got.Field(i)
		fields[field.Name] = field.Tag.Get("json")
	}
	want := map[string]string{
		"PromptVersion":     "promptVersion",
		"RubricVersion":     "rubricVersion",
		"ModelID":           "modelId",
		"Language":          "language",
		"FeatureFlag":       "featureFlag",
		"DataSourceVersion": "dataSourceVersion",
	}
	if !reflect.DeepEqual(fields, want) {
		t.Fatalf("provenance fields = %+v, want %+v", fields, want)
	}
	for _, forbidden := range []string{"FeatureKey", "ModelProfileName", "Provider", "Cost", "Latency"} {
		if _, ok := fields[forbidden]; ok {
			t.Fatalf("runtime field %s must not be exposed on wire provenance", forbidden)
		}
	}
}

func TestAssistantActionProvenanceJSONShape(t *testing.T) {
	raw, err := json.Marshal(AssistantActionProvenance{
		PromptVersion:     "p",
		RubricVersion:     "not_applicable",
		ModelID:           "model-profile:practice.turn_observe.default",
		Language:          "en",
		FeatureFlag:       "none",
		DataSourceVersion: "registry.v1",
	})
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
	if got["rubricVersion"] != "not_applicable" {
		t.Fatalf("non-scoring action rubric version = %q", got["rubricVersion"])
	}
}

func TestAssistantActionProvenanceCrossActionParity(t *testing.T) {
	service := SessionEventService{}
	actions := []string{
		assistantActionShowHint,
		assistantActionAskQuestion,
		assistantActionAskFollowUp,
		assistantActionSessionWait,
		assistantActionSessionCompleted,
	}
	wantKeys := []string{"dataSourceVersion", "featureFlag", "language", "modelId", "promptVersion", "rubricVersion"}
	for _, action := range actions {
		t.Run(action, func(t *testing.T) {
			record := service.assistantAction(action, "turn-1", "question", "intent", sharedtypes.SessionStatusRunning, "en", false)
			raw, err := json.Marshal(record.Provenance)
			if err != nil {
				t.Fatalf("marshal provenance: %v", err)
			}
			var got map[string]string
			if err := json.Unmarshal(raw, &got); err != nil {
				t.Fatalf("unmarshal provenance: %v", err)
			}
			if !reflect.DeepEqual(sortedKeys(got), wantKeys) {
				t.Fatalf("action %s provenance keys = %v, want %v", action, sortedKeys(got), wantKeys)
			}
		})
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

func intPtr(value int) *int {
	return &value
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
