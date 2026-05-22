package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	_ "github.com/lib/pq"

	api "github.com/monshunter/easyinterview/backend/internal/api/generated"
	"github.com/monshunter/easyinterview/backend/internal/auth"
	"github.com/monshunter/easyinterview/backend/internal/jdmatch/generators"
	jdmatchhandler "github.com/monshunter/easyinterview/backend/internal/jdmatch/handler"
	"github.com/monshunter/easyinterview/backend/internal/middleware/idempotency"
	"github.com/monshunter/easyinterview/backend/internal/platform/config"
	"github.com/monshunter/easyinterview/backend/internal/runner"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
	sharedevents "github.com/monshunter/easyinterview/backend/internal/shared/events"
	sharedjobs "github.com/monshunter/easyinterview/backend/internal/shared/jobs"
)

func testLoader(t *testing.T) *config.Loader {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	root := wd
	for i := 0; i < 6; i++ {
		if _, err := os.Stat(filepath.Join(root, "shared", "conventions.yaml")); err == nil {
			break
		}
		root = filepath.Dir(root)
	}
	loader, err := config.LoadCanonical(config.CanonicalOptions{
		AppEnv:    "test",
		ConfigDir: filepath.Join(root, "config"),
	})
	if err != nil {
		t.Skipf("config.LoadCanonical: %v", err)
	}
	return loader
}

func openJDMatchScenarioDB(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://easyinterview:dev@localhost:5432/easyinterview?sslmode=disable"
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := db.Ping(); err != nil {
		t.Skipf("postgres ping failed (%v); skipping jdmatch cmd/api scenario", err)
	}
	return db
}

