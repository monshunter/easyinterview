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
	apipractice "github.com/monshunter/easyinterview/backend/internal/api/practice"
	apireports "github.com/monshunter/easyinterview/backend/internal/api/reports"
	"github.com/monshunter/easyinterview/backend/internal/auth"
	"github.com/monshunter/easyinterview/backend/internal/middleware/idempotency"
	"github.com/monshunter/easyinterview/backend/internal/platform/config"
	"github.com/monshunter/easyinterview/backend/internal/platform/featureflag"
	domainresume "github.com/monshunter/easyinterview/backend/internal/resume"
	resumehandler "github.com/monshunter/easyinterview/backend/internal/resume/handler"
	resumejobs "github.com/monshunter/easyinterview/backend/internal/resume/jobs"
	"github.com/monshunter/easyinterview/backend/internal/runner"
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
		TokenGenerator:        apiFixedTokenGenerator("123456"),
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
	allowed := map[string]struct{}{
		"http://frontend.local:6180": {},
	}
	handler := withLocalDevCORS("dev", allowed, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	preflight := httptest.NewRequest(http.MethodOptions, "/api/v1/runtime-config", nil)
	preflight.Header.Set("Origin", "http://frontend.local:6180")
	preflight.Header.Set("Access-Control-Request-Headers", "Content-Type,Idempotency-Key")
	preflightRec := httptest.NewRecorder()
	handler.ServeHTTP(preflightRec, preflight)
	if preflightRec.Code != http.StatusNoContent {
		t.Fatalf("preflight status = %d body=%s", preflightRec.Code, preflightRec.Body.String())
	}
	if called {
		t.Fatalf("preflight should not call next handler")
	}
	if got := preflightRec.Header().Get("Access-Control-Allow-Origin"); got != "http://frontend.local:6180" {
		t.Fatalf("allow origin = %q", got)
	}
	if got := preflightRec.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
		t.Fatalf("allow credentials = %q", got)
	}
	if got := preflightRec.Header().Get("Access-Control-Allow-Headers"); strings.Contains(got, "Prefer") {
		t.Fatalf("real-mode CORS should not allow fixture Prefer header: %q", got)
	}

	get := httptest.NewRequest(http.MethodGet, "/api/v1/runtime-config", nil)
	get.Header.Set("Origin", "http://frontend.local:6180")
	getRec := httptest.NewRecorder()
	handler.ServeHTTP(getRec, get)
	if !called {
		t.Fatalf("GET should call next handler")
	}
	if got := getRec.Header().Get("Access-Control-Allow-Origin"); got != "http://frontend.local:6180" {
		t.Fatalf("GET allow origin = %q", got)
	}
}

func TestLocalDevCORSRejectsUnknownPreflightAndStaysDisabledOutsideDev(t *testing.T) {
	allowed := map[string]struct{}{
		"http://frontend.local:6180": {},
	}
	handler := withLocalDevCORS("dev", allowed, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	prod := withLocalDevCORS("prod", allowed, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		prodCalled = true
		w.WriteHeader(http.StatusOK)
	}))
	get := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	get.Header.Set("Origin", "http://frontend.local:6180")
	prodRec := httptest.NewRecorder()
	prod.ServeHTTP(prodRec, get)
	if !prodCalled {
		t.Fatalf("non-dev handler should pass through")
	}
	if got := prodRec.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("non-dev CORS header = %q", got)
	}
}

