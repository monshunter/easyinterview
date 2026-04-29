package profile_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient/profile"
)

// TestLoaderConcurrentReadAndReload runs concurrent readers against repeated
// reloads. Combined with `go test -race`, it covers the spec §5 risk
// "loader 替换内部 map 时 reader 拿到 partial 状态" — at least 100 reload
// rounds with parallel readers must produce no race and no nil pointer.
func TestLoaderConcurrentReadAndReload(t *testing.T) {
	dir := t.TempDir()
	const profileName = "practice.followup.default"
	writeYAML := func(version string) {
		body := fmt.Sprintf(`name: %s
task_type: chat
default:
  provider: stub
  model: stub-chat-1
timeout_ms: 1000
version: %s
`, profileName, version)
		if err := os.WriteFile(filepath.Join(dir, "p.yaml"), []byte(body), 0o600); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}
	}
	writeYAML("1.0.0")

	loader, err := profile.NewLoader(profile.Options{Dir: dir, PollInterval: -1})
	if err != nil {
		t.Fatalf("NewLoader: %v", err)
	}
	defer loader.Close()

	const rounds = 100
	const readers = 8

	var wg sync.WaitGroup
	stop := make(chan struct{})

	for i := 0; i < readers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stop:
					return
				default:
				}
				p, err := loader.Resolve(profileName)
				if err != nil {
					t.Errorf("Resolve: %v", err)
					return
				}
				if p == nil {
					t.Errorf("Resolve returned nil profile")
					return
				}
				if p.Default.Provider != "stub" {
					t.Errorf("provider mutated mid-snapshot: %q", p.Default.Provider)
					return
				}
			}
		}()
	}

	for round := 0; round < rounds; round++ {
		writeYAML(fmt.Sprintf("1.0.%d", round))
		if err := loader.Reload(context.Background()); err != nil {
			close(stop)
			wg.Wait()
			t.Fatalf("Reload round %d: %v", round, err)
		}
	}

	close(stop)
	wg.Wait()
}
