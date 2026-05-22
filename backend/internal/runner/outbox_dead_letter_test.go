//go:build integration

package runner

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"
)

func TestOutboxDispatcher_DeadLetterAtAttemptFive(t *testing.T) {
	db := openRunnerTestDB(t)
	store := NewSQLOutboxStore(db)
	clock := &mutableClock{now: time.Now().UTC()}
	id := "0197d143-0000-7000-8000-000000000001"
	insertOutboxEvent(t, db, id, "runner.deadletter.event", `{}`, clock.Now().Add(-time.Hour))

	d := NewOutboxDispatcher(OutboxDispatcherOptions{Store: store, Now: clock.Now})
	d.RegisterConsumer("runner.deadletter.event", OutboxConsumerFunc(func(context.Context, OutboxEvent) error {
		return errors.New("provider exploded with secret token abc123")
	}))

	for i := 0; i < 5; i++ {
		if _, err := d.RunOnce(context.Background()); err != nil {
			t.Fatalf("attempt %d RunOnce: %v", i+1, err)
		}
		var status string
		var attempts int32
		var nextAt time.Time
		if err := db.QueryRow(`select publish_status, publish_attempts, next_attempt_at from outbox_events where id=$1`, id).Scan(&status, &attempts, &nextAt); err != nil {
			t.Fatalf("read back: %v", err)
		}
		if i < 4 {
			if status != "pending" {
				t.Fatalf("attempt %d status = %s, want pending", i+1, status)
			}
			clock.Set(nextAt)
		} else {
			if status != "failed" || attempts != 5 {
				t.Fatalf("attempt 5 status=%s attempts=%d, want failed/5", status, attempts)
			}
		}
	}
}

func TestOutboxDispatcher_RedactsLastError(t *testing.T) {
	db := openRunnerTestDB(t)
	store := NewSQLOutboxStore(db)
	clock := &mutableClock{now: time.Now().UTC()}
	id := "0197d143-0000-7000-8000-000000000002"
	insertOutboxEvent(t, db, id, "runner.redact.event", `{}`, clock.Now().Add(-time.Hour))

	d := NewOutboxDispatcher(OutboxDispatcherOptions{Store: store, Now: clock.Now})
	d.RegisterConsumer("runner.redact.event", OutboxConsumerFunc(func(context.Context, OutboxEvent) error {
		return errors.New("raw provider response: magic-link-token-SHOULD-NOT-PERSIST")
	}))

	if _, err := d.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce: %v", err)
	}
	var code, msg sql.NullString
	if err := db.QueryRow(`select last_error_code, last_error_message from outbox_events where id=$1`, id).Scan(&code, &msg); err != nil {
		t.Fatalf("read back: %v", err)
	}
	if code.String != "OUTBOX_CONSUMER_FAILED" {
		t.Fatalf("last_error_code = %q, want OUTBOX_CONSUMER_FAILED", code.String)
	}
	if msg.String != "consumer rejected event" {
		t.Fatalf("last_error_message = %q, want redacted summary", msg.String)
	}
	if containsToken(msg.String, "magic-link-token-SHOULD-NOT-PERSIST") || containsToken(msg.String, "raw provider response") {
		t.Fatalf("redaction red line violated: %q", msg.String)
	}
}

func containsToken(haystack, needle string) bool {
	return len(haystack) >= len(needle) && (haystack == needle || indexOf(haystack, needle) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
