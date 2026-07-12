# 002 — Conversation Message Loop and Completion

> **版本**: 2.4
> **状态**: completed
> **更新日期**: 2026-07-12

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)
**关联 Test Plan**: [test-plan](./test-plan.md)
**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## 1 目标

实现 `sendPracticeMessage` 连续聊天与 `completePracticeSession`：普通 user/assistant messages 按序持久化，不区分题目、回答或追问；失败可用同一 `clientMessageId` 恢复且不重复消息。

## 2 Operation Matrix

| operationId | fixture | frontend consumer | handler | persistence | AI | scenario |
|-------------|---------|-------------------|---------|-------------|----|----------|
| `sendPracticeMessage` | `sendPracticeMessage.json` | Practice conversation hook | new message handler/service/store | `practice_messages`, task-runs | `practice.session.chat` | P0.044/P0.046 |
| `completePracticeSession` | existing fixture | finish hook | existing completion path | session/report/job/outbox/idempotency | report job | P0.047/P0.056 |

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

## 6 验收标准

- Multiple message pairs append in stable order with no question classification.
- Retries/concurrency cannot duplicate messages or provider calls after replay.
- Completion creates one report job and conversation-level handoff.
- Completion commits one auditable round fact that TargetJob read models can project without a mutable progress column.
- Wrong-resume plan/session facts cannot complete or advance the TargetJob's canonical prefix, even when both resumes belong to the same user.
- No raw message leaks outside authorized content/prompt/report paths.

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

## 8 修订记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-07-12 | 2.4 | Bind completion facts to the TargetJob resume and separate immutable system policy from JSON-encoded untrusted follow-up context. |
| 2026-07-12 | 2.3 | Reopen Phase 8 so the committed session-completion event is the durable round-progress fact. |
| 2026-07-12 | 2.2 | Reopen follow-up messaging so every AI call uses the complete resume source snapshot and fails closed without evidence. |
| 2026-07-12 | 2.1 | Reopen for send/complete race protection and executable failure/recovery scenario evidence. |
| 2026-07-12 | 2.0 | Replace answer/turn event loop with message conversation loop. |
