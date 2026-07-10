package review

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/runner"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

type fakeGenStatusStore struct {
	update ReportStatusUpdate
	err    error
}

func (s *fakeGenStatusStore) UpdateFeedbackReportStatus(_ context.Context, update ReportStatusUpdate) error {
	s.update = update
	return s.err
}

type fakeGenReportService struct {
	outcome ReportOutcome
	sawJob  AsyncJob
}

func (s *fakeGenReportService) GenerateReport(_ context.Context, job AsyncJob) ReportOutcome {
	s.sawJob = job
	return s.outcome
}

func TestGenerateHandler_TransitionsThenGeneratesAndPropagatesOutcome(t *testing.T) {
	store := &fakeGenStatusStore{}
	svc := &fakeGenReportService{outcome: ReportOutcome{Succeeded: true, AsyncJobFinalized: true}}
	h := NewGenerateHandler(GenerateHandlerOptions{Store: store, Service: svc, Now: func() time.Time { return time.Unix(0, 0) }})

	out := h.Handle(context.Background(), runner.ClaimedJob{
		JobID:       "job-1",
		JobType:     ReportGenerateJobType,
		ResourceID:  "report-1",
		Attempts:    1,
		MaxAttempts: 5,
	})
	if store.update.From != sharedtypes.ReportStatusQueued || store.update.To != sharedtypes.ReportStatusGenerating {
		t.Fatalf("status transition = %+v, want queued->generating", store.update)
	}
	if store.update.ReportID != "report-1" {
		t.Fatalf("status transition report id = %q", store.update.ReportID)
	}
	if svc.sawJob.JobID != "job-1" || svc.sawJob.ResourceID != "report-1" {
		t.Fatalf("service saw job %+v", svc.sawJob)
	}
	if !out.Succeeded || !out.AsyncJobFinalized {
		t.Fatalf("outcome = %+v, want succeeded + finalized", out)
	}
}

func TestGenerateHandler_RequeuesWhenStatusTransitionFails(t *testing.T) {
	store := &fakeGenStatusStore{err: errors.New("illegal transition")}
	svc := &fakeGenReportService{outcome: ReportOutcome{Succeeded: true}}
	h := NewGenerateHandler(GenerateHandlerOptions{Store: store, Service: svc})

	out := h.Handle(context.Background(), runner.ClaimedJob{JobID: "job-1", ResourceID: "report-1", Attempts: 1, MaxAttempts: 5})
	if out.Succeeded {
		t.Fatalf("outcome should not be succeeded when transition fails: %+v", out)
	}
	if !out.Retryable {
		t.Fatalf("transition failure must be retryable: %+v", out)
	}
	if svc.sawJob.JobID != "" {
		t.Fatalf("service must not run when status transition fails")
	}
}

func TestGenerateHandler_PropagatesRetryableFailure(t *testing.T) {
	assertGenerateHandlerRetryableFailure(t, false)
}

func TestGenerateHandler_NormalizesFinalizedRetryableFailureThroughKernel(t *testing.T) {
	assertGenerateHandlerRetryableFailure(t, true)
}

func assertGenerateHandlerRetryableFailure(t *testing.T, serviceFinalized bool) {
	t.Helper()
	store := &fakeGenStatusStore{}
	svc := &fakeGenReportService{outcome: ReportOutcome{
		Retryable:         true,
		ErrorCode:         "AI_PROVIDER_TIMEOUT",
		ErrorMessage:      "timeout",
		AsyncJobFinalized: serviceFinalized,
	}}
	h := NewGenerateHandler(GenerateHandlerOptions{Store: store, Service: svc})

	out := h.Handle(context.Background(), runner.ClaimedJob{JobID: "job-1", ResourceID: "report-1", Attempts: 3, MaxAttempts: 5})
	if out.Succeeded || !out.Retryable || out.ErrorCode != "AI_PROVIDER_TIMEOUT" {
		t.Fatalf("outcome = %+v, want retryable timeout", out)
	}
	if out.AsyncJobFinalized {
		t.Fatalf("retryable report failures must be finalized by the runner kernel, serviceFinalized=%v", serviceFinalized)
	}
}
