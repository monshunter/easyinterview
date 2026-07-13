package review

import (
	"context"
	"fmt"

	reviewdomain "github.com/monshunter/easyinterview/backend/internal/review"
)

func (r *Repository) UpdateFeedbackReportStatus(ctx context.Context, update reviewdomain.ReportStatusUpdate) error {
	if err := r.checkDB(); err != nil {
		return err
	}
	if update.ReportID == "" {
		return fmt.Errorf("UpdateFeedbackReportStatus requires reportID")
	}
	if !reviewdomain.CanTransitionReportStatus(update.From, update.To) {
		return fmt.Errorf("%w: %s -> %s", reviewdomain.ErrIllegalTransition, update.From, update.To)
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin update feedback report status: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	if err := lockCurrentAsyncJobLease(ctx, tx, update.AsyncJobID, update.ClaimedAttempts); err != nil {
		return err
	}
	res, err := tx.ExecContext(ctx, `
update feedback_reports
set status = $1,
    updated_at = $2
where id = $3 and status in ($4, $1)`,
		string(update.To),
		update.Now,
		update.ReportID,
		string(update.From),
	)
	if err != nil {
		return fmt.Errorf("update feedback_reports status: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("update feedback_reports status rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("%w: stale status %s -> %s", reviewdomain.ErrIllegalTransition, update.From, update.To)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit update feedback report status: %w", err)
	}
	return nil
}