// TestJDMatchHTTPScenario runs the live cmd/api HTTP matrix required by
// backend-jobs-recommendations/001 §5.6: all 12 JobMatch routes are mounted
// behind SessionMiddleware, the 5 side-effect operations are wrapped by IK,
// cross-user reads fail closed, and replayed side effects do not duplicate
// watchlist / search_run / saved_search rows.
func TestJDMatchHTTPScenario(t *testing.T) {
	if os.Getenv("DATABASE_URL") == "" {
		t.Skip("DATABASE_URL not set; skipping live JD-Match scenario")
	}
	db := openJDMatchScenarioDB(t)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	cleanupJDMatchScenarioEmails(t, db, "jdmatch-a@example.com", "jdmatch-b@example.com")
	authService := auth.NewPasswordlessService(auth.PasswordlessServiceOptions{
		Store:               auth.NewSQLStore(db),
		ChallengePepper:     "jdmatch-test-pepper",
		SessionCookieSecret: "jdmatch-test-session-secret",
	})
	cookieA, userA := loginJDMatchScenarioUser(t, ctx, db, "jdmatch-a@example.com", "Alice Example", "challenge-token-a", "session-token-a")
	cookieB, userB := loginJDMatchScenarioUser(t, ctx, db, "jdmatch-b@example.com", "Bob Example", "challenge-token-b", "session-token-b")
	t.Cleanup(func() { cleanupJDMatchUsers(t, db, userA, userB) })

	const (
		recA1 = "01918fa4-0000-7000-8000-0000000cc101"
		recA2 = "01918fa4-0000-7000-8000-0000000cc102"
	)
	seedJDMatchRecommendation(t, ctx, db, userA, recA1, 92)
	seedJDMatchRecommendation(t, ctx, db, userA, recA2, 78)
	seedJDMatchAgentScan(t, ctx, db, userA)

	// Build the JD-Match runtime with the same wiring helper main()
	// uses, then expose it through the real route helper so auth and IK
	// middleware are exercised instead of being bypassed by direct calls.
	runtime, err := buildJDMatchRuntime(testLoader(t), db, nil, stubJDMatchAI{}, generatorAIAdapter{})
	if err != nil {
		t.Fatalf("buildJDMatchRuntime: %v", err)
	}
	routes := runtime.Routes
	mux := http.NewServeMux()
	addJDMatchRoutes(mux, authService, routes)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	for _, c := range []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/api/v1/jd-match/profile"},
		{http.MethodGet, "/api/v1/jd-match/agent-status"},
		{http.MethodGet, "/api/v1/jd-match/recommendations"},
		{http.MethodGet, "/api/v1/jd-match/recommendations/" + recA1},
		{http.MethodPost, "/api/v1/jd-match/recommendations/" + recA1 + "/dismiss"},
		{http.MethodGet, "/api/v1/jd-match/watchlist"},
		{http.MethodPost, "/api/v1/jd-match/watchlist"},
		{http.MethodDelete, "/api/v1/jd-match/watchlist/" + recA1},
		{http.MethodPost, "/api/v1/jd-match/search"},
		{http.MethodGet, "/api/v1/jd-match/saved-searches"},
		{http.MethodPost, "/api/v1/jd-match/saved-searches"},
		{http.MethodGet, "/api/v1/jd-match/market-signals"},
	} {
		res := doJDMatchRequest(t, srv.URL, c.method, c.path, nil, "", "")
		if res.status != http.StatusUnauthorized {
			t.Fatalf("missing-session %s %s status=%d body=%s", c.method, c.path, res.status, res.body)
		}
	}

	resp := doJDMatchRequest(t, srv.URL, http.MethodGet, "/api/v1/jd-match/profile", cookieA, "", "")
	if resp.status != http.StatusOK {
		t.Fatalf("profile status = %d, want 200; body=%s", resp.status, resp.body)
	}
	var profileBody api.JobMatchProfile
	if err := json.Unmarshal(resp.body, &profileBody); err != nil {
		t.Fatalf("profile decode: %v", err)
	}
	if profileBody.DisplayName == "" {
		t.Fatalf("profile displayName must be non-empty")
	}
	if profileBody.AvatarUrl != nil {
		t.Fatalf("profile avatarUrl must be nil at P0 baseline, got %v", profileBody.AvatarUrl)
	}
	if profileBody.Skills == nil || len(profileBody.Skills) != 0 {
		t.Fatalf("profile skills must be empty []: %#v", profileBody.Skills)
	}

	resp = doJDMatchRequest(t, srv.URL, http.MethodGet, "/api/v1/jd-match/agent-status", cookieA, "", "")
	if resp.status != http.StatusOK {
		t.Fatalf("agent-status status = %d, want 200; body=%s", resp.status, resp.body)
	}

	resp = doJDMatchRequest(t, srv.URL, http.MethodGet, "/api/v1/jd-match/recommendations?pageSize=20", cookieA, "", "")
	if resp.status != http.StatusOK {
		t.Fatalf("recommendations status = %d body=%s", resp.status, resp.body)
	}

	resp = doJDMatchRequest(t, srv.URL, http.MethodGet, "/api/v1/jd-match/recommendations/"+recA1, cookieA, "", "")
	if resp.status != http.StatusOK {
		t.Fatalf("recommendation detail status = %d body=%s", resp.status, resp.body)
	}
	resp = doJDMatchRequest(t, srv.URL, http.MethodGet, "/api/v1/jd-match/recommendations/"+recA1, cookieB, "", "")
	if resp.status != http.StatusNotFound {
		t.Fatalf("cross-user detail status = %d, want 404; body=%s", resp.status, resp.body)
	}

	dismissBody := `{"reason":"wrong_level","freeNote":"too senior for this pass"}`
	resp = doJDMatchRequest(t, srv.URL, http.MethodPost, "/api/v1/jd-match/recommendations/"+recA1+"/dismiss", cookieA, "dismiss-1", dismissBody)
	if resp.status != http.StatusOK {
		t.Fatalf("dismiss status = %d body=%s", resp.status, resp.body)
	}
	resp = doJDMatchRequest(t, srv.URL, http.MethodPost, "/api/v1/jd-match/recommendations/"+recA1+"/dismiss", cookieA, "dismiss-1", dismissBody)
	if resp.status != http.StatusOK || resp.header.Get(idempotency.ReplayHeader) != "true" {
		t.Fatalf("dismiss replay status=%d replay=%q body=%s", resp.status, resp.header.Get(idempotency.ReplayHeader), resp.body)
	}
	assertJDMatchRowCount(t, db, `select count(*) from jd_match_recommendations where user_id = $1 and dismissed_at is not null`, userA, 1)

	resp = doJDMatchRequest(t, srv.URL, http.MethodGet, "/api/v1/jd-match/watchlist", cookieA, "", "")
	if resp.status != http.StatusOK {
		t.Fatalf("watchlist list status = %d body=%s", resp.status, resp.body)
	}
	addWatchBody := `{"jobMatchId":"` + recA2 + `","label":"priority"}`
	resp = doJDMatchRequest(t, srv.URL, http.MethodPost, "/api/v1/jd-match/watchlist", cookieA, "watch-add-1", addWatchBody)
	if resp.status != http.StatusOK {
		t.Fatalf("add watchlist status = %d body=%s", resp.status, resp.body)
	}
	resp = doJDMatchRequest(t, srv.URL, http.MethodPost, "/api/v1/jd-match/watchlist", cookieA, "watch-add-1", addWatchBody)
	if resp.status != http.StatusOK || resp.header.Get(idempotency.ReplayHeader) != "true" {
		t.Fatalf("add watchlist replay status=%d replay=%q body=%s", resp.status, resp.header.Get(idempotency.ReplayHeader), resp.body)
	}
	resp = doJDMatchRequest(t, srv.URL, http.MethodPost, "/api/v1/jd-match/watchlist", cookieA, "watch-add-2", addWatchBody)
	if resp.status != http.StatusOK {
		t.Fatalf("duplicate add watchlist status=%d body=%s", resp.status, resp.body)
	}
	assertJDMatchRowCount(t, db, `select count(*) from watchlist_items where user_id = $1 and linked_job_match_id = $2`, userA, recA2, 1)
	resp = doJDMatchRequest(t, srv.URL, http.MethodPost, "/api/v1/jd-match/watchlist", cookieB, "watch-cross-1", addWatchBody)
	if resp.status != http.StatusNotFound {
		t.Fatalf("cross-user add watchlist status=%d want 404 body=%s", resp.status, resp.body)
	}
	resp = doJDMatchRequest(t, srv.URL, http.MethodDelete, "/api/v1/jd-match/watchlist/"+recA2, cookieA, "watch-remove-1", "")
	if resp.status != http.StatusNoContent {
		t.Fatalf("remove watchlist status=%d body=%s", resp.status, resp.body)
	}
	resp = doJDMatchRequest(t, srv.URL, http.MethodDelete, "/api/v1/jd-match/watchlist/"+recA2, cookieA, "watch-remove-1", "")
	if resp.status != http.StatusNoContent || resp.header.Get(idempotency.ReplayHeader) != "true" {
		t.Fatalf("remove watchlist replay status=%d replay=%q body=%s", resp.status, resp.header.Get(idempotency.ReplayHeader), resp.body)
	}
	assertJDMatchRowCount(t, db, `select count(*) from watchlist_items where user_id = $1 and linked_job_match_id = $2`, userA, recA2, 0)

	searchBody := `{"query":"frontend platform remote","filters":{"remote":true}}`
	resp = doJDMatchRequest(t, srv.URL, http.MethodPost, "/api/v1/jd-match/search", cookieA, "search-1", searchBody)
	if resp.status != http.StatusOK {
		t.Fatalf("search status=%d body=%s", resp.status, resp.body)
	}
	var searchResp struct {
		SearchRunID string `json:"searchRunId"`
		Items       []struct {
			ID string `json:"id"`
		} `json:"items"`
	}
	if err := json.Unmarshal(resp.body, &searchResp); err != nil {
		t.Fatalf("search decode: %v", err)
	}
	resp = doJDMatchRequest(t, srv.URL, http.MethodPost, "/api/v1/jd-match/search", cookieA, "search-1", searchBody)
	if resp.status != http.StatusOK || resp.header.Get(idempotency.ReplayHeader) != "true" {
		t.Fatalf("search replay status=%d replay=%q body=%s", resp.status, resp.header.Get(idempotency.ReplayHeader), resp.body)
	}
	assertJDMatchRowCount(t, db, `select count(*) from jd_match_search_runs where user_id = $1`, userA, 1)
	var searchCompleted sharedevents.JdMatchSearchCompletedPayload
	var searchCompletedRaw []byte
	if err := db.QueryRowContext(ctx, `
select payload
from outbox_events
where event_name = $1 and aggregate_type = 'search_run' and aggregate_id = $2
order by created_at desc
limit 1`,
		string(sharedevents.EventNameJdMatchSearchCompleted),
		searchResp.SearchRunID,
	).Scan(&searchCompletedRaw); err != nil {
		t.Fatalf("read search completed outbox: %v", err)
	}
	if err := json.Unmarshal(searchCompletedRaw, &searchCompleted); err != nil {
		t.Fatalf("decode search completed outbox: %v", err)
	}
	if searchCompleted.UserID != userA || searchCompleted.SearchRunID != searchResp.SearchRunID || searchCompleted.ResultCount != len(searchResp.Items) || searchCompleted.CompletedAt == "" {
		t.Fatalf("search completed outbox drift: %+v", searchCompleted)
	}
	if bytes.Contains(searchCompletedRaw, []byte("frontend platform remote")) || bytes.Contains(searchCompletedRaw, []byte("remote")) {
		t.Fatalf("search completed outbox leaked query/filters: %s", searchCompletedRaw)
	}

	resp = doJDMatchRequest(t, srv.URL, http.MethodGet, "/api/v1/jd-match/saved-searches", cookieA, "", "")
	if resp.status != http.StatusOK {
		t.Fatalf("saved searches list status = %d body=%s", resp.status, resp.body)
	}
	savedBody := `{"label":"frontend remote","query":"frontend platform remote","filters":{"remote":true}}`
	resp = doJDMatchRequest(t, srv.URL, http.MethodPost, "/api/v1/jd-match/saved-searches", cookieA, "saved-1", savedBody)
	if resp.status != http.StatusOK {
		t.Fatalf("create saved search status=%d body=%s", resp.status, resp.body)
	}
	resp = doJDMatchRequest(t, srv.URL, http.MethodPost, "/api/v1/jd-match/saved-searches", cookieA, "saved-1", savedBody)
	if resp.status != http.StatusOK || resp.header.Get(idempotency.ReplayHeader) != "true" {
		t.Fatalf("saved search replay status=%d replay=%q body=%s", resp.status, resp.header.Get(idempotency.ReplayHeader), resp.body)
	}
	assertJDMatchRowCount(t, db, `select count(*) from saved_searches where user_id = $1`, userA, 1)

	resp = doJDMatchRequest(t, srv.URL, http.MethodGet, "/api/v1/jd-match/market-signals?window=7d", cookieA, "", "")
	if resp.status != http.StatusOK {
		t.Fatalf("market-signals status = %d body=%s", resp.status, resp.body)
	}
	var marketBody api.MarketSignalsResponse
	if err := json.Unmarshal(resp.body, &marketBody); err != nil {
		t.Fatalf("market-signals decode: %v", err)
	}
	if len(marketBody.Signals) != 4 {
		t.Fatalf("market-signals must return 4 signals, got %d", len(marketBody.Signals))
	}

	invalidRuntime, err := buildJDMatchRuntime(testLoader(t), db, nil, jdmatchScenarioSearchAI{err: jdmatchhandler.SearchInvalidOutputErr}, generatorAIAdapter{})
	if err != nil {
		t.Fatalf("build invalid search runtime: %v", err)
	}
	invalidMux := http.NewServeMux()
	addJDMatchRoutes(invalidMux, authService, invalidRuntime.Routes)
	invalidSrv := httptest.NewServer(invalidMux)
	defer invalidSrv.Close()
	resp = doJDMatchRequest(t, invalidSrv.URL, http.MethodPost, "/api/v1/jd-match/search", cookieA, "search-invalid-1", searchBody)
	if resp.status != http.StatusBadGateway {
		t.Fatalf("search invalid output status=%d body=%s", resp.status, resp.body)
	}
	var invalidErr api.ApiErrorResponse
	if err := json.Unmarshal(resp.body, &invalidErr); err != nil {
		t.Fatalf("search invalid error decode: %v", err)
	}
	if invalidErr.Error.Code != sharederrors.CodeAiOutputInvalid {
		t.Fatalf("search invalid code=%q want %q", invalidErr.Error.Code, sharederrors.CodeAiOutputInvalid)
	}
	if invalidErr.Error.Retryable {
		t.Fatalf("search invalid output must be non-retryable")
	}
	assertJDMatchRowCount(t, db, `select count(*) from jd_match_search_runs where user_id = $1`, userA, 1)

	if _, err := routes.PrivacyDeleteFunc(ctx, userA); err != nil {
		t.Fatalf("privacy delete: %v", err)
	}
}

