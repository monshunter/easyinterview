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

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient"
	api "github.com/monshunter/easyinterview/backend/internal/api/generated"
	apijobs "github.com/monshunter/easyinterview/backend/internal/api/jobs"
	apireports "github.com/monshunter/easyinterview/backend/internal/api/reports"
	"github.com/monshunter/easyinterview/backend/internal/auth"
	"github.com/monshunter/easyinterview/backend/internal/middleware/idempotency"
	"github.com/monshunter/easyinterview/backend/internal/platform/config"
	"github.com/monshunter/easyinterview/backend/internal/platform/featureflag"
	domainresume "github.com/monshunter/easyinterview/backend/internal/resume"
	resumehandler "github.com/monshunter/easyinterview/backend/internal/resume/handler"
	resumejobs "github.com/monshunter/easyinterview/backend/internal/resume/jobs"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
	"github.com/monshunter/easyinterview/backend/internal/shared/jobs"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
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

func TestLocalDevCORSAllowsFrontendRealModeOrigins(t *testing.T) {
	called := false
	handler := withLocalDevCORS("dev", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	preflight := httptest.NewRequest(http.MethodOptions, "/api/v1/runtime-config", nil)
	preflight.Header.Set("Origin", "http://127.0.0.1:4174")
	preflight.Header.Set("Access-Control-Request-Headers", "Content-Type,Idempotency-Key")
	preflightRec := httptest.NewRecorder()
	handler.ServeHTTP(preflightRec, preflight)
	if preflightRec.Code != http.StatusNoContent {
		t.Fatalf("preflight status = %d body=%s", preflightRec.Code, preflightRec.Body.String())
	}
	if called {
		t.Fatalf("preflight should not call next handler")
	}
	if got := preflightRec.Header().Get("Access-Control-Allow-Origin"); got != "http://127.0.0.1:4174" {
		t.Fatalf("allow origin = %q", got)
	}
	if got := preflightRec.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
		t.Fatalf("allow credentials = %q", got)
	}
	if got := preflightRec.Header().Get("Access-Control-Allow-Headers"); strings.Contains(got, "Prefer") {
		t.Fatalf("real-mode CORS should not allow fixture Prefer header: %q", got)
	}

	get := httptest.NewRequest(http.MethodGet, "/api/v1/runtime-config", nil)
	get.Header.Set("Origin", "http://localhost:5173")
	getRec := httptest.NewRecorder()
	handler.ServeHTTP(getRec, get)
	if !called {
		t.Fatalf("GET should call next handler")
	}
	if got := getRec.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:5173" {
		t.Fatalf("GET allow origin = %q", got)
	}
}

func TestLocalDevCORSRejectsUnknownPreflightAndStaysDisabledOutsideDev(t *testing.T) {
	handler := withLocalDevCORS("dev", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("unknown preflight should not reach next handler")
	}))
	preflight := httptest.NewRequest(http.MethodOptions, "/api/v1/me", nil)
	preflight.Header.Set("Origin", "https://example.invalid")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, preflight)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("unknown preflight status = %d body=%s", rec.Code, rec.Body.String())
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("unknown origin should not be echoed, got %q", got)
	}

	prodCalled := false
	prod := withLocalDevCORS("prod", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		prodCalled = true
		w.WriteHeader(http.StatusOK)
	}))
	get := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	get.Header.Set("Origin", "http://127.0.0.1:4174")
	prodRec := httptest.NewRecorder()
	prod.ServeHTTP(prodRec, get)
	if !prodCalled {
		t.Fatalf("non-dev handler should pass through")
	}
	if got := prodRec.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("non-dev CORS header = %q", got)
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
		{http.MethodPost, "/api/v1/resume-versions"},
		{http.MethodPost, "/api/v1/resume/tailor"},
		{http.MethodGet, "/api/v1/resume/tailor-runs/018f2a40-0000-7000-9000-0000000000c1"},
		{http.MethodGet, "/api/v1/resumes/018f2a40-0000-7000-9000-0000000000a1"},
		{http.MethodGet, "/api/v1/resumes/018f2a40-0000-7000-9000-0000000000a1/versions"},
		{http.MethodGet, "/api/v1/resume-versions/018f2a40-0000-7000-9000-0000000000b1"},
		{http.MethodPatch, "/api/v1/resume-versions/018f2a40-0000-7000-9000-0000000000b1"},
		{http.MethodPost, "/api/v1/resumes/018f2a40-0000-7000-9000-0000000000a1/structured-master"},
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

func TestResumeConfirmStructuredMasterHTTPScenario(t *testing.T) {
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
	sessionTime := time.Date(2026, 5, 17, 17, 30, 0, 0, time.UTC)
	service := auth.NewPasswordlessService(auth.PasswordlessServiceOptions{
		Store: &apiAuthStore{
			session: auth.SessionRecord{
				ID:        "session-1",
				UserID:    "user-1",
				Status:    auth.SessionStatusActive,
				ExpiresAt: sessionTime.Add(time.Hour),
			},
			user: auth.UserContext{ID: "user-1", Email: "candidate@example.com"},
		},
		SessionCookieSecret: "session-secret",
		Now:                 func() time.Time { return sessionTime },
	})
	version := api.ResumeVersion{
		Id:            "version-1",
		ResumeAssetId: "asset-1",
		VersionType:   sharedtypes.ResumeVersionTypeStructuredMaster,
		DisplayName:   "Structured master",
		StructuredProfile: map[string]any{"headline": "Senior engineer", "provenance": map[string]any{
			"promptVersion": "resume_profile.v1", "rubricVersion": "not_applicable", "modelId": "model-1", "language": "en", "featureFlag": "none", "dataSourceVersion": "asset.v1",
		}},
		Provenance: api.GenerationProvenance{
			PromptVersion: "resume_profile.v1", RubricVersion: "not_applicable", ModelId: "model-1", Language: "en", FeatureFlag: "none", DataSourceVersion: "asset.v1",
		},
		Suggestions: []any{},
		CreatedAt:   sessionTime.Format(time.RFC3339),
		UpdatedAt:   sessionTime.Format(time.RFC3339),
	}
	idemStore := &apiIdempotencyStore{}
	resumeSvc := &apiConfirmResumeService{out: version}
	handler := buildAPIHandlerWithUploadAndHandlers(
		loader,
		apiRuntimeFlags{},
		service,
		targetjob.NewHandler(),
		practiceRoutes{},
		uploadRoutes{},
		resumeRoutes{
			Handler: resumehandler.New(resumehandler.Options{Service: resumeSvc, Session: currentUserFromContext}),
			Idempotency: idempotency.New(idempotency.MiddlewareOptions{
				Store: idemStore,
				Now:   func() time.Time { return sessionTime },
				NewID: func() string { return "idem-rec-1" },
			}),
		},
	)

	t.Run("missing session", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/resumes/asset-1/structured-master", strings.NewReader(validAPIConfirmBody()))
		req.Header.Set(idempotency.HeaderName, "idem-confirm")

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
		}
		if !strings.Contains(rec.Body.String(), `"code":"AUTH_UNAUTHORIZED"`) {
			t.Fatalf("expected auth envelope, got %s", rec.Body.String())
		}
	})

	t.Run("missing idempotency key", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := authenticatedAPIRequest(http.MethodPost, "/api/v1/resumes/asset-1/structured-master", validAPIConfirmBody(), "")

		handler.ServeHTTP(rec, req)

		assertAPIStatusCode(t, rec, http.StatusUnprocessableEntity, sharederrors.CodeValidationFailed)
	})

	t.Run("happy path persists idempotency resource", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := authenticatedAPIRequest(http.MethodPost, "/api/v1/resumes/asset-1/structured-master", validAPIConfirmBody(), "idem-confirm")

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
		}
		if resumeSvc.calls != 1 || resumeSvc.in.UserID != "user-1" || resumeSvc.in.ResumeAssetID != "asset-1" {
			t.Fatalf("service calls=%d input=%+v", resumeSvc.calls, resumeSvc.in)
		}
		if idemStore.completeIn.Operation != "confirmResumeStructuredMaster" || idemStore.completeIn.ResourceType != "resume_version" || idemStore.completeIn.ResourceID != "version-1" {
			t.Fatalf("idempotency completion = %+v", idemStore.completeIn)
		}
		if rec.Header().Get("X-Idempotency-Resource-Type") != "" || rec.Header().Get("X-Idempotency-Resource-ID") != "" {
			t.Fatalf("internal idempotency headers leaked: %v", rec.Header())
		}
	})

	t.Run("idempotency replay bypasses service", func(t *testing.T) {
		replayStore := &apiIdempotencyStore{reservation: idempotency.Reservation{
			State:          idempotency.StateReplay,
			RecordID:       "idem-rec-replay",
			ResponseStatus: http.StatusCreated,
			ResponseBody:   []byte(`{"id":"version-replay","resumeAssetId":"asset-1","versionType":"structured_master","displayName":"Structured master","structuredProfile":{},"provenance":{"promptVersion":"p","rubricVersion":"r","modelId":"m","language":"en","featureFlag":"f","dataSourceVersion":"d"},"suggestions":[],"createdAt":"2026-05-17T17:30:00Z","updatedAt":"2026-05-17T17:30:00Z"}`),
		}}
		replaySvc := &apiConfirmResumeService{out: version}
		replayHandler := buildAPIHandlerWithUploadAndHandlers(
			loader,
			apiRuntimeFlags{},
			service,
			targetjob.NewHandler(),
			practiceRoutes{},
			uploadRoutes{},
			resumeRoutes{
				Handler:     resumehandler.New(resumehandler.Options{Service: replaySvc, Session: currentUserFromContext}),
				Idempotency: idempotency.New(idempotency.MiddlewareOptions{Store: replayStore, Now: func() time.Time { return sessionTime }}),
			},
		)
		rec := httptest.NewRecorder()
		req := authenticatedAPIRequest(http.MethodPost, "/api/v1/resumes/asset-1/structured-master", validAPIConfirmBody(), "idem-confirm")

		replayHandler.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated || rec.Header().Get(idempotency.ReplayHeader) != "true" {
			t.Fatalf("replay status=%d header=%q body=%s", rec.Code, rec.Header().Get(idempotency.ReplayHeader), rec.Body.String())
		}
		if replaySvc.calls != 0 {
			t.Fatalf("service calls on replay = %d, want 0", replaySvc.calls)
		}
	})

	t.Run("concurrent unique conflict maps to 409", func(t *testing.T) {
		conflictSvc := &apiConfirmResumeService{err: domainresume.ErrStructuredMasterAlreadyExists}
		conflictHandler := buildAPIHandlerWithUploadAndHandlers(
			loader,
			apiRuntimeFlags{},
			service,
			targetjob.NewHandler(),
			practiceRoutes{},
			uploadRoutes{},
			resumeRoutes{
				Handler:     resumehandler.New(resumehandler.Options{Service: conflictSvc, Session: currentUserFromContext}),
				Idempotency: idempotency.New(idempotency.MiddlewareOptions{Store: &apiIdempotencyStore{}, Now: func() time.Time { return sessionTime }}),
			},
		)
		rec := httptest.NewRecorder()
		req := authenticatedAPIRequest(http.MethodPost, "/api/v1/resumes/asset-1/structured-master", validAPIConfirmBody(), "idem-conflict")

		conflictHandler.ServeHTTP(rec, req)

		assertAPIStatusCode(t, rec, http.StatusConflict, sharederrors.CodeResumeStructuredMasterAlreadyExists)
	})
}

