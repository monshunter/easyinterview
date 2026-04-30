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
		"APP_ENV":      "prod",
		"DATABASE_URL": "postgres://example.invalid/easyinterview",
	}, nil, &stderr)

	if exitCode == 0 {
		t.Fatal("expected prod down to fail")
	}
	if !strings.Contains(stderr.String(), "MIGRATE_DOWN_FORCE=1") {
		t.Fatalf("stderr should describe the force gate, got %q", stderr.String())
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
