package registry

import (
	"encoding/json"
	"fmt"
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

	// Each baseline ships the canonical multi prompt.
	for fk, langs := range snap.prompts {
		if _, ok := langs["multi"]; !ok {
			t.Errorf("feature_key %s missing multi prompt", fk)
		}
	}
	for fk, langs := range snap.rubrics {
		if _, ok := langs["multi"]; !ok {
			t.Errorf("feature_key %s missing multi rubric", fk)
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

func TestLoadOutputSchemaLanguageIndependent(t *testing.T) {
	t.Parallel()
	promptsRoot, rubricsRoot := repoConfigRoots(t)

	snap, err := loadFromDisk(promptsRoot, rubricsRoot)
	if err != nil {
		t.Fatalf("loadFromDisk failed: %v", err)
	}

	schema := snap.prompts["target.import.parse"]["multi"].outputSchema
	if schema == nil {
		t.Fatal("output schema must be loaded for canonical multi prompt")
	}
	if got := schemaType(t, *schema); got != "object" {
		t.Fatalf("target.import.parse schema type: want object, got %s", got)
	}

	recommendation := snap.prompts["jd_match.recommendation"]["multi"].outputSchema
	if recommendation == nil {
		t.Fatal("jd_match.recommendation output schema missing")
	}
	if got := schemaType(t, *recommendation); got != "array" {
		t.Fatalf("jd_match.recommendation schema type: want array, got %s", got)
	}
}

func TestLoadMissingCanonicalMultiRejected(t *testing.T) {
	t.Parallel()
	promptsRoot, rubricsRoot := tempBaselineCopy(t)

	writePromptLanguageOverride(t, promptsRoot, "target.import.parse", "en")
	writeRubricLanguageOverride(t, rubricsRoot, "target.import.parse", "en")
	for _, path := range []string{
		filepath.Join(promptsRoot, "target.import.parse", "v0.1.0.yaml"),
		filepath.Join(promptsRoot, "target.import.parse", "v0.1.0.md"),
		filepath.Join(rubricsRoot, "target.import.parse", "v0.1.0.yaml"),
	} {
		if err := os.Remove(path); err != nil {
			t.Fatalf("remove %s: %v", path, err)
		}
	}

	if _, err := loadFromDisk(promptsRoot, rubricsRoot); err == nil {
		t.Fatalf("expected missing canonical multi error, got nil")
	} else if !strings.Contains(err.Error(), "missing canonical multi") {
		t.Fatalf("expected missing canonical multi message, got: %v", err)
	}
}

func TestLoadOrphanLanguageOverrideRejected(t *testing.T) {
	t.Parallel()
	promptsRoot, rubricsRoot := tempBaselineCopy(t)

	writePromptLanguageOverride(t, promptsRoot, "target.import.parse", "en")
	writeRubricLanguageOverride(t, rubricsRoot, "target.import.parse", "en")
	srcRubric := filepath.Join(rubricsRoot, "target.import.parse", "v0.1.0.en.yaml")
	if err := os.Remove(srcRubric); err != nil && !os.IsNotExist(err) {
		t.Fatalf("remove override rubric: %v", err)
	}

	if _, err := loadFromDisk(promptsRoot, rubricsRoot); err == nil {
		t.Fatalf("expected orphan language override error, got nil")
	} else if !strings.Contains(err.Error(), "no matching rubric") {
		t.Fatalf("expected matching-rubric message, got: %v", err)
	}
}

func writePromptLanguageOverride(t *testing.T, promptsRoot, featureKey, language string) {
	t.Helper()

	featureDir := filepath.Join(promptsRoot, featureKey)
	body, err := os.ReadFile(filepath.Join(featureDir, "v0.1.0.md"))
	if err != nil {
		t.Fatalf("read multi prompt body: %v", err)
	}
	meta := PromptMeta{
		FeatureKey: featureKey,
		Version:    "v0.1.0",
		Language:   language,
		Status:     "active",
		CreatedAt:  "2026-05-09T12:00:00Z",
	}
	hash, err := computeTemplateHash(body, meta)
	if err != nil {
		t.Fatalf("compute override hash: %v", err)
	}
	if err := os.WriteFile(filepath.Join(featureDir, "v0.1.0."+language+".md"), body, 0o644); err != nil {
		t.Fatalf("write override prompt body: %v", err)
	}
	yamlBody := fmt.Sprintf(
		"feature_key: %q\nversion: %q\nlanguage: %q\ntemplate_hash: %q\nstatus: %q\ncreated_at: %q\n",
		meta.FeatureKey,
		meta.Version,
		meta.Language,
		hash,
		meta.Status,
		meta.CreatedAt,
	)
	if err := os.WriteFile(filepath.Join(featureDir, "v0.1.0."+language+".yaml"), []byte(yamlBody), 0o644); err != nil {
		t.Fatalf("write override prompt yaml: %v", err)
	}
}

func writeRubricLanguageOverride(t *testing.T, rubricsRoot, featureKey, language string) {
	t.Helper()

	featureDir := filepath.Join(rubricsRoot, featureKey)
	body, err := os.ReadFile(filepath.Join(featureDir, "v0.1.0.yaml"))
	if err != nil {
		t.Fatalf("read multi rubric yaml: %v", err)
	}
	replaced := strings.Replace(string(body), `language: "multi"`, fmt.Sprintf("language: %q", language), 1)
	if replaced == string(body) {
		t.Fatalf("multi rubric language field not found for %s", featureKey)
	}
	if err := os.WriteFile(filepath.Join(featureDir, "v0.1.0."+language+".yaml"), []byte(replaced), 0o644); err != nil {
		t.Fatalf("write override rubric yaml: %v", err)
	}
}

func TestLoadMissingOutputSchemaRejected(t *testing.T) {
	t.Parallel()
	promptsRoot, rubricsRoot := tempBaselineCopy(t)

	target := filepath.Join(promptsRoot, "target.import.parse", "v0.1.0.schema.json")
	if err := os.Remove(target); err != nil {
		t.Fatalf("remove schema: %v", err)
	}

	if _, err := loadFromDisk(promptsRoot, rubricsRoot); err == nil {
		t.Fatalf("expected missing-schema error, got nil")
	} else if !strings.Contains(err.Error(), "missing output schema") {
		t.Fatalf("expected missing-schema message, got: %v", err)
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

func schemaType(t *testing.T, raw json.RawMessage) string {
	t.Helper()
	var schema struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(raw, &schema); err != nil {
		t.Fatalf("parse schema: %v", err)
	}
	return schema.Type
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
