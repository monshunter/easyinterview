package v000017

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/monshunter/easyinterview/backend/internal/migrations"
)

const (
	uniquePlanID    = "00000000-0000-0000-0000-000000000101"
	zeroPlanID      = "00000000-0000-0000-0000-000000000102"
	ambiguousPlanID = "00000000-0000-0000-0000-000000000103"
	overflowPlanID  = "00000000-0000-0000-0000-000000000104"
)

func TestRunDryRunDoesNotMutateUniqueZeroOrAmbiguousMatches(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	expectCandidateBatch(mock, candidateRows())
	if err := Run(context.Background(), db, migrations.BackfillModeDryRun); err != nil {
		t.Fatalf("Run dry-run returned error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestRunApplyBackfillsOnlyUniqueDurationMatch(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	expectCandidateBatch(mock, candidateRows())
	mock.ExpectExec("update practice_plans").
		WithArgs(uniquePlanID, "round-2-technical", 2).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := Run(context.Background(), db, migrations.BackfillModeApply); err != nil {
		t.Fatalf("Run apply returned error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestRunApplyIsRerunSafeAfterRowsAreAlreadyBackfilled(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	expectCandidateBatch(mock, sqlmock.NewRows([]string{"id", "time_budget_minutes", "summary"}).
		AddRow(uniquePlanID, 45, structuredRoundsJSON()))
	mock.ExpectExec("update practice_plans").
		WithArgs(uniquePlanID, "round-2-technical", 2).
		WillReturnResult(sqlmock.NewResult(0, 1))
	expectCandidateBatch(mock, sqlmock.NewRows([]string{"id", "time_budget_minutes", "summary"}))

	if err := Run(context.Background(), db, migrations.BackfillModeApply); err != nil {
		t.Fatalf("first Run apply returned error: %v", err)
	}
	if err := Run(context.Background(), db, migrations.BackfillModeApply); err != nil {
		t.Fatalf("second Run apply returned error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestRunRejectsUnknownMode(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if err := Run(context.Background(), db, migrations.BackfillMode("unknown")); err == nil {
		t.Fatal("expected unknown mode to fail")
	}
}

func expectCandidateBatch(mock sqlmock.Sqlmock, rows *sqlmock.Rows) {
	mock.ExpectQuery(`(?s)select p\.id::text, p\.time_budget_minutes, j\.summary.*j\.user_id = p\.user_id.*j\.resume_id = p\.resume_id.*j\.deleted_at is null`).
		WithArgs("00000000-0000-0000-0000-000000000000", 500).
		WillReturnRows(rows)
}

func candidateRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{"id", "time_budget_minutes", "summary"}).
		AddRow(uniquePlanID, 45, structuredRoundsJSON()).
		AddRow(zeroPlanID, 90, structuredRoundsJSON()).
		AddRow(ambiguousPlanID, 45, `{"interviewRounds":[{"sequence":1,"type":"hr","durationMinutes":45},{"sequence":2,"type":"technical","durationMinutes":45}]}`).
		AddRow(overflowPlanID, 30, `{"interviewRounds":[{"sequence":2147483648,"type":"technical","durationMinutes":30}]}`)
}

func structuredRoundsJSON() string {
	return `{"interviewRounds":[{"sequence":1,"type":"hr","durationMinutes":30},{"sequence":2,"type":"technical","durationMinutes":45}]}`
}
