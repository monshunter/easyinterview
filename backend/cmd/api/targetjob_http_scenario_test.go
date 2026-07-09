package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient"
	api "github.com/monshunter/easyinterview/backend/internal/api/generated"
	"github.com/monshunter/easyinterview/backend/internal/auth"
	"github.com/monshunter/easyinterview/backend/internal/platform/config"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
	"github.com/monshunter/easyinterview/backend/internal/targetjob"
	"github.com/monshunter/easyinterview/backend/internal/targetjob/urlfetch"
)

const (
	targetJobHTTPScenarioUserID   = "scenario-user-targetjob-http"
	targetJobHTTPScenarioResumeID = "scenario-resume-targetjob-http"
)

func TestE2EP0010HTTPTextImportParseReady(t *testing.T) {
	h := newTargetJobHTTPScenarioHarness(t, targetJobHTTPScenarioOptions{})

	importBody := api.ImportTargetJobRequest{
		Source: map[string]any{
			"type":    "manual_text",
			"rawText": "Private scenario JD text that must stay out of evidence logs.",
		},
		TargetLanguage:  "zh-CN",
		ResumeId:        targetJobHTTPScenarioResumeID,
		TitleHint:       strPtr("Senior Frontend Engineer"),
		CompanyNameHint: strPtr("Acme"),
	}
	raw := h.doJSON(t, http.MethodPost, "/api/v1/targets/import", "e2e-p0-010-import", importBody, http.StatusAccepted)
	var imported api.TargetJobWithJob
	decodeJSON(t, raw, &imported)
	if imported.TargetJobId == "" || imported.Job.Status != sharedtypes.JobStatusQueued {
		t.Fatalf("unexpected import response: %+v", imported)
	}

	duplicateRaw := h.doJSON(t, http.MethodPost, "/api/v1/targets/import", "e2e-p0-010-import", importBody, http.StatusAccepted)
	var duplicate api.TargetJobWithJob
	decodeJSON(t, duplicateRaw, &duplicate)
	if duplicate.TargetJobId != imported.TargetJobId || h.store.targetCount() != 1 {
		t.Fatalf("idempotent import did not return existing target: duplicate=%+v targets=%d", duplicate, h.store.targetCount())
	}

	h.runDrainerOnce(t, true)

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
	updateRaw := h.doJSON(t, http.MethodPatch, "/api/v1/targets/"+imported.TargetJobId, "e2e-p0-010-update", api.UpdateTargetJobRequest{
		Status: &status,
		Notes:  strPtr("Recruiter asked for platform examples."),
	}, http.StatusOK)
	var updated api.TargetJob
	decodeJSON(t, updateRaw, &updated)
	if updated.Status != sharedtypes.TargetJobStatusPreparing || updated.AnalysisStatus != sharedtypes.TargetJobParseStatusReady {
		t.Fatalf("update changed wrong fields: %+v", updated)
	}

	assertNoEvidenceLeak(t, h.store.outboxPayloads(), "Private scenario JD text", "prompt body", "response body", "Authorization:")
}

