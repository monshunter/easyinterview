package practice

import (
	"context"
	"encoding/json"
	stderrs "errors"
	"io"
	"net/http"
	"strings"

	api "github.com/monshunter/easyinterview/backend/internal/api/generated"
	"github.com/monshunter/easyinterview/backend/internal/middleware/idempotency"
	domain "github.com/monshunter/easyinterview/backend/internal/practice"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
)

type SessionResolver func(ctx context.Context) (userID string, ok bool)

type planService interface {
	CreatePracticePlan(ctx context.Context, in domain.CreatePlanRequest) (domain.PlanRecord, error)
	GetPracticePlan(ctx context.Context, userID, planID string) (domain.PlanRecord, error)
	GetPracticeSession(ctx context.Context, userID, sessionID string) (domain.SessionRecord, error)
	StartPracticeSession(ctx context.Context, in domain.StartSessionRequest) (domain.SessionRecord, error)
	AppendSessionEvent(ctx context.Context, in domain.AppendSessionEventRequest) (domain.AppendSessionEventResult, error)
	CompletePracticeSession(ctx context.Context, in domain.CompletePracticeSessionRequest) (domain.CompleteSessionResult, error)
	CreatePracticeVoiceTurn(ctx context.Context, in domain.CreatePracticeVoiceTurnRequest) (domain.PracticeVoiceTurnResult, error)
}

type HandlerOptions struct {
	Service              planService
	Session              SessionResolver
	IdempotencyKeyPepper string
}

type Handler struct {
	service              planService
	session              SessionResolver
	idempotencyKeyPepper string
}

func NewHandler(opts HandlerOptions) *Handler {
	return &Handler{service: opts.Service, session: opts.Session, idempotencyKeyPepper: opts.IdempotencyKeyPepper}
}

func (h *Handler) CreatePracticePlan(w http.ResponseWriter, r *http.Request) {
	if h == nil || h.service == nil {
		writeAPIError(w, http.StatusInternalServerError, sharederrors.CodeValidationFailed, "practice service is not configured", nil)
		return
	}
	userID, ok := h.resolveUser(r)
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, sharederrors.CodeAuthUnauthorized, "authentication required", nil)
		return
	}
	var body api.CreatePracticePlanRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeAPIError(w, http.StatusBadRequest, sharederrors.CodeValidationFailed, "request body is malformed", nil)
		return
	}

	res, err := h.service.CreatePracticePlan(r.Context(), domain.CreatePlanRequest{
		UserID:               userID,
		TargetJobID:          body.TargetJobId,
		ResumeAssetID:        body.ResumeAssetId,
		SourceReportID:       stringValue(body.SourceReportId),
		SourceDebriefID:      stringValue(body.SourceDebriefId),
		Goal:                 body.Goal,
		Mode:                 body.Mode,
		InterviewerPersona:   body.InterviewerPersona,
		Difficulty:           body.Difficulty,
		Language:             body.Language,
		TimeBudgetMinutes:    body.TimeBudgetMinutes,
		QuestionBudget:       body.QuestionBudget,
		FocusCompetencyCodes: body.FocusCompetencyCodes,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, toAPIPracticePlan(res))
}

func (h *Handler) GetPracticePlan(w http.ResponseWriter, r *http.Request, planID string) {
	if h == nil || h.service == nil {
		writeAPIError(w, http.StatusInternalServerError, sharederrors.CodeValidationFailed, "practice service is not configured", nil)
		return
	}
	userID, ok := h.resolveUser(r)
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, sharederrors.CodeAuthUnauthorized, "authentication required", nil)
		return
	}
	res, err := h.service.GetPracticePlan(r.Context(), userID, planID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toAPIPracticePlan(res))
}

func (h *Handler) StartPracticeSession(w http.ResponseWriter, r *http.Request) {
	if h == nil || h.service == nil {
		writeAPIError(w, http.StatusInternalServerError, sharederrors.CodeValidationFailed, "practice service is not configured", nil)
		return
	}
	userID, ok := h.resolveUser(r)
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, sharederrors.CodeAuthUnauthorized, "authentication required", nil)
		return
	}
	idempotencyKey := strings.TrimSpace(r.Header.Get(idempotency.HeaderName))
	if idempotencyKey == "" {
		writeAPIError(w, http.StatusBadRequest, sharederrors.CodeValidationFailed, "Idempotency-Key header is required", nil)
		return
	}
	rawBody, err := io.ReadAll(r.Body)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, sharederrors.CodeValidationFailed, "request body is malformed", nil)
		return
	}
	var body api.StartPracticeSessionRequest
	if err := json.Unmarshal(rawBody, &body); err != nil {
		writeAPIError(w, http.StatusBadRequest, sharederrors.CodeValidationFailed, "request body is malformed", nil)
		return
	}
	hintsEnabled := false
	if body.HintsEnabled != nil {
		hintsEnabled = *body.HintsEnabled
	}
	res, err := h.service.StartPracticeSession(r.Context(), domain.StartSessionRequest{
		UserID:             userID,
		PlanID:             body.PlanId,
		HintsEnabled:       hintsEnabled,
		IdempotencyKeyHash: idempotency.HashKey(idempotencyKey, h.idempotencyKeyPepper),
		RequestFingerprint: idempotency.Fingerprint(r.Method, r.URL.EscapedPath(), r.URL.RawQuery, rawBody),
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, toAPIPracticeSession(res))
}

