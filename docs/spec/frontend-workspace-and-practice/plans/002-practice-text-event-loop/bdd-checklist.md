# 002 Practice Continuous Conversation BDD Checklist

> **版本**: 2.7
> **状态**: completed
> **更新日期**: 2026-07-14

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## E2E.P0.044
- [x] Revise/run/record conversation happy-path evidence.
- [x] Revision 2026-07-13 holds send pending and proves immediate user row, composer clear/lock, accessible interviewer-thinking, then reloads a server `replyStatus=pending` row and proves same-ID/no-resend recovery, success dedupe and clean 1440/390 screenshots. (`pending-reply` fixture is current and validated.)
  <!-- verified: 2026-07-13 evidence="E2E.P0.044 PASS; focused frontend/API/service/store gates and 6 Playwright desktop/mobile cases produced four verified pending-state PNGs." -->
## E2E.P0.045
- [x] Revise/run/record simplified UI and disabled-phone evidence.
## E2E.P0.046
- [x] Revise/run/record failure/retry evidence.
- [x] Remediation: execute loader refresh and same-message retry screen assertions. (E2E.P0.046 PASS)
- [x] Revision 2026-07-13 proves typed transport/`ApiClientError.apiError.retryable` classification and thinking removal；AI failure → reload restores server original text/same clientMessageId as `retryable_failed`, row-local retry preserves draft and converges to one user/reply pair；terminal validation/auth/not-found/conflict/mismatch has no retry and re-reads truth；all unresolved states block Finish；1440/390 failed-state screenshots match source. (`retryable-failed-reply`, `terminal-failed-reply` and terminal error fixtures are current and validated.)
  <!-- verified: 2026-07-13 evidence="E2E.P0.046 PASS against current migrations in an isolated PostgreSQL database; 4 desktop/mobile Playwright cases and four verified failure-state PNGs passed; cleanup left no temporary database." -->
## E2E.P0.047
- [x] Revise/run/record completion evidence.
- [x] Remediation: execute completion retry routing and Finish CTA lifecycle assertions. (E2E.P0.047 PASS)

## E2E.P0.047 Phase 7 zero-answer completion

- [x] Prepare opening-only and one-committed-user-message sessions without raw message evidence.
- [x] Trigger frontend native-disabled/zh-en described-reason assertions and consume backend authoritative rejection/no-side-effect markers.
- [x] Verify `ZERO_ANSWER_FINISH_DISABLED_PASS`, `ZERO_ANSWER_COMPLETION_REJECTED_PASS`, one-answer stable reportId handoff and exact replay; cleanup scenario rows.
  <!-- verified: 2026-07-13 evidence="E2E.P0.047 serial setup/trigger/verify/cleanup PASS against an isolated migrated PostgreSQL database." -->

## E2E.P0.047 Phase 8 reportId-only handoff

- [x] After one-answer completion, assert browser URL/history navigation and downstream report request contain only reportId; copied target/plan/session/resume/round/status/error fields are absent.
- [x] Replay completion and prove the same reportId locator returns without duplicate report navigation state.
  <!-- verified: 2026-07-13 evidence="PracticeScreen URL/history/downstream locator test and P0.047 completion Idempotency-Key plus backend replay markers PASS." -->

## Phase 10 T-B/P-A lease-aligned recovery

- [x] P0.044 captures fresh desktop/mobile immediate and persisted pending states, proves no duplicate send before lease convergence, and emits immediate/persisted/fingerprint markers.
- [x] P0.046 proves exact 95-second abort + same-ID reconciliation, every authoritative server status, uncertain-read fail-lock and stale-response suppression.
- [x] P0.046 consumes the four exact real PostgreSQL lease/generation concurrency tests and emits lease/fence/concurrency markers.
- [x] Historical Phase 10 P0.046 captured the generic terminal state at desktop/mobile and proved the then-current `parse(targetJobId)` action；Phase 11 supersedes only this destination.
- [x] Both scenarios verify one tracked source SHA-256 plus every screenshot SHA-256/dimensions/viewport；source drift, historical artifacts, FAIL or no-tests fail closed.
- [x] Current serial setup → trigger → verify → cleanup passes for P0.044 then P0.046 and records all exact Phase 10 markers.
  <!-- verified: 2026-07-14 evidence="Fresh P0.044/P0.046 serial runs passed all current source-fingerprint, marker, eight-PNG geometry/hash and isolated PostgreSQL cleanup gates." -->

## Phase 11 safe Markdown/GFM and Workspace-detail recovery

- [x] P0.044 renders persisted user and assistant GFM through the shared safe projection and records current 1440/390 DOM/style/bbox/viewport/screenshot evidence with zero document overflow.
  <!-- verified: 2026-07-14 evidence="P0.044 run c71ceb11-300f-4b35-b68c-2ec726b8d4f7 passed 8/8 browser checks and emitted PRACTICE_SAFE_GFM_PROJECTION_PASS with six current screenshots." -->
- [x] P0.046 proves raw HTML/event handlers are inert, remote images issue no request, unsafe URI is rejected and safe external links are hardened.
  <!-- verified: 2026-07-14 evidence="P0.046 run 04fd68b6-1ed7-49d7-b271-a37c236ea541 emitted PRACTICE_MARKDOWN_SECURITY_PASS and passed desktop/mobile hostile-Markdown browser checks." -->
- [x] P0.046 retries the exact original raw text with the same `clientMessageId`, preserves the next draft and never derives payload from rendered/normalized Markdown.
  <!-- verified: 2026-07-14 evidence="P0.046 emitted PRACTICE_RAW_RETRY_PASS and passed exact raw text/clientMessageId browser checks at desktop/mobile." -->
- [x] P0.046 proves terminal recovery has one exact `/workspace?targetJobId` detail action and no row retry, query-free workspace, `planId`, technical copy or current-scope `parse(targetJobId)` path.
  <!-- verified: 2026-07-14 evidence="P0.046 emitted PRACTICE_TERMINAL_PLAN_RECOVERY_PASS route=workspace target_job_id_only=true query_free=false parse=false plan_id=false row_retry=false and passed desktop/mobile terminal parity." -->
- [x] Current serial P0.044 then P0.046 pass using refreshed tracked-source and screenshot hashes；no sibling scenario is added.
  <!-- verified: 2026-07-14 evidence="Serial runs c71ceb11-300f-4b35-b68c-2ec726b8d4f7 / 04fd68b6-1ed7-49d7-b271-a37c236ea541 passed on shared source SHA 85f46c92fd79aa61d6894b86d4b80bc8ac58d1acda9efac9cb8cc1419f761fdf with 12 current screenshot hashes and P0.046 database residual=0." -->
