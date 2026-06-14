package debriefs

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	api "github.com/monshunter/easyinterview/backend/internal/api/generated"
	domain "github.com/monshunter/easyinterview/backend/internal/debrief"
	"github.com/monshunter/easyinterview/backend/internal/middleware/idempotency"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

func TestHandlerPackageCompiles(t *testing.T) {
	t.Helper()
	if NewHandler(HandlerOptions{}) == nil {
		t.Fatalf("NewHandler returned nil")
	}
}

func TestCreateDebrief_ValidationError_EmptyQuestions(t *testing.T) {
	service := &fakeDebriefService{}
	handler := NewHandler(HandlerOptions{Service: service, Session: staticSession("user-1")})
	rec := httptest.NewRecorder()

	handler.CreateDebrief(rec, newCreateDebriefRequest(t, api.CreateDebriefRequest{
		TargetJobId: "01918fa0-0000-7000-8000-00000000d001",
		RoundType:   sharedtypes.DebriefRoundTypeBehavioral,
		Language:    "zh-CN",
		Questions:   []api.DebriefQuestionInput{},
	}))

	assertAPIError(t, rec, http.StatusUnprocessableEntity, sharederrors.CodeValidationFailed)
	if service.calls != 0 {
		t.Fatalf("validation failure should not call service, calls=%d", service.calls)
	}
}

func TestCreateDebrief_ValidationError_LongQuestionText(t *testing.T) {
	service := &fakeDebriefService{}
	handler := NewHandler(HandlerOptions{Service: service, Session: staticSession("user-1")})
	rec := httptest.NewRecorder()

	handler.CreateDebrief(rec, newCreateDebriefRequest(t, api.CreateDebriefRequest{
		TargetJobId: "01918fa0-0000-7000-8000-00000000d001",
		RoundType:   sharedtypes.DebriefRoundTypeBehavioral,
		Language:    "zh-CN",
		Questions: []api.DebriefQuestionInput{{
			QuestionText:    strings.Repeat("x", 4001),
			MyAnswerSummary: "summary",
		}},
	}))

	assertAPIError(t, rec, http.StatusUnprocessableEntity, sharederrors.CodeValidationFailed)
	if service.calls != 0 {
		t.Fatalf("validation failure should not call service, calls=%d", service.calls)
	}
}

func TestCreateDebrief_HappyResponse(t *testing.T) {
	now := time.Date(2026, 5, 16, 9, 0, 0, 0, time.UTC)
	service := &fakeDebriefService{result: fixtureCreateDebriefResult(now)}
	handler := NewHandler(HandlerOptions{Service: service, Session: staticSession("user-1")})
	rec := httptest.NewRecorder()
	interviewerRole := sharedtypes.InterviewerRoleHiringManager

	handler.CreateDebrief(rec, newCreateDebriefRequest(t, api.CreateDebriefRequest{
		TargetJobId:     "01918fa0-0000-7000-8000-00000000d001",
		RoundType:       sharedtypes.DebriefRoundTypeBehavioral,
		InterviewerRole: &interviewerRole,
		Language:        "zh-CN",
		Questions: []api.DebriefQuestionInput{{
			QuestionText:    "Tell me about a cross-functional project.",
			MyAnswerSummary: "I described a design-system migration.",
		}},
	}))

	if rec.Code != http.StatusAccepted {
		t.Fatalf("status: want %d, got %d body=%s", http.StatusAccepted, rec.Code, rec.Body.String())
	}
	var out api.DebriefWithJob
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode DebriefWithJob: %v", err)
	}
	if out.DebriefId != service.result.DebriefID || out.Job.JobType != api.JobTypeDebriefGenerate || out.Job.ResourceType != api.ResourceTypeDebrief {
		t.Fatalf("unexpected debrief response: %+v", out)
	}
	if service.calls != 1 || service.last.UserID != "user-1" || service.last.InterviewerRole != interviewerRole {
		t.Fatalf("service mapping drifted: calls=%d last=%+v", service.calls, service.last)
	}
}

