package requestbody

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	api "github.com/monshunter/easyinterview/backend/internal/api/generated"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
)

// Limit rejects request bodies larger than maxBytes before routing to a
// handler. The accepted body is restored byte-for-byte for downstream decoders.
func Limit(maxBytes int64, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Body == nil || maxBytes <= 0 {
			next.ServeHTTP(w, r)
			return
		}
		body, err := io.ReadAll(io.LimitReader(r.Body, maxBytes+1))
		_ = r.Body.Close()
		if err != nil {
			writeError(w, http.StatusBadRequest, "request body could not be read")
			return
		}
		if int64(len(body)) > maxBytes {
			writeError(w, http.StatusRequestEntityTooLarge, "request body is too large")
			return
		}
		r.Body = io.NopCloser(bytes.NewReader(body))
		r.ContentLength = int64(len(body))
		next.ServeHTTP(w, r)
	})
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(api.ApiErrorResponse{Error: api.ApiError{Code: sharederrors.CodeValidationFailed, Message: message}})
}
