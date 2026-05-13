package practice

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	api "github.com/monshunter/easyinterview/backend/internal/api/generated"
	"github.com/monshunter/easyinterview/backend/internal/middleware/idempotency"
	domain "github.com/monshunter/easyinterview/backend/internal/practice"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

func TestAppendSessionEventReturns200ForSupportedKinds(t *testing.T) {
	kinds := []string{"answer_submitted", "hint_requested", "turn_skipped", "session_paused", "session_resumed"}
	for _, kind := range kinds {
		t.Run(kind, func(t *testing.T) {
			service := &fakePlanService{appendResult: fixtureAppendResult(kind)}
			handler := newTestHandler(service)

			rec := httptest.NewRecorder()
			handler.AppendSessionEvent(rec, newAppendEventHTTPRequest(t, kind, false), "session-1")
			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, body=%s", rec.Code, rec.Body.String())
			}
			var out api.SessionEventResult
			if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
				t.Fatalf("decode SessionEventResult: %v", err)
			}
			if !out.Acknowledged || out.AssistantAction.Provenance.PromptVersion == "" {
				t.Fatalf("unexpected response: %+v", out)
			}
			if service.appendRequest.Kind != kind || service.appendRequest.UserID != "user-1" || service.appendRequest.SessionID != "session-1" {
				t.Fatalf("request not mapped to service: %+v", service.appendRequest)
			}
		})
	}
}

func TestAppendSessionEventRejectsIdempotencyKeyHeader(t *testing.T) {
	handler := newTestHandler(&fakePlanService{appendResult: fixtureAppendResult("answer_submitted")})

	rec := httptest.NewRecorder()
	handler.AppendSessionEvent(rec, newAppendEventHTTPRequest(t, "answer_submitted", true), "session-1")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "use_client_event_id") {
		t.Fatalf("expected policy detail, body=%s", rec.Body.String())
	}
}

