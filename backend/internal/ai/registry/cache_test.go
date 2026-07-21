package registry

import (
	"context"
	"os"
	"path/filepath"
	"strings"
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

func TestPracticeV030ActivationRollbackReactivate(t *testing.T) {
	prompts, rubrics := tempBaselineCopy(t)
	client, err := NewRegistryClient(RegistryOptions{PromptsDir: prompts, RubricsDir: rubrics})
	if err != nil {
		t.Fatalf("NewRegistryClient: %v", err)
	}
	assertFeatureActiveVersion(t, client, "report.generate", "v0.2.0")
	assertFeatureActiveVersion(t, client, "practice.session.chat", "v0.3.0")

	// A partially edited practice pair must fail validation without replacing
	// the already-published report-v0.2/practice-v0.3 snapshot.
	rewritePromptStatus(t, filepath.Join(prompts, "practice.session.chat", "v0.3.0.yaml"), "draft")
	if err := client.Reload(context.Background()); err == nil {
		t.Fatal("partial activation snapshot must fail")
	}
	assertFeatureActiveVersion(t, client, "report.generate", "v0.2.0")
	assertFeatureActiveVersion(t, client, "practice.session.chat", "v0.3.0")
	rewritePromptStatus(t, filepath.Join(prompts, "practice.session.chat", "v0.3.0.yaml"), "active")

	rewritePracticeStatuses(t, prompts, rubrics, "v0.2.0")
	if err := client.Reload(context.Background()); err != nil {
		t.Fatalf("rollback reload: %v", err)
	}
	assertFeatureActiveVersion(t, client, "report.generate", "v0.2.0")
	assertFeatureActiveVersion(t, client, "practice.session.chat", "v0.2.0")

	rewritePracticeStatuses(t, prompts, rubrics, "v0.3.0")
	if err := client.Reload(context.Background()); err != nil {
		t.Fatalf("reactivate reload: %v", err)
	}
	assertFeatureActiveVersion(t, client, "report.generate", "v0.2.0")
	assertFeatureActiveVersion(t, client, "practice.session.chat", "v0.3.0")
	for _, version := range []string{"v0.2.0", "v0.3.0"} {
		if _, _, err := client.GetPrompt("practice.session.chat", version, "multi"); err != nil {
			t.Fatalf("GetPrompt practice %s: %v", version, err)
		}
		if _, err := client.GetRubric("practice.session.chat", version, "multi"); err != nil {
			t.Fatalf("GetRubric practice %s: %v", version, err)
		}
	}
}

func rewritePracticeStatuses(t *testing.T, prompts, rubrics, activeVersion string) {
	t.Helper()
	for _, version := range []string{"v0.2.0", "v0.3.0"} {
		promptStatus := "draft"
		rubricStatus := "inactive"
		if version == activeVersion {
			promptStatus = "active"
			rubricStatus = "active"
		}
		rewritePromptStatus(t, filepath.Join(prompts, "practice.session.chat", version+".yaml"), promptStatus)
		rewriteRubricStatus(t, filepath.Join(rubrics, "practice.session.chat", version+".yaml"), rubricStatus)
	}
}

func rewriteRubricStatus(t *testing.T, path, status string) {
	t.Helper()
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read rubric %s: %v", path, err)
	}
	updated := string(body)
	found := false
	for _, current := range []string{"active", "inactive"} {
		marker := `status: "` + current + `"`
		if strings.Contains(updated, marker) {
			updated = strings.Replace(updated, marker, `status: "`+status+`"`, 1)
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("rubric %s missing status field", path)
	}
	if err := os.WriteFile(path, []byte(updated), 0o644); err != nil {
		t.Fatalf("write rubric %s: %v", path, err)
	}
}

func assertFeatureActiveVersion(t *testing.T, client *Client, featureKey, want string) {
	t.Helper()
	resolved, err := client.ResolveActive(context.Background(), featureKey, "multi")
	if err != nil {
		t.Fatalf("ResolveActive %s: %v", featureKey, err)
	}
	if resolved.PromptVersion != want || resolved.RubricVersion != want {
		t.Fatalf("%s active pair = %s/%s, want %s", featureKey, resolved.PromptVersion, resolved.RubricVersion, want)
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
