package review_test

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	reviewdomain "github.com/monshunter/easyinterview/backend/internal/review"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
	reviewstore "github.com/monshunter/easyinterview/backend/internal/store/review"
)

func TestGetFeedbackReportUserScopedWithAssessments(t *testing.T) {
	db, mock, cleanup := newMockReviewStore(t)
	defer cleanup()
	repo := reviewstore.NewRepository(db)
	now := time.Date(2026, 5, 15, 10, 20, 30, 0, time.UTC)

	reportRows := sqlmock.NewRows([]string{
		"id", "session_id", "target_job_id", "status", "preparedness_level",
		"highlights", "issues", "next_actions", "prompt_version", "rubric_version",
		"model_id", "provider", "language", "feature_flag", "data_source_version",
		"retry_focus_turn_ids", "error_code", "created_at", "updated_at",
	}).AddRow(
		"0197d120-0000-7000-8000-000000000501",
		"0197d120-0000-7000-8000-000000000502",
		"0197d120-0000-7000-8000-000000000503",
		"ready",
		"basically_ready",
		[]byte(`[{"dimension":"depth","evidence":"clear","confidence":"high"}]`),
		[]byte(`[]`),
		[]byte(`[{"type":"next_round","label":"Next round"}]`),
		"v0.1.0",
		"v0.1.0",
		"model-profile:report.generate.default",
		nil,
		"en",
		"none",
		"registry.v1",
		[]byte(`["0197d120-0000-7000-8000-000000000504"]`),
		nil,
		now,
		now,
	)
	mock.ExpectQuery(regexp.QuoteMeta(`
select fr.id, fr.session_id, fr.target_job_id, fr.status, fr.preparedness_level,
       fr.highlights, fr.issues, fr.next_actions, fr.prompt_version, fr.rubric_version,
       fr.model_id, fr.provider, fr.language, fr.feature_flag, fr.data_source_version,
       fr.retry_focus_turn_ids, fr.error_code, fr.created_at, fr.updated_at
from feedback_reports fr
where fr.id = $1 and fr.user_id = $2`)).
		WithArgs("0197d120-0000-7000-8000-000000000501", "user-1").
		WillReturnRows(reportRows)
	assessmentRows := sqlmock.NewRows([]string{"turn_id", "question_intent", "dimension_results", "review_status", "included_in_retry_plan"}).
		AddRow("0197d120-0000-7000-8000-000000000504", "architecture", []byte(`{"depth":{"status":"meets_bar","confidence":"high"}}`), "open", false)
	mock.ExpectQuery(regexp.QuoteMeta(`
select qa.turn_id, coalesce(qa.question_intent, ''), qa.dimension_results,
       qa.review_status, qa.included_in_retry_plan
from question_assessments qa
join practice_turns pt on pt.id = qa.turn_id
where qa.report_id = $1
order by pt.turn_index asc, qa.created_at asc`)).
		WithArgs("0197d120-0000-7000-8000-000000000501").
		WillReturnRows(assessmentRows)

	got, err := repo.GetFeedbackReport(context.Background(), "user-1", "0197d120-0000-7000-8000-000000000501")
	if err != nil {
		t.Fatalf("GetFeedbackReport: %v", err)
	}
	if got.Status != sharedtypes.ReportStatusReady || got.Provenance == nil || len(got.QuestionAssessments) != 1 {
		t.Fatalf("report = %+v", got)
	}
	if got.QuestionAssessments[0].DimensionResults["depth"].Status != sharedtypes.DimensionStatusMeetsBar {
		t.Fatalf("assessment = %+v", got.QuestionAssessments[0])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestGetFeedbackReportMapsNoRowsToReportNotFound(t *testing.T) {
	db, mock, cleanup := newMockReviewStore(t)
	defer cleanup()
	repo := reviewstore.NewRepository(db)

	mock.ExpectQuery("from feedback_reports fr").WithArgs("missing", "user-1").WillReturnRows(sqlmock.NewRows([]string{
		"id", "session_id", "target_job_id", "status", "preparedness_level",
		"highlights", "issues", "next_actions", "prompt_version", "rubric_version",
		"model_id", "provider", "language", "feature_flag", "data_source_version",
		"retry_focus_turn_ids", "error_code", "created_at", "updated_at",
	}))
	_, err := repo.GetFeedbackReport(context.Background(), "user-1", "missing")
	if !errors.Is(err, reviewdomain.ErrReportNotFound) {
		t.Fatalf("err = %v, want ErrReportNotFound", err)
	}
}

func TestListTargetJobReportsCursorPagination(t *testing.T) {
	db, mock, cleanup := newMockReviewStore(t)
	defer cleanup()
	repo := reviewstore.NewRepository(db)
	firstCreated := time.Date(2026, 5, 15, 10, 20, 30, 0, time.UTC)
	secondCreated := firstCreated.Add(-time.Minute)
	cursor := reviewdomain.EncodeCursor(firstCreated, "0197d120-0000-7000-8000-000000000501")

	rows := sqlmock.NewRows([]string{
		"id", "session_id", "target_job_id", "status", "preparedness_level",
		"highlights", "issues", "next_actions", "prompt_version", "rubric_version",
		"model_id", "provider", "language", "feature_flag", "data_source_version",
		"retry_focus_turn_ids", "error_code", "created_at", "updated_at",
	}).AddRow(
		"0197d120-0000-7000-8000-000000000502",
		"0197d120-0000-7000-8000-000000000602",
		"target-1",
		"generating",
		nil,
		[]byte(`[]`),
		[]byte(`[]`),
		[]byte(`[]`),
		nil,
		nil,
		nil,
		nil,
		"en",
		"none",
		"not_applicable",
		[]byte(`[]`),
		nil,
		secondCreated,
		secondCreated,
	).AddRow(
		"0197d120-0000-7000-8000-000000000503",
		"0197d120-0000-7000-8000-000000000603",
		"target-1",
		"queued",
		nil,
		[]byte(`[]`),
		[]byte(`[]`),
		[]byte(`[]`),
		nil,
		nil,
		nil,
		nil,
		"en",
		"none",
		"not_applicable",
		[]byte(`[]`),
		nil,
		secondCreated.Add(-time.Minute),
		secondCreated.Add(-time.Minute),
	)
	mock.ExpectQuery("from target_jobs").
		WithArgs("target-1", "user-1").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectQuery("from feedback_reports fr").
		WithArgs("user-1", "target-1", firstCreated, "0197d120-0000-7000-8000-000000000501", 2).
		WillReturnRows(rows)
	mock.ExpectQuery("from question_assessments qa").WithArgs("0197d120-0000-7000-8000-000000000502").WillReturnRows(sqlmock.NewRows([]string{"turn_id", "question_intent", "dimension_results", "review_status", "included_in_retry_plan"}))

	got, err := repo.ListTargetJobReports(context.Background(), reviewdomain.ListTargetJobReportsInput{
		UserID:      "user-1",
		TargetJobID: "target-1",
		Cursor:      cursor,
		PageSize:    1,
	})
	if err != nil {
		t.Fatalf("ListTargetJobReports: %v", err)
	}
	if len(got.Items) != 1 || !got.HasMore || got.NextCursor == "" || got.PageSize != 1 {
		t.Fatalf("list result = %+v", got)
	}
}

func TestListTargetJobReportsRequiresOwnedTarget(t *testing.T) {
	db, mock, cleanup := newMockReviewStore(t)
	defer cleanup()
	repo := reviewstore.NewRepository(db)

	mock.ExpectQuery("from target_jobs").
		WithArgs("target-cross-user", "user-1").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	_, err := repo.ListTargetJobReports(context.Background(), reviewdomain.ListTargetJobReportsInput{
		UserID:      "user-1",
		TargetJobID: "target-cross-user",
		PageSize:    20,
	})
	if !errors.Is(err, reviewdomain.ErrReportNotFound) {
		t.Fatalf("err = %v, want ErrReportNotFound", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
