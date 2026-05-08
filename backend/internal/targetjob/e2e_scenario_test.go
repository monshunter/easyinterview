package targetjob_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
	"github.com/monshunter/easyinterview/backend/internal/targetjob"
	"github.com/monshunter/easyinterview/backend/internal/targetjob/urlfetch"
)

const (
	scenarioUserID   = "018f2a40-0000-7000-9000-0000000000b1"
	scenarioTargetID = "018f2a40-0000-7000-9000-0000000000a1"
)

func TestE2EP0010TextImportParseReady(t *testing.T) {
	svc, store := newServiceWithFake(
		scenarioTargetID,
		"018f2a40-0000-7000-9000-0000000000f1",
		"018f2a40-0000-7000-9000-0000000000c1",
		"018f2a40-0000-7000-9000-0000000000e1",
		"018f2a40-0000-7000-9000-0000000000d1",
	)
	imported, err := svc.ImportTargetJob(context.Background(), targetjob.ImportRequest{
		UserID:          scenarioUserID,
		IdempotencyKey:  "e2e-p0-010-import",
		TargetLanguage:  "zh-CN",
		TitleHint:       "Senior Frontend Engineer",
		CompanyNameHint: "Acme",
		Source: map[string]any{
			"type":    "manual_text",
			"rawText": "We are hiring a Senior Frontend Engineer to lead React platform work.",
		},
	})
	if err != nil {
		t.Fatalf("manual_text import: %v", err)
	}
	if imported.TargetJobID != scenarioTargetID || imported.Job.Status != sharedtypes.JobStatusQueued {
		t.Fatalf("unexpected import response: %+v", imported)
	}
	assertPayloadOmits(t, store.captured.OutboxEventPayload, "Senior Frontend Engineer", "React platform")
	var requested map[string]any
	if err := json.Unmarshal(store.captured.OutboxEventPayload, &requested); err != nil {
		t.Fatalf("decode target.import.requested payload: %v", err)
	}
	if requested["sourceType"] != "text" {
		t.Fatalf("manual_text import must emit event sourceType=text, got %v", requested["sourceType"])
	}

	exec, parseStore, _, ai, _ := newParseExecutorWithFakes(t)
	ai.resp = aiclient.CompleteResponse{Content: happyResponseJSON}
	parseStore.target = targetjob.TargetJobRecord{
		ID:             imported.TargetJobID,
		UserID:         scenarioUserID,
		SourceType:     targetjob.SourceTypeManualText,
		TargetLanguage: "zh-CN",
		RawJDText:      "Lead React platform and design system programs.",
	}
	outcome := exec.Handle(context.Background(), targetjob.ClaimedJob{
		JobID: "018f2a40-0000-7000-9000-0000000000f1", JobType: "target_import", ResourceType: "target_job", ResourceID: imported.TargetJobID,
	})
	if !outcome.Succeeded {
		t.Fatalf("parse outcome = %+v", outcome)
	}
	if parseStore.applyResultIn == nil || parseStore.applyResultIn.LatestParseJobID != "018f2a40-0000-7000-9000-0000000000f1" {
		t.Fatalf("parse result did not persist latest job id: %+v", parseStore.applyResultIn)
	}
	assertPayloadOmits(t, parseStore.parsedOutboxPayload, "Lead React platform", "prompt", "response")
	if !parseStore.sourceRefreshCalled {
		t.Fatal("target.parsed must enqueue source_refresh placeholder")
	}

	now := time.Date(2026, 5, 9, 23, 0, 0, 0, time.UTC)
	store.getRecord = targetjob.TargetJobRecord{
		ID:             imported.TargetJobID,
		UserID:         scenarioUserID,
		Status:         sharedtypes.TargetJobStatusDraft,
		AnalysisStatus: sharedtypes.TargetJobParseStatusReady,
		Title:          "Senior Frontend Engineer",
		CompanyName:    "Acme",
		TargetLanguage: "zh-CN",
		SourceType:     targetjob.SourceTypeManualText,
		Summary:        parseStore.applyResultIn.Summary,
		FitSummary:     parseStore.applyResultIn.FitSummary,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	store.getRequirements = parseStore.applyResultIn.Requirements
	detail, err := svc.GetTargetJob(context.Background(), scenarioUserID, imported.TargetJobID)
	if err != nil {
		t.Fatalf("get target job: %v", err)
	}
	if detail.AnalysisStatus != sharedtypes.TargetJobParseStatusReady || len(detail.Requirements) == 0 {
		t.Fatalf("detail does not expose ready parse result: %+v", detail)
	}
	if detail.Summary == nil || detail.Summary.Provenance.PromptVersion == "" || len(detail.Summary.CoreThemes) == 0 {
		t.Fatalf("detail missing parsed summary provenance: %+v", detail.Summary)
	}
	if detail.FitSummary == nil || detail.FitSummary.Provenance.ModelId == "" {
		t.Fatalf("detail missing fit summary provenance: %+v", detail.FitSummary)
	}

	store.listResult = targetjob.ListResult{Items: []targetjob.TargetJobRecord{store.getRecord}}
	list, err := svc.ListTargetJobs(context.Background(), targetjob.ListRequest{UserID: scenarioUserID, PageSize: 20})
	if err != nil {
		t.Fatalf("list target jobs: %v", err)
	}
	if len(list.Items) != 1 || list.Items[0].Id != imported.TargetJobID {
		t.Fatalf("list did not include imported job: %+v", list)
	}

	status := sharedtypes.TargetJobStatusPreparing
	notes := "Recruiter asked for frontend platform examples."
	store.updateResult = store.getRecord
	store.updateResult.Status = status
	updated, err := svc.UpdateTargetJob(context.Background(), targetjob.UpdateRequest{
		UserID:         scenarioUserID,
		TargetJobID:    imported.TargetJobID,
		IdempotencyKey: "e2e-p0-010-update",
		Status:         &status,
		Notes:          &notes,
	})
	if err != nil {
		t.Fatalf("update target job: %v", err)
	}
	if updated.Status != sharedtypes.TargetJobStatusPreparing || updated.AnalysisStatus != sharedtypes.TargetJobParseStatusReady {
		t.Fatalf("update changed wrong fields: %+v", updated)
	}
}

func TestE2EP0011URLImportFetchAndParse(t *testing.T) {
	svc, store := newServiceWithFake(
		"018f2a40-0000-7000-9000-0000000000a2",
		"018f2a40-0000-7000-9000-0000000000f2",
		"018f2a40-0000-7000-9000-0000000000c2",
		"018f2a40-0000-7000-9000-0000000000e2",
	)
	imported, err := svc.ImportTargetJob(context.Background(), targetjob.ImportRequest{
		UserID:         scenarioUserID,
		IdempotencyKey: "e2e-p0-011-url",
		TargetLanguage: "en",
		Source:         map[string]any{"type": "url", "url": "https://jobs.example.com/role/1?token=secret#frag"},
	})
	if err != nil {
		t.Fatalf("url import: %v", err)
	}
	if imported.Job.Status != sharedtypes.JobStatusQueued {
		t.Fatalf("url import must queue target_import job: %+v", imported.Job)
	}
	if strings.Contains(store.captured.SourceURL, "token=secret") || strings.Contains(store.captured.SourceURL, "#frag") {
		t.Fatalf("stored URL must strip query and fragment: %q", store.captured.SourceURL)
	}
	assertPayloadOmits(t, store.captured.OutboxEventPayload, "token=secret", "https://jobs.example.com/role/1?token=secret")

	exec, parseStore, _, ai, fetcher := newParseExecutorWithFakes(t)
	ai.resp = aiclient.CompleteResponse{Content: happyResponseJSON}
	fetchedAt := time.Date(2026, 5, 9, 23, 10, 0, 0, time.UTC)
	fetcher.res = urlfetch.FetchResult{
		SanitizedURL: "https://jobs.example.com/role/1",
		Body:         "Fetched public JD text for a frontend platform role.",
		FetchedAt:    fetchedAt,
	}
	parseStore.target = targetjob.TargetJobRecord{
		ID:             imported.TargetJobID,
		UserID:         scenarioUserID,
		SourceType:     targetjob.SourceTypeURL,
		SourceURL:      "https://jobs.example.com/role/1?token=secret",
		TargetLanguage: "en",
	}
	parseStore.sources = []targetjob.SourceRecord{{ID: "src-url-1", SourceType: targetjob.SourceTypeURL}}
	outcome := exec.Handle(context.Background(), targetjob.ClaimedJob{JobID: imported.Job.Id, JobType: "target_import", ResourceID: imported.TargetJobID})
	if !outcome.Succeeded {
		t.Fatalf("url parse outcome = %+v", outcome)
	}
	if parseStore.sourceSnapshotURL != "https://jobs.example.com/role/1" || parseStore.sourceSnapshotText == "" || parseStore.sourceSnapshotAt == nil {
		t.Fatalf("URL snapshot not persisted safely: url=%q text=%q at=%v", parseStore.sourceSnapshotURL, parseStore.sourceSnapshotText, parseStore.sourceSnapshotAt)
	}
	assertPayloadOmits(t, parseStore.parsedOutboxPayload, "token=secret", "Fetched public JD text")

	badSvc, _ := newServiceWithFake("bad-target", "bad-job")
	_, err = badSvc.ImportTargetJob(context.Background(), targetjob.ImportRequest{
		UserID:         scenarioUserID,
		IdempotencyKey: "e2e-p0-011-invalid",
		TargetLanguage: "en",
		Source:         map[string]any{"type": "url", "url": "http://169.254.169.254/latest/meta-data"},
	})
	var apiErr *targetjob.ServiceImportError
	if !errors.As(err, &apiErr) || apiErr.Code != sharederrors.CodeTargetImportSourceInvalid {
		t.Fatalf("invalid URL did not map to TARGET_IMPORT_SOURCE_INVALID: %v", err)
	}
}

func TestE2EP0012ParseFailureRetryableAndNonRetryable(t *testing.T) {
	cases := []struct {
		name      string
		configure func(*pipelineFakeStore, *fakeRegistry, *fakeAIClient)
		code      string
		retryable bool
	}{
		{
			name: "provider-timeout",
			configure: func(_ *pipelineFakeStore, _ *fakeRegistry, ai *fakeAIClient) {
				ai.err = errors.New("upstream failed with AI_PROVIDER_TIMEOUT")
			},
			code:      sharederrors.CodeAiProviderTimeout,
			retryable: true,
		},
		{
			name: "invalid-output",
			configure: func(_ *pipelineFakeStore, _ *fakeRegistry, ai *fakeAIClient) {
				ai.resp = aiclient.CompleteResponse{Content: "not-json"}
			},
			code:      sharederrors.CodeAiOutputInvalid,
			retryable: false,
		},
		{
			name: "registry-disabled",
			configure: func(_ *pipelineFakeStore, registry *fakeRegistry, _ *fakeAIClient) {
				registry.err = targetjob.ErrPromptUnsupported
			},
			code:      sharederrors.CodeAiProviderConfigInvalid,
			retryable: false,
		},
		{
			name: "secret-missing",
			configure: func(_ *pipelineFakeStore, _ *fakeRegistry, ai *fakeAIClient) {
				ai.err = errors.New("AI_PROVIDER_SECRET_MISSING")
			},
			code:      sharederrors.CodeAiProviderSecretMissing,
			retryable: false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			exec, store, registry, ai, _ := newParseExecutorWithFakes(t)
			store.target = targetjob.TargetJobRecord{
				ID:             "target-" + tc.name,
				UserID:         scenarioUserID,
				SourceType:     targetjob.SourceTypeManualText,
				TargetLanguage: "en",
				RawJDText:      "Private JD body that must not leak into failure evidence.",
			}
			tc.configure(store, registry, ai)
			outcome := exec.Handle(context.Background(), targetjob.ClaimedJob{JobID: "job-" + tc.name, JobType: "target_import", ResourceID: store.target.ID})
			if outcome.Succeeded || outcome.ErrorCode != tc.code || outcome.Retryable != tc.retryable {
				t.Fatalf("unexpected failure outcome: %+v", outcome)
			}
			if store.failedOutboxPayload == nil {
				t.Fatal("target.analysis.failed payload missing")
			}
			assertPayloadOmits(t, store.failedOutboxPayload, "Private JD body", "Authorization:", "provider secret")
		})
	}
}

