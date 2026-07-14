# E2E.P0.046 message failure and recovery

Scenario ID: `E2E.P0.046`

Mode: automated

Parallel-safe: no. The PostgreSQL integration gate uses fixed, scenario-owned UUIDs and the Playwright parity runner uses the shared local preview port.

## Contract

Given a running Practice session with a candidate message identified by `clientMessageId`, when the provider, reply commit, reply-state finalization, 90-second lease, or 95-second frontend timeout is crossed, then the current service classifies and persists the existing retryable or terminal state without fabricating an assistant reply. A retryable row exposes one row-local retry that reuses the original ID and byte-exact raw Markdown text without replacing the next draft; concurrent reservation, stale generation, missing-ID reconciliation, same-ID mismatch, cross-user access, and privacy deletion remain fail-closed. Raw HTML/event handlers, remote images, and unsafe URIs stay inert while safe external links are hardened.

The same gate constructs UTF-8 strings in memory to prove exact message/session byte limits and limit+1 rejection before the generated client, store, or AI call; no large message fixture is committed.

An authoritative `terminal_failed` row has no retry or duplicate error banner. It keeps the composer locked and offers one source-matched secondary/small CTA whose only route is the read-only current Workspace detail, exactly `/workspace?targetJobId=...`.

## Prerequisites

- `DATABASE_URL` supplies PostgreSQL server credentials whose role can create and drop a temporary database. If it is unset, the trigger reads it from `deploy/dev-stack/.env`; the referenced database schema is never used.
- Host `createdb` and `dropdb` clients are available.
- Repository frontend dependencies and Playwright Chromium are installed.
- The normal local scenario dependencies may be prepared with `test/scenarios/env-setup.sh`. The trigger creates one unique database, applies the repository's current migrations, and force-drops that database through an exit trap; it never reads or rewrites the stale shared baseline schema.

## Run

```bash
test/scenarios/e2e/p0-046-practice-text-loop-failure-and-recovery/scripts/setup.sh
test/scenarios/e2e/p0-046-practice-text-loop-failure-and-recovery/scripts/trigger.sh
test/scenarios/e2e/p0-046-practice-text-loop-failure-and-recovery/scripts/verify.sh
test/scenarios/e2e/p0-046-practice-text-loop-failure-and-recovery/scripts/cleanup.sh
```

## Evidence

The verifier requires:

- `PRACTICE_ISOLATED_POSTGRES_MIGRATIONS_PASS`, the previous real PostgreSQL recovery test plus four exact independent-connection lease/generation/concurrency tests, and `PRACTICE_ISOLATED_POSTGRES_CLEANUP_PASS` in one unique migrated database;
- exact `PRACTICE_PENDING_LEASE_RECOVERY_PASS`, `PRACTICE_STALE_GENERATION_FENCED_PASS`, `PRACTICE_CONCURRENT_RESERVATION_PASS`, `PRACTICE_POST_TIMEOUT_RECONCILIATION_PASS`, and `PRACTICE_TERMINAL_PLAN_RECOVERY_PASS` markers;
- current API, service, and repository PASS markers for retryable persistence, detached bounded finalization, commit-error finalization precedence, exact replay, mismatch, and atomic failure transitions;
- repository exact aggregate and limit+1 PASS markers for both user-message reservation and assistant-message commit, with the session row lock serializing concurrent decisions;
- frontend verbose evidence for exact 95,000 ms abort/reconcile, both stale-read completion orders, missing-ID/read-failure fail lock, and terminal exact route;
- exact `PRACTICE_MARKDOWN_SECURITY_PASS` and `PRACTICE_RAW_RETRY_PASS` markers backed by unit plus browser assertions for zero remote-image requests, inert raw HTML/events, unsafe-URI rejection, hardened links, byte-exact raw retry, same ID, and preserved next draft;
- Practice Playwright PASS markers for hostile Markdown, exact raw row retry, retryable, and terminal failed states in both configured projects, including terminal CTA DOM/style/bbox/viewport and Workspace click-through assertions;
- one shared [`practice-source-fingerprint-paths.json`](../practice-source-fingerprint-paths.json) hash captured by the trigger and byte-recomputed by the verifier; source drift fails closed;
- six stable PNGs under `.test-output/e2e/p0-046-practice-text-loop-failure-and-recovery/screenshots/` for retryable-failed, terminal-failed, and hostile-markdown:
  - desktop CSS viewport `1440x900`, PNG `1440x900`;
  - mobile CSS viewport `390x844`, DPR 3, PNG `1170x2532`.

`result.json` records the current run ID/source fingerprint and, for every PNG, its actual SHA-256, CSS viewport, DPR, and decoded PNG dimensions. Missing, pre-setup, failed, skipped, or no-test evidence is rejected. Raw JD, resume, message, cookie, and provider-secret values are rejected from the trigger log.
