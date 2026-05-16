package review

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

type RunnerOptions struct {
	Store        Store
	Service      ReportService
	PollInterval time.Duration
	Now          func() time.Time
	Logger       *slog.Logger
}

type Runner struct {
	store        Store
	service      ReportService
	pollInterval time.Duration
	now          func() time.Time
	logger       *slog.Logger

	startOnce sync.Once
	stopOnce  sync.Once
	cancel    context.CancelFunc
	wg        sync.WaitGroup
}

func NewRunner(opts RunnerOptions) *Runner {
	if opts.PollInterval <= 0 {
		opts.PollInterval = 250 * time.Millisecond
	}
	if opts.Now == nil {
		opts.Now = func() time.Time { return time.Now().UTC() }
	}
	if opts.Logger == nil {
		opts.Logger = slog.Default()
	}
	if opts.Service == nil {
		opts.Service = NewService()
	}
	return &Runner{
		store:        opts.Store,
		service:      opts.Service,
		pollInterval: opts.PollInterval,
		now:          opts.Now,
		logger:       opts.Logger,
	}
}

func (r *Runner) Start(ctx context.Context) {
	r.startOnce.Do(func() {
		ctx, r.cancel = context.WithCancel(ctx)
		r.wg.Add(1)
		go r.runWorker(ctx, "report-runner-0")
	})
}

func (r *Runner) Stop(ctx context.Context) error {
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

func (r *Runner) RunOnce(ctx context.Context) (bool, error) {
	if r == nil || r.store == nil {
		return false, fmt.Errorf("review runner store is nil")
	}
	now := r.now()
	job, ok, err := r.store.LeaseAsyncJob(ctx, ReportGenerateJobType, now)
	if err != nil {
		return false, err
	}
	if !ok {
		return false, nil
	}
	if err := r.store.UpdateFeedbackReportStatus(ctx, ReportStatusUpdate{
		ReportID: job.ResourceID,
		From:     sharedtypes.ReportStatusQueued,
		To:       sharedtypes.ReportStatusGenerating,
		Now:      now,
	}); err != nil {
		return true, err
	}
	outcome := r.service.GenerateReport(ctx, job)
	if outcome.AsyncJobFinalized {
		return true, nil
	}
	if outcome.Succeeded {
		return true, r.store.UpdateAsyncJobSucceeded(ctx, job.JobID, r.now())
	}
	return true, r.store.UpdateAsyncJobFailed(ctx, AsyncJobFailure{
		JobID:       job.JobID,
		Retryable:   outcome.Retryable,
		ErrorCode:   outcome.ErrorCode,
		Error:       outcome.ErrorMessage,
		AvailableAt: r.now().Add(ComputeReportFailureBackoff(job.Attempts)),
		Now:         r.now(),
	})
}

func (r *Runner) runWorker(ctx context.Context, workerID string) {
	defer r.wg.Done()
	timer := time.NewTimer(r.pollInterval)
	defer timer.Stop()
	for {
		if ctx.Err() != nil {
			return
		}
		processed, err := r.RunOnce(ctx)
		if err != nil && !errors.Is(err, context.Canceled) {
			r.logger.WarnContext(ctx, "review.runner poll failed",
				slog.String("worker", workerID),
				slog.String("error", err.Error()),
			)
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
		timer.Reset(r.pollInterval)
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
		}
	}
}
