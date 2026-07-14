# 002 — Conversation Message Loop and Completion

> **版本**: 2.9
> **状态**: active
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
| `getPracticeSession` | current `getPracticeSession.json::{default,prototype-baseline,reply-pending,reply-retryable-failed,reply-terminal-failed,reply-complete,missing-session}` | Practice session loader/remount recovery | `api/practice.GetPracticeSession` → `practice.GetPracticeSession` → SQL `GetSession`；Phase 11 在授权事务内惰性收敛过期 pending lease | current `client_message_id/reply_status` read projection；Phase 11 增加内部 `reply_generation/reply_lease_expires_at` | none | P0.044/P0.046 |
| `sendPracticeMessage` | current `sendPracticeMessage.json::{default,ai-timeout-retryable,auth-unauthorized,validation-empty-text,session-not-found,reply-pending-conflict,client-message-mismatch,retry-success-same-client-message}` | Practice message hook/row-local retry | `api/practice.SendPracticeMessage` → `practice.SendPracticeMessage` → SQL reserve/fail/commit；Phase 11 传递 expected generation | current `client_message_id/reply_status`；Phase 11 增加 90 秒 lease/generation fence；task-runs | `practice.session.chat` | P0.044/P0.046 |
| `completePracticeSession` | `openapi/fixtures/PracticeSessions/completePracticeSession.json` | Practice finish hook | `backend-practice/002` handler/service/store（唯一 completion owner） | session/terminal messages/report-context.v1/job/outbox/idempotency | transaction 内无 AI；随后 report job | P0.047 owner artifact；P0.056/058 只消费 marker |

## 3 质量门禁分类

- **Plan 类型**: feature-behavior + API + backend + persistence。
- **TDD 策略**: Red tests cover message replay, concurrency, failure recovery, ordering, language, privacy and completion before implementation；Phase 11 严格按 migration contract RED → store transition RED/GREEN → real PostgreSQL concurrency RED/GREEN → service/API regression → BDD evidence 执行。
- **BDD 策略**: P0.044/P0.046/P0.047 cover happy, failure/recovery and completion.
- **替代验证 gate**: fixture/codegen, store SQL, race, privacy lint, full backend.

## 4 Coverage Matrix

| Behavior | Category | Phase | Verification | Negative |
|----------|----------|-------|--------------|----------|
| send user/reply | primary | 1-2 | service/store/API + P0.044 | answer_submitted/AssistantAction |
| replay | boundary | 2 | same ID/reply uniqueness tests | duplicate user/assistant rows |
| concurrent message | conflict | 2 | pending reply conflict tests | two in-flight user messages |
| AI failure/retry | recovery | 3 | failure matrix + P0.046 | canned reply/session_wait action |
| language/schema | contract | 3 | one-repair tests | wrong-language persisted reply |
| complete | primary/idempotency | 4 | completion tests + P0.047 | turn count/question assessment handoff |
| privacy | security | 5 | redaction/outbox/task tests | raw message outside content store |
| send/complete race | lifecycle/boundary | 6 | service/store race regression + P0.047 | late reply reopens completing session |
| failure scenario evidence | BDD/gate | 6 | P0.046 named failure/replay/mismatch markers | happy-path-only false PASS |
| zero-answer / pending reply | failure/boundary | 9 | exact service/store/API tests + P0.047 | empty conversation creates report/job |
| frozen report context | cross-owner handoff | 9 | one-view DB tests + owner artifact | review rebuilds from mutable entities |
| refresh/reload recovery | failure/recovery/persistence | 10 | SQL/API/OpenAPI/frontend composed P0.046 | client-only retry state; missing original ID; duplicate reply |
| typed error boundary | contract/recovery | 10 | generated `ApiClientError` JSON/non-JSON/empty/Abort/transport matrix | parsing `error.message`; retrying terminal errors |
| pending lease convergence | failure/recovery/concurrency | 11 | migration/store unit + GET/same-ID reserve tests + P0.046 | immortal pending; background-job-only recovery; client clock ownership |
| stale worker generation fence | concurrency/idempotency | 11 | four real PostgreSQL concurrent tests + service propagation | G1 commits/fails G2; duplicate assistant; generation exposed publicly |
| evidence freshness | BDD/regression | 11 | shared source fingerprint manifest + P0.044/P0.046 verifier | stale screenshots/logs accepted after source change |

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
- Make P0.046 execute provider failure, exact replay, mismatch, pending retry and concurrent-new-message assertions; make P0.047 prove a late reply cannot reopen the session.

