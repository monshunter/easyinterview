package review

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	reviewdomain "github.com/monshunter/easyinterview/backend/internal/review"
	"github.com/monshunter/easyinterview/backend/internal/runner"
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

func validReportPersistence(now time.Time) reviewdomain.ReportResultPersistence {
	return reviewdomain.ReportResultPersistence{
		UserID: "user-1", ReportID: "report-1", SessionID: "session-1", TargetJobID: "target-1",
		AsyncJobID: "job-1", ClaimedAttempts: 2, OutboxEventID: "outbox-1", AuditEventID: "audit-1",
		PromptVersion: "v1", RubricVersion: "v1", ModelID: "model-1", Provider: "provider-1",
		Language: "en", FeatureFlag: "none", DataSourceVersion: "registry.v1",
		Now: now,
		Content: reviewdomain.ReportContentDraft{
			Summary: "Direct grounded summary", PreparednessLevel: sharedtypes.ReadinessTierBasicallyReady,
			DimensionAssessments:     []reviewdomain.DimensionAssessmentDraft{{Code: "technical_depth", Label: "Technical depth", Status: sharedtypes.DimensionStatusMeetsBar, Confidence: sharedtypes.ConfidenceHigh}},
			Highlights:               []reviewdomain.ReportEvidenceDraft{{DimensionCode: "technical_depth", Evidence: "Explained the tradeoff", Confidence: sharedtypes.ConfidenceHigh, SourceMessageSeqNos: []int32{2}}},
			Issues:                   []reviewdomain.ReportEvidenceDraft{},
			NextActions:              []reviewdomain.ReportNextActionDraft{{Type: "retry_current_round", Label: "Practice rollout tradeoffs"}},
			RetryFocusDimensionCodes: []string{"technical_depth"},
		},
	}
}

func TestPersistReportUsesPostgresTextArrayForRetryFocus(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	now := time.Unix(4, 0).UTC()
	in := validReportPersistence(now)

	mock.ExpectBegin()
	mock.ExpectQuery(`(?s)select attempts.*from async_jobs.*where id = \$1.*status = 'running'.*attempts = \$2.*for update`).
		WithArgs(in.AsyncJobID, in.ClaimedAttempts).
		WillReturnRows(sqlmock.NewRows([]string{"attempts"}).AddRow(in.ClaimedAttempts))
	mock.ExpectExec(`update feedback_reports`).
		WithArgs(
			in.Content.Summary, string(in.Content.PreparednessLevel), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			in.PromptVersion, in.RubricVersion, in.ModelID, in.Provider, in.Language,
			in.FeatureFlag, in.DataSourceVersion, postgresTextArray(`{"technical_depth"}`),
			sqlmock.AnyArg(), in.Now, in.ReportID,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`insert into outbox_events`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`insert into audit_events`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`(?s)update async_jobs.*where id = \$2.*status = 'running'.*attempts = \$3`).
		WithArgs(in.Now, in.AsyncJobID, in.ClaimedAttempts).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	if err := NewRepository(db).PersistReport(context.Background(), in); err != nil {
		t.Fatalf("PersistReport: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPersistReportStaleLeaseReturnsBeforeBusinessWrites(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	in := validReportPersistence(time.Unix(4, 0).UTC())

	mock.ExpectBegin()
	mock.ExpectQuery(`(?s)select attempts.*from async_jobs.*for update`).
		WithArgs(in.AsyncJobID, in.ClaimedAttempts).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectRollback()

	err = NewRepository(db).PersistReport(context.Background(), in)
	if !errors.Is(err, runner.ErrStaleLease) {
		t.Fatalf("PersistReport err=%v want ErrStaleLease", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("stale lease reached a report/outbox/audit/job write: %v", err)
	}
}

func TestPersistReportRejectsMissingProvenanceInsteadOfWritingFallbacks(t *testing.T) {
	tests := []struct {
		name   string
		field  string
		mutate func(*reviewdomain.ReportResultPersistence)
	}{
		{name: "prompt version", field: "prompt_version", mutate: func(in *reviewdomain.ReportResultPersistence) { in.PromptVersion = "" }},
		{name: "rubric version", field: "rubric_version", mutate: func(in *reviewdomain.ReportResultPersistence) { in.RubricVersion = "" }},
		{name: "model", field: "model_id", mutate: func(in *reviewdomain.ReportResultPersistence) { in.ModelID = "" }},
		{name: "provider", field: "provider", mutate: func(in *reviewdomain.ReportResultPersistence) { in.Provider = "" }},
		{name: "language", field: "language", mutate: func(in *reviewdomain.ReportResultPersistence) { in.Language = "" }},
		{name: "feature flag", field: "feature_flag", mutate: func(in *reviewdomain.ReportResultPersistence) { in.FeatureFlag = "" }},
		{name: "data source", field: "data_source_version", mutate: func(in *reviewdomain.ReportResultPersistence) { in.DataSourceVersion = "" }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, _, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			in := validReportPersistence(time.Unix(4, 0).UTC())
			tt.mutate(&in)

			err = NewRepository(db).PersistReport(context.Background(), in)
			if err == nil || !strings.Contains(err.Error(), tt.field) {
				t.Fatalf("expected missing %s error before DB write, got %v", tt.field, err)
			}
		})
	}
}

func TestUpdateFeedbackReportStatusAllowsGeneratingRetry(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	now := time.Unix(5, 0).UTC()
	mock.ExpectBegin()
	mock.ExpectQuery(`(?s)select attempts.*from async_jobs.*for update`).
		WithArgs("job-1", int32(2)).
		WillReturnRows(sqlmock.NewRows([]string{"attempts"}).AddRow(2))
	mock.ExpectExec(`where id = \$3 and status in \(\$4, \$1\)`).
		WithArgs(string(sharedtypes.ReportStatusGenerating), now, "report-1", string(sharedtypes.ReportStatusQueued)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err = NewRepository(db).UpdateFeedbackReportStatus(context.Background(), reviewdomain.ReportStatusUpdate{
		ReportID: "report-1", AsyncJobID: "job-1", ClaimedAttempts: 2,
		From: sharedtypes.ReportStatusQueued, To: sharedtypes.ReportStatusGenerating, Now: now,
	})
	if err != nil {
		t.Fatalf("UpdateFeedbackReportStatus: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestUpdateFeedbackReportStatusStaleLeaseReturnsBeforeReportWrite(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	now := time.Unix(5, 0).UTC()
	mock.ExpectBegin()
	mock.ExpectQuery(`(?s)select attempts.*from async_jobs.*for update`).
		WithArgs("job-1", int32(1)).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectRollback()

	err = NewRepository(db).UpdateFeedbackReportStatus(context.Background(), reviewdomain.ReportStatusUpdate{
		ReportID: "report-1", AsyncJobID: "job-1", ClaimedAttempts: 1,
		From: sharedtypes.ReportStatusQueued, To: sharedtypes.ReportStatusGenerating, Now: now,
	})
	if !errors.Is(err, runner.ErrStaleLease) {
		t.Fatalf("UpdateFeedbackReportStatus err=%v want ErrStaleLease", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("stale status transition wrote feedback_report: %v", err)
	}
}
