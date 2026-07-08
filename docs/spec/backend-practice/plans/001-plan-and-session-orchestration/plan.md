# 001 — Plan and Session Orchestration

> **版本**: 1.6
> **状态**: completed
> **更新日期**: 2026-07-08

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)
**关联 Test Plan**: [test-plan](./test-plan.md)
**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## 1 目标

完成 backend-practice baseline plan/session foundation：

- `createPracticePlan` / `getPracticePlan` / `startPracticeSession` / `getPracticeSession` real backend handler、service、store。
- `idempotency_records` user-scoped 幂等基础。
- `practice.session.started` outbox emit。
- `startPracticeSession` 三段式首题生成。
- flat Resume `resumeId` 绑定与 `resumes.structured_profile` 首题 prompt context。

## 2 当前 Operation Matrix

| operationId | fixture | frontend consumer | backend handler | persistence | AI dependency | scenario coverage |
|-------------|---------|-------------------|-----------------|-------------|---------------|-------------------|
| `createPracticePlan` | `openapi/fixtures/PracticePlans/createPracticePlan.json` | `frontend-workspace-and-practice` start flow | `backend/internal/api/practice.CreatePracticePlan` | `practice_plans`, `idempotency_records`, `audit_events` | none | `E2E.P0.022`, `E2E.P0.025` |
| `getPracticePlan` | `openapi/fixtures/PracticePlans/getPracticePlan.json` | workspace plan refresh | `backend/internal/api/practice.GetPracticePlan` | `practice_plans` | none | `E2E.P0.022`, `E2E.P0.025` |
| `startPracticeSession` | `openapi/fixtures/PracticeSessions/startPracticeSession.json` | workspace handoff to `practice` | `backend/internal/api/practice.StartPracticeSession` | `practice_sessions`, `practice_turns`, `practice_session_events`, `outbox_events`, `idempotency_records`, `ai_task_runs` | `practice.session.first_question` + A3 `AIClient.Complete` | `E2E.P0.023`, `E2E.P0.024`, `E2E.P0.025`, `E2E.P0.026` |
| `getPracticeSession` | `openapi/fixtures/PracticeSessions/getPracticeSession.json` | practice refresh / resume | `backend/internal/api/practice.GetPracticeSession` | `practice_sessions`, `practice_turns` | none | `E2E.P0.023`, `E2E.P0.025` |

## 3 质量门禁分类

- **Plan 类型**: `feature-behavior` + `contract` + `migration` + `code-internal`
- **TDD 策略**: 已按 `/implement backend-practice/001-plan-and-session-orchestration backend` → `/tdd` 执行；每个 checklist item 有 Go focused test、contract test、drift gate 或 scenario gate。
- **BDD 策略**: 已维护 [bdd-plan](./bdd-plan.md) 与 [bdd-checklist](./bdd-checklist.md)，覆盖 `E2E.P0.022`-`E2E.P0.026`。
- **替代验证 gate**: shared/openapi/events/migration drift gate、practice focused Go tests、privacy grep、idempotency conflict/replay tests、docs/index checks。

## 4 实施结果

### Phase 1: Contract Foundation

PracticeMode / PracticeGoal / error codes / OpenAPI / event refs / migration baseline / idempotency table 与 plan 001 backend foundation 对齐。

### Phase 2: Plan And Session Success Path

`createPracticePlan`、`getPracticePlan`、`startPracticeSession`、`getPracticeSession` 已接入 real handler / service / store。`startPracticeSession` 使用 reserve → AI → commit 三段式，AI 调用不在 DB transaction 内。

### Phase 3: Failure, Idempotency, Isolation

AI error mapping、failed reservation retry、success replay、body mismatch conflict、cross-user isolation、TTL reset、single active session guard 和 shared peppered key hash 已覆盖。

### Phase 4: Privacy And Observability

AI task run metadata、audit metadata、outbox payload、metric label 与 log redaction 已覆盖；practice event/job payload 不包含题目、回答、hint、prompt、response 或 secret 明文。

### Phase 5: Flat Resume Binding

`createPracticePlan` 使用 `resumeId` 与 `practice_plans.resume_id`。`startPracticeSession` reservation 从 flat `resumes.structured_profile` 读取简历摘要，并渲染到 first-question prompt 的 `{{resume_profile}}` context。

### Phase 6: PracticePlan resumeId response remediation

`createPracticePlan` / `getPracticePlan` responses must return the persisted `resumeId` from `practice_plans.resume_id`, matching the request contract and enabling frontend current-plan refresh to keep the bound resume.

Verified focused practice suites, OpenAPI generated contract, fixture validation and cmd/api E2E focused gates on 2026-07-08.

## 5 验收标准

- 4 个 operation 的 focused Go tests 与 handler/store/service tests 通过。
- `E2E.P0.022`-`E2E.P0.026` 场景资产与执行项已完成。
- practice 包内 flat-resume residual grep 零命中。
- `startPracticeSession` first-question prompt 不含未渲染模板 token，且携带 language、role、skills、resume profile、rubric dimensions、practice goal。
- plan / checklist / test / BDD 文档 Header 与 INDEX 同步。

## 6 风险与应对

| 风险 | 应对 |
|------|------|
| AI 调用进入长事务 | `session_starter` test 断言 AI outside transaction |
| 首题缺少扁平简历上下文 | `TestStartPracticeSessionRunsThreeStepFlowWithAIOutsideTransactions` 断言 `resume_profile` 渲染 |
| SQL reservation 漏传 resume profile | `TestSQLRepositoryReserveSessionStartReusesFailedRetryableRecord` 断言 reservation carries structured profile |
| 幂等 replay 读可变 session 状态 | stored response snapshot test 固定 replay source |
| prompt / answer 明文泄露 | redaction tests + scenario grep gate |

## 7 关联文档

- [Spec](../../spec.md)
- [Checklist](./checklist.md)
- [Test Plan](./test-plan.md)
- [BDD Plan](./bdd-plan.md)
- [OpenAPI v1 Contract](../../../openapi-v1-contract/spec.md)
- [DB Migrations Baseline](../../../db-migrations-baseline/spec.md)
- [Event and Outbox Contract](../../../event-and-outbox-contract/spec.md)
- [Prompt Rubric Registry](../../../prompt-rubric-registry/spec.md)
