package bootstrap_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient"
	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient/bootstrap"
)

// writeJudgeRuntimeConfig writes a runtime config whose judge.default profile
// routes through a non-placeholder judge_compatible provider.
func writeJudgeRuntimeConfig(t *testing.T) (string, string) {
	t.Helper()
	dir := t.TempDir()
	registryPath := filepath.Join(dir, "ai-providers.yaml")
	if err := os.WriteFile(registryPath, []byte(`providers:
  - name: judge-deepseek
    protocol: judge_compatible
    base_url_env: AI_PROVIDER_BASE_URL
    api_key_env: AI_PROVIDER_API_KEY
    capabilities: [judge]
    version: 1.0.0
`), 0o600); err != nil {
		t.Fatalf("write registry: %v", err)
	}
	profilePath := filepath.Join(dir, "ai-profiles.yaml")
	if err := os.WriteFile(profilePath, []byte(`profiles:
  - name: judge.default
    capability: judge
    status: active
    default:
      provider_ref: judge-deepseek
      model: deepseek-v4-pro
    timeout_ms: 5000
    route: judge.default
    version: 1.0.0
`), 0o600); err != nil {
		t.Fatalf("write profile: %v", err)
	}
	return registryPath, profilePath
}

// TestNewClientResolvesJudgeCompatibleProvider asserts the production resolver
// materializes a judge_compatible adapter (no "protocol not implemented") and
// routes a CompleteJudge call through it (plan 004 §2.2).
func TestNewClientResolvesJudgeCompatibleProvider(t *testing.T) {
	var sawPath, sawAuth bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/chat/completions") {
			sawPath = true
		}
		if r.Header.Get("Authorization") == "Bearer judge-secret" {
			sawAuth = true
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"model":"deepseek-v4-pro","choices":[{"message":{"content":"{\"scores\":[]}"},"finish_reason":"stop"}],"usage":{"prompt_tokens":4,"completion_tokens":3,"total_tokens":7}}`))
	}))
	defer server.Close()

	registryPath, profilePath := writeJudgeRuntimeConfig(t)
	runtime, err := bootstrap.NewClient(bootstrap.Options{
		Config: aiclient.Config{
			AppEnv:               "prod",
			ProviderRegistryPath: registryPath,
			ModelProfilePath:     profilePath,
		},
		SecretSource: mapSecret{
			"AI_PROVIDER_BASE_URL": server.URL,
			"AI_PROVIDER_API_KEY":  "judge-secret",
		},
	})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	defer runtime.Close()

	payload := aiclient.CompletePayload{
		Messages: []aiclient.Message{
			{Role: "system", Content: "score the output against the rubric"},
			{Role: "user", Content: "{\"output\":\"x\"}"},
		},
		Metadata: aiclient.CallMetadata{FeatureKey: "practice.session.follow_up", PromptVersion: "v0.1.0", RubricVersion: "v0.1.0", Language: "multi"},
	}
	resp, meta, err := runtime.Client.CompleteJudge(context.Background(), "judge.default", payload)
	if err != nil {
		t.Fatalf("CompleteJudge: %v", err)
	}
	if !sawPath {
		t.Fatalf("judge adapter did not POST to a chat/completions endpoint")
	}
	if !sawAuth {
		t.Fatalf("judge adapter did not send resolved API key")
	}
	if meta.Capability != aiclient.CapabilityJudge {
		t.Fatalf("meta.Capability: want judge, got %q", meta.Capability)
	}
	if resp.Content == "" {
		t.Fatalf("expected non-empty judge content")
	}
}
