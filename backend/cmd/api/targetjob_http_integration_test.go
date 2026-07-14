package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient"
	api "github.com/monshunter/easyinterview/backend/internal/api/generated"
	"github.com/monshunter/easyinterview/backend/internal/auth"
	"github.com/monshunter/easyinterview/backend/internal/platform/config"
	"github.com/monshunter/easyinterview/backend/internal/runner"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
	"github.com/monshunter/easyinterview/backend/internal/targetjob"
)

const (
	targetJobHTTPIntegrationUserID   = "integration-user-targetjob-http"
	targetJobHTTPIntegrationResumeID = "integration-resume-targetjob-http"
)

func TestTargetJobHTTPTextImportParseReady(t *testing.T) {
	h := newTargetJobHTTPIntegrationHarness(t, targetJobHTTPIntegrationOptions{})

	const rawText = "Private integration JD text that must stay out of evidence logs."
	importBody := api.ImportTargetJobRequest{
		RawText:        rawText,
		TargetLanguage: "zh-CN",
		ResumeId:       targetJobHTTPIntegrationResumeID,
	}
	raw := h.doJSON(t, http.MethodPost, "/api/v1/targets/import", "targetjob-integration-import", importBody, http.StatusAccepted)
	var imported api.TargetJobWithJob
	decodeJSON(t, raw, &imported)
	if imported.TargetJobId == "" || imported.Job.Status != sharedtypes.JobStatusQueued {
		t.Fatalf("unexpected import response: %+v", imported)
	}

	duplicateRaw := h.doJSON(t, http.MethodPost, "/api/v1/targets/import", "targetjob-integration-import", importBody, http.StatusAccepted)
	var duplicate api.TargetJobWithJob
	decodeJSON(t, duplicateRaw, &duplicate)
	if duplicate.TargetJobId != imported.TargetJobId || h.store.targetCount() != 1 {
		t.Fatalf("idempotent import did not return existing target: duplicate=%+v targets=%d", duplicate, h.store.targetCount())
	}
	if got := h.store.targets[imported.TargetJobId].RawJDText; got != rawText {
		t.Fatalf("paste-only import did not persist exact rawText: got=%q want=%q", got, rawText)
	}

	h.runRunnerOnce(t, true)

	listRaw := h.doJSON(t, http.MethodGet, "/api/v1/targets?pageSize=20", "", nil, http.StatusOK)
	var list api.PaginatedTargetJob
	decodeJSON(t, listRaw, &list)
	if len(list.Items) != 1 || list.Items[0].Id != imported.TargetJobId || list.PageInfo.PageSize != 20 {
		t.Fatalf("list did not expose imported job with pageInfo: %+v", list)
	}

	detailRaw := h.doJSON(t, http.MethodGet, "/api/v1/targets/"+imported.TargetJobId, "", nil, http.StatusOK)
	var detail api.TargetJob
	decodeJSON(t, detailRaw, &detail)
	if detail.AnalysisStatus != sharedtypes.TargetJobParseStatusReady || len(detail.Requirements) == 0 {
		t.Fatalf("detail did not expose ready parse result: %+v", detail)
	}
	if detail.Summary == nil || detail.Summary.Provenance.PromptVersion == "" || detail.FitSummary == nil {
		t.Fatalf("detail missing summary provenance: summary=%+v fit=%+v", detail.Summary, detail.FitSummary)
	}

	status := sharedtypes.TargetJobStatusPreparing
	updateRaw := h.doJSON(t, http.MethodPatch, "/api/v1/targets/"+imported.TargetJobId, "targetjob-integration-update", api.UpdateTargetJobRequest{
		Status: &status,
		Notes:  strPtr("Recruiter asked for platform examples."),
	}, http.StatusOK)
	var updated api.TargetJob
	decodeJSON(t, updateRaw, &updated)
	if updated.Status != sharedtypes.TargetJobStatusPreparing || updated.AnalysisStatus != sharedtypes.TargetJobParseStatusReady {
		t.Fatalf("update changed wrong fields: %+v", updated)
	}

	assertNoEvidenceLeak(t, h.store.outboxPayloads(), rawText, "prompt body", "response body", "Authorization:")
}

