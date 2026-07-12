package main

import (
	"context"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/monshunter/easyinterview/backend/internal/migrations"
)

func TestPracticePlanRoundIdentityBackfillRegistrationAndLedgerIdempotency(t *testing.T) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	manifest := filepath.Join(filepath.Dir(file), "..", "..", "..", "migrations", "backfill", "manifest.yaml")
	entries, err := migrations.LoadBackfillManifest(manifest)
	if err != nil {
		t.Fatalf("LoadBackfillManifest returned error: %v", err)
	}
	if len(entries) != 1 || entries[0].Version != 17 || entries[0].Name != "practice_plan_round_identity" {
		t.Fatalf("unexpected backfill manifest entries: %#v", entries)
	}
	registry := migrations.RegisteredBackfills()
	if registry[entries[0].Name] == nil {
		t.Fatalf("backfill %q was not registered by cmd/migrate", entries[0].Name)
	}

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	expectLedgerRun(mock, entries[0], "dry_run")
	expectEmptyCandidateBatch(mock)
	expectLedgerSuccess(mock)
	expectLedgerRun(mock, entries[0], "apply")
	expectEmptyCandidateBatch(mock)
	expectLedgerSuccess(mock)
	mock.ExpectQuery("select exists").
		WithArgs(17, "dry_run", entries[0].Checksum).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectQuery("select exists").
		WithArgs(17, "apply", entries[0].Checksum).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	cmd := migrations.Command{AppEnv: "dev"}
	if err := migrations.RunBackfillEntries(context.Background(), db, cmd, entries, registry); err != nil {
		t.Fatalf("first RunBackfillEntries returned error: %v", err)
	}
	if err := migrations.RunBackfillEntries(context.Background(), db, cmd, entries, registry); err != nil {
		t.Fatalf("second RunBackfillEntries returned error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func expectLedgerRun(mock sqlmock.Sqlmock, entry migrations.BackfillEntry, mode string) {
	mock.ExpectQuery("select exists").
		WithArgs(entry.Version, mode, entry.Checksum).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	mock.ExpectExec("insert into schema_backfills").
		WithArgs(entry.Version, entry.Name, mode, entry.Checksum).
		WillReturnResult(sqlmock.NewResult(0, 1))
}

func expectEmptyCandidateBatch(mock sqlmock.Sqlmock) {
	mock.ExpectQuery("select p.id::text, p.time_budget_minutes, j.summary").
		WithArgs("00000000-0000-0000-0000-000000000000", 500).
		WillReturnRows(sqlmock.NewRows([]string{"id", "time_budget_minutes", "summary"}))
}

func expectLedgerSuccess(mock sqlmock.Sqlmock) {
	mock.ExpectExec("update schema_backfills").
		WillReturnResult(sqlmock.NewResult(0, 1))
}
