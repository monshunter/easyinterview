package registry

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// snapshot is the in-memory shape produced by the loader. The cache
// (cache.go) wraps it in atomic.Value; the resolver (resolver.go) reads it.
type snapshot struct {
	prompts map[string]map[string]promptEntry // featureKey -> language -> entry
	rubrics map[string]map[string]rubricEntry // featureKey -> language -> entry
}

type promptEntry struct {
	meta         PromptMeta
	body         string
	yamlPath     string
	mdPath       string
	outputSchema *json.RawMessage
}

type rubricEntry struct {
	schema   RubricSchema
	yamlPath string
}

// loadFromDisk walks promptsRoot and rubricsRoot and builds a snapshot.
// It performs schema validation (status enum, required fields), template
// hash verification, and prompt/rubric language-set parity checks.
func loadFromDisk(promptsRoot, rubricsRoot string) (*snapshot, error) {
	snap := &snapshot{
		prompts: map[string]map[string]promptEntry{},
		rubrics: map[string]map[string]rubricEntry{},
	}

	if err := loadPrompts(promptsRoot, snap); err != nil {
		return nil, err
	}
	if err := loadRubrics(rubricsRoot, snap); err != nil {
		return nil, err
	}
	if err := validateLanguageParity(snap); err != nil {
		return nil, err
	}
	return snap, nil
}

func loadPrompts(root string, snap *snapshot) error {
	if root == "" {
		return fmt.Errorf("registry: prompts root is required")
	}
	outputSchemas := map[string]*json.RawMessage{}
	return filepath.WalkDir(root, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() || !strings.HasSuffix(path, ".yaml") {
			return nil
		}
		entry, err := readPrompt(path)
		if err != nil {
			return err
		}
		schema, err := readOutputSchema(path, entry.meta, outputSchemas)
		if err != nil {
			return err
		}
		entry.outputSchema = schema
		entry.meta.OutputSchema = schema
		bucket, ok := snap.prompts[entry.meta.FeatureKey]
		if !ok {
			bucket = map[string]promptEntry{}
			snap.prompts[entry.meta.FeatureKey] = bucket
		}
		if _, dup := bucket[entry.meta.Language]; dup {
			return fmt.Errorf("registry: duplicate prompt for %s/%s", entry.meta.FeatureKey, entry.meta.Language)
		}
		bucket[entry.meta.Language] = *entry
		return nil
	})
}

var outputSchemaExemptFeatureKeys = map[string]struct{}{
	"practice.voice.stt":     {},
	"practice.voice.tts":     {},
	"practice.dictation.stt": {},
}

func readOutputSchema(yamlPath string, meta PromptMeta, cache map[string]*json.RawMessage) (*json.RawMessage, error) {
	if _, exempt := outputSchemaExemptFeatureKeys[meta.FeatureKey]; exempt {
		return nil, nil
	}
	cacheKey := meta.FeatureKey + "\x00" + meta.Version
	if schema, ok := cache[cacheKey]; ok {
		return schema, nil
	}
	schemaPath := filepath.Join(filepath.Dir(yamlPath), meta.Version+".schema.json")
	body, err := os.ReadFile(schemaPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("registry: missing output schema %s for %s/%s", schemaPath, meta.FeatureKey, meta.Version)
		}
		return nil, fmt.Errorf("registry: read output schema %s: %w", schemaPath, err)
	}
	var parsed any
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, fmt.Errorf("registry: parse output schema %s: %w", schemaPath, err)
	}
	raw := json.RawMessage(append([]byte(nil), body...))
	cache[cacheKey] = &raw
	return &raw, nil
}

func loadRubrics(root string, snap *snapshot) error {
	if root == "" {
		return fmt.Errorf("registry: rubrics root is required")
	}
	return filepath.WalkDir(root, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() || !strings.HasSuffix(path, ".yaml") {
			return nil
		}
		entry, err := readRubric(path)
		if err != nil {
			return err
		}
		bucket, ok := snap.rubrics[entry.schema.FeatureKey]
		if !ok {
			bucket = map[string]rubricEntry{}
			snap.rubrics[entry.schema.FeatureKey] = bucket
		}
		if _, dup := bucket[entry.schema.Language]; dup {
			return fmt.Errorf("registry: duplicate rubric for %s/%s", entry.schema.FeatureKey, entry.schema.Language)
		}
		bucket[entry.schema.Language] = *entry
		return nil
	})
}

