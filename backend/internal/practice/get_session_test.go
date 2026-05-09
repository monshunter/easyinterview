package practice

import (
	"context"
	"errors"
	"testing"

	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

func TestServiceGetPracticeSessionScopesByUser(t *testing.T) {
	store := &recordingPlanStore{
		getSessionRecord: SessionRecord{ID: "session-1", PlanID: "plan-1", Status: sharedtypes.SessionStatusRunning},
	}
	service := NewService(ServiceOptions{Store: store})

	session, err := service.GetPracticeSession(context.Background(), "user-1", "session-1")
	if err != nil {
		t.Fatalf("GetPracticeSession returned error: %v", err)
	}
	if session.ID != "session-1" {
		t.Fatalf("unexpected session: %+v", session)
	}
	if store.getSessionUserID != "user-1" || store.getSessionID != "session-1" {
		t.Fatalf("store was not scoped by user and session: user=%q session=%q", store.getSessionUserID, store.getSessionID)
	}
}

func TestServiceGetPracticeSessionMapsMissingRowsToPracticeSessionNotFound(t *testing.T) {
	store := &recordingPlanStore{getSessionErr: ErrSessionNotFound}
	service := NewService(ServiceOptions{Store: store})

	_, err := service.GetPracticeSession(context.Background(), "user-1", "missing-session")
	var svcErr *ServiceError
	if !errors.As(err, &svcErr) {
		t.Fatalf("expected ServiceError, got %T: %v", err, err)
	}
	if svcErr.Code != sharederrors.CodePracticeSessionNotFound {
		t.Fatalf("code = %q", svcErr.Code)
	}
}