func TestCreateDebrief_IdempotencyEnabled(t *testing.T) {
	now := time.Date(2026, 5, 16, 9, 0, 0, 0, time.UTC)
	service := &fakeDebriefService{result: fixtureCreateDebriefResult(now)}
	route, store := newIdempotentCreateDebriefRoute(service, now)
	body := validCreateDebriefRequest()

	first := httptest.NewRecorder()
	route.ServeHTTP(first, newIdempotentCreateDebriefRequest(t, body))
	second := httptest.NewRecorder()
	route.ServeHTTP(second, newIdempotentCreateDebriefRequest(t, body))

	if first.Code != http.StatusAccepted || second.Code != http.StatusAccepted {
		t.Fatalf("unexpected statuses: first=%d second=%d secondBody=%s", first.Code, second.Code, second.Body.String())
	}
	if second.Header().Get(idempotency.ReplayHeader) != "true" {
		t.Fatalf("expected idempotency replay header on second response")
	}
	if service.calls != 1 {
		t.Fatalf("idempotency replay should call service once, got %d", service.calls)
	}
	rec := store.singleRecord(t)
	if rec.resourceType != string(api.ResourceTypeDebrief) || rec.resourceID != service.result.DebriefID {
		t.Fatalf("idempotency resource drifted: %+v", rec)
	}
	if first.Header().Get("X-Idempotency-Resource-Type") != "" || first.Header().Get("X-Idempotency-Resource-ID") != "" {
		t.Fatalf("internal idempotency resource headers leaked to client: %+v", first.Header())
	}
}

func TestCreateDebrief_IdempotencyReplay_SameBody(t *testing.T) {
	now := time.Date(2026, 5, 16, 9, 0, 0, 0, time.UTC)
	service := &fakeDebriefService{result: fixtureCreateDebriefResult(now)}
	route, _ := newIdempotentCreateDebriefRoute(service, now)
	body := validCreateDebriefRequest()

	first := httptest.NewRecorder()
	route.ServeHTTP(first, newIdempotentCreateDebriefRequest(t, body))
	second := httptest.NewRecorder()
	route.ServeHTTP(second, newIdempotentCreateDebriefRequest(t, body))

	if first.Code != http.StatusAccepted || second.Code != http.StatusAccepted {
		t.Fatalf("unexpected statuses: first=%d second=%d secondBody=%s", first.Code, second.Code, second.Body.String())
	}
	if second.Header().Get(idempotency.ReplayHeader) != "true" {
		t.Fatalf("expected idempotency replay header on second response")
	}
	if second.Body.String() != first.Body.String() {
		t.Fatalf("replay body mismatch: first=%s second=%s", first.Body.String(), second.Body.String())
	}
	if service.calls != 1 {
		t.Fatalf("idempotency replay should call service once, got %d", service.calls)
	}
}

func TestCreateDebrief_IdempotencyMismatch_DifferentBody(t *testing.T) {
	now := time.Date(2026, 5, 16, 9, 0, 0, 0, time.UTC)
	service := &fakeDebriefService{result: fixtureCreateDebriefResult(now)}
	route, _ := newIdempotentCreateDebriefRoute(service, now)
	body := validCreateDebriefRequest()

	first := httptest.NewRecorder()
	route.ServeHTTP(first, newIdempotentCreateDebriefRequest(t, body))
	body.Language = "en"
	second := httptest.NewRecorder()
	route.ServeHTTP(second, newIdempotentCreateDebriefRequest(t, body))

	if first.Code != http.StatusAccepted || second.Code != http.StatusConflict {
		t.Fatalf("unexpected statuses: first=%d second=%d secondBody=%s", first.Code, second.Code, second.Body.String())
	}
	assertAPIError(t, second, http.StatusConflict, sharederrors.CodeIdempotencyKeyMismatch)
	if service.calls != 1 {
		t.Fatalf("fingerprint mismatch should not call service again, got %d", service.calls)
	}
}

