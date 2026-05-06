# Passwordless Session Bootstrap

> **版本**: 1.4
> **状态**: active
> **更新日期**: 2026-05-06

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)
**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)
**关联 BDD Checklist**: [bdd-checklist](./bdd-checklist.md)

## 1 目标

实现 P0 后端认证最小闭环：邮箱 challenge 创建、C1 backend-internal mail dispatcher（Go goroutine / 后台线程）与 dev mail sink、challenge 验证、first-party session cookie、protected session middleware、`/me`、`DELETE /me` idempotent auth handoff、logout，并为既有 A4 public runtime config handler 注入 session-aware resolver。该闭环支撑 `frontend-shell` 的操作级登录拦截与 pending action 恢复。

## 2 背景

ADR-Q1 已锁定自建 passwordless email magic link + first-party session cookie。B2 OpenAPI 已定义 Auth endpoints，generated Go `ServerInterface` 当前包含 `startAuthEmailChallenge`、`verifyAuthEmailChallenge`、`getMe`、`deleteMe`、`logout` 和 `getRuntimeConfig`；B4 baseline 已包含 `auth_challenges` / `sessions` / `external_identities` 支撑表；B3 已冻结 internal-only `email_dispatch` job payload 红线。当前 C8 `backend-async-runtime` 仍未创建，且产品尚无真实用户；本 plan 明确采用 C1 backend 进程内后台派发器作为过渡实现，不把独立 worker 进程作为本地 BDD 或 P0 auth 闭环前置。

## 3 质量门禁分类

- **Plan 类型**: `feature-behavior` + `backend` + `contract`。
- **TDD 策略**: 通过 `/implement backend-auth/001-passwordless-session-bootstrap backend` -> `/tdd` 执行；每个 checklist item 先写 focused Go test / handler contract test / store test，再实现最小代码；测试断言写在 checklist 的 `验证:` 后。
- **BDD 策略**: 需要 BDD。本 plan 引入认证 API 行为和 session cookie 工作流，必须维护 `bdd-plan.md`、`bdd-checklist.md`，并在主 checklist 中使用 `BDD-Gate:` 引用 `E2E.P0.003`。
- **替代验证 gate**: 不适用；BDD gate 是 API 行为验证入口。补充 gate 包括 focused Go tests、OpenAPI generated contract tests、migration dry-run / table contract、config lint、B3 `email_dispatch` helper tests、backend-internal mail dispatcher drain/shutdown tests、F1 auth metric registry preflight、privacy log/audit grep、auth metric label checks、`make docs-check`。

## 4 实施步骤

### Phase 1: Storage and config boundaries

#### 1.1 锁定 auth storage

复用 B4 baseline 中的 `users`、`user_settings`、`auth_challenges`、`sessions` 表，建立 store 接口与 SQL 实现；确认 `external_identities` 仅作为 P1 SSO 空表槽存在，本 plan 不读写该表、不暴露 store 方法。不新增 migration，除非 B4 spec 先修订。滑动续期使用当前 `sessions.updated_at` 作为 last-seen touch；如实现需要独立 `last_seen_at` 字段，先停止并修订 ADR-Q1 / B4。

#### 1.2 锁定 config / secret 边界

从 A4 secrets/config 读取 `SESSION_COOKIE_SECRET`、`AUTH_CHALLENGE_TOKEN_PEPPER`、`EMAIL_PROVIDER`、`EMAIL_PROVIDER_API_KEY` 与固定 `ei_session` cookie name；缺必需配置或 secret 必须 fail-fast。challenge TTL 固定 15 分钟、session TTL 固定 30 天、同邮箱或同 IP 1 分钟第 3 次及以上触发 rate-limit / dedupe，dev mail sink 默认值由 C1 代码常量持有并在包级文档记录；若需要配置化，先停止并修订 A4 `secrets-and-config` spec / config truth source。

#### 1.3 锁定 generated Auth surface 和 session middleware

使用 B2 generated `ServerInterface` / operation registry 接入 Auth endpoints。`POST /auth/email/start`、`GET /auth/email/verify`、`GET /runtime-config` 保持 public；`POST /auth/logout` 走 optional-session / always-clear-cookie 路径，缺 cookie、无效 session 或重复调用都必须进入 handler 并保持幂等；其余 protected endpoint 通过 C1 first-party session middleware / current-user resolver 校验。`DELETE /me` 由 C1 完成认证、session 撤销和 privacy_delete handoff，实际删除执行仍归 C8 / B4。

### Phase 2: Challenge issue and delivery

#### 2.1 实现 `startAuthEmailChallenge`

接收邮箱、创建 challenge token hash、记录 IP / UA hash，并通过 C1 backend-internal mail dispatcher 发送一次性链接到 dev mail sink / provider adapter。明文 token 只允许短暂存在于一次性发送边界；dev/test 读取必须通过 mail sink retrieval，不允许 log-only token delivery。

