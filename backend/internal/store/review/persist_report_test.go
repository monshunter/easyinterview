package review

import (
	"context"
	"database/sql/driver"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	reviewdomain "github.com/monshunter/easyinterview/backend/internal/review"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

type postgresTextArray string

func (want postgresTextArray) Match(value driver.Value) bool {
	switch typed := value.(type) {
	case string:
		return typed == string(want)
	case []byte:
		return string(typed) == string(want)
	default:
		return false
	}
}

func TestPersistReportUsesPostgresTextArrayForRetryFocus(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	now := time.Unix(4, 0).UTC()
	in := reviewdomain.ReportResultPersistence{
		UserID: "user-1", ReportID: "report-1", SessionID: "session-1", TargetJobID: "target-1",
		AsyncJobID: "job-1", OutboxEventID: "outbox-1", AuditEventID: "audit-1",
		PreparednessLevel: sharedtypes.ReadinessTierBasicallyReady,
		PromptVersion:     "v1", RubricVersion: "v1", ModelID: "model-1", Provider: "provider-1",
		Language: "en", FeatureFlag: "none", DataSourceVersion: "registry.v1",
		RetryFocusCompetencyCodes: []string{"technical_depth"}, Now: now,
		Content: reviewdomain.ReportContentDraft{
			Highlights:  []reviewdomain.ReportEvidenceDraft{},
			Issues:      []reviewdomain.ReportEvidenceDraft{},
			NextActions: []reviewdomain.ReportNextActionDraft{{Type: "retry_current_round", Label: "Practice rollout tradeoffs"}},
		},
		DimensionAssessments: []reviewdomain.DimensionAssessmentDraft{},
	}

	mock.ExpectBegin()
	mock.ExpectExec(`update feedback_reports`).
		WithArgs(
			string(in.PreparednessLevel), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			in.PromptVersion, in.RubricVersion, in.ModelID, in.Provider, in.Language,
			in.FeatureFlag, in.DataSourceVersion, postgresTextArray(`{"technical_depth"}`),
			sqlmock.AnyArg(), in.Now, in.ReportID,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`insert into outbox_events`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`insert into audit_events`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`update async_jobs`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	if err := NewRepository(db).PersistReport(context.Background(), in); err != nil {
		t.Fatalf("PersistReport: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestUpdateFeedbackReportStatusAllowsGeneratingRetry(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	now := time.Unix(5, 0).UTC()
	mock.ExpectExec(`where id = \$3 and status in \(\$4, \$1\)`).
		WithArgs(string(sharedtypes.ReportStatusGenerating), now, "report-1", string(sharedtypes.ReportStatusQueued)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = NewRepository(db).UpdateFeedbackReportStatus(context.Background(), reviewdomain.ReportStatusUpdate{
		ReportID: "report-1", From: sharedtypes.ReportStatusQueued, To: sharedtypes.ReportStatusGenerating, Now: now,
	})
	if err != nil {
		t.Fatalf("UpdateFeedbackReportStatus: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
