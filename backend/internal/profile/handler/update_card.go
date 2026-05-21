package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	api "github.com/monshunter/easyinterview/backend/internal/api/generated"
	"github.com/monshunter/easyinterview/backend/internal/middleware/idempotency"
	"github.com/monshunter/easyinterview/backend/internal/profile"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
)

// UpdateExperienceCard binds PATCH /api/v1/profiles/me/experience-cards/{cardId}.
// Patch semantics (spec D-2); cross-user access returns 404 + RESOURCE_NOT_FOUND
// (spec D-8) without exposing existence.
func (h *Handler) UpdateExperienceCard(w http.ResponseWriter, r *http.Request, cardID string) {
	if h == nil || h.store == nil {
		writeAPIError(w, http.StatusInternalServerError, sharederrors.CodeValidationFailed, "profile service is not configured", nil)
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
	cardID = strings.TrimSpace(cardID)
	if cardID == "" {
		writeAPIError(w, http.StatusUnprocessableEntity, sharederrors.CodeValidationFailed, "cardId is required", nil)
		return
	}
	var body api.UpdateExperienceCardRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeAPIError(w, http.StatusBadRequest, sharederrors.CodeValidationFailed, "request body is malformed", nil)
		return
	}
	patch := profile.ExperienceCardPatch{
		Title:       body.Title,
		CompanyName: body.CompanyName,
		Situation:   body.Situation,
		Task:        body.Task,
		Action:      body.Action,
		Result:      body.Result,
		Language:    body.Language,
	}
	if body.Skills != nil {
		skills := append([]string{}, body.Skills...)
		patch.Skills = &skills
	}
	rec, err := h.store.UpdateExperienceCard(r.Context(), cardID, userID, patch)
	if err != nil {
		writeServiceError(w, err, "update experience card failed")
		return
	}
	writeJSON(w, http.StatusOK, mapExperienceCard(rec))
}
