package practice

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	api "github.com/monshunter/easyinterview/backend/internal/api/generated"
	"github.com/monshunter/easyinterview/backend/internal/middleware/idempotency"
	domain "github.com/monshunter/easyinterview/backend/internal/practice"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

func TestCreatePracticePlanReturns201WithGeneratedPracticePlan(t *testing.T) {
	service := &fakePlanService{
		record: fixturePlanRecord(),
	}
	handler := newTestHandler(service)

	req := newCreatePlanHTTPRequest(t, fixtureCreatePlanRequest())
	rec := httptest.NewRecorder()
	handler.CreatePracticePlan(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var out api.PracticePlan
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode PracticePlan: %v", err)
	}
	if out.Id != service.record.ID || out.Status != "ready" || out.TargetJobId != service.record.TargetJobID {
		t.Fatalf("unexpected response: %+v", out)
	}
	if service.last.UserID != "user-1" || service.last.TargetJobID != "01918fa0-0000-7000-8000-000000002000" {
		t.Fatalf("request not mapped to service: %+v", service.last)
	}
}

func TestCreatePracticePlanReturns422ForValidationError(t *testing.T) {
	service := &fakePlanService{
		err: &domain.ServiceError{
			Code:    sharederrors.CodeValidationFailed,
			Message: "sourceReportId is required for this practice goal",
			Details: map[string]any{
				"goal":  string(sharedtypes.PracticeGoalNextRound),
				"field": "sourceReportId",
			},
		},
	}
	handler := newTestHandler(service)
	body := fixtureCreatePlanRequest()
	body.Goal = sharedtypes.PracticeGoalNextRound

	rec := httptest.NewRecorder()
	handler.CreatePracticePlan(rec, newCreatePlanHTTPRequest(t, body))
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	assertAPIError(t, rec, sharederrors.CodeValidationFailed, false)
	if !strings.Contains(rec.Body.String(), "sourceReportId") {
		t.Fatalf("validation details missing from error response: %s", rec.Body.String())
	}
}

func TestCreatePracticePlanMapsDerivedSourceIds(t *testing.T) {
	sourceDebriefID := "01918fa0-0000-7000-8000-00000000d001"
	service := &fakePlanService{
		record: func() domain.PlanRecord {
			plan := fixturePlanRecord()
			plan.Goal = sharedtypes.PracticeGoalDebrief
			plan.SourceDebriefID = sourceDebriefID
			return plan
		}(),
	}
	handler := newTestHandler(service)
	body := fixtureCreatePlanRequest()
	body.Goal = sharedtypes.PracticeGoalDebrief
	body.SourceDebriefId = &sourceDebriefID

	rec := httptest.NewRecorder()
	handler.CreatePracticePlan(rec, newCreatePlanHTTPRequest(t, body))
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if service.last.SourceDebriefID != sourceDebriefID {
		t.Fatalf("sourceDebriefId not mapped to service: %+v", service.last)
	}
	var out api.PracticePlan
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode PracticePlan: %v", err)
	}
	if out.SourceDebriefId == nil || *out.SourceDebriefId != sourceDebriefID {
		t.Fatalf("sourceDebriefId not mapped to response: %+v", out)
	}
}

func TestCreatePracticePlanUsesIdempotencyReplay(t *testing.T) {
	service := &fakePlanService{record: fixturePlanRecord()}
	handler := newTestHandler(service)
	store := newRouteMemoryStore()
	mw := idempotency.New(idempotency.MiddlewareOptions{
		Store: store,
		Now:   func() time.Time { return time.Date(2026, 5, 9, 12, 0, 0, 0, time.UTC) },
		NewID: func() string { return "idempotency-record-1" },
	})
	route := mw.Handler("practice", "createPracticePlan", userFromRequestContext, http.HandlerFunc(handler.CreatePracticePlan))

	reqBody := fixtureCreatePlanRequest()
	first := httptest.NewRecorder()
	route.ServeHTTP(first, newCreatePlanHTTPRequest(t, reqBody))
	second := httptest.NewRecorder()
	route.ServeHTTP(second, newCreatePlanHTTPRequest(t, reqBody))

	if first.Code != http.StatusCreated || second.Code != http.StatusCreated {
		t.Fatalf("unexpected statuses: first=%d second=%d secondBody=%s", first.Code, second.Code, second.Body.String())
	}
	if second.Header().Get(idempotency.ReplayHeader) != "true" {
		t.Fatalf("expected idempotency replay header on second response")
	}
	if service.calls != 1 {
		t.Fatalf("idempotency replay should call service once, got %d", service.calls)
	}
}

