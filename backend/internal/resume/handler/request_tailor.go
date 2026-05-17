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

type RequestTailorService interface {
	RequestResumeTailor(ctx context.Context, in resume.RequestTailorRunInput) (api.ResumeTailorRunWithJob, error)
}

// RequestResumeTailor binds POST /api/v1/resume/tailor.
func (h *Handler) RequestResumeTailor(w http.ResponseWriter, r *http.Request) {
	if h == nil {
		writeAPIError(w, http.StatusInternalServerError, sharederrors.CodeValidationFailed, "resume service is not configured", nil)
		return
	}
	service, ok := h.service.(RequestTailorService)
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
	var body api.RequestResumeTailorRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeAPIError(w, http.StatusBadRequest, sharederrors.CodeValidationFailed, "request body is malformed", nil)
		return
	}
	in, err := validateRequestTailorInput(userID, idempotencyKey, body)
	if err != nil {
		writeAPIError(w, http.StatusUnprocessableEntity, sharederrors.CodeValidationFailed, err.Error(), requestTailorValidationDetails(err))
		return
	}
	out, err := service.RequestResumeTailor(r.Context(), in)
	if err != nil {
		writeTailorRunError(w, err, "resume tailor request failed")
		return
	}
	idempotency.SetResponseResource(w, "resume_tailor_run", out.TailorRunId)
	writeJSON(w, http.StatusAccepted, out)
}

func validateRequestTailorInput(userID string, idempotencyKey string, body api.RequestResumeTailorRequest) (resume.RequestTailorRunInput, error) {
	targetJobID := strings.TrimSpace(body.TargetJobId)
	resumeAssetID := strings.TrimSpace(body.ResumeAssetId)
	mode := strings.TrimSpace(body.Mode)
	if targetJobID == "" {
		return resume.RequestTailorRunInput{}, validationError("targetJobId is required")
	}
	if resumeAssetID == "" {
		return resume.RequestTailorRunInput{}, validationError("resumeAssetId is required")
	}
	switch mode {
	case "gap_review", "bullet_suggestions":
	default:
		return resume.RequestTailorRunInput{}, validationError("mode is invalid")
	}
	return resume.RequestTailorRunInput{
		UserID:         strings.TrimSpace(userID),
		TargetJobID:    targetJobID,
		ResumeAssetID:  resumeAssetID,
		Mode:           mode,
		IdempotencyKey: strings.TrimSpace(idempotencyKey),
	}, nil
}

func requestTailorValidationDetails(err error) map[string]any {
	if err == nil {
		return nil
	}
	switch err.Error() {
	case "targetJobId is required":
		return map[string]any{"field": "targetJobId"}
	case "resumeAssetId is required":
		return map[string]any{"field": "resumeAssetId"}
	case "mode is invalid":
		return map[string]any{"field": "mode"}
	default:
		return nil
	}
}

func writeTailorRunError(w http.ResponseWriter, err error, fallbackMessage string) {
	switch {
	case errors.Is(err, resume.ErrValidationFailed):
		writeAPIError(w, http.StatusUnprocessableEntity, sharederrors.CodeValidationFailed, "resume validation failed", nil)
	case errors.Is(err, resume.ErrNotFound):
		writeAPIError(w, http.StatusNotFound, sharederrors.CodeTargetJobNotFound, "Resume asset, target job, or tailor run not found", nil)
	default:
		writeAPIError(w, http.StatusInternalServerError, sharederrors.CodeValidationFailed, fallbackMessage, nil)
	}
}
