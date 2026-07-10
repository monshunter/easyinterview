package providerregistry

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient"
	"github.com/monshunter/easyinterview/backend/internal/platform/config"
)

var (
	// ErrSecretMissing is the A4 SecretSource missing-secret sentinel.
	ErrSecretMissing = config.ErrSecretMissing
	// ErrProviderConfigInvalid marks registry/profile configuration drift.
	ErrProviderConfigInvalid = errors.New("providerregistry: provider config invalid")
	// ErrProviderSecretMissing marks a selected non-test network provider
	// whose declared secret env refs are not provisioned.
	ErrProviderSecretMissing = errors.New("providerregistry: provider secret missing")
)

// ResolvedProvider materializes one provider selected by a profile. BaseURL
// and APIKey are empty for stub providers and may be empty for network
// providers only in APP_ENV=test.
type ResolvedProvider struct {
	Entry   aiclient.ProviderRegistryEntry
	BaseURL string
	APIKey  string
}

// DefaultPollInterval keeps hot reload well inside the ≤30s SLA.
const DefaultPollInterval = 5 * time.Second

// Registry is the parsed provider connection catalog from
// config/ai-providers.yaml.
type Registry struct {
	providers map[string]aiclient.ProviderRegistryEntry
}

// Provider returns the provider entry named by ref.
func (r *Registry) Provider(ref string) (aiclient.ProviderRegistryEntry, bool) {
	if r == nil {
		return aiclient.ProviderRegistryEntry{}, false
	}
	p, ok := r.providers[ref]
	return p, ok
}

