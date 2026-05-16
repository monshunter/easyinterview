package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient"
	apireports "github.com/monshunter/easyinterview/backend/internal/api/reports"
	"github.com/monshunter/easyinterview/backend/internal/auth"
	"github.com/monshunter/easyinterview/backend/internal/platform/config"
	"github.com/monshunter/easyinterview/backend/internal/platform/featureflag"
	resumehandler "github.com/monshunter/easyinterview/backend/internal/resume/handler"
	resumejobs "github.com/monshunter/easyinterview/backend/internal/resume/jobs"
	"github.com/monshunter/easyinterview/backend/internal/shared/jobs"
	"github.com/monshunter/easyinterview/backend/internal/targetjob"
	"github.com/monshunter/easyinterview/backend/internal/upload/objectstore"
	"github.com/monshunter/easyinterview/backend/internal/upload/store"
)

func TestBuildFlagsClientLoadsPostHogPublicAllowlist(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"featureFlags":{"practice_hint_enabled":true,"ai_fallback_model_enabled":true}}`))
	}))
	defer server.Close()

	dir := t.TempDir()
	flagsPath := filepath.Join(dir, "feature-flags.yaml")
	writeAPIFile(t, flagsPath, `
flags:
  practice_hint_enabled:
    enabled: false
    public: true
  ai_fallback_model_enabled:
    enabled: true
    public: false
`)
	writeAPIFile(t, filepath.Join(dir, "config.yaml"), `
runtime:
  appVersion: "1.2.3"
  defaultUiLanguage: zh-CN
featureFlag:
  source: posthog
  filePath: "`+flagsPath+`"
  posthogHost: "`+server.URL+`"
  posthogSelfHosted: true
  posthogProjectApiKey: "ph-key"
`)
	loader, err := config.Load(config.Options{ConfigDir: dir})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	client, err := buildFlagsClient(loader, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatalf("buildFlagsClient: %v", err)
	}
	rc := config.BuildRuntimeConfig(context.Background(), config.RuntimeConfigInput{
		Loader:      loader,
		Flags:       client,
		FlagContext: featureflag.FlagContext{AnonymousDistinctID: "anon-1", AppEnv: "prod"},
	})
	if _, ok := rc.FeatureFlags["practice_hint_enabled"]; !ok {
		t.Fatalf("public flag missing from runtime-config: %+v", rc.FeatureFlags)
	}
	if _, ok := rc.FeatureFlags["ai_fallback_model_enabled"]; ok {
		t.Fatalf("operator-only flag leaked: %+v", rc.FeatureFlags)
	}
}

func TestBuildAPIHandlerMountsAuthRoutesAndSessionAwareRuntimeConfig(t *testing.T) {
	dir := t.TempDir()
	writeAPIFile(t, filepath.Join(dir, "config.yaml"), `
runtime:
  appVersion: "1.2.3"
  defaultUiLanguage: zh-CN
featureFlag:
  posthogPublicKey: "ph-public"