#### 2.2 实现 rate-limit / dedupe 基线

同一邮箱或同一 IP 1 分钟内第 3 次及以上请求必须返回稳定 accepted / rate-limited response，不泄露账号存在性；dedupe key 不得包含邮箱明文。

#### 2.3 接入 `email_dispatch` redacted payload

派发邮件时必须使用 B3 generated `BuildEmailDispatchPayload` helper 构造过渡期派发输入，payload 仅包含 `authChallengeId` / `userId` / `templateKey` / `locale` / `deliverySecretRef` / `dedupeKey`。`rawMagicLinkToken`、`magicLinkUrl`、`recipientEmail`、`recipientEmailHash`、`emailBody`、`emailSubject` 等 redacted 字段不得进入 in-process queue、dev sink、future outbox、async job payload、日志或 audit。

#### 2.4 实现 backend-internal mail dispatcher

在 backend API 进程内实现 C1-owned 后台派发器：handler 只入队并返回 B2 `202` accepted，后台 goroutine / 线程 drain 队列并写入 dev mail sink 或邮件 provider adapter；测试覆盖不启动 C8 worker 进程时仍可读取 challenge link、队列关闭 / graceful shutdown、派发失败不会泄露邮箱或 token。后续切换到 C8 worker 时必须沿用 2.3 的 B3 payload redaction contract。

### Phase 3: Verify, session, and current user

#### 3.1 实现 `verifyAuthEmailChallenge`

验证一次性 token，创建或读取 user，写入 server-side session，并通过 `Set-Cookie` 返回 opaque session。

#### 3.2 实现 session middleware / current-user resolver

读取并校验 `ei_session`，用 server-side `sessions` 表作为真理源，支持 active / revoked / expired 状态；缺 cookie、无效 session、过期 session 返回 B1 error envelope。续期只更新 `sessions.updated_at` / expiry，不把 cookie 明文写入日志或 response。

#### 3.3 实现 `/me`

根据 session cookie 读取当前用户，返回 masked email、displayName、语言偏好等 B2 schema 字段；未登录返回 B1 error envelope。

#### 3.4 实现 logout

撤销 server-side session 并清除 cookie；缺 cookie、无效 session 和重复 logout 都必须返回可预期结果，不泄露账号状态。

#### 3.5 实现 `deleteMe` auth handoff

`DELETE /me` 复用 C1 session middleware 获取当前用户，支持 B2 `Idempotency-Key` 或等价 active-request dedupe，撤销当前 session，并按 B2 `202 + PrivacyRequestWithJob` contract handoff 给 C8 `privacy_delete` owner；相同 idempotency key 或同一 active 删除请求不得创建重复 `privacy_delete` job。本 plan 不实现隐私删除 worker 或 B4 删除矩阵。

### Phase 4: Runtime config resolver, privacy, and observability

#### 4.1 接入 A4 `/runtime-config` session resolver

复用 A4 `NewRuntimeConfigHandler` 与 allowlist，只为其提供 C1 session-aware resolver；未登录路径保持 public response，有效 session 只能影响 A4 已允许公开的用户级偏好，不得扩大 response 字段。

#### 4.2 补隐私和可观测红线

日志、metric label、audit 不得包含 magic token、session cookie、完整邮箱、secret 或 PII。

#### 4.3 接入 auth metrics / audit 最小事件

按 ADR-Q1/F1 口径记录 `auth_challenge_started_total`、`auth_session_minted_total`、`auth_failure_total` 等 auth metrics；实施前必须先确认指标已登记到 F1 baseline metrics 字典，若未登记则先停止并修订 F1 owner spec / plan。label 只能使用 F1 allowed labels。challenge started、session minted、logout、delete handoff、failure audit 只记录 ID / hash / 状态与 trace，不记录邮箱明文、token、session id 或 URL。

### Phase 5: BDD and handoff

#### 5.1 执行 passwordless session BDD gate

按 `bdd-plan.md` 和 `bdd-checklist.md` 验证 email challenge -> backend-internal mail dispatcher -> dev mail sink retrieval -> verify -> `/me` -> runtime-config session resolver -> logout -> repeated logout / deleteMe idempotency / error paths 行为。

#### 5.2 Handoff 给 frontend-shell

记录前端可依赖的 Auth API、cookie 行为、mock/dev mail sink 和错误路径。

### Phase 6: L2 remediation

#### 6.1 修复 `DELETE /me` idempotency user scope

`DELETE /me` 的 active `privacy_delete` handoff 去重必须按当前用户隔离。相同 `Idempotency-Key` 只能复用同一用户的 active 删除请求，不得让其他用户拿到既有用户的 `privacy_request` / job。实现前补两用户同 key 的 store / handler 回归测试。

#### 6.2 修复 runtime Auth wiring

