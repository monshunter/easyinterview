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
	"reflect"
	"testing"
	"time"

	"github.com/lib/pq"

	"github.com/monshunter/easyinterview/backend/internal/auth"
	jdmatchhandler "github.com/monshunter/easyinterview/backend/internal/jdmatch/handler"
)

func TestJDMatchFixtureParity(t *testing.T) {
	if os.Getenv("DATABASE_URL") == "" {
		t.Skip("DATABASE_URL not set; skipping live JD-Match fixture parity scenario")
	}
	db := openJDMatchScenarioDB(t)
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()

	cleanupJDMatchScenarioEmails(t, db, "jdmatch-fixture@example.com")
	authService := authServiceForJDMatchFixture(db)
	cookie, userID := loginJDMatchScenarioUser(t, ctx, db, "jdmatch-fixture@example.com", "Alice Example", "fixture-challenge-token", "fixture-session-token")
	t.Cleanup(func() { cleanupJDMatchUsers(t, db, userID) })

	clock := &jdmatchFixtureClock{now: time.Date(2026, 5, 10, 9, 0, 0, 0, time.UTC)}
	ids := &jdmatchFixtureIDs{}
	routes := buildJDMatchRoutesWithOptions(testLoader(t), db, jdmatchFixtureSearchAI{}, jdmatchRouteOptions{
		Now:   clock.Now,
		NewID: ids.Next,
	})
	mux := http.NewServeMux()
	addJDMatchRoutes(mux, authService, routes)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	t.Run("profile structural parity", func(t *testing.T) {
		cleanupJDMatchFixtureData(t, db, userID)
		resp := requestJDMatchFixture(t, srv.URL, http.MethodGet, "/api/v1/jd-match/profile", cookie, "", nil, fixtureRequestID(t, "getJobMatchProfile"))
		assertJDMatchProfileStructuralParity(t, "getJobMatchProfile", resp)
	})

	t.Run("agent status", func(t *testing.T) {
		cleanupJDMatchFixtureData(t, db, userID)
		seedJDMatchFixtureAgentScan(t, ctx, db, userID)
		resp := requestJDMatchFixture(t, srv.URL, http.MethodGet, "/api/v1/jd-match/agent-status", cookie, "", nil, fixtureRequestID(t, "getAgentScanStatus"))
		assertJDMatchFixtureResponse(t, "getAgentScanStatus", resp)
	})

	t.Run("recommendations list", func(t *testing.T) {
		cleanupJDMatchFixtureData(t, db, userID)
		seedJDMatchFixtureRecommendations(t, ctx, db, userID, fixtureItems(t, "listJobRecommendations"))
		seedJDMatchFixtureWatchlistRow(t, ctx, db, userID, "01918fa0-0000-7000-8000-00000000b002", "01918fa0-0000-7000-8000-00000000a002", nil, nil, "2026-05-06T11:30:00Z")
		resp := requestJDMatchFixture(t, srv.URL, http.MethodGet, "/api/v1/jd-match/recommendations?pageSize=20", cookie, "", nil, fixtureRequestID(t, "listJobRecommendations"))
		assertJDMatchFixtureResponse(t, "listJobRecommendations", resp)
	})

	t.Run("recommendation detail", func(t *testing.T) {
		cleanupJDMatchFixtureData(t, db, userID)
		seedJDMatchFixtureRecommendation(t, ctx, db, userID, fixtureBody(t, "getJobRecommendation"), 0)
		resp := requestJDMatchFixture(t, srv.URL, http.MethodGet, "/api/v1/jd-match/recommendations/01918fa0-0000-7000-8000-00000000a001", cookie, "", nil, fixtureRequestID(t, "getJobRecommendation"))
		assertJDMatchFixtureResponse(t, "getJobRecommendation", resp)
	})

	t.Run("dismiss", func(t *testing.T) {
		cleanupJDMatchFixtureData(t, db, userID)
		clock.now = time.Date(2026, 5, 10, 9, 0, 0, 0, time.UTC)
		seedJDMatchFixtureRecommendation(t, ctx, db, userID, map[string]any{
			"id": "01918fa0-0000-7000-8000-00000000a002", "title": "Staff Frontend Engineer · Platform", "company": "Lumen Labs", "location": "Remote · APAC", "posted": "5 days ago", "score": float64(78),
			"fit": map[string]any{"must": float64(3), "total": float64(5), "plus": float64(2), "totalPlus": float64(4)}, "reasons": []any{}, "risks": []any{}, "highlights": []any{}, "provenance": fixtureProvenance(),
		}, 0)
		req := fixtureRequest(t, "markJobNotRelevant")
		resp := requestJDMatchFixture(t, srv.URL, http.MethodPost, "/api/v1/jd-match/recommendations/01918fa0-0000-7000-8000-00000000a002/dismiss", cookie, req.IdempotencyKey, req.Body, fixtureRequestID(t, "markJobNotRelevant"))
		assertJDMatchFixtureResponse(t, "markJobNotRelevant", resp)
	})

	t.Run("watchlist list", func(t *testing.T) {
		cleanupJDMatchFixtureData(t, db, userID)
		for _, item := range fixtureItems(t, "listWatchlist") {
			item := item.(map[string]any)
			seedJDMatchFixtureRecommendationFromWatchlist(t, ctx, db, userID, item)
			seedJDMatchFixtureWatchlist(t, ctx, db, userID, item)
		}
		resp := requestJDMatchFixture(t, srv.URL, http.MethodGet, "/api/v1/jd-match/watchlist", cookie, "", nil, fixtureRequestID(t, "listWatchlist"))
		assertJDMatchFixtureResponse(t, "listWatchlist", resp)
	})

	t.Run("watchlist add", func(t *testing.T) {
		cleanupJDMatchFixtureData(t, db, userID)
		clock.now = time.Date(2026, 5, 10, 9, 0, 0, 0, time.UTC)
		ids.Reset("01918fa0-0000-7000-8000-00000000b001")
		seedJDMatchFixtureRecommendationFromWatchlist(t, ctx, db, userID, fixtureBody(t, "addToWatchlist"))
		req := fixtureRequest(t, "addToWatchlist")
		resp := requestJDMatchFixture(t, srv.URL, http.MethodPost, "/api/v1/jd-match/watchlist", cookie, req.IdempotencyKey, req.Body, fixtureRequestID(t, "addToWatchlist"))
		assertJDMatchFixtureResponse(t, "addToWatchlist", resp)
	})

	t.Run("watchlist remove", func(t *testing.T) {
		cleanupJDMatchFixtureData(t, db, userID)
		seedJDMatchFixtureRecommendationFromWatchlist(t, ctx, db, userID, map[string]any{"linkedJobMatchId": "01918fa0-0000-7000-8000-00000000a001", "title": "Senior Frontend Engineer · Design Systems", "company": "Acme", "tone": "ok"})
		seedJDMatchFixtureWatchlistRow(t, ctx, db, userID, "01918fa0-0000-7000-8000-00000000b001", "01918fa0-0000-7000-8000-00000000a001", nil, nil, "2026-05-08T10:00:00Z")
		req := fixtureRequest(t, "removeFromWatchlist")
		resp := requestJDMatchFixture(t, srv.URL, http.MethodDelete, "/api/v1/jd-match/watchlist/01918fa0-0000-7000-8000-00000000a001", cookie, req.IdempotencyKey, nil, fixtureRequestID(t, "removeFromWatchlist"))
		assertJDMatchFixtureResponse(t, "removeFromWatchlist", resp)
	})

	t.Run("search", func(t *testing.T) {
		cleanupJDMatchFixtureData(t, db, userID)
		ids.Reset("01918fa0-0000-7000-8000-00000000c100", "01918fa0-0000-7000-8000-00000000c101")
		seedJDMatchFixtureRecommendations(t, ctx, db, userID, fixtureItems(t, "searchJobs"))
		req := fixtureRequest(t, "searchJobs")
		resp := requestJDMatchFixture(t, srv.URL, http.MethodPost, "/api/v1/jd-match/search", cookie, req.IdempotencyKey, req.Body, fixtureRequestID(t, "searchJobs"))
		assertJDMatchFixtureResponse(t, "searchJobs", resp)
	})

	t.Run("saved searches list", func(t *testing.T) {
		cleanupJDMatchFixtureData(t, db, userID)
		seedJDMatchFixtureSavedSearches(t, ctx, db, userID, fixtureItems(t, "listSavedSearches"))
		resp := requestJDMatchFixture(t, srv.URL, http.MethodGet, "/api/v1/jd-match/saved-searches", cookie, "", nil, fixtureRequestID(t, "listSavedSearches"))
		assertJDMatchFixtureResponse(t, "listSavedSearches", resp)
	})

	t.Run("saved search create", func(t *testing.T) {
		cleanupJDMatchFixtureData(t, db, userID)
		clock.now = time.Date(2026, 5, 10, 9, 0, 0, 0, time.UTC)
		ids.Reset("01918fa0-0000-7000-8000-00000000c004")
		req := fixtureRequest(t, "createSavedSearch")
		resp := requestJDMatchFixture(t, srv.URL, http.MethodPost, "/api/v1/jd-match/saved-searches", cookie, req.IdempotencyKey, req.Body, fixtureRequestID(t, "createSavedSearch"))
		assertJDMatchFixtureResponse(t, "createSavedSearch", resp)
	})

	t.Run("market signals", func(t *testing.T) {
		cleanupJDMatchFixtureData(t, db, userID)
		clock.now = time.Date(2026, 5, 10, 5, 0, 0, 0, time.UTC)
		resp := requestJDMatchFixture(t, srv.URL, http.MethodGet, "/api/v1/jd-match/market-signals?window=7d", cookie, "", nil, fixtureRequestID(t, "getMarketSignals"))
		assertJDMatchFixtureResponse(t, "getMarketSignals", resp)
	})
}

