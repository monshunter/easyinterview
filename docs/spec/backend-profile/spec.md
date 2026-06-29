# Backend Profile Spec

> **版本**: 1.3
> **状态**: deprecated
> **更新日期**: 2026-06-29

## 1 背景与目标

`backend-profile` 曾承接候选人画像和 experience cards 的后端 API、store、privacy delete internal API 与聚合读取接口。当前 [product-scope D-22](../product-scope/spec.md#31-已锁定决策) 已选择方案 B：删除用户画像业务模块，保留账号资料补全和设置隐私，但不再维护 Candidate Profile / Experience Card 语义。

本 subject 因此退役。本文只作为历史索引和防回流说明保留，不再是可实施的设计真理源。任何涉及用户画像业务模块恢复的需求，必须先修订 [product-scope](../product-scope/spec.md) 并重新派生新 owner plan；不得从本退役文档或历史 plan 直接恢复 Profile tag、candidate profile DB 表、experience cards、画像页面或画像 internal API。

## 2 范围

### 2.1 In Scope

- 标记 `backend-profile` subject 已退役。
- 区分仍保留的 Auth `completeMyProfile` 账号资料补全和已删除的候选人画像业务模块。
- 指向当前 owner：[product-scope/001-core-loop-module-pruning](../product-scope/plans/001-core-loop-module-pruning/plan.md)。
- 保留历史 plan 目录作为审计记录。

### 2.2 Out of Scope

- 不再定义任何 live Profile API、candidate profile store、experience cards CRUD、画像聚合 internal API、画像隐私删除 runner、画像 UI 或画像 AI 能力。
- 不再为岗位推荐、简历、报告或复盘提供画像证据源。

## 3 用户决策 / 待确认事项

| ID | 决策 | 锁定值 | 影响 |
|----|------|--------|------|
| D-1 | Subject 状态 | deprecated | 本 subject 不再派生实现计划，不再作为当前工程真理源 |
| D-2 | Auth profile completion | 保留 | `completeMyProfile` 只表示首次账号资料补全，不恢复候选人画像模块 |
| D-3 | 当前 owner | `product-scope/001-core-loop-module-pruning` | 模块删除、零残留 gate 和回归验证归 product-scope owner |
| D-4 | 历史内容 | 保留历史 plan 目录但不得复用为恢复依据 | 历史 PASS、历史 checklist 和旧 operation matrix 只能作为审计线索 |

## 4 设计约束

- 当前代码、OpenAPI、fixtures、migrations 和场景资产不得重新引入 Candidate Profile / Experience Card runtime surface。
- `profile` route 不再是目标 route；旧输入必须归一到当前核心入口或设置语义，不得渲染画像页面。
- `structuredProfile` 作为简历结构化字段仍然有效，不代表候选人画像模块。
- AI model profile / provider profile 术语仍然有效，不代表用户画像业务模块。

## 5 模块边界

| 边界 | Owner | 说明 |
|------|-------|------|
| 产品范围 | product-scope | D-22 负责裁剪决策和当前验收标准 |
| Auth 账号资料补全 | backend-auth / frontend-shell | `completeMyProfile` 保留，但仅用于登录后资料补全 |
| API / fixtures | openapi-v1-contract | 当前 baseline 不含 Profile tag |
| DB / migrations | db-migrations-baseline | 当前 baseline 不含 candidate_profiles 或 experience_cards |

## 6 验收标准

| ID | 场景 | Given | When | Then | 对应 Plan |
|----|------|-------|------|------|-----------|
| C-1 | 退役零残留 | D-22 已生效 | 跑 product-scope/001 zero-reference gate | active code/contract/generated/fixtures/migrations/config/scenarios 不出现候选人画像 runtime surface | product-scope/001-core-loop-module-pruning |
| C-2 | Auth profile completion 不混淆 | 首次登录资料补全仍存在 | 检查 Auth tag 与 UI 设置入口 | 只保留账号资料补全，不出现 CandidateProfile / ExperienceCard API 或画像页入口 | product-scope/001-core-loop-module-pruning |

## 7 关联计划

- [product-scope/001-core-loop-module-pruning](../product-scope/plans/001-core-loop-module-pruning/plan.md)
- 历史审计记录：[`001-candidate-profile-and-experience-cards`](./plans/001-candidate-profile-and-experience-cards/plan.md)
