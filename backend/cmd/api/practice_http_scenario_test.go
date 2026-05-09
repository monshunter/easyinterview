package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient"
	"github.com/monshunter/easyinterview/backend/internal/ai/registry"
	api "github.com/monshunter/easyinterview/backend/internal/api/generated"
	apipractice "github.com/monshunter/easyinterview/backend/internal/api/practice"
	"github.com/monshunter/easyinterview/backend/internal/auth"
	"github.com/monshunter/easyinterview/backend/internal/middleware/idempotency"
	"github.com/monshunter/easyinterview/backend/internal/platform/config"
	domainpractice "github.com/monshunter/easyinterview/backend/internal/practice"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
	sharedevents "github.com/monshunter/easyinterview/backend/internal/shared/events"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
	storepractice "github.com/monshunter/easyinterview/backend/internal/store/practice"
	"github.com/monshunter/easyinterview/backend/internal/targetjob"
)

const (
	practiceHTTPScenarioUserAID = "scenario-user-practice-a"
	practiceHTTPScenarioUserBID = "scenario-user-practice-b"
)

func TestE2EP0022PracticePlanBaselineCreateAndRead(t *testing.T) {
	h := newPracticeHTTPScenarioHarness(t)
	body := api.CreatePracticePlanRequest{
		TargetJobId:          "target-job-p0-022-a",
		ResumeAssetId:        strPtr("resume-asset-p0-022-a"),
		Goal:                 sharedtypes.PracticeGoalBaseline,
		Mode:                 sharedtypes.PracticeModeAssisted,
		InterviewerPersona:   sharedtypes.InterviewerRoleHiringManager,
		Difficulty:           "standard",
		Language:             "zh-CN",
		TimeBudgetMinutes:    30,
		QuestionBudget:       6,
		FocusCompetencyCodes: []string{"communication", "design-systems"},
	}

	raw := h.doJSON(t, practiceHTTPScenarioUserAID, http.MethodPost, "/api/v1/practice/plans", "e2e-p0-022-create-plan", body, http.StatusCreated)
	var created api.PracticePlan
	decodeJSON(t, raw, &created)
	if created.Id == "" || created.Status != "ready" || created.Goal != sharedtypes.PracticeGoalBaseline || created.TargetJobId != body.TargetJobId {
		t.Fatalf("unexpected createPracticePlan response: %+v", created)
	}
	if h.store.planCount() != 1 || h.store.auditCount() != 1 {
		t.Fatalf("createPracticePlan should write one plan and one audit row, plans=%d audits=%d", h.store.planCount(), h.store.auditCount())
	}

	duplicateRaw := h.doJSON(t, practiceHTTPScenarioUserAID, http.MethodPost, "/api/v1/practice/plans", "e2e-p0-022-create-plan", body, http.StatusCreated)
	var duplicate api.PracticePlan
	decodeJSON(t, duplicateRaw, &duplicate)
	if duplicate.Id != created.Id || h.store.planCount() != 1 || h.store.auditCount() != 1 {
		t.Fatalf("idempotency replay should not duplicate side effects: duplicate=%+v plans=%d audits=%d", duplicate, h.store.planCount(), h.store.auditCount())
	}

	detailRaw := h.doJSON(t, practiceHTTPScenarioUserAID, http.MethodGet, "/api/v1/practice/plans/"+created.Id, "", nil, http.StatusOK)
	var detail api.PracticePlan
	decodeJSON(t, detailRaw, &detail)
	if detail.Id != created.Id || detail.Status != "ready" || detail.QuestionBudget != body.QuestionBudget {
		t.Fatalf("getPracticePlan did not return the created plan: %+v", detail)
	}

	crossUserRaw := h.doJSON(t, practiceHTTPScenarioUserBID, http.MethodGet, "/api/v1/practice/plans/"+created.Id, "", nil, http.StatusNotFound)
	var crossUser api.ApiErrorResponse
	decodeJSON(t, crossUserRaw, &crossUser)
	if crossUser.Error.Code != sharederrors.CodePracticePlanNotFound || crossUser.Error.Retryable {
		t.Fatalf("cross-user getPracticePlan should hide existence with PRACTICE_PLAN_NOT_FOUND: %+v", crossUser.Error)
	}

	assertNoEvidenceLeak(t, h.store.auditPayloads(), "question_text", "answer_text", "hint_text", "prompt body", "response body", "legacy debrief replay value")
}

