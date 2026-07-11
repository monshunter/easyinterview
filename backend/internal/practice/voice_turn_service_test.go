package practice

import (
	"context"
	"encoding/json"
	stderrs "errors"
	"reflect"
	"strings"
	"testing"

	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient"
	"github.com/monshunter/easyinterview/backend/internal/ai/registry"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

func TestCreatePracticeVoiceTurnRunsIndependentSTTChatTTSProfiles(t *testing.T) {
	session := SessionRecord{
		ID:          "session-1",
		PlanID:      "plan-1",
		TargetJobID: "target-1",
		Status:      sharedtypes.SessionStatusRunning,
		Language:    "zh-CN",
		TurnCount:   1,
		CurrentTurn: &TurnRecord{ID: "turn-1", TurnIndex: 1, QuestionText: "请介绍一次系统设计经历。", Status: string(TurnStatusAsked)},
	}
	ai := &voiceTurnAIClient{
		transcription: "我主导了一次订单系统拆分",
		chatContent:   firstQuestionJSON(t, "请继续说说你如何验证拆分结果。", "voice.follow_up"),
		synthesis:     []byte("tts-audio"),
		sttMeta:       aiclient.AICallMeta{Provider: "doubao-stt", ModelID: "stt-model", ModelProfileName: "practice.voice.stt.default", LatencyMs: 120},
		chatMeta:      aiclient.AICallMeta{Provider: "deepseek-chat", ModelID: "chat-model", ModelProfileName: "practice.followup.default", LatencyMs: 240},
		ttsMeta:       aiclient.AICallMeta{Provider: "minimax-tts", ModelID: "tts-model", ModelProfileName: "practice.voice.tts.default", LatencyMs: 180},
	}
	service := NewService(ServiceOptions{
		Store: &recordingPlanStore{
			getSessionRecord: session,
			getRecord: PlanRecord{
				ID: session.PlanID, TargetJobID: session.TargetJobID,
				Goal: sharedtypes.PracticeGoalBaseline, Mode: sharedtypes.PracticeModeAssisted, QuestionBudget: 3,
			},
		},
		Registry: &voiceTurnPromptResolver{resolutions: map[string]registry.PromptResolution{
			"practice.voice.stt": {
				FeatureKey:          "practice.voice.stt",
				PromptVersion:       "stt-prompt-v1",
				RubricVersion:       "not_applicable",
				ModelProfileName:    "practice.voice.stt.default",
				FeatureFlag:         "none",
				DataSourceVersion:   "registry.v1",
				UserMessageTemplate: "Transcribe the candidate answer.",
			},
			"practice.session.follow_up": {
				FeatureKey:          "practice.session.follow_up",
				PromptVersion:       "chat-prompt-v1",
				RubricVersion:       "rubric-v1",
				ModelProfileName:    "practice.followup.default",
				FeatureFlag:         "none",
				DataSourceVersion:   "registry.v1",
				SystemMessage:       "Generate strict JSON follow-up questions.",
				OutputSchema:        practiceOutputSchema(`{"type":"object","required":["questionText","questionIntent"],"properties":{"questionText":{"type":"string"},"questionIntent":{"type":"string"}}}`),
				UserMessageTemplate: questionTestResolution().UserMessageTemplate,
			},
			"practice.voice.tts": {
				FeatureKey:          "practice.voice.tts",
				PromptVersion:       "tts-prompt-v1",
				RubricVersion:       "not_applicable",
				ModelProfileName:    "practice.voice.tts.default",
				FeatureFlag:         "none",
				DataSourceVersion:   "registry.v1",
				UserMessageTemplate: "Speak the assistant reply.",
			},
		}},
		AI:    ai,
		NewID: sequenceIDs("voice-turn-1", "tts-chunk-1"),
	})

	result, err := service.CreatePracticeVoiceTurn(context.Background(), CreatePracticeVoiceTurnRequest{
		UserID:            "user-1",
		SessionID:         "session-1",
		ClientVoiceTurnID: "client-voice-turn-1",
		TurnID:            "turn-1",
		Language:          "zh-CN",
		PracticeMode:      sharedtypes.PracticeModeAssisted,
		Audio: PracticeVoiceAudioInput{
			Content:     []byte("tiny-audio"),
			ContentType: "audio/webm",
			DurationMs:  900,
			ByteLength:  10,
		},
	})
	if err != nil {
		t.Fatalf("CreatePracticeVoiceTurn: %v", err)
	}
	if got, want := ai.calls, []string{
		"transcribe:practice.voice.stt.default",
		"complete:practice.followup.default",
		"synthesize:practice.voice.tts.default",
	}; !reflect.DeepEqual(got, want) {
		t.Fatalf("AI call order/profile drift:\n got: %#v\nwant: %#v", got, want)
	}
	if result.VoiceTurnID != "voice-turn-1" ||
		result.UserTranscriptFinal != "我主导了一次订单系统拆分" ||
		result.AssistantTextDraft != "请继续说说你如何验证拆分结果。" {
		t.Fatalf("voice turn result drift: %+v", result)
	}
	if result.ProviderMetaSummary.STTProfile != "practice.voice.stt.default" ||
		result.ProviderMetaSummary.STTProvider != "doubao-stt" ||
		result.ProviderMetaSummary.ChatProfile != "practice.followup.default" ||
		result.ProviderMetaSummary.ChatProvider != "deepseek-chat" ||
		result.ProviderMetaSummary.TTSProfile != "practice.voice.tts.default" ||
		result.ProviderMetaSummary.TTSProvider != "minimax-tts" {
		t.Fatalf("provider meta summary did not preserve independent profile/provider triples: %+v", result.ProviderMetaSummary)
	}
	if len(result.TTSChunks) != 1 || result.TTSChunks[0].ChunkID != "tts-chunk-1" ||
		result.TTSChunks[0].ContentType != "audio/mpeg" || result.TTSChunks[0].ByteLength != int32(len(ai.synthesis)) ||
		result.TTSChunks[0].TextHash == "" {
		t.Fatalf("tts chunk metadata drift: %+v", result.TTSChunks)
	}
	if string(ai.transcribeInput.Audio) != "tiny-audio" ||
		ai.transcribeInput.Metadata.FeatureKey != "practice.voice.stt" ||
		ai.completePayload.Metadata.FeatureKey != "practice.session.follow_up" ||
		ai.synthesisInput.Metadata.FeatureKey != "practice.voice.tts" {
		t.Fatalf("AI metadata or payload drift: stt=%+v chat=%+v tts=%+v", ai.transcribeInput, ai.completePayload.Metadata, ai.synthesisInput)
	}
	if len(ai.transcribeInput.Metadata.OutputSchema) != 0 || len(ai.synthesisInput.Metadata.OutputSchema) != 0 {
		t.Fatalf("STT/TTS metadata must not carry OutputSchema: stt=%+v tts=%+v", ai.transcribeInput.Metadata, ai.synthesisInput.Metadata)
	}
	if len(ai.completePayload.Metadata.OutputSchema) == 0 {
		t.Fatalf("voice chat follow-up metadata OutputSchema must be populated")
	}
}

func TestCreatePracticeVoiceTurnStopsWhenSTTFails(t *testing.T) {
	ai := defaultVoiceTurnAIClient(t)
	ai.sttErr = sharederrors.Wrap(sharederrors.CodeAiProviderSecretMissing, "missing STT secret", false)
	service := newVoiceTurnTestService(t, ai)

	_, err := service.CreatePracticeVoiceTurn(context.Background(), validVoiceTurnRequest())
	if code := serviceErrorCode(t, err); code != sharederrors.CodeAiProviderSecretMissing {
		t.Fatalf("expected STT config error, got code=%s err=%v", code, err)
	}
	if got, want := ai.calls, []string{"transcribe:practice.voice.stt.default"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("STT failure must not call chat/TTS:\n got: %#v\nwant: %#v", got, want)
	}
}

func TestCreatePracticeVoiceTurnStopsWhenChatFailsBeforeTTS(t *testing.T) {
	ai := defaultVoiceTurnAIClient(t)
	ai.chatErr = sharederrors.Wrap(sharederrors.CodeAiFallbackExhausted, "chat fallback exhausted", true)
	service := newVoiceTurnTestService(t, ai)

	_, err := service.CreatePracticeVoiceTurn(context.Background(), validVoiceTurnRequest())
	if code := serviceErrorCode(t, err); code != sharederrors.CodeAiFallbackExhausted {
		t.Fatalf("expected chat failure, got code=%s err=%v", code, err)
	}
	if got, want := ai.calls, []string{
		"transcribe:practice.voice.stt.default",
		"complete:practice.followup.default",
	}; !reflect.DeepEqual(got, want) {
		t.Fatalf("chat failure must not call TTS:\n got: %#v\nwant: %#v", got, want)
	}
}

func TestCreatePracticeVoiceTurnUsesPersistedLanguageAndRepairsWrongLanguageBeforeTTS(t *testing.T) {
	ai := defaultVoiceTurnAIClient(t)
	ai.chatContents = []string{
		firstQuestionJSON(t, "How did you validate the split?", "voice.follow_up"),
		firstQuestionJSON(t, "你如何验证这次拆分的结果？", "voice.follow_up"),
	}
	store := &recordingPlanStore{getSessionRecord: voiceTurnSession()}
	service := newVoiceTurnTestServiceWithStore(t, store, ai)
	request := validVoiceTurnRequest()
	request.Language = "en-US"

	result, err := service.CreatePracticeVoiceTurn(context.Background(), request)
	if err != nil {
		t.Fatalf("CreatePracticeVoiceTurn: %v", err)
	}
	if result.AssistantTextDraft != "你如何验证这次拆分的结果？" {
		t.Fatalf("voice result did not use repaired session-language output: %+v", result)
	}
	if got, want := ai.calls, []string{
		"transcribe:practice.voice.stt.default",
		"complete:practice.followup.default",
		"complete:practice.followup.default",
		"synthesize:practice.voice.tts.default",
	}; !reflect.DeepEqual(got, want) {
		t.Fatalf("voice repair call order drift:\n got: %#v\nwant: %#v", got, want)
	}
	if len(ai.completePayloads) != 2 || ai.transcribeInput.Language != "zh-CN" || ai.synthesisInput.Language != "zh-CN" {
		t.Fatalf("persisted session language did not win: stt=%q chat=%+v tts=%q", ai.transcribeInput.Language, ai.completePayloads, ai.synthesisInput.Language)
	}
	initial := ai.completePayloads[0].Messages[len(ai.completePayloads[0].Messages)-1].Content
	repair := ai.completePayloads[1].Messages[len(ai.completePayloads[1].Messages)-1].Content
	if !strings.Contains(initial, "attempt=initial") || !strings.Contains(repair, "attempt=repair") ||
		!strings.Contains(initial, "language=zh-CN") || !strings.Contains(repair, "language=zh-CN") {
		t.Fatalf("voice repair markers drift: initial=%q repair=%q", initial, repair)
	}
}

func TestCreatePracticeVoiceTurnUsesServerPlanStatusAndStoredCommittedContext(t *testing.T) {
	ai := defaultVoiceTurnAIClient(t)
	session := voiceTurnSession()
	session.CurrentTurn.Status = string(TurnStatusFollowUpRequested)
	store := &recordingPlanStore{
		getSessionRecord: session,
		getRecord: PlanRecord{
			ID:             session.PlanID,
			TargetJobID:    session.TargetJobID,
			Goal:           sharedtypes.PracticeGoalRetryCurrentRound,
			Mode:           sharedtypes.PracticeModeStrict,
			QuestionBudget: 3,
		},
		committedContext: CommittedVoiceContext{
			HasCommittedContext:    true,
			CommittedAssistantText: "stored committed assistant context",
		},
	}
	service := newVoiceTurnTestServiceWithStore(t, store, ai)
	request := validVoiceTurnRequest()
	request.PracticeMode = sharedtypes.PracticeModeAssisted
	request.CommittedContext = CommittedVoiceContext{
		HasCommittedContext:    true,
		CommittedAssistantText: "forged request context",
	}

	if _, err := service.CreatePracticeVoiceTurn(context.Background(), request); err != nil {
		t.Fatalf("CreatePracticeVoiceTurn: %v", err)
	}
	userMessage := ai.completePayload.Messages[len(ai.completePayload.Messages)-1].Content
	for _, want := range []string{
		"goal=retry_current_round",
		"mode=strict",
		"status=follow_up_requested",
		"stored committed assistant context",
	} {
		if !strings.Contains(userMessage, want) {
			t.Fatalf("server-owned voice context missing %q: %s", want, userMessage)
		}
	}
	if strings.Contains(userMessage, "forged request context") {
		t.Fatalf("request committed context overrode store: %s", userMessage)
	}
	if !store.loadCommittedContextCalled || store.getPlanID != session.PlanID || store.getUserID != "user-1" {
		t.Fatalf("server context loads drifted: plan=%q user=%q committed=%v", store.getPlanID, store.getUserID, store.loadCommittedContextCalled)
	}
}

func TestCreatePracticeVoiceTurnPersistsGeneratedFollowUpOnSameTurn(t *testing.T) {
	ai := defaultVoiceTurnAIClient(t)
	store := &recordingPlanStore{getSessionRecord: voiceTurnSession()}
	service := newVoiceTurnTestServiceWithStore(t, store, ai)

	result, err := service.CreatePracticeVoiceTurn(context.Background(), validVoiceTurnRequest())
	if err != nil {
		t.Fatalf("CreatePracticeVoiceTurn: %v", err)
	}
	if got := store.voiceTurn.Outcome.AssistantAction.Type; got != assistantActionAskFollowUp {
		t.Fatalf("voice first answer action = %q, want %q", got, assistantActionAskFollowUp)
	}
	if next := store.voiceTurn.Outcome.NextTurn; next == nil ||
		next.ID != "turn-1" ||
		next.Status != string(TurnStatusFollowUpRequested) ||
		next.FollowUpCount != 1 {
		t.Fatalf("voice first answer did not reuse same-turn transition: %+v", next)
	}
	if action := store.voiceTurn.Outcome.AssistantAction; action.QuestionText != result.AssistantTextDraft || action.QuestionIntent != "voice.follow_up" {
		t.Fatalf("generated follow-up missing from server-owned outcome: %+v", action)
	}
	if store.voiceTurn.NextQuestion != nil || store.voiceTurn.Outcome.OutboxRecord != nil {
		t.Fatalf("first answer must not create a new turn/outbox: next=%+v outbox=%+v", store.voiceTurn.NextQuestion, store.voiceTurn.Outcome.OutboxRecord)
	}
	if result.Session.CurrentTurn == nil ||
		result.Session.CurrentTurn.ID != "turn-1" ||
		result.Session.CurrentTurn.QuestionText != result.AssistantTextDraft ||
		result.Session.CurrentTurn.QuestionIntent != "voice.follow_up" ||
		result.Session.CurrentTurn.FollowUpCount != 1 {
		t.Fatalf("result session did not expose persisted same-turn follow-up: %+v", result.Session.CurrentTurn)
	}
}

func TestCreatePracticeVoiceTurnPersistsAnswerObservationSummary(t *testing.T) {
	ai := defaultVoiceTurnAIClient(t)
	ai.observationContent = `{"cue":"","answerSummary":"候选人说明了拆分目标、验证方法和回滚边界。"}`
	session := voiceTurnSession()
	store := &recordingPlanStore{
		getSessionRecord: session,
		getRecord: PlanRecord{
			ID: session.PlanID, TargetJobID: session.TargetJobID,
			Goal: sharedtypes.PracticeGoalBaseline, Mode: sharedtypes.PracticeModeAssisted, QuestionBudget: 3,
		},
	}
	resolutions := voiceTurnTestResolutions()
	resolutions[hintFeatureKey] = registry.PromptResolution{
		FeatureKey:          hintFeatureKey,
		PromptVersion:       "observe-prompt-v1",
		RubricVersion:       "not_applicable",
		ModelProfileName:    "practice.turn.observe.default",
		FeatureFlag:         "none",
		DataSourceVersion:   "registry.v1",
		UserMessageTemplate: "question={{question}} answer={{partial_answer}} language={{language}}",
	}
	service := NewService(ServiceOptions{
		Store:    store,
		Registry: &voiceTurnPromptResolver{resolutions: resolutions},
		AI:       ai,
		NewID:    sequenceIDs("voice-turn-1", "tts-chunk-1", "voice-event-1"),
	})

	_, err := service.CreatePracticeVoiceTurn(context.Background(), validVoiceTurnRequest())
	if err != nil {
		t.Fatalf("CreatePracticeVoiceTurn: %v", err)
	}
	if got := store.voiceTurn.Outcome.AnswerSummary; got != "候选人说明了拆分目标、验证方法和回滚边界。" {
		t.Fatalf("voice answer observation summary was not persisted in the state-machine outcome: %q", got)
	}
	if got, want := ai.calls, []string{
		"transcribe:practice.voice.stt.default",
		"complete:practice.turn.observe.default",
		"complete:practice.followup.default",
		"synthesize:practice.voice.tts.default",
	}; !reflect.DeepEqual(got, want) {
		t.Fatalf("voice answer observation call order drift:\n got: %#v\nwant: %#v", got, want)
	}
}

func TestCreatePracticeVoiceTurnAfterFollowUpCreatesNextQuestionAndCompletesOldTurn(t *testing.T) {
	session := voiceTurnSession()
	session.CurrentTurn.Status = string(TurnStatusFollowUpRequested)
	session.CurrentTurn.FollowUpCount = 1
	session.CurrentTurn.QuestionText = "请说明你如何验证这次拆分的结果。"
	session.CurrentTurn.QuestionIntent = "evidence.follow_up"
	ai := defaultVoiceTurnAIClient(t)
	ai.chatContent = firstQuestionJSON(t, "请再介绍一个你主导的跨团队系统取舍案例。", "system_design.tradeoff")
	store := &recordingPlanStore{
		getSessionRecord: session,
		getRecord: PlanRecord{
			ID: session.PlanID, TargetJobID: session.TargetJobID,
			Goal: sharedtypes.PracticeGoalBaseline, Mode: sharedtypes.PracticeModeAssisted, QuestionBudget: 3,
		},
	}
	service := newVoiceTurnTestServiceWithStore(t, store, ai)

	result, err := service.CreatePracticeVoiceTurn(context.Background(), validVoiceTurnRequest())
	if err != nil {
		t.Fatalf("CreatePracticeVoiceTurn: %v", err)
	}
	if got := store.voiceTurn.Outcome.AssistantAction.Type; got != assistantActionAskQuestion {
		t.Fatalf("voice second answer action = %q, want %q", got, assistantActionAskQuestion)
	}
	if old := store.voiceTurn.Outcome.NextTurn; old == nil || old.ID != "turn-1" || old.Status != string(TurnStatusAssessed) || old.FollowUpCount != 1 {
		t.Fatalf("voice second answer did not assess old turn: %+v", old)
	}
	next := store.voiceTurn.NextQuestion
	if next == nil || next.ID == "turn-1" || next.TurnIndex != 2 ||
		next.QuestionText != result.AssistantTextDraft || next.QuestionIntent != "system_design.tradeoff" ||
		next.Status != string(TurnStatusAsked) {
		t.Fatalf("voice second answer did not create generated next turn: %+v", next)
	}
	if store.voiceTurn.Outcome.OutboxRecord == nil || store.voiceTurn.OutboxEventID == "" {
		t.Fatalf("voice second answer lost practice-turn-completed outbox: %+v", store.voiceTurn)
	}
	if result.Session.TurnCount != 2 || result.Session.CurrentTurn == nil || result.Session.CurrentTurn.ID != next.ID {
		t.Fatalf("voice second answer did not advance session current turn: %+v", result.Session)
	}
	prompt := ai.completePayload.Messages[len(ai.completePayload.Messages)-1].Content
	if !strings.Contains(prompt, "kind=next_question") || !strings.Contains(prompt, "question=请说明你如何验证这次拆分的结果。") {
		t.Fatalf("voice next-question prompt did not use persisted follow-up: %s", prompt)
	}
}

func TestCreatePracticeVoiceTurnAtQuestionBudgetCompletesWithoutQuestionOrTTS(t *testing.T) {
	session := voiceTurnSession()
	session.CurrentTurn.Status = string(TurnStatusFollowUpRequested)
	session.CurrentTurn.FollowUpCount = 1
	store := &recordingPlanStore{
		getSessionRecord: session,
		getRecord: PlanRecord{
			ID: session.PlanID, TargetJobID: session.TargetJobID,
			Goal: sharedtypes.PracticeGoalBaseline, Mode: sharedtypes.PracticeModeAssisted, QuestionBudget: 1,
		},
	}
	ai := defaultVoiceTurnAIClient(t)
	service := newVoiceTurnTestServiceWithStore(t, store, ai)

	result, err := service.CreatePracticeVoiceTurn(context.Background(), validVoiceTurnRequest())
	if err != nil {
		t.Fatalf("CreatePracticeVoiceTurn: %v", err)
	}
	if got, want := ai.calls, []string{"transcribe:practice.voice.stt.default"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("budget completion must not generate a canned question or TTS:\n got: %#v\nwant: %#v", got, want)
	}
	if result.AssistantTextDraft != "" || len(result.TTSChunks) != 0 || result.TTSError != nil {
		t.Fatalf("budget completion must return no assistant question/audio: %+v", result)
	}
	if result.Session.Status != sharedtypes.SessionStatusCompleted ||
		store.voiceTurn.Outcome.AssistantAction.Type != assistantActionSessionCompleted ||
		store.voiceTurn.NextQuestion != nil ||
		store.voiceTurn.Outcome.OutboxRecord == nil {
		t.Fatalf("budget completion did not reuse text transition: result=%+v store=%+v", result, store.voiceTurn)
	}
}

func TestCreatePracticeVoiceTurnSecondLanguageMismatchSkipsTTSAndPersistence(t *testing.T) {
	ai := defaultVoiceTurnAIClient(t)
	ai.chatContents = []string{
		firstQuestionJSON(t, "How did you validate the split?", "voice.follow_up"),
		firstQuestionJSON(t, "What tradeoff did you make?", "voice.tradeoff"),
	}
	store := &recordingPlanStore{getSessionRecord: voiceTurnSession()}
	service := newVoiceTurnTestServiceWithStore(t, store, ai)

	_, err := service.CreatePracticeVoiceTurn(context.Background(), validVoiceTurnRequest())
	if code := serviceErrorCode(t, err); code != sharederrors.CodeAiOutputInvalid {
		t.Fatalf("expected AI_OUTPUT_INVALID, got code=%s err=%v", code, err)
	}
	if got, want := ai.calls, []string{
		"transcribe:practice.voice.stt.default",
		"complete:practice.followup.default",
		"complete:practice.followup.default",
	}; !reflect.DeepEqual(got, want) {
		t.Fatalf("double-invalid voice turn must stop before TTS:\n got: %#v\nwant: %#v", got, want)
	}
	if store.voiceTurn.VoiceTurnID != "" {
		t.Fatalf("double-invalid voice turn must not persist a result: %+v", store.voiceTurn)
	}
}

func TestCreatePracticeVoiceTurnReturnsTranscriptAndAssistantTextWhenTTSFails(t *testing.T) {
	ai := defaultVoiceTurnAIClient(t)
	ai.ttsErr = sharederrors.Wrap(sharederrors.CodeAiProviderTimeout, "tts timeout", true)
	service := newVoiceTurnTestService(t, ai)

	result, err := service.CreatePracticeVoiceTurn(context.Background(), validVoiceTurnRequest())
	if err != nil {
		t.Fatalf("TTS failure should return a partial voice turn result, got err=%v", err)
	}
	if result.UserTranscriptFinal != ai.transcription || result.AssistantTextDraft != "请继续说说你如何验证拆分结果。" {
		t.Fatalf("TTS failure lost transcript/chat text: %+v", result)
	}
	if len(result.TTSChunks) != 0 {
		t.Fatalf("TTS failure must not return chunk metadata: %+v", result.TTSChunks)
	}
	if result.TTSError == nil ||
		result.TTSError.Code != sharederrors.CodeAiProviderTimeout ||
		!result.TTSError.Retryable {
		t.Fatalf("TTS failure must return structured retryable ttsError: %+v", result.TTSError)
	}
	if got, want := ai.calls, []string{
		"transcribe:practice.voice.stt.default",
		"complete:practice.followup.default",
		"synthesize:practice.voice.tts.default",
	}; !reflect.DeepEqual(got, want) {
		t.Fatalf("TTS failure call order/profile drift:\n got: %#v\nwant: %#v", got, want)
	}
}

func TestCreatePracticeVoiceTurnP0ProducerEmitsZeroOrOneTTSChunk(t *testing.T) {
	tests := []struct {
		name       string
		configure  func(*voiceTurnAIClient)
		wantChunks int
	}{
		{name: "successful synthesis emits exactly one chunk", wantChunks: 1},
		{
			name: "tts failure emits no chunks",
			configure: func(ai *voiceTurnAIClient) {
				ai.ttsErr = sharederrors.Wrap(sharederrors.CodeAiProviderTimeout, "tts timeout", true)
			},
			wantChunks: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ai := defaultVoiceTurnAIClient(t)
			if tt.configure != nil {
				tt.configure(ai)
			}
			result, err := newVoiceTurnTestService(t, ai).CreatePracticeVoiceTurn(context.Background(), validVoiceTurnRequest())
			if err != nil {
				t.Fatalf("CreatePracticeVoiceTurn: %v", err)
			}
			if got := len(result.TTSChunks); got != tt.wantChunks || got > 1 {
				t.Fatalf("P0 voice producer chunk count = %d, want %d and never more than one", got, tt.wantChunks)
			}
		})
	}
}

