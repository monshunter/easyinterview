package practice

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	stderrs "errors"
	"fmt"
	"strings"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient"
	"github.com/monshunter/easyinterview/backend/internal/ai/registry"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

const (
	voiceSTTFeatureKey = "practice.voice.stt"
	voiceTTSFeatureKey = "practice.voice.tts"
)

type CreatePracticeVoiceTurnRequest struct {
	UserID                   string
	SessionID                string
	ClientVoiceTurnID        string
	TurnID                   string
	Language                 string
	PracticeMode             sharedtypes.PracticeMode
	Audio                    PracticeVoiceAudioInput
	ManualTranscriptFallback string
	CommittedContext         CommittedVoiceContext
}

type PracticeVoiceAudioInput struct {
	Content     []byte
	ContentType string
	DurationMs  int32
	ByteLength  int32
}

type PracticeVoiceTurnResult struct {
	VoiceTurnID         string
	UserTranscriptFinal string
	AssistantTextDraft  string
	TTSChunks           []PracticeVoiceTTSChunk
	TTSError            *PracticeVoiceTTSError
	ProviderMetaSummary PracticeVoiceProviderMetaSummary
	Session             SessionRecord
}

type PracticeVoiceTTSChunk struct {
	ChunkID     string
	Sequence    int32
	ContentType string
	DurationMs  int32
	ByteLength  int32
	TextHash    string
	AudioRef    string
}

type PracticeVoiceProviderMetaSummary struct {
	STTProfile    string
	STTProvider   string
	STTLatencyMs  *int32
	ChatProfile   string
	ChatProvider  string
	ChatLatencyMs *int32
	TTSProfile    string
	TTSProvider   string
	TTSLatencyMs  *int32
}

type PracticeVoiceTTSError struct {
	Code      string
	Message   string
	Retryable bool
}

type PracticeVoiceTurnStoreInput struct {
	EventID             string
	UserID              string
	SessionID           string
	ClientVoiceTurnID   string
	TurnID              string
	VoiceTurnID         string
	UserTranscriptFinal string
	AssistantTextDraft  string
	AudioByteLength     int32
	AudioDurationMs     int32
	TTSChunks           []PracticeVoiceTTSChunk
	TTSError            *PracticeVoiceTTSError
	ProviderMetaSummary PracticeVoiceProviderMetaSummary
	Session             SessionRecord
	OccurredAt          time.Time
}

