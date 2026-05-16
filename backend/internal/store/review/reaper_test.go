package review_test

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	reviewstore "github.com/monshunter/easyinterview/backend/internal/store/review"
)

func TestReaperReclaimsExpiredLease(t *testing.T) {
	now := time.Date(2026, 5, 15, 15, 30, 0, 0, time.UTC)
	olderThan := now.Add(-2 * time.Minute)
	db, mock, cleanup := newMockReviewStore(t)
	defer cleanup()
	repo := reviewstore.NewRepository(db)

	mock.ExpectExec(regexp.QuoteMeta(`
update async_jobs
set status = 'queued',
    locked_at = null,
    updated_at = $3
where job_type = $1
  and status = 'running'
  and locked_at is not null
  and locked_at < $2`)).
		WithArgs("report_generate", olderThan, now).
		WillReturnResult(sqlmock.NewResult(0, 1))

	count, err := repo.ReclaimExpiredLeases(context.Background(), "report_generate", olderThan, now)
	if err != nil {
		t.Fatalf("ReclaimExpiredLeases: %v", err)
	}
	if count != 1 {
		t.Fatalf("reclaimed = %d, want 1", count)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