func TestResumeUpdateVersionHTTPScenario(t *testing.T) {
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
	sessionTime := time.Date(2026, 5, 17, 19, 30, 0, 0, time.UTC)
	service := auth.NewPasswordlessService(auth.PasswordlessServiceOptions{
		Store: &apiAuthStore{
			session: auth.SessionRecord{
				ID:        "session-1",
				UserID:    "user-1",
				Status:    auth.SessionStatusActive,
				ExpiresAt: sessionTime.Add(time.Hour),
			},
			user: auth.UserContext{ID: "user-1", Email: "candidate@example.com"},
		},
		SessionCookieSecret: "session-secret",
		Now:                 func() time.Time { return sessionTime },
	})
	version := apiVersion("version-1", "asset-1", sessionTime)
	idemStore := &apiIdempotencyStore{}
	resumeSvc := &apiUpdateVersionService{out: version}
	handler := buildAPIHandlerWithUploadAndHandlers(
		loader,
		apiRuntimeFlags{},
		service,
		targetjob.NewHandler(),
		practiceRoutes{},
		uploadRoutes{},
		resumeRoutes{
			Handler: resumehandler.New(resumehandler.Options{Service: resumeSvc, Session: currentUserFromContext}),
			Idempotency: idempotency.New(idempotency.MiddlewareOptions{
				Store: idemStore,
				Now:   func() time.Time { return sessionTime },
				NewID: func() string { return "idem-rec-update" },
			}),
		},
	)

	t.Run("missing session", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/resume-versions/version-1", strings.NewReader(validAPIUpdateBody()))
		req.Header.Set(idempotency.HeaderName, "idem-update")

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
		}
		if !strings.Contains(rec.Body.String(), `"code":"AUTH_UNAUTHORIZED"`) {
			t.Fatalf("expected auth envelope, got %s", rec.Body.String())
		}
	})

	t.Run("missing idempotency key", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := authenticatedAPIRequest(http.MethodPatch, "/api/v1/resume-versions/version-1", validAPIUpdateBody(), "")

		handler.ServeHTTP(rec, req)

		assertAPIStatusCode(t, rec, http.StatusUnprocessableEntity, sharederrors.CodeValidationFailed)
	})

	t.Run("happy path persists idempotency resource", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := authenticatedAPIRequest(http.MethodPatch, "/api/v1/resume-versions/version-1", validAPIUpdateBody(), "idem-update")

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
		}
		if resumeSvc.calls != 1 || resumeSvc.in.UserID != "user-1" || resumeSvc.in.VersionID != "version-1" {
			t.Fatalf("service calls=%d input=%+v", resumeSvc.calls, resumeSvc.in)
		}
		if resumeSvc.in.DisplayName == nil || *resumeSvc.in.DisplayName != "Updated version" {
			t.Fatalf("displayName = %#v", resumeSvc.in.DisplayName)
		}
		if !resumeSvc.in.FocusAngleSet || resumeSvc.in.FocusAngle != nil {
			t.Fatalf("focusAngle patch = set:%v value:%#v", resumeSvc.in.FocusAngleSet, resumeSvc.in.FocusAngle)
		}
		if idemStore.reserveIn.Operation != "updateResumeVersion" || idemStore.completeIn.Operation != "updateResumeVersion" {
			t.Fatalf("idempotency operation reserve=%+v complete=%+v", idemStore.reserveIn, idemStore.completeIn)
		}
		if idemStore.completeIn.ResourceType != "resume_version" || idemStore.completeIn.ResourceID != "version-1" {
			t.Fatalf("idempotency completion = %+v", idemStore.completeIn)
		}
		if rec.Header().Get("X-Idempotency-Resource-Type") != "" || rec.Header().Get("X-Idempotency-Resource-ID") != "" {
			t.Fatalf("internal idempotency headers leaked: %v", rec.Header())
		}
	})

	t.Run("idempotency replay bypasses service", func(t *testing.T) {
		replayStore := &apiIdempotencyStore{reservation: idempotency.Reservation{
			State:          idempotency.StateReplay,
			RecordID:       "idem-rec-replay",
			ResponseStatus: http.StatusOK,
			ResponseBody:   []byte(`{"id":"version-replay","resumeAssetId":"asset-1","versionType":"structured_master","displayName":"Updated version","structuredProfile":{},"provenance":{"promptVersion":"p","rubricVersion":"r","modelId":"m","language":"en","featureFlag":"f","dataSourceVersion":"d"},"suggestions":[],"createdAt":"2026-05-17T19:30:00Z","updatedAt":"2026-05-17T19:30:00Z"}`),
		}}
		replaySvc := &apiUpdateVersionService{out: version}
		replayHandler := buildAPIHandlerWithUploadAndHandlers(
			loader,
			apiRuntimeFlags{},
			service,
			targetjob.NewHandler(),
			practiceRoutes{},
			uploadRoutes{},
			resumeRoutes{
				Handler:     resumehandler.New(resumehandler.Options{Service: replaySvc, Session: currentUserFromContext}),
				Idempotency: idempotency.New(idempotency.MiddlewareOptions{Store: replayStore, Now: func() time.Time { return sessionTime }}),
			},
		)
		rec := httptest.NewRecorder()
		req := authenticatedAPIRequest(http.MethodPatch, "/api/v1/resume-versions/version-1", validAPIUpdateBody(), "idem-update")

		replayHandler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK || rec.Header().Get(idempotency.ReplayHeader) != "true" {
			t.Fatalf("replay status=%d header=%q body=%s", rec.Code, rec.Header().Get(idempotency.ReplayHeader), rec.Body.String())
		}
		if replaySvc.calls != 0 {
			t.Fatalf("service calls on replay = %d, want 0", replaySvc.calls)
		}
	})

	t.Run("idempotency mismatch maps to 409", func(t *testing.T) {
		mismatchHandler := buildAPIHandlerWithUploadAndHandlers(
			loader,
			apiRuntimeFlags{},
			service,
			targetjob.NewHandler(),
			practiceRoutes{},
			uploadRoutes{},
			resumeRoutes{
				Handler:     resumehandler.New(resumehandler.Options{Service: &apiUpdateVersionService{out: version}, Session: currentUserFromContext}),
				Idempotency: idempotency.New(idempotency.MiddlewareOptions{Store: &apiIdempotencyStore{err: idempotency.ErrFingerprintMismatch}, Now: func() time.Time { return sessionTime }}),
			},
		)
		rec := httptest.NewRecorder()
		req := authenticatedAPIRequest(http.MethodPatch, "/api/v1/resume-versions/version-1", validAPIUpdateBody(), "idem-update")

		mismatchHandler.ServeHTTP(rec, req)

		assertAPIStatusCode(t, rec, http.StatusConflict, sharederrors.CodeIdempotencyKeyMismatch)
	})

	t.Run("server-owned field rejected", func(t *testing.T) {
		rejectSvc := &apiUpdateVersionService{out: version}
		rejectHandler := buildAPIHandlerWithUploadAndHandlers(
			loader,
			apiRuntimeFlags{},
			service,
			targetjob.NewHandler(),
			practiceRoutes{},
			uploadRoutes{},
			resumeRoutes{
				Handler:     resumehandler.New(resumehandler.Options{Service: rejectSvc, Session: currentUserFromContext}),
				Idempotency: idempotency.New(idempotency.MiddlewareOptions{Store: &apiIdempotencyStore{}, Now: func() time.Time { return sessionTime }}),
			},
		)
		rec := httptest.NewRecorder()
		req := authenticatedAPIRequest(http.MethodPatch, "/api/v1/resume-versions/version-1", `{"versionType":"targeted"}`, "idem-update")

		rejectHandler.ServeHTTP(rec, req)

		assertAPIStatusCode(t, rec, http.StatusUnprocessableEntity, sharederrors.CodeValidationFailed)
		if rejectSvc.calls != 0 {
			t.Fatalf("service calls = %d, want 0", rejectSvc.calls)
		}
	})

	t.Run("not found envelope", func(t *testing.T) {
		notFoundHandler := buildAPIHandlerWithUploadAndHandlers(
			loader,
			apiRuntimeFlags{},
			service,
			targetjob.NewHandler(),
			practiceRoutes{},
			uploadRoutes{},
			resumeRoutes{
				Handler:     resumehandler.New(resumehandler.Options{Service: &apiUpdateVersionService{out: version, err: domainresume.ErrNotFound}, Session: currentUserFromContext}),
				Idempotency: idempotency.New(idempotency.MiddlewareOptions{Store: &apiIdempotencyStore{}, Now: func() time.Time { return sessionTime }}),
			},
		)
		rec := httptest.NewRecorder()
		req := authenticatedAPIRequest(http.MethodPatch, "/api/v1/resume-versions/missing", validAPIUpdateBody(), "idem-update")

		notFoundHandler.ServeHTTP(rec, req)

		assertAPIStatusCode(t, rec, http.StatusNotFound, sharederrors.CodeTargetJobNotFound)
	})
}