func TestJDMatchAgentScanDrainerScenario(t *testing.T) {
	if os.Getenv("DATABASE_URL") == "" {
		t.Skip("DATABASE_URL not set; skipping live JD-Match drainer scenario")
	}
	db := openJDMatchScenarioDB(t)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	const (
		userID = "01918fa4-0000-7000-8000-0000000ee101"
		jobID  = "01918fa4-0000-7000-8000-0000000ee201"
		recID  = "01918fa4-0000-7000-8000-0000000ee301"
	)
	cleanupJDMatchDrainerScenario(t, db, userID, jobID, recID)
	seedJDMatchUser(t, ctx, db, userID, "jdmatch-drainer@example.com", "JDMatch Drainer")
	seedJDMatchRecommendation(t, ctx, db, userID, recID, 77)
	t.Cleanup(func() {
		cleanupJDMatchDrainerScenario(t, db, userID, jobID, recID)
		cleanupJDMatchUsers(t, db, userID)
	})

	ai := &jdmatchScenarioGeneratorAI{body: mustJDMatchScenarioJSON(t, []map[string]any{
		{
			"jobMatchId":          recID,
			"title":               "Backend Platform Engineer",
			"company":             "Acme",
			"companyTag":          "SaaS",
			"level":               "Senior",
			"location":            "Remote",
			"comp":                "$190k",
			"posted":              "today",
			"score":               91,
			"fit":                 map[string]int{"must": 4, "total": 5, "plus": 2, "totalPlus": 3},
			"reasons":             []string{"backend ownership", "runtime systems"},
			"risks":               []string{"fast hiring loop"},
			"highlights":          []string{"platform scale"},
			"interviewHypotheses": []string{},
		},
	})}
	runtime, err := buildJDMatchRuntime(testLoader(t), db, nil, stubJDMatchAI{}, ai)
	if err != nil {
		t.Fatalf("buildJDMatchRuntime: %v", err)
	}

	if _, err := db.ExecContext(ctx, `
insert into async_jobs (
  id, job_type, resource_type, resource_id, dedupe_key, status, payload,
  available_at, created_at, updated_at
) values (
  $1, $2, 'user', $3, $4, 'queued', $5::jsonb, now(), now(), now()
)`,
		jobID,
		string(sharedjobs.JobTypeJdMatchAgentScan),
		userID,
		"jdmatch-agent-scan-"+userID,
		`{"userId":"`+userID+`"}`,
	); err != nil {
		t.Fatalf("seed async job: %v", err)
	}

	kernel := newTestKernel(runner.NewSQLStore(db), runtime.Handlers)
	processed, err := kernel.RunOnce(ctx)
	if err != nil {
		t.Fatalf("RunOnce: %v", err)
	}
	if !processed {
		t.Fatal("RunOnce processed=false, want true")
	}

	var (
		jobStatus string
		attempts  int
		done      bool
		errorCode sql.NullString
		errorMsg  sql.NullString
	)
	if err := db.QueryRowContext(ctx, `
select status, attempts, completed_at is not null, error_code, error_message
from async_jobs
where id = $1`, jobID).Scan(&jobStatus, &attempts, &done, &errorCode, &errorMsg); err != nil {
		t.Fatalf("read async job status: %v", err)
	}
	if jobStatus != "succeeded" || attempts != 1 || !done {
		var scanStatus, scanError sql.NullString
		_ = db.QueryRowContext(ctx, `
select status, error_message
from agent_scans
where user_id = $1
order by created_at desc
limit 1`, userID).Scan(&scanStatus, &scanError)
		t.Fatalf("async job finalized status=%q attempts=%d completed=%v error_code=%q error_message=%q scan_status=%q scan_error=%q", jobStatus, attempts, done, errorCode.String, errorMsg.String, scanStatus.String, scanError.String)
	}

	var recommendationUser string
	var recommendationScore int
	var modelID string
	if err := db.QueryRowContext(ctx, `
select user_id::text, score, model_id
from jd_match_recommendations
where id = $1`, recID).Scan(&recommendationUser, &recommendationScore, &modelID); err != nil {
		t.Fatalf("read recommendation: %v", err)
	}
	if recommendationUser != userID || recommendationScore != 91 || modelID != "jd_match.recommendation.default" {
		t.Fatalf("recommendation drift user=%q score=%d model=%q", recommendationUser, recommendationScore, modelID)
	}

	var (
		scanID              string
		scanStatus          string
		recommendationCount int
		lastScanSet         bool
		nextScanSet         bool
	)
	if err := db.QueryRowContext(ctx, `
select id::text, status, recommendation_count, last_scan_at is not null, next_scan_at is not null
from agent_scans
where user_id = $1
order by created_at desc
limit 1`, userID).Scan(&scanID, &scanStatus, &recommendationCount, &lastScanSet, &nextScanSet); err != nil {
		t.Fatalf("read agent scan: %v", err)
	}
	if scanStatus != "idle" || recommendationCount != 1 || !lastScanSet || !nextScanSet {
		t.Fatalf("agent scan drift status=%q count=%d last=%v next=%v", scanStatus, recommendationCount, lastScanSet, nextScanSet)
	}

	var payload sharedevents.JdMatchRecommendationCompletedPayload
	var rawPayload []byte
	if err := db.QueryRowContext(ctx, `
select payload
from outbox_events
where event_name = $1 and aggregate_type = 'agent_scan' and aggregate_id = $2
order by created_at desc
limit 1`,
		string(sharedevents.EventNameJdMatchRecommendationCompleted),
		scanID,
	).Scan(&rawPayload); err != nil {
		t.Fatalf("read outbox event: %v", err)
	}
	if err := json.Unmarshal(rawPayload, &payload); err != nil {
		t.Fatalf("decode outbox payload: %v", err)
	}
	if payload.UserID != userID || payload.AgentScanID != scanID || payload.RecommendationCount != 1 || payload.CompletedAt == "" {
		t.Fatalf("outbox payload drift: %+v", payload)
	}
	candidateProfile, _ := ai.payload["candidateProfile"].(json.RawMessage)
	jobsPool, _ := ai.payload["jobsPool"].(json.RawMessage)
	if !json.Valid(candidateProfile) || !strings.Contains(string(candidateProfile), "JDMatch Drainer") {
		t.Fatalf("agent_scan candidate profile payload missing runtime context: %s", string(candidateProfile))
	}
	if !json.Valid(jobsPool) || !strings.Contains(string(jobsPool), recID) {
		t.Fatalf("agent_scan jobs pool payload missing jobMatchId seed: %s", string(jobsPool))
	}

	kernel.Start(ctx)
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer shutdownCancel()
	if err := kernel.Shutdown(shutdownCtx); err != nil {
		t.Fatalf("Shutdown: %v", err)
	}
}

