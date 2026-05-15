package review_test

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	reviewdomain "github.com/monshunter/easyinterview/backend/internal/review"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
	reviewstore "github.com/monshunter/easyinterview/backend/internal/store/review"
)

func TestLeaseAsyncJobUsesAttemptsAndLockedAt(t *testing.T) {
	now := time.Date(2026, 5, 15, 14, 0, 0, 0, time.UTC)
	db, mock, cleanup := newMockReviewStore(t)
	defer cleanup()
	repo := reviewstore.NewRepository(db)

	rows := sqlmock.NewRows([]string{"id", "job_type", "resource_type", "resource_id", "payload", "attempts", "max_attempts", "available_at", "locked_at"}).
		AddRow("0197d120-0000-7000-8000-000000000020", "report_generate", "feedback_report", "0197d120-0000-7000-8000-000000000021", []byte(`{"reportId":"0197d120-0000-7000-8000-000000000021"}`), 1, 5, now, now)
	mock.ExpectQuery(regexp.QuoteMeta(`
update async_jobs
set status = 'running',
    attempts = attempts + 1,
    locked_at = $2,
    updated_at = $2
where id = (
  select id from async_jobs
  where status = 'queued' and available_at <= $2 and job_type = $1
  order by available_at asc, created_at asc
  for update skip locked
  limit 1
)
returning id, job_type, resource_type, resource_id, payload, attempts, max_attempts, available_at, locked_at`)).
		WithArgs("report_generate", now).
		WillReturnRows(rows)

	job, ok, err := repo.LeaseAsyncJob(context.Background(), "report_generate", now)
	if err != nil {
		t.Fatalf("LeaseAsyncJob: %v", err)
	}
	if !ok {
		t.Fatal("LeaseAsyncJob ok=false, want true")
	}
	if job.Attempts != 1 || job.LockedAt == nil || !job.LockedAt.Equal(now) {
		t.Fatalf("leased job attempts/locked_at = %+v", job)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestUpdateAsyncJobSucceededClearsLockedAt(t *testing.T) {
	now := time.Date(2026, 5, 15, 14, 30, 0, 0, time.UTC)
	db, mock, cleanup := newMockReviewStore(t)
	defer cleanup()
	repo := reviewstore.NewRepository(db)

	mock.ExpectExec(regexp.QuoteMeta(`
update async_jobs
set status = 'succeeded',
    completed_at = $1,
    updated_at = $1,
    locked_at = null,
    error_code = null,
    error_message = null
where id = $2`)).
		WithArgs(now, "0197d120-0000-7000-8000-000000000022").
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := repo.UpdateAsyncJobSucceeded(context.Background(), "0197d120-0000-7000-8000-000000000022", now); err != nil {
		t.Fatalf("UpdateAsyncJobSucceeded: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestUpdateAsyncJobFailedClearsLockedAt(t *testing.T) {
	now := time.Date(2026, 5, 15, 15, 0, 0, 0, time.UTC)
	db, mock, cleanup := newMockReviewStore(t)
	defer cleanup()
	repo := reviewstore.NewRepository(db)

	mock.ExpectExec(regexp.QuoteMeta(`
update async_jobs
set status = case when $2 and attempts < max_attempts then 'queued' else 'failed' end,
    completed_at = case when $2 and attempts < max_attempts then null else $1 end,
    available_at = case when $2 and attempts < max_attempts then $5 else available_at end,
    updated_at = $1,
    locked_at = null,
    error_code = $3,
    error_message = $4
where id = $6`)).
		WithArgs(now, true, "AI_PROVIDER_TIMEOUT", "timeout", now.Add(30*time.Second), "0197d120-0000-7000-8000-000000000023").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.UpdateAsyncJobFailed(context.Background(), reviewdomain.AsyncJobFailure{
		JobID:       "0197d120-0000-7000-8000-000000000023",
		Retryable:   true,
		ErrorCode:   "AI_PROVIDER_TIMEOUT",
		Error:       "timeout",
		AvailableAt: now.Add(30 * time.Second),
		Now:         now,
	})
	if err != nil {
		t.Fatalf("UpdateAsyncJobFailed: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestReviewJobTypeConstant(t *testing.T) {
	if reviewdomain.ReportGenerateJobType != "report_generate" {
		t.Fatalf("ReportGenerateJobType = %q", reviewdomain.ReportGenerateJobType)
	}
	if string(sharedtypes.JobStatusRunning) != "running" {
		t.Fatalf("JobStatusRunning = %q", sharedtypes.JobStatusRunning)
	}
}
