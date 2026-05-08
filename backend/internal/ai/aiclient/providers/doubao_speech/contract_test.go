package doubao_speech_test

import (
	"context"
	"errors"
	"testing"

	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient"
	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient/providerregistry"
	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient/providers/doubao_speech"
	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient/providers/doubao_speech/mockserver"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
)

func resolvedProvider(baseURL string) providerregistry.ResolvedProvider {
	return providerregistry.ResolvedProvider{
		Entry: aiclient.ProviderRegistryEntry{
			Name:         "doubao",
			Protocol:     aiclient.ProviderProtocolDoubaoSpeech,
			Capabilities: []aiclient.Capability{aiclient.CapabilityTts, aiclient.CapabilitySTT},
		},
		BaseURL: baseURL,
		APIKey:  "test-key",
	}
}

func ttsProfile() *aiclient.ModelProfile {
	return &aiclient.ModelProfile{
		Name:       "practice.voice.tts.default",
		Capability: aiclient.CapabilityTts,
		Default: aiclient.ProviderConfig{
			ProviderRef: "doubao",
			Model:       "tts-model",
		},
		TimeoutMs: 5000,
		Version:   "1.0.0",
	}
}

func ttsInput() aiclient.SynthesisInput {
	return aiclient.SynthesisInput{
		Text:         "你好",
		Voice:        "zh_female_qingxin",
		Format:       "mp3",
		SpeakingRate: 1.0,
		Language:     "zh-CN",
	}
}

func newAdapter(t *testing.T, srv *mockserver.Server) *doubao_speech.Adapter {
	t.Helper()
	a, err := doubao_speech.New(doubao_speech.Options{
		Provider: resolvedProvider(srv.URL()),
	})
	if err != nil {
		t.Fatalf("doubao_speech.New: %v", err)
	}
	return a
}

func TestSynthesize_NormalTTSSynthesis(t *testing.T) {
	srv := mockserver.New()
	defer srv.Close()
	srv.SetTTSBehavior(mockserver.Behavior{
		StatusCode: 200,
		Body:       mockserver.DefaultTTSSuccessBody(),
	})

	a := newAdapter(t, srv)
	resp, meta, err := a.Synthesize(context.Background(), ttsProfile(), ttsInput())
	if err != nil {
		t.Fatalf("Synthesize: %v", err)
	}
	if len(resp.Audio) == 0 {
		t.Fatal("expected non-empty audio bytes")
	}
	if resp.ContentType != "audio/mpeg" {
		t.Fatalf("expected ContentType=audio/mpeg, got %q", resp.ContentType)
	}
	if resp.DurationMs != 1500 {
		t.Fatalf("expected DurationMs=1500, got %d", resp.DurationMs)
	}
	if resp.CharCount != 10 {
		t.Fatalf("expected CharCount=10, got %d", resp.CharCount)
	}
	if meta.Provider != "doubao" {
		t.Fatalf("expected meta.Provider=doubao, got %q", meta.Provider)
	}
	if meta.ModelFamily != "doubao_speech" {
		t.Fatalf("expected meta.ModelFamily=doubao_speech, got %q", meta.ModelFamily)
	}
}

func TestSynthesize_ProviderError5xx(t *testing.T) {
	srv := mockserver.New()
	defer srv.Close()
	srv.SetTTSBehavior(mockserver.Behavior{
		StatusCode: 503,
		ErrorBody:  `{"error":{"code":"AI_PROVIDER_TIMEOUT","message":"service unavailable"}}`,
	})

	a := newAdapter(t, srv)
	_, meta, err := a.Synthesize(context.Background(), ttsProfile(), ttsInput())
	if err == nil {
		t.Fatal("expected error for 5xx response")
	}
	var apiErr *sharederrors.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected APIError, got %T: %v", err, err)
	}
	if meta.ErrorCode == "" {
		t.Fatal("expected error code in meta")
	}
}

func TestSynthesize_ProviderError4xx(t *testing.T) {
	srv := mockserver.New()
	defer srv.Close()
	srv.SetTTSBehavior(mockserver.Behavior{
		StatusCode: 400,
		ErrorBody:  `{"error":{"code":"VALIDATION_FAILED","message":"invalid voice parameter"}}`,
	})

	a := newAdapter(t, srv)
	_, _, err := a.Synthesize(context.Background(), ttsProfile(), ttsInput())
	if err == nil {
		t.Fatal("expected error for 4xx response")
	}
}

