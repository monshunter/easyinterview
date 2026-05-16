package practice

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	api "github.com/monshunter/easyinterview/backend/internal/api/generated"
	"github.com/monshunter/easyinterview/backend/internal/middleware/idempotency"
	domain "github.com/monshunter/easyinterview/backend/internal/practice"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

func TestStartPracticeSessionReturns201WithCurrentTurn(t *testing.T) {
	service := &fakePlanService{startRecord: fixtureSessionRecord()}
	handler := newTestHandlerWithPepper(service, "test-pepper")

	req := newStartSessionHTTPRequest(t, api.StartPracticeSessionRequest{PlanId: "plan-1", HintsEnabled: pointerBool(true)})
	rec := httptest.NewRecorder()
	handler.StartPracticeSession(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var out api.PracticeSession
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode PracticeSession: %v", err)
	}
	if out.Status != sharedtypes.SessionStatusRunning || out.CurrentTurn == nil || out.CurrentTurn.Status != "asked" {
		t.Fatalf("unexpected session response: %+v", out)
	}
	if service.startUserID != "user-1" || service.startPlanID != "plan-1" || !service.startHintsEnabled {
		t.Fatalf("start request not mapped to service: user=%q plan=%q hints=%v", service.startUserID, service.startPlanID, service.startHintsEnabled)
	}
	if service.startKeyHash == "" || service.startFingerprint == "" {
		t.Fatalf("idempotency metadata was not mapped to service")
	}
	if service.startKeyHash != idempotency.HashKey("session-key-1", "test-pepper") {
		t.Fatalf("idempotency key hash = %q, want peppered hash", service.startKeyHash)
	}
}

func TestStartPracticeSessionDebriefReturnsSourceCurrentTurn(t *testing.T) {
	service := &fakePlanService{
		startRecord: func() domain.SessionRecord {
			session := fixtureSessionRecord()
			session.CurrentTurn.QuestionText = "__DEBRIEF_FIRST_QUESTION__"
			session.CurrentTurn.QuestionIntent = "debrief.source_question"
			return session
		}(),
	}
	handler := newTestHandler(service)

	rec := httptest.NewRecorder()
	handler.StartPracticeSession(rec, newStartSessionHTTPRequest(t, api.StartPracticeSessionRequest{PlanId: "plan-debrief"}))
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var out api.PracticeSession
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode PracticeSession: %v", err)
	}
	if out.Status != sharedtypes.SessionStatusRunning ||
		out.CurrentTurn == nil ||
		out.CurrentTurn.QuestionText != "__DEBRIEF_FIRST_QUESTION__" ||
		out.CurrentTurn.QuestionIntent == nil ||
		*out.CurrentTurn.QuestionIntent != "debrief.source_question" {
		t.Fatalf("unexpected debrief response: %+v", out)
	}
}

func TestStartPracticeSessionFixtureParityDefault(t *testing.T) {
	fixture := loadStartSessionFixture(t)
	scenario := fixture.Scenarios["default"]
	service := &fakePlanService{startRecord: sessionRecordFromFixture(scenario.Response.Body)}
	handler := newTestHandler(service)

	rec := httptest.NewRecorder()
	handler.StartPracticeSession(rec, newStartSessionHTTPRequest(t, scenario.Request.Body))
	if rec.Code != scenario.Response.Status {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, scenario.Response.Status, rec.Body.String())
	}
	assertJSONEqual(t, mustJSON(t, scenario.Response.Body), rec.Body.Bytes())
}

func TestStartPracticeSessionMapsAIServiceErrorTo502(t *testing.T) {
	service := &fakePlanService{startErr: &domain.ServiceError{Code: sharederrors.CodeAiProviderTimeout, Message: "AI provider request timed out"}}
	handler := newTestHandler(service)

	rec := httptest.NewRecorder()
	handler.StartPracticeSession(rec, newStartSessionHTTPRequest(t, api.StartPracticeSessionRequest{PlanId: "plan-1"}))
	if rec.Code != http.StatusBadGateway {
		t.Fatalf("status = %d, body=%s", rec.Code, rec.Body.String())
	}
	var out api.ApiErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if out.Error.Code != sharederrors.CodeAiProviderTimeout || !out.Error.Retryable {
		t.Fatalf("unexpected error response: %+v", out.Error)
	}
}

func newStartSessionHTTPRequest(t *testing.T, body api.StartPracticeSessionRequest) *http.Request {
	t.Helper()
	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/practice/sessions", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", "session-key-1")
	return req.WithContext(contextWithUser(req.Context(), "user-1"))
}

type startSessionFixture struct {
	Scenarios map[string]struct {
		Request struct {
			Body api.StartPracticeSessionRequest `json:"body"`
		} `json:"request"`
		Response struct {
			Status int                 `json:"status"`
			Body   api.PracticeSession `json:"body"`
		} `json:"response"`
	} `json:"scenarios"`
}

func loadStartSessionFixture(t *testing.T) startSessionFixture {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("..", "..", "..", "..", "openapi", "fixtures", "PracticeSessions", "startPracticeSession.json"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	var fixture startSessionFixture
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatalf("decode fixture: %v", err)
	}
	return fixture
}

func fixtureSessionRecord() domain.SessionRecord {
	askedAt := time.Date(2026, 4, 28, 13, 45, 12, 0, time.UTC)
	return domain.SessionRecord{
		ID:           "01918fa0-0000-7000-8000-000000005000",
		PlanID:       "plan-1",
		TargetJobID:  "target-1",
		Status:       sharedtypes.SessionStatusRunning,
		Language:     "zh-CN",
		HintsEnabled: true,
		TurnCount:    1,
		CurrentTurn: &domain.TurnRecord{
			ID:             "turn-1",
			TurnIndex:      1,
			QuestionText:   "Question?",
			QuestionIntent: "behavioral",
			Status:         "asked",
			AskedAt:        askedAt,
		},
		CreatedAt: askedAt.Add(-time.Hour),
		UpdatedAt: askedAt,
	}
}

func sessionRecordFromFixture(session api.PracticeSession) domain.SessionRecord {
	createdAt, _ := time.Parse(time.RFC3339, session.CreatedAt)
	updatedAt, _ := time.Parse(time.RFC3339, session.UpdatedAt)
	var turn *domain.TurnRecord
	if session.CurrentTurn != nil {
		askedAt := time.Time{}
		if session.CurrentTurn.AskedAt != nil {
			askedAt, _ = time.Parse(time.RFC3339, *session.CurrentTurn.AskedAt)
		}
		intent := ""
		if session.CurrentTurn.QuestionIntent != nil {
			intent = *session.CurrentTurn.QuestionIntent
		}
		turn = &domain.TurnRecord{
			ID:             session.CurrentTurn.Id,
			TurnIndex:      session.CurrentTurn.TurnIndex,
			QuestionText:   session.CurrentTurn.QuestionText,
			QuestionIntent: intent,
			Status:         session.CurrentTurn.Status,
			AskedAt:        askedAt,
		}
	}
	return domain.SessionRecord{
		ID:           session.Id,
		PlanID:       session.PlanId,
		TargetJobID:  session.TargetJobId,
		Status:       session.Status,
		Language:     session.Language,
		HintsEnabled: session.HintsEnabled,
		TurnCount:    session.TurnCount,
		CurrentTurn:  turn,
		CreatedAt:    createdAt,
		UpdatedAt:    updatedAt,
	}
}

func pointerBool(value bool) *bool {
	return &value
}
