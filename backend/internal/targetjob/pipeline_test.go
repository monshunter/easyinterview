package targetjob_test

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient"
	"github.com/monshunter/easyinterview/backend/internal/targetjob"
	"github.com/monshunter/easyinterview/backend/internal/targetjob/urlfetch"
)

// ----- pipeline test fakes -----

type pipelineFakeStore struct {
	queuedJobs []targetjob.ClaimedJob
	finalize   []targetjob.JobOutcome
	finalizeID []string

	// parse executor side
	target              targetjob.TargetJobRecord
	sources             []targetjob.SourceRecord
	getErr              error
	applyResultIn       *targetjob.ApplyParseResultInput
	applyResultErr      error
	parsedOutboxPayload []byte
	failedOutboxPayload []byte
	sourceRefreshCalled bool
	sourceFreshnessUpd  string
	sourceSnapshotURL   string
	sourceSnapshotText  string
	sourceSnapshotAt    *time.Time
	updateAnalysisFail  int
	pollMu              chan struct{}
}

func (s *pipelineFakeStore) ImportTargetJob(context.Context, targetjob.ImportTargetJobInput) (targetjob.ImportTargetJobResult, error) {
	return targetjob.ImportTargetJobResult{}, nil
}
func (s *pipelineFakeStore) InsertTargetJob(context.Context, targetjob.TargetJobRecord) error {
	return nil
}
func (s *pipelineFakeStore) InsertTargetJobSource(context.Context, targetjob.SourceRecord) error {
	return nil
}
func (s *pipelineFakeStore) GetTargetJobByUser(context.Context, string, string) (targetjob.TargetJobRecord, []targetjob.RequirementRecord, []targetjob.SourceRecord, error) {
	return targetjob.TargetJobRecord{}, nil, nil, nil
}
func (s *pipelineFakeStore) ListTargetJobsForUser(context.Context, string, targetjob.ListFilter) (targetjob.ListResult, error) {
	return targetjob.ListResult{}, nil
}
func (s *pipelineFakeStore) LookupUpdateDedupe(context.Context, string, string) (targetjob.TargetJobRecord, []targetjob.RequirementRecord, bool, error) {
	return targetjob.TargetJobRecord{}, nil, false, nil
}
func (s *pipelineFakeStore) UpdateTargetJobLifecycle(context.Context, string, string, targetjob.UpdateLifecycleFields, time.Time) (targetjob.TargetJobRecord, error) {
	return targetjob.TargetJobRecord{}, nil
}
func (s *pipelineFakeStore) ApplyParseResult(_ context.Context, in targetjob.ApplyParseResultInput) error {
	cp := in
	s.applyResultIn = &cp
	return s.applyResultErr
}
func (s *pipelineFakeStore) UpdateSourceFreshness(_ context.Context, _ string, status targetjob.FreshnessStatus, _ time.Time) error {
	s.sourceFreshnessUpd = string(status)
	return nil
}
func (s *pipelineFakeStore) UpdateSourceSnapshot(_ context.Context, _ string, sanitizedURL string, snapshotText string, fetchedAt time.Time, _ time.Time) error {
	s.sourceSnapshotURL = sanitizedURL
	s.sourceSnapshotText = snapshotText
	cp := fetchedAt
	s.sourceSnapshotAt = &cp
	return nil
}
func (s *pipelineFakeStore) LookupFileAttachmentForUser(context.Context, string, string) (targetjob.FileAttachmentRecord, error) {
	return targetjob.FileAttachmentRecord{}, nil
}
func (s *pipelineFakeStore) ClaimNextAsyncJob(_ context.Context, _ []string, _ time.Time) (targetjob.ClaimedJob, bool, error) {
	if s.pollMu != nil {
		<-s.pollMu
	}
	if len(s.queuedJobs) == 0 {
		return targetjob.ClaimedJob{}, false, nil
	}
	job := s.queuedJobs[0]
	s.queuedJobs = s.queuedJobs[1:]
	return job, true, nil
}
func (s *pipelineFakeStore) FinalizeAsyncJob(_ context.Context, jobID string, outcome targetjob.JobOutcome, _ time.Time) error {
	s.finalizeID = append(s.finalizeID, jobID)
	s.finalize = append(s.finalize, outcome)
	return nil
}
func (s *pipelineFakeStore) EnqueueSourceRefresh(context.Context, string, string, time.Time) error {
	s.sourceRefreshCalled = true
	return nil
}
func (s *pipelineFakeStore) WriteParseFailedOutbox(_ context.Context, _ string, _ string, payload []byte, _ time.Time) error {
	s.failedOutboxPayload = append([]byte{}, payload...)
	return nil
}
func (s *pipelineFakeStore) WriteTargetParsedOutbox(_ context.Context, _ string, _ string, payload []byte, _ time.Time) error {
	s.parsedOutboxPayload = append([]byte{}, payload...)
	return nil
}
func (s *pipelineFakeStore) GetTargetJobForParse(context.Context, string) (targetjob.TargetJobRecord, []targetjob.SourceRecord, error) {
	if s.getErr != nil {
		return targetjob.TargetJobRecord{}, nil, s.getErr
	}
	return s.target, s.sources, nil
}
func (s *pipelineFakeStore) UpdateTargetJobAnalysisFailure(context.Context, string, time.Time) error {
	s.updateAnalysisFail++
	return nil
}

