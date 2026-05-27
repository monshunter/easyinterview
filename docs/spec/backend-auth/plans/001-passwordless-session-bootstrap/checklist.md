# Passwordless Session Bootstrap Checklist

> **版本**: 1.6
> **状态**: completed
> **更新日期**: 2026-05-27

**关联计划**: [plan](./plan.md)

## Phase 1: Storage and config boundaries

- [x] 1.1 锁定 auth storage；验证: store tests 覆盖 `users`、`user_settings`、`auth_challenges`、`sessions` 表读写，确认 `external_identities` 仅作为 P1 SSO 空表槽存在且 C1 不提供 P0 读写 store 方法，确认无需新增 migration；滑动续期使用 `sessions.updated_at`；若需 `last_seen_at` 或 schema 变更，先停止并修订 ADR-Q1 / B4 owner spec
- [x] 1.2 锁定 config / secret 边界；验证: config tests 覆盖 `SESSION_COOKIE_SECRET`、`AUTH_CHALLENGE_TOKEN_PEPPER`、`EMAIL_PROVIDER`、`EMAIL_PROVIDER_API_KEY` 缺失时 fail-fast，固定 `ei_session` cookie name，且 15 分钟 challenge TTL / 30 天 session TTL / 1 分钟第 3 次 rate-limit / dev mail sink 默认值作为 C1 代码常量有测试和包级文档；若需新增配置先停止并修订 A4
- [x] 1.3 锁定 generated Auth surface 和 session middleware；验证: compile / contract tests 断言 B2 generated `ServerInterface` 的 `startAuthEmailChallenge`、`verifyAuthEmailChallenge`、`getMe`、`deleteMe`、`logout`、`getRuntimeConfig` 均被 C1/A4 wiring 覆盖，public endpoints 为 auth start / verify / runtime-config，logout 为 optional-session / always-clear-cookie 特例，其余 protected endpoint 走 first-party session middleware

## Phase 2: Challenge issue and delivery

- [x] 2.1 实现 `startAuthEmailChallenge`；验证: handler/service tests 覆盖 accepted response、token hash 入库、IP / UA hash、通过 C1 backend-internal mail dispatcher 入队，dev mail sink retrieval 收到一次性链接，且应用日志不输出 token、完整 URL、邮箱明文或邮件正文
- [x] 2.2 实现 rate-limit / dedupe 基线；验证: tests 覆盖同邮箱或同 IP 1 分钟内第 3 次及以上请求不泄露账号存在性，响应仍符合 B2 schema，dedupe key 不含邮箱明文
- [x] 2.3 接入 B3 `email_dispatch` redacted payload；验证: tests 使用 generated `BuildEmailDispatchPayload` 构造 allowed payload，通过包含 `rawMagicLinkToken` / `magicLinkUrl` / `recipientEmail` / `emailBody` 任一 redacted field 的 negative case，并确认 in-process queue / dev sink / future outbox / async payload / log / audit 不含 redacted fields
- [x] 2.4 实现 C1 backend-internal mail dispatcher；验证: tests 覆盖 handler 不等待邮件 provider 即返回 B2 `202`、后台 goroutine / 线程 drain 队列写入 dev mail sink、派发失败不泄露 token / 邮箱、graceful shutdown drain 或明确丢弃策略可观测，且不启动独立 worker 进程也能完成本地邮件读取

## Phase 3: Verify, session, and current user

- [x] 3.1 实现 `verifyAuthEmailChallenge`；验证: tests 覆盖成功签发 `ei_session` cookie、过期 token、重复 verify、无效 token、session_hash 入库且不返回 cookie 明文
- [x] 3.2 实现 session middleware / current-user resolver；验证: middleware tests 覆盖缺 cookie、无效 session、expired / revoked session 返回 B1 error envelope，active session 更新 `sessions.updated_at` / expiry 且不记录 cookie 明文
- [x] 3.3 实现 `/me`；验证: handler tests 覆盖有效 session 返回 masked email / displayName / language，缺 cookie 或无效 session 返回 B1 error envelope
- [x] 3.4 实现 logout；验证: tests 覆盖有效 session 撤销、缺 cookie / 无效 session 仍进入 handler 并 Set-Cookie 清除、重复 logout 幂等和无账号存在性泄露
- [x] 3.5 实现 `deleteMe` auth handoff；验证: handler tests 覆盖有效 session 返回 B2 `202 + PrivacyRequestWithJob` 兼容响应并撤销 session，`Idempotency-Key` 或等价 active-request dedupe 使重复请求返回同一 active `privacy_delete` job 或同义终态且不创建重复 job，缺/无效 session 返回 B1 error envelope，实际 privacy_delete runner / 删除矩阵不在 C1 中实现

## Phase 4: Runtime config resolver, privacy, and observability