func TestCreatePracticeVoiceTurnPersistsBusinessTextOutsideAIMetadata(t *testing.T) {
	ai := defaultVoiceTurnAIClient(t)
	ai.transcription = "candidate transcript privacy-token"
	ai.chatContent = firstQuestionJSON(t, "请结合候选人的项目背景继续追问设计目标、执行过程、验证方式、风险取舍、结果指标和回滚方案，并把 assistant committed privacy-token 作为隐私测试标记保留在问题中。", "voice.follow_up")
	ai.synthesis = []byte("tts-audio-privacy-token")
	store := &recordingPlanStore{getSessionRecord: voiceTurnSession()}
	service := newVoiceTurnTestServiceWithStore(t, store, ai)

	result, err := service.CreatePracticeVoiceTurn(context.Background(), CreatePracticeVoiceTurnRequest{
		UserID:            "user-1",
		SessionID:         "session-1",
		ClientVoiceTurnID: "client-voice-turn-1",
		TurnID:            "turn-1",
		Language:          "zh-CN",
		PracticeMode:      sharedtypes.PracticeModeAssisted,
		Audio: PracticeVoiceAudioInput{
			Content:     []byte("raw-audio-privacy-token"),
			ContentType: "audio/webm",
			DurationMs:  900,
			ByteLength:  23,
		},
	})
	if err != nil {
		t.Fatalf("CreatePracticeVoiceTurn: %v", err)
	}
	if store.voiceTurn.UserTranscriptFinal != ai.transcription ||
		store.voiceTurn.AssistantTextDraft != result.AssistantTextDraft {
		t.Fatalf("business transcript/text must persist through explicit voice turn event input: %+v", store.voiceTurn)
	}
	if store.voiceTurn.AudioByteLength != 23 ||
		store.voiceTurn.AudioDurationMs != 900 ||
		len(store.voiceTurn.TTSChunks) != 1 ||
		store.voiceTurn.TTSChunks[0].TextHash == "" {
		t.Fatalf("voice turn event summary should keep only audio/TTS metadata: %+v", store.voiceTurn)
	}
	safeMetadata := mustMarshalJSON(t, map[string]any{
		"stt_metadata":          ai.transcribeInput.Metadata,
		"chat_metadata":         ai.completePayload.Metadata,
		"tts_metadata":          ai.synthesisInput.Metadata,
		"provider_meta_summary": result.ProviderMetaSummary,
		"tts_chunks":            result.TTSChunks,
		"store_provider_meta":   store.voiceTurn.ProviderMetaSummary,
		"store_tts_chunks":      store.voiceTurn.TTSChunks,
		"store_audio_length":    store.voiceTurn.AudioByteLength,
		"store_audio_duration":  store.voiceTurn.AudioDurationMs,
	})
	for _, forbidden := range []string{
		"raw-audio-privacy-token",
		"tts-audio-privacy-token",
		"candidate transcript privacy-token",
		"assistant committed privacy-token",
	} {
		if strings.Contains(safeMetadata, forbidden) {
			t.Fatalf("AI metadata or summary leaked plaintext %q: %s", forbidden, safeMetadata)
		}
	}
}

