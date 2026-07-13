package runner

import "time"

// MaxAttempts is the shared retry ceiling for every async_jobs job_type
// (spec D-4 / ADR-Q2 §3.4). A retryable failure at attempts >= MaxAttempts is
// finalized as dead instead of being requeued.
const MaxAttempts int32 = 5

// defaultBackoffSchedule is the business async-job retry policy. The Nth retry
// waits schedule[N-1] and later retries remain capped at 80 seconds.
var defaultBackoffSchedule = []time.Duration{
	10 * time.Second,
	20 * time.Second,
	40 * time.Second,
	80 * time.Second,
}

// outboxBackoffSchedule is intentionally independent from business jobs. Event
// delivery remains infrastructure work and retains the slower delivery policy.
var outboxBackoffSchedule = []time.Duration{
	30 * time.Second,
	2 * time.Minute,
	10 * time.Minute,
	1 * time.Hour,
	6 * time.Hour,
}

// BackoffPolicy maps an attempt count to the delay before the next retry. It is
// the only place retry spacing is defined; domain handlers must not invent
// their own backoff.
type BackoffPolicy struct {
	schedule []time.Duration
}

// DefaultBackoffPolicy returns the business async-job policy:
// [10s, 20s, 40s, 80s], capped at 80s.
func DefaultBackoffPolicy() BackoffPolicy {
	return BackoffPolicy{schedule: defaultBackoffSchedule}
}

// DefaultOutboxBackoffPolicy returns the infrastructure delivery policy:
// [30s, 2m, 10m, 1h, 6h].
func DefaultOutboxBackoffPolicy() BackoffPolicy {
	return BackoffPolicy{schedule: outboxBackoffSchedule}
}

// Next returns the delay before the next retry for the given attempt count.
// attempts < 1 returns the first delay; attempts >= len(schedule) returns the
// last delay (the schedule is clamped at both ends).
func (p BackoffPolicy) Next(attempts int32) time.Duration {
	schedule := p.schedule
	if len(schedule) == 0 {
		schedule = defaultBackoffSchedule
	}
	if attempts < 1 {
		return schedule[0]
	}
	idx := int(attempts) - 1
	if idx >= len(schedule) {
		return schedule[len(schedule)-1]
	}
	return schedule[idx]
}
