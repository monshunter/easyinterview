package migrations

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestLoadBackfillManifest(t *testing.T) {
	path := writeBackfillManifest(t, `
backfills:
  - version: 1
    name: baseline_noop
    checksum: sha256:baseline
    reversible: true
    dryRun: true
`)

	entries, err := LoadBackfillManifest(path)
	if err != nil {
		t.Fatalf("LoadBackfillManifest returned error: %v", err)
	}

	want := []BackfillEntry{{
		Version:    1,
		Name:       "baseline_noop",
		Checksum:   "sha256:baseline",
		Reversible: true,
		DryRun:     true,
	}}
	if !reflect.DeepEqual(entries, want) {
		t.Fatalf("entries mismatch\nwant: %#v\n got: %#v", want, entries)
	}
}

func TestRunBackfillEntriesWritesDryRunAndApplyLedger(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mockBackfillLedgerInsert(mock)
	mockBackfillLedgerInsert(mock)
	var modes []BackfillMode

	err = RunBackfillEntries(context.Background(), db, Command{AppEnv: "dev"}, []BackfillEntry{{
		Version:  1,
		Name:     "baseline_noop",
		Checksum: "sha256:baseline",
		DryRun:   true,
	}}, BackfillRegistry{
		"baseline_noop": func(_ context.Context, _ *sql.DB, mode BackfillMode) error {
			modes = append(modes, mode)
			return nil
		},
	})
	if err != nil {
		t.Fatalf("RunBackfillEntries returned error: %v", err)
	}
	if !reflect.DeepEqual(modes, []BackfillMode{BackfillModeDryRun, BackfillModeApply}) {
		t.Fatalf("unexpected modes: %#v", modes)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestRunBackfillEntriesSkipsRepeatedSuccess(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectQuery("select exists").
		WithArgs(1, "dry_run", "sha256:baseline").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectQuery("select exists").
		WithArgs(1, "apply", "sha256:baseline").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	ran := false

	err = RunBackfillEntries(context.Background(), db, Command{AppEnv: "dev"}, []BackfillEntry{{
		Version:  1,
		Name:     "baseline_noop",
		Checksum: "sha256:baseline",
		DryRun:   true,
	}}, BackfillRegistry{
		"baseline_noop": func(context.Context, *sql.DB, BackfillMode) error {
			ran = true
			return nil
		},
	})
	if err != nil {
		t.Fatalf("RunBackfillEntries returned error: %v", err)
	}
	if ran {
		t.Fatal("expected repeated success to skip backfill function")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestRunBackfillEntriesRejectsProdForce(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	err = RunBackfillEntries(context.Background(), db, Command{AppEnv: "prod", ForceBackfill: true}, []BackfillEntry{{
		Version:  1,
		Name:     "baseline_noop",
		Checksum: "sha256:baseline",
	}}, BackfillRegistry{"baseline_noop": func(context.Context, *sql.DB, BackfillMode) error { return nil }})
	if err == nil {
		t.Fatal("expected prod force to be rejected")
	}
}

func mockBackfillLedgerInsert(mock sqlmock.Sqlmock) {
	mock.ExpectQuery("select exists").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	mock.ExpectExec("insert into schema_backfills").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("update schema_backfills").
		WillReturnResult(sqlmock.NewResult(1, 1))
}

func writeBackfillManifest(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "manifest.yaml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}
