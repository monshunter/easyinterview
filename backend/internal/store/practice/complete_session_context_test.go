package practice

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	domain "github.com/monshunter/easyinterview/backend/internal/practice"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

type frozenReportContextMatcher struct{}

func TestCompletionTransactionUsesRepeatableRead(t *testing.T) {
	opts := completionTransactionOptions()
	if opts == nil || opts.Isolation != sql.LevelRepeatableRead || opts.ReadOnly {
		t.Fatalf("completion transaction options=%+v want read-write repeatable-read", opts)
	}
}

func (frozenReportContextMatcher) Match(value driver.Value) bool {
	raw, ok := value.([]byte)
	if !ok {
		if text, stringOK := value.(string); stringOK {
			raw = []byte(text)
		} else {
			return false
		}
	}
	var snapshot domain.ReportContextSnapshot
	if err := json.Unmarshal(raw, &snapshot); err != nil || domain.ValidateReportContextSnapshot(snapshot) != nil {
		return false
	}
	return snapshot.SchemaVersion == domain.ReportContextSchemaVersion &&
		snapshot.TargetJob.RawJD == "complete jd" && snapshot.Resume.SourceSnapshot == "complete resume" &&
		snapshot.Round.ID == "round-1-technical" && snapshot.HasNextRound &&
		snapshot.Conversation.MessageCount == 3 && snapshot.Conversation.LastMessageSeqNo == 3
}

func TestE2EP0047FreezesReportContext(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	now := time.Unix(12, 0).UTC()
	in := domain.CompleteSessionStoreInput{
		UserID: "user-1", SessionID: "session-1", ReportID: "report-1", JobID: "job-1",
		SessionEventID: "event-1", OutboxEventID: "outbox-1", AuditEventID: "audit-1",
		ClientCompletedAt: now.Add(-time.Second), Now: now,
	}

	expectCompletionSessionAndEligibility(mock, in, now)
	mock.ExpectQuery(`(?s)select pp\.id, pp\.goal.*from practice_sessions ps.*for update of pp, tj, r`).
		WithArgs(in.UserID, in.SessionID).
		WillReturnRows(sqlmock.NewRows([]string{
			"plan_id", "goal", "interviewer_persona", "difficulty", "plan_language", "time_budget_minutes", "resume_id", "focus_dimension_codes", "round_id", "round_sequence",
			"target_job_id", "title", "company_name", "seniority_level", "target_language", "raw_jd_text", "summary",
			"resume_id", "resume_display_name", "resume_language", "resume_source_snapshot", "structured_profile", "session_language",
		}).AddRow(
			"plan-1", "baseline", "hiring_manager", "standard", "en", 45, "resume-1", `{system_design}`, "round-1-technical", 1,
			"target-1", "Platform Engineer", "Acme", "senior", "en", "complete jd", reportContextSummary(),
			"resume-1", "Backend Resume", "en", "complete resume", `{"skills":["Go"]}`, "en",
		))
	mock.ExpectQuery(`select kind, label, coalesce\(description,''\), evidence_level, display_order.*from target_job_requirements`).
		WithArgs("target-1").
		WillReturnRows(sqlmock.NewRows([]string{"kind", "label", "description", "evidence_level", "display_order"}).
			AddRow("must_have", "Go", "production Go", "explicit", 1))
	mock.ExpectQuery(`select count\(\*\), coalesce\(max\(seq_no\),0\).*from practice_messages`).
		WithArgs(in.SessionID).
		WillReturnRows(sqlmock.NewRows([]string{"message_count", "last_seq"}).AddRow(3, 3))
	mock.ExpectExec(`update practice_sessions`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`select coalesce\(max\(seq_no\),0\)\+1 from practice_session_events`).
		WithArgs(in.SessionID).WillReturnRows(sqlmock.NewRows([]string{"seq_no"}).AddRow(2))
	mock.ExpectExec(`insert into practice_session_events`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`(?s)insert into feedback_reports \(.*generation_context`).
		WithArgs(in.ReportID, in.UserID, in.SessionID, "target-1", string(sharedtypes.ReportStatusQueued), frozenReportContextMatcher{}, in.Now).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`(?s)insert into async_jobs \(\s*id, job_type, resource_type, resource_id, dedupe_key, status,\s*payload, available_at, created_at, updated_at\s*\) values \(\$1,\$2,'feedback_report',\$3,\$4,\$5,\$6,\$7,\$7,\$7\)`).
		WithArgs(in.JobID, "report_generate", in.ReportID, in.SessionID, string(sharedtypes.JobStatusQueued), sqlmock.AnyArg(), in.Now).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`insert into outbox_events`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`insert into audit_events`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	result, err := NewSQLRepository(db).CompleteSession(context.Background(), in)
	if err != nil {
		t.Fatalf("CompleteSession: %v", err)
	}
	if result.GenerationContext.SchemaVersion != domain.ReportContextSchemaVersion || !result.GenerationContext.HasNextRound {
		t.Fatalf("generation context=%+v", result.GenerationContext)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
	t.Log("REPORT_CONTEXT_SNAPSHOT_PASS")
}

