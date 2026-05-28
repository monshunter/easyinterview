# Backend Auth BDD Plan

> **版本**: 1.3
> **状态**: completed
> **更新日期**: 2026-05-28

## Phase 5: Passwordless session API behavior

| 场景 ID | 场景 | Given | When | Then | 验证入口 |
|---------|------|-------|------|------|----------|
| E2E.P0.003 | Passwordless session cookie | 用户使用有效邮箱请求 challenge，C1 backend-internal mail dispatcher 可把一次性 6 位 code 写入 dev mail sink，场景可构造无效 / 重复 code、cookie jar 与 deleteMe idempotency key | 验证 challenge 后请求 `/me`、读取 `/runtime-config`、执行 logout 并重复调用 logout；在独立登录分支中重复调用 `DELETE /me` | 服务端签发 `ei_session`，`/me` 返回 masked user context，runtime-config 只返回 A4 allowlist，logout 清除 session 且重复调用幂等；`DELETE /me` 返回 B2 `202 + PrivacyRequestWithJob` 并对重复请求复用同一 active 删除请求或同义终态；无效 / 重复 code 与 logout 后 `/me` 返回 B1 error envelope；日志 / payload 无 secret / PII 明文；不启动独立 worker 进程也能完成邮件读取 | `test/scenarios/e2e/p0-003-passwordless-session-cookie/` |
| E2E.P0.101 | Auth email-code single-entry login + profile setup | Real frontend/backend/Mailpit dev stack 可用；邮箱是唯一账号标识；displayName 不唯一；新邮箱尚未补全 profile | 用户从单一登录入口提交新邮箱，从 Mailpit 读取 6 位 code 并在前端 `auth_verify` 输入；未补全时关闭浏览器或换浏览器后用同一邮箱重新登录；随后调用资料补全；再退出并用同一邮箱登录 | start 阶段不泄露邮箱是否存在；新邮箱 verify 后创建唯一用户和 session，`/me.profileCompletionRequired=true`；未补全用户每次重新登录仍返回资料未补全；`PATCH /me` 写入 displayName / terms / completion 后 `/me.profileCompletionRequired=false`；后续同邮箱登录直接进入正常账号；邮件正文不包含 magic link 或 `/auth/verify?token=`；界面不出现注册入口或 prototype fallback | `test/scenarios/e2e/p0-101-auth-email-code-login-register/` |
