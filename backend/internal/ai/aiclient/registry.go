package aiclient

// ProviderRegistryEntry mirrors one config/ai-providers.yaml providers[] item.
// The checked-in registry stores env var names only; runtime secret values are
// resolved by A4 SecretSource when a selected provider is materialized.
type ProviderRegistryEntry struct {
	Name         string           `yaml:"name"`
	Protocol     ProviderProtocol `yaml:"protocol"`
	BaseURLEnv   string           `yaml:"base_url_env,omitempty"`
	APIKeyEnv    string           `yaml:"api_key_env,omitempty"`
	Capabilities []Capability     `yaml:"capabilities"`
	Version      string           `yaml:"version"`
}

// Supports reports whether the provider declares the requested capability.
func (p ProviderRegistryEntry) Supports(capability Capability) bool {
	for _, c := range p.Capabilities {
		if c == capability {
			return true
		}
	}
	return false
}
