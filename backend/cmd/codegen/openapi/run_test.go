package main

import (
	"os"
	"path/filepath"
	"testing"
)

// TestRun_Idempotent verifies that two consecutive `Run` invocations produce
// byte-identical generated files. This is what `make codegen-check` relies on
// to detect drift via `git diff --exit-code`.
func TestRun_Idempotent(t *testing.T) {
	repoRoot := mustFindRepoRoot(t)
	tmp := t.TempDir()

	// Mirror the relevant inputs into a writable temp tree so the test does
	// not modify the repo's openapi.yaml on disk.
	openapiSrc := filepath.Join(repoRoot, "openapi", "openapi.yaml")
	openapiDst := filepath.Join(tmp, "openapi", "openapi.yaml")
	if err := os.MkdirAll(filepath.Dir(openapiDst), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	mustCopy(t, openapiSrc, openapiDst)

	conventionsPath := filepath.Join(repoRoot, "shared", "conventions.yaml")
	templatesDir := filepath.Join(repoRoot, "openapi", "templates")
	mirrorTemplates := filepath.Join(tmp, "openapi", "templates")
	if err := mirrorDir(filepath.Join(repoRoot, "openapi", "templates"), mirrorTemplates); err != nil {
		t.Fatalf("mirror templates: %v", err)
	}
	_ = templatesDir // unused; mirror serves the test

	// First run.
	if err := Run(openapiDst, conventionsPath, mirrorTemplates, tmp, false); err != nil {
		t.Fatalf("first Run: %v", err)
	}
	hashes1 := snapshotHashes(t, tmp)

	// Second run.
	if err := Run(openapiDst, conventionsPath, mirrorTemplates, tmp, false); err != nil {
		t.Fatalf("second Run: %v", err)
	}
	hashes2 := snapshotHashes(t, tmp)

	for path, h1 := range hashes1 {
		if h2 := hashes2[path]; h1 != h2 {
			t.Errorf("non-idempotent output for %s: %s vs %s", path, h1, h2)
		}
	}
}

// TestRun_DriftPropagatesFromConventions verifies that adding a new enum
// value to the loaded Conventions changes the openapi.yaml B1-AUTO block.
// We exercise the rewrite path (not the YAML on disk) by feeding a synthetic
// conventions struct into `syncB1AutoBlock` directly.
func TestRun_DriftPropagatesFromConventions(t *testing.T) {
	repoRoot := mustFindRepoRoot(t)
	tmp := t.TempDir()
	dst := filepath.Join(tmp, "openapi.yaml")
	mustCopy(t, filepath.Join(repoRoot, "openapi", "openapi.yaml"), dst)

	original, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("read: %v", err)
	}

	// Re-run the canonical sync; output should be byte-stable.
	conv, err := loadConventions(filepath.Join(repoRoot, "shared", "conventions.yaml"))
	if err != nil {
		t.Fatalf("loadConventions: %v", err)
	}
	if err := syncB1AutoBlock(dst, conv); err != nil {
		t.Fatalf("sync 1: %v", err)
	}
	stable, _ := os.ReadFile(dst)
	if string(original) != string(stable) {
		t.Fatalf("first sync produced drift even though conventions.yaml is unchanged")
	}

	// Mutate the conventions struct (one of the B1 enums) and verify the
	// next sync rewrites openapi.yaml.
	for i := range conv.Enums {
		if conv.Enums[i].Name == "MistakeStatus" {
			conv.Enums[i].Values = append(conv.Enums[i].Values, "x_test_drift_value")
			break
		}
	}
	if err := syncB1AutoBlock(dst, conv); err != nil {
		t.Fatalf("sync 2: %v", err)
	}
	drifted, _ := os.ReadFile(dst)
	if string(stable) == string(drifted) {
		t.Fatal("expected drift after mutating MistakeStatus values")
	}
}

func mustFindRepoRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	cur := wd
	for cur != "/" {
		if _, err := os.Stat(filepath.Join(cur, "go.work")); err == nil {
			return cur
		}
		if _, err := os.Stat(filepath.Join(cur, "openapi", "openapi.yaml")); err == nil {
			return cur
		}
		cur = filepath.Dir(cur)
	}
	t.Fatalf("could not locate repo root from %s", wd)
	return ""
}

func mustCopy(t *testing.T, src, dst string) {
	t.Helper()
	data, err := os.ReadFile(src)
	if err != nil {
		t.Fatalf("read %s: %v", src, err)
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(dst, data, 0o644); err != nil {
		t.Fatalf("write %s: %v", dst, err)
	}
}

func mirrorDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, path)
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, info.Mode())
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, info.Mode())
	})
}

func snapshotHashes(t *testing.T, root string) map[string]string {
	t.Helper()
	out := map[string]string{}
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		out[path] = sha256hex(data)
		return nil
	})
	if err != nil {
		t.Fatalf("walk: %v", err)
	}
	return out
}
