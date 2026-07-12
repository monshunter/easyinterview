package practice

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	domain "github.com/monshunter/easyinterview/backend/internal/practice"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

type metadataWithoutQuestionFields struct{}

func (metadataWithoutQuestionFields) Match(value driver.Value) bool {
	raw, ok := value.([]byte)
	if !ok {
		return false
	}
	lower := strings.ToLower(string(raw))
	return !strings.Contains(lower, "mode") && !strings.Contains(lower, "question") && !strings.Contains(lower, "hint")
}

type nonNullEmptyTextArray struct{}

func (nonNullEmptyTextArray) Match(value driver.Value) bool {
	switch typed := value.(type) {
	case string:
		return typed == "{}"
	case []byte:
		return string(typed) == "{}"
	default:
		return false
	}
}

func TestSQLRepositoryCreatePlanUsesConversationColumns(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	now := time.Unix(1, 0).UTC()
	in := domain.CreatePlanStoreInput{PlanID: "plan-1", AuditEventID: "audit-1", UserID: "user-1", TargetJobID: "target-1",
		ResumeID: "resume-1", Goal: sharedtypes.PracticeGoalBaseline, InterviewerPersona: sharedtypes.InterviewerRoleHiringManager,
		Difficulty: "standard", Language: "zh-CN", TimeBudgetMinutes: 30, Now: now}
	mock.ExpectBegin()
	query := regexp.QuoteMeta("insert into practice_plans") + `(?s).*interviewer_persona, difficulty, language, time_budget_minutes.*resume_id, focus_competency_codes`
	mock.ExpectQuery(query).WithArgs(in.PlanID, in.UserID, in.TargetJobID, "", string(in.Goal), string(in.InterviewerPersona),
		in.Difficulty, in.Language, in.TimeBudgetMinutes, in.ResumeID, nonNullEmptyTextArray{}, in.Now).
		WillReturnRows(sqlmock.NewRows([]string{"id", "target_job_id", "source_report_id", "goal", "interviewer_persona", "difficulty", "language", "time_budget_minutes", "resume_id", "status", "created_at"}).
			AddRow(in.PlanID, in.TargetJobID, nil, string(in.Goal), string(in.InterviewerPersona), in.Difficulty, in.Language, in.TimeBudgetMinutes, in.ResumeID, "ready", now))
	mock.ExpectExec(regexp.QuoteMeta("insert into audit_events")).WithArgs(in.AuditEventID, in.UserID, in.UserID, in.PlanID, metadataWithoutQuestionFields{}, in.Now).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	if _, err := NewSQLRepository(db).CreatePlan(context.Background(), in); err != nil {
		t.Fatalf("CreatePlan: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLRepositoryCompleteSessionUsesLifecycleOnlyEventColumns(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	now := time.Unix(3, 0).UTC()
	in := domain.CompleteSessionStoreInput{
		UserID: "user-1", SessionID: "session-1", ReportID: "report-1", JobID: "job-1",
		SessionEventID: "event-1", OutboxEventID: "outbox-1", AuditEventID: "audit-1",
		ClientCompletedAt: now.Add(-time.Second), Now: now,
	}

	mock.ExpectBegin()
	mock.ExpectQuery(`select id, plan_id, target_job_id, status, language, created_at, updated_at`).
		WithArgs(in.UserID, in.SessionID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "plan_id", "target_job_id", "status", "language", "created_at", "updated_at"}).
			AddRow(in.SessionID, "plan-1", "target-1", string(sharedtypes.SessionStatusRunning), "en", now.Add(-time.Minute), now.Add(-time.Minute)))
	mock.ExpectQuery(`select fr.id`).
		WithArgs(in.UserID, in.SessionID, "report_generate", in.SessionID).
		WillReturnRows(sqlmock.NewRows([]string{"report_id", "job_id", "job_type", "resource_type", "resource_id", "status", "error_code", "created_at", "updated_at"}))
	mock.ExpectExec(`update practice_sessions`).
		WithArgs(string(sharedtypes.SessionStatusCompleting), in.Now, in.SessionID, in.UserID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`select coalesce\(max\(seq_no\),0\)\+1 from practice_session_events`).
		WithArgs(in.SessionID).
		WillReturnRows(sqlmock.NewRows([]string{"seq_no"}).AddRow(2))
	eventInsert := regexp.QuoteMeta("insert into practice_session_events (\n  id, session_id, seq_no, event_type, payload, created_at\n)")
	mock.ExpectExec(eventInsert).
		WithArgs(in.SessionEventID, in.SessionID, 2, sqlmock.AnyArg(), in.Now).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`insert into feedback_reports`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`insert into async_jobs`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`insert into outbox_events`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`insert into audit_events`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	result, err := NewSQLRepository(db).CompleteSession(context.Background(), in)
	if err != nil {
		t.Fatalf("CompleteSession: %v", err)
	}
	if result.ReportID != in.ReportID || result.Job.ID != in.JobID {
		t.Fatalf("unexpected completion result: %+v", result)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLRepositoryGetSessionReturnsOrderedMessages(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	now := time.Unix(2, 0).UTC()
	mock.ExpectQuery(`select id, plan_id, target_job_id, status, language, created_at, updated_at`).WithArgs("user-1", "session-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "plan_id", "target_job_id", "status", "language", "created_at", "updated_at"}).
			AddRow("session-1", "plan-1", "target-1", string(sharedtypes.SessionStatusRunning), "zh-CN", now, now))
	mock.ExpectQuery(`select id, role, content, seq_no, created_at from practice_messages`).WithArgs("session-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "role", "content", "seq_no", "created_at"}).
			AddRow("m1", "assistant", "你好", 1, now).AddRow("m2", "user", "你好", 2, now))
	session, err := NewSQLRepository(db).GetSession(context.Background(), "user-1", "session-1")
	if err != nil {
		t.Fatalf("GetSession: %v", err)
	}
	if len(session.Messages) != 2 || session.Messages[0].SeqNo != 1 || session.Messages[1].SeqNo != 2 {
		t.Fatalf("unexpected messages: %+v", session.Messages)
	}
}

func TestSQLRepositoryReservePracticeMessageRetriesPendingUserMessage(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	now := time.Unix(2, 0).UTC()
	mock.ExpectBegin()
	mock.ExpectQuery(`select s.id, s.plan_id, s.target_job_id`).WithArgs("session-1", "user-1").WillReturnRows(
		sqlmock.NewRows([]string{"id", "plan_id", "target_job_id", "goal", "interviewer_persona", "language", "title", "seniority_level", "top_skills", "structured_profile", "focus_competency_codes", "created_at", "updated_at"}).
			AddRow("session-1", "plan-1", "target-1", string(sharedtypes.PracticeGoalBaseline), string(sharedtypes.InterviewerRoleHiringManager), "zh-CN", "后端工程师", "senior", "Go", `{}`, `{technical_depth}`, now, now))
	mock.ExpectQuery(`select u.id, u.role, u.content, u.seq_no`).WithArgs("session-1", "client-1").WillReturnRows(
		sqlmock.NewRows([]string{"id", "role", "content", "seq_no", "created_at", "assistant_id", "assistant_content", "assistant_seq", "assistant_created"}).
			AddRow("m2", "user", "继续", 2, now, nil, nil, nil, nil))
	mock.ExpectQuery(`select id, role, content, seq_no, created_at from practice_messages`).WithArgs("session-1").WillReturnRows(
		sqlmock.NewRows([]string{"id", "role", "content", "seq_no", "created_at"}).
			AddRow("m1", "assistant", "你好", 1, now).AddRow("m2", "user", "继续", 2, now))
	mock.ExpectCommit()

	reservation, err := NewSQLRepository(db).ReservePracticeMessage(context.Background(), domain.ReservePracticeMessageInput{
		UserMessageID: "new-id", UserID: "user-1", SessionID: "session-1", ClientMessageID: "client-1", Text: "继续", Now: now,
	})
	if err != nil {
		t.Fatalf("ReservePracticeMessage: %v", err)
	}
	if reservation.UserMessage.ID != "m2" || len(reservation.History) != 1 || reservation.History[0].ID != "m1" {
		t.Fatalf("unexpected retry reservation: %+v", reservation)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLRepositoryReservePracticeMessageRejectsNewMessageWhileReplyPending(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	now := time.Unix(2, 0).UTC()
	mock.ExpectBegin()
	mock.ExpectQuery(`select s.id, s.plan_id, s.target_job_id`).WithArgs("session-1", "user-1").WillReturnRows(
		sqlmock.NewRows([]string{"id", "plan_id", "target_job_id", "goal", "interviewer_persona", "language", "title", "seniority_level", "top_skills", "structured_profile", "focus_competency_codes", "created_at", "updated_at"}).
			AddRow("session-1", "plan-1", "target-1", string(sharedtypes.PracticeGoalBaseline), string(sharedtypes.InterviewerRoleHiringManager), "zh-CN", "后端工程师", "senior", "Go", `{}`, `{technical_depth}`, now, now))
	mock.ExpectQuery(`select u.id, u.role, u.content, u.seq_no`).WithArgs("session-1", "client-new").WillReturnRows(
		sqlmock.NewRows([]string{"id", "role", "content", "seq_no", "created_at", "assistant_id", "assistant_content", "assistant_seq", "assistant_created"}))
	mock.ExpectQuery(`select exists`).WithArgs("session-1").WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectRollback()

	_, err = NewSQLRepository(db).ReservePracticeMessage(context.Background(), domain.ReservePracticeMessageInput{
		UserMessageID: "m-new", UserID: "user-1", SessionID: "session-1", ClientMessageID: "client-new", Text: "another", Now: now,
	})
	if err != domain.ErrSessionConflict {
		t.Fatalf("error=%v want ErrSessionConflict", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLRepositoryCommitPracticeMessageRejectsClosedSession(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	now := time.Unix(4, 0).UTC()
	in := domain.CommitPracticeMessageInput{
		UserID: "user-1", SessionID: "session-1", UserMessageID: "m2",
		AssistantMessageID: "m3", AssistantText: "我们继续。", Now: now,
	}

	mock.ExpectBegin()
	mock.ExpectQuery(`select m.id, m.role, m.content, m.seq_no, m.created_at`).
		WithArgs(in.UserMessageID, in.SessionID, in.UserID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "role", "content", "seq_no", "created_at"}).
			AddRow("m2", "user", "继续", 2, now))
	mock.ExpectExec(`insert into practice_messages`).
		WithArgs(in.AssistantMessageID, in.SessionID, 3, in.AssistantText, in.UserMessageID, in.Now).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`update practice_sessions set status=\$1, updated_at=\$2 where id=\$3 and user_id=\$4 and status in \(\$5,\$6\)`).
		WithArgs(string(sharedtypes.SessionStatusRunning), in.Now, in.SessionID, in.UserID,
			string(sharedtypes.SessionStatusRunning), string(sharedtypes.SessionStatusWaitingUserInput)).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectRollback()

	_, err = NewSQLRepository(db).CommitPracticeMessage(context.Background(), in)
	if err != domain.ErrSessionConflict {
		t.Fatalf("error=%v want ErrSessionConflict", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPracticeOutboxPayloadContainsOnlyLifecycleData(t *testing.T) {
	payload, err := BuildPracticeSessionCompletedPayload(PracticeSessionCompletedInput{Language: "zh-CN", PlanID: "plan-1", SessionID: "session-1", TargetJobID: "target-1"})
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(payload)
	for _, stale := range []string{"content", "question", "turn", "hint", "mode"} {
		if strings.Contains(strings.ToLower(string(raw)), stale) {
			t.Fatalf("payload leaks %s: %s", stale, raw)
		}
	}
}