func (h *Handler) GetPracticeSession(w http.ResponseWriter, r *http.Request, sessionID string) {
	if h == nil || h.service == nil {
		writeAPIError(w, http.StatusInternalServerError, sharederrors.CodeValidationFailed, "practice service is not configured", nil)
		return
	}
	userID, ok := h.resolveUser(r)
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, sharederrors.CodeAuthUnauthorized, "authentication required", nil)
		return
	}
	res, err := h.service.GetPracticeSession(r.Context(), userID, sessionID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toAPIPracticeSession(res))
}

func (h *Handler) resolveUser(r *http.Request) (string, bool) {
	if h.session == nil {
		return "", false
	}
	userID, ok := h.session(r.Context())
	return strings.TrimSpace(userID), ok && strings.TrimSpace(userID) != ""
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}

func toAPIPracticePlan(plan domain.PlanRecord) api.PracticePlan {
	out := api.PracticePlan{
		Id:                 plan.ID,
		TargetJobId:        plan.TargetJobID,
		Goal:               plan.Goal,
		Mode:               plan.Mode,
		InterviewerPersona: plan.InterviewerPersona,
		Difficulty:         plan.Difficulty,
		Language:           plan.Language,
		TimeBudgetMinutes:  plan.TimeBudgetMinutes,
		QuestionBudget:     plan.QuestionBudget,
		Status:             plan.Status,
		CreatedAt:          plan.CreatedAt.UTC().Format(timeFormatRFC3339),
	}
	if strings.TrimSpace(plan.SourceReportID) != "" {
		out.SourceReportId = &plan.SourceReportID
	}
	if strings.TrimSpace(plan.SourceDebriefID) != "" {
		out.SourceDebriefId = &plan.SourceDebriefID
	}
	return out
}

func toAPIPracticeSession(session domain.SessionRecord) api.PracticeSession {
	var currentTurn *api.PracticeTurn
	if session.CurrentTurn != nil {
		askedAt := session.CurrentTurn.AskedAt.UTC().Format(timeFormatRFC3339)
		intent := session.CurrentTurn.QuestionIntent
		currentTurn = &api.PracticeTurn{
			Id:             session.CurrentTurn.ID,
			TurnIndex:      session.CurrentTurn.TurnIndex,
			QuestionText:   session.CurrentTurn.QuestionText,
			QuestionIntent: &intent,
			Status:         session.CurrentTurn.Status,
			AskedAt:        &askedAt,
		}
	}
	return api.PracticeSession{
		Id:           session.ID,
		PlanId:       session.PlanID,
		TargetJobId:  session.TargetJobID,
		Status:       session.Status,
		Language:     session.Language,
		HintsEnabled: session.HintsEnabled,
		TurnCount:    session.TurnCount,
		CurrentTurn:  currentTurn,
		CreatedAt:    session.CreatedAt.UTC().Format(timeFormatRFC3339),
		UpdatedAt:    session.UpdatedAt.UTC().Format(timeFormatRFC3339),
	}
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	raw, err := json.Marshal(body)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, sharederrors.CodeValidationFailed, "response encoding failed", nil)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(raw)
}

func writeServiceError(w http.ResponseWriter, err error) {
	var svcErr *domain.ServiceError
	if stderrs.As(err, &svcErr) {
		status := http.StatusInternalServerError
		switch svcErr.Code {
		case sharederrors.CodeValidationFailed:
			status = http.StatusUnprocessableEntity
		case sharederrors.CodePracticePlanNotFound:
			status = http.StatusNotFound
		case sharederrors.CodePracticeSessionNotFound:
			status = http.StatusNotFound
		case sharederrors.CodePracticeSessionConflict:
			status = http.StatusConflict
		case sharederrors.CodeAiProviderTimeout,
			sharederrors.CodeAiOutputInvalid,
			sharederrors.CodeAiProviderSecretMissing,
			sharederrors.CodeAiProviderConfigInvalid,
			sharederrors.CodeAiUnsupportedCapability:
			status = http.StatusBadGateway
		case sharederrors.CodeAiFallbackExhausted:
			status = http.StatusServiceUnavailable
		}
		writeAPIError(w, status, svcErr.Code, svcErr.Message, svcErr.Details)
		return
	}
	writeAPIError(w, http.StatusInternalServerError, sharederrors.CodeValidationFailed, "practice request failed", nil)
}

func writeAPIError(w http.ResponseWriter, status int, code string, message string, details map[string]any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	meta := sharederrors.CodeRegistry[code]
	raw, _ := json.Marshal(api.ApiErrorResponse{
		Error: api.ApiError{
			Code:      code,
			Message:   message,
			RequestID: "",
			Retryable: meta.Retryable,
			Details:   details,
		},
	})
	_, _ = w.Write(raw)
}

func derefString(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}

const timeFormatRFC3339 = "2006-01-02T15:04:05Z07:00"

type createPracticePlanSurface interface {
	CreatePracticePlan(w http.ResponseWriter, r *http.Request)
	GetPracticePlan(w http.ResponseWriter, r *http.Request, planID string)
	StartPracticeSession(w http.ResponseWriter, r *http.Request)
	GetPracticeSession(w http.ResponseWriter, r *http.Request, sessionID string)
	AppendSessionEvent(w http.ResponseWriter, r *http.Request, sessionID string)
	CompletePracticeSession(w http.ResponseWriter, r *http.Request, sessionID string)
}

var _ createPracticePlanSurface = (*Handler)(nil)
