package profile

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient"
)

// DefaultPollInterval is the polling cadence when the loader runs in
// background mode without an external watcher. The 5 s default keeps the
// hot-reload SLA (≤30 s, spec §6 C-4) comfortably satisfied.
const DefaultPollInterval = 5 * time.Second

// Options control loader construction. Callers typically use NewLoader.
type Options struct {
	// Path is the single YAML profile catalog file
	// (AI_MODEL_PROFILE_PATH).
	Path string
	// PollInterval overrides DefaultPollInterval. Zero falls back to the
	// default; negative disables background reload (Reload(ctx) still works).
	PollInterval time.Duration
	// Now is injectable for tests; nil falls back to time.Now.
	Now func() time.Time
	// OnWarn receives reload errors from the background poller after the old
	// snapshot has been preserved.
	OnWarn func(error)
}

// snapshot holds an immutable view of the loaded profiles plus the last
// scan timestamp.
type snapshot struct {
	profiles map[string]*aiclient.ModelProfile
	loadedAt time.Time
}

// Loader reads and serves the Model Profile YAML catalog. It implements
// aiclient.ProfileResolver.
type Loader struct {
	opts Options

	current atomic.Pointer[snapshot]
	// reload guards single-flight reload execution (the public Reload may
	// race with the background poller).
	reload sync.Mutex

	// stop signals the background poller to exit.
	stop     chan struct{}
	stopOnce sync.Once
}

