package jobs_test

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient"
	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient/observability"
	resumejobs "github.com/monshunter/easyinterview/backend/internal/resume/jobs"
	resumestore "github.com/monshunter/easyinterview/backend/internal/resume/store"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
	"github.com/monshunter/easyinterview/backend/internal/shared/events"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
	"github.com/monshunter/easyinterview/backend/internal/targetjob"
)

func TestParseHandlerUsesTwoSourceInputsAndWritesReadyOutbox(t *testing.T) {
	now := time.Date(2026, 5, 13, 7, 0, 0, 0, time.UTC)
	cases := []struct {
		name            string
		asset           resumestore.ParseAssetRecord
		objectText      string
		aiContent       string
		wantPrompt      string
		wantSnapshot    string
		forbidPrompt    string
		wantDisplayName string
		wantReadObject  bool
	}{
		{
			name: "paste original_text",
			asset: resumestore.ParseAssetRecord{
				ID:           "asset-paste",
				UserID:       "user-1",
				Language:     "en",
				ParseStatus:  sharedtypes.TargetJobParseStatusQueued,
				SourceType:   "paste",
				OriginalText: "paste resume text",
			},
			wantPrompt:   "paste resume text",
			wantSnapshot: "paste resume text",
		},
		{
			name: "upload file object",
			asset: resumestore.ParseAssetRecord{
				ID:            "asset-upload",
				UserID:        "user-1",
				Language:      "en",
				ParseStatus:   sharedtypes.TargetJobParseStatusQueued,
				SourceType:    "upload",
				FileObjectID:  "file-1",
				FileObjectKey: "user-1/resume/file.txt",
			},
			objectText:     "uploaded resume text",
			wantPrompt:     "uploaded resume text",
			wantSnapshot:   "uploaded resume text",
			wantReadObject: true,
		},
		{
			name: "filters generic llm name and uses root headline",
			asset: resumestore.ParseAssetRecord{
				ID:           "asset-generic-name",
				UserID:       "user-1",
				Language:     "zh-CN",
				ParseStatus:  sharedtypes.TargetJobParseStatusQueued,
				SourceType:   "paste",
				OriginalText: "张三\n后端平台工程师",
			},
			aiContent: `{
  "headline": "后端平台工程师",
  "basics": {"name": "粘贴的简历"},
  "experiences": [],
  "projects": [{"name": "Ferry"}],
  "education": [],
  "skills": ["Go"],
  "languages": ["zh-CN"]
}`,
			wantPrompt:      "张三\n后端平台工程师",
			wantSnapshot:    "张三\n后端平台工程师",
			wantDisplayName: "后端平台工程师",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			store := &fakeParseStore{asset: tc.asset}
			objects := &fakeObjectReader{objects: map[string]string{tc.asset.FileObjectKey: tc.objectText}}
			aiContent := tc.aiContent
			if aiContent == "" {
				aiContent = validResumeParseJSON
			}
			ai := &captureAI{resp: aiclient.CompleteResponse{Content: aiContent}}
			handler := resumejobs.NewParseHandler(resumejobs.ParseHandlerOptions{
				Store:    store,
				Registry: fakeRegistry{resolution: parseResolution()},
				AI:       ai,
				Objects:  objects,
				NewID:    idSeq("event-1"),
				Now:      func() time.Time { return now },
			})

			outcome := handler.Handle(context.Background(), targetjob.ClaimedJob{
				JobID: "job-1", JobType: "resume_parse", ResourceType: "resume_asset", ResourceID: tc.asset.ID, Attempts: 1, MaxAttempts: 5,
			})

			if !outcome.Succeeded {
				t.Fatalf("Handle outcome = %+v", outcome)
			}
			if len(store.markParsing) != 1 || store.markParsing[0].AssetID != tc.asset.ID {
				t.Fatalf("MarkParsing calls = %+v", store.markParsing)
			}
			if store.success == nil {
				t.Fatal("expected CompleteParseSuccess")
			}
			if !json.Valid(store.success.ParsedSummary) || !strings.Contains(string(store.success.ParsedSummary), `"skills"`) {
				t.Fatalf("parsed summary = %s", store.success.ParsedSummary)
			}
			if !json.Valid(store.success.StructuredProfile) || !strings.Contains(string(store.success.StructuredProfile), `"skills"`) {
				t.Fatalf("structured profile = %s", store.success.StructuredProfile)
			}
			wantDisplayName := tc.wantDisplayName
			if wantDisplayName == "" {
				wantDisplayName = "Ada Lovelace - Engineer"
			}
			if store.success.DisplayName == nil || *store.success.DisplayName != wantDisplayName {
				t.Fatalf("display name = %#v, want %s", store.success.DisplayName, wantDisplayName)
			}
			if store.success.ParsedTextSnapshot != tc.wantSnapshot {
				t.Fatalf("parsed text snapshot = %q, want %q", store.success.ParsedTextSnapshot, tc.wantSnapshot)
			}
			var payload events.ResumeParseCompletedPayload
			if err := json.Unmarshal(store.success.OutboxEventPayload, &payload); err != nil {
				t.Fatalf("decode outbox payload: %v", err)
			}
			if payload.ResumeID != tc.asset.ID || payload.UserID != tc.asset.UserID || payload.ParseStatus != sharedtypes.TargetJobParseStatusReady {
				t.Fatalf("outbox payload = %+v", payload)
			}
			if bytes := string(store.success.OutboxEventPayload); strings.Contains(bytes, tc.wantPrompt) || strings.Contains(bytes, "Ada Lovelace") {
				t.Fatalf("outbox payload leaked resume content: %s", bytes)
			}
			prompt := ai.lastUserMessage()
			if !strings.Contains(prompt, tc.wantPrompt) {
				t.Fatalf("prompt %q does not contain %q", prompt, tc.wantPrompt)
			}
			if tc.forbidPrompt != "" && strings.Contains(prompt, tc.forbidPrompt) {
				t.Fatalf("prompt leaked forbidden input %q: %s", tc.forbidPrompt, prompt)
			}
			if (objects.readCalls == 1) != tc.wantReadObject {
				t.Fatalf("object read calls = %d, want read=%v", objects.readCalls, tc.wantReadObject)
			}
			if ai.profileName != "resume.parse.default" ||
				ai.payload.Metadata.FeatureKey != resumejobs.FeatureKeyResumeParse ||
				ai.payload.Metadata.TaskRun.Capability != aiclient.AITaskRunTaskResumeParse ||
				ai.payload.Metadata.TaskRun.ResourceType != aiclient.AITaskRunResourceResumeAsset ||
				ai.payload.Metadata.TaskRun.ResourceID != tc.asset.ID ||
				ai.payload.Metadata.TaskRun.UserID != tc.asset.UserID {
				t.Fatalf("AI metadata drift: profile=%q metadata=%+v", ai.profileName, ai.payload.Metadata)
			}
			if len(ai.payload.Metadata.OutputSchema) == 0 {
				t.Fatalf("AI metadata OutputSchema must be populated")
			}
		})
	}
}