type fakeRegistry struct {
	resolution targetjob.PromptResolution
	err        error
}

func (f *fakeRegistry) Resolve(_ context.Context, _ string, _ string) (targetjob.PromptResolution, error) {
	if f.err != nil {
		return targetjob.PromptResolution{}, f.err
	}
	return f.resolution, nil
}

type fakeAIClient struct {
	resp            aiclient.CompleteResponse
	err             error
	lastProfileName string
	lastPayload     aiclient.CompletePayload
}

func (f *fakeAIClient) Complete(_ context.Context, profileName string, payload aiclient.CompletePayload) (aiclient.CompleteResponse, aiclient.AICallMeta, error) {
	f.lastProfileName = profileName
	f.lastPayload = payload
	if f.err != nil {
		return aiclient.CompleteResponse{}, aiclient.AICallMeta{}, f.err
	}
	return f.resp, aiclient.AICallMeta{}, nil
}
func (f *fakeAIClient) Transcribe(context.Context, string, aiclient.TranscriptionInput) (aiclient.TranscriptionResponse, aiclient.AICallMeta, error) {
	return aiclient.TranscriptionResponse{}, aiclient.AICallMeta{}, errors.New("not implemented")
}
func (f *fakeAIClient) Stream(context.Context, string, aiclient.CompletePayload) (<-chan aiclient.AIStreamEvent, error) {
	return nil, errors.New("not implemented")
}
func (f *fakeAIClient) Synthesize(context.Context, string, aiclient.SynthesisInput) (aiclient.SynthesisResponse, aiclient.AICallMeta, error) {
	return aiclient.SynthesisResponse{}, aiclient.AICallMeta{}, errors.New("not implemented")
}

type fakeFetcher struct {
	res urlfetch.FetchResult
	err error
}

func (f *fakeFetcher) Fetch(context.Context, string) (urlfetch.FetchResult, error) {
	if f.err != nil {
		return urlfetch.FetchResult{}, f.err
	}
	return f.res, nil
}

// idSeq returns a deterministic IDGenerator for a fixed list. After the
// list is exhausted, it returns indexed strings.
func idSeq(prefix string) targetjob.IDGenerator {
	var n int32
	return func() string {
		next := atomic.AddInt32(&n, 1)
		return fmt.Sprintf("%s-%d", prefix, next)
	}
}

func newParseExecutorWithFakes(t *testing.T) (*targetjob.ParseExecutor, *pipelineFakeStore, *fakeRegistry, *fakeAIClient, *fakeFetcher) {
	t.Helper()
	store := &pipelineFakeStore{}
	registry := &fakeRegistry{
		resolution: targetjob.PromptResolution{
			PromptVersion:     "v1.0.0",
			RubricVersion:     "v1.0.0",
			ModelProfileName:  "target.import.default",
			DataSourceVersion: "v1",
		},
	}
	ai := &fakeAIClient{}
	fetcher := &fakeFetcher{}
	exec := targetjob.NewParseExecutor(targetjob.ParseExecutorOptions{
		Store:    store,
		Registry: registry,
		AI:       ai,
		Fetcher:  fetcher,
		NewID:    idSeq("id"),
		Now:      func() time.Time { return time.Date(2026, 5, 9, 22, 0, 0, 0, time.UTC) },
	})
	return exec, store, registry, ai, fetcher
}