- [x] 4.1 接入 A4 `/runtime-config` session resolver；验证: tests 断言 C1 只向 A4 handler 注入 session-aware resolver，未登录保持 public response，有效 session 只影响 A4 allowlist 内用户级偏好，secret / internal flag 不出 response
- [x] 4.2 补隐私和可观测红线；验证: privacy tests / grep 确认日志、metric label、audit 不含 raw challenge secret、session cookie、完整邮箱、secret 或 PII 明文
- [x] 4.3 接入 auth metrics / audit 最小事件；验证: F1 registry preflight 断言 `auth_challenge_started_total`、`auth_session_minted_total`、`auth_failure_total` 等 metric 已登记到 F1 baseline metrics 字典或 F1 owner plan；tests / lint 断言 metric 只使用 F1 allowed labels，challenge started / session minted / logout / delete handoff / failure audit 只含 ID / hash / 状态 / trace，不含 token、session id、邮箱明文或 URL

## Phase 5: BDD and handoff

- [x] 5.1 BDD-Gate: 验证 E2E.P0.003 通过
  <!-- verified: 2026-05-06 method=scenario bddChecklist=complete scenario=E2E.P0.003 run=.test-output/runs/20260506T1911-backend-auth-p0-003/e2e/E2E.P0.003/result.json -->
- [x] 5.2 Handoff 给 frontend-shell；验证: backend README 或 package docs 说明 Auth API、cookie 行为、dev mail sink、错误码和前端 pendingAction 接入边界
- [x] 5.3 active-scope 负向搜索通过；验证: backend-auth / API wiring active code 不引入 Bearer token P0 主认证、OAuth / SSO P0 行为、`external_identities` P0 读写 store、明文 token/session 存储、log-only raw code delivery、独立 worker 前置依赖或旧 AI gateway / voice route 口径；允许 A3 provider adapter 内部使用 provider-side `Authorization: Bearer`，不得误判为浏览器主认证
  <!-- verified: 2026-05-06 method=rg scope=backend/internal/auth,backend/cmd/api,backend/internal/api/generated allowed=negative-doc-comments+internal-session-hash-only -->

## Phase 6: L2 remediation

- [x] 6.1 修复 `DELETE /me` idempotency user scope；验证: store / handler tests 覆盖两个不同用户使用相同 `Idempotency-Key` 时不会复用彼此 active `privacy_delete` handoff，同一用户重复 key 仍返回同一 active request / job
  <!-- verified: 2026-05-06 command="cd backend && go test ./internal/auth -run TestSQLStorePrivacyDeleteDedupeKeyIsScopedByUser -count=1" -->
- [x] 6.2 修复 runtime Auth wiring；验证: `cmd/api` wiring tests 断言 auth start / verify / logout / `/me` / `DELETE /me` routes mounted，protected `/me` 经过 C1 session middleware，`/runtime-config` 使用 C1 session-aware resolver 而非 anonymous nil resolver
  <!-- verified: 2026-05-06 command="cd backend && go test ./cmd/api -run TestBuildAPIHandlerMountsAuthRoutesAndSessionAwareRuntimeConfig -count=1" -->
- [x] 6.3 修复 session cookie Secure policy；验证: handler tests 覆盖 verify minted cookie 与 logout / deleteMe clear cookie 共用 cookie policy，prod/staging Secure=true，dev 降级行为显式可测，属性仍为 HttpOnly + SameSite=Lax + Path=/
  <!-- verified: 2026-05-06 command="cd backend && go test ./internal/auth -run 'Test(VerifyAuthEmailChallengeConsumesTokenAndSetsSessionCookie|SessionCookiePolicyAllowsDevInsecureButKeepsProdSecure|LogoutRevokesCurrentSessionAndClearsCookie|LogoutWithoutSessionIsIdempotentAndClearsCookie|LogoutCanUseExplicitDevInsecureCookiePolicy|DeleteMeCreatesPrivacyDeleteHandoffRevokesSessionAndIsIdempotent)' -count=1" -->
- [x] 6.4 修复 challenge rate-limit SQL scope；验证: SQL store tests 断言同邮箱或同 IP 1 分钟窗口统计不再过滤 `status='pending'`，consumed / pending recent challenge 均计入第 3 次 rate-limit / dedupe 基线
  <!-- verified: 2026-05-06 command="cd backend && go test ./internal/auth -run TestSQLStoreCountRecentChallengesCountsAllRecentAttempts -count=1" -->
- [x] 6.5 修复 logout revoke failure response；验证: handler tests 覆盖有效 session revoke 失败时返回 B1 error envelope 而不是 204，同时清 cookie 且响应不泄露 session id / cookie / secret
  <!-- verified: 2026-05-06 command="cd backend && go test ./internal/auth -run TestLogoutRevokeFailureReturnsErrorEnvelopeAndClearsCookie -count=1" -->
- [x] 6.6 修复 runtime Auth secret fail-fast；验证: `cmd/api` builder tests 覆盖 `AUTH_CHALLENGE_TOKEN_PEPPER` / `SESSION_COOKIE_SECRET` 缺失时返回明确错误并不构造 background dispatcher，本地 dev 不得用空 pepper / 空 session secret 启动 C1 session runtime
  <!-- verified: 2026-05-06 command="cd backend && go test ./cmd/api -run TestBuildAuthServiceRejectsEmptyAuthSecrets -count=1" -->
