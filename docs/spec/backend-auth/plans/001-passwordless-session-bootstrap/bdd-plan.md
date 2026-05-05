# Backend Auth BDD Plan

> **版本**: 1.0
> **状态**: active
> **更新日期**: 2026-05-05

## Phase 5: Passwordless session API behavior

| 场景 ID | 场景 | Given | When | Then | 验证入口 |
|---------|------|-------|------|------|----------|
| E2E.P0.003 | Passwordless session cookie | 用户使用有效邮箱请求 challenge，dev mail sink 可读取一次性链接 | 验证 challenge 后请求 `/me` 并执行 logout | 服务端签发 `ei_session`，`/me` 返回 masked user context，logout 清除 session 且重复调用幂等 | `test/scenarios/e2e/p0-003-passwordless-session-cookie/` |
