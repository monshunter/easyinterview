# Email-Code Session Bootstrap Checklist

> **版本**: 3.0
> **状态**: completed
> **更新日期**: 2026-07-19

**关联计划**: [plan](./plan.md)

## Phase 1: Storage and config boundaries

- [x] 1.1 锁定 auth storage；验证: store tests 覆盖 `users`、`user_settings`、`auth_challenges`、`sessions` 表读写，确认 `external_identities` 仅作为 P1 SSO 空表槽存在且 C1 不提供 P0 读写 store 方法，确认无需新增 migration；滑动续期使用 `sessions.updated_at`；若需 `last_seen_at` 或 schema 变更，先停止并修订 ADR-Q1 / B4 owner spec
- [x] 1.2 锁定 config / secret 边界；验证: config tests 覆盖 `SESSION_COOKIE_SECRET`、`AUTH_CHALLENGE_TOKEN_PEPPER` 与当时的 email provider 缺失 fail-fast，固定 `ei_session` cookie name，且 15 分钟 challenge TTL / 30 天 session TTL / 1 分钟第 3 次 rate-limit / dev mail sink 默认值作为 C1 代码常量有测试和包级文档；生产 SMTP 的当前配置合同由 Phase 11 与 A4 最新修订取代
- [x] 1.3 锁定 generated Auth surface 和 session middleware；验证: compile / contract tests 断言 B2 generated `ServerInterface` 的 `startAuthEmailChallenge`、`verifyAuthEmailChallenge`、`getMe`、`deleteMe`、`logout`、`getRuntimeConfig` 均被 C1/A4 wiring 覆盖，public endpoints 为 auth start / verify / runtime-config，logout 为 optional-session / always-clear-cookie 特例，其余 protected endpoint 走 first-party session middleware

## Phase 2: Challenge issue and delivery

- [x] 2.1 实现 `startAuthEmailChallenge`；验证: handler/service tests 覆盖 accepted response、token hash 入库、IP / UA hash、通过 `EmailDispatchEnqueuer` 写入 `email_dispatch` async job，dev mail sink / Mailpit retrieval 收到 code-only delivery，且应用日志不输出 token、完整 URL、邮箱明文或邮件正文
- [x] 2.2 实现 rate-limit / dedupe 基线；验证: tests 覆盖同邮箱或同 IP 1 分钟内第 3 次及以上请求不泄露账号存在性，响应仍符合 B2 schema，dedupe key 不含邮箱明文
- [x] 2.3 接入 B3 `email_dispatch` redacted payload；验证: tests 使用 generated `BuildEmailDispatchPayload` 构造 allowed payload，通过包含 `rawEmailCode` / `emailVerificationUrl` / `recipientEmail` / `emailBody` 任一 redacted field 的 negative case，并确认 async job payload / log / audit 不含 redacted fields
- [x] 2.4 实现 `email_dispatch` producer / handler delivery；验证: tests 覆盖 handler 写入 async job 后返回 B2 `202`、backend-async-runner lease 后调用 `EmailDispatchHandler` 写入 dev mail sink / Mailpit writer、派发失败不泄露 token / 邮箱、runner shutdown / graceful drain 路径可观测，且不启动独立后台执行进程也能完成本地邮件读取

## Phase 3: Verify, session, and current user

- [x] 3.1 实现 `verifyAuthEmailChallenge`；验证: tests 覆盖成功签发 `ei_session` cookie、过期 token、重复 verify、无效 token、session_hash 入库且不返回 cookie 明文
- [x] 3.2 实现 session middleware / current-user resolver；验证: middleware tests 覆盖缺 cookie、无效 session、expired / revoked session 返回 B1 error envelope，active session 更新 `sessions.updated_at` / expiry 且不记录 cookie 明文
- [x] 3.3 实现 `/me`；验证: handler tests 覆盖有效 session 返回账号 email / displayName，缺 cookie 或无效 session 返回 B1 error envelope
- [x] 3.4 实现 logout；验证: tests 覆盖有效 session 撤销、缺 cookie / 无效 session 仍进入 handler 并 Set-Cookie 清除、重复 logout 幂等和无账号存在性泄露
- [x] 3.5 实现 `deleteMe` auth handoff；验证: handler tests 覆盖有效 session 返回 B2 `202 + PrivacyRequestWithJob` 兼容响应并撤销 session，`Idempotency-Key` 或等价 active-request dedupe 使重复请求返回同一 active `privacy_delete` job 或同义终态且不创建重复 job，缺/无效 session 返回 B1 error envelope，实际 privacy_delete runner / 删除矩阵不在 C1 中实现

