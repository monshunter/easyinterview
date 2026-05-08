package minimax_speech_test

import (
	"context"
	"errors"
	"testing"

	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient"
	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient/providerregistry"
	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient/providers/minimax_speech"
	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient/providers/minimax_speech/mockserver"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
)

func resolvedProvider(baseURL string) providerregistry.ResolvedProvider {
	return providerregistry.ResolvedProvider{
		Entry: aiclient.ProviderRegistryEntry{
			Name:         "minimax",
			Protocol:     aiclient.ProviderProtocolMinimaxSpeech,
			Capabilities: []aiclient.Capability{aiclient.CapabilityTts},
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
			ProviderRef: "minimax",
			Model:       "speech-02-turbo",
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

func newAdapter(t *testing.T, srv *mockserver.Server) *minimax_speech.Adapter {
	t.Helper()
	a, err := minimax_speech.New(minimax_speech.Options{
		Provider: resolvedProvider(srv.URL()),
	})
	if err != nil {
		t.Fatalf("minimax_speech.New: %v", err)
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
	if resp.DurationMs != 1200 {
		t.Fatalf("expected DurationMs=1200, got %d", resp.DurationMs)
	}
	if meta.Provider != "minimax" {
		t.Fatalf("expected meta.Provider=minimax, got %q", meta.Provider)
	}
	if meta.ModelFamily != "minimax_speech" {
		t.Fatalf("expected meta.ModelFamily=minimax_speech, got %q", meta.ModelFamily)
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
	_, _, err := a.Synthesize(context.Background(), ttsProfile(), ttsInput())
	if err == nil {
		t.Fatal("expected error for 5xx response")
	}
}

func TestTranscribe_ReturnsUnsupportedCapability(t *testing.T) {
	srv := mockserver.New()
	defer srv.Close()
	a := newAdapter(t, srv)

	_, _, err := a.Transcribe(context.Background(), ttsProfile(), aiclient.TranscriptionInput{
		Audio:       []byte("test"),
		Filename:    "test.webm",
		ContentType: "audio/webm",
	})
	assertCode(t, err, sharederrors.CodeAiUnsupportedCapability)
}

func TestComplete_ReturnsUnsupportedCapability(t *testing.T) {
	srv := mockserver.New()
	defer srv.Close()
	a := newAdapter(t, srv)

	_, _, err := a.Complete(context.Background(), ttsProfile(), aiclient.CompletePayload{
		Messages: []aiclient.Message{{Role: "user", Content: "test"}},
	})
	assertCode(t, err, sharederrors.CodeAiUnsupportedCapability)
}

func TestNew_RejectsNonMinimaxProtocol(t *testing.T) {
	srv := mockserver.New()
	defer srv.Close()
	rp := resolvedProvider(srv.URL())
	rp.Entry.Protocol = aiclient.ProviderProtocolOpenAICompatible

	_, err := minimax_speech.New(minimax_speech.Options{Provider: rp})
	if err == nil {
		t.Fatal("expected error for non-minimax protocol")
	}
}

func TestNew_RejectsMissingBaseURL(t *testing.T) {
	srv := mockserver.New()
	defer srv.Close()
	rp := resolvedProvider(srv.URL())
	rp.BaseURL = ""

	_, err := minimax_speech.New(minimax_speech.Options{Provider: rp})
	if err == nil {
		t.Fatal("expected error for missing BaseURL")
	}
}

func TestNew_RejectsMissingAPIKey(t *testing.T) {
	srv := mockserver.New()
	defer srv.Close()
	rp := resolvedProvider(srv.URL())
	rp.APIKey = ""

	_, err := minimax_speech.New(minimax_speech.Options{Provider: rp})
	if err == nil {
		t.Fatal("expected error for missing APIKey")
	}
}

func assertCode(t *testing.T, err error, code string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error")
	}
	var apiErr *sharederrors.APIError
	if !errors.As(err, &apiErr) || apiErr.Code != code {
		t.Fatalf("expected error code %q, got %v", code, err)
	}
}
