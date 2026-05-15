package review

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	reviewdomain "github.com/monshunter/easyinterview/backend/internal/review"
)

func (r *Repository) LeaseAsyncJob(ctx context.Context, jobType string, now time.Time) (reviewdomain.AsyncJob, bool, error) {
	if err := r.checkDB(); err != nil {
		return reviewdomain.AsyncJob{}, false, err
	}
	if jobType == "" {
		return reviewdomain.AsyncJob{}, false, fmt.Errorf("LeaseAsyncJob requires jobType")
	}
	var (
		job      reviewdomain.AsyncJob
		payload  []byte
		lockedAt sql.NullTime
	)
	err := r.db.QueryRowContext(ctx, `
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
returning id, job_type, resource_type, resource_id, payload, attempts, max_attempts, available_at, locked_at`,
		jobType,
		now,
	).Scan(
		&job.JobID,
		&job.JobType,
		&job.ResourceType,
		&job.ResourceID,
		&payload,
		&job.Attempts,
		&job.MaxAttempts,
		&job.AvailableAt,
		&lockedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return reviewdomain.AsyncJob{}, false, nil
	}
	if err != nil {
		return reviewdomain.AsyncJob{}, false, fmt.Errorf("lease async_jobs: %w", err)
	}
	if len(payload) > 0 {
		job.Payload = append([]byte{}, payload...)
	}
	if lockedAt.Valid {
		job.LockedAt = &lockedAt.Time
	}
	return job, true, nil
}

func (r *Repository) UpdateAsyncJobSucceeded(ctx context.Context, jobID string, now time.Time) error {
	if err := r.checkDB(); err != nil {
		return err
	}
	if jobID == "" {
		return fmt.Errorf("UpdateAsyncJobSucceeded requires jobID")
	}
	_, err := r.db.ExecContext(ctx, `
update async_jobs
set status = 'succeeded',
    completed_at = $1,
    updated_at = $1,
    locked_at = null,
    error_code = null,
    error_message = null
where id = $2`, now, jobID)
	if err != nil {
		return fmt.Errorf("update async_jobs succeeded: %w", err)
	}
	return nil
}

func (r *Repository) UpdateAsyncJobFailed(ctx context.Context, in reviewdomain.AsyncJobFailure) error {
	if err := r.checkDB(); err != nil {
		return err
	}
	if in.JobID == "" {
		return fmt.Errorf("UpdateAsyncJobFailed requires jobID")
	}
	_, err := r.db.ExecContext(ctx, `
update async_jobs
set status = case when $2 and attempts < max_attempts then 'queued' else 'failed' end,
    completed_at = case when $2 and attempts < max_attempts then null else $1 end,
    available_at = case when $2 and attempts < max_attempts then $5 else available_at end,
    updated_at = $1,
    locked_at = null,
    error_code = $3,
    error_message = $4
where id = $6`,
		in.Now,
		in.Retryable,
		in.ErrorCode,
		nullableString(in.Error),
		in.AvailableAt,
		in.JobID,
	)
	if err != nil {
		return fmt.Errorf("update async_jobs failed: %w", err)
	}
	return nil
}
