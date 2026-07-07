package review

import (
	"context"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/runner"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

// ReportStatusStore is the persistence subset GenerateHandler needs to move a
// feedback_report from queued to generating before the AI run. Lease / finalize
// of the async_jobs row is owned by the runner kernel (spec D-3), not here.
type ReportStatusStore interface {
	UpdateFeedbackReportStatus(ctx context.Context, update ReportStatusUpdate) error
}

// GenerateHandlerOptions wires a GenerateHandler.
type GenerateHandlerOptions struct {
	Store   ReportStatusStore
	Service ReportService
	Now     func() time.Time
}

// GenerateHandler is the report_generate runner.Handler. The kernel leases the
// async_jobs row, this handler flips the report to generating and runs the AI
// report service. Success may be finalized by the report service transaction;
// failures are always normalized back through the kernel so the shared
// BackoffPolicy owns retry/dead-letter.
type GenerateHandler struct {
	store   ReportStatusStore
	service ReportService
	now     func() time.Time
}

// NewGenerateHandler constructs a GenerateHandler.
func NewGenerateHandler(opts GenerateHandlerOptions) *GenerateHandler {
	now := opts.Now
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	return &GenerateHandler{store: opts.Store, service: opts.Service, now: now}
}

// Handle satisfies runner.Handler.
func (h *GenerateHandler) Handle(ctx context.Context, job runner.ClaimedJob) runner.JobOutcome {
	if h == nil || h.service == nil {
		return runner.JobOutcome{Retryable: true, ErrorCode: sharederrors.CodeAiOutputInvalid, ErrorMessage: "review generate handler is not configured"}
	}
	now := h.now()
	if h.store != nil {
		if err := h.store.UpdateFeedbackReportStatus(ctx, ReportStatusUpdate{
			ReportID: job.ResourceID,
			From:     sharedtypes.ReportStatusQueued,
			To:       sharedtypes.ReportStatusGenerating,
			Now:      now,
		}); err != nil {
			// The report row is not in a leaseable state (e.g. already
			// generating after a reaped lease). Requeue with backoff rather than
			// finalizing so the kernel reaper / retry path can recover.
			return runner.JobOutcome{Retryable: true, ErrorCode: sharederrors.CodeValidationFailed, ErrorMessage: err.Error()}
		}
	}
	outcome := h.service.GenerateReport(ctx, AsyncJob{
		JobID:        job.JobID,
		JobType:      job.JobType,
		ResourceType: job.ResourceType,
		ResourceID:   job.ResourceID,
		Payload:      job.Payload,
		Attempts:     job.Attempts,
		MaxAttempts:  job.MaxAttempts,
		AvailableAt:  job.AvailableAt,
	})
	return runner.JobOutcome{
		Succeeded:         outcome.Succeeded,
		Retryable:         outcome.Retryable,
		ErrorCode:         outcome.ErrorCode,
		ErrorMessage:      outcome.ErrorMessage,
		AsyncJobFinalized: outcome.Succeeded && outcome.AsyncJobFinalized,
	}
}
