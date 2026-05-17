package practice

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	domain "github.com/monshunter/easyinterview/backend/internal/practice"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

func TestSQLRepositoryListSessionsFiltersByUserTargetAndStatus(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()
	repo := NewSQLRepository(db)
	createdAt := time.Date(2026, 5, 15, 13, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2026, 5, 15, 14, 12, 0, 0, time.UTC)

	mock.ExpectQuery(`(?s)from practice_sessions s.*where s.user_id = \$1.*s.target_job_id = \$2.*s.status = \$3.*order by s.updated_at desc, s.id desc`).
		WithArgs("user-1", "target-1", string(sharedtypes.SessionStatusCompleted), 6).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "plan_id", "target_job_id", "status", "language", "hints_enabled", "turn_count", "created_at", "updated_at",
		}).AddRow(
			"session-1",
			"plan-1",
			"target-1",
			string(sharedtypes.SessionStatusCompleted),
			"zh-CN",
			true,
			8,
			createdAt,
			updatedAt,
		))

	result, err := repo.ListSessions(context.Background(), domain.ListSessionsInput{
		UserID:      "user-1",
		TargetJobID: "target-1",
		Status:      sharedtypes.SessionStatusCompleted,
		PageSize:    5,
	})
	if err != nil {
		t.Fatalf("ListSessions returned error: %v", err)
	}
	if len(result.Items) != 1 || result.Items[0].ID != "session-1" || result.Items[0].Status != sharedtypes.SessionStatusCompleted {
		t.Fatalf("unexpected sessions: %+v", result.Items)
	}
	if result.PageSize != 5 || result.HasMore {
		t.Fatalf("unexpected page info: %+v", result)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}
