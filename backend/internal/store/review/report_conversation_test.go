package review

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	reviewdomain "github.com/monshunter/easyinterview/backend/internal/review"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

func TestGetReportConversationProjectsOwnedOrderedMessagesForEveryReportStatus(t *testing.T) {
	now := time.Date(2026, 7, 15, 8, 0, 0, 0, time.UTC)
	for _, status := range []sharedtypes.ReportStatus{
		sharedtypes.ReportStatusQueued,
		sharedtypes.ReportStatusGenerating,
		sharedtypes.ReportStatusReady,
		sharedtypes.ReportStatusFailed,
	} {
		t.Run(string(status), func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()

			mock.ExpectQuery(reportConversationReportQueryPattern()).
				WithArgs("report-1", "user-1").
				WillReturnRows(reportConversationReportRow(t, status))
			mock.ExpectQuery(reportConversationMessagesQueryPattern()).
				WithArgs("session-1").
				WillReturnRows(sqlmock.NewRows([]string{"seq_no", "role", "content", "created_at"}).
					AddRow(1, "user", "我主导过一次迁移。", now).
					AddRow(2, "assistant", "请说明取舍和结果。", now.Add(12*time.Second)))

			got, err := NewRepository(db).GetReportConversation(context.Background(), "user-1", "report-1")
			if err != nil {
				t.Fatalf("GetReportConversation: %v", err)
			}
			if got.ReportID != "report-1" || got.Status != status || got.Context.SourcePlanID != "plan-1" {
				t.Fatalf("conversation identity/status/context=%+v", got)
			}
			wantMessages := []reviewdomain.ReportConversationMessageRecord{
				{Sequence: 1, Role: "user", Content: "我主导过一次迁移。", CreatedAt: now},
				{Sequence: 2, Role: "assistant", Content: "请说明取舍和结果。", CreatedAt: now.Add(12 * time.Second)},
			}
			if !reflect.DeepEqual(got.Messages, wantMessages) {
				t.Fatalf("messages=%#v want=%#v", got.Messages, wantMessages)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestGetReportConversationAllowsAnOwnedEmptyMessageProjection(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	mock.ExpectQuery(reportConversationReportQueryPattern()).
		WithArgs("report-1", "user-1").
		WillReturnRows(reportConversationReportRow(t, sharedtypes.ReportStatusReady))
	mock.ExpectQuery(reportConversationMessagesQueryPattern()).
		WithArgs("session-1").
		WillReturnRows(sqlmock.NewRows([]string{"seq_no", "role", "content", "created_at"}))

	got, err := NewRepository(db).GetReportConversation(context.Background(), "user-1", "report-1")
	if err != nil {
		t.Fatalf("GetReportConversation: %v", err)
	}
	if got.Messages == nil || len(got.Messages) != 0 {
		t.Fatalf("messages=%#v, want non-nil empty projection", got.Messages)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestGetReportConversationHidesMissingAndCrossUserReports(t *testing.T) {
	for _, userID := range []string{"user-1", "other-user"} {
		t.Run(userID, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			mock.ExpectQuery(reportConversationReportQueryPattern()).
				WithArgs("report-1", userID).
				WillReturnRows(sqlmock.NewRows(reportConversationReportColumns()))

			_, err = NewRepository(db).GetReportConversation(context.Background(), userID, "report-1")
			if !errors.Is(err, reviewdomain.ErrReportNotFound) {
				t.Fatalf("GetReportConversation error=%v, want ErrReportNotFound", err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestGetReportConversationFailsClosedForInvalidOwnershipOrMessageProjection(t *testing.T) {
	now := time.Date(2026, 7, 15, 8, 0, 0, 0, time.UTC)
	for _, tc := range []struct {
		name        string
		report      *sqlmock.Rows
		messages    *sqlmock.Rows
		expectReads bool
	}{
		{
			name: "empty report identity",
			report: sqlmock.NewRows(reportConversationReportColumns()).
				AddRow("", "session-1", "target-1", "ready", mustFrozenReportContextJSON(t), "session-1", "user-1", "target-1"),
		},
		{
			name: "report session binding",
			report: sqlmock.NewRows(reportConversationReportColumns()).
				AddRow("report-1", "session-1", "target-1", "ready", mustFrozenReportContextJSON(t), "session-2", "user-1", "target-1"),
		},
		{
			name: "session user binding",
			report: sqlmock.NewRows(reportConversationReportColumns()).
				AddRow("report-1", "session-1", "target-1", "ready", mustFrozenReportContextJSON(t), "session-1", "other-user", "target-1"),
		},
		{
			name: "target job binding",
			report: sqlmock.NewRows(reportConversationReportColumns()).
				AddRow("report-1", "session-1", "target-1", "ready", mustFrozenReportContextJSON(t), "session-1", "user-1", "target-2"),
		},
		{
			name:        "unknown message role",
			report:      reportConversationReportRow(t, sharedtypes.ReportStatusReady),
			messages:    sqlmock.NewRows([]string{"seq_no", "role", "content", "created_at"}).AddRow(1, "system", "private", now),
			expectReads: true,
		},
		{
			name:        "blank message content",
			report:      reportConversationReportRow(t, sharedtypes.ReportStatusReady),
			messages:    sqlmock.NewRows([]string{"seq_no", "role", "content", "created_at"}).AddRow(1, "user", " \n\t", now),
			expectReads: true,
		},
		{
			name:        "non increasing sequence",
			report:      reportConversationReportRow(t, sharedtypes.ReportStatusReady),
			messages:    sqlmock.NewRows([]string{"seq_no", "role", "content", "created_at"}).AddRow(2, "user", "first", now).AddRow(2, "assistant", "second", now),
			expectReads: true,
		},
		{
			name:        "missing message created at",
			report:      reportConversationReportRow(t, sharedtypes.ReportStatusReady),
			messages:    sqlmock.NewRows([]string{"seq_no", "role", "content", "created_at"}).AddRow(1, "user", "complete", time.Time{}),
			expectReads: true,
		},
		{
			name: "frozen session binding",
			report: sqlmock.NewRows(reportConversationReportColumns()).
				AddRow("report-1", "session-2", "target-1", "ready", mustReportConversationFrozenContext(t, "session-1"), "session-2", "user-1", "target-1"),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			mock.ExpectQuery(reportConversationReportQueryPattern()).
				WithArgs("report-1", "user-1").
				WillReturnRows(tc.report)
			if tc.expectReads {
				mock.ExpectQuery(reportConversationMessagesQueryPattern()).
					WithArgs("session-1").
					WillReturnRows(tc.messages)
			}

			_, err = NewRepository(db).GetReportConversation(context.Background(), "user-1", "report-1")
			if !errors.Is(err, reviewdomain.ErrReportConversationInvalid) {
				t.Fatalf("GetReportConversation error=%v, want ErrReportConversationInvalid", err)
			}
			if strings.Contains(err.Error(), "private") || strings.Contains(err.Error(), "first") || strings.Contains(err.Error(), "second") {
				t.Fatalf("invalid projection error leaked message body: %v", err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestReportConversationReadPathHasNoSideEffectsOrSessionFallback(t *testing.T) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve report conversation test path")
	}
	raw, err := os.ReadFile(filepath.Join(filepath.Dir(currentFile), "report_conversation.go"))
	if err != nil {
		t.Fatalf("read report conversation source: %v", err)
	}
	for _, forbidden := range []string{
		"ExecContext", "BeginTx", "aiclient", "audit_events", "outbox_events", "async_jobs",
		"GetPracticeSession", "ListPracticeSessions", "limit ", "offset ",
	} {
		if strings.Contains(string(raw), forbidden) {
			t.Fatalf("report conversation read path must not contain %q", forbidden)
		}
	}
}

func reportConversationReportQueryPattern() string {
	return `(?s)select fr\.id::text, fr\.session_id::text, fr\.target_job_id::text, fr\.status, fr\.generation_context,\s+ps\.id::text, ps\.user_id::text, ps\.target_job_id::text\s+from feedback_reports fr\s+join practice_sessions ps on ps\.id = fr\.session_id\s+where fr\.id = \$1 and fr\.user_id = \$2`
}

func reportConversationMessagesQueryPattern() string {
	return `(?s)select seq_no, role, content, created_at\s+from practice_messages\s+where session_id = \$1\s+order by seq_no asc`
}

func reportConversationReportColumns() []string {
	return []string{"report_id", "session_id", "target_job_id", "status", "generation_context", "bound_session_id", "session_user_id", "session_target_job_id"}
}

func reportConversationReportRow(t *testing.T, status sharedtypes.ReportStatus) *sqlmock.Rows {
	t.Helper()
	return sqlmock.NewRows(reportConversationReportColumns()).
		AddRow("report-1", "session-1", "target-1", string(status), mustFrozenReportContextJSON(t), "session-1", "user-1", "target-1")
}

func mustReportConversationFrozenContext(t *testing.T, sessionID string) []byte {
	t.Helper()
	snapshot := frozenReportContextSnapshot(t)
	snapshot.Conversation.SessionID = sessionID
	raw, err := json.Marshal(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}
