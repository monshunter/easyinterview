package auth

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/runner"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
	"github.com/monshunter/easyinterview/backend/internal/shared/jobs"
)

// asyncJobExecer is the minimal DB surface the email_dispatch enqueuer needs.
type asyncJobExecer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

// EmailDispatchEnqueuer is the C1 producer (spec D-10): it uses the runner
// kernel by inserting an
// async_jobs(job_type=email_dispatch) row that the runner kernel leases. It
// satisfies the MailDispatcher interface so the email-code service enqueue
// path is unchanged.
type EmailDispatchEnqueuer struct {
	db    asyncJobExecer
	newID func() string
	now   func() time.Time
}

// NewEmailDispatchEnqueuer wires an enqueuer.
func NewEmailDispatchEnqueuer(db asyncJobExecer, newID func() string, now func() time.Time) *EmailDispatchEnqueuer {
	if newID == nil {
		newID = func() string { return "" }
	}
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	return &EmailDispatchEnqueuer{db: db, newID: newID, now: now}
}

// Enqueue inserts an email_dispatch async_jobs row. The challenge id is the
// resource_id so privacy delete / job lookup can join on it.
func (e *EmailDispatchEnqueuer) Enqueue(ctx context.Context, payload jobs.EmailDispatchPayload) error {
	if e == nil || e.db == nil {
		return fmt.Errorf("email dispatch enqueuer db is nil")
	}
	challengeID := payload["authChallengeId"]
	if challengeID == "" {
		return fmt.Errorf("%s payload missing authChallengeId", jobs.JobTypeEmailDispatch)
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal %s payload: %w", jobs.JobTypeEmailDispatch, err)
	}
	now := e.now().UTC()
	if _, err := e.db.ExecContext(ctx, `
insert into async_jobs (
  id, job_type, resource_type, resource_id, dedupe_key, status, payload,
  available_at, created_at, updated_at
) values ($1, $2, 'auth_challenge', $3, null, 'queued', $4::jsonb, $5, $5, $5)`,
		e.newID(), string(jobs.JobTypeEmailDispatch), challengeID, string(raw), now); err != nil {
		return fmt.Errorf("enqueue %s async job: %w", jobs.JobTypeEmailDispatch, err)
	}
	return nil
}

var _ MailDispatcher = (*EmailDispatchEnqueuer)(nil)

// EmailDispatchHandler is the kernel runner.Handler for email_dispatch jobs
// (spec D-10). It revalidates the payload against the B3 redaction red line and
// delivers it through the existing DeliveryWriter sink.
type EmailDispatchHandler struct {
	writer DeliveryWriter
}

// NewEmailDispatchHandler wires the handler.
func NewEmailDispatchHandler(writer DeliveryWriter) *EmailDispatchHandler {
	return &EmailDispatchHandler{writer: writer}
}

// Handle satisfies runner.Handler.
func (h *EmailDispatchHandler) Handle(_ context.Context, job runner.ClaimedJob) runner.JobOutcome {
	raw := map[string]string{}
	if len(job.Payload) > 0 {
		if err := json.Unmarshal(job.Payload, &raw); err != nil {
			return runner.JobOutcome{ErrorCode: sharederrors.CodeValidationFailed, ErrorMessage: fmt.Sprintf("%s payload is invalid JSON", jobs.JobTypeEmailDispatch)}
		}
	}
	payload, err := jobs.BuildEmailDispatchPayload(raw)
	if err != nil {
		// Forbidden / unknown fields are a permanent payload contract failure.
		return runner.JobOutcome{ErrorCode: sharederrors.CodeValidationFailed, ErrorMessage: err.Error()}
	}
	if h.writer == nil {
		return runner.JobOutcome{Retryable: true, ErrorCode: "EMAIL_DISPATCH_FAILED", ErrorMessage: "delivery writer unavailable"}
	}
	if err := h.writer.Write(payload); err != nil {
		message := "email delivery failed"
		if safe, ok := err.(interface{ SafeMessage() string }); ok {
			message = safe.SafeMessage()
		}
		return runner.JobOutcome{Retryable: true, ErrorCode: "EMAIL_DISPATCH_FAILED", ErrorMessage: message}
	}
	return runner.JobOutcome{Succeeded: true}
}

var _ runner.Handler = (*EmailDispatchHandler)(nil)
