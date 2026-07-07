# 004 — Report-derived Practice Plans BDD Plan

> **版本**: 1.2
> **状态**: completed
> **更新日期**: 2026-07-06

**关联计划**: [plan](./plan.md) / [checklist](./checklist.md)
**关联 BDD Checklist**: [bdd-checklist](./bdd-checklist.md)

## 0 BDD 框架与编号

本 plan 当前只保留 report-derived source 的 BDD proof：

- 套件: `e2e`
- 阶段: `P0`
- 当前场景: `E2E.P0.070`, `E2E.P0.072`
- 执行入口: `cd backend && go test ./cmd/api -run 'TestE2EP0070|TestE2EP0072' -count=1`

## 1 场景矩阵

| 场景 ID | 名称 | 类别 | 关联 Phase | 覆盖 |
|---------|------|------|------------|------|
| `E2E.P0.070` | Report-derived practice plan create/read + idempotency replay | primary + cross-layer contract | Phase 1 | backend-practice C-2 |
| `E2E.P0.072` | Report source validation, isolation, and privacy | failure/recovery + privacy | Phase 1 + Phase 2 | backend-practice C-13 / C-16 |

## 2 Scenarios

| 场景 ID | Given | When | Then | 验证入口 |
|---------|-------|------|------|----------|
| `E2E.P0.070` | 用户 A 有 ready feedback report、target job、resume；source 属于同一 target job | 用户 A 分别调用 `createPracticePlan(goal='retry_current_round', sourceReportId=...)`、`createPracticePlan(goal='next_round', sourceReportId=...)` 并重放同一 Idempotency-Key | 两个 plan 均 201；`getPracticePlan` 返回 `sourceReportId`；same key replay 返回同 response；DB 写入 `source_report_id`；audit 仅含 ids/counts | `test/scenarios/e2e/p0-070-practice-derived-plan-create-read-replay/scripts/trigger.sh` -> `verify.sh` (`TestE2EP0070PracticeDerivedPlanCreateReadReplay`) |
| `E2E.P0.072` | 用户 A/B 各有 report source；另有 missing / wrong-target source | 用户 A 用缺失、跨用户、wrong-target source 调用 create/read flows；运行 privacy assertions | 错误 envelope 不泄露跨用户 source 内容；invalid source 返回 canonical envelope；audit/log/runner evidence 不含 raw source data | `test/scenarios/e2e/p0-072-practice-derived-source-isolation-privacy/scripts/trigger.sh` -> `verify.sh` (`TestE2EP0072PracticeDerivedSourceValidationIsolationPrivacy`) |

## 3 数据隔离与污染恢复

- 每个 scenario 使用独立 user / target job / report / plan / session / idempotency key。
- cleanup 顺序遵循 `test/scenarios/README.md`：scenario 自身数据优先，其次共享组件，最后才重建环境。
- 不预设 Helm chart、Kind namespace 或外部平台名称；当前执行入口为 Go HTTP scenario。
