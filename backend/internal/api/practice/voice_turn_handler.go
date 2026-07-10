package practice

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"

	api "github.com/monshunter/easyinterview/backend/internal/api/generated"
	"github.com/monshunter/easyinterview/backend/internal/middleware/idempotency"
	domain "github.com/monshunter/easyinterview/backend/internal/practice"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
)

func (h *Handler) CreatePracticeVoiceTurn(w http.ResponseWriter, r *http.Request, sessionID string) {
	if h == nil || h.service == nil {
		writeAPIError(w, http.StatusInternalServerError, sharederrors.CodeValidationFailed, "practice service is not configured", nil)
		return
	}
	userID, ok := h.resolveUser(r)
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, sharederrors.CodeAuthUnauthorized, "authentication required", nil)
		return
	}
	var body api.CreatePracticeVoiceTurnRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeAPIError(w, http.StatusBadRequest, sharederrors.CodeValidationFailed, "request body is malformed", nil)
		return
	}
	audioContent, err := decodePracticeVoiceAudio(body.Audio.ContentBase64)
	if err != nil {
		writeAPIError(w, http.StatusUnprocessableEntity, sharederrors.CodeValidationFailed, "audio.contentBase64 must be valid base64", map[string]any{"field": "audio.contentBase64"})
		return
	}
	byteLength := int32(len(audioContent))
	if body.Audio.ByteLength != nil {
		byteLength = *body.Audio.ByteLength
	}
	result, err := h.service.CreatePracticeVoiceTurn(r.Context(), domain.CreatePracticeVoiceTurnRequest{
		UserID:                   userID,
		SessionID:                sessionID,
		ClientVoiceTurnID:        body.ClientVoiceTurnId,
		TurnID:                   body.TurnId,
		Language:                 body.Language,
		PracticeMode:             body.PracticeMode,
		Audio: domain.PracticeVoiceAudioInput{
			Content:     audioContent,
			ContentType: body.Audio.ContentType,
			DurationMs:  body.Audio.DurationMs,
			ByteLength:  byteLength,
		},
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	idempotency.SetResponseResource(w, "practice_voice_turn", result.VoiceTurnID)
	writeJSON(w, http.StatusOK, toAPIPracticeVoiceTurnResult(result))
}

func decodePracticeVoiceAudio(raw string) ([]byte, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	return base64.StdEncoding.DecodeString(raw)
}

func toAPIPracticeVoiceTurnResult(result domain.PracticeVoiceTurnResult) api.PracticeVoiceTurnResult {
	chunks := make([]api.PracticeVoiceTTSChunk, 0, len(result.TTSChunks))
	for _, chunk := range result.TTSChunks {
		chunks = append(chunks, api.PracticeVoiceTTSChunk{
			ChunkId:     chunk.ChunkID,
			Sequence:    chunk.Sequence,
			ContentType: chunk.ContentType,
			DurationMs:  chunk.DurationMs,
			ByteLength:  chunk.ByteLength,
			TextHash:    chunk.TextHash,
			AudioRef:    chunk.AudioRef,
		})
	}
	return api.PracticeVoiceTurnResult{
		VoiceTurnId:         result.VoiceTurnID,
		UserTranscriptFinal: result.UserTranscriptFinal,
		AssistantTextDraft:  result.AssistantTextDraft,
		TtsChunks:           chunks,
		ProviderMetaSummary: api.PracticeVoiceProviderMetaSummary{
			SttProfile:    result.ProviderMetaSummary.STTProfile,
			SttProvider:   result.ProviderMetaSummary.STTProvider,
			SttLatencyMs:  result.ProviderMetaSummary.STTLatencyMs,
			ChatProfile:   result.ProviderMetaSummary.ChatProfile,
			ChatProvider:  result.ProviderMetaSummary.ChatProvider,
			ChatLatencyMs: result.ProviderMetaSummary.ChatLatencyMs,
			TtsProfile:    result.ProviderMetaSummary.TTSProfile,
			TtsProvider:   result.ProviderMetaSummary.TTSProvider,
			TtsLatencyMs:  result.ProviderMetaSummary.TTSLatencyMs,
		},
		Session:  toAPIPracticeSession(result.Session),
		TtsError: toAPIPracticeVoiceTTSError(result.TTSError),
	}
}

func toAPIPracticeVoiceTTSError(ttsErr *domain.PracticeVoiceTTSError) *api.PracticeVoiceTTSError {
	if ttsErr == nil {
		return nil
	}
	return &api.PracticeVoiceTTSError{
		Code:      ttsErr.Code,
		Message:   ttsErr.Message,
		Retryable: ttsErr.Retryable,
	}
}