func TestCreatePracticePlanIdempotencyReplayPreservesDerivedSource(t *testing.T) {
	sourceDebriefID := "01918fa0-0000-7000-8000-00000000d001"
	service := &fakePlanService{
		record: func() domain.PlanRecord {
			plan := fixturePlanRecord()
			plan.Goal = sharedtypes.PracticeGoalDebrief
			plan.SourceDebriefID = sourceDebriefID
			return plan
		}(),
	}
	handler := newTestHandler(service)
	store := newRouteMemoryStore()
	mw := idempotency.New(idempotency.MiddlewareOptions{
		Store: store,
		Now:   func() time.Time { return time.Date(2026, 5, 9, 12, 0, 0, 0, time.UTC) },
		NewID: func() string { return "idempotency-record-1" },
	})
	route := mw.Handler("practice", "createPracticePlan", userFromRequestContext, http.HandlerFunc(handler.CreatePracticePlan))

	reqBody := fixtureCreatePlanRequest()
	reqBody.Goal = sharedtypes.PracticeGoalDebrief
	reqBody.SourceDebriefId = &sourceDebriefID
	first := httptest.NewRecorder()
	route.ServeHTTP(first, newCreatePlanHTTPRequest(t, reqBody))
	second := httptest.NewRecorder()
	route.ServeHTTP(second, newCreatePlanHTTPRequest(t, reqBody))

	if first.Code != http.StatusCreated || second.Code != http.StatusCreated {
		t.Fatalf("unexpected statuses: first=%d second=%d secondBody=%s", first.Code, second.Code, second.Body.String())
	}
	if second.Header().Get(idempotency.ReplayHeader) != "true" {
		t.Fatalf("expected idempotency replay header on second response")
	}
	var out api.PracticePlan
	if err := json.Unmarshal(second.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode replay response: %v", err)
	}
	if out.SourceDebriefId == nil || *out.SourceDebriefId != sourceDebriefID {
		t.Fatalf("replay lost sourceDebriefId: %+v", out)
	}
	if service.calls != 1 {
		t.Fatalf("idempotency replay should call service once, got %d", service.calls)
	}
}

func TestCreatePracticePlanIdempotencyRejectsDifferentSourceWithSameKey(t *testing.T) {
	sourceDebriefID := "01918fa0-0000-7000-8000-00000000d001"
	otherDebriefID := "01918fa0-0000-7000-8000-00000000d002"
	service := &fakePlanService{
		record: func() domain.PlanRecord {
			plan := fixturePlanRecord()
			plan.Goal = sharedtypes.PracticeGoalDebrief
			plan.SourceDebriefID = sourceDebriefID
			return plan
		}(),
	}
	handler := newTestHandler(service)
	store := newRouteMemoryStore()
	mw := idempotency.New(idempotency.MiddlewareOptions{
		Store: store,
		Now:   func() time.Time { return time.Date(2026, 5, 9, 12, 0, 0, 0, time.UTC) },
		NewID: func() string { return "idempotency-record-1" },
	})
	route := mw.Handler("practice", "createPracticePlan", userFromRequestContext, http.HandlerFunc(handler.CreatePracticePlan))

	reqBody := fixtureCreatePlanRequest()
	reqBody.Goal = sharedtypes.PracticeGoalDebrief
	reqBody.SourceDebriefId = &sourceDebriefID
	first := httptest.NewRecorder()
	route.ServeHTTP(first, newCreatePlanHTTPRequest(t, reqBody))
	reqBody.SourceDebriefId = &otherDebriefID
	second := httptest.NewRecorder()
	route.ServeHTTP(second, newCreatePlanHTTPRequest(t, reqBody))

	if first.Code != http.StatusCreated || second.Code != http.StatusConflict {
		t.Fatalf("unexpected statuses: first=%d second=%d secondBody=%s", first.Code, second.Code, second.Body.String())
	}
	assertAPIError(t, second, sharederrors.CodePracticeSessionConflict, false)
	if service.calls != 1 {
		t.Fatalf("fingerprint mismatch should not call service again, got %d", service.calls)
	}
}

