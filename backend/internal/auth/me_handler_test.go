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

func TestGetMeReturnsMaskedCurrentUser(t *testing.T) {
	store := &meStore{user: auth.UserContext{
		ID:                        "user-1",
		Email:                     "candidate@example.com",
		DisplayName:               "Candidate",
		UILanguage:                "zh-CN",
		PreferredPracticeLanguage: "en",
		AnalyticsOptIn:            true,
	}}
	service := auth.NewPasswordlessService(auth.PasswordlessServiceOptions{Store: store})
	handler := auth.NewHandler(auth.HandlerOptions{Passwordless: service})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	req = req.WithContext(auth.ContextWithCurrentSession(req.Context(), auth.CurrentSession{
		SessionID: "session-1",
		UserID:    "user-1",
		ExpiresAt: time.Now().Add(auth.SessionTTL),
	}))
	rec := httptest.NewRecorder()

	handler.GetMe(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["id"] != "user-1" || body["displayName"] != "Candidate" || body["uiLanguage"] != "zh-CN" || body["preferredPracticeLanguage"] != "en" {
		t.Fatalf("bad user context: %+v", body)
	}
	if body["emailMasked"] == "" || body["emailMasked"] == "candidate@example.com" {
		t.Fatalf("email was not masked: %+v", body)
	}
	if contains(rec.Body.String(), "candidate@example.com") {
		t.Fatalf("full email leaked: %s", rec.Body.String())
	}
}

func TestGetMeWithoutSessionReturnsAuthEnvelope(t *testing.T) {
	handler := auth.NewHandler(auth.HandlerOptions{})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	rec := httptest.NewRecorder()

	handler.GetMe(rec, req)

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
}

type meStore struct {
	user auth.UserContext
}

func (s *meStore) CountRecentChallenges(context.Context, string, string, time.Time) (int, error) {
	return 0, nil
}

func (s *meStore) CreateChallenge(context.Context, auth.ChallengeRecord) error {
	panic("not used")
}

func (s *meStore) ConsumeChallenge(context.Context, string, time.Time) (auth.ChallengeRecord, error) {
	panic("not used")
}

func (s *meStore) FindOrCreateUserByEmail(context.Context, string, string, time.Time) (auth.UserContext, error) {
	panic("not used")
}

func (s *meStore) CreateSession(context.Context, auth.SessionRecord) error {
	panic("not used")
}

func (s *meStore) GetSessionByHash(context.Context, string, time.Time) (auth.SessionRecord, error) {
	panic("not used")
}

func (s *meStore) GetUserContext(context.Context, string) (auth.UserContext, error) {
	return s.user, nil
}

func (s *meStore) TouchSession(context.Context, string, time.Time, time.Time) error {
	panic("not used")
}

func (s *meStore) RevokeSession(context.Context, string, time.Time) error {
	panic("not used")
}

func (s *meStore) CreatePrivacyDeleteHandoff(context.Context, string, string, string, string, time.Time) (auth.PrivacyDeleteHandoff, error) {
	panic("not used")
}