func TestLocalDevCORSOriginsComeFromVerifyBaseURL(t *testing.T) {
	dir := t.TempDir()
	writeAPIFile(t, filepath.Join(dir, "config.yaml"), `
email:
  verifyBaseURL: "http://frontend.local:6180/auth/verify"
`)
	loader, err := config.Load(config.Options{ConfigDir: dir})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	origins := localDevCORSOrigins(loader)
	if _, ok := origins["http://frontend.local:6180"]; !ok {
		t.Fatalf("expected frontend origin from email.verifyBaseURL, got %#v", origins)
	}
	if _, ok := origins["http://127.0.0.1:5173"]; ok {
		t.Fatalf("unexpected hard-coded frontend origin leaked into CORS origins: %#v", origins)
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
		{http.MethodPatch, "/api/v1/resumes/018f2a40-0000-7000-9000-0000000000a1"},
		{http.MethodPost, "/api/v1/resumes/018f2a40-0000-7000-9000-0000000000a1/duplicate"},
		{http.MethodPost, "/api/v1/resumes/018f2a40-0000-7000-9000-0000000000a1/archive"},
		{http.MethodPost, "/api/v1/resumes/018f2a40-0000-7000-9000-0000000000a1/exports"},
		{http.MethodPost, "/api/v1/resume/tailor"},
		{http.MethodGet, "/api/v1/resume/tailor-runs/018f2a40-0000-7000-9000-0000000000c1"},
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

func TestBuildAPIHandlerMountsPracticeRoutesBehindSessionMiddleware(t *testing.T) {
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
	handler := buildAPIHandlerWithUploadReportJobsAndHandlers(
		loader,
		apiRuntimeFlags{},
		service,
		targetjob.NewHandler(),
		practiceRoutes{Handler: apipractice.NewHandler(apipractice.HandlerOptions{})},
		uploadRoutes{},
		resumeRoutes{},
		reportRoutes{},
		jobsRoutes{},
	)

	cases := []struct {
		method string
		path   string
		body   string
	}{
		{http.MethodPost, "/api/v1/practice/plans", `{}`},
		{http.MethodGet, "/api/v1/practice/plans/018f2a40-0000-7000-9000-0000000000a1", ""},
		{http.MethodGet, "/api/v1/practice/sessions", ""},
		{http.MethodPost, "/api/v1/practice/sessions", `{}`},
		{http.MethodGet, "/api/v1/practice/sessions/018f2a40-0000-7000-9000-0000000000b1", ""},
		{http.MethodPost, "/api/v1/practice/sessions/018f2a40-0000-7000-9000-0000000000b1/events", `{}`},
		{http.MethodPost, "/api/v1/practice/sessions/018f2a40-0000-7000-9000-0000000000b1/complete", `{}`},
		{http.MethodPost, "/api/v1/practice/sessions/018f2a40-0000-7000-9000-0000000000b1/voice-turns", `{}`},
	}
	for _, tc := range cases {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
			req.Header.Set("Idempotency-Key", "idem-1")
			handler.ServeHTTP(rec, req)
			if rec.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d body=%s; route is not mounted behind auth middleware", rec.Code, rec.Body.String())
			}
			if !strings.Contains(rec.Body.String(), `"code":"AUTH_UNAUTHORIZED"`) {
				t.Fatalf("expected auth middleware envelope, got %s", rec.Body.String())
			}
		})
	}
}

