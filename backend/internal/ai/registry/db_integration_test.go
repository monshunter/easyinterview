package registry_test

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"testing"

	"gopkg.in/yaml.v3"
)

// TestSeedMigrationCoversBaselineFeatureKeys statically validates the F3
// seed migration written by plan §4.4 against the on-disk truth source.
// It does not start a Postgres instance — the dockertest /
// pgtestdb harness is not part of this repo's testing surface yet, so the
// active gate is the static SQL parse below plus the schema check inside
// `make migrate-check` (which exercises the actual `up -> down -> up`
// path under DATABASE_URL when one is configured).
//
// Plan §4.7 verification slot. When the repo gains a Postgres-backed
// integration harness, this test should be promoted to a `//go:build
// integration` runtime test that re-runs the migration chain end-to-end.
func TestSeedMigrationCoversBaselineFeatureKeys(t *testing.T) {
	t.Parallel()

	repoRoot := walkUpToRepoRoot(t)
	migrationPath := filepath.Join(repoRoot, "migrations",
		"000002_seed_baseline_prompt_rubric_versions.up.sql")
	body, err := os.ReadFile(migrationPath)
	if err != nil {
		t.Fatalf("read seed migration: %v", err)
	}

	rows := extractInsertRows(string(body))
	if got := rows["prompt_versions"]; len(got) != 22 {
		t.Errorf("prompt_versions seed rows: want 22, got %d", len(got))
	}
	if got := rows["rubric_versions"]; len(got) != 22 {
		t.Errorf("rubric_versions seed rows: want 22, got %d", len(got))
	}

	wantFeatures := []string{
		"target.import.parse",
		"practice.session.first_question",
		"practice.session.follow_up",
		"practice.turn.lightweight_observe",
		"report.generate",
		"report.question_assessment",
		"resume.parse",
		"resume.tailor.gap_review",
		"resume.tailor.bullet_suggestions",
		"debrief.generate",
		"debrief.suggest_questions",
	}
	sort.Strings(wantFeatures)

	for _, table := range []string{"prompt_versions", "rubric_versions"} {
		fk := uniqueFeatureKeys(rows[table])
		sort.Strings(fk)
		if got, want := fk, wantFeatures; !equalStrings(got, want) {
			t.Errorf("%s feature_keys: want %v, got %v", table, want, got)
		}
	}

	// Cross-file template_hash drift: each prompt INSERT row's
	// template_hash must equal the on-disk yaml meta's template_hash.
	for _, row := range rows["prompt_versions"] {
		yp := filepath.Join(repoRoot, "config", "prompts", row.featureKey, yamlBasename(row))
		bytes, err := os.ReadFile(yp)
		if err != nil {
			t.Errorf("missing baseline yaml %s: %v", yp, err)
			continue
		}
		var meta struct {
			TemplateHash string `yaml:"template_hash"`
		}
		if err := yaml.Unmarshal(bytes, &meta); err != nil {
			t.Errorf("parse %s: %v", yp, err)
			continue
		}
		if row.templateHash != meta.TemplateHash {
			t.Errorf("%s template_hash drift: yaml=%s seed=%s",
				yp, meta.TemplateHash, row.templateHash)
		}
	}
}

type insertRow struct {
	featureKey   string
	version      string
	language     string
	templateHash string
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

func uniqueFeatureKeys(rows []insertRow) []string {
	set := map[string]struct{}{}
	for _, r := range rows {
		set[r.featureKey] = struct{}{}
	}
	out := make([]string, 0, len(set))
	for fk := range set {
		out = append(out, fk)
	}
	return out
}

func yamlBasename(row insertRow) string {
	if row.language == "multi" {
		return row.version + ".yaml"
	}
	return row.version + "." + row.language + ".yaml"
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
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
