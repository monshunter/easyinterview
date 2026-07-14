package main

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient"
	resumejobs "github.com/monshunter/easyinterview/backend/internal/resume/jobs"
	resumestore "github.com/monshunter/easyinterview/backend/internal/resume/store"
	"github.com/monshunter/easyinterview/backend/internal/runner"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
	"github.com/monshunter/easyinterview/backend/internal/shared/jobs"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

func TestResumeParseRunnerIntegration(t *testing.T) {
	now := time.Date(2026, 5, 13, 10, 0, 0, 0, time.UTC)
	assetID := "01918fa0-0000-7000-8000-00000000a001"
	userID := "01918fa0-0000-7000-8000-00000000a002"
	parseStore := &apiResumeParseStore{asset: resumestore.ParseAssetRecord{
		ID:           assetID,
		UserID:       userID,
		Language:     "en",
		ParseStatus:  sharedtypes.TargetJobParseStatusQueued,
		SourceType:   "paste",
		OriginalText: "Private resume body",
	}}
	asyncStore := &apiResumeAsyncStore{job: runner.ClaimedJob{
		JobID:        "01918fa0-0000-7000-8000-00000000a003",
		JobType:      string(jobs.JobTypeResumeParse),
		ResourceType: "resume_asset",
		ResourceID:   assetID,
		Attempts:     1,
		MaxAttempts:  5,
	}}
	handler := resumejobs.NewParseHandler(resumejobs.ParseHandlerOptions{
		Store:    parseStore,
		Registry: apiResumeRegistry{},
		AI:       resumejobs.NewDeterministicParseAIClient(&apiNoopAIClient{}),
		NewID:    apiFixedIDs("01918fa0-0000-7000-8000-00000000a004"),
		Now:      func() time.Time { return now },
	})
	kernel := newIntegrationJobRuntime(asyncStore, func() time.Time { return now }, map[string]runner.Handler{
		string(jobs.JobTypeResumeParse): handler,
	})

	processed, err := kernel.RunOnce(context.Background())
	if err != nil {
		t.Fatalf("RunOnce: %v", err)
	}
	if !processed || !asyncStore.outcome.Succeeded {
		t.Fatalf("runner did not process resume_parse successfully: processed=%v outcome=%+v", processed, asyncStore.outcome)
	}
	if parseStore.success == nil || parseStore.success.AssetID != assetID || parseStore.success.ParsedTextSnapshot != "# Private resume body" {
		t.Fatalf("parse success not persisted: %+v", parseStore.success)
	}
	if parseStore.success.DisplayName == nil || *parseStore.success.DisplayName != "Fixture Candidate - Engineer" {
		t.Fatalf("display name = %#v, want Fixture Candidate - Engineer", parseStore.success.DisplayName)
	}
	var outbox map[string]any
	if err := json.Unmarshal(parseStore.success.OutboxEventPayload, &outbox); err != nil {
		t.Fatalf("decode outbox: %v", err)
	}
	if outbox["resumeId"] != assetID || outbox["parseStatus"] != "ready" {
		t.Fatalf("outbox payload drift: %+v", outbox)
	}
}