## Phase 4: Runtime config resolver, privacy, and observability

- [x] 4.1 接入 A4 `/runtime-config` session resolver；验证: tests 断言 C1 只向 A4 handler 注入 session-aware resolver，未登录保持 public response，有效 session 只影响 A4 allowlist 内用户级偏好，secret / internal flag 不出 response
- [x] 4.2 补隐私和可观测红线；验证: privacy tests / grep 确认日志、metric label、audit 不含 raw challenge secret、session cookie、完整邮箱、secret 或 PII 明文
- [x] 4.3 接入 auth metrics / audit 最小事件；验证: F1 registry preflight 断言 `auth_challenge_started_total`、`auth_session_minted_total`、`auth_failure_total` 等 metric 已登记到 F1 baseline metrics 字典或 F1 owner plan；tests / lint 断言 metric 只使用 F1 allowed labels，challenge started / session minted / logout / delete handoff / failure audit 只含 ID / hash / 状态 / trace，不含 token、session id、邮箱明文或 URL

## Phase 5: BDD and handoff

- [x] 5.2 Handoff 给 frontend-shell；验证: backend README 或 package docs 说明 Auth API、cookie 行为、dev mail sink、错误码和前端 pendingAction 接入边界
- [x] 5.3 active-scope 负向搜索通过；验证: backend-auth / API wiring active code 不引入 Bearer token P0 主认证、OAuth / SSO P0 行为、`external_identities` P0 读写 store、明文 token/session 存储、log-only raw code delivery、独立 worker 前置依赖或 out-of-scope AI gateway / voice route 口径；允许 A3 provider adapter 内部使用 provider-side `Authorization: Bearer`，不得误判为浏览器主认证
  <!-- verified: 2026-05-06 method=rg scope=backend/internal/auth,backend/cmd/api,backend/internal/api/generated allowed=negative-doc-comments+internal-session-hash-only -->

## Phase 6: L2 remediation

- [x] 6.1 修复 `DELETE /me` idempotency user scope；验证: store / handler tests 覆盖两个不同用户使用相同 `Idempotency-Key` 时不会复用彼此 active `privacy_delete` handoff，同一用户重复 key 仍返回同一 active request / job
- [x] 6.2 修复 runtime Auth wiring；验证: `cmd/api` wiring tests 断言 auth start / verify / logout / `/me` / `DELETE /me` routes mounted，protected `/me` 经过 C1 session middleware，`/runtime-config` 使用 C1 session-aware resolver 而非 anonymous nil resolver
- [x] 6.3 修复 session cookie Secure policy；验证: handler tests 覆盖 verify minted cookie 与 logout / deleteMe clear cookie 共用 cookie policy，prod/staging Secure=true，dev 降级行为显式可测，属性仍为 HttpOnly + SameSite=Lax + Path=/
- [x] 6.4 修复 challenge rate-limit SQL scope；验证: SQL store tests 断言同邮箱或同 IP 1 分钟窗口统计不再过滤 `status='pending'`，consumed / pending recent challenge 均计入第 3 次 rate-limit / dedupe 基线
- [x] 6.5 修复 logout revoke failure response；验证: handler tests 覆盖有效 session revoke 失败时返回 B1 error envelope 而不是 204，同时清 cookie 且响应不泄露 session id / cookie / secret
- [x] 6.6 修复 runtime Auth secret fail-fast；验证: `cmd/api` builder tests 覆盖 `AUTH_CHALLENGE_TOKEN_PEPPER` / `SESSION_COOKIE_SECRET` 缺失时返回明确错误并不构造 auth email dispatch runtime，本地 dev 不得用空 pepper / 空 session secret 启动 C1 session runtime
- [x] 6.7 修复 logout optional-session resolver error；验证: `cmd/api` route tests 覆盖 cookie-bearing logout 在 session resolver / store error 时返回 B1 error envelope 而不是 204，缺失 / invalid / expired / revoked session 仍保持 optional-session 幂等清 cookie
- [x] 6.8 修复 logout revoke race 的 touch zero-row 归类；验证: middleware tests 覆盖 `ResolveSession` 读到 active session 后 `TouchSession` 返回 `sql.ErrNoRows` 的并发 revoke race，`logout` optional path 仍进入 handler 并幂等清 cookie，required protected path 归类为 B1 `AUTH_UNAUTHORIZED`

