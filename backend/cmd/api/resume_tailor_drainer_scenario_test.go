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
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
	"github.com/monshunter/easyinterview/backend/internal/shared/jobs"
	"github.com/monshunter/easyinterview/backend/internal/targetjob"
)

func TestResumeTailorDrainerHTTPScenario(t *testing.T) {
	now := time.Date(2026, 6, 13, 14, 0, 0, 0, time.UTC)
	tailorRunID := "0195f2d0-4a44-7fc2-8f77-1f9c4cfa0001"
	resumeID := "0195f2d0-4a44-7fc2-8f77-1f9c4cfa0002"
	store := newAPITailorStore(apiTailorContext(tailorRunID, resumeID))
	asyncStore := &apiResumeAsyncStore{job: targetjob.ClaimedJob{
		JobID:        "0195f2d0-4a44-7fc2-8f77-1f9c4cfa0003",
		JobType:      string(jobs.JobTypeResumeTailor),
		ResourceType: "resume_tailor_run",
		ResourceID:   tailorRunID,
		Payload:      []byte(`{"resumeId":"` + resumeID + `","targetJobId":"0195f2d0-4a44-7fc2-8f77-1f9c4cfc0003","mode":"gap_review"}`),
		Attempts:     1,
		MaxAttempts:  5,
	}}
	taskRuns := &apiTailorTaskRuns{}
	handler := resumejobs.NewTailorHandler(resumejobs.TailorHandlerOptions{
		Store:      store,
		Registry:   apiTailorRegistry{},
		AI:         &apiTailorAI{responses: []aiclient.CompleteResponse{{Content: apiValidTailorJSON}}},
		AITaskRuns: taskRuns,
		NewID:      apiFixedIDs("0195f2d0-4a44-7fc2-8f77-1f9c4cfa0101", "0195f2d0-4a44-7fc2-8f77-1f9c4cfa0102", "0195f2d0-4a44-7fc2-8f77-1f9c4cfa0103", "0195f2d0-4a44-7fc2-8f77-1f9c4cfa0104", "0195f2d0-4a44-7fc2-8f77-1f9c4cfa0105"),
		Now:        func() time.Time { return now },
	})
	drainer := targetjob.NewDrainer(targetjob.DrainerOptions{
		Store: asyncStore,
		Handlers: map[string]targetjob.JobHandler{
			string(jobs.JobTypeResumeTailor): handler,
		},
		Now: func() time.Time { return now },
	})

	processed, err := drainer.RunOnce(context.Background())
	if err != nil {
		t.Fatalf("RunOnce: %v", err)
	}
	if !processed || !asyncStore.outcome.Succeeded {
		t.Fatalf("drainer did not process resume_tailor successfully: processed=%v outcome=%+v", processed, asyncStore.outcome)
	}
	success := store.successes[tailorRunID]
	if success == nil || success.ResumeID != resumeID || len(success.Suggestions) != 3 {
		t.Fatalf("tailor success not persisted: %+v", success)
	}
	var outbox map[string]any
	if err := json.Unmarshal(success.OutboxEventPayload, &outbox); err != nil {
		t.Fatalf("decode outbox: %v", err)
	}
	if len(outbox) != 5 || outbox["tailorRunId"] != tailorRunID || outbox["resumeId"] != resumeID || outbox["status"] != "ready" {
		t.Fatalf("outbox payload drift: %+v", outbox)
	}
	if len(taskRuns.rows) != 1 || taskRuns.rows[0].Capability != aiclient.AITaskRunTaskResumeTailor || taskRuns.rows[0].Status != aiclient.AITaskRunStatusSuccess {
		t.Fatalf("ai_task_runs drift: %+v", taskRuns.rows)
	}
}

