package auth_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/auth"
	"github.com/monshunter/easyinterview/backend/internal/platform/config"
	"github.com/monshunter/easyinterview/backend/internal/platform/featureflag"
)

func TestE2EP0003PasswordlessSessionCookie(t *testing.T) {
	now := time.Date(2026, 5, 6, 12, 0, 0, 0, time.UTC)
	store := newPasswordlessScenarioStore()
	sink := auth.NewDevMailSink(auth.DevMailSinkOptions{VerifyBaseURL: "http://api.test/api/v1/auth/email/verify"})
	dispatcher := auth.NewBackgroundMailDispatcher(auth.BackgroundMailDispatcherOptions{Writer: sink})
	t.Cleanup(func() {
		if err := dispatcher.Shutdown(context.Background()); err != nil {
			t.Fatalf("dispatcher shutdown: %v", err)
		}
	})
	registry := auth.NewInMemoryAuthMetricRegistry()
	audit := &recordingAuthAudit{}
	service := auth.NewPasswordlessService(auth.PasswordlessServiceOptions{
		Store:                 store,
		Dispatcher:            dispatcher,
		DeliverySecrets:       sink,
		TokenGenerator:        &sequenceTokenGenerator{tokens: []string{"scenario-magic-token-1", "scenario-magic-token-2"}},
		SessionTokenGenerator: &sequenceTokenGenerator{tokens: []string{"scenario-session-token-1", "scenario-session-token-2"}},
		ChallengePepper:       "scenario-pepper",
		SessionCookieSecret:   "scenario-session-secret",
		Metrics:               auth.RegisterAuthMetrics(registry, auth.AuthMetricsOptions{Service: "backend"}),
		Audit:                 audit,
		Now:                   func() time.Time { return now },
		NewID: fixedIDs(
			"challenge-bdd-1",
			"user-bdd-1",
			"session-bdd-1",
			"challenge-bdd-2",
			"user-bdd-ignored",
			"session-bdd-2",
			"privacy-request-bdd",
			"job-bdd",
		),
	})
	handler := auth.NewHandler(auth.HandlerOptions{Passwordless: service})
	traceparent := "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01"
	email := "scenario.user+e2e-p0-003@example.test"

	startChallenge(t, handler, email, traceparent)
	firstLink := waitMagicLink(t, sink, "challenge-bdd-1")
	firstToken := tokenFromMagicLink(t, firstLink)
	assertAPIErrorStatus(t, verifyChallenge(t, handler, "invalid-token", traceparent), http.StatusUnauthorized)
	verifyRec := verifyChallenge(t, handler, firstToken, traceparent)
	firstCookie := requireSessionCookie(t, verifyRec)
	assertAPIErrorStatus(t, verifyChallenge(t, handler, firstToken, traceparent), http.StatusUnauthorized)

	meRoute := auth.SessionMiddleware(service, "getMe", http.HandlerFunc(handler.GetMe))
	assertAPIErrorStatus(t, serveWithCookie(meRoute, http.MethodGet, "/api/v1/me", nil), http.StatusUnauthorized)
	assertAPIErrorStatus(t, serveWithCookie(meRoute, http.MethodGet, "/api/v1/me", &http.Cookie{Name: auth.SessionCookieName, Value: "invalid-session"}), http.StatusUnauthorized)
	meRec := serveWithCookie(meRoute, http.MethodGet, "/api/v1/me", firstCookie)
	if meRec.Code != http.StatusOK {
		t.Fatalf("/me status = %d body=%s", meRec.Code, meRec.Body.String())
	}
	if contains(meRec.Body.String(), email) || !contains(meRec.Body.String(), "@example.test") {
		t.Fatalf("/me masking failed: %s", meRec.Body.String())
	}

	runtimeHandler := config.NewRuntimeConfigHandler(config.RuntimeConfigHandlerOptions{
		Loader: newRuntimeConfigAuthLoader(t),
		Flags: runtimeFlags{snapshot: map[string]featureflag.FlagDecision{
			"practice_hint_enabled":     {Enabled: true, Public: true},
			"ai_fallback_model_enabled": {Enabled: true, Public: false},
		}},
		FlagContextFunc: func(*http.Request) featureflag.FlagContext {
			return featureflag.FlagContext{AppEnv: "dev"}
		},
		SessionResolver: service.RuntimeConfigSessionResolver(),
	})
	runtimeRec := serveWithCookie(runtimeHandler, http.MethodGet, "/api/v1/runtime-config", firstCookie)
	if runtimeRec.Code != http.StatusOK {
		t.Fatalf("/runtime-config status = %d body=%s", runtimeRec.Code, runtimeRec.Body.String())
	}
	if !contains(runtimeRec.Body.String(), "postHogPublicKey") || contains(runtimeRec.Body.String(), "ai_fallback_model_enabled") {
		t.Fatalf("/runtime-config allowlist mismatch: %s", runtimeRec.Body.String())
	}

	logoutRoute := auth.SessionMiddleware(service, "logout", http.HandlerFunc(handler.Logout))
	logoutRec := serveWithCookie(logoutRoute, http.MethodPost, "/api/v1/auth/logout", firstCookie)
	if logoutRec.Code != http.StatusNoContent {
		t.Fatalf("logout status = %d body=%s", logoutRec.Code, logoutRec.Body.String())
	}
	repeatedLogout := serveWithCookie(logoutRoute, http.MethodPost, "/api/v1/auth/logout", firstCookie)
	if repeatedLogout.Code != http.StatusNoContent {
		t.Fatalf("repeated logout status = %d body=%s", repeatedLogout.Code, repeatedLogout.Body.String())
	}
	assertAPIErrorStatus(t, serveWithCookie(meRoute, http.MethodGet, "/api/v1/me", firstCookie), http.StatusUnauthorized)

	startChallenge(t, handler, email, traceparent)
	secondToken := tokenFromMagicLink(t, waitMagicLink(t, sink, "challenge-bdd-2"))
	secondCookie := requireSessionCookie(t, verifyChallenge(t, handler, secondToken, traceparent))
	deleteRoute := auth.SessionMiddleware(service, "deleteMe", http.HandlerFunc(handler.DeleteMe))
	firstDelete := serveDeleteMe(deleteRoute, secondCookie, "delete-key-bdd")
	if firstDelete.Code != http.StatusAccepted {
		t.Fatalf("deleteMe status = %d body=%s", firstDelete.Code, firstDelete.Body.String())
	}
	current := auth.CurrentSession{SessionID: "session-bdd-2", UserID: "user-bdd-1"}
	secondDeleteReq := httptest.NewRequest(http.MethodDelete, "/api/v1/me", nil)
	secondDeleteReq.Header.Set("Idempotency-Key", "delete-key-bdd")
	secondDeleteReq = secondDeleteReq.WithContext(auth.ContextWithCurrentSession(secondDeleteReq.Context(), current))
	secondDelete := httptest.NewRecorder()
	handler.DeleteMe(secondDelete, secondDeleteReq)
	if secondDelete.Code != http.StatusAccepted {
		t.Fatalf("repeated deleteMe status = %d body=%s", secondDelete.Code, secondDelete.Body.String())
	}
	if privacyRequestID(t, firstDelete) != "privacy-request-bdd" || privacyRequestID(t, secondDelete) != "privacy-request-bdd" {
		t.Fatalf("deleteMe idempotency mismatch: first=%s second=%s", firstDelete.Body.String(), secondDelete.Body.String())
	}
	if store.privacyHandoffCount() != 1 {
		t.Fatalf("duplicate privacy handoffs created: %d", store.privacyHandoffCount())
	}
	if got := store.sessionStatus("session-bdd-2"); got != auth.SessionStatusRevoked {
		t.Fatalf("deleteMe did not revoke session, status=%s", got)
	}

	observed := strings.Join([]string{
		meRec.Body.String(),
		runtimeRec.Body.String(),
		firstDelete.Body.String(),
		secondDelete.Body.String(),
		fmt.Sprintf("%+v", registry.CounterLabelValues(auth.MetricAuthChallengeStartedTotal)),
		fmt.Sprintf("%+v", registry.CounterLabelValues(auth.MetricAuthSessionMintedTotal)),
		fmt.Sprintf("%+v", registry.CounterLabelValues(auth.MetricAuthLogoutTotal)),
		fmt.Sprintf("%+v", registry.CounterLabelValues(auth.MetricAuthDeleteHandoffTotal)),
		fmt.Sprintf("%+v", registry.CounterLabelValues(auth.MetricAuthFailureTotal)),
		fmt.Sprintf("%+v", audit.events),
		fmt.Sprintf("%+v", sink),
		strings.Join(dispatcher.ErrorSummaries(), "\n"),
	}, "\n")
	for _, forbidden := range []string{
		"scenario-magic-token-1",
		"scenario-magic-token-2",
		"scenario-session-token-1",
		"scenario-session-token-2",
		email,
		"scenario-pepper",
		"scenario-session-secret",
		"http://api.test/api/v1/auth/email/verify",
	} {
		if contains(observed, forbidden) {
			t.Fatalf("scenario evidence leaked %q: %s", forbidden, observed)
		}
	}
}