## Phase 8: Unified email login and profile completion

- [x] 8.1 OpenAPI contract and generated clients；验证: `AuthEmailStartRequest` 只保留 email / safe request fields，不含 `purpose` 或 `displayName`；`UserContext.profileCompletionRequired` 为必填；新增 `completeMyProfile` / `PATCH /me` request/response；fixtures 更新；`make codegen-openapi` 通过且 generated Go / TS artefacts 与 schema 一致
  <!-- verified: 2026-05-28 commands="make codegen-openapi; python3 scripts/lint/openapi_inventory.py openapi/openapi.yaml; python3 scripts/lint/validate_fixtures.py --repo-root .; make lint-mock-contract; make openapi-diff" evidence="OpenAPI inventory and fixtures synced to 60 operations; generated Go/TS artifacts include completeMyProfile and profileCompletionRequired" -->
- [x] 8.2 Persistence and migration；验证: 新 migration 为 `users` 添加 `profile_completed_at` / `terms_accepted_at` 并 backfill 既有 active displayName 用户；migration up/down 或 dry-run gate 通过；store tests 覆盖新邮箱用户保持未补全、既有用户为已补全
- [x] 8.3 Unified challenge start and verify semantics；验证: service/handler tests 覆盖 start 不检查账号存在、不返回 duplicate register / unknown login 差异；verify 既有邮箱直接登录、新邮箱创建未补全账号和 session；out-of-scope `purpose=signup/login` 不参与当前分支，displayName 不在 verify 前持久化
- [x] 8.4 `/me` and `completeMyProfile`；验证: handler/store tests 覆盖已登录未补全用户 `/me` 返回 200 + `profileCompletionRequired=true`，未登录仍为 B1 error；`PATCH /me` 要求 session、trimmed displayName 非空、`acceptedTerms=true`，成功后返回 `profileCompletionRequired=false` 且不修改邮箱或创建新账号
- [x] 8.5 Privacy / metrics / out-of-scope negative gates；验证: focused grep/test 断言 raw code、session cookie、完整邮箱不进日志/audit/metric label；当前 active backend/openapi/frontend generated truth 不含注册分流完成证据、duplicate-register start rejection、displayName-before-verify、password/OAuth auth wire 或 email URL callback
  <!-- verified: 2026-05-28 commands="make lint-config; rg -n 'purpose=signup|purpose=login|duplicate-register|duplicate register|AuthRegisterScreen|email URL callback|/auth/verify\\?token=|displayName-before-verify|OAuth|password auth|Bearer token' backend/internal/auth backend/cmd/api openapi/openapi.yaml frontend/src/app frontend/src/api -g '!**/*_test.go' -g '!**/*.test.ts' -g '!**/*.test.tsx'" evidence="lint-config PASS; scoped negative search only found backend/internal/auth/doc.go redline comment" -->
- [x] 8.6 BDD-Gate: `BDD.AUTH.EMAIL.001` 由 [BDD checklist](./bdd-checklist.md) 关联 email-code/session/profile-completion owner behavior tests。
- [x] 8.6a E2E-HANDOFF: `E2E.P0.101` 仅覆盖 real frontend/backend/Mailpit 登录闭环；本轮未运行，current-run 状态仍为 `Ready`。
- [x] 8.7 阶段单测完成证据统一为仓库根 `make test`；focused Auth/cmd/api tests 只作开发反馈。

## Phase 9: unauthorized account handler test consolidation

- [x] 9.1 Record scoped `internal/auth` `dupl` RED and confirm the two exact old test names have no external owner/gate consumers.
  <!-- verified: 2026-07-10 method=auth-unauthorized-envelope-test-dupl evidence="Scoped dupl -t 100 reports the GetMe/DeleteMe unauthenticated tests as internal/auth's only clone group; repo-wide exact-name search finds only their declarations." -->