func TestE2EP0023PracticeSessionStartAndFirstQuestion(t *testing.T) {
	h := newPracticeHTTPScenarioHarness(t)
	h.store.seedReadyPlan(domainpractice.CreatePlanStoreInput{
		PlanID:               "practice-plan-p0-023",
		AuditEventID:         "audit-p0-023",
		UserID:               practiceHTTPScenarioUserAID,
		TargetJobID:          "target-job-p0-023-a",
		ResumeAssetID:        "resume-asset-p0-023-a",
		Goal:                 sharedtypes.PracticeGoalBaseline,
		Mode:                 sharedtypes.PracticeModeAssisted,
		InterviewerPersona:   sharedtypes.InterviewerRoleHiringManager,
		Difficulty:           "standard",
		Language:             "zh-CN",
		TimeBudgetMinutes:    30,
		QuestionBudget:       6,
		FocusCompetencyCodes: []string{"system-design"},
		Now:                  fixedScenarioNow(),
	})

	raw := h.doJSON(t, practiceHTTPScenarioUserAID, http.MethodPost, "/api/v1/practice/sessions", "e2e-p0-023-start-session", api.StartPracticeSessionRequest{
		PlanId:       "practice-plan-p0-023",
		HintsEnabled: practiceBoolPtr(true),
	}, http.StatusCreated)
	var started api.PracticeSession
	decodeJSON(t, raw, &started)
	if started.Id == "" || started.Status != sharedtypes.SessionStatusRunning || started.CurrentTurn == nil {
		t.Fatalf("unexpected startPracticeSession response: %+v", started)
	}
	if started.CurrentTurn.TurnIndex != 1 || started.CurrentTurn.Status != "asked" || strings.TrimSpace(started.CurrentTurn.QuestionText) == "" {
		t.Fatalf("first turn was not synchronously returned: %+v", started.CurrentTurn)
	}
	if started.CurrentTurn.QuestionIntent == nil || *started.CurrentTurn.QuestionIntent != "behavioral.leadership.design_system" {
		t.Fatalf("first turn intent mismatch: %+v", started.CurrentTurn)
	}

	detailRaw := h.doJSON(t, practiceHTTPScenarioUserAID, http.MethodGet, "/api/v1/practice/sessions/"+started.Id, "", nil, http.StatusOK)
	var detail api.PracticeSession
	decodeJSON(t, detailRaw, &detail)
	if detail.Id != started.Id || detail.CurrentTurn == nil || detail.CurrentTurn.Id != started.CurrentTurn.Id {
		t.Fatalf("getPracticeSession did not return the running session with current turn: %+v", detail)
	}
	if h.store.turnCount(started.Id) != 1 || h.store.sessionEventCount(started.Id) != 1 || h.store.outboxCount() != 1 {
		t.Fatalf("session start should persist one turn, one event, and one outbox row: turns=%d events=%d outbox=%d", h.store.turnCount(started.Id), h.store.sessionEventCount(started.Id), h.store.outboxCount())
	}
	if h.store.aiCalledInsideTransaction {
		t.Fatalf("AI call must happen outside the repository transaction window")
	}
	status := h.store.idempotencyStatus(practiceHTTPScenarioUserAID, "practice", "startPracticeSession", "e2e-p0-023-start-session")
	if status != idempotency.StatusSucceeded {
		t.Fatalf("startPracticeSession idempotency status = %q, want %q", status, idempotency.StatusSucceeded)
	}

	var startedPayload sharedevents.PracticeSessionStartedPayload
	decodeJSON(t, h.store.outboxPayloads()[0], &startedPayload)
	if startedPayload.SessionID != started.Id || startedPayload.PlanID != "practice-plan-p0-023" || startedPayload.TargetJobID != "target-job-p0-023-a" {
		t.Fatalf("practice.session.started payload mismatch: %+v", startedPayload)
	}
	assertNoEvidenceLeak(t, h.store.outboxPayloads(), started.CurrentTurn.QuestionText, "question_text", "answer_text", "hint_text", "prompt body", "response body", "provider secret")
}