type jdmatchHTTPResult struct {
	status int
	header http.Header
	body   []byte
}

type jdmatchScenarioSearchAI struct {
	err error
}

func (s jdmatchScenarioSearchAI) Search(context.Context, string, string, json.RawMessage) (jdmatchhandler.SearchAIResult, error) {
	if s.err != nil {
		return jdmatchhandler.SearchAIResult{}, s.err
	}
	return stubJDMatchAI{}.Search(context.Background(), "", "", nil)
}

func doJDMatchRequest(t *testing.T, baseURL, method, path string, cookie *http.Cookie, ik string, body string) jdmatchHTTPResult {
	t.Helper()
	var reader io.Reader
	if body != "" {
		reader = bytes.NewBufferString(body)
	}
	req, err := http.NewRequest(method, baseURL+path, reader)
	if err != nil {
		t.Fatalf("new request %s %s: %v", method, path, err)
	}
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if cookie != nil {
		req.AddCookie(cookie)
	}
	if ik != "" {
		req.Header.Set(idempotency.HeaderName, ik)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, path, err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read %s %s body: %v", method, path, err)
	}
	return jdmatchHTTPResult{status: resp.StatusCode, header: resp.Header.Clone(), body: raw}
}

func loginJDMatchScenarioUser(t *testing.T, ctx context.Context, db *sql.DB, email, displayName, challengeToken, sessionToken string) (*http.Cookie, string) {
	t.Helper()
	tokenSuffix := strconv.FormatInt(time.Now().UnixNano(), 10)
	challengeToken = challengeToken + "-" + tokenSuffix
	sessionToken = sessionToken + "-" + tokenSuffix
	_, _ = db.ExecContext(ctx, `delete from auth_challenges`)
	sink := auth.NewDevMailSink(auth.DevMailSinkOptions{VerifyBaseURL: "http://api.test/api/v1/auth/email/verify"})
	service := auth.NewPasswordlessService(auth.PasswordlessServiceOptions{
		Store:                 auth.NewSQLStore(db),
		Dispatcher:            auth.NewImmediateMailDispatcher(sink),
		DeliverySecrets:       sink,
		TokenGenerator:        apiFixedTokenGenerator(challengeToken),
		SessionTokenGenerator: apiFixedTokenGenerator(sessionToken),
		ChallengePepper:       "jdmatch-test-pepper",
		SessionCookieSecret:   "jdmatch-test-session-secret",
	})
	if _, err := service.StartEmailChallenge(ctx, auth.StartEmailChallengeInput{
		Email:      email,
		RemoteAddr: "127.0.0.1:12345",
		UserAgent:  "jdmatch-test",
	}); err != nil {
		t.Fatalf("start auth challenge for %s: %v", email, err)
	}
	verified, err := service.VerifyEmailChallenge(ctx, auth.VerifyEmailChallengeInput{
		Token:      challengeToken,
		RemoteAddr: "127.0.0.1:12345",
		UserAgent:  "jdmatch-test",
	})
	if err != nil {
		t.Fatalf("verify auth challenge for %s: %v", email, err)
	}
	if _, err := db.ExecContext(ctx, `update users set display_name = $2 where id = $1`, verified.UserID, displayName); err != nil {
		t.Fatalf("update user display name: %v", err)
	}
	return &http.Cookie{Name: auth.SessionCookieName, Value: verified.SessionToken}, verified.UserID
}