func TestCreatePracticePlanFixtureParityDefault(t *testing.T) {
	fixture := loadCreatePlanFixture(t)
	defaultScenario := fixture.Scenarios["default"]
	service := &fakePlanService{record: planRecordFromFixture(defaultScenario.Response.Body)}
	handler := newTestHandler(service)

	req := newCreatePlanHTTPRequest(t, defaultScenario.Request.Body)
	rec := httptest.NewRecorder()
	handler.CreatePracticePlan(rec, req)
	if rec.Code != defaultScenario.Response.Status {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, defaultScenario.Response.Status, rec.Body.String())
	}
	assertJSONEqual(t, mustJSON(t, defaultScenario.Response.Body), rec.Body.Bytes())
}

func newTestHandler(service planService) *Handler {
	return newTestHandlerWithPepper(service, "")
}

func newTestHandlerWithPepper(service planService, pepper string) *Handler {
	return NewHandler(HandlerOptions{
		Service: service,
		Session: func(ctx context.Context) (string, bool) {
			if userID, ok := ctx.Value(testUserKey{}).(string); ok && strings.TrimSpace(userID) != "" {
				return userID, true
			}
			return "", false
		},
		IdempotencyKeyPepper: pepper,
	})
}

type fakePlanService struct {
	mu                sync.Mutex
	record            domain.PlanRecord
	err               error
	last              domain.CreatePlanRequest
	calls             int
	getRecord         domain.PlanRecord
	getErr            error
	getUserID         string
	getPlanID         string
	getSessionRecord  domain.SessionRecord
	getSessionErr     error
	getSessionUserID  string
	getSessionID      string
	startRecord       domain.SessionRecord
	startErr          error
	startUserID       string
	startPlanID       string
	startHintsEnabled bool
	startKeyHash      string
	startFingerprint  string
	appendResult      domain.AppendSessionEventResult
	appendErr         error
	appendRequest     domain.AppendSessionEventRequest
	completeResult    domain.CompleteSessionResult
	completeErr       error
	completeRequest   domain.CompletePracticeSessionRequest
}

func (s *fakePlanService) CreatePracticePlan(ctx context.Context, in domain.CreatePlanRequest) (domain.PlanRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calls++
	s.last = in
	if s.err != nil {
		return domain.PlanRecord{}, s.err
	}
	return s.record, nil
}

func (s *fakePlanService) GetPracticePlan(ctx context.Context, userID, planID string) (domain.PlanRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.getUserID = userID
	s.getPlanID = planID
	if s.getErr != nil {
		return domain.PlanRecord{}, s.getErr
	}
	return s.getRecord, nil
}

func (s *fakePlanService) GetPracticeSession(ctx context.Context, userID, sessionID string) (domain.SessionRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.getSessionUserID = userID
	s.getSessionID = sessionID
	if s.getSessionErr != nil {
		return domain.SessionRecord{}, s.getSessionErr
	}
	return s.getSessionRecord, nil
}

func (s *fakePlanService) StartPracticeSession(ctx context.Context, in domain.StartSessionRequest) (domain.SessionRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.startUserID = in.UserID
	s.startPlanID = in.PlanID
	s.startHintsEnabled = in.HintsEnabled
	s.startKeyHash = in.IdempotencyKeyHash
	s.startFingerprint = in.RequestFingerprint
	if s.startErr != nil {
		return domain.SessionRecord{}, s.startErr
	}
	return s.startRecord, nil
}

func (s *fakePlanService) AppendSessionEvent(ctx context.Context, in domain.AppendSessionEventRequest) (domain.AppendSessionEventResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.appendRequest = in
	if s.appendErr != nil {
		return domain.AppendSessionEventResult{}, s.appendErr
	}
	return s.appendResult, nil
}

