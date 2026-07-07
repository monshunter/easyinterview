package auth_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/auth"
	"github.com/monshunter/easyinterview/backend/internal/platform/config"
	"github.com/monshunter/easyinterview/backend/internal/platform/featureflag"
)

func TestRuntimeConfigSessionResolverOnlyAffectsA4Allowlist(t *testing.T) {
	now := time.Date(2026, 5, 6, 11, 5, 0, 0, time.UTC)
	store := &sessionStore{
		session: auth.SessionRecord{ID: "session-1", UserID: "user-1", Status: auth.SessionStatusActive, ExpiresAt: now.Add(auth.SessionTTL)},
	}
	userStore := &runtimeConfigStore{sessionStore: store, user: auth.UserContext{
		ID:             "user-1",
		Email:          "candidate@example.com",
		AnalyticsOptIn: true,
	}}
	service := auth.NewEmailCodeService(auth.EmailCodeServiceOptions{
		Store:               userStore,
		SessionCookieSecret: "session-secret",
		Now:                 func() time.Time { return now },
	})
	handler := config.NewRuntimeConfigHandler(config.RuntimeConfigHandlerOptions{
		Loader: newRuntimeConfigAuthLoader(t),
		Flags: runtimeFlags{snapshot: map[string]featureflag.FlagDecision{
			"practice_hint_enabled":     {Enabled: true, Public: true},
			"ai_fallback_model_enabled": {Enabled: true, Public: false},
		}},
		FlagContextFunc: func(*http.Request) featureflag.FlagContext {
			return featureflag.FlagContext{AppEnv: "dev"}
		},
		SessionResolver: service.RuntimeConfigSessionResolver(),
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/runtime-config", nil)
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["analyticsEnabled"] != true || body["postHogPublicKey"] != "ph-public" {
		t.Fatalf("session opt-in was not projected through A4 allowlist: %+v", body)
	}
	if contains(rec.Body.String(), "candidate@example.com") || contains(rec.Body.String(), "should-not-leak") {
		t.Fatalf("runtime-config leaked non-allowlisted data: %s", rec.Body.String())
	}
	flags := body["featureFlags"].(map[string]any)
	if _, ok := flags["practice_hint_enabled"]; !ok {
		t.Fatalf("public flag missing: %+v", flags)
	}
	if _, ok := flags["ai_fallback_model_enabled"]; ok {
		t.Fatalf("operator flag leaked: %+v", flags)
	}

	anonymous := httptest.NewRecorder()
	handler.ServeHTTP(anonymous, httptest.NewRequest(http.MethodGet, "/api/v1/runtime-config", nil))
	if contains(anonymous.Body.String(), "ph-public") {
		t.Fatalf("anonymous runtime-config leaked posthog key: %s", anonymous.Body.String())
	}
}

type runtimeConfigStore struct {
	*sessionStore
	user auth.UserContext
}

func (s *runtimeConfigStore) GetUserContext(context.Context, string) (auth.UserContext, error) {
	return s.user, nil
}

type runtimeFlags struct {
	snapshot map[string]featureflag.FlagDecision
}

func (f runtimeFlags) IsEnabled(key string, _ featureflag.FlagContext) bool {
	return f.snapshot[key].Enabled
}

func (f runtimeFlags) Variant(key string, _ featureflag.FlagContext) string {
	return f.snapshot[key].Variant
}

func (f runtimeFlags) Snapshot(_ featureflag.FlagContext) map[string]featureflag.FlagDecision {
	out := make(map[string]featureflag.FlagDecision, len(f.snapshot))
	for k, v := range f.snapshot {
		out[k] = v
	}
	return out
}

func newRuntimeConfigAuthLoader(t *testing.T) *config.Loader {
	t.Helper()
	dir := t.TempDir()
	writeRuntimeFile(t, filepath.Join(dir, "config.yaml"), `
runtime:
  appVersion: "1.2.3"
  defaultUiLanguage: zh-CN
featureFlag:
  posthogPublicKey: "ph-public"
auth:
  sessionCookieSecret: "should-not-leak"
`)
	loader, err := config.Load(config.Options{ConfigDir: dir})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	return loader
}

func writeRuntimeFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