- [x] 6.7 修复 logout optional-session resolver error；验证: `cmd/api` route tests 覆盖 cookie-bearing logout 在 session resolver / store error 时返回 B1 error envelope 而不是 204，缺失 / invalid / expired / revoked session 仍保持 optional-session 幂等清 cookie
  <!-- verified: 2026-05-06 command="cd backend && go test ./cmd/api -run 'TestBuild(AuthServiceRejectsEmptyAuthSecrets|APIHandlerLogoutPropagatesSessionResolverErrors|APIHandlerLogoutKeepsKnownSessionErrorsOptional)' -count=1" -->
- [x] 6.8 修复 logout revoke race 的 touch zero-row 归类；验证: middleware tests 覆盖 `ResolveSession` 读到 active session 后 `TouchSession` 返回 `sql.ErrNoRows` 的并发 revoke race，`logout` optional path 仍进入 handler 并幂等清 cookie，required protected path 归类为 B1 `AUTH_UNAUTHORIZED`
  <!-- verified: 2026-05-06 command="cd backend && go test ./internal/auth -run TestSessionMiddlewareTreatsTouchLostRaceAsAuthState -count=1" -->

## Phase 7: Email code and registration display-name remediation

- [x] 7.1 OpenAPI additive contract and generated clients；验证: `AuthEmailStartRequest` 包含 optional `purpose`（`login` / `signup`）与 `displayName`，verify query 描述/fixture/schema 体现 6 位数字 code，`make codegen-openapi` 后 Go / TS generated artefacts 与 `openapi/openapi.yaml` 一致
  <!-- verified: 2026-05-27 command="make codegen-openapi" evidence="OpenAPI auth email start/verify contract regenerated into backend and frontend generated artefacts" -->
- [x] 7.2 Six-digit code + 5-minute TTL；验证: focused Go tests 覆盖默认 challenge generator 只产生 6 位数字 code、`ChallengeTTL == 5*time.Minute`、challenge code hash 入库且 session token 仍使用 secure opaque token
  <!-- verified: 2026-05-27 command="cd backend && go test ./internal/auth -count=1 && go test ./... " evidence="auth crypto/config/service/store tests and full backend package suite passed" -->
- [x] 7.3 Registration display-name persistence and email uniqueness；验证: handler/service/store tests 覆盖注册页传入 `purpose=signup` + displayName 后暂存到 challenge、verify 成功后新建唯一 email user 并写入 displayName、重复注册同一 email 在 start 阶段返回 409 且不创建 challenge / 不发 code / 不覆盖 displayName、登录页 `purpose=login` 只登录已注册 email 且未知 email 不隐式创建账号
  <!-- verified: 2026-05-27 command="cd backend && go test ./internal/auth -count=1 && go test ./cmd/api -run 'Test(AuthEmail|BuildAuth|LocalDevCORS|BuildAPIHandlerMountsAuthRoutesAndSessionAwareRuntimeConfig)' -count=1 && go test ./..." evidence="duplicate signup rejected at start before challenge creation; signup/login purpose and displayName persistence covered" -->
- [x] 7.4 Code-only Mailpit / SMTP email；验证: SMTP writer tests 覆盖邮件标题/HTML/text 只展示 6 位 code 和 5 分钟有效期，不包含 magic link、`/auth/verify?token=`、完整 URL、delivery secret、raw email 以外的内部字段或日志泄露；dev mail sink retrieval 使用 `CodeForChallenge`
  <!-- verified: 2026-05-27 command="cd backend && go test ./internal/auth -count=1 && bash test/scenarios/e2e/p0-101-auth-email-code-login-register/scripts/trigger.sh" evidence="SMTP/dev sink tests passed; P0.101 Mailpit mailSubject observed with mailCode redacted and no URL callback" -->
- [x] 7.5 BDD-Gate: 验证 E2E.P0.101 通过；验证: real frontend/backend/Mailpit 使用同一邮箱完成 register -> logout -> login，注册后和再次登录后 TopBar 展示同一账号 displayName，重复注册同一 email 在发码前被拒绝且不覆盖 displayName，负向断言 `刘哲` / `Liu Zhe` / `liuzhe@example.com` 不出现
  <!-- verified: 2026-05-27 command="bash test/scenarios/env-redeploy.sh all && bash test/scenarios/env-verify.sh && bash test/scenarios/e2e/p0-101-auth-email-code-login-register/scripts/setup.sh && bash test/scenarios/e2e/p0-101-auth-email-code-login-register/scripts/trigger.sh && bash test/scenarios/e2e/p0-101-auth-email-code-login-register/scripts/verify.sh && bash test/scenarios/e2e/p0-101-auth-email-code-login-register/scripts/cleanup.sh" evidence="P0.101 PASS: register/login meStatus=200, duplicate-register finalUrl=/auth/register meStatus=401 mailSubject=not-sent consoleErrors=0 pageErrors=0 httpFailures=0" -->
