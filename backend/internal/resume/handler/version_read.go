package handler

import (
	"context"
	"errors"
	"net/http"
	"strings"

	api "github.com/monshunter/easyinterview/backend/internal/api/generated"
	"github.com/monshunter/easyinterview/backend/internal/resume"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
)

type VersionGetService interface {
	GetResumeVersion(ctx context.Context, userID string, versionID string) (api.ResumeVersion, error)
}

type VersionListService interface {
	ListResumeVersions(ctx context.Context, in resume.ListVersionRequest) (api.PaginatedResumeVersion, error)
}

// GetResumeVersion binds GET /api/v1/resume-versions/{resumeVersionId}.
func (h *Handler) GetResumeVersion(w http.ResponseWriter, r *http.Request, resumeVersionID string) {
	if h == nil {
		writeAPIError(w, http.StatusInternalServerError, sharederrors.CodeValidationFailed, "resume service is not configured", nil)
		return
	}
	service, ok := h.service.(VersionGetService)
	if !ok {
		writeAPIError(w, http.StatusInternalServerError, sharederrors.CodeValidationFailed, "resume service is not configured", nil)
		return
	}
	userID, ok := h.resolveUser(r)
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, sharederrors.CodeAuthUnauthorized, "authentication required", nil)
		return
	}
	out, err := service.GetResumeVersion(r.Context(), userID, resumeVersionID)
	if err != nil {
		writeVersionReadError(w, err, "Resume version not found", "resume version get failed")
		return
	}
	writeJSON(w, http.StatusOK, out)
}

// ListResumeVersions binds GET /api/v1/resumes/{resumeAssetId}/versions.
func (h *Handler) ListResumeVersions(w http.ResponseWriter, r *http.Request, resumeAssetID string) {
	if h == nil {
		writeAPIError(w, http.StatusInternalServerError, sharederrors.CodeValidationFailed, "resume service is not configured", nil)
		return
	}
	service, ok := h.service.(VersionListService)
	if !ok {
		writeAPIError(w, http.StatusInternalServerError, sharederrors.CodeValidationFailed, "resume service is not configured", nil)
		return
	}
	userID, ok := h.resolveUser(r)
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, sharederrors.CodeAuthUnauthorized, "authentication required", nil)
		return
	}
	q := r.URL.Query()
	in := resume.ListVersionRequest{
		UserID:        userID,
		ResumeAssetID: strings.TrimSpace(resumeAssetID),
		Cursor:        q.Get("cursor"),
	}
	if pageSizeStr := strings.TrimSpace(q.Get("pageSize")); pageSizeStr != "" {
		var n int
		if _, err := fmtSscanInt(pageSizeStr, &n); err == nil {
			in.PageSize = n
		}
	}
	out, err := service.ListResumeVersions(r.Context(), in)
	if err != nil {
		writeVersionReadError(w, err, "Resume version not found", "resume version list failed")
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func writeVersionReadError(w http.ResponseWriter, err error, notFoundMessage string, fallbackMessage string) {
	if errors.Is(err, resume.ErrNotFound) {
		writeAPIError(w, http.StatusNotFound, sharederrors.CodeTargetJobNotFound, notFoundMessage, nil)
		return
	}
	if errors.Is(err, resume.ErrInvalidCursor) {
		writeAPIError(w, http.StatusUnprocessableEntity, sharederrors.CodeValidationFailed, "resume validation failed", nil)
		return
	}
	writeAPIError(w, http.StatusInternalServerError, sharederrors.CodeValidationFailed, fallbackMessage, nil)
}
