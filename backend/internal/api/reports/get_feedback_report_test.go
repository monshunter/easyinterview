package reports

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	api "github.com/monshunter/easyinterview/backend/internal/api/generated"
	reviewdomain "github.com/monshunter/easyinterview/backend/internal/review"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

func TestGetFeedbackReportHandlerReturnsReadyReport(t *testing.T) {
	now := time.Date(2026, 5, 15, 10, 20, 30, 0, time.UTC)
	svc := &fakeReportService{getReport: sampleReport(now)}
	handler := NewHandler(HandlerOptions{Service: svc, Session: sessionFromContext})

	rec := httptest.NewRecorder()
	handler.GetFeedbackReport(rec, requestWithUser(http.MethodGet, "/api/v1/reports/report-1", "user-1"), "report-1")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	var out api.FeedbackReport
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode FeedbackReport: %v", err)
	}
	if svc.getUserID != "user-1" || svc.getReportID != "report-1" {
		t.Fatalf("service request = %s/%s", svc.getUserID, svc.getReportID)
	}
	if out.Status != sharedtypes.ReportStatusReady || out.Provenance == nil || len(out.QuestionAssessments) != 1 {
		t.Fatalf("unexpected report: %+v", out)
	}
}

func TestGetFeedbackReportHandlerMapsCrossUserToReportNotFound(t *testing.T) {
	svc := &fakeReportService{getErr: reviewdomain.ErrReportNotFound}
	handler := NewHandler(HandlerOptions{Service: svc, Session: sessionFromContext})

	rec := httptest.NewRecorder()
	handler.GetFeedbackReport(rec, requestWithUser(http.MethodGet, "/api/v1/reports/report-1", "user-1"), "report-1")

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	assertAPIError(t, rec, sharederrors.CodeReportNotFound)
}

func TestListTargetJobReportsHandlerParsesCursorAndPageSize(t *testing.T) {
	now := time.Date(2026, 5, 15, 10, 20, 30, 0, time.UTC)
	cursor := reviewdomain.EncodeCursor(now, "0197d120-0000-7000-8000-000000000501")
	svc := &fakeReportService{listResult: reviewdomain.PaginatedFeedbackReportRecord{
		Items:    []reviewdomain.FeedbackReportRecord{sampleReport(now)},
		PageInfo: reviewdomain.PageInfo{PageSize: 5, HasMore: true, NextCursor: cursor},
	}}
	handler := NewHandler(HandlerOptions{Service: svc, Session: sessionFromContext})

	rec := httptest.NewRecorder()
	handler.ListTargetJobReports(rec, requestWithUser(http.MethodGet, "/api/v1/targets/target-1/reports?pageSize=5&cursor="+cursor, "user-1"), "target-1")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if svc.listRequest.UserID != "user-1" || svc.listRequest.TargetJobID != "target-1" || svc.listRequest.PageSize != 5 || svc.listRequest.Cursor != cursor {
		t.Fatalf("list request = %+v", svc.listRequest)
	}
	var out api.PaginatedFeedbackReport
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode PaginatedFeedbackReport: %v", err)
	}
	if !out.PageInfo.HasMore || out.PageInfo.NextCursor == nil || *out.PageInfo.NextCursor != cursor {
		t.Fatalf("pageInfo = %+v", out.PageInfo)
	}
}

func TestListTargetJobReportsHandlerMapsInvalidCursor(t *testing.T) {
	svc := &fakeReportService{listErr: reviewdomain.ErrInvalidCursor}
	handler := NewHandler(HandlerOptions{Service: svc, Session: sessionFromContext})

	rec := httptest.NewRecorder()
	handler.ListTargetJobReports(rec, requestWithUser(http.MethodGet, "/api/v1/targets/target-1/reports?cursor=bad", "user-1"), "target-1")

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	assertAPIError(t, rec, sharederrors.CodeValidationFailed)
}

func sampleReport(now time.Time) reviewdomain.FeedbackReportRecord {
	tier := sharedtypes.ReadinessTierBasicallyReady
	return reviewdomain.FeedbackReportRecord{
		ID:                "0197d120-0000-7000-8000-000000000501",
		SessionID:         "0197d120-0000-7000-8000-000000000502",
		TargetJobID:       "0197d120-0000-7000-8000-000000000503",
		Status:            sharedtypes.ReportStatusReady,
		PreparednessLevel: &tier,
		Highlights:        []reviewdomain.ReportEvidenceRecord{{Dimension: "depth", Evidence: "clear", Confidence: sharedtypes.ConfidenceHigh}},
		Issues:            []reviewdomain.ReportEvidenceRecord{},
		NextActions:       []reviewdomain.ReportNextActionRecord{{Type: string(reviewdomain.NextActionNextRound), Label: "Next round"}},
		QuestionAssessments: []reviewdomain.QuestionAssessmentRecord{{
			TurnID:              "0197d120-0000-7000-8000-000000000504",
			QuestionIntent:      "architecture",
			DimensionResults:    map[string]reviewdomain.DimensionResultRecord{"depth": {Status: sharedtypes.DimensionStatusMeetsBar, Confidence: sharedtypes.ConfidenceHigh}},
			ReviewStatus:        sharedtypes.QuestionReviewStatusOpen,
			IncludedInRetryPlan: false,
		}},
		RetryFocusTurnIDs: []string{},
		Provenance:        &reviewdomain.GenerationProvenanceRecord{PromptVersion: "v0.1.0", RubricVersion: "v0.1.0", ModelID: "model-profile:report.generate.default", Language: "en", FeatureFlag: "none", DataSourceVersion: "registry.v1"},
		CreatedAt:         now,
		UpdatedAt:         now,
	}
}

type fakeReportService struct {
	getReport   reviewdomain.FeedbackReportRecord
	getErr      error
	getUserID   string
	getReportID string

	listResult  reviewdomain.PaginatedFeedbackReportRecord
	listErr     error
	listRequest reviewdomain.ListTargetJobReportsRequest
}

func (f *fakeReportService) GetFeedbackReport(_ context.Context, userID, reportID string) (reviewdomain.FeedbackReportRecord, error) {
	f.getUserID = userID
	f.getReportID = reportID
	return f.getReport, f.getErr
}

func (f *fakeReportService) ListTargetJobReports(_ context.Context, in reviewdomain.ListTargetJobReportsRequest) (reviewdomain.PaginatedFeedbackReportRecord, error) {
	f.listRequest = in
	return f.listResult, f.listErr
}

func requestWithUser(method, path, userID string) *http.Request {
	req := httptest.NewRequest(method, path, nil)
	return req.WithContext(context.WithValue(req.Context(), testUserKey{}, userID))
}

type testUserKey struct{}

func sessionFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(testUserKey{}).(string)
	return userID, ok && strings.TrimSpace(userID) != ""
}

func assertAPIError(t *testing.T, rec *httptest.ResponseRecorder, code string) {
	t.Helper()
	var out api.ApiErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode ApiErrorResponse: %v body=%s", err, rec.Body.String())
	}
	if out.Error.Code != code {
		t.Fatalf("error code = %s, want %s", out.Error.Code, code)
	}
}

var _ = errors.Is
