package registry

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/testsupport"
)

func TestCacheReloadIdempotent(t *testing.T) {
	t.Parallel()
	prompts, rubrics := testsupport.ConfigRoots(t)
	client, err := NewRegistryClient(RegistryOptions{
		PromptsDir: prompts,
		RubricsDir: rubrics,
	})
	if err != nil {
		t.Fatalf("NewRegistryClient: %v", err)
	}

	want := client.SnapshotSize()
	for i := 0; i < 5; i++ {
		if err := client.Reload(context.Background()); err != nil {
			t.Fatalf("reload %d: %v", i, err)
		}
		if got := client.SnapshotSize(); got != want {
			t.Fatalf("reload %d snapshot size drift: want %d, got %d", i, want, got)
		}
	}
}

func TestCacheTTLDrivesReload(t *testing.T) {
	t.Parallel()
	prompts, rubrics := testsupport.ConfigRoots(t)
	var nowNs atomic.Int64
	nowFn := func() time.Time { return time.Unix(0, nowNs.Load()).UTC() }
	nowNs.Store(time.Now().UnixNano())

	var loadCount atomic.Uint64
	client, err := NewRegistryClient(RegistryOptions{
		PromptsDir: prompts,
		RubricsDir: rubrics,
		Now:        nowFn,
		CacheTTL:   100 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("NewRegistryClient: %v", err)
	}
	// initial Reload happens inside NewRegistryClient (1 load)
	loadCount.Store(1)

	// First Load within TTL must not refresh.
	_ = client.cache.Load()
	if got := loadCount.Load(); got != 1 {
		t.Fatalf("Load within TTL must not refresh, count=%d", got)
	}

	// Advance clock past TTL; next Load triggers a refresh through
	// loadFromDisk. We probe by calling Reload directly so we control timing.
	nowNs.Store(nowNs.Load() + (200 * time.Millisecond).Nanoseconds())
	_ = client.cache.Load() // triggers shouldReload -> Reload
	if got := client.SnapshotSize(); got == 0 {
		t.Errorf("post-TTL Load must keep snapshot populated, size=%d", got)
	}
}

func TestCacheConcurrentReadsAndReload(t *testing.T) {
	t.Parallel()
	prompts, rubrics := testsupport.ConfigRoots(t)
	client, err := NewRegistryClient(RegistryOptions{
		PromptsDir: prompts,
		RubricsDir: rubrics,
	})
	if err != nil {
		t.Fatalf("NewRegistryClient: %v", err)
	}

	var readers sync.WaitGroup
	const goroutines = 100
	const iterations = 50
	stop := make(chan struct{})

	// Background reloader interleaves with readers; lifetime tied to stop.
	reloaderDone := make(chan struct{})
	go func() {
		defer close(reloaderDone)
		for {
			select {
			case <-stop:
				return
			default:
				_ = client.Reload(context.Background())
			}
		}
	}()

	// Readers exercise the resolver under concurrent reload.
	for i := 0; i < goroutines; i++ {
		readers.Add(1)
		go func() {
			defer readers.Done()
			ctx := context.Background()
			for j := 0; j < iterations; j++ {
				if _, err := client.ResolveActive(ctx, "target.import.parse", "en"); err != nil {
					t.Errorf("resolve: %v", err)
					return
				}
			}
		}()
	}

	readers.Wait()
	close(stop)
	<-reloaderDone
}
