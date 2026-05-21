package handler

import (
	"encoding/json"
	"net/http"

	api "github.com/monshunter/easyinterview/backend/internal/api/generated"
	"github.com/monshunter/easyinterview/backend/internal/profile"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
)

// UpdateMyProfile binds PATCH /api/v1/profiles/me. Patch semantics (spec D-2):
// only supplied fields are written; empty string is a legal value (clears the
// column); yearsOfExperience must be >= 0; profile_version monotonically
// increments on every successful write.
func (h *Handler) UpdateMyProfile(w http.ResponseWriter, r *http.Request) {
	if h == nil || h.store == nil || h.settings == nil {
		writeAPIError(w, http.StatusInternalServerError, sharederrors.CodeValidationFailed, "profile service is not configured", nil)
		return
	}
	userID, ok := h.resolveUser(r)
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, sharederrors.CodeAuthUnauthorized, "authentication required", nil)
		return
	}
	var body api.UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeAPIError(w, http.StatusBadRequest, sharederrors.CodeValidationFailed, "request body is malformed", nil)
		return
	}
	patch, err := validateProfilePatch(body)
	if err != nil {
		writeAPIError(w, http.StatusUnprocessableEntity, sharederrors.CodeValidationFailed, err.Error(), nil)
		return
	}
	ctx := r.Context()
	defaults, derr := h.settings.GetUserSettings(ctx, userID)
	if derr != nil {
		writeServiceError(w, derr, "profile defaults lookup failed")
		return
	}
	rec, err := h.store.UpsertLite(ctx, userID, patch, defaults)
	if err != nil {
		writeServiceError(w, err, "profile update failed")
		return
	}
	writeJSON(w, http.StatusOK, mapCandidateProfile(rec))
}

func validateProfilePatch(body api.UpdateProfileRequest) (profile.ProfilePatch, error) {
	patch := profile.ProfilePatch{
		Headline:                  body.Headline,
		YearsOfExperience:         body.YearsOfExperience,
		CurrentRole:               body.CurrentRole,
		PreferredPracticeLanguage: body.PreferredPracticeLanguage,
		UiLanguage:                body.UiLanguage,
		Region:                    body.Region,
	}
	if patch.YearsOfExperience != nil && *patch.YearsOfExperience < 0 {
		return profile.ProfilePatch{}, validationErr("yearsOfExperience must be >= 0")
	}
	return patch, nil
}

type validationErr string

func (e validationErr) Error() string { return string(e) }