func TestParseHandlerExtractsReadableUploadText(t *testing.T) {
	now := time.Date(2026, 7, 7, 9, 0, 0, 0, time.UTC)
	cases := []struct {
		name       string
		objectKey  string
		body       []byte
		wantText   string
		forbidText []string
	}{
		{
			name:       "pdf",
			objectKey:  "user-1/resume/tan-ai-backend.pdf",
			body:       minimalPDFWithText("PDF Extracted Resume Text"),
			wantText:   "PDF Extracted Resume Text",
			forbidText: []string{"%PDF", "tan-ai-backend.pdf"},
		},
		{
			name:       "docx",
			objectKey:  "user-1/resume/tan-ai-backend.docx",
			body:       minimalDOCXWithText(t, "DOCX Extracted Resume Text"),
			wantText:   "DOCX Extracted Resume Text",
			forbidText: []string{"PK", "word/document.xml"},
		},
		{
			name:      "markdown",
			objectKey: "user-1/resume/tan-ai-backend.md",
			body:      []byte("# Markdown Resume\nGo platform engineer"),
			wantText:  "# Markdown Resume Go platform engineer",
		},
		{
			name:      "text",
			objectKey: "user-1/resume/tan-ai-backend.txt",
			body:      []byte("Plain Text Resume\nAI infrastructure engineer"),
			wantText:  "Plain Text Resume AI infrastructure engineer",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			store := &fakeParseStore{asset: resumestore.ParseAssetRecord{
				ID:            "asset-upload-" + tc.name,
				UserID:        "user-1",
				Language:      "zh-CN",
				ParseStatus:   sharedtypes.TargetJobParseStatusQueued,
				SourceType:    "upload",
				FileObjectID:  "file-1",
				FileObjectKey: tc.objectKey,
			}}
			objects := &fakeObjectReader{objectBytes: map[string][]byte{tc.objectKey: tc.body}}
			ai := &captureAI{resp: aiclient.CompleteResponse{Content: validResumeParseJSON}}
			handler := resumejobs.NewParseHandler(resumejobs.ParseHandlerOptions{
				Store:    store,
				Registry: fakeRegistry{resolution: parseResolution()},
				AI:       ai,
				Objects:  objects,
				NewID:    idSeq("event-1"),
				Now:      func() time.Time { return now },
			})

			outcome := handler.Handle(context.Background(), targetjob.ClaimedJob{
				JobID: "job-1", JobType: "resume_parse", ResourceType: "resume_asset", ResourceID: store.asset.ID, Attempts: 1, MaxAttempts: 5,
			})

			if !outcome.Succeeded {
				t.Fatalf("Handle outcome = %+v", outcome)
			}
			if store.success == nil {
				t.Fatal("expected CompleteParseSuccess")
			}
			if got := normalizeComparableText(store.success.ParsedTextSnapshot); got != tc.wantText {
				t.Fatalf("parsed text snapshot = %q, want %q", got, tc.wantText)
			}
			prompt := normalizeComparableText(ai.lastUserMessage())
			if !strings.Contains(prompt, tc.wantText) {
				t.Fatalf("prompt %q does not contain readable text %q", prompt, tc.wantText)
			}
			for _, forbidden := range tc.forbidText {
				if strings.Contains(store.success.ParsedTextSnapshot, forbidden) || strings.Contains(ai.lastUserMessage(), forbidden) {
					t.Fatalf("upload extraction leaked forbidden token %q into snapshot/prompt", forbidden)
				}
			}
		})
	}
}