type practiceHTTPScenarioHarness struct {
	handler http.Handler
	store   *scenarioPracticeStore
	cookies map[string]*http.Cookie
}

func newPracticeHTTPScenarioHarness(t *testing.T) *practiceHTTPScenarioHarness {
	t.Helper()
	dir := t.TempDir()
	writeAPIFile(t, filepath.Join(dir, "config.yaml"), `
runtime:
  appVersion: "1.2.3"
  defaultUiLanguage: zh-CN
auth:
  challengeTokenPepper: "scenario-challenge-pepper"
`)
	loader, err := config.Load(config.Options{AppEnv: "test", ConfigDir: dir})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	store := newScenarioPracticeStore()
	authStore := newPracticeScenarioAuthStore("scenario-session-secret")
	cookies := map[string]*http.Cookie{
		practiceHTTPScenarioUserAID: authStore.addSession(practiceHTTPScenarioUserAID, "candidate-a@example.com", "raw-session-token-a"),
		practiceHTTPScenarioUserBID: authStore.addSession(practiceHTTPScenarioUserBID, "candidate-b@example.com", "raw-session-token-b"),
	}
	authService := auth.NewPasswordlessService(auth.PasswordlessServiceOptions{
		Store:               authStore,
		SessionCookieSecret: "scenario-session-secret",
		Now:                 fixedScenarioNow,
	})
	service := domainpractice.NewService(domainpractice.ServiceOptions{
		Store:    store,
		Registry: &scenarioPracticeRegistry{},
		AI:       &scenarioPracticeAIClient{store: store},
		Now:      fixedScenarioNow,
		NewID:    store.nextID,
	})
	practiceHandler := apipractice.NewHandler(apipractice.HandlerOptions{
		Service: service,
		Session: currentUserFromContext,
	})
	routeIdempotency := idempotency.New(idempotency.MiddlewareOptions{
		Store:     store,
		Now:       fixedScenarioNow,
		NewID:     store.nextID,
		KeyPepper: "scenario-challenge-pepper",
	})

	return &practiceHTTPScenarioHarness{
		handler: buildAPIHandlerWithHandlers(loader, apiRuntimeFlags{}, authService, targetjob.NewHandler(targetjob.HandlerOptions{}), practiceRoutes{
			Handler:     practiceHandler,
			Idempotency: routeIdempotency,
		}),
		store:   store,
		cookies: cookies,
	}
}

func (h *practiceHTTPScenarioHarness) doJSON(t *testing.T, userID, method, path string, idempotencyKey string, body any, wantStatus int) []byte {
	t.Helper()
	var reader *bytes.Reader
	if body == nil {
		reader = bytes.NewReader(nil)
	} else {
		raw, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request body: %v", err)
		}
		reader = bytes.NewReader(raw)
	}
	req := httptest.NewRequest(method, path, reader)
	cookie, ok := h.cookies[userID]
	if !ok {
		t.Fatalf("missing scenario cookie for user %q", userID)
	}
	req.AddCookie(cookie)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if idempotencyKey != "" {
		req.Header.Set(idempotency.HeaderName, idempotencyKey)
	}
	rec := httptest.NewRecorder()
	h.handler.ServeHTTP(rec, req)
	if rec.Code != wantStatus {
		t.Fatalf("%s %s status=%d want=%d body=%s", method, path, rec.Code, wantStatus, rec.Body.String())
	}
	return rec.Body.Bytes()
}

type scenarioPracticeRegistry struct{}

func (r *scenarioPracticeRegistry) ResolveActive(ctx context.Context, featureKey, language string) (registry.PromptResolution, error) {
	return registry.PromptResolution{
		FeatureKey:          featureKey,
		PromptVersion:       "prompt.v1",
		RubricVersion:       "rubric.v1",
		ModelProfileName:    "practice.first_question.default",
		DataSourceVersion:   "registry.v1",
		FeatureFlag:         "none",
		UserMessageTemplate: "ask the first interview question",
	}, nil
}

type scenarioPracticeAIClient struct {
	store *scenarioPracticeStore
}