func TestResumeTailorDrainerFailureScenario(t *testing.T) {
	now := time.Date(2026, 6, 13, 14, 30, 0, 0, time.UTC)
	timeoutRunID := "0195f2d0-4a44-7fc2-8f77-1f9c4cfb0001"
	invalidRunID := "0195f2d0-4a44-7fc2-8f77-1f9c4cfb0002"
	retryRunID := "0195f2d0-4a44-7fc2-8f77-1f9c4cfb0003"
	resumeID := "0195f2d0-4a44-7fc2-8f77-1f9c4cfb0004"
	store := newAPITailorStore(
		apiTailorContext(timeoutRunID, resumeID),
		apiTailorContext(invalidRunID, resumeID),
		apiTailorContext(retryRunID, resumeID),
	)
	asyncStore := &apiResumeRetryAsyncStore{jobs: []targetjob.ClaimedJob{
		apiTailorClaimedJob(timeoutRunID, resumeID, 1),
		apiTailorClaimedJob(invalidRunID, resumeID, 1),
		apiTailorClaimedJob(retryRunID, resumeID, 1),
		apiTailorClaimedJob(retryRunID, resumeID, 2),
	}}
	taskRuns := &apiTailorTaskRuns{}
	handler := resumejobs.NewTailorHandler(resumejobs.TailorHandlerOptions{
		Store:    store,
		Registry: apiTailorRegistry{},
		AI: &apiTailorAI{
			errs: []error{
				errors.New(sharederrors.CodeAiProviderTimeout + " provider slow"),
				nil,
				errors.New(sharederrors.CodeAiProviderTimeout + " provider slow"),
			},
			responses: []aiclient.CompleteResponse{
				{Content: `not-json`},
				{Content: apiValidTailorJSON},
			},
		},
		AITaskRuns: taskRuns,
		NewID: apiFixedIDs(
			"0195f2d0-4a44-7fc2-8f77-1f9c4cfb0101", "0195f2d0-4a44-7fc2-8f77-1f9c4cfb0102",
			"0195f2d0-4a44-7fc2-8f77-1f9c4cfb0103", "0195f2d0-4a44-7fc2-8f77-1f9c4cfb0104",
			"0195f2d0-4a44-7fc2-8f77-1f9c4cfb0105", "0195f2d0-4a44-7fc2-8f77-1f9c4cfb0106",
			"0195f2d0-4a44-7fc2-8f77-1f9c4cfb0107", "0195f2d0-4a44-7fc2-8f77-1f9c4cfb0108",
		),
		Now: func() time.Time { return now },
	})
	drainer := targetjob.NewDrainer(targetjob.DrainerOptions{
		Store: asyncStore,
		Handlers: map[string]targetjob.JobHandler{
			string(jobs.JobTypeResumeTailor): handler,
		},
		Now: func() time.Time { return now },
	})

	for i := 0; i < 4; i++ {
		processed, err := drainer.RunOnce(context.Background())
		if err != nil {
			t.Fatalf("RunOnce %d: %v", i+1, err)
		}
		if !processed {
			t.Fatalf("RunOnce %d did not process a job", i+1)
		}
	}
	if len(asyncStore.outcomes) != 4 {
		t.Fatalf("outcomes = %+v", asyncStore.outcomes)
	}
	// D-20: failures surface through JobOutcome; the runner kernel persists the
	// async_jobs row status (there is no resume_tailor_runs status to flip).
	if asyncStore.outcomes[0].ErrorCode != sharederrors.CodeAiProviderTimeout || !asyncStore.outcomes[0].Retryable {
		t.Fatalf("timeout outcome = %+v", asyncStore.outcomes[0])
	}
	if asyncStore.outcomes[1].ErrorCode != sharederrors.CodeAiOutputInvalid || asyncStore.outcomes[1].Retryable {
		t.Fatalf("invalid output outcome = %+v", asyncStore.outcomes[1])
	}
	if asyncStore.outcomes[2].ErrorCode != sharederrors.CodeAiProviderTimeout || !asyncStore.outcomes[2].Retryable {
		t.Fatalf("retry first outcome = %+v", asyncStore.outcomes[2])
	}
	if !asyncStore.outcomes[3].Succeeded {
		t.Fatalf("retry second outcome = %+v", asyncStore.outcomes[3])
	}
	// Only the retried run reaches a ready result + outbox; failed runs persist nothing.
	if store.successes[retryRunID] == nil || len(store.successes) != 1 {
		t.Fatalf("ready-only success/outbox drift: successes=%+v", store.successes)
	}
	if store.successes[timeoutRunID] != nil || store.successes[invalidRunID] != nil {
		t.Fatalf("failed runs must not persist a ready result: successes=%+v", store.successes)
	}
	if len(taskRuns.rows) != 4 {
		t.Fatalf("ai_task_runs rows = %+v", taskRuns.rows)
	}
}

func apiTailorClaimedJob(tailorRunID string, resumeID string, attempts int32) targetjob.ClaimedJob {
	return targetjob.ClaimedJob{
		JobID:        tailorRunID,
		JobType:      string(jobs.JobTypeResumeTailor),
		ResourceType: "resume_tailor_run",
		ResourceID:   tailorRunID,
		Payload:      []byte(`{"resumeId":"` + resumeID + `","targetJobId":"0195f2d0-4a44-7fc2-8f77-1f9c4cfc0003","mode":"gap_review"}`),
		Attempts:     attempts,
		MaxAttempts:  5,
	}
}

func apiTailorContext(tailorRunID string, resumeID string) resumestore.TailorJobContext {
	return resumestore.TailorJobContext{
		TailorRunID:       tailorRunID,
		UserID:            "0195f2d0-4a44-7fc2-8f77-1f9c4cfc0001",
		ResumeID:          resumeID,
		TargetJobID:       "0195f2d0-4a44-7fc2-8f77-1f9c4cfc0003",
		Mode:              "gap_review",
		Language:          "en",
		ResumeSummary:     json.RawMessage(`{"headline":"Senior engineer"}`),
		StructuredProfile: json.RawMessage(`{"sections":[{"bullets":["Led migration."]}]}`),
		TargetSummary:     json.RawMessage(`{"requirements":["distributed systems"]}`),
		TargetTitle:       "Staff Backend Engineer",
		TargetSeniority:   "staff",
		OriginalBullet:    "Led migration.",
	}
}

