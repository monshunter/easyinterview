# E2E.P0.076 resume branch version sync paths

## 1. Purpose

Validate the synchronous `branchResumeVersion` paths for `seedStrategy=copy_master` and `seedStrategy=blank`: route wiring, idempotency, parent / target ownership, targeted version persistence, fixture parity, and privacy / retired-vocabulary redlines.

## 2. Requirements

- backend-resume C-10
- `docs/spec/backend-resume/plans/002-versions-tailor-runs-and-save-v1/bdd-plan.md`

## 3. Given / When / Then

Given a ready resume asset, one active structured master version owned by user A, one ready target job owned by user A, user B without access, and the B2 `branchResumeVersion` fixture.

When user A branches with `copy_master`, branches with `blank`, replays the idempotency key, sends an invalid seed strategy, and user B or a foreign target job is used.

Then copy-master creates a targeted version that keeps parent profile content with server-reset branch provenance, blank creates an empty editable profile, idempotency replay returns the first result, invalid input returns 422, cross-user parent / target access returns 404, and no async job is created for synchronous strategies.

## 4. Scripts

- `scripts/setup.sh`: prepares output directories and copies seed / expected outcome notes into `.test-output`.
- `scripts/trigger.sh`: runs fixture validation, focused `cmd/api` HTTP scenario, handler fixture parity, service tests, store unit tests, and live DB integration gates.
- `scripts/verify.sh`: rejects skipped or no-op gates, checks required runner markers and PASS evidence, reruns fixture parity, and performs privacy / retired-vocabulary negative searches.
- `scripts/cleanup.sh`: records cleanup completion while preserving logs under `.test-output/`.

## 5. Evidence

Scenario evidence is written to `.test-output/e2e/p0-076-resume-branch-version-sync-paths/`:

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

The `cmd/api` HTTP scenario proves route and middleware behavior. Store integration tests prove live database copy / blank semantics, parent / target isolation, and no-orphan rollback with concrete `DATABASE_URL`. Missing DB availability or skipped integration gates are scenario failures, not PASS.
