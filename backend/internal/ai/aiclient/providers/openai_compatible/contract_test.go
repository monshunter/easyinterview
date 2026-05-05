package openaicompatible_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient"
	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient/providers/openai_compatible"
	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient/providers/openai_compatible/mockserver"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
)

const (
	chatModelID      = "chat-primary-2026-05-05"
	chatModelFamily  = "chat-primary"
	embedModelID     = "embed-small"
	embedModelFamily = "embed-small"
)

func chatProfile(timeoutMs int) *aiclient.ModelProfile {
	return &aiclient.ModelProfile{
		Name:     "practice.followup.default",
		TaskType: aiclient.TaskTypeChat,
		Default: aiclient.ProviderConfig{
			Provider: openaicompatible.Name,
			Model:    chatModelID,
		},
		TimeoutMs:    timeoutMs,
		Route: "practice.followup",
		Version:      "1.0.0",
	}
}

func embedProfile(timeoutMs int) *aiclient.ModelProfile {
	return &aiclient.ModelProfile{
		Name:     "review.embed.default",
		TaskType: aiclient.TaskTypeEmbed,
		Default: aiclient.ProviderConfig{
			Provider: openaicompatible.Name,
			Model:    embedModelID,
		},
		TimeoutMs: timeoutMs,
		Version:   "1.0.0",
	}
}

func samplePayload() aiclient.CompletePayload {
	return aiclient.CompletePayload{
		Messages: []aiclient.Message{
			{Role: "system", Content: "you are an interviewer."},
			{Role: "user", Content: "Tell me about a project."},
		},
		Metadata: aiclient.CallMetadata{
			FeatureKey:    "practice.followup",
			PromptVersion: "p1",
			RubricVersion: "r1",
			Language:      "en",
		},
	}
}

func newAdapter(t *testing.T, srv *mockserver.Server) *openaicompatible.Adapter {
	t.Helper()
	a, err := openaicompatible.New(openaicompatible.Options{
		BaseURL: srv.URL(),
		APIKey:  "test-key",
	})
	if err != nil {
		t.Fatalf("openai_compatible.New: %v", err)
	}
	return a
}

func TestComplete_NormalChatCompletion(t *testing.T) {
	srv := mockserver.New()
	defer srv.Close()
	a := newAdapter(t, srv)

	resp, meta, err := a.Complete(aiclient.WithRequestID(context.Background(), "req-1"), chatProfile(5000), samplePayload())
	if err != nil {
		t.Fatalf("Complete: %v", err)
	}
	if !strings.HasPrefix(resp.Content, "mock response for ") {
		t.Fatalf("unexpected content: %q", resp.Content)
	}
	if meta.Provider != openaicompatible.Name {
		t.Fatalf("expected meta.Provider=%q, got %q", openaicompatible.Name, meta.Provider)
	}
	if meta.ModelID != chatModelID {
		t.Fatalf("expected ModelID=%q, got %q", chatModelID, meta.ModelID)
	}
	if meta.ModelFamily != chatModelFamily {
		t.Fatalf("expected ModelFamily=%q, got %q", chatModelFamily, meta.ModelFamily)
	}
	if meta.InputTokens == 0 {
		t.Fatalf("expected non-zero input tokens, got %+v", meta)
	}
	if meta.OutputTokens == 0 {
		t.Fatalf("expected non-zero output tokens, got %+v", meta)
	}
	if meta.LatencyMs < 0 {
		t.Fatalf("expected non-negative latency, got %d", meta.LatencyMs)
	}

	requests := srv.Captured()
	if len(requests) != 1 {
		t.Fatalf("expected 1 captured request, got %d", len(requests))
	}
	got := requests[0]
	if got.Path != "/v1/chat/completions" {
		t.Fatalf("expected /v1/chat/completions, got %q", got.Path)
	}
	if got.Authorization != "Bearer test-key" {
		t.Fatalf("expected Bearer auth, got %q", got.Authorization)
	}
	if got.ContentType != "application/json" {
		t.Fatalf("expected JSON content type, got %q", got.ContentType)
	}
	if got.RequestID != "req-1" {
		t.Fatalf("expected X-Request-ID propagation, got %q", got.RequestID)
	}
}