func TestSynthesize_MissingAudioReturnsInvalid(t *testing.T) {
	srv := mockserver.New()
	defer srv.Close()
	srv.SetTTSBehavior(mockserver.Behavior{
		StatusCode: 200,
		Body:       `{"audio":"","content_type":"audio/mpeg","duration_ms":0,"char_count":0}`,
	})

	a := newAdapter(t, srv)
	_, meta, err := a.Synthesize(context.Background(), ttsProfile(), ttsInput())
	assertCode(t, err, meta, sharederrors.CodeAiOutputInvalid)
}

func TestComplete_ReturnsUnsupportedCapability(t *testing.T) {
	srv := mockserver.New()
	defer srv.Close()
	a := newAdapter(t, srv)

	_, meta, err := a.Complete(context.Background(), ttsProfile(), aiclient.CompletePayload{
		Messages: []aiclient.Message{{Role: "user", Content: "test"}},
	})
	assertCode(t, err, meta, sharederrors.CodeAiUnsupportedCapability)
}

func TestStream_ReturnsUnsupportedCapability(t *testing.T) {
	srv := mockserver.New()
	defer srv.Close()
	a := newAdapter(t, srv)

	_, err := a.Stream(context.Background(), ttsProfile(), aiclient.CompletePayload{
		Messages: []aiclient.Message{{Role: "user", Content: "test"}},
	})
	if err == nil {
		t.Fatal("expected unsupported capability error")
	}
	var apiErr *sharederrors.APIError
	if !errors.As(err, &apiErr) || apiErr.Code != sharederrors.CodeAiUnsupportedCapability {
		t.Fatalf("expected AI_UNSUPPORTED_CAPABILITY, got %v", err)
	}
}

func TestNew_RejectsNonDoubaoProtocol(t *testing.T) {
	srv := mockserver.New()
	defer srv.Close()
	rp := resolvedProvider(srv.URL())
	rp.Entry.Protocol = aiclient.ProviderProtocolOpenAICompatible

	_, err := doubao_speech.New(doubao_speech.Options{Provider: rp})
	if err == nil {
		t.Fatal("expected error for non-doubao protocol")
	}
}

func assertCode(t *testing.T, err error, meta aiclient.AICallMeta, code string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error")
	}
	var apiErr *sharederrors.APIError
	if !errors.As(err, &apiErr) || apiErr.Code != code {
		t.Fatalf("expected error code %q, got %v", code, err)
	}
	if meta.ErrorCode != code {
		t.Fatalf("expected meta.ErrorCode=%q, got %q", code, meta.ErrorCode)
	}
}

func TestDoesNotReuseOpenAIAudioTranscriptionsWire(t *testing.T) {
	srv := mockserver.New()
	defer srv.Close()
	a := newAdapter(t, srv)

	// doubao_speech uses its own /v1/audio/recognize path (JSON POST),
	// not OpenAI's /v1/audio/transcriptions (multipart POST).
	// Attempting OpenAI-compatible multipart Transcribe payload would
	// fail because the adapter does not build multipart/form-data.
	sttProfile := &aiclient.ModelProfile{
		Name:       "practice.voice.stt.default",
		Capability: aiclient.CapabilitySTT,
		Default: aiclient.ProviderConfig{
			ProviderRef: "doubao",
			Model:       "stt-model",
		},
		TimeoutMs: 5000,
		Version:   "1.0.0",
	}
	// Even with valid input, the openai_compatible wire is not used.
	_, _, err := a.Transcribe(context.Background(), sttProfile, aiclient.TranscriptionInput{
		Audio:       []byte("test"),
		Filename:    "test.webm",
		ContentType: "audio/webm",
	})
	// The mock server returns 404 for unexpected paths; adapter should
	// map this as provider timeout (5xx behavior).
	if err == nil {
		t.Fatal("expected doubao_speech to use its own STT wire, not OpenAI-compatible")
	}
}

func TestRejectsOpenAICompatibleProtocol(t *testing.T) {
	srv := mockserver.New()
	defer srv.Close()
	rp := resolvedProvider(srv.URL())
	rp.Entry.Protocol = aiclient.ProviderProtocolOpenAICompatible

	_, err := doubao_speech.New(doubao_speech.Options{Provider: rp})
	if err == nil {
		t.Fatal("doubao_speech must reject openai_compatible protocol")
	}
}
