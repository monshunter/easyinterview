package practice

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	api "github.com/monshunter/easyinterview/backend/internal/api/generated"
	domain "github.com/monshunter/easyinterview/backend/internal/practice"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
)

func TestGetPracticeSessionReturnsUserScopedSession(t *testing.T) {
	service := &fakePlanService{getSessionRecord: fixtureSessionRecord()}
	handler := newTestHandler(service)

	rec := httptest.NewRecorder()
	handler.GetPracticeSession(rec, newAuthenticatedGETRequest("user-1"), fixtureSessionRecord().ID)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var out api.PracticeSession
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode PracticeSession: %v", err)
	}
	if out.Id != fixtureSessionRecord().ID || out.CurrentTurn == nil {
		t.Fatalf("unexpected response: %+v", out)
	}
	if service.getSessionUserID != "user-1" || service.getSessionID != fixtureSessionRecord().ID {
		t.Fatalf("service was not called with user scoped lookup: user=%q session=%q", service.getSessionUserID, service.getSessionID)
	}
}

func TestGetPracticeSessionReturns404WithoutLeakingCrossUserExistence(t *testing.T) {
	service := &fakePlanService{
		getSessionErr: &domain.ServiceError{
			Code:    sharederrors.CodePracticeSessionNotFound,
			Message: "practice session not found",
		},
	}
	handler := newTestHandler(service)

	rec := httptest.NewRecorder()
	handler.GetPracticeSession(rec, newAuthenticatedGETRequest("user-b"), "session-owned-by-user-a")
	assertScopedResourceNotFound(t, rec, sharederrors.CodePracticeSessionNotFound, service.getSessionUserID, "user-b")
}
