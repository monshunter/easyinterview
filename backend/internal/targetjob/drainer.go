package targetjob

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"
)

// JobHandler is the contract every async_jobs job_type implements. The
// drainer takes care of claim / finalize bookkeeping; handlers focus on
// the per-job side effects and translate failures into JobOutcome values.
type JobHandler interface {
	Handle(ctx context.Context, job ClaimedJob) JobOutcome
}

// JobHandlerFunc is the conventional adapter for inline handler bodies.
type JobHandlerFunc func(ctx context.Context, job ClaimedJob) JobOutcome

// Handle satisfies JobHandler.
func (f JobHandlerFunc) Handle(ctx context.Context, job ClaimedJob) JobOutcome {
	return f(ctx, job)
}

// DrainerOptions configures a Drainer instance. Handlers maps async_jobs
// job_type values to per-type handler implementations; only the keys
// listed are claimed by this drainer (so unrelated job types do not get
// stolen by a partial deployment).
type DrainerOptions struct {
	Store        Store
	Handlers     map[string]JobHandler
	Workers      int
	PollInterval time.Duration
	Now          func() time.Time
	// Logger is optional. When nil, slog.Default is used.
	Logger *slog.Logger
}

// Drainer is the in-process target_import / source_refresh job runner.
// Spec D-5 / plan 4.1 require this to live inside the cmd/api process
// without a separate worker binary; the structure mirrors backend-auth's
// BackgroundMailDispatcher with an added DB-backed claim path so jobs
// survive process restarts.
type Drainer struct {
	store        Store
	handlers     map[string]JobHandler
	jobTypes     []string
	workers      int
	pollInterval time.Duration
	now          func() time.Time
	logger       *slog.Logger

	startOnce sync.Once
	stopOnce  sync.Once
	cancel    context.CancelFunc
	wg        sync.WaitGroup
}

// NewDrainer constructs a Drainer. Handlers are required.
func NewDrainer(opts DrainerOptions) *Drainer {
	if opts.Workers <= 0 {
		opts.Workers = 2
	}
	if opts.PollInterval <= 0 {
		opts.PollInterval = 250 * time.Millisecond
	}
	if opts.Now == nil {
		opts.Now = func() time.Time { return time.Now().UTC() }
	}
	if opts.Logger == nil {
		opts.Logger = slog.Default()
	}
	jobTypes := make([]string, 0, len(opts.Handlers))
	for k := range opts.Handlers {
		jobTypes = append(jobTypes, k)
	}
	return &Drainer{
		store:        opts.Store,
		handlers:     opts.Handlers,
		jobTypes:     jobTypes,
		workers:      opts.Workers,
		pollInterval: opts.PollInterval,
		now:          opts.Now,
		logger:       opts.Logger,
	}
}

// Start launches the worker pool. Calling Start twice is a no-op.
func (d *Drainer) Start(ctx context.Context) {
	d.startOnce.Do(func() {
		ctx, d.cancel = context.WithCancel(ctx)
		for i := 0; i < d.workers; i++ {
			d.wg.Add(1)
			go d.runWorker(ctx, i)
		}
	})
}

// RunOnce performs a single claim+handle cycle synchronously. Handy for
// tests that want deterministic stepping without timer races; it returns
// true when a job was processed.
func (d *Drainer) RunOnce(ctx context.Context) (bool, error) {
	job, ok, err := d.store.ClaimNextAsyncJob(ctx, d.jobTypes, d.now())
	if err != nil {
		return false, err
	}
	if !ok {
		return false, nil
	}
	d.dispatch(ctx, job)
	return true, nil
}

func (d *Drainer) Handles(jobType string) bool {
	if d == nil {
		return false
	}
	_, ok := d.handlers[jobType]
	return ok
}

// Shutdown cancels the worker context and waits for in-flight jobs to
// finish (or for the supplied context to expire, whichever comes first).
// Calling Shutdown more than once is a no-op.
func (d *Drainer) Shutdown(ctx context.Context) error {
	d.stopOnce.Do(func() {
		if d.cancel != nil {
			d.cancel()
		}
	})
	done := make(chan struct{})
	go func() {
		d.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (d *Drainer) runWorker(ctx context.Context, _ int) {
	defer d.wg.Done()
	timer := time.NewTimer(d.pollInterval)
	defer timer.Stop()
	for {
		if ctx.Err() != nil {
			return
		}
		processed, err := d.RunOnce(ctx)
		if err != nil && !errors.Is(err, context.Canceled) {
			d.logger.WarnContext(ctx, "targetjob.drainer claim failed", slog.String("error", err.Error()))
		}
		if processed {
			// Try again immediately when there is work to do.
			continue
		}
		// Reset the timer for the next idle cycle.
		if !timer.Stop() {
			select {
			case <-timer.C:
			default:
			}
		}
		timer.Reset(d.pollInterval)
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
		}
	}
}

func (d *Drainer) dispatch(ctx context.Context, job ClaimedJob) {
	handler, ok := d.handlers[job.JobType]
	if !ok {
		// Unknown handler — finalize as a non-retryable failure so the
		// row stops cycling instead of getting requeued forever.
		_ = d.store.FinalizeAsyncJob(ctx, job.JobID, JobOutcome{
			ErrorCode:    "TARGET_IMPORT_FAILED",
			ErrorMessage: "no handler registered for job_type " + job.JobType,
		}, d.now())
		return
	}
	outcome := handler.Handle(ctx, job)
	if err := d.store.FinalizeAsyncJob(ctx, job.JobID, outcome, d.now()); err != nil {
		d.logger.ErrorContext(ctx, "targetjob.drainer finalize failed",
			slog.String("error", err.Error()),
			slog.String("job_id", job.JobID),
			slog.String("job_type", job.JobType),
		)
	}
}
