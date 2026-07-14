package practice

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"reflect"
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
	return !strings.Contains(lower, "mode") && !strings.Contains(lower, "question") && !strings.Contains(lower, "hint") &&
		strings.Contains(lower, `"round_id":"round-1-technical"`) && strings.Contains(lower, `"round_sequence":1`)
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
		ResumeID: "resume-1", RoundID: "round-1-technical", Goal: sharedtypes.PracticeGoalBaseline, InterviewerPersona: sharedtypes.InterviewerRoleHiringManager,
		Difficulty: "standard", Language: "zh-CN", TimeBudgetMinutes: 30, Now: now}
	mock.ExpectBegin()
	query := `(?s)tj\.resume_id = \$10.*summary#>>'\{provenance,promptVersion\}'.*canonical_rounds.*pp\.resume_id = \$10.*session_completed.*insert into practice_plans.*round_id, round_sequence.*interviewer_persona, difficulty, language, time_budget_minutes.*resume_id, focus_dimension_codes`
	mock.ExpectQuery(query).WithArgs(in.PlanID, in.UserID, in.TargetJobID, "", string(in.Goal), string(in.InterviewerPersona),
		in.Difficulty, in.Language, in.TimeBudgetMinutes, in.ResumeID, nonNullEmptyTextArray{}, in.Now, in.RoundID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "target_job_id", "source_report_id", "goal", "round_id", "round_sequence", "interviewer_persona", "difficulty", "language", "time_budget_minutes", "resume_id", "focus_dimension_codes", "status", "created_at"}).
			AddRow(in.PlanID, in.TargetJobID, nil, string(in.Goal), in.RoundID, 1, string(in.InterviewerPersona), in.Difficulty, in.Language, in.TimeBudgetMinutes, in.ResumeID, `{}`, "ready", now))
	mock.ExpectExec(regexp.QuoteMeta("insert into audit_events")).WithArgs(in.AuditEventID, in.UserID, in.UserID, in.PlanID, metadataWithoutQuestionFields{}, in.Now).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	plan, err := NewSQLRepository(db).CreatePlan(context.Background(), in)
	if err != nil {
		t.Fatalf("CreatePlan: %v", err)
	}
	if plan.RoundID != in.RoundID || plan.RoundSequence != 1 {
		t.Fatalf("round identity = %+v", plan)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLRepositoryGetPlanKeepsLegacyNullRoundIdentity(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	now := time.Unix(1, 0).UTC()
	mock.ExpectQuery(`select id, target_job_id, source_report_id::text, goal, round_id, round_sequence`).
		WithArgs("user-1", "plan-legacy").
		WillReturnRows(sqlmock.NewRows([]string{"id", "target_job_id", "source_report_id", "goal", "round_id", "round_sequence", "interviewer_persona", "difficulty", "language", "time_budget_minutes", "resume_id", "focus_dimension_codes", "status", "created_at"}).
			AddRow("plan-legacy", "target-1", nil, string(sharedtypes.PracticeGoalBaseline), nil, nil, string(sharedtypes.InterviewerRoleHiringManager), "standard", "zh-CN", 30, "resume-1", `{}`, "ready", now))

	plan, err := NewSQLRepository(db).GetPlan(context.Background(), "user-1", "plan-legacy")
	if err != nil {
		t.Fatalf("GetPlan: %v", err)
	}
	if plan.RoundID != "" || plan.RoundSequence != 0 {
		t.Fatalf("legacy null identity must remain unguessed: %+v", plan)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLRepositoryReserveSessionStartPrefersCompleteResumeSourceSnapshot(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	now := time.Unix(5, 0).UTC()
	in := domain.StartSessionReservationInput{
		IdempotencyRecordID: "idem-1", SessionID: "session-1", UserID: "user-1", PlanID: "plan-1",
		IdempotencyKeyHash: "hash-1", RequestFingerprint: "fp-1", ExpiresAt: now.Add(time.Hour), Now: now,
	}
	const tailMarker = "START_RESUME_SNAPSHOT_TAIL_0712"

	mock.ExpectBegin()
	mock.ExpectExec(`select pg_advisory_xact_lock`).WithArgs(sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`select id, request_fingerprint, status, response_body, expires_at`).
		WithArgs(in.UserID, in.IdempotencyKeyHash).
		WillReturnRows(sqlmock.NewRows([]string{"id", "request_fingerprint", "status", "response_body", "expires_at"}))
	mock.ExpectExec(`insert into idempotency_records`).
		WithArgs(in.IdempotencyRecordID, in.UserID, in.IdempotencyKeyHash, in.RequestFingerprint, "pending", in.ExpiresAt, in.Now).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`(?s)with selected_plan as \(.*p\.focus_dimension_codes.*fr\.dimension_assessments.*fr\.issues.*p\.round_id.*r\.parsed_text_snapshot.*r\.original_text.*r\.structured_profile.*join target_jobs tj.*tj\.resume_id = p\.resume_id.*join feedback_reports fr.*btrim\(entry\.value->>'type'\) round_type.*2147483647.*jsonb_array_elements.*insert into practice_sessions`).
		WithArgs(in.SessionID, in.UserID, in.PlanID, in.Now).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "plan_id", "target_job_id", "goal", "interviewer_persona", "language", "role_title",
			"seniority", "top_skills", "resume_context", "focus_dimension_codes", "dimension_assessments", "issues",
			"round_id", "round_sequence", "round_type", "round_name", "round_focus", "created_at", "updated_at",
		}).AddRow(in.SessionID, in.PlanID, "target-1", string(sharedtypes.PracticeGoalRetryCurrentRound), string(sharedtypes.InterviewerRoleHiringManager),
			"zh-CN", "后端工程师", "senior", "Go", "# Complete resume\n"+tailMarker, `{system_design}`,
			`[{"code":"system_design","label":"系统设计","status":"needs_work","confidence":"high"}]`,
			`[{"dimensionCode":"system_design","evidence":"缺少容量估算","confidence":"high","sourceMessageSeqNos":[2]}]`,
			"round-1-technical", 1, "technical", "技术面", "系统设计", now, now))
	mock.ExpectCommit()

	reservation, err := NewSQLRepository(db).ReserveSessionStart(context.Background(), in)
	if err != nil {
		t.Fatalf("ReserveSessionStart: %v", err)
	}
	if !strings.Contains(reservation.ResumeContext, tailMarker) {
		t.Fatalf("resume context lost source snapshot tail: %+v", reservation)
	}
	if reservation.RoundID != "round-1-technical" || reservation.RoundSequence != 1 || reservation.RoundFocus != "系统设计" ||
		!reflect.DeepEqual(reservation.SemanticFocus, []domain.SemanticFocusDimension{{Code: "system_design", Label: "系统设计", Issues: []string{"缺少容量估算"}}}) {
		t.Fatalf("round context mismatch: %+v", reservation)
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
	mock.ExpectQuery(`(?s)select exists\(.*role='user'.*not exists \(`).
		WithArgs(in.SessionID).
		WillReturnRows(sqlmock.NewRows([]string{"has_user", "has_pending_reply"}).AddRow(true, false))
	expectCompletionReportContextLoad(mock, in, "en", 3, 3)
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

func TestE2EP0047RejectsZeroAnswerCompletion(t *testing.T) {
	t.Run("zero committed user messages", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatal(err)
		}
		defer db.Close()
		now := time.Unix(7, 0).UTC()
		in := domain.CompleteSessionStoreInput{UserID: "user-1", SessionID: "session-zero", Now: now}

		mock.ExpectBegin()
		mock.ExpectQuery(`select id, plan_id, target_job_id, status, language, created_at, updated_at`).
			WithArgs(in.UserID, in.SessionID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "plan_id", "target_job_id", "status", "language", "created_at", "updated_at"}).
				AddRow(in.SessionID, "plan-1", "target-1", string(sharedtypes.SessionStatusRunning), "zh-CN", now, now))
		mock.ExpectQuery(`select fr.id`).
			WithArgs(in.UserID, in.SessionID, "report_generate", in.SessionID).
			WillReturnRows(sqlmock.NewRows([]string{"report_id", "job_id", "job_type", "resource_type", "resource_id", "status", "error_code", "created_at", "updated_at"}))
		mock.ExpectQuery(`(?s)select exists\(.*role='user'.*not exists \(`).
			WithArgs(in.SessionID).
			WillReturnRows(sqlmock.NewRows([]string{"has_user", "has_pending_reply"}).AddRow(false, false))
		mock.ExpectRollback()

		_, err = NewSQLRepository(db).CompleteSession(context.Background(), in)
		if !errors.Is(err, domain.ErrSessionNotReportable) {
			t.Fatalf("error=%v want ErrSessionNotReportable", err)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("zero-answer completion produced side effects: %v", err)
		}
	})

	t.Run("pending assistant reply", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatal(err)
		}
		defer db.Close()
		now := time.Unix(8, 0).UTC()
		in := domain.CompleteSessionStoreInput{UserID: "user-1", SessionID: "session-pending", Now: now}

		mock.ExpectBegin()
		mock.ExpectQuery(`select id, plan_id, target_job_id, status, language, created_at, updated_at`).
			WithArgs(in.UserID, in.SessionID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "plan_id", "target_job_id", "status", "language", "created_at", "updated_at"}).
				AddRow(in.SessionID, "plan-1", "target-1", string(sharedtypes.SessionStatusRunning), "zh-CN", now, now))
		mock.ExpectQuery(`select fr.id`).
			WithArgs(in.UserID, in.SessionID, "report_generate", in.SessionID).
			WillReturnRows(sqlmock.NewRows([]string{"report_id", "job_id", "job_type", "resource_type", "resource_id", "status", "error_code", "created_at", "updated_at"}))
		mock.ExpectQuery(`(?s)select exists\(.*role='user'.*not exists \(`).
			WithArgs(in.SessionID).
			WillReturnRows(sqlmock.NewRows([]string{"has_user", "has_pending_reply"}).AddRow(true, true))
		mock.ExpectRollback()

		_, err = NewSQLRepository(db).CompleteSession(context.Background(), in)
		if !errors.Is(err, domain.ErrSessionNotReportable) {
			t.Fatalf("error=%v want ErrSessionNotReportable", err)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("pending-reply completion produced side effects: %v", err)
		}
	})

	t.Run("one answered user message", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatal(err)
		}
		defer db.Close()
		now := time.Unix(9, 0).UTC()
		in := domain.CompleteSessionStoreInput{
			UserID: "user-1", SessionID: "session-answered", ReportID: "report-1", JobID: "job-1",
			SessionEventID: "event-1", OutboxEventID: "outbox-1", AuditEventID: "audit-1",
			ClientCompletedAt: now.Add(-time.Second), Now: now,
		}

		mock.ExpectBegin()
		mock.ExpectQuery(`select id, plan_id, target_job_id, status, language, created_at, updated_at`).
			WithArgs(in.UserID, in.SessionID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "plan_id", "target_job_id", "status", "language", "created_at", "updated_at"}).
				AddRow(in.SessionID, "plan-1", "target-1", string(sharedtypes.SessionStatusRunning), "zh-CN", now, now))
		mock.ExpectQuery(`select fr.id`).
			WithArgs(in.UserID, in.SessionID, "report_generate", in.SessionID).
			WillReturnRows(sqlmock.NewRows([]string{"report_id", "job_id", "job_type", "resource_type", "resource_id", "status", "error_code", "created_at", "updated_at"}))
		mock.ExpectQuery(`(?s)select exists\(.*role='user'.*not exists \(`).
			WithArgs(in.SessionID).
			WillReturnRows(sqlmock.NewRows([]string{"has_user", "has_pending_reply"}).AddRow(true, false))
		expectCompletionReportContextLoad(mock, in, "zh-CN", 3, 3)
		mock.ExpectExec(`update practice_sessions`).WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectQuery(`select coalesce\(max\(seq_no\),0\)\+1 from practice_session_events`).
			WithArgs(in.SessionID).WillReturnRows(sqlmock.NewRows([]string{"seq_no"}).AddRow(4))
		mock.ExpectExec(`insert into practice_session_events`).WillReturnResult(sqlmock.NewResult(0, 1))
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
	})
}