func TestCreatePracticeVoiceTurnReturnsPlayableAudioRefWithoutPersistingAudioData(t *testing.T) {
	ai := defaultVoiceTurnAIClient(t)
	ai.synthesis = []byte("tts-audio")
	store := &recordingPlanStore{getSessionRecord: voiceTurnSession()}
	service := newVoiceTurnTestServiceWithStore(t, store, ai)

	result, err := service.CreatePracticeVoiceTurn(context.Background(), validVoiceTurnRequest())
	if err != nil {
		t.Fatalf("CreatePracticeVoiceTurn: %v", err)
	}
	if len(result.TTSChunks) != 1 {
		t.Fatalf("expected a playable TTS chunk, got %+v", result.TTSChunks)
	}
	if got := result.TTSChunks[0].AudioRef; !strings.HasPrefix(got, "data:audio/mpeg;base64,") {
		t.Fatalf("response audioRef must be browser-playable, got %q", got)
	}
	if len(store.voiceTurn.TTSChunks) != 1 {
		t.Fatalf("expected persisted chunk summary, got %+v", store.voiceTurn.TTSChunks)
	}
	if got := store.voiceTurn.TTSChunks[0].AudioRef; !strings.HasPrefix(got, "voice-turn://voice-turn-1/chunks/tts-chunk-1") {
		t.Fatalf("stored audioRef must remain opaque and non-audio-bearing, got %q", got)
	}
}

