package reports

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	api "github.com/monshunter/easyinterview/backend/internal/api/generated"
	reviewdomain "github.com/monshunter/easyinterview/backend/internal/review"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

type projectionReportService struct {
	report reviewdomain.FeedbackReportRecord
	err    error
}

func (s projectionReportService) GetFeedbackReport(context.Context, string, string) (reviewdomain.FeedbackReportRecord, error) {
	return s.report, s.err
}

func (s projectionReportService) GetReportConversation(context.Context, string, string) (reviewdomain.ReportConversationRecord, error) {
	return reviewdomain.ReportConversationRecord{}, s.err
}

func (s projectionReportService) ListTargetJobReports(context.Context, reviewdomain.ListTargetJobReportsRequest) (reviewdomain.TargetJobReportsOverviewRecord, error) {
	return reviewdomain.TargetJobReportsOverviewRecord{}, s.err
}

func TestFeedbackReportProjectionUsesExactFrozenContextForEveryStatus(t *testing.T) {
	contextProjection := reviewdomain.ReportContextProjection{
		SourcePlanID: "plan-1", TargetJobTitle: "平台工程师", TargetJobCompany: "Acme",
		ResumeID: "resume-1", ResumeDisplayName: "平台工程简历",
		RoundID: "round-1-technical", RoundSequence: 1, RoundName: "技术面", RoundType: "technical",
		Language: "zh-CN", HasNextRound: true,
	}
	wantContext := api.ReportContextSnapshot{
		SourcePlanId: "plan-1", TargetJobTitle: "平台工程师", TargetJobCompany: "Acme",
		ResumeId: "resume-1", ResumeDisplayName: "平台工程简历",
		RoundId: "round-1-technical", RoundSequence: 1, RoundName: "技术面", RoundType: "technical",
		Language: "zh-CN", HasNextRound: true,
	}
	for _, status := range []sharedtypes.ReportStatus{
		sharedtypes.ReportStatusQueued,
		sharedtypes.ReportStatusGenerating,
		sharedtypes.ReportStatusReady,
		sharedtypes.ReportStatusFailed,
	} {
		t.Run(string(status), func(t *testing.T) {
			report := reviewdomain.FeedbackReportRecord{
				ID: "report-1", SessionID: "session-1", TargetJobID: "target-1", Status: status,
				Context: contextProjection, CreatedAt: time.Unix(1, 0).UTC(), UpdatedAt: time.Unix(2, 0).UTC(),
			}
			got := toAPIFeedbackReport(report)
			if got.RetryFocusDimensionCodes == nil {
				t.Fatal("API retryFocusDimensionCodes must encode an empty focus as [] instead of null")
			}
			if !reflect.DeepEqual(got.Context, wantContext) {
				t.Fatalf("API context mismatch:\n got: %#v\nwant: %#v", got.Context, wantContext)
			}
			raw, err := json.Marshal(got.Context)
			if err != nil {
				t.Fatal(err)
			}
			var fields map[string]any
			if err := json.Unmarshal(raw, &fields); err != nil {
				t.Fatal(err)
			}
			if len(fields) != 11 {
				t.Fatalf("API context has non-minimal fields: %s", raw)
			}
		})
	}
}

