# Passwordless Session Bootstrap Checklist

> **版本**: 1.0
> **状态**: active
> **更新日期**: 2026-05-05

**关联计划**: [plan](./plan.md)

## Phase 1: Storage and config boundaries

- [ ] 1.1 锁定 auth storage；验证: store tests 覆盖 `users`、`auth_challenges`、`sessions` 表读写，确认无需新增 migration；若需变更 schema，先停止并修订 B4 owner spec
- [ ] 1.2 锁定 config / secret 边界；验证: config tests 覆盖 `SESSION_COOKIE_SECRET`、`AUTH_CHALLENGE_TOKEN_PEPPER`、`EMAIL_PROVIDER`、`EMAIL_PROVIDER_API_KEY` 缺失时 fail-fast，固定 `ei_session` cookie name，且 challenge TTL / session TTL / rate-limit / dev mail sink 默认值作为 C1 代码常量有测试和包级文档；若需新增配置先停止并修订 A4

## Phase 2: Challenge issue and delivery

- [ ] 2.1 实现 `startAuthEmailChallenge`；验证: handler/service tests 覆盖 accepted response、token hash 入库、IP / UA hash、dev mail sink 收到脱敏 challenge link
- [ ] 2.2 实现 rate-limit / dedupe 基线；验证: tests 覆盖同邮箱重复请求不泄露账号存在性，响应仍符合 B2 schema

## Phase 3: Verify, session, and current user

- [ ] 3.1 实现 `verifyAuthEmailChallenge`；验证: tests 覆盖成功签发 `ei_session` cookie、过期 token、重复 verify、无效 token、session_hash 入库且不返回 cookie 明文
- [ ] 3.2 实现 `/me`；验证: handler tests 覆盖有效 session 返回 masked email / displayName / language，缺 cookie 或无效 session 返回 B1 error envelope
- [ ] 3.3 实现 logout；验证: tests 覆盖有效 session 撤销、重复 logout 幂等、Set-Cookie 清除和无账号存在性泄露

## Phase 4: Runtime config resolver, privacy, and observability

- [ ] 4.1 接入 A4 `/runtime-config` session resolver；验证: tests 断言 C1 只向 A4 handler 注入 session-aware resolver，未登录保持 public response，有效 session 只影响 A4 allowlist 内用户级偏好，secret / internal flag 不出 response
- [ ] 4.2 补隐私和可观测红线；验证: privacy tests / grep 确认日志、metric label、audit 不含 magic token、session cookie、完整邮箱、secret 或 PII 明文

## Phase 5: BDD and handoff

- [ ] 5.1 BDD-Gate: 验证 E2E.P0.003 通过
- [ ] 5.2 Handoff 给 frontend-shell；验证: backend README 或 package docs 说明 Auth API、cookie 行为、dev mail sink、错误码和前端 pendingAction 接入边界
- [ ] 5.3 active-scope 负向搜索通过；验证: backend active code 不引入 Bearer token P0 主认证、OAuth / SSO P0 行为、明文 token/session 存储或旧 AI gateway / voice route 口径
