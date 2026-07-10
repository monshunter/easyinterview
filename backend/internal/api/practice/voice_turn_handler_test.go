package practice

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	api "github.com/monshunter/easyinterview/backend/internal/api/generated"
	"github.com/monshunter/easyinterview/backend/internal/middleware/idempotency"
	domain "github.com/monshunter/easyinterview/backend/internal/practice"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

func TestCreatePracticeVoiceTurnReturns200AndMapsRequest(t *testing.T) {
	service := &fakePlanService{voiceResult: fixturePracticeVoiceTurnResult()}
	handler := newTestHandler(service)

	rec := httptest.NewRecorder()
	handler.CreatePracticeVoiceTurn(rec, newPracticeVoiceTurnHTTPRequest(t, fixtureCreatePracticeVoiceTurnRequest()), "session-voice-1")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if service.voiceCalls != 1 {
		t.Fatalf("voice calls = %d, want 1", service.voiceCalls)
	}
	if service.voiceRequest.UserID != "user-1" ||
		service.voiceRequest.SessionID != "session-voice-1" ||
		service.voiceRequest.ClientVoiceTurnID != "client-voice-turn-1" ||
		service.voiceRequest.TurnID != "turn-1" ||
		service.voiceRequest.Language != "zh-CN" ||
		service.voiceRequest.PracticeMode != sharedtypes.PracticeModeAssisted {
		t.Fatalf("request not mapped to service: %+v", service.voiceRequest)
	}
	if string(service.voiceRequest.Audio.Content) != "OggS" ||
		service.voiceRequest.Audio.ContentType != "audio/webm" ||
		service.voiceRequest.Audio.DurationMs != 4320 ||
		service.voiceRequest.Audio.ByteLength != 128 {
		t.Fatalf("audio request not mapped to service: %+v content=%q", service.voiceRequest.Audio, service.voiceRequest.Audio.Content)
	}

	var out api.PracticeVoiceTurnResult
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode PracticeVoiceTurnResult: %v", err)
	}
	if out.VoiceTurnId != "voice-turn-1" ||
		out.UserTranscriptFinal != "我主导了设计系统迁移。" ||
		out.AssistantTextDraft != "你如何处理最高风险团队？" ||
		len(out.TtsChunks) != 1 ||
		out.TtsChunks[0].AudioRef != "fixture-audio://voice-turn-1/chunk-1" ||
		out.ProviderMetaSummary.SttProfile != "practice.voice.stt.default" ||
		out.ProviderMetaSummary.ChatProfile != "practice.followup.default" ||
		out.ProviderMetaSummary.TtsProfile != "practice.voice.tts.default" ||
		out.Session.Id != "session-voice-1" ||
		out.Session.CurrentTurn == nil ||
		out.Session.CurrentTurn.Status != string(domain.TurnStatusFollowUpRequested) ||
		out.TtsError != nil {
		t.Fatalf("response not mapped from service result: %+v", out)
	}
}

func TestCreatePracticeVoiceTurnRejectsInvalidBase64(t *testing.T) {
	service := &fakePlanService{voiceResult: fixturePracticeVoiceTurnResult()}
	handler := newTestHandler(service)
	body := fixtureCreatePracticeVoiceTurnRequest()
	body.Audio.ContentBase64 = "@@@"

	rec := httptest.NewRecorder()
	handler.CreatePracticeVoiceTurn(rec, newPracticeVoiceTurnHTTPRequest(t, body), "session-voice-1")

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	assertAPIError(t, rec, sharederrors.CodeValidationFailed, false)
	if service.voiceCalls != 0 {
		t.Fatalf("invalid base64 should not call service, got %d calls", service.voiceCalls)
	}
}

