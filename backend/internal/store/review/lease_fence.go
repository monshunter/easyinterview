package review

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/runner"
)

func lockCurrentAsyncJobLease(ctx context.Context, tx *sql.Tx, jobID string, claimedAttempts int32) error {
	if jobID == "" {
		return fmt.Errorf("async job lease requires jobID")
	}
	if claimedAttempts < 1 {
		return fmt.Errorf("async job lease requires positive claimedAttempts")
	}
	var attempts int32
	err := tx.QueryRowContext(ctx, `
select attempts
from async_jobs
where id = $1
  and status = 'running'
  and attempts = $2
for update`, jobID, claimedAttempts).Scan(&attempts)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("%w: job_id=%s claimed_attempts=%d", runner.ErrStaleLease, jobID, claimedAttempts)
	}
	if err != nil {
		return fmt.Errorf("lock current async job lease: %w", err)
	}
	return nil
}

func renewCurrentAsyncJobLease(ctx context.Context, tx *sql.Tx, jobID string, claimedAttempts int32, now time.Time) error {
	res, err := tx.ExecContext(ctx, `
update async_jobs
set locked_at = $1,
    updated_at = $1
where id = $2
  and status = 'running'
  and attempts = $3`, now, jobID, claimedAttempts)
	if err != nil {
		return fmt.Errorf("renew current async job lease: %w", err)
	}
	return requireCurrentAsyncJobLeaseWrite(res, jobID, claimedAttempts)
}

func requireCurrentAsyncJobLeaseWrite(res sql.Result, jobID string, claimedAttempts int32) error {
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("update async job lease rows affected: %w", err)
	}
	if rows != 1 {
		return fmt.Errorf("%w: job_id=%s claimed_attempts=%d", runner.ErrStaleLease, jobID, claimedAttempts)
	}
	return nil
}
