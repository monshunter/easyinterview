package judgecompatible_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient"
	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient/providerregistry"
	judgecompatible "github.com/monshunter/easyinterview/backend/internal/ai/aiclient/providers/judge_compatible"
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
		Default: aiclient.ProviderConfig{
			ProviderRef: "judge-deepseek",
			Model:       "deepseek-v4-pro",
			Params: map[string]any{
				"temperature":     0.0,
				"thinking":        "disabled",
				"response_format": "json_object",
			},
		},
		MaxTokens: 6144,
		TimeoutMs: timeoutMs,
		Route:     "judge.default",
		Version:   "1.0.0",
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
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read request: %v", err)
		}
		var request map[string]any
		if err := json.Unmarshal(body, &request); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		thinking, ok := request["thinking"].(map[string]any)
		if !ok || thinking["type"] != "disabled" {
			t.Fatalf("thinking mode must be disabled for deterministic judge JSON: %+v", request["thinking"])
		}
		responseFormat, ok := request["response_format"].(map[string]any)
		if !ok || responseFormat["type"] != "json_object" {
			t.Fatalf("judge request must require a JSON object: %+v", request["response_format"])
		}
		if request["max_tokens"] != float64(6144) {
			t.Fatalf("max_tokens=%v, want 6144", request["max_tokens"])
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

func TestCompleteUsesConfiguredResponseBodyByteLimit(t *testing.T) {
	body := []byte(`{"model":"deepseek-v4-pro","choices":[{"message":{"content":"{\"scores\":[]}"},"finish_reason":"stop"}],"usage":{"prompt_tokens":5,"completion_tokens":4}}`)
	for _, tc := range []struct {
		name      string
		limit     int64
		wantError bool
	}{
		{name: "exact byte limit", limit: int64(len(body))},
		{name: "one byte over", limit: int64(len(body) - 1), wantError: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write(body) }))
			defer server.Close()
			adapter, err := judgecompatible.New(judgecompatible.Options{Provider: resolved(server.URL, "k"), MaxResponseBodyBytes: tc.limit})
			if err != nil {
				t.Fatalf("New: %v", err)
			}
			_, _, err = adapter.Complete(context.Background(), judgeProfile(5000), aiclient.CompletePayload{Messages: []aiclient.Message{{Role: "user", Content: "score this"}}})
			var apiErr *sharederrors.APIError
			if tc.wantError {
				if !errors.As(err, &apiErr) || apiErr.Code != sharederrors.CodeAiOutputInvalid {
					t.Fatalf("err=%v want %s", err, sharederrors.CodeAiOutputInvalid)
				}
			} else if err != nil {
				t.Fatalf("Complete: %v", err)
			}
		})
	}
}

func TestCompleteRejectsReasoningOnlyResponseWithoutLeakingReasoning(t *testing.T) {
	const privateReasoning = "private chain of thought must never leave the adapter"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"model":"deepseek-v4-pro","choices":[{"message":{"content":"","reasoning_content":"` + privateReasoning + `"},"finish_reason":"length"}],"usage":{"prompt_tokens":1292,"completion_tokens":2048}}`))
	}))
	defer server.Close()

	adapter, err := judgecompatible.New(judgecompatible.Options{Provider: resolved(server.URL, "k")})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	_, meta, err := adapter.Complete(
		context.Background(),
		judgeProfile(5000),
		aiclient.CompletePayload{Messages: []aiclient.Message{{Role: "user", Content: "score this"}}},
	)
	if err == nil {
		t.Fatal("expected reasoning-only response to fail closed")
	}
	var apiErr *sharederrors.APIError
	if !errors.As(err, &apiErr) || apiErr.Code != sharederrors.CodeAiOutputInvalid {
		t.Fatalf("expected AI_OUTPUT_INVALID, got %v", err)
	}
	if strings.Contains(err.Error(), privateReasoning) {
		t.Fatalf("error leaked private reasoning: %v", err)
	}
	if !strings.Contains(err.Error(), "finish_reason=length") || !strings.Contains(err.Error(), "reasoning_content_present=true") {
		t.Fatalf("error lacks redacted root-cause metadata: %v", err)
	}
	if meta.InputTokens != 1292 || meta.OutputTokens != 2048 || meta.ValidationStatus != aiclient.ValidationStatusInvalid {
		t.Fatalf("invalid-response meta lost usage/provenance: %+v", meta)
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