func TestCreatePracticeVoiceTurnIdempotencyReplayPreservesVoiceTurnResource(t *testing.T) {
	service := &fakePlanService{voiceResult: fixturePracticeVoiceTurnResult()}
	handler := newTestHandler(service)
	store := newRouteMemoryStore()
	mw := idempotency.New(idempotency.MiddlewareOptions{
		Store: store,
		Now:   func() time.Time { return time.Date(2026, 5, 17, 8, 50, 0, 0, time.UTC) },
		NewID: func() string { return "idempotency-record-voice-1" },
	})
	route := mw.Handler("practice", "createPracticeVoiceTurn", userFromRequestContext, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler.CreatePracticeVoiceTurn(w, r, "session-voice-1")
	}))

	first := httptest.NewRecorder()
	route.ServeHTTP(first, newPracticeVoiceTurnHTTPRequest(t, fixtureCreatePracticeVoiceTurnRequest()))
	second := httptest.NewRecorder()
	route.ServeHTTP(second, newPracticeVoiceTurnHTTPRequest(t, fixtureCreatePracticeVoiceTurnRequest()))

	if first.Code != http.StatusOK || second.Code != http.StatusOK {
		t.Fatalf("unexpected statuses: first=%d second=%d secondBody=%s", first.Code, second.Code, second.Body.String())
	}
	if second.Header().Get(idempotency.ReplayHeader) != "true" {
		t.Fatalf("expected idempotency replay header on second response")
	}
	if service.voiceCalls != 1 {
		t.Fatalf("idempotency replay should call service once, got %d", service.voiceCalls)
	}
	if first.Header().Get("X-Idempotency-Resource-ID") != "" || first.Header().Get("X-Idempotency-Resource-Type") != "" {
		t.Fatalf("internal idempotency resource headers leaked to client: %v", first.Header())
	}
	if len(store.records) != 1 {
		t.Fatalf("idempotency records = %d, want 1", len(store.records))
	}
	for _, rec := range store.records {
		if rec.resourceType != "practice_voice_turn" || rec.resourceID != "voice-turn-1" {
			t.Fatalf("voice turn resource handoff not persisted: %+v", rec)
		}
	}
	var out api.PracticeVoiceTurnResult
	if err := json.Unmarshal(second.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode replay response: %v", err)
	}
	if out.VoiceTurnId != "voice-turn-1" || len(out.TtsChunks) != 1 {
		t.Fatalf("replay lost voice turn payload: %+v", out)
	}
}

func newPracticeVoiceTurnHTTPRequest(t *testing.T, body api.CreatePracticeVoiceTurnRequest) *http.Request {
	t.Helper()
	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/practice/sessions/session-voice-1/voice-turns", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(idempotency.HeaderName, "voice-key-1")
	return req.WithContext(contextWithUser(req.Context(), "user-1"))
}

func fixtureCreatePracticeVoiceTurnRequest() api.CreatePracticeVoiceTurnRequest {
	byteLength := int32(128)
	return api.CreatePracticeVoiceTurnRequest{
		ClientVoiceTurnId: "client-voice-turn-1",
		TurnId:            "turn-1",
		Language:          "zh-CN",
		PracticeMode:      sharedtypes.PracticeModeAssisted,
		Audio: api.PracticeVoiceAudioInput{
			ContentBase64: "T2dnUw==",
			ContentType:   "audio/webm",
			DurationMs:    4320,
			ByteLength:    &byteLength,
		},
	}
}

func fixturePracticeVoiceTurnResult() domain.PracticeVoiceTurnResult {
	session := fixtureSessionRecord()
	session.ID = "session-voice-1"
	session.PlanID = "plan-voice-1"
	session.TargetJobID = "target-voice-1"
	session.CurrentTurn.ID = "turn-1"
	session.CurrentTurn.Status = string(domain.TurnStatusFollowUpRequested)
	sttLatency := int32(120)
	chatLatency := int32(240)
	ttsLatency := int32(180)
	return domain.PracticeVoiceTurnResult{
		VoiceTurnID:         "voice-turn-1",
		UserTranscriptFinal: "我主导了设计系统迁移。",
		AssistantTextDraft:  "你如何处理最高风险团队？",
		TTSChunks: []domain.PracticeVoiceTTSChunk{{
			ChunkID:     "tts-chunk-1",
			Sequence:    0,
			ContentType: "audio/mpeg",
			DurationMs:  2840,
			ByteLength:  2048,
			TextHash:    "sha256:voice-turn-1",
			AudioRef:    "fixture-audio://voice-turn-1/chunk-1",
		}},
		ProviderMetaSummary: domain.PracticeVoiceProviderMetaSummary{
			STTProfile:    "practice.voice.stt.default",
			STTProvider:   "fixture-stt",
			STTLatencyMs:  &sttLatency,
			ChatProfile:   "practice.followup.default",
			ChatProvider:  "fixture-chat",
			ChatLatencyMs: &chatLatency,
			TTSProfile:    "practice.voice.tts.default",
			TTSProvider:   "fixture-tts",
			TTSLatencyMs:  &ttsLatency,
		},
		Session: session,
	}
}