### Phase 7: Complete resume grounding for follow-up messages

- RED store/service tests prove `sendPracticeMessage` loads the same complete resume source precedence as session start and preserves a long-input tail marker in every AI payload.
- GREEN message reservation exposes the shared `ResumeContext` without character/token slicing; the common generator returns typed `VALIDATION_FAILED` before prompt resolve/AI when context is empty and never writes an assistant reply.
- Keep immutable interviewer policy in the system role and JSON-encode JD, complete resume, persisted round, persona and ordered conversation history as untrusted user data. Embedded instruction-like text cannot escape into policy; persona only controls tone/perspective and cannot invent facts or replace the persisted round.
- P0.044/P0.046 trigger/verify require named full-snapshot and no-context tests while preserving same-client-message recovery semantics.

### Phase 8: Completion ledger as round-progress fact

- RED completion store/service tests require one durable `session_completed` fact in the same transaction as `completed_at`, report/job/outbox creation, and exact idempotent replay. Only sessions whose plan resume equals `target_jobs.resume_id` may contribute to that TargetJob's completed-round ledger.
- GREEN preserves the existing event as the sole completed-round ledger input; duplicate completion requests, duplicate sessions for one round and report retries do not create duplicate progress entries.
- Progress becomes visible immediately after completion commit, independent of report queued/generating/ready/failed state; no frontend/local storage write is part of completion.
- P0.047 and the cross-layer P0.098 gate prove first-round completion advances TargetJob projection to the next canonical round and final completion yields no current round.

### Phase 9: Reportable completion and frozen report context

`backend-practice/002` 是本阶段唯一 completion owner；`backend-review/001` 不得复制 completion 查询、事务或零回答判断，只能消费冻结快照与 owner evidence。

- Completion requires at least one committed candidate `user` message and no pending assistant reply. A zero-answer session returns typed `VALIDATION_FAILED`, remains running, and creates no completion fact/report/job/outbox/idempotency success record.
- In the same successful completion transaction, freeze current `report-context.v1` from TargetJob raw/structured data, bound Resume source/profile, canonical round ladder/current round, source Plan settings, session language and terminal message count/last sequence. No AI call occurs in this transaction.
- Focused owner command is `cd backend && go test ./internal/api/practice ./internal/practice ./internal/store/practice -run '^(TestE2EP0047RejectsZeroAnswerCompletion|TestE2EP0047FreezesReportContext|TestE2EP0047CompletionReplayPreservesReportContext)$' -count=1 -v`.
- P0.047 writes `.test-output/e2e/p0-047-practice-text-loop-complete-and-generating-handoff/completion-backend-evidence.json` with exact top-level keys `schemaVersion`, `scenarioId`, `command`, `tests`, `markers`, `database`, `result`. `schemaVersion` is `practice-completion-evidence.v1`; `scenarioId` is `E2E.P0.047`; `tests` records the three exact names/statuses; `markers` contains `ZERO_ANSWER_COMPLETION_REJECTED_PASS`, `REPORT_CONTEXT_SNAPSHOT_PASS`, `REPORT_CONTEXT_REPLAY_PASS`; `database` records only redacted booleans/counts/context version; `result` is `PASS` only when the command exits 0, every exact `=== RUN`/`--- PASS:` and marker exists, no `--- FAIL:`/package `FAIL`/`no tests to run` appears, zero-answer has no side effects, and one-answer replay preserves the same snapshot.
- P0.056/P0.058 consume the schema-valid P0.047 artifact/markers later; they cannot substitute their own completion implementation or infer PASS from frontend Vitest.

