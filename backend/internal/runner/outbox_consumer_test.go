package runner

import (
	"context"
	"testing"
	"time"
)

func TestDispatcherMissingConsumerDoesNotAck(t *testing.T) {
	now := time.Date(2026, 5, 22, 12, 0, 0, 0, time.UTC)
	store := newFakeOutboxStore()
	store.enqueue("evt-1", "unowned.event", []byte(`{}`), now.Add(-time.Minute))

	d := NewOutboxDispatcher(OutboxDispatcherOptions{Store: store, Now: fixedClock(now)})
	published, err := d.RunOnce(context.Background())
	if err != nil {
		t.Fatalf("RunOnce: %v", err)
	}
	if published != 0 {
		t.Fatalf("published = %d, want 0 (no consumer must not ack)", published)
	}
	row := store.get("evt-1")
	if row == nil || row.status != "pending" {
		t.Fatalf("row status = %v, want pending (retry path, never published)", row)
	}
	if row.attempts != 1 || row.lastErrCode != "OUTBOX_NO_CONSUMER" {
		t.Fatalf("row attempts/err = %d/%s, want 1/OUTBOX_NO_CONSUMER", row.attempts, row.lastErrCode)
	}
}

func TestDispatcherDryRunConsumerRequiresExplicitRegistration(t *testing.T) {
	now := time.Date(2026, 5, 22, 12, 0, 0, 0, time.UTC)
	store := newFakeOutboxStore()
	store.enqueue("evt-1", "dryrun.event", []byte(`{}`), now.Add(-time.Minute))

	clock := &mutableClock{now: now}
	d := NewOutboxDispatcher(OutboxDispatcherOptions{Store: store, Now: clock.Now})
	// Without registration the event must not be published.
	if published, _ := d.RunOnce(context.Background()); published != 0 {
		t.Fatalf("published before registration = %d, want 0", published)
	}
	row := store.get("evt-1")
	if row.status == "published" {
		t.Fatalf("event published without an explicitly registered consumer")
	}
	// Advance past the no-consumer retry backoff so the row is re-claimable.
	clock.Set(row.nextAttemptAt)

	// Explicitly inject a dry-run consumer that always acks.
	d.RegisterConsumer("dryrun.event", OutboxConsumerFunc(func(context.Context, OutboxEvent) error { return nil }))
	if published, _ := d.RunOnce(context.Background()); published != 1 {
		t.Fatalf("published after explicit registration = %d, want 1", published)
	}
	if row := store.get("evt-1"); row == nil || row.status != "published" {
		t.Fatalf("row status = %v, want published after explicit dry-run registration", row)
	}
}

func TestOutboxDispatcher_DeadLetterAtMaxAttempts(t *testing.T) {
	now := time.Date(2026, 5, 22, 12, 0, 0, 0, time.UTC)
	store := newFakeOutboxStore()
	store.enqueue("evt-1", "always.fails", []byte(`{}`), now.Add(-time.Hour))

	clock := &mutableClock{now: now}
	d := NewOutboxDispatcher(OutboxDispatcherOptions{Store: store, Now: clock.Now})
	d.RegisterConsumer("always.fails", OutboxConsumerFunc(func(context.Context, OutboxEvent) error {
		return context.DeadlineExceeded
	}))
	// Drive five attempts, advancing the clock past each backoff window.
	for i := 0; i < 5; i++ {
		if _, err := d.RunOnce(context.Background()); err != nil {
			t.Fatalf("attempt %d RunOnce: %v", i+1, err)
		}
		row := store.get("evt-1")
		if i < 4 {
			if row.status != "pending" {
				t.Fatalf("attempt %d status = %s, want pending", i+1, row.status)
			}
			clock.Set(row.nextAttemptAt)
		} else {
			if row.status != "failed" {
				t.Fatalf("attempt %d status = %s, want failed (dead-letter)", i+1, row.status)
			}
		}
	}
	row := store.get("evt-1")
	// Redaction: the persisted message must be the stable redacted summary, not
	// the raw consumer error string.
	if row.lastErrMsg != "consumer rejected event" {
		t.Fatalf("last_error_message = %q, want redacted summary", row.lastErrMsg)
	}
}