`)
	loader, err := config.Load(config.Options{ConfigDir: dir})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	sink := auth.NewDevMailSink(auth.DevMailSinkOptions{VerifyBaseURL: "/api/v1/auth/email/verify"})
	store := &apiAuthStore{
		session: auth.SessionRecord{
			ID:        "session-1",
			UserID:    "user-1",
			Status:    auth.SessionStatusActive,
			ExpiresAt: time.Now().Add(auth.SessionTTL),
		},
		user: auth.UserContext{
			ID:             "user-1",
			Email:          "candidate@example.com",
			AnalyticsOptIn: true,
		},
	}
	service := auth.NewPasswordlessService(auth.PasswordlessServiceOptions{
		Store:                 store,
		Dispatcher:            auth.NewImmediateMailDispatcher(sink),
		DeliverySecrets:       sink,
		TokenGenerator:        apiFixedTokenGenerator("magic-token"),
		SessionTokenGenerator: apiFixedTokenGenerator("session-token"),
		ChallengePepper:       "pepper",
		SessionCookieSecret:   "session-secret",
		Now:                   func() time.Time { return time.Date(2026, 5, 6, 20, 0, 0, 0, time.UTC) },
		NewID:                 apiFixedIDs("challenge-1"),
	})
	handler := buildAPIHandler(loader, apiRuntimeFlags{}, service, nil)

	start := httptest.NewRecorder()
	handler.ServeHTTP(start, httptest.NewRequest(http.MethodPost, "/api/v1/auth/email/start", bytes.NewBufferString(`{"email":"Candidate@Example.COM"}`)))
	if start.Code != http.StatusAccepted {
		t.Fatalf("start route status = %d body=%s", start.Code, start.Body.String())
	}

	meMissing := httptest.NewRecorder()
	handler.ServeHTTP(meMissing, httptest.NewRequest(http.MethodGet, "/api/v1/me", nil))
	if meMissing.Code != http.StatusUnauthorized {
		t.Fatalf("/me without cookie status = %d body=%s", meMissing.Code, meMissing.Body.String())
	}

	runtimeReq := httptest.NewRequest(http.MethodGet, "/api/v1/runtime-config", nil)
	runtimeReq.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
	runtimeRec := httptest.NewRecorder()
	handler.ServeHTTP(runtimeRec, runtimeReq)
	if runtimeRec.Code != http.StatusOK {
		t.Fatalf("runtime-config status = %d body=%s", runtimeRec.Code, runtimeRec.Body.String())
	}
	if !strings.Contains(runtimeRec.Body.String(), `"postHogPublicKey":"ph-public"`) {
		t.Fatalf("runtime-config did not use session resolver: %s", runtimeRec.Body.String())
	}

	logout := httptest.NewRecorder()
	handler.ServeHTTP(logout, httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil))
	if logout.Code != http.StatusNoContent {
		t.Fatalf("logout route status = %d body=%s", logout.Code, logout.Body.String())
	}
}

func TestBuildAPIHandlerMountsTargetJobRoutesBehindSessionMiddleware(t *testing.T) {
	dir := t.TempDir()
	writeAPIFile(t, filepath.Join(dir, "config.yaml"), `
runtime:
  appVersion: "1.2.3"
  defaultUiLanguage: zh-CN
`)
	loader, err := config.Load(config.Options{ConfigDir: dir})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	service := auth.NewPasswordlessService(auth.PasswordlessServiceOptions{
		Store:               &apiAuthStore{},
		SessionCookieSecret: "session-secret",
	})
	handler := buildAPIHandler(loader, apiRuntimeFlags{}, service, nil)

	cases := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/api/v1/targets"},
		{http.MethodPost, "/api/v1/targets/import"},
		{http.MethodGet, "/api/v1/targets/018f2a40-0000-7000-9000-0000000000a1"},
		{http.MethodPatch, "/api/v1/targets/018f2a40-0000-7000-9000-0000000000a1"},
	}
	for _, tc := range cases {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, httptest.NewRequest(tc.method, tc.path, nil))
			if rec.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d body=%s; route is not mounted behind auth middleware", rec.Code, rec.Body.String())
			}
			if !strings.Contains(rec.Body.String(), `"code":"AUTH_UNAUTHORIZED"`) {
				t.Fatalf("expected auth middleware envelope, got %s", rec.Body.String())
			}
		})
	}
}

func TestBuildAPIHandlerMountsUploadPresignBehindSessionMiddleware(t *testing.T) {
	dir := t.TempDir()
	writeAPIFile(t, filepath.Join(dir, "config.yaml"), `
runtime:
  appVersion: "1.2.3"
  defaultUiLanguage: zh-CN
objectStorage:
  provider: filesystem
upload:
  presignTTLSeconds: 600
  maxBytes:
    resume: 10485760
    targetJobAttachment: 10485760
    privacyExport: 5242880
