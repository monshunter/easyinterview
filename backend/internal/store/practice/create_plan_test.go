package practice

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lib/pq"
	"github.com/monshunter/easyinterview/backend/internal/middleware/idempotency"
	domain "github.com/monshunter/easyinterview/backend/internal/practice"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
	sharedevents "github.com/monshunter/easyinterview/backend/internal/shared/events"
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
			auditMetadataArg{
				t: t,
				expected: map[string]string{
					"plan_id":       in.PlanID,
					"goal":          string(in.Goal),
					"mode":          string(in.Mode),
					"language":      in.Language,
					"target_job_id": in.TargetJobID,
				},
				forbidden: []string{"question_text", "answer_text", "hint_text", "prompt body", "response body"},
			},
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

func TestSQLRepositoryCommitSessionStartWritesAuditMetadataWithoutQuestionText(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()
	repo := NewSQLRepository(db)
	now := time.Date(2026, 5, 9, 12, 45, 0, 0, time.UTC)
	in := validCommitSessionStartInput(now)

	mock.ExpectBegin()
	mock.ExpectExec(`insert into practice_turns`).
		WithArgs(in.TurnID, in.SessionID, in.QuestionText, in.QuestionIntent, string(in.InterviewerPersona), in.StartedAt).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`insert into practice_session_events`).
		WithArgs(in.SessionEventID, in.SessionID, sqlmock.AnyArg(), in.StartedAt).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`insert into outbox_events`).
		WithArgs(in.OutboxEventID, string(sharedevents.EventNamePracticeSessionStarted), in.SessionID, sqlmock.AnyArg(), in.StartedAt).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`insert into audit_events`).
		WithArgs(
			in.AuditEventID,
			in.UserID,
			in.UserID,
			in.SessionID,
			auditMetadataArg{
				t: t,
				expected: map[string]string{
					"plan_id":       in.PlanID,
					"session_id":    in.SessionID,
					"goal":          string(in.Goal),
					"mode":          string(in.Mode),
					"language":      in.Language,
					"target_job_id": in.TargetJobID,
				},
				forbidden: []string{in.QuestionText, in.QuestionIntent, "question_text", "answer_text", "hint_text", "prompt body", "response body"},
			},
			in.StartedAt,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`update practice_sessions`).
		WithArgs(string(sharedtypes.SessionStatusRunning), in.StartedAt, in.SessionID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "plan_id", "target_job_id", "status", "language", "hints_enabled",
			"turn_count", "created_at", "updated_at",
		}).AddRow(
			in.SessionID,
			in.PlanID,
			in.TargetJobID,
			string(sharedtypes.SessionStatusRunning),
			in.Language,
			true,
			1,
			in.CreatedAt,
			in.StartedAt,
		))
	mock.ExpectExec(`update idempotency_records`).
		WithArgs(string(idempotency.StatusSucceeded), in.SessionID, sqlmock.AnyArg(), in.StartedAt, in.IdempotencyRecordID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	session, err := repo.CommitSessionStart(context.Background(), in)
	if err != nil {
		t.Fatalf("CommitSessionStart returned error: %v", err)
	}
	if session.ID != in.SessionID || session.CurrentTurn == nil || session.CurrentTurn.QuestionText != in.QuestionText {
		t.Fatalf("unexpected committed session: %+v", session)
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
	mock.ExpectQuery(`select id, request_fingerprint, status, resource_id::text, response_body, expires_at`).
		WithArgs(in.UserID, in.IdempotencyKeyHash).
		WillReturnRows(sqlmock.NewRows([]string{"id", "request_fingerprint", "status", "resource_id", "response_body", "expires_at"}).
			AddRow("idem-existing", in.RequestFingerprint, string(idempotency.StatusFailedRetry), nil, nil, in.ExpiresAt))
	mock.ExpectExec(`update idempotency_records`).
		WithArgs(in.RequestFingerprint, string(idempotency.StatusPending), in.ExpiresAt, in.Now, "idem-existing", in.UserID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`with selected_plan`).
		WithArgs(in.SessionID, in.UserID, in.PlanID, in.HintsEnabled, in.Now).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "plan_id", "target_job_id", "goal", "mode", "interviewer_persona",
			"language", "role_title", "seniority", "top_skills", "hints_enabled", "created_at", "updated_at",
		}).AddRow(
			in.SessionID,
			in.PlanID,
			"target-1",
			string(sharedtypes.PracticeGoalBaseline),
			string(sharedtypes.PracticeModeAssisted),
			string(sharedtypes.InterviewerRoleHiringManager),
			"zh-CN",
			"Staff Frontend Architect",
			"staff",
			"React, design systems",
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

func TestSQLRepositoryReserveSessionStartReplaysStoredResponseBody(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()
	repo := NewSQLRepository(db)
	now := time.Date(2026, 5, 9, 12, 22, 0, 0, time.UTC)
	in := domain.StartSessionReservationInput{
		IdempotencyRecordID: "new-idem-ignored",
		SessionID:           "new-session-ignored",
		UserID:              "user-1",
		PlanID:              "plan-ignored",
		HintsEnabled:        true,
		IdempotencyKeyHash:  "key-hash",
		RequestFingerprint:  "fingerprint",
		ExpiresAt:           now.Add(24 * time.Hour),
		Now:                 now,
	}
	snapshot := domain.SessionRecord{
		ID:           "session-original",
		PlanID:       "plan-1",
		TargetJobID:  "target-1",
		Status:       sharedtypes.SessionStatusRunning,
		Language:     "zh-CN",
		HintsEnabled: true,
		TurnCount:    1,
		CreatedAt:    now.Add(-time.Minute),
		UpdatedAt:    now,
		CurrentTurn: &domain.TurnRecord{
			ID:             "turn-1",
			TurnIndex:      1,
			QuestionText:   "original first question",
			QuestionIntent: "behavioral.leadership",
			Status:         "asked",
			AskedAt:        now,
		},
	}
	responseBody, err := marshalSessionResponseBody(snapshot)
	if err != nil {
		t.Fatalf("marshal response snapshot: %v", err)
	}

	mock.ExpectBegin()
	mock.ExpectExec(`select pg_advisory_xact_lock`).
		WithArgs("user-1\x00practice\x00startPracticeSession\x00key-hash").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`select id, request_fingerprint, status, resource_id::text, response_body, expires_at`).
		WithArgs(in.UserID, in.IdempotencyKeyHash).
		WillReturnRows(sqlmock.NewRows([]string{"id", "request_fingerprint", "status", "resource_id", "response_body", "expires_at"}).
			AddRow("idem-existing", in.RequestFingerprint, string(idempotency.StatusSucceeded), snapshot.ID, string(responseBody), in.ExpiresAt))
	mock.ExpectCommit()

	reservation, err := repo.ReserveSessionStart(context.Background(), in)
	if err != nil {
		t.Fatalf("ReserveSessionStart returned error: %v", err)
	}
	if reservation.ReplaySession == nil {
		t.Fatalf("expected replay session from stored response body")
	}
	if reservation.ReplaySession.ID != snapshot.ID ||
		reservation.ReplaySession.Status != sharedtypes.SessionStatusRunning ||
		reservation.ReplaySession.CurrentTurn == nil ||
		reservation.ReplaySession.CurrentTurn.QuestionText != "original first question" {
		t.Fatalf("unexpected replay session: %+v", reservation.ReplaySession)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestSQLRepositoryReserveSessionStartResetsExpiredPendingRecord(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()
	repo := NewSQLRepository(db)
	now := time.Date(2026, 5, 10, 12, 0, 0, 0, time.UTC)
	in := domain.StartSessionReservationInput{
		IdempotencyRecordID: "new-idem-ignored",
		SessionID:           "session-after-expiry",
		UserID:              "user-1",
		PlanID:              "plan-1",
		HintsEnabled:        true,
		IdempotencyKeyHash:  "key-hash",
		RequestFingerprint:  "new-fingerprint",
		ExpiresAt:           now.Add(24 * time.Hour),
		Now:                 now,
	}

	mock.ExpectBegin()
	mock.ExpectExec(`select pg_advisory_xact_lock`).
		WithArgs("user-1\x00practice\x00startPracticeSession\x00key-hash").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`select id, request_fingerprint, status, resource_id::text, response_body, expires_at`).
		WithArgs(in.UserID, in.IdempotencyKeyHash).
		WillReturnRows(sqlmock.NewRows([]string{"id", "request_fingerprint", "status", "resource_id", "response_body", "expires_at"}).
			AddRow("idem-expired", "old-fingerprint", string(idempotency.StatusPending), nil, nil, now.Add(-time.Second)))
	mock.ExpectExec(`update idempotency_records`).
		WithArgs(in.RequestFingerprint, string(idempotency.StatusPending), in.ExpiresAt, in.Now, "idem-expired", in.UserID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`with selected_plan`).
		WithArgs(in.SessionID, in.UserID, in.PlanID, in.HintsEnabled, in.Now).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "plan_id", "target_job_id", "goal", "mode", "interviewer_persona",
			"language", "role_title", "seniority", "top_skills", "hints_enabled", "created_at", "updated_at",
		}).AddRow(
			in.SessionID,
			in.PlanID,
			"target-1",
			string(sharedtypes.PracticeGoalBaseline),
			string(sharedtypes.PracticeModeAssisted),
			string(sharedtypes.InterviewerRoleHiringManager),
			"zh-CN",
			"Staff Frontend Architect",
			"staff",
			"React, design systems",
			true,
			in.Now,
			in.Now,
		))
	mock.ExpectCommit()

	reservation, err := repo.ReserveSessionStart(context.Background(), in)
	if err != nil {
		t.Fatalf("ReserveSessionStart returned error: %v", err)
	}
	if reservation.IdempotencyRecordID != "idem-expired" || reservation.ReplaySession != nil {
		t.Fatalf("expired pending record should reset into fresh execution: %+v", reservation)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestSQLRepositoryReserveSessionStartResetsExpiredSucceededRecord(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()
	repo := NewSQLRepository(db)
	now := time.Date(2026, 5, 10, 12, 5, 0, 0, time.UTC)
	in := domain.StartSessionReservationInput{
		IdempotencyRecordID: "new-idem-ignored",
		SessionID:           "session-after-success-expiry",
		UserID:              "user-1",
		PlanID:              "plan-1",
		HintsEnabled:        true,
		IdempotencyKeyHash:  "key-hash",
		RequestFingerprint:  "new-fingerprint",
		ExpiresAt:           now.Add(24 * time.Hour),
		Now:                 now,
	}
	snapshot := domain.SessionRecord{
		ID:           "old-session",
		PlanID:       "old-plan",
		TargetJobID:  "old-target",
		Status:       sharedtypes.SessionStatusRunning,
		Language:     "zh-CN",
		HintsEnabled: true,
		TurnCount:    1,
		CreatedAt:    now.Add(-25 * time.Hour),
		UpdatedAt:    now.Add(-25 * time.Hour),
	}
	responseBody, err := marshalSessionResponseBody(snapshot)
	if err != nil {
		t.Fatalf("marshal response snapshot: %v", err)
	}

	mock.ExpectBegin()
	mock.ExpectExec(`select pg_advisory_xact_lock`).
		WithArgs("user-1\x00practice\x00startPracticeSession\x00key-hash").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`select id, request_fingerprint, status, resource_id::text, response_body, expires_at`).
		WithArgs(in.UserID, in.IdempotencyKeyHash).
		WillReturnRows(sqlmock.NewRows([]string{"id", "request_fingerprint", "status", "resource_id", "response_body", "expires_at"}).
			AddRow("idem-expired-success", "old-fingerprint", string(idempotency.StatusSucceeded), snapshot.ID, string(responseBody), now.Add(-time.Second)))
	mock.ExpectExec(`update idempotency_records`).
		WithArgs(in.RequestFingerprint, string(idempotency.StatusPending), in.ExpiresAt, in.Now, "idem-expired-success", in.UserID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`with selected_plan`).
		WithArgs(in.SessionID, in.UserID, in.PlanID, in.HintsEnabled, in.Now).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "plan_id", "target_job_id", "goal", "mode", "interviewer_persona",
			"language", "role_title", "seniority", "top_skills", "hints_enabled", "created_at", "updated_at",
		}).AddRow(
			in.SessionID,
			in.PlanID,
			"target-1",
			string(sharedtypes.PracticeGoalBaseline),
			string(sharedtypes.PracticeModeAssisted),
			string(sharedtypes.InterviewerRoleHiringManager),
			"zh-CN",
			"Staff Frontend Architect",
			"staff",
			"React, design systems",
			true,
			in.Now,
			in.Now,
		))
	mock.ExpectCommit()

	reservation, err := repo.ReserveSessionStart(context.Background(), in)
	if err != nil {
		t.Fatalf("ReserveSessionStart returned error: %v", err)
	}
	if reservation.IdempotencyRecordID != "idem-expired-success" || reservation.ReplaySession != nil || reservation.SessionID != in.SessionID {
		t.Fatalf("expired succeeded record should not replay old response: %+v", reservation)
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
	mock.ExpectQuery(`select id, request_fingerprint, status, resource_id::text, response_body, expires_at`).
		WithArgs(in.UserID, in.IdempotencyKeyHash).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectExec(`insert into idempotency_records`).
		WithArgs(in.IdempotencyRecordID, in.UserID, in.IdempotencyKeyHash, in.RequestFingerprint, string(idempotency.StatusPending), in.ExpiresAt, in.Now).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`with selected_plan`).
		WithArgs(in.SessionID, in.UserID, in.PlanID, in.HintsEnabled, in.Now).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "plan_id", "target_job_id", "goal", "mode", "interviewer_persona",
			"language", "role_title", "seniority", "top_skills", "hints_enabled", "created_at", "updated_at",
		}).AddRow(
			in.SessionID,
			in.PlanID,
			"target-b",
			string(sharedtypes.PracticeGoalBaseline),
			string(sharedtypes.PracticeModeAssisted),
			string(sharedtypes.InterviewerRoleHiringManager),
			"en",
			"Product Manager",
			"mid",
			"prioritization, stakeholder communication",
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
	mock.ExpectQuery(`select id, request_fingerprint, status, resource_id::text, response_body, expires_at`).
		WithArgs(in.UserID, in.IdempotencyKeyHash).
		WillReturnRows(sqlmock.NewRows([]string{"id", "request_fingerprint", "status", "resource_id", "response_body", "expires_at"}).
			AddRow("idem-existing", "fingerprint-original", string(idempotency.StatusFailedRetry), nil, nil, in.ExpiresAt))
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
	mock.ExpectQuery(`select id, request_fingerprint, status, resource_id::text, response_body, expires_at`).
		WithArgs(in.UserID, in.IdempotencyKeyHash).
		WillReturnRows(sqlmock.NewRows([]string{"id", "request_fingerprint", "status", "resource_id", "response_body", "expires_at"}).
			AddRow("idem-existing", in.RequestFingerprint, string(idempotency.StatusPending), nil, nil, in.ExpiresAt))
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
	mock.ExpectQuery(`select id, request_fingerprint, status, resource_id::text, response_body, expires_at`).
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

func validCommitSessionStartInput(now time.Time) domain.CommitSessionStartInput {
	return domain.CommitSessionStartInput{
		IdempotencyRecordID: "01918fa0-0000-7000-8000-000000005000",
		SessionID:           "01918fa0-0000-7000-8000-000000006000",
		UserID:              "01918fa0-0000-7000-8000-000000000001",
		PlanID:              "01918fa0-0000-7000-8000-000000004000",
		TargetJobID:         "01918fa0-0000-7000-8000-000000002000",
		Goal:                sharedtypes.PracticeGoalBaseline,
		Mode:                sharedtypes.PracticeModeAssisted,
		InterviewerPersona:  sharedtypes.InterviewerRoleHiringManager,
		Language:            "zh-CN",
		HintsEnabled:        true,
		TurnID:              "01918fa0-0000-7000-8000-000000007000",
		SessionEventID:      "01918fa0-0000-7000-8000-000000008000",
		OutboxEventID:       "01918fa0-0000-7000-8000-000000009000",
		AuditEventID:        "01918fa0-0000-7000-8000-000000010000",
		QuestionText:        "请描述一次跨团队设计系统迁移。",
		QuestionIntent:      "behavioral.leadership.design_system",
		StartedAt:           now,
		CreatedAt:           now.Add(-time.Minute),
	}
}

type auditMetadataArg struct {
	t         *testing.T
	expected  map[string]string
	forbidden []string
}

func (a auditMetadataArg) Match(value driver.Value) bool {
	var raw []byte
	switch v := value.(type) {
	case []byte:
		raw = append([]byte{}, v...)
	case string:
		raw = []byte(v)
	default:
		a.t.Errorf("audit metadata has unexpected type %T", value)
		return false
	}
	for _, forbidden := range a.forbidden {
		if forbidden != "" && containsBytes(raw, []byte(forbidden)) {
			a.t.Errorf("audit metadata leaked forbidden evidence %q: %s", forbidden, string(raw))
			return false
		}
	}
	var got map[string]string
	if err := json.Unmarshal(raw, &got); err != nil {
		a.t.Errorf("audit metadata is not string map JSON: %v; raw=%s", err, string(raw))
		return false
	}
	if len(got) != len(a.expected) {
		a.t.Errorf("audit metadata keys drifted: got=%v want=%v", got, a.expected)
		return false
	}
	for key, want := range a.expected {
		if got[key] != want {
			a.t.Errorf("audit metadata[%s]=%q want %q; full=%v", key, got[key], want, got)
			return false
		}
	}
	return true
}

func containsBytes(haystack, needle []byte) bool {
	if len(needle) == 0 || len(needle) > len(haystack) {
		return false
	}
	for i := 0; i <= len(haystack)-len(needle); i++ {
		match := true
		for j := range needle {
			if haystack[i+j] != needle[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}