func TestBuildAPIHandlerDoesNotMountRetiredDebriefOrProfileRoutes(t *testing.T) {
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
	handler := buildAPIHandlerWithUploadReportJobsAndHandlers(
		loader,
		apiRuntimeFlags{},
		service,
		targetjob.NewHandler(),
		practiceRoutes{Handler: apipractice.NewHandler(apipractice.HandlerOptions{})},
		uploadRoutes{},
		resumeRoutes{},
		reportRoutes{},
		jobsRoutes{},
	)

	for _, tc := range []struct {
		method string
		path   string
		body   string
	}{
		{http.MethodPost, "/api/v1/debriefs", `{}`},
		{http.MethodPost, "/api/v1/debriefs/question-suggestions", `{}`},
		{http.MethodGet, "/api/v1/debriefs/018f2a40-0000-7000-9000-0000000000d1", ""},
		{http.MethodGet, "/api/v1/profiles/me", ""},
		{http.MethodPatch, "/api/v1/profiles/me", `{}`},
		{http.MethodGet, "/api/v1/profiles/me/experience-cards", ""},
		{http.MethodPost, "/api/v1/profiles/me/experience-cards", `{}`},
		{http.MethodPatch, "/api/v1/profiles/me/experience-cards/018f2a40-0000-7000-9000-0000000000c1", `{}`},
	} {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Fatalf("%s %s status = %d body=%s; retired route must not be mounted", tc.method, tc.path, rec.Code, rec.Body.String())
		}
	}
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
	sessionTime := time.Date(2026, 6, 13, 10, 45, 0, 0, time.UTC)
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
	targetJobID := "target-1"
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
			Id:          "tailor-run-1",
			Status:      "queued",
			TargetJobId: &targetJobID,
			ResumeId:    "resume-1",
			Suggestions: []api.ResumeTailorBulletSuggestion{},
			CreatedAt:   sessionTime.Format(time.RFC3339),
			UpdatedAt:   sessionTime.Format(time.RFC3339),
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
		if resumeSvc.requestCalls != 1 || resumeSvc.requestIn.UserID != "user-1" || resumeSvc.requestIn.ResumeID != "resume-1" || resumeSvc.requestIn.TargetJobID != "target-1" || resumeSvc.requestIn.Mode != "gap_review" {
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
			ResponseBody:   []byte(`{"tailorRunId":"tailor-run-replay","job":{"id":"job-replay","jobType":"resume_tailor","status":"queued","resourceType":"resume_tailor_run","resourceId":"tailor-run-replay","errorCode":null,"createdAt":"2026-06-13T10:45:00Z","updatedAt":"2026-06-13T10:45:00Z"}}`),
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
		assertAPIStatusCode(t, rec, http.StatusNotFound, sharederrors.CodeResourceNotFound)
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
	handler := buildAPIHandlerWithUploadReportJobsAndHandlers(
		loader,
		apiRuntimeFlags{},
		service,
		targetjob.NewHandler(),
		practiceRoutes{},
		uploadRoutes{},
		resumeRoutes{},
		reportRoutes{},
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

	runtime, err := buildTargetJobRuntime(loader, nil, slog.New(slog.NewTextHandler(io.Discard, nil)), &apiUploadFileDeleter{})
	if err != nil {
		t.Fatalf("buildTargetJobRuntime: %v", err)
	}
	if !runtime.Handles(string(jobs.JobTypePrivacyDelete)) {
		t.Fatalf("runtime does not contribute handler for %s", jobs.JobTypePrivacyDelete)
	}
}

func TestPrivacyDeleteRemovesAccountIdentityAfterJobCompletion(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

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

	runtime, err := buildTargetJobRuntime(loader, db, slog.New(slog.NewTextHandler(io.Discard, nil)), &apiUploadFileDeleter{})
	if err != nil {
		t.Fatalf("buildTargetJobRuntime: %v", err)
	}
	defer runtime.Close()
	handler := runtime.Handlers[string(jobs.JobTypePrivacyDelete)]
	if handler == nil {
		t.Fatalf("privacy delete handler not registered: %+v", runtime.Handlers)
	}

	const (
		requestID = "018f2a40-0000-7000-9000-000000000201"
		userID    = "018f2a40-0000-7000-9000-000000000101"
		email     = "manual-uat-full-funnel@example.test"
	)
	mock.ExpectQuery("from privacy_requests").
		WithArgs(requestID).
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "status"}).AddRow(userID, "queued"))
	mock.ExpectExec("update privacy_requests").
		WithArgs(requestID, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectBegin()
	mock.ExpectQuery("select email from users").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"email"}).AddRow(email))
	mock.ExpectExec("update privacy_requests").
		WithArgs(requestID, sqlmock.AnyArg(), 0, userID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("delete from async_jobs").
		WithArgs(userID).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("delete from resumes").
		WithArgs(userID).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("delete from auth_challenges").
		WithArgs(userID, email).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("delete from users").
		WithArgs(userID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	outcome := handler.Handle(context.Background(), runner.ClaimedJob{
		JobID:        "privacy-job-1",
		JobType:      string(jobs.JobTypePrivacyDelete),
		ResourceType: "privacy_request",
		ResourceID:   requestID,
	})
	if !outcome.Succeeded || outcome.ErrorCode != "" || outcome.Retryable {
		t.Fatalf("privacy delete outcome = %+v", outcome)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestDrainer_DebriefGenerateNotRegistered(t *testing.T) {
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
	if runtime.Handles("debrief_generate") {
		t.Fatalf("runtime must not contribute retired debrief_generate handler")
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
  verifyBaseURL: "http://127.0.0.1:5173/auth/verify"
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

type apiTailorRunService struct {
	requestCalls int
	requestIn    domainresume.RequestTailorRunInput
	requestOut   api.ResumeTailorRunWithJob
	requestErr   error

	getUserID      string
	getTailorRunID string
	getOut         api.ResumeTailorRun
	getErr         error
}

func (s *apiTailorRunService) RegisterResume(context.Context, domainresume.RegisterInput) (api.ResumeWithJob, error) {
	return api.ResumeWithJob{}, errors.New("not implemented")
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

func validAPIRequestTailorBody(mode string) string {
	return `{"targetJobId":"target-1","resumeId":"resume-1","mode":"` + mode + `"}`
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

func (s *apiAuthStore) CreateUserByEmail(context.Context, string, string, string, time.Time) (auth.UserContext, error) {
	return s.user, nil
}

func (s *apiAuthStore) FindUserByEmail(context.Context, string) (auth.UserContext, error) {
	return s.user, nil
}

func (s *apiAuthStore) CompleteUserProfile(context.Context, string, string, time.Time) (auth.UserContext, error) {
	s.user.ProfileCompletionRequired = false
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
