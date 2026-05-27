package auth

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/api/generated"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
)

type HandlerOptions struct {
	Passwordless *PasswordlessService
	CookiePolicy *CookiePolicy
}

type Handler struct {
	passwordless *PasswordlessService
	cookiePolicy CookiePolicy
}

func NewHandler(opts HandlerOptions) *Handler {
	policy := CookiePolicyForAppEnv("prod")
	if opts.CookiePolicy != nil {
		policy = *opts.CookiePolicy
	}
	return &Handler{passwordless: opts.Passwordless, cookiePolicy: policy}
}

type CookiePolicy struct {
	Secure bool
}

func CookiePolicyForAppEnv(appEnv string) CookiePolicy {
	switch strings.ToLower(strings.TrimSpace(appEnv)) {
	case "dev", "test":
		return CookiePolicy{Secure: false}
	default:
		return CookiePolicy{Secure: true}
	}
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
	purpose := ChallengePurposeLogin
	if body.Purpose != nil {
		purpose = ChallengePurpose(*body.Purpose)
	}
	displayName := ""
	if body.DisplayName != nil {
		displayName = *body.DisplayName
	}
	ctx := ContextWithAuthTraceID(r.Context(), TraceIDFromTraceparent(r.Header.Get("traceparent")))
	_, err := h.passwordless.StartEmailChallenge(ctx, StartEmailChallengeInput{
		Email:        body.Email,
		Purpose:      purpose,
		DisplayName:  displayName,
		ReturnTo:     returnTo,
		RemoteAddr:   r.RemoteAddr,
		UserAgent:    r.UserAgent(),
		AcceptLocale: r.Header.Get("Accept-Language"),
	})
	if err != nil {
		if errors.Is(err, ErrEmailRegistered) {
			writeAPIError(w, http.StatusConflict, sharederrors.CodeValidationFailed, "email is already registered; sign in instead", false)
			return
		}
		writeAPIError(w, http.StatusBadRequest, sharederrors.CodeValidationFailed, "challenge could not be accepted", false)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func (h *Handler) VerifyAuthEmailChallenge(w http.ResponseWriter, r *http.Request) {
	if h == nil || h.passwordless == nil {
		writeAPIError(w, http.StatusInternalServerError, sharederrors.CodeValidationFailed, "passwordless service is not configured", false)
		return
	}
	ctx := ContextWithAuthTraceID(r.Context(), TraceIDFromTraceparent(r.Header.Get("traceparent")))
	result, err := h.passwordless.VerifyEmailChallenge(ctx, VerifyEmailChallengeInput{
		Token:      r.URL.Query().Get("token"),
		RemoteAddr: r.RemoteAddr,
		UserAgent:  r.UserAgent(),
	})
	if err != nil {
		status := http.StatusBadRequest
		code := sharederrors.CodeValidationFailed
		message := "challenge could not be verified"
		if errors.Is(err, ErrChallengeInvalid) || errors.Is(err, ErrChallengeExpired) || errors.Is(err, ErrChallengeConsumed) {
			status = http.StatusUnauthorized
			code = sharederrors.CodeAuthUnauthorized
			message = "challenge code is invalid or expired"
		}
		if errors.Is(err, ErrEmailRegistered) {
			status = http.StatusConflict
			code = sharederrors.CodeValidationFailed
			message = "email is already registered; sign in instead"
		}
		if errors.Is(err, ErrUserNotFound) {
			status = http.StatusUnauthorized
			code = sharederrors.CodeAuthUnauthorized
			message = "email is not registered; create an account first"
		}
		writeAPIError(w, status, code, message, false)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    result.SessionToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   h.cookiePolicy.Secure,
		SameSite: http.SameSiteLaxMode,
		Expires:  result.SessionExpiresAt,
		MaxAge:   int(time.Until(result.SessionExpiresAt).Seconds()),
	})
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(generated.Session{
		UserId:           result.UserID,
		SessionExpiresAt: result.SessionExpiresAt.Format(time.RFC3339),
	})
}

func (h *Handler) GetMe(w http.ResponseWriter, r *http.Request) {
	current, ok := CurrentSessionFromContext(r.Context())
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, sharederrors.CodeAuthUnauthorized, "authentication required or invalid", false)
		return
	}
	if h == nil || h.passwordless == nil {
		writeAPIError(w, http.StatusInternalServerError, sharederrors.CodeValidationFailed, "passwordless service is not configured", false)
		return
	}
	user, err := h.passwordless.CurrentUser(r.Context(), current.UserID)
	if err != nil {
		writeAPIError(w, http.StatusUnauthorized, sharederrors.CodeAuthUnauthorized, "authentication required or invalid", false)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(generated.UserContext{
		Id:                        user.ID,
		EmailMasked:               maskEmail(user.Email),
		DisplayName:               user.DisplayName,
		UiLanguage:                user.UILanguage,
		PreferredPracticeLanguage: user.PreferredPracticeLanguage,
	})
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	if current, ok := CurrentSessionFromContext(r.Context()); ok {
		if h == nil || h.passwordless == nil {
			h.clearSessionCookie(w)
			writeAPIError(w, http.StatusInternalServerError, sharederrors.CodeValidationFailed, "passwordless service is not configured", false)
			return
		}
		ctx := ContextWithAuthTraceID(r.Context(), TraceIDFromTraceparent(r.Header.Get("traceparent")))
		if err := h.passwordless.Logout(ctx, current); err != nil {
			h.clearSessionCookie(w)
			writeAPIError(w, http.StatusInternalServerError, sharederrors.CodeValidationFailed, "logout could not revoke session", false)
			return
		}
	}
	h.clearSessionCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) DeleteMe(w http.ResponseWriter, r *http.Request) {
	current, ok := CurrentSessionFromContext(r.Context())
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, sharederrors.CodeAuthUnauthorized, "authentication required or invalid", false)
		return
	}
	if h == nil || h.passwordless == nil {
		writeAPIError(w, http.StatusInternalServerError, sharederrors.CodeValidationFailed, "passwordless service is not configured", false)
		return
	}
	ctx := ContextWithAuthTraceID(r.Context(), TraceIDFromTraceparent(r.Header.Get("traceparent")))
	handoff, err := h.passwordless.DeleteMe(ctx, current, r.Header.Get("Idempotency-Key"))
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, sharederrors.CodeValidationFailed, "privacy delete handoff could not be created", false)
		return
	}
	h.clearSessionCookie(w)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(generated.PrivacyRequestWithJob{
		PrivacyRequestId: handoff.PrivacyRequestID,
		Job: generated.Job{
			Id:           handoff.JobID,
			JobType:      generated.JobTypePrivacyDelete,
			ResourceType: generated.ResourceTypePrivacyRequest,
			ResourceId:   handoff.PrivacyRequestID,
			Status:       generated.JobStatus("queued"),
			CreatedAt:    handoff.CreatedAt.Format(time.RFC3339),
			UpdatedAt:    handoff.UpdatedAt.Format(time.RFC3339),
		},
	})
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

func (h *Handler) clearSessionCookie(w http.ResponseWriter) {
	policy := CookiePolicyForAppEnv("prod")
	if h != nil {
		policy = h.cookiePolicy
	}
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   policy.Secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}

func maskEmail(email string) string {
	at := -1
	for i, ch := range email {
		if ch == '@' {
			at = i
			break
		}
	}
	if at <= 0 {
		return "***"
	}
	local := email[:at]
	domain := email[at:]
	if len(local) == 1 {
		return local[:1] + "***" + domain
	}
	return local[:1] + "***" + local[len(local)-1:] + domain
}
