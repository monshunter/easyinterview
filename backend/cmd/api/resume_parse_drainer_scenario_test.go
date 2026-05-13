package main

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient"
	resumejobs "github.com/monshunter/easyinterview/backend/internal/resume/jobs"
	resumestore "github.com/monshunter/easyinterview/backend/internal/resume/store"
	"github.com/monshunter/easyinterview/backend/internal/shared/jobs"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
	"github.com/monshunter/easyinterview/backend/internal/targetjob"
)

func TestResumeParseDrainerHTTPScenario(t *testing.T) {
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
	asyncStore := &apiResumeAsyncStore{job: targetjob.ClaimedJob{
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
	drainer := targetjob.NewDrainer(targetjob.DrainerOptions{
		Store: asyncStore,
		Handlers: map[string]targetjob.JobHandler{
			string(jobs.JobTypeResumeParse): handler,
		},
		Now: func() time.Time { return now },
	})

	processed, err := drainer.RunOnce(context.Background())
	if err != nil {
		t.Fatalf("RunOnce: %v", err)
	}
	if !processed || !asyncStore.outcome.Succeeded {
		t.Fatalf("drainer did not process resume_parse successfully: processed=%v outcome=%+v", processed, asyncStore.outcome)
	}
	if parseStore.success == nil || parseStore.success.AssetID != assetID || parseStore.success.ParsedTextSnapshot != "Private resume body" {
		t.Fatalf("parse success not persisted: %+v", parseStore.success)
	}
	var outbox map[string]any
	if err := json.Unmarshal(parseStore.success.OutboxEventPayload, &outbox); err != nil {
		t.Fatalf("decode outbox: %v", err)
	}
	if outbox["resumeAssetId"] != assetID || outbox["parseStatus"] != "ready" {
		t.Fatalf("outbox payload drift: %+v", outbox)
	}
}

type apiResumeAsyncStore struct {
	job     targetjob.ClaimedJob
	claimed bool
	outcome targetjob.JobOutcome
}

func (s *apiResumeAsyncStore) ClaimNextAsyncJob(_ context.Context, _ []string, _ time.Time) (targetjob.ClaimedJob, bool, error) {
	if s.claimed {
		return targetjob.ClaimedJob{}, false, nil
	}
	s.claimed = true
	return s.job, true, nil
}

func (s *apiResumeAsyncStore) FinalizeAsyncJob(_ context.Context, jobID string, outcome targetjob.JobOutcome, _ time.Time) error {
	s.outcome = outcome
	return nil
}

type apiResumeParseStore struct {
	asset   resumestore.ParseAssetRecord
	success *resumestore.CompleteParseSuccessInput
	failure *resumestore.CompleteParseFailureInput
}

func (s *apiResumeParseStore) GetForParse(context.Context, string) (resumestore.ParseAssetRecord, error) {
	return s.asset, nil
}

func (s *apiResumeParseStore) MarkParsing(context.Context, resumestore.StatusUpdateInput) error {
	s.asset.ParseStatus = sharedtypes.TargetJobParseStatusProcessing
	return nil
}

func (s *apiResumeParseStore) CompleteParseSuccess(_ context.Context, in resumestore.CompleteParseSuccessInput) error {
	cp := in
	s.success = &cp
	return nil
}

func (s *apiResumeParseStore) CompleteParseFailure(_ context.Context, in resumestore.CompleteParseFailureInput) error {
	cp := in
	s.failure = &cp
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
