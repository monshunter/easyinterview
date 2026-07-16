# Email-Code Session Bootstrap

> **版本**: 2.7
> **状态**: completed
> **更新日期**: 2026-07-16

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)
**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)
**关联 BDD Checklist**: [bdd-checklist](./bdd-checklist.md)

## 1 目标

实现 P0 后端认证最小闭环：邮箱 challenge 创建、`email_dispatch` async job producer + runner handler delivery、dev mail sink / Mailpit delivery writer、challenge 验证、first-party session cookie、protected session middleware、`/me`、`DELETE /me` idempotent auth handoff、logout，并为既有 A4 public runtime config handler 注入 session-aware resolver。该闭环支撑 `frontend-shell` 的操作级登录拦截与 pending action 恢复。

## 2 背景

ADR-Q1 已锁定自建 email-code challenge + first-party session cookie。B2 OpenAPI 已定义 Auth endpoints，generated Go `ServerInterface` 当前包含 `startAuthEmailChallenge`、`verifyAuthEmailChallenge`、`getMe`、`deleteMe`、`logout` 和 `getRuntimeConfig`；B4 baseline 已包含 `auth_challenges` / `sessions` / `external_identities` 支撑表；B3 已冻结 internal-only `email_dispatch` job payload 红线。`backend-runtime-topology` 已锁定 P0 不保留独立 worker 进程；当前实现通过 C1 `EmailDispatchEnqueuer` 写入 `async_jobs(email_dispatch)`，再由 backend-async-runner in-process kernel 调用 `EmailDispatchHandler` 投递 code-only 邮件。当前登录挑战为 6 位数字验证码、5 分钟有效期和 code-only 邮件；邮箱是唯一账号标识，displayName 不唯一且不参与账号唯一性判断。注册和登录已合并为单一邮箱验证码入口：发码前不区分 purpose、不泄露账号存在性；新邮箱 verify 后创建资料未补全账号，`/me.profileCompletionRequired` 驱动前端资料补全，`PATCH /me` 完成 displayName + 条款确认。

## 3 质量门禁分类

- **Plan 类型**: `feature-behavior` + `backend` + `contract`。
- **TDD 策略**: 通过 `/implement backend-auth/001-email-code-session-bootstrap backend` -> `/tdd` 执行；focused Go test / handler contract test / store test 只用于开发反馈，阶段完成由仓库根 `make test` 承接前后端全量单测。
- **BDD 策略**: `BDD.AUTH.EMAIL.001` 由代码层 owner tests 验证 email-code、session 与 profile-completion，`BDD.AUTH.EMAIL.002` 验证 Mailpit/SMTP provider、TLS/auth、失败与隐私行为，`BDD.AUTH.EMAIL.003` 验证 producer/consumer 位于不同 backend 实例时仍能通过共享 Redis 投递同一 6 位验证码；三者由仓库根 `make test` 统一回归，真实 Redis 跨 client 由独立 integration gate 承接。`E2E.P0.101` 仅作为 real frontend/backend/Mailpit 链路的独立 handoff，只有显式真实运行后才产生 PASS；外部 SMTP 使用显式脱敏 live smoke，不创建配置型 E2E。OpenAPI contract、config lint、privacy grep 与 docs checks 作为独立 gate，不包装为 E2E。

### 3.1 Operation Matrix

| operationId | fixture | frontend consumer | backend handler | persistence | AI dependency | scenario coverage |
|-------------|---------|-------------------|-----------------|-------------|---------------|-------------------|
| `getMe` | `Auth/getMe.json` | AppRuntimeProvider、Settings、auth/profile guards | current-user handler | users + session；analytics opt-in internal read | none | auth/settings domain + extended `E2E.P0.101` |
| `completeMyProfile` | `Auth/completeMyProfile.json` | AuthProfileSetupScreen | profile completion handler | users display/profile/terms | none | `BDD.AUTH.EMAIL.001` + `E2E.P0.101` |
| `deleteMe` | `Auth/deleteMe.json` | Settings destructive confirmation | account delete handoff | user soft delete、all-session revoke、privacy job | none | backend contract + `BDD.SHELL.SETTINGS.DELETE.001` |
| `logout` | `Auth/logout.json` | AuthLogoutScreen from Settings | optional-session logout handler | session revocation | none | auth/settings domain + `E2E.P0.101` |

