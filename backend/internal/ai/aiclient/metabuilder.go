package aiclient

import "fmt"

// metaBuilder merges profile-derived fields, business CallMetadata, and the
// provider's partial AICallMeta into a single canonical AICallMeta.
// Business code never holds a metaBuilder.
type metaBuilder struct{}

// merge produces the final AICallMeta. The provider's partial meta wins for
// Provider / ModelFamily / ModelID / InputTokens / OutputTokens /
// CostUSDMicros / LatencyMs / FallbackChain / Route / ErrorCode (everything
// the upstream knows). Profile fields and CallMetadata fill the rest. The
// builder validates that mandatory fields are populated for success cases.
func (b metaBuilder) merge(profile *ModelProfile, callMeta CallMetadata, partial AICallMeta) (AICallMeta, error) {
	out := partial
	out.Capability = profile.Capability
	out.ModelProfileName = profile.Name
	out.ModelProfileVersion = profile.Version
	out.PromptVersion = callMeta.PromptVersion
	out.RubricVersion = callMeta.RubricVersion
	out.Language = callMeta.Language
	out.FeatureKey = callMeta.FeatureKey
	// FeatureFlag defaults to the documented "none" sentinel when no
	// PostHog flag is active so AITaskRunRow + audit pipeline carry a
	// non-empty value (B4 typed column has the same default).
	if callMeta.FeatureFlag == "" {
		out.FeatureFlag = "none"
	} else {
		out.FeatureFlag = callMeta.FeatureFlag
	}
	// DataSourceVersion defaults to "not_applicable" with the same
	// rationale; F3 baseline ships a real value, but call sites that
	// have no source (for example synthetic eval calls) emit the sentinel.
	if callMeta.DataSourceVersion == "" {
		out.DataSourceVersion = "not_applicable"
	} else {
		out.DataSourceVersion = callMeta.DataSourceVersion
	}

	if out.Provider == "" {
		out.Provider = profile.Default.ProviderRef
	}
	if out.ModelID == "" {
		out.ModelID = profile.Default.Model
	}
	if out.Route == "" {
		out.Route = profile.Route
	}
	if len(out.FallbackChain) == 0 {
		out.FallbackChain = []string{profile.Default.ProviderRef}
	}
	if out.ValidationStatus == "" && out.ErrorCode == "" {
		out.ValidationStatus = ValidationStatusOK
	}

	if out.ErrorCode == "" {
		if out.Provider == "" {
			return AICallMeta{}, fmt.Errorf("aiclient: provider missing in success meta")
		}
		if out.ModelID == "" {
			return AICallMeta{}, fmt.Errorf("aiclient: model_id missing in success meta")
		}
		if out.FeatureKey == "" {
			return AICallMeta{}, fmt.Errorf("aiclient: feature_key missing in success meta")
		}
	}
	return out, nil
}