func TestE2EP0013ManualFormReady(t *testing.T) {
	svc, store := newServiceWithFake(
		"018f2a40-0000-7000-9000-0000000000a3",
		"018f2a40-0000-7000-9000-0000000000f3",
		"018f2a40-0000-7000-9000-0000000000d3",
	)
	resp, err := svc.ImportTargetJob(context.Background(), targetjob.ImportRequest{
		UserID:         scenarioUserID,
		IdempotencyKey: "e2e-p0-013-manual-form",
		TargetLanguage: "zh-CN",
		Source: map[string]any{
			"type":           "manual_form",
			"title":          "Frontend Architect",
			"companyName":    "Acme",
			"rawDescription": "Lead frontend architecture across 12 squads. Must have React platform experience.",
		},
	})
	if err != nil {
		t.Fatalf("manual_form import: %v", err)
	}
	if resp.Job.Status != sharedtypes.JobStatusSucceeded || resp.Job.JobType != "target_import" {
		t.Fatalf("manual_form must return terminal target_import job: %+v", resp.Job)
	}
	if store.captured.OutboxEventID != "" || store.captured.JobPayload != nil || store.captured.SourceID != "" {
		t.Fatalf("manual_form must not create runner source/outbox/job payload: %+v", store.captured)
	}
	if store.captured.InitialAnalysisStatus != sharedtypes.TargetJobParseStatusReady || len(store.captured.DraftRequirements) == 0 {
		t.Fatalf("manual_form must be ready with draft requirements: %+v", store.captured)
	}

	now := time.Date(2026, 5, 9, 23, 20, 0, 0, time.UTC)
	store.getRecord = targetjob.TargetJobRecord{
		ID:             resp.TargetJobID,
		UserID:         scenarioUserID,
		Status:         sharedtypes.TargetJobStatusDraft,
		AnalysisStatus: sharedtypes.TargetJobParseStatusReady,
		Title:          "Frontend Architect",
		CompanyName:    "Acme",
		TargetLanguage: "zh-CN",
		SourceType:     targetjob.SourceTypeManualForm,
		RawJDText:      "Lead frontend architecture across 12 squads.",
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	store.getRequirements = store.captured.DraftRequirements
	detail, err := svc.GetTargetJob(context.Background(), scenarioUserID, resp.TargetJobID)
	if err != nil {
		t.Fatalf("get manual_form target job: %v", err)
	}
	if detail.AnalysisStatus != sharedtypes.TargetJobParseStatusReady || len(detail.Requirements) == 0 || detail.SourceType != "manual_form" {
		t.Fatalf("manual_form detail not ready: %+v", detail)
	}
	store.listResult = targetjob.ListResult{Items: []targetjob.TargetJobRecord{store.getRecord}}
	list, err := svc.ListTargetJobs(context.Background(), targetjob.ListRequest{UserID: scenarioUserID})
	if err != nil {
		t.Fatalf("list manual_form target job: %v", err)
	}
	if len(list.Items) != 1 || list.Items[0].Id != resp.TargetJobID {
		t.Fatalf("manual_form job missing from list: %+v", list)
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