## 4 实施步骤

### Phase 1: Storage and config boundaries

#### 1.1 锁定 auth storage

复用 B4 baseline 中的 `users`、`user_settings`、`auth_challenges`、`sessions` 表，建立 store 接口与 SQL 实现；确认 `external_identities` 仅作为 P1 SSO 空表槽存在，本 plan 不读写该表、不暴露 store 方法。不新增 migration，除非 B4 spec 先修订。滑动续期使用当前 `sessions.updated_at` 作为 last-seen touch；如实现需要独立 `last_seen_at` 字段，先停止并修订 ADR-Q1 / B4。

#### 1.2 锁定 config / secret 边界

从 A4 secrets/config 读取 `SESSION_COOKIE_SECRET`、`AUTH_CHALLENGE_TOKEN_PEPPER`、`EMAIL_PROVIDER`、标准 SMTP 配置与固定 `ei_session` cookie name；缺必需配置或 secret 必须 fail-fast。challenge TTL 固定 5 分钟、session TTL 固定 30 天、同邮箱或同 IP 1 分钟第 3 次及以上触发 rate-limit / dedupe，dev mail sink 默认值由 C1 代码常量持有并在包级文档记录；若需要配置化，先停止并修订 A4 `secrets-and-config` spec / config truth source。

#### 1.3 锁定 generated Auth surface 和 session middleware

使用 B2 generated `ServerInterface` / operation registry 接入 Auth endpoints。`POST /auth/email/start`、`GET /auth/email/verify`、`GET /runtime-config` 保持 public；`POST /auth/logout` 走 optional-session / always-clear-cookie 路径，缺 cookie、无效 session 或重复调用都必须进入 handler 并保持幂等；其余 protected endpoint 通过 C1 first-party session middleware / current-user resolver 校验。`DELETE /me` 由 C1 完成认证、session 撤销和 privacy_delete handoff，实际删除执行仍归 backend internal runner / B4。

### Phase 2: Challenge issue and delivery

#### 2.1 实现 `startAuthEmailChallenge`

接收邮箱和可选 `returnTo`，创建 6 位数字 challenge code hash、记录 IP / UA hash，并通过 `EmailDispatchEnqueuer` 写入 `email_dispatch` async job，由 runner handler 发送一次性验证码到 dev mail sink / Mailpit provider adapter。明文 code 只允许短暂存在于一次性发送边界；dev/test 读取必须通过 mail sink retrieval，不允许 log-only code delivery。发码阶段不区分注册/登录，不接收或持久化 displayName，也不得泄露邮箱是否已存在。

#### 2.2 实现 rate-limit / dedupe 基线

同一邮箱或同一 IP 1 分钟内第 3 次及以上请求必须返回稳定 accepted / rate-limited response，不泄露账号存在性；dedupe key 不得包含邮箱明文。

#### 2.3 接入 `email_dispatch` redacted payload

派发邮件时必须使用 B3 generated `BuildEmailDispatchPayload` helper 构造派发输入，payload 仅包含 `authChallengeId` / `userId` / `templateKey` / `locale` / `deliverySecretRef` / `dedupeKey`。`rawEmailCode`、`emailVerificationUrl`、`recipientEmail`、`recipientEmailHash`、`emailBody`、`emailSubject` 等 redacted 字段不得进入 async job payload、日志或 audit。

#### 2.4 实现 `email_dispatch` producer / handler delivery

`startAuthEmailChallenge` 只写入 `async_jobs(job_type='email_dispatch')` 并返回 B2 `202` accepted；backend-async-runner kernel lease 后调用 `auth.EmailDispatchHandler`，由 `DeliveryWriter` 写入 dev mail sink 或 Mailpit SMTP adapter。测试覆盖不启动独立后台执行进程时仍可读取 6 位验证码、runner shutdown / graceful drain 路径可观测、派发失败不会泄露邮箱或 code。

### Phase 3: Verify, session, and current user

#### 3.1 实现 `verifyAuthEmailChallenge`

验证一次性 6 位验证码后按 normalized email 查找用户：既有邮箱直接登录，新邮箱创建资料未补全账号和默认 user_settings。成功后写入 server-side session，并通过 `Set-Cookie` 返回 opaque session；新账号的 displayName 与条款确认只能由 `PATCH /me` 完成。

