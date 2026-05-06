package auth

import (
	"context"
	"strings"
	"sync"
)

const (
	MetricAuthChallengeStartedTotal = "auth_challenge_started_total"
	MetricAuthSessionMintedTotal    = "auth_session_minted_total"
	MetricAuthLogoutTotal           = "auth_logout_total"
	MetricAuthDeleteHandoffTotal    = "auth_delete_handoff_total"
	MetricAuthFailureTotal          = "auth_failure_total"

	AuthAuditActionChallengeStarted = "auth.challenge_started"
	AuthAuditActionSessionMinted    = "auth.session_minted"
	AuthAuditActionLogout           = "auth.logout"
	AuthAuditActionDeleteHandoff    = "auth.delete_handoff"
	AuthAuditActionFailure          = "auth.failure"
)

var (
	AuthServiceResultLabelKeys   = []string{"service", "result"}
	AuthOperationResultLabelKeys = []string{"service", "operation", "result"}

	f1AllowedAuthMetricLabels = map[string]struct{}{
		"service":   {},
		"operation": {},
		"result":    {},
	}
)

type AuthMetricRegisterer interface {
	Counter(name, help string, labelKeys []string) AuthMetricCounter
}

type AuthMetricCounter interface {
	Inc(labelValues ...string)
}

type AuthMetricsOptions struct {
	Service string
}

type AuthMetrics struct {
	service          string
	challengeStarted AuthMetricCounter
	sessionMinted    AuthMetricCounter
	logout           AuthMetricCounter
	deleteHandoff    AuthMetricCounter
	failure          AuthMetricCounter
}

type AuthAuditRecorder interface {
	RecordAuthAuditEvent(context.Context, AuthAuditEvent) error
}

type AuthAuditEvent struct {
	Action           string
	Operation        string
	Result           string
	UserIDHash       string
	ChallengeID      string
	PrivacyRequestID string
	JobID            string
	TraceID          string
}

func IsF1AllowedAuthMetricLabel(label string) bool {
	_, ok := f1AllowedAuthMetricLabels[label]
	return ok
}

func RegisterAuthMetrics(r AuthMetricRegisterer, opts AuthMetricsOptions) AuthMetrics {
	service := strings.TrimSpace(opts.Service)
	if service == "" {
		service = "backend"
	}
	metrics := AuthMetrics{service: service}
	if r == nil {
		return metrics
	}
	metrics.challengeStarted = r.Counter(MetricAuthChallengeStartedTotal, "Total passwordless auth challenge start requests.", AuthServiceResultLabelKeys)
	metrics.sessionMinted = r.Counter(MetricAuthSessionMintedTotal, "Total passwordless sessions minted after challenge verification.", AuthServiceResultLabelKeys)
	metrics.logout = r.Counter(MetricAuthLogoutTotal, "Total passwordless logout requests with an authenticated session.", AuthServiceResultLabelKeys)
	metrics.deleteHandoff = r.Counter(MetricAuthDeleteHandoffTotal, "Total authenticated privacy delete handoffs created by auth.", AuthServiceResultLabelKeys)
	metrics.failure = r.Counter(MetricAuthFailureTotal, "Total passwordless auth failures partitioned by operation.", AuthOperationResultLabelKeys)
	return metrics
}

func (m AuthMetrics) recordChallengeStarted(result string) {
	if m.challengeStarted != nil {
		m.challengeStarted.Inc(m.serviceName(), authMetricResult(result))
	}
}

func (m AuthMetrics) recordSessionMinted(result string) {
	if m.sessionMinted != nil {
		m.sessionMinted.Inc(m.serviceName(), authMetricResult(result))
	}
}

func (m AuthMetrics) recordLogout(result string) {
	if m.logout != nil {
		m.logout.Inc(m.serviceName(), authMetricResult(result))
	}
}

func (m AuthMetrics) recordDeleteHandoff(result string) {
	if m.deleteHandoff != nil {
		m.deleteHandoff.Inc(m.serviceName(), authMetricResult(result))
	}
}

