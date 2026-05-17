# E2E.P0.079 resume suggestion accept reject terminal

## 1. Purpose

Validate the resume-tailor suggestion terminal decision flow for backend-resume Phase 8: users can accept or reject pending suggestions exactly once, idempotency replay is stable, already-decided suggestions return the documented 409 envelope, and accepting a suggestion does not mutate `resume_versions.structured_profile`.

## 2. Requirements

- backend-resume C-16
- `docs/spec/backend-resume/plans/002-versions-tailor-runs-and-save-v1/bdd-plan.md`

## 3. Given / When / Then

Given user A owns a ready targeted resume version with pending suggestions and user B owns a separate suggestion.

When user A accepts one suggestion, rejects another, repeats the same idempotency key, retries against an already terminal suggestion, and attempts to decide user B's suggestion.

Then the API returns updated `ResumeVersion` snapshots with terminal suggestion status and `decidedAt`; idempotency replay bypasses duplicate side effects; already terminal suggestions return `VALIDATION_FAILED` with `details.reason=SUGGESTION_ALREADY_DECIDED`; cross-user access returns 404; and the version `structured_profile` remains unchanged.

## 4. Scripts

- `scripts/setup.sh`: prepares output directories and copies seed / expected outcome notes into `.test-output`.
- `scripts/trigger.sh`: runs fixture validation plus focused `cmd/api`, handler, service, and live store gates for accept/reject decisions.
- `scripts/verify.sh`: rejects skipped or no-op gates, checks required runner markers and PASS evidence, and performs privacy / retired-vocabulary negative searches.
- `scripts/cleanup.sh`: records cleanup completion while preserving logs under `.test-output/`.

## 5. Evidence

Scenario evidence is written to `.test-output/e2e/p0-079-resume-suggestion-accept-reject-terminal/`:

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

This scenario verifies the backend API route and persistence state machine through deterministic focused tests. It does not require a deployed frontend because Phase 8 owns backend terminal decision semantics and fixture parity.