func (c *scenarioPracticeAIClient) Complete(ctx context.Context, profileName string, payload aiclient.CompletePayload) (aiclient.CompleteResponse, aiclient.AICallMeta, error) {
	if c.store != nil {
		c.store.recordAIObservation()
	}
	if profileName != "practice.first_question.default" {
		return aiclient.CompleteResponse{}, aiclient.AICallMeta{}, fmt.Errorf("unexpected profile %q", profileName)
	}
	if payload.Metadata.FeatureKey != "practice.session.first_question" ||
		payload.Metadata.PromptVersion == "" ||
		payload.Metadata.RubricVersion == "" ||
		payload.Metadata.Language == "" ||
		payload.Metadata.FeatureFlag == "" ||
		payload.Metadata.DataSourceVersion == "" {
		return aiclient.CompleteResponse{}, aiclient.AICallMeta{}, fmt.Errorf("incomplete AI metadata: %+v", payload.Metadata)
	}
	return aiclient.CompleteResponse{Content: `{"questionText":"请用 STAR 描述你主导设计系统迁移的项目，重点说明跨团队协调过程。","questionIntent":"behavioral.leadership.design_system"}`}, aiclient.AICallMeta{}, nil
}

func (c *scenarioPracticeAIClient) Transcribe(ctx context.Context, profileName string, input aiclient.TranscriptionInput) (aiclient.TranscriptionResponse, aiclient.AICallMeta, error) {
	return aiclient.TranscriptionResponse{}, aiclient.AICallMeta{}, nil
}

func (c *scenarioPracticeAIClient) Stream(ctx context.Context, profileName string, payload aiclient.CompletePayload) (<-chan aiclient.AIStreamEvent, error) {
	return nil, nil
}

func (c *scenarioPracticeAIClient) Synthesize(ctx context.Context, profileName string, input aiclient.SynthesisInput) (aiclient.SynthesisResponse, aiclient.AICallMeta, error) {
	return aiclient.SynthesisResponse{}, aiclient.AICallMeta{}, nil
}

type scenarioPracticeStore struct {
	seq                       int
	plans                     map[string]scenarioPracticePlan
	sessions                  map[string]scenarioPracticeSession
	turns                     map[string][]domainpractice.TurnRecord
	sessionEvents             map[string][]scenarioPracticeSessionEvent
	idempotencyRecords        map[string]scenarioPracticeIdempotencyRecord
	outbox                    []scenarioOutboxEvent
	audits                    [][]byte
	prerequisiteTargetOwner   map[string]string
	prerequisiteResumeOwner   map[string]string
	inTransaction             bool
	aiCalledInsideTransaction bool
}

type scenarioPracticePlan struct {
	domainpractice.PlanRecord
	UserID        string
	ResumeAssetID string
}

type scenarioPracticeSession struct {
	domainpractice.SessionRecord
	UserID string
}

type scenarioPracticeSessionEvent struct {
	ID        string
	SeqNo     int
	EventType string
	Payload   []byte
}

type scenarioOutboxEvent struct {
	EventName sharedevents.EventName
	Payload   []byte
}

type scenarioPracticeIdempotencyRecord struct {
	RecordID    string
	UserID      string
	Domain      string
	Operation   string
	KeyHash     string
	Fingerprint string
	Status      idempotency.Status
	ExpiresAt   time.Time
	Response    []byte
	HTTPStatus  int
}

func newScenarioPracticeStore() *scenarioPracticeStore {
	s := &scenarioPracticeStore{
		plans:                   map[string]scenarioPracticePlan{},
		sessions:                map[string]scenarioPracticeSession{},
		turns:                   map[string][]domainpractice.TurnRecord{},
		sessionEvents:           map[string][]scenarioPracticeSessionEvent{},
		idempotencyRecords:      map[string]scenarioPracticeIdempotencyRecord{},
		prerequisiteTargetOwner: map[string]string{},
		prerequisiteResumeOwner: map[string]string{},
	}
	s.prerequisiteTargetOwner["target-job-p0-022-a"] = practiceHTTPScenarioUserAID
	s.prerequisiteResumeOwner["resume-asset-p0-022-a"] = practiceHTTPScenarioUserAID
	s.prerequisiteTargetOwner["target-job-p0-023-a"] = practiceHTTPScenarioUserAID
	s.prerequisiteResumeOwner["resume-asset-p0-023-a"] = practiceHTTPScenarioUserAID
	return s
}

func (s *scenarioPracticeStore) nextID() string {
	s.seq++
	return "practice-scenario-id-" + strconv.Itoa(s.seq)
}

