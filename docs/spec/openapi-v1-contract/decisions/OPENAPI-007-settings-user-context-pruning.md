# OPENAPI-007 · Settings UserContext pruning

> **ID**: OPENAPI-007
> **状态**: accepted
> **日期**: 2026-07-15
> **版本**: 1.1

## 1 背景

`UserContext` 当前 required 暴露 `uiLanguage` 与 `preferredPracticeLanguage`，但正式前端不消费这两个后端字段：界面语言由 TopBar 本地 display preference 与 `Accept-Language` 承接，练习语言由具体岗位/练习计划业务合同承接。Settings 反而以静态列表展示手机号、界面语言和时区，形成没有真实数据源的假完整性。项目尚未上线，用户于 2026-07-15 明确批准设置简化方案 A。

## 2 决策

- `UserContext` 显式增加 `additionalProperties: false`，并收敛为 required `{id,email,displayName,profileCompletionRequired}`。
- authenticated `/me` 与 `PATCH /me` success 返回完整账号 email，供 Settings 正常显示；删除 `emailMasked`，不保留 alias/双字段。完整 email 不进入日志、场景证据或 public unauthenticated response。
- 从 OpenAPI source、Auth fixtures、Go/TS generated artifacts、backend mapper/store 与 frontend test builders 删除 `uiLanguage`、`preferredPracticeLanguage`；不提供 optional alias、默认值或兼容字段。
- `GET /api/v1/me` 与 `PATCH /api/v1/me` 的 method/path/operationId/status 保持不变；`DELETE /api/v1/me`、email-code/profile-completion/session 语义保持不变。
- TopBar 语言继续由前端 display preference + `Accept-Language` 表达；practice language 继续由对应业务 request/plan 表达，不迁回 `UserContext`。
- `user_settings.analytics_opt_in` 仍是 runtime-config 的有效后端事实，不属于本次 public `UserContext` 删除面。

## 3 影响

| 边界 | 受影响的项 | Owner |
|------|-----------|-------|
| 契约 | `UserContext` required/properties、Auth fixtures、baseline | B2 001/002/003 |
| 后端 | auth `UserContext`、store projection、handler mapping/tests | backend-auth/001 |
| 前端 | generated type、runtime/test builders、Settings runtime consumer | frontend-shell/001 |
| 数据库 | 删除 `user_settings.ui_language/preferred_practice_language/region/timezone`，保留 analytics opt-in | db-migrations-baseline/001 |
| Mock | dev mock 与 fixture projection | B2 002 + mock-contract-suite consumer |

## 4 迁移与回滚

- 003 Phase 12 在 baseline 不变时从 merge-base old baseline 到 proposed OpenAPI 生成 exact five-key findings，覆盖两个字段从 required/properties 删除及 `additionalProperties: false` closure；本设计阶段不手写 expected-findings JSON。
- 001/002、backend/frontend consumer 与 B4 migration gates 全部通过后，才允许原地 re-freeze v1.0.0 baseline。
- 任一 consumer 仍读取旧字段时整体回滚本 correction；不得把字段改为 optional 或在 mapper 中补常量来维持旧形态。
- 本项目未上线，不做灰度或双写；rollback 只通过整批 revert，并在重新实施前保持旧 baseline 未变。

## 5 验证边界

- B2 只拥有 schema/fixture/codegen/exact-diff；用户可见 Settings 行为由 frontend-shell BDD 与扩展后的 `E2E.P0.101` 承接。
- backend-auth focused contract 证明 `/me` 只返回四字段且 authenticated email 为完整账号值；B4 migration contract 证明四列删除、analytics opt-in 保留和 up/down/up 可逆。
- scoped negative search 允许本 ADR、history、plan 与 negative tests 提及旧字段；production OpenAPI/generated/backend/frontend/runtime fixture 不得有正向引用。

## 6 审计

| 项 | 内容 |
|----|------|
| 提议人 | product owner |
| Review | user explicitly approved Scheme A on 2026-07-15 |
| 实施分支 | `feat/settings-simplification-0715` |
| expected finding oracle | 由 openapi-v1-contract/003 Phase 12 在 TDD RED 后生成 |
| baseline | `openapi/baseline/openapi-v1.0.0.yaml`；consumer/migration gates 全绿后才允许 re-freeze |
| history | `2026-07-15 | 1.63 | OPENAPI-007 Settings UserContext pruning` |

## 7 修订记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-07-15 | 1.0 | 接受 UserContext 四字段最小投影与 settings display-preference DB 列删除。 |
| 2026-07-15 | 1.1 | 用户确认 Settings 正常显示完整账号 email；四字段合同将 `emailMasked` 原位替换为 `email`，不保留兼容 alias，并维持日志/场景证据脱敏边界。 |