func TestParseHandlerFailurePathsMarkFailedAndSkipCompletedOutbox(t *testing.T) {
	now := time.Date(2026, 5, 13, 7, 30, 0, 0, time.UTC)
	cases := []struct {
		name       string
		ai         *captureAI
		job        targetjob.ClaimedJob
		wantCode   string
		wantRetry  bool
		wantFailed bool
	}{
		{
			name:       "invalid json output",
			ai:         &captureAI{resp: aiclient.CompleteResponse{Content: "not-json"}},
			job:        targetjob.ClaimedJob{JobID: "job-1", JobType: "resume_parse", ResourceType: "resume_asset", ResourceID: "asset-1", Attempts: 1, MaxAttempts: 5},
			wantCode:   sharederrors.CodeAiOutputInvalid,
			wantFailed: true,
		},
		{
			name:       "retryable timeout before exhaustion",
			ai:         &captureAI{err: errors.New(sharederrors.CodeAiProviderTimeout + " provider slow")},
			job:        targetjob.ClaimedJob{JobID: "job-1", JobType: "resume_parse", ResourceType: "resume_asset", ResourceID: "asset-1", Attempts: 1, MaxAttempts: 5},
			wantCode:   sharederrors.CodeAiProviderTimeout,
			wantRetry:  true,
			wantFailed: true,
		},
		{
			name:       "retryable timeout exhausted",
			ai:         &captureAI{err: errors.New(sharederrors.CodeAiProviderTimeout + " provider slow")},
			job:        targetjob.ClaimedJob{JobID: "job-1", JobType: "resume_parse", ResourceType: "resume_asset", ResourceID: "asset-1", Attempts: 5, MaxAttempts: 5},
			wantCode:   sharederrors.CodeAiProviderTimeout,
			wantRetry:  true,
			wantFailed: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			store := &fakeParseStore{asset: resumestore.ParseAssetRecord{
				ID:           "asset-1",
				UserID:       "user-1",
				Language:     "en",
				ParseStatus:  sharedtypes.TargetJobParseStatusQueued,
				SourceType:   "paste",
				OriginalText: "private resume body",
			}}
			handler := resumejobs.NewParseHandler(resumejobs.ParseHandlerOptions{
				Store:    store,
				Registry: fakeRegistry{resolution: parseResolution()},
				AI:       tc.ai,
				NewID:    idSeq("event-1"),
				Now:      func() time.Time { return now },
			})

			outcome := handler.Handle(context.Background(), tc.job)

			if outcome.Succeeded || outcome.ErrorCode != tc.wantCode || outcome.Retryable != tc.wantRetry {
				t.Fatalf("outcome = %+v, want code=%s retry=%v", outcome, tc.wantCode, tc.wantRetry)
			}
			if tc.wantFailed {
				if store.failure == nil || store.failure.ErrorCode != tc.wantCode {
					t.Fatalf("failure = %+v, want code %s", store.failure, tc.wantCode)
				}
			} else if store.failure != nil {
				t.Fatalf("retryable non-exhausted failure should not mark resume failed: %+v", store.failure)
			}
			if store.success != nil {
				t.Fatalf("failure must not write completed outbox: %+v", store.success)
			}
			if strings.Contains(outcome.ErrorMessage, "private resume body") {
				t.Fatalf("outcome leaked resume body: %+v", outcome)
			}
		})
	}
}