func (s *scenarioPracticeStore) CreatePlan(_ context.Context, in domainpractice.CreatePlanStoreInput) (domainpractice.PlanRecord, error) {
	if s.prerequisiteTargetOwner[in.TargetJobID] != in.UserID || s.prerequisiteResumeOwner[in.ResumeAssetID] != in.UserID {
		return domainpractice.PlanRecord{}, domainpractice.ErrPlanPrerequisiteNotFound
	}
	plan := domainpractice.PlanRecord{
		ID:                 in.PlanID,
		TargetJobID:        in.TargetJobID,
		Goal:               in.Goal,
		Mode:               in.Mode,
		InterviewerPersona: in.InterviewerPersona,
		Difficulty:         in.Difficulty,
		Language:           in.Language,
		TimeBudgetMinutes:  in.TimeBudgetMinutes,
		QuestionBudget:     in.QuestionBudget,
		Status:             "ready",
		CreatedAt:          in.Now,
	}
	s.plans[in.PlanID] = scenarioPracticePlan{PlanRecord: plan, UserID: in.UserID, ResumeAssetID: in.ResumeAssetID}
	audit, err := json.Marshal(map[string]any{
		"plan_id":       in.PlanID,
		"goal":          string(in.Goal),
		"mode":          string(in.Mode),
		"language":      in.Language,
		"target_job_id": in.TargetJobID,
	})
	if err != nil {
		return domainpractice.PlanRecord{}, err
	}
	s.audits = append(s.audits, audit)
	return plan, nil
}

func (s *scenarioPracticeStore) seedReadyPlan(in domainpractice.CreatePlanStoreInput) domainpractice.PlanRecord {
	plan, err := s.CreatePlan(context.Background(), in)
	if err != nil {
		panic(err)
	}
	return plan
}

func (s *scenarioPracticeStore) GetPlan(_ context.Context, userID, planID string) (domainpractice.PlanRecord, error) {
	plan, ok := s.plans[planID]
	if !ok || plan.UserID != userID {
		return domainpractice.PlanRecord{}, domainpractice.ErrPlanNotFound
	}
	return plan.PlanRecord, nil
}

func (s *scenarioPracticeStore) GetSession(_ context.Context, userID, sessionID string) (domainpractice.SessionRecord, error) {
	session, ok := s.sessions[sessionID]
	if !ok || session.UserID != userID {
		return domainpractice.SessionRecord{}, domainpractice.ErrSessionNotFound
	}
	return session.SessionRecord, nil
}

func (s *scenarioPracticeStore) ReserveSessionStart(_ context.Context, in domainpractice.StartSessionReservationInput) (domainpractice.SessionReservation, error) {
	plan, ok := s.plans[in.PlanID]
	if !ok || plan.UserID != in.UserID || plan.Status != "ready" {
		return domainpractice.SessionReservation{}, domainpractice.ErrPlanNotFound
	}
	s.inTransaction = true
	defer func() { s.inTransaction = false }()

	key := s.idempotencyRecordKey(in.UserID, "practice", "startPracticeSession", in.IdempotencyKeyHash)
	s.idempotencyRecords[key] = scenarioPracticeIdempotencyRecord{
		RecordID:    in.IdempotencyRecordID,
		UserID:      in.UserID,
		Domain:      "practice",
		Operation:   "startPracticeSession",
		KeyHash:     in.IdempotencyKeyHash,
		Fingerprint: in.RequestFingerprint,
		Status:      idempotency.StatusPending,
		ExpiresAt:   in.ExpiresAt,
	}
	session := domainpractice.SessionRecord{
		ID:           in.SessionID,
		PlanID:       in.PlanID,
		TargetJobID:  plan.TargetJobID,
		Status:       sharedtypes.SessionStatusQueued,
		Language:     plan.Language,
		HintsEnabled: in.HintsEnabled,
		TurnCount:    0,
		CreatedAt:    in.Now,
		UpdatedAt:    in.Now,
	}
	s.sessions[in.SessionID] = scenarioPracticeSession{SessionRecord: session, UserID: in.UserID}
	return domainpractice.SessionReservation{
		IdempotencyRecordID: in.IdempotencyRecordID,
		SessionID:           in.SessionID,
		PlanID:              in.PlanID,
		TargetJobID:         plan.TargetJobID,
		Goal:                plan.Goal,
		Mode:                plan.Mode,
		InterviewerPersona:  plan.InterviewerPersona,
		Language:            plan.Language,
		HintsEnabled:        in.HintsEnabled,
		CreatedAt:           in.Now,
		UpdatedAt:           in.Now,
	}, nil
}