// NewLoader constructs a Loader and runs an initial scan synchronously so
// callers that do `Resolve` immediately see profiles.
func NewLoader(opts Options) (*Loader, error) {
	if opts.Path == "" {
		return nil, errors.New("profile: Path is required")
	}
	if opts.Now == nil {
		opts.Now = time.Now
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

// Close stops the background poller. Subsequent Resolve calls keep working
// against the last snapshot; subsequent Reload calls still execute on
// demand.
func (l *Loader) Close() {
	l.stopOnce.Do(func() { close(l.stop) })
}

// Resolve implements aiclient.ProfileResolver.
func (l *Loader) Resolve(name string) (*aiclient.ModelProfile, error) {
	snap := l.current.Load()
	if snap == nil {
		return nil, fmt.Errorf("profile: loader not initialized")
	}
	p, ok := snap.profiles[name]
	if !ok {
		return nil, fmt.Errorf("profile: %q not found in %s", name, l.opts.Path)
	}
	return p, nil
}

// Names returns the sorted list of profile names from the latest snapshot.
func (l *Loader) Names() []string {
	snap := l.current.Load()
	if snap == nil {
		return nil
	}
	names := make([]string, 0, len(snap.profiles))
	for n := range snap.profiles {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

// Reload re-scans the profile catalog and atomically swaps in a new
// snapshot. In-flight callers that have already captured a *ModelProfile
// pointer keep using it; subsequent Resolve calls observe the new snapshot.
func (l *Loader) Reload(ctx context.Context) error {
	l.reload.Lock()
	defer l.reload.Unlock()

	if err := ctx.Err(); err != nil {
		return err
	}

	loaded, err := readCatalog(l.opts.Path)
	if err != nil {
		return err
	}
	profiles := make(map[string]*aiclient.ModelProfile, len(loaded))
	for _, p := range loaded {
		if existing, dup := profiles[p.Name]; dup {
			return fmt.Errorf("profile: duplicate name %q in %s; first duplicate was %q", p.Name, l.opts.Path, existing.Name)
		}
		profiles[p.Name] = p
	}
	l.current.Store(&snapshot{
		profiles: profiles,
		loadedAt: l.opts.Now(),
	})
	return nil
}

// LoadedAt returns the timestamp of the last successful scan. The hot-reload
// concurrency test uses this to assert ≤30 s convergence.
func (l *Loader) LoadedAt() time.Time {
	snap := l.current.Load()
	if snap == nil {
		return time.Time{}
	}
	return snap.loadedAt
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

func readCatalog(path string) ([]*aiclient.ModelProfile, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("profile: open %s: %w", path, err)
	}
	defer f.Close()
	body, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("profile: read %s: %w", path, err)
	}
	var doc yaml.Node
	dec := yaml.NewDecoder(strings.NewReader(string(body)))
	dec.KnownFields(false)
	if err := dec.Decode(&doc); err != nil {
		return nil, fmt.Errorf("profile: parse %s: %w", path, err)
	}
	root := yamlRoot(&doc)
	if root == nil || root.Kind != yaml.MappingNode {
		return nil, profileValidationError(path, yamlNodeLine(root), "catalog must be a mapping with top-level 'profiles'")
	}

	var profilesNode *yaml.Node
	for i := 0; i+1 < len(root.Content); i += 2 {
		key := root.Content[i]
		value := root.Content[i+1]
		switch key.Value {
		case "profiles":
			profilesNode = value
		default:
			return nil, profileValidationError(path, yamlNodeLine(key), "unsupported top-level field %q; use 'profiles'", key.Value)
		}
	}
	if profilesNode == nil {
		return nil, profileValidationError(path, yamlNodeLine(root), "missing required top-level field 'profiles'")
	}
	if profilesNode.Kind != yaml.SequenceNode {
		return nil, profileValidationError(path, yamlNodeLine(profilesNode), "'profiles' must be a sequence")
	}
	if len(profilesNode.Content) == 0 {
		return nil, profileValidationError(path, yamlNodeLine(profilesNode), "'profiles' must contain at least one profile")
	}

	out := make([]*aiclient.ModelProfile, 0, len(profilesNode.Content))
	for _, item := range profilesNode.Content {
		if item.Kind != yaml.MappingNode {
			return nil, profileValidationError(path, yamlNodeLine(item), "profile entry must be a mapping")
		}
		if err := rejectRetiredSchemaKeys(path, item); err != nil {
			return nil, err
		}
		var raw aiclient.ModelProfile
		if err := item.Decode(&raw); err != nil {
			return nil, fmt.Errorf("profile: parse %s: %w", path, err)
		}
		if err := validateProfile(path, item, &raw); err != nil {
			return nil, err
		}
		out = append(out, &raw)
	}
	return out, nil
}

func validateProfile(path string, doc *yaml.Node, raw *aiclient.ModelProfile) error {
	if raw.Name == "" {
		return profileValidationError(path, fieldLine(doc, "name"), "missing required field 'name'")
	}
	if raw.Capability == "" {
		return profileValidationError(path, fieldLine(doc, "capability"), "missing required field 'capability'")
	}
	switch raw.Capability {
	case aiclient.CapabilityChat,
		aiclient.CapabilitySTT,
		aiclient.CapabilityRealtime,
		aiclient.CapabilityJudge:
	default:
		return profileValidationError(path, fieldLine(doc, "capability"), "has unsupported capability %q (allowed: chat | stt | realtime | judge)", raw.Capability)
	}
	if raw.Status == "" {
		return profileValidationError(path, fieldLine(doc, "status"), "missing required field 'status'")
	}
	switch raw.Status {
	case aiclient.ProfileStatusActive:
	case aiclient.ProfileStatusDisabled, aiclient.ProfileStatusUnsupported:
		if strings.TrimSpace(raw.UnsupportedReason) == "" {
			return profileValidationError(path, fieldLine(doc, "unsupported_reason"), "status %q requires 'unsupported_reason'", raw.Status)
		}
	default:
		return profileValidationError(path, fieldLine(doc, "status"), "has unsupported status %q (allowed: active | disabled | unsupported)", raw.Status)
	}
	if raw.Default.ProviderRef == "" {
		return profileValidationError(path, fieldLine(doc, "default", "provider_ref"), "missing required field 'default.provider_ref'")
	}
	if raw.Default.Model == "" {
		return profileValidationError(path, fieldLine(doc, "default", "model"), "missing required field 'default.model'")
	}
	if raw.TimeoutMs <= 0 {
		return profileValidationError(path, fieldLine(doc, "timeout_ms"), "missing or non-positive 'timeout_ms'")
	}
	if raw.Version == "" {
		return profileValidationError(path, fieldLine(doc, "version"), "missing required field 'version'")
	}
	return nil
}

func rejectRetiredSchemaKeys(path string, doc *yaml.Node) error {
	root := yamlRoot(doc)
	if root == nil || root.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i+1 < len(root.Content); i += 2 {
		key := root.Content[i]
		value := root.Content[i+1]
		switch key.Value {
		case "task_type":
			return profileValidationError(path, yamlNodeLine(key), "retired schema key 'task_type'; use 'capability'")
		case "default":
			if err := rejectRetiredMappingKey(path, value, "provider", "default.provider", "default.provider_ref"); err != nil {
				return err
			}
		case "fallback":
			if value.Kind != yaml.SequenceNode {
				continue
			}
			for _, item := range value.Content {
				if err := rejectRetiredMappingKey(path, item, "provider", "fallback[].provider", "fallback[].provider_ref"); err != nil {
					return err
				}
				if err := rejectRetiredMappingKey(path, item, "trigger", "fallback[].trigger", "fallback[].when"); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func rejectRetiredMappingKey(path string, node *yaml.Node, keyName, retired, replacement string) error {
	if node == nil || node.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i+1 < len(node.Content); i += 2 {
		key := node.Content[i]
		if key.Value == keyName {
			return profileValidationError(path, yamlNodeLine(key), "retired schema key '%s'; use '%s'", retired, replacement)
		}
	}
	return nil
}

func profileValidationError(path string, line int, format string, args ...any) error {
	if line <= 0 {
		line = 1
	}
	return fmt.Errorf("profile: %s:line %d %s", path, line, fmt.Sprintf(format, args...))
}

func fieldLine(doc *yaml.Node, path ...string) int {
	current := yamlRoot(doc)
	lastLine := yamlNodeLine(current)
	for _, part := range path {
		if current == nil || current.Kind != yaml.MappingNode {
			return lastLine
		}
		found := false
		for i := 0; i+1 < len(current.Content); i += 2 {
			key := current.Content[i]
			value := current.Content[i+1]
			if key.Value == part {
				lastLine = yamlNodeLine(key)
				current = value
				found = true
				break
			}
		}
		if !found {
			return lastLine
		}
	}
	return lastLine
}

func yamlRoot(doc *yaml.Node) *yaml.Node {
	if doc == nil {
		return nil
	}
	if doc.Kind == yaml.DocumentNode && len(doc.Content) > 0 {
		return doc.Content[0]
	}
	return doc
}

func yamlNodeLine(node *yaml.Node) int {
	if node == nil || node.Line <= 0 {
		return 1
	}
	return node.Line
}
