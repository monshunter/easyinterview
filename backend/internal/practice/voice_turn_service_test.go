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
		Store: &recordingPlanStore{getSessionRecord: session},
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
				UserMessageTemplate: "Generate a concise follow-up.",
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

func TestCreatePracticeVoiceTurnPersistsBusinessTextOutsideAIMetadata(t *testing.T) {
	ai := defaultVoiceTurnAIClient(t)
	ai.transcription = "candidate transcript privacy-token"
	ai.chatContent = firstQuestionJSON(t, "assistant committed privacy-token", "voice.follow_up")
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

func TestVoiceFollowUpPayloadInjectsCommittedContextWithoutUnplayedDraft(t *testing.T) {
	resolution := registry.PromptResolution{
		FeatureKey:          followUpFeatureKey,
		PromptVersion:       "chat-prompt-v1",
		RubricVersion:       "rubric-v1",
		ModelProfileName:    "practice.followup.default",
		FeatureFlag:         "none",
		DataSourceVersion:   "registry.v1",
		UserMessageTemplate: "Generate a concise follow-up for {{transcript}}.",
	}
	payload := voiceFollowUpPayload(
		resolution,
		"user-1",
		voiceTurnSession(),
		"zh-CN",
		sharedtypes.PracticeModeAssisted,
		"new user answer",
		CommittedVoiceContext{
			VoiceTurnID:            "voice-turn-previous",
			HasCommittedContext:    true,
			CommittedAssistantText: "played assistant content",
			CommittedTextLength:    24,
			Interrupted:            true,
			InterruptionNote:       "Assistant playback was interrupted at 1480ms.",
		},
	)
	userMessage := payload.Messages[len(payload.Messages)-1].Content
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
	return NewService(ServiceOptions{
		Store: store,
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
				UserMessageTemplate: "Generate a concise follow-up.",
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
		NewID: sequenceIDs("voice-turn-1", "tts-chunk-1", "voice-event-1"),
	})
}

func voiceTurnSession() SessionRecord {
	return SessionRecord{
		ID:          "session-1",
		PlanID:      "plan-1",
		TargetJobID: "target-1",
		Status:      sharedtypes.SessionStatusRunning,
		Language:    "zh-CN",
		TurnCount:   1,
		CurrentTurn: &TurnRecord{ID: "turn-1", TurnIndex: 1, QuestionText: "请介绍一次系统设计经历。", Status: string(TurnStatusAsked)},
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

	transcription string
	chatContent   string
	synthesis     []byte

	sttMeta  aiclient.AICallMeta
	chatMeta aiclient.AICallMeta
	ttsMeta  aiclient.AICallMeta
	sttErr   error
	chatErr  error
	ttsErr   error

	transcribeInput aiclient.TranscriptionInput
	completePayload aiclient.CompletePayload
	synthesisInput  aiclient.SynthesisInput
}

func (c *voiceTurnAIClient) Complete(ctx context.Context, profileName string, payload aiclient.CompletePayload) (aiclient.CompleteResponse, aiclient.AICallMeta, error) {
	c.calls = append(c.calls, "complete:"+profileName)
	c.completePayload = payload
	if c.chatErr != nil {
		return aiclient.CompleteResponse{}, c.chatMeta, c.chatErr
	}
	return aiclient.CompleteResponse{Content: c.chatContent}, c.chatMeta, nil
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
