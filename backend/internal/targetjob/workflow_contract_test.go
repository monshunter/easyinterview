package targetjob_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient"
	"github.com/monshunter/easyinterview/backend/internal/runner"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
	"github.com/monshunter/easyinterview/backend/internal/targetjob"
)

const (
	contractUserID   = "018f2a40-0000-7000-9000-0000000000b1"
	contractResumeID = "018f2a40-0000-7000-9000-0000000000r1"
	contractTargetID = "018f2a40-0000-7000-9000-0000000000a1"
)

func TestRawTextImportIdempotencyAndParseReady(t *testing.T) {
	svc, store := newServiceWithFake(
		contractTargetID,
		"018f2a40-0000-7000-9000-0000000000f1",
		"018f2a40-0000-7000-9000-0000000000e1",
		"018f2a40-0000-7000-9000-0000000000a2",
		"018f2a40-0000-7000-9000-0000000000f2",
		"018f2a40-0000-7000-9000-0000000000e2",
	)
	request := targetjob.ImportRequest{
		UserID:         contractUserID,
		IdempotencyKey: "targetjob-contract-import",
		TargetLanguage: "zh-CN",
		ResumeID:       contractResumeID,
		RawText:        "  We are hiring a Senior Frontend Engineer to lead React platform work.  ",
	}
	imported, err := svc.ImportTargetJob(context.Background(), request)
	if err != nil {
		t.Fatalf("rawText import: %v", err)
	}
	if imported.TargetJobID != contractTargetID || imported.Job.Status != sharedtypes.JobStatusQueued {
		t.Fatalf("unexpected import response: %+v", imported)
	}
	if store.captured.RawJDText != strings.TrimSpace(request.RawText) || store.captured.ResumeID != contractResumeID {
		t.Fatalf("paste-only import did not preserve rawText/resume binding: %+v", store.captured)
	}
	assertPayloadOmits(t, store.captured.OutboxEventPayload, request.RawText, "Senior Frontend Engineer", "sourceType", "sourceUrl")
	assertPayloadOmits(t, store.captured.JobPayload, request.RawText, "Senior Frontend Engineer", "sourceType", "sourceUrl")
	firstDedupeKey := store.captured.DedupeKey

	createdAt, err := time.Parse(time.RFC3339, imported.Job.CreatedAt)
	if err != nil {
		t.Fatalf("parse imported job createdAt: %v", err)
	}
	updatedAt, err := time.Parse(time.RFC3339, imported.Job.UpdatedAt)
	if err != nil {
		t.Fatalf("parse imported job updatedAt: %v", err)
	}
	store.result = targetjob.ImportTargetJobResult{
		TargetJobID:  imported.TargetJobID,
		JobID:        imported.Job.Id,
		JobStatus:    imported.Job.Status,
		JobCreatedAt: createdAt,
		JobUpdatedAt: updatedAt,
		Existing:     true,
	}
	replayed, err := svc.ImportTargetJob(context.Background(), request)
	if err != nil {
		t.Fatalf("idempotent replay: %v", err)
	}
	if replayed.TargetJobID != imported.TargetJobID || replayed.Job.Id != imported.Job.Id || store.captured.DedupeKey != firstDedupeKey {
		t.Fatalf("idempotent replay drifted: first=%+v replay=%+v dedupe=%q/%q", imported, replayed, firstDedupeKey, store.captured.DedupeKey)
	}

	exec, parseStore, _, ai := newParseExecutorWithFakes(t)
	ai.resp = aiclient.CompleteResponse{Content: happyResponseJSON}
	parseStore.target = targetjob.TargetJobRecord{
		ID:             imported.TargetJobID,
		UserID:         contractUserID,
		TargetLanguage: "zh-CN",
		RawJDText:      strings.TrimSpace(request.RawText),
	}
	outcome := exec.Handle(context.Background(), runner.ClaimedJob{
		JobID:        imported.Job.Id,
		JobType:      "target_import",
		ResourceType: "target_job",
		ResourceID:   imported.TargetJobID,
	})
	if !outcome.Succeeded {
		t.Fatalf("parse outcome = %+v", outcome)
	}
	if parseStore.completeSuccessIn == nil || parseStore.completeSuccessIn.AnalysisStatus != sharedtypes.TargetJobParseStatusReady || len(parseStore.completeSuccessIn.Requirements) == 0 {
		t.Fatalf("parse result did not atomically commit ready state: %+v", parseStore.completeSuccessIn)
	}
	assertPayloadOmits(t, parseStore.parsedOutboxPayload, request.RawText, "prompt", "response", "sourceType", "sourceUrl")

	now := time.Date(2026, 5, 9, 23, 0, 0, 0, time.UTC)
	store.getRecord = targetjob.TargetJobRecord{
		ID:             imported.TargetJobID,
		UserID:         contractUserID,
		Status:         sharedtypes.TargetJobStatusDraft,
		AnalysisStatus: sharedtypes.TargetJobParseStatusReady,
		Title:          parseStore.completeSuccessIn.Title,
		CompanyName:    parseStore.completeSuccessIn.CompanyName,
		TargetLanguage: "zh-CN",
		Summary:        parseStore.completeSuccessIn.Summary,
		FitSummary:     parseStore.completeSuccessIn.FitSummary,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	store.getRequirements = parseStore.completeSuccessIn.Requirements
	detail, err := svc.GetTargetJob(context.Background(), contractUserID, imported.TargetJobID)
	if err != nil {
		t.Fatalf("get target job: %v", err)
	}
	if detail.AnalysisStatus != sharedtypes.TargetJobParseStatusReady || len(detail.Requirements) == 0 || detail.Summary == nil || detail.FitSummary == nil {
		t.Fatalf("detail does not expose ready parse result: %+v", detail)
	}
}

func TestParseFailureRetryableAndNonRetryable(t *testing.T) {
	cases := []struct {
		name      string
		configure func(*fakeRegistry, *fakeAIClient)
		code      string
		retryable bool
	}{
		{
			name: "provider-timeout",
			configure: func(_ *fakeRegistry, ai *fakeAIClient) {
				ai.err = errors.New("upstream failed with AI_PROVIDER_TIMEOUT")
			},
			code:      sharederrors.CodeAiProviderTimeout,
			retryable: true,
		},
		{
			name: "invalid-output",
			configure: func(_ *fakeRegistry, ai *fakeAIClient) {
				ai.resp = aiclient.CompleteResponse{Content: "not-json"}
			},
			code:      sharederrors.CodeAiOutputInvalid,
			retryable: false,
		},
		{
			name: "registry-disabled",
			configure: func(registry *fakeRegistry, _ *fakeAIClient) {
				registry.err = targetjob.ErrPromptUnsupported
			},
			code:      sharederrors.CodeAiProviderConfigInvalid,
			retryable: false,
		},
		{
			name: "secret-missing",
			configure: func(_ *fakeRegistry, ai *fakeAIClient) {
				ai.err = errors.New("AI_PROVIDER_SECRET_MISSING")
			},
			code:      sharederrors.CodeAiProviderSecretMissing,
			retryable: false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			exec, store, registry, ai := newParseExecutorWithFakes(t)
			store.target = targetjob.TargetJobRecord{
				ID:             "target-" + tc.name,
				UserID:         contractUserID,
				TargetLanguage: "en",
				RawJDText:      "Private JD body that must not leak into failure evidence.",
			}
			tc.configure(registry, ai)
			outcome := exec.Handle(context.Background(), runner.ClaimedJob{
				JobID: "job-" + tc.name, JobType: "target_import", ResourceType: "target_job", ResourceID: store.target.ID,
			})
			if outcome.Succeeded || outcome.ErrorCode != tc.code || outcome.Retryable != tc.retryable {
				t.Fatalf("unexpected failure outcome: %+v", outcome)
			}
			if store.failedOutboxPayload == nil {
				t.Fatal("target.analysis.failed payload missing")
			}
			assertPayloadOmits(t, store.failedOutboxPayload, "Private JD body", "Authorization:", "provider secret", "sourceType", "sourceUrl")
		})
	}
}

func assertPayloadOmits(t *testing.T, payload []byte, forbidden ...string) {
	t.Helper()
	raw := string(payload)
	for _, token := range forbidden {
		if token != "" && strings.Contains(raw, token) {
			t.Fatalf("payload leaked forbidden token %q: %s", token, raw)
		}
	}
}