const happyResponseJSON = `{
  "coreThemes": ["api"],
  "interviewHypotheses": ["microservices"],
  "strengths": ["Go"],
  "gaps": ["k8s"],
  "riskSignals": [],
  "requirements": [
    {"kind":"must_have","label":"Go","evidenceLevel":"explicit"},
    {"kind":"interview_focus","label":"system design"}
  ]
}`

func TestParseExecutor_HappyPath(t *testing.T) {
	exec, store, _, ai, _ := newParseExecutorWithFakes(t)
	ai.resp = aiclient.CompleteResponse{Content: happyResponseJSON}
	store.target = targetjob.TargetJobRecord{
		ID:             "tgt-1",
		UserID:         "user-1",
		SourceType:     targetjob.SourceTypeManualText,
		TargetLanguage: "en",
		RawJDText:      "JD text",
	}
	outcome := exec.Handle(context.Background(), targetjob.ClaimedJob{
		JobID: "j-1", JobType: "target_import", ResourceType: "target_job", ResourceID: "tgt-1",
	})
	if !outcome.Succeeded {
		t.Fatalf("happy path must succeed, got %+v", outcome)
	}
	if store.applyResultIn == nil || len(store.applyResultIn.Requirements) != 2 {
		t.Fatalf("expected 2 requirements applied, got %+v", store.applyResultIn)
	}
	if store.applyResultIn.LatestParseJobID != "j-1" {
		t.Fatalf("latest parse job id = %q", store.applyResultIn.LatestParseJobID)
	}
	if ai.lastProfileName != "target.import.default" {
		t.Fatalf("profileName = %q", ai.lastProfileName)
	}
	if got := ai.lastPayload.Metadata; got.FeatureKey != "target.import.parse" ||
		got.PromptVersion != "v1.0.0" ||
		got.RubricVersion != "v1.0.0" ||
		got.Language != "en" ||
		got.DataSourceVersion != "v1" {
		t.Fatalf("AI metadata did not carry F3 resolution: %+v", got)
	}
	if store.parsedOutboxPayload == nil {
		t.Fatal("target.parsed outbox payload missing")
	}
	if !store.sourceRefreshCalled {
		t.Fatal("source_refresh placeholder not enqueued")
	}
}

func TestParseExecutor_F3FailClosed(t *testing.T) {
	exec, store, registry, _, _ := newParseExecutorWithFakes(t)
	registry.err = targetjob.ErrPromptUnsupported
	store.target = targetjob.TargetJobRecord{ID: "tgt-1", SourceType: targetjob.SourceTypeManualText, RawJDText: "x"}
	outcome := exec.Handle(context.Background(), targetjob.ClaimedJob{ResourceID: "tgt-1"})
	if outcome.Succeeded || outcome.Retryable || outcome.ErrorCode != "AI_PROVIDER_CONFIG_INVALID" {
		t.Fatalf("F3 failure must map to AI_PROVIDER_CONFIG_INVALID non-retryable, got %+v", outcome)
	}
	if store.failedOutboxPayload == nil {
		t.Fatal("target.analysis.failed outbox payload missing")
	}
}

func TestParseExecutor_AIProviderTimeout_Retryable(t *testing.T) {
	exec, store, _, ai, _ := newParseExecutorWithFakes(t)
	ai.err = errors.New("provider error: AI_PROVIDER_TIMEOUT context deadline")
	store.target = targetjob.TargetJobRecord{ID: "tgt-1", SourceType: targetjob.SourceTypeManualText, RawJDText: "x"}
	outcome := exec.Handle(context.Background(), targetjob.ClaimedJob{ResourceID: "tgt-1"})
	if outcome.ErrorCode != "AI_PROVIDER_TIMEOUT" || !outcome.Retryable {
		t.Fatalf("AI timeout must be retryable, got %+v", outcome)
	}
}

func TestParseExecutor_AIOutputInvalid_NonRetryable(t *testing.T) {
	exec, store, _, ai, _ := newParseExecutorWithFakes(t)
	ai.resp = aiclient.CompleteResponse{Content: "not-json"}
	store.target = targetjob.TargetJobRecord{ID: "tgt-1", SourceType: targetjob.SourceTypeManualText, RawJDText: "x"}
	outcome := exec.Handle(context.Background(), targetjob.ClaimedJob{ResourceID: "tgt-1"})
	if outcome.ErrorCode != "AI_OUTPUT_INVALID" || outcome.Retryable {
		t.Fatalf("non-JSON response must map to AI_OUTPUT_INVALID non-retryable, got %+v", outcome)
	}
}

