# Frontend Shell Auth and Settings BDD Plan

> **版本**: 1.17
> **状态**: completed
> **更新日期**: 2026-07-14

**关联 Plan**: [plan](./plan.md)

## 有真实 E2E owner 的行为

| 场景 | Given | When | Then | 真实 E2E |
|---|---|---|---|---|
| Email-code 登录与资料补全 | real frontend、backend 与 Mailpit 已运行 | 用户获取验证码、登录、补全资料并重新登录 | shell 读取真实 session 与 profile 状态，补全后展示已登录用户 | `E2E.P0.101` |

## Domain behavior

| Behavior ID | Given | When | Then | 验证入口 |
|-------------|-------|------|------|----------|
| `BDD.SHELL.AUTH.001` | 用户处于 anonymous、profile-incomplete 或 authenticated 状态 | 访问受保护页面、完成登录/profile setup、恢复 pending action 或切换 shell 设置 | shell 按 route/session 状态渲染并安全恢复；显示设置不持久化业务事实 | `frontend/src/app/AppAuthDispatch.test.tsx` + `frontend/src/app/__tests__/auth-pending-action-resume.test.tsx`，由根 `make test` 承接 |

`E2E.P0.101` 是 email-code/profile setup 的独立 suite handoff；pending action、通用 guard 和 settings 不归入该 E2E。
