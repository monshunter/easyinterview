package runner

import (
	"context"
	"sync"
	"testing"
	"time"
)

type mutableClock struct {
	mu  sync.Mutex
	now time.Time
}

func (c *mutableClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *mutableClock) Set(t time.Time) {
	c.mu.Lock()
	c.now = t
	c.mu.Unlock()
}

// TestAllHandlersUseSharedBackoff proves that every retryable failure, no matter
// which domain handler produced it, is requeued by the kernel using the single
// BackoffPolicy table and dead-lettered at MaxAttempts. Because the kernel owns
// finalize, no domain handler can pick a different backoff.
func TestAllHandlersUseSharedBackoff(t *testing.T) {
	base := time.Date(2026, 5, 22, 12, 0, 0, 0, time.UTC)
	clock := &mutableClock{now: base}
	policy := DefaultBackoffPolicy()

	for _, jobType := range []string{"report_generate", "resume_tailor", "target_import", "email_dispatch"} {
		t.Run(jobType, func(t *testing.T) {
			clock.Set(base)
			store := newFakeStore()
			store.enqueue("job", jobType, 0, base.Add(-time.Minute), base.Add(-time.Minute))

			rt := New(Options{Store: store, Config: testConfig(), Now: clock.Now})
			rt.Register(jobType, JobHandlerFunc(func(_ context.Context, _ ClaimedJob) JobOutcome {
				return JobOutcome{Retryable: true, ErrorCode: "TRANSIENT"}
			}))

			// attempts 1..4 requeue with the shared backoff; attempt 5 dead-letters.
			for attempt := int32(1); attempt <= 5; attempt++ {
				now := clock.Now()
				processed, err := rt.RunOnce(context.Background())
				if err != nil {
					t.Fatalf("attempt %d RunOnce: %v", attempt, err)
				}
				if !processed {
					t.Fatalf("attempt %d: expected a job to be processed", attempt)
				}
				row := store.get("job")
				if attempt < MaxAttempts {
					if row.status != "queued" {
						t.Fatalf("attempt %d status = %s, want queued", attempt, row.status)
					}
					want := now.Add(policy.Next(attempt))
					if !row.availableAt.Equal(want) {
						t.Fatalf("attempt %d availableAt = %s, want %s (shared backoff)", attempt, row.availableAt, want)
					}
					// advance the clock past the backoff so the row is re-claimable.
					clock.Set(row.availableAt)
				} else {
					if row.status != "dead" {
						t.Fatalf("attempt %d status = %s, want dead at MaxAttempts", attempt, row.status)
					}
				}
			}
		})
	}
}