- [x] 9.2 Replace both tests with one table-driven test while preserving named GET/DELETE cases, handler calls, 401 status, JSON envelope and exact error code assertions.
- [x] 9.3 仓库根 `make test` 完成前后端全量单测回归；vet/staticcheck 与 owner/product/docs/pruning 作为独立 closeout gates。

## Phase 10: OPENAPI-007 minimal current-user projection

- [x] 10.1 RED-GATE: generated/store/handler tests fail while public/internal UserContext, SQL scan or builders read/fill `uiLanguage/preferredPracticeLanguage/emailMasked`, or while the public response differs from exact id/email/displayName/profileCompletionRequired；internal `analytics_opt_in` remains explicitly allowed only for runtime-config resolution. Evidence (2026-07-15): after the full-email correction, current generated/handler/frontend contracts remain RED on `emailMasked` versus required `email`.
- [x] 10.2 STORE-GATE: remove obsolete language and `emailMasked` fields from auth types/query/scan；map the existing account email to exact internal `email`；retain only internal `analytics_opt_in` read for runtime-config and keep new-account `user_settings` creation.
  <!-- verified: 2026-07-15 method=focused-auth-store evidence="current-user query/scan returns account email without retired settings fields; focused internal/auth and cmd/api packages PASS" -->
- [x] 10.3 HANDLER-GATE: getMe/completeMyProfile success serialize exact four-field generated UserContext with complete authenticated email and unchanged profile-completion semantics；no `emailMasked` alias and unauthenticated/error behavior remains B1-compliant.
  <!-- verified: 2026-07-15 method=handler-contract evidence="GET/PATCH /me exact full-email response PASS; auth observability test excludes email from audit/metrics/mail sink while allowing it in authenticated response" -->
- [x] 10.4 MIGRATION/HANDOFF: B4 001 Phase 13 drops ui/practice-language/region/timezone with analytics retained；frontend/mock typed consumers compile with `email` and without defaults/aliases before B2 re-freeze.
- [x] 10.5 BDD-GATE: update `BDD.AUTH.EMAIL.001` static owner evidence and `E2E.P0.101` settings handoff；account-delete behavior remains backend contract + frontend Settings BDD, not a new E2E.
- [x] 10.6 REGRESSION-GATE: focused auth/store/runtime-config, root `make test`, generated/codegen, migration, contexts/docs/diff and production old-field zero-reference gates pass before restoring `completed`.

## Phase 11: Production SMTP delivery

- [x] 11.1 A4-HANDOFF: `secrets-and-config/001` 配置 owner 以 RED/GREEN 覆盖 `EMAIL_PROVIDER=mailpit|smtp`、host/port/from、username/password secret、`none|starttls|tls`；staging/prod 禁止 Mailpit/none，删除 `EMAIL_PROVIDER_API_KEY` 当前合同并通过 `make lint-config` 与 zero-reference gate。
  <!-- verified: 2026-07-16 method=a4-phase14 evidence="all four A4 Phase 14 checklist items and focused/lint gates pass" -->
- [x] 11.2 SMTP-TRANSPORT: `backend/internal/auth` RED/GREEN tests 覆盖 Mailpit plain/no-auth、STARTTLS、隐式 TLS、AUTH、TLS >=1.2、unsupported STARTTLS、invalid address 和脱敏错误；实现可注入 transport，不把凭据、邮箱或 raw code写入错误/log/job payload。
  <!-- verified: 2026-07-16 method=focused-auth evidence="SMTP writer and TLS config tests pass for no-auth, STARTTLS, implicit TLS, TLS1.2 floor, injected transport and redacted failures" -->
- [x] 11.3 RUNTIME-WIRING: `backend/cmd/api` RED/GREEN tests 证明 `mailpit` 与 `smtp` 都选择 SMTP writer，标准 SMTP 读取 secret password 并传递 TLS mode/username，未知 provider fail-fast 且不回落 dev sink；Compose/env 模板允许同一组变量选择本地 Mailpit 或外部 SMTP。
  <!-- verified: 2026-07-16 method=focused-api+compose evidence="cmd/api chooses Mailpit none/no-auth or SMTP STARTTLS/auth, rejects unknown provider, and Compose passes provider variables" -->
