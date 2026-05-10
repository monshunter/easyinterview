package practice

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	domain "github.com/monshunter/easyinterview/backend/internal/practice"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

func TestSQLRepositoryGetPlanScopesByUser(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()
	repo := NewSQLRepository(db)
	createdAt := time.Date(2026, 4, 28, 12, 0, 0, 0, time.UTC)

	mock.ExpectQuery(`select id, target_job_id, goal, mode, interviewer_persona, difficulty`).
		WithArgs("user-1", "plan-1").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "target_job_id", "goal", "mode", "interviewer_persona", "difficulty",
			"language", "time_budget_minutes", "question_budget", "status", "created_at",
		}).AddRow(
			"plan-1",
			"target-1",
			string(sharedtypes.PracticeGoalBaseline),
			string(sharedtypes.PracticeModeAssisted),
			string(sharedtypes.InterviewerRoleHiringManager),
			"standard",
			"zh-CN",
			30,
			6,
			"ready",
			createdAt,
		))

	plan, err := repo.GetPlan(context.Background(), "user-1", "plan-1")
	if err != nil {
		t.Fatalf("GetPlan returned error: %v", err)
	}
	if plan.ID != "plan-1" || plan.TargetJobID != "target-1" || plan.CreatedAt != createdAt {
		t.Fatalf("unexpected plan: %+v", plan)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestSQLRepositoryGetPlanMapsNoRowsToNotFound(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()
	repo := NewSQLRepository(db)

	mock.ExpectQuery(`select id, target_job_id, goal, mode, interviewer_persona, difficulty`).
		WithArgs("user-b", "plan-owned-by-user-a").
		WillReturnError(sql.ErrNoRows)

	_, err := repo.GetPlan(context.Background(), "user-b", "plan-owned-by-user-a")
	if !errors.Is(err, domain.ErrPlanNotFound) {
		t.Fatalf("error = %v, want ErrPlanNotFound", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}