// promptYAML mirrors the on-disk schema. Field tags hold the YAML keys.
type promptYAML struct {
	FeatureKey   string `yaml:"feature_key"`
	Version      string `yaml:"version"`
	Language     string `yaml:"language"`
	TemplateHash string `yaml:"template_hash"`
	Status       string `yaml:"status"`
	CreatedAt    string `yaml:"created_at"`
}

func readPrompt(yamlPath string) (*promptEntry, error) {
	yamlBytes, err := os.ReadFile(yamlPath)
	if err != nil {
		return nil, fmt.Errorf("registry: read %s: %w", yamlPath, err)
	}
	var raw promptYAML
	if err := yaml.Unmarshal(yamlBytes, &raw); err != nil {
		return nil, fmt.Errorf("registry: parse %s: %w", yamlPath, err)
	}
	if raw.FeatureKey == "" || raw.Version == "" || raw.Language == "" || raw.Status == "" || raw.TemplateHash == "" {
		return nil, fmt.Errorf("registry: prompt %s missing required meta field", yamlPath)
	}
	switch raw.Status {
	case "draft", "active", "deprecated":
	default:
		return nil, fmt.Errorf("registry: prompt %s status %q invalid", yamlPath, raw.Status)
	}

	mdPath := strings.TrimSuffix(yamlPath, ".yaml") + ".md"
	bodyBytes, err := os.ReadFile(mdPath)
	if err != nil {
		return nil, fmt.Errorf("registry: missing markdown body %s: %w", mdPath, err)
	}

	meta := PromptMeta{
		FeatureKey:   raw.FeatureKey,
		Version:      raw.Version,
		Language:     raw.Language,
		TemplateHash: raw.TemplateHash,
		Status:       raw.Status,
		CreatedAt:    raw.CreatedAt,
	}
	computed, err := computeTemplateHash(bodyBytes, meta)
	if err != nil {
		return nil, fmt.Errorf("registry: %s template_hash: %w", yamlPath, err)
	}
	if computed != raw.TemplateHash {
		return nil, fmt.Errorf(
			"registry: %s template_hash drift (yaml=%s, computed=%s)",
			yamlPath, raw.TemplateHash, computed,
		)
	}

	return &promptEntry{
		meta:     meta,
		body:     string(bodyBytes),
		yamlPath: yamlPath,
		mdPath:   mdPath,
	}, nil
}