func TestE2EP0047CompletionReplayPreservesReportContext(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	now := time.Unix(13, 0).UTC()
	in := domain.CompleteSessionStoreInput{UserID: "user-1", SessionID: "session-1", Now: now}
	snapshot, err := domain.BuildReportContextSnapshot(domain.ReportContextSnapshotInput{
		TargetJob:    domain.ReportTargetJobSnapshot{ID: "target-1", Title: "Role", Language: "en", RawJD: "jd", Summary: json.RawMessage(reportContextSummary())},
		Resume:       domain.ReportResumeSnapshot{ID: "resume-1", DisplayName: "Resume", Language: "en", SourceSnapshot: "resume", StructuredProfile: json.RawMessage(`{}`)},
		Plan:         domain.ReportPlanSnapshot{ID: "plan-1", Goal: "baseline", InterviewerPersona: "hiring_manager", Difficulty: "standard", Language: "en", TimeBudgetMinutes: 45, ResumeID: "resume-1", RoundID: "round-1-technical", RoundSequence: 1},
		Conversation: domain.ReportConversationCoordinate{SessionID: in.SessionID, Language: "en", MessageCount: 3, LastMessageSeqNo: 3},
	})
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(snapshot)

	mock.ExpectBegin()
	mock.ExpectQuery(`select id, plan_id, target_job_id, status, language, created_at, updated_at`).
		WithArgs(in.UserID, in.SessionID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "plan_id", "target_job_id", "status", "language", "created_at", "updated_at"}).
			AddRow(in.SessionID, "plan-1", "target-1", string(sharedtypes.SessionStatusCompleting), "en", now, now))
	mock.ExpectQuery(`(?s)select fr\.id, fr\.generation_context`).
		WithArgs(in.UserID, in.SessionID, "report_generate", in.SessionID).
		WillReturnRows(sqlmock.NewRows([]string{"report_id", "generation_context", "job_id", "job_type", "resource_type", "resource_id", "status", "error_code", "created_at", "updated_at"}).
			AddRow("report-1", raw, "job-1", "report_generate", "feedback_report", "report-1", "queued", nil, now, now))
	mock.ExpectCommit()

	result, err := NewSQLRepository(db).CompleteSession(context.Background(), in)
	if err != nil {
		t.Fatalf("CompleteSession replay: %v", err)
	}
	if !result.Replay || result.GenerationContext.SchemaVersion != domain.ReportContextSchemaVersion || result.GenerationContext.TargetJob.RawJD != "jd" {
		t.Fatalf("replay=%+v", result)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("replay attempted snapshot mutation: %v", err)
	}
	t.Log("REPORT_CONTEXT_REPLAY_PASS")
}

func expectCompletionSessionAndEligibility(mock sqlmock.Sqlmock, in domain.CompleteSessionStoreInput, now time.Time) {
	mock.ExpectBegin()
	mock.ExpectQuery(`select id, plan_id, target_job_id, status, language, created_at, updated_at`).
		WithArgs(in.UserID, in.SessionID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "plan_id", "target_job_id", "status", "language", "created_at", "updated_at"}).
			AddRow(in.SessionID, "plan-1", "target-1", string(sharedtypes.SessionStatusRunning), "en", now, now))
	mock.ExpectQuery(`select fr.id`).
		WithArgs(in.UserID, in.SessionID, "report_generate", in.SessionID).
		WillReturnRows(sqlmock.NewRows([]string{"report_id", "job_id", "job_type", "resource_type", "resource_id", "status", "error_code", "created_at", "updated_at"}))
	mock.ExpectQuery(`(?s)select exists\(.*role='user'.*not exists \(`).
		WithArgs(in.SessionID).
		WillReturnRows(sqlmock.NewRows([]string{"has_user", "has_pending_reply"}).AddRow(true, false))
}

func expectCompletionReportContextLoad(mock sqlmock.Sqlmock, in domain.CompleteSessionStoreInput, language string, messageCount, lastMessageSeqNo int) {
	mock.ExpectQuery(`(?s)select pp\.id, pp\.goal.*from practice_sessions ps.*for update of pp, tj, r`).
		WithArgs(in.UserID, in.SessionID).
		WillReturnRows(sqlmock.NewRows([]string{
			"plan_id", "goal", "interviewer_persona", "difficulty", "plan_language", "time_budget_minutes", "resume_id", "focus_dimension_codes", "round_id", "round_sequence",
			"target_job_id", "title", "company_name", "seniority_level", "target_language", "raw_jd_text", "summary",
			"resume_id", "resume_display_name", "resume_language", "resume_source_snapshot", "structured_profile", "session_language",
		}).AddRow(
			"plan-1", "baseline", "hiring_manager", "standard", language, 45, "resume-1", `{system_design}`, "round-1-technical", 1,
			"target-1", "Platform Engineer", "Acme", "senior", language, "complete jd", reportContextSummary(),
			"resume-1", "Backend Resume", language, "complete resume", `{"skills":["Go"]}`, language,
		))
	mock.ExpectQuery(`select kind, label, coalesce\(description,''\), evidence_level, display_order.*from target_job_requirements`).
		WithArgs("target-1").
		WillReturnRows(sqlmock.NewRows([]string{"kind", "label", "description", "evidence_level", "display_order"}).
			AddRow("must_have", "Go", "production Go", "explicit", 1))
	mock.ExpectQuery(`select count\(\*\), coalesce\(max\(seq_no\),0\).*from practice_messages`).
		WithArgs(in.SessionID).
		WillReturnRows(sqlmock.NewRows([]string{"message_count", "last_seq"}).AddRow(messageCount, lastMessageSeqNo))
}

func reportContextSummary() string {
	return `{"interviewRounds":[{"sequence":1,"type":"technical","name":"Technical","durationMinutes":45,"focus":"system design"},{"sequence":2,"type":"manager","name":"Manager","durationMinutes":30,"focus":"ownership"}],"provenance":{"promptVersion":"v0.1.0","rubricVersion":"v0.1.0","modelId":"fixture-model","language":"en","dataSourceVersion":"target-job.v1"}}`
}
