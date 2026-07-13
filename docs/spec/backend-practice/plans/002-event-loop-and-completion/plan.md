# 002 — Conversation Message Loop and Completion

> **版本**: 2.7
> **状态**: active
> **更新日期**: 2026-07-13

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)
**关联 Test Plan**: [test-plan](./test-plan.md)
**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## 1 目标

实现 `sendPracticeMessage` 连续聊天，并作为 `completePracticeSession` 的唯一 owner 负责可报告性校验、完成事务与 `report-context.v1` 原子快照：普通 user/assistant messages 按序持久化，不区分题目、回答或追问；失败可用同一 `clientMessageId` 恢复且不重复消息。Phase 10 将该恢复能力扩展到刷新/重挂载：后端持久化 reply status，`getPracticeSession` 返回原 user `clientMessageId/replyStatus`，前端不依赖浏览器存储。

## 2 Operation Matrix

| operationId | fixture | frontend consumer | handler | persistence | AI | scenario |
|-------------|---------|-------------------|---------|-------------|----|----------|
| `getPracticeSession` | `getPracticeSession.json::{default,prototype-baseline,pending-reply,retryable-failed,terminal-failed}`（后三项由 Phase 10 新增） | Practice session loader/remount recovery | `api/practice.GetPracticeSession` → `practice.GetPracticeSession` → SQL `GetSession` | `practice_messages.client_message_id/reply_status` read projection | none | P0.044/P0.046 |
| `sendPracticeMessage` | `sendPracticeMessage.json::{default,ai-timeout,validation,auth-not-found,session-not-found,conflict,mismatch}`（缺失项由 Phase 10 新增） | Practice message hook/row-local retry | `api/practice.SendPracticeMessage` → `practice.SendPracticeMessage` → SQL reserve/fail/commit | `practice_messages.client_message_id/reply_status`, task-runs | `practice.session.chat` | P0.044/P0.046 |
| `completePracticeSession` | `openapi/fixtures/PracticeSessions/completePracticeSession.json` | Practice finish hook | `backend-practice/002` handler/service/store（唯一 completion owner） | session/terminal messages/report-context.v1/job/outbox/idempotency | transaction 内无 AI；随后 report job | P0.047 owner artifact；P0.056/058 只消费 marker |

## 3 质量门禁分类

- **Plan 类型**: feature-behavior + API + backend + persistence。
- **TDD 策略**: Red tests cover message replay, concurrency, failure recovery, ordering, language, privacy and completion before implementation.
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

## 6 验收标准

- Multiple message pairs append in stable order with no question classification.
- Retries/concurrency cannot duplicate messages or provider calls after replay.
- Completion creates one report job and conversation-level handoff.
- Completion commits one auditable round fact that TargetJob read models can project without a mutable progress column.
- Wrong-resume plan/session facts cannot complete or advance the TargetJob's canonical prefix, even when both resumes belong to the same user.
- Zero-answer/pending-reply completion is rejected without side effects; successful completion atomically freezes one immutable current-shape report context under this plan's sole ownership.
- No raw message leaks outside authorized content/prompt/report paths.
- Reload/remount cannot strand a persisted user message: server reply state reconstructs pending/failure UI and same-ID retry converges to one reply.

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

## 8 修订记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-07-13 | 2.7 | Add server-persisted reply status, refresh-safe same-ID recovery and typed API-error handoff. |
| 2026-07-12 | 2.6 | Make 002 the sole reportable-completion/snapshot owner and lock exact P0.047 backend evidence. |
| 2026-07-12 | 2.5 | Require one candidate message before completion and atomically freeze report-context.v1. |
| 2026-07-12 | 2.4 | Bind completion facts to the TargetJob resume and separate immutable system policy from JSON-encoded untrusted follow-up context. |
| 2026-07-12 | 2.3 | Reopen Phase 8 so the committed session-completion event is the durable round-progress fact. |
| 2026-07-12 | 2.2 | Reopen follow-up messaging so every AI call uses the complete resume source snapshot and fails closed without evidence. |
| 2026-07-12 | 2.1 | Reopen for send/complete race protection and executable failure/recovery scenario evidence. |
| 2026-07-12 | 2.0 | Replace answer/turn event loop with message conversation loop. |
