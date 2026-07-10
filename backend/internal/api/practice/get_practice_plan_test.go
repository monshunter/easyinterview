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
	if out.Id != fixturePlanRecord().ID || out.Status != "ready" || out.ResumeId != fixturePlanRecord().ResumeID {
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
	assertScopedResourceNotFound(t, rec, sharederrors.CodePracticePlanNotFound, service.getUserID, "user-b")
}

func newAuthenticatedGETRequest(userID string) *http.Request {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/practice/plans/plan-1", nil)
	return req.WithContext(contextWithUser(req.Context(), userID))
}
