package reports

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	api "github.com/monshunter/easyinterview/backend/internal/api/generated"
	reviewdomain "github.com/monshunter/easyinterview/backend/internal/review"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

func TestListTargetJobReportsHandlerProjectsOnlyClosedMinimalOverviewAndIgnoresQuery(t *testing.T) {
	now := time.Date(2026, 7, 14, 8, 0, 0, 0, time.UTC)
	code := sharederrors.CodeAiOutputInvalid
	service := &targetJobReportsHandlerService{overview: reviewdomain.TargetJobReportsOverviewRecord{
		TargetJobID: "target-1",
		Rounds: []reviewdomain.TargetJobReportRoundOverviewRecord{
			{
				Round:         reviewdomain.PracticeRoundRefRecord{RoundID: "round-1-technical", RoundSequence: 1},
				CurrentReport: &reviewdomain.TargetJobCurrentReportSummaryRecord{ID: "report-ready", GeneratedAt: now},
				LatestAttempt: &reviewdomain.TargetJobReportAttemptSummaryRecord{ID: "report-failed", Status: sharedtypes.ReportStatusFailed, ErrorCode: &code, CreatedAt: now.Add(time.Minute)},
			},
			{
				Round:         reviewdomain.PracticeRoundRefRecord{RoundID: "round-2-manager", RoundSequence: 2},
				LatestAttempt: &reviewdomain.TargetJobReportAttemptSummaryRecord{ID: "report-generating", Status: sharedtypes.ReportStatusGenerating, CreatedAt: now.Add(2 * time.Minute)},
			},
			{Round: reviewdomain.PracticeRoundRefRecord{RoundID: "round-3-culture", RoundSequence: 3}},
		},
	}}
	handler := NewHandler(HandlerOptions{
		Service: service,
		Session: func(context.Context) (string, bool) { return "user-1", true },
	})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/targets/target-1/reports?cursor=ignored&pageSize=999", nil)
	request.Header.Set("X-Request-ID", "req_2026-07-14-report-overview-current-ready")

	handler.ListTargetJobReports(recorder, request, "target-1")
	if recorder.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	if service.request.UserID != "user-1" || service.request.TargetJobID != "target-1" {
		t.Fatalf("service request=%#v", service.request)
	}
	if got := recorder.Header().Get("X-Request-ID"); got != "req_2026-07-14-report-overview-current-ready" {
		t.Fatalf("X-Request-ID=%q", got)
	}
	assertPrivateReportHeaders(t, recorder)
	var got api.TargetJobReportsOverview
	if err := json.Unmarshal(recorder.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.TargetJobId != "target-1" || len(got.Rounds) != 3 || got.Rounds[0].CurrentReport == nil || got.Rounds[0].LatestAttempt == nil || got.Rounds[1].CurrentReport != nil || got.Rounds[1].LatestAttempt == nil || got.Rounds[1].LatestAttempt.ErrorCode != nil || got.Rounds[2].CurrentReport != nil || got.Rounds[2].LatestAttempt != nil {
		t.Fatalf("overview=%#v", got)
	}
	var raw map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &raw); err != nil {
		t.Fatal(err)
	}
	assertExactJSONKeys(t, raw, "targetJobId", "rounds")
	rounds := raw["rounds"].([]any)
	first := rounds[0].(map[string]any)
	assertExactJSONKeys(t, first, "round", "currentReport", "latestAttempt")
	assertExactJSONKeys(t, first["round"].(map[string]any), "roundId", "roundSequence")
	assertExactJSONKeys(t, first["currentReport"].(map[string]any), "id", "generatedAt")
	assertExactJSONKeys(t, first["latestAttempt"].(map[string]any), "id", "status", "errorCode", "createdAt")
	second := rounds[1].(map[string]any)
	if latest := second["latestAttempt"].(map[string]any); latest["errorCode"] != nil {
		t.Fatalf("non-failed errorCode=%v, want explicit null", latest["errorCode"])
	}
	third := rounds[2].(map[string]any)
	assertExactJSONKeys(t, third, "round", "currentReport", "latestAttempt")
	if third["currentReport"] != nil || third["latestAttempt"] != nil {
		t.Fatalf("empty canonical round pointers=%v/%v, want explicit nulls", third["currentReport"], third["latestAttempt"])
	}
	for _, forbidden := range []string{"summary", "pageInfo", "cursor", "pageSize", "provenance", "modelId", "rubricVersion", "sessionId", "roundName", "roundType"} {
		if containsJSONKey(recorder.Body.Bytes(), forbidden) {
			t.Fatalf("overview leaked forbidden key %q: %s", forbidden, recorder.Body.String())
		}
	}
}

