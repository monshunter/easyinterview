package runner

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"sync"
	"time"

	sharedjobs "github.com/monshunter/easyinterview/backend/internal/shared/jobs"
)

const defaultOutboxBatch = 100

// OutboxRow is a pending outbox_events row claimed for publishing.
type OutboxRow struct {
	EventID         string
	EventName       string
	AggregateType   string
	AggregateID     string
	Payload         []byte
	PublishAttempts int32
}

// OutboxEvent is the event handed to a registered consumer.
type OutboxEvent struct {
	EventID       string
	EventName     string
	AggregateType string
	AggregateID   string
	Payload       []byte
}

// OutboxConsumer publishes a single outbox event to a downstream sink
// (analytics double-write, audit, etc.). Consumers must be idempotent: the
// dispatcher guarantees at-least-once delivery.
type OutboxConsumer interface {
	Consume(ctx context.Context, event OutboxEvent) error
}

// OutboxConsumerFunc adapts a function to OutboxConsumer.
type OutboxConsumerFunc func(ctx context.Context, event OutboxEvent) error

// Consume satisfies OutboxConsumer.
func (f OutboxConsumerFunc) Consume(ctx context.Context, event OutboxEvent) error {
	return f(ctx, event)
}

// OutboxResult is the per-row disposition the dispatcher returns to the store.
type OutboxResult struct {
	Published bool
	// ErrorCode / ErrorMessage are persisted (redacted) when not published.
	ErrorCode    string
	ErrorMessage string
}

// OutboxBatchOutcome summarizes a single scan.
type OutboxBatchOutcome struct {
	Published    int
	Retried      int
	DeadLettered int
}

// OutboxStore is the persistence boundary for the dispatcher.
type OutboxStore interface {
	// ProcessPendingBatch claims up to batch pending rows with FOR UPDATE SKIP
	// LOCKED ordered by next_attempt_at asc, created_at asc, invokes fn for each,
	// and applies the disposition atomically: published rows are marked
	// published; otherwise publish_attempts is incremented and the row is
	// rescheduled with backoff, or dead-lettered to failed at MaxAttempts.
	ProcessPendingBatch(ctx context.Context, now time.Time, batch int, backoff BackoffPolicy, fn func(OutboxRow) OutboxResult) (OutboxBatchOutcome, error)
	// CountPending returns the number of currently pending rows.
	CountPending(ctx context.Context) (int64, error)
}

// SQLOutboxStore is the Postgres-backed OutboxStore. Retry column names follow
// B3 §2.1 (publish_attempts / next_attempt_at / locked_at / last_error_code /
// last_error_message).
type SQLOutboxStore struct {
	db *sql.DB
}

// NewSQLOutboxStore wires a SQLOutboxStore.
func NewSQLOutboxStore(db *sql.DB) *SQLOutboxStore { return &SQLOutboxStore{db: db} }

func (s *SQLOutboxStore) CountPending(ctx context.Context) (int64, error) {
	if s == nil || s.db == nil {
		return 0, fmt.Errorf("outbox store db is nil")
	}
	var n int64
	if err := s.db.QueryRowContext(ctx, `select count(*) from outbox_events where publish_status = 'pending'`).Scan(&n); err != nil {
		return 0, fmt.Errorf("count pending outbox: %w", err)
	}
	return n, nil
}

