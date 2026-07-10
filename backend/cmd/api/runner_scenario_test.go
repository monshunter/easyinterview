package main

import (
	"io"
	"log/slog"
	"time"

	privacyrunner "github.com/monshunter/easyinterview/backend/internal/privacy/runner"
	resumejobs "github.com/monshunter/easyinterview/backend/internal/resume/jobs"
	"github.com/monshunter/easyinterview/backend/internal/runner"
	"github.com/monshunter/easyinterview/backend/internal/targetjob"
)

var (
	_ runner.Handler = (*targetjob.ParseExecutor)(nil)
	_ runner.Handler = (*targetjob.SourceRefreshHandler)(nil)
	_ runner.Handler = (*privacyrunner.PrivacyDeleteHandler)(nil)
	_ runner.Handler = (*resumejobs.ParseHandler)(nil)
	_ runner.Handler = (*resumejobs.TailorHandler)(nil)
)

func newScenarioJobRuntime(store runner.LeaseStore, now func() time.Time, handlers map[string]runner.Handler) *runner.Runtime {
	runtime := runner.New(runner.Options{
		Store: store,
		Config: runner.Config{
			ScanInterval:   10 * time.Millisecond,
			LeaseTimeout:   5 * time.Minute,
			ReaperInterval: time.Minute,
			ShutdownGrace:  2 * time.Second,
			QueueWeights:   runner.QueueWeights{Critical: 6, Default: 3, Low: 1},
		},
		Now:    now,
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	})
	for jobType, handler := range handlers {
		runtime.Register(jobType, handler)
	}
	return runtime
}