func TestListTargetJobReportsHandlerMapsHidden404AndInvalidContextToOperationFixtures(t *testing.T) {
	for _, tc := range []struct {
		name     string
		scenario string
		err      error
	}{
		{name: "hidden target not found", scenario: "target-not-found", err: reviewdomain.ErrReportNotFound},
		{name: "invalid frozen context", scenario: "invalid-frozen-context", err: reviewdomain.ErrReportContextInvalid},
		{name: "missing frozen context", scenario: "missing-frozen-context", err: reviewdomain.ErrReportContextMissing},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fixture := loadTargetJobReportsErrorFixture(t, tc.scenario)
			handler := NewHandler(HandlerOptions{
				Service: &targetJobReportsHandlerService{err: tc.err},
				Session: func(context.Context) (string, bool) { return "user-1", true },
			})
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, "/api/v1/targets/target-1/reports", nil)
			request.Header.Set("X-Request-ID", fixture.Headers["X-Request-ID"])
			handler.ListTargetJobReports(recorder, request, "target-1")
			if recorder.Code != fixture.Status {
				t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
			}
			if got := recorder.Header().Get("X-Request-ID"); got != fixture.Headers["X-Request-ID"] {
				t.Fatalf("X-Request-ID=%q want=%q", got, fixture.Headers["X-Request-ID"])
			}
			var got api.ApiErrorResponse
			if err := json.Unmarshal(recorder.Body.Bytes(), &got); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, fixture.Body) || containsJSONKey(recorder.Body.Bytes(), "rounds") {
				t.Fatalf("response=%#v want=%#v raw=%s", got, fixture.Body, recorder.Body.String())
			}
		})
	}
}

type targetJobReportsErrorFixture struct {
	Scenarios map[string]struct {
		Response targetJobReportsFixtureResponse `json:"response"`
	} `json:"scenarios"`
}

type targetJobReportsFixtureResponse struct {
	Status  int                  `json:"status"`
	Headers map[string]string    `json:"headers"`
	Body    api.ApiErrorResponse `json:"body"`
}

func loadTargetJobReportsErrorFixture(t *testing.T, scenario string) targetJobReportsFixtureResponse {
	t.Helper()
	path := filepath.Join("..", "..", "..", "..", "openapi", "fixtures", "Reports", "listTargetJobReports.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read listTargetJobReports fixture: %v", err)
	}
	var fixture targetJobReportsErrorFixture
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatalf("decode listTargetJobReports fixture: %v", err)
	}
	entry, ok := fixture.Scenarios[scenario]
	if !ok {
		t.Fatalf("missing listTargetJobReports fixture scenario %q", scenario)
	}
	return entry.Response
}

type targetJobReportsHandlerService struct {
	overview reviewdomain.TargetJobReportsOverviewRecord
	err      error
	request  reviewdomain.ListTargetJobReportsRequest
}

func (s *targetJobReportsHandlerService) GetFeedbackReport(context.Context, string, string) (reviewdomain.FeedbackReportRecord, error) {
	return reviewdomain.FeedbackReportRecord{}, s.err
}

func (s *targetJobReportsHandlerService) GetReportConversation(context.Context, string, string) (reviewdomain.ReportConversationRecord, error) {
	return reviewdomain.ReportConversationRecord{}, s.err
}

func (s *targetJobReportsHandlerService) ListTargetJobReports(_ context.Context, request reviewdomain.ListTargetJobReportsRequest) (reviewdomain.TargetJobReportsOverviewRecord, error) {
	s.request = request
	return s.overview, s.err
}

func assertExactJSONKeys(t *testing.T, got map[string]any, want ...string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("keys=%v want=%v", got, want)
	}
	for _, key := range want {
		if _, ok := got[key]; !ok {
			t.Fatalf("missing key %q in %v", key, got)
		}
	}
}
