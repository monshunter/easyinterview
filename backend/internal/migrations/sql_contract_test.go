package migrations

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestBaselineMigrationEnablesVectorAndKeepsDownSafe(t *testing.T) {
	root := repoRoot(t)
	up := readFile(t, filepath.Join(root, "migrations", "000001_create_baseline.up.sql"))
	down := readFile(t, filepath.Join(root, "migrations", "000001_create_baseline.down.sql"))

	if !strings.Contains(strings.ToLower(up), "create extension if not exists vector") {
		t.Fatalf("baseline up migration must enable vector extension idempotently")
	}
	if strings.Contains(strings.ToLower(down), "drop extension") {
		t.Fatalf("baseline down migration must not drop vector by default")
	}
}

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", ".."))
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}
