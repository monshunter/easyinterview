package featureflag_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/platform/featureflag"
)

func TestPostHogProviderCallsDecideEndpoint(t *testing.T) {
	var hits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		if r.URL.Path != "/decide" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("v") != "3" {
			t.Errorf("missing v=3 query param: %s", r.URL.RawQuery)
		}
		body, _ := io.ReadAll(r.Body)
		var payload map[string]any
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Errorf("decode payload: %v", err)
		}
		if payload["api_key"] != "ph-key" {
			t.Errorf("expected api_key=ph-key, got %v", payload["api_key"])
		}
		if payload["distinct_id"] != "anon-1" {
			t.Errorf("expected distinct_id=anon-1, got %v", payload["distinct_id"])
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"featureFlags":{"practice_hint_enabled":true,"ai_fallback_model_enabled":false}}`))
	}))
	defer server.Close()

	p, err := featureflag.NewPostHogProvider(featureflag.PostHogProviderOptions{
		Host:        server.URL,
		APIKey:      "ph-key",
		SelfHosted:  true,
		AppEnv:      "staging",
		Public:      map[string]bool{"practice_hint_enabled": true, "ai_fallback_model_enabled": false},
		CacheTTL:    0,
		HTTPClient:  server.Client(),
		EvalTimeout: 500 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("NewPostHogProvider: %v", err)
	}

	ctx := featureflag.FlagContext{AnonymousDistinctID: "anon-1", AppEnv: "staging"}
	if !p.IsEnabled("practice_hint_enabled", ctx) {
		t.Errorf("expected practice_hint_enabled=true")
	}
	if p.IsEnabled("ai_fallback_model_enabled", ctx) {
		t.Errorf("expected ai_fallback_model_enabled=false")
	}
	snap := p.Snapshot(ctx)
	if !snap["practice_hint_enabled"].Public {
		t.Errorf("practice flag should be marked public from allowlist")
	}
	if snap["ai_fallback_model_enabled"].Public {
		t.Errorf("ai_fallback_model_enabled must remain operator-only")
	}
}

func TestPostHogProviderUsesLastKnownGoodOn5xx(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := calls.Add(1)
		w.Header().Set("Content-Type", "application/json")
		if n == 1 {
			_, _ = w.Write([]byte(`{"featureFlags":{"practice_hint_enabled":true}}`))
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	p, err := featureflag.NewPostHogProvider(featureflag.PostHogProviderOptions{
		Host: server.URL, APIKey: "k", SelfHosted: true, AppEnv: "staging",
		Public:   map[string]bool{"practice_hint_enabled": true},
		CacheTTL: 0, HTTPClient: server.Client(), EvalTimeout: 500 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("NewPostHogProvider: %v", err)
	}

	ctx := featureflag.FlagContext{AnonymousDistinctID: "anon", AppEnv: "staging"}
	if !p.IsEnabled("practice_hint_enabled", ctx) {
		t.Fatal("first call must populate cache")
	}
	// Second call hits 5xx; provider must fall back to last-known-good.
	if !p.IsEnabled("practice_hint_enabled", ctx) {
		t.Fatal("expected last-known-good cache hit on 5xx")
	}
}

func TestPostHogProviderReturnsErrorWhenNoCacheAvailable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	p, err := featureflag.NewPostHogProvider(featureflag.PostHogProviderOptions{
		Host: server.URL, APIKey: "k", SelfHosted: true, AppEnv: "staging",
		Public: map[string]bool{}, CacheTTL: 0, HTTPClient: server.Client(),
	})
	if err != nil {
		t.Fatalf("NewPostHogProvider: %v", err)
	}

	if got := p.IsEnabled("practice_hint_enabled", featureflag.FlagContext{}); got {
		t.Errorf("missing cache + 5xx must return false (degraded)")
	}
	// Snapshot should be empty so runtime-config exposes nothing it cannot vouch for.
	if len(p.Snapshot(featureflag.FlagContext{})) != 0 {
		t.Errorf("snapshot must remain empty on degraded boot")
	}
}

func TestPostHogProviderRejectsConfigWhenSelfHostedFalseInProd(t *testing.T) {
	for _, env := range []string{"staging", "prod"} {
		_, err := featureflag.NewPostHogProvider(featureflag.PostHogProviderOptions{
			Host: "https://posthog.example", APIKey: "k", SelfHosted: false, AppEnv: env,
		})
		if err == nil {
			t.Errorf("APP_ENV=%s with self-hosted=false must fail-fast", env)
		}
		if err != nil && !strings.Contains(err.Error(), "POSTHOG_SELF_HOSTED") {
			t.Errorf("error should mention POSTHOG_SELF_HOSTED: %v", err)
		}
	}
}

func TestPostHogProviderRejectsMissingAPIKey(t *testing.T) {
	_, err := featureflag.NewPostHogProvider(featureflag.PostHogProviderOptions{
		Host: "https://posthog.example", SelfHosted: true, AppEnv: "prod",
	})
	if err == nil {
		t.Fatal("expected missing API key to fail-fast")
	}
	if !strings.Contains(err.Error(), "POSTHOG_PROJECT_API_KEY") {
		t.Fatalf("error should mention POSTHOG_PROJECT_API_KEY, got %v", err)
	}
}

func TestPostHogProviderAllowsSelfHostedFalseInDev(t *testing.T) {
	_, err := featureflag.NewPostHogProvider(featureflag.PostHogProviderOptions{
		Host: "https://posthog.example", APIKey: "k", SelfHosted: false, AppEnv: "dev",
	})
	if err != nil {
		t.Errorf("APP_ENV=dev should permit self-hosted=false; got: %v", err)
	}
}
