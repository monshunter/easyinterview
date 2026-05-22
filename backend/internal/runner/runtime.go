package runner

import (
	"context"
	"errors"
	"log/slog"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

// Options configures a Runtime.
type Options struct {
	Store   LeaseStore
	Config  Config
	Backoff BackoffPolicy
	Now     func() time.Time
	Logger  *slog.Logger
	Metrics Metrics
}

// Runtime is the single in-process async job kernel (spec D-1). It owns the
// handler registry, lease/finalize bookkeeping, retry backoff, the lease loop,
// the reaper, graceful shutdown, and (Phase 3) the outbox dispatcher. Business
// state stays in domain handlers; the kernel never deserializes business
// payloads when finalizing.
type Runtime struct {
	store    LeaseStore
	config   Config
	backoff  BackoffPolicy
	now      func() time.Time
	logger   *slog.Logger
	metrics  Metrics

	mu       sync.RWMutex
	handlers map[string]Handler

	dispatcher *OutboxDispatcher

	startOnce     sync.Once
	stopping      atomic.Bool
	loopCancel    context.CancelFunc
	handlerCtx    context.Context
	handlerCancel context.CancelFunc
	wg            sync.WaitGroup
	inflight      sync.WaitGroup
}

// New constructs a Runtime. Backoff defaults to the spec D-4 policy.
func New(opts Options) *Runtime {
	if opts.Now == nil {
		opts.Now = func() time.Time { return time.Now().UTC() }
	}
	if opts.Logger == nil {
		opts.Logger = slog.Default()
	}
	if opts.Metrics == nil {
		opts.Metrics = nopMetrics{}
	}
	if len(opts.Backoff.schedule) == 0 {
		opts.Backoff = DefaultBackoffPolicy()
	}
	return &Runtime{
		store:    opts.Store,
		config:   opts.Config,
		backoff:  opts.Backoff,
		now:      opts.Now,
		logger:   opts.Logger,
		metrics:  opts.Metrics,
		handlers: make(map[string]Handler),
	}
}

// Register binds a handler to a job_type. Registering the same job_type twice
// replaces the prior handler. Register must be called before Start.
func (r *Runtime) Register(jobType string, handler Handler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.handlers[jobType] = handler
}

// SetOutboxDispatcher attaches an outbox dispatcher so the runtime drives its
// scan loop in Start and stops it during Shutdown (spec D-8 step d).
func (r *Runtime) SetOutboxDispatcher(d *OutboxDispatcher) {
	r.dispatcher = d
}

// Handles reports whether a handler is registered for jobType.
func (r *Runtime) Handles(jobType string) bool {
	if r == nil {
		return false
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.handlers[jobType]
	return ok
}

// jobTypeList returns every registered job_type (unordered).
func (r *Runtime) jobTypeList() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]string, 0, len(r.handlers))
	for jt := range r.handlers {
		out = append(out, jt)
	}
	return out
}

// leaseBuckets groups registered job_types by priority bucket and returns the
// buckets ordered by descending queue weight (spec D-9). The returned slices
// are sorted for deterministic claim order.
func (r *Runtime) leaseBuckets() [][]string {
	r.mu.RLock()
	byBucket := map[Priority][]string{}
	for jt := range r.handlers {
		p := PriorityForJobType(jt)
		byBucket[p] = append(byBucket[p], jt)
	}
	r.mu.RUnlock()

	order := append([]Priority{}, priorityOrder...)
	sort.SliceStable(order, func(i, j int) bool {
		return r.config.QueueWeights.weightFor(order[i]) > r.config.QueueWeights.weightFor(order[j])
	})
	out := make([][]string, 0, len(order))
	for _, p := range order {
		types := byBucket[p]
		if len(types) == 0 {
			continue
		}
		sort.Strings(types)
		out = append(out, types)
	}
	return out
}

// RunOnce performs a single priority-ordered claim+handle cycle and reports
// whether a job was processed. It is the synchronous test/integration driver
// required by spec D-13.
func (r *Runtime) RunOnce(ctx context.Context) (bool, error) {
	return r.runOnce(ctx, ctx)
}

func (r *Runtime) runOnce(ctx, handlerCtx context.Context) (bool, error) {
	if r.stopping.Load() {
		return false, nil
	}
	for _, jobTypes := range r.leaseBuckets() {
		job, ok, err := r.store.LeaseAsyncJob(ctx, jobTypes, r.now())
		if err != nil {
			return false, err
		}
		if !ok {
			continue
		}
		r.dispatch(handlerCtx, job)
		return true, nil
	}
	return false, nil
}

// ReapOnce reclaims expired leases across every registered job_type once.
func (r *Runtime) ReapOnce(ctx context.Context) (int64, error) {
	return r.newReaper().RunOnce(ctx)
}

func (r *Runtime) newReaper() *Reaper {
	return NewReaper(ReaperOptions{
		Store:        r.store,
		JobTypes:     r.jobTypeList(),
		LeaseTimeout: r.config.LeaseTimeout,
		Now:          r.now,
		Logger:       r.logger,
		Metrics:      r.metrics,
	})
}

func (r *Runtime) handlerFor(jobType string) (Handler, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	h, ok := r.handlers[jobType]
	return h, ok
}

