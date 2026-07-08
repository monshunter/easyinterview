package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	api "github.com/monshunter/easyinterview/backend/internal/api/generated"
	"github.com/monshunter/easyinterview/backend/internal/middleware/idempotency"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
	uploadservice "github.com/monshunter/easyinterview/backend/internal/upload/service"
)

type SessionResolver func(ctx context.Context) (userID string, ok bool)

type PresignService interface {
	CreateUploadPresign(ctx context.Context, in uploadservice.CreatePresignInput) (api.UploadPresign, error)
}

type Options struct {
	Service           PresignService
	Session           SessionResolver
	PresignTTL        time.Duration
	MaxBytesByPurpose map[string]int64
}

type Handler struct {
	service           PresignService
	session           SessionResolver
	presignTTL        time.Duration
	maxBytesByPurpose map[string]int64
}

func New(opts Options) *Handler {
	return &Handler{
		service:           opts.Service,
		session:           opts.Session,
		presignTTL:        opts.PresignTTL,
		maxBytesByPurpose: cloneMaxBytes(opts.MaxBytesByPurpose),
	}
}

func (h *Handler) CreateUploadPresign(w http.ResponseWriter, r *http.Request) {
	if h == nil || h.service == nil {
		writeAPIError(w, http.StatusInternalServerError, sharederrors.CodeValidationFailed, "upload service is not configured", nil)
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
	var body api.UploadPresignRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeAPIError(w, http.StatusBadRequest, sharederrors.CodeValidationFailed, "request body is malformed", nil)
		return
	}
	purpose := strings.TrimSpace(body.Purpose)
	maxBytes, ok := h.maxBytesByPurpose[purpose]
	if !ok {
		writeAPIError(w, http.StatusUnprocessableEntity, sharederrors.CodeValidationFailed, "purpose is not supported", map[string]any{"field": "purpose"})
		return
	}
	if strings.TrimSpace(body.FileName) == "" {
		writeAPIError(w, http.StatusUnprocessableEntity, sharederrors.CodeValidationFailed, "fileName is required", map[string]any{"field": "fileName"})
		return
	}
	if strings.TrimSpace(body.ContentType) == "" {
		writeAPIError(w, http.StatusUnprocessableEntity, sharederrors.CodeValidationFailed, "contentType is required", map[string]any{"field": "contentType"})
		return
	}
	if body.ByteSize <= 0 || body.ByteSize > maxBytes {
		writeAPIError(w, http.StatusUnprocessableEntity, sharederrors.CodeValidationFailed, "byteSize exceeds purpose limit", map[string]any{"field": "byteSize"})
		return
	}
	if purpose == "resume" && !isSupportedResumeUpload(body.FileName, body.ContentType) {
		writeAPIError(w, http.StatusUnprocessableEntity, sharederrors.CodeValidationFailed, "resume upload type is not supported", map[string]any{"field": "fileName"})
		return
	}
	out, err := h.service.CreateUploadPresign(r.Context(), uploadservice.CreatePresignInput{
		UserID:         userID,
		IdempotencyKey: idempotencyKey,
		Purpose:        purpose,
		FileName:       strings.TrimSpace(body.FileName),
		ContentType:    strings.TrimSpace(body.ContentType),
		ByteSize:       body.ByteSize,
		PresignTTL:     h.presignTTL,
		MaxBytes:       maxBytes,
	})
	if err != nil {
		if errors.Is(err, uploadservice.ErrValidationFailed) {
			writeAPIError(w, http.StatusUnprocessableEntity, sharederrors.CodeValidationFailed, "upload presign validation failed", nil)
			return
		}
		writeAPIError(w, http.StatusInternalServerError, sharederrors.CodeValidationFailed, "upload presign failed", nil)
		return
	}
	writeJSON(w, http.StatusCreated, out)
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

func isSupportedResumeUpload(fileName string, contentType string) bool {
	lowerName := strings.ToLower(strings.TrimSpace(fileName))
	mediaType := strings.ToLower(strings.TrimSpace(contentType))
	switch {
	case strings.HasSuffix(lowerName, ".pdf"):
		return mediaType == "application/pdf"
	case strings.HasSuffix(lowerName, ".md"), strings.HasSuffix(lowerName, ".markdown"):
		return mediaType == "text/markdown" || mediaType == "text/plain"
	case strings.HasSuffix(lowerName, ".txt"):
		return mediaType == "text/plain"
	default:
		return false
	}
}

func cloneMaxBytes(in map[string]int64) map[string]int64 {
	out := make(map[string]int64, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}