func cleanupJDMatchScenarioEmails(t *testing.T, db *sql.DB, emails ...string) {
	t.Helper()
	for _, email := range emails {
		_, _ = db.Exec(`delete from sessions where user_id in (select id from users where email = $1)`, email)
		_, _ = db.Exec(`delete from auth_challenges where email = $1`, email)
		_, _ = db.Exec(`delete from users where email = $1`, email)
	}
}

type jdmatchScenarioGeneratorAI struct {
	body    []byte
	payload map[string]any
}

func (a *jdmatchScenarioGeneratorAI) Complete(_ context.Context, _ string, payload map[string]any) (generators.CompleteResult, error) {
	a.payload = payload
	return generators.CompleteResult{
		Body:              a.body,
		PromptVersion:     "jd_match_recommendation.v1",
		RubricVersion:     "jd_match_recommendation_rubric.v1",
		ModelProfileName:  "jd_match.recommendation.default",
		Language:          "zh-CN",
		FeatureFlag:       "none",
		DataSourceVersion: "jd_match.v1",
	}, nil
}

func mustJDMatchScenarioJSON(t *testing.T, v any) []byte {
	t.Helper()
	raw, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal scenario JSON: %v", err)
	}
	return raw
}

func cleanupJDMatchDrainerScenario(t *testing.T, db *sql.DB, userID, jobID, recID string) {
	t.Helper()
	_, _ = db.Exec(`delete from outbox_events where event_name = $1 and payload->>'userId' = $2`, string(sharedevents.EventNameJdMatchRecommendationCompleted), userID)
	_, _ = db.Exec(`delete from async_jobs where id = $1 or resource_id = $2`, jobID, userID)
	_, _ = db.Exec(`delete from watchlist_items where linked_job_match_id = $1`, recID)
	_, _ = db.Exec(`delete from jd_match_recommendations where id = $1 or user_id = $2`, recID, userID)
	_, _ = db.Exec(`delete from agent_scans where user_id = $1`, userID)
}

