package runner

import (
	"context"
	"testing"
	"time"
)

// TestOutboxDispatcher_SkipsSourceEventOnly verifies the dispatcher publishes a
// practice.session.completed event to its consumer but does NOT enqueue a
// second report_generate job, because report_generate is source_event_only and
// the producing transaction already created the job (spec D-7).
func TestOutboxDispatcher_SkipsSourceEventOnly(t *testing.T) {
	now := time.Date(2026, 5, 22, 12, 0, 0, 0, time.UTC)
	store := newFakeOutboxStore()
	store.enqueue("evt-practice", "practice.session.completed", []byte(`{}`), now.Add(-time.Minute))

	var scheduled []string
	d := NewOutboxDispatcher(OutboxDispatcherOptions{
		Store: store,
		Now:   fixedClock(now),
		EventJobBinding: map[string]string{
			"practice.session.completed": "report_generate",
		},
		JobScheduler: func(_ context.Context, jobType string, _ OutboxEvent) error {
			scheduled = append(scheduled, jobType)
			return nil
		},
	})
	d.RegisterConsumer("practice.session.completed", OutboxConsumerFunc(func(context.Context, OutboxEvent) error { return nil }))

	published, err := d.RunOnce(context.Background())
	if err != nil {
		t.Fatalf("RunOnce: %v", err)
	}
	if published != 1 {
		t.Fatalf("published = %d, want 1", published)
	}
	if len(scheduled) != 0 {
		t.Fatalf("scheduled follow-up jobs = %v, want none (source_event_only)", scheduled)
	}
}

// TestOutboxDispatcher_SchedulesFollowUpForNonSourceEvent confirms the guard is
// load-bearing: a non-source-event-only binding does schedule a follow-up job.
func TestOutboxDispatcher_SchedulesFollowUpForNonSourceEvent(t *testing.T) {
	now := time.Date(2026, 5, 22, 12, 0, 0, 0, time.UTC)
	store := newFakeOutboxStore()
	store.enqueue("evt-import", "target.import.requested", []byte(`{}`), now.Add(-time.Minute))

	var scheduled []string
	d := NewOutboxDispatcher(OutboxDispatcherOptions{
		Store:           store,
		Now:             fixedClock(now),
		EventJobBinding: map[string]string{"target.import.requested": "target_import"},
		JobScheduler: func(_ context.Context, jobType string, _ OutboxEvent) error {
			scheduled = append(scheduled, jobType)
			return nil
		},
	})
	d.RegisterConsumer("target.import.requested", OutboxConsumerFunc(func(context.Context, OutboxEvent) error { return nil }))

	if _, err := d.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce: %v", err)
	}
	if len(scheduled) != 1 || scheduled[0] != "target_import" {
		t.Fatalf("scheduled = %v, want [target_import] (trigger_creates_job)", scheduled)
	}
}