type jdmatchFixtureClock struct{ now time.Time }

func (c *jdmatchFixtureClock) Now() time.Time { return c.now }

type jdmatchFixtureIDs struct{ ids []string }

func (s *jdmatchFixtureIDs) Reset(ids ...string) { s.ids = append([]string{}, ids...) }

func (s *jdmatchFixtureIDs) Next() string {
	if len(s.ids) == 0 {
		return "01918fa0-0000-7000-8000-00000000ffff"
	}
	id := s.ids[0]
	s.ids = s.ids[1:]
	return id
}

type jdmatchFixtureSearchAI struct{}

func (jdmatchFixtureSearchAI) Search(context.Context, string, string, json.RawMessage) (jdmatchhandler.SearchAIResult, error) {
	return jdmatchhandler.SearchAIResult{
		MatchedJobMatchIDs: []string{"01918fa0-0000-7000-8000-00000000a001", "01918fa0-0000-7000-8000-00000000a002"},
		PromptVersion:      "jd_match_search.v1",
		RubricVersion:      "jd_match_search_rubric.v1",
		ModelProfileName:   "model-profile:contract.default",
		Language:           "zh-CN",
		FeatureFlag:        "none",
		DataSourceVersion:  "jd_match.v1",
	}, nil
}

func authServiceForJDMatchFixture(db *sql.DB) *auth.PasswordlessService {
	return auth.NewPasswordlessService(auth.PasswordlessServiceOptions{
		Store:               auth.NewSQLStore(db),
		ChallengePepper:     "jdmatch-test-pepper",
		SessionCookieSecret: "jdmatch-test-session-secret",
	})
}