func TestFeedbackReportProjectionPreservesDirectReadyFieldsAndStripsInternalAnchors(t *testing.T) {
	summary := "回答结构清楚，技术取舍需要量化证据。"
	preparedness := sharedtypes.ReadinessTierNeedsPractice
	report := reviewdomain.FeedbackReportRecord{
		ID: "report-1", SessionID: "session-1", TargetJobID: "target-1", Status: sharedtypes.ReportStatusReady,
		Summary: &summary,
		Context: reviewdomain.ReportContextProjection{
			SourcePlanID: "plan-1", TargetJobTitle: "平台工程师", ResumeID: "resume-1", ResumeDisplayName: "简历",
			RoundID: "round-1-technical", RoundSequence: 1, RoundName: "技术面", RoundType: "technical", Language: "zh-CN", HasNextRound: true,
		},
		PreparednessLevel:        &preparedness,
		DimensionAssessments:     []reviewdomain.DimensionAssessmentRecord{{Code: "d1", Label: "技术取舍", Status: sharedtypes.DimensionStatusNeedsWork, Confidence: sharedtypes.ConfidenceHigh}},
		Highlights:               []reviewdomain.ReportEvidenceRecord{{DimensionCode: "d2", Evidence: "结构清楚", Confidence: sharedtypes.ConfidenceHigh, SourceMessageSeqNos: []int32{2}}},
		Issues:                   []reviewdomain.ReportEvidenceRecord{{DimensionCode: "d1", Evidence: "缺量化证据", Confidence: sharedtypes.ConfidenceMedium, SourceMessageSeqNos: []int32{2}}},
		NextActions:              []reviewdomain.ReportNextActionRecord{{Type: "retry_current_round", Label: "补齐证据"}},
		RetryFocusDimensionCodes: []string{"d1"},
		Provenance:               &reviewdomain.GenerationProvenanceRecord{PromptVersion: "v0.2.0", RubricVersion: "v0.2.0", ModelID: "model-1", Language: "zh-CN", FeatureFlag: "none", DataSourceVersion: "report-context.v1"},
		CreatedAt:                time.Unix(1, 0).UTC(), UpdatedAt: time.Unix(2, 0).UTC(),
	}

	got := toAPIFeedbackReport(report)
	if got.Summary == nil || *got.Summary != summary || len(got.DimensionAssessments) != 1 || got.DimensionAssessments[0].Code != "d1" || got.DimensionAssessments[0].Label != "技术取舍" || len(got.Highlights) != 1 || got.Highlights[0].DimensionCode != "d2" || len(got.RetryFocusDimensionCodes) != 1 || got.RetryFocusDimensionCodes[0] != "d1" {
		t.Fatalf("direct ready projection lost fields: %#v", got)
	}
	raw, err := json.Marshal(got)
	if err != nil {
		t.Fatal(err)
	}
	if json.Valid(raw) && containsJSONKey(raw, "sourceMessageSeqNos") {
		t.Fatalf("internal anchors leaked into API: %s", raw)
	}
}

func TestFeedbackReportProjectionKeepsCrossUserReadAsNotFound(t *testing.T) {
	handler := NewHandler(HandlerOptions{
		Service: projectionReportService{err: reviewdomain.ErrReportNotFound},
		Session: func(context.Context) (string, bool) { return "other-user", true },
	})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/reports/report-1", nil)

	handler.GetFeedbackReport(recorder, request, "report-1")
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("cross-user status = %d, want 404; body=%s", recorder.Code, recorder.Body.String())
	}
	assertPrivateReportHeaders(t, recorder)
	var response api.ApiErrorResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if response.Error.Code != "REPORT_NOT_FOUND" {
		t.Fatalf("cross-user code = %q, want REPORT_NOT_FOUND", response.Error.Code)
	}
}

func TestFeedbackReportHTTPResponsesDisablePrivateCaching(t *testing.T) {
	report := reviewdomain.FeedbackReportRecord{
		ID: "report-1", SessionID: "session-1", TargetJobID: "target-1", Status: sharedtypes.ReportStatusGenerating,
		Context: reviewdomain.ReportContextProjection{
			SourcePlanID: "plan-1", TargetJobTitle: "Platform Engineer", ResumeID: "resume-1", ResumeDisplayName: "Resume",
			RoundID: "round-1-technical", RoundSequence: 1, RoundName: "Technical", RoundType: "technical", Language: "en",
		},
		CreatedAt: time.Unix(1, 0).UTC(), UpdatedAt: time.Unix(2, 0).UTC(),
	}
	handler := NewHandler(HandlerOptions{
		Service: projectionReportService{report: report},
		Session: func(context.Context) (string, bool) { return "user-1", true },
	})
	recorder := httptest.NewRecorder()
	handler.GetFeedbackReport(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/reports/report-1", nil), "report-1")
	if recorder.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	assertPrivateReportHeaders(t, recorder)
}

func assertPrivateReportHeaders(t *testing.T, recorder *httptest.ResponseRecorder) {
	t.Helper()
	if got := recorder.Header().Get("Cache-Control"); got != "private, no-store" {
		t.Fatalf("Cache-Control=%q, want private, no-store", got)
	}
	if got := recorder.Header().Get("Pragma"); got != "no-cache" {
		t.Fatalf("Pragma=%q, want no-cache", got)
	}
}

func containsJSONKey(raw []byte, key string) bool {
	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		return false
	}
	var walk func(any) bool
	walk = func(current any) bool {
		switch typed := current.(type) {
		case map[string]any:
			for field, nested := range typed {
				if field == key || walk(nested) {
					return true
				}
			}
		case []any:
			for _, nested := range typed {
				if walk(nested) {
					return true
				}
			}
		}
		return false
	}
	return walk(value)
}
