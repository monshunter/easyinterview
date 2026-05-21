// Package handler hosts the 12 JD-Match HTTP handlers. Each endpoint
// lives in its own file in this package; handler.go only defines the
// shared Handler struct, Options, and session/idempotency helpers.
package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	api "github.com/monshunter/easyinterview/backend/internal/api/generated"
	"github.com/monshunter/easyinterview/backend/internal/jdmatch"
	"github.com/monshunter/easyinterview/backend/internal/jdmatch/service"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
)

// SessionResolver returns the authenticated user id resolved from the
// request context (typically populated by the session middleware in
// cmd/api).
type SessionResolver func(ctx context.Context) (userID string, ok bool)

// AgentScanReader is the read-side projection the GetAgentScanStatus
// handler depends on; the full store.Repository satisfies it.
type AgentScanReader interface {
	GetLatestAgentScanForUser(ctx context.Context, userID string) (jdmatch.AgentScanRecord, error)
}

// ProfileBuilder is the orchestrator dependency for GetJobMatchProfile.
// In production it wraps service.BuildJobMatchProfile bound to a
// ProfileDeps instance; tests inject a deterministic stub.
type ProfileBuilder func(ctx context.Context, userID string) (service.JobMatchProfileResult, error)

// Options bundles all Handler dependencies. Session, AgentScans, and
// ProfileBuilder are mandatory.
type Options struct {
	Session        SessionResolver
	AgentScans     AgentScanReader
	ProfileBuilder ProfileBuilder
}

// Handler implements every JD-Match HTTP endpoint method on the
// generated server interface.
type Handler struct {
	session        SessionResolver
	agentScans     AgentScanReader
	profileBuilder ProfileBuilder
}

// New constructs a Handler with the supplied options.
func New(opts Options) *Handler {
	return &Handler{
		session:        opts.Session,
		agentScans:     opts.AgentScans,
		profileBuilder: opts.ProfileBuilder,
	}
}

func (h *Handler) resolveUser(r *http.Request) (string, bool) {
	if h == nil || h.session == nil {
		return "", false
	}
	uid, ok := h.session(r.Context())
	uid = strings.TrimSpace(uid)
	return uid, ok && uid != ""
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	raw, err := json.Marshal(body)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, sharederrors.CodeValidationFailed, "jdmatch response encoding failed", nil)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(raw)
}

func writeAPIError(w http.ResponseWriter, status int, code string, message string, details map[string]any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	payload := api.ApiError{
		Code:      code,
		Message:   message,
		Retryable: status >= 500,
	}
	if details != nil {
		payload.Details = details
	}
	_ = json.NewEncoder(w).Encode(payload)
}

func writeServiceError(w http.ResponseWriter, err error, fallback string) {
	switch {
	case errors.Is(err, jdmatch.ErrNotFound):
		writeAPIError(w, http.StatusNotFound, sharederrors.CodeResourceNotFound, "resource not found", nil)
	case errors.Is(err, jdmatch.ErrUserIDRequired):
		writeAPIError(w, http.StatusBadRequest, sharederrors.CodeValidationFailed, "userID required", nil)
	case errors.Is(err, jdmatch.ErrInvalidStatus):
		writeAPIError(w, http.StatusBadRequest, sharederrors.CodeValidationFailed, "invalid status", nil)
	default:
		writeAPIError(w, http.StatusInternalServerError, sharederrors.CodeValidationFailed, fallback, nil)
	}
}
