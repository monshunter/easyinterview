# 002 — Conversation Message Loop and Completion Checklist

> **版本**: 2.9
> **状态**: active
> **更新日期**: 2026-07-14

**关联计划**: [plan](./plan.md)

## Phase 1: Message reservation
- [x] 1.1 RED-GREEN: add message domain/store reservation/replay tests and implementation.
- [x] 1.2 RED-GREEN: enforce client/reply uniqueness and concurrent-new-message conflict.

## Phase 2: Assistant reply
- [x] 2.1 RED: service/API tests require ordinary message pair and no AssistantAction/turn fields.
- [x] 2.2 GREEN: implement chat context, AI call, assistant persistence and replay.
- [x] 2.3 BDD-Gate: P0.044 happy conversation passes.

## Phase 3: Failure and repair
- [x] 3.1 RED-GREEN: timeout/config/provider/schema/language matrix keeps user message retryable and writes no invalid reply.
- [x] 3.2 RED-GREEN: same ID retry and request mismatch behavior is deterministic.
- [x] 3.3 BDD-Gate: P0.046 failure/recovery passes.

## Phase 4: Completion
- [x] 4.1 RED-GREEN: completion creates one conversation-level report/job/outbox with no turn focus.
- [x] 4.2 BDD-Gate: P0.047 completion/generating handoff passes.

## Phase 5: Privacy and closeout
- [x] 5.1 RED-GREEN: ownership/privacy/race/redaction tests pass.
- [x] 5.2 Run focused/full backend, codegen/fixture/migration/prompt/docs/diff gates.

## Phase 6: Review remediation
- [x] 6.1 RED-GREEN: assistant commit locks/checks mutable session state, rolls back after completion wins, and maps the conflict without reopening the session. (`go test ./backend/internal/practice -count=1`; `go test ./backend/internal/store/practice -count=1`)
- [x] 6.2 RED-GREEN: P0.046/P0.047 trigger and verifier require provider-failure, replay, mismatch, pending-retry and late-reply lifecycle PASS markers. (`python3 -m pytest scripts/lint/scenario_script_contract_test.py -q -k practice_failure_and_completion`; focused Go tests; `bash -n`)
- [x] 6.3 BDD-Gate: P0.046 and P0.047 failure/recovery and completion scenarios pass. (serial `setup.sh` → `trigger.sh` → `verify.sh` → `cleanup.sh`, both PASS)

## Phase 7: Complete resume grounding for follow-up messages
- [x] 7.1 RED: send store/service tests require the same snapshot/original/profile precedence, complete long-input tail marker, and zero AI/assistant reply for empty context.<!-- verified: 2026-07-12 method=go-test-red tests=TestSQLRepositoryReservePracticeMessageRetriesPendingUserMessage,TestSendPracticeMessageFailsClosedWithoutResumeContextAndSkipsAI -->
- [x] 7.2 GREEN: message reservation returns shared `ResumeContext` with no slicing and common generation fails closed before prompt resolve/AI when empty.<!-- verified: 2026-07-12 method=go-test packages=internal/practice,internal/store/practice -->
- [x] 7.3 E2E.P0.044/P0.046 trigger/verify require named tail-marker and empty-context tests, preserving retry/replay evidence and rejecting skip/no-op.<!-- verified: 2026-07-12 method=scenario both=PASS -->
- [x] 7.4 BDD-Gate: P0.044/P0.046 pass with complete follow-up grounding and empty-context recovery evidence.<!-- verified: 2026-07-12 method=scenario bddChecklist=complete -->
- [x] 7.5 RED-GREEN: follow-up payload tests plus prompt lint/eval prove system-role policy, JSON escaping of untrusted JD/resume/round/persona/history and persona-style-only behavior.<!-- verified: 2026-07-12 method=go+pytest test=TestSendPracticeMessageUsesOrdinaryConversationHistory evidence="full tail marker, system/user roles, closing-tag JSON escape" -->

## Phase 8: Completion ledger as round-progress fact

