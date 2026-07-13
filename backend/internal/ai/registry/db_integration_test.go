package registry_test

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// TestSeedMigrationCoversBaselineFeatureKeys statically validates the F3
// seed migrations against the on-disk prompt/rubric truth source.
// It does not start a Postgres instance — the dockertest /
// pgtestdb harness is not part of this repo's testing surface yet, so the
// active gate is the static SQL parse below plus the schema check inside
// `make migrate-check` (which exercises the actual `up -> down -> up`
// path under DATABASE_URL when one is configured).
//
// When the repo gains a Postgres-backed
// integration harness, this test should be promoted to a `//go:build
// integration` runtime test that re-runs the migration chain end-to-end.
func TestSeedMigrationCoversBaselineFeatureKeys(t *testing.T) {
	t.Parallel()

	repoRoot := walkUpToRepoRoot(t)
	wantPrompts := expectedPromptRows(t, repoRoot)
	wantRubrics := expectedRubricRows(t, repoRoot)
	rows := extractCurrentMigrationRows(t, repoRoot)

	assertSeedRows(t, "prompt_versions", wantPrompts, rows["prompt_versions"])
	assertSeedRows(t, "rubric_versions", wantRubrics, rows["rubric_versions"])
}

type insertRow struct {
	featureKey   string
	version      string
	language     string
	templateHash string
}

func expectedPromptRows(t *testing.T, repoRoot string) map[string]insertRow {
	t.Helper()

	paths, err := filepath.Glob(filepath.Join(repoRoot, "config", "prompts", "*", "v*.yaml"))
	if err != nil {
		t.Fatalf("glob prompt baselines: %v", err)
	}
	sort.Strings(paths)

	out := map[string]insertRow{}
	for _, path := range paths {
		bytes, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read prompt baseline %s: %v", path, err)
		}
		var meta struct {
			FeatureKey   string `yaml:"feature_key"`
			Version      string `yaml:"version"`
			Language     string `yaml:"language"`
			TemplateHash string `yaml:"template_hash"`
			Status       string `yaml:"status"`
		}
		if err := yaml.Unmarshal(bytes, &meta); err != nil {
			t.Fatalf("parse prompt baseline %s: %v", path, err)
		}
		if meta.Status != "active" {
			continue
		}
		if meta.Language != "multi" {
			continue
		}
		row := insertRow{
			featureKey:   meta.FeatureKey,
			version:      meta.Version,
			language:     meta.Language,
			templateHash: meta.TemplateHash,
		}
		addExpectedRow(t, out, row, path)
	}
	if len(out) == 0 {
		t.Fatal("no active prompt baselines found")
	}
	return out
}

func expectedRubricRows(t *testing.T, repoRoot string) map[string]insertRow {
	t.Helper()

	paths, err := filepath.Glob(filepath.Join(repoRoot, "config", "rubrics", "*", "v*.yaml"))
	if err != nil {
		t.Fatalf("glob rubric baselines: %v", err)
	}
	sort.Strings(paths)

	out := map[string]insertRow{}
	for _, path := range paths {
		bytes, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read rubric baseline %s: %v", path, err)
		}
		var meta struct {
			FeatureKey string `yaml:"feature_key"`
			Version    string `yaml:"version"`
			Language   string `yaml:"language"`
			Status     string `yaml:"status"`
		}
		if err := yaml.Unmarshal(bytes, &meta); err != nil {
			t.Fatalf("parse rubric baseline %s: %v", path, err)
		}
		if meta.Status != "active" || meta.Language != "multi" {
			continue
		}
		row := insertRow{
			featureKey: meta.FeatureKey,
			version:    meta.Version,
			language:   meta.Language,
		}
		addExpectedRow(t, out, row, path)
	}
	if len(out) == 0 {
		t.Fatal("no rubric baselines found")
	}
	return out
}

