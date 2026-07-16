# Backend Auth BDD Plan

> **版本**: 1.12
> **状态**: completed
> **更新日期**: 2026-07-16

**关联 Plan**: [plan](./plan.md)

## Domain behavior

| Behavior ID | Given | When | Then | 验证入口 |
|-------------|-------|------|------|----------|
| `BDD.AUTH.EMAIL.001` | 新邮箱或已有账号发起 email-code 登录 | challenge、verify、读取最小 `/me`、补全 profile、logout/relogin | session/profile 状态按账号持久化；UserContext 只有 id/full email/display name/completion flag；完整 email 只在 authenticated response 中返回且不进入日志；非法/重放请求 fail closed | backend Auth contract/integration tests，由根 `make test` 承接 |
| `BDD.AUTH.EMAIL.002` | operator 选择合法 `mailpit` 或 `smtp` 配置；SMTP 凭据仅存在 secret source | 用户发起同一个 email-code challenge，runner 执行 `email_dispatch` | Mailpit 以本地无认证 SMTP 投递；标准 SMTP 必须在 STARTTLS 或隐式 TLS 后认证并投递；错误 fail closed 且不暴露凭据、完整邮箱或 raw code | backend Auth SMTP transport + cmd/api provider behavior tests，由根 `make test` 承接 |
| `BDD.AUTH.EMAIL.003` | 两个 backend 实例共享同一 Redis URL 和 challenge pepper；实例 A 创建 6 位验证码、challenge 与 job | 实例 B lease `email_dispatch` 并通过独立 Redis client 读取 delivery secret | 实例 B 解密并投递同一验证码；发送成功后删除；发送失败可重试至 5 分钟 TTL；Redis key/value、DB/job/error 不暴露 raw code/ref | backend Auth domain behavior tests + real Redis cross-client integration；不包装为 E2E |

## 当前真实 E2E handoff

| E2E ID | Given | When | Then | Owner |
|--------|-------|------|------|-------|
| `E2E.P0.101` | real frontend、backend 与 Mailpit 已运行；新邮箱尚无已补全账号 | 用户登录、补全资料，点击设置齿轮核对姓名/完整邮箱，再退出并重新登录 | session/profile 状态真实持久化；Settings 显示同一 `/me` 账号字段而证据脱敏；logout 清 session；同邮箱重登进入已补全账号 | `e2e-scenarios-p0/001`；本 plan 只登记业务 handoff |

`E2E.P0.101` 原地增加真实 Settings 字段与 logout，不承接 shell pendingAction 通用矩阵、配置 wiring 或破坏性的账号删除。
