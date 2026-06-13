package handler

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	api "github.com/monshunter/easyinterview/backend/internal/api/generated"
	"github.com/monshunter/easyinterview/backend/internal/resume"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
)

type ListService interface {
	ListResumes(ctx context.Context, in resume.ListRequest) (api.PaginatedResume, error)
}

// ListResumes binds GET /api/v1/resumes.
func (h *Handler) ListResumes(w http.ResponseWriter, r *http.Request) {
	if h == nil {
		writeAPIError(w, http.StatusInternalServerError, sharederrors.CodeValidationFailed, "resume service is not configured", nil)
		return
	}
	service, ok := h.service.(ListService)
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
	in := resume.ListRequest{
		UserID: userID,
		Cursor: q.Get("cursor"),
	}
	if pageSizeStr := strings.TrimSpace(q.Get("pageSize")); pageSizeStr != "" {
		var n int
		if _, err := fmtSscanInt(pageSizeStr, &n); err == nil {
			in.PageSize = n
		}
	}
	out, err := service.ListResumes(r.Context(), in)
	if err != nil {
		writeResumeServiceError(w, err, "resume list failed")
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func fmtSscanInt(in string, target *int) (int, error) {
	var v int
	n, err := fmt.Sscanf(in, "%d", &v)
	if err == nil {
		*target = v
	}
	return n, err
}
