package runner

import (
	"context"
	"testing"
	"time"
)

func TestReaper_ReclaimsExpiredLeases(t *testing.T) {
	now := time.Date(2026, 5, 22, 12, 0, 0, 0, time.UTC)
	store := newFakeStore()
	// Three job types stuck running, locked well before the timeout window.
	store.enqueueRunning("ti-1", "target_import", now.Add(-10*time.Minute), 1)
	store.enqueueRunning("pd-1", "privacy_delete", now.Add(-10*time.Minute), 2)
	store.enqueueRunning("dg-1", "debrief_generate", now.Add(-10*time.Minute), 1)
	// A freshly locked row must NOT be reclaimed.
	store.enqueueRunning("ti-2", "target_import", now.Add(-time.Second), 1)

	r := NewReaper(ReaperOptions{
		Store:        store,
		JobTypes:     []string{"target_import", "privacy_delete", "debrief_generate"},
		LeaseTimeout: 5 * time.Minute,
		Now:          fixedClock(now),
	})
	reclaimed, err := r.RunOnce(context.Background())
	if err != nil {
		t.Fatalf("RunOnce: %v", err)
	}
	if reclaimed != 3 {
		t.Fatalf("reclaimed = %d, want 3", reclaimed)
	}
	for _, id := range []string{"ti-1", "pd-1", "dg-1"} {
		if row := store.get(id); row == nil || row.status != "queued" || row.lockedAt != nil {
			t.Fatalf("%s = %v, want queued + locked_at nil", id, row)
		}
	}
	if row := store.get("ti-2"); row == nil || row.status != "running" {
		t.Fatalf("ti-2 = %v, want still running (lease not expired)", row)
	}
}

func TestReaper_DoesNotIncrementAttempts(t *testing.T) {
	now := time.Date(2026, 5, 22, 12, 0, 0, 0, time.UTC)
	store := newFakeStore()
	store.enqueueRunning("ti-1", "target_import", now.Add(-10*time.Minute), 3)

	r := NewReaper(ReaperOptions{
		Store:        store,
		JobTypes:     []string{"target_import"},
		LeaseTimeout: 5 * time.Minute,
		Now:          fixedClock(now),
	})
	if _, err := r.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce: %v", err)
	}
	row := store.get("ti-1")
	if row == nil || row.attempts != 3 {
		t.Fatalf("attempts = %v, want unchanged at 3 (lease timeout is not a business failure)", row)
	}
}
