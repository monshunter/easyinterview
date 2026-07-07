package review

import (
	"os"
	"testing"
)

// TestNoNonCurrentRunnerFiles is the structural negative gate for spec D-12: the
// non-current review.Runner / review.Reaper / lease-backoff files were deleted when
// report_generate moved to the runner kernel (GenerateHandler). They must not
// reappear in this package.
func TestNoNonCurrentRunnerFiles(t *testing.T) {
	for _, name := range []string{"runner.go", "reaper.go", "lease.go"} {
		if _, err := os.Stat(name); err == nil {
			t.Fatalf("non-current review file %s still exists; report generation is owned by runner.Runtime + review.GenerateHandler", name)
		} else if !os.IsNotExist(err) {
			t.Fatalf("stat %s: %v", name, err)
		}
	}
}
