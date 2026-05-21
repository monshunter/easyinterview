package handler

import (
	"net/http"

	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
)

// GetJobMatchProfile is the JD-Match profile endpoint (D-18 sparse
// baseline + D-19 structural parity). The handler resolves the
// authenticated user, hands off to the orchestrator, and writes the
// generated JobMatchProfile DTO.
func (h *Handler) GetJobMatchProfile(w http.ResponseWriter, r *http.Request) {
	if h == nil || h.profileBuilder == nil {
		writeAPIError(w, http.StatusInternalServerError, sharederrors.CodeValidationFailed, "jdmatch profile service is not configured", nil)
		return
	}
	userID, ok := h.resolveUser(r)
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, sharederrors.CodeAuthUnauthorized, "authentication required", nil)
		return
	}
	res, err := h.profileBuilder(r.Context(), userID)
	if err != nil {
		writeServiceError(w, err, "jdmatch profile aggregation failed")
		return
	}
	writeJSON(w, http.StatusOK, res.Profile)
}
