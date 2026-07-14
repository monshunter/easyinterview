package config_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/monshunter/easyinterview/backend/internal/platform/config"
	"github.com/monshunter/easyinterview/backend/internal/platform/featureflag"
)

type stubFlags struct {
	snapshot map[string]featureflag.FlagDecision
}

func (s stubFlags) IsEnabled(key string, _ featureflag.FlagContext) bool {
	return s.snapshot[key].Enabled
}
func (s stubFlags) Variant(key string, _ featureflag.FlagContext) string {
	return s.snapshot[key].Variant
}
func (s stubFlags) Snapshot(_ featureflag.FlagContext) map[string]featureflag.FlagDecision {
	out := make(map[string]featureflag.FlagDecision, len(s.snapshot))
	for k, v := range s.snapshot {
		out[k] = v
	}
	return out
}

func newRuntimeLoader(t *testing.T) *config.Loader {
	t.Helper()
	dir := t.TempDir()
	writeYAML(t, filepath.Join(dir, "config.yaml"), `
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

func TestBuildRuntimeConfigAllowlistAndOptOut(t *testing.T) {
	loader := newRuntimeLoader(t)
	flags := stubFlags{snapshot: map[string]featureflag.FlagDecision{
		"report_evidence_v2_enabled":      {Enabled: true, Public: true},
		"report_retry_plan_enabled":       {Enabled: false, Public: true, Variant: "v1"},
		"readiness_signals_enabled":       {Enabled: true, Public: true},
		"ai_fallback_model_enabled":       {Enabled: true, Public: true},
		"mistake_book_export_enabled":     {Enabled: true, Public: true},
		"growth_dashboard_v1_enabled":     {Enabled: true, Public: true},
		"mock_session_dual_track_enabled": {Enabled: true, Public: true},
	}}

	rc := config.BuildRuntimeConfig(context.Background(), config.RuntimeConfigInput{
		Loader:         loader,
		Flags:          flags,
		FlagContext:    featureflag.FlagContext{AppEnv: "dev"},
		AnalyticsOptIn: false,
	})

	if rc.AppVersion != "1.2.3" || rc.DefaultUILanguage != "zh-CN" {
		t.Errorf("public scalars wrong: %+v", rc)
	}
	if rc.AnalyticsEnabled {
		t.Errorf("analyticsEnabled must be false when opt-out")
	}
	if rc.PostHogPublicKey != "" {
		t.Errorf("postHogPublicKey must be empty when opt-out")
	}
	for _, key := range []string{
		"report_evidence_v2_enabled",
		"report_retry_plan_enabled",
		"readiness_signals_enabled",
	} {
		if _, ok := rc.FeatureFlags[key]; !ok {
			t.Errorf("current public flag %s missing", key)
		}
	}
	if _, ok := rc.FeatureFlags["ai_fallback_model_enabled"]; ok {
		t.Errorf("operator-only flag must be filtered")
	}
	for _, key := range []string{
		"mistake_book_export_enabled",
		"growth_dashboard_v1_enabled",
		"mock_session_dual_track_enabled",
	} {
		if _, ok := rc.FeatureFlags[key]; ok {
			t.Errorf("removed product-scope flag %s must be filtered", key)
		}
	}
	if rc.FeatureFlags["report_retry_plan_enabled"].Variant != "v1" {
		t.Errorf("variant pass-through broken")
	}
	limits, err := loader.ContentLimits()
	if err != nil {
		t.Fatalf("ContentLimits: %v", err)
	}
	wantPublicLimits := config.PublicContentLimits{
		ResumeUploadBytes:        limits.ResumeUploadBytes,
		ResumePasteTextBytes:     limits.ResumeMaxPasteTextBytes,
		TargetJobRawTextBytes:    limits.TargetJobMaxRawTextBytes,
		PracticeMessageBytes:     limits.PracticeMaxMessageBytes,
		PracticeSessionTextBytes: limits.PracticeMaxSessionTextBytes,
	}
	if rc.ContentLimits != wantPublicLimits {
		t.Errorf("public content limits wrong: %+v", rc.ContentLimits)
	}

	// JSON marshal must not leak secrets.
	body, err := json.Marshal(rc)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if got := string(body); contains(got, "should-not-leak") {
		t.Errorf("secret leaked into runtime-config JSON: %s", got)
	}
}

func TestBuildRuntimeConfigAnalyticsOptInIncludesPublicKey(t *testing.T) {
	loader := newRuntimeLoader(t)
	rc := config.BuildRuntimeConfig(context.Background(), config.RuntimeConfigInput{
		Loader:         loader,
		Flags:          stubFlags{snapshot: map[string]featureflag.FlagDecision{}},
		AnalyticsOptIn: true,
	})
	if !rc.AnalyticsEnabled {
		t.Errorf("analyticsEnabled must be true on opt-in")
	}
	if rc.PostHogPublicKey != "ph-public" {
		t.Errorf("expected postHogPublicKey to be exposed; got %q", rc.PostHogPublicKey)
	}
}

func TestBuildRuntimeConfigEvaluatesColdPostHogSnapshot(t *testing.T) {
	loader := newRuntimeLoader(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/decide" {
			t.Errorf("unexpected PostHog path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"featureFlags":{"report_evidence_v2_enabled":true,"ai_fallback_model_enabled":true}}`))
	}))
	defer server.Close()

	flags, err := featureflag.NewPostHogProvider(featureflag.PostHogProviderOptions{
		Host:       server.URL,
		APIKey:     "ph-key",
		SelfHosted: true,
		AppEnv:     "prod",
		Public: map[string]bool{
			"report_evidence_v2_enabled": true,
			"ai_fallback_model_enabled":  false,
		},
		HTTPClient: server.Client(),
	})
	if err != nil {
		t.Fatalf("NewPostHogProvider: %v", err)
	}

	rc := config.BuildRuntimeConfig(context.Background(), config.RuntimeConfigInput{
		Loader:      loader,
		Flags:       flags,
		FlagContext: featureflag.FlagContext{AnonymousDistinctID: "anon-1", AppEnv: "prod"},
	})
	if _, ok := rc.FeatureFlags["report_evidence_v2_enabled"]; !ok {
		t.Fatalf("cold PostHog evaluation did not project public flag: %+v", rc.FeatureFlags)
	}
	if _, ok := rc.FeatureFlags["ai_fallback_model_enabled"]; ok {
		t.Fatalf("operator-only flag leaked: %+v", rc.FeatureFlags)
	}
}

