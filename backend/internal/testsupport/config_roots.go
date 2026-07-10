package testsupport

import (
	"os"
	"path/filepath"
	"testing"
)

// ConfigRoots locates the repository prompt and rubric truth sources for tests and benchmarks.
func ConfigRoots(tb testing.TB) (string, string) {
	tb.Helper()
	wd, err := os.Getwd()
	if err != nil {
		tb.Fatalf("getwd: %v", err)
		return "", ""
	}

	dir := wd
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			repoRoot := filepath.Dir(dir)
			return filepath.Join(repoRoot, "config", "prompts"), filepath.Join(repoRoot, "config", "rubrics")
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			tb.Skipf("could not locate backend go.mod from %s", wd)
			return "", ""
		}
		dir = parent
	}
}
