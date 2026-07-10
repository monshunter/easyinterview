# 004 — Report-derived Practice Plans

> **版本**: 1.2
> **状态**: completed
> **更新日期**: 2026-07-06

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)
**关联 Test Plan**: [test-plan](./test-plan.md)
**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## 1 目标

本计划只承接 report-derived practice plan 当前契约：

- `createPracticePlan` 支持 `goal IN ('retry_current_round','next_round') + sourceReportId`，写入 `practice_plans.source_report_id`。
- `getPracticePlan` 返回当前 plan 的 `sourceReportId`，同用户重放保持一致。
- `startPracticeSession` 对 report-derived plan 走 `practice.session.first_question` / A3 首题生成路径，不存在 source question bypass。
- `goal='debrief'`、`sourceDebriefId`、`source_debrief_id`、`PracticeGoalDebrief` 和 debrief-derived first-turn seeding 是禁止输入 / 禁止契约字段，只能出现在负向断言中。

## 2 背景

当前 `backend-practice/spec.md` 把 `PracticeGoal` 定义为 `baseline / retry_current_round / next_round`，并把 `sourceReportId` 作为唯一派生计划 source 字段。

本次修订不新增代码；它让 completed plan / test / BDD / context 与当前 report-derived 契约一致，`/implement` 或 `/plan-code-review` 从当前 owner 继续。

## 3 质量门禁分类

- **Plan 类型**: docs-only reconciliation for an already-implemented feature-behavior / contract plan。
- **TDD 策略**: 不适用：本次只修订文档和 context；当前实现证据来自现有 focused Go tests、OpenAPI inventory、scenario wrapper 与 negative grep。
- **BDD 策略**: 当前 BDD gate 为 `E2E.P0.070` / `E2E.P0.072`。
- **替代验证 gate**: `validate_context.py`、focused Go test existence/search、P0.070/P0.072 scenario docs、OpenAPI operation/enum search、negative grep、`make docs-check`、`git diff --check`。

## 3.1 Operation Matrix

| `operationId` | fixture | frontend consumer | backend handler | persistence | AI dependency | scenario coverage |
|---------------|---------|-------------------|-----------------|-------------|---------------|-------------------|
| `createPracticePlan` | `openapi/fixtures/PracticePlans/createPracticePlan.json` `report-derived` scenarios | Report next actions and workspace/practice owner create retry / next-round plans | `backend/internal/api/practice.Handler.CreatePracticePlan` + `backend/internal/practice.Service.CreatePracticePlan` + `backend/internal/store/practice.SQLRepository.CreatePlan` | `practice_plans.source_report_id` / `audit_events` / `idempotency_records` | none | `E2E.P0.070`, `E2E.P0.072` |
| `getPracticePlan` | `openapi/fixtures/PracticePlans/getPracticePlan.json` sourceReportId shape | Workspace/practice state refresh | `Handler.GetPracticePlan` + `Service.GetPracticePlan` + `SQLRepository.GetPlan` | `practice_plans` read | none | `E2E.P0.070` |
| `startPracticeSession` | `openapi/fixtures/PracticeSessions/startPracticeSession.json` current plan goals only | Interview session start | `Handler.StartPracticeSession` + `Service.StartPracticeSession` + `SQLRepository.ReserveSessionStart` / `CommitSessionStart` | `practice_sessions` / `practice_turns` / `practice_session_events` / `outbox_events` / `idempotency_records` | `practice.session.first_question` for baseline / retry_current_round / next_round | Covered by active backend-practice start-session gates; no source-question bypass scenario |

## 3.2 Coverage Matrix

| 行 | 类别 | source | verification | negative_scope |
|----|------|--------|--------------|----------------|
| R1 | Primary | retry_current_round / next_round report-derived plan creation | service/store/API derived-source tests + `E2E.P0.070` | no owner=004 rejection for valid report-derived sources |
| R2 | Failure / recovery | missing / cross-user / wrong-target report source | service/store/API source isolation tests + `E2E.P0.072` | no source existence leak across users |
| R3 | Cross-layer contract | B2 `sourceReportId`, B4 `source_report_id`, generated Go/TS | OpenAPI inventory + generated artifact search + fixture validation owner gates | no `sourceDebriefId` / `source_debrief_id` positive fields |
| R4 | Regression / negative | prohibited source fields / goals | negative grep across runtime/generated/fixtures/scenario docs | no `PracticeGoalDebrief`, `goal='debrief'`, debrief start scenario, or debrief first-question bypass |

## 4 实施步骤

### Phase 1: Report-derived Plan Contract

#### 1.1 Source report request / response contract

Keep `sourceReportId` as the only derived-plan source field in `CreatePracticePlanRequest` and `PracticePlan`.

#### 1.2 Report source validation

Keep service/store validation for `retry_current_round` and `next_round`: `sourceReportId` is required, must belong to the same user and target job, and must not leak cross-user source existence.

#### 1.3 Start-session behavior

Keep report-derived starts on the regular AI first-question path. Do not add a source-question bypass or raw-question seed.

### Phase 2: Source Boundary Reconciliation

#### 2.1 Prohibited source removal

Ensure current docs, context, fixtures, generated clients, runtime code, and scenarios do not list `sourceDebriefId`, `source_debrief_id`, `PracticeGoalDebrief`, or `goal='debrief'` as current positive contract.

#### 2.2 Scenario set reconciliation

Keep `E2E.P0.070` and `E2E.P0.072` as the active scenario proof.

## 5 验收标准

- Current plan/test/BDD/context describe only report-derived `retry_current_round` / `next_round` behavior as positive scope.
- `sourceDebriefId` / `source_debrief_id` / `PracticeGoalDebrief` / `goal='debrief'` do not appear in active runtime, generated artifacts, OpenAPI fixtures, or this plan's positive gates.
- `E2E.P0.070` and `E2E.P0.072` remain the only scenario IDs owned by this plan.
- `validate_context.py`, `make docs-check`, and `git diff --check` pass.

## 6 风险与应对

| 风险 | 应对措施 |
|------|----------|
| 范围外 source 字段看起来可用 | 只在负向断言中枚举禁止字段，正向 contract 只列 `sourceReportId` / `source_report_id` |
| 不匹配的 scenario IDs 或 no-op commands 被当作 proof | BDD/test docs 限定为 P0.070/P0.072，并验证匹配 tests/scripts 存在 |
| 禁止 source 字段回流到 generated artifacts | Negative grep 覆盖 OpenAPI、generated Go/TS、fixtures、backend runtime 和 scenario docs |

## 7 修订记录

| 日期 | 版本 | 变更 | 原因 |
|------|------|------|------|
| 2026-07-06 | 1.2 | Rename owner path to `004-report-derived-practice-plans`; current contract remains report-derived retry / next-round only. | Product-scope pruning requires current owner docs to use current owner language. |
| 2026-07-06 | 1.1 | Reconcile completed plan after product-scope D-22: current positive scope is report-derived retry / next-round only; out-of-scope source fields move to negative assertions. | Completed plan/context was still a discovery source and could reintroduce deleted work. |
| 2026-05-16 | 1.0 | Initial implementation of derived practice plans. | Initial contract delivery. |
