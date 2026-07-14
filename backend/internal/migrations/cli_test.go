package migrations

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

type mapEnv map[string]string

func (e mapEnv) Getenv(key string) string {
	return e[key]
}

func TestRunRejectsMissingDatabaseURLForMigrationCommands(t *testing.T) {
	var stderr bytes.Buffer

	exitCode := Run(context.Background(), []string{"up"}, mapEnv{
		"APP_ENV": "dev",
	}, nil, &stderr)

	if exitCode == 0 {
		t.Fatal("expected missing DATABASE_URL to fail")
	}
	if !strings.Contains(stderr.String(), "DATABASE_URL") {
		t.Fatalf("stderr should mention DATABASE_URL, got %q", stderr.String())
	}
}

func TestRunRejectsProdDownWithoutForceBeforeOpeningDatabase(t *testing.T) {
	var stderr bytes.Buffer

	exitCode := Run(context.Background(), []string{"down"}, mapEnv{
		"APP_ENV": "prod",
	}, nil, &stderr)

	if exitCode == 0 {
		t.Fatal("expected prod down to fail")
	}
	if !strings.Contains(stderr.String(), "MIGRATE_DOWN_FORCE=1") {
		t.Fatalf("stderr should describe the force gate, got %q", stderr.String())
	}
	if strings.Contains(stderr.String(), "DATABASE_URL") {
		t.Fatalf("prod guard must run before DATABASE_URL validation, got %q", stderr.String())
	}
}

func TestRunDispatchesCreateThroughGenerator(t *testing.T) {
	tmp := t.TempDir()
	var stdout bytes.Buffer

	exitCode := Run(context.Background(), []string{"--migrations-dir", tmp, "create", "add_test_table"}, mapEnv{}, &stdout, nil)

	if exitCode != 0 {
		t.Fatalf("expected create to succeed, exit=%d", exitCode)
	}
	if !strings.Contains(stdout.String(), "000001_add_test_table.up.sql") {
		t.Fatalf("stdout should include created file name, got %q", stdout.String())
	}
}

func TestRunDispatchesPrivacyMatrixDryRunWithoutDatabase(t *testing.T) {
	var stdout bytes.Buffer

	exitCode := Run(context.Background(), []string{"privacy-matrix", "--dry-run"}, mapEnv{}, &stdout, nil)

	if exitCode != 0 {
		t.Fatalf("expected dry-run to succeed, exit=%d", exitCode)
	}
	for _, want := range []string{"prompt_versions: retain", "schema_backfills: retain", "ai_task_runs: hard_delete_after_audit_summary"} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("dry-run output should contain %q, got %q", want, stdout.String())
		}
	}
}

func TestPrivacyMatrixCoversBaselineRetainTables(t *testing.T) {
	var stdout bytes.Buffer

	WritePrivacyMatrix(&stdout)
	out := stdout.String()
	for _, want := range []string{
		"users: sync_soft_delete_then_hard_delete",
		"auth_challenges: hard_delete",
		"sessions: hard_delete",
		"external_identities: hard_delete",
		"prompt_versions: retain",
		"rubric_versions: retain",
		"schema_migrations: retain",
		"schema_backfills: retain",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("privacy matrix missing %q in output:\n%s", want, out)
		}
	}
}

func TestPrivacyMatrixCoversEveryBaselineTableExactly(t *testing.T) {
	var stdout bytes.Buffer

	WritePrivacyMatrix(&stdout)
	got := map[string]string{}
	for _, line := range strings.Split(strings.TrimSpace(stdout.String()), "\n") {
		table, disposition, ok := strings.Cut(line, ": ")
		if !ok {
			t.Fatalf("unexpected privacy matrix line %q", line)
		}
		if previous, exists := got[table]; exists {
			t.Fatalf("privacy matrix has duplicate table %q: %q and %q", table, previous, disposition)
		}
		got[table] = disposition
	}

	for _, table := range []string{
		"users",
		"user_settings",
		"file_objects",
		"resumes",
		"target_jobs",
		"target_job_requirements",
		"practice_plans",
		"idempotency_records",
		"practice_sessions",
		"practice_session_events",
		"practice_messages",
		"feedback_reports",
		"source_records",
		"prompt_versions",
		"rubric_versions",
		"ai_task_runs",
		"async_jobs",
		"outbox_events",
		"privacy_requests",
		"audit_events",
		"auth_challenges",
		"sessions",
		"external_identities",
		"schema_migrations",
		"schema_backfills",
	} {
		if _, ok := got[table]; !ok {
			t.Fatalf("privacy matrix missing baseline table %q", table)
		}
	}
	if _, ok := got["mistake_entries"]; ok {
		t.Fatalf("privacy matrix must not restore removed mistake_entries")
	}
	for _, removed := range []string{"practice_turns", "question_assessments"} {
		if _, ok := got[removed]; ok {
			t.Fatalf("privacy matrix must not restore removed table %q", removed)
		}
	}
	// product-scope v2.1 D-17 dropped the 5 jd_match module tables
	// (jd_match_recommendations, watchlist_items, saved_searches,
	// agent_scans, jd_match_search_runs), trimming 35 -> 30. product-scope
	// v2.1 D-20 resume flatten renamed resume_assets -> resumes and dropped
	// resume_tailor_runs (resume_versions / resume_version_suggestions were
	// already covered by the resumes cascade, not separate matrix rows),
	// trimming 30 -> 29. product-scope D-22 removed candidate profile,
	// experience card, and debrief tables, trimming 29 -> 26. The
	// idempotency_records table remains current and user-owned. Conversation
	// simplification replaces practice_turns with practice_messages and removes
	// question_assessments. TargetJob paste-only then removes target_job_sources,
	// so the matrix covers 25 entries.
	if len(got) != 25 {
		t.Fatalf("privacy matrix should cover exactly 25 public baseline tables, got %d: %#v", len(got), got)
	}
}