func TestTargetJobHTTPParseFailureRetryableAndNonRetryable(t *testing.T) {
	cases := []struct {
		name      string
		options   targetJobHTTPIntegrationOptions
		wantCode  string
		retryable bool
	}{
		{
			name:      "provider-timeout",
			options:   targetJobHTTPIntegrationOptions{ai: &integrationAIClient{err: errors.New("provider error: AI_PROVIDER_TIMEOUT")}},
			wantCode:  sharederrors.CodeAiProviderTimeout,
			retryable: true,
		},
		{
			name:      "invalid-output",
			options:   targetJobHTTPIntegrationOptions{ai: &integrationAIClient{resp: aiclient.CompleteResponse{Content: "not-json"}}},
			wantCode:  sharederrors.CodeAiOutputInvalid,
			retryable: false,
		},
		{
			name:      "registry-disabled",
			options:   targetJobHTTPIntegrationOptions{registry: integrationRegistry{err: targetjob.ErrPromptUnsupported}},
			wantCode:  sharederrors.CodeAiProviderConfigInvalid,
			retryable: false,
		},
		{
			name:      "secret-missing",
			options:   targetJobHTTPIntegrationOptions{ai: &integrationAIClient{err: errors.New("AI_PROVIDER_SECRET_MISSING provider secret missing")}},
			wantCode:  sharederrors.CodeAiProviderSecretMissing,
			retryable: false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := newTargetJobHTTPIntegrationHarness(t, tc.options)
			raw := h.doJSON(t, http.MethodPost, "/api/v1/targets/import", "targetjob-integration-failure-"+tc.name, api.ImportTargetJobRequest{
				RawText:        "Private JD body that must not leak.",
				TargetLanguage: "en",
				ResumeId:       targetJobHTTPIntegrationResumeID,
			}, http.StatusAccepted)
			var imported api.TargetJobWithJob
			decodeJSON(t, raw, &imported)

			h.runRunnerOnce(t, true)
			outcome := h.store.finalizedOutcome(imported.Job.Id)
			if outcome.Succeeded || outcome.ErrorCode != tc.wantCode || outcome.Retryable != tc.retryable {
				t.Fatalf("unexpected finalize outcome: %+v", outcome)
			}
			h.doJSON(t, http.MethodGet, "/api/v1/targets/"+imported.TargetJobId, "", nil, http.StatusNotFound)
			listRaw := h.doJSON(t, http.MethodGet, "/api/v1/targets?pageSize=20", "", nil, http.StatusOK)
			var list api.PaginatedTargetJob
			decodeJSON(t, listRaw, &list)
			for _, item := range list.Items {
				if item.Id == imported.TargetJobId {
					t.Fatalf("failed parse target must not be visible in list: %+v", list)
				}
			}
			payload := h.store.lastFailedPayload()
			var failed struct {
				ErrorCode string `json:"errorCode"`
				Retryable bool   `json:"retryable"`
			}
			decodeJSON(t, payload, &failed)
			if failed.ErrorCode != tc.wantCode || failed.Retryable != tc.retryable {
				t.Fatalf("target.analysis.failed payload mismatch: %+v", failed)
			}
			assertNoEvidenceLeak(t, h.store.outboxPayloads(), "Private JD body", "provider secret", "prompt body", "response body", "Authorization:")
		})
	}
}

type targetJobHTTPIntegrationOptions struct {
	ai       aiclient.AIClient
	registry targetjob.PromptRegistryClient
}

type targetJobHTTPIntegrationHarness struct {
	handler http.Handler
	runtime *runner.Runtime
	store   *integrationTargetJobStore
	cookie  *http.Cookie
}

