package runner

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

func testConfig() Config {
	return Config{
		ScanInterval:   10 * time.Millisecond,
		LeaseTimeout:   5 * time.Minute,
		ReaperInterval: 0, // disabled by default in tests
		ShutdownGrace:  2 * time.Second,
		QueueWeights:   QueueWeights{Critical: 6, Default: 3, Low: 1},
	}
}

func fixedClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestRuntime_RegisterAndRunOnce(t *testing.T) {
	now := time.Date(2026, 5, 22, 12, 0, 0, 0, time.UTC)
	store := newFakeStore()
	store.enqueue("job-1", "target_import", 0, now.Add(-time.Minute), now.Add(-time.Minute))

	rt := New(Options{Store: store, Config: testConfig(), Now: fixedClock(now)})

	var handled ClaimedJob
	rt.Register("target_import", JobHandlerFunc(func(_ context.Context, job ClaimedJob) JobOutcome {
		handled = job
		return JobOutcome{Succeeded: true}
	}))

	if !rt.Handles("target_import") {
		t.Fatalf("expected runtime to report it handles target_import")
	}

	processed, err := rt.RunOnce(context.Background())
	if err != nil {
		t.Fatalf("RunOnce: %v", err)
	}
	if !processed {
		t.Fatalf("expected RunOnce to process a job")
	}
	if handled.JobID != "job-1" {
		t.Fatalf("handler saw job %q, want job-1", handled.JobID)
	}
	if handled.Attempts != 1 {
		t.Fatalf("claimed attempts = %d, want 1 (incremented on claim)", handled.Attempts)
	}
	if row := store.get("job-1"); row == nil || row.status != "succeeded" {
		t.Fatalf("row status = %v, want succeeded", row)
	}

	// No more queued rows -> RunOnce reports no work.
	processed, err = rt.RunOnce(context.Background())
	if err != nil {
		t.Fatalf("RunOnce (empty): %v", err)
	}
	if processed {
		t.Fatalf("expected no job to process on empty queue")
	}
}

func TestRuntime_FinalizeUsesTimestampAfterHandlerReturns(t *testing.T) {
	base := time.Date(2026, 5, 22, 12, 0, 0, 0, time.UTC)
	finished := base.Add(45 * time.Second)
	clock := &mutableClock{now: base}
	store := newFakeStore()
	store.enqueue("job-retry", "resume_parse", 0, base.Add(-time.Minute), base.Add(-time.Minute))

	rt := New(Options{Store: store, Config: testConfig(), Now: clock.Now})
	rt.Register("resume_parse", JobHandlerFunc(func(_ context.Context, _ ClaimedJob) JobOutcome {
		clock.Set(finished)
		return JobOutcome{Retryable: true, ErrorCode: "AI_PROVIDER_TIMEOUT"}
	}))

	if processed, err := rt.RunOnce(context.Background()); err != nil || !processed {
		t.Fatalf("RunOnce retryable processed=%v err=%v", processed, err)
	}
	row := store.get("job-retry")
	wantAvailable := finished.Add(DefaultBackoffPolicy().Next(1))
	if !row.availableAt.Equal(wantAvailable) {
		t.Fatalf("retry availableAt = %s, want %s", row.availableAt, wantAvailable)
	}

	clock.Set(base)
	store = newFakeStore()
	store.enqueue("job-failed", "resume_parse", 0, base.Add(-time.Minute), base.Add(-time.Minute))
	rt = New(Options{Store: store, Config: testConfig(), Now: clock.Now})
	rt.Register("resume_parse", JobHandlerFunc(func(_ context.Context, _ ClaimedJob) JobOutcome {
		clock.Set(finished)
		return JobOutcome{ErrorCode: "AI_OUTPUT_INVALID"}
	}))

	if processed, err := rt.RunOnce(context.Background()); err != nil || !processed {
		t.Fatalf("RunOnce failed processed=%v err=%v", processed, err)
	}
	row = store.get("job-failed")
	if row.completedAt == nil || !row.completedAt.Equal(finished) {
		t.Fatalf("failed completedAt = %v, want %s", row.completedAt, finished)
	}
}