func TestResumeVersionReadHTTPScenario(t *testing.T) {
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
	sessionTime := time.Date(2026, 5, 17, 18, 30, 0, 0, time.UTC)
	service := auth.NewPasswordlessService(auth.PasswordlessServiceOptions{
		Store: &apiAuthStore{
			session: auth.SessionRecord{
				ID:        "session-1",
				UserID:    "user-1",
				Status:    auth.SessionStatusActive,
				ExpiresAt: sessionTime.Add(time.Hour),
			},
			user: auth.UserContext{ID: "user-1", Email: "candidate@example.com"},
		},
		SessionCookieSecret: "session-secret",
		Now:                 func() time.Time { return sessionTime },
	})
	version := apiVersion("version-1", "asset-1", sessionTime)
	resumeSvc := &apiVersionReadService{
		getOut: version,
		listOut: api.PaginatedResumeVersion{
			Items:    []api.ResumeVersion{version},
			PageInfo: api.PageInfo{PageSize: 20, HasMore: false},
		},
	}
	handler := buildAPIHandlerWithUploadAndHandlers(
		loader,
		apiRuntimeFlags{},
		service,
		targetjob.NewHandler(),
		practiceRoutes{},
		uploadRoutes{},
		resumeRoutes{Handler: resumehandler.New(resumehandler.Options{Service: resumeSvc, Session: currentUserFromContext})},
	)

	t.Run("missing session protects version routes", func(t *testing.T) {
		for _, path := range []string{"/api/v1/resume-versions/version-1", "/api/v1/resumes/asset-1/versions"} {
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
			if rec.Code != http.StatusUnauthorized {
				t.Fatalf("%s status = %d body=%s", path, rec.Code, rec.Body.String())
			}
		}
	})

	t.Run("get version passes path param", func(t *testing.T) {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, authenticatedAPIRequest(http.MethodGet, "/api/v1/resume-versions/version-1", "", ""))
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
		}
		if resumeSvc.getUserID != "user-1" || resumeSvc.getVersionID != "version-1" {
			t.Fatalf("get scope user=%q version=%q", resumeSvc.getUserID, resumeSvc.getVersionID)
		}
	})

	t.Run("list versions passes asset and pagination", func(t *testing.T) {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, authenticatedAPIRequest(http.MethodGet, "/api/v1/resumes/asset-1/versions?pageSize=20&cursor=cursor-1", "", ""))
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
		}
		if resumeSvc.listIn.UserID != "user-1" || resumeSvc.listIn.ResumeAssetID != "asset-1" || resumeSvc.listIn.PageSize != 20 || resumeSvc.listIn.Cursor != "cursor-1" {
			t.Fatalf("list input = %+v", resumeSvc.listIn)
		}
	})

	t.Run("not found and invalid cursor envelopes", func(t *testing.T) {
		notFoundSvc := &apiVersionReadService{getErr: domainresume.ErrNotFound, listErr: domainresume.ErrNotFound}
		notFoundHandler := buildAPIHandlerWithUploadAndHandlers(
			loader,
			apiRuntimeFlags{},
			service,
			targetjob.NewHandler(),
			practiceRoutes{},
			uploadRoutes{},
			resumeRoutes{Handler: resumehandler.New(resumehandler.Options{Service: notFoundSvc, Session: currentUserFromContext})},
		)
		rec := httptest.NewRecorder()
		notFoundHandler.ServeHTTP(rec, authenticatedAPIRequest(http.MethodGet, "/api/v1/resume-versions/missing", "", ""))
		assertAPIStatusCode(t, rec, http.StatusNotFound, sharederrors.CodeTargetJobNotFound)

		invalidSvc := &apiVersionReadService{listErr: domainresume.ErrInvalidCursor}
		invalidHandler := buildAPIHandlerWithUploadAndHandlers(
			loader,
			apiRuntimeFlags{},
			service,
			targetjob.NewHandler(),
			practiceRoutes{},
			uploadRoutes{},
			resumeRoutes{Handler: resumehandler.New(resumehandler.Options{Service: invalidSvc, Session: currentUserFromContext})},
		)
		rec = httptest.NewRecorder()
		invalidHandler.ServeHTTP(rec, authenticatedAPIRequest(http.MethodGet, "/api/v1/resumes/asset-1/versions?cursor=bad", "", ""))
		assertAPIStatusCode(t, rec, http.StatusUnprocessableEntity, sharederrors.CodeValidationFailed)
	})
}

