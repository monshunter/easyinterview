// Package targetjob implements the backend TargetJob domain owned by
// docs/spec/backend-targetjob/spec.md and the
// 001-targetjob-import-and-parse-bootstrap plan.
//
// # Public surface
//
// The Handler binds the five TargetJob operations from the generated OpenAPI
// server interface: import, list, detail, update, and archive. Import accepts
// exactly rawText, targetLanguage, and resumeId, plus the Idempotency-Key
// header. Every valid import returns a queued target_import job.
//
// # Persistence and async behavior
//
// target_jobs.raw_jd_text is the only JD text fact. The import transaction
// writes the TargetJob, target.import.requested outbox event, and queued
// target_import job. The parse executor reads raw_jd_text, resolves the F3
// prompt contract, calls A3, then atomically persists structured requirements
// and summary data with target.parsed. There are no URL, file, structured-form,
// source snapshot, or source refresh branches.
//
// # Idempotency
//
// Mutating operations require Idempotency-Key. Keys are hashed with the user id
// and a process secret before persistence, so callers can safely retry without
// duplicating the TargetJob or async job.
//
// # Privacy and observability red lines
//
// Raw JD text, AI prompt or response bodies, provider secrets, authorization
// tokens, and file object URLs must not enter logs, metric labels, audit
// metadata, outbox payloads, or async job payloads. Payload builders enforce
// this contract before persistence. Metrics use only the bounded labels defined
// in observability.go.
//
// # Error behavior
//
// Blank rawText is rejected before enqueue with VALIDATION_FAILED. Retryable AI
// failures use AI_PROVIDER_TIMEOUT or AI_FALLBACK_EXHAUSTED. Invalid output and
// provider configuration failures are non-retryable. Parse failure writes
// target.analysis.failed and removes the incomplete TargetJob transactionally;
// a user retry starts a new paste import.
//
// # BDD entry points
//
// E2E.P0.010 covers direct paste import, idempotency, parse, and ready readback.
// E2E.P0.012 covers retryable and non-retryable AI failures. Both scenarios use
// the generated rawText request and the in-process runner kernel.
package targetjob
