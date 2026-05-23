package runner

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

// LeaseStore is the persistence boundary the kernel needs. The SQLStore below
// is the production implementation; tests inject an in-memory fake. Per spec
// D-3 the kernel owns this SQL directly so domain stores no longer keep
// duplicate claim / finalize copies.
type LeaseStore interface {
	// LeaseAsyncJob atomically claims the oldest queued row whose job_type is in
	// jobTypes and whose available_at <= now, flipping it to running, bumping
	// attempts, and stamping locked_at. (false, nil) means nothing to claim.
	LeaseAsyncJob(ctx context.Context, jobTypes []string, now time.Time) (ClaimedJob, bool, error)
	// FinalizeAsyncJob applies a handler outcome. Succeeded -> succeeded;
	// retryable -> queued (or dead at max attempts) with available_at set to the
	// supplied backoff target; non-retryable -> failed.
	FinalizeAsyncJob(ctx context.Context, jobID string, outcome JobOutcome, availableAt time.Time, now time.Time) error
	// ReclaimExpiredLeases requeues running rows whose locked_at is older than
	// olderThan. attempts is NOT incremented (lease timeout, not a business
	// failure). Returns the number of reclaimed rows.
	ReclaimExpiredLeases(ctx context.Context, jobTypes []string, olderThan time.Time, now time.Time) (int64, error)
}

// SQLStore is the Postgres-backed LeaseStore. Column names follow the B4
// baseline (locked_at / attempts / max_attempts / available_at / status) and
// the lease/reaper queries match spec §4.2 exactly.
type SQLStore struct {
	db *sql.DB
}

// NewSQLStore wires a SQLStore against db.
func NewSQLStore(db *sql.DB) *SQLStore { return &SQLStore{db: db} }

func (s *SQLStore) checkDB() error {
	if s == nil || s.db == nil {
		return fmt.Errorf("runner store db is nil")
	}
	return nil
}

// LeaseAsyncJob implements the fixed spec §4.2 claim query.
func (s *SQLStore) LeaseAsyncJob(ctx context.Context, jobTypes []string, now time.Time) (ClaimedJob, bool, error) {
	if err := s.checkDB(); err != nil {
		return ClaimedJob{}, false, err
	}
	if len(jobTypes) == 0 {
		return ClaimedJob{}, false, fmt.Errorf("LeaseAsyncJob requires at least one job type")
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
	var (
		claimed ClaimedJob
		payload []byte
	)
	err := s.db.QueryRowContext(ctx, query, args...).Scan(
		&claimed.JobID,
		&claimed.JobType,
		&claimed.ResourceType,
		&claimed.ResourceID,
		&payload,
		&claimed.Attempts,
		&claimed.MaxAttempts,
		&claimed.AvailableAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return ClaimedJob{}, false, nil
	}
	if err != nil {
		return ClaimedJob{}, false, fmt.Errorf("lease async_jobs: %w", err)
	}
	if len(payload) > 0 {
		claimed.Payload = append([]byte{}, payload...)
	}
	return claimed, true, nil
}

// FinalizeAsyncJob persists a handler outcome. The retryable terminal split
// (dead at attempts >= max_attempts, queued otherwise) is enforced in SQL so it
// stays atomic with the row read.
func (s *SQLStore) FinalizeAsyncJob(ctx context.Context, jobID string, outcome JobOutcome, availableAt time.Time, now time.Time) error {
	if err := s.checkDB(); err != nil {
		return err
	}
	if jobID == "" {
		return fmt.Errorf("FinalizeAsyncJob requires jobID")
	}
	if outcome.Succeeded {
		_, err := s.db.ExecContext(ctx, `
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
		_, err := s.db.ExecContext(ctx, `
update async_jobs
set status = case when attempts >= max_attempts then 'dead' else 'queued' end,
    completed_at = case when attempts >= max_attempts then $1::timestamptz else null end,
    available_at = $2,
    updated_at = $1,
    locked_at = null,
    error_code = $3,
    error_message = $4
where id = $5`,
			now,
			availableAt,
			outcome.ErrorCode,
			nullableString(outcome.ErrorMessage),
			jobID,
		)
		if err != nil {
			return fmt.Errorf("finalize async_jobs retryable: %w", err)
		}
		return nil
	}
	_, err := s.db.ExecContext(ctx, `
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

// ReclaimExpiredLeases implements the fixed spec §4.2 reaper query. attempts is
// left unchanged; available_at is advanced to now so the row is immediately
// re-claimable after a crash.
func (s *SQLStore) ReclaimExpiredLeases(ctx context.Context, jobTypes []string, olderThan time.Time, now time.Time) (int64, error) {
	if err := s.checkDB(); err != nil {
		return 0, err
	}
	if len(jobTypes) == 0 {
		return 0, fmt.Errorf("ReclaimExpiredLeases requires at least one job type")
	}
	placeholders := make([]string, 0, len(jobTypes))
	args := make([]any, 0, len(jobTypes)+2)
	args = append(args, olderThan, now)
	for i, jt := range jobTypes {
		placeholders = append(placeholders, fmt.Sprintf("$%d", i+3))
		args = append(args, jt)
	}
	query := fmt.Sprintf(`
update async_jobs
set status = 'queued',
    locked_at = null,
    available_at = $2,
    updated_at = $2
where status = 'running'
  and locked_at is not null
  and locked_at <= $1
  and job_type in (%s)`, strings.Join(placeholders, ","))
	res, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("reclaim expired async_jobs leases: %w", err)
	}
	count, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("reclaim expired async_jobs leases rows affected: %w", err)
	}
	return count, nil
}

func nullableString(s string) any {
	if s == "" {
		return nil
	}
	return s
}