func seedJDMatchUser(t *testing.T, ctx context.Context, db *sql.DB, userID, email, displayName string) {
	t.Helper()
	if _, err := db.ExecContext(ctx, `
insert into users (id, email, display_name, status)
values ($1, $2, $3, 'active')
on conflict (id) do update set email = excluded.email, display_name = excluded.display_name, status = 'active', deleted_at = null`,
		userID, email, displayName); err != nil {
		t.Fatalf("seed user %s: %v", userID, err)
	}
}

func cleanupJDMatchUsers(t *testing.T, db *sql.DB, userIDs ...string) {
	t.Helper()
	ctx := context.Background()
	for _, uid := range userIDs {
		_, _ = db.ExecContext(ctx, `delete from watchlist_items where user_id = $1`, uid)
		_, _ = db.ExecContext(ctx, `delete from saved_searches where user_id = $1`, uid)
		_, _ = db.ExecContext(ctx, `delete from jd_match_search_runs where user_id = $1`, uid)
		_, _ = db.ExecContext(ctx, `delete from jd_match_recommendations where user_id = $1`, uid)
		_, _ = db.ExecContext(ctx, `delete from agent_scans where user_id = $1`, uid)
		_, _ = db.ExecContext(ctx, `delete from candidate_profiles where user_id = $1`, uid)
		_, _ = db.ExecContext(ctx, `delete from user_settings where user_id = $1`, uid)
		_, _ = db.ExecContext(ctx, `delete from users where id = $1`, uid)
	}
}