func TestResumeParseRunnerRetryableFailureIntegration(t *testing.T) {
	now := time.Date(2026, 5, 13, 10, 30, 0, 0, time.UTC)
	assetID := "01918fa0-0000-7000-8000-00000000b001"
	userID := "01918fa0-0000-7000-8000-00000000b002"
	parseStore := &apiResumeParseStore{asset: resumestore.ParseAssetRecord{
		ID:           assetID,
		UserID:       userID,
		Language:     "en",
		ParseStatus:  sharedtypes.TargetJobParseStatusQueued,
		SourceType:   "paste",
		OriginalText: "Private resume body",
	}}
	asyncStore := &apiResumeRetryAsyncStore{jobs: []runner.ClaimedJob{
		{
			JobID:        "01918fa0-0000-7000-8000-00000000b003",
			JobType:      string(jobs.JobTypeResumeParse),
			ResourceType: "resume_asset",
			ResourceID:   assetID,
			Attempts:     1,
			MaxAttempts:  5,
		},
		{
			JobID:        "01918fa0-0000-7000-8000-00000000b003",
			JobType:      string(jobs.JobTypeResumeParse),
			ResourceType: "resume_asset",
			ResourceID:   assetID,
			Attempts:     2,
			MaxAttempts:  5,
		},
	}}
	handler := resumejobs.NewParseHandler(resumejobs.ParseHandlerOptions{
		Store:    parseStore,
		Registry: apiResumeRegistry{},
		AI: &apiSequenceAIClient{
			errs: []error{errors.New(sharederrors.CodeAiProviderTimeout + " provider slow")},
			responses: []aiclient.CompleteResponse{{
				Content:      `{"displayName":"Fixture Candidate","basics":{"name":"Fixture Candidate"},"experiences":[],"projects":[],"education":[],"skills":["Go"],"languages":["en"]}`,
				FinishReason: "stop",
			}},
		},
		NewID: apiFixedIDs("01918fa0-0000-7000-8000-00000000b004"),
		Now:   func() time.Time { return now },
	})
	kernel := newIntegrationJobRuntime(asyncStore, func() time.Time { return now }, map[string]runner.Handler{
		string(jobs.JobTypeResumeParse): handler,
	})

	processed, err := kernel.RunOnce(context.Background())
	if err != nil {
		t.Fatalf("first RunOnce: %v", err)
	}
	if !processed || len(asyncStore.outcomes) != 1 || !asyncStore.outcomes[0].Retryable {
		t.Fatalf("first retryable outcome drift: processed=%v outcomes=%+v", processed, asyncStore.outcomes)
	}
	if parseStore.failure == nil || parseStore.failure.ErrorCode != sharederrors.CodeAiProviderTimeout || parseStore.asset.ParseStatus != sharedtypes.TargetJobParseStatusFailed {
		t.Fatalf("retryable failure not persisted: failure=%+v status=%s", parseStore.failure, parseStore.asset.ParseStatus)
	}

	processed, err = kernel.RunOnce(context.Background())
	if err != nil {
		t.Fatalf("second RunOnce: %v", err)
	}
	if !processed || len(asyncStore.outcomes) != 2 || !asyncStore.outcomes[1].Succeeded {
		t.Fatalf("second success outcome drift: processed=%v outcomes=%+v", processed, asyncStore.outcomes)
	}
	if parseStore.success == nil || parseStore.asset.ParseStatus != sharedtypes.TargetJobParseStatusReady {
		t.Fatalf("retry success not persisted: success=%+v status=%s", parseStore.success, parseStore.asset.ParseStatus)
	}
	if parseStore.success.DisplayName == nil || *parseStore.success.DisplayName != "Fixture Candidate" {
		t.Fatalf("retry success display name = %#v, want Fixture Candidate", parseStore.success.DisplayName)
	}
	if parseStore.markParsing != 2 {
		t.Fatalf("MarkParsing calls = %d, want 2", parseStore.markParsing)
	}
	var outbox map[string]any
	if err := json.Unmarshal(parseStore.success.OutboxEventPayload, &outbox); err != nil {
		t.Fatalf("decode outbox: %v", err)
	}
	if outbox["resumeId"] != assetID || outbox["parseStatus"] != "ready" {
		t.Fatalf("outbox payload drift: %+v", outbox)
	}
}

type apiResumeAsyncStore struct {
	job     runner.ClaimedJob
	claimed bool
	outcome runner.JobOutcome
}

func (s *apiResumeAsyncStore) LeaseAsyncJob(_ context.Context, _ []string, _ time.Time) (runner.ClaimedJob, bool, error) {
	if s.claimed {
		return runner.ClaimedJob{}, false, nil
	}
	s.claimed = true
	return s.job, true, nil
}

func (s *apiResumeAsyncStore) FinalizeAsyncJob(_ context.Context, jobID string, _ int32, outcome runner.JobOutcome, _ time.Time, _ time.Time) error {
	s.outcome = outcome
	return nil
}

func (s *apiResumeAsyncStore) ReclaimExpiredLeases(context.Context, []string, time.Time, time.Time) (int64, error) {
	return 0, nil
}

type apiResumeRetryAsyncStore struct {
	jobs     []runner.ClaimedJob
	next     int
	outcomes []runner.JobOutcome
}