func TestSuggestDebriefQuestions_CountBoundary(t *testing.T) {
	cases := []struct {
		name       string
		count      *int32
		wantStatus int
		wantCalls  int
		wantCount  int32
	}{
		{name: "default", count: nil, wantStatus: http.StatusOK, wantCalls: 1, wantCount: 6},
		{name: "min", count: int32Ptr(1), wantStatus: http.StatusOK, wantCalls: 1, wantCount: 1},
		{name: "max", count: int32Ptr(10), wantStatus: http.StatusOK, wantCalls: 1, wantCount: 10},
		{name: "zero", count: int32Ptr(0), wantStatus: http.StatusUnprocessableEntity},
		{name: "too high", count: int32Ptr(11), wantStatus: http.StatusUnprocessableEntity},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			service := &fakeDebriefService{suggestResult: domain.SuggestQuestionsResult{Suggestions: []domain.SuggestedQuestion{{
				QuestionText:   "How did you measure adoption?",
				WhyLikelyAsked: "The JD stresses metrics.",
				Source:         sharedtypes.DebriefQuestionSourceJd,
			}}}}
			handler := NewHandler(HandlerOptions{Service: service, Session: staticSession("user-1")})
			rec := httptest.NewRecorder()

			handler.SuggestDebriefQuestions(rec, newSuggestDebriefQuestionsRequest(t, api.SuggestDebriefQuestionsRequest{
				TargetJobId: "01918fa0-0000-7000-8000-00000000d001",
				Language:    "zh-CN",
				Count:       tc.count,
			}))

			if rec.Code != tc.wantStatus {
				t.Fatalf("status: want %d, got %d body=%s", tc.wantStatus, rec.Code, rec.Body.String())
			}
			if service.suggestCalls != tc.wantCalls {
				t.Fatalf("service calls: want %d, got %d", tc.wantCalls, service.suggestCalls)
			}
			if tc.wantStatus == http.StatusOK {
				if service.lastSuggest.Count != tc.wantCount {
					t.Fatalf("count mapped to service: want %d, got %+v", tc.wantCount, service.lastSuggest)
				}
				var out api.SuggestDebriefQuestionsResponse
				if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
					t.Fatalf("decode response: %v", err)
				}
				if len(out.Suggestions) != 1 || out.Suggestions[0].Source != sharedtypes.DebriefQuestionSourceJd {
					t.Fatalf("unexpected response: %+v", out)
				}
			}
		})
	}
}

func TestSuggestDebriefQuestions_MapsResumeIDToService(t *testing.T) {
	service := &fakeDebriefService{suggestResult: domain.SuggestQuestionsResult{Suggestions: []domain.SuggestedQuestion{{
		QuestionText:   "How did the platform migration change your scope?",
		WhyLikelyAsked: "The resume profile highlights platform ownership.",
		Source:         sharedtypes.DebriefQuestionSourceResume,
	}}}}
	handler := NewHandler(HandlerOptions{Service: service, Session: staticSession("user-1")})
	rec := httptest.NewRecorder()
	resumeID := "01918fa0-0000-7000-8000-00000000a001"

	handler.SuggestDebriefQuestions(rec, newSuggestDebriefQuestionsRequest(t, api.SuggestDebriefQuestionsRequest{
		TargetJobId: "01918fa0-0000-7000-8000-00000000d001",
		ResumeId:    &resumeID,
		Language:    "zh-CN",
	}))

	if rec.Code != http.StatusOK {
		t.Fatalf("status: want %d, got %d body=%s", http.StatusOK, rec.Code, rec.Body.String())
	}
	if service.suggestCalls != 1 || service.lastSuggest.ResumeID != resumeID {
		t.Fatalf("resumeId not mapped to service: calls=%d last=%+v", service.suggestCalls, service.lastSuggest)
	}
}

func TestSuggestDebriefQuestions_MapsSessionIDToService(t *testing.T) {
	service := &fakeDebriefService{suggestResult: domain.SuggestQuestionsResult{Suggestions: []domain.SuggestedQuestion{{
		QuestionText:   "How did you measure adoption?",
		WhyLikelyAsked: "The mock report highlights a metrics gap.",
		Source:         sharedtypes.DebriefQuestionSourceMockReport,
	}}}}
	handler := NewHandler(HandlerOptions{Service: service, Session: staticSession("user-1")})
	rec := httptest.NewRecorder()
	sessionID := "01918fa0-0000-7000-8000-000000005000"

	handler.SuggestDebriefQuestions(rec, newSuggestDebriefQuestionsRequest(t, api.SuggestDebriefQuestionsRequest{
		TargetJobId: "01918fa0-0000-7000-8000-00000000d001",
		SessionId:   &sessionID,
		Language:    "zh-CN",
	}))

	if rec.Code != http.StatusOK {
		t.Fatalf("status: want %d, got %d body=%s", http.StatusOK, rec.Code, rec.Body.String())
	}
	if service.suggestCalls != 1 || service.lastSuggest.SessionID != sessionID {
		t.Fatalf("sessionId not mapped to service: calls=%d last=%+v", service.suggestCalls, service.lastSuggest)
	}
}

