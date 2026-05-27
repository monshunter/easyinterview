package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
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
	"github.com/monshunter/easyinterview/backend/internal/middleware/idempotency"
	"github.com/monshunter/easyinterview/backend/internal/platform/config"
	"github.com/monshunter/easyinterview/backend/internal/profile"
	profilehandler "github.com/monshunter/easyinterview/backend/internal/profile/handler"
	profileservice "github.com/monshunter/easyinterview/backend/internal/profile/service"
	profilestore "github.com/monshunter/easyinterview/backend/internal/profile/store"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
	"github.com/monshunter/easyinterview/backend/internal/shared/idx"
	"github.com/monshunter/easyinterview/backend/internal/targetjob"
)

const (
	profileScenarioUserA = "01918fa3-0000-7000-8000-0000000aa101"
	profileScenarioUserB = "01918fa3-0000-7000-8000-0000000bb201"
	profileScenarioUserC = "01918fa3-0000-7000-8000-0000000cc301"
)

func openProfileScenarioDB(t *testing.T) *sql.DB {
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
		t.Skipf("postgres ping failed (%v); skipping profile cmd/api scenario", err)
	}
	return db
}

func setupProfileScenarioUsers(t *testing.T, ctx context.Context, db *sql.DB) {
	t.Helper()
	for _, id := range []string{profileScenarioUserA, profileScenarioUserB, profileScenarioUserC} {
		if _, err := db.ExecContext(ctx, `
insert into users (id, email, status)
values ($1, $2, 'active')
on conflict (id) do nothing`, id, id+"@profile-scenario.local"); err != nil {
			t.Fatalf("seed user %s: %v", id, err)
		}
		if _, err := db.ExecContext(ctx, `
insert into user_settings (user_id, ui_language, preferred_practice_language, region)
values ($1, 'zh-CN', 'en', 'CN-SH')
on conflict (user_id) do update set
  ui_language = excluded.ui_language,
  preferred_practice_language = excluded.preferred_practice_language,
  region = excluded.region`, id); err != nil {
			t.Fatalf("seed user_settings %s: %v", id, err)
		}
	}
	t.Cleanup(func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		ids := []string{profileScenarioUserA, profileScenarioUserB, profileScenarioUserC}
		for _, id := range ids {
			_, _ = db.ExecContext(cleanupCtx, `delete from experience_cards where user_id = $1`, id)
			_, _ = db.ExecContext(cleanupCtx, `delete from candidate_profiles where user_id = $1`, id)
			_, _ = db.ExecContext(cleanupCtx, `delete from audit_events where user_id = $1`, id)
			_, _ = db.ExecContext(cleanupCtx, `delete from idempotency_records where user_id = $1`, id)
			_, _ = db.ExecContext(cleanupCtx, `delete from user_settings where user_id = $1`, id)
			_, _ = db.ExecContext(cleanupCtx, `delete from users where id = $1`, id)
		}
	})
}

type profileScenarioFixture struct {
	handler http.Handler
	routes  profileRoutes
	authA   *apiAuthStore
	authB   *apiAuthStore
	authC   *apiAuthStore
	userA   string
	userB   string
	userC   string
	cleanup func()
}

func buildProfileScenarioFixture(t *testing.T, db *sql.DB) profileScenarioFixture {
	t.Helper()
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

	// Provision an apiAuthStore per user so SessionMiddleware can resolve
	// each cookie to its own user.
	mkAuth := func(id, email string) *apiAuthStore {
		return &apiAuthStore{
			session: auth.SessionRecord{
				ID:        "session-" + id,
				UserID:    id,
				Status:    auth.SessionStatusActive,
				ExpiresAt: time.Now().Add(auth.SessionTTL),
			},
			user: auth.UserContext{ID: id, Email: email},
		}
	}
	storeA := mkAuth(profileScenarioUserA, "user-a@profile-scenario.local")
	storeB := mkAuth(profileScenarioUserB, "user-b@profile-scenario.local")
	storeC := mkAuth(profileScenarioUserC, "user-c@profile-scenario.local")

	authService := auth.NewPasswordlessService(auth.PasswordlessServiceOptions{
		Store: &multiUserAuthStore{users: map[string]*apiAuthStore{
			"raw-session-user-a": storeA,
			"raw-session-user-b": storeB,
			"raw-session-user-c": storeC,
		}},
		SessionCookieSecret: "session-secret",
	})

	repo := profilestore.NewRepository(db)
	settings := profilestore.NewSettingsReader(db)
	audit := profilestore.NewAuditTombstoneWriter(db)
	svc := profileservice.New(profileservice.Options{Store: repo, Audit: audit})
	hand := profilehandler.New(profilehandler.Options{
		Store: repo, Settings: settings, Session: currentUserFromContext, NewID: idx.NewID,
	})
	routes := profileRoutes{
		Handler: hand,
		Store:   repo,
		Service: svc,
		Idempotency: idempotency.New(idempotency.MiddlewareOptions{
			Store:     idempotency.NewSQLStore(db),
			KeyPepper: "scenario-pepper",
			TTL:       24 * time.Hour,
		}),
	}

	handler := buildAPIHandlerWithUploadReportDebriefJobsProfileAndHandlers(
		loader, apiRuntimeFlags{}, authService, targetjob.NewHandler(),
		practiceRoutes{}, uploadRoutes{}, resumeRoutes{}, reportRoutes{}, debriefRoutes{}, jobsRoutes{}, routes,
	)
	return profileScenarioFixture{
		handler: handler,
		routes:  routes,
		authA:   storeA,
		authB:   storeB,
		authC:   storeC,
		userA:   profileScenarioUserA,
		userB:   profileScenarioUserB,
		userC:   profileScenarioUserC,
	}
}

