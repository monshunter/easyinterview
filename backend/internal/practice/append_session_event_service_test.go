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
	ai := &fakeAIClient{
		contents: []string{
			`{"cue":"Good evidence cue.","answerSummary":"Candidate explained aligning 12 teams and measuring adoption."}`,
			firstQuestionJSON(t, "What was the strongest objection?", "behavioral.depth"),
		},
		store: store,
	}
	service := NewService(ServiceOptions{
		Store: store,
		Registry: &fakePromptResolver{resolutions: map[string]registry.PromptResolution{
			hintFeatureKey: {
				PromptVersion:       "observe.prompt.v1",
				RubricVersion:       "observe.rubric.v1",
				ModelProfileName:    "practice.turn_observe.default",
				FeatureFlag:         "observe_v1",
				DataSourceVersion:   "registry.v1",
				OutputSchema:        practiceOutputSchema(`{"type":"object","required":["cue","answerSummary"],"properties":{"cue":{"type":"string"},"answerSummary":{"type":"string"}}}`),
				UserMessageTemplate: "observe answer",
			},
			followUpFeatureKey: {
				PromptVersion:       "followup.prompt.v1",
				RubricVersion:       "followup.rubric.v1",
				ModelProfileName:    "practice.follow_up.default",
				FeatureFlag:         "follow_up_v1",
				DataSourceVersion:   "registry.v1",
				OutputSchema:        practiceOutputSchema(`{"type":"object","required":["questionText","questionIntent"],"properties":{"questionText":{"type":"string"},"questionIntent":{"type":"string"}}}`),
				UserMessageTemplate: "ask a follow up",
			},
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
	if !reflect.DeepEqual(store.steps, []string{"reserve-event", "ai", "ai", "append-event"}) {
		t.Fatalf("steps = %v", store.steps)
	}
	if !ai.calledOutsideTransaction {
		t.Fatalf("turn observation and follow-up AI must run outside repository transaction")
	}
	if result.AssistantAction.Type != assistantActionAskFollowUp ||
		result.AssistantAction.QuestionText != "What was the strongest objection?" ||
		result.AssistantAction.Provenance.PromptVersion != "followup.prompt.v1" {
		t.Fatalf("unexpected follow-up action: %+v", result.AssistantAction)
	}
	if store.appendEvent.Outcome.AnswerSummary != "Candidate explained aligning 12 teams and measuring adoption." {
		t.Fatalf("answer summary was not carried to the store: %+v", store.appendEvent.Outcome)
	}
	if len(ai.payloads) != 2 ||
		ai.payloads[0].Metadata.FeatureKey != hintFeatureKey ||
		ai.payloads[0].Metadata.TaskRun.Capability != aiclient.AITaskRunTaskHintGenerate ||
		ai.payloads[1].Metadata.FeatureKey != followUpFeatureKey ||
		ai.payloads[1].Metadata.TaskRun.Capability != aiclient.AITaskRunTaskFollowupGenerate {
		t.Fatalf("unexpected AI payloads: %+v", ai.payloads)
	}
	if len(ai.payload.Metadata.OutputSchema) == 0 {
		t.Fatalf("follow-up metadata OutputSchema must be populated")
	}
	if store.eventReservationInput.EventID != "event-1" {
		t.Fatalf("reservation event id = %q, want event-1", store.eventReservationInput.EventID)
	}
	if store.appendEvent.EventID != "event-1" || store.appendEvent.OutboxEventID != "outbox-1" {
		t.Fatalf("append ids not generated: %+v", store.appendEvent)
	}
}

func TestAppendSessionEventVoicePlaybackCommitsWithoutAI(t *testing.T) {
	now := time.Date(2026, 5, 17, 8, 51, 4, 0, time.UTC)
	store := &recordingPlanStore{
		eventReservation: SessionEventReservation{
			UserID:  "user-1",
			Session: sessionEventTestSession(1),
			Plan:    sessionEventTestPlan(3, sharedtypes.PracticeGoalBaseline),
			LatestTurn: TurnRecord{
				ID:             "turn-1",
				TurnIndex:      1,
				QuestionText:   "Question?",
				QuestionIntent: "behavioral.depth",
				Status:         string(TurnStatusFollowUpRequested),
				FollowUpCount:  1,
				AskedAt:        now.Add(-time.Minute),
			},
		},
	}
	ai := &fakeAIClient{content: firstQuestionJSON(t, "AI must not run", "must-not-run"), store: store}
	service := NewService(ServiceOptions{
		Store: store,
		AI:    ai,
		Now:   func() time.Time { return now },
		NewID: sequenceIDs("event-voice-1", "outbox-voice-1"),
	})

	result, err := service.AppendSessionEvent(context.Background(), AppendSessionEventRequest{
		UserID:        "user-1",
		SessionID:     "session-1",
		ClientEventID: "client-event-voice-1",
		Kind:          sessionEventKindTTSChunkPlayed,
		OccurredAt:    now,
		Payload: map[string]any{
			"voiceTurnId":      "voice-turn-1",
			"chunkId":          "chunk-1",
			"playedTextHash":   "sha256:chunk-1",
			"playedTextLength": 36,
			"playbackOffsetMs": 2840,
		},
	})
	if err != nil {
		t.Fatalf("AppendSessionEvent returned error: %v", err)
	}
	if !reflect.DeepEqual(store.steps, []string{"reserve-event", "append-event"}) {
		t.Fatalf("voice playback should not call AI, steps=%v", store.steps)
	}
	if result.AssistantAction.Type != assistantActionSessionWait || result.AssistantAction.TurnID != "turn-1" {
		t.Fatalf("unexpected assistant action: %+v", result.AssistantAction)
	}
	if store.appendEvent.Kind != sessionEventKindTTSChunkPlayed ||
		store.appendEvent.ClientEventID != "client-event-voice-1" ||
		store.appendEvent.RequestPayload["voiceTurnId"] != "voice-turn-1" {
		t.Fatalf("voice event append input drift: %+v", store.appendEvent)
	}
	if store.appendEvent.Outcome.AuditMetadata["played_text_hash"] != "sha256:chunk-1" {
		t.Fatalf("voice playback audit summary missing: %+v", store.appendEvent.Outcome.AuditMetadata)
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

func TestAppendSessionEventReplayReturnsStoredErrorBeforeResult(t *testing.T) {
	replay := AppendSessionEventResult{Acknowledged: true}
	replayErr := &ServiceError{
		Code:    sharederrors.CodePracticeSessionConflict,
		Message: "client event payload mismatch",
		Details: map[string]any{
			"policy": "client_event_payload_mismatch",
		},
	}
	store := &recordingPlanStore{
		eventReservation: SessionEventReservation{
			ReplayResult: &replay,
			ReplayError:  replayErr,
		},
	}
	service := NewService(ServiceOptions{Store: store})

	_, err := service.AppendSessionEvent(context.Background(), AppendSessionEventRequest{
		UserID:        "user-1",
		SessionID:     "session-1",
		ClientEventID: "client-event-1",
		Kind:          "session_resumed",
	})
	var svcErr *ServiceError
	if !errors.As(err, &svcErr) || svcErr.Code != sharederrors.CodePracticeSessionConflict {
		t.Fatalf("expected stored replay error, got %v", err)
	}
	if !reflect.DeepEqual(store.steps, []string{"reserve-event"}) {
		t.Fatalf("error replay should not append side effects, steps=%v", store.steps)
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

func TestServiceAppliesHintAIForAssisted(t *testing.T) {
	now := time.Date(2026, 4, 28, 13, 47, 32, 0, time.UTC)
	store := &recordingPlanStore{
		eventReservation: SessionEventReservation{
			UserID:  "01918fa0-0010-7a00-8a00-000000000001",
			Session: sessionEventTestSession(1),
			Plan: func() PlanRecord {
				plan := sessionEventTestPlan(3, sharedtypes.PracticeGoalBaseline)
				plan.Mode = sharedtypes.PracticeModeAssisted
				plan.TargetJobID = "01918fa0-0020-7a00-8a00-000000000002"
				return plan
			}(),
			LatestTurn: sessionEventTestTurn(1),
		},
	}
	ai := &fakeAIClient{content: `{"cue":"Use one measurable tradeoff."}`, store: store}
	service := NewService(ServiceOptions{
		Store: store,
		Registry: &fakePromptResolver{resolution: registry.PromptResolution{
			PromptVersion:       "hint.prompt.v1",
			RubricVersion:       "hint.rubric.v1",
			ModelProfileName:    "practice.turn_observe.default",
			FeatureFlag:         "none",
			DataSourceVersion:   "registry.v1",
			OutputSchema:        practiceOutputSchema(`{"type":"object","required":["cue"],"properties":{"cue":{"type":"string"}}}`),
			UserMessageTemplate: "give a hint",
		}},
		AI:    ai,
		Now:   func() time.Time { return now },
		NewID: sequenceIDs("event-1", "outbox-1"),
	})

	result, err := service.AppendSessionEvent(context.Background(), AppendSessionEventRequest{
		UserID:        "01918fa0-0010-7a00-8a00-000000000001",
		SessionID:     "session-1",
		ClientEventID: "client-event-1",
		Kind:          "hint_requested",
		OccurredAt:    now,
		Payload:       map[string]any{"turnId": "turn-1"},
	})
	if err != nil {
		t.Fatalf("AppendSessionEvent returned error: %v", err)
	}
	if !reflect.DeepEqual(store.steps, []string{"reserve-event", "ai", "append-event"}) {
		t.Fatalf("steps = %v", store.steps)
	}
	if result.AssistantAction.Type != assistantActionShowHint || result.AssistantAction.Hint != "Use one measurable tradeoff." {
		t.Fatalf("unexpected hint action: %+v", result.AssistantAction)
	}
	if result.AssistantAction.Provenance.RubricVersion != "not_applicable" ||
		result.AssistantAction.Provenance.PromptVersion != "hint.prompt.v1" {
		t.Fatalf("unexpected provenance: %+v", result.AssistantAction.Provenance)
	}
	if store.appendEvent.Outcome.OutboxRecord != nil || store.appendEvent.Outcome.NextTurn != nil {
		t.Fatalf("hint must not advance turn lifecycle: %+v", store.appendEvent.Outcome)
	}
	if ai.payload.Metadata.TaskRun.Capability != aiclient.AITaskRunTaskHintGenerate ||
		ai.payload.Metadata.FeatureKey != hintFeatureKey {
		t.Fatalf("unexpected AI metadata: %+v", ai.payload.Metadata)
	}
	if len(ai.payload.Metadata.OutputSchema) == 0 {
		t.Fatalf("hint metadata OutputSchema must be populated")
	}
}

func TestServiceAppliesHintAIForStrictMode(t *testing.T) {
	now := time.Date(2026, 4, 28, 13, 47, 32, 0, time.UTC)
	store := &recordingPlanStore{
		eventReservation: SessionEventReservation{
			UserID:  "01918fa0-0010-7a00-8a00-000000000001",
			Session: sessionEventTestSession(1),
			Plan: func() PlanRecord {
				plan := sessionEventTestPlan(3, sharedtypes.PracticeGoalBaseline)
				plan.Mode = sharedtypes.PracticeModeStrict
				return plan
			}(),
			LatestTurn: sessionEventTestTurn(1),
		},
	}
	ai := &fakeAIClient{content: `{"cue":"Use a measurable tradeoff."}`, store: store}
	service := NewService(ServiceOptions{
		Store:    store,
		Registry: &fakePromptResolver{resolution: hintTestResolution()},
		AI:       ai,
		Now:      func() time.Time { return now },
		NewID:    sequenceIDs("event-1", "outbox-1"),
	})

	result, err := service.AppendSessionEvent(context.Background(), AppendSessionEventRequest{
		UserID:        "01918fa0-0010-7a00-8a00-000000000001",
		SessionID:     "session-1",
		ClientEventID: "client-event-1",
		Kind:          "hint_requested",
		OccurredAt:    now,
		Payload:       map[string]any{"turnId": "turn-1"},
	})
	if err != nil {
		t.Fatalf("AppendSessionEvent returned error: %v", err)
	}
	if result.AssistantAction.Type != assistantActionShowHint || result.AssistantAction.Hint != "Use a measurable tradeoff." {
		t.Fatalf("unexpected strict hint action: %+v", result.AssistantAction)
	}
	if !reflect.DeepEqual(store.steps, []string{"reserve-event", "ai", "append-event"}) {
		t.Fatalf("strict-mode hint should call AI and append, steps=%v", store.steps)
	}
}

func TestApplyHintAISuccess(t *testing.T) {
	reservation := hintTestReservation()
	ai := &fakeAIClient{
		content: `{"cue":"Tie the answer to a concrete metric.","severity":"nudge","dimensionHint":"evidence"}`,
		meta: aiclient.AICallMeta{
			ModelID:          "stub-chat-1",
			ValidationStatus: aiclient.ValidationStatusOK,
		},
	}
	service := NewService(ServiceOptions{
		Registry: &fakePromptResolver{resolution: hintTestResolution()},
		AI:       ai,
	})
	outcome := hintPendingOutcome(reservation)

	service.applyHintAI(context.Background(), reservation, map[string]any{"answerText": "short answer"}, outcome)

	if outcome.AssistantAction.Type != assistantActionShowHint ||
		outcome.AssistantAction.Hint != "Tie the answer to a concrete metric." ||
		outcome.AssistantAction.RequiresAI {
		t.Fatalf("unexpected hint outcome: %+v", outcome.AssistantAction)
	}
	if outcome.AssistantAction.Provenance.PromptVersion != "hint.prompt.v1" ||
		outcome.AssistantAction.Provenance.RubricVersion != "not_applicable" ||
		outcome.AssistantAction.Provenance.ModelID != "stub-chat-1" ||
		outcome.AssistantAction.Provenance.DataSourceVersion != "registry.v1" {
		t.Fatalf("unexpected hint provenance: %+v", outcome.AssistantAction.Provenance)
	}
	if ai.payload.Metadata.TaskRun.Capability != aiclient.AITaskRunTaskHintGenerate ||
		ai.payload.Metadata.TaskRun.ResourceType != aiclient.AITaskRunResourceTargetJob ||
		ai.payload.Metadata.TaskRun.ResourceID != reservation.Plan.TargetJobID ||
		ai.payload.Metadata.FeatureKey != hintFeatureKey {
		t.Fatalf("unexpected task metadata: %+v", ai.payload.Metadata)
	}
	if len(ai.payload.Metadata.OutputSchema) == 0 {
		t.Fatalf("hint metadata OutputSchema must be populated")
	}
}

func TestParseHintAcceptsLightweightObserveCueSchema(t *testing.T) {
	hint, err := parseHint(`{"cue":"Anchor the answer in one measurable decision.","severity":"nudge","dimensionHint":"evidence"}`)
	if err != nil {
		t.Fatalf("parseHint returned error: %v", err)
	}
	if hint != "Anchor the answer in one measurable decision." {
		t.Fatalf("hint = %q", hint)
	}
}

func TestParseHintUsesCanonicalCue(t *testing.T) {
	t.Run("rejects alias-only output", func(t *testing.T) {
		if _, err := parseHint(`{"hint":"Alias hint."}`); err == nil {
			t.Fatal("expected alias-only hint output to fail")
		}
	})

	t.Run("canonical key wins over unknown alias", func(t *testing.T) {
		hint, err := parseHint(`{"cue":"Canonical cue.","hint":"Alias hint."}`)
		if err != nil {
			t.Fatalf("parseHint returned error: %v", err)
		}
		if hint != "Canonical cue." {
			t.Fatalf("hint = %q", hint)
		}
	})
}

func TestParseTurnObservationExtractsAnswerSummary(t *testing.T) {
	obs, err := parseTurnObservation(`{"cue":"Anchor the answer.","answerSummary":"Candidate covered queue sizing, rollback, and adoption metrics.","severity":"nudge"}`)
	if err != nil {
		t.Fatalf("parseTurnObservation returned error: %v", err)
	}
	if obs.Hint != "Anchor the answer." || obs.AnswerSummary != "Candidate covered queue sizing, rollback, and adoption metrics." {
		t.Fatalf("unexpected observation: %+v", obs)
	}
}

func TestParseTurnObservationUsesCanonicalAnswerSummary(t *testing.T) {
	t.Run("ignores alias-only summary", func(t *testing.T) {
		obs, err := parseTurnObservation(`{"cue":"Canonical cue.","answer_summary":"Alias summary."}`)
		if err != nil {
			t.Fatalf("parseTurnObservation returned error: %v", err)
		}
		if obs.Hint != "Canonical cue." || obs.AnswerSummary != "" {
			t.Fatalf("unexpected observation: %+v", obs)
		}
	})

	t.Run("canonical summary wins over unknown alias", func(t *testing.T) {
		obs, err := parseTurnObservation(`{"cue":"Canonical cue.","answerSummary":"Canonical summary.","answer_summary":"Alias summary."}`)
		if err != nil {
			t.Fatalf("parseTurnObservation returned error: %v", err)
		}
		if obs.Hint != "Canonical cue." || obs.AnswerSummary != "Canonical summary." {
			t.Fatalf("unexpected observation: %+v", obs)
		}
	})
}

func TestAppendSessionEventCompletedTurnPersistsObservationSummary(t *testing.T) {
	now := time.Date(2026, 5, 26, 19, 12, 0, 0, time.UTC)
	store := &recordingPlanStore{
		eventReservation: SessionEventReservation{
			UserID:  "user-1",
			Session: sessionEventTestSession(1),
			Plan: func() PlanRecord {
				plan := sessionEventTestPlan(1, sharedtypes.PracticeGoalBaseline)
				plan.TargetJobID = "target-1"
				return plan
			}(),
			LatestTurn: sessionEventTestTurn(1),
		},
	}
	service := NewService(ServiceOptions{
		Store: store,
		Registry: &fakePromptResolver{resolution: registry.PromptResolution{
			PromptVersion:       "observe.prompt.v1",
			RubricVersion:       "observe.rubric.v1",
			ModelProfileName:    "practice.turn_observe.default",
			FeatureFlag:         "none",
			DataSourceVersion:   "registry.v1",
			OutputSchema:        practiceOutputSchema(`{"type":"object","required":["cue","answerSummary"],"properties":{"cue":{"type":"string"},"answerSummary":{"type":"string"}}}`),
			UserMessageTemplate: "Question: {{question}}\nPartial answer: {{partial_answer}}\nRespond in {{language}}.",
		}},
		AI:    &fakeAIClient{content: `{"cue":"Metric cue.","answerSummary":"Candidate described the service boundary, tradeoffs, and rollback path."}`, store: store},
		Now:   func() time.Time { return now },
		NewID: sequenceIDs("event-1", "outbox-1"),
	})

	result, err := service.AppendSessionEvent(context.Background(), AppendSessionEventRequest{
		UserID:        "user-1",
		SessionID:     "session-1",
		ClientEventID: "client-event-1",
		Kind:          sessionEventKindAnswerSubmitted,
		OccurredAt:    now,
		Payload: map[string]any{
			"turnId":     "turn-1",
			"answerText": "I would isolate writes behind a queue, define rollback gates, and track adoption metrics.",
		},
	})
	if err != nil {
		t.Fatalf("AppendSessionEvent returned error: %v", err)
	}
	if result.AssistantAction.Type != assistantActionSessionCompleted {
		t.Fatalf("expected completed action, got %+v", result.AssistantAction)
	}
	if store.appendEvent.Outcome.AnswerSummary != "Candidate described the service boundary, tradeoffs, and rollback path." {
		t.Fatalf("answer summary was not persisted in outcome: %+v", store.appendEvent.Outcome)
	}
}

func TestApplyHintAIBuildsPromptFromF3Template(t *testing.T) {
	reservation := hintTestReservation()
	reservation.LatestTurn.QuestionText = "Tell me about a cross-team migration."
	ai := &fakeAIClient{content: `{"cue":"Use a metric.","severity":"nudge","dimensionHint":"evidence"}`}
	service := NewService(ServiceOptions{
		Registry: &fakePromptResolver{resolution: registry.PromptResolution{
			PromptVersion:       "hint.prompt.v1",
			RubricVersion:       "hint.rubric.v1",
			ModelProfileName:    "practice.turn_observe.default",
			FeatureFlag:         "none",
			DataSourceVersion:   "registry.v1",
			UserMessageTemplate: "Question: {{question}}\nPartial answer: {{partial_answer}}\nElapsed seconds: {{elapsed_seconds}}\nRespond in {{language}}.",
		}},
		AI: ai,
	})
	outcome := hintPendingOutcome(reservation)

	service.applyHintAI(context.Background(), reservation, map[string]any{"answerText": "I aligned 12 teams.", "elapsedSeconds": float64(42)}, outcome)

	rawPrompt := ""
	for _, msg := range ai.payload.Messages {
		rawPrompt += msg.Content + "\n"
	}
	for _, required := range []string{
		"Question: Tell me about a cross-team migration.",
		"Partial answer: I aligned 12 teams.",
		"Elapsed seconds: 42",
		"Respond in zh-CN.",
	} {
		if !strings.Contains(rawPrompt, required) {
			t.Fatalf("hint prompt missing %q: %s", required, rawPrompt)
		}
	}
	for _, forbidden := range []string{"{{question}}", "{{partial_answer}}", "{{elapsed_seconds}}", "{{language}}"} {
		if strings.Contains(rawPrompt, forbidden) {
			t.Fatalf("hint prompt left template marker %q: %s", forbidden, rawPrompt)
		}
	}
}

func TestApplyHintAIGracefulDegradeMatrix(t *testing.T) {
	cases := []struct {
		name     string
		resolver *fakePromptResolver
		ai       *fakeAIClient
		wantCode string
		wantRows int
	}{
		{
			name:     "f3 prompt unsupported",
			resolver: &fakePromptResolver{err: registry.ErrPromptUnsupported},
			ai:       &fakeAIClient{content: `{"cue":"ignored"}`},
			wantCode: sharederrors.CodeAiProviderConfigInvalid,
			wantRows: 1,
		},
		{
			name:     "f3 language unsupported",
			resolver: &fakePromptResolver{err: registry.ErrLanguageUnsupported},
			ai:       &fakeAIClient{content: `{"cue":"ignored"}`},
			wantCode: sharederrors.CodeAiProviderConfigInvalid,
			wantRows: 1,
		},
		{
			name:     "a3 secret missing",
			resolver: &fakePromptResolver{resolution: hintTestResolution()},
			ai:       &fakeAIClient{err: sharederrors.Wrap(sharederrors.CodeAiProviderSecretMissing, "missing secret", false)},
			wantCode: sharederrors.CodeAiProviderSecretMissing,
		},
		{
			name:     "a3 timeout",
			resolver: &fakePromptResolver{resolution: hintTestResolution()},
			ai:       &fakeAIClient{err: context.DeadlineExceeded},
			wantCode: sharederrors.CodeAiProviderTimeout,
		},
		{
			name:     "a3 invalid output",
			resolver: &fakePromptResolver{resolution: hintTestResolution()},
			ai:       &fakeAIClient{err: sharederrors.Wrap(sharederrors.CodeAiOutputInvalid, "schema failed", false)},
			wantCode: sharederrors.CodeAiOutputInvalid,
		},
		{
			name:     "a3 capability mismatch",
			resolver: &fakePromptResolver{resolution: hintTestResolution()},
			ai:       &fakeAIClient{err: sharederrors.Wrap(sharederrors.CodeAiUnsupportedCapability, "capability mismatch", false)},
			wantCode: sharederrors.CodeAiUnsupportedCapability,
		},
		{
			name:     "parsed hint empty",
			resolver: &fakePromptResolver{resolution: hintTestResolution()},
			ai:       &fakeAIClient{content: `{"cue":"   "}`},
			wantCode: sharederrors.CodeAiOutputInvalid,
			wantRows: 1,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			reservation := hintTestReservation()
			rows := &recordingAITaskRunWriter{}
			service := NewService(ServiceOptions{
				Registry:   tc.resolver,
				AI:         tc.ai,
				AITaskRuns: rows,
			})
			outcome := hintPendingOutcome(reservation)

			service.applyHintAI(context.Background(), reservation, map[string]any{}, outcome)

			if outcome.AssistantAction.Type != assistantActionSessionWait ||
				outcome.AssistantAction.Hint != "" ||
				outcome.NextSessionStatus != sharedtypes.SessionStatusRunning {
				t.Fatalf("expected graceful session_wait, got %+v", outcome)
			}
			if got := outcome.AuditMetadata["hint_degrade_reason"]; got != tc.wantCode {
				t.Fatalf("hint_degrade_reason = %v, want %s", got, tc.wantCode)
			}
			if len(rows.rows) != tc.wantRows {
				t.Fatalf("task run rows = %+v, want %d rows", rows.rows, tc.wantRows)
			}
			for _, row := range rows.rows {
				if row.Capability != aiclient.AITaskRunTaskHintGenerate ||
					row.ValidationStatus != aiclient.ValidationStatusInvalid ||
					row.ErrorCode != tc.wantCode {
					t.Fatalf("unexpected failed task run row: %+v", row)
				}
			}
		})
	}
}

func TestApplyHintAIPrivacyRedaction(t *testing.T) {
	reservation := hintTestReservation()
	reservation.LatestTurn.QuestionText = "question_text_secret"
	rows := &recordingAITaskRunWriter{}
	service := NewService(ServiceOptions{
		Registry:   &fakePromptResolver{err: registry.ErrPromptUnsupported},
		AI:         &fakeAIClient{content: `{"cue":"ignored response body secret"}`},
		AITaskRuns: rows,
	})
	outcome := hintPendingOutcome(reservation)

	service.applyHintAI(context.Background(), reservation, map[string]any{"answerText": "answer_text_secret"}, outcome)

	raw := mustMarshalString(t, map[string]any{
		"action":         outcome.AssistantAction,
		"auditMetadata":  outcome.AuditMetadata,
		"ai_task_runs":   rows.rows,
		"practice_event": "hint_requested",
	})
	for _, forbidden := range []string{"question_text_secret", "answer_text_secret", "hint_text", "response body secret", "provider secret"} {
		if strings.Contains(raw, forbidden) {
			t.Fatalf("hint degradation surface leaked %q: %s", forbidden, raw)
		}
	}
}

func TestApplyHintAIGracefulDegradeOnRegistryFailure(t *testing.T) {
	now := time.Date(2026, 4, 28, 13, 47, 32, 0, time.UTC)
	runs := &recordingAITaskRunWriter{}
	store := &recordingPlanStore{
		eventReservation: SessionEventReservation{
			UserID:  "01918fa0-0010-7a00-8a00-000000000001",
			Session: sessionEventTestSession(1),
			Plan: func() PlanRecord {
				plan := sessionEventTestPlan(3, sharedtypes.PracticeGoalBaseline)
				plan.Mode = sharedtypes.PracticeModeAssisted
				plan.TargetJobID = "01918fa0-0020-7a00-8a00-000000000002"
				return plan
			}(),
			LatestTurn: sessionEventTestTurn(1),
		},
	}
	service := NewService(ServiceOptions{
		Store:      store,
		Registry:   &fakePromptResolver{err: registry.ErrPromptUnsupported},
		AI:         &fakeAIClient{content: `{"cue":"ignored"}`, store: store},
		AITaskRuns: runs,
		Now:        func() time.Time { return now },
		NewID:      sequenceIDs("event-1", "outbox-1"),
	})

	result, err := service.AppendSessionEvent(context.Background(), AppendSessionEventRequest{
		UserID:        "01918fa0-0010-7a00-8a00-000000000001",
		SessionID:     "session-1",
		ClientEventID: "client-event-1",
		Kind:          "hint_requested",
		OccurredAt:    now,
		Payload:       map[string]any{"turnId": "turn-1"},
	})
	if err != nil {
		t.Fatalf("AppendSessionEvent returned error: %v", err)
	}
	if result.AssistantAction.Type != assistantActionSessionWait || result.AssistantAction.Hint != "" {
		t.Fatalf("expected session_wait degrade without hint, got %+v", result.AssistantAction)
	}
	if got := store.appendEvent.Outcome.AuditMetadata["hint_degrade_reason"]; got != sharederrors.CodeAiProviderConfigInvalid {
		t.Fatalf("degrade reason = %v", got)
	}
	if len(runs.rows) != 1 || runs.rows[0].Capability != aiclient.AITaskRunTaskHintGenerate ||
		runs.rows[0].ErrorCode != sharederrors.CodeAiProviderConfigInvalid {
		t.Fatalf("unexpected task run rows: %+v", runs.rows)
	}
}

func hintTestReservation() SessionEventReservation {
	return SessionEventReservation{
		UserID:  "01918fa0-0010-7a00-8a00-000000000001",
		Session: sessionEventTestSession(1),
		Plan: func() PlanRecord {
			plan := sessionEventTestPlan(3, sharedtypes.PracticeGoalBaseline)
			plan.Mode = sharedtypes.PracticeModeAssisted
			plan.TargetJobID = "01918fa0-0020-7a00-8a00-000000000002"
			return plan
		}(),
		LatestTurn: sessionEventTestTurn(1),
	}
}

func hintTestResolution() registry.PromptResolution {
	return registry.PromptResolution{
		PromptVersion:       "hint.prompt.v1",
		RubricVersion:       "hint.rubric.v1",
		ModelProfileName:    "practice.turn_observe.default",
		FeatureFlag:         "none",
		DataSourceVersion:   "registry.v1",
		OutputSchema:        practiceOutputSchema(`{"type":"object","required":["cue"],"properties":{"cue":{"type":"string"}}}`),
		UserMessageTemplate: "give a hint",
	}
}

func hintPendingOutcome(reservation SessionEventReservation) *SessionEventOutcome {
	return &SessionEventOutcome{
		Acknowledged:      true,
		NextSessionStatus: sharedtypes.SessionStatusRunning,
		AssistantAction: AssistantActionRecord{
			Type:          assistantActionShowHint,
			TurnID:        reservation.LatestTurn.ID,
			SessionStatus: sharedtypes.SessionStatusRunning,
			RequiresAI:    true,
			Provenance:    (SessionEventService{}).assistantAction(assistantActionShowHint, reservation.LatestTurn.ID, "", "", sharedtypes.SessionStatusRunning, reservation.Session.Language, true).Provenance,
		},
		AuditMetadata: map[string]any{"event_kind": sessionEventKindHintRequested, "mode": string(sharedtypes.PracticeModeAssisted)},
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