func TestE2EP0011HTTPURLImportFetchAndParse(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/role/1" {
			t.Fatalf("unexpected URL fetch path: %s", r.URL.String())
		}
		_, _ = io.WriteString(w, "Fetched public JD text that must not appear in scenario evidence.")
	}))
	defer server.Close()
	serverURL, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("parse fixture server URL: %v", err)
	}

	fetcher := urlfetch.New(urlfetch.FetcherOptions{
		UserAgent: targetjob.URLFetchUserAgent("scenario"),
		Resolver: func(context.Context, string) ([]net.IP, error) {
			return []net.IP{net.ParseIP("203.0.113.10")}, nil
		},
		HTTPClient: &http.Client{Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // local httptest certificate
			DialContext: func(ctx context.Context, network string, _ string) (net.Conn, error) {
				var d net.Dialer
				return d.DialContext(ctx, network, serverURL.Host)
			},
		}},
		Now: fixedScenarioNow,
	})
	h := newTargetJobHTTPScenarioHarness(t, targetJobHTTPScenarioOptions{fetcher: fetcher})

	raw := h.doJSON(t, http.MethodPost, "/api/v1/targets/import", "e2e-p0-011-url", api.ImportTargetJobRequest{
		Source:         map[string]any{"type": "url", "url": "https://jobs.example.test/role/1?token=secret#frag"},
		TargetLanguage: "en",
		ResumeId:       targetJobHTTPScenarioResumeID,
	}, http.StatusAccepted)
	var imported api.TargetJobWithJob
	decodeJSON(t, raw, &imported)
	h.runDrainerOnce(t, true)

	detailRaw := h.doJSON(t, http.MethodGet, "/api/v1/targets/"+imported.TargetJobId, "", nil, http.StatusOK)
	var detail api.TargetJob
	decodeJSON(t, detailRaw, &detail)
	if detail.AnalysisStatus != sharedtypes.TargetJobParseStatusReady || detail.SourceUrl == nil || strings.Contains(*detail.SourceUrl, "token=secret") {
		t.Fatalf("URL detail not ready or not sanitized: %+v", detail)
	}
	source := h.store.firstSource(imported.TargetJobId)
	if source.SnapshotText == "" || source.FetchedAt == nil || source.FreshnessStatus != targetjob.FreshnessFresh {
		t.Fatalf("URL source snapshot not persisted: %+v", source)
	}
	if strings.Contains(source.URL, "token=secret") || strings.Contains(source.URL, "#frag") {
		t.Fatalf("source URL leaked secret material: %q", source.URL)
	}

	errorRaw := h.doJSON(t, http.MethodPost, "/api/v1/targets/import", "e2e-p0-011-invalid", api.ImportTargetJobRequest{
		Source:         map[string]any{"type": "url", "url": "http://169.254.169.254/latest/meta-data"},
		TargetLanguage: "en",
		ResumeId:       targetJobHTTPScenarioResumeID,
	}, http.StatusBadRequest)
	var errResp api.ApiErrorResponse
	decodeJSON(t, errorRaw, &errResp)
	if errResp.Error.Code != sharederrors.CodeTargetImportSourceInvalid || errResp.Error.Retryable {
		t.Fatalf("invalid URL did not map to TARGET_IMPORT_SOURCE_INVALID: %+v", errResp.Error)
	}

	assertNoEvidenceLeak(t, h.store.outboxPayloads(), "token=secret", "Fetched public JD text", "latest/meta-data", "Authorization:")
}

