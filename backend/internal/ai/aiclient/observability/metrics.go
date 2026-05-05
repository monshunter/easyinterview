package observability

import "sync"

// Registerer is the small surface this package needs from a metrics
// backend. F1 observability-stack will satisfy it with a real
// prometheus.Registerer; tests use the in-memory NewInMemoryRegistry. The
// abstraction keeps backend/go.mod free of a Prometheus dependency until
// F1 lands its plan.
type Registerer interface {
	Counter(name, help string, labelKeys []string) Counter
	Histogram(name, help string, labelKeys []string, buckets []float64) Histogram
}

// Counter is the subset of prometheus.CounterVec used by the decorator.
type Counter interface {
	Inc(labelValues ...string)
	Add(value float64, labelValues ...string)
}

// Histogram is the subset of prometheus.HistogramVec used by the decorator.
type Histogram interface {
	Observe(value float64, labelValues ...string)
}

// Metric family names are the seven canonical names spec §2.1 / D-6 lock.
const (
	MetricRunsTotal                = "ai_task_runs_total"
	MetricLatencySeconds           = "ai_task_latency_seconds"
	MetricInputTokensTotal         = "ai_task_input_tokens_total"
	MetricOutputTokensTotal        = "ai_task_output_tokens_total"
	MetricCostUSDTotal             = "ai_task_cost_usd_total"
	MetricOutputValidationFailures = "ai_output_validation_failures_total"
	MetricFallbackTotal            = "ai_fallback_total"
)

// LabelKeys are the standard label sets for each metric family. F1 may
// extend the set in their own plan, but every required key listed here
// must remain present (spec §2.1).
var (
	StandardLabelKeys = []string{
		"provider", "model_family", "model_profile_name", "route", "capability", "language", "result",
	}
	FallbackLabelKeys = []string{
		"provider", "model_family", "model_profile_name", "route", "capability", "language", "result", "from_model_family", "to_model_family",
	}
)

// metricSet bundles the seven metric handles after registration.
type metricSet struct {
	runs              Counter
	latency           Histogram
	inputTokens       Counter
	outputTokens      Counter
	cost              Counter
	validationFailure Counter
	fallback          Counter
}

func registerMetrics(r Registerer) metricSet {
	return metricSet{
		runs:              r.Counter(MetricRunsTotal, "Total AIClient calls partitioned by result.", StandardLabelKeys),
		latency:           r.Histogram(MetricLatencySeconds, "End-to-end AIClient call latency in seconds.", StandardLabelKeys, defaultLatencyBuckets()),
		inputTokens:       r.Counter(MetricInputTokensTotal, "Total input tokens consumed.", StandardLabelKeys),
		outputTokens:      r.Counter(MetricOutputTokensTotal, "Total output tokens produced.", StandardLabelKeys),
		cost:              r.Counter(MetricCostUSDTotal, "Total AIClient call cost in USD micros.", StandardLabelKeys),
		validationFailure: r.Counter(MetricOutputValidationFailures, "AIClient validateOutput failures (AI_OUTPUT_INVALID).", StandardLabelKeys),
		fallback:          r.Counter(MetricFallbackTotal, "AIClient fallback chain transitions.", FallbackLabelKeys),
	}
}

func defaultLatencyBuckets() []float64 {
	return []float64{0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10, 30}
}

// InMemoryRegistry is a test-only Registerer that stores counter and
// histogram values keyed by label tuples so privacy and correctness tests
// can assert metric output without standing up a Prometheus runtime.
type InMemoryRegistry struct {
	mu         sync.Mutex
	counters   map[string]*inMemCounter
	histograms map[string]*inMemHistogram
}

// NewInMemoryRegistry constructs an empty registry.
func NewInMemoryRegistry() *InMemoryRegistry {
	return &InMemoryRegistry{
		counters:   map[string]*inMemCounter{},
		histograms: map[string]*inMemHistogram{},
	}
}

// Counter implements Registerer.
func (r *InMemoryRegistry) Counter(name, _ string, labelKeys []string) Counter {
	r.mu.Lock()
	defer r.mu.Unlock()
	if c, ok := r.counters[name]; ok {
		return c
	}
	c := &inMemCounter{name: name, labelKeys: append([]string{}, labelKeys...), values: map[string]float64{}}
	r.counters[name] = c
	return c
}

// Histogram implements Registerer.
func (r *InMemoryRegistry) Histogram(name, _ string, labelKeys []string, _ []float64) Histogram {
	r.mu.Lock()
	defer r.mu.Unlock()
	if h, ok := r.histograms[name]; ok {
		return h
	}
	h := &inMemHistogram{name: name, labelKeys: append([]string{}, labelKeys...), observations: map[string][]float64{}}
	r.histograms[name] = h
	return h
}

// CounterValue returns the accumulated value for one (name, labels) tuple.
// Missing tuples return 0 to mirror prometheus semantics.
func (r *InMemoryRegistry) CounterValue(name string, labelValues ...string) float64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	c, ok := r.counters[name]
	if !ok {
		return 0
	}
	return c.values[joinLabels(labelValues)]
}

// HistogramObservations returns the recorded observations for one
// (name, labels) tuple.
func (r *InMemoryRegistry) HistogramObservations(name string, labelValues ...string) []float64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	h, ok := r.histograms[name]
	if !ok {
		return nil
	}
	out := append([]float64{}, h.observations[joinLabels(labelValues)]...)
	return out
}

// CounterLabelValues returns every label tuple that has been recorded for
// a counter. Tests use it to assert that no PII slipped into label values.
func (r *InMemoryRegistry) CounterLabelValues(name string) [][]string {
	r.mu.Lock()
	defer r.mu.Unlock()
	c, ok := r.counters[name]
	if !ok {
		return nil
	}
	out := make([][]string, 0, len(c.values))
	for k := range c.values {
		out = append(out, splitLabels(k))
	}
	return out
}

// CounterRegistered returns true if the named counter has been registered
// (regardless of whether it has any observations).
func (r *InMemoryRegistry) CounterRegistered(name string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	_, ok := r.counters[name]
	return ok
}

// HistogramRegistered returns true if the named histogram has been
// registered.
func (r *InMemoryRegistry) HistogramRegistered(name string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	_, ok := r.histograms[name]
	return ok
}

type inMemCounter struct {
	mu        sync.Mutex
	name      string
	labelKeys []string
	values    map[string]float64
}

func (c *inMemCounter) Inc(labelValues ...string) { c.Add(1, labelValues...) }
func (c *inMemCounter) Add(value float64, labelValues ...string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.values[joinLabels(labelValues)] += value
}

type inMemHistogram struct {
	mu           sync.Mutex
	name         string
	labelKeys    []string
	observations map[string][]float64
}

func (h *inMemHistogram) Observe(value float64, labelValues ...string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	key := joinLabels(labelValues)
	h.observations[key] = append(h.observations[key], value)
}

func joinLabels(labelValues []string) string {
	if len(labelValues) == 0 {
		return ""
	}
	out := ""
	for i, v := range labelValues {
		if i > 0 {
			out += "\x1f"
		}
		out += v
	}
	return out
}

func splitLabels(key string) []string {
	if key == "" {
		return nil
	}
	out := []string{}
	cur := ""
	for _, r := range key {
		if r == '\x1f' {
			out = append(out, cur)
			cur = ""
			continue
		}
		cur += string(r)
	}
	out = append(out, cur)
	return out
}