func addExpectedRow(t *testing.T, rows map[string]insertRow, row insertRow, path string) {
	t.Helper()

	key := rowKey(row)
	if key == "||" {
		t.Fatalf("empty baseline coordinate in %s", path)
	}
	if _, ok := rows[key]; ok {
		t.Fatalf("duplicate baseline coordinate %s in %s", key, path)
	}
	rows[key] = row
}

func extractCurrentMigrationRows(t *testing.T, repoRoot string) map[string][]insertRow {
	t.Helper()

	paths, err := filepath.Glob(filepath.Join(repoRoot, "migrations", "*prompt_rubric*.up.sql"))
	if err != nil {
		t.Fatalf("glob prompt/rubric migrations: %v", err)
	}
	sort.Strings(paths)
	if len(paths) == 0 {
		t.Fatal("no prompt/rubric migrations found")
	}

	allRows := map[string]map[string]insertRow{
		"prompt_versions": {},
		"rubric_versions": {},
	}
	activeVersions := map[string]map[string]string{
		"prompt_versions": {},
		"rubric_versions": {},
	}
	updateRe := regexp.MustCompile(`(?is)UPDATE\s+(prompt_versions|rubric_versions)\s+SET\s+is_active\s*=\s*\(version\s*=\s*'([^']+)'\)\s+WHERE\s+feature_key\s+IN\s*\(([^)]+)\)\s+AND\s+language\s*=\s*'([^']+)'`)
	keyRe := regexp.MustCompile(`'([^']+)'`)
	for _, path := range paths {
		body, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read prompt/rubric migration %s: %v", path, err)
		}
		rows := extractInsertRows(string(body))
		for table, inserted := range rows {
			for _, row := range inserted {
				key := rowKey(row)
				if _, duplicate := allRows[table][key]; duplicate {
					t.Fatalf("%s duplicate migration row %s", table, key)
				}
				allRows[table][key] = row
				if strings.Contains(filepath.Base(path), "seed_baseline_prompt_rubric") {
					activeVersions[table][row.featureKey+"|"+row.language] = row.version
				}
			}
		}
		for _, match := range updateRe.FindAllStringSubmatch(string(body), -1) {
			table, version, rawKeys, language := match[1], match[2], match[3], match[4]
			for _, key := range keyRe.FindAllStringSubmatch(rawKeys, -1) {
				activeVersions[table][key[1]+"|"+language] = version
			}
		}
	}

	// Later module-removal migrations (e.g. product-scope v2.1 D-17 dropping
	// the jd_match feature keys) delete previously seeded rows. The net DB
	// state — not the raw seed inserts — is what must match the on-disk
	// config truth source, so subtract out-of-scope feature keys here.
	outOfScope := outOfScopeFeatureKeys(t, repoRoot)
	out := map[string][]insertRow{"prompt_versions": {}, "rubric_versions": {}}
	for table, coordinates := range activeVersions {
		for coordinate, version := range coordinates {
			parts := strings.Split(coordinate, "|")
			if len(parts) != 2 || outOfScope[parts[0]] {
				continue
			}
			key := parts[0] + "|" + version + "|" + parts[1]
			row, ok := allRows[table][key]
			if !ok {
				t.Errorf("%s active migration coordinate has no inserted row: %s", table, key)
				continue
			}
			out[table] = append(out[table], row)
		}
		sort.Slice(out[table], func(i, j int) bool { return rowKey(out[table][i]) < rowKey(out[table][j]) })
	}
	return out
}

// outOfScopeFeatureKeys parses `DELETE FROM prompt_versions ... feature_key IN
// (...)` statements from module-removal migrations so the static seed gate
// tracks the post-migration net state.
func outOfScopeFeatureKeys(t *testing.T, repoRoot string) map[string]bool {
	t.Helper()

	out := map[string]bool{}
	paths, err := filepath.Glob(filepath.Join(repoRoot, "migrations", "*drop*_module.up.sql"))
	if err != nil {
		t.Fatalf("glob removal migrations: %v", err)
	}
	deleteRe := regexp.MustCompile(`DELETE FROM (?:prompt|rubric)_versions WHERE feature_key IN \(([^)]+)\)`)
	keyRe := regexp.MustCompile(`'([^']+)'`)
	for _, path := range paths {
		body, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read removal migration %s: %v", path, err)
		}
		for _, m := range deleteRe.FindAllStringSubmatch(string(body), -1) {
			for _, key := range keyRe.FindAllStringSubmatch(m[1], -1) {
				out[key[1]] = true
			}
		}
	}
	return out
}

