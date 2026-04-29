package featureflag

import (
	"crypto/sha256"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

// defaultReloadInterval is the upper bound implied by spec D-7 (≤ 30s).
// Tests may override via FileProviderOptions.ReloadInterval.
const defaultReloadInterval = 30 * time.Second

// FileProviderOptions configures FileFlagProvider.
type FileProviderOptions struct {
	// Path points to config/feature-flags.yaml (or test fixture).
	Path string
	// ReloadInterval bounds the polling cadence used to detect mtime/hash
	// changes. Defaults to 30 seconds (spec D-7).
	ReloadInterval time.Duration
	// Logger receives warnings on parse failures. Defaults to slog.Default().
	Logger *slog.Logger
}

// FileFlagProvider reads flags from a YAML file and hot-reloads on
// mtime/content changes. Schema: `flags: { <key>: { enabled, variant?,
// public } }`. The provider holds the current evaluated map under
// sync.RWMutex; reads are lock-free fast path via Snapshot copies.
type FileFlagProvider struct {
	path           string
	reloadInterval time.Duration
	logger         *slog.Logger

	mu    sync.RWMutex
	flags map[string]FlagDecision

	lastModTime time.Time
	lastHash    [32]byte

	stop chan struct{}
	wg   sync.WaitGroup
}

// NewFileProvider loads the file synchronously and starts a polling
// goroutine. Returns an error if the initial load fails so misconfigured
// processes fail-fast (spec C-3 prerequisite).
func NewFileProvider(opts FileProviderOptions) (*FileFlagProvider, error) {
	if opts.Path == "" {
		return nil, fmt.Errorf("featureflag: file provider path is required")
	}
	interval := opts.ReloadInterval
	if interval <= 0 {
		interval = defaultReloadInterval
	}
	logger := opts.Logger
	if logger == nil {
		logger = slog.Default()
	}
	p := &FileFlagProvider{
		path:           opts.Path,
		reloadInterval: interval,
		logger:         logger,
		flags:          map[string]FlagDecision{},
		stop:           make(chan struct{}),
	}
	if err := p.loadOnce(); err != nil {
		return nil, err
	}
	p.wg.Add(1)
	go p.watchLoop()
	return p, nil
}

// Close stops the polling goroutine. Safe to call multiple times.
func (p *FileFlagProvider) Close() {
	select {
	case <-p.stop:
		return
	default:
	}
	close(p.stop)
	p.wg.Wait()
}

// IsEnabled implements FeatureFlagClient.
func (p *FileFlagProvider) IsEnabled(key string, _ FlagContext) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.flags[key].Enabled
}

// Variant implements FeatureFlagClient.
func (p *FileFlagProvider) Variant(key string, _ FlagContext) string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.flags[key].Variant
}

// Snapshot returns a defensive copy of the current evaluated map. File-backed
// flags do not depend on FlagContext, but the argument keeps the runtime-config
// projection contract consistent with PostHog.
func (p *FileFlagProvider) Snapshot(_ FlagContext) map[string]FlagDecision {
	p.mu.RLock()
	defer p.mu.RUnlock()
	out := make(map[string]FlagDecision, len(p.flags))
	for k, v := range p.flags {
		out[k] = v
	}
	return out
}

// LoadPublicFlagMap reads a feature-flags.yaml file and returns key -> public.
// PostHog-backed runtime-config uses the local baseline as the visibility
// allowlist so first-request projections can still filter operator-only flags.
func LoadPublicFlagMap(path string) (map[string]bool, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	flags, err := parseFeatureFlags(raw)
	if err != nil {
		return nil, err
	}
	out := make(map[string]bool, len(flags))
	for key, decision := range flags {
		out[key] = decision.Public
	}
	return out, nil
}

func (p *FileFlagProvider) watchLoop() {
	defer p.wg.Done()
	ticker := time.NewTicker(p.reloadInterval)
	defer ticker.Stop()
	for {
		select {
		case <-p.stop:
			return
		case <-ticker.C:
			if err := p.maybeReload(); err != nil {
				p.logger.Warn("featureflag: file reload failed; keeping last snapshot", "path", p.path, "error", err.Error())
			}
		}
	}
}

func (p *FileFlagProvider) maybeReload() error {
	info, err := os.Stat(p.path)
	if err != nil {
		return err
	}
	if info.ModTime().Equal(p.lastModTime) {
		return nil
	}
	return p.loadOnce()
}

type fileSchema struct {
	Flags map[string]struct {
		Enabled bool   `yaml:"enabled"`
		Variant string `yaml:"variant"`
		Public  bool   `yaml:"public"`
	} `yaml:"flags"`
}

func (p *FileFlagProvider) loadOnce() error {
	raw, err := os.ReadFile(p.path)
	if err != nil {
		return err
	}
	hash := sha256.Sum256(raw)
	if hash == p.lastHash {
		// Same content; refresh mtime so future reloads compare from here.
		info, statErr := os.Stat(p.path)
		if statErr == nil {
			p.lastModTime = info.ModTime()
		}
		return nil
	}
	next, err := parseFeatureFlags(raw)
	if err != nil {
		return err
	}
	p.mu.Lock()
	p.flags = next
	p.mu.Unlock()
	p.lastHash = hash
	if info, statErr := os.Stat(p.path); statErr == nil {
		p.lastModTime = info.ModTime()
	}
	return nil
}

func parseFeatureFlags(raw []byte) (map[string]FlagDecision, error) {
	var probe map[string]any
	if err := yaml.Unmarshal(raw, &probe); err != nil {
		return nil, fmt.Errorf("parse feature-flags.yaml: %w", err)
	}
	if _, ok := probe["flags"]; !ok {
		return nil, fmt.Errorf("parse feature-flags.yaml: missing top-level `flags` key")
	}
	var schema fileSchema
	if err := yaml.Unmarshal(raw, &schema); err != nil {
		return nil, fmt.Errorf("parse feature-flags.yaml: %w", err)
	}
	next := make(map[string]FlagDecision, len(schema.Flags))
	for key, entry := range schema.Flags {
		next[key] = FlagDecision{Enabled: entry.Enabled, Variant: entry.Variant, Public: entry.Public}
	}
	return next, nil
}
