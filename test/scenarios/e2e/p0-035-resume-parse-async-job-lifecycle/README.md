# E2E.P0.035 resume parse async job lifecycle

## 1. Purpose

Validate the backend-resume async parse lifecycle from queued `resume_parse` job to deterministic AI parse, complete input tail-marker preservation, deterministic source snapshots, structured-only model output, `finish_reason=length` fail-closed behavior, long-resume output budget, ready/failed state transitions, typed AI observability, ready-only outbox emission, and privacy redlines.

## 2. Requirements

- backend-resume C-3, C-4, C-13
- `docs/spec/backend-resume/plans/001-asset-register-parse-and-listing/bdd-plan.md`

## 3. Given / When / Then

Given registered resumes for `upload` and `paste` sources, an in-process
`resume_parse` runner kernel, and a deterministic AI client implementing the A3/F3 contracts.

When the runner kernel claims queued jobs and invokes the resume parse handler for structured-only success, invalid output, `finish_reason=length`, timeout, and retry-exhausted variants.

Then `resume.parse.default.max_tokens` is at least 8192 for structured output safety; the complete source, including a long-input tail marker, reaches the AI prompt and a deterministic `parsed_text_snapshot`; the model does not echo full Markdown; `finish_reason=length` writes `AI_OUTPUT_INVALID`, keeps the complete snapshot, and emits no completed event. Configured extracted-text exact/limit+1 cases are constructed in memory, and overflow fails before AI. DOCX and unreadable PDF input fail before AI; success writes ready state, structured fields, LLM-derived `displayName`, typed task metadata, and one ready-only completed event.

## 4. Scripts

- `scripts/setup.sh`: prepares output directories and copies expected evidence notes.
- `scripts/trigger.sh`: runs the profile output-budget gate, `cmd/api` runner kernel scenario, runtime wiring test, parse handler tests, and live DB integration gate.
- `scripts/verify.sh`: rejects skips/no-op focused gates, checks required parse lifecycle evidence, and performs privacy / current-scope negative searches.
- `scripts/cleanup.sh`: records cleanup completion while preserving logs under `.test-output/`.

## 5. Evidence

Scenario evidence is written to `.test-output/e2e/p0-035-resume-parse-async-job-lifecycle/`:

- `setup.log`
- `trigger.log`
- `verify.log`
- `cleanup.log`
- `expected-outcome.md`

## 6. Baseline

This scenario is backend-owned and intentionally uses the `cmd/api` in-process runner kernel path. It proves there is no separate worker binary or `WORKER_*` config dependency for the P0 resume parse baseline.

## 7. Offline Limits

The deterministic AI client is acceptable because this scenario validates runtime wiring, contracts, state changes, and privacy boundaries. Missing DB availability for the integration-tag store gate is a scenario failure, not a PASS.