func TestRuntimeConfigHandlerReturnsAllowlistJSON(t *testing.T) {
	loader := newRuntimeLoader(t)
	flags := stubFlags{snapshot: map[string]featureflag.FlagDecision{
		"report_evidence_v2_enabled": {Enabled: true, Public: true},
	}}
	handler := config.NewRuntimeConfigHandler(config.RuntimeConfigHandlerOptions{
		Loader: loader,
		Flags:  flags,
		// Nil SessionResolver exercises the anonymous opt-out default.
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/runtime-config", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status: %d body=%s", rec.Code, rec.Body.String())
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("content-type: %q", ct)
	}
	var payload map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode: %v", err)
	}
	for _, key := range []string{"appVersion", "defaultUiLanguage", "analyticsEnabled", "featureFlags", "contentLimits"} {
		if _, ok := payload[key]; !ok {
			t.Errorf("missing key: %s", key)
		}
	}
	limits, ok := payload["contentLimits"].(map[string]any)
	if !ok || len(limits) != 5 {
		t.Fatalf("contentLimits must be a closed five-field object: %#v", payload["contentLimits"])
	}
	for _, internal := range []string{"reportMaxFramedInputBytes", "httpMaxRequestBodyBytes", "aiProviderMaxResponseBodyBytes", "resumeMaxExtractedTextBytes"} {
		if _, leaked := limits[internal]; leaked {
			t.Errorf("internal limit leaked: %s", internal)
		}
	}
	if contains(rec.Body.String(), "should-not-leak") {
		t.Errorf("secret leaked into handler response: %s", rec.Body.String())
	}
}

func contains(haystack, needle string) bool {
	if needle == "" {
		return true
	}
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}
