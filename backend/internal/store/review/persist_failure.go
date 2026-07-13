package review

import (
	"context"
	"fmt"

	reviewdomain "github.com/monshunter/easyinterview/backend/internal/review"
	sharedevents "github.com/monshunter/easyinterview/backend/internal/shared/events"
)

type PersistReportFailureInput = reviewdomain.ReportFailurePersistence

func (r *Repository) PersistReportFailure(ctx context.Context, in PersistReportFailureInput) error {
	if err := r.checkDB(); err != nil {
		return err
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin persist report failure: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	if err := lockCurrentAsyncJobLease(ctx, tx, in.AsyncJobID, in.ClaimedAttempts); err != nil {
		return err
	}
	// A report handler can spend most of its lease window in provider calls.
	// Renew the still-current generation before committing the failure-side
	// effects so the kernel gets a complete lease window to finalize it.
	if err := renewCurrentAsyncJobLease(ctx, tx, in.AsyncJobID, in.ClaimedAttempts, in.Now); err != nil {
		return err
	}

	willRetry := in.Retryable && in.ClaimedAttempts < in.MaxAttempts
	reportStatus := "failed"
	if willRetry {
		reportStatus = "queued"
	}
	res, err := tx.ExecContext(ctx, `
update feedback_reports
set status = $1,
    error_code = $2,
    generated_at = case when $1 = 'failed' then $3::timestamptz else null end,
    updated_at = $3
where id = $4 and status = 'generating'`, reportStatus, in.ErrorCode, in.Now, in.ReportID)
	if err != nil {
		return fmt.Errorf("update feedback_reports failed: %w", err)
	}
	if err := requireOneRow(res, "update feedback_reports failed"); err != nil {
		return err
	}
	payload, err := BuildReportGenerationFailedPayload(ReportGenerationFailedInput{
		ReportID:  in.ReportID,
		SessionID: in.SessionID,
		ErrorCode: in.ErrorCode,
		Retryable: willRetry,
	})
	if err != nil {
		return err
	}
	if err := insertReviewOutbox(ctx, tx, in.OutboxEventID, string(sharedevents.EventNameReportGenerationFailed), in.ReportID, payload, in.Now); err != nil {
		return err
	}
	if err := insertReviewAudit(ctx, tx, in.AuditEventID, in.UserID, "feedback_report.generation_failed", in.ReportID, "failure", map[string]any{"errorCode": in.ErrorCode, "retryable": willRetry}, in.Now); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit persist report failure: %w", err)
	}
	return nil
}
