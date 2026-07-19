package auth_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/auth"
)

func TestGetMeReturnsMaskedCurrentUser(t *testing.T) {
	store := &meStore{user: auth.UserContext{
		ID:          "user-1",
		Email:       "candidate@example.com",
		DisplayName: "Candidate",
	}}
	service := auth.NewEmailCodeService(auth.EmailCodeServiceOptions{Store: store})
	handler := auth.NewHandler(auth.HandlerOptions{EmailCode: service})
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
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["id"] != "user-1" || body["displayName"] != "Candidate" {
		t.Fatalf("bad user context: %+v", body)
	}
	wantKeys := map[string]bool{"id": true, "email": true, "displayName": true, "profileCompletionRequired": true, "displayPreferences": true}
	if len(body) != len(wantKeys) {
		t.Fatalf("user context keys=%v, want exact four-field projection", body)
	}
	for key := range body {
		if !wantKeys[key] {
			t.Fatalf("unexpected user context field %q in %+v", key, body)
		}
	}
	if body["email"] != "candidate@example.com" {
		t.Fatalf("email = %v, want complete account email", body["email"])
	}
	if body["profileCompletionRequired"] != false {
		t.Fatalf("profile completion flag = %+v", body["profileCompletionRequired"])
	}
	if _, exists := body["emailMasked"]; exists {
		t.Fatalf("legacy emailMasked field leaked: %+v", body)
	}
}

func TestGetMeReturnsProfileCompletionRequiredForIncompleteUser(t *testing.T) {
	store := &meStore{user: auth.UserContext{
		ID:                        "user-incomplete",
		Email:                     "new-user@example.com",
		ProfileCompletionRequired: true,
	}}
	service := auth.NewEmailCodeService(auth.EmailCodeServiceOptions{Store: store})
	handler := auth.NewHandler(auth.HandlerOptions{EmailCode: service})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	req = req.WithContext(auth.ContextWithCurrentSession(req.Context(), auth.CurrentSession{
		SessionID: "session-1",
		UserID:    "user-incomplete",
		ExpiresAt: time.Now().Add(auth.SessionTTL),
	}))
	rec := httptest.NewRecorder()

	handler.GetMe(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["profileCompletionRequired"] != true {
		t.Fatalf("profile completion flag = %+v body=%s", body["profileCompletionRequired"], rec.Body.String())
	}
}

func TestGetMeReturnsAccountThemeProjection(t *testing.T) {
	store := &meStore{user: auth.UserContext{
		ID:          "user-themed",
		Email:       "themed@example.com",
		DisplayName: "Themed Candidate",
	}}
	service := auth.NewEmailCodeService(auth.EmailCodeServiceOptions{Store: store})
	handler := auth.NewHandler(auth.HandlerOptions{EmailCode: service})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	req = req.WithContext(auth.ContextWithCurrentSession(req.Context(), auth.CurrentSession{SessionID: "session-1", UserID: "user-themed"}))
	rec := httptest.NewRecorder()

	handler.GetMe(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	preferences, ok := body["displayPreferences"].(map[string]any)
	if !ok || preferences["theme"] != "ocean" || preferences["customAccent"] != nil {
		t.Fatalf("displayPreferences = %#v, want ocean with null custom accent", body["displayPreferences"])
	}
}

func TestUpdateMeAcceptsThemeOnlyAndReturnsFullContext(t *testing.T) {
	store := &meStore{user: auth.UserContext{ID: "user-1", Email: "candidate@example.com", DisplayName: "Candidate"}}
	service := auth.NewEmailCodeService(auth.EmailCodeServiceOptions{Store: store})
	handler := auth.NewHandler(auth.HandlerOptions{EmailCode: service})
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/me", strings.NewReader(`{"displayPreferences":{"theme":"plum","customAccent":null}}`))
	req = req.WithContext(auth.ContextWithCurrentSession(req.Context(), auth.CurrentSession{SessionID: "session-1", UserID: "user-1"}))
	rec := httptest.NewRecorder()

	handler.UpdateMe(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	preferences, ok := body["displayPreferences"].(map[string]any)
	if !ok || preferences["theme"] != "plum" {
		t.Fatalf("displayPreferences = %#v, want persisted plum", body["displayPreferences"])
	}
}

func TestUpdateMeRequiresSessionAndTermsThenClearsFlag(t *testing.T) {
	store := &meStore{user: auth.UserContext{
		ID:                        "user-incomplete",
		Email:                     "new-user@example.com",
		ProfileCompletionRequired: true,
	}}
	service := auth.NewEmailCodeService(auth.EmailCodeServiceOptions{Store: store})
	handler := auth.NewHandler(auth.HandlerOptions{EmailCode: service})

	unauth := httptest.NewRecorder()
	handler.UpdateMe(unauth, httptest.NewRequest(http.MethodPatch, "/api/v1/me", nil))
	if unauth.Code != http.StatusUnauthorized {
		t.Fatalf("unauth status = %d body=%s", unauth.Code, unauth.Body.String())
	}

	noTermsReq := httptest.NewRequest(http.MethodPatch, "/api/v1/me", strings.NewReader(`{"displayName":" Alice Candidate ","acceptedTerms":false}`))
	noTermsReq = noTermsReq.WithContext(auth.ContextWithCurrentSession(noTermsReq.Context(), auth.CurrentSession{SessionID: "session-1", UserID: "user-incomplete"}))
	noTerms := httptest.NewRecorder()
	handler.UpdateMe(noTerms, noTermsReq)
	if noTerms.Code != http.StatusBadRequest {
		t.Fatalf("terms status = %d body=%s", noTerms.Code, noTerms.Body.String())
	}

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/me", strings.NewReader(`{"displayName":" Alice Candidate ","acceptedTerms":true}`))
	req = req.WithContext(auth.ContextWithCurrentSession(req.Context(), auth.CurrentSession{SessionID: "session-1", UserID: "user-incomplete"}))
	rec := httptest.NewRecorder()
	handler.UpdateMe(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["displayName"] != "Alice Candidate" || body["profileCompletionRequired"] != false {
		t.Fatalf("bad completion response: %+v", body)
	}
	if len(body) != 5 {
		t.Fatalf("completion user context keys=%v, want exact five-field projection", body)
	}
}

func TestAccountHandlersWithoutSessionReturnAuthEnvelope(t *testing.T) {
	handler := auth.NewHandler(auth.HandlerOptions{})
	tests := []struct {
		name   string
		method string
		handle func(http.ResponseWriter, *http.Request)
	}{
		{name: "get me", method: http.MethodGet, handle: handler.GetMe},
		{name: "delete me", method: http.MethodDelete, handle: handler.DeleteMe},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, "/api/v1/me", nil)
			rec := httptest.NewRecorder()

			tc.handle(rec, req)

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

func (s *meStore) CreateUserByEmail(context.Context, string, string, string, time.Time) (auth.UserContext, error) {
	panic("not used")
}

func (s *meStore) FindUserByEmail(context.Context, string) (auth.UserContext, error) {
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

func (s *meStore) UpdateUserContext(_ context.Context, _ string, in auth.UpdateUserContextInput, _ time.Time) (auth.UserContext, error) {
	if in.DisplayName != nil {
		s.user.DisplayName = *in.DisplayName
		s.user.ProfileCompletionRequired = false
	}
	if in.DisplayPreferences != nil {
		s.user.DisplayPreferences = *in.DisplayPreferences
	}
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
