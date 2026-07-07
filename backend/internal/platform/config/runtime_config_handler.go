package config

import (
	"encoding/json"
	"net/http"

	"github.com/monshunter/easyinterview/backend/internal/platform/featureflag"
)

// SessionAnalyticsResolver returns the analytics opt-in state for the
// caller. Implementations are owned by C1 backend-auth; the minimal stub
// in this package treats every request as anonymous opt-out.
type SessionAnalyticsResolver func(r *http.Request) bool

// RuntimeConfigHandlerOptions configures NewRuntimeConfigHandler.
type RuntimeConfigHandlerOptions struct {
	Loader          *Loader
	Flags           featureflag.FeatureFlagClient
	FlagContextFunc func(r *http.Request) featureflag.FlagContext
	SessionResolver SessionAnalyticsResolver
}

// NewRuntimeConfigHandler returns the minimal HTTP handler that powers
// `GET /api/v1/runtime-config`. The OpenAPI contract truth source lives
// in B2 openapi-v1-contract; this handler only wires the field allowlist
// declared in BuildRuntimeConfig and serializes the result as JSON.
//
// backend-auth may provide a session-aware SessionResolver; frontend-shell
// consumes the response through the runtime-config fetcher. Neither owner may
// mutate the field allowlist without a spec revision (D-2).
func NewRuntimeConfigHandler(opts RuntimeConfigHandlerOptions) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		input := RuntimeConfigInput{
			Loader: opts.Loader,
			Flags:  opts.Flags,
		}
		if opts.FlagContextFunc != nil {
			input.FlagContext = opts.FlagContextFunc(r)
		}
		if opts.SessionResolver != nil {
			input.AnalyticsOptIn = opts.SessionResolver(r)
		}
		rc := BuildRuntimeConfig(ctx, input)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(rc)
	})
}
