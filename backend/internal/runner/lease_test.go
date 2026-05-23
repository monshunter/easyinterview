package runner

import (
	"context"
	"testing"
	"time"
)

// TestLeaseAsyncJob_RespectsPriorityBuckets verifies the runtime leases from a
// higher-priority bucket before a lower one when both have an available row
// with the same available_at (spec D-9).
func TestLeaseAsyncJob_RespectsPriorityBuckets(t *testing.T) {
	now := time.Date(2026, 5, 22, 12, 0, 0, 0, time.UTC)
	store := newFakeStore()
	// low-priority and critical-priority rows, identical timing.
	store.enqueue("low-1", "email_dispatch", 0, now.Add(-time.Minute), now.Add(-time.Minute))
	store.enqueue("crit-1", "report_generate", 0, now.Add(-time.Minute), now.Add(-time.Minute))

	rt := New(Options{Store: store, Config: testConfig(), Now: fixedClock(now)})
	var order []string
	rt.Register("email_dispatch", JobHandlerFunc(func(_ context.Context, job ClaimedJob) JobOutcome {
		order = append(order, job.JobType)
		return JobOutcome{Succeeded: true}
	}))
	rt.Register("report_generate", JobHandlerFunc(func(_ context.Context, job ClaimedJob) JobOutcome {
		order = append(order, job.JobType)
		return JobOutcome{Succeeded: true}
	}))

	for {
		processed, err := rt.RunOnce(context.Background())
		if err != nil {
			t.Fatalf("RunOnce: %v", err)
		}
		if !processed {
			break
		}
	}
	if len(order) != 2 || order[0] != "report_generate" {
		t.Fatalf("processing order = %v, want critical (report_generate) first", order)
	}
}

func TestFinalizeAsyncJob_PermanentFailureAtMax(t *testing.T) {
	now := time.Date(2026, 5, 22, 12, 0, 0, 0, time.UTC)
	store := newFakeStore()
	// attempts already at max-1; the claim bumps it to max, so a retryable
	// failure must finalize as dead, not requeued.
	store.enqueue("job-dead", "target_import", MaxAttempts-1, now.Add(-time.Minute), now.Add(-time.Minute))

	rt := New(Options{Store: store, Config: testConfig(), Now: fixedClock(now)})
	rt.Register("target_import", JobHandlerFunc(func(_ context.Context, _ ClaimedJob) JobOutcome {
		return JobOutcome{Retryable: true, ErrorCode: "TRANSIENT", ErrorMessage: "boom"}
	}))

	if _, err := rt.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce: %v", err)
	}
	row := store.get("job-dead")
	if row == nil || row.status != "dead" {
		t.Fatalf("row status = %v, want dead at max attempts", row)
	}
	if row.attempts != MaxAttempts {
		t.Fatalf("attempts = %d, want %d", row.attempts, MaxAttempts)
	}
}

func TestFinalizeAsyncJob_NonRetryableFailure(t *testing.T) {
	now := time.Date(2026, 5, 22, 12, 0, 0, 0, time.UTC)
	store := newFakeStore()
	store.enqueue("job-fail", "target_import", 0, now.Add(-time.Minute), now.Add(-time.Minute))

	rt := New(Options{Store: store, Config: testConfig(), Now: fixedClock(now)})
	rt.Register("target_import", JobHandlerFunc(func(_ context.Context, _ ClaimedJob) JobOutcome {
		return JobOutcome{Retryable: false, ErrorCode: "PERMANENT", ErrorMessage: "nope"}
	}))

	if _, err := rt.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce: %v", err)
	}
	row := store.get("job-fail")
	if row == nil || row.status != "failed" {
		t.Fatalf("row status = %v, want failed for non-retryable", row)
	}
}

func TestFinalizeAsyncJob_RetryableRequeuesWithBackoff(t *testing.T) {
	now := time.Date(2026, 5, 22, 12, 0, 0, 0, time.UTC)
	store := newFakeStore()
	store.enqueue("job-retry", "target_import", 0, now.Add(-time.Minute), now.Add(-time.Minute))

	rt := New(Options{Store: store, Config: testConfig(), Now: fixedClock(now)})
	rt.Register("target_import", JobHandlerFunc(func(_ context.Context, _ ClaimedJob) JobOutcome {
		return JobOutcome{Retryable: true, ErrorCode: "TRANSIENT"}
	}))

	if _, err := rt.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce: %v", err)
	}
	row := store.get("job-retry")
	if row == nil || row.status != "queued" {
		t.Fatalf("row status = %v, want queued for retryable below max", row)
	}
	// attempts == 1 after claim -> backoff Next(1) == 30s.
	wantAvailable := now.Add(30 * time.Second)
	if !row.availableAt.Equal(wantAvailable) {
		t.Fatalf("availableAt = %s, want %s (30s backoff)", row.availableAt, wantAvailable)
	}
}
