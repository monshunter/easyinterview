# E2E.P0.033 file presign register roundtrip

## 1. Purpose

Validate the backend-upload baseline path for `createUploadPresign`, object registration, idempotency replay, purpose validation, cross-user isolation, and privacy deletion ordering.

## 2. Requirements

- backend-upload C-1, C-2, C-3, C-4, C-6, C-7, C-8
- `docs/spec/backend-upload/plans/001-file-objects-and-presign-baseline/bdd-plan.md`

## 3. Scripts

- `scripts/setup.sh`: prepares deterministic scenario files and records current environment capability.
- `scripts/trigger.sh`: runs the backend-upload focused test suite that exercises presign, register, object store, and privacy delete behavior.
- `scripts/verify.sh`: checks fixture parity, state-machine evidence, privacy tombstone redlines, and retired-pattern negative searches.
- `scripts/cleanup.sh`: removes temporary binary inputs while preserving logs under `.test-output/`.

## 4. Evidence

Scenario evidence is written to `.test-output/e2e/p0-033-file-presign-register-roundtrip/`:

- `setup.log`
- `trigger.log`
- `verify.log`
- `cleanup.log`
- `expected-outcome.md`

## 5. Live Stack Contract

When the A2 dev stack and MinIO credentials are available, `trigger.sh` also runs integration-tag tests. Without `DATABASE_URL` / `OBJECT_STORAGE_*`, those integration tests follow the repository contract and skip live network checks while preserving unit and fixture gates.

## 6. Baseline

This baseline is intentionally backend-owned. It does not claim that `resume_assets` or `target_jobs` privacy order is globally owned by backend-upload; those cross-domain chains remain with B4 and each business owner.

## 7. Offline Limits

The current harness validates the upload roundtrip through unit, sqlmock, fixture parity, and integration-tag smoke gates. Full HTTP `DELETE /api/v1/me` execution requires the future backend async runner wiring and the A2 dev stack to be active in the shared test environment.
