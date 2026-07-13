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

// TestAllBusinessHandlersUseBusinessBackoff proves that every retryable
// business job uses the same 10s/20s/40s/80s policy. Outbox delivery has its
// own infrastructure policy and is tested separately.
func TestAllBusinessHandlersUseBusinessBackoff(t *testing.T) {
	base := time.Date(2026, 5, 22, 12, 0, 0, 0, time.UTC)
	clock := &mutableClock{now: base}
	policy := DefaultBackoffPolicy()

	for _, tc := range []struct {
		jobType     string
		maxAttempts int32
	}{
		{jobType: "report_generate", maxAttempts: 4},
		{jobType: "resume_tailor", maxAttempts: MaxAttempts},
		{jobType: "target_import", maxAttempts: MaxAttempts},
		{jobType: "email_dispatch", maxAttempts: MaxAttempts},
	} {
		t.Run(tc.jobType, func(t *testing.T) {
			clock.Set(base)
			store := newFakeStore()
			store.enqueue("job", tc.jobType, 0, base.Add(-time.Minute), base.Add(-time.Minute))
			store.setMaxAttempts("job", tc.maxAttempts)

			rt := New(Options{Store: store, Config: testConfig(), Now: clock.Now})
			rt.Register(tc.jobType, JobHandlerFunc(func(_ context.Context, _ ClaimedJob) JobOutcome {
				return JobOutcome{Retryable: true, ErrorCode: "TRANSIENT"}
			}))

			for attempt := int32(1); attempt <= tc.maxAttempts; attempt++ {
				now := clock.Now()
				processed, err := rt.RunOnce(context.Background())
				if err != nil {
					t.Fatalf("attempt %d RunOnce: %v", attempt, err)
				}
				if !processed {
					t.Fatalf("attempt %d: expected a job to be processed", attempt)
				}
				row := store.get("job")
				if attempt < tc.maxAttempts {
					if row.status != "queued" {
						t.Fatalf("attempt %d status = %s, want queued", attempt, row.status)
					}
					want := now.Add(policy.Next(attempt))
					if !row.availableAt.Equal(want) {
						t.Fatalf("attempt %d availableAt = %s, want %s (business backoff)", attempt, row.availableAt, want)
					}
					// advance the clock past the backoff so the row is re-claimable.
					clock.Set(row.availableAt)
				} else {
					if row.status != "dead" {
						t.Fatalf("attempt %d status = %s, want dead at max_attempts=%d", attempt, row.status, tc.maxAttempts)
					}
					if tc.jobType == "report_generate" && !row.availableAt.Equal(now) {
						t.Fatalf("report attempt4 scheduled a forbidden 80s retry: availableAt=%s now=%s", row.availableAt, now)
					}
				}
			}
		})
	}
}
