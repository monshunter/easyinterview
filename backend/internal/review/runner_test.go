package review

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"sync"
	"testing"
	"time"

	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

func TestRunnerLeasesAndAdvancesToGenerating(t *testing.T) {
	now := time.Date(2026, 5, 15, 12, 0, 0, 0, time.UTC)
	store := &fakeRunnerStore{
		job: AsyncJob{
			JobID:      "0197d120-0000-7000-8000-000000000001",
			JobType:    "report_generate",
			ResourceID: "0197d120-0000-7000-8000-000000000002",
			LockedAt:   &now,
		},
		ok: true,
	}
	service := &fakeReportService{outcome: ReportOutcome{Succeeded: true}}
	runner := NewRunner(RunnerOptions{
		Store:        store,
		Service:      service,
		Now:          func() time.Time { return now },
		PollInterval: time.Millisecond,
		Logger:       slog.New(slog.NewTextHandler(io.Discard, nil)),
	})

	processed, err := runner.RunOnce(context.Background())
	if err != nil {
		t.Fatalf("RunOnce: %v", err)
	}
	if !processed {
		t.Fatal("RunOnce processed=false, want true")
	}
	if store.leasedJobType != "report_generate" {
		t.Fatalf("leased job type = %q", store.leasedJobType)
	}
	if store.statusUpdate.ReportID != store.job.ResourceID ||
		store.statusUpdate.From != sharedtypes.ReportStatusQueued ||
		store.statusUpdate.To != sharedtypes.ReportStatusGenerating {
		t.Fatalf("status update = %+v", store.statusUpdate)
	}
	if service.job.JobID != store.job.JobID || service.job.LockedAt == nil {
		t.Fatalf("service job = %+v", service.job)
	}
	if store.succeededJobID != store.job.JobID {
		t.Fatalf("succeeded job id = %q", store.succeededJobID)
	}
}

func TestRunnerSkipsAsyncJobUpdateWhenServiceFinalized(t *testing.T) {
	now := time.Date(2026, 5, 15, 12, 5, 0, 0, time.UTC)
	store := &fakeRunnerStore{
		job: AsyncJob{
			JobID:      "0197d120-0000-7000-8000-000000000011",
			JobType:    "report_generate",
			ResourceID: "0197d120-0000-7000-8000-000000000012",
		},
		ok: true,
	}
	runner := NewRunner(RunnerOptions{
		Store:        store,
		Service:      &fakeReportService{outcome: ReportOutcome{Succeeded: true, AsyncJobFinalized: true}},
		Now:          func() time.Time { return now },
		PollInterval: time.Millisecond,
		Logger:       slog.New(slog.NewTextHandler(io.Discard, nil)),
	})

	processed, err := runner.RunOnce(context.Background())
	if err != nil {
		t.Fatalf("RunOnce: %v", err)
	}
	if !processed {
		t.Fatal("RunOnce processed=false, want true")
	}
	if store.succeededJobID != "" || store.failed.JobID != "" {
		t.Fatalf("runner updated async job after service finalized: succeeded=%q failed=%+v", store.succeededJobID, store.failed)
	}
}

func TestRunnerContinuesPollingAfterLeaseError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	now := time.Date(2026, 5, 15, 12, 30, 0, 0, time.UTC)
	store := &fakeRunnerStore{
		leaseErrs: []error{errors.New("temporary lease error")},
		callCh:    make(chan int, 4),
	}
	runner := NewRunner(RunnerOptions{
		Store:        store,
		Service:      &fakeReportService{},
		Now:          func() time.Time { return now },
		PollInterval: time.Millisecond,
		Logger:       slog.New(slog.NewTextHandler(io.Discard, nil)),
	})

	runner.Start(ctx)
	defer func() {
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), time.Second)
		defer shutdownCancel()
		if err := runner.Stop(shutdownCtx); err != nil {
			t.Fatalf("Stop: %v", err)
		}
	}()

	deadline := time.After(time.Second)
	for {
		select {
		case call := <-store.callCh:
			if call >= 2 {
				return
			}
		case <-deadline:
			t.Fatal("runner did not continue polling after lease error")
		}
	}
}

