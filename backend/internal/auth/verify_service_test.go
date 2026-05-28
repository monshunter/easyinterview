package auth_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/auth"
)

func TestVerifyAuthEmailChallengeConsumesTokenAndSetsSessionCookie(t *testing.T) {
	store := &verifyStore{
		challenge: auth.ChallengeRecord{
			ID:        "challenge-1",
			Email:     "candidate@example.com",
			ExpiresAt: time.Date(2026, 5, 6, 10, 30, 0, 0, time.UTC),
		},
		user: auth.UserContext{
			ID:                        "018f2a40-0000-7000-9000-000000000100",
			Email:                     "candidate@example.com",
			DisplayName:               "Candidate",
			UILanguage:                "zh-CN",
			PreferredPracticeLanguage: "en",
			AnalyticsOptIn:            true,
		},
	}
	now := time.Date(2026, 5, 6, 10, 15, 0, 0, time.UTC)
	service := auth.NewPasswordlessService(auth.PasswordlessServiceOptions{
		Store:                 store,
		SessionTokenGenerator: fixedTokenGenerator("raw-session-token"),
		ChallengePepper:       "pepper",
		SessionCookieSecret:   "session-secret",
		Now:                   func() time.Time { return now },
		NewID:                 fixedIDs("018f2a40-0000-7000-9000-000000000101"),
	})
	handler := auth.NewHandler(auth.HandlerOptions{Passwordless: service})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/email/verify?token=123456", nil)
	req.RemoteAddr = "203.0.113.30:5588"
	req.Header.Set("User-Agent", "unit-test-agent")
	rec := httptest.NewRecorder()

	handler.VerifyAuthEmailChallenge(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if store.consumedTokenHash == "" || store.consumedTokenHash == "123456" {
		t.Fatalf("challenge token must be looked up by hash, got %q", store.consumedTokenHash)
	}
	if store.session.SessionHash == "" || store.session.SessionHash == "raw-session-token" {
		t.Fatalf("session must be stored by hash, got %q", store.session.SessionHash)
	}
	if !store.session.ExpiresAt.Equal(now.Add(auth.SessionTTL)) {
		t.Fatalf("session expiresAt = %s", store.session.ExpiresAt)
	}
	cookies := rec.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("cookies = %#v", cookies)
	}
	cookie := cookies[0]
	if cookie.Name != auth.SessionCookieName {
		t.Fatalf("cookie name = %q", cookie.Name)
	}
	if cookie.Value == "" || cookie.Value == store.session.SessionHash {
		t.Fatalf("cookie must contain opaque raw session token, got %q", cookie.Value)
	}
	if !cookie.HttpOnly || !cookie.Secure || cookie.SameSite != http.SameSiteLaxMode || cookie.Path != "/" {
		t.Fatalf("cookie attributes = %#v", cookie)
	}
	if contains(rec.Body.String(), "raw-session-token") {
		t.Fatalf("response body leaked session token: %s", rec.Body.String())
	}
	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body["userId"] != store.user.ID || body["sessionExpiresAt"] == "" {
		t.Fatalf("bad session response: %+v", body)
	}
}

func TestSessionCookiePolicyAllowsDevInsecureButKeepsProdSecure(t *testing.T) {
	prod := auth.CookiePolicyForAppEnv("prod")
	if !prod.Secure {
		t.Fatalf("prod cookie policy must be secure: %#v", prod)
	}
	staging := auth.CookiePolicyForAppEnv("staging")
	if !staging.Secure {
		t.Fatalf("staging cookie policy must be secure: %#v", staging)
	}
	dev := auth.CookiePolicyForAppEnv("dev")
	if dev.Secure {
		t.Fatalf("dev cookie policy must allow explicit insecure local cookie: %#v", dev)
	}
}

func TestVerifyAuthEmailChallengeRejectsInvalidExpiredOrConsumedToken(t *testing.T) {
	for name, err := range map[string]error{
		"invalid":  auth.ErrChallengeInvalid,
		"expired":  auth.ErrChallengeExpired,
		"consumed": auth.ErrChallengeConsumed,
	} {
		t.Run(name, func(t *testing.T) {
			store := &verifyStore{consumeErr: err}
			service := auth.NewPasswordlessService(auth.PasswordlessServiceOptions{
				Store:                 store,
				SessionTokenGenerator: fixedTokenGenerator("raw-session-token"),
				ChallengePepper:       "pepper",
				SessionCookieSecret:   "session-secret",
				Now:                   func() time.Time { return time.Date(2026, 5, 6, 10, 15, 0, 0, time.UTC) },
				NewID:                 fixedIDs("018f2a40-0000-7000-9000-000000000102"),
			})
			handler := auth.NewHandler(auth.HandlerOptions{Passwordless: service})
			req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/email/verify?token=000000", nil)
			rec := httptest.NewRecorder()

			handler.VerifyAuthEmailChallenge(rec, req)

			if rec.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
			}
			if len(rec.Result().Cookies()) != 0 {
				t.Fatalf("invalid verify must not set cookie: %#v", rec.Result().Cookies())
			}
		})
	}
}