func TestRuntime_StartDoesNotLetCriticalJobStarveEmailDispatch(t *testing.T) {
	now := time.Date(2026, 5, 22, 12, 0, 0, 0, time.UTC)
	store := newFakeStore()
	store.enqueue("job-report", "report_generate", 0, now.Add(-time.Minute), now.Add(-time.Minute))
	store.enqueue("job-email", "email_dispatch", 0, now.Add(-time.Minute), now.Add(-time.Minute))

	rt := New(Options{Store: store, Config: testConfig(), Now: fixedClock(now)})

	reportStarted := make(chan struct{})
	releaseReport := make(chan struct{})
	var reportStartedOnce sync.Once
	var releaseOnce sync.Once
	rt.Register("report_generate", JobHandlerFunc(func(_ context.Context, _ ClaimedJob) JobOutcome {
		reportStartedOnce.Do(func() { close(reportStarted) })
		<-releaseReport
		return JobOutcome{Succeeded: true}
	}))

	emailHandled := make(chan struct{})
	var emailHandledOnce sync.Once
	rt.Register("email_dispatch", JobHandlerFunc(func(_ context.Context, _ ClaimedJob) JobOutcome {
		emailHandledOnce.Do(func() { close(emailHandled) })
		return JobOutcome{Succeeded: true}
	}))

	rt.Start(context.Background())
	defer func() {
		releaseOnce.Do(func() { close(releaseReport) })
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := rt.Shutdown(shutdownCtx); err != nil {
			t.Fatalf("Shutdown: %v", err)
		}
	}()

	select {
	case <-reportStarted:
	case <-time.After(time.Second):
		t.Fatalf("critical report job did not start")
	}
	select {
	case <-emailHandled:
	case <-time.After(200 * time.Millisecond):
		t.Fatalf("email_dispatch was starved behind a running critical job")
	}
	releaseOnce.Do(func() { close(releaseReport) })
}

func TestRuntime_GracefulShutdown(t *testing.T) {
	now := time.Date(2026, 5, 22, 12, 0, 0, 0, time.UTC)
	store := newFakeStore()
	store.enqueue("job-1", "target_import", 0, now.Add(-time.Minute), now.Add(-time.Minute))

	rt := New(Options{Store: store, Config: testConfig(), Now: fixedClock(now)})

	release := make(chan struct{})
	entered := make(chan struct{})
	var once sync.Once
	rt.Register("target_import", JobHandlerFunc(func(_ context.Context, _ ClaimedJob) JobOutcome {
		once.Do(func() { close(entered) })
		<-release
		return JobOutcome{Succeeded: true}
	}))

	rt.Start(context.Background())
	select {
	case <-entered:
	case <-time.After(time.Second):
		t.Fatalf("handler did not start within 1s")
	}

	// Let the in-flight handler finish during the grace window.
	close(release)
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := rt.Shutdown(shutdownCtx); err != nil {
		t.Fatalf("Shutdown: %v", err)
	}
	if row := store.get("job-1"); row == nil || row.status != "succeeded" {
		t.Fatalf("row status = %v, want succeeded after graceful drain", row)
	}
}

func TestRuntime_ShutdownTimeoutPropagates(t *testing.T) {
	now := time.Date(2026, 5, 22, 12, 0, 0, 0, time.UTC)
	store := newFakeStore()
	store.enqueue("job-1", "target_import", 0, now.Add(-time.Minute), now.Add(-time.Minute))

	rt := New(Options{Store: store, Config: testConfig(), Now: fixedClock(now)})

	release := make(chan struct{})
	entered := make(chan struct{})
	var once sync.Once
	rt.Register("target_import", JobHandlerFunc(func(_ context.Context, _ ClaimedJob) JobOutcome {
		once.Do(func() { close(entered) })
		<-release // ignore ctx: simulate a stuck handler
		return JobOutcome{Succeeded: true}
	}))

	rt.Start(context.Background())
	select {
	case <-entered:
	case <-time.After(time.Second):
		t.Fatalf("handler did not start within 1s")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	err := rt.Shutdown(shutdownCtx)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Shutdown err = %v, want context.DeadlineExceeded", err)
	}
	close(release) // unblock the goroutine for clean teardown
}
