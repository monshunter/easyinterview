// Package targetjob implements the backend TargetJob domain owned by
// docs/spec/backend-targetjob/spec.md and the
// 001-targetjob-import-and-parse-bootstrap plan.
//
// # Public surface (B2 OpenAPI v1 operations)
//
// The Handler binds the four TargetJob operations from the B2 generated
// ServerInterface:
//
//   - POST /targets/import      importTargetJob   (sync; 202 + TargetJobWithJob)
//   - GET  /targets             listTargetJobs    (cursor-paginated)
//   - GET  /targets/{id}        getTargetJob      (full detail incl. requirements)
//   - PATCH /targets/{id}       updateTargetJob   (lifecycle status + notes)
//
// # Idempotency
//
// importTargetJob and updateTargetJob both require an Idempotency-Key
// header. The key is hashed with a per-process pepper plus the user id
// (sha256, prefixed with target_import:) so two different users can use
// the same client-generated key without colliding. Same-user repeats
// return the same targetJobId and the same async_jobs row; the partial
// unique index idx_async_jobs_active_dedupe protects runner-bound paths
// against race-window duplicates. manual_form path inserts a status =
// succeeded marker row for SELECT-based dedupe.
//
// # Source variants and async behaviour
//
//   - url / manual_text / file: sync stage writes target_jobs (analysis_status =
//     queued) + target_job_sources + outbox_events(target.import.requested) +
//     async_jobs(target_import) inside one transaction. The drainer picks the
//     async_jobs row up, fetches the source body for url paths, asks F3 for
//     prompt / rubric / model_profile, calls A3, persists requirements +
//     summary, emits target.parsed, then enqueues an internal-only
//     source_refresh placeholder.
//   - manual_form: sync stage writes target_jobs (analysis_status = ready) +
//     target_job_requirements(must_have draft) + async_jobs(succeeded). It
//     does not emit target.import.requested or target.parsed and is not
//     consumed by the drainer (D-13 / D-11).
//
// # Privacy and observability red lines
//
// Spec C-9 / D-8 forbid raw_jd_text, full source_url, file object URL,
// AI prompt / response body, provider secret, and Authorization tokens
// from log lines, metric labels, audit metadata, outbox payloads, and
// async_jobs payloads. The package enforces this through ForbiddenOutboxFields
// (assertNoForbiddenOutboxFields) and ParseExecutor.redactErrorMessage. F1
// metric names are registered in observability.go; the strict label
// allowlist is asserted by observability_test.go.
//
// # URL fetch boundary
//
// urlfetch.Fetcher implements the spec D-7 envelope: https-only scheme,
// host required, fragment / userinfo stripped, DNS-resolved IP rejection
// for loopback / multicast / link-local / IsPrivate / unspecified /
// 100.64.0.0/10 CGNAT / 198.18.0.0/15 benchmarking / 169.254.169.254
// metadata, redirect hook reapplies the policy per hop, body cap at 1
// MiB, and a User-Agent of EasyInterview JD-Crawler/<version>. Errors
// surface as ErrInvalidSource (B1 TARGET_IMPORT_SOURCE_INVALID, non-
// retryable) or ErrSourceUnavailable (B1
// TARGET_IMPORT_SOURCE_UNAVAILABLE, retryable).
//
// # Error code mapping (D-10)
//
// Retryable: AI_PROVIDER_TIMEOUT, AI_FALLBACK_EXHAUSTED,
// TARGET_IMPORT_SOURCE_UNAVAILABLE.
// Non-retryable: AI_OUTPUT_INVALID, AI_UNSUPPORTED_CAPABILITY,
// AI_PROVIDER_SECRET_MISSING, AI_PROVIDER_CONFIG_INVALID,
// TARGET_IMPORT_SOURCE_INVALID, TARGET_INVALID_STATE_TRANSITION.
//
// # F1 metrics
//
// target_job_imports_total, target_job_parse_duration_seconds, and
// target_job_parse_failures_total are the only TargetJob metrics. Their
// labels are limited to service / operation / job_type / source_type /
// language / result / error_code; high-cardinality or PII-bearing labels
// are rejected by IsF1AllowedTargetJobMetricLabel.
//
// # Frontend mock -> real cutover
//
// frontend-home-job-picks-and-parse can switch the parse screen from
// fixture-backed mocks to the real backend by swapping its generated
// client wiring. The wire shape (request body, headers including
// Idempotency-Key, response body including Job and TargetJobWithJob)
// matches openapi/openapi.yaml exactly; no frontend rewrites are needed
// once cmd/api/main.go registers this package's Handler.
//
// # BDD entry points
//
// Acceptance scenarios live under
// test/scenarios/e2e/p0-010-targetjob-text-import-parse-ready/,
// p0-011-targetjob-url-import-fetch-and-parse/,
// p0-012-targetjob-parse-failure-retryable/,
// p0-013-targetjob-manual-form-ready/. In the current repository state
// these gates execute through repo-tracked go-test scenario scripts
// (setup -> trigger -> verify -> cleanup) and record result.json evidence
// under .test-output/runs/<run-id>/e2e/E2E.P0.010..013/.
package targetjob