func newTargetJobHTTPIntegrationHarness(t *testing.T, opts targetJobHTTPIntegrationOptions) *targetJobHTTPIntegrationHarness {
	t.Helper()
	dir := t.TempDir()
	writeAPIFile(t, filepath.Join(dir, "config.yaml"), `
runtime:
  appVersion: "1.2.3"
  defaultUiLanguage: zh-CN
`)
	loader, err := config.Load(config.Options{AppEnv: "test", ConfigDir: dir})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	store := newIntegrationTargetJobStore()
	authStore := &apiAuthStore{
		session: auth.SessionRecord{
			ID:        "integration-session-1",
			UserID:    targetJobHTTPIntegrationUserID,
			Status:    auth.SessionStatusActive,
			ExpiresAt: fixedIntegrationNow().Add(auth.SessionTTL),
			CreatedAt: fixedIntegrationNow(),
			UpdatedAt: fixedIntegrationNow(),
		},
		user: auth.UserContext{ID: targetJobHTTPIntegrationUserID, Email: "candidate@example.com"},
	}
	authService := auth.NewEmailCodeService(auth.EmailCodeServiceOptions{
		Store:               authStore,
		SessionCookieSecret: "session-secret",
		Now:                 fixedIntegrationNow,
	})
	service := targetjob.NewService(targetjob.ServiceOptions{
		Store:        store,
		NewID:        store.nextID,
		Now:          fixedIntegrationNow,
		DedupePepper: "integration-dedupe-pepper",
	})
	targetJobHandler := targetjob.NewHandler(targetjob.HandlerOptions{
		Service: service,
		Session: func(ctx context.Context) (string, bool) {
			current, ok := auth.CurrentSessionFromContext(ctx)
			if !ok || strings.TrimSpace(current.UserID) == "" {
				return "", false
			}
			return current.UserID, true
		},
	})
	aiClient := opts.ai
	if aiClient == nil {
		aiClient = targetjob.NewDeterministicParseAIClient(&integrationAIClient{resp: aiclient.CompleteResponse{Content: "delegated"}})
	}
	registry := opts.registry
	if registry == nil {
		registry = newStaticTestPromptRegistry()
	}
	executor := targetjob.NewParseExecutor(targetjob.ParseExecutorOptions{
		Store:    store,
		Registry: registry,
		AI:       aiClient,
		NewID:    store.nextID,
		Now:      fixedIntegrationNow,
	})
	runtime := newIntegrationJobRuntime(store, fixedIntegrationNow, map[string]runner.Handler{
		"target_import": executor,
	})
	return &targetJobHTTPIntegrationHarness{
		handler: buildAPIHandler(loader, apiRuntimeFlags{}, authService, targetJobHandler, practiceRoutes{}, uploadRoutes{}, resumeRoutes{}, reportRoutes{}, jobsRoutes{}),
		runtime: runtime,
		store:   store,
		cookie:  &http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"},
	}
}

func (h *targetJobHTTPIntegrationHarness) doJSON(t *testing.T, method, path string, idempotencyKey string, body any, wantStatus int) []byte {
	return doHTTPContractJSONWithCookie(t, h.handler, h.cookie, targetjob.IdempotencyKeyHeader, method, path, idempotencyKey, body, wantStatus)
}

func (h *targetJobHTTPIntegrationHarness) runRunnerOnce(t *testing.T, wantProcessed bool) {
	t.Helper()
	processed, err := h.runtime.RunOnce(context.Background())
	if err != nil {
		t.Fatalf("RunOnce: %v", err)
	}
	if processed != wantProcessed {
		t.Fatalf("RunOnce processed=%v want=%v", processed, wantProcessed)
	}
}

func fixedIntegrationNow() time.Time {
	return time.Date(2026, 5, 9, 23, 30, 0, 0, time.UTC)
}

func strPtr(v string) *string { return &v }

func decodeJSON[T any](t *testing.T, raw []byte, out *T) {
	t.Helper()
	if err := json.Unmarshal(raw, out); err != nil {
		t.Fatalf("decode JSON: %v; body=%s", err, string(raw))
	}
}

func assertNoEvidenceLeak(t *testing.T, payloads [][]byte, forbidden ...string) {
	t.Helper()
	joined := string(bytes.Join(payloads, []byte("\n")))
	for _, token := range forbidden {
		if token != "" && strings.Contains(joined, token) {
			t.Fatalf("integration output leaked forbidden token %q", token)
		}
	}
}