func TestSuggestDebriefQuestions_Unauthenticated_401(t *testing.T) {
	service := &fakeDebriefService{}
	handler := NewHandler(HandlerOptions{Service: service})
	rec := httptest.NewRecorder()

	handler.SuggestDebriefQuestions(rec, newSuggestDebriefQuestionsRequest(t, api.SuggestDebriefQuestionsRequest{
		TargetJobId: "01918fa0-0000-7000-8000-00000000d001",
		Language:    "zh-CN",
		Count:       int32Ptr(6),
	}))

	assertAPIError(t, rec, http.StatusUnauthorized, sharederrors.CodeAuthUnauthorized)
	if service.suggestCalls != 0 {
		t.Fatalf("unauthenticated request should not call service, calls=%d", service.suggestCalls)
	}
}

func TestGetDebrief_DraftResponse(t *testing.T) {
	now := time.Date(2026, 5, 16, 16, 0, 0, 0, time.UTC)
	service := &fakeDebriefService{getResult: fixtureDraftDebriefRecord(now)}
	handler := NewHandler(HandlerOptions{Service: service, Session: staticSession("user-1")})
	rec := httptest.NewRecorder()

	handler.GetDebrief(rec, httptest.NewRequest(http.MethodGet, "/api/v1/debriefs/debrief-1", nil), "debrief-1")

	if rec.Code != http.StatusOK {
		t.Fatalf("status: want %d, got %d body=%s", http.StatusOK, rec.Code, rec.Body.String())
	}
	var out api.Debrief
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode Debrief: %v", err)
	}
	if out.Id != service.getResult.ID ||
		out.Status != sharedtypes.DebriefStatusDraft ||
		out.Provenance != nil ||
		len(out.RiskItems) != 0 ||
		len(out.Questions) != 1 ||
		out.Questions[0].AiAnalysis != nil {
		t.Fatalf("draft response drifted: %+v", out)
	}
	if service.getCalls != 1 || service.getUserID != "user-1" || service.getDebriefID != "debrief-1" {
		t.Fatalf("service get mapping drifted: calls=%d user=%s debrief=%s", service.getCalls, service.getUserID, service.getDebriefID)
	}
}

func TestGetDebrief_CompletedResponse(t *testing.T) {
	now := time.Date(2026, 5, 16, 16, 0, 0, 0, time.UTC)
	service := &fakeDebriefService{getResult: fixtureCompletedDebriefRecord(now)}
	handler := NewHandler(HandlerOptions{Service: service, Session: staticSession("user-1")})
	rec := httptest.NewRecorder()

	handler.GetDebrief(rec, httptest.NewRequest(http.MethodGet, "/api/v1/debriefs/debrief-1", nil), "debrief-1")

	if rec.Code != http.StatusOK {
		t.Fatalf("status: want %d, got %d body=%s", http.StatusOK, rec.Code, rec.Body.String())
	}
	var out api.Debrief
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode Debrief: %v", err)
	}
	if out.Id != service.getResult.ID ||
		out.Status != sharedtypes.DebriefStatusCompleted ||
		out.Provenance == nil ||
		out.Provenance.DataSourceVersion != "debrief/debrief-1@v1" ||
		len(out.RiskItems) != 1 ||
		len(out.NextRoundChecklist) != 1 ||
		out.ThankYouDraft == nil ||
		len(out.Questions) != 1 ||
		out.Questions[0].AiAnalysis == nil {
		t.Fatalf("completed response drifted: %+v", out)
	}
}

func TestGetDebrief_CrossUser404(t *testing.T) {
	service := &fakeDebriefService{getErr: domain.ErrDebriefNotFound}
	handler := NewHandler(HandlerOptions{Service: service, Session: staticSession("user-1")})
	rec := httptest.NewRecorder()

	handler.GetDebrief(rec, httptest.NewRequest(http.MethodGet, "/api/v1/debriefs/debrief-1", nil), "debrief-1")

	assertAPIError(t, rec, http.StatusNotFound, sharederrors.CodeDebriefNotFound)
}

func TestGetDebrief_NotFound404(t *testing.T) {
	service := &fakeDebriefService{getErr: domain.ErrDebriefNotFound}
	handler := NewHandler(HandlerOptions{Service: service, Session: staticSession("user-1")})
	rec := httptest.NewRecorder()

	handler.GetDebrief(rec, httptest.NewRequest(http.MethodGet, "/api/v1/debriefs/debrief-missing", nil), "debrief-missing")

	assertAPIError(t, rec, http.StatusNotFound, sharederrors.CodeDebriefNotFound)
}

