package practice

import (
	"context"
	"errors"
	"testing"

	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

func TestServiceListPracticeSessionsScopesAndValidatesFilters(t *testing.T) {
	store := &recordingPlanStore{
		listSessionsResult: ListSessionsResult{
			Items: []SessionRecord{
				{ID: "session-1", PlanID: "plan-1", Status: sharedtypes.SessionStatusCompleted},
			},
			PageSize: 5,
		},
	}
	service := NewService(ServiceOptions{Store: store})

	result, err := service.ListPracticeSessions(context.Background(), ListSessionsRequest{
		UserID:      " user-1 ",
		TargetJobID: " target-1 ",
		Status:      sharedtypes.SessionStatusCompleted,
		PageSize:    5,
		Cursor:      " cursor-1 ",
	})
	if err != nil {
		t.Fatalf("ListPracticeSessions returned error: %v", err)
	}
	if len(result.Items) != 1 || result.PageSize != 5 {
		t.Fatalf("unexpected result: %+v", result)
	}
	if store.listSessionsInput.UserID != "user-1" ||
		store.listSessionsInput.TargetJobID != "target-1" ||
		store.listSessionsInput.Status != sharedtypes.SessionStatusCompleted ||
		store.listSessionsInput.PageSize != 5 ||
		store.listSessionsInput.Cursor != "cursor-1" {
		t.Fatalf("store input not normalized: %+v", store.listSessionsInput)
	}
}

func TestServiceListPracticeSessionsRejectsInvalidStatus(t *testing.T) {
	service := NewService(ServiceOptions{Store: &recordingPlanStore{}})

	_, err := service.ListPracticeSessions(context.Background(), ListSessionsRequest{
		UserID: "user-1",
		Status: sharedtypes.SessionStatus("bogus"),
	})

	var svcErr *ServiceError
	if !errors.As(err, &svcErr) {
		t.Fatalf("expected ServiceError, got %T: %v", err, err)
	}
	if svcErr.Code != sharederrors.CodeValidationFailed || svcErr.Details["field"] != "status" {
		t.Fatalf("unexpected status error: %+v", svcErr)
	}
}
