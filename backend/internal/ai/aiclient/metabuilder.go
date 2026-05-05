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
	}
	return out, nil
}
