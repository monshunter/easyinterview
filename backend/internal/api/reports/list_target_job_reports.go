package reports

import (
	"errors"
	"net/http"
	"strings"

	reviewdomain "github.com/monshunter/easyinterview/backend/internal/review"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
)

func (h *Handler) ListTargetJobReports(w http.ResponseWriter, r *http.Request, targetJobID string) {
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
	req := reviewdomain.ListTargetJobReportsRequest{
		UserID:      userID,
		TargetJobID: targetJobID,
	}
	res, err := h.service.ListTargetJobReports(r.Context(), req)
	if err != nil {
		writeTargetJobReportsError(w, requestID, err)
		return
	}
	writeJSON(w, http.StatusOK, toAPITargetJobReportsOverview(res))
}

func writeTargetJobReportsError(w http.ResponseWriter, requestID string, err error) {
	switch {
	case errors.Is(err, reviewdomain.ErrReportNotFound):
		writeAPIErrorWithRequestID(w, http.StatusNotFound, sharederrors.CodeTargetJobNotFound, "Target job not found.", requestID, map[string]any{"resource": "target_job"})
	case errors.Is(err, reviewdomain.ErrReportContextMissing):
		writeAPIErrorWithRequestID(w, http.StatusInternalServerError, sharederrors.CodeAiOutputInvalid, "Report context could not be validated.", requestID, map[string]any{"reason": "missing_generation_context"})
	case errors.Is(err, reviewdomain.ErrReportContextInvalid):
		writeAPIErrorWithRequestID(w, http.StatusInternalServerError, sharederrors.CodeAiOutputInvalid, "Report context could not be validated.", requestID, map[string]any{"reason": "invalid_generation_context"})
	default:
		writeAPIErrorWithRequestID(w, http.StatusInternalServerError, sharederrors.CodeValidationFailed, "report request failed", requestID, nil)
	}
}