#### 3.2 实现 session middleware / current-user resolver

读取并校验 `ei_session`，用 server-side `sessions` 表作为真理源，支持 active / revoked / expired 状态；缺 cookie、无效 session、过期 session 返回 B1 error envelope。续期只更新 `sessions.updated_at` / expiry，不把 cookie 明文写入日志或 response。

#### 3.3 实现 `/me`

根据 session cookie 读取当前用户，返回完整 email、displayName 等当前 B2 schema 字段；未登录返回 B1 error envelope。完整 email 仅进入 authenticated response，不写日志。

#### 3.4 实现 logout

撤销 server-side session 并清除 cookie；缺 cookie、无效 session 和重复 logout 都必须返回可预期结果，不泄露账号状态。

#### 3.5 实现 `deleteMe` auth handoff

`DELETE /me` 复用 C1 session middleware 获取当前用户，支持 B2 `Idempotency-Key` 或等价 active-request dedupe，撤销当前 session，并按 B2 `202 + PrivacyRequestWithJob` contract handoff 给 backend internal runner / `privacy_delete` owner；相同 idempotency key 或同一 active 删除请求不得创建重复 `privacy_delete` job。本 plan 不实现隐私删除 runner 或 B4 删除矩阵。

### Phase 4: Runtime config resolver, privacy, and observability

#### 4.1 接入 A4 `/runtime-config` session resolver

复用 A4 `NewRuntimeConfigHandler` 与 allowlist，只为其提供 C1 session-aware resolver；未登录路径保持 public response，有效 session 只能影响 A4 已允许公开的用户级偏好，不得扩大 response 字段。

#### 4.2 补隐私和可观测红线

日志、metric label、audit 不得包含 raw challenge secret、session cookie、完整邮箱、secret 或 PII。

#### 4.3 接入 auth metrics / audit 最小事件

按 ADR-Q1/F1 口径记录 `auth_challenge_started_total`、`auth_session_minted_total`、`auth_failure_total` 等 auth metrics；实施前必须先确认指标已登记到 F1 baseline metrics 字典，若未登记则先停止并修订 F1 owner spec / plan。label 只能使用 F1 allowed labels。challenge started、session minted、logout、delete handoff、failure audit 只记录 ID / hash / 状态与 trace，不记录邮箱明文、token、session id 或 URL。

### Phase 5: BDD and handoff

#### 5.1 完成代码级 BDD 并登记真实 E2E handoff

按 `bdd-plan.md` 和 `bdd-checklist.md` 由 `BDD.AUTH.EMAIL.001` 关联 email challenge -> verify -> `/me` -> profile completion -> logout/relogin 的代码层行为证据。`E2E.P0.101` 只登记真实 Mailpit 主链路；本轮当前 run `e2e-p0-101-20260715114513-19516` 已 PASS，静态 INDEX 生命周期状态保持 `Ready`。runtime-config wiring、deleteMe idempotency 和错误矩阵由代码层 gate 承接，不扩张该 E2E。

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

真实 `cmd/api` runtime 构建 `EmailCodeService` 前必须确认 `AUTH_CHALLENGE_TOKEN_PEPPER` 与 `SESSION_COOKIE_SECRET` 非空。dev / test 可以由外部 init 或 env 注入，但不得在缺失 secret 时继续启动并写入可预测 challenge / session hash。实现前补 `cmd/api` builder 回归测试，断言空 secret 返回明确错误且不启动 mail dispatcher。

#### 6.7 修复 logout optional-session resolver error

`logout` 仍保持 optional-session：缺 cookie、invalid / expired / revoked session 可以进入幂等清 cookie路径；但 session resolver / store 错误必须返回 B1 error envelope，不得被 middleware 当作匿名 logout 静默吞掉。实现前补真实 `cmd/api` route 回归测试，断言 cookie-bearing logout 遇到 store error 返回 500 且不泄露 cookie / session id。

#### 6.8 修复 logout revoke race 的 touch zero-row 归类

并发 logout 可能让 `ResolveSession` 先读到 active session，随后 `TouchSession` 因另一个请求已 revoke 而返回零行。该场景必须按已知认证状态失效处理，optional logout 继续走幂等清 cookie，而不是返回 500。实现前补 middleware focused test，断言 `TouchSession` 返回 `sql.ErrNoRows` 时 `logout` optional path 仍进入 handler。