func (s *Service) CreatePracticeVoiceTurn(ctx context.Context, in CreatePracticeVoiceTurnRequest) (PracticeVoiceTurnResult, error) {
	if s == nil || s.store == nil {
		return PracticeVoiceTurnResult{}, fmt.Errorf("practice service is not initialised")
	}
	if s.registry == nil || s.ai == nil {
		return PracticeVoiceTurnResult{}, aiConfigError()
	}
	userID := strings.TrimSpace(in.UserID)
	sessionID := strings.TrimSpace(in.SessionID)
	clientVoiceTurnID := strings.TrimSpace(in.ClientVoiceTurnID)
	turnID := strings.TrimSpace(in.TurnID)
	if userID == "" {
		return PracticeVoiceTurnResult{}, fmt.Errorf("userId is required")
	}
	if sessionID == "" {
		return PracticeVoiceTurnResult{}, sessionNotFoundError()
	}
	if clientVoiceTurnID == "" {
		return PracticeVoiceTurnResult{}, validationError("clientVoiceTurnId is required", map[string]any{"field": "clientVoiceTurnId"})
	}
	if turnID == "" {
		return PracticeVoiceTurnResult{}, validationError("turnId is required", map[string]any{"field": "turnId"})
	}
	if !validPracticeMode(in.PracticeMode) {
		return PracticeVoiceTurnResult{}, validationError("practiceMode is invalid", map[string]any{"field": "practiceMode"})
	}
	if len(in.Audio.Content) == 0 && strings.TrimSpace(in.ManualTranscriptFallback) == "" {
		return PracticeVoiceTurnResult{}, validationError("audio.contentBase64 is required", map[string]any{"field": "audio.contentBase64"})
	}
	if strings.TrimSpace(in.Audio.ContentType) == "" && len(in.Audio.Content) > 0 {
		return PracticeVoiceTurnResult{}, validationError("audio.contentType is required", map[string]any{"field": "audio.contentType"})
	}
	if in.Audio.DurationMs < 0 {
		return PracticeVoiceTurnResult{}, validationError("audio.durationMs is invalid", map[string]any{"field": "audio.durationMs"})
	}

	session, err := s.store.GetSession(ctx, userID, sessionID)
	if stderrs.Is(err, ErrSessionNotFound) {
		return PracticeVoiceTurnResult{}, sessionNotFoundError()
	}
	if err != nil {
		return PracticeVoiceTurnResult{}, err
	}
	if isClosedSessionStatus(session.Status) {
		return PracticeVoiceTurnResult{}, sessionConflictError()
	}
	if session.CurrentTurn == nil || strings.TrimSpace(session.CurrentTurn.ID) != turnID {
		return PracticeVoiceTurnResult{}, sessionConflictError()
	}
	language := strings.TrimSpace(in.Language)
	if language == "" {
		language = strings.TrimSpace(session.Language)
	}
	if language == "" {
		language = "en"
	}

	sttRes, err := s.registry.ResolveActive(ctx, voiceSTTFeatureKey, language)
	if err != nil {
		return PracticeVoiceTurnResult{}, serviceErrorFromRegistry(err)
	}
	sttResp, sttMeta, err := s.ai.Transcribe(ctx, sttRes.ModelProfileName, voiceTranscriptionInput(sttRes, userID, session, language, in.Audio))
	if err != nil {
		return PracticeVoiceTurnResult{}, serviceErrorFromAI(err)
	}
	transcript := strings.TrimSpace(sttResp.Text)
	if transcript == "" {
		transcript = strings.TrimSpace(in.ManualTranscriptFallback)
	}
	if transcript == "" {
		return PracticeVoiceTurnResult{}, serviceErrorFromAI(sharederrors.Wrap(sharederrors.CodeAiOutputInvalid, "voice transcript is empty", false))
	}

	chatRes, err := s.registry.ResolveActive(ctx, followUpFeatureKey, language)
	if err != nil {
		return PracticeVoiceTurnResult{}, serviceErrorFromRegistry(err)
	}
	committedContext := in.CommittedContext
	if !committedContext.HasCommittedContext && strings.TrimSpace(committedContext.InterruptionNote) == "" {
		loaded, err := s.store.LoadCommittedVoiceContext(ctx, userID, sessionID)
		if err != nil {
			return PracticeVoiceTurnResult{}, err
		}
		committedContext = loaded
	}
	chatResp, chatMeta, err := s.ai.Complete(ctx, chatRes.ModelProfileName, voiceFollowUpPayload(chatRes, userID, session, language, in.PracticeMode, transcript, committedContext))
	if err != nil {
		return PracticeVoiceTurnResult{}, serviceErrorFromAI(err)
	}
	question, err := parseFirstQuestion(chatResp.Content)
	if err != nil {
		return PracticeVoiceTurnResult{}, serviceErrorFromAI(err)
	}
	assistantText := strings.TrimSpace(question.Text)
	if assistantText == "" {
		return PracticeVoiceTurnResult{}, serviceErrorFromAI(sharederrors.Wrap(sharederrors.CodeAiOutputInvalid, "voice assistant text is empty", false))
	}

	voiceTurnID := s.newID()
	result := PracticeVoiceTurnResult{
		VoiceTurnID:         voiceTurnID,
		UserTranscriptFinal: transcript,
		AssistantTextDraft:  assistantText,
		Session:             session,
	}

	ttsRes, err := s.registry.ResolveActive(ctx, voiceTTSFeatureKey, language)
	if err != nil {
		ttsErr, ok := practiceVoiceTTSErrorFromError(serviceErrorFromRegistry(err))
		if !ok {
			return PracticeVoiceTurnResult{}, err
		}
		result.TTSError = ttsErr
		result.ProviderMetaSummary = voiceProviderMetaSummary(sttMeta, chatMeta, aiclient.AICallMeta{}, sttRes, chatRes, registry.PromptResolution{})
		if err := s.recordPracticeVoiceTurn(ctx, in, &result); err != nil {
			return PracticeVoiceTurnResult{}, err
		}
		return result, nil
	}
	ttsResp, ttsMeta, err := s.ai.Synthesize(ctx, ttsRes.ModelProfileName, voiceSynthesisInput(ttsRes, userID, session, language, assistantText))
	if err != nil {
		ttsErr, ok := practiceVoiceTTSErrorFromError(serviceErrorFromAI(err))
		if !ok {
			return PracticeVoiceTurnResult{}, err
		}
		result.TTSError = ttsErr
		result.ProviderMetaSummary = voiceProviderMetaSummary(sttMeta, chatMeta, ttsMeta, sttRes, chatRes, ttsRes)
		if err := s.recordPracticeVoiceTurn(ctx, in, &result); err != nil {
			return PracticeVoiceTurnResult{}, err
		}
		return result, nil
	}

	result.TTSChunks = []PracticeVoiceTTSChunk{voiceTTSChunk(voiceTurnID, s.newID(), assistantText, ttsResp)}
	result.ProviderMetaSummary = voiceProviderMetaSummary(sttMeta, chatMeta, ttsMeta, sttRes, chatRes, ttsRes)
	if err := s.recordPracticeVoiceTurn(ctx, in, &result); err != nil {
		return PracticeVoiceTurnResult{}, err
	}
	return result, nil
}

