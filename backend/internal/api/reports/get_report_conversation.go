package reports

import (
	"errors"
	"net/http"
	"strings"

	reviewdomain "github.com/monshunter/easyinterview/backend/internal/review"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
)

func (h *Handler) GetReportConversation(w http.ResponseWriter, r *http.Request, reportID string) {
	requestID := strings.TrimSpace(r.Header.Get("X-Request-ID"))
	if requestID != "" {
		w.Header().Set("X-Request-ID", requestID)
	}
	if h == nil || h.service == nil {
		writeAPIErrorWithRequestID(w, http.StatusInternalServerError, sharederrors.CodeValidationFailed, "report service is not configured", requestID, nil)
		return
	}
	userID, ok := h.resolveUser(r)
	if !ok {
		writeAPIErrorWithRequestID(w, http.StatusUnauthorized, sharederrors.CodeAuthUnauthorized, "authentication required", requestID, nil)
		return
	}
	conversation, err := h.service.GetReportConversation(r.Context(), userID, reportID)
	if err != nil {
		writeReportConversationError(w, requestID, err)
		return
	}
	writeJSON(w, http.StatusOK, toAPIReportConversation(conversation))
}

func writeReportConversationError(w http.ResponseWriter, requestID string, err error) {
	switch {
	case errors.Is(err, reviewdomain.ErrReportNotFound):
		writeAPIErrorWithRequestID(w, http.StatusNotFound, sharederrors.CodeReportNotFound, "feedback report not found or not accessible", requestID, nil)
	case errors.Is(err, reviewdomain.ErrReportConversationInvalid):
		writeAPIErrorWithRequestID(w, http.StatusInternalServerError, sharederrors.CodeAiOutputInvalid, "report conversation projection is invalid", requestID, nil)
	default:
		writeAPIErrorWithRequestID(w, http.StatusInternalServerError, sharederrors.CodeValidationFailed, "report request failed", requestID, nil)
	}
}
