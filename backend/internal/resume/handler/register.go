package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	api "github.com/monshunter/easyinterview/backend/internal/api/generated"
	"github.com/monshunter/easyinterview/backend/internal/middleware/idempotency"
	"github.com/monshunter/easyinterview/backend/internal/resume"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
)

// SessionResolver returns the authenticated user id for the request.
type SessionResolver func(ctx context.Context) (userID string, ok bool)

type RegisterService interface {
	RegisterResume(ctx context.Context, in resume.RegisterInput) (api.ResumeWithJob, error)
}

type Options struct {
	Service RegisterService
	Session SessionResolver
}

type Handler struct {
	service RegisterService
	session SessionResolver
}

func New(opts Options) *Handler {
	return &Handler{service: opts.Service, session: opts.Session}
}

// RegisterResume binds POST /api/v1/resumes.
func (h *Handler) RegisterResume(w http.ResponseWriter, r *http.Request) {
	if h == nil || h.service == nil {
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
		writeAPIError(w, http.StatusBadRequest, sharederrors.CodeValidationFailed, "Idempotency-Key header is required", nil)
		return
	}
	var body api.RegisterResumeRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeAPIError(w, http.StatusBadRequest, sharederrors.CodeValidationFailed, "request body is malformed", nil)
		return
	}
	in, err := validateRegisterInput(userID, idempotencyKey, body)
	if err != nil {
		writeAPIError(w, http.StatusUnprocessableEntity, sharederrors.CodeValidationFailed, err.Error(), nil)
		return
	}
	out, err := h.service.RegisterResume(r.Context(), in)
	if err != nil {
		writeResumeServiceError(w, err, "resume register failed")
		return
	}
	writeJSON(w, http.StatusAccepted, out)
}

func validateRegisterInput(userID string, idempotencyKey string, body api.RegisterResumeRequest) (resume.RegisterInput, error) {
	title := strings.TrimSpace(body.Title)
	language := strings.TrimSpace(body.Language)
	sourceType := ""
	if body.SourceType != nil {
		sourceType = strings.TrimSpace(*body.SourceType)
	}
	fileObjectID := ""
	if body.FileObjectId != nil {
		fileObjectID = strings.TrimSpace(*body.FileObjectId)
	}
	rawText := ""
	if body.RawText != nil {
		rawText = strings.TrimSpace(*body.RawText)
	}
	if title == "" {
		return resume.RegisterInput{}, validationError("title is required")
	}
	if language == "" {
		return resume.RegisterInput{}, validationError("language is required")
	}
	switch sourceType {
	case "upload":
		if fileObjectID == "" || rawText != "" {
			return resume.RegisterInput{}, validationError("upload source requires fileObjectId only")
		}
	case "paste":
		if rawText == "" || fileObjectID != "" {
			return resume.RegisterInput{}, validationError("paste source requires rawText only")
		}
	default:
		return resume.RegisterInput{}, validationError("sourceType must be upload or paste")
	}
	return resume.RegisterInput{
		UserID:         strings.TrimSpace(userID),
		IdempotencyKey: strings.TrimSpace(idempotencyKey),
		SourceType:     sourceType,
		FileObjectID:   fileObjectID,
		RawText:        rawText,
		Title:          title,
		Language:       language,
	}, nil
}

type validationError string

func (e validationError) Error() string { return string(e) }

func cloneMap(in map[string]any) map[string]any {
	if in == nil {
		return nil
	}
	out := make(map[string]any, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
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

func writeAPIError(w http.ResponseWriter, status int, code string, message string, details map[string]any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	meta := sharederrors.CodeRegistry[code]
	raw, _ := json.Marshal(api.ApiErrorResponse{Error: api.ApiError{
		Code:      code,
		Message:   message,
		RequestID: "",
		Retryable: meta.Retryable,
		Details:   details,
	}})
	_, _ = w.Write(raw)
}