func (s *Service) recordPracticeVoiceTurn(ctx context.Context, in CreatePracticeVoiceTurnRequest, result *PracticeVoiceTurnResult) error {
	if result == nil {
		return fmt.Errorf("practice voice turn result is required")
	}
	session, err := s.store.RecordPracticeVoiceTurn(ctx, PracticeVoiceTurnStoreInput{
		EventID:             s.newID(),
		UserID:              strings.TrimSpace(in.UserID),
		SessionID:           strings.TrimSpace(in.SessionID),
		ClientVoiceTurnID:   strings.TrimSpace(in.ClientVoiceTurnID),
		TurnID:              strings.TrimSpace(in.TurnID),
		VoiceTurnID:         result.VoiceTurnID,
		UserTranscriptFinal: result.UserTranscriptFinal,
		AssistantTextDraft:  result.AssistantTextDraft,
		AudioByteLength:     voiceAudioByteLength(in.Audio),
		AudioDurationMs:     in.Audio.DurationMs,
		TTSChunks:           practiceVoiceTTSChunksForStore(result.VoiceTurnID, result.TTSChunks),
		TTSError:            clonePracticeVoiceTTSError(result.TTSError),
		ProviderMetaSummary: result.ProviderMetaSummary,
		Session:             result.Session,
		OccurredAt:          s.now().UTC(),
	})
	if stderrs.Is(err, ErrSessionNotFound) {
		return sessionNotFoundError()
	}
	if stderrs.Is(err, ErrSessionConflict) {
		return sessionConflictError()
	}
	if err != nil {
		return err
	}
	result.Session = session
	return nil
}

func voiceTranscriptionInput(resolution registry.PromptResolution, userID string, session SessionRecord, language string, audio PracticeVoiceAudioInput) aiclient.TranscriptionInput {
	return aiclient.TranscriptionInput{
		Audio:       append([]byte(nil), audio.Content...),
		Filename:    "practice-voice-turn.webm",
		ContentType: strings.TrimSpace(audio.ContentType),
		Language:    language,
		Prompt:      strings.TrimSpace(resolution.UserMessageTemplate),
		Metadata:    voiceCallMetadata(resolution, userID, session, language, voiceSTTFeatureKey),
	}
}

