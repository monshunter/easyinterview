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

func TestDeleteMeSoftDeletesUserRevokesAllSessionsAndCreatesPrivacyHandoff(t *testing.T) {
	now := time.Date(2026, 5, 6, 11, 0, 0, 0, time.UTC)
	store := &deleteMeStore{handoff: auth.PrivacyDeleteHandoff{
		PrivacyRequestID: "privacy-request-1",
		JobID:            "job-1",
		CreatedAt:        now,
		UpdatedAt:        now,
	}}
	service := auth.NewPasswordlessService(auth.PasswordlessServiceOptions{
		Store: store,
		Now:   func() time.Time { return now },
		NewID: fixedIDs("privacy-request-1", "job-1"),
	})
	handler := auth.NewHandler(auth.HandlerOptions{Passwordless: service})

	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/me", nil)
		req.Header.Set("Idempotency-Key", "delete-key-1")
		req = req.WithContext(auth.ContextWithCurrentSession(req.Context(), auth.CurrentSession{
			SessionID: "session-1",
			UserID:    "user-1",
		}))
		rec := httptest.NewRecorder()

		handler.DeleteMe(rec, req)

		if rec.Code != http.StatusAccepted {
			t.Fatalf("attempt %d status = %d body=%s", i, rec.Code, rec.Body.String())
		}
		var body map[string]any
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if body["privacyRequestId"] != "privacy-request-1" {
			t.Fatalf("privacyRequestId = %+v", body)
		}
		job := body["job"].(map[string]any)
		if job["id"] != "job-1" || job["jobType"] != "privacy_delete" || job["resourceType"] != "privacy_request" {
			t.Fatalf("bad job = %+v", job)
		}
		cookies := rec.Result().Cookies()
		if len(cookies) != 1 || cookies[0].Name != auth.SessionCookieName || cookies[0].MaxAge >= 0 || !cookies[0].Secure {
			t.Fatalf("deleteMe clear cookie = %#v", cookies)
		}
	}
	if store.revokedSessionID != "" {
		t.Fatalf("DeleteMe must let the handoff store revoke all user sessions atomically, got single-session revoke %q", store.revokedSessionID)
	}
	if store.softDeletedUserID != "user-1" {
		t.Fatalf("soft-deleted user = %q", store.softDeletedUserID)
	}
	if store.revokedAllSessionsUserID != "user-1" {
		t.Fatalf("revoked all sessions for user = %q", store.revokedAllSessionsUserID)
	}
	if store.lastIdempotencyKey != "delete-key-1" {
		t.Fatalf("idempotency key = %q", store.lastIdempotencyKey)
	}
	if store.createCalls != 2 {
		t.Fatalf("handler should ask store for idempotent handoff each time, calls=%d", store.createCalls)
	}
}

func TestDeleteMeWithoutSessionReturnsAuthEnvelope(t *testing.T) {
	handler := auth.NewHandler(auth.HandlerOptions{})
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/me", nil)
	rec := httptest.NewRecorder()

	handler.DeleteMe(rec, req)

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

type deleteMeStore struct {
	handoff                  auth.PrivacyDeleteHandoff
	revokedSessionID         string
	softDeletedUserID        string
	revokedAllSessionsUserID string
	lastIdempotencyKey       string
	createCalls              int
}

func (s *deleteMeStore) CountRecentChallenges(context.Context, string, string, time.Time) (int, error) {
	return 0, nil
}

func (s *deleteMeStore) CreateChallenge(context.Context, auth.ChallengeRecord) error {
	panic("not used")
}

func (s *deleteMeStore) ConsumeChallenge(context.Context, string, time.Time) (auth.ChallengeRecord, error) {
	panic("not used")
}

func (s *deleteMeStore) FindOrCreateUserByEmail(context.Context, string, string, time.Time) (auth.UserContext, error) {
	panic("not used")
}

func (s *deleteMeStore) CreateSession(context.Context, auth.SessionRecord) error {
	panic("not used")
}

func (s *deleteMeStore) GetSessionByHash(context.Context, string, time.Time) (auth.SessionRecord, error) {
	panic("not used")
}

func (s *deleteMeStore) GetUserContext(context.Context, string) (auth.UserContext, error) {
	panic("not used")
}

func (s *deleteMeStore) TouchSession(context.Context, string, time.Time, time.Time) error {
	panic("not used")
}

func (s *deleteMeStore) RevokeSession(_ context.Context, sessionID string, _ time.Time) error {
	s.revokedSessionID = sessionID
	return nil
}

func (s *deleteMeStore) CreatePrivacyDeleteHandoff(_ context.Context, userID string, idempotencyKey string, privacyRequestID string, jobID string, now time.Time) (auth.PrivacyDeleteHandoff, error) {
	s.createCalls++
	s.lastIdempotencyKey = idempotencyKey
	s.softDeletedUserID = userID
	s.revokedAllSessionsUserID = userID
	if s.handoff.PrivacyRequestID == "" {
		s.handoff = auth.PrivacyDeleteHandoff{PrivacyRequestID: privacyRequestID, JobID: jobID, CreatedAt: now, UpdatedAt: now}
	}
	return s.handoff, nil
}
