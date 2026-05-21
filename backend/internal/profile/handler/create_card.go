package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	api "github.com/monshunter/easyinterview/backend/internal/api/generated"
	"github.com/monshunter/easyinterview/backend/internal/middleware/idempotency"
	"github.com/monshunter/easyinterview/backend/internal/profile"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
)

// CreateExperienceCard binds POST /api/v1/profiles/me/experience-cards.
// IK is enforced upstream by the idempotency middleware; this handler also
// guards against missing keys for the unit-test path. source_type is forced
// to "manual" regardless of body (spec D-6); confidence defaults to "medium"
// (spec D-7).
func (h *Handler) CreateExperienceCard(w http.ResponseWriter, r *http.Request) {
	if h == nil || h.store == nil || h.settings == nil || h.newID == nil {
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
	var body api.CreateExperienceCardRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeAPIError(w, http.StatusBadRequest, sharederrors.CodeValidationFailed, "request body is malformed", nil)
		return
	}
	attrs, err := validateCreateCard(body)
	if err != nil {
		writeAPIError(w, http.StatusUnprocessableEntity, sharederrors.CodeValidationFailed, err.Error(), nil)
		return
	}
	ctx := r.Context()
	if err := h.ensureProfileExists(ctx, userID); err != nil {
		writeServiceError(w, err, "profile seed failed")
		return
	}
	source := profile.ExperienceCardSource{
		SourceType: profile.SourceTypeManual,
		Confidence: profile.ConfidenceDefaultMedium,
	}
	rec, err := h.store.CreateExperienceCard(ctx, h.newID(), userID, attrs, source)
	if err != nil {
		writeServiceError(w, err, "create experience card failed")
		return
	}
	writeJSON(w, http.StatusCreated, mapExperienceCard(rec))
}

// ensureProfileExists makes sure a candidate_profile row exists for userID so
// experience_cards.profile_id FK can resolve. Mirrors the spec §2.1 contract
// that experience cards always attach to the user's candidate profile.
func (h *Handler) ensureProfileExists(ctx context.Context, userID string) error {
	rec, err := h.store.GetCandidateProfileByUser(ctx, userID)
	if err != nil && !errors.Is(err, profile.ErrNotFound) {
		return err
	}
	if rec != nil {
		return nil
	}
	defaults, err := h.settings.GetUserSettings(ctx, userID)
	if err != nil {
		return err
	}
	if _, err := h.store.SeedCandidateProfile(ctx, userID, defaults); err != nil {
		// Concurrent seed race: another caller inserted first. The follow-up
		// CreateExperienceCard read will succeed against the row already in
		// place.
		if errors.Is(err, profile.ErrValidationFailed) {
			return nil
		}
		return err
	}
	return nil
}

func validateCreateCard(body api.CreateExperienceCardRequest) (profile.ExperienceCardAttrs, error) {
	title := strings.TrimSpace(body.Title)
	companyName := strings.TrimSpace(body.CompanyName)
	situation := strings.TrimSpace(body.Situation)
	task := strings.TrimSpace(body.Task)
	action := strings.TrimSpace(body.Action)
	result := strings.TrimSpace(body.Result)
	language := strings.TrimSpace(body.Language)
	if title == "" {
		return profile.ExperienceCardAttrs{}, validationErr("title is required")
	}
	if companyName == "" {
		return profile.ExperienceCardAttrs{}, validationErr("companyName is required")
	}
	if situation == "" {
		return profile.ExperienceCardAttrs{}, validationErr("situation is required")
	}
	if task == "" {
		return profile.ExperienceCardAttrs{}, validationErr("task is required")
	}
	if action == "" {
		return profile.ExperienceCardAttrs{}, validationErr("action is required")
	}
	if result == "" {
		return profile.ExperienceCardAttrs{}, validationErr("result is required")
	}
	if language == "" {
		return profile.ExperienceCardAttrs{}, validationErr("language is required")
	}
	skills := append([]string{}, body.Skills...)
	return profile.ExperienceCardAttrs{
		Title:       title,
		CompanyName: companyName,
		Situation:   situation,
		Task:        task,
		Action:      action,
		Result:      result,
		Skills:      skills,
		Language:    language,
	}, nil
}