### Phase 10: Server-recoverable message reply state

- RED store/service/API tests first prove that a failed user reservation survives without a public `clientMessageId/replyStatus`, `getPracticeSession` cannot reconstruct the retry, and a reload followed by a new ID conflicts. OpenAPI/generated/frontend RED must also prove structured `ApiErrorResponse.error.retryable` is currently lost by the TS client.
- GREEN extends pre-release `practice_messages` with user-only `reply_status` (`pending / retryable_failed / terminal_failed / complete`). Reserve inserts or re-enters `pending`; AI/provider/contract failure atomically records failed status before returning; assistant commit writes the unique reply and `complete` in one transaction. A different ID remains blocked while an unresolved user row exists.
- `PracticeMessage` read projection returns user `clientMessageId/replyStatus` and omits both from assistant messages. Refresh/remount consumes `getPracticeSession`: `pending` restores thinking, `retryable_failed` restores a row-local same-ID retry, `terminal_failed` restores terminal recovery, and `complete` has exactly one following reply. No URL/browser storage/fixture state may restore business status.
- The OpenAPI-generated TS runtime must expose typed `ApiClientError` with public `status` and `apiError` fields；JSON `ApiErrorResponse`、non-JSON、empty response、Abort and transport failures have explicit tests. Frontend must never parse `Error.message` to infer retryability.
- BDD-Gate P0.046 executes `AI failure → persisted retryable_failed → reload/get session → same ID retry → complete + unique assistant reply`, plus pending/terminal states and cross-user/privacy negatives. P0.044 retains immediate optimistic/pending success coverage; screenshots remain frontend owner evidence.

### Phase 11: Lease-bounded generation fencing and lazy convergence

- **RED 11.1 — migration/state table**: first add failing migration SQL-contract and store/domain tests for user-only `reply_generation` plus `reply_lease_expires_at`. Generation starts at 1 and survives terminal/complete history；only `pending` has a non-null lease；assistant rows expose neither field. These internal fields must not enter OpenAPI or generated clients.
- **GREEN 11.2 — 90-second lease**: reserve writes `pending(G1, serverNow+90s)`；a same-ID retry from `retryable_failed(Gn)` writes `pending(Gn+1, serverNow+90s)`；Fail/Commit clear lease while retaining generation。`Service.now` is the sole authoritative clock so unit and PostgreSQL tests can deterministically cross the boundary.
- **RED/GREEN 11.3 — lazy convergence**: `getPracticeSession` locks the authorized session, converts every expired `pending(Gn)` to `retryable_failed(Gn)` and reads the converged session in the same transaction. A same-ID reserve may instead atomically take over an expired row as `pending(Gn+1,newLease)`；an unexpired pending, terminal/complete row, payload mismatch or different new ID remains a typed conflict. No cron/background worker is required.
- **RED/GREEN 11.4 — generation fence**: reserve returns the internal generation to service；`CommitPracticeMessage` and `FailPracticeMessage` require expected generation and may mutate only `pending + generation match`. Stale G1 Commit/Fail after G2 reserve returns typed conflict with zero state/assistant writes；valid G2 can commit exactly one assistant reply.
- **Real PostgreSQL 11.5**: use independent DB connections plus a start barrier, not sequential calls, for exactly `TestIntegrationPracticeReplyConcurrentNewIDsReserveOnce`, `TestIntegrationPracticeReplyConcurrentSameIDInitialReserveOnce`, `TestIntegrationPracticeReplyConcurrentExpiredSameIDRetryAdvancesOneGeneration`, and `TestIntegrationPracticeReplyStaleGenerationFencedAfterGETRecovery`. The fourth test pauses G1, expires it through GET, reserves G2, releases stale G1 Commit and Fail, then proves only G2 can write one assistant reply.
- **Contract/BDD 11.6**: public `PracticeMessage` remains `clientMessageId + replyStatus` only；OpenAPI/codegen/fixtures remain compatible. P0.044/P0.046 consume one tracked Practice source manifest, record its SHA-256 fingerprint in trigger output, reject verifier-time drift, and record screenshot SHA-256/dimensions/viewport. Any source change invalidates prior evidence；historical PASS cannot close Phase 11.
- P0.046 exact markers are `PRACTICE_PENDING_LEASE_RECOVERY_PASS`, `PRACTICE_STALE_GENERATION_FENCED_PASS`, `PRACTICE_CONCURRENT_RESERVATION_PASS`, `PRACTICE_POST_TIMEOUT_RECONCILIATION_PASS`, `PRACTICE_TERMINAL_PLAN_RECOVERY_PASS` and `PRACTICE_EVIDENCE_FINGERPRINT_PASS`；P0.044 must emit `PRACTICE_IMMEDIATE_PENDING_PASS`, `PRACTICE_PERSISTED_PENDING_PASS` and the same fingerprint marker.

