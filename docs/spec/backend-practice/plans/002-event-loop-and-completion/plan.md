# 002 — Conversation Message Loop and Completion

> **版本**: 2.10
> **状态**: completed
> **更新日期**: 2026-07-14

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)
**关联 Test Plan**: [test-plan](./test-plan.md)
**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## 1 目标

实现 `sendPracticeMessage` 连续聊天，并作为 `completePracticeSession` 的唯一 owner 负责可报告性校验、完成事务与 `report-context.v1` 原子快照：普通 user/assistant messages 按序持久化，不区分题目、回答或追问；失败可用同一 `clientMessageId` 恢复且不重复消息。Phase 10 将该恢复能力扩展到刷新/重挂载：后端持久化 reply status，`getPracticeSession` 返回原 user `clientMessageId/replyStatus`，前端不依赖浏览器存储。Phase 11 按已确认 T-B/P-A 为 pending reservation 增加 90 秒 lease、内部 generation fence 与 GET / 同 ID reserve 惰性收敛，使进程中断和迟到 worker 也能确定性恢复。

## 2 Operation Matrix

| operationId | fixture | frontend consumer | handler | persistence | AI | scenario |
|-------------|---------|-------------------|---------|-------------|----|----------|
| `getPracticeSession` | current session/reply-state fixtures | Practice loader/remount recovery | practice read owner | session + messages/reply state/lease | none | 当前无真实 E2E owner；root `make test` |
| `sendPracticeMessage` | current send/retry fixtures | Practice send/retry | practice send owner | messages/reply state/lease/task runs | `practice.session.chat` | 当前无真实 E2E owner；root `make test` |
| `completePracticeSession` | current completion fixtures | Practice Finish | practice completion owner | session/report-context/job/outbox/idempotency | no AI in transaction | `E2E.P0.098` 仅真实 completion API 与 progress refresh；root `make test` for other paths |

## 3 质量门禁分类

- **Plan 类型**: feature-behavior + API + backend + persistence。
- **TDD 策略**: Red tests cover message replay, concurrency, failure recovery, ordering, language, privacy and completion before implementation；focused tests only provide development feedback，阶段完成由根 `make test` 承接。
- **BDD 策略**: `BDD.PRACTICE.EVENT_LOOP.001` 由代码层 owner tests 验证 send/retry/completion 行为，并由仓库根 `make test` 统一回归；`E2E.P0.098` 仅作为真实登录、completion API、Home/Workspace/TargetJob refresh/detail read 的独立 handoff，只有显式真实运行后才产生 PASS。chat/session start/plan creation 当前无真实 E2E owner。
- **替代验证 gate**: fixture/codegen, store SQL, PostgreSQL integration, race, privacy lint.

## 4 Coverage Matrix

| Behavior | Category | Phase | Verification | Negative |
|----------|----------|-------|--------------|----------|
| replay | boundary | 2 | same ID/reply uniqueness tests | duplicate user/assistant rows |
| concurrent message | conflict | 2 | pending reply conflict tests | two in-flight user messages |
| language/schema | contract | 3 | one-repair tests | wrong-language persisted reply |
| privacy | security | 5 | redaction/outbox/task tests | raw message outside content store |
| frozen report context | cross-owner handoff | 9 | one-view DB tests + owner artifact | review rebuilds from mutable entities |
| typed error boundary | contract/recovery | 10 | generated `ApiClientError` JSON/non-JSON/empty/Abort/transport matrix | parsing `error.message`; retrying terminal errors |
| stale worker generation fence | concurrency/idempotency | 11 | four real PostgreSQL concurrent tests + service propagation | G1 commits/fails G2; duplicate assistant; generation exposed publicly |

## 5 实施步骤

### Phase 1: Message domain and reservation
- Define user/assistant message records and `clientMessageId` / `replyToMessageId` uniqueness.
- Reserve or replay user message in a short transaction; reject concurrent new message while reply is missing.

### Phase 2: Assistant reply
- Load canonical plan/session context and ordered messages.
- Execute chat outside transaction, persist one assistant reply, return pair + current session.
- Replay returns stored pair without AI call.

### Phase 3: Failure and repair
- One schema/language repair; provider/config/timeout do not business-repair.
- Failed user message remains retryable; same ID retries generation and never duplicates.

### Phase 4: Completion
- Complete running session idempotently and enqueue conversation-level report.
- No turn assessment/job or question focus data.

### Phase 5: Privacy and closeout
- Redaction/ownership/race/full gates and BDD scenarios.

### Phase 6: Review remediation
- Reject assistant commits when completion has already moved the session out of a mutable state, and map the store conflict through the service/API boundary.

### Phase 7: Complete resume grounding for follow-up messages

