// Package handler implements the 5 Profile HTTP handlers defined in
// backend-profile spec §2.1. Handlers depend only on the domain Store +
// SettingsReader contracts so cmd/api wiring (auth session resolver +
// idempotency middleware) stays orthogonal.
package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	api "github.com/monshunter/easyinterview/backend/internal/api/generated"
	"github.com/monshunter/easyinterview/backend/internal/profile"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
)

// SessionResolver returns the authenticated user id for the request context.
type SessionResolver func(ctx context.Context) (userID string, ok bool)

// NewID mirrors idx.NewID so tests can inject deterministic ULIDs.
type NewID func() string

// Options bundles handler dependencies.
type Options struct {
	Store        profile.Store
	Settings     profile.SettingsReader
	Session      SessionResolver
	NewID        NewID
	MaxPageSize  int32
	DefaultPage  int32
	CursorPepper string // currently unused; reserved for future opaque cursor signing
}

// Handler hosts all 5 Profile endpoints. The struct stays small — each
// endpoint lives in its own file in this package.
type Handler struct {
	store       profile.Store
	settings    profile.SettingsReader
	session     SessionResolver
	newID       NewID
	maxPageSize int32
	defaultPage int32
}

// New constructs a Handler with the supplied options. Required fields:
// Store, Settings, Session, NewID.
func New(opts Options) *Handler {
	maxPage := opts.MaxPageSize
	if maxPage <= 0 {
		maxPage = profile.MaxExperienceCardSize
	}
	defPage := opts.DefaultPage
	if defPage <= 0 {
		defPage = profile.DefaultExperienceCardSize
	}
	return &Handler{
		store:       opts.Store,
		settings:    opts.Settings,
		session:     opts.Session,
		newID:       opts.NewID,
		maxPageSize: maxPage,
		defaultPage: defPage,
	}
}

func (h *Handler) resolveUser(r *http.Request) (string, bool) {
	if h == nil || h.session == nil {
		return "", false
	}
	userID, ok := h.session(r.Context())
	userID = strings.TrimSpace(userID)
	return userID, ok && userID != ""
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	raw, err := json.Marshal(body)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, sharederrors.CodeValidationFailed, "profile response encoding failed", nil)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(raw)
}

func writeAPIError(w http.ResponseWriter, status int, code string, message string, details map[string]any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	meta := sharederrors.CodeRegistry[code]
	raw, _ := json.Marshal(api.ApiErrorResponse{Error: api.ApiError{
		Code:      code,
		Message:   message,
		RequestID: "",
		Retryable: meta.Retryable,
		Details:   details,
	}})
	_, _ = w.Write(raw)
}

func writeServiceError(w http.ResponseWriter, err error, fallbackMessage string) {
	switch {
	case errors.Is(err, profile.ErrNotFound), errors.Is(err, profile.ErrCrossUser):
		writeAPIError(w, http.StatusNotFound, sharederrors.CodeResourceNotFound, "resource not found", nil)
	case errors.Is(err, profile.ErrInvalidCursor):
		writeAPIError(w, http.StatusUnprocessableEntity, sharederrors.CodeValidationFailed, "cursor is invalid", nil)
	case errors.Is(err, profile.ErrValidationFailed):
		writeAPIError(w, http.StatusUnprocessableEntity, sharederrors.CodeValidationFailed, err.Error(), nil)
	default:
		writeAPIError(w, http.StatusInternalServerError, sharederrors.CodeValidationFailed, fallbackMessage, nil)
	}
}

func mapCandidateProfile(rec *profile.CandidateProfileRecord) api.CandidateProfile {
	if rec == nil {
		return api.CandidateProfile{}
	}
	return api.CandidateProfile{
		Headline:                  copyStringPtr(rec.Headline),
		YearsOfExperience:         copyInt32Ptr(rec.YearsOfExperience),
		CurrentRole:               copyStringPtr(rec.CurrentRole),
		PreferredPracticeLanguage: rec.PreferredPracticeLanguage,
		UiLanguage:                rec.UILanguage,
		Region:                    copyStringPtr(rec.Region),
	}
}

func mapExperienceCard(rec *profile.ExperienceCardRecord) api.ExperienceCard {
	if rec == nil {
		return api.ExperienceCard{}
	}
	skills := append([]string{}, rec.Skills...)
	return api.ExperienceCard{
		Id:          rec.ID,
		Title:       rec.Title,
		CompanyName: rec.CompanyName,
		Situation:   rec.Situation,
		Task:        rec.Task,
		Action:      rec.Action,
		Result:      rec.Result,
		Skills:      skills,
		Language:    rec.Language,
		CreatedAt:   rec.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   rec.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z"),
	}
}

func copyStringPtr(s *string) *string {
	if s == nil {
		return nil
	}
	v := *s
	return &v
}

func copyInt32Ptr(v *int32) *int32 {
	if v == nil {
		return nil
	}
	out := *v
	return &out
}
