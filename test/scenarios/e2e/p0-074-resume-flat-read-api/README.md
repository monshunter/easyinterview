# E2E.P0.074 flat resume read API

## 1. Purpose

Validate the backend-resume flat read path: `getResume` / `listResumes`
fixture parity, service/store user scoping, pagination, generated route catalog
boundaries, and non-current API route 404 behavior.

## 2. Requirements

- backend-resume C-6, C-14, C-15
- `docs/spec/backend-resume/plans/002-tailor-runs-and-save-v1/bdd-plan.md`

## 3. Given / When / Then

Given flat ready resume rows owned by user A, user B without access, B2 fixtures
for `getResume` and `listResumes`, and route catalog tests for non-current API
inputs.

When user A reads one resume and lists resumes with cursor pagination, user B
attempts cross-user access, and non-current endpoints are probed.

Then the API returns fixture-compatible flat resume payloads, hides cross-user
records as 404, proves stable `updated_at DESC, id DESC` pagination, confirms
non-current route inputs are absent from generated routes, and keeps raw resume /
suggestion text out of scenario evidence.

## 4. Scripts

- `scripts/setup.sh`: prepares output directories and copies seed / expected outcome notes into `.test-output`.
- `scripts/trigger.sh`: runs fixture validation, non-current route/catalog tests,
  handler fixture parity, and service/store flat read tests.
- `scripts/verify.sh`: rejects skipped or no-op gates, checks required runner markers and PASS evidence, reruns fixture parity, and performs privacy / non-current-vocabulary negative searches.
- `scripts/cleanup.sh`: records cleanup completion while preserving logs under `.test-output/`.

## 5. Evidence

Scenario evidence is written to `.test-output/e2e/p0-074-resume-flat-read-api/`:

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