func voiceFollowUpPayload(resolution registry.PromptResolution, userID string, session SessionRecord, language string, mode sharedtypes.PracticeMode, transcript string, committedContext CommittedVoiceContext) aiclient.CompletePayload {
	userContent := strings.TrimSpace(resolution.UserMessageTemplate)
	if userContent == "" {
		userContent = "Generate a concise follow-up question for the candidate's latest spoken answer."
	}
	replacer := strings.NewReplacer(
		"{{language}}", fallbackString(language, "en"),
		"{{practice_mode}}", fallbackString(string(mode), string(sharedtypes.PracticeModeAssisted)),
		"{{current_question}}", fallbackString(currentQuestionText(session), "current question unavailable"),
		"{{transcript}}", transcript,
	)
	userContent = strings.TrimSpace(replacer.Replace(userContent))
	if !strings.Contains(userContent, transcript) {
		userContent += "\nCandidate transcript:\n" + transcript
	}
	if committed := renderCommittedVoiceContext(committedContext); committed != "" {
		userContent += "\nCommitted assistant context:\n" + committed
	}
	messages := make([]aiclient.Message, 0, 2)
	if strings.TrimSpace(resolution.SystemMessage) != "" {
		messages = append(messages, aiclient.Message{Role: "system", Content: resolution.SystemMessage})
	}
	messages = append(messages, aiclient.Message{Role: "user", Content: userContent})
	return aiclient.CompletePayload{
		Messages: messages,
		Metadata: voiceCallMetadata(resolution, userID, session, language, followUpFeatureKey),
	}
}

func renderCommittedVoiceContext(ctx CommittedVoiceContext) string {
	parts := make([]string, 0, 2)
	if ctx.HasCommittedContext && strings.TrimSpace(ctx.CommittedAssistantText) != "" {
		parts = append(parts, "Previously heard assistant content: "+strings.TrimSpace(ctx.CommittedAssistantText))
	}
	if strings.TrimSpace(ctx.InterruptionNote) != "" {
		parts = append(parts, "Interruption note: "+strings.TrimSpace(ctx.InterruptionNote))
	}
	return strings.Join(parts, "\n")
}

func voiceSynthesisInput(resolution registry.PromptResolution, userID string, session SessionRecord, language string, assistantText string) aiclient.SynthesisInput {
	return aiclient.SynthesisInput{
		Text:     assistantText,
		Format:   "mp3",
		Language: language,
		Metadata: voiceCallMetadata(resolution, userID, session, language, voiceTTSFeatureKey),
	}
}

func voiceCallMetadata(resolution registry.PromptResolution, userID string, session SessionRecord, language string, featureKey string) aiclient.CallMetadata {
	return aiclient.CallMetadata{
		FeatureKey:        featureKey,
		PromptVersion:     resolution.PromptVersion,
		RubricVersion:     resolution.RubricVersion,
		Language:          language,
		FeatureFlag:       resolution.FeatureFlag,
		DataSourceVersion: resolution.DataSourceVersion,
		TaskRun: aiclient.AITaskRunContext{
			UserID:       userID,
			Capability:   aiclient.AITaskRunTaskFollowupGenerate,
			ResourceType: aiclient.AITaskRunResourceTargetJob,
			ResourceID:   session.TargetJobID,
		},
	}
}

func voiceTTSChunk(voiceTurnID, chunkID, assistantText string, resp aiclient.SynthesisResponse) PracticeVoiceTTSChunk {
	contentType := strings.TrimSpace(resp.ContentType)
	if contentType == "" {
		contentType = "audio/mpeg"
	}
	return PracticeVoiceTTSChunk{
		ChunkID:     chunkID,
		Sequence:    1,
		ContentType: contentType,
		DurationMs:  int32(resp.DurationMs),
		ByteLength:  int32(len(resp.Audio)),
		TextHash:    textSHA256(assistantText),
		AudioRef:    voiceAudioDataURL(contentType, resp.Audio),
	}
}