type jdmatchFixtureHTTPResult struct {
	Status int
	Header http.Header
	Body   []byte
}

func requestJDMatchFixture(t *testing.T, baseURL, method, path string, cookie *http.Cookie, ik string, body []byte, requestID string) jdmatchFixtureHTTPResult {
	t.Helper()
	var reader io.Reader
	if len(body) > 0 {
		reader = bytes.NewReader(body)
	}
	req, err := http.NewRequest(method, baseURL+path, reader)
	if err != nil {
		t.Fatalf("new request %s %s: %v", method, path, err)
	}
	if len(body) > 0 {
		req.Header.Set("Content-Type", "application/json")
	}
	if cookie != nil {
		req.AddCookie(cookie)
	}
	if ik != "" {
		req.Header.Set("Idempotency-Key", ik)
	}
	if requestID != "" {
		req.Header.Set("X-Request-ID", requestID)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, path, err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read response body: %v", err)
	}
	return jdmatchFixtureHTTPResult{Status: resp.StatusCode, Header: resp.Header.Clone(), Body: raw}
}

type jdmatchFixtureRequest struct {
	IdempotencyKey string
	Body           []byte
}

func fixtureRequest(t *testing.T, operationID string) jdmatchFixtureRequest {
	t.Helper()
	defaultScenario := fixtureDefaultScenario(t, operationID)
	var out jdmatchFixtureRequest
	if headers, _ := defaultScenario["request"].(map[string]any)["headers"].(map[string]any); headers != nil {
		if raw, _ := headers["Idempotency-Key"].(string); raw != "" {
			out.IdempotencyKey = raw
		}
	}
	if body, ok := defaultScenario["request"].(map[string]any)["body"]; ok {
		raw, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("%s request body marshal: %v", operationID, err)
		}
		out.Body = raw
	}
	return out
}