func (s *scenarioPracticeStore) CommitSessionStart(_ context.Context, in domainpractice.CommitSessionStartInput) (domainpractice.SessionRecord, error) {
	session, ok := s.sessions[in.SessionID]
	if !ok {
		return domainpractice.SessionRecord{}, domainpractice.ErrSessionNotFound
	}
	s.inTransaction = true
	defer func() { s.inTransaction = false }()

	turn := domainpractice.TurnRecord{
		ID:             in.TurnID,
		TurnIndex:      1,
		QuestionText:   in.QuestionText,
		QuestionIntent: in.QuestionIntent,
		Status:         "asked",
		AskedAt:        in.StartedAt,
	}
	s.turns[in.SessionID] = append(s.turns[in.SessionID], turn)
	eventPayload, err := json.Marshal(map[string]any{"sessionId": in.SessionID, "turnId": in.TurnID, "turnIndex": 1})
	if err != nil {
		return domainpractice.SessionRecord{}, err
	}
	s.sessionEvents[in.SessionID] = append(s.sessionEvents[in.SessionID], scenarioPracticeSessionEvent{
		ID:        in.SessionEventID,
		SeqNo:     1,
		EventType: "session_started",
		Payload:   eventPayload,
	})
	startedPayload, err := storepractice.BuildPracticeSessionStartedPayload(storepractice.PracticeSessionStartedInput{
		Goal:        in.Goal,
		Language:    in.Language,
		Mode:        in.Mode,
		PlanID:      in.PlanID,
		SessionID:   in.SessionID,
		TargetJobID: in.TargetJobID,
	})
	if err != nil {
		return domainpractice.SessionRecord{}, err
	}
	outboxPayload, err := json.Marshal(startedPayload)
	if err != nil {
		return domainpractice.SessionRecord{}, err
	}
	s.outbox = append(s.outbox, scenarioOutboxEvent{EventName: sharedevents.EventNamePracticeSessionStarted, Payload: outboxPayload})

	session.Status = sharedtypes.SessionStatusRunning
	session.TurnCount = 1
	session.CurrentTurn = &turn
	session.UpdatedAt = in.StartedAt
	s.sessions[in.SessionID] = session

	responseBody, err := json.Marshal(api.PracticeSession{
		Id:           session.ID,
		PlanId:       session.PlanID,
		TargetJobId:  session.TargetJobID,
		Status:       session.Status,
		Language:     session.Language,
		HintsEnabled: session.HintsEnabled,
		TurnCount:    session.TurnCount,
		CurrentTurn: &api.PracticeTurn{
			Id:             turn.ID,
			TurnIndex:      turn.TurnIndex,
			QuestionText:   turn.QuestionText,
			QuestionIntent: &turn.QuestionIntent,
			Status:         turn.Status,
			AskedAt:        strPtr(turn.AskedAt.UTC().Format(time.RFC3339)),
		},
		CreatedAt: session.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt: session.UpdatedAt.UTC().Format(time.RFC3339),
	})
	if err != nil {
		return domainpractice.SessionRecord{}, err
	}
	for key, record := range s.idempotencyRecords {
		if record.RecordID == in.IdempotencyRecordID {
			record.Status = idempotency.StatusSucceeded
			record.Response = responseBody
			record.HTTPStatus = http.StatusCreated
			s.idempotencyRecords[key] = record
			break
		}
	}
	return session.SessionRecord, nil
}