func TestCreatePracticeVoiceTurnLoadsCommittedContextFromStoredPlaybackEvents(t *testing.T) {
	ai := defaultVoiceTurnAIClient(t)
	store := &recordingPlanStore{
		getSessionRecord: voiceTurnSession(),
		committedContext: CommittedVoiceContext{
			VoiceTurnID:            "voice-turn-previous",
			HasCommittedContext:    true,
			CommittedAssistantText: "previous assistant words heard by user",
			CommittedTextLength:    38,
			Interrupted:            true,
			InterruptionNote:       "Assistant playback was interrupted at 1480ms.",
		},
	}
	service := newVoiceTurnTestServiceWithStore(t, store, ai)

	_, err := service.CreatePracticeVoiceTurn(context.Background(), validVoiceTurnRequest())
	if err != nil {
		t.Fatalf("CreatePracticeVoiceTurn: %v", err)
	}

	userMessage := ai.completePayload.Messages[len(ai.completePayload.Messages)-1].Content
	if !store.loadCommittedContextCalled {
		t.Fatalf("service did not load committed context from stored playback events")
	}
	if !strings.Contains(userMessage, "previous assistant words heard by user") ||
		!strings.Contains(userMessage, "Assistant playback was interrupted at 1480ms.") {
		t.Fatalf("stored committed context missing from prompt: %s", userMessage)
	}
}

