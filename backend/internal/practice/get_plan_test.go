package practice

import (
	"context"
	"errors"
	"testing"

	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
)

func TestServiceGetPracticePlanScopesByUser(t *testing.T) {
	store := &recordingPlanStore{
		getRecord: PlanRecord{ID: "plan-1", TargetJobID: "target-1", Status: "ready"},
	}
	service := NewService(ServiceOptions{Store: store})

	plan, err := service.GetPracticePlan(context.Background(), "user-1", "plan-1")
	if err != nil {
		t.Fatalf("GetPracticePlan returned error: %v", err)
	}
	if plan.ID != "plan-1" {
		t.Fatalf("unexpected plan: %+v", plan)
	}
	if store.getUserID != "user-1" || store.getPlanID != "plan-1" {
		t.Fatalf("store was not scoped by user and plan: user=%q plan=%q", store.getUserID, store.getPlanID)
	}
}

func TestServiceGetPracticePlanMapsMissingRowsToPracticePlanNotFound(t *testing.T) {
	store := &recordingPlanStore{getErr: ErrPlanNotFound}
	service := NewService(ServiceOptions{Store: store})

	_, err := service.GetPracticePlan(context.Background(), "user-1", "missing-plan")
	var svcErr *ServiceError
	if !errors.As(err, &svcErr) {
		t.Fatalf("expected ServiceError, got %T: %v", err, err)
	}
	if svcErr.Code != sharederrors.CodePracticePlanNotFound {
		t.Fatalf("code = %q", svcErr.Code)
	}
}
