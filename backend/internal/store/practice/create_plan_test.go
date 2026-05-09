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

func TestSQLRepositoryCreatePlanWritesPlanAndAuditInOneTransaction(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()
	repo := NewSQLRepository(db)
	now := time.Date(2026, 5, 9, 12, 0, 0, 0, time.UTC)
	in := validCreatePlanStoreInput(now)

	mock.ExpectBegin()
	mock.ExpectQuery(`insert into practice_plans`).
		WithArgs(
			in.PlanID,
			in.UserID,
			in.TargetJobID,
			string(in.Goal),
			string(in.Mode),
			string(in.InterviewerPersona),
			in.Difficulty,
			in.Language,
			in.TimeBudgetMinutes,
			in.QuestionBudget,
			in.ResumeAssetID,
			sqlmock.AnyArg(),
			in.Now,
		).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "target_job_id", "goal", "mode", "interviewer_persona", "difficulty",
			"language", "time_budget_minutes", "question_budget", "status", "created_at",
		}).AddRow(
			in.PlanID,
			in.TargetJobID,
			string(in.Goal),
			string(in.Mode),
			string(in.InterviewerPersona),
			in.Difficulty,
			in.Language,
			in.TimeBudgetMinutes,
			in.QuestionBudget,
			"ready",
			in.Now,
		))
	mock.ExpectExec(`insert into audit_events`).
		WithArgs(
			in.AuditEventID,
			in.UserID,
			in.UserID,
			in.PlanID,
			sqlmock.AnyArg(),
			in.Now,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	plan, err := repo.CreatePlan(context.Background(), in)
	if err != nil {
		t.Fatalf("CreatePlan returned error: %v", err)
	}
	if plan.ID != in.PlanID || plan.Status != "ready" || plan.CreatedAt != in.Now {
		t.Fatalf("unexpected plan: %+v", plan)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestSQLRepositoryCreatePlanReturnsPrerequisiteErrorWhenTargetOrResumeMissing(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()
	repo := NewSQLRepository(db)
	in := validCreatePlanStoreInput(time.Date(2026, 5, 9, 12, 0, 0, 0, time.UTC))

	mock.ExpectBegin()
	mock.ExpectQuery(`insert into practice_plans`).
		WithArgs(
			in.PlanID,
			in.UserID,
			in.TargetJobID,
			string(in.Goal),
			string(in.Mode),
			string(in.InterviewerPersona),
			in.Difficulty,
			in.Language,
			in.TimeBudgetMinutes,
			in.QuestionBudget,
			in.ResumeAssetID,
			sqlmock.AnyArg(),
			in.Now,
		).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectRollback()

	_, err := repo.CreatePlan(context.Background(), in)
	if !errors.Is(err, domain.ErrPlanPrerequisiteNotFound) {
		t.Fatalf("error = %v, want ErrPlanPrerequisiteNotFound", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func newMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock, func()) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	return db, mock, func() { _ = db.Close() }
}

func validCreatePlanStoreInput(now time.Time) domain.CreatePlanStoreInput {
	return domain.CreatePlanStoreInput{
		PlanID:               "01918fa0-0000-7000-8000-000000004000",
		AuditEventID:         "01918fa0-0000-7000-8000-000000004001",
		UserID:               "01918fa0-0000-7000-8000-000000000001",
		TargetJobID:          "01918fa0-0000-7000-8000-000000002000",
		ResumeAssetID:        "01918fa0-0000-7000-8000-000000001000",
		Goal:                 sharedtypes.PracticeGoalBaseline,
		Mode:                 sharedtypes.PracticeModeAssisted,
		InterviewerPersona:   sharedtypes.InterviewerRoleHiringManager,
		Difficulty:           "standard",
		Language:             "zh-CN",
		TimeBudgetMinutes:    30,
		QuestionBudget:       6,
		FocusCompetencyCodes: []string{"communication", "design-systems"},
		Now:                  now,
	}
}
