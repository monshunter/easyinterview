package reports

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	api "github.com/monshunter/easyinterview/backend/internal/api/generated"
	reviewdomain "github.com/monshunter/easyinterview/backend/internal/review"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

type reportConversationHandlerService struct {
	conversation reviewdomain.ReportConversationRecord
	err          error
	userID       string
	reportID     string
}

func (s *reportConversationHandlerService) GetFeedbackReport(context.Context, string, string) (reviewdomain.FeedbackReportRecord, error) {
	return reviewdomain.FeedbackReportRecord{}, s.err
}

func (s *reportConversationHandlerService) GetReportConversation(_ context.Context, userID, reportID string) (reviewdomain.ReportConversationRecord, error) {
	s.userID = userID
	s.reportID = reportID
	return s.conversation, s.err
}

func (s *reportConversationHandlerService) ListTargetJobReports(context.Context, reviewdomain.ListTargetJobReportsRequest) (reviewdomain.TargetJobReportsOverviewRecord, error) {
	return reviewdomain.TargetJobReportsOverviewRecord{}, s.err
}

func TestGetReportConversationProjectsClosedOrderedMessagesForEveryReportStatus(t *testing.T) {
	now := time.Date(2026, 7, 15, 8, 0, 0, 0, time.UTC)
	for _, status := range []sharedtypes.ReportStatus{
		sharedtypes.ReportStatusQueued,
		sharedtypes.ReportStatusGenerating,
		sharedtypes.ReportStatusReady,
		sharedtypes.ReportStatusFailed,
	} {
		t.Run(string(status), func(t *testing.T) {
			service := &reportConversationHandlerService{conversation: reviewdomain.ReportConversationRecord{
				ReportID: "report-1", Status: status,
				Context: reportConversationContext(),
				Messages: []reviewdomain.ReportConversationMessageRecord{
					{Sequence: 1, Role: "user", Content: "我主导了迁移。", CreatedAt: now},
					{Sequence: 2, Role: "assistant", Content: "请说明取舍。", CreatedAt: now.Add(12 * time.Second)},
				},
			}}
			handler := NewHandler(HandlerOptions{Service: service, Session: func(context.Context) (string, bool) { return "user-1", true }})
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, "/api/v1/reports/report-1/conversation", nil)
			request.Header.Set("X-Request-ID", "req-report-conversation")

			handler.GetReportConversation(recorder, request, "report-1")
			if recorder.Code != http.StatusOK || service.userID != "user-1" || service.reportID != "report-1" {
				t.Fatalf("status=%d input=%q/%q body=%s", recorder.Code, service.userID, service.reportID, recorder.Body.String())
			}
			if got := recorder.Header().Get("X-Request-ID"); got != "req-report-conversation" {
				t.Fatalf("X-Request-ID=%q", got)
			}
			assertPrivateReportHeaders(t, recorder)
			var body map[string]any
			if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			assertExactJSONKeys(t, body, "reportId", "reportStatus", "context", "messages")
			if body["reportId"] != "report-1" || body["reportStatus"] != string(status) {
				t.Fatalf("body=%v", body)
			}
			messages, ok := body["messages"].([]any)
			if !ok || len(messages) != 2 {
				t.Fatalf("messages=%#v", body["messages"])
			}
			for _, message := range messages {
				assertExactJSONKeys(t, message.(map[string]any), "sequence", "role", "content", "createdAt")
			}
			for _, forbidden := range []string{"sessionId", "messageId", "clientMessageId", "replyStatus", "replyGeneration", "id"} {
				if containsJSONKey(recorder.Body.Bytes(), forbidden) {
					t.Fatalf("response leaked %q: %s", forbidden, recorder.Body.String())
				}
			}
		})
	}
}

