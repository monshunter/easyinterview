# E2E.P0.078 resume tailor failure and retry

## 1. Purpose

Validate the `resume_tailor` async failure and retry semantics for backend-resume Phase 7: AI timeout is retryable, invalid AI output is terminal, retry moves a failed run back through generating to ready, and completed outbox events are ready-only.

## 2. Requirements

- backend-resume C-16
- `docs/spec/backend-resume/plans/002-versions-tailor-runs-and-save-v1/bdd-plan.md`

## 3. Given / When / Then

Given three queued resume tailor runs owned by user A and a deterministic `cmd/api` in-process drainer.

When the drainer processes timeout, invalid-output, and timeout-then-success variants.

Then failed runs persist `resume_tailor_runs.status='failed'` with `AI_PROVIDER_TIMEOUT` or `AI_OUTPUT_INVALID`; retryable state remains in `async_jobs` outcome metadata; retry can re-enter generating and finish ready; `ai_task_runs` records every AI attempt; and `resume.tailor.completed` is emitted only for the final ready run.

## 4. Scripts

- `scripts/setup.sh`: prepares output directories and copies seed / expected outcome notes into `.test-output`.
- `scripts/trigger.sh`: runs focused `cmd/api`, job handler, and store gates for failure, retry, task-run, and ready-only outbox behavior.
- `scripts/verify.sh`: rejects skipped or no-op gates, checks required runner markers and PASS evidence, and performs privacy / retired-vocabulary negative searches.
- `scripts/cleanup.sh`: records cleanup completion while preserving logs under `.test-output/`.

## 5. Evidence

Scenario evidence is written to `.test-output/e2e/p0-078-resume-tailor-failure-and-retry/`:

- `setup.log`
- `trigger.log`
- `verify.log`
- `cleanup.log`
- `seed-input.md`
- `expected-outcome.md`

## 6. Isolation

- Environment: shared local scenario environment.
- Parallel safe: No.
- Cleanup is idempotent and preserves evidence logs.

## 7. Offline Limits

This scenario verifies the backend drainer through deterministic focused tests. It does not require a separate worker binary because this plan explicitly owns the `cmd/api` in-process drainer topology.
