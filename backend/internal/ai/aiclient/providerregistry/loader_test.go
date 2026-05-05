package providerregistry_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient"
	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient/providerregistry"
)

func writeRegistry(t *testing.T, body string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "ai-providers.yaml")
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	return path
}

func TestLoadParsesProviderRegistrySchema(t *testing.T) {
	path := writeRegistry(t, `providers:
  - name: unit-test-stub
    protocol: stub
    capabilities: [chat, embed]
    version: 1.0.0
  - name: default-openai-compatible
    protocol: openai_compatible
    base_url_env: AI_PROVIDER_BASE_URL
    api_key_env: AI_PROVIDER_API_KEY
    capabilities: [chat, embed]
    version: 1.0.0
`)

	reg, err := providerregistry.Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	stubProvider, ok := reg.Provider("unit-test-stub")
	if !ok {
		t.Fatalf("expected unit-test-stub provider")
	}
	if stubProvider.Protocol != aiclient.ProviderProtocolStub {
		t.Fatalf("expected stub protocol, got %q", stubProvider.Protocol)
	}
	if stubProvider.BaseURLEnv != "" || stubProvider.APIKeyEnv != "" {
		t.Fatalf("stub provider must not require env refs: %+v", stubProvider)
	}
	if !stubProvider.Supports(aiclient.CapabilityChat) || !stubProvider.Supports(aiclient.CapabilityEmbed) {
		t.Fatalf("stub capabilities not parsed: %+v", stubProvider.Capabilities)
	}

	openAIProvider, ok := reg.Provider("default-openai-compatible")
	if !ok {
		t.Fatalf("expected default-openai-compatible provider")
	}
	if openAIProvider.Protocol != aiclient.ProviderProtocolOpenAICompatible {
		t.Fatalf("expected openai_compatible protocol, got %q", openAIProvider.Protocol)
	}
	if openAIProvider.BaseURLEnv != "AI_PROVIDER_BASE_URL" || openAIProvider.APIKeyEnv != "AI_PROVIDER_API_KEY" {
		t.Fatalf("env refs not parsed: %+v", openAIProvider)
	}
}

func TestLoadRejectsSchemaViolations(t *testing.T) {
	cases := map[string]string{
		"duplicate-provider-name": `providers:
  - name: duplicate
    protocol: stub
    capabilities: [chat]
    version: 1.0.0
  - name: duplicate
    protocol: stub
    capabilities: [embed]
    version: 1.0.0
`,
		"unknown-protocol": `providers:
  - name: bad
    protocol: vendor_sdk
    capabilities: [chat]
    version: 1.0.0
`,
		"unknown-capability": `providers:
  - name: bad
    protocol: stub
    capabilities: [image]
    version: 1.0.0
`,
		"network-provider-missing-env-refs": `providers:
  - name: bad
    protocol: openai_compatible
    capabilities: [chat]
    version: 1.0.0
`,
	}

	for label, body := range cases {
		t.Run(label, func(t *testing.T) {
			_, err := providerregistry.Load(writeRegistry(t, body))
			if err == nil {
				t.Fatalf("expected Load to reject %s", label)
			}
		})
	}
}

type mapSecret map[string]string

func (m mapSecret) Get(name string) (string, error) {
	v, ok := m[name]
	if !ok {
		return "", providerregistry.ErrSecretMissing
	}
	return v, nil
}

func TestResolveSelectedProvidersUsesA4SecretSource(t *testing.T) {
	path := writeRegistry(t, `providers:
  - name: unit-test-stub
    protocol: stub
    capabilities: [chat]
    version: 1.0.0
  - name: default-openai-compatible
    protocol: openai_compatible
    base_url_env: AI_PROVIDER_BASE_URL
    api_key_env: AI_PROVIDER_API_KEY
    capabilities: [chat]
    version: 1.0.0
`)
	reg, err := providerregistry.Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	profile := &aiclient.ModelProfile{
		Name:     "practice.followup.default",
		TaskType: aiclient.TaskTypeChat,
		Default:  aiclient.ProviderConfig{Provider: "default-openai-compatible", Model: "chat-model"},
	}
	resolved, err := reg.ResolveSelectedProviders(profile, "prod", mapSecret{
		"AI_PROVIDER_BASE_URL": "https://provider.example",
		"AI_PROVIDER_API_KEY":  "secret",
	})
	if err != nil {
		t.Fatalf("ResolveSelectedProviders: %v", err)
	}
	got := resolved["default-openai-compatible"]
	if got.BaseURL != "https://provider.example" || got.APIKey != "secret" {
		t.Fatalf("secret values not resolved: %+v", got)
	}
}

