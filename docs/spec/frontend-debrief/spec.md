# Frontend Debrief Spec

> **版本**: 1.8
> **状态**: deprecated
> **更新日期**: 2026-06-29

## 1 背景与目标

`frontend-debrief` 曾承接真实面试复盘页面、TopBar 入口、URL route、DebriefScreen、复盘问题建议、复盘分析和复盘面试 handoff。当前 [product-scope D-22](../product-scope/spec.md#31-已锁定决策) 已选择方案 B：P0 核心闭环只保留 JD / 简历 -> 模拟面试 -> 报告 -> 复练当前轮 / 进入下一轮，不再维护真实面试复盘模块或用户画像入口。

本 subject 因此退役。本文只作为历史索引和防回流说明保留，不再是可实施的 UI 真理源。任何涉及复盘 UI 恢复的需求，必须先修订 [product-scope](../product-scope/spec.md)、`docs/ui-design/` 和 `ui-design/`，再重新派生新 owner plan；不得从本退役文档或历史 plan 直接恢复 route、TopBar 入口、DebriefScreen、i18n key、pixel parity spec 或场景资产。

## 2 范围

### 2.1 In Scope

- 标记 `frontend-debrief` subject 已退役。
- 指向当前 owner：[product-scope/001-core-loop-module-pruning](../product-scope/plans/001-core-loop-module-pruning/plan.md)。
- 保留历史 plan 目录作为审计记录。

### 2.2 Out of Scope

- 不再定义任何 live `debrief` / `debrief_full` route、TopBar entry、DebriefScreen、复盘 Context、复盘 i18n、复盘 dev mock flow、复盘 pixel parity spec 或复盘 BDD 正向场景。
- 不再定义复盘面试 handoff、复盘语音模式、复盘 picker、复盘问题建议或复盘分析 UI。

## 3 用户决策 / 待确认事项

| ID | 决策 | 锁定值 | 影响 |
|----|------|--------|------|
| D-1 | Subject 状态 | deprecated | 本 subject 不再派生实现计划，不再作为当前工程真理源 |
| D-2 | 当前 UI 真理源 | `docs/ui-design/` + `ui-design/` 的 D-22 后三入口模型 | 正式 frontend 必须复刻当前原型，不得恢复旧复盘页面 |
| D-3 | 当前 owner | `product-scope/001-core-loop-module-pruning` | 模块删除、零残留 gate 和回归验证归 product-scope owner |
| D-4 | 历史内容 | 保留历史 plan 目录但不得复用为恢复依据 | 历史 PASS、历史 checklist 和旧 parity spec 只能作为审计线索 |

## 4 设计约束

- 当前 `frontend/` 和 `ui-design/` 不得重新引入复盘页面、用户画像页、复盘 TopBar entry 或画像用户菜单项。
- 旧 `/debrief`、`/#route=debrief`、`/#route=debrief_full` 和 `profile` 输入只能作为 legacy-negative normalization 输入，不能 materialize 业务页面。
- 防回归测试或负向 lint 可以保留退役词面，但必须明确用于禁止旧模块回流。

## 5 模块边界

| 边界 | Owner | 说明 |
|------|-------|------|
| 产品范围 | product-scope | D-22 负责裁剪决策和当前验收标准 |
| UI 真理源 | docs/ui-design / ui-design | 当前原型为三入口和设置/登出用户菜单 |
| 正式 frontend | frontend-shell + 当前业务 owner | 不再包含 DebriefScreen 或 ProfileScreen |
| 场景资产 | test/scenarios/e2e | 只保留 legacy-negative 和核心闭环正向场景 |

## 6 验收标准

| ID | 场景 | Given | When | Then | 对应 Plan |
|----|------|-------|------|------|-----------|
| C-1 | 退役零残留 | D-22 已生效 | 跑 product-scope/001 zero-reference gate | active frontend/ui-design/scenario assets 不出现复盘或画像 runtime surface | product-scope/001-core-loop-module-pruning |
| C-2 | Legacy input 不 materialize | 用户打开旧 route/hash/path | App normalize 路由 | URL 折回当前入口，不渲染 DebriefScreen 或 ProfileScreen | product-scope/001-core-loop-module-pruning |

## 7 关联计划

- [product-scope/001-core-loop-module-pruning](../product-scope/plans/001-core-loop-module-pruning/plan.md)
- 历史审计记录：[`001-debrief-screen-and-handoff`](./plans/001-debrief-screen-and-handoff/plan.md)