真实 `cmd/api` runtime 必须挂载 C1 Auth endpoints，并用 C1 session middleware 包装 generated Auth surface 与 protected operation policy；`/runtime-config` 必须接入 C1 session-aware resolver，而不是只提供 anonymous resolver。实现前补 `cmd/api` wiring 测试证明 auth routes mounted、protected `/me` 需要 session、runtime-config 使用 resolver。

#### 6.3 修复 session cookie Secure policy

`verifyAuthEmailChallenge` 和 logout / deleteMe 清 cookie 必须使用同一 cookie policy，保证 ADR-Q1 / OpenAPI 锁定的 `HttpOnly; Secure; SameSite=Lax` 可测试；dev 可按环境降级 Secure，但必须显式测试。实现前补 secure cookie attribute 回归测试。

#### 6.4 修复 challenge rate-limit SQL scope

同邮箱或同 IP 1 分钟第 3 次及以上请求的 SQL 统计必须覆盖最近 challenge attempt，而不是只统计 pending challenge。实现前补 SQL expectation / service 回归测试，证明 consumed challenge 也计入窗口。

#### 6.5 修复 logout revoke failure response

logout 撤销 server-side session 失败时不得静默返回成功。handler 可以继续清 cookie，但必须返回 B1 error envelope，并记录 redacted failure evidence。实现前补 store revoke failure 的 handler 回归测试。

#### 6.6 修复 runtime Auth secret fail-fast

真实 `cmd/api` runtime 构建 `PasswordlessService` 前必须确认 `AUTH_CHALLENGE_TOKEN_PEPPER` 与 `SESSION_COOKIE_SECRET` 非空。dev / test 可以由外部 init 或 env 注入，但不得在缺失 secret 时继续启动并写入可预测 challenge / session hash。实现前补 `cmd/api` builder 回归测试，断言空 secret 返回明确错误且不启动 mail dispatcher。

#### 6.7 修复 logout optional-session resolver error

`logout` 仍保持 optional-session：缺 cookie、invalid / expired / revoked session 可以进入幂等清 cookie路径；但 session resolver / store 错误必须返回 B1 error envelope，不得被 middleware 当作匿名 logout 静默吞掉。实现前补真实 `cmd/api` route 回归测试，断言 cookie-bearing logout 遇到 store error 返回 500 且不泄露 cookie / session id。

#### 6.8 修复 logout revoke race 的 touch zero-row 归类

并发 logout 可能让 `ResolveSession` 先读到 active session，随后 `TouchSession` 因另一个请求已 revoke 而返回零行。该场景必须按已知认证状态失效处理，optional logout 继续走幂等清 cookie，而不是返回 500。实现前补 middleware focused test，断言 `TouchSession` 返回 `sql.ErrNoRows` 时 `logout` optional path 仍进入 handler。

## 5 验收标准

- `startAuthEmailChallenge`、backend-internal mail dispatcher、`verifyAuthEmailChallenge`、session middleware、`getMe`、`deleteMe` idempotent auth handoff、`logout`、A4 runtime-config session resolver integration 的 focused Go tests 通过。
- session cookie 符合 ADR-Q1 与 OpenAPI 描述；server-side session 是真理源。
- 未登录、过期 token、重复 verify、无效 token、缺 cookie、revoked/expired session、缺 cookie logout、logout 幂等、delete handoff idempotency、缺 secret 都有测试覆盖。
- B3 `email_dispatch` payload helper gate 通过，raw token / URL / 邮箱明文 / 邮件正文不会进入 in-process queue、dev sink、future outbox、async payload、log 或 audit。
- 日志 / metric / audit privacy grep 无 secret / PII 明文；auth metrics 名称已由 F1 登记或承接，label 只使用 F1 allowed labels。
- BDD-Gate `E2E.P0.003` 通过。

## 6 风险与应对

| 风险 | 应对措施 |
|------|----------|
| 认证实现绕过 B2 schema | handler tests 必须使用 generated OpenAPI types / server contract |
| token 或 session secret 进入日志 | Phase 4.2 privacy test 和 grep gate 强制覆盖 |
| P0 误引入 OAuth / SSO | spec Out of Scope 和 checklist negative search 拦截 |
| 过早依赖独立 C8 worker 阻塞本地验证 | 本 plan 在 C1 内实现 backend-internal goroutine / 后台线程派发器与 dev mail sink；C8 worker 是未来替换边界，不是 P0 auth 闭环前置 |
| C1 绕过 B3 email payload 红线 | Phase 2.3 强制使用 generated `BuildEmailDispatchPayload`，并用 negative tests 拒绝 redacted fields |
| C1 抢占 privacy deletion 执行 | Phase 3.5 只做 auth/session handoff，C8 / B4 删除执行不进入本 plan |
| C1 新增未登记 auth metric | Phase 4.3 先跑 F1 registry preflight；未登记则先修订 F1，不在 C1 私造 metric |
