package main

import (
	"context"
	"io"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/platform/config"
	"github.com/monshunter/easyinterview/backend/internal/runner"
	"github.com/monshunter/easyinterview/backend/internal/shared/jobs"
)

// TestMain_SingleRuntimeShutdown proves spec C-16 / D-1 / D-8: every executable
// canonical job_type is registered on a single runner.Runtime, a single
// Shutdown call drains the whole kernel, and the two reserved-but-unregistered
// job types (privacy_export, jd_match_search) are not handled.
func TestMain_SingleRuntimeShutdown(t *testing.T) {
	kernel := runner.New(runner.Options{Store: runner.NewSQLStore(nil), Config: testRunnerConfig()})

	// Mirror cmd/api wiring: register the nine executable job types on one
	// kernel via the registerRunnerHandlers aggregator.
	noop := runner.JobHandlerFunc(func(context.Context, runner.ClaimedJob) runner.JobOutcome {
		return runner.JobOutcome{Succeeded: true}
	})
	executable := []jobs.JobType{
		jobs.JobTypeTargetImport,
		jobs.JobTypeSourceRefresh,
		jobs.JobTypePrivacyDelete,
		jobs.JobTypeDebriefGenerate,
		jobs.JobTypeResumeParse,
		jobs.JobTypeResumeTailor,
		jobs.JobTypeReportGenerate,
		jobs.JobTypeJdMatchAgentScan,
		jobs.JobTypeEmailDispatch,
	}
	handlers := map[string]runner.Handler{}
	for _, jt := range executable {
		handlers[string(jt)] = noop
	}
	registerRunnerHandlers(kernel, handlers)

	for _, jt := range executable {
		if !kernel.Handles(string(jt)) {
			t.Fatalf("single runtime does not handle %s", jt)
		}
	}
	// Reserved job types must never be registered by this plan.
	for _, reserved := range []jobs.JobType{jobs.JobTypePrivacyExport, jobs.JobTypeJdMatchSearch} {
		if kernel.Handles(string(reserved)) {
			t.Fatalf("single runtime must not handle reserved job type %s", reserved)
		}
	}

	kernel.Start(context.Background())
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := kernel.Shutdown(shutdownCtx); err != nil {
		t.Fatalf("single runtime Shutdown: %v", err)
	}
}

func TestMainRunnerKernelDrivesOutboxDispatcher(t *testing.T) {
	store := &recordingOutboxStore{processed: make(chan struct{})}
	kernel, err := buildRunnerKernel(runnerKernelOptions{
		Async:       testAsyncConfig(),
		Logger:      slog.New(slog.NewTextHandler(io.Discard, nil)),
		OutboxStore: store,
	})
	if err != nil {
		t.Fatalf("buildRunnerKernel: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	kernel.Start(ctx)
	select {
	case <-store.processed:
	case <-time.After(time.Second):
		t.Fatal("runner kernel did not start the attached outbox dispatcher")
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer shutdownCancel()
	if err := kernel.Shutdown(shutdownCtx); err != nil {
		t.Fatalf("kernel shutdown: %v", err)
	}
}

func testAsyncConfig() config.AsyncConfig {
	return config.AsyncConfig{
		QueueWeights: config.AsyncQueueWeights{
			Critical: 6,
			Default:  3,
			Low:      1,
		},
		LeaseTimeoutSeconds:   300,
		ShutdownGraceSeconds:  2,
		ReaperIntervalSeconds: 60,
		ScanIntervalSeconds:   1,
	}
}

type recordingOutboxStore struct {
	once      sync.Once
	processed chan struct{}
}

func (s *recordingOutboxStore) ProcessPendingBatch(context.Context, time.Time, int, runner.BackoffPolicy, func(runner.OutboxRow) runner.OutboxResult) (runner.OutboxBatchOutcome, error) {
	s.once.Do(func() { close(s.processed) })
	return runner.OutboxBatchOutcome{}, nil
}

func (s *recordingOutboxStore) CountPending(context.Context) (int64, error) {
	return 0, nil
}