func seedJDMatchRecommendation(t *testing.T, ctx context.Context, db *sql.DB, userID, recID string, score int) {
	t.Helper()
	_, _ = db.ExecContext(ctx, `delete from watchlist_items where linked_job_match_id = $1`, recID)
	_, _ = db.ExecContext(ctx, `delete from jd_match_recommendations where id = $1`, recID)
	if _, err := db.ExecContext(ctx, `
insert into jd_match_recommendations (
  id, user_id, title, company, company_tag, level, location, comp, posted_label,
  score, fit, reasons, risks, highlights, prompt_version, rubric_version, model_id,
  language, feature_flag, data_source_version, recommended_at, updated_at
) values (
  $1, $2, 'Frontend Platform Engineer', 'Acme', 'SaaS', 'Senior', 'Remote', '$180k', '2d ago',
  $3, '{"must":4,"total":5,"plus":2,"totalPlus":3}'::jsonb, ARRAY['React', 'Platform'], ARRAY['Fast pace'], ARRAY['Design systems'],
  'prompt-v1', 'rubric-v1', 'jd_match.recommendation.default', 'zh-CN', 'none', 'jd_match.v1', now(), now()
)`, recID, userID, score); err != nil {
		t.Fatalf("seed recommendation %s: %v", recID, err)
	}
}

