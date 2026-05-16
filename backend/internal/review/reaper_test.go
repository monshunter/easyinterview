package review

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"
)

func TestReaperReclaimsExpiredLease(t *testing.T) {
	now := time.Date(2026, 5, 15, 13, 0, 0, 0, time.UTC)
	store := &fakeRunnerStore{reaperCount: 3}
	reaper := NewReaper(ReaperOptions{
		Store:        store,
		LeaseTimeout: 2 * time.Minute,
		Now:          func() time.Time { return now },
		Logger:       slog.New(slog.NewTextHandler(io.Discard, nil)),
	})

	reclaimed, err := reaper.RunOnce(context.Background())
	if err != nil {
		t.Fatalf("RunOnce: %v", err)
	}
	if reclaimed != 3 {
		t.Fatalf("reclaimed = %d, want 3", reclaimed)
	}
	if store.reaperJobType != ReportGenerateJobType {
		t.Fatalf("reaper jobType = %q", store.reaperJobType)
	}
	if !store.reaperOlderThan.Equal(now.Add(-2*time.Minute)) || !store.reaperNow.Equal(now) {
		t.Fatalf("reaper times olderThan=%s now=%s", store.reaperOlderThan, store.reaperNow)
	}
}
