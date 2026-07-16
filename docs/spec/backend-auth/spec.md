# Backend Auth Spec

> **版本**: 2.7
> **状态**: active
> **更新日期**: 2026-07-16

## 1 背景与目标

`backend-auth` 承接 ADR-Q1 的 P0 后端认证实现：自建 email-code challenge + first-party session cookie。它为 `frontend-shell` 的操作级登录拦截、pending action 恢复和 Settings 真实账号/隐私动作提供后端支撑。

本 subject 的目标是落地最小可用认证后端，同时保持 product-scope 的隐私红线和 B2 OpenAPI 契约。

## 2 范围

### 2.1 In Scope

- `POST /api/v1/auth/email/start` 邮箱挑战创建。
- `GET /api/v1/auth/email/verify` 邮箱挑战验证并签发 first-party session cookie。
- `GET /api/v1/me` 当前用户读取。
- `PATCH /api/v1/me` 首次登录资料补全：保存 displayName、条款确认时间和 profile completion 状态。
- `POST /api/v1/auth/logout` 清除 session。
- first-party session middleware / current-user resolver：保护除 B2 public endpoints 外的 P0 API；`logout` 使用 optional-session / always-clear-cookie 路径以保持幂等；`DELETE /api/v1/me` 提供认证态、同步软删 `users.deleted_at` / `users.status='deleted'`、撤销该用户所有 session 和 idempotent privacy_delete handoff。
- 为既有 `GET /api/v1/runtime-config` 注入 C1 session-aware resolver，供 A4 handler 合并用户级公开偏好。
- C1 auth service 通过 `EmailDispatchEnqueuer` 写入 `async_jobs(job_type='email_dispatch')`；backend-async-runner 当前注册 `EmailDispatchHandler`，通过 `DeliveryWriter` 投递到 Mailpit、标准 SMTP 或 dev sink；`ImmediateMailDispatcher` 仅作为测试 helper。
- 复用 B3 internal-only `email_dispatch` payload contract 与 generated `BuildEmailDispatchPayload` redaction policy 构造邮件派发输入；payload 只允许 helper 支持的脱敏字段。
- session cookie 属性、安全默认值、过期、幂等 logout、错误码映射。
- auth metrics / audit 事件最小生产边界：记录 started / minted / failure / logout / delete handoff 等可观测事实，但不落 secret / PII 明文。

### 2.2 Out of Scope

- 不实现 OAuth / SSO / 企业账号体系。
- 不实现 Team / EDU、订阅或计费能力。
- 不实现完整隐私导出；P0 隐私导出延后，逐域硬删除执行按 product-scope / B4 / backend internal runner owner 另行计划。C1 只负责 `DELETE /me` 的认证、请求受理期用户软删、全部 session 撤销与 privacy_delete handoff。
- 不实现独立后台执行进程、Asynq dispatcher 或生产级 outbox consumer；`email_dispatch` 当前由 backend internal runner 的 in-process kernel 承接，不保留 C1 私有 goroutine queue 或 parallel mail-dispatch lifecycle。
- 不新建独立 Redis 集群或第二套 cache 配置；邮件验证码共享存储复用 A2/A4 已有 `REDIS_URL` 与 Redis 7 服务。
- 不在日志中输出验证码、完整邮箱、session secret 或 PII。
- 当前项目未上线，Auth operation shape 可以按 active spec 进行 breaking cleanup；涉及 OpenAPI wire shape 的变更必须同步修订 `openapi-v1-contract`、fixtures 和 generated clients，不保留注册入口兼容层。

## 3 用户决策 / 待确认事项

