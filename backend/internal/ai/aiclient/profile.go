package aiclient

// ProviderConfig is the (provider, model, params) triple used as either the
// default routing target or a fallback entry inside a ModelProfile.
type ProviderConfig struct {
	ProviderRef string         `yaml:"provider_ref"`
	Model       string         `yaml:"model"`
	Params      map[string]any `yaml:"params,omitempty"`
}

// FallbackEntry is one ordered fallback hop. The When strings are taken
// verbatim from the YAML and consumed by the central fallback policy.
type FallbackEntry struct {
	ProviderConfig `yaml:",inline"`
	When           []string `yaml:"when,omitempty"`
}

// RateLimit captures the per-profile rate-limit hint surfaced to the provider
// endpoint. A3 client does not enforce these locally.
type RateLimit struct {
	RPS int `yaml:"rps,omitempty"`
	TPM int `yaml:"tpm,omitempty"`
}

// ProfileStatus controls whether a profile is executable.
type ProfileStatus string

const (
	ProfileStatusActive      ProfileStatus = "active"
	ProfileStatusDisabled    ProfileStatus = "disabled"
	ProfileStatusUnsupported ProfileStatus = "unsupported"
)

// ModelProfile mirrors the spec §2.1 Model Profile schema. The struct shape
// is the canonical YAML target; new fields require a spec version bump.
type ModelProfile struct {
	Name                string          `yaml:"name"`
	Capability          Capability      `yaml:"capability"`
	Status              ProfileStatus   `yaml:"status"`
	UnsupportedReason   string          `yaml:"unsupported_reason,omitempty"`
	Default             ProviderConfig  `yaml:"default"`
	Fallback            []FallbackEntry `yaml:"fallback,omitempty"`
	TimeoutMs           int             `yaml:"timeout_ms"`
	ContextWindowTokens int             `yaml:"context_window_tokens,omitempty"`
	MaxTokens           int             `yaml:"max_tokens,omitempty"`
	RateLimit           RateLimit       `yaml:"rate_limit,omitempty"`
	Route               string          `yaml:"route,omitempty"`
	Version             string          `yaml:"version"`
	PrivacyPolicy       string          `yaml:"privacy_policy,omitempty"`
}

// ProfileResolver looks up a Model Profile by name. The hot-reloading loader
// implements this interface; tests may use a static map.
type ProfileResolver interface {
	Resolve(name string) (*ModelProfile, error)
}