func TestParseHandlerRetriesFailedAssetBackToProcessing(t *testing.T) {
	now := time.Date(2026, 5, 13, 7, 45, 0, 0, time.UTC)
	store := &fakeParseStore{asset: resumestore.ParseAssetRecord{
		ID:           "asset-1",
		UserID:       "user-1",
		Language:     "en",
		ParseStatus:  sharedtypes.TargetJobParseStatusFailed,
		SourceType:   "paste",
		OriginalText: "resume retry body",
	}}
	handler := resumejobs.NewParseHandler(resumejobs.ParseHandlerOptions{
		Store:    store,
		Registry: fakeRegistry{resolution: parseResolution()},
		AI:       &captureAI{resp: aiclient.CompleteResponse{Content: validResumeParseJSON}},
		NewID:    idSeq("event-1"),
		Now:      func() time.Time { return now },
	})

	outcome := handler.Handle(context.Background(), targetjob.ClaimedJob{
		JobID: "job-1", JobType: "resume_parse", ResourceType: "resume_asset", ResourceID: "asset-1", Attempts: 2, MaxAttempts: 5,
	})

	if !outcome.Succeeded {
		t.Fatalf("Handle outcome = %+v", outcome)
	}
	if len(store.markParsing) != 1 || store.markParsing[0].AssetID != "asset-1" {
		t.Fatalf("MarkParsing calls = %+v", store.markParsing)
	}
	if store.success == nil {
		t.Fatal("expected CompleteParseSuccess")
	}
	if store.success.DisplayName == nil || *store.success.DisplayName != "Ada Lovelace - Engineer" {
		t.Fatalf("display name = %#v, want Ada Lovelace - Engineer", store.success.DisplayName)
	}
	if strings.Contains(string(store.success.OutboxEventPayload), "resume retry body") {
		t.Fatalf("outbox payload leaked resume content: %s", store.success.OutboxEventPayload)
	}
}

