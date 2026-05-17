package jobs_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient"
	resumejobs "github.com/monshunter/easyinterview/backend/internal/resume/jobs"
	resumestore "github.com/monshunter/easyinterview/backend/internal/resume/store"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
	"github.com/monshunter/easyinterview/backend/internal/shared/jobs"
	"github.com/monshunter/easyinterview/backend/internal/targetjob"
)

func TestTailorHandlerHappyPathWritesReadySuggestionsTaskRunAndPrivateOutbox(t *testing.T) {
	now := time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC)
	tailorRunID := "0195f2d0-4a44-7fc2-8f77-1f9c4cf7a001"
	userID := "0195f2d0-4a44-7fc2-8f77-1f9c4cf7a002"
	versionID := "0195f2d0-4a44-7fc2-8f77-1f9c4cf7a003"
	assetID := "0195f2d0-4a44-7fc2-8f77-1f9c4cf7a004"
	targetID := "0195f2d0-4a44-7fc2-8f77-1f9c4cf7a005"
	privateSuggestedBullet := "PRIVATE_SUGGESTED_BULLET"
	store := &fakeTailorStore{ctx: resumestore.TailorJobContext{
		TailorRunID:       tailorRunID,
		UserID:            userID,
		ResumeVersionID:   versionID,
		ResumeAssetID:     assetID,
		TargetJobID:       targetID,
		Mode:              "gap_review",
		Language:          "en",
		ResumeSummary:     json.RawMessage(`{"headline":"Senior backend engineer"}`),
		StructuredProfile: json.RawMessage(`{"sections":[{"bullets":["Led migration."]}]}`),
		TargetSummary:     json.RawMessage(`{"requirements":["platform scale"]}`),
		TargetTitle:       "Staff Backend Engineer",
		TargetSeniority:   "staff",
		OriginalBullet:    "Led migration.",
	}}
	ai := &captureTailorAI{resp: aiclient.CompleteResponse{Content: `{
	  "matchSummary": {"strengths":["Strong systems evidence"],"gaps":["Add scale metrics"]},
	  "suggestions": [
	    {"originalBullet":"Led migration.","suggestedBullet":"Led migration across 12 teams.","reason":"Adds scope."},
	    {"originalBullet":"Built services.","suggestedBullet":"` + privateSuggestedBullet + `","reason":"Adds outcome."},
	    {"originalBullet":"Improved reliability.","suggestedBullet":"Improved reliability with SLOs.","reason":"Adds measurement."}
	  ]
	}`}}
	taskRuns := &memTaskRunWriter{}
	handler := resumejobs.NewTailorHandler(resumejobs.TailorHandlerOptions{
		Store:      store,
		Registry:   tailorRegistry{},
		AI:         ai,
		AITaskRuns: taskRuns,
		NewID: idSeq(
			"0195f2d0-4a44-7fc2-8f77-1f9c4cf7b001",
			"0195f2d0-4a44-7fc2-8f77-1f9c4cf7b002",
			"0195f2d0-4a44-7fc2-8f77-1f9c4cf7b003",
			"0195f2d0-4a44-7fc2-8f77-1f9c4cf7b004",
			"0195f2d0-4a44-7fc2-8f77-1f9c4cf7b005",
		),
		Now: func() time.Time { return now },
	})

	outcome := handler.Handle(context.Background(), targetjob.ClaimedJob{
		JobID: "0195f2d0-4a44-7fc2-8f77-1f9c4cf7c001", JobType: string(jobs.JobTypeResumeTailor), ResourceType: "resume_tailor_run", ResourceID: tailorRunID, Payload: []byte(`{"tailorRunId":"` + tailorRunID + `","resumeVersionId":"` + versionID + `"}`), Attempts: 1, MaxAttempts: 5,
	})

	if !outcome.Succeeded {
		t.Fatalf("Handle outcome = %+v", outcome)
	}
	if len(store.generating) != 1 || store.generating[0].TailorRunID != tailorRunID {
		t.Fatalf("generating calls = %+v", store.generating)
	}
	if store.loadedResumeVersionID != versionID {
		t.Fatalf("loaded resume version id = %q, want %q", store.loadedResumeVersionID, versionID)
	}
	if store.success == nil {
		t.Fatal("expected CompleteTailorRunSuccess")
	}
	if store.success.TailorRunID != tailorRunID || store.success.ResumeVersionID != versionID || len(store.success.Suggestions) != 3 {
		t.Fatalf("success input drift: %+v", store.success)
	}
	var outbox map[string]any
	if err := json.Unmarshal(store.success.OutboxEventPayload, &outbox); err != nil {
		t.Fatalf("decode outbox payload: %v", err)
	}
	wantKeys := map[string]bool{"tailorRunId": true, "resumeAssetId": true, "targetJobId": true, "mode": true, "status": true}
	if len(outbox) != len(wantKeys) {
		t.Fatalf("outbox field count = %d payload=%+v", len(outbox), outbox)
	}
	for key := range wantKeys {
		if _, ok := outbox[key]; !ok {
			t.Fatalf("outbox missing key %s: %+v", key, outbox)
		}
	}
	if outbox["tailorRunId"] != tailorRunID || outbox["status"] != "ready" || outbox["mode"] != "gap_review" {
		t.Fatalf("outbox identity/status drift: %+v", outbox)
	}
	if strings.Contains(string(store.success.OutboxEventPayload), privateSuggestedBullet) ||
		strings.Contains(string(store.success.OutboxEventPayload), "Strong systems evidence") ||
		strings.Contains(string(store.success.OutboxEventPayload), "Staff Backend Engineer") {
		t.Fatalf("outbox leaked private AI or prompt content: %s", store.success.OutboxEventPayload)
	}
	if ai.profileName != "resume.tailor.gap_review" ||
		ai.payload.Metadata.FeatureKey != resumejobs.FeatureKeyResumeTailorGapReview ||
		ai.payload.Metadata.TaskRun.Capability != aiclient.AITaskRunTaskResumeTailor ||
		ai.payload.Metadata.TaskRun.ResourceType != aiclient.AITaskRunResourceResumeTailorRun ||
		ai.payload.Metadata.TaskRun.ResourceID != tailorRunID ||
		ai.payload.Metadata.TaskRun.UserID != userID {
		t.Fatalf("AI metadata drift: profile=%q metadata=%+v", ai.profileName, ai.payload.Metadata)
	}
	rows := taskRuns.Rows()
	if len(rows) != 1 {
		t.Fatalf("expected one ai_task_runs row, got %+v", rows)
	}
	row := rows[0]
	if row.FeatureKey != resumejobs.FeatureKeyResumeTailorGapReview ||
		row.Capability != aiclient.AITaskRunTaskResumeTailor ||
		row.ResourceType != aiclient.AITaskRunResourceResumeTailorRun ||
		row.ResourceID != tailorRunID ||
		row.Status != aiclient.AITaskRunStatusSuccess ||
		row.OutputSchemaVersion != "resume.tailor.v1" {
		t.Fatalf("ai_task_runs row drift: %+v", row)
	}
}