- [x] 8.1 RED-GREEN: completion commits exactly one `session_completed` event with `completed_at`, report/job/outbox and returns exact replay without duplicate lifecycle facts.<!-- verified: 2026-07-12 method=unit+P0.047 test=TestSQLRepositoryCompleteSessionReplayDoesNotAppendSecondCompletedFact -->
- [x] 8.2 RED-GREEN: only completion facts whose plan resume equals `target_jobs.resume_id` are admitted; duplicate completed sessions/events for one round and report status changes project one completed round and never mutate a TargetJob progress column.<!-- verified: 2026-07-12 method=P0.098 real-postgres markers="wrong-resume-completion-ignored,target-report-status-independent,out-of-order-gap-hidden" -->
- [x] 8.3 BDD-Gate: P0.047 executes completion/replay event evidence; P0.098 proves first-to-next and final-round projection after real completion.<!-- verified: 2026-07-12 method=scenario-run both=PASS -->
- [x] 8.4 Run focused/full backend, migration/OpenAPI, privacy, context/docs/index/diff and no-frontend-business-persistence gates.<!-- verified: 2026-07-12 evidence="completion replay+bound-resume integration; P0.098; storage negative search; make test; migration/OpenAPI/context/docs/index/diff" -->

## Phase 9: Reportable completion and frozen context

- [x] 9.1 RED-GREEN: sole-owner test `TestE2EP0047RejectsZeroAnswerCompletion` proves zero committed user messages or pending assistant reply returns VALIDATION_FAILED, keeps session mutable and writes no completion/report/job/outbox/idempotency success; one committed user message succeeds. (`cd backend && go test ./internal/api/practice ./internal/practice ./internal/store/practice -run '^TestE2EP0047RejectsZeroAnswerCompletion$' -count=1 -v`; real PostgreSQL)
  <!-- verified: 2026-07-12 method=red-green+postgres evidence="RED undefined reportability contract; GREEN exact test PASS in api/practice/store packages, adjacent package suites PASS, integration TestIntegrationE2EP0047RejectsZeroAnswerCompletion PASS against version=17 Postgres with zero invalid side effects and one-answer success; marker ZERO_ANSWER_COMPLETION_REJECTED_PASS" -->
- [x] 9.2 RED-GREEN: sole-owner tests `TestE2EP0047FreezesReportContext` and `TestE2EP0047CompletionReplayPreservesReportContext` prove successful completion atomically writes full report-context.v1 + terminal coordinate from one consistent DB view; concurrent target/resume mutation, mismatch and replay pass with no AI call. (`cd backend && go test ./internal/api/practice ./internal/practice ./internal/store/practice -run '^(TestE2EP0047FreezesReportContext|TestE2EP0047CompletionReplayPreservesReportContext)$' -count=1 -v`)
  <!-- verified: 2026-07-12 method=red-green+postgres-v18 evidence="exact API/domain/store tests emit REPORT_CONTEXT_SNAPSHOT_PASS and REPORT_CONTEXT_REPLAY_PASS; completion uses read-write repeatable-read; a tagged Postgres advisory gate proves a concurrent TargetJob mutation blocks behind completion row locks, then report-context.v1/terminal 3:3 persists from the pre-mutation view; replay stays byte-equivalent after further TargetJob/Resume edits, mismatch writes zero side effects, and a second isolated rerun passes" -->
- [x] 9.3 RED-GREEN: frontend Practice Finish is disabled before the first committed user message and exposes a localized accessible reason; backend remains authoritative.
  <!-- verified: 2026-07-12 method=focused-vitest evidence="24/24 PASS across PracticeScreen, useCompletePracticeSession, target display and zh/en coverage; TestE2EP0047RejectsZeroAnswerCompletion proves opening-only and draft input do not count, native disabled + aria-describedby exposes localized reason, pending assistant reply also disables, and no completion request is sent" -->
- [x] 9.4 BDD-Gate: P0.047 runs the three exact owner tests and writes `completion-backend-evidence.json` schema `practice-completion-evidence.v1` with `ZERO_ANSWER_COMPLETION_REJECTED_PASS`, `REPORT_CONTEXT_SNAPSHOT_PASS`, `REPORT_CONTEXT_REPLAY_PASS`; PASS requires command exit 0, exact RUN/PASS markers, no FAIL/no-test marker, zero-answer no-side-effect DB assertions and same-snapshot replay. P0.056/058 consume rather than duplicate it.
  <!-- verified: 2026-07-12 method=scenario-run evidence="E2E.P0.047 setup/trigger/verify/cleanup PASS; exact 3-package owner command and tagged v18 PostgreSQL test pass with no FAIL/no-test; verifier alone writes schema-valid redacted artifact; cleanup leaves only completion-backend-evidence.json" -->

## Phase 10: Server-recoverable message reply state

- [x] 10.1 RED: store/service/API/OpenAPI tests prove failed reservations lack durable/public recovery status and `getPracticeSession` cannot return the original replay identity; generated TS error tests prove `retryable` is dropped.
  <!-- verified: 2026-07-13 method=tdd-red evidence="Migration/store/domain/API/schema/generated-client tests failed on absent durable status, missing readback identity, generic PracticeMessage and dropped typed error metadata." -->
