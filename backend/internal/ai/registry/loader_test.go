package registry

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monshunter/easyinterview/backend/internal/testsupport"
	"gopkg.in/yaml.v3"
)

// TestLoad covers the loader's happy path against the real
// config/prompts/ + config/rubrics/ shipped by plan items 1.2 + 1.3.
func TestLoadHappyPath(t *testing.T) {
	t.Parallel()
	promptsRoot, rubricsRoot := testsupport.ConfigRoots(t)

	snap, err := loadFromDisk(promptsRoot, rubricsRoot)
	if err != nil {
		t.Fatalf("loadFromDisk failed: %v", err)
	}

	// 6 baseline feature_keys remain after conversation simplification removed
	// the jd_match and debrief/profile prompt owners; both directories must
	// the structured question prompt owners.
	if got := len(snap.prompts); got != 6 {
		t.Fatalf("prompts: want 6 feature_keys, got %d", got)
	}
	if got := len(snap.rubrics); got != 6 {
		t.Fatalf("rubrics: want 6 feature_keys, got %d", got)
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

func TestLoadPreservesReportPromptAndRubricVersions(t *testing.T) {
	t.Parallel()
	promptsRoot, rubricsRoot := testsupport.ConfigRoots(t)

	snap, err := loadFromDisk(promptsRoot, rubricsRoot)
	if err != nil {
		t.Fatalf("loadFromDisk failed: %v", err)
	}

	promptVersions := snap.prompts["report.generate"]["multi"]
	if _, ok := promptVersions["v0.1.0"]; !ok {
		t.Fatal("report.generate prompt rollback version v0.1.0 must remain loaded")
	}
	if _, ok := promptVersions["v0.2.0"]; !ok {
		t.Fatal("report.generate prompt candidate version v0.2.0 must be loaded")
	}
	rubricVersions := snap.rubrics["report.generate"]["multi"]
	if _, ok := rubricVersions["v0.1.0"]; !ok {
		t.Fatal("report.generate rubric rollback version v0.1.0 must remain loaded")
	}
	if _, ok := rubricVersions["v0.2.0"]; !ok {
		t.Fatal("report.generate rubric candidate version v0.2.0 must be loaded")
	}
}

func TestLoadRejectsUnknownRubricStatus(t *testing.T) {
	t.Parallel()
	promptsRoot, rubricsRoot := tempBaselineCopy(t)
	target := filepath.Join(rubricsRoot, "report.generate", "v0.2.0.yaml")
	body, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read rubric: %v", err)
	}
	body = []byte(strings.Replace(string(body), `status: "active"`, `status: "retired"`, 1))
	if err := os.WriteFile(target, body, 0o644); err != nil {
		t.Fatalf("write rubric: %v", err)
	}

	if _, err := loadFromDisk(promptsRoot, rubricsRoot); err == nil {
		t.Fatal("expected unknown rubric status to fail")
	} else if !strings.Contains(err.Error(), "status") || !strings.Contains(err.Error(), "invalid") {
		t.Fatalf("expected invalid-status diagnostic, got %v", err)
	}
}

func TestLoadRejectsZeroActivePromptVersion(t *testing.T) {
	t.Parallel()
	promptsRoot, rubricsRoot := tempBaselineCopy(t)
	target := filepath.Join(promptsRoot, "target.import.parse", "v0.1.0.yaml")
	rewritePromptStatus(t, target, "draft")

	if _, err := loadFromDisk(promptsRoot, rubricsRoot); err == nil {
		t.Fatal("expected zero active prompt versions to fail")
	} else if !strings.Contains(err.Error(), "exactly one active prompt") {
		t.Fatalf("expected active prompt diagnostic, got %v", err)
	}
}

func TestLoadRejectsPromptRubricVersionParityDrift(t *testing.T) {
	t.Parallel()
	promptsRoot, rubricsRoot := tempBaselineCopy(t)
	if err := os.Remove(filepath.Join(rubricsRoot, "report.generate", "v0.2.0.yaml")); err != nil {
		t.Fatalf("remove v0.2 rubric: %v", err)
	}

	if _, err := loadFromDisk(promptsRoot, rubricsRoot); err == nil {
		t.Fatal("expected prompt/rubric version parity drift to fail")
	} else if !strings.Contains(err.Error(), "version parity") {
		t.Fatalf("expected version parity diagnostic, got %v", err)
	}
}

func TestLoadRejectsOpenGroundedReportSchema(t *testing.T) {
	t.Parallel()
	promptsRoot, rubricsRoot := tempBaselineCopy(t)
	target := filepath.Join(promptsRoot, "report.generate", "v0.2.0.schema.json")
	body, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read schema: %v", err)
	}
	var schema map[string]any
	if err := json.Unmarshal(body, &schema); err != nil {
		t.Fatalf("parse schema: %v", err)
	}
	delete(schema, "additionalProperties")
	body, err = json.Marshal(schema)
	if err != nil {
		t.Fatalf("marshal schema: %v", err)
	}
	if err := os.WriteFile(target, body, 0o644); err != nil {
		t.Fatalf("write schema: %v", err)
	}

	if _, err := loadFromDisk(promptsRoot, rubricsRoot); err == nil {
		t.Fatal("expected open grounded report schema to fail")
	} else if !strings.Contains(err.Error(), "additionalProperties=false") {
		t.Fatalf("expected closed-schema diagnostic, got %v", err)
	}
}

