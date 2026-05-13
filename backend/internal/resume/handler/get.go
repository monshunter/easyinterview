package handler

import (
	"context"
	"errors"
	"net/http"

	api "github.com/monshunter/easyinterview/backend/internal/api/generated"
	"github.com/monshunter/easyinterview/backend/internal/resume"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
)

type GetService interface {
	GetResume(ctx context.Context, userID string, resumeAssetID string) (api.ResumeAsset, error)
}

// GetResume binds GET /api/v1/resumes/{resumeAssetId}.
func (h *Handler) GetResume(w http.ResponseWriter, r *http.Request, resumeAssetID string) {
	service, ok := h.service.(GetService)
	if h == nil || !ok {
		writeAPIError(w, http.StatusInternalServerError, sharederrors.CodeValidationFailed, "resume service is not configured", nil)
		return
	}
	userID, ok := h.resolveUser(r)
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, sharederrors.CodeAuthUnauthorized, "authentication required", nil)
		return
	}
	out, err := service.GetResume(r.Context(), userID, resumeAssetID)
	if err != nil {
		if errors.Is(err, resume.ErrNotFound) {
			writeAPIError(w, http.StatusNotFound, sharederrors.CodeTargetJobNotFound, "resume not found", nil)
			return
		}
		writeAPIError(w, http.StatusInternalServerError, sharederrors.CodeValidationFailed, "resume get failed", nil)
		return
	}
	writeJSON(w, http.StatusOK, out)
}