func seedJDMatchAgentScan(t *testing.T, ctx context.Context, db *sql.DB, userID string) {
	t.Helper()
	if _, err := db.ExecContext(ctx, `
insert into agent_scans (id, user_id, status, started_at, finished_at, last_scan_at, next_scan_at, recommendation_count)
values ('01918fa4-0000-7000-8000-0000000dd101', $1, 'idle', now(), now(), now(), now() + interval '1 day', 2)
on conflict (id) do update set user_id = excluded.user_id, status = excluded.status, last_scan_at = excluded.last_scan_at, next_scan_at = excluded.next_scan_at`, userID); err != nil {
		t.Fatalf("seed agent scan: %v", err)
	}
}

func assertJDMatchRowCount(t *testing.T, db *sql.DB, query string, arg1 any, arg2OrWant ...any) {
	t.Helper()
	args := []any{arg1}
	want := 0
	if len(arg2OrWant) == 1 {
		want = arg2OrWant[0].(int)
	} else if len(arg2OrWant) == 2 {
		args = append(args, arg2OrWant[0])
		want = arg2OrWant[1].(int)
	} else {
		t.Fatalf("assertJDMatchRowCount expects one or two trailing args, got %d", len(arg2OrWant))
	}
	var got int
	if err := db.QueryRow(query, args...).Scan(&got); err != nil {
		t.Fatalf("count query failed: %v", err)
	}
	if got != want {
		t.Fatalf("count query got %d, want %d: %s", got, want, query)
	}
}