func fixtureCreateDebriefResult(now time.Time) domain.CreateDebriefResult {
	return domain.CreateDebriefResult{
		DebriefID: "01918fa0-0000-7000-8000-00000000d010",
		Job: domain.JobRecord{
			ID:           "01918fa0-0000-7000-8000-00000000d011",
			JobType:      api.JobTypeDebriefGenerate,
			ResourceType: api.ResourceTypeDebrief,
			ResourceID:   "01918fa0-0000-7000-8000-00000000d010",
			Status:       sharedtypes.JobStatusQueued,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
	}
}

func fixtureDraftDebriefRecord(now time.Time) domain.DebriefRecord {
	return domain.DebriefRecord{
		ID:              "debrief-1",
		TargetJobID:     "target-1",
		Status:          sharedtypes.DebriefStatusDraft,
		RoundType:       sharedtypes.DebriefRoundTypeBehavioral,
		InterviewerRole: string(sharedtypes.InterviewerRoleHiringManager),
		Questions: []domain.QuestionRecord{{
			QuestionText:        "Tell me about scope.",
			MyAnswerSummary:     "I explained ownership.",
			InterviewerReaction: "Asked for metrics.",
		}},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func fixtureCompletedDebriefRecord(now time.Time) domain.DebriefRecord {
	analysis := fixtureDraftDebriefRecord(now)
	analysis.Status = sharedtypes.DebriefStatusCompleted
	analysis.Questions[0].AIAnalysis = "Add outcome numbers."
	analysis.RiskItems = []domain.RiskItem{{Label: "Metrics missing", Severity: "medium"}}
	analysis.NextRoundChecklist = []domain.NextRoundChecklistItem{{Label: "Prepare launch metrics", Rationale: "Interviewer asked follow-up."}}
	analysis.ThankYouDraft = "Thanks for the conversation."
	analysis.Provenance = &domain.Provenance{
		PromptVersion:     "v0.1.0",
		RubricVersion:     "v0.1.0",
		ModelID:           "stub-model",
		Language:          "zh-CN",
		FeatureFlag:       "none",
		DataSourceVersion: "debrief/debrief-1@v1",
	}
	return analysis
}

func newIdempotentCreateDebriefRoute(service *fakeDebriefService, now time.Time) (http.Handler, *debriefIdempotencyStore) {
	handler := NewHandler(HandlerOptions{Service: service, Session: staticSession("user-1")})
	store := newDebriefIdempotencyStore()
	mw := idempotency.New(idempotency.MiddlewareOptions{
		Store: store,
		Now:   func() time.Time { return now },
		NewID: func() string { return "idempotency-record-1" },
	})
	return mw.Handler("debrief", "createDebrief", staticRequestUser("user-1"), http.HandlerFunc(handler.CreateDebrief)), store
}

func newCreateDebriefRequest(t *testing.T, body api.CreateDebriefRequest) *http.Request {
	t.Helper()
	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/debriefs", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	return req
}

func newIdempotentCreateDebriefRequest(t *testing.T, body api.CreateDebriefRequest) *http.Request {
	t.Helper()
	req := newCreateDebriefRequest(t, body)
	req.Header.Set(idempotency.HeaderName, "debrief-idempotency-key-1")
	return req
}

func newSuggestDebriefQuestionsRequest(t *testing.T, body api.SuggestDebriefQuestionsRequest) *http.Request {
	t.Helper()
	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/debriefs/question-suggestions", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	return req
}

func int32Ptr(value int32) *int32 {
	return &value
}

func validCreateDebriefRequest() api.CreateDebriefRequest {
	return api.CreateDebriefRequest{
		TargetJobId: "01918fa0-0000-7000-8000-00000000d001",
		RoundType:   sharedtypes.DebriefRoundTypeBehavioral,
		Language:    "zh-CN",
		Questions: []api.DebriefQuestionInput{{
			QuestionText:    "Tell me about a cross-functional project.",
			MyAnswerSummary: "I described a design-system migration.",
		}},
	}
}

func staticSession(userID string) SessionResolver {
	return func(context.Context) (string, bool) { return userID, true }
}

func staticRequestUser(userID string) idempotency.UserIDResolver {
	return func(*http.Request) (string, bool) { return userID, true }
}

func assertAPIError(t *testing.T, rec *httptest.ResponseRecorder, wantStatus int, wantCode string) {
	t.Helper()
	if rec.Code != wantStatus {
		t.Fatalf("status: want %d, got %d body=%s", wantStatus, rec.Code, rec.Body.String())
	}
	var out api.ApiErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode error response: %v body=%s", err, rec.Body.String())
	}
	if out.Error.Code != wantCode {
		t.Fatalf("error code: want %s, got %+v", wantCode, out.Error)
	}
}

type fakeDebriefService struct {
	calls         int
	last          domain.CreateDebriefRequest
	result        domain.CreateDebriefResult
	suggestCalls  int
	lastSuggest   domain.SuggestQuestionsRequest
	suggestResult domain.SuggestQuestionsResult
	getCalls      int
	getUserID     string
	getDebriefID  string
	getResult     domain.DebriefRecord
	getErr        error
}

func (s *fakeDebriefService) CreateDebrief(_ context.Context, in domain.CreateDebriefRequest) (domain.CreateDebriefResult, error) {
	s.calls++
	s.last = in
	return s.result, nil
}

func (s *fakeDebriefService) SuggestQuestions(_ context.Context, in domain.SuggestQuestionsRequest) (domain.SuggestQuestionsResult, error) {
	s.suggestCalls++
	s.lastSuggest = in
	return s.suggestResult, nil
}

func (s *fakeDebriefService) GetDebrief(_ context.Context, userID string, debriefID string) (domain.DebriefRecord, error) {
	s.getCalls++
	s.getUserID = userID
	s.getDebriefID = debriefID
	if s.getErr != nil {
		return domain.DebriefRecord{}, s.getErr
	}
	return s.getResult, nil
}

type debriefIdempotencyStore struct {
	mu      sync.Mutex
	records map[string]debriefIdempotencyRecord
}

type debriefIdempotencyRecord struct {
	recordID     string
	fingerprint  string
	status       idempotency.Status
	response     []byte
	httpStatus   int
	resourceID   string
	resourceType string
	expiresAt    time.Time
}

func newDebriefIdempotencyStore() *debriefIdempotencyStore {
	return &debriefIdempotencyStore{records: map[string]debriefIdempotencyRecord{}}
}

func (s *debriefIdempotencyStore) Reserve(_ context.Context, in idempotency.ReservationInput) (idempotency.Reservation, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := strings.Join([]string{in.UserID, in.Domain, in.Operation, in.IdempotencyKeyHash}, "\x00")
	rec, ok := s.records[key]
	if ok && !in.Now.Before(rec.expiresAt) {
		ok = false
	}
	if !ok {
		s.records[key] = debriefIdempotencyRecord{
			recordID:    in.RecordID,
			fingerprint: in.RequestFingerprint,
			status:      idempotency.StatusPending,
			expiresAt:   in.ExpiresAt,
		}
		return idempotency.Reservation{State: idempotency.StateExecute, RecordID: in.RecordID}, nil
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
			ResourceType:   rec.resourceType,
			ResourceID:     rec.resourceID,
		}, nil
	}
	return idempotency.Reservation{}, idempotency.ErrPending
}

func (s *debriefIdempotencyStore) MarkSucceeded(_ context.Context, in idempotency.CompletionInput) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for key, rec := range s.records {
		if rec.recordID == in.RecordID {
			rec.status = idempotency.StatusSucceeded
			rec.response = append([]byte(nil), in.ResponseBody...)
			rec.httpStatus = in.ResponseStatus
			rec.resourceType = in.ResourceType
			rec.resourceID = in.ResourceID
			s.records[key] = rec
			return nil
		}
	}
	return idempotency.ErrReservationNotFound
}

func (s *debriefIdempotencyStore) MarkFailed(_ context.Context, in idempotency.CompletionInput) error {
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

func (s *debriefIdempotencyStore) singleRecord(t *testing.T) debriefIdempotencyRecord {
	t.Helper()
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.records) != 1 {
		t.Fatalf("expected exactly one idempotency record, got %d", len(s.records))
	}
	for _, rec := range s.records {
		return rec
	}
	t.Fatalf("idempotency record missing")
	return debriefIdempotencyRecord{}
}
