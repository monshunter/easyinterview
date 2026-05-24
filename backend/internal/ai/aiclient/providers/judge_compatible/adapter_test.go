package judgecompatible_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient"
	judgecompatible "github.com/monshunter/easyinterview/backend/internal/ai/aiclient/providers/judge_compatible"
	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient/providerregistry"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
)

func judgeEntry() aiclient.ProviderRegistryEntry {
	return aiclient.ProviderRegistryEntry{
		Name:     "judge-deepseek",
		Protocol: aiclient.ProviderProtocolJudgeCompatible,
	}
}

func resolved(baseURL, apiKey string) providerregistry.ResolvedProvider {
	return providerregistry.ResolvedProvider{Entry: judgeEntry(), BaseURL: baseURL, APIKey: apiKey}
}

func judgeProfile(timeoutMs int) *aiclient.ModelProfile {
	return &aiclient.ModelProfile{
		Name:       "judge.default",
		Capability: aiclient.CapabilityJudge,
		Status:     aiclient.ProfileStatusActive,
		Default:    aiclient.ProviderConfig{ProviderRef: "judge-deepseek", Model: "deepseek-v4-pro"},
		TimeoutMs:  timeoutMs,
		Route:      "judge.default",
		Version:    "1.0.0",
	}
}

func TestNewFailsFastOnMissingSecret(t *testing.T) {
	if _, err := judgecompatible.New(judgecompatible.Options{Provider: resolved("", "k")}); err == nil {
		t.Fatal("expected error when BaseURL missing")
	}
	if _, err := judgecompatible.New(judgecompatible.Options{Provider: resolved("http://x", "")}); err == nil {
		t.Fatal("expected error when APIKey missing")
	}
}

func TestNewRejectsWrongProtocol(t *testing.T) {
	entry := judgeEntry()
	entry.Protocol = aiclient.ProviderProtocolOpenAICompatible
	_, err := judgecompatible.New(judgecompatible.Options{Provider: providerregistry.ResolvedProvider{Entry: entry, BaseURL: "http://x", APIKey: "k"}})
	if err == nil {
		t.Fatal("expected protocol mismatch error")
	}
}

func TestCompletePostsAndParses(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != judgecompatible.PathChatCompletions {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer k" {
			t.Fatalf("missing bearer auth")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"model":"deepseek-v4-pro","choices":[{"message":{"content":"{\"scores\":[]}"},"finish_reason":"stop"}],"usage":{"prompt_tokens":5,"completion_tokens":4}}`))
	}))
	defer server.Close()

	adapter, err := judgecompatible.New(judgecompatible.Options{Provider: resolved(server.URL, "k")})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	payload := aiclient.CompletePayload{Messages: []aiclient.Message{{Role: "user", Content: "score this"}}}
	resp, meta, err := adapter.Complete(context.Background(), judgeProfile(5000), payload)
	if err != nil {
		t.Fatalf("Complete: %v", err)
	}
	if resp.Content == "" {
		t.Fatal("expected non-empty content")
	}
	if meta.Capability != aiclient.CapabilityJudge {
		t.Fatalf("meta.Capability: want judge, got %q", meta.Capability)
	}
	if meta.InputTokens != 5 || meta.OutputTokens != 4 {
		t.Fatalf("token usage not propagated: in=%d out=%d", meta.InputTokens, meta.OutputTokens)
	}
}

func TestUnsupportedMethodsFailClose(t *testing.T) {
	adapter, err := judgecompatible.New(judgecompatible.Options{Provider: resolved("http://x", "k")})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	p := judgeProfile(0)
	if _, _, err := adapter.Transcribe(context.Background(), p, aiclient.TranscriptionInput{}); !isUnsupported(err) {
		t.Fatalf("Transcribe: want unsupported, got %v", err)
	}
	if _, err := adapter.Stream(context.Background(), p, aiclient.CompletePayload{}); !isUnsupported(err) {
		t.Fatalf("Stream: want unsupported, got %v", err)
	}
	if _, _, err := adapter.Synthesize(context.Background(), p, aiclient.SynthesisInput{}); !isUnsupported(err) {
		t.Fatalf("Synthesize: want unsupported, got %v", err)
	}
}

func isUnsupported(err error) bool {
	var apiErr *sharederrors.APIError
	return errors.As(err, &apiErr) && apiErr.Code == sharederrors.CodeAiUnsupportedCapability
}