func (s *apiResumeRetryAsyncStore) LeaseAsyncJob(_ context.Context, _ []string, _ time.Time) (runner.ClaimedJob, bool, error) {
	if s.next >= len(s.jobs) {
		return runner.ClaimedJob{}, false, nil
	}
	job := s.jobs[s.next]
	s.next++
	return job, true, nil
}

func (s *apiResumeRetryAsyncStore) FinalizeAsyncJob(_ context.Context, _ string, _ int32, outcome runner.JobOutcome, _ time.Time, _ time.Time) error {
	s.outcomes = append(s.outcomes, outcome)
	return nil
}

func (s *apiResumeRetryAsyncStore) ReclaimExpiredLeases(context.Context, []string, time.Time, time.Time) (int64, error) {
	return 0, nil
}

type apiResumeParseStore struct {
	asset       resumestore.ParseAssetRecord
	markParsing int
	success     *resumestore.CompleteParseSuccessInput
	failure     *resumestore.CompleteParseFailureInput
}

func (s *apiResumeParseStore) GetForParse(context.Context, string) (resumestore.ParseAssetRecord, error) {
	return s.asset, nil
}

func (s *apiResumeParseStore) MarkParsing(context.Context, resumestore.StatusUpdateInput) error {
	s.markParsing++
	s.asset.ParseStatus = sharedtypes.TargetJobParseStatusProcessing
	return nil
}

func (s *apiResumeParseStore) CompleteParseSuccess(_ context.Context, in resumestore.CompleteParseSuccessInput) error {
	cp := in
	s.success = &cp
	s.asset.ParseStatus = sharedtypes.TargetJobParseStatusReady
	return nil
}

func (s *apiResumeParseStore) CompleteParseFailure(_ context.Context, in resumestore.CompleteParseFailureInput) error {
	cp := in
	s.failure = &cp
	s.asset.ParseStatus = sharedtypes.TargetJobParseStatusFailed
	return nil
}

type apiResumeRegistry struct{}

func (apiResumeRegistry) Resolve(context.Context, string, string) (resumejobs.PromptResolution, error) {
	return resumejobs.PromptResolution{
		PromptVersion:       "v0.1.0",
		RubricVersion:       "v0.1.0",
		ModelProfileName:    "resume.parse.default",
		DataSourceVersion:   "registry.v1",
		FeatureFlag:         "none",
		UserMessageTemplate: "{{resume_text}}",
	}, nil
}

var _ aiclient.AIClient = (*apiNoopAIClient)(nil)

type apiSequenceAIClient struct {
	responses []aiclient.CompleteResponse
	errs      []error
	calls     int
}

func (c *apiSequenceAIClient) Complete(context.Context, string, aiclient.CompletePayload) (aiclient.CompleteResponse, aiclient.AICallMeta, error) {
	idx := c.calls
	c.calls++
	meta := aiclient.AICallMeta{
		Provider:         "stub",
		ModelFamily:      "stub",
		ModelID:          "resume-parse-fixture",
		FallbackChain:    []string{"stub/resume-parse-fixture"},
		ValidationStatus: aiclient.ValidationStatusOK,
	}
	if idx < len(c.errs) && c.errs[idx] != nil {
		return aiclient.CompleteResponse{}, meta, c.errs[idx]
	}
	responseIdx := idx - len(c.errs)
	if responseIdx >= 0 && responseIdx < len(c.responses) {
		return c.responses[responseIdx], meta, nil
	}
	return aiclient.CompleteResponse{}, meta, errors.New("unexpected Complete call")
}

func (c *apiSequenceAIClient) Transcribe(context.Context, string, aiclient.TranscriptionInput) (aiclient.TranscriptionResponse, aiclient.AICallMeta, error) {
	return aiclient.TranscriptionResponse{}, aiclient.AICallMeta{}, errors.New("unexpected Transcribe call")
}

func (c *apiSequenceAIClient) Stream(context.Context, string, aiclient.CompletePayload) (<-chan aiclient.AIStreamEvent, error) {
	return nil, errors.New("unexpected Stream call")
}

func (c *apiSequenceAIClient) Synthesize(context.Context, string, aiclient.SynthesisInput) (aiclient.SynthesisResponse, aiclient.AICallMeta, error) {
	return aiclient.SynthesisResponse{}, aiclient.AICallMeta{}, errors.New("unexpected Synthesize call")
}
