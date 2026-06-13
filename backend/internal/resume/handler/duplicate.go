package handler

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	api "github.com/monshunter/easyinterview/backend/internal/api/generated"
	"github.com/monshunter/easyinterview/backend/internal/middleware/idempotency"
	"github.com/monshunter/easyinterview/backend/internal/resume"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
)

type DuplicateResumeService interface {
	DuplicateResume(ctx context.Context, in resume.DuplicateResumeRequest) (api.Resume, error)
}

// DuplicateResume binds POST /api/v1/resumes/{resumeId}/duplicate.
func (h *Handler) DuplicateResume(w http.ResponseWriter, r *http.Request, resumeID string) {
	if h == nil {
		writeAPIError(w, http.StatusInternalServerError, sharederrors.CodeValidationFailed, "resume service is not configured", nil)
		return
	}
	service, ok := h.service.(DuplicateResumeService)
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
	raw, err := io.ReadAll(r.Body)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, sharederrors.CodeValidationFailed, "request body is malformed", nil)
		return
	}
	in, err := validateDuplicateResumeInput(userID, resumeID, raw)
	if err != nil {
		writeAPIError(w, http.StatusUnprocessableEntity, sharederrors.CodeValidationFailed, err.Error(), updateResumeValidationDetails(err))
		return
	}
	out, err := service.DuplicateResume(r.Context(), in)
	if err != nil {
		writeUpdateResumeError(w, err)
		return
	}
	idempotency.SetResponseResource(w, "resume", out.Id)
	writeJSON(w, http.StatusCreated, out)
}

func validateDuplicateResumeInput(userID string, resumeID string, raw []byte) (resume.DuplicateResumeRequest, error) {
	in := resume.DuplicateResumeRequest{
		UserID:         strings.TrimSpace(userID),
		SourceResumeID: strings.TrimSpace(resumeID),
	}
	trimmed := strings.TrimSpace(string(raw))
	if trimmed == "" {
		return in, nil
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		return resume.DuplicateResumeRequest{}, validationError("request body is malformed")
	}
	for key := range fields {
		switch key {
		case "displayName", "structuredProfile":
		default:
			return resume.DuplicateResumeRequest{}, validationError(key + " is not editable")
		}
	}
	if rawDisplayName, ok := fields["displayName"]; ok {
		if isJSONNull(rawDisplayName) {
			return resume.DuplicateResumeRequest{}, validationError("displayName must not be null")
		}
		var displayName string
		if err := json.Unmarshal(rawDisplayName, &displayName); err != nil {
			return resume.DuplicateResumeRequest{}, validationError("displayName must be a string")
		}
		displayName = strings.TrimSpace(displayName)
		if displayName == "" {
			return resume.DuplicateResumeRequest{}, validationError("displayName must not be blank")
		}
		in.DisplayName = &displayName
		in.DisplayNameSet = true
	}
	if rawProfile, ok := fields["structuredProfile"]; ok {
		if isJSONNull(rawProfile) {
			return resume.DuplicateResumeRequest{}, validationError("structuredProfile must not be null")
		}
		var profile map[string]any
		if err := json.Unmarshal(rawProfile, &profile); err != nil {
			return resume.DuplicateResumeRequest{}, validationError("structuredProfile must be an object")
		}
		// Provenance echoed back inside structuredProfile is stripped by the
		// service before persisting (D-20).
		in.StructuredProfile = cloneMap(profile)
	}
	return in, nil
}