// Providers returns provider entries sorted by name.
func (r *Registry) Providers() []aiclient.ProviderRegistryEntry {
	if r == nil {
		return nil
	}
	out := make([]aiclient.ProviderRegistryEntry, 0, len(r.providers))
	for _, p := range r.providers {
		out = append(out, p)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// SecretSource is the A4 runtime secret resolver shape used by provider
// registry materialization.
type SecretSource interface {
	Get(name string) (string, error)
}

// Options controls registry loader construction.
type Options struct {
	Path         string
	PollInterval time.Duration
	OnWarn       func(error)
}

type registrySnapshot struct {
	registry *Registry
	loadedAt time.Time
}

// Loader serves an atomically-swapped provider registry snapshot.
type Loader struct {
	opts Options

	current atomic.Pointer[registrySnapshot]
	reload  sync.Mutex

	stop     chan struct{}
	stopOnce sync.Once
}

// NewLoader loads the registry once and starts the background poller unless
// PollInterval is negative.
func NewLoader(opts Options) (*Loader, error) {
	if opts.Path == "" {
		return nil, fmt.Errorf("providerregistry: Path is required")
	}
	l := &Loader{
		opts: opts,
		stop: make(chan struct{}),
	}
	if err := l.Reload(context.Background()); err != nil {
		return nil, err
	}
	if opts.PollInterval >= 0 {
		go l.pollLoop()
	}
	return l, nil
}

// Close stops the background poller.
func (l *Loader) Close() {
	if l == nil {
		return
	}
	l.stopOnce.Do(func() { close(l.stop) })
}

// Provider returns a provider from the latest good snapshot.
func (l *Loader) Provider(ref string) (aiclient.ProviderRegistryEntry, bool) {
	snap := l.current.Load()
	if snap == nil || snap.registry == nil {
		return aiclient.ProviderRegistryEntry{}, false
	}
	return snap.registry.Provider(ref)
}

// Providers returns sorted providers from the latest good snapshot.
func (l *Loader) Providers() []aiclient.ProviderRegistryEntry {
	snap := l.current.Load()
	if snap == nil || snap.registry == nil {
		return nil
	}
	return snap.registry.Providers()
}

// ResolveSelectedProviders validates selected providers against the latest good
// registry snapshot.
func (l *Loader) ResolveSelectedProviders(profile *aiclient.ModelProfile, appEnv string, secrets SecretSource) (map[string]ResolvedProvider, error) {
	snap := l.current.Load()
	if snap == nil || snap.registry == nil {
		return nil, fmt.Errorf("%w: registry loader is not initialized", ErrProviderConfigInvalid)
	}
	return snap.registry.ResolveSelectedProviders(profile, appEnv, secrets)
}

// LoadedAt returns the timestamp of the last successful registry scan.
func (l *Loader) LoadedAt() time.Time {
	snap := l.current.Load()
	if snap == nil {
		return time.Time{}
	}
	return snap.loadedAt
}

// Reload parses the registry and atomically swaps it in only after full
// validation succeeds.
func (l *Loader) Reload(ctx context.Context) error {
	l.reload.Lock()
	defer l.reload.Unlock()
	if err := ctx.Err(); err != nil {
		return err
	}
	reg, err := Load(l.opts.Path)
	if err != nil {
		return err
	}
	l.current.Store(&registrySnapshot{registry: reg, loadedAt: time.Now()})
	return nil
}

func (l *Loader) pollLoop() {
	interval := l.opts.PollInterval
	if interval == 0 {
		interval = DefaultPollInterval
	}
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		select {
		case <-l.stop:
			return
		case <-t.C:
			if err := l.Reload(context.Background()); err != nil && l.opts.OnWarn != nil {
				l.opts.OnWarn(err)
			}
		}
	}
}

// ResolveSelectedProviders validates that every provider selected by profile
// exists, supports the profile capability, and has required network secrets in
// non-test environments.
func (r *Registry) ResolveSelectedProviders(profile *aiclient.ModelProfile, appEnv string, secrets SecretSource) (map[string]ResolvedProvider, error) {
	if r == nil {
		return nil, fmt.Errorf("%w: registry is nil", ErrProviderConfigInvalid)
	}
	if profile == nil {
		return nil, fmt.Errorf("%w: profile is nil", ErrProviderConfigInvalid)
	}
	if len(profile.Fallback) > 2 {
		return nil, fmt.Errorf("%w: profile %q fallback chain has %d hops; maximum is 2", ErrProviderConfigInvalid, profile.Name, len(profile.Fallback))
	}
	selected := selectedProviderRefs(profile)
	if len(selected) == 0 {
		return nil, fmt.Errorf("%w: profile %q selects no provider", ErrProviderConfigInvalid, profile.Name)
	}
	out := make(map[string]ResolvedProvider, len(selected))
	for _, ref := range selected {
		entry, ok := r.Provider(ref)
		if !ok {
			return nil, fmt.Errorf("%w: profile %q references unknown provider %q", ErrProviderConfigInvalid, profile.Name, ref)
		}
		if !entry.Supports(profile.Capability) {
			return nil, fmt.Errorf("%w: profile %q capability %q is not supported by provider %q", ErrProviderConfigInvalid, profile.Name, profile.Capability, ref)
		}
		resolved, err := ResolveProviderEntry(entry, appEnv, secrets)
		if err != nil {
			return nil, fmt.Errorf("provider %q: %w", ref, err)
		}
		out[ref] = resolved
	}
	return out, nil
}

func selectedProviderRefs(profile *aiclient.ModelProfile) []string {
	seen := map[string]bool{}
	var out []string
	add := func(ref string) {
		ref = strings.TrimSpace(ref)
		if ref == "" || seen[ref] {
			return
		}
		seen[ref] = true
		out = append(out, ref)
	}
	add(profile.Default.ProviderRef)
	for _, fb := range profile.Fallback {
		add(fb.ProviderRef)
	}
	return out
}

func resolveSecret(source SecretSource, name string) (string, error) {
	if source == nil {
		return "", ErrSecretMissing
	}
	value, err := source.Get(name)
	if err != nil {
		return "", err
	}
	if value == "" {
		return "", ErrSecretMissing
	}
	return value, nil
}

// ResolveProviderEntry materializes a single provider entry by resolving its
// env-secret refs through A4 SecretSource.
func ResolveProviderEntry(entry aiclient.ProviderRegistryEntry, appEnv string, secrets SecretSource) (ResolvedProvider, error) {
	resolved := ResolvedProvider{Entry: entry}
	if entry.Protocol == aiclient.ProviderProtocolStub {
		return resolved, nil
	}
	baseURL, err := resolveSecret(secrets, entry.BaseURLEnv)
	if err != nil {
		if appEnv == aiclient.AppEnvTest && errors.Is(err, ErrSecretMissing) {
			return resolved, nil
		}
		return resolved, fmt.Errorf("%w: missing %s", ErrProviderSecretMissing, entry.BaseURLEnv)
	}
	apiKey, err := resolveSecret(secrets, entry.APIKeyEnv)
	if err != nil {
		if appEnv == aiclient.AppEnvTest && errors.Is(err, ErrSecretMissing) {
			return resolved, nil
		}
		return resolved, fmt.Errorf("%w: missing %s", ErrProviderSecretMissing, entry.APIKeyEnv)
	}
	resolved.BaseURL = baseURL
	resolved.APIKey = apiKey
	return resolved, nil
}

// Load parses and validates a provider registry YAML file.
func Load(path string) (*Registry, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("providerregistry: read %s: %w", path, err)
	}
	var raw struct {
		Providers []aiclient.ProviderRegistryEntry `yaml:"providers"`
	}
	if err := yaml.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("providerregistry: parse %s: %w", path, err)
	}
	if len(raw.Providers) == 0 {
		return nil, fmt.Errorf("providerregistry: %s must declare at least one provider", path)
	}
	entries := make(map[string]aiclient.ProviderRegistryEntry, len(raw.Providers))
	for _, p := range raw.Providers {
		if err := validateProvider(p); err != nil {
			return nil, fmt.Errorf("providerregistry: %s: provider %q: %w", path, p.Name, err)
		}
		if _, exists := entries[p.Name]; exists {
			return nil, fmt.Errorf("providerregistry: %s: duplicate provider name %q", path, p.Name)
		}
		entries[p.Name] = p
	}
	return &Registry{providers: entries}, nil
}

