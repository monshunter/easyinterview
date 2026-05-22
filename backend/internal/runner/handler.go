package runner

import (
	"context"
	"time"
)

// ClaimedJob is the structured handoff between the kernel lease query and a
// registered Handler. It mirrors the async_jobs row columns the runtime needs
// for retry / completion bookkeeping; handlers treat it as read-only.
type ClaimedJob struct {
	JobID        string
	JobType      string
	ResourceType string
	ResourceID   string
	Payload      []byte
	Attempts     int32
	MaxAttempts  int32
	AvailableAt  time.Time
}

// JobOutcome captures a handler result so the runtime can finalize the
// async_jobs row consistently (spec §4.1 / D-4).
//
// Succeeded marks the row succeeded. When Succeeded is false, Retryable
// distinguishes a transient failure (requeue with backoff, or dead-letter at
// max attempts) from a permanent failure (status='failed'). ErrorCode /
// ErrorMessage are persisted as the redacted finalize reason. AsyncJobFinalized
// is the migration-only escape hatch: handlers that already finalized the
// async_jobs row inside their own transaction set it true so the runtime skips
// its finalize step (removable once every handler delegates finalize to the
// kernel).
type JobOutcome struct {
	Succeeded         bool
	Retryable         bool
	ErrorCode         string
	ErrorMessage      string
	AsyncJobFinalized bool
}

// Handler is the contract every registered job_type implements. The runtime
// owns claim / finalize bookkeeping; handlers focus on per-job side effects and
// translate failures into JobOutcome values.
type Handler interface {
	Handle(ctx context.Context, job ClaimedJob) JobOutcome
}

// JobHandlerFunc adapts an inline function to the Handler interface.
type JobHandlerFunc func(ctx context.Context, job ClaimedJob) JobOutcome

// Handle satisfies Handler.
func (f JobHandlerFunc) Handle(ctx context.Context, job ClaimedJob) JobOutcome {
	return f(ctx, job)
}
