# Passwordless Session Bootstrap Checklist

> **版本**: 1.1
> **状态**: active
> **更新日期**: 2026-05-06

**关联计划**: [plan](./plan.md)

## Phase 1: Storage and config boundaries

- [x] 1.1 锁定 auth storage；验证: store tests 覆盖 `users`、`user_settings`、`auth_challenges`、`sessions` 表读写，确认 `external_identities` 仅作为 P1 SSO 空表槽存在且 C1 不提供 P0 读写 store 方法，确认无需新增 migration；滑动续期使用 `sessions.updated_at`；若需 `last_seen_at` 或 schema 变更，先停止并修订 ADR-Q1 / B4 owner spec
- [x] 1.2 锁定 config / secret 边界；验证: config tests 覆盖 `SESSION_COOKIE_SECRET`、`AUTH_CHALLENGE_TOKEN_PEPPER`、`EMAIL_PROVIDER`、`EMAIL_PROVIDER_API_KEY` 缺失时 fail-fast，固定 `ei_session` cookie name，且 15 分钟 challenge TTL / 30 天 session TTL / 1 分钟第 3 次 rate-limit / dev mail sink 默认值作为 C1 代码常量有测试和包级文档；若需新增配置先停止并修订 A4
- [x] 1.3 锁定 generated Auth surface 和 session middleware；验证: compile / contract tests 断言 B2 generated `ServerInterface` 的 `startAuthEmailChallenge`、`verifyAuthEmailChallenge`、`getMe`、`deleteMe`、`logout`、`getRuntimeConfig` 均被 C1/A4 wiring 覆盖，public endpoints 为 auth start / verify / runtime-config，logout 为 optional-session / always-clear-cookie 特例，其余 protected endpoint 走 first-party session middleware

## Phase 2: Challenge issue and delivery

- [x] 2.1 实现 `startAuthEmailChallenge`；验证: handler/service tests 覆盖 accepted response、token hash 入库、IP / UA hash、通过 C1 backend-internal mail dispatcher 入队，dev mail sink retrieval 收到一次性链接，且应用日志不输出 token、完整 URL、邮箱明文或邮件正文
- [x] 2.2 实现 rate-limit / dedupe 基线；验证: tests 覆盖同邮箱或同 IP 1 分钟内第 3 次及以上请求不泄露账号存在性，响应仍符合 B2 schema，dedupe key 不含邮箱明文
- [x] 2.3 接入 B3 `email_dispatch` redacted payload；验证: tests 使用 generated `BuildEmailDispatchPayload` 构造 allowed payload，通过包含 `rawMagicLinkToken` / `magicLinkUrl` / `recipientEmail` / `emailBody` 任一 redacted field 的 negative case，并确认 in-process queue / dev sink / future outbox / async payload / log / audit 不含 redacted fields
- [x] 2.4 实现 C1 backend-internal mail dispatcher；验证: tests 覆盖 handler 不等待邮件 provider 即返回 B2 `202`、后台 goroutine / 线程 drain 队列写入 dev mail sink、派发失败不泄露 token / 邮箱、graceful shutdown drain 或明确丢弃策略可观测，且不启动 C8 worker 进程也能完成本地邮件读取

## Phase 3: Verify, session, and current user

- [x] 3.1 实现 `verifyAuthEmailChallenge`；验证: tests 覆盖成功签发 `ei_session` cookie、过期 token、重复 verify、无效 token、session_hash 入库且不返回 cookie 明文
- [x] 3.2 实现 session middleware / current-user resolver；验证: middleware tests 覆盖缺 cookie、无效 session、expired / revoked session 返回 B1 error envelope，active session 更新 `sessions.updated_at` / expiry 且不记录 cookie 明文
- [x] 3.3 实现 `/me`；验证: handler tests 覆盖有效 session 返回 masked email / displayName / language，缺 cookie 或无效 session 返回 B1 error envelope
- [x] 3.4 实现 logout；验证: tests 覆盖有效 session 撤销、缺 cookie / 无效 session 仍进入 handler 并 Set-Cookie 清除、重复 logout 幂等和无账号存在性泄露
- [x] 3.5 实现 `deleteMe` auth handoff；验证: handler tests 覆盖有效 session 返回 B2 `202 + PrivacyRequestWithJob` 兼容响应并撤销 session，`Idempotency-Key` 或等价 active-request dedupe 使重复请求返回同一 active `privacy_delete` job 或同义终态且不创建重复 job，缺/无效 session 返回 B1 error envelope，实际 privacy_delete worker / 删除矩阵不在 C1 中实现

## Phase 4: Runtime config resolver, privacy, and observability

- [x] 4.1 接入 A4 `/runtime-config` session resolver；验证: tests 断言 C1 只向 A4 handler 注入 session-aware resolver，未登录保持 public response，有效 session 只影响 A4 allowlist 内用户级偏好，secret / internal flag 不出 response
- [x] 4.2 补隐私和可观测红线；验证: privacy tests / grep 确认日志、metric label、audit 不含 magic token、session cookie、完整邮箱、secret 或 PII 明文
- [x] 4.3 接入 auth metrics / audit 最小事件；验证: F1 registry preflight 断言 `auth_challenge_started_total`、`auth_session_minted_total`、`auth_failure_total` 等 metric 已登记到 F1 baseline metrics 字典或 F1 owner plan；tests / lint 断言 metric 只使用 F1 allowed labels，challenge started / session minted / logout / delete handoff / failure audit 只含 ID / hash / 状态 / trace，不含 token、session id、邮箱明文或 URL

## Phase 5: BDD and handoff

- [x] 5.1 BDD-Gate: 验证 E2E.P0.003 通过
  <!-- verified: 2026-05-06 method=scenario bddChecklist=complete scenario=E2E.P0.003 run=.test-output/runs/20260506T1911-backend-auth-p0-003/e2e/E2E.P0.003/result.json -->
- [x] 5.2 Handoff 给 frontend-shell；验证: backend README 或 package docs 说明 Auth API、cookie 行为、dev mail sink、错误码和前端 pendingAction 接入边界
- [x] 5.3 active-scope 负向搜索通过；验证: backend-auth / API wiring active code 不引入 Bearer token P0 主认证、OAuth / SSO P0 行为、`external_identities` P0 读写 store、明文 token/session 存储、log-only magic token delivery、独立 C8 worker 前置依赖或旧 AI gateway / voice route 口径；允许 A3 provider adapter 内部使用 provider-side `Authorization: Bearer`，不得误判为浏览器主认证
  <!-- verified: 2026-05-06 method=rg scope=backend/internal/auth,backend/cmd/api,backend/internal/api/generated allowed=negative-doc-comments+internal-session-hash-only -->
