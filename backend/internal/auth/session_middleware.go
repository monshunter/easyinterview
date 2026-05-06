package auth

import (
	"context"
	"errors"
	"net/http"

	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
)

type currentSessionContextKey struct{}

func ContextWithCurrentSession(ctx context.Context, current CurrentSession) context.Context {
	return context.WithValue(ctx, currentSessionContextKey{}, current)
}

func CurrentSessionFromContext(ctx context.Context) (CurrentSession, bool) {
	current, ok := ctx.Value(currentSessionContextKey{}).(CurrentSession)
	return current, ok
}

func SessionMiddleware(service *PasswordlessService, operationID string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requirement, ok := SessionPolicyForOperation(operationID)
		if !ok {
			writeAPIError(w, http.StatusInternalServerError, sharederrors.CodeValidationFailed, "operation session policy is not configured", false)
			return
		}
		if requirement == SessionPublic {
			next.ServeHTTP(w, r)
			return
		}
		cookie, err := r.Cookie(SessionCookieName)
		if err != nil || cookie.Value == "" {
			if requirement == SessionOptional {
				next.ServeHTTP(w, r)
				return
			}
			writeAPIError(w, http.StatusUnauthorized, sharederrors.CodeAuthUnauthorized, "authentication required or invalid", false)
			return
		}
		current, err := service.ResolveSession(r.Context(), cookie.Value)
		if err != nil {
			if requirement == SessionOptional {
				next.ServeHTTP(w, r)
				return
			}
			if errors.Is(err, ErrSessionInvalid) || errors.Is(err, ErrSessionExpired) || errors.Is(err, ErrSessionRevoked) {
				writeAPIError(w, http.StatusUnauthorized, sharederrors.CodeAuthUnauthorized, "authentication required or invalid", false)
				return
			}
			writeAPIError(w, http.StatusInternalServerError, sharederrors.CodeValidationFailed, "session could not be resolved", false)
			return
		}
		next.ServeHTTP(w, r.WithContext(ContextWithCurrentSession(r.Context(), current)))
	})
}
