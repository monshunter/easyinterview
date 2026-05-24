package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	api "github.com/monshunter/easyinterview/backend/internal/api/generated"
	"github.com/monshunter/easyinterview/backend/internal/auth"
	"github.com/monshunter/easyinterview/backend/internal/middleware/idempotency"
	"github.com/monshunter/easyinterview/backend/internal/platform/config"
	"github.com/monshunter/easyinterview/backend/internal/runner"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
	"github.com/monshunter/easyinterview/backend/internal/targetjob"

	_ "github.com/lib/pq"
)

const (
	fullFunnelSeedEmail      = "full-funnel-seed@example.com"
	fullFunnelAuthPepper     = "full-funnel-test-pepper"
	fullFunnelSessionSecret  = "full-funnel-test-session-secret"
	fullFunnelSeedResumeText = "Full funnel seed resume text with Go, React, and async job ownership."
)

func TestE2EP0FullFunnelReadyResumeSeedUsesRegisterResumeAndRunner(t *testing.T) {
	h := newFullFunnelResumeSeedHarness(t)

	seed := h.seedReadyResume(t)
	if seed.ResumeAssetID == "" || seed.ParseJobID == "" {
		t.Fatalf("seed did not return resumeAssetId and parse job id: %+v", seed)
	}
	h.assertReadyResume(t, seed)
	h.cleanupSeed(t, seed)
	h.assertSeedCleaned(t, seed)
}

type fullFunnelResumeSeed struct {
	UserID        string
	ResumeAssetID string
	ParseJobID    string
}

type fullFunnelResumeSeedHarness struct {
	ctx     context.Context
	db      *sql.DB
	handler http.Handler
	kernel  *runner.Runtime
	cookie  *http.Cookie
	userID  string
}

func newFullFunnelResumeSeedHarness(t *testing.T) *fullFunnelResumeSeedHarness {
	t.Helper()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL not set; skipping full-funnel resume seed scenario")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("open postgres: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	t.Cleanup(cancel)
	if err := db.PingContext(ctx); err != nil {
		t.Skipf("postgres ping failed (%v); skipping full-funnel resume seed scenario", err)
	}

	cleanupFullFunnelScenarioEmail(t, db, fullFunnelSeedEmail)
	cookie, userID := loginFullFunnelScenarioUser(t, ctx, db)
	t.Cleanup(func() { cleanupFullFunnelScenarioUser(t, db, userID) })

	loader := loadFullFunnelResumeSeedConfig(t)
	authService := auth.NewPasswordlessService(auth.PasswordlessServiceOptions{
		Store:               auth.NewSQLStore(db),
		ChallengePepper:     fullFunnelAuthPepper,
		SessionCookieSecret: fullFunnelSessionSecret,
	})
	resumeRuntime, err := buildResumeRuntime(
		loader,
		db,
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		uploadRoutes{},
		&apiNoopAIClient{},
	)
	if err != nil {
		t.Fatalf("buildResumeRuntime: %v", err)
	}
	handler := buildAPIHandlerWithUploadAndHandlers(
		loader,
		apiRuntimeFlags{},
		authService,
		targetjob.NewHandler(),
		practiceRoutes{},
		uploadRoutes{},
		resumeRuntime.Routes(),
	)
	kernel := newTestKernel(runner.NewSQLStore(db), resumeRuntime.Handlers)

	return &fullFunnelResumeSeedHarness{
		ctx:     ctx,
		db:      db,
		handler: handler,
		kernel:  kernel,
		cookie:  cookie,
		userID:  userID,
	}
}

func loadFullFunnelResumeSeedConfig(t *testing.T) *config.Loader {
	t.Helper()
	canonical := loadE2EP0ConfigPreflightConfig(t)
	if canonical.AppEnv() != "test" {
		t.Fatalf("canonical AppEnv=%q, want test", canonical.AppEnv())
	}
	root := scenarioRepoRoot(t)
	dir := t.TempDir()
	writeAPIFile(t, dir+"/config.yaml", `
app:
  env: test
runtime:
  appVersion: "full-funnel-seed-test"
  defaultUiLanguage: zh-CN
auth:
  challengeTokenPepper: "`+fullFunnelAuthPepper+`"
  sessionCookieSecret: "`+fullFunnelSessionSecret+`"
ai:
  providerRegistryPath: "`+root+`/config/ai-providers.yaml"
  modelProfilePath: "`+root+`/config/ai-profiles.yaml"
  promptsDir: "`+root+`/config/prompts"
  rubricsDir: "`+root+`/config/rubrics"
`)
	loader, err := config.Load(config.Options{AppEnv: "test", ConfigDir: dir})
	if err != nil {
		t.Fatalf("load full-funnel seed config: %v", err)
	}
	return loader
}