func TestTailorHandlerModeRoutingAndFailurePaths(t *testing.T) {
	now := time.Date(2026, 5, 18, 12, 30, 0, 0, time.UTC)
	tailorRunID := "0195f2d0-4a44-7fc2-8f77-1f9c4cf8a001"
	userID := "0195f2d0-4a44-7fc2-8f77-1f9c4cf8a002"
	versionID := "0195f2d0-4a44-7fc2-8f77-1f9c4cf8a003"
	baseCtx := resumestore.TailorJobContext{
		TailorRunID:       tailorRunID,
		UserID:            userID,
		ResumeVersionID:   versionID,
		ResumeAssetID:     "0195f2d0-4a44-7fc2-8f77-1f9c4cf8a004",
		TargetJobID:       "0195f2d0-4a44-7fc2-8f77-1f9c4cf8a005",
		Mode:              "bullet_suggestions",
		Language:          "en",
		ResumeSummary:     json.RawMessage(`{"headline":"Engineer"}`),
		StructuredProfile: json.RawMessage(`{"sections":[{"bullets":["Built services."]}]}`),
		TargetSummary:     json.RawMessage(`{"requirements":["latency"]}`),
		OriginalBullet:    "Built services.",
	}
	cases := []struct {
		name          string
		ai            *captureTailorAI
		wantCode      string
		wantRetryable bool
		wantRunStatus aiclient.AITaskRunStatus
	}{
		{
			name: "bullet suggestions route",
			ai: &captureTailorAI{resp: aiclient.CompleteResponse{Content: `{
			  "matchSummary": {"strengths":["Has backend depth"],"gaps":[]},
			  "suggestions": [{"suggestedBullet":"Built low-latency services.","why_better":"Adds latency outcome.","kept_facts":["Built services"]}]
			}`}},
			wantRunStatus: aiclient.AITaskRunStatusSuccess,
		},
		{
			name:          "timeout marks failed retryable without completed outbox",
			ai:            &captureTailorAI{err: errors.New(sharederrors.CodeAiProviderTimeout + " provider slow")},
			wantCode:      sharederrors.CodeAiProviderTimeout,
			wantRetryable: true,
			wantRunStatus: aiclient.AITaskRunStatusTimeout,
		},
		{
			name:          "invalid output marks failed non retryable",
			ai:            &captureTailorAI{resp: aiclient.CompleteResponse{Content: `not-json`}},
			wantCode:      sharederrors.CodeAiOutputInvalid,
			wantRunStatus: aiclient.AITaskRunStatusFailed,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			store := &fakeTailorStore{ctx: baseCtx}
			taskRuns := &memTaskRunWriter{}
			handler := resumejobs.NewTailorHandler(resumejobs.TailorHandlerOptions{
				Store:      store,
				Registry:   tailorRegistry{},
				AI:         tc.ai,
				AITaskRuns: taskRuns,
				NewID: idSeq(
					"0195f2d0-4a44-7fc2-8f77-1f9c4cf8b001",
					"0195f2d0-4a44-7fc2-8f77-1f9c4cf8b002",
					"0195f2d0-4a44-7fc2-8f77-1f9c4cf8b003",
				),
				Now: func() time.Time { return now },
			})
			outcome := handler.Handle(context.Background(), targetjob.ClaimedJob{
				JobID: "0195f2d0-4a44-7fc2-8f77-1f9c4cf8c001", JobType: string(jobs.JobTypeResumeTailor), ResourceType: "resume_tailor_run", ResourceID: tailorRunID, Payload: []byte(`{"tailorRunId":"` + tailorRunID + `","resumeVersionId":"` + versionID + `"}`), Attempts: 1, MaxAttempts: 5,
			})

			rows := taskRuns.Rows()
			if len(rows) != 1 || rows[0].Status != tc.wantRunStatus {
				t.Fatalf("ai_task_runs rows = %+v, want status %s", rows, tc.wantRunStatus)
			}
			if tc.wantCode == "" {
				if !outcome.Succeeded || store.success == nil || store.failure != nil {
					t.Fatalf("success path outcome=%+v success=%+v failure=%+v", outcome, store.success, store.failure)
				}
				if tc.ai.profileName != "resume.tailor.bullet_suggestions" || tc.ai.payload.Metadata.FeatureKey != resumejobs.FeatureKeyResumeTailorBulletSuggestions {
					t.Fatalf("bullet route drift: profile=%q metadata=%+v", tc.ai.profileName, tc.ai.payload.Metadata)
				}
				if len(store.success.Suggestions) != 1 || store.success.Suggestions[0].OriginalBullet != "Built services." || store.success.Suggestions[0].SuggestedBullet != "Built low-latency services." {
					t.Fatalf("mapped suggestions = %+v", store.success.Suggestions)
				}
				return
			}
			if outcome.Succeeded || outcome.ErrorCode != tc.wantCode || outcome.Retryable != tc.wantRetryable {
				t.Fatalf("failure outcome = %+v, want code=%s retry=%v", outcome, tc.wantCode, tc.wantRetryable)
			}
			if store.failure == nil || store.failure.ErrorCode != tc.wantCode {
				t.Fatalf("failure input = %+v, want code %s", store.failure, tc.wantCode)
			}
			if store.success != nil {
				t.Fatalf("failure must not write ready outbox: %+v", store.success)
			}
		})
	}
}

