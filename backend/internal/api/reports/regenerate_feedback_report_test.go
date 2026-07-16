package reports

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	api "github.com/monshunter/easyinterview/backend/internal/api/generated"
	reviewdomain "github.com/monshunter/easyinterview/backend/internal/review"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
	sharedjobs "github.com/monshunter/easyinterview/backend/internal/shared/jobs"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

func TestRegenerateFeedbackReportReturnsSameIDQueuedJobAndResourceMarker(t *testing.T) {
	now := time.Date(2026, 7, 16, 10, 0, 0, 0, time.UTC)
	service := &regenerateReportHandlerService{result: reviewdomain.RegenerateReportResult{
		ReportID: "report-1",
		Job: reviewdomain.ReportJobRecord{
			ID: "job-1", JobType: string(sharedjobs.JobTypeReportGenerate), ResourceType: "feedback_report",
			ResourceID: "report-1", Status: sharedtypes.JobStatusQueued, CreatedAt: now, UpdatedAt: now,
		},
	}}
	handler := NewHandler(HandlerOptions{Service: service, Session: func(context.Context) (string, bool) { return "user-1", true }})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/reports/report-1/regenerate", nil)
	request.Header.Set("X-Request-ID", "req-report-regenerate")

	handler.RegenerateFeedbackReport(recorder, request, "report-1")
	if recorder.Code != http.StatusAccepted {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	if service.input.UserID != "user-1" || service.input.ReportID != "report-1" || service.calls != 1 {
		t.Fatalf("service input=%+v calls=%d", service.input, service.calls)
	}
	if recorder.Header().Get("X-Idempotency-Resource-Type") != "feedback_report" || recorder.Header().Get("X-Idempotency-Resource-ID") != "report-1" || recorder.Header().Get("X-Request-ID") != "req-report-regenerate" {
		t.Fatalf("headers=%v", recorder.Header())
	}
	assertPrivateReportHeaders(t, recorder)
	var response api.ReportWithJob
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if response.ReportId != "report-1" || response.Job.ResourceId != "report-1" || response.Job.Id != "job-1" || response.Job.JobType != api.JobTypeReportGenerate || response.Job.ResourceType != api.ResourceTypeFeedbackReport || response.Job.Status != sharedtypes.JobStatusQueued {
		t.Fatalf("response=%+v", response)
	}
}

func TestRegenerateFeedbackReportMapsTypedEligibilityFailures(t *testing.T) {
	for _, tc := range []struct {
		name       string
		err        error
		wantStatus int
		wantCode   string
		retryable  bool
	}{
		{name: "hidden not found", err: reviewdomain.ErrReportNotFound, wantStatus: http.StatusNotFound, wantCode: sharederrors.CodeReportNotFound},
		{name: "active old job", err: reviewdomain.ErrReportNotReady, wantStatus: http.StatusConflict, wantCode: sharederrors.CodeReportNotReady, retryable: true},
		{name: "context too large", err: reviewdomain.ErrReportContextTooLarge, wantStatus: http.StatusConflict, wantCode: sharederrors.CodeReportContextTooLarge},
		{name: "invalid state", err: reviewdomain.ErrReportInvalidStateTransition, wantStatus: http.StatusConflict, wantCode: sharederrors.CodeReportInvalidStateTransition},
	} {
		t.Run(tc.name, func(t *testing.T) {
			service := &regenerateReportHandlerService{err: fmtWrapped(tc.err)}
			handler := NewHandler(HandlerOptions{Service: service, Session: func(context.Context) (string, bool) { return "user-1", true }})
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodPost, "/api/v1/reports/report-1/regenerate", nil)
			request.Header.Set("X-Request-ID", "req-report-regenerate-error")
			handler.RegenerateFeedbackReport(recorder, request, "report-1")
			if recorder.Code != tc.wantStatus {
				t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
			}
			var response api.ApiErrorResponse
			if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
				t.Fatal(err)
			}
			if response.Error.Code != tc.wantCode || response.Error.Retryable != tc.retryable || response.Error.RequestID != "req-report-regenerate-error" || strings.Contains(recorder.Body.String(), "store detail") {
				t.Fatalf("error=%+v body=%s", response.Error, recorder.Body.String())
			}
		})
	}
}

func fmtWrapped(err error) error {
	return errors.Join(err, errors.New("store detail must stay private"))
}

type regenerateReportHandlerService struct {
	result reviewdomain.RegenerateReportResult
	err    error
	input  reviewdomain.RegenerateReportRequest
	calls  int
}

func (s *regenerateReportHandlerService) RegenerateReport(_ context.Context, in reviewdomain.RegenerateReportRequest) (reviewdomain.RegenerateReportResult, error) {
	s.calls++
	s.input = in
	return s.result, s.err
}

func (*regenerateReportHandlerService) GetFeedbackReport(context.Context, string, string) (reviewdomain.FeedbackReportRecord, error) {
	return reviewdomain.FeedbackReportRecord{}, nil
}

func (*regenerateReportHandlerService) GetReportConversation(context.Context, string, string) (reviewdomain.ReportConversationRecord, error) {
	return reviewdomain.ReportConversationRecord{}, nil
}

func (*regenerateReportHandlerService) ListTargetJobReports(context.Context, reviewdomain.ListTargetJobReportsRequest) (reviewdomain.TargetJobReportsOverviewRecord, error) {
	return reviewdomain.TargetJobReportsOverviewRecord{}, nil
}

func TestHandlerExposesGeneratedFailedReportRegenerationOperation(t *testing.T) {
	method, ok := reflect.TypeOf((*Handler)(nil)).MethodByName("RegenerateFeedbackReport")
	if !ok {
		t.Fatal("reports Handler does not implement generated RegenerateFeedbackReport operation")
	}
	if method.Type.NumIn() != 4 {
		t.Fatalf("RegenerateFeedbackReport input count=%d, want receiver + writer + request + reportID", method.Type.NumIn())
	}
}

func TestRegenerateFeedbackReportHandlerKeepsSameReportAndMapsTypedFailures(t *testing.T) {
	raw, err := os.ReadFile("regenerate_feedback_report.go")
	if err != nil {
		t.Fatalf("read regenerate_feedback_report.go: %v", err)
	}
	source := string(raw)
	for _, required := range []string{
		"RegenerateFeedbackReport",
		"RegenerateReportRequest",
		"http.StatusAccepted",
		"idempotency.SetResponseResource",
		"ReportId:",
		"CodeReportNotFound",
		"CodeReportNotReady",
		"CodeReportContextTooLarge",
		"CodeReportInvalidStateTransition",
	} {
		if !strings.Contains(source, required) {
			t.Fatalf("regeneration handler missing %q", required)
		}
	}
	for _, forbidden := range []string{"json.NewDecoder", "aiclient", ".Complete(", "generation_context", "practice_messages"} {
		if strings.Contains(source, forbidden) {
			t.Fatalf("regeneration handler must not contain %q", forbidden)
		}
	}
}
