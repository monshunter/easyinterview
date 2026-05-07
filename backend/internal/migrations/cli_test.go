package migrations

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestRunRejectsMissingDatabaseURLForMigrationCommands(t *testing.T) {
	var stderr bytes.Buffer

	exitCode := Run(context.Background(), []string{"up"}, StaticEnv{
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

	exitCode := Run(context.Background(), []string{"down"}, StaticEnv{
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

	exitCode := Run(context.Background(), []string{"--migrations-dir", tmp, "create", "add_test_table"}, StaticEnv{}, &stdout, nil)

	if exitCode != 0 {
		t.Fatalf("expected create to succeed, exit=%d", exitCode)
	}
	if !strings.Contains(stdout.String(), "000001_add_test_table.up.sql") {
		t.Fatalf("stdout should include created file name, got %q", stdout.String())
	}
}

func TestRunDispatchesPrivacyMatrixDryRunWithoutDatabase(t *testing.T) {
	var stdout bytes.Buffer

	exitCode := Run(context.Background(), []string{"privacy-matrix", "--dry-run"}, StaticEnv{}, &stdout, nil)

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
		"candidate_profiles",
		"experience_cards",
		"file_objects",
		"resume_assets",
		"target_jobs",
		"target_job_requirements",
		"target_job_sources",
		"practice_plans",
		"practice_sessions",
		"practice_session_events",
		"practice_turns",
		"feedback_reports",
		"question_assessments",
		"resume_tailor_runs",
		"debriefs",
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
	if len(got) != 30 {
		t.Fatalf("privacy matrix should cover exactly 30 public baseline tables, got %d: %#v", len(got), got)
	}
}
