package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	api "github.com/monshunter/easyinterview/backend/internal/api/generated"
	"github.com/monshunter/easyinterview/backend/internal/middleware/idempotency"
	"github.com/monshunter/easyinterview/backend/internal/resume"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
)

type ConfirmStructuredMasterService interface {
	ConfirmStructuredMaster(ctx context.Context, in resume.ConfirmStructuredMasterInput) (api.ResumeVersion, error)
}

// ConfirmResumeStructuredMaster binds
// POST /api/v1/resumes/{resumeAssetId}/structured-master.
func (h *Handler) ConfirmResumeStructuredMaster(w http.ResponseWriter, r *http.Request, resumeAssetID string) {
	if h == nil {
		writeAPIError(w, http.StatusInternalServerError, sharederrors.CodeValidationFailed, "resume service is not configured", nil)
		return
	}
	service, ok := h.service.(ConfirmStructuredMasterService)
	if !ok {
		writeAPIError(w, http.StatusInternalServerError, sharederrors.CodeValidationFailed, "resume service is not configured", nil)
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
	var body api.ConfirmResumeStructuredMasterRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeAPIError(w, http.StatusBadRequest, sharederrors.CodeValidationFailed, "request body is malformed", nil)
		return
	}
	in, err := validateConfirmStructuredMasterInput(userID, resumeAssetID, body)
	if err != nil {
		writeAPIError(w, http.StatusUnprocessableEntity, sharederrors.CodeValidationFailed, err.Error(), nil)
		return
	}
	out, err := service.ConfirmStructuredMaster(r.Context(), in)
	if err != nil {
		writeConfirmStructuredMasterError(w, err)
		return
	}
	idempotency.SetResponseResource(w, "resume_version", out.Id)
	writeJSON(w, http.StatusCreated, out)
}

func validateConfirmStructuredMasterInput(userID string, resumeAssetID string, body api.ConfirmResumeStructuredMasterRequest) (resume.ConfirmStructuredMasterInput, error) {
	displayName := strings.TrimSpace(body.DisplayName)
	if displayName == "" {
		return resume.ConfirmStructuredMasterInput{}, validationError("displayName must not be blank")
	}
	profile, ok := body.StructuredProfile.(map[string]any)
	if !ok || len(profile) == 0 {
		return resume.ConfirmStructuredMasterInput{}, validationError("structuredProfile is required")
	}
	if provenance, ok := profile["provenance"].(map[string]any); !ok || len(provenance) == 0 {
		return resume.ConfirmStructuredMasterInput{}, validationError("structuredProfile.provenance is required")
	}
	language := ""
	if body.Language != nil {
		language = strings.TrimSpace(*body.Language)
	}
	return resume.ConfirmStructuredMasterInput{
		UserID:            strings.TrimSpace(userID),
		ResumeAssetID:     strings.TrimSpace(resumeAssetID),
		DisplayName:       displayName,
		Language:          language,
		StructuredProfile: cloneMap(profile),
	}, nil
}

func writeConfirmStructuredMasterError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, resume.ErrStructuredMasterAlreadyExists):
		writeAPIError(w, http.StatusConflict, sharederrors.CodeResumeStructuredMasterAlreadyExists, "structured master resume version already exists for this resume asset", nil)
	case errors.Is(err, resume.ErrAssetParseNotReady):
		writeAPIError(w, http.StatusUnprocessableEntity, sharederrors.CodeValidationFailed, "resume asset parse is not ready", map[string]any{"reason": "PARSE_NOT_READY"})
	case errors.Is(err, resume.ErrValidationFailed):
		writeAPIError(w, http.StatusUnprocessableEntity, sharederrors.CodeValidationFailed, "resume validation failed", nil)
	case errors.Is(err, resume.ErrNotFound):
		writeAPIError(w, http.StatusNotFound, sharederrors.CodeTargetJobNotFound, "resume not found", nil)
	default:
		writeAPIError(w, http.StatusInternalServerError, sharederrors.CodeValidationFailed, "resume structured master create failed", nil)
	}
}