func TestTailorDrainerHandlesOnlyResumeTailor(t *testing.T) {
	now := time.Date(2026, 5, 18, 13, 0, 0, 0, time.UTC)
	tailorRunID := "0195f2d0-4a44-7fc2-8f77-1f9c4cf9a001"
	asyncStore := &tailorAsyncStore{job: targetjob.ClaimedJob{
		JobID:        "0195f2d0-4a44-7fc2-8f77-1f9c4cf9a002",
		JobType:      string(jobs.JobTypeResumeTailor),
		ResourceType: "resume_tailor_run",
		ResourceID:   tailorRunID,
		Payload:      []byte(`{"tailorRunId":"` + tailorRunID + `","resumeVersionId":"0195f2d0-4a44-7fc2-8f77-1f9c4cf9a003"}`),
		Attempts:     1,
		MaxAttempts:  5,
	}}
	store := &fakeTailorStore{ctx: resumestore.TailorJobContext{
		TailorRunID:       tailorRunID,
		UserID:            "0195f2d0-4a44-7fc2-8f77-1f9c4cf9a004",
		ResumeVersionID:   "0195f2d0-4a44-7fc2-8f77-1f9c4cf9a003",
		ResumeAssetID:     "0195f2d0-4a44-7fc2-8f77-1f9c4cf9a005",
		TargetJobID:       "0195f2d0-4a44-7fc2-8f77-1f9c4cf9a006",
		Mode:              "gap_review",
		Language:          "en",
		ResumeSummary:     json.RawMessage(`{}`),
		StructuredProfile: json.RawMessage(`{"sections":[{"bullets":["Led migration."]}]}`),
		TargetSummary:     json.RawMessage(`{}`),
		OriginalBullet:    "Led migration.",
	}}
	handler := resumejobs.NewTailorHandler(resumejobs.TailorHandlerOptions{
		Store:    store,
		Registry: tailorRegistry{},
		AI: &captureTailorAI{resp: aiclient.CompleteResponse{Content: `{
		  "matchSummary": {"strengths":["Strong"],"gaps":[]},
		  "suggestions": [{"originalBullet":"Led migration.","suggestedBullet":"Led migration across teams.","reason":"Adds scope."}]
		}`}},
		AITaskRuns: &memTaskRunWriter{},
		NewID:      idSeq("0195f2d0-4a44-7fc2-8f77-1f9c4cf9b001", "0195f2d0-4a44-7fc2-8f77-1f9c4cf9b002", "0195f2d0-4a44-7fc2-8f77-1f9c4cf9b003"),
		Now:        func() time.Time { return now },
	})
	drainer := targetjob.NewDrainer(targetjob.DrainerOptions{
		Store: asyncStore,
		Handlers: map[string]targetjob.JobHandler{
			string(jobs.JobTypeResumeTailor): handler,
		},
		Now: func() time.Time { return now },
	})
	if !drainer.Handles(string(jobs.JobTypeResumeTailor)) || drainer.Handles(string(jobs.JobTypeResumeParse)) {
		t.Fatalf("drainer handler set drift")
	}

	processed, err := drainer.RunOnce(context.Background())
	if err != nil {
		t.Fatalf("RunOnce: %v", err)
	}
	if !processed || !asyncStore.outcome.Succeeded || store.success == nil {
		t.Fatalf("drainer did not complete resume_tailor: processed=%v outcome=%+v success=%+v", processed, asyncStore.outcome, store.success)
	}
}

