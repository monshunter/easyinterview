# E2E.P0.076 flat resume duplicate sync paths

## 1. Purpose

Validate the synchronous `duplicateResume` path for the D-20 flat resume model:
route wiring, idempotency, source ownership, flat resume copy persistence,
fixture parity, and privacy / out-of-scope-vocabulary redlines.

## 2. Requirements

- backend-resume C-10
- `docs/spec/backend-resume/plans/002-tailor-runs-and-save-v1/bdd-plan.md`

## 3. Given / When / Then

Given a ready flat resume owned by user A, user B without access, and the B2
`duplicateResume` fixture.

When user A duplicates a resume, optionally overlays editable profile fields,
replays the idempotency key, sends invalid input, or user B accesses the source.

Then duplicate creates a new flat resume that keeps source snapshots and resets
server provenance, idempotency replay returns the first result, invalid input
returns 422, cross-user source access returns 404, and rollback leaves no orphan
rows.

## 4. Scripts

- `scripts/setup.sh`: prepares output directories and copies seed / expected outcome notes into `.test-output`.
- `scripts/trigger.sh`: runs fixture validation, focused `cmd/api` HTTP scenario, handler fixture parity, service tests, store unit tests, and live DB integration gates.
- `scripts/verify.sh`: rejects skipped or no-op gates, checks required runner markers and PASS evidence, reruns fixture parity, and performs privacy / out-of-scope-vocabulary negative searches.
- `scripts/cleanup.sh`: records cleanup completion while preserving logs under `.test-output/`.

## 5. Evidence

Scenario evidence is written to `.test-output/e2e/p0-076-resume-duplicate-save-as-new/`:

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

Focused handler/service/store gates prove route and middleware behavior, source
copy semantics, source isolation, and no-orphan rollback. Skipped focused gates
are scenario failures, not PASS.
