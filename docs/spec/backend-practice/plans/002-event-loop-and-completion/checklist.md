# 002 — Conversation Message Loop and Completion Checklist

> **版本**: 2.10
> **状态**: completed
> **更新日期**: 2026-07-14

**关联计划**: [plan](./plan.md)

## Phase 1: Message reservation
- [x] 1.1 RED-GREEN: add message domain/store reservation/replay tests and implementation.
- [x] 1.2 RED-GREEN: enforce client/reply uniqueness and concurrent-new-message conflict.

## Phase 2: Assistant reply
- [x] 2.1 RED: service/API tests require ordinary message pair and no AssistantAction/turn fields.
- [x] 2.2 GREEN: implement chat context, AI call, assistant persistence and replay.

## Phase 3: Failure and repair
- [x] 3.1 RED-GREEN: timeout/config/provider/schema/language matrix keeps user message retryable and writes no invalid reply.
- [x] 3.2 RED-GREEN: same ID retry and request mismatch behavior is deterministic.

## Phase 4: Completion
- [x] 4.1 RED-GREEN: completion creates one conversation-level report/job/outbox with no turn focus.

## Phase 5: Privacy and closeout
- [x] 5.1 RED-GREEN: ownership/privacy/race/redaction tests pass.
- [x] 5.2 仓库根 `make test` 完成前后端全量单测回归；codegen/fixture/migration/prompt/docs/diff 作为独立 gates。

## Phase 6: Review remediation
- [x] 6.1 RED-GREEN: assistant commit locks/checks mutable session state, rolls back after completion wins, and maps the conflict without reopening the session. (`go test ./backend/internal/practice -count=1`; `go test ./backend/internal/store/practice -count=1`)

## Phase 7: Complete resume grounding for follow-up messages
- [x] 7.1 RED: send store/service tests require the same snapshot/original/profile precedence, complete long-input tail marker, and zero AI/assistant reply for empty context.<!-- verified: 2026-07-12 method=go-test-red tests=TestSQLRepositoryReservePracticeMessageRetriesPendingUserMessage,TestSendPracticeMessageFailsClosedWithoutResumeContextAndSkipsAI -->
- [x] 7.2 GREEN: message reservation returns shared `ResumeContext` with no slicing and common generation fails closed before prompt resolve/AI when empty.<!-- verified: 2026-07-12 method=go-test packages=internal/practice,internal/store/practice -->
- [x] 7.5 RED-GREEN: follow-up payload tests plus prompt lint/eval prove system-role policy, JSON escaping of untrusted JD/resume/round/persona/history and persona-style-only behavior.<!-- verified: 2026-07-12 method=go+pytest test=TestSendPracticeMessageUsesOrdinaryConversationHistory evidence="full tail marker, system/user roles, closing-tag JSON escape" -->

## Phase 8: Completion ledger as round-progress fact

- [x] 8.2 RED-GREEN: only completion facts whose plan resume equals `target_jobs.resume_id` are admitted; duplicate completed sessions/events for one round and report status changes project one completed round and never mutate a TargetJob progress column；focused DB tests only provide development feedback.
- [x] 8.3 BDD-Gate: `BDD.PRACTICE.EVENT_LOOP.001` 由 [BDD checklist](./bdd-checklist.md) 关联 send/retry/completion owner behavior tests。
- [x] 8.3a E2E-HANDOFF: `E2E.P0.098` 仅承接真实 completion/progress refresh；本轮未运行，current-run 状态仍为 `Ready`。
- [x] 8.4 Repository-root `make test` provides frontend/backend unit regression；migration/OpenAPI, privacy, context/docs/index/diff and browser-persistence negative searches remain separate gates.

## Phase 9: Reportable completion and frozen context

- [x] 9.1 RED-GREEN: focused completion code tests prove zero committed user messages or a pending assistant reply returns `VALIDATION_FAILED`, keeps the session mutable and writes no completion/report/job/outbox/idempotency success；one committed user message succeeds.
- [x] 9.2 RED-GREEN: focused snapshot/replay code tests prove successful completion atomically writes full report-context.v1 and terminal coordinate from one consistent DB view；concurrent mutation, mismatch and replay cause no AI call or duplicate side effect.
- [x] 9.3 RED-GREEN: frontend Practice Finish is disabled before the first committed user message and exposes a localized accessible reason; backend remains authoritative.
- [x] 9.4 Phase completion is reported by repository-root `make test`；focused frontend/backend tests are development feedback and PostgreSQL consistency remains a separate integration gate.

## Phase 10: Server-recoverable message reply state

- [x] 10.1 RED: store/service/API/OpenAPI tests prove failed reservations lack durable/public recovery status and `getPracticeSession` cannot return the original replay identity; generated TS error tests prove `retryable` is dropped.
  <!-- verified: 2026-07-13 method=tdd-red evidence="Migration/store/domain/API/schema/generated-client tests failed on absent durable status, missing readback identity, generic PracticeMessage and dropped typed error metadata." -->
- [x] 10.2 GREEN: baseline migration and store add user-only `reply_status=pending|retryable_failed|terminal_failed|complete`; reserve/fail/commit transitions are atomic, user-scoped and preserve unique user/reply rows.
  <!-- verified: 2026-07-13 method=unit+real-postgres evidence="User-only four-state persistence, detached bounded failure finalization, retryable-only CAS, transactional assistant+complete commit, replay, isolation and uniqueness all PASS." -->