func TestResumeBranchVersionHTTPScenario(t *testing.T) {
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
	sessionTime := time.Date(2026, 5, 17, 20, 45, 0, 0, time.UTC)
	service := auth.NewPasswordlessService(auth.PasswordlessServiceOptions{
		Store: &apiAuthStore{
			session: auth.SessionRecord{
				ID:        "session-1",
				UserID:    "user-1",
				Status:    auth.SessionStatusActive,
				ExpiresAt: sessionTime.Add(time.Hour),
			},
			user: auth.UserContext{ID: "user-1", Email: "candidate@example.com"},
		},
		SessionCookieSecret: "session-secret",
		Now:                 func() time.Time { return sessionTime },
	})
	version := apiTargetedVersion("version-1", "asset-1", "parent-1", "target-1", sharedtypes.ResumeSeedStrategyCopyMaster, sessionTime)
	idemStore := &apiIdempotencyStore{}
	resumeSvc := &apiBranchVersionService{result: domainresume.BranchVersionResult{Status: http.StatusCreated, Version: version}}
	handler := buildAPIHandlerWithUploadAndHandlers(
		loader,
		apiRuntimeFlags{},
		service,
		targetjob.NewHandler(),
		practiceRoutes{},
		uploadRoutes{},
		resumeRoutes{
			Handler: resumehandler.New(resumehandler.Options{Service: resumeSvc, Session: currentUserFromContext}),
			Idempotency: idempotency.New(idempotency.MiddlewareOptions{
				Store: idemStore,
				Now:   func() time.Time { return sessionTime },
				NewID: func() string { return "idem-rec-branch" },
			}),
		},
	)

	t.Run("missing session", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/resume-versions", strings.NewReader(validAPIBranchBody("copy_master")))
		req.Header.Set(idempotency.HeaderName, "idem-branch")

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
		}
		if !strings.Contains(rec.Body.String(), `"code":"AUTH_UNAUTHORIZED"`) {
			t.Fatalf("expected auth envelope, got %s", rec.Body.String())
		}
	})

	t.Run("missing idempotency key", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := authenticatedAPIRequest(http.MethodPost, "/api/v1/resume-versions", validAPIBranchBody("copy_master"), "")

		handler.ServeHTTP(rec, req)

		assertAPIStatusCode(t, rec, http.StatusUnprocessableEntity, sharederrors.CodeValidationFailed)
	})

	t.Run("copy master persists idempotency resource", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := authenticatedAPIRequest(http.MethodPost, "/api/v1/resume-versions", validAPIBranchBody("copy_master"), "idem-branch")

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
		}
		if resumeSvc.calls != 1 || resumeSvc.in.UserID != "user-1" || resumeSvc.in.ParentVersionID != "parent-1" || resumeSvc.in.TargetJobID != "target-1" || resumeSvc.in.SeedStrategy != sharedtypes.ResumeSeedStrategyCopyMaster {
			t.Fatalf("service calls=%d input=%+v", resumeSvc.calls, resumeSvc.in)
		}
		if idemStore.completeIn.Operation != "branchResumeVersion" || idemStore.completeIn.ResourceType != "resume_version" || idemStore.completeIn.ResourceID != "version-1" {
			t.Fatalf("idempotency completion = %+v", idemStore.completeIn)
		}
		if rec.Header().Get("X-Idempotency-Resource-Type") != "" || rec.Header().Get("X-Idempotency-Resource-ID") != "" {
			t.Fatalf("internal idempotency headers leaked: %v", rec.Header())
		}
	})

	t.Run("ai select returns accepted job", func(t *testing.T) {
		aiVersion := apiTargetedVersion("version-ai", "asset-1", "parent-1", "target-1", sharedtypes.ResumeSeedStrategyAiSelect, sessionTime)
		aiSvc := &apiBranchVersionService{result: domainresume.BranchVersionResult{Status: http.StatusAccepted, Accepted: &api.BranchResumeVersionAccepted{
			ResumeVersionId: "version-ai",
			Version:         aiVersion,
			Job: api.Job{
				Id:           "job-1",
				JobType:      api.JobTypeResumeTailor,
				ResourceType: api.ResourceTypeResumeTailorRun,
				ResourceId:   "tailor-run-1",
				Status:       sharedtypes.JobStatusQueued,
				CreatedAt:    sessionTime.Format(time.RFC3339),
				UpdatedAt:    sessionTime.Format(time.RFC3339),
			},
		}}}
		aiHandler := buildAPIHandlerWithUploadAndHandlers(
			loader,
			apiRuntimeFlags{},
			service,
			targetjob.NewHandler(),
			practiceRoutes{},
			uploadRoutes{},
			resumeRoutes{
				Handler:     resumehandler.New(resumehandler.Options{Service: aiSvc, Session: currentUserFromContext}),
				Idempotency: idempotency.New(idempotency.MiddlewareOptions{Store: &apiIdempotencyStore{}, Now: func() time.Time { return sessionTime }}),
			},
		)
		rec := httptest.NewRecorder()
		req := authenticatedAPIRequest(http.MethodPost, "/api/v1/resume-versions", validAPIBranchBody("ai_select"), "idem-branch-ai")

		aiHandler.ServeHTTP(rec, req)

		if rec.Code != http.StatusAccepted {
			t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
		}
		if aiSvc.in.SeedStrategy != sharedtypes.ResumeSeedStrategyAiSelect {
			t.Fatalf("ai service input = %+v", aiSvc.in)
		}
	})

	t.Run("idempotency replay bypasses service", func(t *testing.T) {
		replayStore := &apiIdempotencyStore{reservation: idempotency.Reservation{
			State:          idempotency.StateReplay,
			RecordID:       "idem-rec-replay",
			ResponseStatus: http.StatusCreated,
			ResponseBody:   []byte(`{"id":"version-replay","resumeAssetId":"asset-1","parentVersionId":"parent-1","versionType":"targeted","targetJobId":"target-1","displayName":"Targeted","seedStrategy":"copy_master","structuredProfile":{},"provenance":{"promptVersion":"p","rubricVersion":"r","modelId":"m","language":"en","featureFlag":"f","dataSourceVersion":"d"},"suggestions":[],"createdAt":"2026-05-17T20:45:00Z","updatedAt":"2026-05-17T20:45:00Z"}`),
		}}
		replaySvc := &apiBranchVersionService{result: domainresume.BranchVersionResult{Status: http.StatusCreated, Version: version}}
		replayHandler := buildAPIHandlerWithUploadAndHandlers(
			loader,
			apiRuntimeFlags{},
			service,
			targetjob.NewHandler(),
			practiceRoutes{},
			uploadRoutes{},
			resumeRoutes{
				Handler:     resumehandler.New(resumehandler.Options{Service: replaySvc, Session: currentUserFromContext}),
				Idempotency: idempotency.New(idempotency.MiddlewareOptions{Store: replayStore, Now: func() time.Time { return sessionTime }}),
			},
		)
		rec := httptest.NewRecorder()
		req := authenticatedAPIRequest(http.MethodPost, "/api/v1/resume-versions", validAPIBranchBody("copy_master"), "idem-branch")

		replayHandler.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated || rec.Header().Get(idempotency.ReplayHeader) != "true" {
			t.Fatalf("replay status=%d header=%q body=%s", rec.Code, rec.Header().Get(idempotency.ReplayHeader), rec.Body.String())
		}
		if replaySvc.calls != 0 {
			t.Fatalf("service calls on replay = %d, want 0", replaySvc.calls)
		}
	})

	t.Run("idempotency mismatch maps to 409", func(t *testing.T) {
		mismatchHandler := buildAPIHandlerWithUploadAndHandlers(
			loader,
			apiRuntimeFlags{},
			service,
			targetjob.NewHandler(),
			practiceRoutes{},
			uploadRoutes{},
			resumeRoutes{
				Handler:     resumehandler.New(resumehandler.Options{Service: &apiBranchVersionService{result: domainresume.BranchVersionResult{Status: http.StatusCreated, Version: version}}, Session: currentUserFromContext}),
				Idempotency: idempotency.New(idempotency.MiddlewareOptions{Store: &apiIdempotencyStore{err: idempotency.ErrFingerprintMismatch}, Now: func() time.Time { return sessionTime }}),
			},
		)
		rec := httptest.NewRecorder()
		req := authenticatedAPIRequest(http.MethodPost, "/api/v1/resume-versions", validAPIBranchBody("copy_master"), "idem-branch")

		mismatchHandler.ServeHTTP(rec, req)

		assertAPIStatusCode(t, rec, http.StatusConflict, sharederrors.CodeIdempotencyKeyMismatch)
	})

	t.Run("validation and not found envelopes", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := authenticatedAPIRequest(http.MethodPost, "/api/v1/resume-versions", validAPIBranchBody("invalid"), "idem-invalid")
		handler.ServeHTTP(rec, req)
		assertAPIStatusCode(t, rec, http.StatusUnprocessableEntity, sharederrors.CodeValidationFailed)

		notFoundHandler := buildAPIHandlerWithUploadAndHandlers(
			loader,
			apiRuntimeFlags{},
			service,
			targetjob.NewHandler(),
			practiceRoutes{},
			uploadRoutes{},
			resumeRoutes{
				Handler:     resumehandler.New(resumehandler.Options{Service: &apiBranchVersionService{result: domainresume.BranchVersionResult{Status: http.StatusCreated, Version: version}, err: domainresume.ErrNotFound}, Session: currentUserFromContext}),
				Idempotency: idempotency.New(idempotency.MiddlewareOptions{Store: &apiIdempotencyStore{}, Now: func() time.Time { return sessionTime }}),
			},
		)
		rec = httptest.NewRecorder()
		req = authenticatedAPIRequest(http.MethodPost, "/api/v1/resume-versions", validAPIBranchBody("copy_master"), "idem-not-found")
		notFoundHandler.ServeHTTP(rec, req)
		assertAPIStatusCode(t, rec, http.StatusNotFound, sharederrors.CodeTargetJobNotFound)
	})
}

