# E2E.P0.033 file presign register roundtrip

## 1. Purpose

Validate the backend-upload baseline path for `createUploadPresign`, object registration, idempotency replay, purpose validation, cross-user isolation, and privacy data-erasure ordering.

## 2. Requirements

- backend-upload C-1, C-2, C-3, C-4, C-6, C-7, C-8
- `docs/spec/backend-upload/plans/001-file-objects-and-presign-baseline/bdd-plan.md`

## 3. Scripts

- `scripts/setup.sh`: prepares deterministic scenario files and records current environment capability.
- `scripts/trigger.sh`: runs the backend-upload focused test suite plus the live `TestUploadPresignRegisterPrivacyDeleteLiveRoundtrip` gate that exercises HTTP presign, signed PUT, register, and privacy delete behavior.
- `scripts/verify.sh`: checks fixture parity, state-machine evidence, privacy tombstone redlines, and non-current-pattern negative searches.
- `scripts/cleanup.sh`: removes temporary binary inputs while preserving logs under `.test-output/`.

## 4. Evidence

Scenario evidence is written to `.test-output/e2e/p0-033-file-presign-register-roundtrip/`:

- `setup.log`
- `trigger.log`
- `verify.log`
- `cleanup.log`
- `expected-outcome.md`

## 5. Live Stack Contract

`trigger.sh` requires live `DATABASE_URL` and complete `OBJECT_STORAGE_*` settings. Missing live DB / MinIO configuration is a scenario failure, not a skip, because E2E.P0.033 is used as BDD evidence for presign → PUT → register → privacy delete behavior. `verify.sh` also rejects skipped live integration checks, focused gates that match no tests, or trigger logs that lack `TestUploadPresignRegisterPrivacyDeleteLiveRoundtrip`.

## 6. Baseline

This baseline is intentionally backend-owned. It does not claim that `resume_assets` or `target_jobs` privacy order is globally owned by backend-upload; those cross-domain chains remain with B4 and each business owner.

## 7. Offline Limits

Offline unit, sqlmock, and fixture parity gates remain useful focused checks, but they are not sufficient to mark E2E.P0.033 PASS. `verify.sh` rejects trigger logs that contain live integration skips.
