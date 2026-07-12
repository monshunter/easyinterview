# 002 — Conversation Message Loop and Completion

> **版本**: 2.0
> **状态**: active
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

## 6 验收标准

- Multiple message pairs append in stable order with no question classification.
- Retries/concurrency cannot duplicate messages or provider calls after replay.
- Completion creates one report job and conversation-level handoff.
- No raw message leaks outside authorized content/prompt/report paths.

## 7 风险与应对

| 风险 | 应对 |
|------|------|
| failure after user persist | same clientMessageId resumable generation |
| concurrent submits | unique constraints + session-level reservation conflict |
| stale frontend expects AssistantAction | codegen/typecheck/negative search |

## 8 修订记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-07-12 | 2.0 | Replace answer/turn event loop with message conversation loop. |