`)
	loader, err := config.Load(config.Options{ConfigDir: dir})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	service := auth.NewPasswordlessService(auth.PasswordlessServiceOptions{
		Store:               &apiAuthStore{},
		SessionCookieSecret: "session-secret",
	})
	handler := buildAPIHandler(loader, apiRuntimeFlags{}, service, nil)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/uploads/presign", strings.NewReader(`{"purpose":"resume","fileName":"resume.pdf","contentType":"application/pdf","byteSize":128}`))
	req.Header.Set("Idempotency-Key", "idem-1")
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d body=%s; upload route is not mounted behind auth middleware", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"code":"AUTH_UNAUTHORIZED"`) {
		t.Fatalf("expected auth middleware envelope, got %s", rec.Body.String())
	}
}

func TestBuildAPIHandlerMountsResumeRoutesBehindSessionMiddleware(t *testing.T) {
	dir := t.TempDir()
	writeAPIFile(t, filepath.Join(dir, "config.yaml"), `
runtime:
  appVersion: "1.2.3"
  defaultUiLanguage: zh-CN
`)
	loader, err := config.Load(config.Options{ConfigDir: dir})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	service := auth.NewPasswordlessService(auth.PasswordlessServiceOptions{
		Store:               &apiAuthStore{},
		SessionCookieSecret: "session-secret",
	})
	handler := buildAPIHandlerWithUploadAndHandlers(
		loader,
		apiRuntimeFlags{},
		service,
		targetjob.NewHandler(),
		practiceRoutes{},
		uploadRoutes{},
		resumeRoutes{Handler: resumehandler.New(resumehandler.Options{})},
	)

	cases := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/api/v1/resumes"},
		{http.MethodPost, "/api/v1/resumes"},
		{http.MethodGet, "/api/v1/resumes/018f2a40-0000-7000-9000-0000000000a1"},
	}
	for _, tc := range cases {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(`{"sourceType":"paste","rawText":"resume","title":"Resume","language":"en"}`))
			req.Header.Set("Idempotency-Key", "idem-1")
			handler.ServeHTTP(rec, req)
			if rec.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d body=%s; resume route is not mounted behind auth middleware", rec.Code, rec.Body.String())
			}
			if !strings.Contains(rec.Body.String(), `"code":"AUTH_UNAUTHORIZED"`) {
				t.Fatalf("expected auth middleware envelope, got %s", rec.Body.String())
			}
		})
	}
}

func TestBuildAPIHandlerMountsReportRoutesBehindSessionMiddleware(t *testing.T) {
	dir := t.TempDir()
	writeAPIFile(t, filepath.Join(dir, "config.yaml"), `
runtime:
  appVersion: "1.2.3"
  defaultUiLanguage: zh-CN
`)
	loader, err := config.Load(config.Options{ConfigDir: dir})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	service := auth.NewPasswordlessService(auth.PasswordlessServiceOptions{
		Store:               &apiAuthStore{},
		SessionCookieSecret: "session-secret",
	})
	handler := buildAPIHandlerWithUploadReportAndHandlers(
		loader,
		apiRuntimeFlags{},
		service,
		targetjob.NewHandler(),
		practiceRoutes{},
		uploadRoutes{},
		resumeRoutes{},
		reportRoutes{Handler: apireports.NewHandler(apireports.HandlerOptions{})},
	)

	cases := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/api/v1/reports/018f2a40-0000-7000-9000-0000000000a1"},
		{http.MethodGet, "/api/v1/targets/018f2a40-0000-7000-9000-0000000000a2/reports"},
	}
	for _, tc := range cases {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(tc.method, tc.path, nil)
			handler.ServeHTTP(rec, req)
			if rec.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d body=%s; report route is not mounted behind auth middleware", rec.Code, rec.Body.String())
			}
			if !strings.Contains(rec.Body.String(), `"code":"AUTH_UNAUTHORIZED"`) {
				t.Fatalf("expected auth middleware envelope, got %s", rec.Body.String())
			}
		})
	}
}