func TestVoiceQuestionTemplateInjectsCommittedContextWithoutUnplayedDraft(t *testing.T) {
	committed := CommittedVoiceContext{
		VoiceTurnID:            "voice-turn-previous",
		HasCommittedContext:    true,
		CommittedAssistantText: "played assistant content",
		CommittedTextLength:    24,
		Interrupted:            true,
		InterruptionNote:       "Assistant playback was interrupted at 1480ms.",
	}
	userMessage, err := renderQuestionTemplate(questionTestResolution().UserMessageTemplate, questionTemplateData{
		Language:         "zh-CN",
		PracticeMode:     string(sharedtypes.PracticeModeAssisted),
		TargetJobID:      "target-1",
		GenerationKind:   questionGenerationFollowUp,
		AttemptMode:      questionAttemptInitial,
		LastQuestion:     "请介绍一次系统设计经历。",
		QuestionIntent:   "system_design",
		LastAnswer:       "new user answer",
		CommittedContext: renderCommittedVoiceContext(committed),
	})
	if err != nil {
		t.Fatalf("renderQuestionTemplate: %v", err)
	}
	if !strings.Contains(userMessage, "played assistant content") ||
		!strings.Contains(userMessage, "Assistant playback was interrupted at 1480ms.") {
		t.Fatalf("committed context/interruption note missing from prompt: %s", userMessage)
	}
	if strings.Contains(userMessage, "unplayed assistant draft") {
		t.Fatalf("unplayed assistant draft leaked into prompt: %s", userMessage)
	}
}

