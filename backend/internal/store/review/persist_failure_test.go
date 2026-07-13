package review

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	reviewdomain "github.com/monshunter/easyinterview/backend/internal/review"
	"github.com/monshunter/easyinterview/backend/internal/runner"
)

func TestPersistReportFailureStaleLeaseReturnsBeforeBusinessWrites(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	now := time.Unix(7, 0).UTC()
	in := reviewdomain.ReportFailurePersistence{
		UserID: "user-1", ReportID: "report-1", SessionID: "session-1",
		AsyncJobID: "job-1", ClaimedAttempts: 1, OutboxEventID: "outbox-1", AuditEventID: "audit-1",
		ErrorCode: "AI_PROVIDER_TIMEOUT", Retryable: true, MaxAttempts: 4, Now: now,
	}

	mock.ExpectBegin()
	mock.ExpectQuery(`(?s)select attempts.*from async_jobs.*for update`).
		WithArgs(in.AsyncJobID, in.ClaimedAttempts).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectRollback()

	err = NewRepository(db).PersistReportFailure(context.Background(), in)
	if !errors.Is(err, runner.ErrStaleLease) {
		t.Fatalf("PersistReportFailure err=%v want ErrStaleLease", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("stale lease reached a report/outbox/audit/job write: %v", err)
	}
}