type fakeTailorStore struct {
	ctx                   resumestore.TailorJobContext
	loadedResumeVersionID string
	generating            []resumestore.TailorRunStatusInput
	success               *resumestore.CompleteTailorRunSuccessInput
	failure               *resumestore.TailorRunFailureInput
}

type tailorAsyncStore struct {
	job     targetjob.ClaimedJob
	claimed bool
	outcome targetjob.JobOutcome
}

func (s *tailorAsyncStore) ClaimNextAsyncJob(_ context.Context, _ []string, _ time.Time) (targetjob.ClaimedJob, bool, error) {
	if s.claimed {
		return targetjob.ClaimedJob{}, false, nil
	}
	s.claimed = true
	return s.job, true, nil
}

func (s *tailorAsyncStore) FinalizeAsyncJob(_ context.Context, _ string, outcome targetjob.JobOutcome, _ time.Time) error {
	s.outcome = outcome
	return nil
}

func (s *fakeTailorStore) MarkTailorRunGenerating(_ context.Context, in resumestore.TailorRunStatusInput) (resumestore.TailorRunRecord, error) {
	s.generating = append(s.generating, in)
	return resumestore.TailorRunRecord{
		ID:            s.ctx.TailorRunID,
		UserID:        s.ctx.UserID,
		TargetJobID:   s.ctx.TargetJobID,
		ResumeAssetID: s.ctx.ResumeAssetID,
		Mode:          s.ctx.Mode,
		Status:        "generating",
	}, nil
}