func TestAppendSessionEventRequiresOccurredAt(t *testing.T) {
	service := &fakePlanService{appendResult: fixtureAppendResult("session_paused")}
	handler := newTestHandler(service)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/practice/sessions/session-1/events", strings.NewReader(`{
		"clientEventId": "client-event-1",
		"kind": "session_paused"
	}`))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(contextWithUser(req.Context(), "user-1"))

	rec := httptest.NewRecorder()
	handler.AppendSessionEvent(rec, req, "session-1")
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, body=%s", rec.Code, rec.Body.String())
	}
	if service.appendRequest.ClientEventID != "" {
		t.Fatalf("service should not be called when occurredAt is missing: %+v", service.appendRequest)
	}
	assertAPIError(t, rec, sharederrors.CodeValidationFailed, false)
	if !strings.Contains(rec.Body.String(), "occurredAt") {
		t.Fatalf("error should identify occurredAt, body=%s", rec.Body.String())
	}
}

func TestAppendSessionEventMapsReplayMismatchAndCrossUser(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want int
		code string
	}{
		{
			name: "replay",
			err:  nil,
			want: http.StatusOK,
		},
		{
			name: "mismatch",
			err: &domain.ServiceError{
				Code:    sharederrors.CodePracticeSessionConflict,
				Message: "clientEventId was already used with a different payload",
				Details: map[string]any{
					"policy": "client_event_payload_mismatch",
				},
			},
			want: http.StatusConflict,
			code: sharederrors.CodePracticeSessionConflict,
		},
		{
			name: "cross-user",
			err:  &domain.ServiceError{Code: sharederrors.CodePracticeSessionNotFound, Message: "practice session not found"},
			want: http.StatusNotFound,
			code: sharederrors.CodePracticeSessionNotFound,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			service := &fakePlanService{appendResult: fixtureAppendResult("answer_submitted"), appendErr: tc.err}
			handler := newTestHandler(service)
			rec := httptest.NewRecorder()
			handler.AppendSessionEvent(rec, newAppendEventHTTPRequest(t, "answer_submitted", false), "session-1")
			if rec.Code != tc.want {
				t.Fatalf("status = %d, want %d; body=%s", rec.Code, tc.want, rec.Body.String())
			}
			if tc.code != "" {
				assertAPIError(t, rec, tc.code, false)
			}
		})
	}
}

func TestCompletePracticeSessionReturns202ReportWithJob(t *testing.T) {
	now := time.Date(2026, 4, 28, 13, 45, 12, 0, time.UTC)
	service := &fakePlanService{completeResult: domain.CompleteSessionResult{
		ReportID: "report-1",
		Job: domain.JobRecord{
			ID:           "job-1",
			JobType:      api.JobTypeReportGenerate,
			ResourceType: api.ResourceTypeFeedbackReport,
			ResourceID:   "report-1",
			Status:       sharedtypes.JobStatusQueued,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
	}}
	handler := newTestHandler(service)

	rec := httptest.NewRecorder()
	handler.CompletePracticeSession(rec, newCompleteHTTPRequest(t, "idem-complete-1"), "session-1")
	if rec.Code != http.StatusAccepted {
		t.Fatalf("status = %d, body=%s", rec.Code, rec.Body.String())
	}
	var out api.ReportWithJob
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode ReportWithJob: %v", err)
	}
	if out.ReportId != "report-1" || out.Job.Id != "job-1" || out.Job.Status != sharedtypes.JobStatusQueued {
		t.Fatalf("unexpected complete response: %+v", out)
	}
	if service.completeRequest.UserID != "user-1" || service.completeRequest.SessionID != "session-1" {
		t.Fatalf("request not mapped to service: %+v", service.completeRequest)
	}
}

func TestCompletePracticeSessionRequiresClientCompletedAt(t *testing.T) {
	service := &fakePlanService{completeResult: domain.CompleteSessionResult{
		ReportID: "report-1",
		Job:      domain.JobRecord{ID: "job-1"},
	}}
	handler := newTestHandler(service)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/practice/sessions/session-1/complete", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(contextWithUser(req.Context(), "user-1"))

	rec := httptest.NewRecorder()
	handler.CompletePracticeSession(rec, req, "session-1")
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, body=%s", rec.Code, rec.Body.String())
	}
	if service.completeRequest.SessionID != "" {
		t.Fatalf("service should not be called when clientCompletedAt is missing: %+v", service.completeRequest)
	}
	assertAPIError(t, rec, sharederrors.CodeValidationFailed, false)
	if !strings.Contains(rec.Body.String(), "clientCompletedAt") {
		t.Fatalf("error should identify clientCompletedAt, body=%s", rec.Body.String())
	}
}

func TestCompletePracticeSessionMiddlewarePersistsResourceAndReplay(t *testing.T) {
	now := time.Date(2026, 4, 28, 13, 45, 12, 0, time.UTC)
	service := &fakePlanService{completeResult: domain.CompleteSessionResult{
		ReportID: "report-1",
		Job: domain.JobRecord{
			ID:           "job-1",
			JobType:      api.JobTypeReportGenerate,
			ResourceType: api.ResourceTypeFeedbackReport,
			ResourceID:   "report-1",
			Status:       sharedtypes.JobStatusQueued,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
	}}
	handler := newTestHandler(service)
	store := newRouteMemoryStore()
	mw := idempotency.New(idempotency.MiddlewareOptions{
		Store: store,
		Now:   func() time.Time { return now },
		NewID: func() string { return "idem-record-1" },
	})
	route := mw.Handler("practice", "completePracticeSession", userFromRequestContext, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler.CompletePracticeSession(w, r, "session-1")
	}))

	first := httptest.NewRecorder()
	route.ServeHTTP(first, newCompleteHTTPRequest(t, "idem-complete-1"))
	second := httptest.NewRecorder()
	route.ServeHTTP(second, newCompleteHTTPRequest(t, "idem-complete-1"))

	if first.Code != http.StatusAccepted || second.Code != http.StatusAccepted {
		t.Fatalf("statuses: first=%d second=%d secondBody=%s", first.Code, second.Code, second.Body.String())
	}
	if second.Header().Get(idempotency.ReplayHeader) != "true" {
		t.Fatalf("expected replay header")
	}
	if first.Header().Get("X-Idempotency-Resource-ID") != "" {
		t.Fatalf("internal idempotency resource header leaked")
	}
	rec := store.records["user-1\x00practice\x00completePracticeSession\x00"+idempotency.HashKey("idem-complete-1", "")]
	if rec.resourceID != "report-1" {
		t.Fatalf("resourceID = %q, want report-1", rec.resourceID)
	}
}

func newAppendEventHTTPRequest(t *testing.T, kind string, withIdempotency bool) *http.Request {
	t.Helper()
	raw, err := json.Marshal(api.PracticeSessionEventRequest{
		ClientEventId: "client-event-1",
		Kind:          kind,
		OccurredAt:    "2026-04-28T13:45:12Z",
		Payload: map[string]any{
			"turnId":     "turn-1",
			"answerText": "answer",
		},
	})
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/practice/sessions/session-1/events", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	if withIdempotency {
		req.Header.Set(idempotency.HeaderName, "wrong-key")
	}
	return req.WithContext(contextWithUser(req.Context(), "user-1"))
}

func newCompleteHTTPRequest(t *testing.T, idemKey string) *http.Request {
	t.Helper()
	raw, err := json.Marshal(api.CompletePracticeSessionRequest{ClientCompletedAt: "2026-04-28T13:45:12Z"})
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/practice/sessions/session-1/complete", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(idempotency.HeaderName, idemKey)
	return req.WithContext(contextWithUser(req.Context(), "user-1"))
}

func fixtureAppendResult(kind string) domain.AppendSessionEventResult {
	session := fixtureSessionRecord()
	action := domain.AssistantActionRecord{
		Type:          "session_wait",
		SessionStatus: session.Status,
		Provenance: domain.AssistantActionProvenance{
			PromptVersion:     "not_applicable",
			RubricVersion:     "not_applicable",
			ModelID:           "model-profile:static",
			Language:          session.Language,
			FeatureFlag:       "none",
			DataSourceVersion: "static",
		},
	}
	switch kind {
	case "answer_submitted":
		action.Type = "ask_follow_up"
		action.TurnID = "turn-1"
		action.QuestionText = "Follow up?"
	case "turn_skipped":
		action.Type = "ask_question"
		action.TurnID = "turn-2"
		action.QuestionText = "Next question?"
	}
	return domain.AppendSessionEventResult{
		Acknowledged:    true,
		Session:         session,
		AssistantAction: action,
	}
}