func TestParseHandlerObservedAIWritesResumeTaskRunColumns(t *testing.T) {
	now := time.Date(2026, 5, 13, 8, 0, 0, 0, time.UTC)
	assetID := "01918fa0-0000-7000-8000-000000000801"
	userID := "01918fa0-0000-7000-8000-000000000802"
	store := &fakeParseStore{asset: resumestore.ParseAssetRecord{
		ID:           assetID,
		UserID:       userID,
		Language:     "en",
		ParseStatus:  sharedtypes.TargetJobParseStatusQueued,
		SourceType:   "paste",
		OriginalText: "resume text",
	}}
	baseAI := &captureAI{resp: aiclient.CompleteResponse{Content: validResumeParseJSON}}
	runWriter := &memTaskRunWriter{}
	ai, err := observability.New(baseAI,
		observability.WithRegisterer(observability.NewInMemoryRegistry()),
		observability.WithLogger(observability.NewMemoryLogger()),
		observability.WithAITaskRunWriter(runWriter),
		observability.WithAuditEventWriter(discardAuditWriter{}),
		observability.WithProfileResolver(staticResolver{
			"resume.parse.default": {
				Name:       "resume.parse.default",
				Capability: aiclient.CapabilityChat,
				Status:     aiclient.ProfileStatusActive,
				Default: aiclient.ProviderConfig{
					ProviderRef: "stub",
					Model:       "fixture-model:resume-parse",
				},
				Route:     "resume.parse",
				TimeoutMs: 5000,
				Version:   "1.0.0",
			},
		}),
		observability.WithNow(func() time.Time { return now }),
	)
	if err != nil {
		t.Fatalf("observability.New: %v", err)
	}
	handler := resumejobs.NewParseHandler(resumejobs.ParseHandlerOptions{
		Store:    store,
		Registry: fakeRegistry{resolution: parseResolution()},
		AI:       ai,
		NewID:    idSeq("event-1"),
		Now:      func() time.Time { return now },
	})

	outcome := handler.Handle(context.Background(), targetjob.ClaimedJob{
		JobID: "job-1", JobType: "resume_parse", ResourceType: "resume_asset", ResourceID: assetID, Attempts: 1, MaxAttempts: 5,
	})

	if !outcome.Succeeded {
		t.Fatalf("Handle outcome = %+v", outcome)
	}
	rows := runWriter.Rows()
	if len(rows) != 1 {
		t.Fatalf("expected one ai_task_runs row, got %+v", rows)
	}
	row := rows[0]
	if row.FeatureKey != resumejobs.FeatureKeyResumeParse ||
		row.PromptVersion != "v0.1.0" ||
		row.RubricVersion != "v0.1.0" ||
		row.ModelProfileName != "resume.parse.default" ||
		row.ModelProfileVersion != "1.0.0" ||
		row.Route != "resume.parse" ||
		row.ValidationStatus != aiclient.ValidationStatusOK ||
		row.Capability != aiclient.AITaskRunTaskResumeParse ||
		row.ResourceType != aiclient.AITaskRunResourceResumeAsset ||
		row.ResourceID != assetID ||
		row.UserID != userID {
		t.Fatalf("ai_task_runs row drift: %+v", row)
	}
}