func TestParseExecutor_BlankJDText_SourceInvalid(t *testing.T) {
	exec, store, _, _, _ := newParseExecutorWithFakes(t)
	store.target = targetjob.TargetJobRecord{ID: "tgt-1", SourceType: targetjob.SourceTypeManualText, RawJDText: "   "}
	outcome := exec.Handle(context.Background(), targetjob.ClaimedJob{ResourceID: "tgt-1"})
	if outcome.ErrorCode != "TARGET_IMPORT_SOURCE_INVALID" || outcome.Retryable {
		t.Fatalf("blank JD must map to TARGET_IMPORT_SOURCE_INVALID non-retryable, got %+v", outcome)
	}
}

func TestParseExecutor_URLFetchUnavailable_Retryable(t *testing.T) {
	exec, store, _, _, fetcher := newParseExecutorWithFakes(t)
	fetcher.err = fmt.Errorf("%w: upstream timeout", urlfetch.ErrSourceUnavailable)
	store.target = targetjob.TargetJobRecord{
		ID: "tgt-1", SourceType: targetjob.SourceTypeURL, SourceURL: "https://jobs.example.com/1",
		RawJDText: "x",
	}
	outcome := exec.Handle(context.Background(), targetjob.ClaimedJob{ResourceID: "tgt-1"})
	if outcome.ErrorCode != "TARGET_IMPORT_SOURCE_UNAVAILABLE" || !outcome.Retryable {
		t.Fatalf("upstream unavailable must be retryable, got %+v", outcome)
	}
}

func TestParseExecutor_URLFetchInvalid_NonRetryable(t *testing.T) {
	exec, store, _, _, fetcher := newParseExecutorWithFakes(t)
	fetcher.err = fmt.Errorf("%w: bad scheme", urlfetch.ErrInvalidSource)
	store.target = targetjob.TargetJobRecord{
		ID: "tgt-1", SourceType: targetjob.SourceTypeURL, SourceURL: "https://jobs.example.com/1",
		RawJDText: "x",
	}
	outcome := exec.Handle(context.Background(), targetjob.ClaimedJob{ResourceID: "tgt-1"})
	if outcome.ErrorCode != "TARGET_IMPORT_SOURCE_INVALID" || outcome.Retryable {
		t.Fatalf("invalid source must map to TARGET_IMPORT_SOURCE_INVALID non-retryable, got %+v", outcome)
	}
}

func TestParseExecutor_URLFetchBodyIsPersistedAndParsed(t *testing.T) {
	exec, store, _, ai, fetcher := newParseExecutorWithFakes(t)
	ai.resp = aiclient.CompleteResponse{Content: happyResponseJSON}
	fetchedAt := time.Date(2026, 5, 9, 22, 30, 0, 0, time.UTC)
	fetcher.res = urlfetch.FetchResult{
		SanitizedURL: "https://jobs.example.com/role/1",
		Body:         "Fetched JD body for a Backend Engineer.",
		FetchedAt:    fetchedAt,
	}
	store.target = targetjob.TargetJobRecord{
		ID:             "tgt-1",
		UserID:         "user-1",
		SourceType:     targetjob.SourceTypeURL,
		SourceURL:      "https://jobs.example.com/role/1?token=secret",
		TargetLanguage: "en",
	}
	store.sources = []targetjob.SourceRecord{{ID: "src-1", SourceType: targetjob.SourceTypeURL}}

	outcome := exec.Handle(context.Background(), targetjob.ClaimedJob{JobID: "j-1", ResourceID: "tgt-1"})
	if !outcome.Succeeded {
		t.Fatalf("URL fetch body should drive parse success, got %+v", outcome)
	}
	if store.sourceSnapshotText != "Fetched JD body for a Backend Engineer." || store.sourceSnapshotAt == nil {
		t.Fatalf("source snapshot not persisted: text=%q at=%v", store.sourceSnapshotText, store.sourceSnapshotAt)
	}
	if store.sourceSnapshotURL != "https://jobs.example.com/role/1" {
		t.Fatalf("sanitized source URL = %q", store.sourceSnapshotURL)
	}
	if store.applyResultIn == nil || len(store.applyResultIn.Requirements) == 0 {
		t.Fatalf("parse result was not applied from fetched URL body: %+v", store.applyResultIn)
	}
}

