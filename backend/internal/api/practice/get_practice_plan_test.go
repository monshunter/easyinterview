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

func TestGetPracticePlanReturnsUserScopedPlan(t *testing.T) {
	service := &fakePlanService{getRecord: fixturePlanRecord()}
	handler := newTestHandler(service)

	rec := httptest.NewRecorder()
	handler.GetPracticePlan(rec, newAuthenticatedGETRequest("user-1"), fixturePlanRecord().ID)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var out api.PracticePlan
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode PracticePlan: %v", err)
	}
	if out.Id != fixturePlanRecord().ID || out.Status != "ready" {
		t.Fatalf("unexpected response: %+v", out)
	}
	if service.getUserID != "user-1" || service.getPlanID != fixturePlanRecord().ID {
		t.Fatalf("service was not called with user scoped lookup: user=%q plan=%q", service.getUserID, service.getPlanID)
	}
}

func TestGetPracticePlanReturns404WithoutLeakingCrossUserExistence(t *testing.T) {
	service := &fakePlanService{
		getErr: &domain.ServiceError{
			Code:    sharederrors.CodePracticePlanNotFound,
			Message: "practice plan not found",
		},
	}
	handler := newTestHandler(service)

	rec := httptest.NewRecorder()
	handler.GetPracticePlan(rec, newAuthenticatedGETRequest("user-b"), "plan-owned-by-user-a")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	assertAPIError(t, rec, sharederrors.CodePracticePlanNotFound, false)
	if service.getUserID != "user-b" {
		t.Fatalf("lookup must use requesting user, got %q", service.getUserID)
	}
}

func TestGetPracticePlanFixtureParity(t *testing.T) {
	fixture := loadGetPlanFixture(t)
	for _, name := range []string{"default", "archived"} {
		t.Run(name, func(t *testing.T) {
			scenario := fixture.Scenarios[name]
			service := &fakePlanService{getRecord: planRecordFromFixture(scenario.Response.Body)}
			handler := newTestHandler(service)
			rec := httptest.NewRecorder()
			handler.GetPracticePlan(rec, newAuthenticatedGETRequest("user-1"), scenario.Response.Body.Id)
			if rec.Code != scenario.Response.Status {
				t.Fatalf("status = %d, want %d; body=%s", rec.Code, scenario.Response.Status, rec.Body.String())
			}
			assertJSONEqual(t, mustJSON(t, scenario.Response.Body), rec.Body.Bytes())
		})
	}

	t.Run("not-found", func(t *testing.T) {
		scenario := fixture.Scenarios["not-found"]
		service := &fakePlanService{
			getErr: &domain.ServiceError{Code: sharederrors.CodePracticePlanNotFound, Message: "practice plan not found"},
		}
		handler := newTestHandler(service)
		rec := httptest.NewRecorder()
		handler.GetPracticePlan(rec, newAuthenticatedGETRequest("user-1"), "missing-plan")
		if rec.Code != scenario.Response.Status {
			t.Fatalf("status = %d, want %d; body=%s", rec.Code, scenario.Response.Status, rec.Body.String())
		}
		assertAPIError(t, rec, sharederrors.CodePracticePlanNotFound, false)
	})
}

func newAuthenticatedGETRequest(userID string) *http.Request {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/practice/plans/plan-1", nil)
	return req.WithContext(contextWithUser(req.Context(), userID))
}

type getPlanFixture struct {
	Scenarios map[string]struct {
		Response struct {
			Status int              `json:"status"`
			Body   api.PracticePlan `json:"body"`
		} `json:"response"`
	} `json:"scenarios"`
}

func loadGetPlanFixture(t *testing.T) getPlanFixture {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("..", "..", "..", "..", "openapi", "fixtures", "PracticePlans", "getPracticePlan.json"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	var fixture getPlanFixture
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatalf("decode fixture: %v", err)
	}
	return fixture
}