type apiTailorStore struct {
	contexts  map[string]resumestore.TailorJobContext
	successes map[string]*resumestore.CompleteTailorRunSuccessInput
}

func newAPITailorStore(contexts ...resumestore.TailorJobContext) *apiTailorStore {
	out := &apiTailorStore{
		contexts:  map[string]resumestore.TailorJobContext{},
		successes: map[string]*resumestore.CompleteTailorRunSuccessInput{},
	}
	for _, ctx := range contexts {
		out.contexts[ctx.TailorRunID] = ctx
	}
	return out
}

func (s *apiTailorStore) GetForTailor(_ context.Context, tailorRunID string) (resumestore.TailorJobContext, error) {
	ctx, ok := s.contexts[tailorRunID]
	if !ok {
		return resumestore.TailorJobContext{}, resumestore.ErrTailorRunNotFound
	}
	return ctx, nil
}

func (s *apiTailorStore) CompleteTailorRunSuccess(_ context.Context, in resumestore.CompleteTailorRunSuccessInput) error {
	cp := in
	s.successes[in.TailorRunID] = &cp
	return nil
}

type apiTailorRegistry struct{}

func (apiTailorRegistry) Resolve(context.Context, string, string) (resumejobs.PromptResolution, error) {
	return resumejobs.PromptResolution{
		PromptVersion:       "v0.1.0",
		RubricVersion:       "v0.1.0",
		ModelProfileName:    "resume.tailor.default",
		DataSourceVersion:   "registry.v1",
		FeatureFlag:         "none",
		UserMessageTemplate: "Resume summary: {{resume_summary}}\nJD summary: {{jd_summary}}\nTarget seniority: {{target_seniority}}\nLanguage: {{language}}",
	}, nil
}

type apiTailorAI struct {
	responses []aiclient.CompleteResponse
	errs      []error
	calls     int
}

func (c *apiTailorAI) Complete(_ context.Context, profileName string, _ aiclient.CompletePayload) (aiclient.CompleteResponse, aiclient.AICallMeta, error) {
	idx := c.calls
	c.calls++
	meta := aiclient.AICallMeta{
		Provider:            "stub",
		ModelFamily:         "stub",
		ModelID:             "fixture-model:resume-tailor",
		ModelProfileName:    profileName,
		ModelProfileVersion: "1.1.0",
		FallbackChain:       []string{"stub/fixture-model:resume-tailor"},
		ValidationStatus:    aiclient.ValidationStatusOK,
	}
	if idx < len(c.errs) && c.errs[idx] != nil {
		meta.ErrorCode = sharederrors.CodeAiProviderTimeout
		return aiclient.CompleteResponse{}, meta, c.errs[idx]
	}
	responseIdx := idx
	errCountBefore := 0
	for i := 0; i < idx && i < len(c.errs); i++ {
		if c.errs[i] != nil {
			errCountBefore++
		}
	}
	responseIdx -= errCountBefore
	if responseIdx >= 0 && responseIdx < len(c.responses) {
		return c.responses[responseIdx], meta, nil
	}
	return aiclient.CompleteResponse{}, meta, errors.New("unexpected Complete call")
}

func (c *apiTailorAI) Transcribe(context.Context, string, aiclient.TranscriptionInput) (aiclient.TranscriptionResponse, aiclient.AICallMeta, error) {
	return aiclient.TranscriptionResponse{}, aiclient.AICallMeta{}, errors.New("unexpected Transcribe call")
}

func (c *apiTailorAI) Stream(context.Context, string, aiclient.CompletePayload) (<-chan aiclient.AIStreamEvent, error) {
	return nil, errors.New("unexpected Stream call")
}

func (c *apiTailorAI) Synthesize(context.Context, string, aiclient.SynthesisInput) (aiclient.SynthesisResponse, aiclient.AICallMeta, error) {
	return aiclient.SynthesisResponse{}, aiclient.AICallMeta{}, errors.New("unexpected Synthesize call")
}

type apiTailorTaskRuns struct {
	rows []aiclient.AITaskRunRow
}

func (w *apiTailorTaskRuns) WriteAITaskRun(_ context.Context, row aiclient.AITaskRunRow) error {
	w.rows = append(w.rows, row)
	return nil
}

const apiValidTailorJSON = `{
  "matchSummary": {"strengths":["Strong systems evidence"],"gaps":["Add scale metrics"]},
  "suggestions": [
    {"originalBullet":"Led migration.","suggestedBullet":"Led migration across 12 teams.","reason":"Adds scope."},
    {"originalBullet":"Built services.","suggestedBullet":"Built reliable services.","reason":"Adds outcome."},
    {"originalBullet":"Improved reliability.","suggestedBullet":"Improved reliability with SLOs.","reason":"Adds measurement."}
  ]
}`
