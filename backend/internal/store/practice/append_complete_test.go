package practice

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	domain "github.com/monshunter/easyinterview/backend/internal/practice"
	sharedevents "github.com/monshunter/easyinterview/backend/internal/shared/events"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

func TestSQLRepositoryAppendSessionEventWritesEventTurnSessionOutboxWithoutAudit(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()
	repo := NewSQLRepository(db)
	now := time.Date(2026, 4, 28, 13, 45, 12, 0, time.UTC)
	in := domain.AppendSessionEventStoreInput{
		EventID:            "event-1",
		OutboxEventID:      "outbox-1",
		UserID:             "user-1",
		SessionID:          "session-1",
		ClientEventID:      "client-event-1",
		Kind:               "answer_submitted",
		OccurredAt:         now,
		RequestFingerprint: "fingerprint-1",
		RequestPayload: map[string]any{
			"answerText":    "answer",
			"followUpCount": 99,
		},
		Outcome: domain.SessionEventOutcome{
			Acknowledged:      true,
			NextSessionStatus: sharedtypes.SessionStatusCompleted,
			NextTurn: &domain.TurnRecord{
				ID:             "turn-1",
				TurnIndex:      1,
				QuestionText:   "Question?",
				QuestionIntent: "behavioral",
				Status:         string(domain.TurnStatusAssessed),
				FollowUpCount:  1,
				AskedAt:        now.Add(-time.Minute),
			},
			AssistantAction: domain.AssistantActionRecord{
				Type:          "session_completed",
				SessionStatus: sharedtypes.SessionStatusCompleted,
				Provenance:    domain.AssistantActionProvenance{PromptVersion: "not_applicable", RubricVersion: "not_applicable", ModelID: "model-profile:static", Language: "zh-CN", FeatureFlag: "none", DataSourceVersion: "static"},
			},
			OutboxRecord: &domain.PracticeTurnCompletedRecord{
				SessionID:        "session-1",
				TurnID:           "turn-1",
				FollowUpCount:    1,
				AnswerCharLength: 6,
				CompletedAt:      now,
			},
		},
	}

	mock.ExpectBegin()
	expectAppendContext(mock, now)
	mock.ExpectQuery(`select payload`).
		WithArgs(in.SessionID, in.ClientEventID).
		WillReturnRows(sqlmock.NewRows([]string{"payload"}).AddRow([]byte(`{"requestFingerprint":"fingerprint-1","pending":true}`)))
	mock.ExpectExec(`update practice_turns`).
		WithArgs(string(domain.TurnStatusAssessed), "answer", 1, now, now, now, in.SessionID, "turn-1").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`update practice_sessions`).
		WithArgs(string(sharedtypes.SessionStatusCompleted), int32(1), now, in.SessionID, in.UserID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`insert into outbox_events`).
		WithArgs(in.OutboxEventID, string(sharedevents.EventNamePracticeTurnCompleted), "turn-1", sqlmock.AnyArg(), now).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`update practice_session_events`).
		WithArgs(sqlmock.AnyArg(), in.SessionID, in.ClientEventID, in.EventID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	result, err := repo.AppendSessionEvent(context.Background(), in)
	if err != nil {
		t.Fatalf("AppendSessionEvent returned error: %v", err)
	}
	if !result.Acknowledged || result.Session.Status != sharedtypes.SessionStatusCompleted {
		t.Fatalf("unexpected result: %+v", result)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestSQLRepositoryReserveSessionEventCreatesPendingReservation(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()
	repo := NewSQLRepository(db)
	now := time.Date(2026, 4, 28, 13, 45, 12, 0, time.UTC)
	in := domain.SessionEventReservationInput{
		EventID:            "event-1",
		UserID:             "user-1",
		SessionID:          "session-1",
		ClientEventID:      "client-event-1",
		Kind:               "answer_submitted",
		RequestFingerprint: "fingerprint-1",
		Now:                now,
	}

	mock.ExpectBegin()
	expectAppendContext(mock, now)
	mock.ExpectQuery(`select payload`).
		WithArgs(in.SessionID, in.ClientEventID).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(`select coalesce\(max\(seq_no\), 0\) \+ 1`).
		WithArgs(in.SessionID).
		WillReturnRows(sqlmock.NewRows([]string{"seq_no"}).AddRow(2))
	mock.ExpectExec(`insert into practice_session_events`).
		WithArgs(in.EventID, in.SessionID, 2, in.Kind, in.ClientEventID, sqlmock.AnyArg(), now).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	result, err := repo.ReserveSessionEvent(context.Background(), in)
	if err != nil {
		t.Fatalf("ReserveSessionEvent returned error: %v", err)
	}
	if result.ReplayResult != nil || result.Session.ID != in.SessionID || result.LatestTurn.ID != "turn-1" {
		t.Fatalf("unexpected reservation: %+v", result)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestSQLRepositoryReserveSessionEventRejectsPendingReservationReplay(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()
	repo := NewSQLRepository(db)
	now := time.Date(2026, 4, 28, 13, 45, 12, 0, time.UTC)
	in := domain.SessionEventReservationInput{
		EventID:            "event-2",
		UserID:             "user-1",
		SessionID:          "session-1",
		ClientEventID:      "client-event-1",
		Kind:               "answer_submitted",
		RequestFingerprint: "fingerprint-1",
		Now:                now,
	}

	mock.ExpectBegin()
	expectAppendContext(mock, now)
	mock.ExpectQuery(`select payload`).
		WithArgs(in.SessionID, in.ClientEventID).
		WillReturnRows(sqlmock.NewRows([]string{"payload"}).AddRow([]byte(`{"requestFingerprint":"fingerprint-1","requestPayload":{"answerText":"answer"},"pending":true}`)))
	mock.ExpectRollback()

	_, err := repo.ReserveSessionEvent(context.Background(), in)
	if !errors.Is(err, domain.ErrSessionConflict) {
		t.Fatalf("error = %v, want ErrSessionConflict", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestSQLRepositoryCompleteSessionWritesReportJobOutboxAndAudit(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()
	repo := NewSQLRepository(db)
	now := time.Date(2026, 4, 28, 14, 0, 0, 0, time.UTC)
	in := domain.CompleteSessionStoreInput{
		UserID:            "user-1",
		SessionID:         "session-1",
		ReportID:          "report-1",
		JobID:             "job-1",
		SessionEventID:    "event-1",
		OutboxEventID:     "outbox-1",
		AuditEventID:      "audit-1",
		ClientCompletedAt: now,
		Now:               now,
	}

	mock.ExpectBegin()
	mock.ExpectQuery(`select id, plan_id, target_job_id, status, language, hints_enabled`).
		WithArgs(in.UserID, in.SessionID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "plan_id", "target_job_id", "status", "language", "hints_enabled", "turn_count", "created_at", "updated_at"}).
			AddRow(in.SessionID, "plan-1", "target-1", string(sharedtypes.SessionStatusRunning), "zh-CN", true, 3, now.Add(-time.Hour), now.Add(-time.Minute)))
	mock.ExpectQuery(`select fr.id`).
		WithArgs(in.UserID, in.SessionID, "report_generate", in.SessionID).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectExec(`update practice_sessions`).
		WithArgs(string(sharedtypes.SessionStatusCompleting), now, in.SessionID, in.UserID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`select coalesce\(max\(seq_no\), 0\) \+ 1`).
		WithArgs(in.SessionID).
		WillReturnRows(sqlmock.NewRows([]string{"seq_no"}).AddRow(4))
	mock.ExpectExec(`insert into practice_session_events`).
		WithArgs(in.SessionEventID, in.SessionID, 4, sqlmock.AnyArg(), now).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`insert into feedback_reports`).
		WithArgs(in.ReportID, in.UserID, in.SessionID, "target-1", string(sharedtypes.ReportStatusQueued), now).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`insert into async_jobs`).
		WithArgs(in.JobID, "report_generate", in.ReportID, in.SessionID, string(sharedtypes.JobStatusQueued), sqlmock.AnyArg(), now).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`insert into outbox_events`).
		WithArgs(in.OutboxEventID, string(sharedevents.EventNamePracticeSessionCompleted), in.SessionID, sqlmock.AnyArg(), now).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`insert into audit_events`).
		WithArgs(in.AuditEventID, in.UserID, in.UserID, in.SessionID, sqlmock.AnyArg(), now).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	result, err := repo.CompleteSession(context.Background(), in)
	if err != nil {
		t.Fatalf("CompleteSession returned error: %v", err)
	}
	if result.ReportID != in.ReportID || result.Job.ID != in.JobID || result.Job.Status != sharedtypes.JobStatusQueued {
		t.Fatalf("unexpected result: %+v", result)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestSQLRepositoryCompleteSessionRejectsIllegalStatusWithoutReport(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()
	repo := NewSQLRepository(db)
	now := time.Date(2026, 4, 28, 14, 0, 0, 0, time.UTC)
	in := domain.CompleteSessionStoreInput{
		UserID:            "user-1",
		SessionID:         "session-1",
		ReportID:          "report-1",
		JobID:             "job-1",
		SessionEventID:    "event-1",
		OutboxEventID:     "outbox-1",
		AuditEventID:      "audit-1",
		ClientCompletedAt: now,
		Now:               now,
	}

	mock.ExpectBegin()
	mock.ExpectQuery(`select id, plan_id, target_job_id, status, language, hints_enabled`).
		WithArgs(in.UserID, in.SessionID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "plan_id", "target_job_id", "status", "language", "hints_enabled", "turn_count", "created_at", "updated_at"}).
			AddRow(in.SessionID, "plan-1", "target-1", string(sharedtypes.SessionStatusFailed), "zh-CN", true, 3, now.Add(-time.Hour), now.Add(-time.Minute)))
	mock.ExpectQuery(`select fr.id`).
		WithArgs(in.UserID, in.SessionID, "report_generate", in.SessionID).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectRollback()

	_, err := repo.CompleteSession(context.Background(), in)
	if !errors.Is(err, domain.ErrSessionConflict) {
		t.Fatalf("error = %v, want ErrSessionConflict", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestSQLRepositoryCompleteSessionReplaysExistingReportBeforeStatusGuard(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()
	repo := NewSQLRepository(db)
	now := time.Date(2026, 4, 28, 14, 0, 0, 0, time.UTC)
	in := domain.CompleteSessionStoreInput{
		UserID:            "user-1",
		SessionID:         "session-1",
		ReportID:          "report-new",
		JobID:             "job-new",
		SessionEventID:    "event-new",
		OutboxEventID:     "outbox-new",
		AuditEventID:      "audit-new",
		ClientCompletedAt: now,
		Now:               now,
	}

	mock.ExpectBegin()
	mock.ExpectQuery(`select id, plan_id, target_job_id, status, language, hints_enabled`).
		WithArgs(in.UserID, in.SessionID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "plan_id", "target_job_id", "status", "language", "hints_enabled", "turn_count", "created_at", "updated_at"}).
			AddRow(in.SessionID, "plan-1", "target-1", string(sharedtypes.SessionStatusFailed), "zh-CN", true, 3, now.Add(-time.Hour), now.Add(-time.Minute)))
	mock.ExpectQuery(`j\.dedupe_key = \$4`).
		WithArgs(in.UserID, in.SessionID, "report_generate", in.SessionID).
		WillReturnRows(sqlmock.NewRows([]string{"report_id", "job_id", "job_type", "resource_type", "resource_id", "status", "error_code", "created_at", "updated_at"}).
			AddRow("report-existing", "job-existing", "report_generate", "feedback_report", "report-existing", string(sharedtypes.JobStatusQueued), nil, now.Add(-time.Minute), now.Add(-time.Minute)))
	mock.ExpectCommit()

	result, err := repo.CompleteSession(context.Background(), in)
	if err != nil {
		t.Fatalf("CompleteSession returned error: %v", err)
	}
	if !result.Replay || result.ReportID != "report-existing" || result.Job.ID != "job-existing" {
		t.Fatalf("unexpected replay result: %+v", result)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestCanCompletePracticeSessionStatusAllowsRunningWaitingAndCompleted(t *testing.T) {
	allowed := map[sharedtypes.SessionStatus]bool{
		sharedtypes.SessionStatusRunning:          true,
		sharedtypes.SessionStatusWaitingUserInput: true,
		sharedtypes.SessionStatusCompleted:        true,
	}
	for _, status := range sharedtypes.AllSessionStatuses {
		if got := canCompletePracticeSessionStatus(status); got != allowed[status] {
			t.Fatalf("canCompletePracticeSessionStatus(%q) = %v, want %v", status, got, allowed[status])
		}
	}
}

func expectAppendContext(mock sqlmock.Sqlmock, now time.Time) {
	mock.ExpectQuery(`select s.id, s.plan_id`).
		WithArgs("user-1", "session-1").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "plan_id", "target_job_id", "status", "language", "hints_enabled",
			"turn_count", "created_at", "updated_at",
			"id", "target_job_id", "goal", "mode", "interviewer_persona", "difficulty",
			"language", "time_budget_minutes", "question_budget", "status", "created_at",
			"id", "turn_index", "question_text", "question_intent", "status", "follow_up_count", "asked_at",
		}).AddRow(
			"session-1", "plan-1", "target-1", string(sharedtypes.SessionStatusRunning), "zh-CN", true,
			1, now.Add(-time.Hour), now.Add(-time.Minute),
			"plan-1", "target-1", string(sharedtypes.PracticeGoalBaseline), string(sharedtypes.PracticeModeAssisted), string(sharedtypes.InterviewerRoleHiringManager), "standard",
			"zh-CN", 30, 3, "ready", now.Add(-2*time.Hour),
			"turn-1", 1, "Question?", "behavioral", string(domain.TurnStatusAsked), 1, now.Add(-time.Minute),
		))
}
