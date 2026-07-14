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

func (h *Handler) SendPracticeMessage(w http.ResponseWriter, r *http.Request, sessionID string) {
	if h == nil || h.service == nil {
		writeAPIError(w, http.StatusInternalServerError, sharederrors.CodeValidationFailed, "practice service is not configured", nil)
		return
	}
	userID, ok := h.resolveUser(r)
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, sharederrors.CodeAuthUnauthorized, "authentication required", nil)
		return
	}
	var body api.SendPracticeMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeAPIError(w, http.StatusBadRequest, sharederrors.CodeValidationFailed, "request body is malformed", nil)
		return
	}
	result, err := h.service.SendPracticeMessage(r.Context(), domain.SendPracticeMessageRequest{
		UserID: userID, SessionID: sessionID, ClientMessageID: body.ClientMessageId, Text: body.Text,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, api.SendPracticeMessageResponse{
		Acknowledged:     result.Acknowledged,
		UserMessage:      toAPIPracticeUserMessage(result.UserMessage),
		AssistantMessage: toAPIPracticeAssistantMessage(result.AssistantMessage),
		Session:          toAPIPracticeSession(result.Session),
	})
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
	completedAt, missing, err := parseRequiredRFC3339(body.ClientCompletedAt)
	if missing {
		writeAPIError(w, http.StatusUnprocessableEntity, sharederrors.CodeValidationFailed, "clientCompletedAt is required", map[string]any{"field": "clientCompletedAt"})
		return
	}
	if err != nil {
		writeAPIError(w, http.StatusUnprocessableEntity, sharederrors.CodeValidationFailed, "clientCompletedAt must be RFC3339", map[string]any{"field": "clientCompletedAt"})
		return
	}
	result, err := h.service.CompletePracticeSession(r.Context(), domain.CompletePracticeSessionRequest{
		UserID: userID, SessionID: sessionID, ClientCompletedAt: completedAt,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	idempotency.SetResponseResource(w, "feedback_report", result.ReportID)
	writeJSON(w, http.StatusAccepted, toAPIReportWithJob(result))
}

func toAPIReportWithJob(result domain.CompleteSessionResult) api.ReportWithJob {
	return api.ReportWithJob{ReportId: result.ReportID, Job: api.Job{
		Id: result.Job.ID, JobType: result.Job.JobType, ResourceType: result.Job.ResourceType,
		ResourceId: result.Job.ResourceID, Status: result.Job.Status,
		ErrorCode: nilAPIErrorCode(result.Job.ErrorCode),
		CreatedAt: result.Job.CreatedAt.UTC().Format(timeFormatRFC3339),
		UpdatedAt: result.Job.UpdatedAt.UTC().Format(timeFormatRFC3339),
	}}
}

func nilAPIErrorCode(code string) *api.ApiErrorCode {
	if strings.TrimSpace(code) == "" {
		return nil
	}
	value := api.ApiErrorCode(strings.TrimSpace(code))
	return &value
}

func parseRequiredRFC3339(raw string) (time.Time, bool, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}, true, nil
	}
	parsed, err := time.Parse(time.RFC3339, raw)
	return parsed, false, err
}
