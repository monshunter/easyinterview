package review

import (
	"context"
	"fmt"
	"math"
	"time"

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

	if _, err := tx.ExecContext(ctx, `
update feedback_reports
set status = 'failed',
    error_code = $1,
    generated_at = $2,
    updated_at = $2
where id = $3`, in.ErrorCode, in.Now, in.ReportID); err != nil {
		return fmt.Errorf("update feedback_reports failed: %w", err)
	}
	var attempts, maxAttempts int
	if err := tx.QueryRowContext(ctx, `select attempts, max_attempts from async_jobs where id=$1 for update`, in.AsyncJobID).Scan(&attempts, &maxAttempts); err != nil {
		return fmt.Errorf("select async_jobs attempts: %w", err)
	}
	requeue := in.Retryable && attempts < maxAttempts
	availableAt := in.Now.Add(computeBackoff(attempts))
	var completedAt any = in.Now
	if requeue {
		completedAt = nil
	}
	status := "failed"
	if requeue {
		status = "queued"
	}
	if _, err := tx.ExecContext(ctx, `
update async_jobs
set status = $1,
    completed_at = $2,
    available_at = $3,
    updated_at = $4,
    locked_at = null,
    error_code = $5,
    error_message = $6
where id = $7`,
		status,
		completedAt,
		availableAt,
		in.Now,
		in.ErrorCode,
		nullableString(in.ErrorCode),
		in.AsyncJobID,
	); err != nil {
		return fmt.Errorf("update async_jobs failure: %w", err)
	}
	payload, err := BuildReportGenerationFailedPayload(ReportGenerationFailedInput{
		ReportID:  in.ReportID,
		SessionID: in.SessionID,
		ErrorCode: in.ErrorCode,
		Retryable: requeue,
	})
	if err != nil {
		return err
	}
	if err := insertReviewOutbox(ctx, tx, in.OutboxEventID, string(sharedevents.EventNameReportGenerationFailed), in.ReportID, payload, in.Now); err != nil {
		return err
	}
	if err := insertReviewAudit(ctx, tx, in.AuditEventID, in.UserID, "feedback_report.generation_failed", in.ReportID, "failure", map[string]any{"errorCode": in.ErrorCode, "retryable": requeue}, in.Now); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit persist report failure: %w", err)
	}
	return nil
}

func computeBackoff(attempts int) time.Duration {
	if attempts < 0 {
		attempts = 0
	}
	pow := math.Pow(2, float64(attempts))
	delay := time.Duration(pow) * 30 * time.Second
	if delay > 30*time.Minute {
		return 30 * time.Minute
	}
	return delay
}
