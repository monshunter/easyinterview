package runner

import (
	"sync"
	"time"
)

// Metric family names (spec §4.4 / ADR-Q2 §3.6 / B3 D-9). These are the
// canonical names the kernel emits; an F1 Prometheus adapter must preserve
// them.
const (
	MetricAsyncJobDuration      = "async_job_duration_seconds"
	MetricAsyncJobsProcessed    = "async_jobs_processed_total"
	MetricAsyncJobQueueDepth    = "async_job_queue_depth"
	MetricAsyncJobLag           = "async_job_lag_seconds"
	MetricOutboxPending         = "outbox_events_pending"
	MetricOutboxPublishDuration = "outbox_publish_duration_seconds"
	MetricOutboxPublishFailures = "outbox_publish_failures_total"
)

// Metrics is the observability sink for the kernel (spec §4.4). When no sink is
// supplied, emissions are discarded; KernelMetrics is the registry-backed
// implementation. Keeping it an interface lets the runtime and outbox
// dispatcher emit without depending on a concrete registry.
type Metrics interface {
	// ObserveJobProcessed records one finalized job with its terminal result
	// (succeeded / failed / dead / retried) and wall-clock duration.
	ObserveJobProcessed(jobType, result string, duration time.Duration)
	// ObserveReaped records rows reclaimed by the reaper for a job_type.
	ObserveReaped(jobType string, count int64)
	// ObserveOutboxPublish records one outbox publish attempt with its result
	// (published / retried / dead) and duration.
	ObserveOutboxPublish(result string, duration time.Duration)
	// ObserveOutboxFailure increments the outbox publish failure counter.
	ObserveOutboxFailure()
	// SetOutboxPending sets the current pending outbox backlog gauge.
	SetOutboxPending(pending float64)
}

// discardMetrics is the default sink used when no Metrics is supplied.
type discardMetrics struct{}

func (discardMetrics) ObserveJobProcessed(string, string, time.Duration) {}
func (discardMetrics) ObserveReaped(string, int64)                       {}
func (discardMetrics) ObserveOutboxPublish(string, time.Duration)        {}
func (discardMetrics) ObserveOutboxFailure()                             {}
func (discardMetrics) SetOutboxPending(float64)                          {}

// Counter / Gauge / Histogram are the minimal metric handle surfaces the kernel
// needs. They mirror the Prometheus *Vec subset so an F1 adapter is a thin
// wrapper.
type Counter interface {
	Inc(labelValues ...string)
	Add(value float64, labelValues ...string)
}

type Gauge interface {
	Set(value float64, labelValues ...string)
}

type Histogram interface {
	Observe(value float64, labelValues ...string)
}

// MetricsRegisterer is the metric backend the kernel registers against. The
// kernel never imports Prometheus directly (backend-runtime-topology D-4);
// F1 supplies a real registerer, tests use NewInMemoryMetricsRegistry.
type MetricsRegisterer interface {
	Counter(name, help string, labelKeys []string) Counter
	Gauge(name, help string, labelKeys []string) Gauge
	Histogram(name, help string, labelKeys []string, buckets []float64) Histogram
}

// KernelMetrics registers and emits the spec §4.4 metric families.
type KernelMetrics struct {
	jobDuration    Histogram
	jobsProcessed  Counter
	queueDepth     Gauge
	jobLag         Gauge
	outboxPending  Gauge
	outboxDuration Histogram
	outboxFailures Counter
}

// NewKernelMetrics registers the seven metric families on r.
func NewKernelMetrics(r MetricsRegisterer) *KernelMetrics {
	durationBuckets := []float64{0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10, 30, 60}
	return &KernelMetrics{
		jobDuration:    r.Histogram(MetricAsyncJobDuration, "Async job handler duration in seconds.", []string{"job_type"}, durationBuckets),
		jobsProcessed:  r.Counter(MetricAsyncJobsProcessed, "Async jobs processed partitioned by result.", []string{"job_type", "result"}),
		queueDepth:     r.Gauge(MetricAsyncJobQueueDepth, "Queued async jobs awaiting lease.", []string{"job_type"}),
		jobLag:         r.Gauge(MetricAsyncJobLag, "Age of the oldest queued async job in seconds.", []string{"job_type"}),
		outboxPending:  r.Gauge(MetricOutboxPending, "Pending outbox events awaiting publish.", nil),
		outboxDuration: r.Histogram(MetricOutboxPublishDuration, "Outbox publish attempt duration in seconds.", []string{"result"}, durationBuckets),
		outboxFailures: r.Counter(MetricOutboxPublishFailures, "Outbox publish failures (retry + dead-letter).", nil),
	}
}

func (m *KernelMetrics) ObserveJobProcessed(jobType, result string, duration time.Duration) {
	m.jobsProcessed.Inc(jobType, result)
	m.jobDuration.Observe(duration.Seconds(), jobType)
}

