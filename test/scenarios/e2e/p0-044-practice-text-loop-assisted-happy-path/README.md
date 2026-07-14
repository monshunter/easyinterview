# E2E.P0.044 continuous practice conversation

Scenario ID: `E2E.P0.044`

Mode: automated

Parallel-safe: no. The Playwright parity runner uses the shared local preview port and frontend build artifacts.

## Contract

Given an authenticated running Practice session, when the candidate submits a text answer, then the answer is rendered immediately with `replyStatus=pending`, the composer and Finish action remain locked until the reply settles, a reload preserves the pending projection before the exact 90-second lease boundary, and a successful commit converges to exactly one user/assistant pair.

The wrapper executes the current frontend state-machine tests, focused API/service/repository owner tests, and the existing Practice Playwright parity tests. The real PostgreSQL recovery matrix belongs to `E2E.P0.046`, so this happy-path wrapper remains environment-independent.

## Prerequisites

- Repository frontend dependencies and Playwright Chromium are installed.

## Run

```bash
test/scenarios/e2e/p0-044-practice-text-loop-assisted-happy-path/scripts/setup.sh
test/scenarios/e2e/p0-044-practice-text-loop-assisted-happy-path/scripts/trigger.sh
test/scenarios/e2e/p0-044-practice-text-loop-assisted-happy-path/scripts/verify.sh
test/scenarios/e2e/p0-044-practice-text-loop-assisted-happy-path/scripts/cleanup.sh
```

## Evidence

The verifier requires:

- API projection and repository PASS markers for user-only `clientMessageId` / `replyStatus`, pending readback, same-ID retry, and atomic reply completion;
- loader and screen verbose evidence plus `TestSQLRepositoryGetSessionKeepsPendingBeforeLeaseBoundary`, with exact `PRACTICE_IMMEDIATE_PENDING_PASS` / `PRACTICE_PERSISTED_PENDING_PASS` markers;
- Practice Playwright PASS markers for immediate pending and persisted pending states in both configured projects;
- one shared [`practice-source-fingerprint-paths.json`](../practice-source-fingerprint-paths.json) hash captured by the trigger and byte-recomputed by the verifier; source drift fails closed;
- four stable PNGs under `.test-output/e2e/p0-044-practice-text-loop-assisted-happy-path/screenshots/`:
  - desktop CSS viewport `1440x900`, PNG `1440x900`;
  - mobile CSS viewport `390x844`, DPR 3, PNG `1170x2532`.

`result.json` records the current run ID/source fingerprint and, for every PNG, its actual SHA-256, CSS viewport, DPR, and decoded PNG dimensions. Missing, pre-setup, failed, skipped, or no-test evidence is rejected. Raw message, cookie, and provider-secret values are rejected from the trigger log.