- [x] 11.4 BDD-Gate: 验证 `BDD.AUTH.EMAIL.002` 通过；domain behavior tests 证明 provider selection、TLS/auth path、delivery failure 与隐私红线，代码层结果不冒充 E2E。
  <!-- verified: 2026-07-16 method=domain-behavior bddChecklist=complete -->
- [x] 11.5 LIVE/REGRESSION: 根 `make test`、`make build`、`make lint-config`、docs/context/diff gates 通过；真实 Mailpit 登录收码闭环 PASS；用户 `.env` 标准 SMTP 完成 TLS/auth/实发，收件人和证据脱敏。MVP 明确只运行一个 active backend 实例。
  <!-- verified: 2026-07-16 evidence="make test/build/lint-config/docs/context/compose/scenario-env gates pass; provider-only full-container Mailpit start->receive-code->verify->session->me passes with host-run app stopped; fresh external SMTP implicit TLS/auth/MAIL FROM/RCPT/DATA and application job succeeded in one attempt; user confirmed EMAIL_FROM_ADDRESS inbox received EasyInterview sign-in code; redacted artifacts record no recipient, code, or credential" -->

## Phase 12: Redis-backed cross-instance delivery secret

- [x] 12.1 RED-STORE: `redis_delivery_secret_store_test.go` 先失败，覆盖 SHA-256 namespaced key、AES-GCM value 不含 code/ref、TTL=5m、两个独立 store 共享读取、miss/expired/decrypt/backend error 脱敏；不新增 config key。证据：首次 focused test 因 `NewRedisDeliverySecretStoreWithClient` / `RedisDeliverySecretStore` 未定义而按预期失败。
- [x] 12.2 GREEN-STORE: 落地 context-aware `DeliverySecretStore.Put/Get/Delete` 与 Redis 实现；test-only `DevMailSink` 适配同一接口；focused auth tests 全绿。证据：`go test ./internal/auth -run TestRedisDeliverySecretStore -count=1`、`go test ./internal/auth -count=1` PASS。
- [x] 12.3 SERVICE-LIFECYCLE: RED/GREEN tests 证明 Put 成功后才 enqueue，Put 失败不 enqueue，payload/enqueue 失败 best-effort delete；SMTP 成功删除、SMTP 失败保留供 retry，delete 失败不触发重复发信且只由 TTL 兜底。证据：RED 分别观察到 enqueue 失败未删除、SMTP 成功未删除及 delete 未调用；GREEN 后 lifecycle focused tests 与 `go test ./internal/auth -count=1` PASS。
- [x] 12.4 RUNTIME-WIRING: `cmd/api` RED/GREEN tests 证明 A4 `redis.url` 被解析、启动 ping fail closed、同一 Redis store 注入 producer/writer并在 shutdown close；Mailpit/SMTP provider selection 不回退进程内 sink。证据：RED 因新 runtime builder / 三参数 `buildAuthService` 未定义而失败；GREEN 后 focused builder tests 与 `go test ./cmd/api -count=1` PASS，启动错误固定脱敏且 ping 失败立即 close。
- [x] 12.5 BDD-Gate: 验证 `BDD.AUTH.EMAIL.003` 通过；domain behavior test 证明 producer/consumer 跨 backend 实例仍投递同一 6 位验证码，Redis/DB/job/error 不泄露 raw code/ref。证据：`TestEmailCodeDeliveryWorksAcrossIndependentRedisBackedInstances` 与 auth package regression PASS；实例 A 生成的 `123456` 由实例 B 的独立 store 解密投递并删除，Redis key/value 与 async payload 无 raw code/ref。
- [x] 12.6 LIVE/REGRESSION: 两个真实 Redis client 完成跨 client Put/Get/Delete + TTL integration；重建 full-container 后 Mailpit challenge->receive->verify/session/me PASS，外部 SMTP 脱敏 smoke PASS；根 `make test`、`make build`、`make lint-config`、docs/context/index/diff/Compose/doctor 全绿。
  <!-- verified: 2026-07-16 evidence="real Redis cross-client integration PASS; Mailpit Chrome full-container login/profile PASS with consoleIssues=0; external SMTP email_dispatch succeeded attempts=1; Redis namespace key count=0 after both deliveries; doctor 6/6; root make test, build, lint-config, docs, context, index, diff and Compose gates PASS" -->
