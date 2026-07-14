package reports

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	api "github.com/monshunter/easyinterview/backend/internal/api/generated"
	reviewdomain "github.com/monshunter/easyinterview/backend/internal/review"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

func TestReportAPIProjectsReadyAndFrozenContextWithoutInternalAnchors(t *testing.T) {
	frozen := reportEvidenceFrozenProjection()
	wantContext := api.ReportContextSnapshot{
		SourcePlanId: frozen.SourcePlanID, TargetJobTitle: frozen.TargetJobTitle, TargetJobCompany: frozen.TargetJobCompany,
		ResumeId: frozen.ResumeID, ResumeDisplayName: frozen.ResumeDisplayName,
		RoundId: frozen.RoundID, RoundSequence: frozen.RoundSequence, RoundName: frozen.RoundName, RoundType: frozen.RoundType,
		Language: frozen.Language, HasNextRound: frozen.HasNextRound,
	}
	for _, status := range []sharedtypes.ReportStatus{
		sharedtypes.ReportStatusQueued,
		sharedtypes.ReportStatusGenerating,
		sharedtypes.ReportStatusReady,
		sharedtypes.ReportStatusFailed,
	} {
		report := reviewdomain.FeedbackReportRecord{
			ID: "report-redacted", SessionID: "session-redacted", TargetJobID: "target-redacted", Status: status,
			Context: frozen, CreatedAt: time.Unix(1, 0).UTC(), UpdatedAt: time.Unix(2, 0).UTC(),
		}
		if got := toAPIFeedbackReport(report); !reflect.DeepEqual(got.Context, wantContext) {
			t.Fatalf("%s API projection did not preserve frozen context", status)
		}
	}

	summary := "The answer explained tradeoffs, but rollback evidence needs concrete steps."
	preparedness := sharedtypes.ReadinessTierNeedsPractice
	record := reviewdomain.FeedbackReportRecord{
		ID: "report-redacted", SessionID: "session-redacted", TargetJobID: "target-redacted", Status: sharedtypes.ReportStatusReady,
		Summary: &summary, Context: frozen, PreparednessLevel: &preparedness,
		DimensionAssessments:     []reviewdomain.DimensionAssessmentRecord{{Code: "technical_depth", Label: "Technical depth", Status: sharedtypes.DimensionStatusNeedsWork, Confidence: sharedtypes.ConfidenceHigh}},
		Highlights:               []reviewdomain.ReportEvidenceRecord{},
		Issues:                   []reviewdomain.ReportEvidenceRecord{{DimensionCode: "technical_depth", Evidence: "Rollback steps were not concrete.", Confidence: sharedtypes.ConfidenceHigh, SourceMessageSeqNos: []int32{2}}},
		NextActions:              []reviewdomain.ReportNextActionRecord{{Type: "retry_current_round", Label: "Add rollback steps and replay this round"}},
		RetryFocusDimensionCodes: []string{"technical_depth"},
		Provenance: &reviewdomain.GenerationProvenanceRecord{
			PromptVersion: "v0.2.0", RubricVersion: "v0.2.0", ModelID: "evidence-model", Language: "en", FeatureFlag: "none", DataSourceVersion: "report-context.v1",
		},
		CreatedAt: time.Unix(1, 0).UTC(), UpdatedAt: time.Unix(2, 0).UTC(),
	}
	got := toAPIFeedbackReport(record)
	if got.Summary == nil || *got.Summary != summary || got.PreparednessLevel == nil || string(*got.PreparednessLevel) != string(preparedness) ||
		len(got.DimensionAssessments) != 1 || got.DimensionAssessments[0].Code != "technical_depth" || got.DimensionAssessments[0].Label != "Technical depth" ||
		len(got.Issues) != 1 || got.Issues[0].DimensionCode != "technical_depth" || len(got.NextActions) != 1 ||
		!reflect.DeepEqual(got.RetryFocusDimensionCodes, []string{"technical_depth"}) || got.Provenance == nil || got.Provenance.PromptVersion != "v0.2.0" || got.Provenance.DataSourceVersion != "report-context.v1" {
		t.Fatal("ready API direct fields drifted")
	}
	raw, err := json.Marshal(got)
	if err != nil {
		t.Fatal(err)
	}
	if containsJSONKey(raw, "sourceMessageSeqNos") || bytes.Contains(raw, []byte(`"sourceMessageSeqNos"`)) {
		t.Fatal("internal grounding anchors leaked into the report API")
	}

	handler := NewHandler(HandlerOptions{
		Service: projectionReportService{err: reviewdomain.ErrReportNotFound},
		Session: func(context.Context) (string, bool) { return "other-user-redacted", true },
	})
	recorder := httptest.NewRecorder()
	handler.GetFeedbackReport(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/reports/redacted", nil), "report-redacted")
	assertReportNotFoundResponse(t, recorder)

	t.Log("REPORT_DIRECT_READY_PASS")
	t.Log("REPORT_FROZEN_CONTEXT_READ_PASS")
}

func TestReportAPIProjectsTerminalFailuresWithoutPartialReadyData(t *testing.T) {
	frozen := reportEvidenceFrozenProjection()
	for _, code := range []string{sharederrors.CodeAiOutputInvalid, sharederrors.CodeReportContextTooLarge} {
		code := code
		t.Run(code, func(t *testing.T) {
			record := reviewdomain.FeedbackReportRecord{
				ID: "report-redacted", SessionID: "session-redacted", TargetJobID: "target-redacted",
				Status: sharedtypes.ReportStatusFailed, ErrorCode: &code, Context: frozen,
				CreatedAt: time.Unix(1, 0).UTC(), UpdatedAt: time.Unix(2, 0).UTC(),
			}
			got := toAPIFeedbackReport(record)
			if got.Status != sharedtypes.ReportStatusFailed || got.ErrorCode == nil || string(*got.ErrorCode) != code || got.Summary != nil || got.PreparednessLevel != nil || got.Provenance != nil ||
				len(got.DimensionAssessments) != 0 || len(got.Highlights) != 0 || len(got.Issues) != 0 || len(got.NextActions) != 0 || len(got.RetryFocusDimensionCodes) != 0 {
				t.Fatal("terminal report API exposed partial ready/retry data")
			}
			if !reflect.DeepEqual(got.Context, toAPIReportContext(frozen)) {
				t.Fatal("terminal report lost frozen context")
			}
			raw, err := json.Marshal(got)
			if err != nil {
				t.Fatal(err)
			}
			if containsJSONKey(raw, "sourceMessageSeqNos") {
				t.Fatal("terminal report API leaked internal anchors")
			}
			meta, ok := sharederrors.CodeRegistry[code]
			if !ok || meta.Retryable {
				t.Fatalf("terminal report code %s is not locked non-retryable", code)
			}
		})
	}

	handler := NewHandler(HandlerOptions{
		Service: projectionReportService{err: reviewdomain.ErrReportNotFound},
		Session: func(context.Context) (string, bool) { return "owned-user-redacted", true },
	})
	recorder := httptest.NewRecorder()
	handler.GetFeedbackReport(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/reports/redacted", nil), "missing-report-redacted")
	assertReportNotFoundResponse(t, recorder)

	t.Log("REPORT_CONTEXT_TOO_LARGE_PASS")
	t.Log("REPORT_FOUR_INVALID_FAIL_CLOSED_PASS")
}

func reportEvidenceFrozenProjection() reviewdomain.ReportContextProjection {
	return reviewdomain.ReportContextProjection{
		SourcePlanID: "plan-redacted", TargetJobTitle: "Frozen Platform Engineer", TargetJobCompany: "Frozen Company",
		ResumeID: "resume-redacted", ResumeDisplayName: "Frozen resume",
		RoundID: "round-1-technical", RoundSequence: 1, RoundName: "Technical", RoundType: "technical",
		Language: "en", HasNextRound: true,
	}
}

func assertReportNotFoundResponse(t *testing.T, recorder *httptest.ResponseRecorder) {
	t.Helper()
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("inaccessible report status=%d, want 404", recorder.Code)
	}
	var response api.ApiErrorResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if response.Error.Code != sharederrors.CodeReportNotFound || response.Error.Retryable {
		t.Fatalf("inaccessible report response=%+v", response.Error)
	}
}
