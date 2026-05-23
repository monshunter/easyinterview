package runner

import (
	"context"
	"log/slog"
	"testing"
	"time"
)

func TestOutboxDispatcher_PropagatesTraceParent(t *testing.T) {
	now := time.Date(2026, 5, 22, 12, 0, 0, 0, time.UTC)
	store := newFakeOutboxStore()
	store.enqueue("evt-1", "debrief.created", []byte(`{"traceparent":"`+sampleTraceparent+`"}`), now.Add(-time.Minute))

	d := NewOutboxDispatcher(OutboxDispatcherOptions{Store: store, Now: fixedClock(now)})
	var seen string
	d.RegisterConsumer("debrief.created", OutboxConsumerFunc(func(ctx context.Context, _ OutboxEvent) error {
		seen = TraceIDFromContext(ctx)
		return nil
	}))
	if _, err := d.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce: %v", err)
	}
	if seen != sampleTraceID {
		t.Fatalf("consumer ctx trace id = %q, want %q", seen, sampleTraceID)
	}
}

func TestOutboxDispatcher_WarnsOnMissingTrace(t *testing.T) {
	now := time.Date(2026, 5, 22, 12, 0, 0, 0, time.UTC)
	store := newFakeOutboxStore()
	store.enqueue("evt-1", "debrief.created", []byte(`{}`), now.Add(-time.Minute))

	capture := newCapturingHandler()
	d := NewOutboxDispatcher(OutboxDispatcherOptions{Store: store, Now: fixedClock(now), Logger: slog.New(capture)})
	d.RegisterConsumer("debrief.created", OutboxConsumerFunc(func(context.Context, OutboxEvent) error { return nil }))

	published, err := d.RunOnce(context.Background())
	if err != nil {
		t.Fatalf("RunOnce: %v", err)
	}
	// Missing trace must downgrade to a warn but still publish.
	if published != 1 {
		t.Fatalf("published = %d, want 1 (publish continues without trace)", published)
	}
	if _, ok := capture.store.field("outbox event missing trace_id; publishing without trace context", "event_id"); !ok {
		t.Fatalf("expected a warn log for missing trace_id")
	}
}
