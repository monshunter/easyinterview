# E2E.P0.074 flat resume reads and retired version routes

## 1. Purpose

Validate the D-20 backend-resume flat read path: `getResume` / `listResumes`
fixture parity, service/store user scoping and pagination, generated route catalog
cleanup, and retired `/resume-versions` / structured-master route 404 behavior.

## 2. Requirements

- backend-resume C-6, C-14, C-15
- `docs/spec/backend-resume/plans/002-versions-tailor-runs-and-save-v1/bdd-plan.md`

## 3. Given / When / Then

Given flat ready resume rows owned by user A, user B without access, B2 fixtures
for `getResume` and `listResumes`, and D-20 route catalog tests for retired
version operations.

When user A reads one resume and lists resumes with cursor pagination, user B
attempts cross-user access, and old version endpoints are probed.

Then the API returns fixture-compatible flat resume payloads, hides cross-user
records as 404, proves stable `updated_at DESC, id DESC` pagination, confirms
old version operations are absent from generated routes, and keeps raw resume /
suggestion text out of scenario evidence.

## 4. Scripts

- `scripts/setup.sh`: prepares output directories and copies seed / expected outcome notes into `.test-output`.
- `scripts/trigger.sh`: runs fixture validation, retired route/catalog tests,
  handler fixture parity, and service/store flat read tests.
- `scripts/verify.sh`: rejects skipped or no-op gates, checks required runner markers and PASS evidence, reruns fixture parity, and performs privacy / retired-vocabulary negative searches.
- `scripts/cleanup.sh`: records cleanup completion while preserving logs under `.test-output/`.

## 5. Evidence

Scenario evidence is written to `.test-output/e2e/p0-074-resume-confirm-master-and-version-reads/`:

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

The focused tests prove route/catalog cleanup, handler fixture parity, and flat
store/service behavior. Skipped focused gates are scenario failures, not PASS.
