package practice

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	domain "github.com/monshunter/easyinterview/backend/internal/practice"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

func TestSQLRepositoryCommitSessionStartRecoveryLocksSessionBeforeFinalizing(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	now := time.Date(2026, 7, 19, 8, 0, 0, 0, time.UTC)
	in := domain.CommitSessionStartRecoveryInput{
		IdempotencyRecordID: "idem-recovery", SessionID: "session-1", UserID: "user-1", RecoveredAt: now,
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`select id from practice_sessions where id=$1 and user_id=$2 for update`)).
		WithArgs(in.SessionID, in.UserID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(in.SessionID))
	mock.ExpectQuery(`select id, plan_id, target_job_id, status, language, created_at, updated_at`).
		WithArgs(in.UserID, in.SessionID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "plan_id", "target_job_id", "status", "language", "created_at", "updated_at"}).
			AddRow(in.SessionID, "plan-1", "target-1", string(sharedtypes.SessionStatusRunning), "zh-CN", now, now))
	mock.ExpectQuery(`select id, role, content, seq_no, client_message_id::text, reply_status, created_at`).
		WithArgs(in.SessionID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "role", "content", "seq_no", "client_message_id", "reply_status", "created_at"}).
			AddRow("message-1", "assistant", "原开场消息", 1, nil, nil, now))
	mock.ExpectExec(`update idempotency_records set status = \$1`).
		WithArgs("succeeded", in.SessionID, sqlmock.AnyArg(), in.RecoveredAt, in.IdempotencyRecordID, in.UserID, "pending").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	recovered, err := NewSQLRepository(db).CommitSessionStartRecovery(context.Background(), in)
	if err != nil {
		t.Fatalf("CommitSessionStartRecovery: %v", err)
	}
	if recovered.ID != in.SessionID || recovered.Status != sharedtypes.SessionStatusRunning || len(recovered.Messages) != 1 {
		t.Fatalf("recovered session=%+v", recovered)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLRepositoryCommitSessionStartRejectsLateWorkerAfterQueuedSessionExpired(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	now := time.Date(2026, 7, 19, 8, 0, 0, 0, time.UTC)
	in := domain.CommitSessionStartInput{
		IdempotencyRecordID: "idem-original", SessionID: "session-1", UserID: "user-1", PlanID: "plan-1", TargetJobID: "target-1",
		Goal: sharedtypes.PracticeGoalBaseline, Language: "zh-CN", MessageID: "message-1", SessionEventID: "event-1",
		OutboxEventID: "outbox-1", AuditEventID: "audit-1", MessageText: "开场消息", StartedAt: now,
	}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("insert into practice_messages")).
		WithArgs(in.MessageID, in.SessionID, in.MessageText, in.StartedAt).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta("insert into practice_session_events")).
		WithArgs(in.SessionEventID, in.SessionID, sqlmock.AnyArg(), in.StartedAt).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta("insert into outbox_events")).
		WithArgs(in.OutboxEventID, "practice.session.started", in.SessionID, sqlmock.AnyArg(), in.StartedAt).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta("insert into audit_events")).
		WithArgs(in.AuditEventID, in.UserID, in.UserID, in.SessionID, sqlmock.AnyArg(), in.StartedAt).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`update practice_sessions set status = \$1, started_at = \$2, updated_at = \$2 where id = \$3 and user_id = \$4 and status = \$5`).
		WithArgs(string(sharedtypes.SessionStatusRunning), in.StartedAt, in.SessionID, in.UserID, string(sharedtypes.SessionStatusQueued)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "plan_id", "target_job_id", "status", "language", "created_at", "updated_at"}))
	mock.ExpectRollback()

	_, err = NewSQLRepository(db).CommitSessionStart(context.Background(), in)
	if !errors.Is(err, domain.ErrSessionConflict) {
		t.Fatalf("CommitSessionStart error=%v want ErrSessionConflict", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLRepositoryFailSessionStartOnlyExpiresQueuedSession(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	now := time.Date(2026, 7, 19, 8, 0, 0, 0, time.UTC)
	in := domain.FailSessionStartInput{
		IdempotencyRecordID: "idem-recovery", SessionID: "session-1", UserID: "user-1",
		ErrorCode: "AI_PROVIDER_TIMEOUT", Retryable: true, FailedAt: now,
	}

	mock.ExpectBegin()
	mock.ExpectExec(`update practice_sessions set status=\$1, failure_code=\$2, updated_at=\$3 where id=\$4 and user_id=\$5 and status=\$6`).
		WithArgs(string(sharedtypes.SessionStatusFailed), in.ErrorCode, in.FailedAt, in.SessionID, in.UserID, string(sharedtypes.SessionStatusQueued)).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectRollback()

	err = NewSQLRepository(db).FailSessionStart(context.Background(), in)
	if !errors.Is(err, domain.ErrSessionConflict) {
		t.Fatalf("FailSessionStart error=%v want ErrSessionConflict", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