- [x] 10.2 GREEN: baseline migration and store add user-only `reply_status=pending|retryable_failed|terminal_failed|complete`; reserve/fail/commit transitions are atomic, user-scoped and preserve unique user/reply rows.
  <!-- verified: 2026-07-13 method=unit+real-postgres evidence="User-only four-state persistence, detached bounded failure finalization, retryable-only CAS, transactional assistant+complete commit, replay, isolation and uniqueness all PASS." -->
- [x] 10.3 GREEN: generated `PracticeMessage` exposes user `clientMessageId/replyStatus`; `getPracticeSession` fixtures cover pending/retryable/terminal/complete and assistant messages omit recovery fields.
  <!-- verified: 2026-07-13 method=openapi+api+fixture evidence="Generated role union and API projection compile; canonical four-state fixtures validate; assistant JSON omits clientMessageId/replyStatus." -->
- [x] 10.4 GREEN: generated TS `ApiClientError` preserves HTTP status plus parsed `ApiErrorResponse`; JSON/non-JSON/empty/Abort/transport tests pass and no consumer parses `Error.message`.
  <!-- verified: 2026-07-13 method=generated-client+frontend-consumer evidence="Valid JSON preserves status/envelope; non-JSON, empty, Abort and transport keep apiError=null and never expose raw bodies. Practice send failure classification uses typed kind/retryable metadata and localized copy, with no Error.message parsing or technical-text leak." -->
- [x] 10.5 BDD-Gate: P0.046 proves AI failure → reload/readback → same-ID retry → one assistant reply, plus pending/terminal/cross-user/privacy gates; P0.044 remains the immediate-send pending/success owner.
  <!-- verified: 2026-07-13 method=scenario-run evidence="E2E.P0.044 and P0.046 serial setup/trigger/verify/cleanup PASS. P0.046 migrated a unique temporary PostgreSQL database, proved retry convergence, terminal readback, same-ID uniqueness, cross-user hiding and privacy cascade, then dropped the database with zero residuals." -->
- [x] 10.6 HISTORICAL-SUPERSEDED: the previous final aggregate gate is not current completion evidence after Phase 11 reopened the owner; current focused/full validation and lifecycle restoration are owned only by 11.8.
  <!-- superseded: 2026-07-14 decision="User approved Scheme A; do not attempt the pre-Phase-11 aggregate against the evolved OpenAPI/database contract or treat historical PASS as current evidence." current-owner="11.8" -->

## Phase 11: Lease-bounded generation fencing and lazy convergence

- [x] 11.1 RED: migration SQL-contract plus store/domain tests define user-only `reply_generation/reply_lease_expires_at`, the exact state-transition table and 90-second boundary before implementation.
  <!-- verified: 2026-07-14 method=go-test-red evidence="Migration contract first failed on missing reply_generation bigint; store/domain focused build then failed on missing PracticeReplyLeaseDuration and reservation ReplyGeneration before any GREEN implementation." -->
- [x] 11.2 GREEN: baseline migration and reservation model persist `pending(Gn, serverNow+90s)`；same-ID retry increments generation；Fail/Commit clear lease；assistant/public API never expose either internal field.
  <!-- verified: 2026-07-14 method=go-unit+isolated-postgres evidence="Baseline joint constraints require positive non-null generation and a pending-only lease; reservations create G1/serverNow+90s, retries own G+1, and Fail/Commit clear the lease without public projection." -->
- [x] 11.3 RED-GREEN: `getPracticeSession` and same-ID reserve each lazily converge an expired pending row under the authorized session lock；GET returns `retryable_failed(Gn)` while reserve atomically owns `pending(Gn+1,newLease)`；unexpired/different-ID/terminal/complete/mismatch paths fail closed.
  <!-- verified: 2026-07-14 method=injected-clock+store-tests evidence="Focused tests first required the service clock, then passed for the pre-boundary, exact <= boundary, GET expiry and same-ID expired G1-to-G2 paths under the authorized session lock." -->
- [x] 11.4 RED-GREEN: reserve returns generation internally；Commit/Fail require expected generation and stale G1 after a G2 reservation returns typed conflict with zero status/reply writes.
  <!-- verified: 2026-07-14 method=store-unit+real-postgres evidence="Commit and Fail compare expected generation after authorization locking; stale G1 operations after G2 return conflict and leave status and assistant rows unchanged." -->
