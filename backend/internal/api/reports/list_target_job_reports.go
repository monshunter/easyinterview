package reports

import (
	"fmt"
	"net/http"
	"strings"

	api "github.com/monshunter/easyinterview/backend/internal/api/generated"
	reviewdomain "github.com/monshunter/easyinterview/backend/internal/review"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
)

func (h *Handler) ListTargetJobReports(w http.ResponseWriter, r *http.Request, targetJobID string) {
	if h == nil || h.service == nil {
		writeAPIError(w, http.StatusInternalServerError, sharederrors.CodeValidationFailed, "report service is not configured", nil)
		return
	}
	userID, ok := h.resolveUser(r)
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, sharederrors.CodeAuthUnauthorized, "authentication required", nil)
		return
	}
	q := r.URL.Query()
	req := reviewdomain.ListTargetJobReportsRequest{
		UserID:      userID,
		TargetJobID: targetJobID,
		Cursor:      q.Get("cursor"),
	}
	if pageSize := strings.TrimSpace(q.Get("pageSize")); pageSize != "" {
		var n int
		if _, err := fmt.Sscanf(pageSize, "%d", &n); err == nil {
			req.PageSize = n
		}
	}
	res, err := h.service.ListTargetJobReports(r.Context(), req)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	out := api.PaginatedFeedbackReport{
		Items: make([]api.FeedbackReport, 0, len(res.Items)),
		PageInfo: api.PageInfo{
			NextCursor: optionalString(res.PageInfo.NextCursor),
			PageSize:   res.PageInfo.PageSize,
			HasMore:    res.PageInfo.HasMore,
		},
	}
	for _, item := range res.Items {
		out.Items = append(out.Items, toAPIFeedbackReport(item))
	}
	writeJSON(w, http.StatusOK, out)
}
