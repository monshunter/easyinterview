package reports

import (
	"errors"
	"net/http"
	"strings"

	api "github.com/monshunter/easyinterview/backend/internal/api/generated"
	"github.com/monshunter/easyinterview/backend/internal/middleware/idempotency"
	reviewdomain "github.com/monshunter/easyinterview/backend/internal/review"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
)

func (h *Handler) RegenerateFeedbackReport(w http.ResponseWriter, r *http.Request, reportID string) {
	requestID := strings.TrimSpace(r.Header.Get("X-Request-ID"))
	if requestID != "" {
		w.Header().Set("X-Request-ID", requestID)
	}
	if h == nil || h.service == nil {
		writeAPIErrorWithRequestID(w, http.StatusInternalServerError, sharederrors.CodeValidationFailed, "report service is not configured", requestID, nil)
		return
	}
	service, ok := h.service.(reportRegenerationService)
	if !ok {
		writeAPIErrorWithRequestID(w, http.StatusInternalServerError, sharederrors.CodeValidationFailed, "report regeneration service is not configured", requestID, nil)
		return
	}
	userID, ok := h.resolveUser(r)
	if !ok {
		writeAPIErrorWithRequestID(w, http.StatusUnauthorized, sharederrors.CodeAuthUnauthorized, "authentication required", requestID, nil)
		return
	}
	result, err := service.RegenerateReport(r.Context(), reviewdomain.RegenerateReportRequest{UserID: userID, ReportID: reportID})
	if err != nil {
		writeRegenerateReportError(w, requestID, err)
		return
	}
	idempotency.SetResponseResource(w, "feedback_report", result.ReportID)
	writeJSON(w, http.StatusAccepted, toAPIRegenerateReport(result))
}

func writeRegenerateReportError(w http.ResponseWriter, requestID string, err error) {
	switch {
	case errors.Is(err, reviewdomain.ErrReportNotFound):
		writeAPIErrorWithRequestID(w, http.StatusNotFound, sharederrors.CodeReportNotFound, "feedback report not found or not accessible", requestID, nil)
	case errors.Is(err, reviewdomain.ErrReportNotReady):
		writeAPIErrorWithRequestID(w, http.StatusConflict, sharederrors.CodeReportNotReady, "report is not ready yet", requestID, nil)
	case errors.Is(err, reviewdomain.ErrReportContextTooLarge):
		writeAPIErrorWithRequestID(w, http.StatusConflict, sharederrors.CodeReportContextTooLarge, "report context exceeds supported generation size", requestID, nil)
	case errors.Is(err, reviewdomain.ErrReportInvalidStateTransition):
		writeAPIErrorWithRequestID(w, http.StatusConflict, sharederrors.CodeReportInvalidStateTransition, "report state transition is not allowed", requestID, nil)
	default:
		writeAPIErrorWithRequestID(w, http.StatusInternalServerError, sharederrors.CodeValidationFailed, "report regeneration failed", requestID, nil)
	}
}

func toAPIRegenerateReport(result reviewdomain.RegenerateReportResult) api.ReportWithJob {
	job := result.Job
	return api.ReportWithJob{
		ReportId: result.ReportID,
		Job: api.Job{
			Id: job.ID, JobType: api.JobType(job.JobType), ResourceType: api.ResourceType(job.ResourceType),
			ResourceId: job.ResourceID, Status: job.Status,
			CreatedAt: job.CreatedAt.UTC().Format(timeFormatRFC3339), UpdatedAt: job.UpdatedAt.UTC().Format(timeFormatRFC3339),
		},
	}
}
