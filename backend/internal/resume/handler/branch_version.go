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
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

type BranchVersionService interface {
	BranchResumeVersion(ctx context.Context, in resume.BranchVersionRequest) (resume.BranchVersionResult, error)
}

// BranchResumeVersion binds POST /api/v1/resume-versions.
func (h *Handler) BranchResumeVersion(w http.ResponseWriter, r *http.Request) {
	if h == nil {
		writeAPIError(w, http.StatusInternalServerError, sharederrors.CodeValidationFailed, "resume service is not configured", nil)
		return
	}
	service, ok := h.service.(BranchVersionService)
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
	var body api.BranchResumeVersionRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeAPIError(w, http.StatusBadRequest, sharederrors.CodeValidationFailed, "request body is malformed", nil)
		return
	}
	in, err := validateBranchVersionInput(userID, idempotencyKey, body)
	if err != nil {
		writeAPIError(w, http.StatusUnprocessableEntity, sharederrors.CodeValidationFailed, err.Error(), branchVersionValidationDetails(err))
		return
	}
	out, err := service.BranchResumeVersion(r.Context(), in)
	if err != nil {
		writeBranchVersionError(w, err)
		return
	}
	writeBranchVersionResult(w, out)
}

func validateBranchVersionInput(userID string, idempotencyKey string, body api.BranchResumeVersionRequest) (resume.BranchVersionRequest, error) {
	parentVersionID := strings.TrimSpace(body.ParentVersionId)
	targetJobID := strings.TrimSpace(body.TargetJobId)
	seedStrategy := sharedtypes.ResumeSeedStrategy(strings.TrimSpace(string(body.SeedStrategy)))
	if parentVersionID == "" {
		return resume.BranchVersionRequest{}, validationError("parentVersionId is required")
	}
	if targetJobID == "" {
		return resume.BranchVersionRequest{}, validationError("targetJobId is required")
	}
	switch seedStrategy {
	case sharedtypes.ResumeSeedStrategyCopyMaster, sharedtypes.ResumeSeedStrategyBlank, sharedtypes.ResumeSeedStrategyAiSelect:
	default:
		return resume.BranchVersionRequest{}, validationError("seedStrategy is invalid")
	}
	if body.DisplayName == nil || strings.TrimSpace(*body.DisplayName) == "" {
		return resume.BranchVersionRequest{}, validationError("Display name is required for this branch request")
	}
	displayName := strings.TrimSpace(*body.DisplayName)
	var focusAngle *string
	if body.FocusAngle != nil {
		trimmed := strings.TrimSpace(*body.FocusAngle)
		if trimmed != "" {
			focusAngle = &trimmed
		}
	}
	return resume.BranchVersionRequest{
		UserID:          strings.TrimSpace(userID),
		ParentVersionID: parentVersionID,
		TargetJobID:     targetJobID,
		SeedStrategy:    seedStrategy,
		DisplayName:     displayName,
		FocusAngle:      focusAngle,
		IdempotencyKey:  strings.TrimSpace(idempotencyKey),
	}, nil
}

func branchVersionValidationDetails(err error) map[string]any {
	if err == nil {
		return nil
	}
	switch err.Error() {
	case "Display name is required for this branch request":
		return map[string]any{"field": "displayName"}
	case "parentVersionId is required":
		return map[string]any{"field": "parentVersionId"}
	case "targetJobId is required":
		return map[string]any{"field": "targetJobId"}
	case "seedStrategy is invalid":
		return map[string]any{"field": "seedStrategy"}
	default:
		return nil
	}
}

func writeBranchVersionResult(w http.ResponseWriter, out resume.BranchVersionResult) {
	if out.Accepted != nil {
		status := out.Status
		if status == 0 {
			status = http.StatusAccepted
		}
		idempotency.SetResponseResource(w, "resume_version", out.Accepted.ResumeVersionId)
		writeJSON(w, status, out.Accepted)
		return
	}
	status := out.Status
	if status == 0 {
		status = http.StatusCreated
	}
	idempotency.SetResponseResource(w, "resume_version", out.Version.Id)
	writeJSON(w, status, out.Version)
}

func writeBranchVersionError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, resume.ErrValidationFailed):
		writeAPIError(w, http.StatusUnprocessableEntity, sharederrors.CodeValidationFailed, "resume validation failed", nil)
	case errors.Is(err, resume.ErrNotFound):
		writeAPIError(w, http.StatusNotFound, sharederrors.CodeTargetJobNotFound, "Resume version or target job not found", nil)
	default:
		writeAPIError(w, http.StatusInternalServerError, sharederrors.CodeValidationFailed, "resume version branch failed", nil)
	}
}
