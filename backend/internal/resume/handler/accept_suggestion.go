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

type SuggestionDecisionService interface {
	AcceptResumeTailorSuggestion(ctx context.Context, in resume.SuggestionDecisionRequest) (api.ResumeVersion, error)
	RejectResumeTailorSuggestion(ctx context.Context, in resume.SuggestionDecisionRequest) (api.ResumeVersion, error)
}

// AcceptResumeTailorSuggestion binds
// POST /api/v1/resume-versions/{resumeVersionId}/suggestions/{suggestionId}/accept.
func (h *Handler) AcceptResumeTailorSuggestion(w http.ResponseWriter, r *http.Request, resumeVersionID string, suggestionID string) {
	h.decideResumeTailorSuggestion(w, r, resumeVersionID, suggestionID, true)
}

func (h *Handler) decideResumeTailorSuggestion(w http.ResponseWriter, r *http.Request, resumeVersionID string, suggestionID string, accept bool) {
	if h == nil {
		writeAPIError(w, http.StatusInternalServerError, sharederrors.CodeValidationFailed, "resume service is not configured", nil)
		return
	}
	service, ok := h.service.(SuggestionDecisionService)
	if !ok {
		writeAPIError(w, http.StatusInternalServerError, sharederrors.CodeValidationFailed, "resume service is not configured", nil)
		return
	}
	userID, ok := h.resolveUser(r)
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, sharederrors.CodeAuthUnauthorized, "authentication required", nil)
		return
	}
	idempotencyKey := strings.TrimSpace(r.Header.Get(idempotency.HeaderName))
	if idempotencyKey == "" {
		writeAPIError(w, http.StatusUnprocessableEntity, sharederrors.CodeValidationFailed, "Idempotency-Key header is required", nil)
		return
	}
	in, err := validateSuggestionDecisionInput(userID, resumeVersionID, suggestionID, idempotencyKey)
	if err != nil {
		writeAPIError(w, http.StatusUnprocessableEntity, sharederrors.CodeValidationFailed, err.Error(), suggestionDecisionValidationDetails(err))
		return
	}
	var out api.ResumeVersion
	if accept {
		out, err = service.AcceptResumeTailorSuggestion(r.Context(), in)
	} else {
		out, err = service.RejectResumeTailorSuggestion(r.Context(), in)
	}
	if err != nil {
		writeSuggestionDecisionError(w, err)
		return
	}
	idempotency.SetResponseResource(w, "resume_version", out.Id)
	writeJSON(w, http.StatusOK, out)
}

func validateSuggestionDecisionInput(userID string, resumeVersionID string, suggestionID string, idempotencyKey string) (resume.SuggestionDecisionRequest, error) {
	resumeVersionID = strings.TrimSpace(resumeVersionID)
	suggestionID = strings.TrimSpace(suggestionID)
	if resumeVersionID == "" {
		return resume.SuggestionDecisionRequest{}, validationError("resumeVersionId is required")
	}
	if suggestionID == "" {
		return resume.SuggestionDecisionRequest{}, validationError("suggestionId is required")
	}
	return resume.SuggestionDecisionRequest{
		UserID:          strings.TrimSpace(userID),
		ResumeVersionID: resumeVersionID,
		SuggestionID:    suggestionID,
		IdempotencyKey:  strings.TrimSpace(idempotencyKey),
	}, nil
}

func suggestionDecisionValidationDetails(err error) map[string]any {
	if err == nil {
		return nil
	}
	switch err.Error() {
	case "resumeVersionId is required":
		return map[string]any{"field": "resumeVersionId"}
	case "suggestionId is required":
		return map[string]any{"field": "suggestionId"}
	default:
		return nil
	}
}

func writeSuggestionDecisionError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, resume.ErrSuggestionAlreadyDecided):
		writeAPIError(w, http.StatusConflict, sharederrors.CodeValidationFailed, "Suggestion has already been decided", map[string]any{"reason": "SUGGESTION_ALREADY_DECIDED"})
	case errors.Is(err, resume.ErrValidationFailed):
		writeAPIError(w, http.StatusUnprocessableEntity, sharederrors.CodeValidationFailed, "resume validation failed", nil)
	case errors.Is(err, resume.ErrNotFound):
		writeAPIError(w, http.StatusNotFound, sharederrors.CodeTargetJobNotFound, "Resume version or suggestion not found", nil)
	default:
		writeAPIError(w, http.StatusInternalServerError, sharederrors.CodeValidationFailed, "resume suggestion decision failed", nil)
	}
}
