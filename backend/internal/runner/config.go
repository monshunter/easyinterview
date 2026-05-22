package runner

import (
	"fmt"
	"time"
)

// Priority is the lease-selection bucket a job_type belongs to (spec D-9). It
// is a secondary selection key only: within a bucket rows are still ordered by
// available_at / created_at.
type Priority string

const (
	PriorityCritical Priority = "critical"
	PriorityDefault  Priority = "default"
	PriorityLow      Priority = "low"
)

// priorityOrder lists buckets from highest to lowest precedence. The runtime
// tries to lease from earlier buckets first.
var priorityOrder = []Priority{PriorityCritical, PriorityDefault, PriorityLow}

// defaultJobTypePriority is the spec D-9 fixed assignment over the 9 currently
// executable job_type values. privacy_export / jd_match_search are not
// registered by this plan and intentionally absent.
var defaultJobTypePriority = map[string]Priority{
	"report_generate":     PriorityCritical,
	"privacy_delete":      PriorityCritical,
	"target_import":       PriorityDefault,
	"resume_parse":        PriorityDefault,
	"resume_tailor":       PriorityDefault,
	"debrief_generate":    PriorityDefault,
	"source_refresh":      PriorityLow,
	"email_dispatch":      PriorityLow,
	"jd_match_agent_scan": PriorityLow,
}

// PriorityForJobType returns the spec D-9 bucket for jobType, defaulting to
// PriorityDefault for any job_type without an explicit assignment.
func PriorityForJobType(jobType string) Priority {
	if p, ok := defaultJobTypePriority[jobType]; ok {
		return p
	}
	return PriorityDefault
}

// QueueWeights is the A4 async.queueWeights typed config. The values seed
// fair-scheduling weights per priority bucket; all three must be positive.
type QueueWeights struct {
	Critical int
	Default  int
	Low      int
}

// Config carries the kernel timing + weight knobs sourced from A4 typed config
// (spec D-9 / D-14). Every duration must be positive; the runtime fails fast
// rather than falling back to code constants.
type Config struct {
	ScanInterval   time.Duration
	LeaseTimeout   time.Duration
	ReaperInterval time.Duration
	ShutdownGrace  time.Duration
	QueueWeights   QueueWeights
}

// ConfigFromSeconds builds a Config from the A4 typed-config seconds values
// (spec D-14). cmd/api uses it to map config.AsyncConfig into kernel durations
// without the runner package importing the config loader.
func ConfigFromSeconds(scanSeconds, leaseSeconds, reaperSeconds, graceSeconds int, weights QueueWeights) Config {
	return Config{
		ScanInterval:   time.Duration(scanSeconds) * time.Second,
		LeaseTimeout:   time.Duration(leaseSeconds) * time.Second,
		ReaperInterval: time.Duration(reaperSeconds) * time.Second,
		ShutdownGrace:  time.Duration(graceSeconds) * time.Second,
		QueueWeights:   weights,
	}
}

// Validate enforces the spec D-14 fail-fast contract: all timings and queue
// weights must be positive.
func (c Config) Validate() error {
	if c.ScanInterval <= 0 {
		return fmt.Errorf("runner config: scanInterval must be positive")
	}
	if c.LeaseTimeout <= 0 {
		return fmt.Errorf("runner config: leaseTimeout must be positive")
	}
	if c.ReaperInterval <= 0 {
		return fmt.Errorf("runner config: reaperInterval must be positive")
	}
	if c.ShutdownGrace <= 0 {
		return fmt.Errorf("runner config: shutdownGrace must be positive")
	}
	if c.QueueWeights.Critical <= 0 || c.QueueWeights.Default <= 0 || c.QueueWeights.Low <= 0 {
		return fmt.Errorf("runner config: queueWeights must declare positive critical/default/low values")
	}
	return nil
}

func (w QueueWeights) weightFor(p Priority) int {
	switch p {
	case PriorityCritical:
		return w.Critical
	case PriorityLow:
		return w.Low
	default:
		return w.Default
	}
}
