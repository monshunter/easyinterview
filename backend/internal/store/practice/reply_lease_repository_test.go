package practice

import (
	"context"
	"database/sql"
	"reflect"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	domain "github.com/monshunter/easyinterview/backend/internal/practice"
)

func TestPracticeReplyLeaseInternalContract(t *testing.T) {
	if domain.PracticeReplyLeaseDuration != 90*time.Second {
		t.Fatalf("lease duration=%s want 90s", domain.PracticeReplyLeaseDuration)
	}
	reservation := domain.PracticeMessageReservation{ReplyGeneration: 1}
	if reservation.ReplyGeneration != 1 {
		t.Fatalf("reservation generation=%d want 1", reservation.ReplyGeneration)
	}
	messageType := reflect.TypeOf(domain.MessageRecord{})
	for _, forbidden := range []string{"ReplyGeneration", "ReplyLeaseExpiresAt", "Generation", "LeaseExpiresAt"} {
		if _, ok := messageType.FieldByName(forbidden); ok {
			t.Fatalf("public MessageRecord exposes internal field %s", forbidden)
		}
	}
}

func TestSQLRepositoryReservePracticeMessageStartsGenerationOneWithExactLease(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	now := time.Date(2026, 7, 14, 8, 0, 0, 123000000, time.UTC)
	expectReplyStateReservationContext(mock, now)
	mock.ExpectQuery(`(?s)select u\.id, u\.role, u\.content, u\.seq_no, u\.client_message_id::text, u\.reply_status, u\.reply_generation, u\.reply_lease_expires_at, u\.created_at.*for update of u`).
		WithArgs("session-1", "client-1").
		WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(`select exists`).WithArgs("session-1").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	mock.ExpectQuery(`select coalesce\(sum\(octet_length\(content\)\),0\) from practice_messages`).WithArgs("session-1").
		WillReturnRows(sqlmock.NewRows([]string{"bytes"}).AddRow(4))
	mock.ExpectQuery(`select coalesce\(max\(seq_no\),0\)\+1`).WithArgs("session-1").
		WillReturnRows(sqlmock.NewRows([]string{"seq"}).AddRow(2))
	mock.ExpectExec(`(?s)insert into practice_messages .*reply_generation, reply_lease_expires_at, created_at.*values`).
		WithArgs("m2", "session-1", 2, "继续", "client-1", string(domain.PracticeReplyStatusPending), int64(1), now.Add(90*time.Second), now).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`select id, role, content, seq_no, client_message_id::text, reply_status, created_at from practice_messages`).WithArgs("session-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "role", "content", "seq_no", "client_message_id", "reply_status", "created_at"}).
			AddRow("m1", "assistant", "你好", 1, nil, nil, now).
			AddRow("m2", "user", "继续", 2, "client-1", string(domain.PracticeReplyStatusPending), now))
	mock.ExpectCommit()

	reservation, err := NewSQLRepository(db).ReservePracticeMessage(context.Background(), domain.ReservePracticeMessageInput{
		UserMessageID: "m2", UserID: "user-1", SessionID: "session-1", ClientMessageID: "client-1", Text: "继续", MaxSessionTextBytes: 10, Now: now,
	})
	if err != nil {
		t.Fatalf("ReservePracticeMessage: %v", err)
	}
	if reservation.ReplyGeneration != 1 {
		t.Fatalf("generation=%d want 1", reservation.ReplyGeneration)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLRepositoryReservePracticeMessageRejectsAggregateLimitPlusOneBeforeInsert(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	now := time.Date(2026, 7, 14, 8, 0, 0, 123000000, time.UTC)
	expectReplyStateReservationContext(mock, now)
	mock.ExpectQuery(`(?s)select u\.id, u\.role, u\.content, u\.seq_no.*for update of u`).
		WithArgs("session-1", "client-1").
		WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(`select exists`).WithArgs("session-1").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	mock.ExpectQuery(`select coalesce\(sum\(octet_length\(content\)\),0\) from practice_messages`).WithArgs("session-1").
		WillReturnRows(sqlmock.NewRows([]string{"bytes"}).AddRow(4))
	mock.ExpectRollback()

	_, err = NewSQLRepository(db).ReservePracticeMessage(context.Background(), domain.ReservePracticeMessageInput{
		UserMessageID: "m2", UserID: "user-1", SessionID: "session-1", ClientMessageID: "client-1", Text: "继续", MaxSessionTextBytes: 9, Now: now,
	})
	if err != domain.ErrPracticeSessionTextLimitExceeded {
		t.Fatalf("error=%v want ErrPracticeSessionTextLimitExceeded", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLRepositoryReservePracticeMessageRetryIncrementsGenerationAndRefreshesExactLease(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	now := time.Date(2026, 7, 14, 8, 0, 0, 123000000, time.UTC)
	expectReplyStateReservationContext(mock, now)
	mock.ExpectQuery(`(?s)select u\.id, u\.role, u\.content, u\.seq_no, u\.client_message_id::text, u\.reply_status, u\.reply_generation, u\.reply_lease_expires_at, u\.created_at.*for update of u`).
		WithArgs("session-1", "client-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "role", "content", "seq_no", "client_message_id", "reply_status", "reply_generation", "reply_lease_expires_at", "created_at", "assistant_id", "assistant_content", "assistant_seq", "assistant_created"}).
			AddRow("m2", "user", "继续", 2, "client-1", string(domain.PracticeReplyStatusRetryableFailed), int64(1), nil, now, nil, nil, nil, nil))
	mock.ExpectQuery(`(?s)update practice_messages m.*set reply_status=\$1,.*reply_generation=reply_generation\+1,.*reply_lease_expires_at=\$2.*returning m\.reply_generation`).
		WithArgs(string(domain.PracticeReplyStatusPending), now.Add(90*time.Second), "m2", "session-1", "user-1", string(domain.PracticeReplyStatusRetryableFailed), int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"reply_generation"}).AddRow(int64(2)))
	mock.ExpectQuery(`select id, role, content, seq_no, client_message_id::text, reply_status, created_at from practice_messages`).WithArgs("session-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "role", "content", "seq_no", "client_message_id", "reply_status", "created_at"}).
			AddRow("m1", "assistant", "你好", 1, nil, nil, now).
			AddRow("m2", "user", "继续", 2, "client-1", string(domain.PracticeReplyStatusPending), now))
	mock.ExpectCommit()

	reservation, err := NewSQLRepository(db).ReservePracticeMessage(context.Background(), domain.ReservePracticeMessageInput{
		UserMessageID: "unused", UserID: "user-1", SessionID: "session-1", ClientMessageID: "client-1", Text: "继续", Now: now,
	})
	if err != nil {
		t.Fatalf("ReservePracticeMessage: %v", err)
	}
	if reservation.ReplyGeneration != 2 || reservation.UserMessage.ID != "m2" {
		t.Fatalf("retry reservation=%+v", reservation)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLRepositoryGetSessionLazilyExpiresPendingAtExactLeaseBoundary(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	now := time.Date(2026, 7, 14, 8, 1, 30, 0, time.UTC)
	mock.ExpectBegin()
	mock.ExpectQuery(`select id from practice_sessions where id=\$1 and user_id=\$2 for update`).
		WithArgs("session-1", "user-1").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("session-1"))
	mock.ExpectExec(`(?s)update practice_messages.*set reply_status=\$1, reply_lease_expires_at=null.*reply_lease_expires_at <= \$3.*reply_status=\$4`).
		WithArgs(string(domain.PracticeReplyStatusRetryableFailed), "session-1", now, string(domain.PracticeReplyStatusPending)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`select id, plan_id, target_job_id, status, language, created_at, updated_at`).WithArgs("user-1", "session-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "plan_id", "target_job_id", "status", "language", "created_at", "updated_at"}).
			AddRow("session-1", "plan-1", "target-1", "running", "zh-CN", now, now))
	mock.ExpectQuery(`select id, role, content, seq_no, client_message_id::text, reply_status, created_at from practice_messages`).WithArgs("session-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "role", "content", "seq_no", "client_message_id", "reply_status", "created_at"}).
			AddRow("m1", "assistant", "你好", 1, nil, nil, now).
			AddRow("m2", "user", "继续", 2, "client-1", string(domain.PracticeReplyStatusRetryableFailed), now))
	mock.ExpectCommit()

	session, err := NewSQLRepository(db).GetSession(context.Background(), "user-1", "session-1", now)
	if err != nil {
		t.Fatalf("GetSession: %v", err)
	}
	if got := session.Messages[1].ReplyStatus; got != domain.PracticeReplyStatusRetryableFailed {
		t.Fatalf("reply status=%s want retryable_failed", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLRepositoryGetSessionKeepsPendingBeforeLeaseBoundary(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	now := time.Date(2026, 7, 14, 8, 1, 29, 999999000, time.UTC)
	mock.ExpectBegin()
	mock.ExpectQuery(`select id from practice_sessions where id=\$1 and user_id=\$2 for update`).
		WithArgs("session-1", "user-1").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("session-1"))
	mock.ExpectExec(`(?s)update practice_messages.*reply_lease_expires_at <= \$3`).
		WithArgs(string(domain.PracticeReplyStatusRetryableFailed), "session-1", now, string(domain.PracticeReplyStatusPending)).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery(`select id, plan_id, target_job_id, status, language, created_at, updated_at`).WithArgs("user-1", "session-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "plan_id", "target_job_id", "status", "language", "created_at", "updated_at"}).
			AddRow("session-1", "plan-1", "target-1", "running", "zh-CN", now, now))
	mock.ExpectQuery(`select id, role, content, seq_no, client_message_id::text, reply_status, created_at from practice_messages`).WithArgs("session-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "role", "content", "seq_no", "client_message_id", "reply_status", "created_at"}).
			AddRow("m1", "assistant", "你好", 1, nil, nil, now).
			AddRow("m2", "user", "继续", 2, "client-1", string(domain.PracticeReplyStatusPending), now))
	mock.ExpectCommit()

	session, err := NewSQLRepository(db).GetSession(context.Background(), "user-1", "session-1", now)
	if err != nil {
		t.Fatalf("GetSession: %v", err)
	}
	if got := session.Messages[1].ReplyStatus; got != domain.PracticeReplyStatusPending {
		t.Fatalf("reply status=%s want pending", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLRepositoryReservePracticeMessageTakesOverExpiredPendingAtExactBoundary(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	now := time.Date(2026, 7, 14, 8, 1, 30, 0, time.UTC)
	expectReplyStateReservationContext(mock, now)
	mock.ExpectQuery(`(?s)select u\.id, u\.role, u\.content, u\.seq_no, u\.client_message_id::text, u\.reply_status, u\.reply_generation, u\.reply_lease_expires_at, u\.created_at.*for update of u`).
		WithArgs("session-1", "client-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "role", "content", "seq_no", "client_message_id", "reply_status", "reply_generation", "reply_lease_expires_at", "created_at", "assistant_id", "assistant_content", "assistant_seq", "assistant_created"}).
			AddRow("m2", "user", "继续", 2, "client-1", string(domain.PracticeReplyStatusPending), int64(1), now, now.Add(-90*time.Second), nil, nil, nil, nil))
	mock.ExpectQuery(`(?s)update practice_messages m.*set reply_status=\$1,.*reply_generation=reply_generation\+1,.*reply_lease_expires_at=\$2.*m\.reply_status=\$6.*m\.reply_generation=\$7.*returning m\.reply_generation`).
		WithArgs(string(domain.PracticeReplyStatusPending), now.Add(90*time.Second), "m2", "session-1", "user-1", string(domain.PracticeReplyStatusPending), int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"reply_generation"}).AddRow(int64(2)))
	mock.ExpectQuery(`select id, role, content, seq_no, client_message_id::text, reply_status, created_at from practice_messages`).WithArgs("session-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "role", "content", "seq_no", "client_message_id", "reply_status", "created_at"}).
			AddRow("m1", "assistant", "你好", 1, nil, nil, now).
			AddRow("m2", "user", "继续", 2, "client-1", string(domain.PracticeReplyStatusPending), now))
	mock.ExpectCommit()

	reservation, err := NewSQLRepository(db).ReservePracticeMessage(context.Background(), domain.ReservePracticeMessageInput{
		UserMessageID: "unused", UserID: "user-1", SessionID: "session-1", ClientMessageID: "client-1", Text: "继续", Now: now,
	})
	if err != nil {
		t.Fatalf("ReservePracticeMessage: %v", err)
	}
	if reservation.ReplyGeneration != 2 {
		t.Fatalf("generation=%d want 2", reservation.ReplyGeneration)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLRepositoryCommitPracticeMessageRejectsStaleGenerationAfterAuthorizationWithZeroWrites(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	now := time.Date(2026, 7, 14, 8, 2, 0, 0, time.UTC)
	in := domain.CommitPracticeMessageInput{
		UserID: "user-1", SessionID: "session-1", UserMessageID: "m2",
		ExpectedReplyGeneration: 1,
		AssistantMessageID:      "m3", AssistantText: "迟到回复", Now: now,
	}
	mock.ExpectBegin()
	mock.ExpectQuery(`select id from practice_sessions where id=\$1 and user_id=\$2 for update`).
		WithArgs(in.SessionID, in.UserID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(in.SessionID))
	mock.ExpectQuery(`(?s)select m\.id, m\.role, m\.content, m\.seq_no, m\.client_message_id::text, m\.reply_status, m\.reply_generation, m\.created_at.*where m\.id=\$1 and m\.session_id=\$2 and m\.role='user'.*for update of m`).
		WithArgs(in.UserMessageID, in.SessionID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "role", "content", "seq_no", "client_message_id", "reply_status", "reply_generation", "created_at"}).
			AddRow("m2", "user", "继续", 2, "client-1", string(domain.PracticeReplyStatusPending), int64(2), now))
	mock.ExpectRollback()

	_, err = NewSQLRepository(db).CommitPracticeMessage(context.Background(), in)
	if err != domain.ErrSessionConflict {
		t.Fatalf("error=%v want ErrSessionConflict", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLRepositoryFailPracticeMessageRejectsStaleGenerationWithZeroWrites(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	in := domain.FailPracticeMessageInput{
		UserID: "user-1", SessionID: "session-1", UserMessageID: "m2",
		ExpectedReplyGeneration: 1, ReplyStatus: domain.PracticeReplyStatusRetryableFailed,
	}
	mock.ExpectBegin()
	mock.ExpectQuery(`select id from practice_sessions where id=\$1 and user_id=\$2 for update`).
		WithArgs(in.SessionID, in.UserID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(in.SessionID))
	mock.ExpectExec(`(?s)update practice_messages m.*set reply_status=\$1, reply_lease_expires_at=null.*m\.reply_status=\$5 and m\.reply_generation=\$6`).
		WithArgs(string(in.ReplyStatus), in.UserMessageID, in.SessionID, in.UserID, string(domain.PracticeReplyStatusPending), in.ExpectedReplyGeneration).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectRollback()

	err = NewSQLRepository(db).FailPracticeMessage(context.Background(), in)
	if err != domain.ErrSessionConflict {
		t.Fatalf("error=%v want ErrSessionConflict", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