### Phase 12: Configured message and session text limits

- **RED 12.1**: construct UTF-8 byte fixtures at 32KiB/32KiB+1 per message and at 256KiB/256KiB+1 persisted session total. Existing 8,000-rune behavior or missing total cap must fail the new tests.
- **GREEN 12.2**: inject A4 `practice.maxMessageBytes=32768` and `practice.maxSessionTextBytes=262144` into `SendPracticeMessage`. Count bytes, not runes; evaluate before reservation and provider call; replay of an already accepted same ID remains idempotent.
- **Persistence 12.3**: session total is computed from authorized persisted `practice_messages` plus the candidate message in the same consistency boundary. Over-limit returns typed `VALIDATION_FAILED` with zero user/assistant/provider side effects; concurrency cannot let two messages jointly bypass the total.
- **Contract/BDD 12.4**: public RuntimeConfig exposes both values for frontend precheck without adding them to Practice operation bodies. P0.046 covers limit/limit+1, multibyte text, reload and same-ID behavior; backend remains authoritative.

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
- P0.044/P0.046 evidence is accepted only when its tracked source fingerprint and every screenshot hash/geometry still match the current tree.
- Practice 单条/会话文本默认边界分别为 32KiB/256KiB UTF-8 bytes；limit 接受，limit+1 在 reservation/provider 前拒绝且无半成品持久化。

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
| 2026-07-14 | 2.9 | Reopen Phase 12 for injected 32KiB message and 256KiB persisted-session UTF-8 byte limits. |
| 2026-07-14 | 2.8 | Add Phase 11 for a 90-second reply lease, internal generation fence, GET/same-ID lazy convergence, four concurrent PostgreSQL gates and fingerprint-bound P0.044/P0.046 evidence. |
| 2026-07-13 | 2.7 | Add server-persisted reply status, refresh-safe same-ID recovery and typed API-error handoff. |
| 2026-07-12 | 2.6 | Make 002 the sole reportable-completion/snapshot owner and lock exact P0.047 backend evidence. |
| 2026-07-12 | 2.5 | Require one candidate message before completion and atomically freeze report-context.v1. |
| 2026-07-12 | 2.4 | Bind completion facts to the TargetJob resume and separate immutable system policy from JSON-encoded untrusted follow-up context. |
| 2026-07-12 | 2.3 | Reopen Phase 8 so the committed session-completion event is the durable round-progress fact. |
| 2026-07-12 | 2.2 | Reopen follow-up messaging so every AI call uses the complete resume source snapshot and fails closed without evidence. |
| 2026-07-12 | 2.1 | Reopen for send/complete race protection and executable failure/recovery scenario evidence. |
| 2026-07-12 | 2.0 | Replace answer/turn event loop with message conversation loop. |
