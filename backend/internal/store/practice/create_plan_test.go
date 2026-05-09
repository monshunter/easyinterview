package practice

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lib/pq"
	"github.com/monshunter/easyinterview/backend/internal/middleware/idempotency"
	domain "github.com/monshunter/easyinterview/backend/internal/practice"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
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

func TestSQLRepositoryFailSessionStartMarksSessionAndIdempotencyRetryable(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()
	repo := NewSQLRepository(db)
	now := time.Date(2026, 5, 9, 12, 15, 0, 0, time.UTC)
	in := domain.FailSessionStartInput{
		IdempotencyRecordID: "idem-1",
		SessionID:           "session-1",
		UserID:              "user-1",
		ErrorCode:           sharederrors.CodeAiProviderTimeout,
		Retryable:           true,
		FailedAt:            now,
	}

	mock.ExpectBegin()
	mock.ExpectExec(`update practice_sessions`).
		WithArgs(string(sharedtypes.SessionStatusFailed), in.ErrorCode, in.FailedAt, in.SessionID, in.UserID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`update idempotency_records`).
		WithArgs(string(idempotency.StatusFailedRetry), in.ErrorCode, in.SessionID, in.FailedAt, in.IdempotencyRecordID, in.UserID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	if err := repo.FailSessionStart(context.Background(), in); err != nil {
		t.Fatalf("FailSessionStart returned error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestSQLRepositoryReserveSessionStartReusesFailedRetryableRecord(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()
	repo := NewSQLRepository(db)
	now := time.Date(2026, 5, 9, 12, 20, 0, 0, time.UTC)
	in := domain.StartSessionReservationInput{
		IdempotencyRecordID: "new-idem-ignored",
		SessionID:           "session-retry",
		UserID:              "user-1",
		PlanID:              "plan-1",
		HintsEnabled:        true,
		IdempotencyKeyHash:  "key-hash",
		RequestFingerprint:  "fingerprint",
		ExpiresAt:           now.Add(24 * time.Hour),
		Now:                 now,
	}

	mock.ExpectBegin()
	mock.ExpectExec(`select pg_advisory_xact_lock`).
		WithArgs("user-1\x00practice\x00startPracticeSession\x00key-hash").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`select id, request_fingerprint, status, resource_id::text`).
		WithArgs(in.UserID, in.IdempotencyKeyHash).
		WillReturnRows(sqlmock.NewRows([]string{"id", "request_fingerprint", "status", "resource_id"}).
			AddRow("idem-existing", in.RequestFingerprint, string(idempotency.StatusFailedRetry), nil))
	mock.ExpectExec(`update idempotency_records`).
		WithArgs(in.RequestFingerprint, string(idempotency.StatusPending), in.ExpiresAt, in.Now, "idem-existing", in.UserID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`with selected_plan`).
		WithArgs(in.SessionID, in.UserID, in.PlanID, in.HintsEnabled, in.Now).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "plan_id", "target_job_id", "goal", "mode", "interviewer_persona",
			"language", "hints_enabled", "created_at", "updated_at",
		}).AddRow(
			in.SessionID,
			in.PlanID,
			"target-1",
			string(sharedtypes.PracticeGoalBaseline),
			string(sharedtypes.PracticeModeAssisted),
			string(sharedtypes.InterviewerRoleHiringManager),
			"zh-CN",
			true,
			in.Now,
			in.Now,
		))
	mock.ExpectCommit()

	reservation, err := repo.ReserveSessionStart(context.Background(), in)
	if err != nil {
		t.Fatalf("ReserveSessionStart returned error: %v", err)
	}
	if reservation.IdempotencyRecordID != "idem-existing" || reservation.SessionID != in.SessionID {
		t.Fatalf("unexpected reservation: %+v", reservation)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestSQLRepositoryReserveSessionStartScopesIdempotencyByUser(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()
	repo := NewSQLRepository(db)
	now := time.Date(2026, 5, 9, 12, 25, 0, 0, time.UTC)
	in := domain.StartSessionReservationInput{
		IdempotencyRecordID: "idem-user-b",
		SessionID:           "session-user-b",
		UserID:              "user-b",
		PlanID:              "plan-user-b",
		IdempotencyKeyHash:  "shared-key-hash",
		RequestFingerprint:  "fingerprint-user-b",
		ExpiresAt:           now.Add(24 * time.Hour),
		Now:                 now,
	}

	mock.ExpectBegin()
	mock.ExpectExec(`select pg_advisory_xact_lock`).
		WithArgs("user-b\x00practice\x00startPracticeSession\x00shared-key-hash").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`select id, request_fingerprint, status, resource_id::text`).
		WithArgs(in.UserID, in.IdempotencyKeyHash).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectExec(`insert into idempotency_records`).
		WithArgs(in.IdempotencyRecordID, in.UserID, in.IdempotencyKeyHash, in.RequestFingerprint, string(idempotency.StatusPending), in.ExpiresAt, in.Now).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`with selected_plan`).
		WithArgs(in.SessionID, in.UserID, in.PlanID, in.HintsEnabled, in.Now).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "plan_id", "target_job_id", "goal", "mode", "interviewer_persona",
			"language", "hints_enabled", "created_at", "updated_at",
		}).AddRow(
			in.SessionID,
			in.PlanID,
			"target-b",
			string(sharedtypes.PracticeGoalBaseline),
			string(sharedtypes.PracticeModeAssisted),
			string(sharedtypes.InterviewerRoleHiringManager),
			"en",
			false,
			in.Now,
			in.Now,
		))
	mock.ExpectCommit()

	reservation, err := repo.ReserveSessionStart(context.Background(), in)
	if err != nil {
		t.Fatalf("ReserveSessionStart returned error: %v", err)
	}
	if reservation.SessionID != in.SessionID || reservation.IdempotencyRecordID != in.IdempotencyRecordID {
		t.Fatalf("unexpected reservation: %+v", reservation)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestSQLRepositoryReserveSessionStartRejectsFingerprintMismatch(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()
	repo := NewSQLRepository(db)
	now := time.Date(2026, 5, 9, 12, 20, 0, 0, time.UTC)
	in := domain.StartSessionReservationInput{
		IdempotencyRecordID: "idem-new",
		SessionID:           "session-1",
		UserID:              "user-1",
		PlanID:              "plan-1",
		IdempotencyKeyHash:  "key-hash",
		RequestFingerprint:  "fingerprint-new",
		ExpiresAt:           now.Add(24 * time.Hour),
		Now:                 now,
	}

	mock.ExpectBegin()
	mock.ExpectExec(`select pg_advisory_xact_lock`).
		WithArgs("user-1\x00practice\x00startPracticeSession\x00key-hash").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`select id, request_fingerprint, status, resource_id::text`).
		WithArgs(in.UserID, in.IdempotencyKeyHash).
		WillReturnRows(sqlmock.NewRows([]string{"id", "request_fingerprint", "status", "resource_id"}).
			AddRow("idem-existing", "fingerprint-original", string(idempotency.StatusFailedRetry), nil))
	mock.ExpectRollback()

	_, err := repo.ReserveSessionStart(context.Background(), in)
	if !errors.Is(err, domain.ErrSessionConflict) {
		t.Fatalf("error = %v, want ErrSessionConflict", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestSQLRepositoryReserveSessionStartRejectsConcurrentPendingRecord(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()
	repo := NewSQLRepository(db)
	now := time.Date(2026, 5, 9, 12, 30, 0, 0, time.UTC)
	in := domain.StartSessionReservationInput{
		IdempotencyRecordID: "idem-new",
		SessionID:           "session-1",
		UserID:              "user-1",
		PlanID:              "plan-1",
		IdempotencyKeyHash:  "key-hash",
		RequestFingerprint:  "fingerprint",
		ExpiresAt:           now.Add(24 * time.Hour),
		Now:                 now,
	}

	mock.ExpectBegin()
	mock.ExpectExec(`select pg_advisory_xact_lock`).
		WithArgs("user-1\x00practice\x00startPracticeSession\x00key-hash").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`select id, request_fingerprint, status, resource_id::text`).
		WithArgs(in.UserID, in.IdempotencyKeyHash).
		WillReturnRows(sqlmock.NewRows([]string{"id", "request_fingerprint", "status", "resource_id"}).
			AddRow("idem-existing", in.RequestFingerprint, string(idempotency.StatusPending), nil))
	mock.ExpectRollback()

	_, err := repo.ReserveSessionStart(context.Background(), in)
	if !errors.Is(err, domain.ErrSessionConflict) {
		t.Fatalf("error = %v, want ErrSessionConflict", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestSQLRepositoryReserveSessionStartMapsActivePlanUniqueViolationToConflict(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()
	repo := NewSQLRepository(db)
	now := time.Date(2026, 5, 9, 12, 35, 0, 0, time.UTC)
	in := domain.StartSessionReservationInput{
		IdempotencyRecordID: "idem-1",
		SessionID:           "session-2",
		UserID:              "user-1",
		PlanID:              "plan-1",
		IdempotencyKeyHash:  "different-key-hash",
		RequestFingerprint:  "fingerprint-2",
		ExpiresAt:           now.Add(24 * time.Hour),
		Now:                 now,
	}

	mock.ExpectBegin()
	mock.ExpectExec(`select pg_advisory_xact_lock`).
		WithArgs("user-1\x00practice\x00startPracticeSession\x00different-key-hash").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`select id, request_fingerprint, status, resource_id::text`).
		WithArgs(in.UserID, in.IdempotencyKeyHash).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectExec(`insert into idempotency_records`).
		WithArgs(in.IdempotencyRecordID, in.UserID, in.IdempotencyKeyHash, in.RequestFingerprint, string(idempotency.StatusPending), in.ExpiresAt, in.Now).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`with selected_plan`).
		WithArgs(in.SessionID, in.UserID, in.PlanID, in.HintsEnabled, in.Now).
		WillReturnError(&pq.Error{Code: "23505", Constraint: "idx_practice_sessions_one_active_per_plan"})
	mock.ExpectRollback()

	_, err := repo.ReserveSessionStart(context.Background(), in)
	if !errors.Is(err, domain.ErrSessionConflict) {
		t.Fatalf("error = %v, want ErrSessionConflict", err)
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