type integrationTargetJobStore struct {
	seq           int
	targets       map[string]targetjob.TargetJobRecord
	requirements  map[string][]targetjob.RequirementRecord
	importDedupe  map[string]string
	updateDedupe  map[string]string
	archiveDedupe map[string]string
	jobByTarget   map[string]string
	jobStatus     map[string]sharedtypes.JobStatus
	jobs          []runner.ClaimedJob
	finalized     map[string]runner.JobOutcome
	outbox        [][]byte
	failed        [][]byte
}

func newIntegrationTargetJobStore() *integrationTargetJobStore {
	return &integrationTargetJobStore{
		targets:       map[string]targetjob.TargetJobRecord{},
		requirements:  map[string][]targetjob.RequirementRecord{},
		importDedupe:  map[string]string{},
		updateDedupe:  map[string]string{},
		archiveDedupe: map[string]string{},
		jobByTarget:   map[string]string{},
		jobStatus:     map[string]sharedtypes.JobStatus{},
		finalized:     map[string]runner.JobOutcome{},
	}
}

func (s *integrationTargetJobStore) nextID() string {
	s.seq++
	return "integration-id-" + strconv.Itoa(s.seq)
}

func (s *integrationTargetJobStore) ImportTargetJob(_ context.Context, in targetjob.ImportTargetJobInput) (targetjob.ImportTargetJobResult, error) {
	if targetID, ok := s.importDedupe[in.DedupeKey]; ok {
		jobID := s.jobByTarget[targetID]
		return targetjob.ImportTargetJobResult{
			TargetJobID:  targetID,
			JobID:        jobID,
			JobStatus:    s.jobStatus[jobID],
			JobCreatedAt: in.Now,
			JobUpdatedAt: in.Now,
			Existing:     true,
		}, nil
	}
	rec := targetjob.TargetJobRecord{
		ID:             in.TargetJobID,
		UserID:         in.UserID,
		Status:         in.InitialLifecycleStatus,
		AnalysisStatus: in.InitialAnalysisStatus,
		Title:          in.Title,
		CompanyName:    in.CompanyName,
		LocationText:   in.LocationText,
		EmploymentType: in.EmploymentType,
		SeniorityLevel: in.SeniorityLevel,
		TargetLanguage: in.TargetLanguage,
		ResumeID:       in.ResumeID,
		RawJDText:      in.RawJDText,
		CreatedAt:      in.Now,
		UpdatedAt:      in.Now,
	}
	s.targets[in.TargetJobID] = rec
	s.importDedupe[in.DedupeKey] = in.TargetJobID
	status := sharedtypes.JobStatusQueued
	s.jobs = append(s.jobs, runner.ClaimedJob{
		JobID:        in.JobID,
		JobType:      "target_import",
		ResourceType: "target_job",
		ResourceID:   in.TargetJobID,
		Payload:      append([]byte{}, in.JobPayload...),
		MaxAttempts:  3,
		AvailableAt:  in.Now,
	})
	if len(in.OutboxEventPayload) > 0 {
		s.outbox = append(s.outbox, append([]byte{}, in.OutboxEventPayload...))
	}
	s.jobByTarget[in.TargetJobID] = in.JobID
	s.jobStatus[in.JobID] = status
	return targetjob.ImportTargetJobResult{
		TargetJobID:  in.TargetJobID,
		JobID:        in.JobID,
		JobStatus:    status,
		JobCreatedAt: in.Now,
		JobUpdatedAt: in.Now,
	}, nil
}

func (s *integrationTargetJobStore) InsertTargetJob(context.Context, targetjob.TargetJobRecord) error {
	panic("not used")
}

func (s *integrationTargetJobStore) GetTargetJobByUser(_ context.Context, userID string, targetJobID string) (targetjob.TargetJobRecord, []targetjob.RequirementRecord, error) {
	rec, ok := s.targets[targetJobID]
	if !ok || rec.UserID != userID || rec.DeletedAt != nil || rec.AnalysisStatus == sharedtypes.TargetJobParseStatusFailed {
		return targetjob.TargetJobRecord{}, nil, targetjob.ErrTargetJobNotFound
	}
	return rec, append([]targetjob.RequirementRecord{}, s.requirements[targetJobID]...), nil
}

