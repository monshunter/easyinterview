package runner

import (
	"context"

	"github.com/monshunter/easyinterview/backend/internal/targetjob"
)

// FromTargetjobHandler adapts a targetjob.JobHandler to the kernel
// Handler interface (spec D-2). It lets Phase 2 register existing domain
// handlers against the kernel without rewriting their bodies; the shim is
// removable once every handler implements runner.Handler natively.
func FromTargetjobHandler(h targetjob.JobHandler) Handler {
	return JobHandlerFunc(func(ctx context.Context, job ClaimedJob) JobOutcome {
		outcome := h.Handle(ctx, toTargetjobClaimedJob(job))
		return JobOutcome{
			Succeeded:    outcome.Succeeded,
			Retryable:    outcome.Retryable,
			ErrorCode:    outcome.ErrorCode,
			ErrorMessage: outcome.ErrorMessage,
		}
	})
}

func toTargetjobClaimedJob(job ClaimedJob) targetjob.ClaimedJob {
	return targetjob.ClaimedJob{
		JobID:        job.JobID,
		JobType:      job.JobType,
		ResourceType: job.ResourceType,
		ResourceID:   job.ResourceID,
		Payload:      job.Payload,
		Attempts:     job.Attempts,
		MaxAttempts:  job.MaxAttempts,
		AvailableAt:  job.AvailableAt,
	}
}