func TestSQLRepositoryCompleteSessionReplayDoesNotAppendSecondCompletedFact(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	now := time.Unix(3, 0).UTC()
	in := domain.CompleteSessionStoreInput{
		UserID: "user-1", SessionID: "session-1", ReportID: "new-report", JobID: "new-job",
		SessionEventID: "new-event", OutboxEventID: "new-outbox", AuditEventID: "new-audit",
		ClientCompletedAt: now.Add(-time.Second), Now: now,
	}
	snapshot, err := domain.BuildReportContextSnapshot(domain.ReportContextSnapshotInput{
		TargetJob:    domain.ReportTargetJobSnapshot{ID: "target-1", Title: "Role", Language: "zh-CN", RawJD: "jd", Summary: json.RawMessage(reportContextSummary())},
		Resume:       domain.ReportResumeSnapshot{ID: "resume-1", DisplayName: "Resume", Language: "zh-CN", SourceSnapshot: "resume", StructuredProfile: json.RawMessage(`{}`)},
		Plan:         domain.ReportPlanSnapshot{ID: "plan-1", Goal: "baseline", InterviewerPersona: "hiring_manager", Difficulty: "standard", Language: "zh-CN", TimeBudgetMinutes: 45, ResumeID: "resume-1", RoundID: "round-1-technical", RoundSequence: 1},
		Conversation: domain.ReportConversationCoordinate{SessionID: in.SessionID, Language: "zh-CN", MessageCount: 3, LastMessageSeqNo: 3},
	})
	if err != nil {
		t.Fatal(err)
	}
	contextRaw, _ := json.Marshal(snapshot)

	mock.ExpectBegin()
	mock.ExpectQuery(`select id, plan_id, target_job_id, status, language, created_at, updated_at`).
		WithArgs(in.UserID, in.SessionID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "plan_id", "target_job_id", "status", "language", "created_at", "updated_at"}).
			AddRow(in.SessionID, "plan-1", "target-1", string(sharedtypes.SessionStatusCompleting), "zh-CN", now.Add(-time.Minute), now))
	mock.ExpectQuery(`(?s)select fr\.id, fr\.generation_context`).
		WithArgs(in.UserID, in.SessionID, "report_generate", in.SessionID).
		WillReturnRows(sqlmock.NewRows([]string{"report_id", "generation_context", "job_id", "job_type", "resource_type", "resource_id", "status", "error_code", "created_at", "updated_at"}).
			AddRow("report-1", contextRaw, "job-1", "report_generate", "feedback_report", "report-1", "queued", nil, now, now))
	mock.ExpectCommit()

	result, err := NewSQLRepository(db).CompleteSession(context.Background(), in)
	if err != nil {
		t.Fatalf("CompleteSession replay: %v", err)
	}
	if !result.Replay || result.ReportID != "report-1" || result.Job.ID != "job-1" {
		t.Fatalf("unexpected replay: %+v", result)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("replay attempted a second completion side effect: %v", err)
	}
}