func TestResolveSelectedProvidersFailFastOnlyForSelectedNonTestNetworkProvider(t *testing.T) {
	path := writeRegistry(t, `providers:
  - name: unit-test-stub
    protocol: stub
    capabilities: [chat]
    version: 1.0.0
  - name: default-openai-compatible
    protocol: openai_compatible
    base_url_env: AI_PROVIDER_BASE_URL
    api_key_env: AI_PROVIDER_API_KEY
    capabilities: [chat]
    version: 1.0.0
`)
	reg, err := providerregistry.Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	stubProfile := &aiclient.ModelProfile{
		Name:     "practice.followup.default",
		TaskType: aiclient.TaskTypeChat,
		Default:  aiclient.ProviderConfig{Provider: "unit-test-stub", Model: "stub-chat"},
	}
	if _, err := reg.ResolveSelectedProviders(stubProfile, "prod", mapSecret{}); err != nil {
		t.Fatalf("unselected network provider must not require secrets: %v", err)
	}

	networkProfile := &aiclient.ModelProfile{
		Name:     "practice.followup.default",
		TaskType: aiclient.TaskTypeChat,
		Default:  aiclient.ProviderConfig{Provider: "default-openai-compatible", Model: "chat-model"},
	}
	if _, err := reg.ResolveSelectedProviders(networkProfile, "prod", mapSecret{}); !errors.Is(err, providerregistry.ErrProviderSecretMissing) {
		t.Fatalf("expected ErrProviderSecretMissing, got %v", err)
	}
	if _, err := reg.ResolveSelectedProviders(networkProfile, aiclient.AppEnvTest, mapSecret{}); err != nil {
		t.Fatalf("test env may resolve selected network provider without actual secret: %v", err)
	}
}

func TestResolveSelectedProvidersRejectsProfileRegistryDrift(t *testing.T) {
	path := writeRegistry(t, `providers:
  - name: unit-test-stub
    protocol: stub
    capabilities: [chat]
    version: 1.0.0
  - name: embed-only
    protocol: stub
    capabilities: [embed]
    version: 1.0.0
`)
	reg, err := providerregistry.Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	cases := map[string]*aiclient.ModelProfile{
		"provider-ref-not-found": {
			Name:     "practice.followup.default",
			TaskType: aiclient.TaskTypeChat,
			Default:  aiclient.ProviderConfig{Provider: "missing-provider", Model: "chat-model"},
		},
		"capability-mismatch": {
			Name:     "practice.followup.default",
			TaskType: aiclient.TaskTypeChat,
			Default:  aiclient.ProviderConfig{Provider: "embed-only", Model: "chat-model"},
		},
		"fallback-over-two-hops": {
			Name:     "practice.followup.default",
			TaskType: aiclient.TaskTypeChat,
			Default:  aiclient.ProviderConfig{Provider: "unit-test-stub", Model: "chat-model"},
			Fallback: []aiclient.FallbackEntry{
				{ProviderConfig: aiclient.ProviderConfig{Provider: "unit-test-stub", Model: "fallback-1"}},
				{ProviderConfig: aiclient.ProviderConfig{Provider: "unit-test-stub", Model: "fallback-2"}},
				{ProviderConfig: aiclient.ProviderConfig{Provider: "unit-test-stub", Model: "fallback-3"}},
			},
		},
	}
	for label, profile := range cases {
		t.Run(label, func(t *testing.T) {
			_, err := reg.ResolveSelectedProviders(profile, "prod", mapSecret{})
			if !errors.Is(err, providerregistry.ErrProviderConfigInvalid) {
				t.Fatalf("expected ErrProviderConfigInvalid, got %v", err)
			}
		})
	}
}

func TestLoaderReloadKeepsOldSnapshotOnFailure(t *testing.T) {
	path := writeRegistry(t, `providers:
  - name: unit-test-stub
    protocol: stub
    capabilities: [chat]
    version: 1.0.0
`)
	loader, err := providerregistry.NewLoader(providerregistry.Options{Path: path, PollInterval: -1})
	if err != nil {
		t.Fatalf("NewLoader: %v", err)
	}
	defer loader.Close()

	initial, ok := loader.Provider("unit-test-stub")
	if !ok {
		t.Fatal("expected initial provider")
	}
	if initial.Version != "1.0.0" {
		t.Fatalf("unexpected initial version: %q", initial.Version)
	}

	if err := os.WriteFile(path, []byte(`providers:
  - name: bad
    protocol: vendor_sdk
    capabilities: [chat]
    version: 2.0.0
`), 0o600); err != nil {
		t.Fatalf("rewrite invalid registry: %v", err)
	}
	if err := loader.Reload(context.Background()); err == nil {
		t.Fatal("expected reload error for invalid registry")
	}
	afterFailure, ok := loader.Provider("unit-test-stub")
	if !ok || afterFailure.Version != "1.0.0" {
		t.Fatalf("failed reload polluted snapshot: ok=%v provider=%+v", ok, afterFailure)
	}
}

func TestLoaderHotReloadPicksUpRegistryEdits(t *testing.T) {
	path := writeRegistry(t, `providers:
  - name: unit-test-stub
    protocol: stub
    capabilities: [chat]
    version: 1.0.0
`)
	loader, err := providerregistry.NewLoader(providerregistry.Options{Path: path, PollInterval: 50 * time.Millisecond})
	if err != nil {
		t.Fatalf("NewLoader: %v", err)
	}
	defer loader.Close()

	if err := os.WriteFile(path, []byte(`providers:
  - name: unit-test-stub
    protocol: stub
    capabilities: [chat, embed]
    version: 1.1.0
`), 0o600); err != nil {
		t.Fatalf("rewrite registry: %v", err)
	}

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		p, ok := loader.Provider("unit-test-stub")
		if ok && p.Version == "1.1.0" && p.Supports(aiclient.CapabilityEmbed) {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("registry hot reload did not converge before deadline")
}
