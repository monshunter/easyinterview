# E2E.P0.079 flat save fixtures and read-only detail boundary

## 1. Purpose

Validate the current flat Resume save contract and the frontend read-only
detail boundary: out-of-scope accept/reject operation inputs are absent, flat
save fixtures remain valid, and the frontend detail does not expose Rewrites or
Edit surfaces.

## 2. Requirements

- backend-resume C-16
- `docs/spec/backend-resume/plans/002-tailor-runs-and-save-v1/bdd-plan.md`

## 3. Given / When / Then

Given the D-20 flat resume API fixtures, generated route catalog, and current
frontend read-only Detail surface.

When fixture validation runs, out-of-scope route tests probe the suggestion
decision family, handler fixture parity runs for flat save operations, and
frontend Vitest exercises the read-only detail negative flow.

Then accept/reject route inputs stay absent from `cmd/api` and generated route
catalog, `updateResume` / `duplicateResume` / `requestResumeTailor` fixture
parity stays green, and detail-level Rewrites/Edit UI remains absent.

## 4. Scripts

- `scripts/setup.sh`: prepares output directories and copies seed / expected outcome notes into `.test-output`.
- `scripts/trigger.sh`: runs fixture validation, out-of-scope route/catalog tests,
  flat save fixture parity, and frontend read-only Detail negative coverage.
- `scripts/verify.sh`: rejects skipped or no-op gates, checks required runner markers and PASS evidence, and performs privacy / out-of-scope vocabulary negative searches.
- `scripts/cleanup.sh`: records cleanup completion while preserving logs under `.test-output/`.

## 5. Evidence

Scenario evidence is written to `.test-output/e2e/p0-079-resume-rewrites-accept-only-save/`:

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

This scenario verifies deterministic route/catalog, fixture, and frontend save
semantics. It does not require external services.
