# Backend Auth Spec

> **版本**: 1.0
> **状态**: active
> **更新日期**: 2026-05-05

## 1 背景与目标

`backend-auth` 承接 ADR-Q1 的 P0 后端认证实现：自建 passwordless email magic link / challenge + first-party session cookie。它为 `frontend-shell` 的操作级登录拦截、pending action 恢复和用户菜单提供后端支撑。

本 subject 的目标是落地最小可用认证后端，同时保持 product-scope 的隐私红线和 B2 OpenAPI 契约。

## 2 范围

### 2.1 In Scope

- `POST /api/v1/auth/email/start` 邮箱挑战创建。
- `GET /api/v1/auth/email/verify` 邮箱挑战验证并签发 first-party session cookie。
- `GET /api/v1/me` 当前用户读取。
- `POST /api/v1/auth/logout` 清除 session。
- 为既有 `GET /api/v1/runtime-config` 注入 C1 session-aware resolver，供 A4 handler 合并用户级公开偏好。
- 本地 dev 邮件 sink / log-only challenge delivery 的可观测与脱敏。
- session cookie 属性、安全默认值、过期、幂等 logout、错误码映射。

### 2.2 Out of Scope

- 不实现 OAuth / SSO / 企业账号体系。
- 不实现 Team / EDU、订阅或计费能力。
- 不实现完整隐私导出；P0 隐私导出延后，删除能力按 product-scope / B4 owner 另行计划。
- 不在日志中输出验证码、magic link token、完整邮箱、session secret 或 PII。
- 不修改 B2 Auth operation shape；需要变更时先修订 `openapi-v1-contract`。

## 3 用户决策 / 待确认事项

| ID | 决策 | 锁定值 | 影响 |
|----|------|--------|------|
| D-1 | 认证方案 | passwordless email challenge + first-party session cookie | 不使用 Bearer token 作为 P0 浏览器主认证形态 |
| D-2 | 登录入口 | 操作级 gate | 后端不要求首页加载前认证 |
| D-3 | Cookie 安全 | HttpOnly、SameSite、Secure 按环境配置 | dev 可降级 Secure，但必须可测试 |
| D-4 | 配置来源 | A4 secrets/config 只提供 `SESSION_COOKIE_SECRET`、`AUTH_CHALLENGE_TOKEN_PEPPER`、`EMAIL_PROVIDER`、`EMAIL_PROVIDER_API_KEY` 和固定 `ei_session` cookie name；TTL、rate-limit 与 dev mail sink 默认值由 C1 代码常量持有并在包级文档记录 | 邮件、cookie、session secret 不私造配置 key；如需把 TTL / mail sink 变成配置，先修订 A4 |
| D-5 | 错误码 | B1 shared error envelope | 认证错误必须使用共享错误 shape |

## 4 设计约束

- Challenge token 必须 hash 后存储或在本地 stub 中以不可逆形式比较；明文 token 只能在一次性发送边界短暂存在。
- Session ID / secret 不得进入日志、metrics label、audit 明文字段或 API response。
- Logout 必须幂等；没有有效 session 时返回可预期结果，不泄露账号存在性。
- `/runtime-config` 由 A4 handler 持有公开 allowlist；C1 只能提供 session-aware resolver，不得扩大 response 字段。
- `/me` 未登录必须返回 B2 / B1 约定的认证错误，不返回假用户。
- 邮箱挑战发送失败、挑战过期、重复验证、缺 cookie、无效 session 都必须有可测试错误路径。

## 5 模块边界

| 边界 | Owner | 说明 |
|------|-------|------|
| API contract | B2 `openapi-v1-contract` | Auth endpoints、response schema、cookie 描述 |
| backend auth | `backend-auth` | handlers、service、store、session、challenge delivery |
| config/secrets | A4 `secrets-and-config` | session secret、challenge pepper、email provider secret、固定 `ei_session` cookie name；TTL / dev mail sink 默认值归 C1 代码常量，新增配置前先修订 A4 |
| frontend gate | `frontend-shell` | pendingAction、登录页面和登录后恢复 |
| DB/session storage | B4 `db-migrations-baseline` | session/challenge 表或等价持久化边界 |

## 6 验收标准

| ID | 场景 | Given | When | Then | 对应 Plan |
|----|------|-------|------|------|-----------|
| C-1 | Passwordless session | 用户请求邮箱挑战 | 验证 challenge | 返回 first-party session cookie，随后 `/me` 返回当前用户 | 001-passwordless-session-bootstrap |
| C-2 | Logout 幂等 | 用户已有或没有有效 session | 调用 logout | cookie 被清除且响应不泄露账号状态 | 001-passwordless-session-bootstrap |
| C-3 | 错误路径 | challenge 过期、重复验证、缺 cookie、配置缺失 | 调用对应 endpoint | 返回 B1 error envelope，日志无 secret / PII 明文 | 001-passwordless-session-bootstrap |
| C-4 | Runtime config session resolver | 前端启动，用户可能携带有效 session | 请求 `/runtime-config` | A4 handler 仍只返回公开 allowlist 字段；C1 session resolver 只影响允许公开的用户级偏好，不泄露 secret / internal flag | 001-passwordless-session-bootstrap |

## 7 关联计划

- [001-passwordless-session-bootstrap](./plans/001-passwordless-session-bootstrap/plan.md)