func TestVerifyNewEmailCreatesIncompleteUserAndSession(t *testing.T) {
	store := &verifyStore{
		challenge: auth.ChallengeRecord{
			ID:        "challenge-new-email",
			Email:     "candidate@example.com",
			Purpose:   auth.ChallengePurposeLogin,
			ExpiresAt: time.Date(2026, 5, 28, 10, 30, 0, 0, time.UTC),
		},
		user: auth.UserContext{
			ID:                        "018f2a40-0000-7000-9000-000000000200",
			Email:                     "candidate@example.com",
			UILanguage:                "zh-CN",
			PreferredPracticeLanguage: "en",
			AnalyticsOptIn:            true,
			ProfileCompletionRequired: true,
		},
		findErr: auth.ErrUserNotFound,
	}
	now := time.Date(2026, 5, 28, 10, 15, 0, 0, time.UTC)
	service := auth.NewPasswordlessService(auth.PasswordlessServiceOptions{
		Store:                 store,
		SessionTokenGenerator: fixedTokenGenerator("raw-session-token"),
		ChallengePepper:       "pepper",
		SessionCookieSecret:   "session-secret",
		Now:                   func() time.Time { return now },
		NewID:                 fixedIDs("018f2a40-0000-7000-9000-000000000201", "018f2a40-0000-7000-9000-000000000202"),
	})

	if _, err := service.VerifyEmailChallenge(context.Background(), auth.VerifyEmailChallengeInput{Token: "123456"}); err != nil {
		t.Fatalf("VerifyEmailChallenge: %v", err)
	}
	if store.createdEmail != "candidate@example.com" || store.createdName != "" {
		t.Fatalf("new email create user = email %q displayName %q", store.createdEmail, store.createdName)
	}
	if store.session.ID == "" || store.session.UserID != store.user.ID {
		t.Fatalf("new email must mint session: %+v", store.session)
	}
}

func TestVerifyExistingEmailLogsInWithoutCreatingUser(t *testing.T) {
	store := &verifyStore{
		challenge: auth.ChallengeRecord{
			ID:        "challenge-existing",
			Email:     "candidate@example.com",
			Purpose:   auth.ChallengePurposeLogin,
			ExpiresAt: time.Date(2026, 5, 28, 10, 30, 0, 0, time.UTC),
		},
		user: auth.UserContext{
			ID:                        "018f2a40-0000-7000-9000-000000000210",
			Email:                     "candidate@example.com",
			DisplayName:               "Alice Candidate",
			UILanguage:                "zh-CN",
			PreferredPracticeLanguage: "en",
			AnalyticsOptIn:            true,
		},
	}
	service := auth.NewPasswordlessService(auth.PasswordlessServiceOptions{
		Store:                 store,
		SessionTokenGenerator: fixedTokenGenerator("raw-session-token"),
		ChallengePepper:       "pepper",
		SessionCookieSecret:   "session-secret",
		Now:                   func() time.Time { return time.Date(2026, 5, 28, 10, 15, 0, 0, time.UTC) },
		NewID:                 fixedIDs("018f2a40-0000-7000-9000-000000000211"),
	})

	if _, err := service.VerifyEmailChallenge(context.Background(), auth.VerifyEmailChallengeInput{Token: "123456"}); err != nil {
		t.Fatalf("VerifyEmailChallenge: %v", err)
	}
	if store.createdEmail != "" {
		t.Fatalf("existing email must not create user, createdEmail=%q", store.createdEmail)
	}
	if store.foundEmail != "candidate@example.com" {
		t.Fatalf("existing email was not looked up, foundEmail=%q", store.foundEmail)
	}
}

type verifyStore struct {
	challenge         auth.ChallengeRecord
	user              auth.UserContext
	session           auth.SessionRecord
	consumeErr        error
	createErr         error
	findErr           error
	consumedTokenHash string
	createdEmail      string
	createdName       string
	foundEmail        string
}

func (s *verifyStore) CountRecentChallenges(context.Context, string, string, time.Time) (int, error) {
	return 0, nil
}

func (s *verifyStore) CreateChallenge(context.Context, auth.ChallengeRecord) error {
	panic("not used")
}

func (s *verifyStore) ConsumeChallenge(_ context.Context, tokenHash string, _ time.Time) (auth.ChallengeRecord, error) {
	s.consumedTokenHash = tokenHash
	if s.consumeErr != nil {
		return auth.ChallengeRecord{}, s.consumeErr
	}
	return s.challenge, nil
}

func (s *verifyStore) CreateUserByEmail(_ context.Context, email string, displayName string, _ string, _ time.Time) (auth.UserContext, error) {
	if s.createErr != nil {
		return auth.UserContext{}, s.createErr
	}
	s.createdEmail = email
	s.createdName = displayName
	return s.user, nil
}

func (s *verifyStore) FindUserByEmail(_ context.Context, email string) (auth.UserContext, error) {
	if s.findErr != nil {
		return auth.UserContext{}, s.findErr
	}
	s.foundEmail = email
	return s.user, nil
}

func (s *verifyStore) CreateSession(_ context.Context, rec auth.SessionRecord) error {
	s.session = rec
	return nil
}

func (s *verifyStore) GetSessionByHash(context.Context, string, time.Time) (auth.SessionRecord, error) {
	panic("not used")
}

func (s *verifyStore) GetUserContext(context.Context, string) (auth.UserContext, error) {
	panic("not used")
}

func (s *verifyStore) TouchSession(context.Context, string, time.Time, time.Time) error {
	panic("not used")
}

func (s *verifyStore) RevokeSession(context.Context, string, time.Time) error {
	panic("not used")
}

func (s *verifyStore) CreatePrivacyDeleteHandoff(context.Context, string, string, string, string, time.Time) (auth.PrivacyDeleteHandoff, error) {
	panic("not used")
}
