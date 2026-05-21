# Backend Auth BDD Plan

> **版本**: 1.1
> **状态**: completed
> **更新日期**: 2026-05-21

## Phase 5: Passwordless session API behavior

| 场景 ID | 场景 | Given | When | Then | 验证入口 |
|---------|------|-------|------|------|----------|
| E2E.P0.003 | Passwordless session cookie | 用户使用有效邮箱请求 challenge，C1 backend-internal mail dispatcher 可把一次性链接写入 dev mail sink，场景可构造无效 / 重复 token、cookie jar 与 deleteMe idempotency key | 验证 challenge 后请求 `/me`、读取 `/runtime-config`、执行 logout 并重复调用 logout；在独立登录分支中重复调用 `DELETE /me` | 服务端签发 `ei_session`，`/me` 返回 masked user context，runtime-config 只返回 A4 allowlist，logout 清除 session 且重复调用幂等；`DELETE /me` 返回 B2 `202 + PrivacyRequestWithJob` 并对重复请求复用同一 active 删除请求或同义终态；无效 / 重复 token 与 logout 后 `/me` 返回 B1 error envelope；日志 / payload 无 secret / PII 明文；不启动独立 worker 进程也能完成邮件读取 | `test/scenarios/e2e/p0-003-passwordless-session-cookie/` |
