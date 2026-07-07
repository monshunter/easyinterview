package auth_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestNoBackgroundDispatcher is the structural negative gate for spec D-10 /
// D-12: the non-current in-process BackgroundMailDispatcher must not reappear in the
// auth package source (tests excluded). Email delivery now flows through
// async_jobs(email_dispatch) via EmailDispatchEnqueuer + EmailDispatchHandler.
func TestNoBackgroundDispatcher(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read auth package dir: %v", err)
	}
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		data, err := os.ReadFile(filepath.Clean(name))
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		if strings.Contains(string(data), "BackgroundMailDispatcher") {
			t.Fatalf("%s still references BackgroundMailDispatcher; non-current in-process dispatcher must stay absent", name)
		}
	}
}
