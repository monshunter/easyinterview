package runner

import (
	"context"
	"log/slog"
	"sync"
	"testing"
	"time"
)

// captureStore is the shared backing store for capturingHandler and any
// derived handlers produced by slog's With(...).
type captureStore struct {
	mu      sync.Mutex
	records []map[string]string
}

func (s *captureStore) field(msg, key string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, rec := range s.records {
		if rec["msg"] == msg {
			v, ok := rec[key]
			return v, ok
		}
	}
	return "", false
}

// capturingHandler is a minimal slog.Handler that records emitted records and
// their merged attributes for assertions. Derived handlers (via WithAttrs)
// share the same captureStore so With(...)-chained logs remain visible.
type capturingHandler struct {
	store *captureStore
	attrs []slog.Attr
}

func newCapturingHandler() *capturingHandler {
	return &capturingHandler{store: &captureStore{}}
}

func (h *capturingHandler) Enabled(context.Context, slog.Level) bool { return true }

func (h *capturingHandler) Handle(_ context.Context, r slog.Record) error {
	fields := map[string]string{"msg": r.Message, "level": r.Level.String()}
	for _, a := range h.attrs {
		fields[a.Key] = a.Value.String()
	}
	r.Attrs(func(a slog.Attr) bool {
		fields[a.Key] = a.Value.String()
		return true
	})
	h.store.mu.Lock()
	h.store.records = append(h.store.records, fields)
	h.store.mu.Unlock()
	return nil
}

type staleFinalizeStore struct{ *fakeStore }

func (s *staleFinalizeStore) FinalizeAsyncJob(context.Context, string, int32, JobOutcome, time.Time, time.Time) error {
	return ErrStaleLease
}

func (h *capturingHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	merged := append([]slog.Attr{}, h.attrs...)
	merged = append(merged, attrs...)
	return &capturingHandler{store: h.store, attrs: merged}
}

func (h *capturingHandler) WithGroup(string) slog.Handler { return h }

const sampleTraceID = "4bf92f3577b34da6a3ce929d0e0e4736"
const sampleTraceparent = "00-" + sampleTraceID + "-00f067aa0ba902b7-01"

func TestRuntime_HandlerInheritsTraceparent(t *testing.T) {
	now := time.Date(2026, 5, 22, 12, 0, 0, 0, time.UTC)
	store := newFakeStore()
	store.enqueue("job-trace", "target_import", 0, now.Add(-time.Minute), now.Add(-time.Minute))
	store.setPayload("job-trace", []byte(`{"traceparent":"`+sampleTraceparent+`"}`))

	rt := New(Options{Store: store, Config: testConfig(), Now: fixedClock(now)})
	var seen string
	rt.Register("target_import", JobHandlerFunc(func(ctx context.Context, _ ClaimedJob) JobOutcome {
		seen = TraceIDFromContext(ctx)
		return JobOutcome{Succeeded: true}
	}))
	if _, err := rt.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce: %v", err)
	}
	if seen != sampleTraceID {
		t.Fatalf("handler ctx trace id = %q, want %q", seen, sampleTraceID)
	}
}

func TestRuntime_HandlerLogsTraceIdField(t *testing.T) {
	now := time.Date(2026, 5, 22, 12, 0, 0, 0, time.UTC)
	store := newFakeStore()
	store.enqueue("job-trace", "target_import", 0, now.Add(-time.Minute), now.Add(-time.Minute))
	store.setPayload("job-trace", []byte(`{"traceparent":"`+sampleTraceparent+`"}`))

	capture := newCapturingHandler()
	logger := slog.New(capture)
	rt := New(Options{Store: store, Config: testConfig(), Now: fixedClock(now), Logger: logger})
	rt.Register("target_import", JobHandlerFunc(func(_ context.Context, _ ClaimedJob) JobOutcome {
		return JobOutcome{Succeeded: true}
	}))
	if _, err := rt.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce: %v", err)
	}
	got, ok := capture.store.field("runner.handle completed", "trace_id")
	if !ok {
		t.Fatalf("runner.handle completed log missing trace_id field")
	}
	if got != sampleTraceID {
		t.Fatalf("trace_id field = %q, want %q", got, sampleTraceID)
	}
}

func TestRuntime_HandlerSkipsTraceWhenMissing(t *testing.T) {
	now := time.Date(2026, 5, 22, 12, 0, 0, 0, time.UTC)
	store := newFakeStore()
	store.enqueue("job-notrace", "target_import", 0, now.Add(-time.Minute), now.Add(-time.Minute))

	rt := New(Options{Store: store, Config: testConfig(), Now: fixedClock(now)})
	var seen = "preset"
	rt.Register("target_import", JobHandlerFunc(func(ctx context.Context, _ ClaimedJob) JobOutcome {
		seen = TraceIDFromContext(ctx)
		return JobOutcome{Succeeded: true}
	}))
	if _, err := rt.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce: %v", err)
	}
	if seen != "" {
		t.Fatalf("expected empty trace id when payload has none, got %q", seen)
	}
}

func TestRuntime_StaleFinalizeIsExpectedFencedDebugSignal(t *testing.T) {
	now := time.Date(2026, 7, 13, 12, 0, 0, 0, time.UTC)
	base := newFakeStore()
	base.enqueue("job-stale", "target_import", 0, now.Add(-time.Minute), now.Add(-time.Minute))
	capture := newCapturingHandler()
	rt := New(Options{Store: &staleFinalizeStore{fakeStore: base}, Config: testConfig(), Now: fixedClock(now), Logger: slog.New(capture)})
	rt.Register("target_import", JobHandlerFunc(func(context.Context, ClaimedJob) JobOutcome {
		return JobOutcome{Succeeded: true}
	}))
	if processed, err := rt.RunOnce(context.Background()); err != nil || !processed {
		t.Fatalf("RunOnce processed=%t err=%v", processed, err)
	}
	if level, ok := capture.store.field("runner.finalize fenced stale lease", "level"); !ok || level != "DEBUG" {
		t.Fatalf("stale finalize signal level=%q present=%t want DEBUG", level, ok)
	}
	if _, ok := capture.store.field("runner.finalize failed", "level"); ok {
		t.Fatal("stale finalize was emitted as an error")
	}
}