func (s *fakeTailorStore) GetForTailor(_ context.Context, tailorRunID string, resumeVersionID string) (resumestore.TailorJobContext, error) {
	s.loadedResumeVersionID = resumeVersionID
	if s.ctx.TailorRunID != tailorRunID {
		return resumestore.TailorJobContext{}, resumestore.ErrTailorRunNotFound
	}
	out := s.ctx
	if strings.TrimSpace(out.ResumeVersionID) == "" {
		out.ResumeVersionID = resumeVersionID
	}
	return out, nil
}

func (s *fakeTailorStore) CompleteTailorRunSuccess(_ context.Context, in resumestore.CompleteTailorRunSuccessInput) error {
	cp := in
	s.success = &cp
	return nil
}

func (s *fakeTailorStore) MarkTailorRunFailed(_ context.Context, in resumestore.TailorRunFailureInput) (resumestore.TailorRunRecord, error) {
	cp := in
	s.failure = &cp
	return resumestore.TailorRunRecord{ID: s.ctx.TailorRunID, Status: "failed"}, nil
}

type tailorRegistry struct{}

func (tailorRegistry) Resolve(_ context.Context, featureKey string, language string) (resumejobs.PromptResolution, error) {
	if strings.TrimSpace(language) == "" {
		return resumejobs.PromptResolution{}, resumejobs.ErrPromptUnsupported
	}
	switch featureKey {
	case resumejobs.FeatureKeyResumeTailorGapReview:
		return resumejobs.PromptResolution{
			PromptVersion:       "v0.1.0",
			RubricVersion:       "v0.1.0",
			ModelProfileName:    "resume.tailor.gap_review",
			DataSourceVersion:   "target_job.v1",
			FeatureFlag:         "none",
			UserMessageTemplate: "Resume summary: {{resume_summary}}\nJD summary: {{jd_summary}}\nTarget seniority: {{target_seniority}}\nLanguage: {{language}}",
		}, nil
	case resumejobs.FeatureKeyResumeTailorBulletSuggestions:
		return resumejobs.PromptResolution{
			PromptVersion:       "v0.1.0",
			RubricVersion:       "v0.1.0",
			ModelProfileName:    "resume.tailor.bullet_suggestions",
			DataSourceVersion:   "target_job.v1",
			FeatureFlag:         "none",
			UserMessageTemplate: "Original bullet: {{original_bullet}}\nTarget context: {{jd_context}}\nTone: {{tone}}\nLanguage: {{language}}",
		}, nil
	default:
		return resumejobs.PromptResolution{}, resumejobs.ErrPromptUnsupported
	}
}

type captureTailorAI struct {
	profileName string
	payload     aiclient.CompletePayload
	resp        aiclient.CompleteResponse
	meta        aiclient.AICallMeta
	err         error
}

func (c *captureTailorAI) Complete(_ context.Context, profileName string, payload aiclient.CompletePayload) (aiclient.CompleteResponse, aiclient.AICallMeta, error) {
	c.profileName = profileName
	c.payload = payload
	meta := c.meta
	if meta.ModelID == "" {
		meta = aiclient.AICallMeta{
			Provider:            "stub",
			ModelFamily:         "stub",
			ModelID:             "fixture-model:resume-tailor",
			ModelProfileName:    profileName,
			ModelProfileVersion: "1.0.0",
			FallbackChain:       []string{"stub/fixture-model:resume-tailor"},
			ValidationStatus:    aiclient.ValidationStatusOK,
		}
	}
	return c.resp, meta, c.err
}

func (c *captureTailorAI) Transcribe(context.Context, string, aiclient.TranscriptionInput) (aiclient.TranscriptionResponse, aiclient.AICallMeta, error) {
	return aiclient.TranscriptionResponse{}, aiclient.AICallMeta{}, errors.New("not implemented")
}

func (c *captureTailorAI) Stream(context.Context, string, aiclient.CompletePayload) (<-chan aiclient.AIStreamEvent, error) {
	return nil, errors.New("not implemented")
}

func (c *captureTailorAI) Synthesize(context.Context, string, aiclient.SynthesisInput) (aiclient.SynthesisResponse, aiclient.AICallMeta, error) {
	return aiclient.SynthesisResponse{}, aiclient.AICallMeta{}, errors.New("not implemented")
}
