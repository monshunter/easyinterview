# E2E.P0.077 resume tailor async dispatch and ready

## 1. Purpose

Validate the resume tailor async path. In Phase 5 this scenario covers only the `branchResumeVersion` `seedStrategy=ai_select` dispatch slice: provisional targeted version creation, queued `resume_tailor_runs`, queued `async_jobs(resume_tailor)`, route idempotency, fixture parity, and privacy / retired-vocabulary redlines.

Later phases extend this same scenario in place for `requestResumeTailor`, `getResumeTailorRun`, drainer `RunOnce(resume_tailor)`, ready suggestions, `ai_task_runs`, and ready-only outbox.

## 2. Requirements

- backend-resume C-10, C-16
- `docs/spec/backend-resume/plans/002-versions-tailor-runs-and-save-v1/bdd-plan.md`

## 3. Given / When / Then

Given a ready resume asset, one active structured master version owned by user A, one ready target job owned by user A, and the B2 `branchResumeVersion` `ai-select-202-with-job` fixture.

When user A branches with `seedStrategy=ai_select`.

Then the API returns 202 with `BranchResumeVersionAccepted`, the provisional version is persisted, one `resume_tailor_runs` row is queued with `mode='gap_review'`, one `async_jobs` row is queued with `job_type='resume_tailor'` and `resource_type='resume_tailor_run'`, and no raw profile / suggestion text leaks into async job payload or scenario evidence.

## 4. Scripts

- `scripts/setup.sh`: prepares output directories and copies seed / expected outcome notes into `.test-output`.
- `scripts/trigger.sh`: runs fixture validation, focused `cmd/api` branch HTTP scenario, handler fixture parity, service tests, store unit tests, and live DB integration gates for the dispatch slice.
- `scripts/verify.sh`: rejects skipped or no-op gates, checks required runner markers and PASS evidence, reruns fixture parity, and performs privacy / retired-vocabulary negative searches.
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

Phase 5 evidence is intentionally limited to the `ai_select` dispatch slice. Full tailor run execution is not claimed until Phase 7 extends this scenario with drainer and ready-state assertions.
