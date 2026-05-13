package practice

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	api "github.com/monshunter/easyinterview/backend/internal/api/generated"
	"github.com/monshunter/easyinterview/backend/internal/middleware/idempotency"
	domain "github.com/monshunter/easyinterview/backend/internal/practice"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
)

func (h *Handler) AppendSessionEvent(w http.ResponseWriter, r *http.Request, sessionID string) {
	if h == nil || h.service == nil {
		writeAPIError(w, http.StatusInternalServerError, sharederrors.CodeValidationFailed, "practice service is not configured", nil)
		return
	}
	userID, ok := h.resolveUser(r)
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, sharederrors.CodeAuthUnauthorized, "authentication required", nil)
		return
	}
	if strings.TrimSpace(r.Header.Get(idempotency.HeaderName)) != "" {
		writeAPIError(w, http.StatusBadRequest, sharederrors.CodeValidationFailed, "Idempotency-Key is not accepted for appendSessionEvent", map[string]any{
			"policy": "use_client_event_id",
		})
		return
	}
	var body api.PracticeSessionEventRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeAPIError(w, http.StatusBadRequest, sharederrors.CodeValidationFailed, "request body is malformed", nil)
		return
	}
	occurredAt, err := parseOptionalRFC3339(body.OccurredAt)
	if err != nil {
		writeAPIError(w, http.StatusUnprocessableEntity, sharederrors.CodeValidationFailed, "occurredAt must be RFC3339", map[string]any{"field": "occurredAt"})
		return
	}
	result, err := h.service.AppendSessionEvent(r.Context(), domain.AppendSessionEventRequest{
		UserID:        userID,
		SessionID:     sessionID,
		ClientEventID: body.ClientEventId,
		Kind:          body.Kind,
		OccurredAt:    occurredAt,
		Payload:       body.Payload,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toAPISessionEventResult(result))
}

func (h *Handler) CompletePracticeSession(w http.ResponseWriter, r *http.Request, sessionID string) {
	if h == nil || h.service == nil {
		writeAPIError(w, http.StatusInternalServerError, sharederrors.CodeValidationFailed, "practice service is not configured", nil)
		return
	}
	userID, ok := h.resolveUser(r)
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, sharederrors.CodeAuthUnauthorized, "authentication required", nil)
		return
	}
	var body api.CompletePracticeSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeAPIError(w, http.StatusBadRequest, sharederrors.CodeValidationFailed, "request body is malformed", nil)
		return
	}
	completedAt, err := parseOptionalRFC3339(body.ClientCompletedAt)
	if err != nil {
		writeAPIError(w, http.StatusUnprocessableEntity, sharederrors.CodeValidationFailed, "clientCompletedAt must be RFC3339", map[string]any{"field": "clientCompletedAt"})
		return
	}
	result, err := h.service.CompletePracticeSession(r.Context(), domain.CompletePracticeSessionRequest{
		UserID:            userID,
		SessionID:         sessionID,
		ClientCompletedAt: completedAt,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	idempotency.SetResponseResource(w, "feedback_report", result.ReportID)
	writeJSON(w, http.StatusAccepted, toAPIReportWithJob(result))
}

func toAPISessionEventResult(result domain.AppendSessionEventResult) api.SessionEventResult {
	return api.SessionEventResult{
		Acknowledged:    result.Acknowledged,
		Session:         toAPIPracticeSession(result.Session),
		AssistantAction: toAPIAssistantAction(result.AssistantAction),
	}
}

func toAPIAssistantAction(action domain.AssistantActionRecord) api.AssistantAction {
	var turnID *string
	if strings.TrimSpace(action.TurnID) != "" {
		value := strings.TrimSpace(action.TurnID)
		turnID = &value
	}
	var questionText *string
	if strings.TrimSpace(action.QuestionText) != "" {
		value := strings.TrimSpace(action.QuestionText)
		questionText = &value
	}
	var hint *string
	if strings.TrimSpace(action.Hint) != "" {
		value := strings.TrimSpace(action.Hint)
		hint = &value
	}
	return api.AssistantAction{
		Type:          action.Type,
		TurnId:        turnID,
		QuestionText:  questionText,
		Hint:          hint,
		SessionStatus: action.SessionStatus,
		Provenance: api.GenerationProvenance{
			PromptVersion:     action.Provenance.PromptVersion,
			RubricVersion:     action.Provenance.RubricVersion,
			ModelId:           action.Provenance.ModelID,
			Language:          action.Provenance.Language,
			FeatureFlag:       action.Provenance.FeatureFlag,
			DataSourceVersion: action.Provenance.DataSourceVersion,
		},
	}
}

func toAPIReportWithJob(result domain.CompleteSessionResult) api.ReportWithJob {
	return api.ReportWithJob{
		ReportId: result.ReportID,
		Job: api.Job{
			Id:           result.Job.ID,
			JobType:      result.Job.JobType,
			ResourceType: result.Job.ResourceType,
			ResourceId:   result.Job.ResourceID,
			Status:       result.Job.Status,
			ErrorCode:    nilAPIErrorCode(result.Job.ErrorCode),
			CreatedAt:    result.Job.CreatedAt.UTC().Format(timeFormatRFC3339),
			UpdatedAt:    result.Job.UpdatedAt.UTC().Format(timeFormatRFC3339),
		},
	}
}

func nilAPIErrorCode(code string) *api.ApiErrorCode {
	if strings.TrimSpace(code) == "" {
		return nil
	}
	value := api.ApiErrorCode(strings.TrimSpace(code))
	return &value
}

func parseOptionalRFC3339(raw string) (time.Time, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}, nil
	}
	parsed, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return time.Time{}, err
	}
	return parsed, nil
}
