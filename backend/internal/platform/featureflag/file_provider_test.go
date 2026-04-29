package featureflag_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/platform/featureflag"
)

func writeFlagsYAML(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func TestFileProviderInitialLoadAndIsEnabled(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "feature-flags.yaml")
	writeFlagsYAML(t, path, `
flags:
  practice_hint_enabled:
    enabled: true
    public: true
  ai_fallback_model_enabled:
    enabled: true
    public: false
`)
	provider, err := featureflag.NewFileProvider(featureflag.FileProviderOptions{Path: path, ReloadInterval: time.Second})
	if err != nil {
		t.Fatalf("NewFileProvider: %v", err)
	}
	defer provider.Close()

	ctx := featureflag.FlagContext{AnonymousDistinctID: "anon", AppEnv: "dev"}
	if !provider.IsEnabled("practice_hint_enabled", ctx) {
		t.Errorf("practice_hint_enabled should be true")
	}
	if !provider.IsEnabled("ai_fallback_model_enabled", ctx) {
		t.Errorf("ai_fallback_model_enabled should be true")
	}
	snap := provider.Snapshot()
	if !snap["practice_hint_enabled"].Public {
		t.Errorf("practice_hint_enabled should be public")
	}
	if snap["ai_fallback_model_enabled"].Public {
		t.Errorf("ai_fallback_model_enabled must remain operator-only")
	}
}

func TestFileProviderHotReloadOnContentChange(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "feature-flags.yaml")
	writeFlagsYAML(t, path, "flags:\n  practice_hint_enabled:\n    enabled: false\n    public: true\n")
	provider, err := featureflag.NewFileProvider(featureflag.FileProviderOptions{Path: path, ReloadInterval: 50 * time.Millisecond})
	if err != nil {
		t.Fatalf("NewFileProvider: %v", err)
	}
	defer provider.Close()
	ctx := featureflag.FlagContext{AppEnv: "dev"}
	if provider.IsEnabled("practice_hint_enabled", ctx) {
		t.Errorf("initial enabled should be false")
	}

	writeFlagsYAML(t, path, "flags:\n  practice_hint_enabled:\n    enabled: true\n    public: true\n")
	// Bump mtime explicitly to ensure detection regardless of FS resolution.
	now := time.Now().Add(time.Second)
	if err := os.Chtimes(path, now, now); err != nil {
		t.Fatalf("chtimes: %v", err)
	}

	deadline, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	for {
		select {
		case <-deadline.Done():
			t.Fatal("hot reload did not pick up change in 2s")
		default:
		}
		if provider.IsEnabled("practice_hint_enabled", ctx) {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
}

func TestFileProviderInvalidYAMLKeepsLastSnapshot(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "feature-flags.yaml")
	writeFlagsYAML(t, path, "flags:\n  practice_hint_enabled:\n    enabled: true\n    public: true\n")
	provider, err := featureflag.NewFileProvider(featureflag.FileProviderOptions{Path: path, ReloadInterval: 50 * time.Millisecond})
	if err != nil {
		t.Fatalf("NewFileProvider: %v", err)
	}
	defer provider.Close()
	if !provider.IsEnabled("practice_hint_enabled", featureflag.FlagContext{}) {
		t.Errorf("initial enabled should be true")
	}

	writeFlagsYAML(t, path, ":::not yaml:::")
	now := time.Now().Add(time.Second)
	_ = os.Chtimes(path, now, now)
	time.Sleep(200 * time.Millisecond)

	if !provider.IsEnabled("practice_hint_enabled", featureflag.FlagContext{}) {
		t.Errorf("invalid YAML must not wipe last-known-good snapshot")
	}
}

func TestFileProviderInitialLoadFailsOnMissingFile(t *testing.T) {
	_, err := featureflag.NewFileProvider(featureflag.FileProviderOptions{Path: "/does/not/exist.yaml"})
	if err == nil {
		t.Fatal("expected error when initial file is missing")
	}
}