// multiUserAuthStore lets the same SessionMiddleware resolve different
// session cookies to different users. Each cookie carries a distinct raw
// token that maps to a per-user apiAuthStore.
type multiUserAuthStore struct {
	users map[string]*apiAuthStore
}

func (s *multiUserAuthStore) CountRecentChallenges(context.Context, string, string, time.Time) (int, error) {
	return 0, nil
}
func (s *multiUserAuthStore) CreateChallenge(context.Context, auth.ChallengeRecord) error { return nil }
func (s *multiUserAuthStore) ConsumeChallenge(context.Context, string, time.Time) (auth.ChallengeRecord, error) {
	return auth.ChallengeRecord{}, nil
}
func (s *multiUserAuthStore) CreateUserByEmail(context.Context, string, string, string, time.Time) (auth.UserContext, error) {
	return auth.UserContext{}, nil
}
func (s *multiUserAuthStore) FindUserByEmail(context.Context, string) (auth.UserContext, error) {
	return auth.UserContext{}, nil
}
func (s *multiUserAuthStore) CreateSession(context.Context, auth.SessionRecord) error { return nil }
func (s *multiUserAuthStore) GetSessionByHash(_ context.Context, hash string, _ time.Time) (auth.SessionRecord, error) {
	for token, store := range s.users {
		if scenarioHashSessionToken(token, "session-secret") == hash {
			return store.session, nil
		}
	}
	return auth.SessionRecord{}, auth.ErrSessionInvalid
}
func (s *multiUserAuthStore) GetUserContext(_ context.Context, userID string) (auth.UserContext, error) {
	for _, store := range s.users {
		if store.user.ID == userID {
			return store.user, nil
		}
	}
	return auth.UserContext{}, auth.ErrSessionInvalid
}

// scenarioHashSessionToken mirrors auth.hashWithPepper so the scenario test
// can resolve cookie token -> hash without exporting the helper from auth.
func scenarioHashSessionToken(token, pepper string) string {
	sum := sha256.Sum256([]byte(pepper + "\x00" + token))
	return hex.EncodeToString(sum[:])
}
func (s *multiUserAuthStore) TouchSession(context.Context, string, time.Time, time.Time) error {
	return nil
}
func (s *multiUserAuthStore) RevokeSession(context.Context, string, time.Time) error { return nil }
func (s *multiUserAuthStore) CreatePrivacyDeleteHandoff(context.Context, string, string, string, string, time.Time) (auth.PrivacyDeleteHandoff, error) {
	panic("not used")
}

func userToken(userID string) string {
	switch userID {
	case profileScenarioUserA:
		return "raw-session-user-a"
	case profileScenarioUserB:
		return "raw-session-user-b"
	case profileScenarioUserC:
		return "raw-session-user-c"
	}
	return ""
}

func doProfileRequest(t *testing.T, h http.Handler, method, path string, userID string, idempotencyKey string, body any, wantStatus int) []byte {
	t.Helper()
	var reqBody *bytes.Reader
	if body == nil {
		reqBody = bytes.NewReader(nil)
	} else {
		raw, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request: %v", err)
		}
		reqBody = bytes.NewReader(raw)
	}
	req := httptest.NewRequest(method, path, reqBody)
	if token := userToken(userID); token != "" {
		req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: token})
	}
	if idempotencyKey != "" {
		req.Header.Set("Idempotency-Key", idempotencyKey)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != wantStatus {
		t.Fatalf("%s %s userID=%s status=%d want=%d body=%s", method, path, userID, rec.Code, wantStatus, rec.Body.String())
	}
	return rec.Body.Bytes()
}

