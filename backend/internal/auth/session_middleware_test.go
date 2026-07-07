package auth_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/auth"
)

func TestSessionMiddlewareResolvesActiveSessionAndTouchesUpdatedAt(t *testing.T) {
	now := time.Date(2026, 5, 6, 10, 30, 0, 0, time.UTC)
	store := &sessionStore{session: auth.SessionRecord{
		ID:        "session-1",
		UserID:    "user-1",
		Status:    auth.SessionStatusActive,
		ExpiresAt: now.Add(auth.SessionTTL),
	}}
	service := auth.NewEmailCodeService(auth.EmailCodeServiceOptions{
		Store:               store,
		SessionCookieSecret: "session-secret",
		Now:                 func() time.Time { return now },
	})
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		current, ok := auth.CurrentSessionFromContext(r.Context())
		if !ok {
			t.Fatal("missing current session in context")
		}
		if current.UserID != "user-1" || current.SessionID != "session-1" {
			t.Fatalf("current session = %+v", current)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	handler := auth.SessionMiddleware(service, "getMe", next)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if !called || rec.Code != http.StatusNoContent {
		t.Fatalf("middleware did not pass active session: called=%v status=%d body=%s", called, rec.Code, rec.Body.String())
	}
	if store.lookupHash == "" || store.lookupHash == "raw-session-token" {
		t.Fatalf("session lookup must use hash, got %q", store.lookupHash)
	}
	if store.touchedSessionID != "session-1" || !store.touchNow.Equal(now) || !store.touchExpiresAt.Equal(now.Add(auth.SessionTTL)) {
		t.Fatalf("touch = id:%s now:%s exp:%s", store.touchedSessionID, store.touchNow, store.touchExpiresAt)
	}
}

func TestSessionMiddlewareRejectsMissingInvalidRevokedOrExpiredSession(t *testing.T) {
	for name, setup := range map[string]struct {
		cookie string
		err    error
	}{
		"missing": {},
		"invalid": {cookie: "bad", err: auth.ErrSessionInvalid},
		"revoked": {cookie: "revoked", err: auth.ErrSessionRevoked},
		"expired": {cookie: "expired", err: auth.ErrSessionExpired},
	} {
		t.Run(name, func(t *testing.T) {
			store := &sessionStore{lookupErr: setup.err}
			service := auth.NewEmailCodeService(auth.EmailCodeServiceOptions{
				Store:               store,
				SessionCookieSecret: "session-secret",
				Now:                 func() time.Time { return time.Date(2026, 5, 6, 10, 30, 0, 0, time.UTC) },
			})
			handler := auth.SessionMiddleware(service, "getMe", http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
				t.Fatal("protected handler must not run")
			}))
			req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
			if setup.cookie != "" {
				req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: setup.cookie})
			}
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
			}
			var body map[string]map[string]any
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatalf("bad error JSON: %v", err)
			}
			if body["error"]["code"] != "AUTH_UNAUTHORIZED" {
				t.Fatalf("error = %+v", body)
			}
		})
	}
}

func TestSessionMiddlewareTreatsTouchLostRaceAsAuthState(t *testing.T) {
	now := time.Date(2026, 5, 6, 22, 0, 0, 0, time.UTC)
	store := &sessionStore{
		session: auth.SessionRecord{
			ID:        "session-1",
			UserID:    "user-1",
			Status:    auth.SessionStatusActive,
			ExpiresAt: now.Add(auth.SessionTTL),
		},
		touchErr: sql.ErrNoRows,
	}
	service := auth.NewEmailCodeService(auth.EmailCodeServiceOptions{
		Store:               store,
		SessionCookieSecret: "session-secret",
		Now:                 func() time.Time { return now },
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
	called := false
	optionalLogout := auth.SessionMiddleware(service, "logout", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusNoContent)
	}))
	rec := httptest.NewRecorder()

	optionalLogout.ServeHTTP(rec, req)

	if !called || rec.Code != http.StatusNoContent {
		t.Fatalf("optional logout should remain idempotent after touch lost race: called=%v status=%d body=%s", called, rec.Code, rec.Body.String())
	}

	protected := auth.SessionMiddleware(service, "getMe", http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("protected handler must not run after touch lost race")
	}))
	protectedRec := httptest.NewRecorder()
	protectedReq := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	protectedReq.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})

	protected.ServeHTTP(protectedRec, protectedReq)

	if protectedRec.Code != http.StatusUnauthorized {
		t.Fatalf("protected touch lost race status = %d body=%s", protectedRec.Code, protectedRec.Body.String())
	}
}

type sessionStore struct {
	session          auth.SessionRecord
	lookupErr        error
	lookupHash       string
	touchedSessionID string
	touchNow         time.Time
	touchExpiresAt   time.Time
	touchErr         error
}

func (s *sessionStore) CountRecentChallenges(context.Context, string, string, time.Time) (int, error) {
	return 0, nil
}

func (s *sessionStore) CreateChallenge(context.Context, auth.ChallengeRecord) error {
	panic("not used")
}

func (s *sessionStore) ConsumeChallenge(context.Context, string, time.Time) (auth.ChallengeRecord, error) {
	panic("not used")
}

func (s *sessionStore) CreateUserByEmail(context.Context, string, string, string, time.Time) (auth.UserContext, error) {
	panic("not used")
}

func (s *sessionStore) FindUserByEmail(context.Context, string) (auth.UserContext, error) {
	panic("not used")
}

func (s *sessionStore) CreateSession(context.Context, auth.SessionRecord) error {
	panic("not used")
}

func (s *sessionStore) GetSessionByHash(_ context.Context, hash string, _ time.Time) (auth.SessionRecord, error) {
	s.lookupHash = hash
	if s.lookupErr != nil {
		return auth.SessionRecord{}, s.lookupErr
	}
	return s.session, nil
}

func (s *sessionStore) GetUserContext(context.Context, string) (auth.UserContext, error) {
	panic("not used")
}

func (s *sessionStore) TouchSession(_ context.Context, sessionID string, now time.Time, expiresAt time.Time) error {
	s.touchedSessionID = sessionID
	s.touchNow = now
	s.touchExpiresAt = expiresAt
	return s.touchErr
}

func (s *sessionStore) RevokeSession(context.Context, string, time.Time) error {
	panic("not used")
}

func (s *sessionStore) CreatePrivacyDeleteHandoff(context.Context, string, string, string, string, time.Time) (auth.PrivacyDeleteHandoff, error) {
	panic("not used")
}