func startChallenge(t *testing.T, handler *auth.Handler, email string, traceparent string) {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/email/start", bytes.NewBufferString(fmt.Sprintf(`{"email":%q}`, email)))
	req.Header.Set("traceparent", traceparent)
	req.RemoteAddr = "203.0.113.60:5588"
	req.Header.Set("User-Agent", "scenario-runner")
	rec := httptest.NewRecorder()
	handler.StartAuthEmailChallenge(rec, req)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("start challenge status = %d body=%s", rec.Code, rec.Body.String())
	}
}

func verifyChallenge(t *testing.T, handler *auth.Handler, token string, traceparent string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/email/verify?token="+url.QueryEscape(token), nil)
	req.Header.Set("traceparent", traceparent)
	req.RemoteAddr = "203.0.113.60:5588"
	req.Header.Set("User-Agent", "scenario-runner")
	rec := httptest.NewRecorder()
	handler.VerifyAuthEmailChallenge(rec, req)
	return rec
}

func waitMagicLink(t *testing.T, sink *auth.DevMailSink, challengeID string) string {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if link, ok := sink.MagicLinkForChallenge(challengeID); ok {
			return link
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("magic link for %s not delivered", challengeID)
	return ""
}

func tokenFromMagicLink(t *testing.T, link string) string {
	t.Helper()
	u, err := url.Parse(link)
	if err != nil {
		t.Fatalf("parse magic link: %v", err)
	}
	token := u.Query().Get("token")
	if token == "" {
		t.Fatalf("magic link missing token: %s", link)
	}
	return token
}

func requireSessionCookie(t *testing.T, rec *httptest.ResponseRecorder) *http.Cookie {
	t.Helper()
	if rec.Code != http.StatusOK {
		t.Fatalf("verify status = %d body=%s", rec.Code, rec.Body.String())
	}
	for _, cookie := range rec.Result().Cookies() {
		if cookie.Name == auth.SessionCookieName && cookie.Value != "" {
			return cookie
		}
	}
	t.Fatalf("verify response missing session cookie: %#v", rec.Result().Cookies())
	return nil
}

func serveWithCookie(handler http.Handler, method string, path string, cookie *http.Cookie) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, nil)
	if cookie != nil {
		req.AddCookie(cookie)
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

func serveDeleteMe(handler http.Handler, cookie *http.Cookie, idempotencyKey string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/me", nil)
	req.Header.Set("Idempotency-Key", idempotencyKey)
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

func assertAPIErrorStatus(t *testing.T, rec *httptest.ResponseRecorder, status int) {
	t.Helper()
	if rec.Code != status {
		t.Fatalf("status = %d want %d body=%s", rec.Code, status, rec.Body.String())
	}
	var body map[string]map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode error envelope: %v body=%s", err, rec.Body.String())
	}
	if _, ok := body["error"]["code"]; !ok {
		t.Fatalf("missing error code: %+v", body)
	}
}

func privacyRequestID(t *testing.T, rec *httptest.ResponseRecorder) string {
	t.Helper()
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode privacy response: %v", err)
	}
	id, _ := body["privacyRequestId"].(string)
	return id
}

