package runner

import (
	"context"
	"log/slog"
	"time"
)

// ReaperOptions configures a Reaper.
type ReaperOptions struct {
	Store        LeaseStore
	JobTypes     []string
	LeaseTimeout time.Duration
	Now          func() time.Time
	Logger       *slog.Logger
	Metrics      Metrics
}

// Reaper requeues async_jobs rows whose lease has expired (the owning process
// crashed or was killed mid-handler). It covers every registered job_type
// (spec D-5) and never increments attempts: a lease timeout is infrastructure
// recovery, not a business failure.
type Reaper struct {
	store        LeaseStore
	jobTypes     []string
	leaseTimeout time.Duration
	now          func() time.Time
	logger       *slog.Logger
	metrics      Metrics
}

// NewReaper constructs a Reaper.
func NewReaper(opts ReaperOptions) *Reaper {
	if opts.Now == nil {
		opts.Now = func() time.Time { return time.Now().UTC() }
	}
	if opts.Logger == nil {
		opts.Logger = slog.Default()
	}
	if opts.Metrics == nil {
		opts.Metrics = nopMetrics{}
	}
	return &Reaper{
		store:        opts.Store,
		jobTypes:     opts.JobTypes,
		leaseTimeout: opts.LeaseTimeout,
		now:          opts.Now,
		logger:       opts.Logger,
		metrics:      opts.Metrics,
	}
}

// RunOnce reclaims expired leases once and returns the number of rows requeued.
func (r *Reaper) RunOnce(ctx context.Context) (int64, error) {
	if r == nil || len(r.jobTypes) == 0 {
		return 0, nil
	}
	now := r.now()
	reclaimed, err := r.store.ReclaimExpiredLeases(ctx, r.jobTypes, now.Add(-r.leaseTimeout), now)
	if err != nil {
		return 0, err
	}
	if reclaimed > 0 {
		r.metrics.ObserveReaped("", reclaimed)
		r.logger.InfoContext(ctx, "runner.reaper reclaimed expired leases",
			slog.Int64("reclaimed", reclaimed),
		)
	}
	return reclaimed, nil
}
