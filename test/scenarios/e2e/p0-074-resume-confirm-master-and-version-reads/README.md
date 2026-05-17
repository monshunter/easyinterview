# E2E.P0.074 resume confirm master and version reads

## 1. Purpose

Validate the backend-resume v1 save path after parse readiness: confirm a structured master resume version, replay the idempotent request, reject duplicates, read the saved version, list versions with cursor pagination, hide cross-user records, and preserve privacy / retired-vocabulary redlines.

## 2. Requirements

- backend-resume C-6, C-14, C-15
- `docs/spec/backend-resume/plans/002-versions-tailor-runs-and-save-v1/bdd-plan.md`

## 3. Given / When / Then

Given a ready resume asset, a processing resume asset, two authenticated users, B2 fixtures for `confirmResumeStructuredMaster`, `getResumeVersion`, and `listResumeVersions`, and migration `000007_resume_versions_structured_master_unique` applied.

When user A confirms the structured master, replays the same idempotency key, retries with a new key, requests the saved version, lists versions across cursor pages, requests an empty asset list, sends an invalid cursor, and user B accesses user A records.

Then the API returns fixture-compatible payloads, writes exactly one active `structured_master` version per asset, returns 409 for duplicate active master creation, returns 422 for invalid input / parse-not-ready, hides cross-user records as 404, proves stable `updated_at DESC, id DESC` pagination, and keeps raw resume / suggestion text out of scenario evidence.

## 4. Scripts

- `scripts/setup.sh`: prepares output directories and copies seed / expected outcome notes into `.test-output`.
- `scripts/trigger.sh`: runs the focused `cmd/api` HTTP scenarios, handler fixture parity, service/store read tests, fixture validation, and live DB integration gate.
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

The `cmd/api` HTTP scenario proves route and middleware behavior. Store integration tests prove live database uniqueness, cross-user isolation, and cursor pagination with concrete `DATABASE_URL`. Missing DB availability or skipped integration gates are scenario failures, not PASS.