func validateProvider(p aiclient.ProviderRegistryEntry) error {
	if p.Name == "" {
		return fmt.Errorf("missing required field 'name'")
	}
	if !isAllowedProtocol(p.Protocol) {
		return fmt.Errorf("unsupported protocol %q", p.Protocol)
	}
	if len(p.Capabilities) == 0 {
		return fmt.Errorf("missing required field 'capabilities'")
	}
	seen := map[aiclient.Capability]bool{}
	for _, c := range p.Capabilities {
		if !isAllowedCapability(c) {
			return fmt.Errorf("unsupported capability %q", c)
		}
		if seen[c] {
			return fmt.Errorf("duplicate capability %q", c)
		}
		seen[c] = true
	}
	if p.Version == "" {
		return fmt.Errorf("missing required field 'version'")
	}
	if p.Protocol != aiclient.ProviderProtocolStub {
		if p.BaseURLEnv == "" || p.APIKeyEnv == "" {
			return fmt.Errorf("network provider must declare base_url_env and api_key_env")
		}
	}
	return nil
}

func isAllowedProtocol(p aiclient.ProviderProtocol) bool {
	switch p {
	case aiclient.ProviderProtocolStub,
		aiclient.ProviderProtocolOpenAICompatible,
		aiclient.ProviderProtocolDoubaoSpeech,
		aiclient.ProviderProtocolMinimaxSpeech,
		aiclient.ProviderProtocolRealtimeAudio,
		aiclient.ProviderProtocolJudgeCompatible:
		return true
	default:
		return false
	}
}

func isAllowedCapability(c aiclient.Capability) bool {
	switch c {
	case aiclient.CapabilityChat,
		aiclient.CapabilitySTT,
		aiclient.CapabilityTts,
		aiclient.CapabilityRealtime,
		aiclient.CapabilityJudge:
		return true
	default:
		return false
	}
}
