package review

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	reviewdomain "github.com/monshunter/easyinterview/backend/internal/review"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

func TestGetFeedbackReportFrozenProjectionForEveryStatus(t *testing.T) {
	now := time.Date(2026, 7, 12, 8, 30, 0, 0, time.UTC)
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
			snapshotRaw := mustFrozenReportContextJSON(t)
			row := feedbackReportProjectionRow(status, snapshotRaw, now)
			mock.ExpectQuery(feedbackReportProjectionQueryPattern()).
				WithArgs("report-1", "user-1").
				WillReturnRows(row)

			got, err := NewRepository(db).GetFeedbackReport(context.Background(), "user-1", "report-1")
			if err != nil {
				t.Fatalf("GetFeedbackReport: %v", err)
			}
			wantContext := reviewdomain.ReportContextProjection{
				SourcePlanID: "plan-1", TargetJobTitle: "平台工程师", TargetJobCompany: "Acme",
				ResumeID: "resume-1", ResumeDisplayName: "平台工程简历",
				RoundID: "round-1-technical", RoundSequence: 1, RoundName: "技术面", RoundType: "technical",
				Language: "zh-CN", HasNextRound: true,
			}
			if !reflect.DeepEqual(got.Context, wantContext) {
				t.Fatalf("frozen public context mismatch:\n got: %#v\nwant: %#v", got.Context, wantContext)
			}
			if got.Status != status || got.SessionID != "session-1" || got.TargetJobID != "target-1" {
				t.Fatalf("report identity/status mismatch: %#v", got)
			}
			if status != sharedtypes.ReportStatusReady {
				if got.Summary != nil || got.PreparednessLevel != nil || got.Provenance != nil || len(got.DimensionAssessments) != 0 || len(got.Highlights) != 0 || len(got.Issues) != 0 || len(got.NextActions) != 0 || len(got.RetryFocusDimensionCodes) != 0 {
					t.Fatalf("non-ready report exposed ready-only fields: %#v", got)
				}
			} else {
				assertReadyProjectionLossless(t, got)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestGetFeedbackReportFrozenProjectionRejectsCrossUserAsNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	mock.ExpectQuery(feedbackReportProjectionQueryPattern()).
		WithArgs("report-1", "other-user").
		WillReturnRows(sqlmock.NewRows(feedbackReportProjectionColumns()))

	_, err = NewRepository(db).GetFeedbackReport(context.Background(), "other-user", "report-1")
	if !errors.Is(err, reviewdomain.ErrReportNotFound) {
		t.Fatalf("cross-user read error = %v, want ErrReportNotFound", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func assertReadyProjectionLossless(t *testing.T, got reviewdomain.FeedbackReportRecord) {
	t.Helper()
	if got.Summary == nil || *got.Summary != "回答结构清楚，技术取舍需要量化证据。" {
		t.Fatalf("summary lost: %#v", got.Summary)
	}
	if got.PreparednessLevel == nil || *got.PreparednessLevel != sharedtypes.ReadinessTierNeedsPractice {
		t.Fatalf("preparedness lost: %#v", got.PreparednessLevel)
	}
	if len(got.DimensionAssessments) != 1 || got.DimensionAssessments[0].Code != "technical_tradeoffs" || got.DimensionAssessments[0].Label != "技术取舍" {
		t.Fatalf("dimension assessment lost: %#v", got.DimensionAssessments)
	}
	if len(got.Highlights) != 1 || got.Highlights[0].DimensionCode != "structured_communication" || !reflect.DeepEqual(got.Highlights[0].SourceMessageSeqNos, []int32{2}) {
		t.Fatalf("highlight or internal anchors lost: %#v", got.Highlights)
	}
	if len(got.Issues) != 1 || got.Issues[0].DimensionCode != "technical_tradeoffs" || !reflect.DeepEqual(got.Issues[0].SourceMessageSeqNos, []int32{2}) {
		t.Fatalf("issue or internal anchors lost: %#v", got.Issues)
	}
	if len(got.NextActions) != 1 || got.NextActions[0].Type != "retry_current_round" || len(got.RetryFocusDimensionCodes) != 1 || got.RetryFocusDimensionCodes[0] != "technical_tradeoffs" {
		t.Fatalf("actions/focus lost: %#v / %#v", got.NextActions, got.RetryFocusDimensionCodes)
	}
	if got.Provenance == nil || got.Provenance.PromptVersion != "v0.2.0" || got.Provenance.DataSourceVersion != "report-context.v1" {
		t.Fatalf("provenance lost: %#v", got.Provenance)
	}
}

func feedbackReportProjectionQueryPattern() string {
	return `(?s)select fr\.id::text, fr\.session_id::text, fr\.target_job_id::text, fr\.status, fr\.error_code, fr\.summary, fr\.generation_context,\s+fr\.preparedness_level, fr\.dimension_assessments, fr\.highlights, fr\.issues, fr\.next_actions,\s+fr\.retry_focus_dimension_codes, fr\.prompt_version, fr\.rubric_version, fr\.model_id,\s+fr\.language, fr\.feature_flag, fr\.data_source_version, fr\.created_at, fr\.updated_at\s+from feedback_reports fr\s+where fr\.id = \$1 and fr\.user_id = \$2`
}

func feedbackReportProjectionColumns() []string {
	return []string{
		"id", "session_id", "target_job_id", "status", "error_code", "summary", "generation_context",
		"preparedness_level", "dimension_assessments", "highlights", "issues", "next_actions",
		"retry_focus_dimension_codes", "prompt_version", "rubric_version", "model_id",
		"language", "feature_flag", "data_source_version", "created_at", "updated_at",
	}
}

func feedbackReportProjectionRow(status sharedtypes.ReportStatus, snapshotRaw []byte, now time.Time) *sqlmock.Rows {
	row := sqlmock.NewRows(feedbackReportProjectionColumns())
	if status != sharedtypes.ReportStatusReady {
		var errorCode any
		if status == sharedtypes.ReportStatusFailed {
			errorCode = "AI_OUTPUT_INVALID"
		}
		return row.AddRow(
			"report-1", "session-1", "target-1", string(status), errorCode, nil, snapshotRaw,
			nil, []byte(`[]`), []byte(`[]`), []byte(`[]`), []byte(`[]`), `{}`,
			nil, nil, nil, "zh-CN", "none", "report-context.v1", now, now,
		)
	}
	return row.AddRow(
		"report-1", "session-1", "target-1", string(status), nil,
		"回答结构清楚，技术取舍需要量化证据。", snapshotRaw,
		string(sharedtypes.ReadinessTierNeedsPractice),
		[]byte(`[{"code":"technical_tradeoffs","label":"技术取舍","status":"needs_work","confidence":"high"}]`),
		[]byte(`[{"dimensionCode":"structured_communication","evidence":"回答按背景、行动和结果展开。","confidence":"high","sourceMessageSeqNos":[2]}]`),
		[]byte(`[{"dimensionCode":"technical_tradeoffs","evidence":"没有比较备选方案。","confidence":"medium","sourceMessageSeqNos":[2]}]`),
		[]byte(`[{"type":"retry_current_round","label":"补齐技术取舍证据。"}]`),
		`{"technical_tradeoffs"}`,
		"v0.2.0", "v0.2.0", "model-profile:report.generate.default", "zh-CN", "none", "report-context.v1", now, now.Add(time.Minute),
	)
}

func mustFrozenReportContextJSON(t *testing.T) []byte {
	t.Helper()
	raw, err := json.Marshal(frozenReportContextSnapshot(t))
	if err != nil {
		t.Fatal(err)
	}
	return raw
}