### Phase 8: Unified email login and profile completion

#### 8.1 OpenAPI contract and generated clients

把 `AuthEmailStartRequest` 收敛为单入口发码请求：`email` 必填，`purpose` / `displayName` 从请求 schema、fixtures、generated Go / TS types 和正式前端调用中移除。`UserContext` 增加必填 `profileCompletionRequired`。新增 `PATCH /api/v1/me` operationId `completeMyProfile`，request body 包含 trimmed `displayName` 与 `acceptedTerms=true`，response 返回更新后的 `UserContext`。运行 `make codegen-openapi`，禁止手改 generated artefacts。

#### 8.2 Persistence and migration

为 `users` 增加资料补全状态字段：`profile_completed_at` 与 `terms_accepted_at`。迁移必须幂等，并把已有带 displayName 的活跃用户 backfill 为已补全，避免既有测试账号被误导到资料补全。新邮箱 verify 创建用户时保持 `profile_completed_at` / `terms_accepted_at` 为空。

#### 8.3 Unified challenge start and verify semantics

`startAuthEmailChallenge` 不再检查或暴露账号是否存在，只创建同一类 email challenge 并发送 6 位 code。`verifyAuthEmailChallenge` 消费 challenge 后按 normalized email 查找用户：既有用户直接 mint session；新邮箱创建资料未补全账号、user_settings 和 session。Out-of-scope `signup/login` purpose 不参与业务分支，不得出现重复注册 409 或未知邮箱登录拒绝语义。

#### 8.4 `/me` and `PATCH /me`

`/me` 对已登录未补全账号返回 200，并设置 `profileCompletionRequired=true`；未登录仍返回 B1 auth error。`completeMyProfile` 必须要求有效 session、非空 trimmed displayName、`acceptedTerms=true`，成功后写入 displayName、`terms_accepted_at`、`profile_completed_at` 并返回 `profileCompletionRequired=false`。该 endpoint 不承接 candidate profile、简历、JD 或面试偏好字段。

#### 8.5 Privacy, metrics, and out-of-scope negative gates

日志、audit、metric label 仍不得包含 raw code、session cookie、完整邮箱或 PII。新增 focused negative tests / grep gate：`purpose=signup/login`、duplicate-register start rejection、displayName stored before verify、password/OAuth auth wire、email URL callback 和 out-of-scope `auth_register` live contract 不得作为当前后端完成证据。

#### 8.6 BDD-Gate: BDD.AUTH.EMAIL.001

代码层 owner tests 必须证明新邮箱 verify 后 `/me.profileCompletionRequired=true`、`PATCH /me` 后为 false、logout/relogin 保留账号资料状态，并覆盖邮箱唯一、displayName 非唯一及非法/重放请求 fail closed。

#### 8.6a E2E-HANDOFF: E2E.P0.101

该场景在被显式运行时再以 real frontend/backend/Mailpit 复核同一主链路；静态资产登记或代码测试不得将其标记为 PASS。

### Phase 9: Unauthorized account handler test consolidation

Replace the duplicate unauthenticated `getMe` / `deleteMe` envelope tests with one table-driven test and named subtests. Preserve each HTTP method, handler invocation, 401 status, JSON decode and exact `AUTH_UNAUTHORIZED` code assertion; do not change production handlers or BDD behavior.

### Phase 10: OPENAPI-007 minimal current-user projection

Use OPENAPI-007 generated types as the only public shape. Focused RED must fail while internal `UserContext`, `GetUserContext` SQL/scan, handler mapping or test builders read/fill `UILanguage` / `PreferredPracticeLanguage` / `emailMasked`. GREEN removes those fields and selects the existing account email plus users identity/profile；`user_settings.analytics_opt_in` remains internal for runtime-config resolution. The Auth handler must serialize exact `id/email/displayName/profileCompletionRequired`, never log raw email and never retain compatibility aliases or language fields.

Coordinate B4 001 Phase 13 before removing DB columns；new account creation still inserts a `user_settings` row so analytics opt-in keeps its default owner. `region/timezone` have no Auth consumer and are removed by B4. Run focused store/handler/runtime-config tests, generated compile, root `make test` and production old-field zero-reference gates. User-visible Settings behavior remains owned by frontend-shell BDD；`E2E.P0.101` may verify the real values after login but this backend phase creates no parallel scenario.

