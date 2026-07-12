# 001 — Plan and Session Orchestration

> **版本**: 2.1
> **状态**: completed
> **更新日期**: 2026-07-12

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)
**关联 Test Plan**: [test-plan](./test-plan.md)
**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## 1 目标

把 Practice plan/session foundation 从 question/turn 模型原地重构为 conversation/message 模型：

- plan 删除 question budget、mode 和 hint configuration。
- baseline migration 删除 `practice_turns/question_assessments`，新增 `practice_messages`。
- OpenAPI 删除 PracticeTurn/AssistantAction/event-answer schema，新增 message schemas 与 `sendPracticeMessage` operation。
- `startPracticeSession` 通过 `practice.session.chat` 生成 opening assistant message。
- `getPracticeSession` 返回 ordered messages。
- 保持 user isolation、idempotency、AI failure recovery、privacy 和 codegen drift gates。

## 2 Operation Matrix

| operationId | fixture | frontend consumer | backend handler | persistence | AI dependency | scenario coverage |
|-------------|---------|-------------------|-----------------|-------------|---------------|-------------------|
| `createPracticePlan` | `PracticePlans/createPracticePlan.json` | start helpers | existing practice handler | `practice_plans`, idempotency | none | P0.022/P0.070/P0.072 |
| `getPracticePlan` | `PracticePlans/getPracticePlan.json` | start helpers | existing practice handler | `practice_plans` | none | P0.022/P0.070 |
| `startPracticeSession` | `PracticeSessions/startPracticeSession.json` | start helpers | practice start handler/service | session + opening `practice_messages` | `practice.session.chat` | P0.023-P0.026 |
| `getPracticeSession` | `PracticeSessions/getPracticeSession.json` | Practice loader | practice read handler/store | session + messages | none | P0.023/P0.025/P0.044 |

## 3 质量门禁分类

- **Plan 类型**: feature-behavior + contract + migration + backend + codegen。
- **TDD 策略**: 每个 phase 先更新/新增 focused Red test，再修改 OpenAPI/shared/migration/service/store；断言来源见 test-plan/checklist。
- **BDD 策略**: 适用。P0.022-P0.026 覆盖 plan、opening message、AI failure、idempotency/isolation 与隐私。
- **替代验证 gate**: OpenAPI codegen/fixture、conventions codegen、migration up-down-up、prompt/schema/eval lint、privacy negative search。

## 4 Coverage Matrix

| Source | Category | Plan phase | Verification | Negative scope |
|--------|----------|------------|--------------|----------------|
| D-24 conversation | cross-layer contract | 1 | codegen/fixture/migration/prompt gates | questionBudget/PracticeTurn/QuestionAssessment |
| plan create/read | primary | 2 | service/store/API tests + P0.022 | mode/hint fields |
| session opening | primary | 3 | starter tests + P0.023 | first_question/currentTurn |
| AI failure retry | failure/recovery | 3 | starter failure matrix + P0.024 | canned opening/duplicate message |
| session read | boundary | 4 | ordered/empty/cross-user tests | local fixture transcript |
| idempotency/isolation | security/boundary | 5 | P0.025 | duplicate session/opening |
| privacy | privacy/observability | 5 | P0.026 + redaction lint | raw message in event/log/audit/task payload |

## 5 实施步骤

### Phase 1: Contract and baseline rebase

- Red tests lock the new 37-operation inventory, message schemas, 11 shared enums, 21 app tables and 6 prompt coordinates.
- Replace `appendSessionEvent` HTTP operation with `sendPracticeMessage` while keeping total endpoint count stable.
- Remove `PracticeMode` / `QuestionReviewStatus` and all question/report-review schemas.
- Rewrite baseline SQL, enum sources, seeds, prompt/rubric/profile/eval coordinates for `practice.session.chat`.
- Regenerate Go/TS artifacts and validate fixtures/baseline.

### Phase 2: PracticePlan simplification

- Remove question budget, mode and hints from request/domain/store.
- Preserve goal, interviewer persona, difficulty, language, time budget, resume/source/focus context.
- Cover baseline and report-derived plans, idempotency and cross-user isolation.
- Run P0.022/P0.070/P0.072 only after the real handler/store path compiles; Phase 1 contract work cannot substitute fixture-only evidence for this BDD gate.

### Phase 3: Session start with opening message

- Replace first-question generator with chat opening using empty ordered history.
- Reserve session, call AI outside transaction, persist one assistant message and session_started fact.
- Retry same IK after timeout without duplicate session/opening/outbox.
- Validate language and schema with one repair; never emit canned text.

### Phase 4: Session read model

- Return ordered messages with stable roles/seq/timestamps.
- Cover empty/loading states, pagination decision (P0 full bounded list), cross-user 404 and deleted context.

### Phase 5: Privacy, migration and BDD closeout

- Prove raw message content exists only in `practice_messages` and authorized prompt/read/report input.
- Run migration/codegen/prompt/fixture/full backend gates and P0.022-P0.026.
- Sync owner docs/index/context; leave plan active until downstream 002/003 and frontend/report owners close.

## 6 验收标准

- Contract truth sources contain message/conversation shapes and zero current question/hint/mode shapes.
- start returns exactly one opening assistant message; retries do not duplicate it.
- get session returns ordered messages and preserves isolation/privacy.
- All checklist, test checklist and BDD checklist items pass before completed.

## 7 风险与应对

| 风险 | 应对 |
|------|------|
| baseline edit leaves stale generated artifacts | codegen-check + baseline diff + negative search |
| AI timeout duplicates opening | three-stage reservation + IK replay tests |
| raw transcript leaks through events/logs | allowlist event payload + redaction tests |
| report consumer still expects turns | downstream backend-review/frontend-report plan gates block closeout |

## 8 修订记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-07-12 | 2.1 | Move the real P0.022 gate from Phase 1 to Phase 2 after handler implementation. |
| 2026-07-12 | 2.0 | Reopen for conversation/message model and opening assistant message. |
