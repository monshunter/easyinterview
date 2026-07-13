package review

import (
	"context"
	"fmt"
)

// AssertCurrentReportJobLease fences provider work to the currently running
// async-job lease generation. Product retry counts remain local to one
// GenerateReport invocation and are never written to feedback_reports.
func (r *Repository) AssertCurrentReportJobLease(ctx context.Context, jobID string, claimedAttempts int32) error {
	if err := r.checkDB(); err != nil {
		return err
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin assert feedback report job lease: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	if err := lockCurrentAsyncJobLease(ctx, tx, jobID, claimedAttempts); err != nil {
		return fmt.Errorf("assert feedback report job lease: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit feedback report job lease assertion: %w", err)
	}
	return nil
}