func TestSQLRepositoryGetSessionReturnsOrderedMessages(t *testing.T) {
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
			AddRow("m2", "user", "你好", 2, "client-1", string(domain.PracticeReplyStatusPending), now))
	mock.ExpectCommit()
	session, err := NewSQLRepository(db).GetSession(context.Background(), "user-1", "session-1", now)
	if err != nil {
		t.Fatalf("GetSession: %v", err)
	}
	if len(session.Messages) != 2 || session.Messages[0].SeqNo != 1 || session.Messages[1].SeqNo != 2 {
		t.Fatalf("unexpected messages: %+v", session.Messages)
	}
}

func TestSQLRepositoryReservePracticeMessagePreservesGroundingOnRetryableFailure(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	now := time.Unix(2, 0).UTC()
	mock.ExpectBegin()
	const tailMarker = "SEND_RESUME_SNAPSHOT_TAIL_0712"
	mock.ExpectQuery(`(?s)select s\.id, s\.plan_id, s\.target_job_id.*r\.parsed_text_snapshot.*r\.original_text.*r\.structured_profile.*p\.focus_dimension_codes.*fr\.dimension_assessments.*fr\.issues.*p\.round_id.*from practice_sessions.*join target_jobs tj.*tj\.resume_id\s*=\s*p\.resume_id.*join feedback_reports fr.*btrim\(entry\.value->>'type'\) round_type.*2147483647.*jsonb_array_elements`).WithArgs("session-1", "user-1").WillReturnRows(
		sqlmock.NewRows([]string{"id", "plan_id", "target_job_id", "goal", "interviewer_persona", "language", "title", "seniority_level", "top_skills", "resume_context", "focus_dimension_codes", "dimension_assessments", "issues", "round_id", "round_sequence", "round_type", "round_name", "round_focus", "created_at", "updated_at"}).
			AddRow("session-1", "plan-1", "target-1", string(sharedtypes.PracticeGoalRetryCurrentRound), string(sharedtypes.InterviewerRoleHiringManager), "zh-CN", "后端工程师", "senior", "Go", "# Complete resume\n"+tailMarker, `{system_design}`,
				`[{"code":"system_design","label":"系统设计","status":"needs_work","confidence":"high"}]`,
				`[{"dimensionCode":"system_design","evidence":"缺少容量估算","confidence":"high","sourceMessageSeqNos":[2]}]`,
				"round-1-technical", 1, "technical", "技术面", "系统设计", now, now))
	mock.ExpectQuery(`select u.id, u.role, u.content, u.seq_no, u.client_message_id::text, u.reply_status`).WithArgs("session-1", "client-1").WillReturnRows(
		sqlmock.NewRows([]string{"id", "role", "content", "seq_no", "client_message_id", "reply_status", "reply_generation", "reply_lease_expires_at", "created_at", "assistant_id", "assistant_content", "assistant_seq", "assistant_created"}).
			AddRow("m2", "user", "继续", 2, "client-1", string(domain.PracticeReplyStatusRetryableFailed), int64(1), nil, now, nil, nil, nil, nil))
	mock.ExpectQuery(`(?s)update practice_messages m.*set reply_status=\$1,.*reply_generation=reply_generation\+1,.*reply_lease_expires_at=\$2.*returning m\.reply_generation`).
		WithArgs(string(domain.PracticeReplyStatusPending), now.Add(domain.PracticeReplyLeaseDuration), "m2", "session-1", "user-1", string(domain.PracticeReplyStatusRetryableFailed), int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"reply_generation"}).AddRow(int64(2)))
	mock.ExpectQuery(`select id, role, content, seq_no, client_message_id::text, reply_status, created_at from practice_messages`).WithArgs("session-1").WillReturnRows(
		sqlmock.NewRows([]string{"id", "role", "content", "seq_no", "client_message_id", "reply_status", "created_at"}).
			AddRow("m1", "assistant", "你好", 1, nil, nil, now).
			AddRow("m2", "user", "继续", 2, "client-1", string(domain.PracticeReplyStatusPending), now))
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
	if !strings.Contains(reservation.Session.ResumeContext, tailMarker) {
		t.Fatalf("send reservation lost source snapshot tail: %+v", reservation.Session)
	}
	if reservation.Session.RoundID != "round-1-technical" || reservation.Session.RoundFocus != "系统设计" {
		t.Fatalf("send reservation round mismatch: %+v", reservation.Session)
	}
	if !reflect.DeepEqual(reservation.Session.SemanticFocus, []domain.SemanticFocusDimension{{Code: "system_design", Label: "系统设计", Issues: []string{"缺少容量估算"}}}) {
		t.Fatalf("send reservation semantic focus mismatch: %+v", reservation.Session.SemanticFocus)
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
		sqlmock.NewRows([]string{"id", "plan_id", "target_job_id", "goal", "interviewer_persona", "language", "title", "seniority_level", "top_skills", "resume_context", "focus_dimension_codes", "dimension_assessments", "issues", "round_id", "round_sequence", "round_type", "round_name", "round_focus", "created_at", "updated_at"}).
			AddRow("session-1", "plan-1", "target-1", string(sharedtypes.PracticeGoalBaseline), string(sharedtypes.InterviewerRoleHiringManager), "zh-CN", "后端工程师", "senior", "Go", `{}`, `{}`, nil, nil, "round-1-technical", 1, "technical", "技术面", "系统设计", now, now))
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
		ExpectedReplyGeneration: 1, AssistantMessageID: "m3", AssistantText: "我们继续。", Now: now,
	}

	mock.ExpectBegin()
	mock.ExpectQuery(`select id from practice_sessions where id=\$1 and user_id=\$2 for update`).WithArgs(in.SessionID, in.UserID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(in.SessionID))
	mock.ExpectQuery(`select m.id, m.role, m.content, m.seq_no, m.client_message_id::text, m.reply_status, m.reply_generation, m.created_at`).
		WithArgs(in.UserMessageID, in.SessionID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "role", "content", "seq_no", "client_message_id", "reply_status", "reply_generation", "created_at"}).
			AddRow("m2", "user", "继续", 2, "client-1", string(domain.PracticeReplyStatusPending), int64(1), now))
	mock.ExpectExec(`insert into practice_messages`).
		WithArgs(in.AssistantMessageID, in.SessionID, 3, in.AssistantText, in.UserMessageID, in.Now).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta("update practice_messages set reply_status=$1, reply_lease_expires_at=null where id=$2 and session_id=$3 and role='user' and reply_status=$4 and reply_generation=$5")).
		WithArgs(string(domain.PracticeReplyStatusComplete), in.UserMessageID, in.SessionID, string(domain.PracticeReplyStatusPending), in.ExpectedReplyGeneration).
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