func validVoiceTurnRequest() CreatePracticeVoiceTurnRequest {
	return CreatePracticeVoiceTurnRequest{
		UserID:            "user-1",
		SessionID:         "session-1",
		ClientVoiceTurnID: "client-voice-turn-1",
		TurnID:            "turn-1",
		Language:          "zh-CN",
		PracticeMode:      sharedtypes.PracticeModeAssisted,
		Audio: PracticeVoiceAudioInput{
			Content:     []byte("tiny-audio"),
			ContentType: "audio/webm",
			DurationMs:  900,
			ByteLength:  10,
		},
	}
}

func defaultVoiceTurnAIClient(t *testing.T) *voiceTurnAIClient {
	t.Helper()
	return &voiceTurnAIClient{
		transcription: "我主导了一次订单系统拆分",
		chatContent:   firstQuestionJSON(t, "请继续说说你如何验证拆分结果。", "voice.follow_up"),
		synthesis:     []byte("tts-audio"),
		sttMeta:       aiclient.AICallMeta{Provider: "doubao-stt", ModelID: "stt-model", ModelProfileName: "practice.voice.stt.default", LatencyMs: 120},
		chatMeta:      aiclient.AICallMeta{Provider: "deepseek-chat", ModelID: "chat-model", ModelProfileName: "practice.followup.default", LatencyMs: 240},
		ttsMeta:       aiclient.AICallMeta{Provider: "minimax-tts", ModelID: "tts-model", ModelProfileName: "practice.voice.tts.default", LatencyMs: 180},
	}
}

