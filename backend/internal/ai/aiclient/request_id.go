package aiclient

import "context"

type requestIDKey struct{}

// WithRequestID stores a request id on the context. The
// openai_compatible adapter passes the value verbatim in the X-Request-ID
// header so an upstream provider endpoint can correlate logs end-to-end. Callers that
// already use middleware-stamped request ids should call this helper at the
// boundary where the request id is set.
func WithRequestID(ctx context.Context, id string) context.Context {
	if id == "" {
		return ctx
	}
	return context.WithValue(ctx, requestIDKey{}, id)
}

// RequestIDFromContext returns the request id stored on the context, or the
// empty string when none is set.
func RequestIDFromContext(ctx context.Context) string {
	v, _ := ctx.Value(requestIDKey{}).(string)
	return v
}