func fixtureRequestID(t *testing.T, operationID string) string {
	t.Helper()
	resp := fixtureResponse(t, operationID)
	headers, _ := resp["headers"].(map[string]any)
	requestID, _ := headers["X-Request-ID"].(string)
	return requestID
}

func fixtureBody(t *testing.T, operationID string) map[string]any {
	t.Helper()
	body, _ := fixtureResponse(t, operationID)["body"].(map[string]any)
	if body == nil {
		t.Fatalf("%s fixture body is not an object", operationID)
	}
	return body
}

func fixtureItems(t *testing.T, operationID string) []any {
	t.Helper()
	items, _ := fixtureBody(t, operationID)["items"].([]any)
	if items == nil {
		t.Fatalf("%s fixture body missing items", operationID)
	}
	return items
}

func fixtureProvenance() map[string]any {
	return map[string]any{
		"promptVersion":     "jd_match_recommendation.v1",
		"rubricVersion":     "jd_match_recommendation_rubric.v1",
		"modelId":           "model-profile:contract.default",
		"language":          "zh-CN",
		"featureFlag":       "none",
		"dataSourceVersion": "jd_match.v1",
	}
}

func fixtureResponse(t *testing.T, operationID string) map[string]any {
	t.Helper()
	resp, _ := fixtureDefaultScenario(t, operationID)["response"].(map[string]any)
	if resp == nil {
		t.Fatalf("%s fixture missing default response", operationID)
	}
	return resp
}

func fixtureDefaultScenario(t *testing.T, operationID string) map[string]any {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(repoRootForJDMatchFixture(t), "openapi", "fixtures", "JobMatch", operationID+".json"))
	if err != nil {
		t.Fatalf("read %s fixture: %v", operationID, err)
	}
	var fixture map[string]any
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatalf("decode %s fixture: %v", operationID, err)
	}
	scenarios, _ := fixture["scenarios"].(map[string]any)
	defaultScenario, _ := scenarios["default"].(map[string]any)
	if defaultScenario == nil {
		t.Fatalf("%s fixture missing default scenario", operationID)
	}
	return defaultScenario
}

func repoRootForJDMatchFixture(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	root := wd
	for i := 0; i < 8; i++ {
		if _, err := os.Stat(filepath.Join(root, "openapi", "fixtures", "JobMatch")); err == nil {
			return root
		}
		root = filepath.Dir(root)
	}
	t.Fatalf("repo root not found from %s", wd)
	return ""
}

