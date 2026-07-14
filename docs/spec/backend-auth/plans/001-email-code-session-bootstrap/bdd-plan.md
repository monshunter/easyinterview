# Backend Auth BDD Plan

> **版本**: 1.8
> **状态**: completed
> **更新日期**: 2026-07-14

**关联 Plan**: [plan](./plan.md)

## Domain behavior

| Behavior ID | Given | When | Then | 验证入口 |
|-------------|-------|------|------|----------|
| `BDD.AUTH.EMAIL.001` | 新邮箱或已有账号发起 email-code 登录 | challenge、verify、补全 profile、logout/relogin | session 与 profile-completion 状态按账号持久化，非法/重放请求 fail closed 且不泄露 code/cookie/email | `backend/internal/auth/email_code_session_contract_test.go` + `backend/cmd/api/auth_email_integration_test.go`，由根 `make test` 承接 |

## 当前真实 E2E handoff

| E2E ID | Given | When | Then | Owner |
|--------|-------|------|------|-------|
| `E2E.P0.101` | real frontend、backend 与 Mailpit 已运行；新邮箱尚无已补全账号 | 用户从单一邮箱入口获取并输入 6 位验证码，补全 displayName 与条款，再退出并重新登录 | 首次验证建立 session 且要求补全资料；补全后 `/me.profileCompletionRequired=false`；同邮箱再次登录进入已补全账号 | `e2e-scenarios-p0/001`；本 plan 只登记业务 handoff |

`E2E.P0.101` 只承接真实 email-code、session 与 profile-completion 链路，不承接 shell pending action、路由或配置 wiring。
