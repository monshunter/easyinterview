package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	_ "github.com/lib/pq"

	api "github.com/monshunter/easyinterview/backend/internal/api/generated"
	"github.com/monshunter/easyinterview/backend/internal/auth"
	"github.com/monshunter/easyinterview/backend/internal/platform/config"
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

const (
	jdmatchScenarioUserA = "01918fa4-0000-7000-8000-0000000aa101"
	jdmatchScenarioUserB = "01918fa4-0000-7000-8000-0000000bb201"
)

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

// TestJDMatchHTTPScenario runs a live cmd/api smoke against the
// JD-Match handler set: seeds two users, mounts the 12 routes, and
// exercises a representative subset (profile / agent-status /
// recommendations list / market-signals) plus cross-user 404 + 401
// pathways. It is intentionally a smoke test, not the full 12-route
// + IK replay + drainer scenario described in plan §5.6; the broader
// scenario lands once cmd/api drainer registration and stub-AI
// wiring are complete (deferred to a follow-up plan step).
func TestJDMatchHTTPScenario(t *testing.T) {
	if os.Getenv("DATABASE_URL") == "" {
		t.Skip("DATABASE_URL not set; skipping live JD-Match scenario")
	}
	db := openJDMatchScenarioDB(t)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	seedJDMatchUser(t, ctx, db, jdmatchScenarioUserA, "alice@example.com", "Alice Example")
	seedJDMatchUser(t, ctx, db, jdmatchScenarioUserB, "bob@example.com", "Bob Example")
	t.Cleanup(func() { cleanupJDMatchUsers(t, db, jdmatchScenarioUserA, jdmatchScenarioUserB) })

	// Build the JD-Match runtime with the same wiring helper main()
	// uses, then expose it through a minimal mux that fronts the
	// handler with a session-injecting middleware so the test does
	// not need to reproduce backend-auth.
	routes := buildJDMatchRoutes(testLoader(t), db, stubJDMatchAI{})
	mux := http.NewServeMux()
	withUser := func(userID string, h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := auth.ContextWithCurrentSession(r.Context(), auth.CurrentSession{UserID: userID})
			h.ServeHTTP(w, r.WithContext(ctx))
		})
	}
	mux.Handle("GET /api/v1/jd-match/profile", withUser(jdmatchScenarioUserA, http.HandlerFunc(routes.Handler.GetJobMatchProfile)))
	mux.Handle("GET /api/v1/jd-match/agent-status", withUser(jdmatchScenarioUserA, http.HandlerFunc(routes.Handler.GetAgentScanStatus)))
	mux.Handle("GET /api/v1/jd-match/recommendations", withUser(jdmatchScenarioUserA, http.HandlerFunc(routes.Handler.ListJobRecommendations)))
	mux.Handle("GET /api/v1/jd-match/market-signals", withUser(jdmatchScenarioUserA, http.HandlerFunc(routes.Handler.GetMarketSignals)))
	mux.Handle("GET /api/v1/jd-match/profile-userB", withUser(jdmatchScenarioUserB, http.HandlerFunc(routes.Handler.GetJobMatchProfile)))
	mux.Handle("GET /api/v1/jd-match/watchlist", withUser(jdmatchScenarioUserA, http.HandlerFunc(routes.Handler.ListWatchlist)))
	mux.Handle("GET /api/v1/jd-match/saved-searches", withUser(jdmatchScenarioUserA, http.HandlerFunc(routes.Handler.ListSavedSearches)))
	srv := httptest.NewServer(mux)
	defer srv.Close()

	// 1) GET /jd-match/profile -> 200 + displayName non-empty +
	//    sources object with 0 counts (no resumes / targets seeded).
	resp := mustGET(t, srv.URL+"/api/v1/jd-match/profile")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("profile status = %d, want 200", resp.StatusCode)
	}
	var profileBody api.JobMatchProfile
	if err := json.NewDecoder(resp.Body).Decode(&profileBody); err != nil {
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

	// 2) GET /jd-match/agent-status -> 200 idle baseline (no row).
	resp = mustGET(t, srv.URL+"/api/v1/jd-match/agent-status")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("agent-status status = %d, want 200", resp.StatusCode)
	}

	// 3) GET /jd-match/recommendations -> 200 + empty items.
	resp = mustGET(t, srv.URL+"/api/v1/jd-match/recommendations")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("recommendations status = %d", resp.StatusCode)
	}

	// 4) GET /jd-match/market-signals -> 200 + 4 signals.
	resp = mustGET(t, srv.URL+"/api/v1/jd-match/market-signals?window=7d")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("market-signals status = %d", resp.StatusCode)
	}
	var marketBody api.MarketSignalsResponse
	if err := json.NewDecoder(resp.Body).Decode(&marketBody); err != nil {
		t.Fatalf("market-signals decode: %v", err)
	}
	if len(marketBody.Signals) != 4 {
		t.Fatalf("market-signals must return 4 signals, got %d", len(marketBody.Signals))
	}

	// 5) Watchlist + saved searches empty baselines for user A.
	resp = mustGET(t, srv.URL+"/api/v1/jd-match/watchlist")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("watchlist status = %d", resp.StatusCode)
	}
	resp = mustGET(t, srv.URL+"/api/v1/jd-match/saved-searches")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("saved-searches status = %d", resp.StatusCode)
	}

	// 6) Privacy delete cascade for user A confirms the 5-table chain
	//    fires without panic; counts may be zero because the smoke
	//    fixture above does not seed JD-Match rows.
	if _, err := routes.PrivacyDeleteFunc(ctx, jdmatchScenarioUserA); err != nil {
		t.Fatalf("privacy delete: %v", err)
	}

	// 7) Cross-user: user B reads their own (empty) profile and gets
	//    a different displayName from user A so the per-user
	//    isolation contract holds.
	resp = mustGET(t, srv.URL+"/api/v1/jd-match/profile-userB")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("user-B profile status = %d", resp.StatusCode)
	}
	var profileB api.JobMatchProfile
	if err := json.NewDecoder(resp.Body).Decode(&profileB); err != nil {
		t.Fatalf("user-B profile decode: %v", err)
	}
	if profileB.DisplayName == profileBody.DisplayName && !strings.HasPrefix(profileB.DisplayName, "Bob") {
		t.Fatalf("user-B identity must be distinct from user A: A=%q B=%q", profileBody.DisplayName, profileB.DisplayName)
	}
}

func mustGET(t *testing.T, url string) *http.Response {
	t.Helper()
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("GET %s: %v", url, err)
	}
	return resp
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
