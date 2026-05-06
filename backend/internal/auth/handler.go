package auth

import (
	"encoding/json"
	"net/http"

	"github.com/monshunter/easyinterview/backend/internal/api/generated"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
)

type HandlerOptions struct {
	Passwordless *PasswordlessService
}

type Handler struct {
	passwordless *PasswordlessService
}

func NewHandler(opts HandlerOptions) *Handler {
	return &Handler{passwordless: opts.Passwordless}
}

func (h *Handler) StartAuthEmailChallenge(w http.ResponseWriter, r *http.Request) {
	var body generated.AuthEmailStartRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&body); err != nil {
		writeAPIError(w, http.StatusBadRequest, sharederrors.CodeValidationFailed, "invalid JSON request body", false)
		return
	}
	if h == nil || h.passwordless == nil {
		writeAPIError(w, http.StatusInternalServerError, sharederrors.CodeValidationFailed, "passwordless service is not configured", false)
		return
	}
	returnTo := ""
	if body.ReturnTo != nil {
		returnTo = *body.ReturnTo
	}
	_, err := h.passwordless.StartEmailChallenge(r.Context(), StartEmailChallengeInput{
		Email:        body.Email,
		ReturnTo:     returnTo,
		RemoteAddr:   r.RemoteAddr,
		UserAgent:    r.UserAgent(),
		AcceptLocale: r.Header.Get("Accept-Language"),
	})
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, sharederrors.CodeValidationFailed, "challenge could not be accepted", false)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func writeAPIError(w http.ResponseWriter, status int, code string, message string, retryable bool) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"error": sharederrors.APIError{
			Code:      code,
			Message:   message,
			Retryable: retryable,
		},
	})
}