func TestParseHandlerPIIRedlineForLogsAuditTaskRunsAndOutbox(t *testing.T) {
	now := time.Date(2026, 5, 13, 8, 15, 0, 0, time.UTC)
	secretResume := "SECRET_RESUME_BODY"
	secretResponse := "SECRET_MODEL_RESPONSE"
	assetID := "01918fa0-0000-7000-8000-000000000811"
	userID := "01918fa0-0000-7000-8000-000000000812"
	store := &fakeParseStore{asset: resumestore.ParseAssetRecord{
		ID:           assetID,
		UserID:       userID,
		Language:     "en",
		ParseStatus:  sharedtypes.TargetJobParseStatusQueued,
		SourceType:   "paste",
		OriginalText: secretResume,
	}}
	baseAI := &captureAI{resp: aiclient.CompleteResponse{Content: strings.Replace(validResumeParseJSON, "Ada Lovelace", secretResponse, 1)}}
	logger := observability.NewMemoryLogger()
	runWriter := &memTaskRunWriter{}
	auditWriter := &memAuditWriter{}
	metrics := observability.NewInMemoryRegistry()
	ai, err := observability.New(baseAI,
		observability.WithRegisterer(metrics),
		observability.WithLogger(logger),
		observability.WithAITaskRunWriter(runWriter),
		observability.WithAuditEventWriter(auditWriter),
		observability.WithProfileResolver(staticResolver{
			"resume.parse.default": {
				Name:       "resume.parse.default",
				Capability: aiclient.CapabilityChat,
				Status:     aiclient.ProfileStatusActive,
				Default: aiclient.ProviderConfig{
					ProviderRef: "stub",
					Model:       "fixture-model:resume-parse",
				},
				Route:     "resume.parse",
				TimeoutMs: 5000,
				Version:   "1.0.0",
			},
		}),
		observability.WithNow(func() time.Time { return now }),
	)
	if err != nil {
		t.Fatalf("observability.New: %v", err)
	}
	handler := resumejobs.NewParseHandler(resumejobs.ParseHandlerOptions{
		Store:    store,
		Registry: fakeRegistry{resolution: parseResolution()},
		AI:       ai,
		NewID:    idSeq("event-1"),
		Now:      func() time.Time { return now },
	})

	outcome := handler.Handle(context.Background(), targetjob.ClaimedJob{
		JobID: "job-1", JobType: "resume_parse", ResourceType: "resume_asset", ResourceID: assetID, Attempts: 1, MaxAttempts: 5,
	})

	if !outcome.Succeeded {
		t.Fatalf("Handle outcome = %+v", outcome)
	}
	snapshot, err := json.Marshal(map[string]any{
		"logs":          logger.Entries(),
		"ai_task_runs":  runWriter.Rows(),
		"audit":         auditWriter.Rows(),
		"metric_labels": metrics.CounterLabelValues(observability.MetricRunsTotal),
		"outbox":        string(store.success.OutboxEventPayload),
	})
	if err != nil {
		t.Fatalf("marshal privacy snapshot: %v", err)
	}
	for _, token := range []string{secretResume, secretResponse} {
		if strings.Contains(string(snapshot), token) {
			t.Fatalf("PII token %q leaked into non-store surfaces: %s", token, snapshot)
		}
	}
	if !strings.Contains(string(store.success.ParsedSummary), secretResponse) {
		t.Fatalf("parsed summary should remain the store-owned parsed payload")
	}
}

const validResumeParseJSON = `{
  "basics": {"name": "Ada Lovelace"},
  "experiences": [{"company": "Analytical Engines", "title": "Engineer", "start": "2024-01", "end": "", "summary": "Built systems", "bullets": ["Led platform work"]}],
  "projects": [],
  "education": [],
  "skills": ["Go", "PostgreSQL"],
  "languages": ["en"]
}`

func parseResolution() resumejobs.PromptResolution {
	return resumejobs.PromptResolution{
		PromptVersion:       "v0.1.0",
		RubricVersion:       "v0.1.0",
		ModelProfileName:    "resume.parse.default",
		DataSourceVersion:   "registry.v1",
		FeatureFlag:         "none",
		OutputSchema:        rawSchema(`{"type":"object","required":["basics"],"properties":{"basics":{"type":"object"}}}`),
		UserMessageTemplate: "Parse this resume:\n{{resume_text}}",
	}
}

func rawSchema(s string) *json.RawMessage {
	raw := json.RawMessage(s)
	return &raw
}