func assertJDMatchFixtureResponse(t *testing.T, operationID string, got jdmatchFixtureHTTPResult) {
	t.Helper()
	resp := fixtureResponse(t, operationID)
	wantStatus := int(resp["status"].(float64))
	if got.Status != wantStatus {
		t.Fatalf("%s status=%d want=%d body=%s", operationID, got.Status, wantStatus, got.Body)
	}
	wantRequestID := fixtureRequestID(t, operationID)
	if got.Header.Get("X-Request-ID") != wantRequestID {
		t.Fatalf("%s X-Request-ID=%q want=%q", operationID, got.Header.Get("X-Request-ID"), wantRequestID)
	}
	wantBody, hasBody := resp["body"]
	if !hasBody || wantBody == nil {
		if len(got.Body) != 0 {
			t.Fatalf("%s body=%s want empty", operationID, got.Body)
		}
		return
	}
	wantRaw, err := json.Marshal(wantBody)
	if err != nil {
		t.Fatalf("%s marshal fixture body: %v", operationID, err)
	}
	assertJSONSemanticallyEqual(t, operationID, wantRaw, got.Body)
}

func assertJDMatchProfileStructuralParity(t *testing.T, operationID string, got jdmatchFixtureHTTPResult) {
	t.Helper()
	resp := fixtureResponse(t, operationID)
	wantStatus := int(resp["status"].(float64))
	if got.Status != wantStatus {
		t.Fatalf("%s status=%d want=%d body=%s", operationID, got.Status, wantStatus, got.Body)
	}
	wantRequestID := fixtureRequestID(t, operationID)
	if got.Header.Get("X-Request-ID") != wantRequestID {
		t.Fatalf("%s X-Request-ID=%q want=%q", operationID, got.Header.Get("X-Request-ID"), wantRequestID)
	}
	var body map[string]any
	if err := json.Unmarshal(got.Body, &body); err != nil {
		t.Fatalf("%s decode body: %v", operationID, err)
	}
	for _, key := range []string{"displayName", "skills", "sources"} {
		if _, ok := body[key]; !ok {
			t.Fatalf("%s missing required field %q in %s", operationID, key, got.Body)
		}
	}
	if _, ok := body["skills"].([]any); !ok {
		t.Fatalf("%s skills must be an array: %s", operationID, got.Body)
	}
	sources, ok := body["sources"].(map[string]any)
	if !ok {
		t.Fatalf("%s sources must be an object: %s", operationID, got.Body)
	}
	for _, key := range []string{"resumes", "jds", "mocks", "debriefs"} {
		if _, ok := sources[key].(float64); !ok {
			t.Fatalf("%s sources.%s must be numeric: %s", operationID, key, got.Body)
		}
	}
}

func assertJSONSemanticallyEqual(t *testing.T, label string, wantRaw, gotRaw []byte) {
	t.Helper()
	var want any
	var got any
	if err := json.Unmarshal(wantRaw, &want); err != nil {
		t.Fatalf("%s decode want: %v", label, err)
	}
	if err := json.Unmarshal(gotRaw, &got); err != nil {
		t.Fatalf("%s decode got: %v body=%s", label, err, gotRaw)
	}
	if !reflect.DeepEqual(want, got) {
		t.Fatalf("%s JSON mismatch\ngot:  %s\nwant: %s", label, gotRaw, wantRaw)
	}
}

func cleanupJDMatchFixtureData(t *testing.T, db *sql.DB, userID string) {
	t.Helper()
	_, _ = db.Exec(`delete from idempotency_records where user_id = $1`, userID)
	_, _ = db.Exec(`delete from outbox_events where payload->>'userId' = $1`, userID)
	_, _ = db.Exec(`delete from async_jobs where resource_id = $1`, userID)
	_, _ = db.Exec(`delete from watchlist_items where user_id = $1`, userID)
	_, _ = db.Exec(`delete from saved_searches where user_id = $1`, userID)
	_, _ = db.Exec(`delete from jd_match_search_runs where user_id = $1`, userID)
	_, _ = db.Exec(`delete from jd_match_recommendations where user_id = $1`, userID)
	_, _ = db.Exec(`delete from agent_scans where user_id = $1`, userID)
}

