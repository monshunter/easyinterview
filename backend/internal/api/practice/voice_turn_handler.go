package practice

import (
	"net/http"

	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
)

func (h *Handler) CreatePracticeVoiceTurn(w http.ResponseWriter, r *http.Request, _ string) {
	if _, ok := h.resolveUser(r); !ok {
		writeAPIError(w, http.StatusUnauthorized, sharederrors.CodeAuthUnauthorized, "authentication required", nil)
		return
	}
	writeAPIError(w, http.StatusUnprocessableEntity, sharederrors.CodeAiUnsupportedCapability, "phone mode is temporarily unavailable", nil)
}
