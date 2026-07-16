package main

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient"
	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient/observability"
	"github.com/monshunter/easyinterview/backend/internal/platform/config"
	"github.com/monshunter/easyinterview/backend/internal/runner"
	"github.com/monshunter/easyinterview/backend/internal/shared/jobs"
	"github.com/monshunter/easyinterview/backend/internal/testsupport"
)

// TestMain_SingleRuntimeShutdown proves spec C-16 / D-1 / D-8: every executable
// canonical job_type is registered on a single runner.Runtime, a single
// Shutdown call drains the whole kernel, and the two reserved-but-unregistered
// job types (privacy_export) are not handled.
func TestMain_SingleRuntimeShutdown(t *testing.T) {
	kernel := runner.New(runner.Options{Store: runner.NewSQLStore(nil), Config: testRunnerConfig()})

	// Mirror cmd/api wiring: register the nine executable job types on one
	// kernel via the registerRunnerHandlers aggregator.
	noop := runner.JobHandlerFunc(func(context.Context, runner.ClaimedJob) runner.JobOutcome {
		return runner.JobOutcome{Succeeded: true}
	})
	executable := []jobs.JobType{
		jobs.JobTypeTargetImport,
		jobs.JobTypePrivacyDelete,
		jobs.JobTypeResumeParse,
		jobs.JobTypeResumeTailor,
		jobs.JobTypeReportGenerate,
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
	for _, reserved := range []jobs.JobType{jobs.JobTypePrivacyExport} {
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

func TestMainReportRuntimeHandlerRegistersIntoStartedKernel(t *testing.T) {
	dir := t.TempDir()
	promptsDir, rubricsDir := testsupport.ConfigRoots(t)
	writeAPIFile(t, filepath.Join(dir, "config.yaml"), `
runtime:
  appVersion: "1.2.3"
  defaultUiLanguage: zh-CN
ai:
  promptsDir: "`+promptsDir+`"
  rubricsDir: "`+rubricsDir+`"
`)
	loader, err := config.Load(config.Options{AppEnv: "test", ConfigDir: dir})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	report, err := buildReportRuntime(loader, nil, &apiNoopAIClient{}, nil)
	if err != nil {
		t.Fatalf("buildReportRuntime: %v", err)
	}

	kernel := runner.New(runner.Options{Store: runner.NewSQLStore(nil), Config: testRunnerConfig()})
	registerRunnerHandlers(kernel, report.Handlers)
	if !kernel.Handles(string(jobs.JobTypeReportGenerate)) {
		t.Fatalf("production report runtime did not register %s", jobs.JobTypeReportGenerate)
	}

	ctx, cancel := context.WithCancel(context.Background())
	kernel.Start(ctx)
	cancel()
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer shutdownCancel()
	if err := kernel.Shutdown(shutdownCtx); err != nil {
		t.Fatalf("report runtime kernel shutdown: %v", err)
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

func TestAIRawCaptureRuntimeUsesConfigDirParentAndHardensPermissions(t *testing.T) {
	root := t.TempDir()
	configDir := filepath.Join(root, "config")
	if err := os.MkdirAll(configDir, 0o700); err != nil {
		t.Fatalf("mkdir config: %v", err)
	}
	writeAPIFile(t, filepath.Join(configDir, "config.yaml"), `
ai:
  debugCaptureRawIO: true
  debugRawIOPath: .test-output/local-dev/ai-raw.ndjson
`)
	loader, err := config.LoadCanonical(config.CanonicalOptions{AppEnv: "test", ConfigDir: configDir})
	if err != nil {
		t.Fatalf("LoadCanonical: %v", err)
	}

	wrap, err := newAPIRawCaptureObservedClient(t, loader, &apiTailorTaskRuns{})
	if err != nil {
		t.Fatalf("construct observed AI client: %v", err)
	}
	payload := apiRawCapturePayload()
	if _, _, err := wrap.Complete(context.Background(), "practice.chat.default", payload); err != nil {
		t.Fatalf("Complete: %v", err)
	}

	wantPath := filepath.Join(root, ".test-output", "local-dev", "ai-raw.ndjson")
	raw, err := os.ReadFile(wantPath)
	if err != nil {
		t.Fatalf("raw capture was not opened at ConfigDir-parent anchor %s: %v", wantPath, err)
	}
	if !strings.Contains(string(raw), `"recordVersion":"ai.complete.raw.v1"`) {
		t.Fatalf("raw capture is not versioned NDJSON: %q", raw)
	}
	assertAPIRawCaptureMode(t, filepath.Dir(wantPath), 0o700)
	assertAPIRawCaptureMode(t, wantPath, 0o600)
}

func TestAIRawCaptureRuntimeTightensExisting0644FileAnd0755Parent(t *testing.T) {
	root := t.TempDir()
	configDir := filepath.Join(root, "config")
	rawDir := filepath.Join(root, ".test-output", "local-dev")
	rawPath := filepath.Join(rawDir, "ai-raw.ndjson")
	if err := os.MkdirAll(configDir, 0o700); err != nil {
		t.Fatalf("mkdir config: %v", err)
	}
	if err := os.MkdirAll(rawDir, 0o755); err != nil {
		t.Fatalf("mkdir raw dir: %v", err)
	}
	if err := os.Chmod(rawDir, 0o755); err != nil {
		t.Fatalf("chmod raw dir: %v", err)
	}
	if err := os.WriteFile(rawPath, []byte(""), 0o644); err != nil {
		t.Fatalf("seed raw file: %v", err)
	}
	if err := os.Chmod(rawPath, 0o644); err != nil {
		t.Fatalf("chmod raw file: %v", err)
	}
	writeAPIFile(t, filepath.Join(configDir, "config.yaml"), `
ai:
  debugCaptureRawIO: true
  debugRawIOPath: .test-output/local-dev/ai-raw.ndjson
`)
	loader, err := config.LoadCanonical(config.CanonicalOptions{AppEnv: "test", ConfigDir: configDir})
	if err != nil {
		t.Fatalf("LoadCanonical: %v", err)
	}
	if _, err := newAPIRawCaptureObservedClient(t, loader, &apiTailorTaskRuns{}); err != nil {
		t.Fatalf("construct observed AI client: %v", err)
	}
	assertAPIRawCaptureMode(t, rawDir, 0o700)
	assertAPIRawCaptureMode(t, rawPath, 0o600)
}

func TestAIRawCaptureRuntimeRejectsSymlinkAndNonRegularPaths(t *testing.T) {
	tests := []struct {
		name  string
		setup func(t *testing.T, root, rawPath string)
	}{
		{
			name: "symlink path component",
			setup: func(t *testing.T, root, _ string) {
				t.Helper()
				outside := filepath.Join(root, "outside")
				if err := os.MkdirAll(outside, 0o700); err != nil {
					t.Fatalf("mkdir outside: %v", err)
				}
				if err := os.Symlink(outside, filepath.Join(root, ".test-output")); err != nil {
					t.Fatalf("symlink component: %v", err)
				}
			},
		},
		{
			name: "symlink target",
			setup: func(t *testing.T, root, rawPath string) {
				t.Helper()
				if err := os.MkdirAll(filepath.Dir(rawPath), 0o700); err != nil {
					t.Fatalf("mkdir raw parent: %v", err)
				}
				outside := filepath.Join(root, "outside.ndjson")
				if err := os.WriteFile(outside, nil, 0o600); err != nil {
					t.Fatalf("write outside target: %v", err)
				}
				if err := os.Symlink(outside, rawPath); err != nil {
					t.Fatalf("symlink target: %v", err)
				}
			},
		},
		{
			name: "non regular target",
			setup: func(t *testing.T, _ string, rawPath string) {
				t.Helper()
				if err := os.MkdirAll(rawPath, 0o700); err != nil {
					t.Fatalf("mkdir directory target: %v", err)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			configDir := filepath.Join(root, "config")
			rawPath := filepath.Join(root, ".test-output", "local-dev", "ai-raw.ndjson")
			if err := os.MkdirAll(configDir, 0o700); err != nil {
				t.Fatalf("mkdir config: %v", err)
			}
			tt.setup(t, root, rawPath)
			writeAPIFile(t, filepath.Join(configDir, "config.yaml"), `
ai:
  debugCaptureRawIO: true
  debugRawIOPath: .test-output/local-dev/ai-raw.ndjson
`)
			loader, err := config.LoadCanonical(config.CanonicalOptions{AppEnv: "test", ConfigDir: configDir})
			if err != nil {
				t.Fatalf("LoadCanonical: %v", err)
			}
			if _, err := newAPIRawCaptureObservedClient(t, loader, &apiTailorTaskRuns{}); err == nil {
				t.Fatal("raw capture startup accepted a symlink/non-regular path")
			}
		})
	}
}

func TestResumeRuntimeRoutesTailorThroughObservedAIWithoutManualTaskRunWrite(t *testing.T) {
	mainSource, err := os.ReadFile("main.go")
	if err != nil {
		t.Fatalf("read main.go: %v", err)
	}
	start := strings.Index(string(mainSource), "func buildResumeRuntime(")
	end := strings.Index(string(mainSource)[start:], "\nfunc registryDirOrDefault(")
	if start < 0 || end < 0 {
		t.Fatal("cannot isolate buildResumeRuntime source")
	}
	body := string(mainSource)[start : start+end]
	if !strings.Contains(strings.Join(strings.Fields(body), " "), "AI: observedAI") {
		t.Fatal("resume-tailor does not receive the same observed AI client as resume parse")
	}
	if strings.Contains(body, "AITaskRuns:") {
		t.Fatal("resume-tailor still wires a manual ai_task_runs writer in addition to the decorator")
	}

	tailorSource, err := os.ReadFile(filepath.Join("..", "..", "internal", "resume", "jobs", "tailor.go"))
	if err != nil {
		t.Fatalf("read resume tailor handler: %v", err)
	}
	for _, duplicateSeam := range []string{"aiTaskRuns aiclient.AITaskRunWriter", "func (h *TailorHandler) writeTaskRun("} {
		if strings.Contains(string(tailorSource), duplicateSeam) {
			t.Fatalf("resume-tailor retains duplicate task-run seam %q", duplicateSeam)
		}
	}
}

func TestAIRawCaptureRuntimeOwnsOneProcessSharedRecorder(t *testing.T) {
	source, err := os.ReadFile("main.go")
	if err != nil {
		t.Fatalf("read main.go: %v", err)
	}
	text := string(source)
	if got := strings.Count(text, `loader.GetString("ai.debugRawIOPath")`); got != 1 {
		t.Fatalf("AI raw path must be consumed once at process startup, occurrences=%d", got)
	}
	if got := strings.Count(text, "observability.WithRawIOCapture("); got != 1 {
		t.Fatalf("one shared recorder option must feed all Complete wrappers, WithRawIOCapture occurrences=%d", got)
	}
	if strings.Contains(text, "observability.WithRawOutputDebugWriter(") {
		t.Fatal("legacy per-wrapper stderr/raw writer remains in API runtime")
	}
}

func TestAIRawCaptureWriteFailureUsesProcessStructuredLoggerWithoutChangingAIResult(t *testing.T) {
	previousLogger := slog.Default()
	var logOutput bytes.Buffer
	slog.SetDefault(slog.New(slog.NewJSONHandler(&logOutput, nil)))
	t.Cleanup(func() { slog.SetDefault(previousLogger) })

	inner := &apiResolvableAIClient{}
	wrap, err := observability.New(inner,
		aiObservabilityOptions(&apiTailorTaskRuns{}, inner.Resolver(), apiFailingRawCapture{})...,
	)
	if err != nil {
		t.Fatalf("construct observed AI client: %v", err)
	}

	response, _, err := wrap.Complete(context.Background(), "practice.chat.default", apiRawCapturePayload())
	if err != nil {
		t.Fatalf("raw capture failure changed AI result: %v", err)
	}
	if response.Content != `{"basics":{}}` {
		t.Fatalf("response content = %q, want provider result", response.Content)
	}

	logs := logOutput.String()
	if got := strings.Count(logs, observability.EventRawCaptureWriteFailed); got != 2 {
		t.Fatalf("process-visible raw capture warnings = %d, want request+response warnings; logs=%q", got, logs)
	}
	if !strings.Contains(logs, `"level":"WARN"`) {
		t.Fatalf("raw capture failure was not emitted at WARN level: %q", logs)
	}
	for _, forbidden := range []string{
		"raw capture storage unavailable at /private/secret/ai-raw.ndjson",
		"capture this Complete call",
	} {
		if strings.Contains(logs, forbidden) {
			t.Fatalf("process logger leaked raw capture detail %q: %q", forbidden, logs)
		}
	}
}

type apiFailingRawCapture struct{}

func (apiFailingRawCapture) RecordRawComplete(observability.RawCompleteRecord) error {
	return errors.New("raw capture storage unavailable at /private/secret/ai-raw.ndjson")
}

func newAPIRawCaptureObservedClient(t *testing.T, loader *config.Loader, runs aiclient.AITaskRunWriter) (*observability.Wrap, error) {
	t.Helper()
	inner := &apiResolvableAIClient{}
	capture, err := openAIRawCapture(loader)
	if err != nil {
		return nil, err
	}
	t.Cleanup(func() { _ = capture.Close() })
	return observability.New(inner, aiObservabilityOptions(runs, inner.Resolver(), capture)...)
}

func apiRawCapturePayload() aiclient.CompletePayload {
	return aiclient.CompletePayload{
		Messages: []aiclient.Message{{Role: "user", Content: "capture this Complete call"}},
		Metadata: aiclient.CallMetadata{
			FeatureKey:        "practice.session.chat",
			PromptVersion:     "p1",
			RubricVersion:     "r1",
			Language:          "en",
			FeatureFlag:       "none",
			DataSourceVersion: "registry.v1",
			TaskRun: aiclient.AITaskRunContext{
				Capability:   aiclient.AITaskRunTaskPracticeChat,
				ResourceType: aiclient.AITaskRunResourceTargetJob,
				ResourceID:   "0195f2d0-4a44-7fc2-8f77-1f9c4ce1ae9e",
			},
		},
	}
}

func assertAPIRawCaptureMode(t *testing.T, path string, want os.FileMode) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat %s: %v", path, err)
	}
	if got := info.Mode().Perm(); got != want {
		t.Errorf("%s mode = %#o, want %#o", path, got, want)
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