| ID | 决策 | 锁定值 | 影响 |
|----|------|--------|------|
| D-1 | 认证方案 | email-code challenge + first-party session cookie | 不使用 Bearer token 作为 P0 浏览器主认证形态 |
| D-2 | 登录入口 | 操作级 gate | 后端不要求首页加载前认证 |
| D-3 | Cookie 安全 | HttpOnly、SameSite、Secure 按环境配置 | dev 可降级 Secure，但必须可测试 |
| D-4 | 配置来源 | A4 secrets/config 提供 `SESSION_COOKIE_SECRET`、`AUTH_CHALLENGE_TOKEN_PEPPER`、`EMAIL_PROVIDER`、`EMAIL_SMTP_HOST` / `EMAIL_SMTP_PORT` / `EMAIL_SMTP_USERNAME` / `EMAIL_SMTP_PASSWORD` / `EMAIL_SMTP_TLS_MODE` / `EMAIL_FROM_ADDRESS` / `EMAIL_VERIFY_BASE_URL` 和固定 `ei_session` cookie name；local dev `EMAIL_VERIFY_BASE_URL` 仅用于 frontend origin / callback 配置边界，backend verify API 仍固定为 `GET /api/v1/auth/email/verify`；TTL、rate-limit 与 in-memory `DevMailSink` 测试默认值由 C1 代码常量持有并在包级文档记录 | 邮件、cookie、session secret 不私造配置 key；新增 email config key 必须先修订 A4 |
| D-5 | 错误码 | B1 shared error envelope | 认证错误必须使用共享错误 shape |
| D-6 | P0 邮件派发 | C1 通过 `async_jobs(job_type='email_dispatch')` + backend internal runner 派发邮件；单测使用 in-memory `DevMailSink`，非 test runtime 使用 Redis delivery secret store；Mailpit / SMTP writer 从 `auth_challenges` 查询收件人，并可跨 backend 实例读取同一 6 位数字验证码；邮件正文只展示验证码和 5 分钟有效期，不包含 email URL callback 或完整 URL | 不要求独立后台执行进程；不把 raw code、完整 URL、邮箱明文、邮件正文或标题写入 async job payload、日志或 audit；本地测试不依赖真实外部邮箱服务或真实邮箱账号 |
| D-7 | Account deletion auth handoff | `DELETE /api/v1/me` 使用 C1 session middleware 验证当前用户，支持 `Idempotency-Key` 或等价 active-request dedupe；受理请求时同步将 `users.deleted_at` 置为当前时间、`users.status='deleted'`，撤销该用户所有 session，并返回 B2 `202 + PrivacyRequestWithJob`；逐域硬删与用户行最终 hard delete 归 backend internal runner / B4 | C1 不扩展删除 schema，不绕过 B2 contract；重复请求不得创建重复 active 删除任务；request/job success 不得早于账户身份清理 gate |
| D-8 | 邮箱账号唯一性与单入口登录 | 邮箱是唯一账号标识，用户只有一个邮箱验证码登录入口；`AuthEmailStartRequest` 不再暴露 `purpose=login/signup` 或 `displayName`，发码前不得泄露邮箱是否已存在；verify 时既有邮箱直接登录，新邮箱创建资料未补全账号并签发 session；displayName 不唯一、不参与账号唯一性判断 | 注册页不再是 live route；重复使用同一邮箱只会登录同一账号，不创建第二个用户；账号唯一性由 normalized email 保证 |
| D-9 | 首次登录资料补全 | 新邮箱首次 verify 创建 `profile_completed_at IS NULL`、`terms_accepted_at IS NULL` 的账号；`/me.profileCompletionRequired=true` 是前端强制进入资料补全页的唯一后端信号；`PATCH /me` 只负责首次资料补全，保存 trimmed displayName、条款确认和完成时间 | 未补全账号即使关闭浏览器、换浏览器重新登录、退出后重新登录、刷新或直开业务 URL，登录后 `/me` 仍返回 profile completion required；完成后 `/me.profileCompletionRequired=false`，后续同邮箱登录直接进入正常登录态 |
| D-10 | Minimal current-user context | accepted OPENAPI-007：`/me` 与 `PATCH /me` success 只返回 `id/email/displayName/profileCompletionRequired`；删除 `emailMasked` 与 UI/practice language 字段，不以 optional/default 保留 | Settings 复用 runtime `/me` 展示完整账号 email；该 authenticated PII 不写入日志/场景证据。TopBar language 和 practice language 由各自 owner 承接。内部 `analytics_opt_in` 仍供 runtime-config resolver 使用，不进入 `UserContext` |
| D-11 | 生产 SMTP provider | `EMAIL_PROVIDER=mailpit` 使用无认证明文 SMTP，仅允许本地开发；`EMAIL_PROVIDER=smtp` 使用用户名/密码认证，并按 `EMAIL_SMTP_TLS_MODE=starttls|tls` 分别支持显式 STARTTLS 与隐式 TLS。staging/prod 禁止 `mailpit`、禁止 `none` TLS，SMTP TLS 最低 1.2；未知 provider / TLS mode、缺 host/port/from/username/password 均启动失败；完整 SMTP 会话受 runner context 与有界 deadline 约束；非 ASCII `Subject` 使用 RFC 2047 encoded-word，UTF-8 text/plain 与 text/html part 使用声明一致的 MIME transfer encoding | 删除未消费的 `EMAIL_PROVIDER_API_KEY`；凭据只从 A4 secret source 读取，不进入日志、错误、job payload、audit 或 runtime-config；DATA 已获最终成功响应后，QUIT/连接清理失败不得触发重复投递 |
| D-12 | Redis delivery secret | 非 test runtime 复用 `REDIS_URL` 建立共享 store；key 为 namespaced `SHA-256(deliverySecretRef)`，value 使用由 `AUTH_CHALLENGE_TOKEN_PEPPER` 经 HKDF-SHA256 固定 context 域隔离派生的 AES-GCM key 加密，TTL 固定为 `ChallengeTTL=5m`。producer 必须先成功写入 secret 才创建可计入限流的 challenge 并 enqueue；secret 写入失败不得留下 challenge，challenge 创建失败 best-effort 删除 secret；consumer 发送成功后删除，失败保留到 TTL 供 job retry；启动时 Redis URL/连接失败必须 fail closed | 多 backend 实例可消费同一任务；Redis、job payload、DB、log、audit 均不出现 raw code。pepper 轮换会使最多 5 分钟内的 pending delivery 失效，这是当前 clean-break 语义，不保留旧 key 兼容解密 |