func (r *Runtime) dispatch(ctx context.Context, job ClaimedJob) {
	r.inflight.Add(1)
	defer r.inflight.Done()

	handler, ok := r.handlerFor(job.JobType)
	now := r.now()
	if !ok {
		// No handler registered: finalize non-retryable so the row stops
		// cycling instead of being requeued forever.
		_ = r.store.FinalizeAsyncJob(ctx, job.JobID, JobOutcome{
			ErrorCode:    "RUNNER_NO_HANDLER",
			ErrorMessage: "no handler registered for job_type " + job.JobType,
		}, now, now)
		r.metrics.ObserveJobProcessed(job.JobType, "failed", 0)
		return
	}

	traceID := traceIDFromPayload(job.Payload)
	handlerCtx := withTraceID(ctx, traceID)
	logger := r.logger
	if traceID != "" {
		logger = logger.With(slog.String("trace_id", traceID))
	}

	start := r.now()
	outcome := handler.Handle(handlerCtx, job)
	duration := r.now().Sub(start)

	result := resultLabel(outcome, job.Attempts, job.MaxAttempts)
	logger.InfoContext(handlerCtx, "runner.handle completed",
		slog.String("job_type", job.JobType),
		slog.String("job_id", job.JobID),
		slog.Int64("attempts", int64(job.Attempts)),
		slog.String("outcome", result),
	)

	if !outcome.AsyncJobFinalized {
		availableAt := now
		if !outcome.Succeeded && outcome.Retryable {
			availableAt = now.Add(r.backoff.Next(job.Attempts))
		}
		if err := r.store.FinalizeAsyncJob(ctx, job.JobID, outcome, availableAt, now); err != nil {
			r.logger.ErrorContext(ctx, "runner.finalize failed",
				slog.String("error", err.Error()),
				slog.String("job_id", job.JobID),
				slog.String("job_type", job.JobType),
			)
		}
	}
	r.metrics.ObserveJobProcessed(job.JobType, result, duration)
}

func resultLabel(outcome JobOutcome, attempts, maxAttempts int32) string {
	switch {
	case outcome.Succeeded:
		return "succeeded"
	case !outcome.Retryable:
		return "failed"
	case attempts >= maxAttempts:
		return "dead"
	default:
		return "retried"
	}
}

// Start launches the lease loop, the reaper loop, and (when attached) the
// outbox dispatcher loop. Calling Start twice is a no-op.
func (r *Runtime) Start(ctx context.Context) {
	r.startOnce.Do(func() {
		loopCtx, loopCancel := context.WithCancel(ctx)
		r.loopCancel = loopCancel
		r.handlerCtx, r.handlerCancel = context.WithCancel(context.WithoutCancel(ctx))

		r.wg.Add(1)
		go r.leaseLoop(loopCtx)

		if r.config.ReaperInterval > 0 {
			r.wg.Add(1)
			go r.reaperLoop(loopCtx)
		}

		if r.dispatcher != nil {
			r.wg.Add(1)
			go r.dispatcher.loop(loopCtx, &r.wg)
		}
	})
}

func (r *Runtime) leaseLoop(ctx context.Context) {
	defer r.wg.Done()
	interval := r.config.ScanInterval
	if interval <= 0 {
		interval = 5 * time.Second
	}
	timer := time.NewTimer(interval)
	defer timer.Stop()
	for {
		if ctx.Err() != nil || r.stopping.Load() {
			return
		}
		processed, err := r.runOnce(ctx, r.handlerCtx)
		if err != nil && !errors.Is(err, context.Canceled) {
			r.logger.WarnContext(ctx, "runner.lease loop claim failed", slog.String("error", err.Error()))
		}
		if processed {
			continue
		}
		if !timer.Stop() {
			select {
			case <-timer.C:
			default:
			}
		}
		timer.Reset(interval)
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
		}
	}
}

func (r *Runtime) reaperLoop(ctx context.Context) {
	defer r.wg.Done()
	ticker := time.NewTicker(r.config.ReaperInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if _, err := r.ReapOnce(ctx); err != nil && !errors.Is(err, context.Canceled) {
				r.logger.WarnContext(ctx, "runner.reaper loop failed", slog.String("error", err.Error()))
			}
		}
	}
}

// Shutdown drains the kernel in the spec D-8 order: stop accepting new leases,
// wait for in-flight handlers (bounded by ctx), then stop the reaper and outbox
// dispatcher loops. If the grace deadline (ctx) expires while a handler is
// still running, it cancels the handler context and returns ctx.Err().
func (r *Runtime) Shutdown(ctx context.Context) error {
	r.stopping.Store(true)
	if r.loopCancel != nil {
		r.loopCancel()
	}

	done := make(chan struct{})
	go func() {
		r.inflight.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-ctx.Done():
		if r.handlerCancel != nil {
			r.handlerCancel()
		}
		return ctx.Err()
	}

	if r.handlerCancel != nil {
		r.handlerCancel()
	}

	loopsDone := make(chan struct{})
	go func() {
		r.wg.Wait()
		close(loopsDone)
	}()
	select {
	case <-loopsDone:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
