package bootstrap_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient"
	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient/bootstrap"
	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient/providerregistry"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
)

type mapSecret map[string]string

func (m mapSecret) Get(name string) (string, error) {
	value, ok := m[name]
	if !ok {
		return "", providerregistry.ErrSecretMissing
	}
	return value, nil
}

func TestNewClientFailsFastWhenSelectedProviderSecretMissing(t *testing.T) {
	registryPath, profilePath := writeRuntimeConfig(t, "deepseek")

	_, err := bootstrap.NewClient(bootstrap.Options{
		Config: aiclient.Config{
			AppEnv:               "prod",
			ProviderRegistryPath: registryPath,
			ModelProfilePath:     profilePath,
		},
		SecretSource: mapSecret{},
	})
	if !errors.Is(err, providerregistry.ErrProviderSecretMissing) {
		t.Fatalf("expected ErrProviderSecretMissing, got %v", err)
	}
}

func TestNewClientLoadsRegistryProfileAndRoutesThroughProviderRef(t *testing.T) {
	registryPath, profilePath := writeRuntimeConfig(t, "deepseek")
	var sawAuth bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got == "Bearer runtime-secret" {
			sawAuth = true
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"model":"chat-runtime-2026-05-05","choices":[{"message":{"content":"runtime ok"},"finish_reason":"stop"}],"usage":{"prompt_tokens":3,"completion_tokens":2,"total_tokens":5}}`))
	}))
	defer server.Close()

	runtime, err := bootstrap.NewClient(bootstrap.Options{
		Config: aiclient.Config{
			AppEnv:               "prod",
			ProviderRegistryPath: registryPath,
			ModelProfilePath:     profilePath,
		},
		SecretSource: mapSecret{
			"AI_PROVIDER_BASE_URL": server.URL,
			"AI_PROVIDER_API_KEY":  "runtime-secret",
		},
	})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	defer runtime.Close()

	resp, meta, err := runtime.Client.Complete(context.Background(), "practice.followup.default", aiclient.CompletePayload{
		Messages: []aiclient.Message{{Role: "user", Content: "hello"}},
		Metadata: aiclient.CallMetadata{
			FeatureKey:    "practice.session.follow_up",
			PromptVersion: "p1",
			RubricVersion: "r1",
			Language:      "en",
		},
	})
	if err != nil {
		t.Fatalf("Complete: %v", err)
	}
	if !sawAuth {
		t.Fatal("provider adapter did not use API key resolved from registry secret ref")
	}
	if resp.Content != "runtime ok" {
		t.Fatalf("unexpected response %q", resp.Content)
	}
	if meta.Provider != "deepseek" || meta.ModelProfileName != "practice.followup.default" {
		t.Fatalf("meta not routed through provider ref/profile: %+v", meta)
	}
}

func TestNewClientRejectsActiveStubProfileOutsideTest(t *testing.T) {
	registryPath, profilePath := writeRuntimeConfig(t, "unit-test-stub")

	_, err := bootstrap.NewClient(bootstrap.Options{
		Config: aiclient.Config{
			AppEnv:               "prod",
			ProviderRegistryPath: registryPath,
			ModelProfilePath:     profilePath,
		},
		SecretSource: mapSecret{},
	})
	if !errors.Is(err, providerregistry.ErrProviderConfigInvalid) {
		t.Fatalf("expected ErrProviderConfigInvalid for active stub profile outside test, got %v", err)
	}
	if code := providerregistry.SharedErrorCode(err); code != sharederrors.CodeAiProviderConfigInvalid {
		t.Fatalf("expected shared config-invalid code, got %q", code)
	}
}

func writeRuntimeConfig(t *testing.T, providerRef string) (string, string) {
	t.Helper()
	dir := t.TempDir()
	registryPath := filepath.Join(dir, "ai-providers.yaml")
	if err := os.WriteFile(registryPath, []byte(`providers:
  - name: unit-test-stub
    protocol: stub
    capabilities: [chat, stt]
    version: 1.0.0
  - name: deepseek
    protocol: openai_compatible
    base_url_env: AI_PROVIDER_BASE_URL
    api_key_env: AI_PROVIDER_API_KEY
    capabilities: [chat]
    version: 1.0.0
`), 0o600); err != nil {
		t.Fatalf("write registry: %v", err)
	}
	profilePath := filepath.Join(dir, "ai-profiles.yaml")
	profileBody := `profiles:
  - name: practice.followup.default
    capability: chat
    status: active
    default:
      provider_ref: ` + providerRef + `
      model: chat-runtime-2026-05-05
    timeout_ms: 5000
    route: practice.followup
    version: 1.0.0
`
	if err := os.WriteFile(profilePath, []byte(profileBody), 0o600); err != nil {
		t.Fatalf("write profile: %v", err)
	}
	return registryPath, profilePath
}