func voiceAudioDataURL(contentType string, audio []byte) string {
	if strings.TrimSpace(contentType) == "" {
		contentType = "audio/mpeg"
	}
	if len(audio) == 0 {
		return ""
	}
	return "data:" + strings.TrimSpace(contentType) + ";base64," + base64.StdEncoding.EncodeToString(audio)
}

func practiceVoiceTTSChunksForStore(voiceTurnID string, chunks []PracticeVoiceTTSChunk) []PracticeVoiceTTSChunk {
	out := make([]PracticeVoiceTTSChunk, 0, len(chunks))
	for _, chunk := range chunks {
		next := chunk
		next.AudioRef = "voice-turn://" + strings.TrimSpace(voiceTurnID) + "/chunks/" + strings.TrimSpace(chunk.ChunkID)
		out = append(out, next)
	}
	return out
}

func voiceProfileName(meta aiclient.AICallMeta, resolution registry.PromptResolution) string {
	if strings.TrimSpace(meta.ModelProfileName) != "" {
		return strings.TrimSpace(meta.ModelProfileName)
	}
	return strings.TrimSpace(resolution.ModelProfileName)
}

func voiceProviderName(meta aiclient.AICallMeta) string {
	if strings.TrimSpace(meta.Provider) != "" {
		return strings.TrimSpace(meta.Provider)
	}
	return "unknown"
}

func voiceProviderMetaSummary(
	sttMeta aiclient.AICallMeta,
	chatMeta aiclient.AICallMeta,
	ttsMeta aiclient.AICallMeta,
	sttRes registry.PromptResolution,
	chatRes registry.PromptResolution,
	ttsRes registry.PromptResolution,
) PracticeVoiceProviderMetaSummary {
	return PracticeVoiceProviderMetaSummary{
		STTProfile:    voiceProfileName(sttMeta, sttRes),
		STTProvider:   voiceProviderName(sttMeta),
		STTLatencyMs:  int32PtrFromInt64(sttMeta.LatencyMs),
		ChatProfile:   voiceProfileName(chatMeta, chatRes),
		ChatProvider:  voiceProviderName(chatMeta),
		ChatLatencyMs: int32PtrFromInt64(chatMeta.LatencyMs),
		TTSProfile:    voiceProfileName(ttsMeta, ttsRes),
		TTSProvider:   voiceProviderName(ttsMeta),
		TTSLatencyMs:  int32PtrFromInt64(ttsMeta.LatencyMs),
	}
}

func practiceVoiceTTSErrorFromError(err error) (*PracticeVoiceTTSError, bool) {
	var svcErr *ServiceError
	if !stderrs.As(err, &svcErr) {
		return nil, false
	}
	message := strings.TrimSpace(svcErr.Message)
	retryable := false
	if meta, ok := sharederrors.CodeRegistry[svcErr.Code]; ok {
		retryable = meta.Retryable
		if message == "" {
			message = meta.Message
		}
	}
	if message == "" {
		message = "TTS generation failed"
	}
	return &PracticeVoiceTTSError{
		Code:      svcErr.Code,
		Message:   message,
		Retryable: retryable,
	}, true
}

func voiceAudioByteLength(audio PracticeVoiceAudioInput) int32 {
	if audio.ByteLength > 0 {
		return audio.ByteLength
	}
	return int32(len(audio.Content))
}

func clonePracticeVoiceTTSError(in *PracticeVoiceTTSError) *PracticeVoiceTTSError {
	if in == nil {
		return nil
	}
	out := *in
	return &out
}

func int32PtrFromInt64(value int64) *int32 {
	if value <= 0 {
		return nil
	}
	converted := int32(value)
	return &converted
}

func currentQuestionText(session SessionRecord) string {
	if session.CurrentTurn == nil {
		return ""
	}
	return strings.TrimSpace(session.CurrentTurn.QuestionText)
}

func textSHA256(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}
