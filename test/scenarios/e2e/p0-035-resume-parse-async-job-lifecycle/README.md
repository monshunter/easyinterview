# E2E.P0.035 resume parse async job lifecycle

## 1. Purpose

Validate the backend-resume async parse lifecycle from queued `resume_parse` job to deterministic AI parse, upload readable text extraction, ready/failed state transitions, LLM-derived `displayName`, typed AI observability, ready-only outbox emission, and privacy redlines.

## 2. Requirements

- backend-resume C-3, C-4, C-13
- `docs/spec/backend-resume/plans/001-asset-register-parse-and-listing/bdd-plan.md`

## 3. Given / When / Then

Given registered resume assets for `upload` and `paste` sources, an in-process
`resume_parse` drainer, and a deterministic A3/F3-compatible AI client.

When the drainer claims queued jobs and invokes the resume parse handler for success, invalid output, timeout, and retry-exhausted variants.

Then upload PDF / DOCX / Markdown / text sources are converted to readable prompt input and `parsed_text_snapshot`; queued rows keep `display_name` empty until parse success; success writes `parsed_summary`, `parsed_text_snapshot`, `parse_status=ready`, LLM-derived `displayName`, typed `ai_task_runs` metadata, and one `resume.parse.completed` outbox event; failures write `parse_status=failed` with `error_code` and no completed event; parse does not create `resume_versions` before Preview Confirm.

## 4. Scripts

- `scripts/setup.sh`: prepares output directories and copies expected evidence notes.
- `scripts/trigger.sh`: runs the `cmd/api` drainer scenario, runtime wiring test, parse handler tests, and live DB integration gate.
- `scripts/verify.sh`: rejects skips/no-op focused gates, checks required parse lifecycle evidence, and performs privacy / non-current negative searches.
- `scripts/cleanup.sh`: records cleanup completion while preserving logs under `.test-output/`.

## 5. Evidence

Scenario evidence is written to `.test-output/e2e/p0-035-resume-parse-async-job-lifecycle/`:

- `setup.log`
- `trigger.log`
- `verify.log`
- `cleanup.log`
- `expected-outcome.md`

## 6. Baseline

This scenario is backend-owned and intentionally uses the `cmd/api` in-process drainer path. It proves there is no separate worker binary or `WORKER_*` config dependency for the P0 resume parse baseline.

## 7. Offline Limits

The deterministic AI client is acceptable because this scenario validates runtime wiring, contracts, state changes, and privacy boundaries. Missing DB availability for the integration-tag store gate is a scenario failure, not a PASS.
