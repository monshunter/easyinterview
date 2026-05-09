package registry

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"time"
)

// RegistryOptions wires NewRegistryClient. PromptsDir / RubricsDir are the
// on-disk truth source roots; Now is overridable so tests can drive the
// 30-second TTL without sleeping.
type RegistryOptions struct {
	PromptsDir string
	RubricsDir string
	Now        func() time.Time
	// CacheTTL overrides the default 30-second TTL; zero or negative
	// disables TTL-driven refresh (tests use this to pin the snapshot).
	CacheTTL time.Duration
}

// Client is the F3 RegistryClient. Construct via NewRegistryClient; never
// build the struct literal directly.
type Client struct {
	cache         *snapshotCache
	fallbackCount atomic.Uint64
}

// NewRegistryClient loads the truth source from disk into the snapshot
// cache. A startup error indicates the caller cannot serve traffic; cmd/api
// must fail fast rather than continue.
func NewRegistryClient(opts RegistryOptions) (*Client, error) {
	if opts.PromptsDir == "" {
		return nil, errors.New("registry: PromptsDir is required")
	}
	if opts.RubricsDir == "" {
		return nil, errors.New("registry: RubricsDir is required")
	}
	ttl := opts.CacheTTL
	if ttl == 0 {
		ttl = 30 * time.Second
	}
	loadFn := func() (*snapshot, error) {
		return loadFromDisk(opts.PromptsDir, opts.RubricsDir)
	}
	cache := newSnapshotCache(loadFn, ttl, opts.Now)
	if err := cache.Reload(context.Background()); err != nil {
		return nil, fmt.Errorf("registry: initial load: %w", err)
	}
	return &Client{cache: cache}, nil
}

// Reload reads the truth source again and atomically swaps the snapshot.
// Tests use it to validate the 30-second TTL contract and to retry after a
// transient disk error.
func (c *Client) Reload(ctx context.Context) error {
	return c.cache.Reload(ctx)
}

// SnapshotSize reports how many (feature_key, language) coordinates are
// loaded. Tests rely on this to assert idempotency and concurrent reload
// invariants without exposing the snapshot struct itself.
func (c *Client) SnapshotSize() int {
	snap := c.cache.Load()
	if snap == nil {
		return 0
	}
	total := 0
	for _, langs := range snap.prompts {
		total += len(langs)
	}
	return total
}
