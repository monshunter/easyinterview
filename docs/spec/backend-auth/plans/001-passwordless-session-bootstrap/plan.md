# Passwordless Session Bootstrap

> **版本**: 1.0
> **状态**: active
> **更新日期**: 2026-05-05

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)
**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)
**关联 BDD Checklist**: [bdd-checklist](./bdd-checklist.md)

## 1 目标

实现 P0 后端认证最小闭环：邮箱 challenge 创建、challenge 验证、first-party session cookie、`/me`、logout，并为既有 A4 public runtime config handler 注入 session-aware resolver。该闭环支撑 `frontend-shell` 的操作级登录拦截与 pending action 恢复。

## 2 背景

ADR-Q1 已锁定自建 passwordless email magic link + first-party session cookie。B2 OpenAPI 已定义 Auth endpoints，B4 baseline 已包含 `auth_challenges` / `sessions` / `external_identities` 支撑表。当前缺少把这些 truth source 接成可运行后端 API 的实现计划。

## 3 质量门禁分类

- **Plan 类型**: `feature-behavior` + `backend` + `contract`。
- **TDD 策略**: 通过 `/implement backend-auth/001-passwordless-session-bootstrap backend` -> `/tdd` 执行；每个 checklist item 先写 focused Go test / handler contract test / store test，再实现最小代码；测试断言写在 checklist 的 `验证:` 后。
- **BDD 策略**: 需要 BDD。本 plan 引入认证 API 行为和 session cookie 工作流，必须维护 `bdd-plan.md`、`bdd-checklist.md`，并在主 checklist 中使用 `BDD-Gate:` 引用 `E2E.P0.003`。
- **替代验证 gate**: 不适用；BDD gate 是 API 行为验证入口。补充 gate 包括 focused Go tests、OpenAPI generated contract tests、migration dry-run / table contract、config lint、privacy log grep、`make docs-check`。

## 4 实施步骤

### Phase 1: Storage and config boundaries

#### 1.1 锁定 auth storage

复用 B4 baseline 中的 `users`、`auth_challenges`、`sessions` 表，建立 store 接口与 SQL 实现；不新增 migration，除非 B4 spec 先修订。

#### 1.2 锁定 config / secret 边界

从 A4 secrets/config 读取 `SESSION_COOKIE_SECRET`、`AUTH_CHALLENGE_TOKEN_PEPPER`、`EMAIL_PROVIDER`、`EMAIL_PROVIDER_API_KEY` 与固定 `ei_session` cookie name；缺必需配置或 secret 必须 fail-fast。challenge TTL、session TTL、rate-limit 和 dev mail sink 默认值由 C1 代码常量持有并在包级文档记录；若需要配置化，先停止并修订 A4 `secrets-and-config` spec / config truth source。

### Phase 2: Challenge issue and delivery

#### 2.1 实现 `startAuthEmailChallenge`

接收邮箱、创建 challenge token hash、记录 IP / UA hash，并通过 dev mail sink / dispatcher 边界发送一次性链接。

#### 2.2 实现 rate-limit / dedupe 基线

同一邮箱短时间重复请求必须返回稳定 accepted response，不泄露账号存在性。

### Phase 3: Verify, session, and current user

#### 3.1 实现 `verifyAuthEmailChallenge`

验证一次性 token，创建或读取 user，写入 server-side session，并通过 `Set-Cookie` 返回 opaque session。

#### 3.2 实现 `/me`

根据 session cookie 读取当前用户，返回 masked email、displayName、语言偏好等 B2 schema 字段；未登录返回 B1 error envelope。

#### 3.3 实现 logout

撤销 server-side session 并清除 cookie；重复 logout 幂等。

### Phase 4: Runtime config resolver, privacy, and observability

#### 4.1 接入 A4 `/runtime-config` session resolver

复用 A4 `NewRuntimeConfigHandler` 与 allowlist，只为其提供 C1 session-aware resolver；未登录路径保持 public response，有效 session 只能影响 A4 已允许公开的用户级偏好，不得扩大 response 字段。

#### 4.2 补隐私和可观测红线

日志、metric label、audit 不得包含 magic token、session cookie、完整邮箱、secret 或 PII。

### Phase 5: BDD and handoff

#### 5.1 执行 passwordless session BDD gate

按 `bdd-plan.md` 和 `bdd-checklist.md` 验证 email challenge -> verify -> `/me` -> logout 行为。

#### 5.2 Handoff 给 frontend-shell

记录前端可依赖的 Auth API、cookie 行为、mock/dev mail sink 和错误路径。

## 5 验收标准

- `startAuthEmailChallenge`、`verifyAuthEmailChallenge`、`getMe`、`logout`、A4 runtime-config session resolver integration 的 focused Go tests 通过。
- session cookie 符合 ADR-Q1 与 OpenAPI 描述；server-side session 是真理源。
- 未登录、过期 token、重复 verify、缺 cookie、logout 幂等、缺 secret 都有测试覆盖。
- 日志 / metric / audit privacy grep 无 secret / PII 明文。
- BDD-Gate `E2E.P0.003` 通过。

## 6 风险与应对

| 风险 | 应对措施 |
|------|----------|
| 认证实现绕过 B2 schema | handler tests 必须使用 generated OpenAPI types / server contract |
| token 或 session secret 进入日志 | Phase 4.2 privacy test 和 grep gate 强制覆盖 |
| P0 误引入 OAuth / SSO | spec Out of Scope 和 checklist negative search 拦截 |
| 缺少邮件基础设施阻塞本地验证 | P0 使用 dev mail sink / log-only delivery；生产邮件供应商另行计划 |