### Phase 11: Production SMTP delivery

#### 11.1 A4 provider/config contract

先由 `secrets-and-config/001` 原地落地 `mailpit|smtp` provider、SMTP username/password/TLS mode 与 staging/prod fail-fast；删除未消费的 `EMAIL_PROVIDER_API_KEY`。配置 owner 只维护一组 typed contract tests，C1 不复制完整配置矩阵。

#### 11.2 SMTP transport TDD

在 `backend/internal/auth` 先补 RED tests，覆盖 Mailpit 无认证明文投递、STARTTLS、隐式 TLS、认证顺序、TLS 最低 1.2、服务器不支持 STARTTLS、无效地址/凭据与错误脱敏；再以可注入 SMTP client factory 落地最小 transport。raw code 仍只存在 transient delivery secret 和出站邮件正文。

#### 11.3 Runtime wiring and provider selection

`cmd/api` 对 `mailpit` 与 `smtp` 都注册同一 `SMTPDeliveryWriter`：Mailpit 使用 `none` 且无认证；标准 SMTP 从 A4 loader 读取 secret password，按 TLS mode 建连并认证。未知 provider 不得静默回落 `DevMailSink`；`DevMailSink` 仅保留单测显式使用。

#### 11.4 Operation matrix and verification

| operationId | fixture | frontend consumer | backend handler | persistence | AI dependency | scenario coverage |
|-------------|---------|-------------------|-----------------|-------------|---------------|-------------------|
| `startAuthEmailChallenge` | `openapi/fixtures/Auth/startAuthEmailChallenge.json` 既有场景 | frontend auth email flow（generated client） | C1 handler → `async_jobs(email_dispatch)` → internal runner → SMTP writer | `auth_challenges`、`async_jobs`；凭据与 raw code 不持久化到 job payload | none | domain `BDD.AUTH.EMAIL.002`；真实 Mailpit 链路复用 `E2E.P0.101`，外部 SMTP 用显式脱敏 smoke 证据，不新建配置型 E2E |

执行 focused Go tests、`make lint-config`、根 `make test`、`make build`、旧 `EMAIL_PROVIDER_API_KEY` current-scope zero-reference，并分别对 Mailpit 与用户 `.env` 的标准 SMTP 做真实投递验收。

Phase 11 的单实例 MVP 边界由 Phase 12 取代：验证码不进入 job payload，而是复用现有 `REDIS_URL` 写入加密、namespaced、5 分钟 TTL 的共享 delivery secret。`dev-container-up` 仍可停止仓库 PID 文件管理的 host-run backend，以保持本地运行拓扑可预测，但正确性不再依赖单实例。

### Phase 12: Redis-backed cross-instance delivery secret

#### 12.1 Redis store contract

先写 RED tests 锁定 namespaced SHA-256 key、AES-GCM encrypted value、`ChallengeTTL`、miss/expired/decrypt/Redis error 脱敏与两个独立 store/client 读取同一 secret。加密 key 只由 `AUTH_CHALLENGE_TOKEN_PEPPER` 通过 HKDF-SHA256 固定 context label 域隔离派生，不新增配置 key；pepper 轮换按 clean break 使 pending secret 失效。

#### 12.2 Service and writer lifecycle

把 `DeliverySecretStore` 改为 context-aware、error-returning `Put/Get/Delete` 合同。`StartEmailChallenge` 在 enqueue 前写入 Redis，写入失败不 enqueue；payload/enqueue 失败 best-effort 删除。SMTP writer 发送成功后删除，发送失败保留到 TTL 供 runner retry；missing/decrypt/backend errors 只返回安全阶段信息。

#### 12.3 Runtime wiring and real Redis gate

`cmd/api` 用 A4 `redis.url` 构造并 ping 一个 Redis client，将同一 store 注入 service 与 writer，并在 shutdown 关闭 client；test runtime 继续显式注入 `DevMailSink`。真实 Redis integration 使用两个独立 client 证明实例 A Put、实例 B Get/Delete；full-container 重建后再跑 Mailpit challenge->receive->verify 与外部 SMTP 脱敏 smoke。

#### 12.4 Operation matrix and verification

