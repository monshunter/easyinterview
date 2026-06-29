# Backend Jobs Recommendations Spec

> **版本**: 2.1
> **状态**: deprecated
> **更新日期**: 2026-06-29

## 1 背景与目标

`backend-jobs-recommendations` 曾承接 JobMatch / 岗位推荐后端、agent scan、watchlist、saved search 和画像聚合。该 subject 已在 product-scope D-17 中整体退役；当前 product-scope D-22 又删除了其历史依赖的候选人画像和复盘计数来源。

本 subject 因此退役。本文只作为历史索引和防回流说明保留，不再是可实施的设计真理源。任何岗位推荐或外部搜岗能力恢复，必须先修订 [product-scope](../product-scope/spec.md) 并重新派生新 owner plan；不得从本退役文档或历史 plan 直接恢复 JobMatch route、API、DB、event、job、AI feature key、画像聚合或场景资产。

## 2 范围

### 2.1 In Scope

- 标记 `backend-jobs-recommendations` subject 已退役。
- 指向当前删除 owner：[product-scope](../product-scope/spec.md) D-17 / D-22。
- 保留历史 plan 目录作为审计记录。

### 2.2 Out of Scope

- 不再定义任何 live JobMatch endpoint、agent scan worker、watchlist/saved-search store、market signal、JD recommendation AI feature key、画像聚合 internal API 或隐私删除链路。
- 不再消费 `backend-profile`、`backend-debrief` 或候选人画像 / 复盘计数。

## 3 用户决策 / 待确认事项

| ID | 决策 | 锁定值 | 影响 |
|----|------|--------|------|
| D-1 | Subject 状态 | deprecated | 本 subject 不再派生实现计划，不再作为当前工程真理源 |
| D-2 | 删除依据 | product-scope D-17 + D-22 | JobMatch 及其画像/复盘依赖均不得恢复为 P0 runtime |
| D-3 | 历史内容 | 保留历史 plan 目录但不得复用为恢复依据 | 历史 PASS、历史 checklist 和旧 operation matrix 只能作为审计线索 |

## 4 设计约束

- 当前代码、OpenAPI、fixtures、migrations、shared events/jobs、AI config 和场景资产不得重新引入 JobMatch runtime surface。
- 当前核心闭环不依赖候选人画像、复盘计数、watchlist、saved search 或外部搜岗 agent。
- 防回归测试或负向 lint 可以保留退役词面，但必须明确用于禁止旧模块回流。

## 5 模块边界

| 边界 | Owner | 说明 |
|------|-------|------|
| 产品范围 | product-scope | D-17 / D-22 负责裁剪决策和当前验收标准 |
| API / fixtures | openapi-v1-contract | 当前 baseline 不含 JobMatch / Profile / Debriefs tags |
| DB / migrations | db-migrations-baseline | 当前 baseline 不含 jd_match、candidate profile 或 debrief runtime 表 |
| AI config | prompt-rubric-registry / ai-provider-and-model-routing | 当前 config 不含 jd_match、profile.update 或 debrief feature key |

## 6 验收标准

| ID | 场景 | Given | When | Then | 对应 Plan |
|----|------|-------|------|------|-----------|
| C-1 | 退役零残留 | D-17 / D-22 已生效 | 跑 current product-scope zero-reference gate | active code/contract/generated/fixtures/migrations/config/scenarios 不出现 JobMatch / Profile / Debrief runtime surface | product-scope |
| C-2 | 历史文档不恢复能力 | 历史 plan 仍保留 | 新实现或 review 读取本 subject | 必须以 product-scope 当前 spec 为准，不得从历史 plan 恢复 JobMatch / 画像 / 复盘依赖 | product-scope |

## 7 关联计划

- [product-scope](../product-scope/spec.md)
- 历史审计记录：[`001-jd-match-real-backend-baseline`](./plans/001-jd-match-real-backend-baseline/plan.md)