func TestResumeTailorEndpointsHTTPScenario(t *testing.T) {
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
	sessionTime := time.Date(2026, 5, 18, 10, 45, 0, 0, time.UTC)
	service := auth.NewPasswordlessService(auth.PasswordlessServiceOptions{
		Store: &apiAuthStore{
			session: auth.SessionRecord{
				ID:        "session-1",
				UserID:    "user-1",
				Status:    auth.SessionStatusActive,
				ExpiresAt: sessionTime.Add(time.Hour),
			},
			user: auth.UserContext{ID: "user-1", Email: "candidate@example.com"},
		},
		SessionCookieSecret: "session-secret",
		Now:                 func() time.Time { return sessionTime },
	})
	idemStore := &apiIdempotencyStore{}
	resumeSvc := &apiTailorRunService{
		requestOut: api.ResumeTailorRunWithJob{
			TailorRunId: "tailor-run-1",
			Job: api.Job{
				Id:           "job-1",
				JobType:      api.JobTypeResumeTailor,
				ResourceType: api.ResourceTypeResumeTailorRun,
				ResourceId:   "tailor-run-1",
				Status:       sharedtypes.JobStatusQueued,
				CreatedAt:    sessionTime.Format(time.RFC3339),
				UpdatedAt:    sessionTime.Format(time.RFC3339),
			},
		},
		getOut: api.ResumeTailorRun{
			Id:            "tailor-run-1",
			Status:        "queued",
			TargetJobId:   "target-1",
			ResumeAssetId: "asset-1",
			Suggestions:   []api.ResumeTailorBulletSuggestion{},
			CreatedAt:     sessionTime.Format(time.RFC3339),
			UpdatedAt:     sessionTime.Format(time.RFC3339),
		},
	}
	handler := buildAPIHandlerWithUploadAndHandlers(
		loader,
		apiRuntimeFlags{},
		service,
		targetjob.NewHandler(),
		practiceRoutes{},
		uploadRoutes{},
		resumeRoutes{
			Handler: resumehandler.New(resumehandler.Options{Service: resumeSvc, Session: currentUserFromContext}),
			Idempotency: idempotency.New(idempotency.MiddlewareOptions{
				Store: idemStore,
				Now:   func() time.Time { return sessionTime },
				NewID: func() string { return "idem-rec-tailor" },
			}),
		},
	)

	t.Run("missing session", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/resume/tailor", strings.NewReader(validAPIRequestTailorBody("gap_review")))
		req.Header.Set(idempotency.HeaderName, "idem-tailor")

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
		}
	})

	t.Run("missing idempotency key", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := authenticatedAPIRequest(http.MethodPost, "/api/v1/resume/tailor", validAPIRequestTailorBody("gap_review"), "")

		handler.ServeHTTP(rec, req)

		assertAPIStatusCode(t, rec, http.StatusUnprocessableEntity, sharederrors.CodeValidationFailed)
	})

	t.Run("request tailor persists idempotency resource", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := authenticatedAPIRequest(http.MethodPost, "/api/v1/resume/tailor", validAPIRequestTailorBody("gap_review"), "idem-tailor")

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusAccepted {
			t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
		}
		if resumeSvc.requestCalls != 1 || resumeSvc.requestIn.UserID != "user-1" || resumeSvc.requestIn.ResumeAssetID != "asset-1" || resumeSvc.requestIn.TargetJobID != "target-1" || resumeSvc.requestIn.Mode != "gap_review" {
			t.Fatalf("service calls=%d input=%+v", resumeSvc.requestCalls, resumeSvc.requestIn)
		}
		if idemStore.completeIn.Operation != "requestResumeTailor" || idemStore.completeIn.ResourceType != "resume_tailor_run" || idemStore.completeIn.ResourceID != "tailor-run-1" {
			t.Fatalf("idempotency completion = %+v", idemStore.completeIn)
		}
	})

	t.Run("get tailor run", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := authenticatedAPIRequest(http.MethodGet, "/api/v1/resume/tailor-runs/tailor-run-1", "", "")

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
		}
		if resumeSvc.getUserID != "user-1" || resumeSvc.getTailorRunID != "tailor-run-1" {
			t.Fatalf("get scope user=%q run=%q", resumeSvc.getUserID, resumeSvc.getTailorRunID)
		}
	})

	t.Run("idempotency replay bypasses service", func(t *testing.T) {
		replayStore := &apiIdempotencyStore{reservation: idempotency.Reservation{
			State:          idempotency.StateReplay,
			RecordID:       "idem-rec-tailor-replay",
			ResponseStatus: http.StatusAccepted,
			ResponseBody:   []byte(`{"tailorRunId":"tailor-run-replay","job":{"id":"job-replay","jobType":"resume_tailor","status":"queued","resourceType":"resume_tailor_run","resourceId":"tailor-run-replay","errorCode":null,"createdAt":"2026-05-18T10:45:00Z","updatedAt":"2026-05-18T10:45:00Z"}}`),
		}}
		replaySvc := &apiTailorRunService{requestOut: resumeSvc.requestOut}
		replayHandler := buildAPIHandlerWithUploadAndHandlers(
			loader,
			apiRuntimeFlags{},
			service,
			targetjob.NewHandler(),
			practiceRoutes{},
			uploadRoutes{},
			resumeRoutes{
				Handler:     resumehandler.New(resumehandler.Options{Service: replaySvc, Session: currentUserFromContext}),
				Idempotency: idempotency.New(idempotency.MiddlewareOptions{Store: replayStore, Now: func() time.Time { return sessionTime }}),
			},
		)
		rec := httptest.NewRecorder()
		req := authenticatedAPIRequest(http.MethodPost, "/api/v1/resume/tailor", validAPIRequestTailorBody("gap_review"), "idem-tailor")

		replayHandler.ServeHTTP(rec, req)

		if rec.Code != http.StatusAccepted || rec.Header().Get(idempotency.ReplayHeader) != "true" {
			t.Fatalf("replay status=%d header=%q body=%s", rec.Code, rec.Header().Get(idempotency.ReplayHeader), rec.Body.String())
		}
		if replaySvc.requestCalls != 0 {
			t.Fatalf("service calls on replay = %d, want 0", replaySvc.requestCalls)
		}
	})

	t.Run("validation and not found envelopes", func(t *testing.T) {
		invalid := httptest.NewRecorder()
		handler.ServeHTTP(invalid, authenticatedAPIRequest(http.MethodPost, "/api/v1/resume/tailor", validAPIRequestTailorBody("unsupported"), "idem-tailor-invalid"))
		assertAPIStatusCode(t, invalid, http.StatusUnprocessableEntity, sharederrors.CodeValidationFailed)

		notFoundSvc := &apiTailorRunService{requestErr: domainresume.ErrNotFound, getErr: domainresume.ErrNotFound}
		notFoundHandler := buildAPIHandlerWithUploadAndHandlers(
			loader,
			apiRuntimeFlags{},
			service,
			targetjob.NewHandler(),
			practiceRoutes{},
			uploadRoutes{},
			resumeRoutes{Handler: resumehandler.New(resumehandler.Options{Service: notFoundSvc, Session: currentUserFromContext})},
		)
		rec := httptest.NewRecorder()
		notFoundHandler.ServeHTTP(rec, authenticatedAPIRequest(http.MethodGet, "/api/v1/resume/tailor-runs/missing", "", ""))
		assertAPIStatusCode(t, rec, http.StatusNotFound, sharederrors.CodeTargetJobNotFound)
	})
}

