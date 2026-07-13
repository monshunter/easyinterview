package review

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/monshunter/easyinterview/backend/internal/runner"
)

func TestAssertCurrentReportJobLeaseChecksRunningClaimWithoutBusinessWrite(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectQuery(`select attempts[\s\S]*from async_jobs[\s\S]*status = 'running'[\s\S]*attempts = \$2[\s\S]*for update`).
		WithArgs("job-1", int32(2)).
		WillReturnRows(sqlmock.NewRows([]string{"attempts"}).AddRow(2))
	mock.ExpectCommit()

	if err := NewRepository(db).AssertCurrentReportJobLease(context.Background(), "job-1", 2); err != nil {
		t.Fatalf("AssertCurrentReportJobLease: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestAssertCurrentReportJobLeaseRejectsStaleGeneration(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectQuery(`select attempts[\s\S]*from async_jobs[\s\S]*status = 'running'[\s\S]*attempts = \$2[\s\S]*for update`).
		WithArgs("job-1", int32(1)).
		WillReturnRows(sqlmock.NewRows([]string{"attempts"}))
	mock.ExpectRollback()

	err = NewRepository(db).AssertCurrentReportJobLease(context.Background(), "job-1", 1)
	if !errors.Is(err, runner.ErrStaleLease) {
		t.Fatalf("AssertCurrentReportJobLease err=%v want ErrStaleLease", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