type sequenceTokenGenerator struct {
	tokens []string
	index  int
}

func (g *sequenceTokenGenerator) GenerateToken() (string, error) {
	if g.index >= len(g.tokens) {
		return "", fmt.Errorf("sequence token generator exhausted")
	}
	token := g.tokens[g.index]
	g.index++
	return token, nil
}

type passwordlessScenarioStore struct {
	challenges      map[string]*scenarioChallenge
	usersByEmail    map[string]auth.UserContext
	usersByID       map[string]auth.UserContext
	sessionsByHash  map[string]auth.SessionRecord
	sessionsByID    map[string]string
	privacyHandoffs map[string]auth.PrivacyDeleteHandoff
}

type scenarioChallenge struct {
	record auth.ChallengeRecord
	status string
}

func newPasswordlessScenarioStore() *passwordlessScenarioStore {
	return &passwordlessScenarioStore{
		challenges:      map[string]*scenarioChallenge{},
		usersByEmail:    map[string]auth.UserContext{},
		usersByID:       map[string]auth.UserContext{},
		sessionsByHash:  map[string]auth.SessionRecord{},
		sessionsByID:    map[string]string{},
		privacyHandoffs: map[string]auth.PrivacyDeleteHandoff{},
	}
}

func (s *passwordlessScenarioStore) CountRecentChallenges(_ context.Context, email string, ipHash string, since time.Time) (int, error) {
	count := 0
	for _, challenge := range s.challenges {
		rec := challenge.record
		if challenge.status == "pending" && !rec.CreatedAt.Before(since) && (rec.Email == email || rec.IPHash == ipHash) {
			count++
		}
	}
	return count, nil
}

