package review

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

type ReaperOptions struct {
	Store        Store
	LeaseTimeout time.Duration
	Interval     time.Duration
	Now          func() time.Time
	Logger       *slog.Logger
}

type Reaper struct {
	store        Store
	leaseTimeout time.Duration
	interval     time.Duration
	now          func() time.Time
	logger       *slog.Logger

	startOnce sync.Once
	stopOnce  sync.Once
	cancel    context.CancelFunc
	wg        sync.WaitGroup
}

func NewReaper(opts ReaperOptions) *Reaper {
	if opts.LeaseTimeout <= 0 {
		opts.LeaseTimeout = 5 * time.Minute
	}
	if opts.Interval <= 0 {
		opts.Interval = time.Minute
	}
	if opts.Now == nil {
		opts.Now = func() time.Time { return time.Now().UTC() }
	}
	if opts.Logger == nil {
		opts.Logger = slog.Default()
	}
	return &Reaper{
		store:        opts.Store,
		leaseTimeout: opts.LeaseTimeout,
		interval:     opts.Interval,
		now:          opts.Now,
		logger:       opts.Logger,
	}
}

func (r *Reaper) RunOnce(ctx context.Context) (int64, error) {
	if r == nil || r.store == nil {
		return 0, fmt.Errorf("review reaper store is nil")
	}
	now := r.now()
	return r.store.ReclaimExpiredLeases(ctx, ReportGenerateJobType, now.Add(-r.leaseTimeout), now)
}

func (r *Reaper) Start(ctx context.Context) {
	r.startOnce.Do(func() {
		ctx, r.cancel = context.WithCancel(ctx)
		r.wg.Add(1)
		go r.loop(ctx)
	})
}

func (r *Reaper) Stop(ctx context.Context) error {
	r.stopOnce.Do(func() {
		if r.cancel != nil {
			r.cancel()
		}
	})
	done := make(chan struct{})
	go func() {
		r.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (r *Reaper) loop(ctx context.Context) {
	defer r.wg.Done()
	timer := time.NewTimer(r.interval)
	defer timer.Stop()
	for {
		if _, err := r.RunOnce(ctx); err != nil && !errors.Is(err, context.Canceled) {
			r.logger.WarnContext(ctx, "review.reaper reclaim failed", slog.String("error", err.Error()))
		}
		if !timer.Stop() {
			select {
			case <-timer.C:
			default:
			}
		}
		timer.Reset(r.interval)
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
		}
	}
}
