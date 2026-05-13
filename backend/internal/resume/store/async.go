package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/targetjob"
)

func (r *Repository) ClaimNextAsyncJob(ctx context.Context, jobTypes []string, now time.Time) (targetjob.ClaimedJob, bool, error) {
	if r == nil || r.db == nil {
		return targetjob.ClaimedJob{}, false, fmt.Errorf("resume store db is nil")
	}
	if len(jobTypes) == 0 {
		return targetjob.ClaimedJob{}, false, fmt.Errorf("ClaimNextAsyncJob requires at least one job type")
	}
	placeholders := make([]string, 0, len(jobTypes))
	args := make([]any, 0, len(jobTypes)+1)
	for i, jt := range jobTypes {
		placeholders = append(placeholders, fmt.Sprintf("$%d", i+1))
		args = append(args, jt)
	}
	args = append(args, now)
	query := fmt.Sprintf(`
update async_jobs
set status = 'running',
    attempts = attempts + 1,
    locked_at = $%[1]d,
    updated_at = $%[1]d
where id = (
  select id from async_jobs
  where status = 'queued' and available_at <= $%[1]d and job_type in (%s)
  order by available_at asc, created_at asc
  for update skip locked
  limit 1
)
returning id, job_type, resource_type, resource_id, payload, attempts, max_attempts, available_at`,
		len(args),
		strings.Join(placeholders, ","),
	)
	var claimed targetjob.ClaimedJob
	var payload []byte
	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&claimed.JobID,
		&claimed.JobType,
		&claimed.ResourceType,
		&claimed.ResourceID,
		&payload,
		&claimed.Attempts,
		&claimed.MaxAttempts,
		&claimed.AvailableAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return targetjob.ClaimedJob{}, false, nil
		}
		return targetjob.ClaimedJob{}, false, fmt.Errorf("claim async_jobs: %w", err)
	}
	if len(payload) > 0 {
		claimed.Payload = append([]byte{}, payload...)
	}
	return claimed, true, nil
}

func (r *Repository) FinalizeAsyncJob(ctx context.Context, jobID string, outcome targetjob.JobOutcome, now time.Time) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("resume store db is nil")
	}
	if jobID == "" {
		return fmt.Errorf("FinalizeAsyncJob requires jobID")
	}
	if outcome.Succeeded {
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
			return fmt.Errorf("finalize async_jobs succeeded: %w", err)
		}
		return nil
	}
	if outcome.Retryable {
		_, err := r.db.ExecContext(ctx, `
update async_jobs
set status = case when attempts >= max_attempts then 'dead' else 'queued' end,
    available_at = $1,
    updated_at = $1,
    locked_at = null,
    error_code = $2,
    error_message = $3
where id = $4`,
			now.Add(15*time.Second),
			outcome.ErrorCode,
			nullableString(outcome.ErrorMessage),
			jobID,
		)
		if err != nil {
			return fmt.Errorf("finalize async_jobs retryable: %w", err)
		}
		return nil
	}
	_, err := r.db.ExecContext(ctx, `
update async_jobs
set status = 'failed',
    completed_at = $1,
    updated_at = $1,
    locked_at = null,
    error_code = $2,
    error_message = $3
where id = $4`,
		now,
		outcome.ErrorCode,
		nullableString(outcome.ErrorMessage),
		jobID,
	)
	if err != nil {
		return fmt.Errorf("finalize async_jobs failed: %w", err)
	}
	return nil
}