func (s *integrationTargetJobStore) ListTargetJobsForUser(_ context.Context, userID string, filter targetjob.ListFilter) (targetjob.ListResult, error) {
	items := make([]targetjob.TargetJobRecord, 0, len(s.targets))
	for _, rec := range s.targets {
		if rec.UserID != userID {
			continue
		}
		if rec.DeletedAt != nil {
			continue
		}
		if rec.AnalysisStatus == sharedtypes.TargetJobParseStatusFailed {
			continue
		}
		if filter.Status != nil && rec.Status != *filter.Status {
			continue
		}
		if filter.AnalysisStatus != nil && rec.AnalysisStatus != *filter.AnalysisStatus {
			continue
		}
		items = append(items, rec)
	}
	return targetjob.ListResult{Items: items}, nil
}

func (s *integrationTargetJobStore) ArchiveTargetJob(_ context.Context, in targetjob.ArchiveTargetJobInput) (targetjob.TargetJobRecord, error) {
	if targetID, ok := s.archiveDedupe[in.DedupeKey]; ok {
		rec, ok := s.targets[targetID]
		if !ok || rec.UserID != in.UserID {
			return targetjob.TargetJobRecord{}, targetjob.ErrTargetJobNotFound
		}
		return rec, nil
	}
	rec, ok := s.targets[in.TargetJobID]
	if !ok || rec.UserID != in.UserID {
		return targetjob.TargetJobRecord{}, targetjob.ErrTargetJobNotFound
	}
	if rec.DeletedAt != nil {
		return targetjob.TargetJobRecord{}, targetjob.ErrTargetJobAlreadyArchived
	}
	deletedAt := in.Now
	rec.Status = sharedtypes.TargetJobStatusArchived
	rec.DeletedAt = &deletedAt
	rec.UpdatedAt = in.Now
	s.targets[in.TargetJobID] = rec
	s.archiveDedupe[in.DedupeKey] = in.TargetJobID
	return rec, nil
}

func (s *integrationTargetJobStore) LookupUpdateDedupe(_ context.Context, userID string, dedupeKey string) (targetjob.TargetJobRecord, []targetjob.RequirementRecord, bool, error) {
	targetID, ok := s.updateDedupe[dedupeKey]
	if !ok {
		return targetjob.TargetJobRecord{}, nil, false, nil
	}
	rec, ok := s.targets[targetID]
	if !ok || rec.UserID != userID {
		return targetjob.TargetJobRecord{}, nil, false, targetjob.ErrTargetJobNotFound
	}
	return rec, append([]targetjob.RequirementRecord{}, s.requirements[targetID]...), true, nil
}

func (s *integrationTargetJobStore) UpdateTargetJobLifecycle(_ context.Context, userID string, targetJobID string, fields targetjob.UpdateLifecycleFields, now time.Time) (targetjob.TargetJobRecord, error) {
	rec, ok := s.targets[targetJobID]
	if !ok || rec.UserID != userID {
		return targetjob.TargetJobRecord{}, targetjob.ErrTargetJobNotFound
	}
	if fields.Status != nil {
		rec.Status = *fields.Status
	}
	if fields.LocationText != nil {
		rec.LocationText = *fields.LocationText
	}
	if fields.Notes != nil {
		rec.Notes = *fields.Notes
	}
	if fields.TitleHint != nil {
		rec.Title = *fields.TitleHint
	}
	if fields.CompanyNameHint != nil {
		rec.CompanyName = *fields.CompanyNameHint
	}
	rec.UpdatedAt = now
	s.targets[targetJobID] = rec
	if fields.DedupeKey != "" {
		s.updateDedupe[fields.DedupeKey] = targetJobID
	}
	return rec, nil
}