func TestBuildUploadRoutesAlignsIdempotencyTTLWithPresignTTL(t *testing.T) {
	dir := t.TempDir()
	writeAPIFile(t, filepath.Join(dir, "config.yaml"), `
runtime:
  appVersion: "1.2.3"
  defaultUiLanguage: zh-CN
objectStorage:
  provider: filesystem
upload:
  presignTTLSeconds: 600
  maxBytes:
    resume: 10485760
    targetJobAttachment: 10485760
    privacyExport: 5242880
`)
	loader, err := config.Load(config.Options{ConfigDir: dir})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	routes, err := buildUploadRoutes(loader, nil)
	if err != nil {
		t.Fatalf("buildUploadRoutes: %v", err)
	}
	if got := routes.Idempotency.TTL(); got != 10*time.Minute {
		t.Fatalf("upload idempotency TTL = %s, want presign TTL %s", got, 10*time.Minute)
	}
}

func TestBuildResumeRuntimeWiresRoutesDrainerAndDeterministicAI(t *testing.T) {
	dir := t.TempDir()
	promptsDir, rubricsDir := repoConfigPromptsRubrics(t)
	writeAPIFile(t, filepath.Join(dir, "config.yaml"), `
runtime:
  appVersion: "1.2.3"
  defaultUiLanguage: zh-CN
ai:
  promptsDir: "`+promptsDir+`"
  rubricsDir: "`+rubricsDir+`"
auth:
  challengeTokenPepper: "pepper"
`)
	loader, err := config.Load(config.Options{AppEnv: "test", ConfigDir: dir})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	runtime, err := buildResumeRuntime(
		loader,
		nil,
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		uploadRoutes{Objects: objectstore.NewFilesystemStore(t.TempDir())},
		&apiNoopAIClient{},
	)
	if err != nil {
		t.Fatalf("buildResumeRuntime: %v", err)
	}
	if runtime.Handler == nil || runtime.Idempotency == nil || runtime.Drainer == nil || runtime.ParseAI == nil {
		t.Fatalf("runtime missing handler/idempotency/drainer/AI wiring: %+v", runtime)
	}
	if !runtime.Drainer.Handles(string(jobs.JobTypeResumeParse)) {
		t.Fatalf("runtime drainer does not handle %s", jobs.JobTypeResumeParse)
	}
	resp, _, err := runtime.ParseAI.Complete(context.Background(), "resume.parse.default", aiclient.CompletePayload{
		Messages: []aiclient.Message{{Role: "user", Content: "Resume text"}},
		Metadata: aiclient.CallMetadata{FeatureKey: resumejobs.FeatureKeyResumeParse, Language: "en"},
	})
	if err != nil {
		t.Fatalf("test runtime resume parse fixture: %v", err)
	}
	var parsed map[string]any
	if err := json.Unmarshal([]byte(resp.Content), &parsed); err != nil {
		t.Fatalf("test runtime resume parse fixture did not return JSON: %v; content=%s", err, resp.Content)
	}
	if _, ok := parsed["basics"]; !ok {
		t.Fatalf("test runtime resume parse fixture missing basics: %+v", parsed)
	}
}

func TestBuildReportRuntimeWiresRoutesRunnerReaperAndAI(t *testing.T) {
	dir := t.TempDir()
	promptsDir, rubricsDir := repoConfigPromptsRubrics(t)
	writeAPIFile(t, filepath.Join(dir, "config.yaml"), `
runtime:
  appVersion: "1.2.3"
  defaultUiLanguage: zh-CN
ai:
  promptsDir: "`+promptsDir+`"
  rubricsDir: "`+rubricsDir+`"
auth:
  challengeTokenPepper: "pepper"
`)
	loader, err := config.Load(config.Options{AppEnv: "test", ConfigDir: dir})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	runtime, err := buildReportRuntime(loader, nil, slog.New(slog.NewTextHandler(io.Discard, nil)), &apiNoopAIClient{})
	if err != nil {
		t.Fatalf("buildReportRuntime: %v", err)
	}
	if runtime.Handler == nil || runtime.Runner == nil || runtime.Reaper == nil || runtime.Service == nil {
		t.Fatalf("runtime missing handler/runner/reaper/service wiring: %+v", runtime)
	}
	if runtime.Routes().Handler != runtime.Handler {
		t.Fatalf("Routes handler mismatch")
	}
}