- RED store/service tests prove `sendPracticeMessage` loads the same complete resume source precedence as session start and preserves a long-input tail marker in every AI payload.
- GREEN message reservation exposes the shared `ResumeContext` without character/token slicing; the common generator returns typed `VALIDATION_FAILED` before prompt resolve/AI when context is empty and never writes an assistant reply.
- Keep immutable interviewer policy in the system role and JSON-encode JD, complete resume, persisted round, persona and ordered conversation history as untrusted user data. Embedded instruction-like text cannot escape into policy; persona only controls tone/perspective and cannot invent facts or replace the persisted round.

### Phase 8: Completion ledger as round-progress fact

- RED completion store/service tests require one durable `session_completed` fact in the same transaction as `completed_at`, report/job/outbox creation, and exact idempotent replay. Only sessions whose plan resume equals `target_jobs.resume_id` may contribute to that TargetJob's completed-round ledger.
- GREEN preserves the existing event as the sole completed-round ledger input; duplicate completion requests, duplicate sessions for one round and report retries do not create duplicate progress entries.
- Progress becomes visible immediately after completion commit, independent of report queued/generating/ready/failed state; no frontend/local storage write is part of completion.

### Phase 9: Reportable completion and frozen report context

`backend-practice/002` 是本阶段唯一 completion owner；`backend-review/001` 不得复制 completion 查询、事务或零回答判断，只能消费冻结快照与 owner evidence。

- Completion requires at least one committed candidate `user` message and no pending assistant reply. A zero-answer session returns typed `VALIDATION_FAILED`, remains running, and creates no completion fact/report/job/outbox/idempotency success record.
- In the same successful completion transaction, freeze current `report-context.v1` from TargetJob raw/structured data, bound Resume source/profile, canonical round ladder/current round, source Plan settings, session language and terminal message count/last sequence. No AI call occurs in this transaction.
- Focused API/service/store tests may be used for development feedback. Phase completion is reported by repository-root `make test`; real database transaction checks remain a separate integration gate.

### Phase 10: Server-recoverable message reply state

- RED store/service/API tests first prove that a failed user reservation survives without a public `clientMessageId/replyStatus`, `getPracticeSession` cannot reconstruct the retry, and a reload followed by a new ID conflicts. OpenAPI/generated/frontend RED must also prove structured `ApiErrorResponse.error.retryable` is currently lost by the TS client.
- GREEN extends pre-release `practice_messages` with user-only `reply_status` (`pending / retryable_failed / terminal_failed / complete`). Reserve inserts or re-enters `pending`; AI/provider/contract failure atomically records failed status before returning; assistant commit writes the unique reply and `complete` in one transaction. A different ID remains blocked while an unresolved user row exists.
- `PracticeMessage` read projection returns user `clientMessageId/replyStatus` and omits both from assistant messages. Refresh/remount consumes `getPracticeSession`: `pending` restores thinking, `retryable_failed` restores a row-local same-ID retry, `terminal_failed` restores terminal recovery, and `complete` has exactly one following reply. No URL/browser storage/fixture state may restore business status.
- The OpenAPI-generated TS runtime must expose typed `ApiClientError` with public `status` and `apiError` fields；JSON `ApiErrorResponse`、non-JSON、empty response、Abort and transport failures have explicit tests. Frontend must never parse `Error.message` to infer retryability.

### Phase 11: Lease-bounded generation fencing and lazy convergence

- **RED 11.1 — migration/state table**: first add failing migration SQL-contract and store/domain tests for user-only `reply_generation` plus `reply_lease_expires_at`. Generation starts at 1 and survives terminal/complete history；only `pending` has a non-null lease；assistant rows expose neither field. These internal fields must not enter OpenAPI or generated clients.
- **GREEN 11.2 — 90-second lease**: reserve writes `pending(G1, serverNow+90s)`；a same-ID retry from `retryable_failed(Gn)` writes `pending(Gn+1, serverNow+90s)`；Fail/Commit clear lease while retaining generation。`Service.now` is the sole authoritative clock so unit and PostgreSQL tests can deterministically cross the boundary.
- **RED/GREEN 11.3 — lazy convergence**: `getPracticeSession` locks the authorized session, converts every expired `pending(Gn)` to `retryable_failed(Gn)` and reads the converged session in the same transaction. A same-ID reserve may instead atomically take over an expired row as `pending(Gn+1,newLease)`；an unexpired pending, terminal/complete row, payload mismatch or different new ID remains a typed conflict. No cron/background worker is required.
- **RED/GREEN 11.4 — generation fence**: reserve returns the internal generation to service；`CommitPracticeMessage` and `FailPracticeMessage` require expected generation and may mutate only `pending + generation match`. Stale G1 Commit/Fail after G2 reserve returns typed conflict with zero state/assistant writes；valid G2 can commit exactly one assistant reply.
- **Real PostgreSQL 11.5**: use independent DB connections plus a start barrier, not sequential calls, for exactly `TestIntegrationPracticeReplyConcurrentNewIDsReserveOnce`, `TestIntegrationPracticeReplyConcurrentSameIDInitialReserveOnce`, `TestIntegrationPracticeReplyConcurrentExpiredSameIDRetryAdvancesOneGeneration`, and `TestIntegrationPracticeReplyStaleGenerationFencedAfterGETRecovery`. The fourth test pauses G1, expires it through GET, reserves G2, releases stale G1 Commit and Fail, then proves only G2 can write one assistant reply.