func (s *integrationTargetJobStore) ApplyParseResult(_ context.Context, in targetjob.ApplyParseResultInput) error {
	rec := s.targets[in.TargetJobID]
	rec.AnalysisStatus = in.AnalysisStatus
	rec.Summary = append([]byte{}, in.Summary...)
	rec.FitSummary = append([]byte{}, in.FitSummary...)
	rec.UpdatedAt = in.Now
	s.targets[in.TargetJobID] = rec
	s.requirements[in.TargetJobID] = normalizeRequirements(in.TargetJobID, in.Requirements, in.Now)
	return nil
}

func (s *integrationTargetJobStore) CompleteParseSuccess(_ context.Context, in targetjob.CompleteParseSuccessInput) error {
	if err := s.ApplyParseResult(context.Background(), targetjob.ApplyParseResultInput{
		TargetJobID:    in.TargetJobID,
		AnalysisStatus: in.AnalysisStatus,
		Summary:        in.Summary,
		FitSummary:     in.FitSummary,
		Requirements:   in.Requirements,
		Now:            in.Now,
	}); err != nil {
		return err
	}
	s.outbox = append(s.outbox, append([]byte{}, in.ParsedEventPayload...))
	return nil
}

func (s *integrationTargetJobStore) CompleteParseFailure(_ context.Context, in targetjob.CompleteParseFailureInput) error {
	payload := append([]byte{}, in.FailedEventPayload...)
	s.failed = append(s.failed, payload)
	s.outbox = append(s.outbox, payload)
	delete(s.targets, in.TargetJobID)
	delete(s.requirements, in.TargetJobID)
	return nil
}

func (s *integrationTargetJobStore) LeaseAsyncJob(_ context.Context, jobTypes []string, _ time.Time) (runner.ClaimedJob, bool, error) {
	allowed := map[string]struct{}{}
	for _, jobType := range jobTypes {
		allowed[jobType] = struct{}{}
	}
	for i, job := range s.jobs {
		if _, ok := allowed[job.JobType]; !ok {
			continue
		}
		s.jobs = append(s.jobs[:i], s.jobs[i+1:]...)
		return job, true, nil
	}
	return runner.ClaimedJob{}, false, nil
}

func (s *integrationTargetJobStore) FinalizeAsyncJob(_ context.Context, jobID string, _ int32, outcome runner.JobOutcome, _ time.Time, _ time.Time) error {
	s.finalized[jobID] = outcome
	if outcome.Succeeded {
		s.jobStatus[jobID] = sharedtypes.JobStatusSucceeded
	} else {
		s.jobStatus[jobID] = sharedtypes.JobStatusFailed
	}
	return nil
}

func (s *integrationTargetJobStore) ReclaimExpiredLeases(context.Context, []string, time.Time, time.Time) (int64, error) {
	return 0, nil
}

func (s *integrationTargetJobStore) WriteParseFailedOutbox(_ context.Context, _ string, _ string, payload []byte, _ time.Time) error {
	s.failed = append(s.failed, append([]byte{}, payload...))
	s.outbox = append(s.outbox, append([]byte{}, payload...))
	return nil
}

func (s *integrationTargetJobStore) WriteTargetParsedOutbox(_ context.Context, _ string, _ string, payload []byte, _ time.Time) error {
	s.outbox = append(s.outbox, append([]byte{}, payload...))
	return nil
}

func (s *integrationTargetJobStore) GetTargetJobForParse(_ context.Context, targetJobID string) (targetjob.TargetJobRecord, error) {
	rec, ok := s.targets[targetJobID]
	if !ok {
		return targetjob.TargetJobRecord{}, targetjob.ErrTargetJobNotFound
	}
	return rec, nil
}

func (s *integrationTargetJobStore) targetCount() int { return len(s.targets) }

func (s *integrationTargetJobStore) outboxPayloads() [][]byte {
	out := make([][]byte, 0, len(s.outbox))
	for _, p := range s.outbox {
		out = append(out, append([]byte{}, p...))
	}
	return out
}

func (s *integrationTargetJobStore) finalizedOutcome(jobID string) runner.JobOutcome {
	return s.finalized[jobID]
}

func (s *integrationTargetJobStore) lastFailedPayload() []byte {
	if len(s.failed) == 0 {
		return nil
	}
	return s.failed[len(s.failed)-1]
}