- [x] 10.3 GREEN: generated `PracticeMessage` exposes user `clientMessageId/replyStatus`; `getPracticeSession` fixtures cover pending/retryable/terminal/complete and assistant messages omit recovery fields.
  <!-- verified: 2026-07-13 method=openapi+api+fixture evidence="Generated role union and API projection compile; canonical four-state fixtures validate; assistant JSON omits clientMessageId/replyStatus." -->
- [x] 10.4 GREEN: generated TS `ApiClientError` preserves HTTP status plus parsed `ApiErrorResponse`; JSON/non-JSON/empty/Abort/transport tests pass and no consumer parses `Error.message`.
  <!-- verified: 2026-07-13 method=generated-client+frontend-consumer evidence="Valid JSON preserves status/envelope; non-JSON, empty, Abort and transport keep apiError=null and never expose raw bodies. Practice send failure classification uses typed kind/retryable metadata and localized copy, with no Error.message parsing or technical-text leak." -->
- [x] 10.6 HISTORICAL-SUPERSEDED: the previous final aggregate gate is not current completion evidence after Phase 11 reopened the owner; current root unit regression and lifecycle restoration are owned only by 11.8.
  <!-- superseded: 2026-07-14 decision="User approved Scheme A; do not attempt the pre-Phase-11 aggregate against the evolved OpenAPI/database contract or treat historical PASS as current evidence." current-owner="11.8" -->

## Phase 11: Lease-bounded generation fencing and lazy convergence

- [x] 11.1 RED: migration SQL-contract plus store/domain tests define user-only `reply_generation/reply_lease_expires_at`, the exact state-transition table and 90-second boundary before implementation.
- [x] 11.2 GREEN: baseline migration and reservation model persist `pending(Gn, serverNow+90s)`；same-ID retry increments generation；Fail/Commit clear lease；assistant/public API never expose either internal field.
  <!-- verified: 2026-07-14 method=go-unit+isolated-postgres evidence="Baseline joint constraints require positive non-null generation and a pending-only lease; reservations create G1/serverNow+90s, retries own G+1, and Fail/Commit clear the lease without public projection." -->
- [x] 11.3 RED-GREEN: `getPracticeSession` and same-ID reserve each lazily converge an expired pending row under the authorized session lock；GET returns `retryable_failed(Gn)` while reserve atomically owns `pending(Gn+1,newLease)`；unexpired/different-ID/terminal/complete/mismatch paths fail closed.
- [x] 11.4 RED-GREEN: reserve returns generation internally；Commit/Fail require expected generation and stale G1 after a G2 reservation returns typed conflict with zero status/reply writes.
  <!-- verified: 2026-07-14 method=store-unit+real-postgres evidence="Commit and Fail compare expected generation after authorization locking; stale G1 operations after G2 return conflict and leave status and assistant rows unchanged." -->
- [x] 11.5 RED-GREEN: the four exact independent-connection PostgreSQL concurrency tests pass: `TestIntegrationPracticeReplyConcurrentNewIDsReserveOnce`, `TestIntegrationPracticeReplyConcurrentSameIDInitialReserveOnce`, `TestIntegrationPracticeReplyConcurrentExpiredSameIDRetryAdvancesOneGeneration`, `TestIntegrationPracticeReplyStaleGenerationFencedAfterGETRecovery`.
  <!-- verified: 2026-07-14 method=isolated-postgres-concurrency evidence="All four exact tests passed with distinct pg_backend_pid values and a shared start barrier; the temporary database was force-dropped with residual count zero." -->
- [x] 11.6 CONTRACT-GATE: service/API/OpenAPI/codegen/fixture tests prove `PracticeMessage` remains `clientMessageId + replyStatus` only and generation/lease never leak into response, URL, logs or frontend state.
  <!-- verified: 2026-07-14 method=service-api-contract evidence="Practice service/API/store/migration packages pass; raw response and scoped production scans expose no generation or lease fields outside backend persistence internals, while current OpenAPI/codegen/fixtures retain only clientMessageId and replyStatus." -->
- [x] 11.8 仓库根 `make test` 完成前后端全量单测回归；migration、OpenAPI/codegen/fixture、integration contract 与 context/docs/diff 作为独立 gates；随后记录当前证据并恢复 completed lifecycle。

## Phase 12: Injected message and session guards

- [x] 12.1 OWNER-GATE: missing/default/override/invalid 与跨字段约束只由 A4 typed contract 覆盖；本 owner 删除重复 config wiring tests。
- [x] 12.2 FOCUSED-GATE: service 注入小型 message/session limits；ASCII/多字节 overflow 在 reservation/provider 前校验，不构造默认大小字符串。
- [x] 12.3 STORE/CONCURRENCY: persisted aggregate + candidate 在一致性边界内裁决；越界零 user/assistant/provider side effect，并发不得联合绕过。
  <!-- verified: 2026-07-14 evidence="Small-limit service/store and real concurrency assertions remain; configuration-only scenario assertions are removed." -->

## 修订记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-07-14 | 2.9 | Revise Phase 12 to owner-only typed config plus small focused business guards; no configuration scenario gate. |
| 2026-07-14 | 2.8 | Reopen with unchecked Phase 11 for lease expiry, generation fencing, real concurrency and freshness-bound scenario evidence. |
| 2026-07-13 | 2.7 | Reopen for durable reply status and refresh-safe same-ID recovery. |
| 2026-07-12 | 2.5 | 要求至少一条 candidate user message 后才能 completion，并原子冻结 report-context.v1。 |
| 2026-07-12 | 2.4 | 完成事实限定 TargetJob 绑定 resume，并增加 system policy / JSON 不可信 follow-up 上下文分层 gate。 |