func TestRunnerRetryPolicyAndPermanentFail(t *testing.T) {
	now := time.Date(2026, 5, 15, 21, 0, 0, 0, time.UTC)
	for _, tc := range []struct {
		attempts int32
		wantWait time.Duration
	}{
		{attempts: 1, wantWait: time.Minute},
		{attempts: 2, wantWait: 2 * time.Minute},
		{attempts: 3, wantWait: 4 * time.Minute},
		{attempts: 4, wantWait: 8 * time.Minute},
		{attempts: 5, wantWait: 16 * time.Minute},
	} {
		t.Run(fmt.Sprintf("attempts_%d", tc.attempts), func(t *testing.T) {
			store := &fakeRunnerStore{
				job: AsyncJob{
					JobID:       "0197d120-0000-7000-8000-000000000701",
					JobType:     "report_generate",
					ResourceID:  "0197d120-0000-7000-8000-000000000702",
					Attempts:    tc.attempts,
					MaxAttempts: 5,
				},
				ok: true,
			}
			runner := NewRunner(RunnerOptions{
				Store:        store,
				Service:      &fakeReportService{outcome: ReportOutcome{ErrorCode: "AI_PROVIDER_TIMEOUT", ErrorMessage: "timeout", Retryable: true}},
				Now:          func() time.Time { return now },
				PollInterval: time.Millisecond,
				Logger:       slog.New(slog.NewTextHandler(io.Discard, nil)),
			})
			if _, err := runner.RunOnce(context.Background()); err != nil {
				t.Fatalf("RunOnce: %v", err)
			}
			if store.failed.JobID != store.job.JobID || store.failed.AvailableAt.Sub(now) != tc.wantWait || !store.failed.Retryable {
				t.Fatalf("failed update = %+v, want wait %s", store.failed, tc.wantWait)
			}
		})
	}
}

type fakeRunnerStore struct {
	mu              sync.Mutex
	job             AsyncJob
	ok              bool
	leaseErrs       []error
	callCh          chan int
	calls           int
	leasedJobType   string
	statusUpdate    ReportStatusUpdate
	succeededJobID  string
	failed          AsyncJobFailure
	reaperJobType   string
	reaperOlderThan time.Time
	reaperNow       time.Time
	reaperCount     int64
}

func (s *fakeRunnerStore) LeaseAsyncJob(_ context.Context, jobType string, _ time.Time) (AsyncJob, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calls++
	if s.callCh != nil {
		select {
		case s.callCh <- s.calls:
		default:
		}
	}
	s.leasedJobType = jobType
	if len(s.leaseErrs) > 0 {
		err := s.leaseErrs[0]
		s.leaseErrs = s.leaseErrs[1:]
		return AsyncJob{}, false, err
	}
	return s.job, s.ok, nil
}

func (s *fakeRunnerStore) UpdateFeedbackReportStatus(_ context.Context, update ReportStatusUpdate) error {
	s.statusUpdate = update
	return nil
}

func (s *fakeRunnerStore) UpdateAsyncJobSucceeded(_ context.Context, jobID string, _ time.Time) error {
	s.succeededJobID = jobID
	return nil
}

func (s *fakeRunnerStore) UpdateAsyncJobFailed(_ context.Context, in AsyncJobFailure) error {
	s.failed = in
	return nil
}

func (s *fakeRunnerStore) ReclaimExpiredLeases(_ context.Context, jobType string, olderThan time.Time, now time.Time) (int64, error) {
	s.reaperJobType = jobType
	s.reaperOlderThan = olderThan
	s.reaperNow = now
	return s.reaperCount, nil
}

type fakeReportService struct {
	job     AsyncJob
	outcome ReportOutcome
}

func (s *fakeReportService) GenerateReport(_ context.Context, job AsyncJob) ReportOutcome {
	s.job = job
	return s.outcome
}
