# E2E.P0.075 resume update version merge and IK

## 1. Purpose

Validate `PATCH /api/v1/resume-versions/{resumeVersionId}` for editable resume version fields: partial `structured_profile` merge, idempotency replay / mismatch, server-owned field rejection, cross-user hiding, soft-delete hiding, fixture parity, and privacy / retired-vocabulary redlines.

## 2. Requirements

- backend-resume C-14
- `docs/spec/backend-resume/plans/002-versions-tailor-runs-and-save-v1/bdd-plan.md`

## 3. Given / When / Then

Given a ready resume asset, one active `structured_master` version owned by user A, one authenticated user B without access, the B2 `updateResumeVersion` fixture, and migration `000007_resume_versions_structured_master_unique` applied.

When user A patches editable fields, replays the same idempotency key, reuses the key with a changed fingerprint, sends a server-owned field, clears nullable fields, and user B or a deleted row is patched.

Then the API returns fixture-compatible payloads, merges the partial profile without dropping existing profile sections, returns IK replay / mismatch semantics, rejects server-owned fields with `422 VALIDATION_FAILED`, hides cross-user and soft-deleted records as 404, and keeps raw resume / profile text out of logs and scenario evidence.

## 4. Scripts

- `scripts/setup.sh`: prepares output directories and copies seed / expected outcome notes into `.test-output`.
- `scripts/trigger.sh`: runs fixture validation, focused `cmd/api` HTTP scenario, handler fixture parity, service tests, store unit tests, and live DB integration gates.
- `scripts/verify.sh`: rejects skipped or no-op gates, checks required runner markers and PASS evidence, reruns fixture parity, and performs privacy / retired-vocabulary negative searches.
- `scripts/cleanup.sh`: records cleanup completion while preserving logs under `.test-output/`.

## 5. Evidence

Scenario evidence is written to `.test-output/e2e/p0-075-resume-update-version-merge-and-ik/`:

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

The `cmd/api` HTTP scenario proves route and middleware behavior. Store integration tests prove live database merge, cross-user isolation, soft-delete isolation, and rollback behavior with concrete `DATABASE_URL`. Missing DB availability or skipped integration gates are scenario failures, not PASS.