func newVoiceTurnTestService(t *testing.T, ai *voiceTurnAIClient) *Service {
	t.Helper()
	return newVoiceTurnTestServiceWithStore(t, &recordingPlanStore{getSessionRecord: voiceTurnSession()}, ai)
}

func newVoiceTurnTestServiceWithStore(t *testing.T, store *recordingPlanStore, ai *voiceTurnAIClient) *Service {
	t.Helper()
	if store.getRecord.ID == "" {
		session := store.getSessionRecord
		store.getRecord = PlanRecord{
			ID:             session.PlanID,
			TargetJobID:    session.TargetJobID,
			Goal:           sharedtypes.PracticeGoalBaseline,
			Mode:           sharedtypes.PracticeModeAssisted,
			QuestionBudget: 3,
		}
	}
	return NewService(ServiceOptions{
		Store:    store,
		Registry: &voiceTurnPromptResolver{resolutions: voiceTurnTestResolutions()},
		AI:       ai,
		NewID:    sequenceIDs("voice-turn-1", "tts-chunk-1", "voice-event-1"),
	})
}

func voiceTurnTestResolutions() map[string]registry.PromptResolution {
	return map[string]registry.PromptResolution{
		"practice.voice.stt": {
			FeatureKey:          "practice.voice.stt",
			PromptVersion:       "stt-prompt-v1",
			RubricVersion:       "not_applicable",
			ModelProfileName:    "practice.voice.stt.default",
			FeatureFlag:         "none",
			DataSourceVersion:   "registry.v1",
			UserMessageTemplate: "Transcribe the candidate answer.",
		},
		"practice.session.follow_up": {
			FeatureKey:          "practice.session.follow_up",
			PromptVersion:       "chat-prompt-v1",
			RubricVersion:       "rubric-v1",
			ModelProfileName:    "practice.followup.default",
			FeatureFlag:         "none",
			DataSourceVersion:   "registry.v1",
			SystemMessage:       "Generate strict JSON follow-up questions.",
			OutputSchema:        practiceOutputSchema(`{"type":"object","required":["questionText","questionIntent"],"properties":{"questionText":{"type":"string"},"questionIntent":{"type":"string"}}}`),
			UserMessageTemplate: questionTestResolution().UserMessageTemplate,
		},
		"practice.voice.tts": {
			FeatureKey:          "practice.voice.tts",
			PromptVersion:       "tts-prompt-v1",
			RubricVersion:       "not_applicable",
			ModelProfileName:    "practice.voice.tts.default",
			FeatureFlag:         "none",
			DataSourceVersion:   "registry.v1",
			UserMessageTemplate: "Speak the assistant reply.",
		},
	}
}

