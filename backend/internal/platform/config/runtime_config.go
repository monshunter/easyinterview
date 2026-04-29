// runtime_config.go owns the public response shape for
// `GET /api/v1/runtime-config` (secrets-and-config spec D-2 / §3.1.2 / C-6).
//
// OpenAPI schema truth source lives in
// docs/spec/openapi-v1-contract (B2). This file persists the field
// allowlist on the backend side; B2 references this builder when freezing
// the OpenAPI schema. A4 must not extend the allowlist below without
// first revising the spec — see plan §5.3.

package config

import (
	"context"

	"github.com/monshunter/easyinterview/backend/internal/platform/featureflag"
)

// RuntimeFlag is the public projection of a feature flag decision. It
// intentionally drops the `Public` field from featureflag.FlagDecision so
// the response cannot accidentally re-expose internal metadata.
type RuntimeFlag struct {
	Enabled bool   `json:"enabled"`
	Variant string `json:"variant,omitempty"`
}

// RuntimeConfig is the JSON shape returned by /api/v1/runtime-config.
// Field set is locked; expansion requires a spec revision (D-2).
type RuntimeConfig struct {
	AppVersion        string                 `json:"appVersion"`
	DefaultUiLanguage string                 `json:"defaultUiLanguage"`
	AnalyticsEnabled  bool                   `json:"analyticsEnabled"`
	FeatureFlags      map[string]RuntimeFlag `json:"featureFlags"`
	PostHogPublicKey  string                 `json:"postHogPublicKey,omitempty"`
}

// RuntimeConfigInput captures everything the builder needs to produce a
// RuntimeConfig response. Loader / Flags are mandatory; AnalyticsOptIn
// reflects the resolved user preference (defaults to false on anonymous
// requests, per spec D-2).
type RuntimeConfigInput struct {
	Loader         *Loader
	Flags          featureflag.FeatureFlagClient
	FlagContext    featureflag.FlagContext
	AnalyticsOptIn bool
}

// BuildRuntimeConfig assembles the response according to the field
// allowlist. Secrets, operator-only flags and unknown fields never enter
// the result. The function is pure given its inputs — useful for unit
// tests that exercise C-6 without going through the HTTP handler.
func BuildRuntimeConfig(_ context.Context, in RuntimeConfigInput) RuntimeConfig {
	rc := RuntimeConfig{
		AnalyticsEnabled: in.AnalyticsOptIn,
		FeatureFlags:     map[string]RuntimeFlag{},
	}
	if in.Loader != nil {
		rc.AppVersion = in.Loader.GetString("runtime.appVersion")
		rc.DefaultUiLanguage = in.Loader.GetString("runtime.defaultUiLanguage")
		if in.AnalyticsOptIn {
			rc.PostHogPublicKey = in.Loader.GetString("featureFlag.posthogPublicKey")
		}
	}
	if snapshotter, ok := in.Flags.(featureflag.SnapshotProvider); ok {
		for key, decision := range snapshotter.Snapshot(in.FlagContext) {
			if !decision.Public {
				continue
			}
			rc.FeatureFlags[key] = RuntimeFlag{
				Enabled: decision.Enabled,
				Variant: decision.Variant,
			}
		}
	}
	return rc
}
