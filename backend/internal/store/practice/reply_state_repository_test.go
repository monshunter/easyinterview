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

func TestSQLRepositoryGetSessionReturnsUserReplyRecoveryStateOnly(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	now := time.Unix(2, 0).UTC()
	mock.ExpectBegin()
	mock.ExpectQuery(`select id from practice_sessions where id=\$1 and user_id=\$2 for update`).WithArgs("session-1", "user-1").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("session-1"))
	mock.ExpectExec(`(?s)update practice_messages.*reply_lease_expires_at <= \$3`).
		WithArgs(string(domain.PracticeReplyStatusRetryableFailed), "session-1", now, string(domain.PracticeReplyStatusPending)).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery(`select id, plan_id, target_job_id, status, language, created_at, updated_at`).WithArgs("user-1", "session-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "plan_id", "target_job_id", "status", "language", "created_at", "updated_at"}).
			AddRow("session-1", "plan-1", "target-1", string(sharedtypes.SessionStatusRunning), "zh-CN", now, now))
	mock.ExpectQuery(`select id, role, content, seq_no, client_message_id::text, reply_status, created_at from practice_messages`).WithArgs("session-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "role", "content", "seq_no", "client_message_id", "reply_status", "created_at"}).
			AddRow("m1", "assistant", "你好", 1, nil, nil, now).
			AddRow("m2", "user", "继续", 2, "client-1", string(domain.PracticeReplyStatusRetryableFailed), now))
	mock.ExpectCommit()

	session, err := NewSQLRepository(db).GetSession(context.Background(), "user-1", "session-1", now)
	if err != nil {
		t.Fatalf("GetSession: %v", err)
	}
	if len(session.Messages) != 2 || session.Messages[0].ClientMessageID != "" || session.Messages[0].ReplyStatus != "" {
		t.Fatalf("assistant recovery fields = %+v", session.Messages)
	}
	if session.Messages[1].ClientMessageID != "client-1" || session.Messages[1].ReplyStatus != domain.PracticeReplyStatusRetryableFailed {
		t.Fatalf("user recovery fields = %+v", session.Messages[1])
	}
}

