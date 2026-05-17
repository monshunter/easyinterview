package handler

import (
	"context"
	"net/http"

	api "github.com/monshunter/easyinterview/backend/internal/api/generated"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
)

type GetTailorRunService interface {
	GetResumeTailorRun(ctx context.Context, userID string, tailorRunID string) (api.ResumeTailorRun, error)
}

// GetResumeTailorRun binds GET /api/v1/resume/tailor-runs/{tailorRunId}.
func (h *Handler) GetResumeTailorRun(w http.ResponseWriter, r *http.Request, tailorRunID string) {
	if h == nil {
		writeAPIError(w, http.StatusInternalServerError, sharederrors.CodeValidationFailed, "resume service is not configured", nil)
		return
	}
	service, ok := h.service.(GetTailorRunService)
	if !ok {
		writeAPIError(w, http.StatusInternalServerError, sharederrors.CodeValidationFailed, "resume service is not configured", nil)
		return
	}
	userID, ok := h.resolveUser(r)
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, sharederrors.CodeAuthUnauthorized, "authentication required", nil)
		return
	}
	out, err := service.GetResumeTailorRun(r.Context(), userID, tailorRunID)
	if err != nil {
		writeTailorRunError(w, err, "resume tailor run get failed")
		return
	}
	writeJSON(w, http.StatusOK, out)
}
