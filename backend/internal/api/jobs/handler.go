package jobs

import (
	"context"
	"encoding/json"
	stderrs "errors"
	"net/http"
	"strings"

	api "github.com/monshunter/easyinterview/backend/internal/api/generated"
	domain "github.com/monshunter/easyinterview/backend/internal/jobs"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
)

type SessionResolver func(ctx context.Context) (userID string, ok bool)

type jobService interface {
	GetJob(ctx context.Context, userID, jobID string) (domain.JobRecord, error)
}

type HandlerOptions struct {
	Service jobService
	Session SessionResolver
}

type Handler struct {
	service jobService
	session SessionResolver
}

func NewHandler(opts HandlerOptions) *Handler {
	return &Handler{service: opts.Service, session: opts.Session}
}

func (h *Handler) GetJob(w http.ResponseWriter, r *http.Request, jobID string) {
	if h == nil || h.service == nil {
		writeAPIError(w, http.StatusInternalServerError, sharederrors.CodeValidationFailed, "job service is not configured", nil)
		return
	}
	userID, ok := h.resolveUser(r)
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, sharederrors.CodeAuthUnauthorized, "authentication required", nil)
		return
	}
	jobID = strings.TrimSpace(jobID)
	if jobID == "" {
		writeAPIError(w, http.StatusNotFound, sharederrors.CodeValidationFailed, "job not found or not accessible", nil)
		return
	}
	result, err := h.service.GetJob(r.Context(), userID, jobID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toAPIJob(result))
}

func (h *Handler) resolveUser(r *http.Request) (string, bool) {
	if h.session == nil {
		return "", false
	}
	userID, ok := h.session(r.Context())
	userID = strings.TrimSpace(userID)
	return userID, ok && userID != ""
}

func writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case stderrs.Is(err, domain.ErrJobNotFound):
		writeAPIError(w, http.StatusNotFound, sharederrors.CodeValidationFailed, "job not found or not accessible", nil)
	default:
		writeAPIError(w, http.StatusInternalServerError, sharederrors.CodeValidationFailed, "job request failed", nil)
	}
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

func toAPIJob(in domain.JobRecord) api.Job {
	job := api.Job{
		Id:           in.ID,
		JobType:      in.JobType,
		ResourceType: in.ResourceType,
		ResourceId:   in.ResourceID,
		Status:       in.Status,
		CreatedAt:    in.CreatedAt.UTC().Format(timeFormatRFC3339),
		UpdatedAt:    in.UpdatedAt.UTC().Format(timeFormatRFC3339),
	}
	if strings.TrimSpace(in.ErrorCode) != "" {
		code := api.ApiErrorCode(in.ErrorCode)
		job.ErrorCode = &code
	}
	return job
}

const timeFormatRFC3339 = "2006-01-02T15:04:05Z07:00"
