// Package targetjob implements the backend TargetJob domain (handler / service /
// store / drainer / parse executor / outbox emit) per
// docs/spec/backend-targetjob/spec.md and the
// 001-targetjob-import-and-parse-bootstrap plan.
//
// This file owns the F1 metric registry contract for the domain. F1
// observability-stack §3.1.1 lists the three metrics; this file is the
// in-process source of truth for metric names, label keys, and the
// label allowlist. High-cardinality or PII-bearing label keys (target id,
// user id, source URL, prompt body, response body, prompt version, etc.)
// must never appear here — they would leak into Prometheus storage and
// violate spec C-9 / D-8.
package targetjob

// Metric name constants registered in
// docs/spec/observability-stack/spec.md §3.1.1 (TargetJob rows).
const (
	MetricTargetJobImportsTotal         = "target_job_imports_total"
	MetricTargetJobParseDurationSeconds = "target_job_parse_duration_seconds"
	MetricTargetJobParseFailuresTotal   = "target_job_parse_failures_total"
)

// Label key tuples per metric. Order matches the F1 dictionary table so the
// downstream registerer (added in a later phase plan) emits stable label
// orderings on the wire.
var (
	TargetJobImportsLabelKeys       = []string{"service", "operation", "source_type", "result", "error_code"}
	TargetJobParseDurationLabelKeys = []string{"service", "job_type", "source_type", "language", "result"}
	TargetJobParseFailuresLabelKeys = []string{"service", "job_type", "source_type", "language", "error_code", "result"}
)

// f1AllowedTargetJobMetricLabels enumerates every label key any TargetJob
// metric is permitted to use. This is a STRICT allowlist; adding a new key
// requires an F1 spec revision (see observability-stack spec §4.1).
var f1AllowedTargetJobMetricLabels = map[string]struct{}{
	"service":     {},
	"operation":   {},
	"job_type":    {},
	"source_type": {},
	"language":    {},
	"result":      {},
	"error_code":  {},
}

// IsF1AllowedTargetJobMetricLabel reports whether label key is permitted by
// F1 for any TargetJob metric. Used by tests and runtime registration to
// fail fast when a forbidden key (target id, user id, URL, prompt body,
// etc.) is requested.
func IsF1AllowedTargetJobMetricLabel(key string) bool {
	_, ok := f1AllowedTargetJobMetricLabels[key]
	return ok
}

// AllowedTargetJobMetricLabels returns a copy of the F1 allowlist as a sorted
// snapshot. Callers must not mutate the result.
func AllowedTargetJobMetricLabels() []string {
	out := make([]string, 0, len(f1AllowedTargetJobMetricLabels))
	for k := range f1AllowedTargetJobMetricLabels {
		out = append(out, k)
	}
	return out
}
