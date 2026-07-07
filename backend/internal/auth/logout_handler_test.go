package auth_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/auth"
)

func TestLogoutRevokesCurrentSessionAndClearsCookie(t *testing.T) {
	store := &logoutStore{}
	service := auth.NewEmailCodeService(auth.EmailCodeServiceOptions{
		Store: store,
		Now:   func() time.Time { return time.Date(2026, 5, 6, 10, 45, 0, 0, time.UTC) },
	})
	handler := auth.NewHandler(auth.HandlerOptions{EmailCode: service})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	req = req.WithContext(auth.ContextWithCurrentSession(req.Context(), auth.CurrentSession{
		SessionID: "session-1",
		UserID:    "user-1",
	}))
	rec := httptest.NewRecorder()

	handler.Logout(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if store.revokedSessionID != "session-1" {
		t.Fatalf("revoked session = %q", store.revokedSessionID)
	}
	cookies := rec.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Name != auth.SessionCookieName || cookies[0].MaxAge >= 0 || !cookies[0].Secure {
		t.Fatalf("clear cookie = %#v", cookies)
	}
}

func TestLogoutWithoutSessionIsIdempotentAndClearsCookie(t *testing.T) {
	handler := auth.NewHandler(auth.HandlerOptions{})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	rec := httptest.NewRecorder()

	handler.Logout(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	cookies := rec.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Name != auth.SessionCookieName || cookies[0].MaxAge >= 0 || !cookies[0].Secure {
		t.Fatalf("clear cookie = %#v", cookies)
	}
}

func TestLogoutCanUseExplicitDevInsecureCookiePolicy(t *testing.T) {
	policy := auth.CookiePolicyForAppEnv("dev")
	handler := auth.NewHandler(auth.HandlerOptions{CookiePolicy: &policy})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	rec := httptest.NewRecorder()

	handler.Logout(rec, req)

	cookies := rec.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Secure {
		t.Fatalf("dev clear cookie should explicitly omit Secure: %#v", cookies)
	}
}

func TestLogoutRevokeFailureReturnsErrorEnvelopeAndClearsCookie(t *testing.T) {
	store := &logoutStore{revokeErr: errors.New("database unavailable for session-1")}
	service := auth.NewEmailCodeService(auth.EmailCodeServiceOptions{
		Store: store,
		Now:   func() time.Time { return time.Date(2026, 5, 6, 20, 30, 0, 0, time.UTC) },
	})
	handler := auth.NewHandler(auth.HandlerOptions{EmailCode: service})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	req = req.WithContext(auth.ContextWithCurrentSession(req.Context(), auth.CurrentSession{
		SessionID: "session-1",
		UserID:    "user-1",
	}))
	rec := httptest.NewRecorder()

	handler.Logout(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	cookies := rec.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Name != auth.SessionCookieName || cookies[0].MaxAge >= 0 || !cookies[0].Secure {
		t.Fatalf("clear cookie = %#v", cookies)
	}
	var body map[string]map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("bad error JSON: %v", err)
	}
	if body["error"]["code"] != "VALIDATION_FAILED" {
		t.Fatalf("error = %+v", body)
	}
	if strings.Contains(rec.Body.String(), "session-1") {
		t.Fatalf("logout failure leaked session id: %s", rec.Body.String())
	}
}

type logoutStore struct {
	revokedSessionID string
	revokeErr        error
}

func (s *logoutStore) CountRecentChallenges(context.Context, string, string, time.Time) (int, error) {
	return 0, nil
}

func (s *logoutStore) CreateChallenge(context.Context, auth.ChallengeRecord) error {
	panic("not used")
}

func (s *logoutStore) ConsumeChallenge(context.Context, string, time.Time) (auth.ChallengeRecord, error) {
	panic("not used")
}

func (s *logoutStore) CreateUserByEmail(context.Context, string, string, string, time.Time) (auth.UserContext, error) {
	panic("not used")
}

func (s *logoutStore) FindUserByEmail(context.Context, string) (auth.UserContext, error) {
	panic("not used")
}

func (s *logoutStore) CreateSession(context.Context, auth.SessionRecord) error {
	panic("not used")
}

func (s *logoutStore) GetSessionByHash(context.Context, string, time.Time) (auth.SessionRecord, error) {
	panic("not used")
}

func (s *logoutStore) GetUserContext(context.Context, string) (auth.UserContext, error) {
	panic("not used")
}

func (s *logoutStore) TouchSession(context.Context, string, time.Time, time.Time) error {
	panic("not used")
}

func (s *logoutStore) RevokeSession(_ context.Context, sessionID string, _ time.Time) error {
	s.revokedSessionID = sessionID
	return s.revokeErr
}

func (s *logoutStore) CreatePrivacyDeleteHandoff(context.Context, string, string, string, string, time.Time) (auth.PrivacyDeleteHandoff, error) {
	panic("not used")
}