func normalizeRequirements(targetID string, in []targetjob.RequirementRecord, now time.Time) []targetjob.RequirementRecord {
	out := make([]targetjob.RequirementRecord, 0, len(in))
	for i, req := range in {
		cp := req
		cp.TargetJobID = targetID
		if cp.DisplayOrder == 0 {
			cp.DisplayOrder = int32(i + 1)
		}
		if cp.EvidenceLevel == "" {
			cp.EvidenceLevel = targetjob.EvidenceExplicit
		}
		if cp.CreatedAt.IsZero() {
			cp.CreatedAt = now
		}
		out = append(out, cp)
	}
	return out
}

type integrationAIClient struct {
	resp aiclient.CompleteResponse
	err  error
}

func (c *integrationAIClient) Complete(_ context.Context, profileName string, payload aiclient.CompletePayload) (aiclient.CompleteResponse, aiclient.AICallMeta, error) {
	if c.err != nil {
		return aiclient.CompleteResponse{}, aiclient.AICallMeta{}, c.err
	}
	return c.resp, aiclient.AICallMeta{
		Provider:          "integration-test-provider",
		ModelFamily:       "fixture",
		ModelID:           "fixture-model:target-import-parse",
		ModelProfileName:  profileName,
		Language:          payload.Metadata.Language,
		PromptVersion:     payload.Metadata.PromptVersion,
		RubricVersion:     payload.Metadata.RubricVersion,
		FeatureKey:        payload.Metadata.FeatureKey,
		FeatureFlag:       payload.Metadata.FeatureFlag,
		DataSourceVersion: payload.Metadata.DataSourceVersion,
		ValidationStatus:  aiclient.ValidationStatusOK,
	}, nil
}

func (c *integrationAIClient) Transcribe(context.Context, string, aiclient.TranscriptionInput) (aiclient.TranscriptionResponse, aiclient.AICallMeta, error) {
	return aiclient.TranscriptionResponse{}, aiclient.AICallMeta{}, errors.New("unexpected Transcribe call in targetjob integrationAIClient")
}

func (c *integrationAIClient) Stream(context.Context, string, aiclient.CompletePayload) (<-chan aiclient.AIStreamEvent, error) {
	return nil, errors.New("unexpected Stream call in targetjob integrationAIClient")
}

func (c *integrationAIClient) Synthesize(context.Context, string, aiclient.SynthesisInput) (aiclient.SynthesisResponse, aiclient.AICallMeta, error) {
	return aiclient.SynthesisResponse{}, aiclient.AICallMeta{}, errors.New("unexpected Synthesize call in targetjob integrationAIClient")
}

type integrationRegistry struct {
	err error
}

func (r integrationRegistry) Resolve(ctx context.Context, featureKey string, language string) (targetjob.PromptResolution, error) {
	if r.err != nil {
		return targetjob.PromptResolution{}, r.err
	}
	return newStaticTestPromptRegistry().Resolve(ctx, featureKey, language)
}

// staticTestPromptRegistry replaces the out-of-scope targetjob.StaticPromptRegistry
// for cmd/api integration tests. It mirrors the F3 RegistryAdapter shape with a
// fixed target.import.parse resolution so HTTP integration tests can assert the
// provenance flow without spinning up a real registry.Client.
type staticTestPromptRegistry struct {
	resolution targetjob.PromptResolution
}

func newStaticTestPromptRegistry() *staticTestPromptRegistry {
	return &staticTestPromptRegistry{
		resolution: targetjob.PromptResolution{
			PromptVersion:       "v0.1.0",
			RubricVersion:       "v0.1.0",
			ModelProfileName:    "target.import.default",
			DataSourceVersion:   "registry.v1",
			FeatureFlag:         "none",
			UserMessageTemplate: "{{jd_text}}",
		},
	}
}

func (r *staticTestPromptRegistry) Resolve(_ context.Context, featureKey string, language string) (targetjob.PromptResolution, error) {
	if featureKey != targetjob.FeatureKeyTargetImportParse || strings.TrimSpace(language) == "" {
		return targetjob.PromptResolution{}, targetjob.ErrPromptUnsupported
	}
	return r.resolution, nil
}
