package handler

import (
	"context"
	"errors"
	"net/http"
	"strings"

	api "github.com/monshunter/easyinterview/backend/internal/api/generated"
	"github.com/monshunter/easyinterview/backend/internal/middleware/idempotency"
	"github.com/monshunter/easyinterview/backend/internal/resume"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
)

type ArchiveResumeService interface {
	ArchiveResume(ctx context.Context, userID string, resumeID string) (api.Resume, error)
}

// ArchiveResume binds POST /api/v1/resumes/{resumeId}/archive.
func (h *Handler) ArchiveResume(w http.ResponseWriter, r *http.Request, resumeID string) {
	if h == nil {
		writeAPIError(w, http.StatusInternalServerError, sharederrors.CodeValidationFailed, "resume service is not configured", nil)
		return
	}
	service, ok := h.service.(ArchiveResumeService)
	if !ok {
		writeAPIError(w, http.StatusInternalServerError, sharederrors.CodeValidationFailed, "resume service is not configured", nil)
		return
	}
	userID, ok := h.resolveUser(r)
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, sharederrors.CodeAuthUnauthorized, "authentication required", nil)
		return
	}
	if strings.TrimSpace(r.Header.Get(idempotency.HeaderName)) == "" {
		writeAPIError(w, http.StatusUnprocessableEntity, sharederrors.CodeValidationFailed, "Idempotency-Key header is required", nil)
		return
	}
	out, err := service.ArchiveResume(r.Context(), userID, resumeID)
	if err != nil {
		if errors.Is(err, resume.ErrNotFound) {
			writeAPIError(w, http.StatusNotFound, sharederrors.CodeResourceNotFound, "resume not found", nil)
			return
		}
		writeAPIError(w, http.StatusInternalServerError, sharederrors.CodeValidationFailed, "resume archive failed", nil)
		return
	}
	idempotency.SetResponseResource(w, "resume", out.Id)
	writeJSON(w, http.StatusAccepted, out)
}
