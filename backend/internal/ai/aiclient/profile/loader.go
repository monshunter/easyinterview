package profile

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
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
	// Dir is the directory holding *.yaml profile files
	// (AI_MODEL_PROFILE_PATH).
	Dir string
	// PollInterval overrides DefaultPollInterval. Zero falls back to the
	// default; negative disables background reload (Reload(ctx) still works).
	PollInterval time.Duration
	// Now is injectable for tests; nil falls back to time.Now.
	Now func() time.Time
}

// snapshot holds an immutable view of the loaded profiles plus the last
// scan timestamp.
type snapshot struct {
	profiles map[string]*aiclient.ModelProfile
	loadedAt time.Time
}

// Loader reads and serves Model Profile YAML files. It implements
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
	if opts.Dir == "" {
		return nil, errors.New("profile: Dir is required")
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
		return nil, fmt.Errorf("profile: %q not found in %s", name, l.opts.Dir)
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

// Reload re-scans the profile directory and atomically swaps in a new
// snapshot. In-flight callers that have already captured a *ModelProfile
// pointer keep using it; subsequent Resolve calls observe the new snapshot.
func (l *Loader) Reload(ctx context.Context) error {
	l.reload.Lock()
	defer l.reload.Unlock()

	if err := ctx.Err(); err != nil {
		return err
	}

	files, err := listYAMLFiles(l.opts.Dir)
	if err != nil {
		return err
	}
	profiles := make(map[string]*aiclient.ModelProfile, len(files))
	for _, path := range files {
		p, err := readProfile(path)
		if err != nil {
			return err
		}
		if existing, dup := profiles[p.Name]; dup {
			return fmt.Errorf("profile: duplicate name %q in %s and %s", p.Name, path, existing.Name)
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
			_ = l.Reload(context.Background())
		}
	}
}

func listYAMLFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("profile: read dir %s: %w", dir, err)
	}
	var out []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, ".yaml") && !strings.HasSuffix(name, ".yml") {
			continue
		}
		out = append(out, filepath.Join(dir, name))
	}
	sort.Strings(out)
	return out, nil
}

func readProfile(path string) (*aiclient.ModelProfile, error) {
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
	var raw aiclient.ModelProfile
	if err := doc.Decode(&raw); err != nil {
		return nil, fmt.Errorf("profile: parse %s: %w", path, err)
	}
	if raw.Name == "" {
		return nil, profileValidationError(path, fieldLine(&doc, "name"), "missing required field 'name'")
	}
	if raw.TaskType == "" {
		return nil, profileValidationError(path, fieldLine(&doc, "task_type"), "missing required field 'task_type'")
	}
	switch raw.TaskType {
	case aiclient.TaskTypeChat, aiclient.TaskTypeEmbed, aiclient.TaskTypeSTT:
	default:
		return nil, profileValidationError(path, fieldLine(&doc, "task_type"), "has unsupported task_type %q (allowed: chat | embed | stt)", raw.TaskType)
	}
	if raw.Default.Provider == "" {
		return nil, profileValidationError(path, fieldLine(&doc, "default", "provider"), "missing required field 'default.provider'")
	}
	if raw.Default.Model == "" {
		return nil, profileValidationError(path, fieldLine(&doc, "default", "model"), "missing required field 'default.model'")
	}
	if raw.TimeoutMs <= 0 {
		return nil, profileValidationError(path, fieldLine(&doc, "timeout_ms"), "missing or non-positive 'timeout_ms'")
	}
	if raw.Version == "" {
		return nil, profileValidationError(path, fieldLine(&doc, "version"), "missing required field 'version'")
	}
	return &raw, nil
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