func TestComplete_BaseURLMayIncludeV1Prefix(t *testing.T) {
	srv := mockserver.New()
	defer srv.Close()
	a, err := openaicompatible.New(openaicompatible.Options{
		BaseURL: srv.URL() + "/v1",
		APIKey:  "test-key",
	})
	if err != nil {
		t.Fatalf("openai_compatible.New: %v", err)
	}

	_, _, err = a.Complete(context.Background(), chatProfile(5000), samplePayload())
	if err != nil {
		t.Fatalf("Complete: %v", err)
	}

	requests := srv.Captured()
	if len(requests) != 1 {
		t.Fatalf("expected 1 captured request, got %d", len(requests))
	}
	if got := requests[0].Path; got != "/v1/chat/completions" {
		t.Fatalf("expected normalized /v1/chat/completions path, got %q", got)
	}
}

func TestEmbed_NormalEmbeddings(t *testing.T) {
	srv := mockserver.New()
	defer srv.Close()
	a := newAdapter(t, srv)

	resp, meta, err := a.Embed(context.Background(), embedProfile(5000), aiclient.EmbedInput{
		Texts: []string{"hello", "world"},
		Metadata: aiclient.CallMetadata{
			FeatureKey:    "review.embed",
			PromptVersion: "p1",
			RubricVersion: "r1",
			Language:      "en",
		},
	})
	if err != nil {
		t.Fatalf("Embed: %v", err)
	}
	if len(resp.Vectors) != 2 {
		t.Fatalf("expected 2 vectors, got %d", len(resp.Vectors))
	}
	if meta.Provider != openaicompatible.Name {
		t.Fatalf("provider mismatch: %q", meta.Provider)
	}
	if meta.ModelID != embedModelID {
		t.Fatalf("model mismatch: %q", meta.ModelID)
	}
	if meta.ModelFamily != embedModelFamily {
		t.Fatalf("expected ModelFamily=%q, got %q", embedModelFamily, meta.ModelFamily)
	}
	if meta.InputTokens == 0 {
		t.Fatalf("expected non-zero input tokens, got %+v", meta)
	}
}

func TestComplete_TimeoutMapsToAIProviderTimeout(t *testing.T) {
	srv := mockserver.New()
	srv.SetChatBehavior(mockserver.Behavior{SleepBeforeRespond: 200 * time.Millisecond})
	defer srv.Close()
	a := newAdapter(t, srv)

	_, meta, err := a.Complete(context.Background(), chatProfile(50), samplePayload())
	if err == nil {
		t.Fatalf("expected timeout error")
	}
	var apiErr *sharederrors.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected APIError, got %T: %v", err, err)
	}
	if apiErr.Code != sharederrors.CodeAiProviderTimeout {
		t.Fatalf("expected %q, got %q", sharederrors.CodeAiProviderTimeout, apiErr.Code)
	}
	if meta.ErrorCode != sharederrors.CodeAiProviderTimeout {
		t.Fatalf("meta.ErrorCode mismatch: %q", meta.ErrorCode)
	}
}

func TestComplete_5xxMapsToAIProviderTimeout(t *testing.T) {
	srv := mockserver.New()
	srv.SetChatBehavior(mockserver.Behavior{StatusCode: 503, ErrorBody: "upstream gone"})
	defer srv.Close()
	a := newAdapter(t, srv)

	_, _, err := a.Complete(context.Background(), chatProfile(2000), samplePayload())
	if err == nil {
		t.Fatalf("expected error for 5xx")
	}
	var apiErr *sharederrors.APIError
	if !errors.As(err, &apiErr) || apiErr.Code != sharederrors.CodeAiProviderTimeout {
		t.Fatalf("expected AI_PROVIDER_TIMEOUT, got %v", err)
	}
}

func TestComplete_4xxParsesErrorEnvelope(t *testing.T) {
	srv := mockserver.New()
	srv.SetChatBehavior(mockserver.Behavior{
		StatusCode: 400,
		ErrorBody:  `{"error":{"code":"AI_OUTPUT_INVALID","message":"bad input"}}`,
	})
	defer srv.Close()
	a := newAdapter(t, srv)

	_, _, err := a.Complete(context.Background(), chatProfile(2000), samplePayload())
	if err == nil {
		t.Fatalf("expected error for 4xx")
	}
	var apiErr *sharederrors.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected APIError, got %T: %v", err, err)
	}
	if apiErr.Code != sharederrors.CodeAiOutputInvalid {
		t.Fatalf("expected upstream code passthrough %q, got %q", sharederrors.CodeAiOutputInvalid, apiErr.Code)
	}
}