func TestLoadRejectsMultipleActivePromptVersions(t *testing.T) {
	t.Parallel()
	promptsRoot, rubricsRoot := tempBaselineCopy(t)
	rewritePromptStatus(t, filepath.Join(promptsRoot, "report.generate", "v0.1.0.yaml"), "active")

	if _, err := loadFromDisk(promptsRoot, rubricsRoot); err == nil {
		t.Fatal("expected multiple active prompt versions to fail")
	} else if !strings.Contains(err.Error(), "exactly one active prompt") {
		t.Fatalf("expected active prompt diagnostic, got %v", err)
	}
}

func TestLoadRejectsZeroActiveRubricVersion(t *testing.T) {
	t.Parallel()
	promptsRoot, rubricsRoot := tempBaselineCopy(t)
	target := filepath.Join(rubricsRoot, "report.generate", "v0.2.0.yaml")
	body, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read rubric: %v", err)
	}
	body = []byte(strings.Replace(string(body), `status: "active"`, `status: "inactive"`, 1))
	if err := os.WriteFile(target, body, 0o644); err != nil {
		t.Fatalf("write rubric: %v", err)
	}

	if _, err := loadFromDisk(promptsRoot, rubricsRoot); err == nil {
		t.Fatal("expected zero active rubric versions to fail")
	} else if !strings.Contains(err.Error(), "exactly one active rubric") {
		t.Fatalf("expected active rubric diagnostic, got %v", err)
	}
}

func TestLoadRejectsDuplicatePromptVersion(t *testing.T) {
	t.Parallel()
	promptsRoot, rubricsRoot := tempBaselineCopy(t)
	featureDir := filepath.Join(promptsRoot, "target.import.parse")
	for _, ext := range []string{".yaml", ".md"} {
		body, err := os.ReadFile(filepath.Join(featureDir, "v0.1.0"+ext))
		if err != nil {
			t.Fatalf("read duplicate source: %v", err)
		}
		if err := os.WriteFile(filepath.Join(featureDir, "v9.9.9"+ext), body, 0o644); err != nil {
			t.Fatalf("write duplicate: %v", err)
		}
	}

	if _, err := loadFromDisk(promptsRoot, rubricsRoot); err == nil {
		t.Fatal("expected duplicate prompt version to fail")
	} else if !strings.Contains(err.Error(), "duplicate prompt") {
		t.Fatalf("expected duplicate prompt diagnostic, got %v", err)
	}
}

func rewritePromptStatus(t *testing.T, target string, status string) {
	t.Helper()
	body, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read prompt: %v", err)
	}
	var raw promptYAML
	if err := yaml.Unmarshal(body, &raw); err != nil {
		t.Fatalf("parse prompt meta: %v", err)
	}
	meta := PromptMeta{
		FeatureKey: raw.FeatureKey,
		Version:    raw.Version,
		Language:   raw.Language,
		Status:     status,
		CreatedAt:  raw.CreatedAt,
	}
	promptBody, err := os.ReadFile(strings.TrimSuffix(target, ".yaml") + ".md")
	if err != nil {
		t.Fatalf("read prompt body: %v", err)
	}
	hash, err := computeTemplateHash(promptBody, meta)
	if err != nil {
		t.Fatalf("compute prompt hash: %v", err)
	}
	body = []byte(strings.Replace(string(body), `status: "`+raw.Status+`"`, `status: "`+status+`"`, 1))
	body = []byte(strings.Replace(string(body), raw.TemplateHash, hash, 1))
	if err := os.WriteFile(target, body, 0o644); err != nil {
		t.Fatalf("write prompt: %v", err)
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
	promptsRoot, rubricsRoot := testsupport.ConfigRoots(t)

	snap, err := loadFromDisk(promptsRoot, rubricsRoot)
	if err != nil {
		t.Fatalf("loadFromDisk failed: %v", err)
	}

	schema := snap.prompts["target.import.parse"]["multi"]["v0.1.0"].outputSchema
	if schema == nil {
		t.Fatal("output schema must be loaded for canonical multi prompt")
	}
	if got := schemaType(t, *schema); got != "object" {
		t.Fatalf("target.import.parse schema type: want object, got %s", got)
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

// tempBaselineCopy clones the in-repo baseline into t.TempDir so individual
// tests can mutate files without polluting the repo working tree.
func tempBaselineCopy(t *testing.T) (string, string) {
	t.Helper()
	srcPrompts, srcRubrics := testsupport.ConfigRoots(t)
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