func TestBuildReportRuntimeRejectsMissingAIClient(t *testing.T) {
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
	if _, err := buildReportRuntime(loader, nil, slog.New(slog.NewTextHandler(io.Discard, nil)), nil); err == nil {
		t.Fatal("buildReportRuntime returned nil error for missing AI client")
	}
}

func TestBuildDebriefRoutesWiresHandlerAndIdempotency(t *testing.T) {
	dir := t.TempDir()
	promptsDir, rubricsDir := repoConfigPromptsRubrics(t)
	writeAPIFile(t, filepath.Join(dir, "config.yaml"), `
runtime:
  appVersion: "1.2.3"
  defaultUiLanguage: zh-CN
ai:
  promptsDir: "`+promptsDir+`"
  rubricsDir: "`+rubricsDir+`"
auth:
  challengeTokenPepper: "pepper"
`)
	loader, err := config.Load(config.Options{AppEnv: "test", ConfigDir: dir})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	routes, err := buildDebriefRoutes(loader, nil, &apiNoopAIClient{})
	if err != nil {
		t.Fatalf("buildDebriefRoutes: %v", err)
	}
	if routes.Handler == nil || routes.Idempotency == nil {
		t.Fatalf("debrief routes missing handler/idempotency: %+v", routes)
	}
}

func TestBuildTargetJobRuntimeWiresDrainerAndAIClient(t *testing.T) {
	dir := t.TempDir()
	providersPath := filepath.Join(dir, "ai-providers.yaml")
	profilesPath := filepath.Join(dir, "ai-profiles.yaml")
	writeAPIFile(t, providersPath, `
providers:
  - name: unit-test-stub
    protocol: stub
    capabilities: [chat]
    version: 1.0.0
`)
	writeAPIFile(t, profilesPath, `
profiles:
  - name: target.import.default
    capability: chat
    status: active
    default:
      provider_ref: unit-test-stub
      model: stub-chat
    fallback: []
    timeout_ms: 1000
    max_tokens: 256
    rate_limit:
      rps: 1
      tpm: 1000
    route: target.import
    version: 1.0.0
`)
	promptsDir, rubricsDir := repoConfigPromptsRubrics(t)
	writeAPIFile(t, filepath.Join(dir, "config.yaml"), `
runtime:
  appVersion: "1.2.3"
  defaultUiLanguage: zh-CN
ai:
  providerRegistryPath: "`+providersPath+`"
  modelProfilePath: "`+profilesPath+`"
  promptsDir: "`+promptsDir+`"
  rubricsDir: "`+rubricsDir+`"
`)
	loader, err := config.Load(config.Options{AppEnv: "test", ConfigDir: dir})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	runtime, err := buildTargetJobRuntime(loader, nil, slog.New(slog.NewTextHandler(io.Discard, nil)), nil)
	if err != nil {
		t.Fatalf("buildTargetJobRuntime: %v", err)
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if err := runtime.Shutdown(shutdownCtx); err != nil {
			t.Fatalf("Shutdown: %v", err)
		}
	}()
	if runtime.Handler == nil || runtime.Drainer == nil || runtime.AI == nil || runtime.AI.Client == nil {
		t.Fatalf("runtime missing handler/drainer/AI wiring: %+v", runtime)
	}
	resp, _, err := runtime.ParseAI.Complete(context.Background(), "target.import.default", aiclient.CompletePayload{
		Messages: []aiclient.Message{{Role: "user", Content: "Backend Engineer JD"}},
		Metadata: aiclient.CallMetadata{
			FeatureKey: targetjob.FeatureKeyTargetImportParse,
			Language:   "en",
		},
	})
	if err != nil {
		t.Fatalf("test runtime target import parse fixture: %v", err)
	}
	var parsed struct {
		Requirements []struct {
			Kind  string `json:"kind"`
			Label string `json:"label"`
		} `json:"requirements"`
	}
	if err := json.Unmarshal([]byte(resp.Content), &parsed); err != nil {
		t.Fatalf("test runtime parse fixture did not return JSON: %v; content=%s", err, resp.Content)
	}
	if len(parsed.Requirements) == 0 || parsed.Requirements[0].Kind != string(targetjob.RequirementMustHave) {
		t.Fatalf("test runtime parse fixture returned invalid requirements: %+v", parsed.Requirements)
	}
}