func (s *fakePlanService) CompletePracticeSession(ctx context.Context, in domain.CompletePracticeSessionRequest) (domain.CompleteSessionResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.completeRequest = in
	if s.completeErr != nil {
		return domain.CompleteSessionResult{}, s.completeErr
	}
	return s.completeResult, nil
}

type testUserKey struct{}

func userFromRequestContext(r *http.Request) (string, bool) {
	if userID, ok := r.Context().Value(testUserKey{}).(string); ok && strings.TrimSpace(userID) != "" {
		return userID, true
	}
	return "", false
}

func contextWithUser(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, testUserKey{}, userID)
}

func newCreatePlanHTTPRequest(t *testing.T, body api.CreatePracticePlanRequest) *http.Request {
	t.Helper()
	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/practice/plans", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(idempotency.HeaderName, "plan-key-1")
	return req.WithContext(contextWithUser(req.Context(), "user-1"))
}

func fixtureCreatePlanRequest() api.CreatePracticePlanRequest {
	return api.CreatePracticePlanRequest{
		TargetJobId:          "01918fa0-0000-7000-8000-000000002000",
		Goal:                 sharedtypes.PracticeGoalBaseline,
		Mode:                 sharedtypes.PracticeModeAssisted,
		InterviewerPersona:   sharedtypes.InterviewerRoleHiringManager,
		Difficulty:           "standard",
		Language:             "zh-CN",
		QuestionBudget:       6,
		TimeBudgetMinutes:    30,
		ResumeAssetId:        "01918fa0-0000-7000-8000-000000001000",
		FocusCompetencyCodes: []string{"communication", "design-systems"},
	}
}

func fixturePlanRecord() domain.PlanRecord {
	return domain.PlanRecord{
		ID:                 "01918fa0-0000-7000-8000-000000004000",
		TargetJobID:        "01918fa0-0000-7000-8000-000000002000",
		Goal:               sharedtypes.PracticeGoalBaseline,
		Mode:               sharedtypes.PracticeModeAssisted,
		InterviewerPersona: sharedtypes.InterviewerRoleHiringManager,
		Difficulty:         "standard",
		Language:           "zh-CN",
		TimeBudgetMinutes:  30,
		QuestionBudget:     6,
		Status:             "ready",
		CreatedAt:          time.Date(2026, 4, 28, 12, 0, 0, 0, time.UTC),
	}
}

type createPlanFixture struct {
	Scenarios map[string]struct {
		Request struct {
			Body api.CreatePracticePlanRequest `json:"body"`
		} `json:"request"`
		Response struct {
			Status int              `json:"status"`
			Body   api.PracticePlan `json:"body"`
		} `json:"response"`
	} `json:"scenarios"`
}

func loadCreatePlanFixture(t *testing.T) createPlanFixture {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("..", "..", "..", "..", "openapi", "fixtures", "PracticePlans", "createPracticePlan.json"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	var fixture createPlanFixture
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatalf("decode fixture: %v", err)
	}
	return fixture
}

func planRecordFromFixture(plan api.PracticePlan) domain.PlanRecord {
	createdAt, _ := time.Parse(time.RFC3339, plan.CreatedAt)
	return domain.PlanRecord{
		ID:                 plan.Id,
		TargetJobID:        plan.TargetJobId,
		SourceReportID:     stringValue(plan.SourceReportId),
		SourceDebriefID:    stringValue(plan.SourceDebriefId),
		Goal:               plan.Goal,
		Mode:               plan.Mode,
		InterviewerPersona: plan.InterviewerPersona,
		Difficulty:         plan.Difficulty,
		Language:           plan.Language,
		TimeBudgetMinutes:  plan.TimeBudgetMinutes,
		QuestionBudget:     plan.QuestionBudget,
		Status:             plan.Status,
		CreatedAt:          createdAt,
	}
}