func TestComplete_4xxUnknownErrorCodeFallsBackToAIOutputInvalid(t *testing.T) {
	srv := mockserver.New()
	srv.SetChatBehavior(mockserver.Behavior{
		StatusCode: 400,
		ErrorBody:  `{"error":{"code":"PROVIDER_PRIVATE_BAD_REQUEST","message":"bad input"}}`,
	})
	defer srv.Close()
	a := newAdapter(t, srv)

	_, meta, err := a.Complete(context.Background(), chatProfile(2000), samplePayload())
	if err == nil {
		t.Fatalf("expected error for 4xx")
	}
	var apiErr *sharederrors.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected APIError, got %T: %v", err, err)
	}
	if apiErr.Code != sharederrors.CodeAiOutputInvalid {
		t.Fatalf("expected unknown upstream code to fall back to %q, got %q", sharederrors.CodeAiOutputInvalid, apiErr.Code)
	}
	if meta.ErrorCode != sharederrors.CodeAiOutputInvalid {
		t.Fatalf("expected meta.ErrorCode=%q, got %q", sharederrors.CodeAiOutputInvalid, meta.ErrorCode)
	}
}

func TestComplete_FallbackHeadersPopulateMeta(t *testing.T) {
	srv := mockserver.New()
	srv.SetChatBehavior(mockserver.Behavior{
		FallbackFrom: "primary/chat",
		FallbackTo:   "fallback/chat",
		Route:        "practice.followup",
	})
	defer srv.Close()
	a := newAdapter(t, srv)

	_, meta, err := a.Complete(context.Background(), chatProfile(2000), samplePayload())
	if err != nil {
		t.Fatalf("Complete: %v", err)
	}
	if len(meta.FallbackChain) != 2 || meta.FallbackChain[0] != "primary/chat" || meta.FallbackChain[1] != "fallback/chat" {
		t.Fatalf("fallback chain not populated: %+v", meta.FallbackChain)
	}
	if meta.Route != "practice.followup" {
		t.Fatalf("route not populated: %q", meta.Route)
	}
}

func TestComplete_MissingChoicesReturnsAIOutputInvalid(t *testing.T) {
	srv := mockserver.New()
	srv.SetChatBehavior(mockserver.Behavior{MissingChoices: true})
	defer srv.Close()
	a := newAdapter(t, srv)

	_, _, err := a.Complete(context.Background(), chatProfile(2000), samplePayload())
	if err == nil {
		t.Fatalf("expected AI_OUTPUT_INVALID for missing choices")
	}
	var apiErr *sharederrors.APIError
	if !errors.As(err, &apiErr) || apiErr.Code != sharederrors.CodeAiOutputInvalid {
		t.Fatalf("expected AI_OUTPUT_INVALID, got %v", err)
	}
}

func TestComplete_NoFallbackHeadersUsesProfileRoute(t *testing.T) {
	srv := mockserver.New()
	defer srv.Close()
	a := newAdapter(t, srv)

	_, meta, err := a.Complete(context.Background(), chatProfile(2000), samplePayload())
	if err != nil {
		t.Fatalf("Complete: %v", err)
	}
	if meta.Route != "practice.followup" {
		t.Fatalf("expected fallback route from profile, got %q", meta.Route)
	}
	if len(meta.FallbackChain) != 0 {
		t.Fatalf("expected empty fallback chain when headers absent, got %+v", meta.FallbackChain)
	}
}

func TestNew_RequiresBaseURLAndAPIKey(t *testing.T) {
	if _, err := openaicompatible.New(openaicompatible.Options{APIKey: "k"}); err == nil {
		t.Fatalf("expected error when BaseURL missing")
	}
	if _, err := openaicompatible.New(openaicompatible.Options{BaseURL: "http://x"}); err == nil {
		t.Fatalf("expected error when APIKey missing")
	}
}