func voiceTurnSession() SessionRecord {
	return SessionRecord{
		ID:          "session-1",
		PlanID:      "plan-1",
		TargetJobID: "target-1",
		Status:      sharedtypes.SessionStatusRunning,
		Language:    "zh-CN",
		TurnCount:   1,
		CurrentTurn: &TurnRecord{ID: "turn-1", TurnIndex: 1, QuestionText: "请介绍一次系统设计经历。", QuestionIntent: "system_design", Status: string(TurnStatusAsked)},
	}
}

func serviceErrorCode(t *testing.T, err error) string {
	t.Helper()
	var svcErr *ServiceError
	if !stderrs.As(err, &svcErr) {
		t.Fatalf("expected ServiceError, got %T %v", err, err)
	}
	return svcErr.Code
}

func mustMarshalJSON(t *testing.T, value any) string {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal json: %v", err)
	}
	return string(raw)
}

type voiceTurnPromptResolver struct {
	resolutions map[string]registry.PromptResolution
	requests    []string
}

func (r *voiceTurnPromptResolver) ResolveActive(ctx context.Context, featureKey, language string) (registry.PromptResolution, error) {
	r.requests = append(r.requests, featureKey+":"+language)
	if res, ok := r.resolutions[featureKey]; ok {
		res.FeatureKey = featureKey
		return res, nil
	}
	return registry.PromptResolution{}, registry.ErrPromptUnsupported
}

type voiceTurnAIClient struct {
	calls []string

	transcription      string
	chatContent        string
	chatContents       []string
	observationContent string
	synthesis          []byte

	sttMeta  aiclient.AICallMeta
	chatMeta aiclient.AICallMeta
	ttsMeta  aiclient.AICallMeta
	sttErr   error
	chatErr  error
	chatErrs []error
	ttsErr   error

	transcribeInput  aiclient.TranscriptionInput
	completePayload  aiclient.CompletePayload
	completePayloads []aiclient.CompletePayload
	synthesisInput   aiclient.SynthesisInput
}

func (c *voiceTurnAIClient) Complete(ctx context.Context, profileName string, payload aiclient.CompletePayload) (aiclient.CompleteResponse, aiclient.AICallMeta, error) {
	c.calls = append(c.calls, "complete:"+profileName)
	c.completePayload = payload
	c.completePayloads = append(c.completePayloads, payload)
	if profileName == "practice.turn.observe.default" {
		return aiclient.CompleteResponse{Content: c.observationContent}, c.chatMeta, nil
	}
	if len(c.chatErrs) > 0 {
		err := c.chatErrs[0]
		c.chatErrs = c.chatErrs[1:]
		if err != nil {
			return aiclient.CompleteResponse{}, c.chatMeta, err
		}
	}
	if c.chatErr != nil {
		return aiclient.CompleteResponse{}, c.chatMeta, c.chatErr
	}
	content := c.chatContent
	if len(c.chatContents) > 0 {
		content = c.chatContents[0]
		c.chatContents = c.chatContents[1:]
	}
	return aiclient.CompleteResponse{Content: content}, c.chatMeta, nil
}

func (c *voiceTurnAIClient) Transcribe(ctx context.Context, profileName string, payload aiclient.TranscriptionInput) (aiclient.TranscriptionResponse, aiclient.AICallMeta, error) {
	c.calls = append(c.calls, "transcribe:"+profileName)
	c.transcribeInput = payload
	if c.sttErr != nil {
		return aiclient.TranscriptionResponse{}, c.sttMeta, c.sttErr
	}
	return aiclient.TranscriptionResponse{Text: c.transcription}, c.sttMeta, nil
}

func (c *voiceTurnAIClient) Stream(ctx context.Context, profileName string, payload aiclient.CompletePayload) (<-chan aiclient.AIStreamEvent, error) {
	return nil, nil
}

func (c *voiceTurnAIClient) Synthesize(ctx context.Context, profileName string, input aiclient.SynthesisInput) (aiclient.SynthesisResponse, aiclient.AICallMeta, error) {
	c.calls = append(c.calls, "synthesize:"+profileName)
	c.synthesisInput = input
	if c.ttsErr != nil {
		return aiclient.SynthesisResponse{}, c.ttsMeta, c.ttsErr
	}
	return aiclient.SynthesisResponse{Audio: c.synthesis, ContentType: "audio/mpeg", DurationMs: 880, CharCount: len([]rune(input.Text))}, c.ttsMeta, nil
}