func (s *scenarioPracticeStore) Reserve(_ context.Context, in idempotency.ReservationInput) (idempotency.Reservation, error) {
	key := s.idempotencyRecordKey(in.UserID, in.Domain, in.Operation, in.IdempotencyKeyHash)
	rec, ok := s.idempotencyRecords[key]
	if !ok || !in.Now.Before(rec.ExpiresAt) {
		s.idempotencyRecords[key] = scenarioPracticeIdempotencyRecord{
			RecordID:    in.RecordID,
			UserID:      in.UserID,
			Domain:      in.Domain,
			Operation:   in.Operation,
			KeyHash:     in.IdempotencyKeyHash,
			Fingerprint: in.RequestFingerprint,
			Status:      idempotency.StatusPending,
			ExpiresAt:   in.ExpiresAt,
		}
		return idempotency.Reservation{State: idempotency.StateExecute, RecordID: in.RecordID}, nil
	}
	if rec.Fingerprint != in.RequestFingerprint {
		return idempotency.Reservation{}, idempotency.ErrFingerprintMismatch
	}
	switch rec.Status {
	case idempotency.StatusPending:
		return idempotency.Reservation{}, idempotency.ErrPending
	case idempotency.StatusSucceeded:
		return idempotency.Reservation{
			State:          idempotency.StateReplay,
			RecordID:       rec.RecordID,
			ResponseStatus: rec.HTTPStatus,
			ResponseBody:   append([]byte{}, rec.Response...),
		}, nil
	case idempotency.StatusFailedRetry:
		rec.Status = idempotency.StatusPending
		rec.ExpiresAt = in.ExpiresAt
		rec.Fingerprint = in.RequestFingerprint
		s.idempotencyRecords[key] = rec
		return idempotency.Reservation{State: idempotency.StateExecute, RecordID: rec.RecordID}, nil
	default:
		return idempotency.Reservation{}, idempotency.ErrUnexpectedStatus
	}
}

func (s *scenarioPracticeStore) MarkSucceeded(_ context.Context, in idempotency.CompletionInput) error {
	for key, rec := range s.idempotencyRecords {
		if rec.RecordID == in.RecordID && rec.UserID == in.UserID && rec.Domain == in.Domain && rec.Operation == in.Operation {
			rec.Status = idempotency.StatusSucceeded
			rec.Response = append([]byte{}, in.ResponseBody...)
			rec.HTTPStatus = in.ResponseStatus
			s.idempotencyRecords[key] = rec
			return nil
		}
	}
	return idempotency.ErrReservationNotFound
}

func (s *scenarioPracticeStore) recordAIObservation() {
	if s.inTransaction {
		s.aiCalledInsideTransaction = true
	}
}

func (s *scenarioPracticeStore) planCount() int {
	return len(s.plans)
}

func (s *scenarioPracticeStore) auditCount() int {
	return len(s.audits)
}

func (s *scenarioPracticeStore) auditPayloads() [][]byte {
	out := make([][]byte, 0, len(s.audits))
	for _, payload := range s.audits {
		out = append(out, append([]byte{}, payload...))
	}
	return out
}

func (s *scenarioPracticeStore) turnCount(sessionID string) int {
	return len(s.turns[sessionID])
}

func (s *scenarioPracticeStore) sessionEventCount(sessionID string) int {
	return len(s.sessionEvents[sessionID])
}

func (s *scenarioPracticeStore) outboxCount() int {
	return len(s.outbox)
}

func (s *scenarioPracticeStore) outboxPayloads() [][]byte {
	out := make([][]byte, 0, len(s.outbox))
	for _, event := range s.outbox {
		out = append(out, append([]byte{}, event.Payload...))
	}
	return out
}

func (s *scenarioPracticeStore) idempotencyStatus(userID, domain, operation, rawKey string) idempotency.Status {
	rec, ok := s.idempotencyRecords[s.idempotencyRecordKey(userID, domain, operation, idempotency.HashKey(rawKey, ""))]
	if !ok {
		return ""
	}
	return rec.Status
}

func (s *scenarioPracticeStore) idempotencyRecordKey(userID, domain, operation, keyHash string) string {
	return strings.Join([]string{userID, domain, operation, keyHash}, "\x00")
}

var (
	_ domainpractice.Store = (*scenarioPracticeStore)(nil)
	_ idempotency.Store    = (*scenarioPracticeStore)(nil)
)

type practiceScenarioAuthStore struct {
	sessionSecret  string
	sessionsByHash map[string]auth.SessionRecord
	sessionsByID   map[string]string
	usersByID      map[string]auth.UserContext
}