func (m AuthMetrics) recordFailure(operation string, result string) {
	if m.failure != nil {
		m.failure.Inc(m.serviceName(), authMetricOperation(operation), authMetricResult(result))
	}
}

func (m AuthMetrics) serviceName() string {
	if strings.TrimSpace(m.service) == "" {
		return "backend"
	}
	return m.service
}

func authMetricOperation(operation string) string {
	operation = strings.TrimSpace(operation)
	if operation == "" {
		return "unknown"
	}
	return operation
}

func authMetricResult(result string) string {
	result = strings.TrimSpace(result)
	if result == "" {
		return "unknown"
	}
	return result
}

type authTraceIDContextKey struct{}

func ContextWithAuthTraceID(ctx context.Context, traceID string) context.Context {
	traceID = strings.TrimSpace(traceID)
	if traceID == "" {
		return ctx
	}
	return context.WithValue(ctx, authTraceIDContextKey{}, traceID)
}

func AuthTraceIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	traceID, _ := ctx.Value(authTraceIDContextKey{}).(string)
	return traceID
}

func TraceIDFromTraceparent(traceparent string) string {
	parts := strings.Split(strings.TrimSpace(traceparent), "-")
	if len(parts) < 4 || len(parts[1]) != 32 {
		return ""
	}
	return parts[1]
}

type InMemoryAuthMetricRegistry struct {
	mu       sync.Mutex
	counters map[string]*inMemoryAuthCounter
}

func NewInMemoryAuthMetricRegistry() *InMemoryAuthMetricRegistry {
	return &InMemoryAuthMetricRegistry{counters: map[string]*inMemoryAuthCounter{}}
}

func (r *InMemoryAuthMetricRegistry) Counter(name, _ string, labelKeys []string) AuthMetricCounter {
	r.mu.Lock()
	defer r.mu.Unlock()
	if c, ok := r.counters[name]; ok {
		return c
	}
	c := &inMemoryAuthCounter{
		labelKeys: append([]string{}, labelKeys...),
		values:    map[string]float64{},
	}
	r.counters[name] = c
	return c
}

func (r *InMemoryAuthMetricRegistry) CounterRegistered(name string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	_, ok := r.counters[name]
	return ok
}

func (r *InMemoryAuthMetricRegistry) CounterLabelKeys(name string) []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	counter, ok := r.counters[name]
	if !ok {
		return nil
	}
	return append([]string{}, counter.labelKeys...)
}

func (r *InMemoryAuthMetricRegistry) CounterValue(name string, labelValues ...string) float64 {
	r.mu.Lock()
	counter, ok := r.counters[name]
	r.mu.Unlock()
	if !ok {
		return 0
	}
	counter.mu.Lock()
	defer counter.mu.Unlock()
	return counter.values[joinAuthMetricLabels(labelValues)]
}

func (r *InMemoryAuthMetricRegistry) CounterLabelValues(name string) [][]string {
	r.mu.Lock()
	counter, ok := r.counters[name]
	r.mu.Unlock()
	if !ok {
		return nil
	}
	counter.mu.Lock()
	defer counter.mu.Unlock()
	out := make([][]string, 0, len(counter.values))
	for key := range counter.values {
		out = append(out, splitAuthMetricLabels(key))
	}
	return out
}

type inMemoryAuthCounter struct {
	mu        sync.Mutex
	labelKeys []string
	values    map[string]float64
}

func (c *inMemoryAuthCounter) Inc(labelValues ...string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.values[joinAuthMetricLabels(labelValues)]++
}

func joinAuthMetricLabels(labelValues []string) string {
	if len(labelValues) == 0 {
		return ""
	}
	return strings.Join(labelValues, "\x1f")
}

func splitAuthMetricLabels(key string) []string {
	if key == "" {
		return nil
	}
	return strings.Split(key, "\x1f")
}
