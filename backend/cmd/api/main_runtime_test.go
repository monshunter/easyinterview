package main

import (
	"context"
	"testing"
	"time"

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