func TestResumeSuggestionAcceptRejectHTTPScenario(t *testing.T) {
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
	sessionTime := time.Date(2026, 5, 18, 11, 45, 0, 0, time.UTC)
	service := auth.NewPasswordlessService(auth.PasswordlessServiceOptions{
		Store: &apiAuthStore{
			session: auth.SessionRecord{
				ID:        "session-1",
				UserID:    "user-1",
				Status:    auth.SessionStatusActive,
				ExpiresAt: sessionTime.Add(time.Hour),
			},
			user: auth.UserContext{ID: "user-1", Email: "candidate@example.com"},
		},
		SessionCookieSecret: "session-secret",
		Now:                 func() time.Time { return sessionTime },
	})
	idemStore := &apiIdempotencyStore{}
	resumeSvc := &apiTailorRunService{
		acceptOut: apiTargetedVersion("version-1", "asset-1", "parent-1", "target-1", sharedtypes.ResumeSeedStrategyCopyMaster, sessionTime),
		rejectOut: apiTargetedVersion("version-1", "asset-1", "parent-1", "target-1", sharedtypes.ResumeSeedStrategyCopyMaster, sessionTime),
	}
	handler := buildAPIHandlerWithUploadAndHandlers(
		loader,
		apiRuntimeFlags{},
		service,
		targetjob.NewHandler(),
		practiceRoutes{},
		uploadRoutes{},
		resumeRoutes{
			Handler: resumehandler.New(resumehandler.Options{Service: resumeSvc, Session: currentUserFromContext}),
			Idempotency: idempotency.New(idempotency.MiddlewareOptions{
				Store: idemStore,
				Now:   func() time.Time { return sessionTime },
				NewID: func() string { return "idem-rec-suggestion" },
			}),
		},
	)

	t.Run("missing session", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/resume-versions/version-1/suggestions/suggestion-1/accept", nil)
		req.Header.Set(idempotency.HeaderName, "idem-accept")

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
		}
	})

	t.Run("missing idempotency key", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := authenticatedAPIRequest(http.MethodPost, "/api/v1/resume-versions/version-1/suggestions/suggestion-1/accept", "", "")

		handler.ServeHTTP(rec, req)

		assertAPIStatusCode(t, rec, http.StatusUnprocessableEntity, sharederrors.CodeValidationFailed)
	})

	t.Run("accept persists idempotency resource", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := authenticatedAPIRequest(http.MethodPost, "/api/v1/resume-versions/version-1/suggestions/suggestion-1/accept", "", "idem-accept")

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
		}
		if resumeSvc.acceptCalls != 1 || resumeSvc.acceptIn.UserID != "user-1" || resumeSvc.acceptIn.ResumeVersionID != "version-1" || resumeSvc.acceptIn.SuggestionID != "suggestion-1" {
			t.Fatalf("accept calls=%d input=%+v", resumeSvc.acceptCalls, resumeSvc.acceptIn)
		}
		if idemStore.completeIn.Operation != "acceptResumeTailorSuggestion" || idemStore.completeIn.ResourceType != "resume_version" || idemStore.completeIn.ResourceID != "version-1" {
			t.Fatalf("idempotency completion = %+v", idemStore.completeIn)
		}
	})

	t.Run("reject route", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := authenticatedAPIRequest(http.MethodPost, "/api/v1/resume-versions/version-1/suggestions/suggestion-2/reject", "", "idem-reject")

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
		}
		if resumeSvc.rejectCalls != 1 || resumeSvc.rejectIn.SuggestionID != "suggestion-2" {
			t.Fatalf("reject calls=%d input=%+v", resumeSvc.rejectCalls, resumeSvc.rejectIn)
		}
	})

	t.Run("idempotency replay bypasses service", func(t *testing.T) {
		replayStore := &apiIdempotencyStore{reservation: idempotency.Reservation{
			State:          idempotency.StateReplay,
			RecordID:       "idem-rec-suggestion-replay",
			ResponseStatus: http.StatusOK,
			ResponseBody:   []byte(`{"id":"version-replay","resumeAssetId":"asset-1","versionType":"targeted","displayName":"Targeted","structuredProfile":{},"provenance":{"promptVersion":"p","rubricVersion":"r","modelId":"m","language":"en","featureFlag":"f","dataSourceVersion":"d"},"suggestions":[],"createdAt":"2026-05-18T11:45:00Z","updatedAt":"2026-05-18T11:45:00Z"}`),
		}}
		replaySvc := &apiTailorRunService{acceptOut: resumeSvc.acceptOut}
		replayHandler := buildAPIHandlerWithUploadAndHandlers(
			loader,
			apiRuntimeFlags{},
			service,
			targetjob.NewHandler(),
			practiceRoutes{},
			uploadRoutes{},
			resumeRoutes{
				Handler:     resumehandler.New(resumehandler.Options{Service: replaySvc, Session: currentUserFromContext}),
				Idempotency: idempotency.New(idempotency.MiddlewareOptions{Store: replayStore, Now: func() time.Time { return sessionTime }}),
			},
		)
		rec := httptest.NewRecorder()
		req := authenticatedAPIRequest(http.MethodPost, "/api/v1/resume-versions/version-1/suggestions/suggestion-1/accept", "", "idem-accept")

		replayHandler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK || rec.Header().Get(idempotency.ReplayHeader) != "true" {
			t.Fatalf("replay status=%d header=%q body=%s", rec.Code, rec.Header().Get(idempotency.ReplayHeader), rec.Body.String())
		}
		if replaySvc.acceptCalls != 0 {
			t.Fatalf("service calls on replay = %d, want 0", replaySvc.acceptCalls)
		}
	})

	t.Run("mismatch already decided and not found envelopes", func(t *testing.T) {
		mismatchHandler := buildAPIHandlerWithUploadAndHandlers(
			loader,
			apiRuntimeFlags{},
			service,
			targetjob.NewHandler(),
			practiceRoutes{},
			uploadRoutes{},
			resumeRoutes{
				Handler:     resumehandler.New(resumehandler.Options{Service: resumeSvc, Session: currentUserFromContext}),
				Idempotency: idempotency.New(idempotency.MiddlewareOptions{Store: &apiIdempotencyStore{err: idempotency.ErrFingerprintMismatch}, Now: func() time.Time { return sessionTime }}),
			},
		)
		rec := httptest.NewRecorder()
		mismatchHandler.ServeHTTP(rec, authenticatedAPIRequest(http.MethodPost, "/api/v1/resume-versions/version-1/suggestions/suggestion-1/accept", "", "idem-accept"))
		assertAPIStatusCode(t, rec, http.StatusConflict, sharederrors.CodeIdempotencyKeyMismatch)

		alreadySvc := &apiTailorRunService{acceptErr: domainresume.ErrSuggestionAlreadyDecided}
		alreadyHandler := buildAPIHandlerWithUploadAndHandlers(
			loader,
			apiRuntimeFlags{},
			service,
			targetjob.NewHandler(),
			practiceRoutes{},
			uploadRoutes{},
			resumeRoutes{Handler: resumehandler.New(resumehandler.Options{Service: alreadySvc, Session: currentUserFromContext})},
		)
		rec = httptest.NewRecorder()
		alreadyHandler.ServeHTTP(rec, authenticatedAPIRequest(http.MethodPost, "/api/v1/resume-versions/version-1/suggestions/suggestion-1/accept", "", "idem-accept-already"))
		assertAPIStatusCode(t, rec, http.StatusConflict, sharederrors.CodeValidationFailed)

		notFoundSvc := &apiTailorRunService{acceptErr: domainresume.ErrNotFound}
		notFoundHandler := buildAPIHandlerWithUploadAndHandlers(
			loader,
			apiRuntimeFlags{},
			service,
			targetjob.NewHandler(),
			practiceRoutes{},
			uploadRoutes{},
			resumeRoutes{Handler: resumehandler.New(resumehandler.Options{Service: notFoundSvc, Session: currentUserFromContext})},
		)
		rec = httptest.NewRecorder()
		notFoundHandler.ServeHTTP(rec, authenticatedAPIRequest(http.MethodPost, "/api/v1/resume-versions/version-1/suggestions/foreign/accept", "", "idem-accept-missing"))
		assertAPIStatusCode(t, rec, http.StatusNotFound, sharederrors.CodeTargetJobNotFound)
	})
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

