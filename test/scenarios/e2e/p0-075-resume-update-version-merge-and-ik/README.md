# E2E.P0.075 flat resume update and IK

## 1. Purpose

Validate `PATCH /api/v1/resumes/{resumeId}` for editable flat resume fields:
structured profile overwrite, idempotency replay / mismatch, server-owned field
rejection, cross-user hiding, fixture parity, and privacy / retired-vocabulary
redlines.

## 2. Requirements

- backend-resume C-14
- `docs/spec/backend-resume/plans/002-versions-tailor-runs-and-save-v1/bdd-plan.md`

## 3. Given / When / Then

Given a ready flat resume owned by user A, one authenticated user B without
access, and the B2 `updateResume` fixture.

When user A patches editable fields, replays the same idempotency key, reuses the key with a changed fingerprint, sends a server-owned field, clears nullable fields, and user B or a deleted row is patched.

Then the API returns fixture-compatible payloads, overwrites editable flat
resume fields while stripping client provenance, returns IK replay / mismatch
semantics, rejects server-owned fields with `422 VALIDATION_FAILED`, hides
cross-user records as 404, and keeps raw resume / profile text out of logs and
scenario evidence.

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

Focused handler/service/store gates prove route, idempotency, validation,
fixture parity, cross-user isolation, and rollback behavior. Skipped focused
gates are scenario failures, not PASS.
