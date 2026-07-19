# OPENAPI-008 · Account theme and generic updateMe

> **ID**: OPENAPI-008
> **状态**: accepted
> **日期**: 2026-07-19
> **版本**: 1.0

## 1 背景

设置页需要把主体色作为账号偏好跨设备持久化。现有 `PATCH /api/v1/me` 的 method/path 已具备正确 owner，但 operationId `completeMyProfile` 与只允许资料补全的 request schema 无法表达后续账号设置。项目尚未上线；用户于 2026-07-19 明确选择方案 B：原地泛化为 `updateMe`，并要求避免页面切换时重复远端读取主体色。

## 2 决策

- 保留 protected `PATCH /api/v1/me`、session-cookie security、200 success 和 profile-completion 语义，operationId 原地改为 `updateMe`。
- `UpdateMeRequest` 是 closed object，可提交完整的 `displayName + acceptedTerms` 资料补全字段对、`displayPreferences`，或二者组合；空请求、半组资料字段和非法主题值 fail closed。
- `UserContext` required 新增 closed `displayPreferences`；当前账号主题为 `ocean|plum`，custom accent 为 nullable `{h,c}`，其中 `0 <= h < 360`、`0 <= c <= 0.28`。
- profile 与 display preferences 同次提交时由后端单事务写入；失败不得部分更新。
- 登录 bootstrap / auth recovery 的既有 `GET /me` 响应直接携带主题。路由切换和 Settings mount 不再请求 `/me`；滑块只本地预览，点击保存只发一次 `PATCH /me`，成功响应直接刷新内存 auth context，不追加 follow-up GET。
- `user_settings` 继续拥有账号 display preferences，新增 `theme/custom_accent_hue/custom_accent_chroma`；不使用 `localStorage`、URL 或前端 fixture 作为业务事实源。

## 3 影响

| 边界 | 受影响项 | Owner |
|------|----------|-------|
| 契约 | operationId、request schema、UserContext、fixtures、baseline | B2 001/002/003 |
| 后端 | generic handler/service、事务 store、current-user projection | backend-auth/001 |
| 数据库 | `user_settings` 三列与约束 | db-migrations-baseline/001 |
| 前端 | runtime hydration、Settings preview/save、TopBar 去主题控件 | frontend-shell/001/002 |
| E2E | 登录后保存主题、跨路由不重复读取、重登恢复 | `E2E.P0.101` 原地扩展 |

## 4 迁移与审计

- baseline 未改动时，`make openapi-diff` 必须保留 3 breaking + 4 additive findings；另以 operation invariant 显式记录 `completeMyProfile -> updateMe`，因为当前通用 wrapper 不把 operationId rename 单列为 finding。
- consumer、migration、fixture/codegen 与 root regression 全绿后，才允许在未上线 v1.0.0 上原地 re-freeze；不保留旧 operation、request schema、fixture 或 compatibility method。
- 回滚只能整批 revert；不得双写旧 operation 或把主题降级为前端本地业务状态。

## 5 修订记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-07-19 | 1.0 | 接受方案 B：泛化 PATCH /me 为 updateMe，并把账号主题加入一次 bootstrap 投影与单次保存事务。 |
