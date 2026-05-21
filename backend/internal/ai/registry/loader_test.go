package registry

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestLoad covers the loader's happy path against the real
// config/prompts/ + config/rubrics/ shipped by plan items 1.2 + 1.3.
func TestLoadHappyPath(t *testing.T) {
	t.Parallel()
	promptsRoot, rubricsRoot := repoConfigRoots(t)

	snap, err := loadFromDisk(promptsRoot, rubricsRoot)
	if err != nil {
		t.Fatalf("loadFromDisk failed: %v", err)
	}

	// Spec §3.1.1 names 13 baseline feature_keys (11 original + jd_match.recommendation
	// + jd_match.search cross-owner additive); both directories must
	// resolve all 13 through the loader.
	if got := len(snap.prompts); got != 13 {
		t.Fatalf("prompts: want 13 feature_keys, got %d", got)
	}
	if got := len(snap.rubrics); got != 13 {
		t.Fatalf("rubrics: want 13 feature_keys, got %d", got)
	}

	// Each baseline ships at least the multi + en pair.
	for fk, langs := range snap.prompts {
		if _, ok := langs["multi"]; !ok {
			t.Errorf("feature_key %s missing multi prompt", fk)
		}
		if len(langs) < 2 {
			t.Errorf("feature_key %s expected >=2 prompt languages, got %d", fk, len(langs))
		}
	}
}

func TestLoadHashDriftRejected(t *testing.T) {
	t.Parallel()
	promptsRoot, rubricsRoot := tempBaselineCopy(t)

	// Mutate a single prompt body without refreshing template_hash.
	target := filepath.Join(promptsRoot, "target.import.parse", "v0.1.0.md")
	body, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	if err := os.WriteFile(target, append(body, '\n'), 0o644); err != nil {
		t.Fatalf("write body: %v", err)
	}

	if _, err := loadFromDisk(promptsRoot, rubricsRoot); err == nil {
		t.Fatalf("expected hash drift error, got nil")
	} else if !strings.Contains(err.Error(), "template_hash drift") {
		t.Fatalf("expected drift message, got: %v", err)
	}
}

func TestLoadMissingMarkdownBody(t *testing.T) {
	t.Parallel()
	promptsRoot, rubricsRoot := tempBaselineCopy(t)

	target := filepath.Join(promptsRoot, "target.import.parse", "v0.1.0.md")
	if err := os.Remove(target); err != nil {
		t.Fatalf("remove body: %v", err)
	}

	if _, err := loadFromDisk(promptsRoot, rubricsRoot); err == nil {
		t.Fatalf("expected missing-body error, got nil")
	} else if !strings.Contains(err.Error(), "missing markdown body") {
		t.Fatalf("expected missing-body message, got: %v", err)
	}
}

func TestLoadInvalidYAML(t *testing.T) {
	t.Parallel()
	promptsRoot, rubricsRoot := tempBaselineCopy(t)

	target := filepath.Join(promptsRoot, "target.import.parse", "v0.1.0.yaml")
	if err := os.WriteFile(target, []byte("not: valid: yaml: content:\n  - "), 0o644); err != nil {
		t.Fatalf("write yaml: %v", err)
	}

	if _, err := loadFromDisk(promptsRoot, rubricsRoot); err == nil {
		t.Fatalf("expected yaml parse error, got nil")
	}
}

// repoConfigRoots returns the absolute paths to the in-repo
// config/prompts and config/rubrics roots, skipping the test if the
// repository layout cannot be located (for example when this binary runs
// outside a checked-out tree).
func repoConfigRoots(t *testing.T) (string, string) {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	// Walk upward until we find go.mod for the backend module.
	dir := wd
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Skipf("could not locate backend go.mod from %s", wd)
			return "", ""
		}
		dir = parent
	}
	repoRoot := filepath.Dir(dir) // backend/.. == repo root
	prompts := filepath.Join(repoRoot, "config", "prompts")
	rubrics := filepath.Join(repoRoot, "config", "rubrics")
	return prompts, rubrics
}

// tempBaselineCopy clones the in-repo baseline into t.TempDir so individual
// tests can mutate files without polluting the repo working tree.
func tempBaselineCopy(t *testing.T) (string, string) {
	t.Helper()
	srcPrompts, srcRubrics := repoConfigRoots(t)
	root := t.TempDir()
	dstPrompts := filepath.Join(root, "config", "prompts")
	dstRubrics := filepath.Join(root, "config", "rubrics")
	if err := copyTree(srcPrompts, dstPrompts); err != nil {
		t.Fatalf("copy prompts: %v", err)
	}
	if err := copyTree(srcRubrics, dstRubrics); err != nil {
		t.Fatalf("copy rubrics: %v", err)
	}
	return dstPrompts, dstRubrics
}

func copyTree(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		body, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		return os.WriteFile(target, body, 0o644)
	})
}
