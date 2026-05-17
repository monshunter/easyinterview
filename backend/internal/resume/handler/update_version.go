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

type VersionUpdateService interface {
	UpdateResumeVersion(ctx context.Context, in resume.UpdateVersionRequest) (api.ResumeVersion, error)
}

// UpdateResumeVersion binds PATCH /api/v1/resume-versions/{resumeVersionId}.
func (h *Handler) UpdateResumeVersion(w http.ResponseWriter, r *http.Request, resumeVersionID string) {
	if h == nil {
		writeAPIError(w, http.StatusInternalServerError, sharederrors.CodeValidationFailed, "resume service is not configured", nil)
		return
	}
	service, ok := h.service.(VersionUpdateService)
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
	in, err := validateUpdateVersionInput(userID, resumeVersionID, raw)
	if err != nil {
		writeAPIError(w, http.StatusUnprocessableEntity, sharederrors.CodeValidationFailed, err.Error(), updateVersionValidationDetails(err))
		return
	}
	out, err := service.UpdateResumeVersion(r.Context(), in)
	if err != nil {
		writeUpdateVersionError(w, err)
		return
	}
	idempotency.SetResponseResource(w, "resume_version", out.Id)
	writeJSON(w, http.StatusOK, out)
}

func validateUpdateVersionInput(userID string, resumeVersionID string, raw []byte) (resume.UpdateVersionRequest, error) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		return resume.UpdateVersionRequest{}, validationError("request body is malformed")
	}
	if len(fields) == 0 {
		return resume.UpdateVersionRequest{}, validationError("at least one editable field is required")
	}
	for key := range fields {
		switch key {
		case "displayName", "focusAngle", "matchScore", "structuredProfile":
		default:
			return resume.UpdateVersionRequest{}, validationError(key + " is not editable")
		}
	}
	in := resume.UpdateVersionRequest{
		UserID:    strings.TrimSpace(userID),
		VersionID: strings.TrimSpace(resumeVersionID),
	}
	if rawDisplayName, ok := fields["displayName"]; ok {
		if isJSONNull(rawDisplayName) {
			return resume.UpdateVersionRequest{}, validationError("displayName must not be null")
		}
		var displayName string
		if err := json.Unmarshal(rawDisplayName, &displayName); err != nil {
			return resume.UpdateVersionRequest{}, validationError("displayName must be a string")
		}
		displayName = strings.TrimSpace(displayName)
		if displayName == "" {
			return resume.UpdateVersionRequest{}, validationError("displayName must not be blank")
		}
		in.DisplayName = &displayName
		in.DisplayNameSet = true
	}
	if rawFocusAngle, ok := fields["focusAngle"]; ok {
		in.FocusAngleSet = true
		if !isJSONNull(rawFocusAngle) {
			var focusAngle string
			if err := json.Unmarshal(rawFocusAngle, &focusAngle); err != nil {
				return resume.UpdateVersionRequest{}, validationError("focusAngle must be a string or null")
			}
			focusAngle = strings.TrimSpace(focusAngle)
			in.FocusAngle = &focusAngle
		}
	}
	if rawMatchScore, ok := fields["matchScore"]; ok {
		in.MatchScoreSet = true
		if !isJSONNull(rawMatchScore) {
			var matchScore float64
			if err := json.Unmarshal(rawMatchScore, &matchScore); err != nil {
				return resume.UpdateVersionRequest{}, validationError("matchScore must be a number or null")
			}
			in.MatchScore = &matchScore
		}
	}
	if rawProfile, ok := fields["structuredProfile"]; ok {
		if isJSONNull(rawProfile) {
			return resume.UpdateVersionRequest{}, validationError("structuredProfile must not be null")
		}
		var profile map[string]any
		if err := json.Unmarshal(rawProfile, &profile); err != nil {
			return resume.UpdateVersionRequest{}, validationError("structuredProfile must be an object")
		}
		if provenance, ok := profile["provenance"]; ok && provenance != nil {
			return resume.UpdateVersionRequest{}, validationError("structuredProfile.provenance is not editable")
		}
		in.StructuredProfile = cloneMap(profile)
		in.StructuredProfileSet = true
	}
	return in, nil
}

func updateVersionValidationDetails(err error) map[string]any {
	if err == nil {
		return nil
	}
	message := err.Error()
	switch {
	case strings.HasPrefix(message, "displayName "):
		return map[string]any{"field": "displayName"}
	case strings.HasPrefix(message, "focusAngle "):
		return map[string]any{"field": "focusAngle"}
	case strings.HasPrefix(message, "matchScore "):
		return map[string]any{"field": "matchScore"}
	case strings.HasPrefix(message, "structuredProfile."):
		return map[string]any{"field": strings.TrimSuffix(strings.Split(message, " ")[0], ".")}
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

func writeUpdateVersionError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, resume.ErrValidationFailed):
		writeAPIError(w, http.StatusUnprocessableEntity, sharederrors.CodeValidationFailed, "resume validation failed", nil)
	case errors.Is(err, resume.ErrNotFound):
		writeAPIError(w, http.StatusNotFound, sharederrors.CodeTargetJobNotFound, "Resume version not found", nil)
	default:
		writeAPIError(w, http.StatusInternalServerError, sharederrors.CodeValidationFailed, "resume version update failed", nil)
	}
}
