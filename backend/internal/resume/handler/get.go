package handler

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	api "github.com/monshunter/easyinterview/backend/internal/api/generated"
	"github.com/monshunter/easyinterview/backend/internal/resume"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
)

type GetService interface {
	GetResume(ctx context.Context, userID string, resumeID string) (api.Resume, error)
}

type SourceService interface {
	GetResumeSource(ctx context.Context, userID string, resumeID string) (resume.SourceFile, error)
}

// GetResume binds GET /api/v1/resumes/{resumeId}.
func (h *Handler) GetResume(w http.ResponseWriter, r *http.Request, resumeID string) {
	if h == nil {
		writeAPIError(w, http.StatusInternalServerError, sharederrors.CodeValidationFailed, "resume service is not configured", nil)
		return
	}
	service, ok := h.service.(GetService)
	if !ok {
		writeAPIError(w, http.StatusInternalServerError, sharederrors.CodeValidationFailed, "resume service is not configured", nil)
		return
	}
	userID, ok := h.resolveUser(r)
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, sharederrors.CodeAuthUnauthorized, "authentication required", nil)
		return
	}
	out, err := service.GetResume(r.Context(), userID, resumeID)
	if err != nil {
		if errors.Is(err, resume.ErrNotFound) {
			writeAPIError(w, http.StatusNotFound, sharederrors.CodeResourceNotFound, "resume not found", nil)
			return
		}
		writeAPIError(w, http.StatusInternalServerError, sharederrors.CodeValidationFailed, "resume get failed", nil)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

// GetResumeSource binds GET /api/v1/resumes/{resumeId}/source.
func (h *Handler) GetResumeSource(w http.ResponseWriter, r *http.Request, resumeID string) {
	if h == nil {
		writeAPIError(w, http.StatusInternalServerError, sharederrors.CodeValidationFailed, "resume service is not configured", nil)
		return
	}
	service, ok := h.service.(SourceService)
	if !ok {
		writeAPIError(w, http.StatusInternalServerError, sharederrors.CodeValidationFailed, "resume service is not configured", nil)
		return
	}
	userID, ok := h.resolveUser(r)
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, sharederrors.CodeAuthUnauthorized, "authentication required", nil)
		return
	}
	out, err := service.GetResumeSource(r.Context(), userID, resumeID)
	if err != nil {
		if errors.Is(err, resume.ErrNotFound) {
			writeAPIError(w, http.StatusNotFound, sharederrors.CodeResourceNotFound, "resume source not found", nil)
			return
		}
		writeResumeServiceError(w, err, "resume source get failed")
		return
	}
	fileName := strings.TrimSpace(out.FileName)
	if fileName == "" {
		fileName = "resume.pdf"
	}
	contentType := strings.TrimSpace(out.ContentType)
	if contentType == "" {
		contentType = "application/pdf"
	}
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", "inline; filename="+strconv.Quote(fileName))
	w.Header().Set("Cache-Control", "private, no-store")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(out.Body)
}
