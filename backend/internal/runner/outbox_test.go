package runner

import (
	"context"
	"sort"
	"sync"
	"testing"
	"time"
)

type fakeOutboxRow struct {
	id            string
	eventName     string
	aggType       string
	aggID         string
	payload       []byte
	status        string
	attempts      int32
	nextAttemptAt time.Time
	createdAt     time.Time
	lastErrCode   string
	lastErrMsg    string
}

type fakeOutboxStore struct {
	mu   sync.Mutex
	rows []*fakeOutboxRow
}

func newFakeOutboxStore() *fakeOutboxStore { return &fakeOutboxStore{} }

func (s *fakeOutboxStore) enqueue(id, eventName string, payload []byte, createdAt time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.rows = append(s.rows, &fakeOutboxRow{
		id: id, eventName: eventName, aggType: "test", aggID: id, payload: payload,
		status: "pending", nextAttemptAt: createdAt, createdAt: createdAt,
	})
}

func (s *fakeOutboxStore) get(id string) *fakeOutboxRow {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, r := range s.rows {
		if r.id == id {
			clone := *r
			return &clone
		}
	}
	return nil
}

func (s *fakeOutboxStore) CountPending(context.Context) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var n int64
	for _, r := range s.rows {
		if r.status == "pending" {
			n++
		}
	}
	return n, nil
}

func (s *fakeOutboxStore) ProcessPendingBatch(_ context.Context, now time.Time, batch int, backoff BackoffPolicy, fn func(OutboxRow) OutboxResult) (OutboxBatchOutcome, error) {
	s.mu.Lock()
	pending := make([]*fakeOutboxRow, 0)
	for _, r := range s.rows {
		if r.status == "pending" && !r.nextAttemptAt.After(now) {
			pending = append(pending, r)
		}
	}
	sort.SliceStable(pending, func(i, j int) bool {
		if !pending[i].nextAttemptAt.Equal(pending[j].nextAttemptAt) {
			return pending[i].nextAttemptAt.Before(pending[j].nextAttemptAt)
		}
		return pending[i].createdAt.Before(pending[j].createdAt)
	})
	if batch > 0 && len(pending) > batch {
		pending = pending[:batch]
	}
	s.mu.Unlock()

	var outcome OutboxBatchOutcome
	for _, r := range pending {
		res := fn(OutboxRow{EventID: r.id, EventName: r.eventName, AggregateType: r.aggType, AggregateID: r.aggID, Payload: r.payload, PublishAttempts: r.attempts})
		s.mu.Lock()
		newAttempts := r.attempts + 1
		switch {
		case res.Published:
			r.status = "published"
			outcome.Published++
		case newAttempts >= MaxAttempts:
			r.status = "failed"
			r.attempts = newAttempts
			r.lastErrCode = res.ErrorCode
			r.lastErrMsg = res.ErrorMessage
			outcome.DeadLettered++
		default:
			r.attempts = newAttempts
			r.nextAttemptAt = now.Add(backoff.Next(newAttempts))
			r.lastErrCode = res.ErrorCode
			r.lastErrMsg = res.ErrorMessage
			outcome.Retried++
		}
		s.mu.Unlock()
	}
	return outcome, nil
}

func TestOutboxDispatcher_PublishesToRegisteredConsumer(t *testing.T) {
	now := time.Date(2026, 5, 22, 12, 0, 0, 0, time.UTC)
	store := newFakeOutboxStore()
	store.enqueue("evt-1", "report.generated", []byte(`{}`), now.Add(-time.Minute))

	d := NewOutboxDispatcher(OutboxDispatcherOptions{Store: store, Now: fixedClock(now)})
	var consumed []string
	d.RegisterConsumer("report.generated", OutboxConsumerFunc(func(_ context.Context, e OutboxEvent) error {
		consumed = append(consumed, e.EventID)
		return nil
	}))
	published, err := d.RunOnce(context.Background())
	if err != nil {
		t.Fatalf("RunOnce: %v", err)
	}
	if published != 1 || len(consumed) != 1 || consumed[0] != "evt-1" {
		t.Fatalf("published=%d consumed=%v", published, consumed)
	}
	if row := store.get("evt-1"); row == nil || row.status != "published" {
		t.Fatalf("row status = %v, want published", row)
	}
}

func TestOutboxDispatcher_BatchSizeLimit(t *testing.T) {
	now := time.Date(2026, 5, 22, 12, 0, 0, 0, time.UTC)
	store := newFakeOutboxStore()
	for i := 0; i < 150; i++ {
		store.enqueue("evt-"+itoa(i), "x.event", []byte(`{}`), now.Add(-time.Minute))
	}
	d := NewOutboxDispatcher(OutboxDispatcherOptions{Store: store, Batch: 100, Now: fixedClock(now)})
	d.RegisterConsumer("x.event", OutboxConsumerFunc(func(context.Context, OutboxEvent) error { return nil }))
	published, err := d.RunOnce(context.Background())
	if err != nil {
		t.Fatalf("RunOnce: %v", err)
	}
	if published != 100 {
		t.Fatalf("published = %d, want 100 (batch limit)", published)
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	digits := []byte{}
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}
