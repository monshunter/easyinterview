// Package observability is the AIClient decorator that wraps every
// Complete / Transcribe (and Stream `done` event) with the four mandatory
// outputs: structured log, metric counters, ai_task_runs row, and
// audit_events row.
//
// Spec §4.3 / D-7 forbids plaintext prompt or response from leaking into
// any of these outputs; only sha256 hashes, character lengths, and the
// profile name are allowed in audit metadata. Phase 3.4 white-box tests
// hold the line.
//
// The decorator is the single observability seam: business code MUST NOT
// call the inner Client directly, otherwise per-call metrics, logs, DB
// rows, and audit events will be skipped (spec §6 D-6).
package observability
