# E2E.P0.077 resume tailor async dispatch and ready

## 1. Purpose

Validate the D-20 flat resume tailor async path: `requestResumeTailor`,
`getResumeTailorRun`, request idempotency, mode validation, queued run/job
creation, run status reads, state transitions, concurrent claim, fixture parity,
cross-user isolation, runner kernel `RunOnce(resume_tailor)`, ready suggestions,
typed `ai_task_runs`, and ready-only `resume.tailor.completed` outbox payload
privacy.

## 2. Requirements

- backend-resume C-10, C-16
- `docs/spec/backend-resume/plans/002-tailor-runs-and-save-v1/bdd-plan.md`

## 3. Given / When / Then

Given a ready flat resume owned by user A, one ready target job owned by user A,
and the B2 `requestResumeTailor` / `getResumeTailorRun` fixtures.

When user A requests a tailor run and polls the run by ID.

Then the API returns 202 with `ResumeTailorRunWithJob`; queued state is held by
`async_jobs(job_type='resume_tailor')` with a `resumeId` payload; getTailorRun
returns queued / generating / ready / failed variants from async job state and
task output; state transitions reject double-claim; the runner kernel marks the job
ready, persists the run result and ephemeral suggestions in `async_jobs.result`,
writes a typed `ai_task_runs` row, and emits one ready-only completed outbox
event whose payload contains only IDs, mode, and status.

## 4. Scripts

- `scripts/setup.sh`: prepares output directories and copies seed / expected outcome notes into `.test-output`.
- `scripts/trigger.sh`: runs fixture validation, focused `cmd/api` tailor HTTP
  scenarios, handler fixture parity, service tests, store unit tests, and
  runner kernel ready-path gates.
- `scripts/verify.sh`: rejects skipped or no-op gates, checks required runner markers and PASS evidence, reruns fixture parity, and performs privacy / out-of-scope-vocabulary negative searches.
- `scripts/cleanup.sh`: records cleanup completion while preserving logs under `.test-output/`.

## 5. Evidence

Scenario evidence is written to `.test-output/e2e/p0-077-resume-tailor-async-dispatch-and-ready/`:

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

This scenario is still local and deterministic: it verifies the in-process runner kernel through focused `cmd/api`, job handler, and live store gates rather than a long-running external worker.
