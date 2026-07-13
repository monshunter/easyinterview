# 004 — Report-derived Practice Plans BDD Plan

> **版本**: 1.8
> **状态**: completed
> **更新日期**: 2026-07-12

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
| `E2E.P0.070` | Report-derived practice plan create/read + idempotency replay | primary + cross-layer contract | Phase 3 | backend-practice C-11 |
| `E2E.P0.072` | Report source validation, isolation, and privacy | failure/recovery + privacy | Phase 3 | backend-practice C-11 / C-12 |

## 2 Scenarios

| 场景 ID | Given | When | Then | 验证入口 |
|---------|-------|------|------|----------|
| `E2E.P0.070` | 用户 A 有 current-shape ready report，分别包含 empty focus、issue-backed code→label+issues focus 与 canonical successor；F3 已激活 immutable practice v0.2 pair | 只用 goal+sourceReportId 调用 retry/next create，启动 session 并重放同一 IK | empty retry 为通用同轮复练且无伪造 focus；非空 retry 投影 codes 且 runtime/F3 v0.2 通过 `semanticFocus/{{semantic_focus_json}}` 传递 label/issues；next 派生 successor/settings 且 focus 为空；read/start/IK replay 精确 | `test/scenarios/e2e/p0-070-practice-derived-plan-create-read-replay/scripts/trigger.sh` -> `verify.sh` |
| `E2E.P0.072` | missing/cross-user/wrong target/resume/round/persona/language/budget/non-ready/missing-context/unsupported-or-duplicate-non-empty-focus source | 调用 derived create/read 并运行 privacy assertions | 统一 fail closed、零 plan insert、零 source leak；空 focus 不进入失败矩阵；输出 named isolation markers | `test/scenarios/e2e/p0-072-practice-derived-source-isolation-privacy/scripts/trigger.sh` -> `verify.sh` |

## 3 数据隔离与污染恢复

- 每个 scenario 使用独立 user / target job / report / plan / session / idempotency key。
- cleanup 顺序遵循 `test/scenarios/README.md`：scenario 自身数据优先，其次共享组件，最后才重建环境。
- 不预设 Helm chart、Kind namespace 或外部平台名称；当前执行入口为 Go HTTP scenario。