- [x] 12.7 SMTP-LIFECYCLE-REMEDIATION: RED/GREEN tests 证明 runner context 贯穿 writer/Redis/DB/SMTP，建连后停滞会被取消且完整 SMTP 会话有界；DATA 最终成功后 QUIT 失败不返回 retryable error、不重复发信。
  <!-- verified: 2026-07-16 method=tdd evidence="RED compile gate rejected context-free SMTP API; GREEN TestEmailDispatchHandler_PassesRunnerContextToDeliveryWriter, TestSMTPTransportHonorsCancellationAfterConnect and TestSMTPTransportTreatsQuitFailureAfterAcceptedDataAsSuccess PASS" -->
- [x] 12.8 CHALLENGE-COMPENSATION: RED/GREEN tests 证明 Redis Put 失败不创建 `auth_challenges`、不污染 rate-limit；Redis Put 成功但 challenge 创建失败 best-effort 删除 secret，错误仍脱敏。
  <!-- verified: 2026-07-16 method=tdd evidence="TestStartEmailChallengeDoesNotEnqueueWhenDeliverySecretStorageFails and TestStartEmailChallengeDeletesDeliverySecretWhenChallengeCreationFails PASS; compensation remains active after request cancellation" -->
- [x] 12.9 BDD-Gate: 重新验证 `BDD.AUTH.EMAIL.003` domain behavior，覆盖跨实例投递、SMTP cancel/accepted-once 与 Redis Put 失败无 challenge；聚焦测试和根 `make test` 通过后恢复 completed。
  <!-- verified: 2026-07-16 commands="go test ./internal/auth -count=1; REDIS_URL=redis://127.0.0.1:6379/0 go test -tags=integration ./internal/auth -run TestRedisDeliverySecretStoreCrossClientIntegration -count=1; make test; make build" evidence="focused Auth PASS; real Redis cross-client PASS; root Python 566 tests/4481 subtests, Go all packages and frontend 1004 tests PASS; build PASS" -->

## Phase 13: Localized SMTP MIME remediation

- [x] 13.1 SMTP-MIME-RED/GREEN: 标准 MIME reader 回归测试先复现 `zh-CN` 原始 UTF-8 Subject/body 的不合规输出，再实现 RFC 2047 Subject 与 quoted-printable text/plain/text/html；断言两种 part 解码后均保留中文标题、说明和同一 6 位验证码。
  <!-- verified: 2026-07-16 method=tdd-red-green evidence="RED: TestSMTPDeliveryWriterEncodesLocalizedMessageAsStandardsCompliantMIME failed on raw UTF-8 Subject. GREEN: the same test, all SMTPDeliveryWriter tests and go test ./internal/auth -count=1 PASS with RFC 2047 Subject and decoded quoted-printable text/plain/text/html assertions." -->
- [x] 13.2 BDD-Gate: 重新验证 `BDD.AUTH.EMAIL.002` domain behavior，覆盖本地化 MIME 无损解码；focused auth、auth package、根 `make test`、`make build`、owner context/docs/index/diff gates通过后恢复 completed。
  <!-- verified: 2026-07-16 method=domain-behavior bddChecklist=complete evidence="TestSMTPDeliveryWriterEncodesLocalizedMessageAsStandardsCompliantMIME exists and PASS; go test ./internal/auth PASS; root make test PASS with Python 567 tests/4481 subtests, Go all packages and frontend 126 files/1004 tests; make build, A3 terminology lint, both owner contexts, docs-check and diff-check PASS. No E2E status changed." -->

## Phase 14: OPENAPI-008 generic updateMe and account theme

- [x] 14.1 RED/GREEN: generated handler/service tests replace completeMyProfile with updateMe and cover profile-only/theme-only/combined plus empty/partial/invalid input.
- [x] 14.2 STORE: joined getMe projection returns displayPreferences；theme/combined updates use one transaction and reject partial writes.
- [x] 14.3 HANDLER: exact generated UserContext includes displayPreferences and preserves session/profile/full-email privacy semantics.
- [x] 14.4 REGRESSION: focused/full auth, root test, real PostgreSQL migration and old production operation zero-reference gates pass before restoring completed.
