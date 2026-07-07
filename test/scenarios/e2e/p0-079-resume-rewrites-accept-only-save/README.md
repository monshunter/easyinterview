# E2E.P0.079 Rewrites accept-only save flow

## 1. Purpose

Validate the current flat Resume save contract: non-current accept/reject
operation inputs are absent, flat save fixtures remain valid, and the frontend
Rewrites tab keeps suggestions ephemeral and accept-only until the user saves
through `updateResume` or `duplicateResume`.

## 2. Requirements

- backend-resume C-16
- `docs/spec/backend-resume/plans/002-tailor-runs-and-save-v1/bdd-plan.md`

## 3. Given / When / Then

Given the D-20 flat resume API fixtures, generated route catalog, and current
frontend Rewrites/Detail save surfaces.

When fixture validation runs, non-current route tests probe the suggestion
decision family, handler fixture parity runs for flat save operations, and frontend
Vitest exercises the accept-only Rewrites save flow.

Then accept/reject route inputs stay absent from `cmd/api` and generated route
catalog, `updateResume` / `duplicateResume` / `requestResumeTailor` fixture
parity stays green, and accepted rewrites are saved only through the flat resume
save paths.

## 4. Scripts

- `scripts/setup.sh`: prepares output directories and copies seed / expected outcome notes into `.test-output`.
- `scripts/trigger.sh`: runs fixture validation, non-current route/catalog tests,
  flat save fixture parity, and frontend Rewrites/Detail Vitest coverage.
- `scripts/verify.sh`: rejects skipped or no-op gates, checks required runner markers and PASS evidence, and performs privacy / non-current-vocabulary negative searches.
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
