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
	"github.com/monshunter/easyinterview/backend/internal/runner"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
	"github.com/monshunter/easyinterview/backend/internal/shared/jobs"
)

func TestTailorHandlerHappyPathWritesReadySuggestionsTaskRunAndPrivateOutbox(t *testing.T) {
	now := time.Date(2026, 6, 13, 12, 0, 0, 0, time.UTC)
	tailorRunID := "0195f2d0-4a44-7fc2-8f77-1f9c4cf7a001"
	userID := "0195f2d0-4a44-7fc2-8f77-1f9c4cf7a002"
	resumeID := "0195f2d0-4a44-7fc2-8f77-1f9c4cf7a004"
	targetID := "0195f2d0-4a44-7fc2-8f77-1f9c4cf7a005"
	privateSuggestedBullet := "PRIVATE_SUGGESTED_BULLET"
	store := &fakeTailorStore{ctx: resumestore.TailorJobContext{
		TailorRunID:       tailorRunID,
		UserID:            userID,
		ResumeID:          resumeID,
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

	outcome := handler.Handle(context.Background(), runner.ClaimedJob{
		JobID: "0195f2d0-4a44-7fc2-8f77-1f9c4cf7c001", JobType: string(jobs.JobTypeResumeTailor), ResourceType: "resume_tailor_run", ResourceID: tailorRunID, Payload: []byte(`{"resumeId":"` + resumeID + `","targetJobId":"` + targetID + `","mode":"gap_review"}`), Attempts: 1, MaxAttempts: 5,
	})

	if !outcome.Succeeded {
		t.Fatalf("Handle outcome = %+v", outcome)
	}
	if store.loadedTailorRunID != tailorRunID {
		t.Fatalf("loaded tailor run id = %q, want %q", store.loadedTailorRunID, tailorRunID)
	}
	if store.success == nil {
		t.Fatal("expected CompleteTailorRunSuccess")
	}
	if store.success.TailorRunID != tailorRunID || store.success.ResumeID != resumeID || store.success.Mode != "gap_review" || len(store.success.Suggestions) != 3 {
		t.Fatalf("success input drift: %+v", store.success)
	}
	if store.success.JobID != "0195f2d0-4a44-7fc2-8f77-1f9c4cf7c001" || store.success.ClaimedAttempts != 1 {
		t.Fatalf("claimed lease generation was not threaded to success transaction: %+v", store.success)
	}
	var outbox map[string]any
	if err := json.Unmarshal(store.success.OutboxEventPayload, &outbox); err != nil {
		t.Fatalf("decode outbox payload: %v", err)
	}
	wantKeys := map[string]bool{"tailorRunId": true, "resumeId": true, "targetJobId": true, "mode": true, "status": true}
	if len(outbox) != len(wantKeys) {
		t.Fatalf("outbox field count = %d payload=%+v", len(outbox), outbox)
	}
	for key := range wantKeys {
		if _, ok := outbox[key]; !ok {
			t.Fatalf("outbox missing key %s: %+v", key, outbox)
		}
	}
	if outbox["tailorRunId"] != tailorRunID || outbox["resumeId"] != resumeID || outbox["status"] != "ready" || outbox["mode"] != "gap_review" {
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
	if len(ai.payload.Metadata.OutputSchema) == 0 {
		t.Fatalf("AI metadata OutputSchema must be populated")
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

func TestOutboxPrivacyForTailorCompletedEvent(t *testing.T) {
	fixture := runTailorPrivacyFixture(t)
	if fixture.store.success == nil {
		t.Fatal("expected completed tailor run")
	}
	raw := string(fixture.store.success.OutboxEventPayload)
	assertNoTailorPrivacyLeak(t, "outbox payload", raw)

	var payload map[string]any
	if err := json.Unmarshal(fixture.store.success.OutboxEventPayload, &payload); err != nil {
		t.Fatalf("decode outbox payload: %v", err)
	}
	wantKeys := map[string]bool{"tailorRunId": true, "resumeId": true, "targetJobId": true, "mode": true, "status": true}
	if len(payload) != len(wantKeys) {
		t.Fatalf("outbox payload fields drifted: %+v", payload)
	}
	for key := range wantKeys {
		if _, ok := payload[key]; !ok {
			t.Fatalf("outbox payload missing %s: %+v", key, payload)
		}
	}
}

func TestAiTaskRunsPrivacyForTailorHandler(t *testing.T) {
	fixture := runTailorPrivacyFixture(t)
	rows := fixture.taskRuns.Rows()
	if len(rows) != 1 {
		t.Fatalf("ai_task_runs rows = %+v, want one row", rows)
	}
	raw, err := json.Marshal(rows[0])
	if err != nil {
		t.Fatalf("marshal ai_task_runs row: %v", err)
	}
	assertNoTailorPrivacyLeak(t, "ai_task_runs row", string(raw))
	if rows[0].RawResponseObjectKey != "" {
		t.Fatalf("tailor task run must not point to raw model response object: %+v", rows[0])
	}
}

func TestAuditPrivacyForTailorHandler(t *testing.T) {
	fixture := runTailorPrivacyFixture(t)
	rows := fixture.taskRuns.Rows()
	if len(rows) != 1 {
		t.Fatalf("ai_task_runs rows = %+v, want one row", rows)
	}
	raw, err := json.Marshal(rows[0].Metadata)
	if err != nil {
		t.Fatalf("marshal audit metadata: %v", err)
	}
	assertNoTailorPrivacyLeak(t, "audit metadata", string(raw))
	if rows[0].Metadata.PromptHash != "" ||
		rows[0].Metadata.ResponseHash != "" ||
		rows[0].Metadata.PromptCharLength != 0 ||
		rows[0].Metadata.ResponseCharLength != 0 ||
		rows[0].Metadata.ProfileName != "" {
		t.Fatalf("tailor handler should not persist prompt/response audit metadata directly: %+v", rows[0].Metadata)
	}
}

func TestTailorHandlerModeRoutingAndFailurePaths(t *testing.T) {
	now := time.Date(2026, 6, 13, 12, 30, 0, 0, time.UTC)
	tailorRunID := "0195f2d0-4a44-7fc2-8f77-1f9c4cf8a001"
	userID := "0195f2d0-4a44-7fc2-8f77-1f9c4cf8a002"
	resumeID := "0195f2d0-4a44-7fc2-8f77-1f9c4cf8a004"
	baseCtx := resumestore.TailorJobContext{
		TailorRunID:       tailorRunID,
		UserID:            userID,
		ResumeID:          resumeID,
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
			  "suggestions": [{"originalBullet":"Built services.","suggestedBullet":"Built low-latency services.","reason":"Adds latency outcome."}]
			}`}},
			wantRunStatus: aiclient.AITaskRunStatusSuccess,
		},
		{
			name: "root match summary aliases are rejected",
			ai: &captureTailorAI{resp: aiclient.CompleteResponse{Content: `{
			  "strengths_to_amplify": [{"topic":"Backend depth","evidence":"Built services"}],
			  "gaps": [{"topic":"Scale","why":"Missing metrics"}]
			}`}},
			wantCode:      sharederrors.CodeAiOutputInvalid,
			wantRunStatus: aiclient.AITaskRunStatusFailed,
		},
		{
			name: "suggestion aliases are rejected",
			ai: &captureTailorAI{resp: aiclient.CompleteResponse{Content: `{
			  "suggestions": [{"original_bullet":"Built services.","rewrite":"Built low-latency services.","why_better":"Adds latency outcome."}]
			}`}},
			wantCode:      sharederrors.CodeAiOutputInvalid,
			wantRunStatus: aiclient.AITaskRunStatusFailed,
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
			outcome := handler.Handle(context.Background(), runner.ClaimedJob{
				JobID: "0195f2d0-4a44-7fc2-8f77-1f9c4cf8c001", JobType: string(jobs.JobTypeResumeTailor), ResourceType: "resume_tailor_run", ResourceID: tailorRunID, Payload: []byte(`{"resumeId":"` + resumeID + `","mode":"bullet_suggestions"}`), Attempts: 1, MaxAttempts: 5,
			})

			rows := taskRuns.Rows()
			if len(rows) != 1 || rows[0].Status != tc.wantRunStatus {
				t.Fatalf("ai_task_runs rows = %+v, want status %s", rows, tc.wantRunStatus)
			}
			if tc.wantCode == "" {
				if !outcome.Succeeded || store.success == nil {
					t.Fatalf("success path outcome=%+v success=%+v", outcome, store.success)
				}
				if tc.ai.profileName != "resume.tailor.bullet_suggestions" || tc.ai.payload.Metadata.FeatureKey != resumejobs.FeatureKeyResumeTailorBulletSuggestions {
					t.Fatalf("bullet route drift: profile=%q metadata=%+v", tc.ai.profileName, tc.ai.payload.Metadata)
				}
				if len(tc.ai.payload.Metadata.OutputSchema) == 0 {
					t.Fatalf("bullet route metadata OutputSchema must be populated")
				}
				if len(store.success.Suggestions) != 1 || store.success.Suggestions[0].OriginalBullet != "Built services." || store.success.Suggestions[0].SuggestedBullet != "Built low-latency services." {
					t.Fatalf("mapped suggestions = %+v", store.success.Suggestions)
				}
				return
			}
			// Failure paths surface through JobOutcome; the runner kernel marks
			// the async_jobs row failed (D-20: no resume_tailor_runs status to flip).
			if outcome.Succeeded || outcome.ErrorCode != tc.wantCode || outcome.Retryable != tc.wantRetryable {
				t.Fatalf("failure outcome = %+v, want code=%s retry=%v", outcome, tc.wantCode, tc.wantRetryable)
			}
			if store.success != nil {
				t.Fatalf("failure must not write ready outbox: %+v", store.success)
			}
		})
	}
}

func TestTailorHandlerBulletSuggestionsCanonicalKeysRoundTrip(t *testing.T) {
	now := time.Date(2026, 6, 13, 12, 40, 0, 0, time.UTC)
	tailorRunID := "0195f2d0-4a44-7fc2-8f77-1f9c4cf8ca01"
	resumeID := "0195f2d0-4a44-7fc2-8f77-1f9c4cf8ca04"
	store := &fakeTailorStore{ctx: resumestore.TailorJobContext{
		TailorRunID:       tailorRunID,
		UserID:            "0195f2d0-4a44-7fc2-8f77-1f9c4cf8ca03",
		ResumeID:          resumeID,
		TargetJobID:       "0195f2d0-4a44-7fc2-8f77-1f9c4cf8ca05",
		Mode:              "bullet_suggestions",
		Language:          "en",
		ResumeSummary:     json.RawMessage(`{"headline":"Engineer"}`),
		StructuredProfile: json.RawMessage(`{"sections":[{"bullets":["Built APIs."]}]}`),
		TargetSummary:     json.RawMessage(`{"requirements":["platform reliability"]}`),
		OriginalBullet:    "Built APIs.",
	}}
	handler := resumejobs.NewTailorHandler(resumejobs.TailorHandlerOptions{
		Store:    store,
		Registry: tailorRegistry{},
		AI: &captureTailorAI{resp: aiclient.CompleteResponse{Content: `{
		  "suggestions": [
		    {
		      "originalBullet":"Built APIs.",
		      "original_bullet":"ALIAS ORIGINAL",
		      "suggestedBullet":"Built reliable APIs for onboarding.",
		      "suggested_bullet":"ALIAS SUGGESTION",
		      "rewrite":"ALIAS REWRITE",
		      "reason":"Adds reliability and product scope.",
		      "why_better":"ALIAS REASON",
		      "whyBetter":"ALIAS CAMEL REASON"
		    }
		  ]
		}`}},
		AITaskRuns: &memTaskRunWriter{},
		NewID:      idSeq("0195f2d0-4a44-7fc2-8f77-1f9c4cf8ca06", "0195f2d0-4a44-7fc2-8f77-1f9c4cf8ca07"),
		Now:        func() time.Time { return now },
	})

	outcome := handler.Handle(context.Background(), runner.ClaimedJob{
		JobID:        "0195f2d0-4a44-7fc2-8f77-1f9c4cf8ca08",
		JobType:      string(jobs.JobTypeResumeTailor),
		ResourceType: "resume_tailor_run",
		ResourceID:   tailorRunID,
		Payload:      []byte(`{"resumeId":"` + resumeID + `","mode":"bullet_suggestions"}`),
		Attempts:     1,
		MaxAttempts:  5,
	})

	if !outcome.Succeeded || store.success == nil {
		t.Fatalf("Handle outcome=%+v success=%+v", outcome, store.success)
	}
	if len(store.success.Suggestions) != 1 {
		t.Fatalf("suggestions = %+v", store.success.Suggestions)
	}
	got := store.success.Suggestions[0]
	if got.OriginalBullet != "Built APIs." ||
		got.SuggestedBullet != "Built reliable APIs for onboarding." ||
		got.Reason != "Adds reliability and product scope." {
		t.Fatalf("canonical suggestion did not round trip: %+v", got)
	}
}

func TestTailorHandlerSuccessPersistenceFailureMarksRetryable(t *testing.T) {
	now := time.Date(2026, 6, 13, 12, 45, 0, 0, time.UTC)
	tailorRunID := "0195f2d0-4a44-7fc2-8f77-1f9c4cf8d001"
	resumeID := "0195f2d0-4a44-7fc2-8f77-1f9c4cf8d004"
	store := &fakeTailorStore{
		ctx: resumestore.TailorJobContext{
			TailorRunID:       tailorRunID,
			UserID:            "0195f2d0-4a44-7fc2-8f77-1f9c4cf8d003",
			ResumeID:          resumeID,
			TargetJobID:       "0195f2d0-4a44-7fc2-8f77-1f9c4cf8d005",
			Mode:              "gap_review",
			Language:          "en",
			ResumeSummary:     json.RawMessage(`{"headline":"Engineer"}`),
			StructuredProfile: json.RawMessage(`{"sections":[{"bullets":["Built services."]}]}`),
			TargetSummary:     json.RawMessage(`{"requirements":["latency"]}`),
			OriginalBullet:    "Built services.",
		},
		completeSuccessErr: errors.New("outbox unavailable"),
	}
	handler := resumejobs.NewTailorHandler(resumejobs.TailorHandlerOptions{
		Store:    store,
		Registry: tailorRegistry{},
		AI: &captureTailorAI{resp: aiclient.CompleteResponse{Content: `{
		  "matchSummary": {"strengths":["Has backend depth"],"gaps":[]},
		  "suggestions": [{"originalBullet":"Built services.","suggestedBullet":"Built reliable services.","reason":"Adds outcome."}]
		}`}},
		AITaskRuns: &memTaskRunWriter{},
		NewID:      idSeq("0195f2d0-4a44-7fc2-8f77-1f9c4cf8e001", "0195f2d0-4a44-7fc2-8f77-1f9c4cf8e002", "0195f2d0-4a44-7fc2-8f77-1f9c4cf8e003"),
		Now:        func() time.Time { return now },
	})

	outcome := handler.Handle(context.Background(), runner.ClaimedJob{
		JobID:        "0195f2d0-4a44-7fc2-8f77-1f9c4cf8f001",
		JobType:      string(jobs.JobTypeResumeTailor),
		ResourceType: "resume_tailor_run",
		ResourceID:   tailorRunID,
		Payload:      []byte(`{"resumeId":"` + resumeID + `","mode":"gap_review"}`),
		Attempts:     1,
		MaxAttempts:  5,
	})

	if outcome.Succeeded || outcome.ErrorCode != sharederrors.CodeTargetImportFailed || !outcome.Retryable {
		t.Fatalf("completion persistence failure outcome = %+v", outcome)
	}
	if store.success == nil {
		t.Fatal("expected success transaction attempt before failure")
	}
}

type fakeTailorStore struct {
	ctx                resumestore.TailorJobContext
	loadedTailorRunID  string
	success            *resumestore.CompleteTailorRunSuccessInput
	completeSuccessErr error
}

func (s *fakeTailorStore) GetForTailor(_ context.Context, tailorRunID string) (resumestore.TailorJobContext, error) {
	s.loadedTailorRunID = tailorRunID
	if s.ctx.TailorRunID != tailorRunID {
		return resumestore.TailorJobContext{}, resumestore.ErrTailorRunNotFound
	}
	return s.ctx, nil
}

func (s *fakeTailorStore) CompleteTailorRunSuccess(_ context.Context, in resumestore.CompleteTailorRunSuccessInput) error {
	cp := in
	s.success = &cp
	return s.completeSuccessErr
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
			OutputSchema:        rawSchema(`{"type":"object","required":["matchSummary"],"properties":{"matchSummary":{"type":"object"}}}`),
			UserMessageTemplate: "Resume summary: {{resume_summary}}\nJD summary: {{jd_summary}}\nTarget seniority: {{target_seniority}}\nLanguage: {{language}}",
		}, nil
	case resumejobs.FeatureKeyResumeTailorBulletSuggestions:
		return resumejobs.PromptResolution{
			PromptVersion:       "v0.1.0",
			RubricVersion:       "v0.1.0",
			ModelProfileName:    "resume.tailor.bullet_suggestions",
			DataSourceVersion:   "target_job.v1",
			FeatureFlag:         "none",
			OutputSchema:        rawSchema(`{"type":"object","required":["suggestions"],"properties":{"suggestions":{"type":"array"}}}`),
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
	return aiclient.TranscriptionResponse{}, aiclient.AICallMeta{}, errors.New("unexpected Transcribe call in resume tailor captureTailorAI")
}

func (c *captureTailorAI) Stream(context.Context, string, aiclient.CompletePayload) (<-chan aiclient.AIStreamEvent, error) {
	return nil, errors.New("unexpected Stream call in resume tailor captureTailorAI")
}

func (c *captureTailorAI) Synthesize(context.Context, string, aiclient.SynthesisInput) (aiclient.SynthesisResponse, aiclient.AICallMeta, error) {
	return aiclient.SynthesisResponse{}, aiclient.AICallMeta{}, errors.New("unexpected Synthesize call in resume tailor captureTailorAI")
}

type tailorPrivacyFixture struct {
	store    *fakeTailorStore
	taskRuns *memTaskRunWriter
}

func runTailorPrivacyFixture(t *testing.T) tailorPrivacyFixture {
	t.Helper()
	now := time.Date(2026, 6, 13, 14, 0, 0, 0, time.UTC)
	tailorRunID := "0195f2d0-4a44-7fc2-8f77-1f9c4cfaa001"
	userID := "0195f2d0-4a44-7fc2-8f77-1f9c4cfaa002"
	resumeID := "0195f2d0-4a44-7fc2-8f77-1f9c4cfaa004"
	store := &fakeTailorStore{ctx: resumestore.TailorJobContext{
		TailorRunID:       tailorRunID,
		UserID:            userID,
		ResumeID:          resumeID,
		TargetJobID:       "0195f2d0-4a44-7fc2-8f77-1f9c4cfaa005",
		Mode:              "gap_review",
		Language:          "en",
		ResumeSummary:     json.RawMessage(`{"headline":"PRIVATE_RESUME_SUMMARY"}`),
		StructuredProfile: json.RawMessage(`{"sections":[{"bullets":["PRIVATE_STRUCTURED_PROFILE"]}]}`),
		TargetSummary:     json.RawMessage(`{"requirements":["PRIVATE_JD_CONTEXT"]}`),
		TargetTitle:       "PRIVATE_TARGET_TITLE",
		RawJDText:         "PRIVATE_PROMPT_BODY",
		OriginalBullet:    "PRIVATE_ORIGINAL_BULLET",
	}}
	taskRuns := &memTaskRunWriter{}
	handler := resumejobs.NewTailorHandler(resumejobs.TailorHandlerOptions{
		Store:    store,
		Registry: tailorRegistry{},
		AI: &captureTailorAI{resp: aiclient.CompleteResponse{Content: `{
		  "matchSummary": {"strengths":["PRIVATE_MATCH_SUMMARY"],"gaps":["PRIVATE_MODEL_RAW_RESPONSE"]},
		  "suggestions": [{"originalBullet":"PRIVATE_ORIGINAL_BULLET","suggestedBullet":"PRIVATE_SUGGESTED_BULLET","reason":"PRIVATE_SUGGESTION_REASON"}]
		}`}},
		AITaskRuns: taskRuns,
		NewID: idSeq(
			"0195f2d0-4a44-7fc2-8f77-1f9cfaa0b001",
			"0195f2d0-4a44-7fc2-8f77-1f9cfaa0b002",
			"0195f2d0-4a44-7fc2-8f77-1f9cfaa0b003",
		),
		Now: func() time.Time { return now },
	})
	outcome := handler.Handle(context.Background(), runner.ClaimedJob{
		JobID:        "0195f2d0-4a44-7fc2-8f77-1f9c4cfaa006",
		JobType:      string(jobs.JobTypeResumeTailor),
		ResourceType: "resume_tailor_run",
		ResourceID:   tailorRunID,
		Payload:      []byte(`{"resumeId":"` + resumeID + `","mode":"gap_review"}`),
		Attempts:     1,
		MaxAttempts:  5,
	})
	if !outcome.Succeeded {
		t.Fatalf("Handle outcome = %+v", outcome)
	}
	return tailorPrivacyFixture{store: store, taskRuns: taskRuns}
}

func assertNoTailorPrivacyLeak(t *testing.T, label string, raw string) {
	t.Helper()
	for _, forbidden := range []string{
		"PRIVATE_RESUME_SUMMARY",
		"PRIVATE_STRUCTURED_PROFILE",
		"PRIVATE_JD_CONTEXT",
		"PRIVATE_TARGET_TITLE",
		"PRIVATE_PROMPT_BODY",
		"PRIVATE_ORIGINAL_BULLET",
		"PRIVATE_MATCH_SUMMARY",
		"PRIVATE_MODEL_RAW_RESPONSE",
		"PRIVATE_SUGGESTED_BULLET",
		"PRIVATE_SUGGESTION_REASON",
		"prompt body",
		"model raw response",
		"suggested bullet",
	} {
		if strings.Contains(raw, forbidden) {
			t.Fatalf("%s leaked %q: %s", label, forbidden, raw)
		}
	}
}