func assertAPIError(t *testing.T, rec *httptest.ResponseRecorder, wantCode string, wantRetryable bool) {
	t.Helper()
	var out api.ApiErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode error response: %v; body=%s", err, rec.Body.String())
	}
	if out.Error.Code != wantCode || out.Error.Retryable != wantRetryable {
		t.Fatalf("error = %+v, want code=%s retryable=%v", out.Error, wantCode, wantRetryable)
	}
}

func mustJSON(t *testing.T, value any) []byte {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal json: %v", err)
	}
	return raw
}

func assertJSONEqual(t *testing.T, want []byte, got []byte) {
	t.Helper()
	var wantValue any
	var gotValue any
	if err := json.Unmarshal(want, &wantValue); err != nil {
		t.Fatalf("decode want: %v", err)
	}
	if err := json.Unmarshal(got, &gotValue); err != nil {
		t.Fatalf("decode got: %v; body=%s", err, string(got))
	}
	wantRaw, _ := json.Marshal(wantValue)
	gotRaw, _ := json.Marshal(gotValue)
	if !bytes.Equal(wantRaw, gotRaw) {
		t.Fatalf("json mismatch\nwant: %s\n got: %s", wantRaw, gotRaw)
	}
}

type routeMemoryStore struct {
	mu      sync.Mutex
	records map[string]routeMemoryRecord
}

type routeMemoryRecord struct {
	recordID    string
	fingerprint string
	status      idempotency.Status
	response    []byte
	httpStatus  int
	resourceID  string
	expiresAt   time.Time
}

func newRouteMemoryStore() *routeMemoryStore {
	return &routeMemoryStore{records: map[string]routeMemoryRecord{}}
}

func (s *routeMemoryStore) Reserve(ctx context.Context, in idempotency.ReservationInput) (idempotency.Reservation, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := strings.Join([]string{in.UserID, in.Domain, in.Operation, in.IdempotencyKeyHash}, "\x00")
	rec, ok := s.records[key]
	if ok && !in.Now.Before(rec.expiresAt) {
		ok = false
	}
	if !ok {
		s.records[key] = routeMemoryRecord{
			recordID:    in.RecordID,
			fingerprint: in.RequestFingerprint,
			status:      idempotency.StatusPending,
			expiresAt:   in.ExpiresAt,
		}
		return idempotency.Reservation{State: idempotency.StateExecute, RecordID: in.RecordID}, nil
	}
	if rec.status == idempotency.StatusFailedTerminal {
		rec.fingerprint = in.RequestFingerprint
		rec.status = idempotency.StatusPending
		rec.response = nil
		rec.httpStatus = 0
		rec.resourceID = ""
		rec.expiresAt = in.ExpiresAt
		s.records[key] = rec
		return idempotency.Reservation{State: idempotency.StateExecute, RecordID: rec.recordID}, nil
	}
	if rec.fingerprint != in.RequestFingerprint {
		return idempotency.Reservation{}, idempotency.ErrFingerprintMismatch
	}
	if rec.status == idempotency.StatusSucceeded {
		return idempotency.Reservation{
			State:          idempotency.StateReplay,
			RecordID:       rec.recordID,
			ResponseBody:   append([]byte(nil), rec.response...),
			ResponseStatus: rec.httpStatus,
		}, nil
	}
	return idempotency.Reservation{}, idempotency.ErrPending
}

func (s *routeMemoryStore) MarkSucceeded(ctx context.Context, in idempotency.CompletionInput) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for key, rec := range s.records {
		if rec.recordID == in.RecordID {
			rec.status = idempotency.StatusSucceeded
			rec.response = append([]byte(nil), in.ResponseBody...)
			rec.httpStatus = in.ResponseStatus
			rec.resourceID = in.ResourceID
			s.records[key] = rec
			return nil
		}
	}
	return idempotency.ErrReservationNotFound
}

func (s *routeMemoryStore) MarkFailed(ctx context.Context, in idempotency.CompletionInput) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for key, rec := range s.records {
		if rec.recordID == in.RecordID {
			rec.status = idempotency.StatusFailedTerminal
			rec.response = append([]byte(nil), in.ResponseBody...)
			rec.httpStatus = in.ResponseStatus
			s.records[key] = rec
			return nil
		}
	}
	return idempotency.ErrReservationNotFound
}