func assertSeedRows(t *testing.T, table string, want map[string]insertRow, gotRows []insertRow) {
	t.Helper()

	got := map[string]insertRow{}
	for _, row := range gotRows {
		key := rowKey(row)
		if _, ok := got[key]; ok {
			t.Errorf("%s duplicate seed row: %s", table, key)
			continue
		}
		got[key] = row
	}

	missing, extra := diffRowKeys(want, got)
	if len(missing) > 0 {
		t.Errorf("%s missing seed rows: %v", table, missing)
	}
	if len(extra) > 0 {
		t.Errorf("%s unexpected seed rows: %v", table, extra)
	}
	if len(gotRows) != len(want) {
		t.Errorf("%s seed row count: want %d, got %d", table, len(want), len(gotRows))
	}

	if table != "prompt_versions" {
		return
	}
	for key, wantRow := range want {
		gotRow, ok := got[key]
		if !ok {
			continue
		}
		if gotRow.templateHash != wantRow.templateHash {
			t.Errorf("%s %s template_hash drift: yaml=%s seed=%s",
				table, key, wantRow.templateHash, gotRow.templateHash)
		}
	}
}

func diffRowKeys(want, got map[string]insertRow) ([]string, []string) {
	missing := make([]string, 0)
	extra := make([]string, 0)
	for key := range want {
		if _, ok := got[key]; !ok {
			missing = append(missing, key)
		}
	}
	for key := range got {
		if _, ok := want[key]; !ok {
			extra = append(extra, key)
		}
	}
	sort.Strings(missing)
	sort.Strings(extra)
	return missing, extra
}

func rowKey(row insertRow) string {
	return row.featureKey + "|" + row.version + "|" + row.language
}

// rowRe matches the leading literal columns shared by both prompt and
// rubric INSERT VALUES tuples (id, feature_key, version, language). The
// trailing column differs between tables (template_hash for prompts,
// schema_json for rubrics) so the regex captures it as an optional
// hex-only group; rubric rows leave it empty.
var rowRe = regexp.MustCompile(
	`\(\s*'[^']+'\s*,\s*'(?P<feature_key>[^']+)'\s*,\s*'(?P<version>[^']+)'\s*,\s*'(?P<language>[^']+)'\s*,\s*('(?P<template_hash>[a-fA-F0-9]+)'|\$)`,
)

func extractInsertRows(sql string) map[string][]insertRow {
	out := map[string][]insertRow{
		"prompt_versions": {},
		"rubric_versions": {},
	}
	insertRe := regexp.MustCompile(`(?is)INSERT\s+INTO\s+(\w+)[^;]*?VALUES\s*(.*?)ON\s+CONFLICT`)
	for _, match := range insertRe.FindAllStringSubmatch(sql, -1) {
		table := match[1]
		body := match[2]
		for _, row := range rowRe.FindAllStringSubmatch(body, -1) {
			// Submatch indexes follow the regex: 1=feature_key,
			// 2=version, 3=language, 4=outer alternation,
			// 5=template_hash (empty for rubric rows).
			out[table] = append(out[table], insertRow{
				featureKey:   row[1],
				version:      row[2],
				language:     row[3],
				templateHash: row[5],
			})
		}
	}
	return out
}

func walkUpToRepoRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	dir := wd
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return filepath.Dir(dir) // backend/.. == repo root
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Skipf("could not locate backend go.mod from %s", wd)
			return ""
		}
		dir = parent
	}
}
