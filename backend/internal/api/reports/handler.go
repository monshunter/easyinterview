package reports

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	api "github.com/monshunter/easyinterview/backend/internal/api/generated"
	reviewdomain "github.com/monshunter/easyinterview/backend/internal/review"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
)

type SessionResolver func(ctx context.Context) (userID string, ok bool)

type reportService interface {
	GetFeedbackReport(ctx context.Context, userID, reportID string) (reviewdomain.FeedbackReportRecord, error)
	ListTargetJobReports(ctx context.Context, in reviewdomain.ListTargetJobReportsRequest) (reviewdomain.PaginatedFeedbackReportRecord, error)
}

type HandlerOptions struct {
	Service reportService
	Session SessionResolver
}

type Handler struct {
	service reportService
	session SessionResolver
}

func NewHandler(opts HandlerOptions) *Handler {
	return &Handler{service: opts.Service, session: opts.Session}
}

func (h *Handler) resolveUser(r *http.Request) (string, bool) {
	if h.session == nil {
		return "", false
	}
	userID, ok := h.session(r.Context())
	userID = strings.TrimSpace(userID)
	return userID, ok && userID != ""
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	raw, err := json.Marshal(body)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, sharederrors.CodeValidationFailed, "response encoding failed", nil)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(raw)
}

func writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, reviewdomain.ErrReportNotFound):
		writeAPIError(w, http.StatusNotFound, sharederrors.CodeReportNotFound, "feedback report not found or not accessible", nil)
	case errors.Is(err, reviewdomain.ErrInvalidCursor):
		writeAPIError(w, http.StatusBadRequest, sharederrors.CodeValidationFailed, "cursor is invalid", map[string]any{"field": "cursor"})
	default:
		writeAPIError(w, http.StatusInternalServerError, sharederrors.CodeValidationFailed, "report request failed", nil)
	}
}

func writeAPIError(w http.ResponseWriter, status int, code string, message string, details map[string]any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	meta := sharederrors.CodeRegistry[code]
	raw, _ := json.Marshal(api.ApiErrorResponse{
		Error: api.ApiError{
			Code:      code,
			Message:   message,
			RequestID: "",
			Retryable: meta.Retryable,
			Details:   details,
		},
	})
	_, _ = w.Write(raw)
}

func optionalString(value string) *string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return &value
}

const timeFormatRFC3339 = "2006-01-02T15:04:05Z07:00"
