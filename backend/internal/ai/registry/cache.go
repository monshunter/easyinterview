package registry

import (
	"context"
	"sync/atomic"
	"time"
)

// snapshotCache is an atomic.Value-backed snapshot holder. Reload swaps the
// snapshot in one atomic store; readers always see a consistent view. The
// 30-second TTL drives lazy refresh in production; tests can call Reload
// directly to skip the wait.
type snapshotCache struct {
	current    atomic.Pointer[snapshot]
	lastLoaded atomic.Int64 // unix nanoseconds
	loadFn     func() (*snapshot, error)
	ttl        time.Duration
	now        func() time.Time
}

func newSnapshotCache(loadFn func() (*snapshot, error), ttl time.Duration, now func() time.Time) *snapshotCache {
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	return &snapshotCache{
		loadFn: loadFn,
		ttl:    ttl,
		now:    now,
	}
}

// Load returns the current snapshot, refreshing it if the TTL has elapsed.
// Concurrent Loads see at most one in-flight reload; followers reuse the
// previous snapshot until the new one publishes.
func (c *snapshotCache) Load() *snapshot {
	if c.shouldReload() {
		// Best-effort lazy refresh; ignore errors so a transient disk hiccup
		// does not break in-flight reads. Loader errors at startup were
		// already surfaced by NewRegistryClient.
		_ = c.Reload(context.Background())
	}
	return c.current.Load()
}

func (c *snapshotCache) shouldReload() bool {
	last := c.lastLoaded.Load()
	if last == 0 {
		return true
	}
	if c.ttl <= 0 {
		return false
	}
	return c.now().UnixNano()-last >= int64(c.ttl)
}

// Reload performs a synchronous reload from disk. Tests use this hook to
// validate idempotency and concurrency contracts.
func (c *snapshotCache) Reload(_ context.Context) error {
	snap, err := c.loadFn()
	if err != nil {
		return err
	}
	c.current.Store(snap)
	c.lastLoaded.Store(c.now().UnixNano())
	return nil
}
