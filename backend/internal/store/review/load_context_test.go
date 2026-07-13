package review

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	practicedomain "github.com/monshunter/easyinterview/backend/internal/practice"
)

func TestLoadReportContextUsesFrozenSnapshotAndTerminalOrderedMessages(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	snapshot := frozenReportContextSnapshot(t)
	raw, err := json.Marshal(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	mock.ExpectQuery(`(?s)select fr\.user_id::text, fr\.id::text, fr\.session_id::text, fr\.target_job_id::text, fr\.generation_context\s+from feedback_reports fr\s+where fr\.id = \$1`).
		WithArgs("report-1").
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "report_id", "session_id", "target_job_id", "generation_context"}).
			AddRow("user-1", "report-1", "session-1", "target-1", raw))
	mock.ExpectQuery(`(?s)select role, content, seq_no\s+from practice_messages\s+where session_id = \$1\s+order by seq_no asc`).
		WithArgs("session-1").
		WillReturnRows(sqlmock.NewRows([]string{"role", "content", "seq_no"}).
			AddRow("assistant", "请介绍一个项目。", 1).
			AddRow("user", "我负责迁移。", 2).
			AddRow("assistant", "你如何验证结果？", 3))

	got, err := NewRepository(db).LoadReportContext(context.Background(), "report-1")
	if err != nil {
		t.Fatalf("LoadReportContext: %v", err)
	}
	if !reflect.DeepEqual(got.FrozenContext, snapshot) {
		t.Fatalf("frozen context drifted:\n got: %#v\nwant: %#v", got.FrozenContext, snapshot)
	}
	if got.Session.UserID != "user-1" || got.Session.ReportID != "report-1" || got.Session.SessionID != "session-1" ||
		got.Session.TargetJobID != "target-1" || got.Session.Language != "zh-CN" {
		t.Fatalf("session projection did not come from frozen context: %#v", got.Session)
	}
	if len(got.Messages) != 3 || got.Messages[0].SeqNo != 1 || got.Messages[1].Role != "user" || got.Messages[2].SeqNo != 3 {
		t.Fatalf("terminal ordered messages were not preserved: %#v", got.Messages)
	}
	contextType := reflect.TypeOf(got)
	for _, staleField := range []string{"Rubric", "ReportPromptVersion", "ReportRubricVersion", "ModelID", "Provider", "FeatureFlag", "DataSourceVersion"} {
		if _, exists := contextType.FieldByName(staleField); exists {
			t.Fatalf("ReportContext retains loader-owned generation metadata field %s", staleField)
		}
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestLoadReportContextRejectsFrozenIdentityAndCoordinateMismatches(t *testing.T) {
	base := frozenReportContextSnapshot(t)
	tests := []struct {
		name           string
		mutateSnapshot func(*practicedomain.ReportContextSnapshot)
		rowReportID    string
		rowSessionID   string
		rowTargetJobID string
		messages       *sqlmock.Rows
		expectMessages bool
	}{
		{name: "report row id", rowReportID: "report-2", rowSessionID: "session-1", rowTargetJobID: "target-1"},
		{name: "session row id", rowReportID: "report-1", rowSessionID: "session-2", rowTargetJobID: "target-1"},
		{name: "target job row id", rowReportID: "report-1", rowSessionID: "session-1", rowTargetJobID: "target-2"},
		{
			name: "session language", rowReportID: "report-1", rowSessionID: "session-1", rowTargetJobID: "target-1",
			mutateSnapshot: func(snapshot *practicedomain.ReportContextSnapshot) { snapshot.Plan.Language = "en" },
		},
		{
			name: "message count", rowReportID: "report-1", rowSessionID: "session-1", rowTargetJobID: "target-1", expectMessages: true,
			messages: sqlmock.NewRows([]string{"role", "content", "seq_no"}).
				AddRow("assistant", "question", 1).
				AddRow("user", "answer", 2),
		},
		{
			name: "last message sequence", rowReportID: "report-1", rowSessionID: "session-1", rowTargetJobID: "target-1", expectMessages: true,
			messages: sqlmock.NewRows([]string{"role", "content", "seq_no"}).
				AddRow("assistant", "question", 1).
				AddRow("user", "answer", 2).
				AddRow("assistant", "follow-up", 4),
		},
		{
			name: "message role ordering", rowReportID: "report-1", rowSessionID: "session-1", rowTargetJobID: "target-1", expectMessages: true,
			messages: sqlmock.NewRows([]string{"role", "content", "seq_no"}).
				AddRow("user", "wrong first role", 1).
				AddRow("assistant", "wrong second role", 2).
				AddRow("assistant", "follow-up", 3),
		},
		{
			name: "message sequence ordering", rowReportID: "report-1", rowSessionID: "session-1", rowTargetJobID: "target-1", expectMessages: true,
			messages: sqlmock.NewRows([]string{"role", "content", "seq_no"}).
				AddRow("assistant", "question", 1).
				AddRow("assistant", "follow-up", 3).
				AddRow("user", "answer", 2),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			snapshot := base
			if tt.mutateSnapshot != nil {
				tt.mutateSnapshot(&snapshot)
			}
			raw, err := json.Marshal(snapshot)
			if err != nil {
				t.Fatal(err)
			}
			mock.ExpectQuery(`(?s)select fr\.user_id::text, fr\.id::text, fr\.session_id::text, fr\.target_job_id::text, fr\.generation_context\s+from feedback_reports fr\s+where fr\.id = \$1`).
				WithArgs("report-1").
				WillReturnRows(sqlmock.NewRows([]string{"user_id", "report_id", "session_id", "target_job_id", "generation_context"}).
					AddRow("user-1", tt.rowReportID, tt.rowSessionID, tt.rowTargetJobID, raw))
			if tt.expectMessages {
				mock.ExpectQuery(`(?s)select role, content, seq_no\s+from practice_messages\s+where session_id = \$1\s+order by seq_no asc`).
					WithArgs("session-1").
					WillReturnRows(tt.messages)
			}
			if _, err := NewRepository(db).LoadReportContext(context.Background(), "report-1"); err == nil {
				t.Fatal("LoadReportContext accepted mismatched frozen context")
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestLoadReportContextRejectsMissingFrozenSnapshotWithoutMutableFallback(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	mock.ExpectQuery(`(?s)select fr\.user_id::text, fr\.id::text, fr\.session_id::text, fr\.target_job_id::text, fr\.generation_context\s+from feedback_reports fr\s+where fr\.id = \$1`).
		WithArgs("report-1").
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "report_id", "session_id", "target_job_id", "generation_context"}).
			AddRow("user-1", "report-1", "session-1", "target-1", []byte(`{}`)))

	if _, err := NewRepository(db).LoadReportContext(context.Background(), "report-1"); err == nil {
		t.Fatal("LoadReportContext rebuilt a missing frozen snapshot")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func frozenReportContextSnapshot(t *testing.T) practicedomain.ReportContextSnapshot {
	t.Helper()
	snapshot, err := practicedomain.BuildReportContextSnapshot(practicedomain.ReportContextSnapshotInput{
		TargetJob: practicedomain.ReportTargetJobSnapshot{
			ID: "target-1", Title: "平台工程师", Company: "Acme", Language: "zh-CN", RawJD: "负责平台可靠性。",
			Summary: json.RawMessage(`{"interviewRounds":[{"sequence":1,"type":"technical","name":"技术面","focus":"系统设计","durationMinutes":45},{"sequence":2,"type":"manager","name":"管理面","focus":"协作与决策","durationMinutes":45}],"provenance":{"promptVersion":"v1","rubricVersion":"v1","modelId":"model-1","language":"zh-CN","dataSourceVersion":"target-summary.v1"}}`),
		},
		Resume: practicedomain.ReportResumeSnapshot{
			ID: "resume-1", DisplayName: "平台工程简历", Language: "zh-CN", SourceSnapshot: "完整简历正文",
			StructuredProfile: json.RawMessage(`{"skills":["Go"]}`),
		},
		Plan: practicedomain.ReportPlanSnapshot{
			ID: "plan-1", Goal: "baseline", InterviewerPersona: "hiring_manager", Difficulty: "medium", Language: "zh-CN",
			TimeBudgetMinutes: 45, ResumeID: "resume-1", RoundID: "round-1-technical", RoundSequence: 1,
			FocusDimensionCodes: []string{"system_design"},
		},
		Conversation: practicedomain.ReportConversationCoordinate{
			SessionID: "session-1", Language: "zh-CN", MessageCount: 3, LastMessageSeqNo: 3,
		},
	})
	if err != nil {
		t.Fatalf("BuildReportContextSnapshot: %v", err)
	}
	return snapshot
}
