package review

import (
	"context"
	"math"
	"time"
)

const DefaultReportFailureBackoff = 30 * time.Second

func ComputeReportFailureBackoff(attempts int32) time.Duration {
	if attempts < 0 {
		attempts = 0
	}
	delay := time.Duration(math.Pow(2, float64(attempts))) * DefaultReportFailureBackoff
	if delay > 30*time.Minute {
		return 30 * time.Minute
	}
	return delay
}

type Store interface {
	LeaseAsyncJob(ctx context.Context, jobType string, now time.Time) (AsyncJob, bool, error)
	UpdateFeedbackReportStatus(ctx context.Context, update ReportStatusUpdate) error
	UpdateAsyncJobSucceeded(ctx context.Context, jobID string, now time.Time) error
	UpdateAsyncJobFailed(ctx context.Context, in AsyncJobFailure) error
	ReclaimExpiredLeases(ctx context.Context, jobType string, olderThan time.Time, now time.Time) (int64, error)
}
