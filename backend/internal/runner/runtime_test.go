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
