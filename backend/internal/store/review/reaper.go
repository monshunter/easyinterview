package review

import (
	"context"
	"fmt"
	"time"
)

func (r *Repository) ReclaimExpiredLeases(ctx context.Context, jobType string, olderThan time.Time, now time.Time) (int64, error) {
	if err := r.checkDB(); err != nil {
		return 0, err
	}
	if jobType == "" {
		return 0, fmt.Errorf("ReclaimExpiredLeases requires jobType")
	}
	res, err := r.db.ExecContext(ctx, `
update async_jobs
set status = 'queued',
    locked_at = null,
    updated_at = $3
where job_type = $1
  and status = 'running'
  and locked_at is not null
  and locked_at < $2`, jobType, olderThan, now)
	if err != nil {
		return 0, fmt.Errorf("reclaim expired async_jobs leases: %w", err)
	}
	count, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("reclaim expired async_jobs leases rows affected: %w", err)
	}
	return count, nil
}