func newPracticeScenarioAuthStore(sessionSecret string) *practiceScenarioAuthStore {
	return &practiceScenarioAuthStore{
		sessionSecret:  sessionSecret,
		sessionsByHash: map[string]auth.SessionRecord{},
		sessionsByID:   map[string]string{},
		usersByID:      map[string]auth.UserContext{},
	}
}

func (s *practiceScenarioAuthStore) addSession(userID, email, rawToken string) *http.Cookie {
	sessionHash := practiceScenarioSessionHash(s.sessionSecret, rawToken)
	user := auth.UserContext{ID: userID, Email: email, DisplayName: "Scenario User", UILanguage: "zh-CN", PreferredPracticeLanguage: "zh-CN", AnalyticsOptIn: true}
	s.usersByID[userID] = user
	session := auth.SessionRecord{
		ID:          "session-" + userID,
		UserID:      userID,
		SessionHash: sessionHash,
		Status:      auth.SessionStatusActive,
		ExpiresAt:   fixedScenarioNow().Add(auth.SessionTTL),
		CreatedAt:   fixedScenarioNow(),
		UpdatedAt:   fixedScenarioNow(),
	}
	s.sessionsByHash[sessionHash] = session
	s.sessionsByID[session.ID] = sessionHash
	return &http.Cookie{Name: auth.SessionCookieName, Value: rawToken}
}

func (s *practiceScenarioAuthStore) CountRecentChallenges(context.Context, string, string, time.Time) (int, error) {
	return 0, nil
}

func (s *practiceScenarioAuthStore) CreateChallenge(context.Context, auth.ChallengeRecord) error {
	return nil
}

func (s *practiceScenarioAuthStore) ConsumeChallenge(context.Context, string, time.Time) (auth.ChallengeRecord, error) {
	return auth.ChallengeRecord{}, auth.ErrChallengeInvalid
}

func (s *practiceScenarioAuthStore) FindOrCreateUserByEmail(context.Context, string, string, time.Time) (auth.UserContext, error) {
	return auth.UserContext{}, auth.ErrSessionInvalid
}

func (s *practiceScenarioAuthStore) CreateSession(_ context.Context, rec auth.SessionRecord) error {
	s.sessionsByHash[rec.SessionHash] = rec
	s.sessionsByID[rec.ID] = rec.SessionHash
	return nil
}

func (s *practiceScenarioAuthStore) GetSessionByHash(_ context.Context, sessionHash string, _ time.Time) (auth.SessionRecord, error) {
	session, ok := s.sessionsByHash[sessionHash]
	if !ok {
		return auth.SessionRecord{}, auth.ErrSessionInvalid
	}
	return session, nil
}

func (s *practiceScenarioAuthStore) GetUserContext(_ context.Context, userID string) (auth.UserContext, error) {
	user, ok := s.usersByID[userID]
	if !ok {
		return auth.UserContext{}, auth.ErrSessionInvalid
	}
	return user, nil
}

func (s *practiceScenarioAuthStore) TouchSession(_ context.Context, sessionID string, now time.Time, expiresAt time.Time) error {
	hash, ok := s.sessionsByID[sessionID]
	if !ok {
		return auth.ErrSessionInvalid
	}
	session := s.sessionsByHash[hash]
	session.UpdatedAt = now
	session.ExpiresAt = expiresAt
	s.sessionsByHash[hash] = session
	return nil
}

func (s *practiceScenarioAuthStore) RevokeSession(_ context.Context, sessionID string, now time.Time) error {
	hash, ok := s.sessionsByID[sessionID]
	if !ok {
		return auth.ErrSessionInvalid
	}
	session := s.sessionsByHash[hash]
	session.Status = auth.SessionStatusRevoked
	session.RevokedAt = now
	session.UpdatedAt = now
	s.sessionsByHash[hash] = session
	return nil
}

func (s *practiceScenarioAuthStore) CreatePrivacyDeleteHandoff(context.Context, string, string, string, string, time.Time) (auth.PrivacyDeleteHandoff, error) {
	return auth.PrivacyDeleteHandoff{}, fmt.Errorf("not used")
}

func practiceScenarioSessionHash(secret, rawToken string) string {
	sum := sha256.Sum256([]byte(secret + "\x00" + strings.TrimSpace(rawToken)))
	return hex.EncodeToString(sum[:])
}

func practiceBoolPtr(value bool) *bool {
	return &value
}
