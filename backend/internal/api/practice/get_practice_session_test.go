package practice

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	assertAPIError(t, rec, sharederrors.CodePracticeSessionNotFound, false)
	if service.getSessionUserID != "user-b" {
		t.Fatalf("lookup must use requesting user, got %q", service.getSessionUserID)
	}
}

func TestGetPracticeSessionFixtureParity(t *testing.T) {
	fixture := loadGetSessionFixture(t)
	for _, name := range []string{"default", "prototype-baseline"} {
		t.Run(name, func(t *testing.T) {
			scenario := fixture.Scenarios[name]
			service := &fakePlanService{getSessionRecord: sessionRecordFromFixture(scenario.Response.Body)}
			handler := newTestHandler(service)
			rec := httptest.NewRecorder()
			handler.GetPracticeSession(rec, newAuthenticatedGETRequest("user-1"), scenario.Response.Body.Id)
			if rec.Code != scenario.Response.Status {
				t.Fatalf("status = %d, want %d; body=%s", rec.Code, scenario.Response.Status, rec.Body.String())
			}
			assertJSONEqual(t, mustJSON(t, scenario.Response.Body), rec.Body.Bytes())
		})
	}

	t.Run("missing-session", func(t *testing.T) {
		scenario := fixture.Scenarios["missing-session"]
		service := &fakePlanService{
			getSessionErr: &domain.ServiceError{Code: sharederrors.CodePracticeSessionNotFound, Message: "practice session not found"},
		}
		handler := newTestHandler(service)
		rec := httptest.NewRecorder()
		handler.GetPracticeSession(rec, newAuthenticatedGETRequest("user-1"), "missing-session")
		if rec.Code != scenario.Response.Status {
			t.Fatalf("status = %d, want %d; body=%s", rec.Code, scenario.Response.Status, rec.Body.String())
		}
		assertAPIError(t, rec, sharederrors.CodePracticeSessionNotFound, false)
	})
}

type getSessionFixture struct {
	Scenarios map[string]struct {
		Response struct {
			Status int                 `json:"status"`
			Body   api.PracticeSession `json:"body"`
		} `json:"response"`
	} `json:"scenarios"`
}

func loadGetSessionFixture(t *testing.T) getSessionFixture {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("..", "..", "..", "..", "openapi", "fixtures", "PracticeSessions", "getPracticeSession.json"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	var fixture getSessionFixture
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatalf("decode fixture: %v", err)
	}
	return fixture
}
