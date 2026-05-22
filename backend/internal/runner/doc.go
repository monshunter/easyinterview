// Package runner is the single in-process async job and outbox kernel for the
// backend (docs/spec/backend-async-runner). It consolidates the per-domain
// drainers / runners that previously lived in targetjob, review, resume, auth,
// and jdmatch into one runner.Runtime that owns:
//
//   - the handler registry (Register / Handles),
//   - lease + finalize bookkeeping over async_jobs (LeaseStore, spec D-3),
//   - the shared retry backoff table (BackoffPolicy, spec D-4),
//   - lease-timeout reclamation (Reaper, spec D-5),
//   - graceful shutdown drain (Shutdown, spec D-8),
//   - and the outbox dispatcher (OutboxDispatcher, spec D-6, Phase 3).
//
// Business logic stays in the owning domain packages; the kernel only manages
// the run lifecycle and never deserializes business payloads when finalizing.
package runner
