package main

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/monshunter/easyinterview/backend/internal/platform/config"
	"github.com/monshunter/easyinterview/backend/internal/platform/featureflag"
)

func TestBuildFlagsClientLoadsPostHogPublicAllowlist(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"featureFlags":{"practice_hint_enabled":true,"ai_fallback_model_enabled":true}}`))
	}))
	defer server.Close()

	dir := t.TempDir()
	flagsPath := filepath.Join(dir, "feature-flags.yaml")
	writeAPIFile(t, flagsPath, `
flags:
  practice_hint_enabled:
    enabled: false
    public: true
  ai_fallback_model_enabled:
    enabled: true
    public: false
`)
	writeAPIFile(t, filepath.Join(dir, "config.yaml"), `
runtime:
  appVersion: "1.2.3"
  defaultUiLanguage: zh-CN
featureFlag:
  source: posthog
  filePath: "`+flagsPath+`"
  posthogHost: "`+server.URL+`"
  posthogSelfHosted: true
  posthogProjectApiKey: "ph-key"
`)
	loader, err := config.Load(config.Options{ConfigDir: dir})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	client, err := buildFlagsClient(loader, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatalf("buildFlagsClient: %v", err)
	}
	rc := config.BuildRuntimeConfig(context.Background(), config.RuntimeConfigInput{
		Loader:      loader,
		Flags:       client,
		FlagContext: featureflag.FlagContext{AnonymousDistinctID: "anon-1", AppEnv: "prod"},
	})
	if _, ok := rc.FeatureFlags["practice_hint_enabled"]; !ok {
		t.Fatalf("public flag missing from runtime-config: %+v", rc.FeatureFlags)
	}
	if _, ok := rc.FeatureFlags["ai_fallback_model_enabled"]; ok {
		t.Fatalf("operator-only flag leaked: %+v", rc.FeatureFlags)
	}
}

func writeAPIFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
