//go:build integration

package runner

import (
	"context"
	"database/sql"
	"sync"
	"testing"
	"time"
)

func insertOutboxEvent(t *testing.T, db *sql.DB, id, eventName string, payload string, nextAttemptAt time.Time) {
	t.Helper()
	ctx := context.Background()
	_, _ = db.ExecContext(ctx, `delete from outbox_events where id = $1`, id)
	_, err := db.ExecContext(ctx, `
insert into outbox_events (
  id, event_name, event_version, aggregate_type, aggregate_id, payload,
  publish_status, publish_attempts, next_attempt_at, created_at
) values ($1, $2, 1, 'test', $1, $3::jsonb, 'pending', 0, $4, $4)`,
		id, eventName, payload, nextAttemptAt)
	if err != nil {
		t.Fatalf("insert outbox_events: %v", err)
	}
	t.Cleanup(func() { _, _ = db.ExecContext(context.Background(), `delete from outbox_events where id = $1`, id) })
}

func TestOutboxDispatcher_ClaimsPendingBatch(t *testing.T) {
	db := openRunnerTestDB(t)
	store := NewSQLOutboxStore(db)
	now := time.Now().UTC()
	ids := []string{
		"0197d140-0000-7000-8000-000000000001",
		"0197d140-0000-7000-8000-000000000002",
		"0197d140-0000-7000-8000-000000000003",
	}
	for i, id := range ids {
		insertOutboxEvent(t, db, id, "runner.it.event", `{}`, now.Add(-time.Duration(i+1)*time.Minute))
	}

	d := NewOutboxDispatcher(OutboxDispatcherOptions{Store: store, Batch: 100, Now: func() time.Time { return now }})
	var mu sync.Mutex
	order := []string{}
	d.RegisterConsumer("runner.it.event", OutboxConsumerFunc(func(_ context.Context, e OutboxEvent) error {
		mu.Lock()
		order = append(order, e.EventID)
		mu.Unlock()
		return nil
	}))
	published, err := d.RunOnce(context.Background())
	if err != nil {
		t.Fatalf("RunOnce: %v", err)
	}
	if published != 3 {
		t.Fatalf("published = %d, want 3", published)
	}
	// Oldest next_attempt_at first.
	if len(order) != 3 || order[0] != ids[2] {
		t.Fatalf("publish order = %v, want oldest-first starting %s", order, ids[2])
	}
	for _, id := range ids {
		var status string
		if err := db.QueryRow(`select publish_status from outbox_events where id=$1`, id).Scan(&status); err != nil {
			t.Fatalf("read back: %v", err)
		}
		if status != "published" {
			t.Fatalf("%s status = %s, want published", id, status)
		}
	}
}

func TestOutboxDispatcher_BatchSizeLimitIntegration(t *testing.T) {
	db := openRunnerTestDB(t)
	store := NewSQLOutboxStore(db)
	now := time.Now().UTC()
	const total = 105
	prefix := "0197d141-0000-7000-8000-0000000"
	for i := 0; i < total; i++ {
		id := prefix + pad5(i)
		insertOutboxEvent(t, db, id, "runner.batch.event", `{}`, now.Add(-time.Hour))
	}
	d := NewOutboxDispatcher(OutboxDispatcherOptions{Store: store, Batch: 100, Now: func() time.Time { return now }})
	d.RegisterConsumer("runner.batch.event", OutboxConsumerFunc(func(context.Context, OutboxEvent) error { return nil }))
	published, err := d.RunOnce(context.Background())
	if err != nil {
		t.Fatalf("RunOnce: %v", err)
	}
	if published != 100 {
		t.Fatalf("published = %d, want 100 (batch cap)", published)
	}
}

func TestOutboxDispatcher_DuplicateEventIdHandledIdempotently(t *testing.T) {
	db := openRunnerTestDB(t)
	store := NewSQLOutboxStore(db)
	now := time.Now().UTC()
	id := "0197d142-0000-7000-8000-000000000001"
	insertOutboxEvent(t, db, id, "runner.idem.event", `{}`, now.Add(-time.Minute))

	processed := map[string]int{}
	var mu sync.Mutex
	consumer := OutboxConsumerFunc(func(_ context.Context, e OutboxEvent) error {
		mu.Lock()
		processed[e.EventID]++
		mu.Unlock()
		return nil
	})
	d := NewOutboxDispatcher(OutboxDispatcherOptions{Store: store, Now: func() time.Time { return now }})
	d.RegisterConsumer("runner.idem.event", consumer)

	if _, err := d.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce 1: %v", err)
	}
	// Simulate at-least-once redelivery: reset the row to pending and run again.
	if _, err := db.Exec(`update outbox_events set publish_status='pending', published_at=null where id=$1`, id); err != nil {
		t.Fatalf("reset row: %v", err)
	}
	if _, err := d.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce 2: %v", err)
	}
	// The consumer saw the event twice (at-least-once), and an idempotent
	// consumer keyed by eventID can dedupe; the dispatcher itself published it
	// each time it was pending.
	mu.Lock()
	count := processed[id]
	mu.Unlock()
	if count != 2 {
		t.Fatalf("consumer invocations = %d, want 2 (at-least-once)", count)
	}
	var status string
	if err := db.QueryRow(`select publish_status from outbox_events where id=$1`, id).Scan(&status); err != nil {
		t.Fatalf("read back: %v", err)
	}
	if status != "published" {
		t.Fatalf("status = %s, want published", status)
	}
}

func pad5(n int) string {
	s := itoa(n)
	for len(s) < 5 {
		s = "0" + s
	}
	return s
}
