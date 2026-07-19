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
	EmailCode    *EmailCodeService
	CookiePolicy *CookiePolicy
}

type Handler struct {
	emailCode    *EmailCodeService
	cookiePolicy CookiePolicy
}

func NewHandler(opts HandlerOptions) *Handler {
	policy := CookiePolicyForAppEnv("prod")
	if opts.CookiePolicy != nil {
		policy = *opts.CookiePolicy
	}
	return &Handler{emailCode: opts.EmailCode, cookiePolicy: policy}
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
	if h == nil || h.emailCode == nil {
		writeAPIError(w, http.StatusInternalServerError, sharederrors.CodeValidationFailed, "email-code service is not configured", false)
		return
	}
	returnTo := ""
	if body.ReturnTo != nil {
		returnTo = *body.ReturnTo
	}
	ctx := ContextWithAuthTraceID(r.Context(), TraceIDFromTraceparent(r.Header.Get("traceparent")))
	_, err := h.emailCode.StartEmailChallenge(ctx, StartEmailChallengeInput{
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

func (h *Handler) VerifyAuthEmailChallenge(w http.ResponseWriter, r *http.Request) {
	if h == nil || h.emailCode == nil {
		writeAPIError(w, http.StatusInternalServerError, sharederrors.CodeValidationFailed, "email-code service is not configured", false)
		return
	}
	ctx := ContextWithAuthTraceID(r.Context(), TraceIDFromTraceparent(r.Header.Get("traceparent")))
	result, err := h.emailCode.VerifyEmailChallenge(ctx, VerifyEmailChallengeInput{
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
	if h == nil || h.emailCode == nil {
		writeAPIError(w, http.StatusInternalServerError, sharederrors.CodeValidationFailed, "email-code service is not configured", false)
		return
	}
	user, err := h.emailCode.CurrentUser(r.Context(), current.UserID)
	if err != nil {
		writeAPIError(w, http.StatusUnauthorized, sharederrors.CodeAuthUnauthorized, "authentication required or invalid", false)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(generatedUserContext(user))
}

func (h *Handler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	current, ok := CurrentSessionFromContext(r.Context())
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, sharederrors.CodeAuthUnauthorized, "authentication required or invalid", false)
		return
	}
	if h == nil || h.emailCode == nil {
		writeAPIError(w, http.StatusInternalServerError, sharederrors.CodeValidationFailed, "email-code service is not configured", false)
		return
	}
	var body generated.UpdateMeRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&body); err != nil {
		writeAPIError(w, http.StatusBadRequest, sharederrors.CodeValidationFailed, "invalid JSON request body", false)
		return
	}
	var preferences *AccountDisplayPreferences
	if body.DisplayPreferences != nil {
		preferences = &AccountDisplayPreferences{Theme: AccountTheme(body.DisplayPreferences.Theme)}
		if body.DisplayPreferences.CustomAccent != nil {
			preferences.CustomAccent = &CustomAccent{H: body.DisplayPreferences.CustomAccent.H, C: body.DisplayPreferences.CustomAccent.C}
		}
	}
	user, err := h.emailCode.UpdateMe(r.Context(), current.UserID, UpdateUserContextInput{
		DisplayName:        body.DisplayName,
		AcceptedTerms:      body.AcceptedTerms,
		DisplayPreferences: preferences,
	})
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, sharederrors.CodeValidationFailed, "account could not be updated", false)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(generatedUserContext(user))
}

func generatedUserContext(user UserContext) generated.UserContext {
	preferences := generated.AccountDisplayPreferences{
		Theme: generated.AccountTheme(user.DisplayPreferences.Theme),
	}
	if preferences.Theme == "" {
		preferences.Theme = generated.AccountThemeOcean
	}
	if user.DisplayPreferences.CustomAccent != nil {
		preferences.CustomAccent = &generated.CustomAccent{
			H: user.DisplayPreferences.CustomAccent.H,
			C: user.DisplayPreferences.CustomAccent.C,
		}
	}
	return generated.UserContext{
		Id:                        user.ID,
		Email:                     user.Email,
		DisplayName:               user.DisplayName,
		ProfileCompletionRequired: user.ProfileCompletionRequired,
		DisplayPreferences:        preferences,
	}
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	if current, ok := CurrentSessionFromContext(r.Context()); ok {
		if h == nil || h.emailCode == nil {
			h.clearSessionCookie(w)
			writeAPIError(w, http.StatusInternalServerError, sharederrors.CodeValidationFailed, "email-code service is not configured", false)
			return
		}
		ctx := ContextWithAuthTraceID(r.Context(), TraceIDFromTraceparent(r.Header.Get("traceparent")))
		if err := h.emailCode.Logout(ctx, current); err != nil {
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
	if h == nil || h.emailCode == nil {
		writeAPIError(w, http.StatusInternalServerError, sharederrors.CodeValidationFailed, "email-code service is not configured", false)
		return
	}
	ctx := ContextWithAuthTraceID(r.Context(), TraceIDFromTraceparent(r.Header.Get("traceparent")))
	handoff, err := h.emailCode.DeleteMe(ctx, current, r.Header.Get("Idempotency-Key"))
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