func (h *fullFunnelResumeSeedHarness) seedReadyResume(t *testing.T) fullFunnelResumeSeed {
	t.Helper()
	raw := h.doJSON(t, http.MethodPost, "/api/v1/resumes", "e2e-p0-098-register-resume", api.RegisterResumeRequest{
		Title:      "Full Funnel Seed Resume",
		Language:   "en",
		SourceType: fullFunnelStringPtr("paste"),
		RawText:    fullFunnelStringPtr(fullFunnelSeedResumeText),
	}, http.StatusAccepted)
	var registered api.ResumeAssetWithJob
	decodeJSON(t, raw, &registered)
	if registered.ResumeAssetId == "" || registered.Job.Id == "" {
		t.Fatalf("registerResume did not return asset/job ids: %+v", registered)
	}
	if registered.Job.JobType != api.JobTypeResumeParse || registered.Job.Status != sharedtypes.JobStatusQueued {
		t.Fatalf("registerResume did not queue resume_parse: %+v", registered.Job)
	}

	processed, err := h.kernel.RunOnce(h.ctx)
	if err != nil {
		t.Fatalf("resume seed RunOnce: %v", err)
	}
	if !processed {
		t.Fatal("resume seed RunOnce processed=false, want true")
	}
	return fullFunnelResumeSeed{UserID: h.userID, ResumeAssetID: registered.ResumeAssetId, ParseJobID: registered.Job.Id}
}

func (h *fullFunnelResumeSeedHarness) assertReadyResume(t *testing.T, seed fullFunnelResumeSeed) {
	t.Helper()
	raw := h.doJSON(t, http.MethodGet, "/api/v1/resumes/"+seed.ResumeAssetID, "", nil, http.StatusOK)
	var detail api.ResumeAsset
	decodeJSON(t, raw, &detail)
	if detail.Id != seed.ResumeAssetID || detail.ParseStatus != sharedtypes.TargetJobParseStatusReady {
		t.Fatalf("seed resume is not ready: %+v", detail)
	}
	if detail.ParsedSummary == nil || len(*detail.ParsedSummary) == 0 {
		t.Fatalf("seed resume missing parsed summary: %+v", detail)
	}
	if detail.ParsedTextSnapshot == nil || *detail.ParsedTextSnapshot != fullFunnelSeedResumeText {
		t.Fatalf("seed resume parsed text snapshot mismatch: %+v", detail.ParsedTextSnapshot)
	}
	var (
		status    string
		attempts  int
		completed bool
	)
	if err := h.db.QueryRowContext(h.ctx, `
select status, attempts, completed_at is not null
from async_jobs
where id = $1`, seed.ParseJobID).Scan(&status, &attempts, &completed); err != nil {
		t.Fatalf("read resume parse job: %v", err)
	}
	if status != "succeeded" || attempts != 1 || !completed {
		t.Fatalf("resume parse job not finalized: status=%q attempts=%d completed=%v", status, attempts, completed)
	}
	assertFullFunnelCount(t, h.db, `
select count(*)
from outbox_events
where aggregate_type = 'resume_asset' and aggregate_id = $1 and event_name = 'resume.parse.completed'`, seed.ResumeAssetID, 1)
}

func (h *fullFunnelResumeSeedHarness) cleanupSeed(t *testing.T, seed fullFunnelResumeSeed) {
	t.Helper()
	cleanupFullFunnelScenarioUser(t, h.db, seed.UserID)
}

func (h *fullFunnelResumeSeedHarness) assertSeedCleaned(t *testing.T, seed fullFunnelResumeSeed) {
	t.Helper()
	assertFullFunnelCount(t, h.db, `select count(*) from users where id = $1`, seed.UserID, 0)
	assertFullFunnelCount(t, h.db, `select count(*) from resume_assets where id = $1`, seed.ResumeAssetID, 0)
	assertFullFunnelCount(t, h.db, `select count(*) from async_jobs where id = $1 or resource_id = $2`, seed.ParseJobID, seed.ResumeAssetID, 0)
	assertFullFunnelCount(t, h.db, `select count(*) from outbox_events where aggregate_id = $1`, seed.ResumeAssetID, 0)
	assertFullFunnelCount(t, h.db, `select count(*) from idempotency_records where user_id = $1`, seed.UserID, 0)
}