type fakeParseStore struct {
	asset       resumestore.ParseAssetRecord
	markParsing []resumestore.StatusUpdateInput
	success     *resumestore.CompleteParseSuccessInput
	failure     *resumestore.CompleteParseFailureInput
}

func (s *fakeParseStore) GetForParse(_ context.Context, assetID string) (resumestore.ParseAssetRecord, error) {
	if s.asset.ID != assetID {
		return resumestore.ParseAssetRecord{}, resumestore.ErrAssetNotFound
	}
	return s.asset, nil
}

func (s *fakeParseStore) MarkParsing(_ context.Context, in resumestore.StatusUpdateInput) error {
	s.markParsing = append(s.markParsing, in)
	s.asset.ParseStatus = sharedtypes.TargetJobParseStatusProcessing
	return nil
}

func (s *fakeParseStore) CompleteParseSuccess(_ context.Context, in resumestore.CompleteParseSuccessInput) error {
	cp := in
	s.success = &cp
	s.asset.ParseStatus = sharedtypes.TargetJobParseStatusReady
	return nil
}

func (s *fakeParseStore) CompleteParseFailure(_ context.Context, in resumestore.CompleteParseFailureInput) error {
	cp := in
	s.failure = &cp
	s.asset.ParseStatus = sharedtypes.TargetJobParseStatusFailed
	return nil
}

type fakeRegistry struct {
	resolution resumejobs.PromptResolution
	err        error
}

func (r fakeRegistry) Resolve(_ context.Context, featureKey string, language string) (resumejobs.PromptResolution, error) {
	if r.err != nil {
		return resumejobs.PromptResolution{}, r.err
	}
	if featureKey != resumejobs.FeatureKeyResumeParse || strings.TrimSpace(language) == "" {
		return resumejobs.PromptResolution{}, resumejobs.ErrPromptUnsupported
	}
	return r.resolution, nil
}

type fakeObjectReader struct {
	objects     map[string]string
	objectBytes map[string][]byte
	readCalls   int
}

func (r *fakeObjectReader) Read(_ context.Context, objectKey string, _ int64) ([]byte, error) {
	r.readCalls++
	if value, ok := r.objectBytes[objectKey]; ok {
		return append([]byte(nil), value...), nil
	}
	value, ok := r.objects[objectKey]
	if !ok {
		return nil, errors.New("object not found")
	}
	return []byte(value), nil
}

func minimalPDFWithText(text string) []byte {
	stream := fmt.Sprintf("BT /F1 16 Tf 72 720 Td (%s) Tj ET", escapePDFText(text))
	objects := []string{
		"1 0 obj\n<< /Type /Catalog /Pages 2 0 R >>\nendobj\n",
		"2 0 obj\n<< /Type /Pages /Kids [3 0 R] /Count 1 >>\nendobj\n",
		"3 0 obj\n<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Resources << /Font << /F1 4 0 R >> >> /Contents 5 0 R >>\nendobj\n",
		"4 0 obj\n<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>\nendobj\n",
		fmt.Sprintf("5 0 obj\n<< /Length %d >>\nstream\n%s\nendstream\nendobj\n", len(stream), stream),
	}
	var buf bytes.Buffer
	buf.WriteString("%PDF-1.4\n")
	offsets := make([]int, 0, len(objects))
	for _, obj := range objects {
		offsets = append(offsets, buf.Len())
		buf.WriteString(obj)
	}
	xref := buf.Len()
	fmt.Fprintf(&buf, "xref\n0 %d\n0000000000 65535 f \n", len(objects)+1)
	for _, offset := range offsets {
		fmt.Fprintf(&buf, "%010d 00000 n \n", offset)
	}
	fmt.Fprintf(&buf, "trailer\n<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n", len(objects)+1, xref)
	return buf.Bytes()
}

