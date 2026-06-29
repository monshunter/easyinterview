# Backend Debrief Spec

> **版本**: 1.7
> **状态**: deprecated
> **更新日期**: 2026-06-29

## 1 背景与目标

`backend-debrief` 曾承接真实面试复盘的后端 API、worker、DB、event、job 和 AI feature key。当前 [product-scope D-22](../product-scope/spec.md#31-已锁定决策) 已选择方案 B：P0 核心闭环只保留 JD / 简历 -> 模拟面试 -> 报告 -> 复练当前轮 / 进入下一轮，不再维护真实面试复盘模块。

本 subject 因此退役。本文只作为历史索引和防回流说明保留，不再是可实施的设计真理源。任何涉及复盘模块恢复的需求，必须先修订 [product-scope](../product-scope/spec.md) 并重新派生新 owner plan；不得从本退役文档或历史 plan 直接恢复 route、API、DB、event、job、AI feature key 或场景资产。

## 2 范围

### 2.1 In Scope

- 标记 `backend-debrief` subject 已退役。
- 指向当前 owner：[product-scope/001-core-loop-module-pruning](../product-scope/plans/001-core-loop-module-pruning/plan.md)。
- 保留历史 plan 目录作为审计记录。

### 2.2 Out of Scope

- 不再定义任何 live backend handler、service、store、worker、migration、OpenAPI operation、fixture、shared event/job、prompt、rubric、eval、scenario 或 internal API。
- 不再定义复盘派生 practice plan、复盘问题建议、复盘报告生成、复盘历史浏览、复盘语音集成或复盘数据保留策略。

## 3 用户决策 / 待确认事项

| ID | 决策 | 锁定值 | 影响 |
|----|------|--------|------|
| D-1 | Subject 状态 | deprecated | 本 subject 不再派生实现计划，不再作为当前工程真理源 |
| D-2 | 当前 owner | `product-scope/001-core-loop-module-pruning` | 模块删除、零残留 gate 和回归验证归 product-scope owner |
| D-3 | 历史内容 | 保留历史 plan 目录但不得复用为恢复依据 | 历史 PASS、历史 checklist 和旧 operation matrix 只能作为审计线索 |

## 4 设计约束

- 当前代码、OpenAPI、fixtures、migrations、shared events/jobs、AI config 和场景资产不得重新引入 debrief runtime surface。
- `listPracticeSessions` 保留为核心 practice recovery contract，不再承担复盘 picker 语义。
- 防回归测试或负向 lint 可以保留退役词面，但必须明确用于禁止旧模块回流。

## 5 模块边界

| 边界 | Owner | 说明 |
|------|-------|------|
| 产品范围 | product-scope | D-22 负责裁剪决策和当前验收标准 |
| API / fixtures | openapi-v1-contract | 当前 baseline 不含 Debriefs tag |
| DB / migrations | db-migrations-baseline | 当前 baseline 不含 debriefs 表或 debrief job/event enum |
| AI config | prompt-rubric-registry / ai-provider-and-model-routing | 当前 config 不含 debrief feature key 或 profile |

## 6 验收标准

| ID | 场景 | Given | When | Then | 对应 Plan |
|----|------|-------|------|------|-----------|
| C-1 | 退役零残留 | D-22 已生效 | 跑 product-scope/001 zero-reference gate | active code/contract/generated/fixtures/migrations/config/scenarios 不出现复盘 runtime surface | product-scope/001-core-loop-module-pruning |
| C-2 | 历史文档不恢复能力 | 历史 plan 仍保留 | 新实现或 review 读取本 subject | 必须以 product-scope 当前 spec 为准，不得从历史 plan 恢复复盘模块 | product-scope/001-core-loop-module-pruning |

## 7 关联计划

- [product-scope/001-core-loop-module-pruning](../product-scope/plans/001-core-loop-module-pruning/plan.md)
- 历史审计记录：[`001-debrief-record-and-analysis`](./plans/001-debrief-record-and-analysis/plan.md)
