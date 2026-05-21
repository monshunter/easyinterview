package handler

import (
	"errors"
	"net/http"

	"github.com/monshunter/easyinterview/backend/internal/profile"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
)

// GetMyProfile binds GET /api/v1/profiles/me. First-time callers trigger a
// seed write using user_settings defaults (spec D-1); subsequent calls return
// the same row without re-seeding.
func (h *Handler) GetMyProfile(w http.ResponseWriter, r *http.Request) {
	if h == nil || h.store == nil || h.settings == nil {
		writeAPIError(w, http.StatusInternalServerError, sharederrors.CodeValidationFailed, "profile service is not configured", nil)
		return
	}
	userID, ok := h.resolveUser(r)
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, sharederrors.CodeAuthUnauthorized, "authentication required", nil)
		return
	}
	ctx := r.Context()
	rec, err := h.store.GetCandidateProfileByUser(ctx, userID)
	if err != nil && !errors.Is(err, profile.ErrNotFound) {
		writeServiceError(w, err, "profile get failed")
		return
	}
	if rec == nil {
		defaults, derr := h.settings.GetUserSettings(ctx, userID)
		if derr != nil {
			writeServiceError(w, derr, "profile defaults lookup failed")
			return
		}
		rec, err = h.store.SeedCandidateProfile(ctx, userID, defaults)
		if err != nil {
			// Race: another concurrent SeedCandidateProfile may have inserted
			// the row first. Fall back to a fresh read so we still return the
			// canonical row instead of failing the request.
			if errors.Is(err, profile.ErrValidationFailed) {
				rec, err = h.store.GetCandidateProfileByUser(ctx, userID)
			}
			if err != nil {
				writeServiceError(w, err, "profile seed failed")
				return
			}
		}
	}
	writeJSON(w, http.StatusOK, mapCandidateProfile(rec))
}
