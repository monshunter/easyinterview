package doubaospeech_test

import (
	"context"
	"errors"
	"mime"
	"testing"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient"
	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient/providerregistry"
	doubaospeech "github.com/monshunter/easyinterview/backend/internal/ai/aiclient/providers/doubao_speech"
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

func newAdapter(t *testing.T, srv *mockserver.Server) *doubaospeech.Adapter {
	t.Helper()
	a, err := doubaospeech.New(doubaospeech.Options{
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

func TestSynthesizeUsesConfiguredResponseBodyByteLimit(t *testing.T) {
	body := mockserver.DefaultTTSSuccessBody()
	for _, tc := range []struct {
		name      string
		limit     int64
		wantError bool
	}{
		{name: "exact byte limit", limit: int64(len(body))},
		{name: "one byte over", limit: int64(len(body) - 1), wantError: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv := mockserver.New()
			defer srv.Close()
			srv.SetTTSBehavior(mockserver.Behavior{StatusCode: 200, Body: body})
			a, err := doubaospeech.New(doubaospeech.Options{Provider: resolvedProvider(srv.URL()), MaxResponseBodyBytes: tc.limit})
			if err != nil {
				t.Fatalf("New: %v", err)
			}
			_, meta, err := a.Synthesize(context.Background(), ttsProfile(), ttsInput())
			if tc.wantError {
				assertCode(t, err, meta, sharederrors.CodeAiOutputInvalid)
			} else if err != nil {
				t.Fatalf("Synthesize: %v", err)
			}
		})
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

func TestSynthesize_RespectsProfileTimeout(t *testing.T) {
	srv := mockserver.New()
	defer srv.Close()
	srv.SetTTSBehavior(mockserver.Behavior{
		StatusCode: 200,
		Body:       mockserver.DefaultTTSSuccessBody(),
		SleepMs:    100,
	})

	profile := ttsProfile()
	profile.TimeoutMs = 10
	a := newAdapter(t, srv)
	start := time.Now()
	_, meta, err := a.Synthesize(context.Background(), profile, ttsInput())
	if time.Since(start) > 500*time.Millisecond {
		t.Fatal("Synthesize did not return promptly after profile timeout")
	}
	assertCode(t, err, meta, sharederrors.CodeAiProviderTimeout)
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

func TestSynthesize_UnsupportedFormatMapsSharedError(t *testing.T) {
	srv := mockserver.New()
	defer srv.Close()
	srv.SetTTSBehavior(mockserver.Behavior{
		StatusCode: 400,
		ErrorBody:  `{"error":{"code":"AI_OUTPUT_INVALID","message":"unsupported text format"}}`,
	})

	input := ttsInput()
	input.Format = "unsupported"
	a := newAdapter(t, srv)
	_, meta, err := a.Synthesize(context.Background(), ttsProfile(), input)
	assertCode(t, err, meta, sharederrors.CodeAiOutputInvalid)
}

func TestTranscribe_UnsupportedAudioFormatMapsSharedError(t *testing.T) {
	srv := mockserver.New()
	defer srv.Close()
	srv.SetSTTBehavior(mockserver.Behavior{
		StatusCode: 400,
		ErrorBody:  `{"error":{"code":"AI_OUTPUT_INVALID","message":"unsupported audio format"}}`,
	})

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
	a := newAdapter(t, srv)
	_, meta, err := a.Transcribe(context.Background(), sttProfile, aiclient.TranscriptionInput{
		Audio:       []byte("test"),
		Filename:    "test.bin",
		ContentType: "application/octet-stream",
	})
	assertCode(t, err, meta, sharederrors.CodeAiOutputInvalid)
}

func TestTranscribe_RespectsProfileTimeout(t *testing.T) {
	srv := mockserver.New()
	defer srv.Close()
	srv.SetSTTBehavior(mockserver.Behavior{
		StatusCode: 200,
		Body:       mockserver.DefaultSTTSuccessBody(),
		SleepMs:    100,
	})

	sttProfile := &aiclient.ModelProfile{
		Name:       "practice.voice.stt.default",
		Capability: aiclient.CapabilitySTT,
		Default: aiclient.ProviderConfig{
			ProviderRef: "doubao",
			Model:       "stt-model",
		},
		TimeoutMs: 10,
		Version:   "1.0.0",
	}
	a := newAdapter(t, srv)
	start := time.Now()
	_, meta, err := a.Transcribe(context.Background(), sttProfile, aiclient.TranscriptionInput{
		Audio:       []byte("test"),
		Filename:    "test.webm",
		ContentType: "audio/webm",
	})
	if time.Since(start) > 500*time.Millisecond {
		t.Fatal("Transcribe did not return promptly after profile timeout")
	}
	assertCode(t, err, meta, sharederrors.CodeAiProviderTimeout)
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

	_, err := doubaospeech.New(doubaospeech.Options{Provider: rp})
	if err == nil {
		t.Fatal("expected error for non-doubao protocol")
	}
}

func TestNew_RejectsMissingBaseURL(t *testing.T) {
	srv := mockserver.New()
	defer srv.Close()
	rp := resolvedProvider(srv.URL())
	rp.BaseURL = ""

	_, err := doubaospeech.New(doubaospeech.Options{Provider: rp})
	if err == nil {
		t.Fatal("expected error for missing BaseURL")
	}
}

func TestNew_RejectsMissingAPIKey(t *testing.T) {
	srv := mockserver.New()
	defer srv.Close()
	rp := resolvedProvider(srv.URL())
	rp.APIKey = ""

	_, err := doubaospeech.New(doubaospeech.Options{Provider: rp})
	if err == nil {
		t.Fatal("expected error for missing APIKey")
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
	srv.SetSTTBehavior(mockserver.Behavior{
		StatusCode: 200,
		Body:       mockserver.DefaultSTTSuccessBody(),
	})
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
	_, _, err := a.Transcribe(context.Background(), sttProfile, aiclient.TranscriptionInput{
		Audio:       []byte("test"),
		Filename:    "test.webm",
		ContentType: "audio/webm",
	})
	if err != nil {
		t.Fatalf("Transcribe: %v", err)
	}
	got := srv.LastRequest()
	if got.Method != "POST" {
		t.Fatalf("expected POST, got %q", got.Method)
	}
	if got.Path != doubaospeech.PathSTTRecognize {
		t.Fatalf("expected %s, got %q", doubaospeech.PathSTTRecognize, got.Path)
	}
	mediaType, _, err := mime.ParseMediaType(got.ContentType)
	if err != nil {
		t.Fatalf("parse Content-Type: %v", err)
	}
	if mediaType != "application/json" {
		t.Fatalf("expected application/json, got %q", mediaType)
	}
}

func TestRejectsOpenAICompatibleProtocol(t *testing.T) {
	srv := mockserver.New()
	defer srv.Close()
	rp := resolvedProvider(srv.URL())
	rp.Entry.Protocol = aiclient.ProviderProtocolOpenAICompatible

	_, err := doubaospeech.New(doubaospeech.Options{Provider: rp})
	if err == nil {
		t.Fatal("doubao_speech must reject openai_compatible protocol")
	}
}
