// Package featureflag exposes the FeatureFlagClient interface (spec D-4)
// and its P0 providers: FileFlagProvider (config/feature-flags.yaml,
// dev / unit-test default) and PostHogFlagProvider (self-hosted PostHog,
// staging / prod). Business code must depend on the interface only and
// never import github.com/posthog/posthog-go directly (Phase 4 lint
// enforces the boundary).
package featureflag