- [x] 11.5 RED-GREEN: the four exact independent-connection PostgreSQL concurrency tests pass: `TestIntegrationPracticeReplyConcurrentNewIDsReserveOnce`, `TestIntegrationPracticeReplyConcurrentSameIDInitialReserveOnce`, `TestIntegrationPracticeReplyConcurrentExpiredSameIDRetryAdvancesOneGeneration`, `TestIntegrationPracticeReplyStaleGenerationFencedAfterGETRecovery`.
  <!-- verified: 2026-07-14 method=isolated-postgres-concurrency evidence="All four exact tests passed with distinct pg_backend_pid values and a shared start barrier; the temporary database was force-dropped with residual count zero." -->
- [x] 11.6 CONTRACT-GATE: service/API/OpenAPI/codegen/fixture tests prove `PracticeMessage` remains `clientMessageId + replyStatus` only and generation/lease never leak into response, URL, logs or frontend state.
  <!-- verified: 2026-07-14 method=service-api-contract evidence="Practice service/API/store/migration packages pass; raw response and scoped production scans expose no generation or lease fields outside backend persistence internals, while current OpenAPI/codegen/fixtures retain only clientMessageId and replyStatus." -->
- [x] 11.7 BDD-Gate: P0.044/P0.046 run against current code and isolated migrated PostgreSQL；required lease/fence/concurrency/95-second-reconcile/terminal-plan/fingerprint markers pass, every screenshot has SHA-256 + dimensions + viewport, and verifier rejects any source-fingerprint drift or historical artifact.
  <!-- reverified: 2026-07-14 method=serial-scenario-run evidence="Current-source P0.044 run cd8c378d-6fcc-4045-a00f-c9129873e511 and P0.046 run dfc68a8f-41de-46e5-9b0c-1aec3fbb67fb passed against source SHA 3e644ae013ee2159937e4853c8f3f32a3f2bd1f1351fd6fe74d9b66aa2ea11d2; screenshots, exact lease/fence/concurrency/recovery markers, cleanup and isolated database residual=0 all passed." -->
- [x] 11.8 Run focused/full backend, migration, OpenAPI/codegen/fixture, scenario contract, context/docs/diff gates；only then record current evidence and restore completed lifecycle.
  <!-- reverified: 2026-07-14 decision="User approved Scheme A" method=current-full-aggregate evidence="Root make test passed UI 62/62, Python 590 tests/5181 subtests, all Go packages and frontend 121 files/977 tests after exact event/table sets and shared-verifier ownership were repaired; current P0.044/P0.046, context, docs, diff and pruning gates pass, while 10.6 remains historical-superseded." -->

## Phase 12: Configured message and session text limits

- [x] 12.1 RED: UTF-8 32KiB/32KiB+1 单条与 256KiB/256KiB+1 会话累计 fixture 暴露旧 8,000-rune/无总量上限行为。
- [x] 12.2 GREEN: service 注入 `practice.maxMessageBytes=32768` / `practice.maxSessionTextBytes=262144`；按 bytes 在 reservation/provider 前校验。
- [x] 12.3 STORE/CONCURRENCY: persisted aggregate + candidate 在一致性边界内裁决；越界零 user/assistant/provider side effect，并发不得联合绕过。
- [x] 12.4 CONTRACT/BDD: RuntimeConfig 两字段、前端 precheck 与 P0.046 limit/limit+1/multibyte/reload/same-ID gates 通过。
- [x] 12.5 VERIFY: focused/full practice/store/API/race、OpenAPI scenario、privacy、contexts/docs/diff 与旧 8,000-rune production-truth negative search 通过；post-commit codegen drift 由 A4 13.8 收口。
  <!-- verified: 2026-07-14 evidence="Service and SQLRepository exact/+1 user/assistant aggregate tests pass; P0.046 fresh current run passes real concurrency, browser, fingerprint and residual=0 gates." -->

## 修订记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-07-14 | 2.9 | Reopen with unchecked Phase 12 for configured Practice message/session byte boundaries. |
| 2026-07-14 | 2.8 | Reopen with unchecked Phase 11 for lease expiry, generation fencing, real concurrency and freshness-bound scenario evidence. |
| 2026-07-13 | 2.7 | Reopen for durable reply status and refresh-safe same-ID recovery. |
| 2026-07-12 | 2.6 | 锁定 002 completion 唯一 owner、精确 P0.047 tests/markers/artifact。 |
| 2026-07-12 | 2.5 | 要求至少一条 candidate user message 后才能 completion，并原子冻结 report-context.v1。 |
| 2026-07-12 | 2.4 | 完成事实限定 TargetJob 绑定 resume，并增加 system policy / JSON 不可信 follow-up 上下文分层 gate。 |
