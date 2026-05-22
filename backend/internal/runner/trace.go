package runner

import (
	"context"
	"encoding/json"
	"regexp"
)

// traceparentField / traceIDField are the JSON keys the kernel inspects on an
// async_jobs payload to recover a W3C trace context (spec D-11). Either a full
// `traceparent` header value or a bare 32-hex `traceId` is accepted.
const (
	traceparentField = "traceparent"
	traceIDField     = "traceId"
)

// w3cTraceparentPattern matches `version-trace_id-parent_id-flags`; group 1 is
// the 32-hex trace-id.
var w3cTraceparentPattern = regexp.MustCompile(`^[0-9a-f]{2}-([0-9a-f]{32})-[0-9a-f]{16}-[0-9a-f]{2}$`)

var hex32Pattern = regexp.MustCompile(`^[0-9a-f]{32}$`)

type traceContextKey struct{}

// traceIDFromPayload extracts a trace-id from an async_jobs payload. It returns
// "" when the payload is empty, not an object, or carries no recognizable
// trace field.
func traceIDFromPayload(payload []byte) string {
	if len(payload) == 0 {
		return ""
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(payload, &fields); err != nil {
		return ""
	}
	if raw, ok := fields[traceparentField]; ok {
		var tp string
		if json.Unmarshal(raw, &tp) == nil {
			if m := w3cTraceparentPattern.FindStringSubmatch(tp); m != nil {
				return m[1]
			}
		}
	}
	if raw, ok := fields[traceIDField]; ok {
		var id string
		if json.Unmarshal(raw, &id) == nil && hex32Pattern.MatchString(id) {
			return id
		}
	}
	return ""
}

// withTraceID stores a recovered trace-id on the context so downstream handlers
// and stores can read it for log correlation.
func withTraceID(ctx context.Context, traceID string) context.Context {
	if traceID == "" {
		return ctx
	}
	return context.WithValue(ctx, traceContextKey{}, traceID)
}

// TraceIDFromContext returns the trace-id recovered by the runtime, or "".
func TraceIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if v, ok := ctx.Value(traceContextKey{}).(string); ok {
		return v
	}
	return ""
}