// computeTemplateHash applies the canonical algorithm from
// config/prompts/README.md §3 and prompt_lint.expected_hash. Both
// implementations must agree byte-for-byte.
func computeTemplateHash(body []byte, meta PromptMeta) (string, error) {
	metaForHash := map[string]string{
		"feature_key": meta.FeatureKey,
		"version":     meta.Version,
		"language":    meta.Language,
		"status":      meta.Status,
		"created_at":  meta.CreatedAt,
	}
	keys := make([]string, 0, len(metaForHash))
	for k := range metaForHash {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	// Build canonical JSON manually so output exactly matches Python's
	// json.dumps(sort_keys=True, ensure_ascii=False, separators=(",", ":")).
	var b strings.Builder
	b.WriteByte('{')
	for i, k := range keys {
		if i > 0 {
			b.WriteByte(',')
		}
		keyBytes, err := json.Marshal(k)
		if err != nil {
			return "", err
		}
		b.Write(keyBytes)
		b.WriteByte(':')
		valBytes, err := json.Marshal(metaForHash[k])
		if err != nil {
			return "", err
		}
		b.Write(valBytes)
	}
	b.WriteByte('}')
	b.WriteByte('\n')

	h := sha256.New()
	h.Write(body)
	h.Write([]byte(b.String()))
	return hex.EncodeToString(h.Sum(nil)), nil
}

// rubricYAML mirrors the on-disk schema for rubrics.
type rubricYAML struct {
	FeatureKey string                `yaml:"feature_key"`
	Version    string                `yaml:"version"`
	Language   string                `yaml:"language"`
	Dimensions []rubricDimensionYAML `yaml:"dimensions"`
}

type rubricDimensionYAML struct {
	Name        string           `yaml:"name"`
	Weight      float64          `yaml:"weight"`
	Description string           `yaml:"description"`
	ScoreLevels []scoreLevelYAML `yaml:"score_levels"`
}

type scoreLevelYAML struct {
	Label       string  `yaml:"label"`
	Threshold   float64 `yaml:"threshold"`
	Description string  `yaml:"description"`
}

func readRubric(yamlPath string) (*rubricEntry, error) {
	yamlBytes, err := os.ReadFile(yamlPath)
	if err != nil {
		return nil, fmt.Errorf("registry: read %s: %w", yamlPath, err)
	}
	var raw rubricYAML
	if err := yaml.Unmarshal(yamlBytes, &raw); err != nil {
		return nil, fmt.Errorf("registry: parse %s: %w", yamlPath, err)
	}
	if raw.FeatureKey == "" || raw.Version == "" || raw.Language == "" {
		return nil, fmt.Errorf("registry: rubric %s missing required meta field", yamlPath)
	}
	if len(raw.Dimensions) == 0 {
		return nil, fmt.Errorf("registry: rubric %s has no dimensions", yamlPath)
	}

	dims := make([]RubricDimension, 0, len(raw.Dimensions))
	for _, d := range raw.Dimensions {
		levels := make([]ScoreLevel, 0, len(d.ScoreLevels))
		for _, sl := range d.ScoreLevels {
			levels = append(levels, ScoreLevel{
				Label:       sl.Label,
				Threshold:   sl.Threshold,
				Description: sl.Description,
			})
		}
		dims = append(dims, RubricDimension{
			Name:        d.Name,
			Weight:      d.Weight,
			Description: d.Description,
			ScoreLevels: levels,
		})
	}

	return &rubricEntry{
		schema: RubricSchema{
			FeatureKey: raw.FeatureKey,
			Version:    raw.Version,
			Language:   raw.Language,
			Dimensions: dims,
		},
		yamlPath: yamlPath,
	}, nil
}

// validateLanguageParity ensures every feature_key has canonical multi
// prompt/rubric entries and any language override is present on both sides.
// Mismatch is a hard failure at startup so suspended baselines do not slip
// into staging.
func validateLanguageParity(snap *snapshot) error {
	// Every prompt feature_key must have a rubric.
	for fk, langs := range snap.prompts {
		rubricLangs, ok := snap.rubrics[fk]
		if !ok {
			return fmt.Errorf("registry: feature_key %q has prompt(s) but no rubric", fk)
		}
		if _, ok := langs["multi"]; !ok {
			return fmt.Errorf("registry: feature_key %q missing canonical multi prompt", fk)
		}
		if _, ok := rubricLangs["multi"]; !ok {
			return fmt.Errorf("registry: feature_key %q missing canonical multi rubric", fk)
		}
		for lang := range langs {
			if _, ok := rubricLangs[lang]; !ok {
				return fmt.Errorf(
					"registry: feature_key %q has prompt language %q but no matching rubric", fk, lang,
				)
			}
		}
	}
	// Every rubric feature_key must have a prompt.
	for fk, langs := range snap.rubrics {
		promptLangs, ok := snap.prompts[fk]
		if !ok {
			return fmt.Errorf("registry: feature_key %q has rubric(s) but no prompt", fk)
		}
		for lang := range langs {
			if _, ok := promptLangs[lang]; !ok {
				return fmt.Errorf(
					"registry: feature_key %q has rubric language %q but no matching prompt", fk, lang,
				)
			}
		}
	}
	return nil
}