## 4 设计约束

- Challenge code 必须 hash 后存储或在本地 stub 中以不可逆形式比较；明文 code 只能在一次性发送边界短暂存在。
- Session ID / secret 不得进入日志、metrics label、audit 明文字段或 API response。
- Challenge token 必须是 6 位 cryptographically random numeric code；Challenge TTL 固定 5 分钟；同邮箱或同 IP 1 分钟内第 3 次及以上请求必须 rate-limit / dedupe，响应不得泄露账号存在性。
- Session 默认 30 天滑动续期；B4 当前 `sessions` 表以 `updated_at` / `revoked_at` 承载续期触点与撤销时间，如实现需要独立 `last_seen_at` 字段，必须先修订 ADR-Q1 / B4 migration owner。
- Logout 必须幂等；没有有效 session 时返回可预期结果，不泄露账号存在性。
- `/runtime-config` 由 A4 handler 持有公开 allowlist；C1 只能提供 session-aware resolver，不得扩大 response 字段。
- `/me` 未登录必须返回 B2 / B1 约定的认证错误，不返回假用户。
- `/me` 对已登录但资料未补全的账号必须返回 200，并显式返回 `profileCompletionRequired=true`；不得把未补全状态误判为未登录。
- `PATCH /me` 只承接首次登录资料补全，不承接候选人画像、简历、JD 或面试偏好业务字段；`displayName` 必须 trim 后非空，`acceptedTerms` 必须为 true，重复补全不得覆盖账号邮箱或创建新账号。
- Auth store current-user projection只读取用户 identity/profile 字段与内部 `analytics_opt_in`；不得读取或填充 `ui_language/preferred_practice_language/region/timezone`，也不得为 OPENAPI-007 删除字段提供兼容常量。
- 邮箱挑战发送失败、挑战过期、重复验证、缺 cookie、无效 session 都必须有可测试错误路径。
- Auth metrics 必须先登记到 F1 baseline metrics 字典或由 F1 owner 明确承接；label 只能使用 F1 allowed labels，禁止 `user_id`、邮箱、session id、token、URL path 明文等高基数或敏感 label。