func TestGetReportConversationMapsHiddenAndFailClosedErrorsWithoutTranscriptLeak(t *testing.T) {
	cases := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   string
	}{
		{name: "cross user hidden", err: reviewdomain.ErrReportNotFound, wantStatus: http.StatusNotFound, wantCode: sharederrors.CodeReportNotFound},
		{name: "corrupt projection", err: fmt.Errorf("corrupt report message: %w: raw transcript must not leak", reviewdomain.ErrReportConversationInvalid), wantStatus: http.StatusInternalServerError, wantCode: sharederrors.CodeAiOutputInvalid},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			service := &reportConversationHandlerService{err: tc.err}
			handler := NewHandler(HandlerOptions{Service: service, Session: func(context.Context) (string, bool) { return "user-1", true }})
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, "/api/v1/reports/report-1/conversation", nil)
			request.Header.Set("X-Request-ID", "req-report-conversation-error")

			handler.GetReportConversation(recorder, request, "report-1")
			if recorder.Code != tc.wantStatus {
				t.Fatalf("status=%d want=%d body=%s", recorder.Code, tc.wantStatus, recorder.Body.String())
			}
			assertPrivateReportHeaders(t, recorder)
			var response api.ApiErrorResponse
			if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
				t.Fatal(err)
			}
			if response.Error.Code != tc.wantCode || response.Error.RequestID != "req-report-conversation-error" || response.Error.Retryable || strings.Contains(recorder.Body.String(), "raw transcript") {
				t.Fatalf("response=%+v body=%s", response.Error, recorder.Body.String())
			}
		})
	}
}

func TestReportConversationProjectionUsesOnlyClosedReadModelFields(t *testing.T) {
	now := time.Date(2026, 7, 15, 8, 0, 0, 0, time.UTC)
	conversation := reviewdomain.ReportConversationRecord{
		ReportID: "report-1", Status: sharedtypes.ReportStatusReady, Context: reportConversationContext(),
		Messages: []reviewdomain.ReportConversationMessageRecord{{Sequence: 1, Role: "user", Content: "**STAR**", CreatedAt: now}},
	}
	got := toAPIReportConversation(conversation)
	if got.ReportId != conversation.ReportID || got.ReportStatus != conversation.Status || !reflect.DeepEqual(got.Context, toAPIReportContext(conversation.Context)) || len(got.Messages) != 1 || got.Messages[0].Sequence != 1 || got.Messages[0].Role != "user" || got.Messages[0].Content != "**STAR**" || got.Messages[0].CreatedAt != now.Format(timeFormatRFC3339) {
		t.Fatalf("projection=%#v", got)
	}
	raw, err := json.Marshal(got)
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"sessionId", "messageId", "clientMessageId", "replyStatus", "replyGeneration", "anchor"} {
		if containsJSONKey(raw, forbidden) {
			t.Fatalf("projection leaked %q: %s", forbidden, raw)
		}
	}
}

func TestReportConversationProjectionPreservesSubsecondCreatedAt(t *testing.T) {
	createdAt := time.Date(2026, 7, 15, 8, 0, 0, 123456000, time.UTC)
	conversation := reviewdomain.ReportConversationRecord{
		ReportID: "report-1", Status: sharedtypes.ReportStatusReady, Context: reportConversationContext(),
		Messages: []reviewdomain.ReportConversationMessageRecord{{Sequence: 1, Role: "user", Content: "具体回答。", CreatedAt: createdAt}},
	}

	got := toAPIReportConversation(conversation)
	if len(got.Messages) != 1 || got.Messages[0].CreatedAt != createdAt.Format(time.RFC3339Nano) {
		t.Fatalf("createdAt=%q want=%q", got.Messages[0].CreatedAt, createdAt.Format(time.RFC3339Nano))
	}
}

func reportConversationContext() reviewdomain.ReportContextProjection {
	return reviewdomain.ReportContextProjection{
		SourcePlanID: "plan-1", TargetJobTitle: "平台工程师", TargetJobCompany: "Acme",
		ResumeID: "resume-1", ResumeDisplayName: "平台工程简历",
		RoundID: "round-1-technical", RoundSequence: 1, RoundName: "技术面", RoundType: "technical",
		Language: "zh-CN", HasNextRound: true,
	}
}