func TestBuildAPIHandlerMountsJobRouteBehindSessionMiddleware(t *testing.T) {
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
	handler := buildAPIHandlerWithUploadReportDebriefJobsAndHandlers(
		loader,
		apiRuntimeFlags{},
		service,
		targetjob.NewHandler(),
		practiceRoutes{},
		uploadRoutes{},
		resumeRoutes{},
		reportRoutes{},
		debriefRoutes{},
		jobsRoutes{Handler: apijobs.NewHandler(apijobs.HandlerOptions{})},
	)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/jobs/018f2a40-0000-7000-9000-0000000000a1", nil)
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d body=%s; job route is not mounted behind auth middleware", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"code":"AUTH_UNAUTHORIZED"`) {
		t.Fatalf("expected auth middleware envelope, got %s", rec.Body.String())
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
	if runtime.Handler == nil || runtime.Idempotency == nil || runtime.Handlers == nil || runtime.ParseAI == nil {
		t.Fatalf("runtime missing handler/idempotency/handlers/AI wiring: %+v", runtime)
	}
	if !runtime.Handles(string(jobs.JobTypeResumeParse)) {
		t.Fatalf("runtime does not contribute handler for %s", jobs.JobTypeResumeParse)
	}
	if !runtime.Handles(string(jobs.JobTypeResumeTailor)) {
		t.Fatalf("runtime does not contribute handler for %s", jobs.JobTypeResumeTailor)
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

func TestBuildReportRuntimeWiresRoutesHandlerAndAI(t *testing.T) {
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
	if runtime.Handler == nil || runtime.Handlers == nil || runtime.Service == nil {
		t.Fatalf("runtime missing handler/handlers/service wiring: %+v", runtime)
	}
	if !runtime.Handles(string(jobs.JobTypeReportGenerate)) {
		t.Fatalf("runtime does not contribute handler for %s", jobs.JobTypeReportGenerate)
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

	runtime, err := buildTargetJobRuntime(loader, nil, slog.New(slog.NewTextHandler(io.Discard, nil)), nil, nil)
	if err != nil {
		t.Fatalf("buildTargetJobRuntime: %v", err)
	}
	defer runtime.Close()
	if runtime.Handler == nil || runtime.Handlers == nil || runtime.AI == nil || runtime.AI.Client == nil {
		t.Fatalf("runtime missing handler/handlers/AI wiring: %+v", runtime)
	}
	if !runtime.Handles(string(jobs.JobTypeTargetImport)) || !runtime.Handles(string(jobs.JobTypeSourceRefresh)) {
		t.Fatalf("runtime does not contribute target_import/source_refresh handlers: %+v", runtime.Handlers)
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

	runtime, err := buildTargetJobRuntime(loader, nil, slog.New(slog.NewTextHandler(io.Discard, nil)), &apiUploadFileDeleter{}, nil)
	if err != nil {
		t.Fatalf("buildTargetJobRuntime: %v", err)
	}
	if !runtime.Handles(string(jobs.JobTypePrivacyDelete)) {
		t.Fatalf("runtime does not contribute handler for %s", jobs.JobTypePrivacyDelete)
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

	runtime, err := buildTargetJobRuntime(loader, nil, slog.New(slog.NewTextHandler(io.Discard, nil)), nil, nil)
	if err != nil {
		t.Fatalf("buildTargetJobRuntime: %v", err)
	}
	if !runtime.Handles(string(jobs.JobTypeDebriefGenerate)) {
		t.Fatalf("runtime does not contribute handler for %s", jobs.JobTypeDebriefGenerate)
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

func TestBuildAuthServiceUsesMailpitDeliveryWriterWhenConfigured(t *testing.T) {
	dir := t.TempDir()
	writeAPIFile(t, filepath.Join(dir, "config.yaml"), `
auth:
  challengeTokenPepper: "pepper"
  sessionCookieSecret: "session-secret"
email:
  provider: "mailpit"
  smtpHost: "127.0.0.1"
  smtpPort: 1025
  fromAddress: "noreply@easyinterview.local"
  verifyBaseURL: "http://127.0.0.1:8080/api/v1/auth/email/verify"
`)
	loader, err := config.Load(config.Options{ConfigDir: dir})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	service, writer, err := buildAuthService(loader, db)
	if err != nil {
		t.Fatalf("buildAuthService: %v", err)
	}
	if service == nil {
		t.Fatal("auth service was not constructed")
	}
	if _, ok := writer.(*auth.SMTPDeliveryWriter); !ok {
		t.Fatalf("delivery writer type = %T, want *auth.SMTPDeliveryWriter", writer)
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

type apiConfirmResumeService struct {
	calls int
	in    domainresume.ConfirmStructuredMasterInput
	out   api.ResumeVersion
	err   error
}

func (s *apiConfirmResumeService) RegisterResume(context.Context, domainresume.RegisterInput) (api.ResumeAssetWithJob, error) {
	return api.ResumeAssetWithJob{}, errors.New("not implemented")
}

func (s *apiConfirmResumeService) ConfirmStructuredMaster(_ context.Context, in domainresume.ConfirmStructuredMasterInput) (api.ResumeVersion, error) {
	s.calls++
	s.in = in
	return s.out, s.err
}

type apiUpdateVersionService struct {
	calls int
	in    domainresume.UpdateVersionRequest
	out   api.ResumeVersion
	err   error
}

func (s *apiUpdateVersionService) RegisterResume(context.Context, domainresume.RegisterInput) (api.ResumeAssetWithJob, error) {
	return api.ResumeAssetWithJob{}, errors.New("not implemented")
}

func (s *apiUpdateVersionService) UpdateResumeVersion(_ context.Context, in domainresume.UpdateVersionRequest) (api.ResumeVersion, error) {
	s.calls++
	s.in = in
	return s.out, s.err
}

type apiBranchVersionService struct {
	calls  int
	in     domainresume.BranchVersionRequest
	result domainresume.BranchVersionResult
	err    error
}

func (s *apiBranchVersionService) RegisterResume(context.Context, domainresume.RegisterInput) (api.ResumeAssetWithJob, error) {
	return api.ResumeAssetWithJob{}, errors.New("not implemented")
}

func (s *apiBranchVersionService) BranchResumeVersion(_ context.Context, in domainresume.BranchVersionRequest) (domainresume.BranchVersionResult, error) {
	s.calls++
	s.in = in
	return s.result, s.err
}

type apiTailorRunService struct {
	requestCalls int
	requestIn    domainresume.RequestTailorRunInput
	requestOut   api.ResumeTailorRunWithJob
	requestErr   error

	getUserID      string
	getTailorRunID string
	getOut         api.ResumeTailorRun
	getErr         error

	acceptCalls int
	acceptIn    domainresume.SuggestionDecisionRequest
	acceptOut   api.ResumeVersion
	acceptErr   error

	rejectCalls int
	rejectIn    domainresume.SuggestionDecisionRequest
	rejectOut   api.ResumeVersion
	rejectErr   error
}

func (s *apiTailorRunService) RegisterResume(context.Context, domainresume.RegisterInput) (api.ResumeAssetWithJob, error) {
	return api.ResumeAssetWithJob{}, errors.New("not implemented")
}

func (s *apiTailorRunService) RequestResumeTailor(_ context.Context, in domainresume.RequestTailorRunInput) (api.ResumeTailorRunWithJob, error) {
	s.requestCalls++
	s.requestIn = in
	return s.requestOut, s.requestErr
}

func (s *apiTailorRunService) GetResumeTailorRun(_ context.Context, userID string, tailorRunID string) (api.ResumeTailorRun, error) {
	s.getUserID = userID
	s.getTailorRunID = tailorRunID
	return s.getOut, s.getErr
}

func (s *apiTailorRunService) AcceptResumeTailorSuggestion(_ context.Context, in domainresume.SuggestionDecisionRequest) (api.ResumeVersion, error) {
	s.acceptCalls++
	s.acceptIn = in
	return s.acceptOut, s.acceptErr
}

func (s *apiTailorRunService) RejectResumeTailorSuggestion(_ context.Context, in domainresume.SuggestionDecisionRequest) (api.ResumeVersion, error) {
	s.rejectCalls++
	s.rejectIn = in
	return s.rejectOut, s.rejectErr
}

type apiVersionReadService struct {
	getUserID    string
	getVersionID string
	getOut       api.ResumeVersion
	getErr       error

	listIn  domainresume.ListVersionRequest
	listOut api.PaginatedResumeVersion
	listErr error
}

func (s *apiVersionReadService) RegisterResume(context.Context, domainresume.RegisterInput) (api.ResumeAssetWithJob, error) {
	return api.ResumeAssetWithJob{}, errors.New("not implemented")
}

func (s *apiVersionReadService) GetResumeVersion(_ context.Context, userID string, versionID string) (api.ResumeVersion, error) {
	s.getUserID = userID
	s.getVersionID = versionID
	return s.getOut, s.getErr
}

func (s *apiVersionReadService) ListResumeVersions(_ context.Context, in domainresume.ListVersionRequest) (api.PaginatedResumeVersion, error) {
	s.listIn = in
	return s.listOut, s.listErr
}

func apiVersion(id, assetID string, now time.Time) api.ResumeVersion {
	return api.ResumeVersion{
		Id:            id,
		ResumeAssetId: assetID,
		VersionType:   sharedtypes.ResumeVersionTypeStructuredMaster,
		DisplayName:   "Structured master",
		StructuredProfile: map[string]any{"headline": "Senior engineer", "provenance": map[string]any{
			"promptVersion": "resume_profile.v1", "rubricVersion": "not_applicable", "modelId": "model-1", "language": "en", "featureFlag": "none", "dataSourceVersion": "asset.v1",
		}},
		Provenance: api.GenerationProvenance{
			PromptVersion: "resume_profile.v1", RubricVersion: "not_applicable", ModelId: "model-1", Language: "en", FeatureFlag: "none", DataSourceVersion: "asset.v1",
		},
		Suggestions: []any{},
		CreatedAt:   now.Format(time.RFC3339),
		UpdatedAt:   now.Format(time.RFC3339),
	}
}

func apiTargetedVersion(id, assetID, parentID, targetID string, strategy sharedtypes.ResumeSeedStrategy, now time.Time) api.ResumeVersion {
	focusAngle := "Platform evidence"
	return api.ResumeVersion{
		Id:              id,
		ResumeAssetId:   assetID,
		ParentVersionId: &parentID,
		VersionType:     sharedtypes.ResumeVersionTypeTargeted,
		TargetJobId:     &targetID,
		DisplayName:     "Targeted",
		SeedStrategy:    &strategy,
		FocusAngle:      &focusAngle,
		StructuredProfile: map[string]any{"provenance": map[string]any{
			"promptVersion": "p", "rubricVersion": "r", "modelId": "m", "language": "en", "featureFlag": "f", "dataSourceVersion": "d",
		}},
		Provenance: api.GenerationProvenance{
			PromptVersion: "p", RubricVersion: "r", ModelId: "m", Language: "en", FeatureFlag: "f", DataSourceVersion: "d",
		},
		Suggestions: []any{},
		CreatedAt:   now.Format(time.RFC3339),
		UpdatedAt:   now.Format(time.RFC3339),
	}
}

type apiIdempotencyStore struct {
	reserveIn   idempotency.ReservationInput
	completeIn  idempotency.CompletionInput
	reservation idempotency.Reservation
	err         error
}

func (s *apiIdempotencyStore) Reserve(_ context.Context, in idempotency.ReservationInput) (idempotency.Reservation, error) {
	s.reserveIn = in
	if s.err != nil {
		return idempotency.Reservation{}, s.err
	}
	if s.reservation.State == "" {
		return idempotency.Reservation{State: idempotency.StateExecute, RecordID: "idem-rec-1"}, nil
	}
	return s.reservation, nil
}

func (s *apiIdempotencyStore) MarkSucceeded(_ context.Context, in idempotency.CompletionInput) error {
	s.completeIn = in
	return nil
}

func (s *apiIdempotencyStore) MarkFailed(_ context.Context, in idempotency.CompletionInput) error {
	s.completeIn = in
	return nil
}

func authenticatedAPIRequest(method, path, body, idempotencyKey string) *http.Request {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
	if idempotencyKey != "" {
		req.Header.Set(idempotency.HeaderName, idempotencyKey)
	}
	return req
}

func validAPIConfirmBody() string {
	return `{"displayName":"Structured master","language":"en","structuredProfile":{"headline":"Senior engineer","provenance":{"promptVersion":"resume_profile.v1","rubricVersion":"not_applicable","modelId":"model-1","language":"en","featureFlag":"none","dataSourceVersion":"asset.v1"}}}`
}

func validAPIUpdateBody() string {
	return `{"displayName":" Updated version ","focusAngle":null,"matchScore":0.82,"structuredProfile":{"summary":"updated"}}`
}

func validAPIBranchBody(strategy string) string {
	return `{"parentVersionId":"parent-1","targetJobId":"target-1","seedStrategy":"` + strategy + `","displayName":" Targeted ","focusAngle":" Platform evidence "}`
}

func validAPIRequestTailorBody(mode string) string {
	return `{"targetJobId":"target-1","resumeAssetId":"asset-1","mode":"` + mode + `"}`
}

func assertAPIStatusCode(t *testing.T, rec *httptest.ResponseRecorder, status int, code string) {
	t.Helper()
	if rec.Code != status {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	var payload api.ApiErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if payload.Error.Code != code {
		t.Fatalf("error code = %q, want %q", payload.Error.Code, code)
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
