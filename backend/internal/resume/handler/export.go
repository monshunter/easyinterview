package handler

import (
	"net/http"
	"strings"

	"github.com/monshunter/easyinterview/backend/internal/middleware/idempotency"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
)

// ExportResume binds POST /api/v1/resumes/{resumeId}/exports. P0 is
// intentionally unavailable and returns 501 + RESUME_EXPORT_NOT_AVAILABLE
// (spec D-6); no model quota is consumed.
func (h *Handler) ExportResume(w http.ResponseWriter, r *http.Request, _ string) {
	if h == nil {
		writeAPIError(w, http.StatusInternalServerError, sharederrors.CodeValidationFailed, "resume service is not configured", nil)
		return
	}
	if _, ok := h.resolveUser(r); !ok {
		writeAPIError(w, http.StatusUnauthorized, sharederrors.CodeAuthUnauthorized, "authentication required", nil)
		return
	}
	if strings.TrimSpace(r.Header.Get(idempotency.HeaderName)) == "" {
		writeAPIError(w, http.StatusUnprocessableEntity, sharederrors.CodeValidationFailed, "Idempotency-Key header is required", nil)
		return
	}
	writeAPIError(w, http.StatusNotImplemented, sharederrors.CodeResumeExportNotAvailable, "resume export is not available in P0", nil)
}