func TestProfileHTTPScenario(t *testing.T) {
	db := openProfileScenarioDB(t)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	setupProfileScenarioUsers(t, ctx, db)
	fx := buildProfileScenarioFixture(t, db)

	// --- E2E.P0.091 candidate profile seed + patch ----------------------------
	rawA1 := doProfileRequest(t, fx.handler, http.MethodGet, "/api/v1/profiles/me", fx.userA, "", nil, http.StatusOK)
	var seed api.CandidateProfile
	if err := json.Unmarshal(rawA1, &seed); err != nil {
		t.Fatalf("seed body: %v", err)
	}
	if seed.Headline != nil || seed.YearsOfExperience != nil || seed.CurrentRole != nil {
		t.Fatalf("seed must return null nullable fields: %+v", seed)
	}
	if seed.PreferredPracticeLanguage != "en" || seed.UiLanguage != "zh-CN" {
		t.Fatalf("seed defaults wrong: %+v", seed)
	}

	// A2 second call returns same row (no re-seed).
	doProfileRequest(t, fx.handler, http.MethodGet, "/api/v1/profiles/me", fx.userA, "", nil, http.StatusOK)

	// A3 patch headline + years.
	headline := "Senior frontend engineer focused on growth-stage SaaS"
	years := int32(5)
	rawA3 := doProfileRequest(t, fx.handler, http.MethodPatch, "/api/v1/profiles/me", fx.userA, "", api.UpdateProfileRequest{
		Headline:          &headline,
		YearsOfExperience: &years,
	}, http.StatusOK)
	var patched api.CandidateProfile
	if err := json.Unmarshal(rawA3, &patched); err != nil {
		t.Fatalf("patch body: %v", err)
	}
	if patched.Headline == nil || *patched.Headline != headline {
		t.Fatalf("headline = %v, want %q", patched.Headline, headline)
	}

	// A5 invalid yearsOfExperience returns 422.
	bad := int32(-1)
	rawA5 := doProfileRequest(t, fx.handler, http.MethodPatch, "/api/v1/profiles/me", fx.userA, "", api.UpdateProfileRequest{YearsOfExperience: &bad}, http.StatusUnprocessableEntity)
	var validation api.ApiErrorResponse
	if err := json.Unmarshal(rawA5, &validation); err != nil {
		t.Fatalf("validation body: %v", err)
	}
	if validation.Error.Code != sharederrors.CodeValidationFailed {
		t.Fatalf("validation error code = %q", validation.Error.Code)
	}

	// B1 user B sees own seed only.
	rawB := doProfileRequest(t, fx.handler, http.MethodGet, "/api/v1/profiles/me", fx.userB, "", nil, http.StatusOK)
	var bSeed api.CandidateProfile
	if err := json.Unmarshal(rawB, &bSeed); err != nil {
		t.Fatalf("user B seed: %v", err)
	}
	if bSeed.Headline != nil {
		t.Fatalf("user B seed leaked: %+v", bSeed)
	}

	// --- E2E.P0.092 experience cards CRUD + IK --------------------------------
	createBody := api.CreateExperienceCardRequest{
		Title:       "Drove design-system migration",
		CompanyName: "Acme",
		Situation:   "fragmented design systems",
		Task:        "unify",
		Action:      "RFC + 6-week rollout",
		Result:      "Reduced UI defects by 38%",
		Skills:      []string{"leadership"},
		Language:    "zh-CN",
	}
	ikCreate := "01918fa3-0000-7000-8000-0000000000c1"
	rawCreate := doProfileRequest(t, fx.handler, http.MethodPost, "/api/v1/profiles/me/experience-cards", fx.userA, ikCreate, createBody, http.StatusCreated)
	var created api.ExperienceCard
	if err := json.Unmarshal(rawCreate, &created); err != nil {
		t.Fatalf("create body: %v", err)
	}
	if created.Title != createBody.Title {
		t.Fatalf("create title = %q", created.Title)
	}

	// IK replay: same key + same body returns first card without duplicating.
	rawReplay := doProfileRequest(t, fx.handler, http.MethodPost, "/api/v1/profiles/me/experience-cards", fx.userA, ikCreate, createBody, http.StatusCreated)
	var replay api.ExperienceCard
	if err := json.Unmarshal(rawReplay, &replay); err != nil {
		t.Fatalf("replay body: %v", err)
	}
	if replay.Id != created.Id {
		t.Fatalf("IK replay produced new id: created=%q replay=%q", created.Id, replay.Id)
	}

	// Cross-user update returns 404 + RESOURCE_NOT_FOUND.
	patchBody := api.UpdateExperienceCardRequest{Result: strPtr("hijacked")}
	ikCross := "01918fa3-0000-7000-8000-0000000000c2"
	rawCross := doProfileRequest(t, fx.handler, http.MethodPatch, "/api/v1/profiles/me/experience-cards/"+created.Id, fx.userB, ikCross, patchBody, http.StatusNotFound)
	var crossErr api.ApiErrorResponse
	if err := json.Unmarshal(rawCross, &crossErr); err != nil {
		t.Fatalf("cross body: %v", err)
	}
	if crossErr.Error.Code != sharederrors.CodeResourceNotFound {
		t.Fatalf("cross-user error code = %q, want RESOURCE_NOT_FOUND", crossErr.Error.Code)
	}

	// Missing IK on create returns 422.
	rawMissingIK := doProfileRequest(t, fx.handler, http.MethodPost, "/api/v1/profiles/me/experience-cards", fx.userA, "", createBody, http.StatusUnprocessableEntity)
	if !strings.Contains(string(rawMissingIK), "Idempotency-Key") {
		t.Fatalf("missing IK body = %s", rawMissingIK)
	}

	// List returns at least the one card user A created.
	rawList := doProfileRequest(t, fx.handler, http.MethodGet, "/api/v1/profiles/me/experience-cards?pageSize=20", fx.userA, "", nil, http.StatusOK)
	var page api.PaginatedExperienceCard
	if err := json.Unmarshal(rawList, &page); err != nil {
		t.Fatalf("list body: %v", err)
	}
	if len(page.Items) == 0 {
		t.Fatal("list returned empty items")
	}

	// Internal API: CountExperienceCardsBySource (D-11).
	counts, err := fx.routes.Service.CountExperienceCardsBySource(ctx, fx.userA)
	if err != nil {
		t.Fatalf("count internal: %v", err)
	}
	if counts[profile.SourceTypeManual] < 1 {
		t.Fatalf("manual count = %d", counts[profile.SourceTypeManual])
	}

	// Internal API: GetCandidateProfileForUser (D-13) — read-only for userC.
	gotC, err := fx.routes.Service.GetCandidateProfileForUser(ctx, fx.userC)
	if err != nil {
		t.Fatalf("internal read userC: %v", err)
	}
	if gotC != nil {
		t.Fatalf("internal read userC returned %+v; want nil (no seed)", gotC)
	}
	// Then verify userC can still seed via the public endpoint.
	rawCseed := doProfileRequest(t, fx.handler, http.MethodGet, "/api/v1/profiles/me", fx.userC, "", nil, http.StatusOK)
	var cSeed api.CandidateProfile
	if err := json.Unmarshal(rawCseed, &cSeed); err != nil {
		t.Fatalf("userC public seed body: %v", err)
	}
	if cSeed.Headline != nil {
		t.Fatalf("userC seed leak: %+v", cSeed)
	}

	// Internal read after seed returns *CandidateProfile.
	gotC2, err := fx.routes.Service.GetCandidateProfileForUser(ctx, fx.userC)
	if err != nil {
		t.Fatalf("internal read userC post-seed: %v", err)
	}
	if gotC2 == nil {
		t.Fatal("internal read userC post-seed returned nil")
	}

	// --- E2E.P0.093 privacy delete lifecycle ---------------------------------
	if err := fx.routes.Service.DeleteCandidateProfileForUser(ctx, fx.userA, "scenario-job"); err != nil {
		t.Fatalf("privacy delete: %v", err)
	}
	// DB state: 0 rows for user A.
	var remainingProfiles, remainingCards int
	if err := db.QueryRowContext(ctx, `select count(*) from candidate_profiles where user_id = $1`, fx.userA).Scan(&remainingProfiles); err != nil {
		t.Fatalf("count remaining profiles: %v", err)
	}
	if err := db.QueryRowContext(ctx, `select count(*) from experience_cards where user_id = $1`, fx.userA).Scan(&remainingCards); err != nil {
		t.Fatalf("count remaining cards: %v", err)
	}
	if remainingProfiles != 0 || remainingCards != 0 {
		t.Fatalf("delete left rows: profiles=%d cards=%d", remainingProfiles, remainingCards)
	}
	// audit_events row exists with no raw card content.
	var auditMetadata []byte
	err = db.QueryRowContext(ctx, `
select metadata::text
  from audit_events
 where user_id = $1 and action = 'profile.privacy_delete'
 order by created_at desc limit 1`, fx.userA).Scan(&auditMetadata)
	if err != nil {
		t.Fatalf("audit lookup: %v", err)
	}
	payload := string(auditMetadata)
	for _, leak := range []string{"Drove design-system", "Reduced UI defects", "Acme", "RFC + 6-week rollout"} {
		if strings.Contains(payload, leak) {
			t.Fatalf("audit metadata leaks raw card content: %s", payload)
		}
	}
	// After delete the next getMyProfile call re-seeds (D-1).
	doProfileRequest(t, fx.handler, http.MethodGet, "/api/v1/profiles/me", fx.userA, "", nil, http.StatusOK)
}
