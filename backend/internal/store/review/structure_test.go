package review

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

func TestReviewStoreDoesNotOwnAsyncJobLeaseFinalizeOrReaper(t *testing.T) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve review store test path")
	}
	dir := filepath.Dir(currentFile)
	for _, forbiddenFile := range []string{"lease_async_job.go", "reaper.go"} {
		if _, err := os.Stat(filepath.Join(dir, forbiddenFile)); err == nil {
			t.Fatalf("review store must not own runner kernel file %s", forbiddenFile)
		} else if !os.IsNotExist(err) {
			t.Fatalf("stat %s: %v", forbiddenFile, err)
		}
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read review store directory: %v", err)
	}
	forbiddenIdentifiers := []string{
		"func (r *Repository) LeaseAsyncJob",
		"func (r *Repository) UpdateAsyncJobSucceeded",
		"func (r *Repository) UpdateAsyncJobFailed",
		"func (r *Repository) ReclaimExpiredLeases",
	}
	forbiddenSQL := []*regexp.Regexp{
		regexp.MustCompile(`(?is)update\s+async_jobs\s+set\s+status\s*=\s*'running'.*attempts\s*=\s*attempts\s*\+\s*1`),
		regexp.MustCompile(`(?is)update\s+async_jobs\s+set\s+status\s*=\s*'queued'.*locked_at\s*=\s*null.*where\s+status\s*=\s*'running'.*locked_at`),
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			t.Fatalf("read %s: %v", entry.Name(), err)
		}
		body := string(raw)
		for _, identifier := range forbiddenIdentifiers {
			if strings.Contains(body, identifier) {
				t.Fatalf("review store production file %s reintroduced runner owner %q", entry.Name(), identifier)
			}
		}
		for _, pattern := range forbiddenSQL {
			if pattern.MatchString(body) {
				t.Fatalf("review store production file %s reintroduced runner lease/reaper SQL %q", entry.Name(), pattern.String())
			}
		}
	}
}