func (m *KernelMetrics) ObserveReaped(jobType string, count int64) {
	m.jobsProcessed.Add(float64(count), firstNonEmptyLabel(jobType), "reaped")
}

func (m *KernelMetrics) ObserveOutboxPublish(result string, duration time.Duration) {
	m.outboxDuration.Observe(duration.Seconds(), result)
}

func (m *KernelMetrics) ObserveOutboxFailure() {
	m.outboxFailures.Inc()
}

func (m *KernelMetrics) SetOutboxPending(pending float64) {
	m.outboxPending.Set(pending)
}

// SetQueueDepth and SetJobLag expose the job-level gauges for callers that
// sample async_jobs backlog (the reaper / a future sampler).
func (m *KernelMetrics) SetQueueDepth(jobType string, depth float64) {
	m.queueDepth.Set(depth, jobType)
}

func (m *KernelMetrics) SetJobLag(jobType string, lagSeconds float64) {
	m.jobLag.Set(lagSeconds, jobType)
}

func firstNonEmptyLabel(s string) string {
	if s == "" {
		return "all"
	}
	return s
}

// InMemoryMetricsRegistry is a registry-agnostic MetricsRegisterer used for P0
// wiring (until F1 lands a Prometheus adapter) and for assertions in tests.
type InMemoryMetricsRegistry struct {
	mu         sync.Mutex
	counters   map[string]*inMemMetric
	gauges     map[string]*inMemMetric
	histograms map[string]*inMemMetric
}

// NewInMemoryMetricsRegistry constructs an empty registry.
func NewInMemoryMetricsRegistry() *InMemoryMetricsRegistry {
	return &InMemoryMetricsRegistry{
		counters:   map[string]*inMemMetric{},
		gauges:     map[string]*inMemMetric{},
		histograms: map[string]*inMemMetric{},
	}
}

type inMemMetric struct {
	mu        sync.Mutex
	name      string
	labelKeys []string
	values    map[string]float64
}

func (m *inMemMetric) Inc(labelValues ...string) { m.Add(1, labelValues...) }
func (m *inMemMetric) Add(value float64, labelValues ...string) {
	m.mu.Lock()
	m.values[joinMetricLabels(labelValues)] += value
	m.mu.Unlock()
}
func (m *inMemMetric) Set(value float64, labelValues ...string) {
	m.mu.Lock()
	m.values[joinMetricLabels(labelValues)] = value
	m.mu.Unlock()
}
func (m *inMemMetric) Observe(value float64, labelValues ...string) {
	m.mu.Lock()
	m.values[joinMetricLabels(labelValues)] += value
	m.mu.Unlock()
}

func (r *InMemoryMetricsRegistry) Counter(name, _ string, labelKeys []string) Counter {
	return r.getOrCreate(r.counters, name, labelKeys)
}
func (r *InMemoryMetricsRegistry) Gauge(name, _ string, labelKeys []string) Gauge {
	return r.getOrCreate(r.gauges, name, labelKeys)
}
func (r *InMemoryMetricsRegistry) Histogram(name, _ string, labelKeys []string, _ []float64) Histogram {
	return r.getOrCreate(r.histograms, name, labelKeys)
}

func (r *InMemoryMetricsRegistry) getOrCreate(into map[string]*inMemMetric, name string, labelKeys []string) *inMemMetric {
	r.mu.Lock()
	defer r.mu.Unlock()
	if m, ok := into[name]; ok {
		return m
	}
	m := &inMemMetric{name: name, labelKeys: append([]string{}, labelKeys...), values: map[string]float64{}}
	into[name] = m
	return m
}

// LabelKeys returns the registered label keys for a metric family, or nil.
func (r *InMemoryMetricsRegistry) LabelKeys(name string) []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, set := range []map[string]*inMemMetric{r.counters, r.gauges, r.histograms} {
		if m, ok := set[name]; ok {
			return append([]string{}, m.labelKeys...)
		}
	}
	return nil
}

// Registered reports whether a metric family with name exists in any kind.
func (r *InMemoryMetricsRegistry) Registered(name string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	_, c := r.counters[name]
	_, g := r.gauges[name]
	_, h := r.histograms[name]
	return c || g || h
}

// Value returns the accumulated value for one (name, labels) tuple.
func (r *InMemoryMetricsRegistry) Value(name string, labelValues ...string) float64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, set := range []map[string]*inMemMetric{r.counters, r.gauges, r.histograms} {
		if m, ok := set[name]; ok {
			m.mu.Lock()
			v := m.values[joinMetricLabels(labelValues)]
			m.mu.Unlock()
			return v
		}
	}
	return 0
}

func joinMetricLabels(labelValues []string) string {
	out := ""
	for i, v := range labelValues {
		if i > 0 {
			out += "\x1f"
		}
		out += v
	}
	return out
}