func seedJDMatchFixtureAgentScan(t *testing.T, ctx context.Context, db *sql.DB, userID string) {
	t.Helper()
	if _, err := db.ExecContext(ctx, `
insert into agent_scans (id, user_id, status, started_at, finished_at, last_scan_at, next_scan_at, recommendation_count, created_at, updated_at)
values ('01918fa0-0000-7000-8000-00000000d001', $1, 'idle', '2026-05-10T05:00:00Z', '2026-05-10T05:00:00Z', '2026-05-10T05:00:00Z', '2026-05-10T13:00:00Z', 3, now(), now())`, userID); err != nil {
		t.Fatalf("seed agent scan: %v", err)
	}
}

func seedJDMatchFixtureRecommendations(t *testing.T, ctx context.Context, db *sql.DB, userID string, items []any) {
	t.Helper()
	for i, raw := range items {
		item := raw.(map[string]any)
		seedJDMatchFixtureRecommendation(t, ctx, db, userID, item, i)
	}
}

func seedJDMatchFixtureRecommendation(t *testing.T, ctx context.Context, db *sql.DB, userID string, item map[string]any, order int) {
	t.Helper()
	fit := item["fit"].(map[string]any)
	provenance, _ := item["provenance"].(map[string]any)
	if provenance == nil {
		provenance = fixtureProvenance()
	}
	if _, err := db.ExecContext(ctx, `
insert into jd_match_recommendations (
  id, user_id, title, company, company_tag, level, location, comp, posted_label, score,
  fit, reasons, risks, highlights, seen, source_url, source_label, network_note,
  similar_interviewers, interview_hypotheses, prompt_version, rubric_version, model_id,
  language, feature_flag, data_source_version, recommended_at, updated_at
) values (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
  $11::jsonb, $12, $13, $14, $15, $16, $17, $18,
  $19, $20, $21, $22, $23, $24, $25, $26, $27, $27
) on conflict (id) do update set
  user_id = excluded.user_id, title = excluded.title, company = excluded.company,
  company_tag = excluded.company_tag, level = excluded.level, location = excluded.location,
  comp = excluded.comp, posted_label = excluded.posted_label, score = excluded.score,
  fit = excluded.fit, reasons = excluded.reasons, risks = excluded.risks, highlights = excluded.highlights,
  seen = excluded.seen, source_url = excluded.source_url, source_label = excluded.source_label,
  network_note = excluded.network_note, similar_interviewers = excluded.similar_interviewers,
  interview_hypotheses = excluded.interview_hypotheses, prompt_version = excluded.prompt_version,
  rubric_version = excluded.rubric_version, model_id = excluded.model_id, language = excluded.language,
  feature_flag = excluded.feature_flag, data_source_version = excluded.data_source_version,
  recommended_at = excluded.recommended_at, updated_at = excluded.updated_at, dismissed_at = null, deleted_at = null`,
		stringField(item, "id"), userID, stringField(item, "title"), stringField(item, "company"),
		nullableString(item["companyTag"]), nullableString(item["level"]), stringField(item, "location"),
		nullableString(item["comp"]), nullableString(item["posted"]), intField(item, "score"),
		mustJSONRaw(t, fit), pq.Array(stringSlice(item["reasons"])), pq.Array(stringSlice(item["risks"])), pq.Array(stringSlice(item["highlights"])),
		boolField(item, "seen"), nullableString(item["sourceUrl"]), nullableString(item["sourceLabel"]), nullableString(item["networkNote"]),
		nullableInt(item["similarInterviewers"]), pq.Array(stringSlice(item["interviewHypotheses"])),
		stringField(provenance, "promptVersion"), stringField(provenance, "rubricVersion"), stringField(provenance, "modelId"),
		stringField(provenance, "language"), stringField(provenance, "featureFlag"), stringField(provenance, "dataSourceVersion"),
		time.Date(2026, 5, 10, 9, 0-order, 0, 0, time.UTC),
	); err != nil {
		t.Fatalf("seed recommendation %s: %v", stringField(item, "id"), err)
	}
}