## 5 模块边界

| 边界 | Owner | 说明 |
|------|-------|------|
| API contract | B2 `openapi-v1-contract` | Auth endpoints、response schema、cookie 描述 |
| backend auth | `backend-auth` | handlers、service、store、session、C1 backend-internal email_dispatch producer/handler、dev sink、Mailpit SMTP writer、challenge delivery |
| event/outbox job contract | B3 `event-and-outbox-contract` + active [`backend-async-runner`](../backend-async-runner/spec.md) | `email_dispatch` 已收口为 `async_jobs(job_type='email_dispatch')`：producer `auth.EmailDispatchEnqueuer` 同库写入 job 行，kernel `auth.EmailDispatchHandler` 经 `runner.Runtime` lease 后通过 `DeliveryWriter` 投递；payload 仍受 `BuildEmailDispatchPayload` redaction 约束 |
| config/secrets | A4 `secrets-and-config` | session secret、challenge pepper、email provider secret、Mailpit SMTP dev keys、固定 `ei_session` cookie name；TTL / rate-limit 默认值归 C1 代码常量，新增配置前先修订 A4 |
| frontend gate | `frontend-shell` | pendingAction、登录页面和登录后恢复 |
| DB/session storage | B4 `db-migrations-baseline` | session/challenge 表或等价持久化边界 |
| privacy deletion execution | backend-auth + backend internal runner / B4 | C1 在 `DELETE /me` 请求受理时同步软删用户身份并撤销该用户所有 session；后续 `privacy_delete` job 由 backend internal runner / B4 执行逐域 hard delete 与用户行最终删除 |
| observability registry | F1 `observability-stack` | auth metric names 与 allowed labels 登记；C1 只消费已登记指标 |

## 6 验收标准