func (s *passwordlessScenarioStore) CreateChallenge(_ context.Context, rec auth.ChallengeRecord) error {
	s.challenges[rec.TokenHash] = &scenarioChallenge{record: rec, status: "pending"}
	return nil
}

func (s *passwordlessScenarioStore) ConsumeChallenge(_ context.Context, tokenHash string, now time.Time) (auth.ChallengeRecord, error) {
	challenge, ok := s.challenges[tokenHash]
	if !ok {
		return auth.ChallengeRecord{}, auth.ErrChallengeInvalid
	}
	if challenge.status == "consumed" {
		return auth.ChallengeRecord{}, auth.ErrChallengeConsumed
	}
	if !challenge.record.ExpiresAt.After(now) {
		return auth.ChallengeRecord{}, auth.ErrChallengeExpired
	}
	challenge.status = "consumed"
	return challenge.record, nil
}

func (s *passwordlessScenarioStore) FindOrCreateUserByEmail(_ context.Context, email string, userID string, _ time.Time) (auth.UserContext, error) {
	if user, ok := s.usersByEmail[email]; ok {
		return user, nil
	}
	user := auth.UserContext{
		ID:                        userID,
		Email:                     email,
		DisplayName:               "Scenario User",
		UILanguage:                "zh-CN",
		PreferredPracticeLanguage: "en",
		AnalyticsOptIn:            true,
	}
	s.usersByEmail[email] = user
	s.usersByID[user.ID] = user
	return user, nil
}

func (s *passwordlessScenarioStore) CreateSession(_ context.Context, rec auth.SessionRecord) error {
	s.sessionsByHash[rec.SessionHash] = rec
	s.sessionsByID[rec.ID] = rec.SessionHash
	return nil
}

func (s *passwordlessScenarioStore) GetSessionByHash(_ context.Context, sessionHash string, _ time.Time) (auth.SessionRecord, error) {
	rec, ok := s.sessionsByHash[sessionHash]
	if !ok {
		return auth.SessionRecord{}, auth.ErrSessionInvalid
	}
	return rec, nil
}

func (s *passwordlessScenarioStore) GetUserContext(_ context.Context, userID string) (auth.UserContext, error) {
	user, ok := s.usersByID[userID]
	if !ok {
		return auth.UserContext{}, auth.ErrSessionInvalid
	}
	return user, nil
}

func (s *passwordlessScenarioStore) TouchSession(_ context.Context, sessionID string, now time.Time, expiresAt time.Time) error {
	hash, ok := s.sessionsByID[sessionID]
	if !ok {
		return auth.ErrSessionInvalid
	}
	rec := s.sessionsByHash[hash]
	rec.UpdatedAt = now
	rec.ExpiresAt = expiresAt
	s.sessionsByHash[hash] = rec
	return nil
}

func (s *passwordlessScenarioStore) RevokeSession(_ context.Context, sessionID string, now time.Time) error {
	hash, ok := s.sessionsByID[sessionID]
	if !ok {
		return auth.ErrSessionInvalid
	}
	rec := s.sessionsByHash[hash]
	rec.Status = auth.SessionStatusRevoked
	rec.RevokedAt = now
	rec.UpdatedAt = now
	s.sessionsByHash[hash] = rec
	return nil
}

func (s *passwordlessScenarioStore) CreatePrivacyDeleteHandoff(_ context.Context, userID string, idempotencyKey string, privacyRequestID string, jobID string, now time.Time) (auth.PrivacyDeleteHandoff, error) {
	key := userID + "\x00" + idempotencyKey
	if existing, ok := s.privacyHandoffs[key]; ok {
		return existing, nil
	}
	handoff := auth.PrivacyDeleteHandoff{
		PrivacyRequestID: privacyRequestID,
		JobID:            jobID,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	s.privacyHandoffs[key] = handoff
	return handoff, nil
}

func (s *passwordlessScenarioStore) privacyHandoffCount() int {
	return len(s.privacyHandoffs)
}

func (s *passwordlessScenarioStore) sessionStatus(sessionID string) auth.SessionStatus {
	hash := s.sessionsByID[sessionID]
	return s.sessionsByHash[hash].Status
}
