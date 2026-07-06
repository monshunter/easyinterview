# 004 — Report-derived Practice Plans (Debrief Retired)

> **版本**: 1.1
> **状态**: completed
> **更新日期**: 2026-07-06

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)
**关联 Test Plan**: [test-plan](./test-plan.md)
**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## 1 目标

本目录名保留历史 `004-derived-plans-debrief` 锚点，但 D-22 后当前有效范围只剩 report-derived practice plan：

- `createPracticePlan` 支持 `goal IN ('retry_current_round','next_round') + sourceReportId`，写入 `practice_plans.source_report_id`。
- `getPracticePlan` 返回当前 plan 的 `sourceReportId`，同用户重放保持一致。
- `startPracticeSession` 对 report-derived plan 仍走 `practice.session.first_question` / A3 首题生成路径，不存在 debrief source 首题 bypass。
- `goal='debrief'`、`sourceDebriefId`、`source_debrief_id`、`PracticeGoalDebrief` 和 debrief-derived first-turn seeding 均已随 product-scope D-22 退役；只能作为 retired-negative gate，不再作为当前功能目标、scenario 或 future handoff。

## 2 背景

2026-05-16 的历史 004 曾同时覆盖 report-derived 与 debrief-derived practice plan。product-scope D-22 删除真实面试复盘模块后，当前 `backend-practice/spec.md` v1.13 已把 `PracticeGoal` 收敛为 `baseline / retry_current_round / next_round`，并明确 `debrief/sourceDebriefId/source_debrief_id` 不再是合法 plan source。

本次修订不新增代码；它清理 completed plan/test/BDD/context 的正向口径，使后续 `/implement` 或 `/plan-code-review` 不会从历史 debrief 主路径重新派生工作。

## 3 质量门禁分类

- **Plan 类型**: docs-only reconciliation for an already-implemented feature-behavior / contract plan。
- **TDD 策略**: 不适用：本次只修订文档和 context；当前实现证据来自现有 focused Go tests、OpenAPI inventory、scenario wrapper 与 retired-token negative grep。
- **BDD 策略**: 保留当前有效 BDD `E2E.P0.070` / `E2E.P0.072`；历史 debrief 专属 `E2E.P0.071` / `E2E.P0.073` 已退役，不再作为完成 gate。
- **替代验证 gate**: `validate_context.py`、focused Go test existence/search、P0.070/P0.072 scenario docs、OpenAPI operation/enum search、retired-token negative grep、`make docs-check`、`git diff --check`。

## 3.1 Operation Matrix

| `operationId` | fixture | frontend consumer | backend handler | persistence | AI dependency | scenario coverage |
|---------------|---------|-------------------|-----------------|-------------|---------------|-------------------|
| `createPracticePlan` | `openapi/fixtures/PracticePlans/createPracticePlan.json` `report-derived` scenarios | Report next actions and workspace/practice owner create retry / next-round plans | `backend/internal/api/practice.Handler.CreatePracticePlan` + `backend/internal/practice.Service.CreatePracticePlan` + `backend/internal/store/practice.SQLRepository.CreatePlan` | `practice_plans.source_report_id` / `audit_events` / `idempotency_records` | none | `E2E.P0.070`, `E2E.P0.072` |
| `getPracticePlan` | `openapi/fixtures/PracticePlans/getPracticePlan.json` sourceReportId shape | Workspace/practice state refresh | `Handler.GetPracticePlan` + `Service.GetPracticePlan` + `SQLRepository.GetPlan` | `practice_plans` read | none | `E2E.P0.070` |
| `startPracticeSession` | `openapi/fixtures/PracticeSessions/startPracticeSession.json` current plan goals only | Interview session start | `Handler.StartPracticeSession` + `Service.StartPracticeSession` + `SQLRepository.ReserveSessionStart` / `CommitSessionStart` | `practice_sessions` / `practice_turns` / `practice_session_events` / `outbox_events` / `idempotency_records` | `practice.session.first_question` for baseline / retry_current_round / next_round | Covered by active backend-practice start-session gates; no debrief bypass scenario remains |

## 3.2 Coverage Matrix

| 行 | 类别 | source | verification | negative_scope |
|----|------|--------|--------------|----------------|
| R1 | Primary | retry_current_round / next_round report-derived plan creation | service/store/API derived-source tests + `E2E.P0.070` | no owner=004 rejection for valid report-derived sources |
| R2 | Failure / recovery | missing / cross-user / wrong-target report source | service/store/API source isolation tests + `E2E.P0.072` | no source existence leak across users |
| R3 | Cross-layer contract | B2 `sourceReportId`, B4 `source_report_id`, generated Go/TS | OpenAPI inventory + generated artifact search + fixture validation owner gates | no `sourceDebriefId` / `source_debrief_id` positive fields |
| R4 | Regression / legacy-negative | D-22 debrief retirement | retired-token grep across runtime/generated/fixtures/scenario docs | no `PracticeGoalDebrief`, `goal='debrief'`, debrief start scenario, or debrief first-question bypass |

## 4 实施步骤

### Phase 1: Report-derived Plan Contract

#### 1.1 Source report request / response contract

Keep `sourceReportId` as the only derived-plan source field in `CreatePracticePlanRequest` and `PracticePlan`.

#### 1.2 Report source validation

Keep service/store validation for `retry_current_round` and `next_round`: `sourceReportId` is required, must belong to the same user and target job, and must not leak cross-user source existence.

#### 1.3 Start-session behavior

Keep report-derived starts on the regular AI first-question path. Do not add a source-question bypass or debrief raw-question seed.

### Phase 2: D-22 Retired-negative Reconciliation

#### 2.1 Debrief source removal

Ensure current docs, context, fixtures, generated clients, runtime code, and scenarios do not list `sourceDebriefId`, `source_debrief_id`, `PracticeGoalDebrief`, or `goal='debrief'` as current positive contract.

#### 2.2 Scenario set reconciliation

Keep `E2E.P0.070` and `E2E.P0.072` as the active scenario proof. Remove historical `E2E.P0.071` / `E2E.P0.073` debrief-specific gates from this owner plan.

## 5 验收标准

- Current plan/test/BDD/context describe only report-derived `retry_current_round` / `next_round` behavior as positive scope.
- `sourceDebriefId` / `source_debrief_id` / `PracticeGoalDebrief` / `goal='debrief'` do not appear in active runtime, generated artifacts, OpenAPI fixtures, or this plan's positive gates.
- `E2E.P0.070` and `E2E.P0.072` remain the only scenario IDs owned by this plan.
- `validate_context.py`, `make docs-check`, and `git diff --check` pass.

## 6 风险与应对

| 风险 | 应对措施 |
|------|----------|
| Historical directory name makes debrief look current | Keep the directory path, but make every current section state report-derived only and debrief retired-negative only |
| Old scenario IDs or no-op commands are treated as proof | Keep BDD/test docs limited to P0.070/P0.072 and verify matching tests/scripts exist |
| Retired source fields re-enter generated artifacts | Retired-token grep covers OpenAPI, generated Go/TS, fixtures, backend runtime, and scenario docs |

## 7 修订记录

| 日期 | 版本 | 变更 | 原因 |
|------|------|------|------|
| 2026-07-06 | 1.1 | Reconcile completed plan after product-scope D-22: current positive scope is report-derived retry / next-round only; debrief-derived source, scenarios, and first-turn seeding are retired-negative. | Completed plan/context was still a discovery source and could reintroduce deleted debrief work. |
| 2026-05-16 | 1.0 | Historical implementation of derived practice plans. | Superseded in-place by D-22 pruning semantics. |