func TestBuildTargetJobRuntimeRegistersPrivacyDeleteHandler(t *testing.T) {
	dir := t.TempDir()
	providersPath := filepath.Join(dir, "ai-providers.yaml")
	profilesPath := filepath.Join(dir, "ai-profiles.yaml")
	writeAPIFile(t, providersPath, `
providers:
  - name: unit-test-stub
    protocol: stub
    capabilities: [chat]
    version: 1.0.0
`)
	writeAPIFile(t, profilesPath, `
profiles:
  - name: target.import.default
    capability: chat
    status: active
    default:
      provider_ref: unit-test-stub
      model: stub-chat
    fallback: []
    timeout_ms: 1000
    max_tokens: 256
    rate_limit:
      rps: 1
      tpm: 1000
    route: target.import
    version: 1.0.0
`)
	promptsDir, rubricsDir := repoConfigPromptsRubrics(t)
	writeAPIFile(t, filepath.Join(dir, "config.yaml"), `
runtime:
  appVersion: "1.2.3"
  defaultUiLanguage: zh-CN
ai:
  providerRegistryPath: "`+providersPath+`"
  modelProfilePath: "`+profilesPath+`"
  promptsDir: "`+promptsDir+`"
  rubricsDir: "`+rubricsDir+`"
`)
	loader, err := config.Load(config.Options{AppEnv: "test", ConfigDir: dir})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	runtime, err := buildTargetJobRuntime(loader, nil, slog.New(slog.NewTextHandler(io.Discard, nil)), &apiUploadFileDeleter{})
	if err != nil {
		t.Fatalf("buildTargetJobRuntime: %v", err)
	}
	if !runtime.Drainer.Handles(string(jobs.JobTypePrivacyDelete)) {
		t.Fatalf("runtime drainer does not handle %s", jobs.JobTypePrivacyDelete)
	}
}

func TestDrainer_DebriefGenerateRegistered(t *testing.T) {
	dir := t.TempDir()
	providersPath := filepath.Join(dir, "ai-providers.yaml")
	profilesPath := filepath.Join(dir, "ai-profiles.yaml")
	writeAPIFile(t, providersPath, `
providers:
  - name: unit-test-stub
    protocol: stub
    capabilities: [chat]
    version: 1.0.0
`)
	writeAPIFile(t, profilesPath, `
profiles:
  - name: target.import.default
    capability: chat
    status: active
    default:
      provider_ref: unit-test-stub
      model: stub-chat
    fallback: []
    timeout_ms: 1000
    max_tokens: 256
    rate_limit:
      rps: 1
      tpm: 1000
    route: target.import
    version: 1.0.0
`)
	promptsDir, rubricsDir := repoConfigPromptsRubrics(t)
	writeAPIFile(t, filepath.Join(dir, "config.yaml"), `
runtime:
  appVersion: "1.2.3"
  defaultUiLanguage: zh-CN
ai:
  providerRegistryPath: "`+providersPath+`"
  modelProfilePath: "`+profilesPath+`"
  promptsDir: "`+promptsDir+`"
  rubricsDir: "`+rubricsDir+`"
`)
	loader, err := config.Load(config.Options{AppEnv: "test", ConfigDir: dir})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	runtime, err := buildTargetJobRuntime(loader, nil, slog.New(slog.NewTextHandler(io.Discard, nil)), nil)
	if err != nil {
		t.Fatalf("buildTargetJobRuntime: %v", err)
	}
	if !runtime.Drainer.Handles(string(jobs.JobTypeDebriefGenerate)) {
		t.Fatalf("runtime drainer does not handle %s", jobs.JobTypeDebriefGenerate)
	}
}

