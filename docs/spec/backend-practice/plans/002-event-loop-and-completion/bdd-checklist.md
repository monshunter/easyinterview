# 002 Conversation Message Loop BDD Checklist

> **版本**: 2.8
> **状态**: completed
> **更新日期**: 2026-07-14

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## E2E.P0.044 Full conversation
- [x] Revise/run/record scenario evidence.
## E2E.P0.046 Failure recovery
- [x] Revise/run/record scenario evidence.
- [x] Remediation: execute and verify provider-failure, exact-replay, mismatch, pending-retry and concurrent-new-message named tests. (E2E.P0.046 PASS)
## E2E.P0.047 Completion
- [x] Revise/run/record scenario evidence.
- [x] Remediation: execute and verify late assistant commit rollback after completion wins. (E2E.P0.047 PASS)

## Phase 7 resume grounding
- [x] P0.044 executes and verifies complete follow-up snapshot tail marker.<!-- verified: 2026-07-12 method=scenario -->
- [x] P0.046 executes and verifies empty-context typed failure, zero AI/assistant reply, and retryable user reservation.<!-- verified: 2026-07-12 method=scenario -->

## Phase 8 completion ledger

- [x] P0.047 executes and verifies atomic completion fact plus exact replay without duplicates.<!-- verified: 2026-07-12 method=scenario-run result=PASS -->
- [x] P0.098 executes wrong-resume completion exclusion plus persisted first-to-next-existing and final projection with no frontend business-state storage.<!-- verified: 2026-07-12 method=real-postgres+scenario-run marker=wrong-resume-completion-ignored=PASS -->

## Phase 9 reportable completion/context

- [x] P0.047 setup includes zero-answer, pending-reply and one-answer sessions plus a run correlation value containing no cookie/raw business content.<!-- verified: 2026-07-12 method=scenario-run -->
- [x] Trigger runs `cd backend && go test ./internal/api/practice ./internal/practice ./internal/store/practice -run '^(TestE2EP0047RejectsZeroAnswerCompletion|TestE2EP0047FreezesReportContext|TestE2EP0047CompletionReplayPreservesReportContext)$' -count=1 -v` and records exact RUN/PASS output.<!-- verified: 2026-07-12 method=scenario-run -->
- [x] Verify requires `ZERO_ANSWER_COMPLETION_REJECTED_PASS`, `REPORT_CONTEXT_SNAPSHOT_PASS`, `REPORT_CONTEXT_REPLAY_PASS`, frontend disabled reason, zero forbidden FAIL/no-test output and redacted DB no-side-effect/same-snapshot assertions.<!-- verified: 2026-07-12 method=scenario-run -->
- [x] Write `completion-backend-evidence.json` with exact `practice-completion-evidence.v1` keys and `result=PASS`; P0.056/058 consume this artifact rather than recreating completion evidence.<!-- verified: 2026-07-12 method=scenario-run -->
- [x] Cleanup removes scenario rows and preserves only the redacted owner artifact/marker evidence.<!-- verified: 2026-07-12 method=scenario-run evidence="only completion-backend-evidence.json remains" -->

## Phase 10 server-recoverable reply state

- [x] P0.044 proves pending readback carries the original user clientMessageId and pending status, then commits exactly one assistant reply.
- [x] P0.046 proves retryable AI failure is persisted before error response, survives reload, and same-ID retry converges without duplicate user/reply rows.
- [x] P0.046 proves terminal failure readback has no retry path, cross-user access stays hidden, and no raw message/error content leaks outside authorized response/session content.
- [x] Current setup/trigger/verify/cleanup run records named backend + frontend recovery markers; historical PASS cannot close Phase 10.

## Phase 11 lease-bounded generation recovery

- [x] P0.044 proves immediate and persisted pending before lease expiry, then GET-based expiry convergence without duplicate send, using fresh desktop/mobile evidence.
- [x] P0.046 executes the four exact real PostgreSQL concurrency tests and verifies lease recovery, one winning G2, stale G1 Commit/Fail fencing and one assistant reply.
- [x] P0.046 proves the 95-second frontend timeout reconciles the same clientMessageId and terminal failure exposes a generic current-plan recovery CTA with no row retry.
- [x] Both scenarios bind setup/trigger/verify evidence to one tracked source fingerprint and per-screenshot SHA-256/dimensions/viewport；verifier-time drift, missing paths, historical PASS, FAIL or no-tests fail closed.
- [x] Serial setup → trigger → verify → cleanup passes on current code and isolated migrated PostgreSQL with every exact Phase 11 marker.
  <!-- verified: 2026-07-14 evidence="Fresh P0.044/P0.046 serial runs passed current migration, contract, real concurrency, marker, fingerprint and eight-PNG evidence gates; isolated database residual=0." -->
