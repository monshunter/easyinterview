package migrations

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateMigrationFilesRejectsMalformedNames(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "1_bad.up.sql")
	writeTestFile(t, dir, "000001_bad.down.sql")

	err := ValidateMigrationFiles(dir)
	if err == nil {
		t.Fatal("expected malformed file name to fail")
	}
	if !strings.Contains(err.Error(), "1_bad.up.sql") {
		t.Fatalf("expected error to include malformed file, got %v", err)
	}
}

func TestValidateMigrationFilesRejectsMissingPair(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "000001_create_baseline.up.sql")

	err := ValidateMigrationFiles(dir)
	if err == nil {
		t.Fatal("expected missing down pair to fail")
	}
	if !strings.Contains(err.Error(), "missing down") {
		t.Fatalf("expected missing pair error, got %v", err)
	}
}

func TestValidateMigrationFilesRejectsGaps(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "000001_create_baseline.up.sql")
	writeTestFile(t, dir, "000001_create_baseline.down.sql")
	writeTestFile(t, dir, "000003_add_table.up.sql")
	writeTestFile(t, dir, "000003_add_table.down.sql")

	err := ValidateMigrationFiles(dir)
	if err == nil {
		t.Fatal("expected version gap to fail")
	}
	if !strings.Contains(err.Error(), "expected version 000002") {
		t.Fatalf("expected version gap error, got %v", err)
	}
}

func TestValidateMigrationFilesAcceptsSequentialPairs(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "000001_create_baseline.up.sql")
	writeTestFile(t, dir, "000001_create_baseline.down.sql")
	writeTestFile(t, dir, "000002_add_example.up.sql")
	writeTestFile(t, dir, "000002_add_example.down.sql")

	if err := ValidateMigrationFiles(dir); err != nil {
		t.Fatalf("expected sequential pairs to pass, got %v", err)
	}
}

func writeTestFile(t *testing.T, dir, name string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte("-- test\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}