func seedJDMatchFixtureRecommendationFromWatchlist(t *testing.T, ctx context.Context, db *sql.DB, userID string, item map[string]any) {
	t.Helper()
	score := 45
	switch item["tone"] {
	case "ok":
		score = 92
	case "warn":
		score = 78
	}
	seedJDMatchFixtureRecommendation(t, ctx, db, userID, map[string]any{
		"id": item["linkedJobMatchId"], "title": item["title"], "company": item["company"], "location": "Remote", "posted": "today", "score": float64(score),
		"fit": map[string]any{"must": float64(1), "total": float64(1), "plus": float64(0), "totalPlus": float64(0)}, "reasons": []any{}, "risks": []any{}, "highlights": []any{}, "provenance": fixtureProvenance(),
	}, 0)
}

func seedJDMatchFixtureWatchlist(t *testing.T, ctx context.Context, db *sql.DB, userID string, item map[string]any) {
	t.Helper()
	seedJDMatchFixtureWatchlistRow(t, ctx, db, userID, stringField(item, "id"), stringField(item, "linkedJobMatchId"), nullableString(item["label"]), nullableString(item["change"]), stringField(item, "addedAt"))
}

func seedJDMatchFixtureWatchlistRow(t *testing.T, ctx context.Context, db *sql.DB, userID, id, linkedID string, label, change any, addedAt string) {
	t.Helper()
	if _, err := db.ExecContext(ctx, `
insert into watchlist_items (id, user_id, linked_job_match_id, label, change_note, added_at)
values ($1, $2, $3, $4, $5, $6::timestamptz)
on conflict (user_id, linked_job_match_id) do update set label = excluded.label, change_note = excluded.change_note, added_at = excluded.added_at`,
		id, userID, linkedID, label, change, addedAt); err != nil {
		t.Fatalf("seed watchlist: %v", err)
	}
}

func seedJDMatchFixtureSavedSearches(t *testing.T, ctx context.Context, db *sql.DB, userID string, items []any) {
	t.Helper()
	for _, raw := range items {
		item := raw.(map[string]any)
		filters := item["filters"]
		if filters == nil {
			filters = map[string]any{}
		}
		if _, err := db.ExecContext(ctx, `
insert into saved_searches (id, user_id, label, query, filters, new_jobs_count, last_run_at, created_at, updated_at)
values ($1, $2, $3, $4, $5::jsonb, $6, $7, $8::timestamptz, $8::timestamptz)`,
			stringField(item, "id"), userID, stringField(item, "label"), stringField(item, "query"), mustJSONRaw(t, filters),
			nullableInt(item["newJobsCount"]), nullableTime(item["lastRunAt"]), stringField(item, "createdAt")); err != nil {
			t.Fatalf("seed saved search: %v", err)
		}
	}
}

func mustJSONRaw(t *testing.T, v any) string {
	t.Helper()
	raw, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal json: %v", err)
	}
	return string(raw)
}

func stringField(m map[string]any, key string) string {
	v, _ := m[key].(string)
	return v
}

func intField(m map[string]any, key string) int {
	return int(m[key].(float64))
}

func boolField(m map[string]any, key string) bool {
	v, _ := m[key].(bool)
	return v
}

func nullableString(v any) any {
	if v == nil {
		return nil
	}
	if s, ok := v.(string); ok {
		return s
	}
	return nil
}

func nullableInt(v any) any {
	if v == nil {
		return nil
	}
	return int(v.(float64))
}

func nullableTime(v any) any {
	if v == nil {
		return nil
	}
	return v.(string)
}

func stringSlice(v any) []string {
	if v == nil {
		return []string{}
	}
	raw := v.([]any)
	out := make([]string, 0, len(raw))
	for _, item := range raw {
		out = append(out, item.(string))
	}
	return out
}