func TestSourceRefreshHandler_MarksStale(t *testing.T) {
	store := &pipelineFakeStore{}
	h := &targetjob.SourceRefreshHandler{Store: store, Now: func() time.Time { return time.Now().UTC() }}
	outcome := h.Handle(context.Background(), targetjob.ClaimedJob{ResourceID: "tgt-1"})
	if !outcome.Succeeded {
		t.Fatalf("source refresh must succeed by default, got %+v", outcome)
	}
	if store.sourceFreshnessUpd != "stale" {
		t.Fatalf("source freshness should be stale, got %q", store.sourceFreshnessUpd)
	}
}

func TestDrainer_RunOnceProcessesQueuedJobAndFinalizes(t *testing.T) {
	store := &pipelineFakeStore{
		queuedJobs: []targetjob.ClaimedJob{
			{JobID: "j-1", JobType: "target_import", ResourceType: "target_job", ResourceID: "tgt-1"},
		},
	}
	called := false
	drainer := targetjob.NewDrainer(targetjob.DrainerOptions{
		Store: store,
		Handlers: map[string]targetjob.JobHandler{
			"target_import": targetjob.JobHandlerFunc(func(_ context.Context, j targetjob.ClaimedJob) targetjob.JobOutcome {
				called = true
				if j.JobID != "j-1" {
					t.Errorf("unexpected jobID: %q", j.JobID)
				}
				return targetjob.JobOutcome{Succeeded: true}
			}),
		},
	})
	processed, err := drainer.RunOnce(context.Background())
	if err != nil {
		t.Fatalf("RunOnce: %v", err)
	}
	if !processed || !called || len(store.finalize) != 1 || !store.finalize[0].Succeeded {
		t.Fatalf("expected one processed job + finalize succeeded; got processed=%v called=%v finalize=%+v", processed, called, store.finalize)
	}
}

func TestDrainer_RunOnceWithUnknownJobTypeFinalizesNonRetryableFailure(t *testing.T) {
	store := &pipelineFakeStore{
		queuedJobs: []targetjob.ClaimedJob{
			{JobID: "j-9", JobType: "ghost_runner", ResourceType: "target_job", ResourceID: "tgt-9"},
		},
	}
	drainer := targetjob.NewDrainer(targetjob.DrainerOptions{
		Store: store,
		Handlers: map[string]targetjob.JobHandler{
			"target_import": targetjob.JobHandlerFunc(func(_ context.Context, _ targetjob.ClaimedJob) targetjob.JobOutcome {
				return targetjob.JobOutcome{Succeeded: true}
			}),
		},
	})
	if _, err := drainer.RunOnce(context.Background()); err != nil {
		t.Fatal(err)
	}
	// note: jobTypes derived from handlers map = ["target_import"]; the
	// fake store returns the queued job regardless of jobTypes filter to
	// exercise the unknown-handler safety path.
	if len(store.finalize) != 1 || store.finalize[0].Succeeded || store.finalize[0].ErrorCode != "TARGET_IMPORT_FAILED" {
		t.Fatalf("expected unknown-handler failure finalize, got %+v", store.finalize)
	}
}

func TestDrainer_StartShutdownDrainsCleanly(t *testing.T) {
	store := &pipelineFakeStore{
		queuedJobs: []targetjob.ClaimedJob{
			{JobID: "j-1", JobType: "target_import", ResourceID: "tgt-1"},
		},
	}
	processed := make(chan struct{}, 1)
	drainer := targetjob.NewDrainer(targetjob.DrainerOptions{
		Store: store,
		Handlers: map[string]targetjob.JobHandler{
			"target_import": targetjob.JobHandlerFunc(func(_ context.Context, _ targetjob.ClaimedJob) targetjob.JobOutcome {
				processed <- struct{}{}
				return targetjob.JobOutcome{Succeeded: true}
			}),
		},
		Workers:      1,
		PollInterval: 10 * time.Millisecond,
	})
	drainer.Start(context.Background())

	select {
	case <-processed:
	case <-time.After(2 * time.Second):
		t.Fatal("worker never picked up queued job")
	}
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := drainer.Shutdown(shutdownCtx); err != nil {
		t.Fatalf("Shutdown: %v", err)
	}
}