### Phase 12: Injected message and session guards

- **OWNER 12.1**: missing/default/override/invalid 与跨字段配置只由 A4 typed owner 覆盖；本 owner 不复制默认数值、loader/composition 或 RuntimeConfig 传播测试。
- **FOCUSED 12.2**: inject small message/session byte limits into `SendPracticeMessage`. Count bytes, not runes; evaluate before reservation and provider call; replay of an already accepted same ID remains idempotent。Fixtures 保持小型并覆盖 ASCII/多字节，不构造默认大小字符串。
- **Persistence 12.3**: session total is computed from authorized persisted `practice_messages` plus the candidate message in the same consistency boundary. Over-limit returns typed `VALIDATION_FAILED` with zero user/assistant/provider side effects; concurrency cannot let two messages jointly bypass the total.

## 6 验收标准

- Multiple message pairs append in stable order with no question classification.
- Retries/concurrency cannot duplicate messages or provider calls after replay.
- Completion creates one report job and conversation-level handoff.
- Completion commits one auditable round fact that TargetJob read models can project without a mutable progress column.
- Wrong-resume plan/session facts cannot complete or advance the TargetJob's canonical prefix, even when both resumes belong to the same user.
- Zero-answer/pending-reply completion is rejected without side effects; successful completion atomically freezes one immutable current-shape report context under this plan's sole ownership.
- No raw message leaks outside authorized content/prompt/report paths.
- Reload/remount cannot strand a persisted user message: server reply state reconstructs pending/failure UI and same-ID retry converges to one reply.
- A pending reservation expires after exactly 90 seconds of server time；GET or same-ID reserve lazily converges it, while generation fencing prevents every stale worker from mutating the newer attempt.
- Practice 默认/override/invalid 归 A4；backend-practice 以小型注入值验证 overflow 在 reservation/provider 前拒绝、会话累计原子裁决且无半成品持久化。

## 7 风险与应对

| 风险 | 应对 |
|------|------|
| failure after user persist | same clientMessageId resumable generation |
| concurrent submits | unique constraints + session-level reservation conflict |
| stale frontend expects AssistantAction | codegen/typecheck/negative search |
| start and send use different resume projections | Phase 7 SQL/service tests require identical precedence and the same tail marker on follow-up calls |
| report generation fails after completion | progress consumes the committed completion event, not report status |
| completion request/session is replayed | store idempotency plus read-side distinct round pairs prevents duplicate progress |
| same-user wrong-resume plan contributes progress | completion/read projection requires the plan resume to equal the TargetJob binding before admitting the fact |
| history or resume carries prompt-like instructions | system policy stays separate; all business context is JSON-encoded untrusted user data |
| refresh loses the only replay identity | persist user reply status and expose the original clientMessageId through the authorized session read model |
| generated client drops retryable metadata | typed ApiClientError is generated and tested; UI never parses error strings |
| server process dies after reserve | 90-second persisted lease plus GET/same-ID-reserve lazy convergence makes the row recoverable without an in-memory timer |
| stale worker returns after a retry owns the row | expected-generation CAS fences both Commit and Fail before assistant insertion/status mutation |
| prior screenshots survive a source change | shared source fingerprint and screenshot SHA-256 are verified at acceptance time; stale artifacts fail closed |

## 8 修订记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-07-14 | 2.10 | Separate code-owned BDD behavior from the independent Ready-only P0.098 real API/UI handoff. |
| 2026-07-13 | 2.7 | Add server-persisted reply status, refresh-safe same-ID recovery and typed API-error handoff. |
| 2026-07-12 | 2.5 | Require one candidate message before completion and atomically freeze report-context.v1. |
| 2026-07-12 | 2.4 | Bind completion facts to the TargetJob resume and separate immutable system policy from JSON-encoded untrusted follow-up context. |
| 2026-07-12 | 2.3 | Reopen Phase 8 so the committed session-completion event is the durable round-progress fact. |
| 2026-07-12 | 2.2 | Reopen follow-up messaging so every AI call uses the complete resume source snapshot and fails closed without evidence. |
| 2026-07-12 | 2.1 | Reopen for send/complete race protection and executable failure/recovery scenario evidence. |
| 2026-07-12 | 2.0 | Replace answer/turn event loop with message conversation loop. |
