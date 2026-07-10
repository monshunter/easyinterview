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

type getResourceFixture struct {
	Scenarios map[string]getResourceFixtureScenario `json:"scenarios"`
}

type getResourceFixtureScenario struct {
	Response struct {
		Status int             `json:"status"`
		Body   json.RawMessage `json:"body"`
	} `json:"response"`
}

func TestGetPracticeResourcesFixtureParity(t *testing.T) {
	t.Run("plan", func(t *testing.T) {
		assertGetPracticeResourceFixtureParity(
			t,
			"PracticePlans",
			"getPracticePlan.json",
			[]string{"default", "archived"},
			"not-found",
			sharederrors.CodePracticePlanNotFound,
			invokeGetPracticePlanFixtureSuccess,
			invokeGetPracticePlanFixtureError,
		)
	})
	t.Run("session", func(t *testing.T) {
		assertGetPracticeResourceFixtureParity(
			t,
			"PracticeSessions",
			"getPracticeSession.json",
			[]string{"default", "prototype-baseline"},
			"missing-session",
			sharederrors.CodePracticeSessionNotFound,
			invokeGetPracticeSessionFixtureSuccess,
			invokeGetPracticeSessionFixtureError,
		)
	})
}

func assertGetPracticeResourceFixtureParity[T any](
	t *testing.T,
	tag string,
	operation string,
	successScenarios []string,
	errorScenario string,
	errorCode string,
	invokeSuccess func(*testing.T, *httptest.ResponseRecorder, T),
	invokeError func(*httptest.ResponseRecorder),
) {
	t.Helper()
	fixture := loadGetResourceFixture(t, tag, operation)
	assertGetResourceSuccessScenarios(t, fixture, successScenarios, func(t *testing.T, rec *httptest.ResponseRecorder, raw json.RawMessage) {
		invokeSuccess(t, rec, decodeGetResourceFixtureBody[T](t, raw))
	})
	assertGetResourceErrorScenario(t, fixture, errorScenario, errorCode, invokeError)
}

func invokeGetPracticePlanFixtureSuccess(t *testing.T, rec *httptest.ResponseRecorder, body api.PracticePlan) {
	t.Helper()
	service := &fakePlanService{getRecord: planRecordFromFixture(body)}
	newTestHandler(service).GetPracticePlan(rec, newAuthenticatedGETRequest("user-1"), body.Id)
}

func invokeGetPracticePlanFixtureError(rec *httptest.ResponseRecorder) {
	service := &fakePlanService{
		getErr: &domain.ServiceError{Code: sharederrors.CodePracticePlanNotFound, Message: "practice plan not found"},
	}
	newTestHandler(service).GetPracticePlan(rec, newAuthenticatedGETRequest("user-1"), "missing-plan")
}

func invokeGetPracticeSessionFixtureSuccess(t *testing.T, rec *httptest.ResponseRecorder, body api.PracticeSession) {
	t.Helper()
	service := &fakePlanService{getSessionRecord: sessionRecordFromFixture(body)}
	newTestHandler(service).GetPracticeSession(rec, newAuthenticatedGETRequest("user-1"), body.Id)
}

func invokeGetPracticeSessionFixtureError(rec *httptest.ResponseRecorder) {
	service := &fakePlanService{
		getSessionErr: &domain.ServiceError{Code: sharederrors.CodePracticeSessionNotFound, Message: "practice session not found"},
	}
	newTestHandler(service).GetPracticeSession(rec, newAuthenticatedGETRequest("user-1"), "missing-session")
}

func loadGetResourceFixture(t *testing.T, tag, operation string) getResourceFixture {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("..", "..", "..", "..", "openapi", "fixtures", tag, operation))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	var fixture getResourceFixture
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatalf("decode fixture: %v", err)
	}
	return fixture
}

func decodeGetResourceFixtureBody[T any](t *testing.T, raw json.RawMessage) T {
	t.Helper()
	var body T
	if err := json.Unmarshal(raw, &body); err != nil {
		t.Fatalf("decode fixture body: %v", err)
	}
	return body
}

func assertGetResourceSuccessScenarios(
	t *testing.T,
	fixture getResourceFixture,
	names []string,
	invoke func(*testing.T, *httptest.ResponseRecorder, json.RawMessage),
) {
	t.Helper()
	for _, name := range names {
		scenario, ok := fixture.Scenarios[name]
		if !ok {
			t.Fatalf("fixture scenario %q is missing", name)
		}
		t.Run(name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			invoke(t, rec, scenario.Response.Body)
			if rec.Code != scenario.Response.Status {
				t.Fatalf("status = %d, want %d; body=%s", rec.Code, scenario.Response.Status, rec.Body.String())
			}
			assertJSONEqual(t, scenario.Response.Body, rec.Body.Bytes())
		})
	}
}

func assertGetResourceErrorScenario(
	t *testing.T,
	fixture getResourceFixture,
	name string,
	wantCode string,
	invoke func(*httptest.ResponseRecorder),
) {
	t.Helper()
	scenario, ok := fixture.Scenarios[name]
	if !ok {
		t.Fatalf("fixture scenario %q is missing", name)
	}
	t.Run(name, func(t *testing.T) {
		rec := httptest.NewRecorder()
		invoke(rec)
		if rec.Code != scenario.Response.Status {
			t.Fatalf("status = %d, want %d; body=%s", rec.Code, scenario.Response.Status, rec.Body.String())
		}
		assertAPIError(t, rec, wantCode, false)
	})
}

func assertScopedResourceNotFound(
	t *testing.T,
	rec *httptest.ResponseRecorder,
	wantCode string,
	gotUserID string,
	wantUserID string,
) {
	t.Helper()
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	assertAPIError(t, rec, wantCode, false)
	if gotUserID != wantUserID {
		t.Fatalf("lookup must use requesting user, got %q", gotUserID)
	}
}