func minimalDOCXWithText(t *testing.T, text string) []byte {
	t.Helper()
	var buf bytes.Buffer
	writer := zip.NewWriter(&buf)
	file, err := writer.Create("word/document.xml")
	if err != nil {
		t.Fatalf("create docx document.xml: %v", err)
	}
	if _, err := file.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?><w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:body><w:p><w:r><w:t>` + text + `</w:t></w:r></w:p></w:body></w:document>`)); err != nil {
		t.Fatalf("write docx document.xml: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close docx zip: %v", err)
	}
	return buf.Bytes()
}

func escapePDFText(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, "(", `\(`)
	value = strings.ReplaceAll(value, ")", `\)`)
	return value
}

func normalizeComparableText(value string) string {
	return strings.Join(strings.Fields(value), " ")
}

type captureAI struct {
	profileName string
	payload     aiclient.CompletePayload
	resp        aiclient.CompleteResponse
	meta        aiclient.AICallMeta
	err         error
}

func (c *captureAI) Complete(_ context.Context, profileName string, payload aiclient.CompletePayload) (aiclient.CompleteResponse, aiclient.AICallMeta, error) {
	c.profileName = profileName
	c.payload = payload
	if c.meta.ModelID == "" {
		c.meta = aiclient.AICallMeta{
			Provider:         "stub",
			ModelFamily:      "stub",
			ModelID:          "fixture-model:resume-parse",
			FallbackChain:    []string{"stub/fixture-model:resume-parse"},
			ValidationStatus: aiclient.ValidationStatusOK,
		}
	}
	return c.resp, c.meta, c.err
}

func (c *captureAI) lastUserMessage() string {
	for i := len(c.payload.Messages) - 1; i >= 0; i-- {
		if c.payload.Messages[i].Role == "user" {
			return c.payload.Messages[i].Content
		}
	}
	return ""
}

func (c *captureAI) Transcribe(context.Context, string, aiclient.TranscriptionInput) (aiclient.TranscriptionResponse, aiclient.AICallMeta, error) {
	return aiclient.TranscriptionResponse{}, aiclient.AICallMeta{}, errors.New("not implemented")
}

func (c *captureAI) Stream(context.Context, string, aiclient.CompletePayload) (<-chan aiclient.AIStreamEvent, error) {
	return nil, errors.New("not implemented")
}

func (c *captureAI) Synthesize(context.Context, string, aiclient.SynthesisInput) (aiclient.SynthesisResponse, aiclient.AICallMeta, error) {
	return aiclient.SynthesisResponse{}, aiclient.AICallMeta{}, errors.New("not implemented")
}

type memTaskRunWriter struct {
	rows []aiclient.AITaskRunRow
}

func (w *memTaskRunWriter) WriteAITaskRun(_ context.Context, row aiclient.AITaskRunRow) error {
	w.rows = append(w.rows, row)
	return nil
}

func (w *memTaskRunWriter) Rows() []aiclient.AITaskRunRow {
	return append([]aiclient.AITaskRunRow{}, w.rows...)
}

type discardAuditWriter struct{}

func (discardAuditWriter) WriteAuditEvent(context.Context, aiclient.AuditEventRow) error {
	return nil
}

type memAuditWriter struct {
	rows []aiclient.AuditEventRow
}

func (w *memAuditWriter) WriteAuditEvent(_ context.Context, row aiclient.AuditEventRow) error {
	w.rows = append(w.rows, row)
	return nil
}

func (w *memAuditWriter) Rows() []aiclient.AuditEventRow {
	return append([]aiclient.AuditEventRow{}, w.rows...)
}

type staticResolver map[string]*aiclient.ModelProfile

func (r staticResolver) Resolve(name string) (*aiclient.ModelProfile, error) {
	profile, ok := r[name]
	if !ok {
		return nil, errors.New("not found: " + name)
	}
	return profile, nil
}

func idSeq(ids ...string) func() string {
	i := 0
	return func() string {
		if i >= len(ids) {
			return ids[len(ids)-1]
		}
		id := ids[i]
		i++
		return id
	}
}