| operationId | fixture | frontend consumer | backend handler | persistence | AI dependency | scenario coverage |
|-------------|---------|-------------------|-----------------|-------------|---------------|-------------------|
| `startAuthEmailChallenge` | `openapi/fixtures/Auth/startAuthEmailChallenge.json` 既有场景 | frontend auth email flow（generated client） | C1 handler -> `async_jobs(email_dispatch)` -> arbitrary backend runner -> SMTP writer | `auth_challenges`、`async_jobs`；Redis encrypted delivery secret（5m TTL）；raw code 不进入 DB/job | none | domain `BDD.AUTH.EMAIL.003` + real Redis cross-client integration；Mailpit 继续复用 `E2E.P0.101` handoff |

## 5 验收标准

- 仓库根 `make test` 覆盖 `startAuthEmailChallenge`、`email_dispatch` delivery、`verifyAuthEmailChallenge`、session middleware、`getMe`、`deleteMe`、`logout` 与 runtime-config resolver；focused Go tests 只用于开发反馈。
- session cookie 符合 ADR-Q1 与 OpenAPI 描述；server-side session 是真理源。
- 未登录、过期 token、重复 verify、无效 token、缺 cookie、revoked/expired session、缺 cookie logout、logout 幂等、delete handoff idempotency、缺 secret 都有测试覆盖。
- B3 `email_dispatch` payload helper gate 通过，raw code / URL / 邮箱明文 / 邮件正文不会进入 async payload、log 或 audit。
- 发码前不泄露账号存在性；新邮箱 verify 后创建资料未补全账号并返回 `profileCompletionRequired=true`；未补全用户重开浏览器、换浏览器或 logout 后重新登录仍必须保持未补全状态；`PATCH /me` 成功后返回 `profileCompletionRequired=false`。
- 日志 / metric / audit privacy grep 无 secret / PII 明文；auth metrics 名称已由 F1 登记或承接，label 只使用 F1 allowed labels。
- `backend/internal/auth` has no duplicate unauthenticated account-envelope test body at the scoped threshold.
- `/me` and profile-completion success use exact OPENAPI-007 four-field `UserContext` with full authenticated `email` and no `emailMasked`; auth store no longer reads obsolete display/practice-language columns while runtime-config analytics behavior remains intact.
- `mailpit` 与 `smtp` 均可通过 `EMAIL_PROVIDER` 选择；生产 SMTP 支持 STARTTLS / 隐式 TLS 和认证，staging/prod 对不安全或缺失配置 fail-fast，且凭据与 raw code 不泄露。
- 共享 Redis store 下，producer 与 consumer 位于不同 backend 实例时仍可投递同一 6 位验证码；Redis key/value、DB/job/log/audit 不暴露 raw code 或原始 secret ref。

## 6 风险与应对

| 风险 | 应对措施 |
|------|----------|
| 认证实现绕过 B2 schema | handler tests 必须使用 generated OpenAPI types / server contract |
| token 或 session secret 进入日志 | Phase 4.2 privacy test 和 grep gate 强制覆盖 |
| P0 误引入 OAuth / SSO | spec Out of Scope 和 checklist negative search 拦截 |
| 过早依赖独立 worker 阻塞本地验证 | 邮件派发由 backend internal runner 的 in-process kernel 承接；C1 不需要独立后台执行进程，dev sink / Mailpit writer 仍能被本地场景验证 |
| Redis 不可用、secret 过期或 pepper 轮换导致 pending job 无法取码 | startup ping 与 producer Put fail closed；consumer 返回脱敏 retryable failure；TTL 与 job retry budget有界，operator 恢复 Redis 后用户可重新发起 challenge |
| SMTP 已接受邮件但 Redis delete 失败导致 secret 暂留 | delivery 仍判定成功，避免重复发信；encrypted value 由 5 分钟 TTL 自动清理，delete error 不输出 key/ref/code |
| C1 绕过 B3 email payload 红线 | Phase 2.3 强制使用 generated `BuildEmailDispatchPayload`，并用 negative tests 拒绝 redacted fields |
| C1 抢占 privacy deletion 执行 | Phase 3.5 只做 auth/session handoff，backend internal runner / B4 删除执行不进入本 plan |
| C1 新增未登记 auth metric | Phase 4.3 先跑 F1 registry preflight；未登记则先修订 F1，不在 C1 私造 metric |
