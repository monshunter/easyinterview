package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	api "github.com/monshunter/easyinterview/backend/internal/api/generated"
	"github.com/monshunter/easyinterview/backend/internal/middleware/idempotency"
	"github.com/monshunter/easyinterview/backend/internal/resume"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
)

type UpdateResumeService interface {
	UpdateResume(ctx context.Context, in resume.UpdateResumeRequest) (api.Resume, error)
}

// UpdateResume binds PATCH /api/v1/resumes/{resumeId}.
func (h *Handler) UpdateResume(w http.ResponseWriter, r *http.Request, resumeID string) {
	if h == nil {
		writeAPIError(w, http.StatusInternalServerError, sharederrors.CodeValidationFailed, "resume service is not configured", nil)
		return
	}
	service, ok := h.service.(UpdateResumeService)
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
	in, err := validateUpdateResumeInput(userID, resumeID, raw)
	if err != nil {
		writeAPIError(w, http.StatusUnprocessableEntity, sharederrors.CodeValidationFailed, err.Error(), updateResumeValidationDetails(err))
		return
	}
	out, err := service.UpdateResume(r.Context(), in)
	if err != nil {
		writeUpdateResumeError(w, err)
		return
	}
	idempotency.SetResponseResource(w, "resume", out.Id)
	writeJSON(w, http.StatusOK, out)
}

func validateUpdateResumeInput(userID string, resumeID string, raw []byte) (resume.UpdateResumeRequest, error) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		return resume.UpdateResumeRequest{}, validationError("request body is malformed")
	}
	if len(fields) == 0 {
		return resume.UpdateResumeRequest{}, validationError("at least one editable field is required")
	}
	for key := range fields {
		switch key {
		case "displayName", "structuredProfile":
		default:
			return resume.UpdateResumeRequest{}, validationError(key + " is not editable")
		}
	}
	in := resume.UpdateResumeRequest{
		UserID:   strings.TrimSpace(userID),
		ResumeID: strings.TrimSpace(resumeID),
	}
	if rawDisplayName, ok := fields["displayName"]; ok {
		if isJSONNull(rawDisplayName) {
			return resume.UpdateResumeRequest{}, validationError("displayName must not be null")
		}
		var displayName string
		if err := json.Unmarshal(rawDisplayName, &displayName); err != nil {
			return resume.UpdateResumeRequest{}, validationError("displayName must be a string")
		}
		displayName = strings.TrimSpace(displayName)
		if displayName == "" {
			return resume.UpdateResumeRequest{}, validationError("displayName must not be blank")
		}
		in.DisplayName = &displayName
		in.DisplayNameSet = true
	}
	if rawProfile, ok := fields["structuredProfile"]; ok {
		if isJSONNull(rawProfile) {
			return resume.UpdateResumeRequest{}, validationError("structuredProfile must not be null")
		}
		var profile map[string]any
		if err := json.Unmarshal(rawProfile, &profile); err != nil {
			return resume.UpdateResumeRequest{}, validationError("structuredProfile must be an object")
		}
		// Clients may echo the AI-generated provenance back inside the saved
		// structuredProfile; the service strips it before persisting (D-20:
		// resume.tailor suggestions are ephemeral, provenance is not editable).
		in.StructuredProfile = cloneMap(profile)
		in.StructuredProfileSet = true
	}
	return in, nil
}

func updateResumeValidationDetails(err error) map[string]any {
	if err == nil {
		return nil
	}
	message := err.Error()
	switch {
	case strings.HasPrefix(message, "displayName "):
		return map[string]any{"field": "displayName"}
	case strings.HasPrefix(message, "structuredProfile "):
		return map[string]any{"field": "structuredProfile"}
	case strings.HasSuffix(message, " is not editable"):
		return map[string]any{"fields": []string{strings.TrimSuffix(message, " is not editable")}}
	default:
		return nil
	}
}

func isJSONNull(raw json.RawMessage) bool {
	return bytes.Equal(bytes.TrimSpace(raw), []byte("null"))
}

func writeUpdateResumeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, resume.ErrValidationFailed):
		writeAPIError(w, http.StatusUnprocessableEntity, sharederrors.CodeValidationFailed, "resume validation failed", nil)
	case errors.Is(err, resume.ErrNotFound):
		writeAPIError(w, http.StatusNotFound, sharederrors.CodeResourceNotFound, "resume not found", nil)
	default:
		writeAPIError(w, http.StatusInternalServerError, sharederrors.CodeValidationFailed, "resume update failed", nil)
	}
}