func TestSQLRepositoryReservePracticeMessageRetriesOnlyRetryableFailure(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	now := time.Unix(2, 0).UTC()
	expectReplyStateReservationContext(mock, now)
	mock.ExpectQuery(`(?s)select u\.id, u\.role, u\.content, u\.seq_no, u\.client_message_id::text, u\.reply_status,.*u\.reply_generation, u\.reply_lease_expires_at, u\.created_at.*for update of u`).
		WithArgs("session-1", "client-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "role", "content", "seq_no", "client_message_id", "reply_status", "reply_generation", "reply_lease_expires_at", "created_at", "assistant_id", "assistant_content", "assistant_seq", "assistant_created"}).
			AddRow("m2", "user", "继续", 2, "client-1", string(domain.PracticeReplyStatusRetryableFailed), int64(1), nil, now, nil, nil, nil, nil))
	mock.ExpectQuery(`(?s)update practice_messages m.*set reply_status=\$1,.*reply_generation=reply_generation\+1,.*reply_lease_expires_at=\$2.*returning m\.reply_generation`).
		WithArgs(string(domain.PracticeReplyStatusPending), now.Add(domain.PracticeReplyLeaseDuration), "m2", "session-1", "user-1", string(domain.PracticeReplyStatusRetryableFailed), int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"reply_generation"}).AddRow(int64(2)))
	mock.ExpectQuery(`select id, role, content, seq_no, client_message_id::text, reply_status, created_at from practice_messages`).WithArgs("session-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "role", "content", "seq_no", "client_message_id", "reply_status", "created_at"}).
			AddRow("m1", "assistant", "你好", 1, nil, nil, now).
			AddRow("m2", "user", "继续", 2, "client-1", string(domain.PracticeReplyStatusPending), now))
	mock.ExpectCommit()

	reservation, err := NewSQLRepository(db).ReservePracticeMessage(context.Background(), domain.ReservePracticeMessageInput{
		UserMessageID: "new-id", UserID: "user-1", SessionID: "session-1", ClientMessageID: "client-1", Text: "继续", Now: now,
	})
	if err != nil {
		t.Fatalf("ReservePracticeMessage: %v", err)
	}
	if reservation.UserMessage.ID != "m2" || reservation.UserMessage.ReplyStatus != domain.PracticeReplyStatusPending || len(reservation.History) != 1 {
		t.Fatalf("retry reservation = %+v", reservation)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLRepositoryReservePracticeMessageRejectsPendingAndTerminalSameID(t *testing.T) {
	for _, status := range []domain.PracticeReplyStatus{domain.PracticeReplyStatusPending, domain.PracticeReplyStatusTerminalFailed} {
		t.Run(string(status), func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			now := time.Unix(2, 0).UTC()
			var lease any
			if status == domain.PracticeReplyStatusPending {
				lease = now.Add(domain.PracticeReplyLeaseDuration)
			}
			expectReplyStateReservationContext(mock, now)
			mock.ExpectQuery(`(?s)select u\.id, u\.role, u\.content, u\.seq_no, u\.client_message_id::text, u\.reply_status,.*u\.reply_generation, u\.reply_lease_expires_at, u\.created_at.*for update of u`).
				WithArgs("session-1", "client-1").
				WillReturnRows(sqlmock.NewRows([]string{"id", "role", "content", "seq_no", "client_message_id", "reply_status", "reply_generation", "reply_lease_expires_at", "created_at", "assistant_id", "assistant_content", "assistant_seq", "assistant_created"}).
					AddRow("m2", "user", "继续", 2, "client-1", string(status), int64(1), lease, now, nil, nil, nil, nil))
			mock.ExpectRollback()

			_, err = NewSQLRepository(db).ReservePracticeMessage(context.Background(), domain.ReservePracticeMessageInput{
				UserMessageID: "new-id", UserID: "user-1", SessionID: "session-1", ClientMessageID: "client-1", Text: "继续", Now: now,
			})
			if !errors.Is(err, domain.ErrSessionConflict) {
				t.Fatalf("error=%v want ErrSessionConflict", err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestSQLRepositoryFailPracticeMessageTransitionsPendingAtomically(t *testing.T) {
	for _, status := range []domain.PracticeReplyStatus{domain.PracticeReplyStatusRetryableFailed, domain.PracticeReplyStatusTerminalFailed} {
		t.Run(string(status), func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			mock.ExpectBegin()
			mock.ExpectQuery(`select id from practice_sessions where id=\$1 and user_id=\$2 for update`).WithArgs("session-1", "user-1").
				WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("session-1"))
			mock.ExpectExec(`(?s)update practice_messages m.*set reply_status=\$1, reply_lease_expires_at=null.*from practice_sessions s.*m\.reply_status=\$5 and m\.reply_generation=\$6`).
				WithArgs(string(status), "m2", "session-1", "user-1", string(domain.PracticeReplyStatusPending), int64(1)).
				WillReturnResult(sqlmock.NewResult(0, 1))
			mock.ExpectCommit()

			err = NewSQLRepository(db).FailPracticeMessage(context.Background(), domain.FailPracticeMessageInput{
				UserID: "user-1", SessionID: "session-1", UserMessageID: "m2", ExpectedReplyGeneration: 1, ReplyStatus: status,
			})
			if err != nil {
				t.Fatalf("FailPracticeMessage: %v", err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestSQLRepositoryCommitPracticeMessageInsertsReplyAndCompletesUserAtomically(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	now := time.Unix(4, 0).UTC()
	in := domain.CommitPracticeMessageInput{
		UserID: "user-1", SessionID: "session-1", UserMessageID: "m2",
		ExpectedReplyGeneration: 1, AssistantMessageID: "m3", AssistantText: "我们继续。", Now: now,
	}
	mock.ExpectBegin()
	mock.ExpectQuery(`select id from practice_sessions where id=\$1 and user_id=\$2 for update`).WithArgs(in.SessionID, in.UserID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(in.SessionID))
	mock.ExpectQuery(`select m.id, m.role, m.content, m.seq_no, m.client_message_id::text, m.reply_status, m.reply_generation, m.created_at`).
		WithArgs(in.UserMessageID, in.SessionID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "role", "content", "seq_no", "client_message_id", "reply_status", "reply_generation", "created_at"}).
			AddRow("m2", "user", "继续", 2, "client-1", string(domain.PracticeReplyStatusPending), int64(1), now))
	mock.ExpectQuery(`select coalesce\(sum\(octet_length\(content\)\),0\) from practice_messages`).WithArgs(in.SessionID).
		WillReturnRows(sqlmock.NewRows([]string{"bytes"}).AddRow(20))
	mock.ExpectExec(`insert into practice_messages`).
		WithArgs(in.AssistantMessageID, in.SessionID, 3, in.AssistantText, in.UserMessageID, in.Now).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta("update practice_messages set reply_status=$1, reply_lease_expires_at=null where id=$2 and session_id=$3 and role='user' and reply_status=$4 and reply_generation=$5")).
		WithArgs(string(domain.PracticeReplyStatusComplete), in.UserMessageID, in.SessionID, string(domain.PracticeReplyStatusPending), in.ExpectedReplyGeneration).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`update practice_sessions set status=\$1, updated_at=\$2 where id=\$3 and user_id=\$4 and status in \(\$5,\$6\)`).
		WithArgs(string(sharedtypes.SessionStatusRunning), in.Now, in.SessionID, in.UserID,
			string(sharedtypes.SessionStatusRunning), string(sharedtypes.SessionStatusWaitingUserInput)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`select id, plan_id, target_job_id, status, language, created_at, updated_at`).WithArgs(in.UserID, in.SessionID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "plan_id", "target_job_id", "status", "language", "created_at", "updated_at"}).
			AddRow(in.SessionID, "plan-1", "target-1", string(sharedtypes.SessionStatusRunning), "zh-CN", now, now))
	mock.ExpectQuery(`select id, role, content, seq_no, client_message_id::text, reply_status, created_at from practice_messages`).WithArgs(in.SessionID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "role", "content", "seq_no", "client_message_id", "reply_status", "created_at"}).
			AddRow("m1", "assistant", "你好", 1, nil, nil, now).
			AddRow("m2", "user", "继续", 2, "client-1", string(domain.PracticeReplyStatusComplete), now).
			AddRow("m3", "assistant", "我们继续。", 3, nil, nil, now))
	mock.ExpectCommit()

	in.MaxSessionTextBytes = 35
	result, err := NewSQLRepository(db).CommitPracticeMessage(context.Background(), in)
	if err != nil {
		t.Fatalf("CommitPracticeMessage: %v", err)
	}
	if result.UserMessage.ReplyStatus != domain.PracticeReplyStatusComplete || result.UserMessage.ClientMessageID != "client-1" || result.AssistantMessage.ReplyStatus != "" {
		t.Fatalf("commit result = %+v", result)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLRepositoryCommitPracticeMessageRejectsAssistantAggregateLimitPlusOneBeforeInsert(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	now := time.Unix(4, 0).UTC()
	in := domain.CommitPracticeMessageInput{
		UserID: "user-1", SessionID: "session-1", UserMessageID: "m2",
		ExpectedReplyGeneration: 1, AssistantMessageID: "m3", AssistantText: "我们继续。", MaxSessionTextBytes: 34, Now: now,
	}
	mock.ExpectBegin()
	mock.ExpectQuery(`select id from practice_sessions where id=\$1 and user_id=\$2 for update`).WithArgs(in.SessionID, in.UserID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(in.SessionID))
	mock.ExpectQuery(`select m.id, m.role, m.content, m.seq_no, m.client_message_id::text, m.reply_status, m.reply_generation, m.created_at`).
		WithArgs(in.UserMessageID, in.SessionID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "role", "content", "seq_no", "client_message_id", "reply_status", "reply_generation", "created_at"}).
			AddRow("m2", "user", "继续", 2, "client-1", string(domain.PracticeReplyStatusPending), int64(1), now))
	mock.ExpectQuery(`select coalesce\(sum\(octet_length\(content\)\),0\) from practice_messages`).WithArgs(in.SessionID).
		WillReturnRows(sqlmock.NewRows([]string{"bytes"}).AddRow(20))
	mock.ExpectRollback()

	_, err = NewSQLRepository(db).CommitPracticeMessage(context.Background(), in)
	if err != domain.ErrPracticeSessionTextLimitExceeded {
		t.Fatalf("error=%v want ErrPracticeSessionTextLimitExceeded", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func expectReplyStateReservationContext(mock sqlmock.Sqlmock, now time.Time) {
	mock.ExpectBegin()
	mock.ExpectQuery(`(?s)select s\.id, s\.plan_id, s\.target_job_id.*r\.parsed_text_snapshot.*p\.round_id.*from practice_sessions.*for update of s`).
		WithArgs("session-1", "user-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "plan_id", "target_job_id", "goal", "interviewer_persona", "language", "title", "seniority_level", "top_skills", "resume_context", "focus_dimension_codes", "dimension_assessments", "issues", "round_id", "round_sequence", "round_type", "round_name", "round_focus", "created_at", "updated_at"}).
			AddRow("session-1", "plan-1", "target-1", string(sharedtypes.PracticeGoalBaseline), "hiring_manager", "zh-CN", "后端工程师", "senior", "Go", "完整简历", `{}`, nil, nil, "round-1-technical", 1, "technical", "技术面", "系统设计", now, now))
}