func (h *fullFunnelResumeSeedHarness) doJSON(t *testing.T, method, path, idempotencyKey string, body any, wantStatus int) []byte {
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
		req.Header.Set(idempotency.HeaderName, idempotencyKey)
	}
	rec := httptest.NewRecorder()
	h.handler.ServeHTTP(rec, req)
	if rec.Code != wantStatus {
		t.Fatalf("%s %s status=%d want=%d body=%s", method, path, rec.Code, wantStatus, rec.Body.String())
	}
	return rec.Body.Bytes()
}

func loginFullFunnelScenarioUser(t *testing.T, ctx context.Context, db *sql.DB) (*http.Cookie, string) {
	t.Helper()
	tokenSuffix := time.Now().UTC().Format("20060102150405.000000000")
	challengeToken := "full-funnel-challenge-" + tokenSuffix
	sessionToken := "full-funnel-session-" + tokenSuffix
	sink := auth.NewDevMailSink(auth.DevMailSinkOptions{VerifyBaseURL: "http://api.test/api/v1/auth/email/verify"})
	service := auth.NewPasswordlessService(auth.PasswordlessServiceOptions{
		Store:                 auth.NewSQLStore(db),
		Dispatcher:            auth.NewImmediateMailDispatcher(sink),
		DeliverySecrets:       sink,
		TokenGenerator:        apiFixedTokenGenerator(challengeToken),
		SessionTokenGenerator: apiFixedTokenGenerator(sessionToken),
		ChallengePepper:       fullFunnelAuthPepper,
		SessionCookieSecret:   fullFunnelSessionSecret,
	})
	if _, err := service.StartEmailChallenge(ctx, auth.StartEmailChallengeInput{
		Email:      fullFunnelSeedEmail,
		RemoteAddr: "127.0.0.1:12345",
		UserAgent:  "full-funnel-seed-test",
	}); err != nil {
		t.Fatalf("start full-funnel auth challenge: %v", err)
	}
	verified, err := service.VerifyEmailChallenge(ctx, auth.VerifyEmailChallengeInput{
		Token:      challengeToken,
		RemoteAddr: "127.0.0.1:12345",
		UserAgent:  "full-funnel-seed-test",
	})
	if err != nil {
		t.Fatalf("verify full-funnel auth challenge: %v", err)
	}
	return &http.Cookie{Name: auth.SessionCookieName, Value: verified.SessionToken}, verified.UserID
}

func cleanupFullFunnelScenarioEmail(t *testing.T, db *sql.DB, email string) {
	t.Helper()
	rows, err := db.Query(`select id::text from users where email = $1`, email)
	if err != nil {
		t.Fatalf("query stale full-funnel user: %v", err)
	}
	defer rows.Close()
	var userIDs []string
	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			t.Fatalf("scan stale full-funnel user: %v", err)
		}
		userIDs = append(userIDs, userID)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate stale full-funnel users: %v", err)
	}
	for _, userID := range userIDs {
		cleanupFullFunnelScenarioUser(t, db, userID)
	}
	_, _ = db.Exec(`delete from auth_challenges where email = $1`, email)
}

func cleanupFullFunnelScenarioUser(t *testing.T, db *sql.DB, userID string) {
	t.Helper()
	if userID == "" {
		return
	}
	_, _ = db.Exec(`
delete from outbox_events
where aggregate_id in (select id from resume_assets where user_id = $1)`, userID)
	_, _ = db.Exec(`
delete from async_jobs
where resource_type = 'resume_asset'
  and resource_id in (select id from resume_assets where user_id = $1)`, userID)
	_, _ = db.Exec(`delete from idempotency_records where user_id = $1`, userID)
	_, _ = db.Exec(`delete from auth_challenges where user_id = $1`, userID)
	_, _ = db.Exec(`delete from sessions where user_id = $1`, userID)
	_, _ = db.Exec(`delete from resume_assets where user_id = $1`, userID)
	_, _ = db.Exec(`delete from users where id = $1`, userID)
}

func assertFullFunnelCount(t *testing.T, db *sql.DB, query string, argsAndWant ...any) {
	t.Helper()
	if len(argsAndWant) < 1 {
		t.Fatal("assertFullFunnelCount requires expected count")
	}
	want, ok := argsAndWant[len(argsAndWant)-1].(int)
	if !ok {
		t.Fatalf("last assertFullFunnelCount argument must be int, got %T", argsAndWant[len(argsAndWant)-1])
	}
	var got int
	if err := db.QueryRow(query, argsAndWant[:len(argsAndWant)-1]...).Scan(&got); err != nil {
		t.Fatalf("count query failed: %v", err)
	}
	if got != want {
		t.Fatalf("count query got %d, want %d: %s", got, want, query)
	}
}

func fullFunnelStringPtr(v string) *string {
	return &v
}