func (s *SQLOutboxStore) ProcessPendingBatch(ctx context.Context, now time.Time, batch int, backoff BackoffPolicy, fn func(OutboxRow) OutboxResult) (OutboxBatchOutcome, error) {
	if s == nil || s.db == nil {
		return OutboxBatchOutcome{}, fmt.Errorf("outbox store db is nil")
	}
	if batch <= 0 {
		batch = defaultOutboxBatch
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return OutboxBatchOutcome{}, fmt.Errorf("begin outbox tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	rows, err := tx.QueryContext(ctx, `
select id, event_name, aggregate_type, aggregate_id, payload, publish_attempts
from outbox_events
where publish_status = 'pending' and next_attempt_at <= $1
order by next_attempt_at asc, created_at asc
for update skip locked
limit $2`, now, batch)
	if err != nil {
		return OutboxBatchOutcome{}, fmt.Errorf("claim pending outbox: %w", err)
	}
	claimed := make([]OutboxRow, 0, batch)
	for rows.Next() {
		var row OutboxRow
		if err := rows.Scan(&row.EventID, &row.EventName, &row.AggregateType, &row.AggregateID, &row.Payload, &row.PublishAttempts); err != nil {
			rows.Close()
			return OutboxBatchOutcome{}, fmt.Errorf("scan outbox row: %w", err)
		}
		claimed = append(claimed, row)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return OutboxBatchOutcome{}, fmt.Errorf("iterate outbox rows: %w", err)
	}
	rows.Close()

	var outcome OutboxBatchOutcome
	for _, row := range claimed {
		res := fn(row)
		newAttempts := row.PublishAttempts + 1
		switch {
		case res.Published:
			if _, err := tx.ExecContext(ctx, `
update outbox_events
set publish_status = 'published', published_at = $1, locked_at = null,
    last_error_code = null, last_error_message = null
where id = $2`, now, row.EventID); err != nil {
				return OutboxBatchOutcome{}, fmt.Errorf("mark outbox published: %w", err)
			}
			outcome.Published++
		case newAttempts >= MaxAttempts:
			if _, err := tx.ExecContext(ctx, `
update outbox_events
set publish_status = 'failed', publish_attempts = $1, locked_at = null,
    last_error_code = $2, last_error_message = $3
where id = $4`, newAttempts, nullableString(res.ErrorCode), nullableString(res.ErrorMessage), row.EventID); err != nil {
				return OutboxBatchOutcome{}, fmt.Errorf("dead-letter outbox: %w", err)
			}
			outcome.DeadLettered++
		default:
			nextAt := now.Add(backoff.Next(newAttempts))
			if _, err := tx.ExecContext(ctx, `
update outbox_events
set publish_attempts = $1, next_attempt_at = $2, locked_at = null,
    last_error_code = $3, last_error_message = $4
where id = $5`, newAttempts, nextAt, nullableString(res.ErrorCode), nullableString(res.ErrorMessage), row.EventID); err != nil {
				return OutboxBatchOutcome{}, fmt.Errorf("reschedule outbox: %w", err)
			}
			outcome.Retried++
		}
	}
	if err := tx.Commit(); err != nil {
		return OutboxBatchOutcome{}, fmt.Errorf("commit outbox tx: %w", err)
	}
	return outcome, nil
}

// OutboxDispatcherOptions configures an OutboxDispatcher.
type OutboxDispatcherOptions struct {
	Store        OutboxStore
	Backoff      BackoffPolicy
	Batch        int
	ScanInterval time.Duration
	Now          func() time.Time
	Logger       *slog.Logger
	Metrics      Metrics
	// EventJobBinding maps an outbox event_name to the job_type it would
	// trigger; JobScheduler enqueues that follow-up job. Both are optional and
	// unset in P0 (the producing transaction creates the job). They exist so the
	// source_event_only guard (spec D-7) is explicit and testable.
	EventJobBinding map[string]string
	JobScheduler    func(ctx context.Context, jobType string, event OutboxEvent) error
}

// OutboxDispatcher publishes pending outbox_events to registered consumers
// (spec D-6). It never marks an event published unless a consumer acked it, and
// it never creates a second async_job for a source_event_only event (D-7).
type OutboxDispatcher struct {
	store        OutboxStore
	backoff      BackoffPolicy
	batch        int
	scanInterval time.Duration
	now          func() time.Time
	logger       *slog.Logger
	metrics      Metrics
	binding      map[string]string
	scheduler    func(ctx context.Context, jobType string, event OutboxEvent) error

	mu        sync.RWMutex
	consumers map[string]OutboxConsumer
}

// NewOutboxDispatcher constructs an OutboxDispatcher.
func NewOutboxDispatcher(opts OutboxDispatcherOptions) *OutboxDispatcher {
	if opts.Now == nil {
		opts.Now = func() time.Time { return time.Now().UTC() }
	}
	if opts.Logger == nil {
		opts.Logger = slog.Default()
	}
	if opts.Metrics == nil {
		opts.Metrics = discardMetrics{}
	}
	if len(opts.Backoff.schedule) == 0 {
		opts.Backoff = DefaultOutboxBackoffPolicy()
	}
	if opts.Batch <= 0 {
		opts.Batch = defaultOutboxBatch
	}
	if opts.ScanInterval <= 0 {
		opts.ScanInterval = 5 * time.Second
	}
	return &OutboxDispatcher{
		store:        opts.Store,
		backoff:      opts.Backoff,
		batch:        opts.Batch,
		scanInterval: opts.ScanInterval,
		now:          opts.Now,
		logger:       opts.Logger,
		metrics:      opts.Metrics,
		binding:      opts.EventJobBinding,
		scheduler:    opts.JobScheduler,
		consumers:    map[string]OutboxConsumer{},
	}
}

// RegisterConsumer binds a consumer to an event_name. A runtime event without a
// registered consumer is never marked published (spec C-13a); test-only dry-run
// consumers must be registered explicitly here.
func (d *OutboxDispatcher) RegisterConsumer(eventName string, consumer OutboxConsumer) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.consumers[eventName] = consumer
}

func (d *OutboxDispatcher) consumer(eventName string) (OutboxConsumer, bool) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	c, ok := d.consumers[eventName]
	return c, ok
}

// RunOnce performs a single scan + publish cycle and returns the number of rows
// published.
func (d *OutboxDispatcher) RunOnce(ctx context.Context) (int, error) {
	now := d.now()
	outcome, err := d.store.ProcessPendingBatch(ctx, now, d.batch, d.backoff, func(row OutboxRow) OutboxResult {
		start := d.now()
		res := d.process(ctx, row)
		result := "published"
		if !res.Published {
			result = "failed"
			d.metrics.ObserveOutboxFailure()
		}
		d.metrics.ObserveOutboxPublish(result, d.now().Sub(start))
		return res
	})
	if pending, perr := d.store.CountPending(ctx); perr == nil {
		d.metrics.SetOutboxPending(float64(pending))
	}
	if err != nil {
		return outcome.Published, err
	}
	return outcome.Published, nil
}

func (d *OutboxDispatcher) process(ctx context.Context, row OutboxRow) OutboxResult {
	consumer, ok := d.consumer(row.EventName)
	if !ok {
		// Missing consumer: never ack. The row stays pending (retry) until it
		// dead-letters, surfacing the gap instead of silently dropping it.
		d.logger.WarnContext(ctx, "outbox event has no registered consumer",
			slog.String("event_name", row.EventName),
			slog.String("event_id", row.EventID),
		)
		return OutboxResult{Published: false, ErrorCode: "OUTBOX_NO_CONSUMER", ErrorMessage: "no consumer registered for event"}
	}

	traceID := traceIDFromPayload(row.Payload)
	hctx := withTraceID(ctx, traceID)
	logger := d.logger
	if traceID != "" {
		logger = logger.With(slog.String("trace_id", traceID))
	} else {
		d.logger.WarnContext(ctx, "outbox event missing trace_id; publishing without trace context",
			slog.String("event_name", row.EventName),
			slog.String("event_id", row.EventID),
		)
	}

	event := OutboxEvent{
		EventID:       row.EventID,
		EventName:     row.EventName,
		AggregateType: row.AggregateType,
		AggregateID:   row.AggregateID,
		Payload:       row.Payload,
	}
	if err := consumer.Consume(hctx, event); err != nil {
		// Redaction red line (spec §4.4 / O-1): never persist the raw consumer
		// error (it may embed provider response / prompt / answer text). Persist
		// only a stable redacted summary.
		logger.WarnContext(hctx, "outbox consumer failed", slog.String("event_name", row.EventName))
		return OutboxResult{Published: false, ErrorCode: "OUTBOX_CONSUMER_FAILED", ErrorMessage: "consumer rejected event"}
	}

	d.maybeScheduleFollowUp(hctx, event, logger)
	logger.InfoContext(hctx, "outbox event published", slog.String("event_name", row.EventName), slog.String("event_id", row.EventID))
	return OutboxResult{Published: true}
}

// maybeScheduleFollowUp enqueues a follow-up async job for events bound to a
// job_type, unless that job_type is source_event_only (spec D-7): those jobs are
// created by the producing transaction, so the dispatcher must not create a
// second one.
func (d *OutboxDispatcher) maybeScheduleFollowUp(ctx context.Context, event OutboxEvent, logger *slog.Logger) {
	if len(d.binding) == 0 || d.scheduler == nil {
		return
	}
	jobType, ok := d.binding[event.EventName]
	if !ok {
		return
	}
	if sharedjobs.IsSourceEventOnly(sharedjobs.JobType(jobType)) {
		logger.InfoContext(ctx, "skipping follow-up job for source_event_only event",
			slog.String("event_name", event.EventName),
			slog.String("job_type", jobType),
		)
		return
	}
	if err := d.scheduler(ctx, jobType, event); err != nil {
		logger.WarnContext(ctx, "outbox follow-up job scheduling failed",
			slog.String("event_name", event.EventName),
			slog.String("job_type", jobType),
		)
	}
}

// loop runs RunOnce on the scan interval until ctx is cancelled. It is started
// by Runtime.Start and stopped by Runtime.Shutdown (spec D-8 step d).
func (d *OutboxDispatcher) loop(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	ticker := time.NewTicker(d.scanInterval)
	defer ticker.Stop()
	for {
		// Drain available work immediately, then wait for the next tick.
		for {
			published, err := d.RunOnce(ctx)
			if err != nil {
				d.logger.WarnContext(ctx, "outbox dispatcher run failed", slog.String("error", err.Error()))
				break
			}
			if published == 0 {
				break
			}
			if ctx.Err() != nil {
				return
			}
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}
