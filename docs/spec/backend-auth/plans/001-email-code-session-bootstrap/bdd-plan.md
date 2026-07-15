# Backend Auth BDD Plan

> **版本**: 1.10
> **状态**: completed
> **更新日期**: 2026-07-15

**关联 Plan**: [plan](./plan.md)

## Domain behavior

| Behavior ID | Given | When | Then | 验证入口 |
|-------------|-------|------|------|----------|
| `BDD.AUTH.EMAIL.001` | 新邮箱或已有账号发起 email-code 登录 | challenge、verify、读取最小 `/me`、补全 profile、logout/relogin | session/profile 状态按账号持久化；UserContext 只有 id/full email/display name/completion flag；完整 email 只在 authenticated response 中返回且不进入日志；非法/重放请求 fail closed | backend Auth contract/integration tests，由根 `make test` 承接 |

## 当前真实 E2E handoff

| E2E ID | Given | When | Then | Owner |
|--------|-------|------|------|-------|
| `E2E.P0.101` | real frontend、backend 与 Mailpit 已运行；新邮箱尚无已补全账号 | 用户登录、补全资料，点击设置齿轮核对姓名/完整邮箱，再退出并重新登录 | session/profile 状态真实持久化；Settings 显示同一 `/me` 账号字段而证据脱敏；logout 清 session；同邮箱重登进入已补全账号 | `e2e-scenarios-p0/001`；本 plan 只登记业务 handoff |

`E2E.P0.101` 原地增加真实 Settings 字段与 logout，不承接 shell pendingAction 通用矩阵、配置 wiring 或破坏性的账号删除。