func TestE2EP0012HTTPParseFailureRetryableAndNonRetryable(t *testing.T) {
	cases := []struct {
		name      string
		options   targetJobHTTPScenarioOptions
		wantCode  string
		retryable bool
	}{
		{
			name:      "provider-timeout",
			options:   targetJobHTTPScenarioOptions{ai: &scenarioAIClient{err: errors.New("provider error: AI_PROVIDER_TIMEOUT")}},
			wantCode:  sharederrors.CodeAiProviderTimeout,
			retryable: true,
		},
		{
			name:      "invalid-output",
			options:   targetJobHTTPScenarioOptions{ai: &scenarioAIClient{resp: aiclient.CompleteResponse{Content: "not-json"}}},
			wantCode:  sharederrors.CodeAiOutputInvalid,
			retryable: false,
		},
		{
			name:      "registry-disabled",
			options:   targetJobHTTPScenarioOptions{registry: scenarioRegistry{err: targetjob.ErrPromptUnsupported}},
			wantCode:  sharederrors.CodeAiProviderConfigInvalid,
			retryable: false,
		},
		{
			name:      "secret-missing",
			options:   targetJobHTTPScenarioOptions{ai: &scenarioAIClient{err: errors.New("AI_PROVIDER_SECRET_MISSING provider secret missing")}},
			wantCode:  sharederrors.CodeAiProviderSecretMissing,
			retryable: false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := newTargetJobHTTPScenarioHarness(t, tc.options)
			raw := h.doJSON(t, http.MethodPost, "/api/v1/targets/import", "e2e-p0-012-"+tc.name, api.ImportTargetJobRequest{
				Source:         map[string]any{"type": "manual_text", "rawText": "Private JD body that must not leak."},
				TargetLanguage: "en",
				ResumeId:       targetJobHTTPScenarioResumeID,
			}, http.StatusAccepted)
			var imported api.TargetJobWithJob
			decodeJSON(t, raw, &imported)

			h.runDrainerOnce(t, true)
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

func TestE2EP0013HTTPManualFormReady(t *testing.T) {
	h := newTargetJobHTTPScenarioHarness(t, targetJobHTTPScenarioOptions{})
	raw := h.doJSON(t, http.MethodPost, "/api/v1/targets/import", "e2e-p0-013-manual-form", api.ImportTargetJobRequest{
		Source: map[string]any{
			"type":           "manual_form",
			"title":          "Frontend Architect",
			"companyName":    "Acme",
			"rawDescription": "Lead frontend architecture across squads. Must have React platform experience.",
		},
		TargetLanguage: "zh-CN",
		ResumeId:       targetJobHTTPScenarioResumeID,
	}, http.StatusAccepted)
	var imported api.TargetJobWithJob
	decodeJSON(t, raw, &imported)
	if imported.Job.Status != sharedtypes.JobStatusSucceeded || imported.Job.JobType != api.JobTypeTargetImport {
		t.Fatalf("manual_form must return terminal target_import job: %+v", imported.Job)
	}
	if processed, err := h.drainer.RunOnce(context.Background()); err != nil || processed {
		t.Fatalf("manual_form must not queue target_import runner, processed=%v err=%v", processed, err)
	}

	detailRaw := h.doJSON(t, http.MethodGet, "/api/v1/targets/"+imported.TargetJobId, "", nil, http.StatusOK)
	var detail api.TargetJob
	decodeJSON(t, detailRaw, &detail)
	if detail.AnalysisStatus != sharedtypes.TargetJobParseStatusReady || len(detail.Requirements) == 0 || detail.SourceType != string(targetjob.SourceTypeManualForm) {
		t.Fatalf("manual_form detail not ready: %+v", detail)
	}
	listRaw := h.doJSON(t, http.MethodGet, "/api/v1/targets", "", nil, http.StatusOK)
	var list api.PaginatedTargetJob
	decodeJSON(t, listRaw, &list)
	if len(list.Items) != 1 || list.Items[0].Id != imported.TargetJobId {
		t.Fatalf("manual_form target missing from list: %+v", list)
	}
	assertNoEvidenceLeak(t, h.store.outboxPayloads(), "Lead frontend architecture across squads", "prompt body", "response body", "Authorization:")
}

type targetJobHTTPScenarioOptions struct {
	ai       aiclient.AIClient
	registry targetjob.PromptRegistryClient
	fetcher  targetjob.URLFetcher
}

type targetJobHTTPScenarioHarness struct {
	handler http.Handler
	drainer *targetjob.Drainer
	store   *scenarioTargetJobStore
	cookie  *http.Cookie
}

func newTargetJobHTTPScenarioHarness(t *testing.T, opts targetJobHTTPScenarioOptions) *targetJobHTTPScenarioHarness {
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
	store := newScenarioTargetJobStore()
	authStore := &apiAuthStore{
		session: auth.SessionRecord{
			ID:        "scenario-session-1",
			UserID:    targetJobHTTPScenarioUserID,
			Status:    auth.SessionStatusActive,
			ExpiresAt: fixedScenarioNow().Add(auth.SessionTTL),
			CreatedAt: fixedScenarioNow(),
			UpdatedAt: fixedScenarioNow(),
		},
		user: auth.UserContext{ID: targetJobHTTPScenarioUserID, Email: "candidate@example.com"},
	}
	authService := auth.NewEmailCodeService(auth.EmailCodeServiceOptions{
		Store:               authStore,
		SessionCookieSecret: "session-secret",
		Now:                 fixedScenarioNow,
	})
	service := targetjob.NewService(targetjob.ServiceOptions{
		Store:        store,
		NewID:        store.nextID,
		Now:          fixedScenarioNow,
		DedupePepper: "scenario-dedupe-pepper",
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
		aiClient = targetjob.NewDeterministicParseAIClient(&scenarioAIClient{resp: aiclient.CompleteResponse{Content: "delegated"}})
	}
	registry := opts.registry
	if registry == nil {
		registry = newStaticTestPromptRegistry()
	}
	fetcher := opts.fetcher
	if fetcher == nil {
		fetcher = &scenarioFetcher{}
	}
	executor := targetjob.NewParseExecutor(targetjob.ParseExecutorOptions{
		Store:    store,
		Registry: registry,
		AI:       aiClient,
		Fetcher:  fetcher,
		NewID:    store.nextID,
		Now:      fixedScenarioNow,
	})
	drainer := targetjob.NewDrainer(targetjob.DrainerOptions{
		Store: store,
		Handlers: map[string]targetjob.JobHandler{
			"target_import":  executor,
			"source_refresh": &targetjob.SourceRefreshHandler{Store: store, Now: fixedScenarioNow},
		},
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	})
	return &targetJobHTTPScenarioHarness{
		handler: buildAPIHandlerWithTargetJobHandler(loader, apiRuntimeFlags{}, authService, targetJobHandler),
		drainer: drainer,
		store:   store,
		cookie:  &http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"},
	}
}

func (h *targetJobHTTPScenarioHarness) doJSON(t *testing.T, method, path string, idempotencyKey string, body any, wantStatus int) []byte {
	t.Helper()
	var reader *bytes.Reader
	if body == nil {
		reader = bytes.NewReader(nil)
	} else {
		raw, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request body: %v", err)
		}
		reader = bytes.NewReader(raw)
	}
	req := httptest.NewRequest(method, path, reader)
	req.AddCookie(h.cookie)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if idempotencyKey != "" {
		req.Header.Set(targetjob.IdempotencyKeyHeader, idempotencyKey)
	}
	rec := httptest.NewRecorder()
	h.handler.ServeHTTP(rec, req)
	if rec.Code != wantStatus {
		t.Fatalf("%s %s status=%d want=%d body=%s", method, path, rec.Code, wantStatus, rec.Body.String())
	}
	return rec.Body.Bytes()
}

func (h *targetJobHTTPScenarioHarness) runDrainerOnce(t *testing.T, wantProcessed bool) {
	t.Helper()
	processed, err := h.drainer.RunOnce(context.Background())
	if err != nil {
		t.Fatalf("RunOnce: %v", err)
	}
	if processed != wantProcessed {
		t.Fatalf("RunOnce processed=%v want=%v", processed, wantProcessed)
	}
}

func fixedScenarioNow() time.Time {
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
			t.Fatalf("scenario evidence leaked forbidden token %q", token)
		}
	}
}

type scenarioTargetJobStore struct {
	seq           int
	targets       map[string]targetjob.TargetJobRecord
	requirements  map[string][]targetjob.RequirementRecord
	sources       map[string][]targetjob.SourceRecord
	importDedupe  map[string]string
	updateDedupe  map[string]string
	archiveDedupe map[string]string
	jobByTarget   map[string]string
	jobStatus     map[string]sharedtypes.JobStatus
	jobs          []targetjob.ClaimedJob
	finalized     map[string]targetjob.JobOutcome
	outbox        [][]byte
	failed        [][]byte
}

func newScenarioTargetJobStore() *scenarioTargetJobStore {
	return &scenarioTargetJobStore{
		targets:       map[string]targetjob.TargetJobRecord{},
		requirements:  map[string][]targetjob.RequirementRecord{},
		sources:       map[string][]targetjob.SourceRecord{},
		importDedupe:  map[string]string{},
		updateDedupe:  map[string]string{},
		archiveDedupe: map[string]string{},
		jobByTarget:   map[string]string{},
		jobStatus:     map[string]sharedtypes.JobStatus{},
		finalized:     map[string]targetjob.JobOutcome{},
	}
}

func (s *scenarioTargetJobStore) nextID() string {
	s.seq++
	return "scenario-id-" + strconv.Itoa(s.seq)
}

func (s *scenarioTargetJobStore) ImportTargetJob(_ context.Context, in targetjob.ImportTargetJobInput) (targetjob.ImportTargetJobResult, error) {
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
		ID:                 in.TargetJobID,
		UserID:             in.UserID,
		Status:             in.InitialLifecycleStatus,
		AnalysisStatus:     in.InitialAnalysisStatus,
		Title:              in.Title,
		CompanyName:        in.CompanyName,
		LocationText:       in.LocationText,
		EmploymentType:     in.EmploymentType,
		SeniorityLevel:     in.SeniorityLevel,
		TargetLanguage:     in.TargetLanguage,
		SourceType:         in.APISourceType,
		SourceURL:          in.SourceURL,
		SourceFileObjectID: in.SourceFileObjectID,
		RawJDText:          in.RawJDText,
		CreatedAt:          in.Now,
		UpdatedAt:          in.Now,
	}
	s.targets[in.TargetJobID] = rec
	s.importDedupe[in.DedupeKey] = in.TargetJobID
	if in.SourceID != "" {
		s.sources[in.TargetJobID] = append(s.sources[in.TargetJobID], targetjob.SourceRecord{
			ID:              in.SourceID,
			TargetJobID:     in.TargetJobID,
			SourceType:      in.APISourceType,
			URL:             in.SourceURL,
			FileObjectID:    in.SourceFileObjectID,
			SnapshotText:    in.SourceSnapshotText,
			FetchedAt:       in.SourceFetchedAt,
			FreshnessStatus: targetjob.FreshnessFresh,
			CreatedAt:       in.Now,
		})
	}
	if len(in.DraftRequirements) > 0 {
		s.requirements[in.TargetJobID] = normalizeRequirements(in.TargetJobID, in.DraftRequirements, in.Now)
	}
	status := sharedtypes.JobStatusSucceeded
	if in.JobID != "" && len(in.JobPayload) > 0 {
		status = sharedtypes.JobStatusQueued
		s.jobs = append(s.jobs, targetjob.ClaimedJob{
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

func (s *scenarioTargetJobStore) InsertTargetJob(context.Context, targetjob.TargetJobRecord) error {
	panic("not used")
}

func (s *scenarioTargetJobStore) InsertTargetJobSource(context.Context, targetjob.SourceRecord) error {
	panic("not used")
}

func (s *scenarioTargetJobStore) GetTargetJobByUser(_ context.Context, userID string, targetJobID string) (targetjob.TargetJobRecord, []targetjob.RequirementRecord, []targetjob.SourceRecord, error) {
	rec, ok := s.targets[targetJobID]
	if !ok || rec.UserID != userID || rec.DeletedAt != nil || rec.AnalysisStatus == sharedtypes.TargetJobParseStatusFailed {
		return targetjob.TargetJobRecord{}, nil, nil, targetjob.ErrTargetJobNotFound
	}
	return rec, append([]targetjob.RequirementRecord{}, s.requirements[targetJobID]...), append([]targetjob.SourceRecord{}, s.sources[targetJobID]...), nil
}

func (s *scenarioTargetJobStore) ListTargetJobsForUser(_ context.Context, userID string, filter targetjob.ListFilter) (targetjob.ListResult, error) {
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

func (s *scenarioTargetJobStore) ArchiveTargetJob(_ context.Context, in targetjob.ArchiveTargetJobInput) (targetjob.TargetJobRecord, error) {
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

func (s *scenarioTargetJobStore) LookupUpdateDedupe(_ context.Context, userID string, dedupeKey string) (targetjob.TargetJobRecord, []targetjob.RequirementRecord, bool, error) {
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

func (s *scenarioTargetJobStore) UpdateTargetJobLifecycle(_ context.Context, userID string, targetJobID string, fields targetjob.UpdateLifecycleFields, now time.Time) (targetjob.TargetJobRecord, error) {
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

func (s *scenarioTargetJobStore) ApplyParseResult(_ context.Context, in targetjob.ApplyParseResultInput) error {
	rec := s.targets[in.TargetJobID]
	rec.AnalysisStatus = in.AnalysisStatus
	rec.Summary = append([]byte{}, in.Summary...)
	rec.FitSummary = append([]byte{}, in.FitSummary...)
	rec.UpdatedAt = in.Now
	s.targets[in.TargetJobID] = rec
	s.requirements[in.TargetJobID] = normalizeRequirements(in.TargetJobID, in.Requirements, in.Now)
	return nil
}

func (s *scenarioTargetJobStore) CompleteParseSuccess(_ context.Context, in targetjob.CompleteParseSuccessInput) error {
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
	if in.SourceRefreshJobID != "" {
		s.jobs = append(s.jobs, targetjob.ClaimedJob{
			JobID:        in.SourceRefreshJobID,
			JobType:      "source_refresh",
			ResourceType: "target_job",
			ResourceID:   in.TargetJobID,
			MaxAttempts:  3,
			AvailableAt:  in.Now,
		})
		s.jobStatus[in.SourceRefreshJobID] = sharedtypes.JobStatusQueued
	}
	return nil
}

func (s *scenarioTargetJobStore) CompleteParseFailure(_ context.Context, in targetjob.CompleteParseFailureInput) error {
	payload := append([]byte{}, in.FailedEventPayload...)
	s.failed = append(s.failed, payload)
	s.outbox = append(s.outbox, payload)
	delete(s.targets, in.TargetJobID)
	delete(s.requirements, in.TargetJobID)
	delete(s.sources, in.TargetJobID)
	return nil
}

func (s *scenarioTargetJobStore) UpdateSourceFreshness(_ context.Context, targetJobID string, freshness targetjob.FreshnessStatus, _ time.Time) error {
	sources := s.sources[targetJobID]
	for i := range sources {
		sources[i].FreshnessStatus = freshness
	}
	s.sources[targetJobID] = sources
	return nil
}

func (s *scenarioTargetJobStore) UpdateSourceSnapshot(_ context.Context, sourceID string, sanitizedURL string, snapshotText string, fetchedAt time.Time, _ time.Time) error {
	for targetID, sources := range s.sources {
		for i := range sources {
			if sources[i].ID == sourceID {
				sources[i].URL = sanitizedURL
				sources[i].SnapshotText = snapshotText
				sources[i].FetchedAt = &fetchedAt
				sources[i].FreshnessStatus = targetjob.FreshnessFresh
				s.sources[targetID] = sources
				rec := s.targets[targetID]
				rec.SourceURL = sanitizedURL
				s.targets[targetID] = rec
				return nil
			}
		}
	}
	return targetjob.ErrTargetJobNotFound
}

func (s *scenarioTargetJobStore) LookupFileAttachmentForUser(context.Context, string, string) (targetjob.FileAttachmentRecord, error) {
	return targetjob.FileAttachmentRecord{}, targetjob.ErrTargetJobNotFound
}

func (s *scenarioTargetJobStore) ClaimNextAsyncJob(_ context.Context, jobTypes []string, _ time.Time) (targetjob.ClaimedJob, bool, error) {
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
	return targetjob.ClaimedJob{}, false, nil
}

func (s *scenarioTargetJobStore) FinalizeAsyncJob(_ context.Context, jobID string, outcome targetjob.JobOutcome, _ time.Time) error {
	s.finalized[jobID] = outcome
	if outcome.Succeeded {
		s.jobStatus[jobID] = sharedtypes.JobStatusSucceeded
	} else {
		s.jobStatus[jobID] = sharedtypes.JobStatusFailed
	}
	return nil
}

func (s *scenarioTargetJobStore) EnqueueSourceRefresh(_ context.Context, jobID string, targetJobID string, now time.Time) error {
	s.jobs = append(s.jobs, targetjob.ClaimedJob{JobID: jobID, JobType: "source_refresh", ResourceType: "target_job", ResourceID: targetJobID, AvailableAt: now})
	return nil
}

func (s *scenarioTargetJobStore) WriteParseFailedOutbox(_ context.Context, _ string, _ string, payload []byte, _ time.Time) error {
	s.failed = append(s.failed, append([]byte{}, payload...))
	s.outbox = append(s.outbox, append([]byte{}, payload...))
	return nil
}

func (s *scenarioTargetJobStore) WriteTargetParsedOutbox(_ context.Context, _ string, _ string, payload []byte, _ time.Time) error {
	s.outbox = append(s.outbox, append([]byte{}, payload...))
	return nil
}

func (s *scenarioTargetJobStore) GetTargetJobForParse(_ context.Context, targetJobID string) (targetjob.TargetJobRecord, []targetjob.SourceRecord, error) {
	rec, ok := s.targets[targetJobID]
	if !ok {
		return targetjob.TargetJobRecord{}, nil, targetjob.ErrTargetJobNotFound
	}
	return rec, append([]targetjob.SourceRecord{}, s.sources[targetJobID]...), nil
}

func (s *scenarioTargetJobStore) UpdateTargetJobAnalysisFailure(_ context.Context, targetJobID string, _ time.Time) error {
	delete(s.targets, targetJobID)
	delete(s.requirements, targetJobID)
	delete(s.sources, targetJobID)
	return nil
}

func (s *scenarioTargetJobStore) targetCount() int { return len(s.targets) }

func (s *scenarioTargetJobStore) firstSource(targetID string) targetjob.SourceRecord {
	if len(s.sources[targetID]) == 0 {
		return targetjob.SourceRecord{}
	}
	return s.sources[targetID][0]
}

func (s *scenarioTargetJobStore) outboxPayloads() [][]byte {
	out := make([][]byte, 0, len(s.outbox))
	for _, p := range s.outbox {
		out = append(out, append([]byte{}, p...))
	}
	return out
}

func (s *scenarioTargetJobStore) finalizedOutcome(jobID string) targetjob.JobOutcome {
	return s.finalized[jobID]
}

func (s *scenarioTargetJobStore) lastFailedPayload() []byte {
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

type scenarioAIClient struct {
	resp aiclient.CompleteResponse
	err  error
}

func (c *scenarioAIClient) Complete(_ context.Context, profileName string, payload aiclient.CompletePayload) (aiclient.CompleteResponse, aiclient.AICallMeta, error) {
	if c.err != nil {
		return aiclient.CompleteResponse{}, aiclient.AICallMeta{}, c.err
	}
	return c.resp, aiclient.AICallMeta{
		Provider:          "scenario-test-provider",
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

func (c *scenarioAIClient) Transcribe(context.Context, string, aiclient.TranscriptionInput) (aiclient.TranscriptionResponse, aiclient.AICallMeta, error) {
	return aiclient.TranscriptionResponse{}, aiclient.AICallMeta{}, errors.New("not implemented")
}

func (c *scenarioAIClient) Stream(context.Context, string, aiclient.CompletePayload) (<-chan aiclient.AIStreamEvent, error) {
	return nil, errors.New("not implemented")
}

func (c *scenarioAIClient) Synthesize(context.Context, string, aiclient.SynthesisInput) (aiclient.SynthesisResponse, aiclient.AICallMeta, error) {
	return aiclient.SynthesisResponse{}, aiclient.AICallMeta{}, errors.New("not implemented")
}

type scenarioRegistry struct {
	err error
}

func (r scenarioRegistry) Resolve(ctx context.Context, featureKey string, language string) (targetjob.PromptResolution, error) {
	if r.err != nil {
		return targetjob.PromptResolution{}, r.err
	}
	return newStaticTestPromptRegistry().Resolve(ctx, featureKey, language)
}

// staticTestPromptRegistry replaces the non-current targetjob.StaticPromptRegistry
// for cmd/api scenario tests. It mirrors the F3 RegistryAdapter shape with a
// fixed target.import.parse resolution so HTTP scenarios can assert the
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

type scenarioFetcher struct{}

func (f *scenarioFetcher) Fetch(context.Context, string) (urlfetch.FetchResult, error) {
	return urlfetch.FetchResult{SanitizedURL: "https://jobs.example.com/role/1", Body: "Scenario JD body", FetchedAt: fixedScenarioNow()}, nil
}