func TestBuildAuthServiceRejectsEmptyAuthSecrets(t *testing.T) {
	dir := t.TempDir()
	writeAPIFile(t, filepath.Join(dir, "config.yaml"), `
auth:
  challengeTokenPepper: ""
  sessionCookieSecret: ""
`)
	loader, err := config.Load(config.Options{ConfigDir: dir})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	service, dispatcher, err := buildAuthService(loader, nil)

	if err == nil {
		t.Fatal("expected empty auth secrets to fail")
	}
	if service != nil || dispatcher != nil {
		t.Fatalf("auth service should not be constructed on missing secrets: service=%#v dispatcher=%#v", service, dispatcher)
	}
	for _, want := range []string{"AUTH_CHALLENGE_TOKEN_PEPPER", "SESSION_COOKIE_SECRET"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("missing %s in error: %v", want, err)
		}
	}
}

func TestBuildAPIHandlerLogoutPropagatesSessionResolverErrors(t *testing.T) {
	dir := t.TempDir()
	writeAPIFile(t, filepath.Join(dir, "config.yaml"), `
runtime:
  appVersion: "1.2.3"
  defaultUiLanguage: zh-CN
`)
	loader, err := config.Load(config.Options{ConfigDir: dir})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	store := &apiAuthStore{lookupErr: errors.New("database unavailable for session-1")}
	service := auth.NewPasswordlessService(auth.PasswordlessServiceOptions{
		Store:               store,
		SessionCookieSecret: "session-secret",
		Now:                 func() time.Time { return time.Date(2026, 5, 6, 21, 0, 0, 0, time.UTC) },
	})
	handler := buildAPIHandler(loader, apiRuntimeFlags{}, service, nil)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("logout resolver error status = %d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"code":"VALIDATION_FAILED"`) {
		t.Fatalf("logout resolver error envelope = %s", rec.Body.String())
	}
	for _, forbidden := range []string{"raw-session-token", "session-1", "session-secret"} {
		if strings.Contains(rec.Body.String(), forbidden) {
			t.Fatalf("logout resolver error leaked %q: %s", forbidden, rec.Body.String())
		}
	}
}

func TestBuildAPIHandlerLogoutKeepsKnownSessionErrorsOptional(t *testing.T) {
	dir := t.TempDir()
	writeAPIFile(t, filepath.Join(dir, "config.yaml"), `
runtime:
  appVersion: "1.2.3"
  defaultUiLanguage: zh-CN
`)
	loader, err := config.Load(config.Options{ConfigDir: dir})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	for name, lookupErr := range map[string]error{
		"invalid": auth.ErrSessionInvalid,
		"expired": auth.ErrSessionExpired,
		"revoked": auth.ErrSessionRevoked,
	} {
		t.Run(name, func(t *testing.T) {
			store := &apiAuthStore{lookupErr: lookupErr}
			service := auth.NewPasswordlessService(auth.PasswordlessServiceOptions{
				Store:               store,
				SessionCookieSecret: "session-secret",
				Now:                 func() time.Time { return time.Date(2026, 5, 6, 21, 0, 0, 0, time.UTC) },
			})
			handler := buildAPIHandler(loader, apiRuntimeFlags{}, service, nil)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
			req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusNoContent {
				t.Fatalf("logout %s status = %d body=%s", name, rec.Code, rec.Body.String())
			}
			cookies := rec.Result().Cookies()
			if len(cookies) != 1 || cookies[0].Name != auth.SessionCookieName || cookies[0].MaxAge >= 0 {
				t.Fatalf("logout %s clear cookie = %#v", name, cookies)
			}
		})
	}
}

func writeAPIFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

type apiRuntimeFlags struct{}

func (apiRuntimeFlags) IsEnabled(string, featureflag.FlagContext) bool { return false }

func (apiRuntimeFlags) Variant(string, featureflag.FlagContext) string { return "" }

func (apiRuntimeFlags) Snapshot(featureflag.FlagContext) map[string]featureflag.FlagDecision {
	return map[string]featureflag.FlagDecision{}
}

type apiFixedTokenGenerator string

func (g apiFixedTokenGenerator) GenerateToken() (string, error) { return string(g), nil }

