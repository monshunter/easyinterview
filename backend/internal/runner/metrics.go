package runner

import "time"

// Metrics is the observability sink for the kernel (spec §4.4). Phase 1 ships a
// no-op default; Phase 3.5 wires the real metric families. Keeping it an
// interface lets the runtime emit without depending on a concrete registry.
type Metrics interface {
	// ObserveJobProcessed records one finalized job with its terminal result
	// (succeeded / failed / dead / retried) and wall-clock duration.
	ObserveJobProcessed(jobType, result string, duration time.Duration)
	// ObserveReaped records rows reclaimed by the reaper for a job_type.
	ObserveReaped(jobType string, count int64)
}

// nopMetrics is the default sink used when no Metrics is supplied.
type nopMetrics struct{}

func (nopMetrics) ObserveJobProcessed(string, string, time.Duration) {}
func (nopMetrics) ObserveReaped(string, int64)                       {}