| ID | 场景 | Given | When | Then | 对应 Plan |
|----|------|-------|------|------|-----------|
| C-1 | Email-code session | 用户请求邮箱挑战 | 验证 challenge | 返回 first-party session cookie；既有账号 `/me.profileCompletionRequired=false`，新邮箱账号 `/me.profileCompletionRequired=true` | 001-email-code-session-bootstrap |
| C-2 | Logout 幂等 | 用户已有或没有有效 session | 调用 logout | cookie 被清除且响应不泄露账号状态 | 001-email-code-session-bootstrap |
| C-3 | 错误路径 | challenge 过期、重复验证、缺 cookie、配置缺失 | 调用对应 endpoint | 返回 B1 error envelope，日志无 secret / PII 明文 | 001-email-code-session-bootstrap |
| C-4 | Runtime config session resolver | 前端启动，用户可能携带有效 session | 请求 `/runtime-config` | A4 handler 仍只返回公开 allowlist 字段；C1 session resolver 只影响允许公开的用户级偏好，不泄露 secret / internal flag | 001-email-code-session-bootstrap |
| C-5 | Auth middleware and delete handoff | 用户携带有效或无效 `ei_session` | 访问 protected Auth operation、logout 或 `DELETE /me` | auth start / verify / runtime-config 不要求 session；logout optional-session 且总是清 cookie；protected endpoints 使用 first-party session；`DELETE /me` 支持 idempotency / active-request dedupe，返回 B2 删除响应，同步设置 `users.deleted_at` / `users.status='deleted'` 并撤销该用户所有 session，逐域 hard delete 仍由 backend internal runner / B4 承接 | 001-email-code-session-bootstrap |
| C-6 | Email dispatch redaction | 用户请求邮箱挑战 | C1 auth flow 写入 `async_jobs(email_dispatch)`，backend-async-runner kernel `EmailDispatchHandler` lease 后写入 dev sink 或 Mailpit SMTP writer | `email_dispatch` payload 只含 allowed fields；raw code / URL / 邮箱明文 / 邮件正文不进入 async_jobs payload、dev sink、outbox、log 或 audit；无需独立后台执行进程即可通过本地验证 | 001-email-code-session-bootstrap |
| C-8 | Local Mailpit sign-in | `EMAIL_PROVIDER=mailpit`，Mailpit 由 local-dev-stack 提供，用户请求 synthetic `.example.test` 邮箱挑战 | `EmailDispatchHandler` 处理 queued job | SMTP writer 从 DB lookup 收件人、从 transient secret store 取 6 位验证码并投递 code-only 邮件到 Mailpit；用户在前端 `/auth/verify` 手动输入验证码后签发 `ei_session`；邮件正文、URL、日志和场景证据不保存 raw code；不使用真实外部邮箱服务、真实邮箱账号或 `backend/cmd` 场景 helper | local-dev-stack/001 Mailpit revision + frontend-shell/001 Phase 8 |
| C-7 | Auth observability | challenge / verify / logout / failure 发生 | 记录 metrics / audit | 指标名已在 F1 baseline 或 F1 承接 gate 中登记，label 符合 F1，audit 只含 ID / hash / 状态，不含 secret / PII 明文 | 001-email-code-session-bootstrap |
| C-9 | Unified email login and profile completion | 用户从单一邮箱验证码入口提交新邮箱或既有邮箱 | verify 后请求 `/me`；未补全用户调用 `PATCH /me` 提交 displayName + acceptedTerms | 发码前不泄露账号存在性；新邮箱创建资料未补全账号并返回 `profileCompletionRequired=true`；关闭浏览器、换浏览器重新登录、退出后重新登录、刷新或直开业务 URL 后仍必须先补全资料；补全成功后同邮箱后续登录返回 `profileCompletionRequired=false`；normalized email 唯一，displayName 不唯一 | 001-email-code-session-bootstrap |
| C-10 | Minimal `/me` projection | authenticated 或 profile-incomplete 用户请求 `/me` / 完成 profile | handler mapping + generated contract | success body 精确包含 id、完整账号 email、display name、profile completion flag；无 `emailMasked`、旧语言字段或其他额外 PII；runtime config analytics 仍读取保留列 | 001-email-code-session-bootstrap Phase 10 + OPENAPI-007 + B4 001 Phase 13 |
| C-11 | Mailpit / production SMTP delivery | 环境选择 `mailpit` 或 `smtp` 且配置满足 A4；生产 SMTP 凭据通过 secret source 注入 | 用户调用既有 `startAuthEmailChallenge`，internal runner 消费 `email_dispatch` | 两种 provider 都投递相同 code-only 邮件；Mailpit 使用本地无认证 SMTP，生产 SMTP 在 TLS 1.2+ 上认证；中文 locale 的主题和两种正文 part 可由标准 MIME reader 无损解码；runner 取消或 deadline 能终止停滞会话；DATA 已被接受后的 QUIT 失败不重复发信；失败返回脱敏 delivery error，raw code、邮箱和凭据不进入持久化 payload / log / audit | 001-email-code-session-bootstrap Phase 11/12/13 |
| C-12 | Cross-instance Redis delivery | 两个 backend 实例共享同一 `REDIS_URL` 和 challenge pepper | 实例 A 创建 challenge/secret/job，实例 B lease `email_dispatch` 并发送 | 实例 B 解密并投递同一 6 位验证码；Redis key 有 5 分钟 TTL，value/key 不泄露 raw code/ref；secret Put 失败不创建 challenge 或消耗限流额度；成功后删除，Redis unavailable/miss/decrypt failure 均 fail closed 且错误脱敏 | 001-email-code-session-bootstrap Phase 12 |

## 7 关联计划

- [001-email-code-session-bootstrap](./plans/001-email-code-session-bootstrap/plan.md)