func apiFixedIDs(ids ...string) func() string {
	index := 0
	return func() string {
		if index >= len(ids) {
			return ids[len(ids)-1]
		}
		id := ids[index]
		index++
		return id
	}
}

type apiAuthStore struct {
	challenge auth.ChallengeRecord
	session   auth.SessionRecord
	user      auth.UserContext
	lookupErr error
}

type apiUploadFileDeleter struct{}

func (d *apiUploadFileDeleter) DeleteFileObjectsForUser(context.Context, string) ([]store.DeletedFileObject, error) {
	return nil, nil
}

type apiNoopAIClient struct{}

func (c *apiNoopAIClient) Complete(context.Context, string, aiclient.CompletePayload) (aiclient.CompleteResponse, aiclient.AICallMeta, error) {
	return aiclient.CompleteResponse{}, aiclient.AICallMeta{}, errors.New("unexpected Complete call")
}

func (c *apiNoopAIClient) Transcribe(context.Context, string, aiclient.TranscriptionInput) (aiclient.TranscriptionResponse, aiclient.AICallMeta, error) {
	return aiclient.TranscriptionResponse{}, aiclient.AICallMeta{}, errors.New("unexpected Transcribe call")
}

func (c *apiNoopAIClient) Stream(context.Context, string, aiclient.CompletePayload) (<-chan aiclient.AIStreamEvent, error) {
	return nil, errors.New("unexpected Stream call")
}

func (c *apiNoopAIClient) Synthesize(context.Context, string, aiclient.SynthesisInput) (aiclient.SynthesisResponse, aiclient.AICallMeta, error) {
	return aiclient.SynthesisResponse{}, aiclient.AICallMeta{}, errors.New("unexpected Synthesize call")
}

func (s *apiAuthStore) CountRecentChallenges(context.Context, string, string, time.Time) (int, error) {
	return 0, nil
}

func (s *apiAuthStore) CreateChallenge(_ context.Context, rec auth.ChallengeRecord) error {
	s.challenge = rec
	return nil
}

func (s *apiAuthStore) ConsumeChallenge(context.Context, string, time.Time) (auth.ChallengeRecord, error) {
	return s.challenge, nil
}

func (s *apiAuthStore) FindOrCreateUserByEmail(context.Context, string, string, time.Time) (auth.UserContext, error) {
	return s.user, nil
}

func (s *apiAuthStore) CreateSession(_ context.Context, rec auth.SessionRecord) error {
	s.session = rec
	return nil
}

func (s *apiAuthStore) GetSessionByHash(context.Context, string, time.Time) (auth.SessionRecord, error) {
	if s.lookupErr != nil {
		return auth.SessionRecord{}, s.lookupErr
	}
	return s.session, nil
}

func (s *apiAuthStore) GetUserContext(context.Context, string) (auth.UserContext, error) {
	return s.user, nil
}

func (s *apiAuthStore) TouchSession(_ context.Context, sessionID string, now time.Time, expiresAt time.Time) error {
	s.session.ID = sessionID
	s.session.UpdatedAt = now
	s.session.ExpiresAt = expiresAt
	return nil
}

func (s *apiAuthStore) RevokeSession(context.Context, string, time.Time) error { return nil }

func (s *apiAuthStore) CreatePrivacyDeleteHandoff(context.Context, string, string, string, string, time.Time) (auth.PrivacyDeleteHandoff, error) {
	panic("not used")
}

// repoConfigPromptsRubrics walks upward from the test working directory
// until it finds the backend go.mod, then returns the in-repo
// config/prompts and config/rubrics absolute paths so cmd/api tests can
// wire a real F3 registry without copying the truth source into a tmpdir.
func repoConfigPromptsRubrics(t *testing.T) (string, string) {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	dir := wd
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Skipf("could not locate backend go.mod from %s", wd)
			return "", ""
		}
		dir = parent
	}
	repoRoot := filepath.Dir(dir)
	return filepath.Join(repoRoot, "config", "prompts"),
		filepath.Join(repoRoot, "config", "rubrics")
}
