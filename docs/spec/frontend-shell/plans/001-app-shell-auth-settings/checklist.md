# App Shell, Auth Gate, and Settings Entrypoints Checklist

> **版本**: 1.17
> **状态**: active
> **更新日期**: 2026-06-12

**关联计划**: [plan](./plan.md)

## Phase 1: App bootstrap and route normalization

- [x] 1.1 建立正式前端 App shell；验证: frontend focused test 断言默认 route 为 `home`，`parse` / `report` / `company_intel` 等上下文 route 保留 App chrome，`practice` / `generating` 隐藏 TopBar
- [x] 1.2 实现 route normalization 与旧 route 拦截；验证: route-state test 覆盖 `welcome`、`growth`、`plan`、`mistakes`、`drill`、`followup`、`experiences`、`star`、`resume`、`onboarding`、`voice` 映射到当前 route 或 Home，不创建独立 screen
- [x] 1.3 接入 runtime config 与 typed API bootstrap；验证: focused test 断言 `getRuntimeConfig` 经 generated client + fixture-backed mock transport 读取，`getMe` 覆盖 `unauthenticated` / `authenticated`，unknown scenario fail loudly，frontend shell 不 import `ui-design/src/data.jsx`
- [x] 1.4 L2 remediation: 删除正式前端 `voice` route alias；验证: route-state / scope focused test 断言 `voice` fallback `home`，且 `normalizeRoute.ts` 不包含 `voice:` alias，语音面试只通过 `practice` params 表达

## Phase 2: TopBar and display controls

- [x] 2.1 实现五入口 TopBar；验证: component test 断言一级导航只包含 `home`、`jd_match`、`workspace`、`resume_versions`、`debrief`
- [x] 2.2 实现全局显示控制；验证: state test 断言主题色、暗色、语言切换在未登录/已登录状态切换后保持稳定
- [x] 2.3 BDD-Gate: 验证 E2E.P0.001 通过
<!-- verified: 2026-05-07 method=scenario bddChecklist=complete -->
- [x] 2.4 I18n remediation: 建立 D1 shell `zh` / `en` message catalog；验证: focused component tests 断言 TopBar、auth shell、profile/settings shell 和 placeholder route shell 在切换 `lang=en` 后展示英文文案，RouteName/testid/params 保持稳定
  <!-- verified: 2026-05-07 method=focused-tests evidence="pnpm --filter @easyinterview/frontend test src/app/i18n/i18nShell.test.tsx PASS; pnpm --filter @easyinterview/frontend test src/app/topbar/TopBar.test.tsx src/app/auth/AuthScreens.test.tsx src/app/screens/ProfileScreen.test.tsx PASS" -->
- [x] 2.5 I18n remediation: 接入 browser locale bootstrap 与 `Accept-Language` display hint；验证: focused runtime/App tests 覆盖浏览器 locale 归一化、English fallback、用户显式切换优先、`getRuntimeConfig` / `getMe` / auth operations 请求带当前 locale header
  <!-- verified: 2026-05-07 method=focused-tests evidence="pnpm --filter @easyinterview/frontend test src/app/i18n/localeRuntime.test.tsx PASS; pnpm --filter @easyinterview/frontend test src/app/runtime/AppRuntimeProvider.test.tsx src/app/AppAuthDispatch.test.tsx src/app/auth/AppPendingAction.test.tsx src/app/i18n/i18nShell.test.tsx PASS" -->
- [x] 2.6 BDD-Gate: 验证 E2E.P0.004 通过
<!-- verified: 2026-05-07 method=scenario bddChecklist=complete -->
- [x] 2.7 I18n remediation: 拆分独立 locale 文件并固化语言切换契约；验证: focused structural test 断言 `zh` / `en` 各自位于独立 locale 文件且 `messages.ts` 不糅合多语言 map；当前控件结构由 002 L2 remediation 按 `ui-design/src/app.jsx` 更新为 TopBar icon dropdown，component / scenario test 断言 `topbar-lang-toggle` 可切换文案与 `Accept-Language`
  <!-- verified: 2026-05-07 method=focused-tests evidence="pnpm --filter @easyinterview/frontend test src/app/i18n/localeFiles.test.ts src/app/i18n/i18nShell.test.tsx src/app/scenarios/p0-004-app-shell-language-switch.test.tsx PASS" -->
- [x] 2.8 I18n remediation: 前端语言偏好独立于登录态；验证: focused runtime/App regression test 断言浏览器语言初始化后，runtime `defaultUiLanguage` 与 `/me.uiLanguage` 不一致也不会覆盖当前 UI 语言或造成 locale 循环请求
  <!-- verified: 2026-05-07 method=focused-tests evidence="pnpm --filter @easyinterview/frontend test src/app/display/DisplayPreferencesProvider.test.tsx src/app/i18n/localeRuntime.test.tsx src/app/topbar/TopBar.test.tsx PASS; pnpm --filter @easyinterview/frontend test PASS (30 files / 138 tests)" -->

## Phase 3: Auth pages and pending action

- [x] 3.1 实现认证页面壳；验证: historical component/route test 曾覆盖 `auth_login`、`auth_register`、`auth_verify`、`auth_reset`、`auth_logout` 渲染和基本跳转；当前 Phase 9 已将 `auth_register` 收敛为 legacy alias，不再 materialize live page；真实 network wire 只使用 generated `startAuthEmailChallenge` / `verifyAuthEmailChallenge` / `getMe` / `completeMyProfile` / `logout`
- [x] 3.2 实现 `requestAuth(pendingAction)`；验证: route-state test 断言未登录点击 `立即面试` 后进入 login，登录成功恢复 `practice` 并保留 planId / targetJobId / jdId / resumeVersionId / roundId
- [x] 3.3 Auth API contract gate；验证: negative search / focused test 断言 frontend shell 不新增 password auth API、OAuth API、Bearer token auth 或自定义 session storage contract；`auth_reset` 保持 UI shell / stub，真实 API 变更必须先修订 C1 / B2
- [x] 3.4 BDD-Gate: 验证 E2E.P0.002 通过
<!-- verified: 2026-05-07 method=scenario bddChecklist=complete -->
- [x] 3.5 L2 remediation: 修复 `auth_verify` token wire 与 pendingAction auth-only params 泄漏；验证: focused test 断言 verify 请求带 `token` query，恢复后的业务 route params 不含 `email` / `displayName`

## Phase 4: User menu, profile, settings

- [x] 4.1 实现用户菜单入口；验证: historical component test 曾断言未登录显示登录/注册；当前 Phase 9 已收敛为未登录只显示单一登录入口，已登录显示 `用户画像`、`设置与隐私`、`退出登录`
- [x] 4.2 实现 settings/profile placeholder shell；验证: route/component test 断言 `profile` 和 `settings` 分离，settings 只维护账号/隐私/字体预设，不恢复旧 Growth / Experiences / Mistakes

## Phase 5: Handoff

- [x] 5.1 记录后续 D2-D6 shell 接入点；验证: frontend README 或 package docs 说明 route table、pendingAction contract、mock runtime 入口和后续 owner 边界
- [x] 5.2 active-scope 负向搜索通过；验证: frontend active code 不含独立 `voice` route、独立 `growth` / `mistakes` / `drill` 页面、prototype data runtime import
- [x] 5.3 记录 UI 真理源 handoff；验证: frontend README 或 package docs 说明正式前端视觉只以 `docs/ui-design/` 与 `ui-design/` 为准，新页面必须先有静态原型，后续实现做原生迁移并通过 parity gate，禁止 AI 自由重设计或引入外部品牌设计系统作为替代参考
- [x] 5.4 Review hardening: 固化真实 build smoke gate；验证: `pnpm --filter @easyinterview/frontend build` 与根 `make build` 均通过，确保 package `build` script 真实化时 HTML/runtime entry 与聚合构建同步可用
  <!-- verified: 2026-05-07 method=build-smoke evidence="pnpm --filter @easyinterview/frontend build PASS; make build PASS" -->

## Phase 6: Auth state and user menu parity remediation

- [x] 6.1 源级复刻已登录用户菜单；验证: `pnpm --filter @easyinterview/frontend test src/app/topbar/TopBar.test.tsx src/app/i18n/i18nShell.test.tsx` 断言已登录用户区先显示头像 chip（initials / displayName / caret），点击后才出现 dropdown，dropdown 含姓名 / masked email header、profile/settings/logout 三项、图标、分隔线、backdrop / Escape 关闭，点击菜单项关闭并派发 `profile` / `settings` / `auth_logout`；负向断言旧 inline 三按钮结构不存在
  <!-- verified: 2026-05-10 method=focused-tests evidence="pnpm --filter @easyinterview/frontend test src/app/topbar/TopBar.test.tsx src/app/i18n/i18nShell.test.tsx PASS (2 files / 9 tests)" -->
- [x] 6.2 修复 fixture-backed dev mock session 状态；验证: `pnpm --filter @easyinterview/frontend test src/api/devMockClient.test.ts src/app/AppAuthDispatch.test.tsx src/app/runtime/AppRuntimeProvider.test.tsx` 断言 `createDevMockClient()` 默认 `/me` 为 401 unauthenticated，`verifyAuthEmailChallenge` 后 `/me` 变 authenticated，`logout` 后 `/me` 变 unauthenticated，显式 `Prefer: example=authenticated|unauthenticated` 仍按 fixture scenario 生效且 unknown scenario fail loudly
  <!-- verified: 2026-05-10 method=focused-tests evidence="pnpm --filter @easyinterview/frontend test src/api/devMockClient.test.ts src/app/AppAuthDispatch.test.tsx src/app/runtime/AppRuntimeProvider.test.tsx PASS (3 files / 13 tests)" -->
- [x] 6.3 BDD-Gate: 验证 E2E.P0.032 通过
  <!-- verified: 2026-05-10 method=scenario evidence="./test/scenarios/e2e/p0-032-dev-mock-auth-state-and-user-menu/scripts/setup.sh && ./test/scenarios/e2e/p0-032-dev-mock-auth-state-and-user-menu/scripts/trigger.sh && ./test/scenarios/e2e/p0-032-dev-mock-auth-state-and-user-menu/scripts/verify.sh && ./test/scenarios/e2e/p0-032-dev-mock-auth-state-and-user-menu/scripts/cleanup.sh PASS" -->
- [x] 6.4 L2 remediation: 浏览器级 authenticated user menu parity gate；验证: `frontend/tests/pixel-parity/topbar.spec.ts` 在 desktop + mobile 两个 chromium project 下通过 mocked Auth API 完成 login → avatar chip → dropdown → logout，断言 dropdown 源码字面量、desktop right alignment、mobile viewport containment 与 logout 后非登录态
  <!-- verified: 2026-05-11 method=playwright evidence="Red: mobile authenticated user menu left=-64.984375 overflow；Green: pnpm --filter @easyinterview/frontend exec playwright test tests/pixel-parity/topbar.spec.ts PASS (22 tests)；pnpm --filter @easyinterview/frontend test:pixel-parity PASS (112 passed)" -->
- [x] 6.5 Phase 6 operation matrix；验证: plan.md 固化 `getRuntimeConfig` / `getMe` / `startAuthEmailChallenge` / `verifyAuthEmailChallenge` / `completeMyProfile` / `logout` 的 operationId、fixture、frontend consumer、backend handler、persistence、AI dependency、scenario coverage；context validator 与 docs-check 通过
  <!-- verified: 2026-05-11 method=docs evidence="plan.md Phase 6 operation matrix updated; validate_context.py frontend target PASS; make docs-check PASS" -->

## Phase 7: Historical real passwordless mail-link remediation

- [x] 7.1 `startAuthEmailChallenge` empty-body success；验证: generated client focused test 使用 `new Response(null, { status: 202 })` 断言 `startAuthEmailChallenge` resolve；App auth dispatch tests 断言登录和注册提交邮箱后不会抛 `Unexpected end of JSON input`，并导航到 `auth_verify`
  <!-- verified: 2026-05-27 command="pnpm --filter @easyinterview/frontend test src/api/generatedClient.test.ts src/api/clientFactory.test.ts src/app/routeUrl.test.ts src/app/auth/AuthScreens.test.tsx src/app/AppAuthDispatch.test.tsx src/app/auth/AuthVisual.test.tsx src/app/scenarios/p0-089-url-routing-auth-privacy.test.tsx && pnpm --filter @easyinterview/frontend build" evidence="75 focused tests passed; frontend build passed; generated client accepts 202 empty body and login/register submit routes to auth_verify" -->
- [x] 7.2 `auth_verify` magic-link callback；验证: routeUrl tests 断言 `auth_verify` 独占允许 `token` query，其他 route 仍丢弃 raw token；AuthVerify/App tests 断言进入 `/auth/verify?token=...` 会自动调用 `verifyAuthEmailChallenge`，成功后 `replace` 到恢复 route 或 Home 且 URL 不再含 token；手动输入 token fallback 继续通过
  <!-- verified: 2026-05-27 command="pnpm --filter @easyinterview/frontend test src/api/generatedClient.test.ts src/api/clientFactory.test.ts src/app/routeUrl.test.ts src/app/auth/AuthScreens.test.tsx src/app/AppAuthDispatch.test.tsx src/app/auth/AuthVisual.test.tsx src/app/scenarios/p0-089-url-routing-auth-privacy.test.tsx" evidence="routeUrl/auth/App tests passed; auth_verify consumes token query, calls verifyAuthEmailChallenge, replaces restored URL, and keeps manual token fallback" -->
- [x] 7.3 Local dev Mailpit handoff；验证: `EMAIL_VERIFY_BASE_URL` 默认值、dev-stack README、P0.100 场景材料、backend SMTP writer test、dev CORS origin 派生测试与 token redline 均指向前端 `/auth/verify` callback；后端 API verify endpoint 仍只由前端 generated client 调用；frontend real mode 必须显式配置 `VITE_EI_API_BASE_URL`
  <!-- verified: 2026-05-27 command="go test ./backend/internal/auth -run 'TestSMTPDeliveryWriter|TestSQLChallengeEmailLookup|TestPasswordlessSessionBDD|TestPasswordlessService|TestAuthObservabilityDoesNotLeak' -count=1 && go test ./backend/cmd/api -run 'TestLocalDevCORS|TestBuildAuthServiceUsesMailpitDeliveryWriterWhenConfigured|TestBuildAuthServiceRejectsEmptyAuthSecrets|TestLocalDevCORSAllowsFrontendRealModeOrigins' -count=1 && go test ./backend/cmd/codegen/openapi -count=1 && make lint-config && make lint-mock-contract && make docs-check && bash -n test/scenarios/e2e/p0-100-real-provider-full-funnel-hybrid/scripts/trigger.sh test/scenarios/e2e/p0-100-real-provider-full-funnel-hybrid/scripts/verify.sh && git diff --check" evidence="backend auth/cmd/api/codegen tests passed; CORS origin derives from EMAIL_VERIFY_BASE_URL; fixtures validate; docs/index/link gates and shell syntax passed" -->

## Phase 8: Historical email-code auth and display-name remediation

> 本阶段是 2026-05-27 的历史完成记录，已被 Phase 9 的单入口邮箱登录与资料补全语义取代。以下 `AuthRegisterScreen`、`purpose=signup/login`、duplicate-register 证据不得再作为当前验收口径。

- [x] 8.1 Historical register/login purpose and displayName pass-through；验证: historical AuthRegisterScreen / AppAuthDispatch tests 曾断言注册提交 `purpose=signup` + trimmed `displayName` 给 `startAuthEmailChallenge`，登录页提交 `purpose=login` 且不传 displayName，verify 成功恢复业务 route 时不携带 displayName；当前统一入口只提交 email，见 Phase 9.2
  <!-- verified: 2026-05-27 command="pnpm --filter @easyinterview/frontend test src/app/auth/AuthScreens.test.tsx src/app/AppAuthDispatch.test.tsx src/app/routeUrl.test.ts src/app/topbar/TopBar.test.tsx" evidence="58 focused tests passed" -->
- [x] 8.2 Six-digit code verify UI；验证: AuthVerifyScreen tests 断言 input 为 numeric one-time-code、最多 6 位、过滤非数字、generated verify query 仍传 `token=<code>`，auth/i18n 文案不含 link/token 口径
  <!-- verified: 2026-05-27 command="pnpm --filter @easyinterview/frontend test src/app/auth/AuthScreens.test.tsx src/app/AppAuthDispatch.test.tsx src/app/routeUrl.test.ts src/app/topbar/TopBar.test.tsx && pnpm --filter @easyinterview/frontend build" evidence="manual code UI tests passed; production build passed" -->
- [x] 8.3 TopBar user fallback cleanup；验证: TopBar tests 断言缺 displayName / emailMasked 时展示中性 fallback，不出现 `刘哲` / `Liu Zhe` / `liuzhe@example.com`
  <!-- verified: 2026-05-27 command="pnpm --filter @easyinterview/frontend test src/app/topbar/TopBar.test.tsx" evidence="TopBar fallback tests passed" -->
- [x] 8.4 Historical BDD-Gate: 验证 E2E.P0.101 通过；验证: historical Playwright real-mode auth email-code 使用同一邮箱完成 register -> logout -> login，注册后和再次登录后 TopBar 显示同一 displayName，重复注册同一 email 在发码前被拒绝且不覆盖 displayName，邮件和 evidence 不含 magic link 或 `/auth/verify?token=`；当前 P0.101 验收见 Phase 9.6
  <!-- verified: 2026-05-27 command="bash test/scenarios/e2e/p0-101-auth-email-code-login-register/scripts/setup.sh && bash test/scenarios/e2e/p0-101-auth-email-code-login-register/scripts/trigger.sh && bash test/scenarios/e2e/p0-101-auth-email-code-login-register/scripts/verify.sh && bash test/scenarios/e2e/p0-101-auth-email-code-login-register/scripts/cleanup.sh" evidence="P0.101 PASS: register/login same email, duplicate-register finalUrl=/auth/register mailSubject=not-sent, consoleErrors=0 pageErrors=0 httpFailures=0" -->

## Phase 9: Unified email login and first-login profile setup

- [x] 9.1 UI truth source and route catalog；验证: `docs/ui-design/auth-and-entry.md`、`ui-design/src/screen-auth.jsx`、`ui-design/src/app.jsx` 已移除可见注册入口，新增 `auth_profile_setup`，TopBar 未登录态只显示登录；正式前端 route tests 断言 `auth_register` 不会 materialize live page，只能 normalize 到当前保留 route 或 Home
  <!-- verified: 2026-05-28 commands="node --test ui-design/ui-design-contract.test.mjs; pnpm --filter @easyinterview/frontend test" evidence="ui-design auth contract PASS; frontend tests cover auth_register legacy normalization and no live route materialization" -->
- [x] 9.2 Unified login start；验证: AuthLogin/App dispatch tests 断言 `startAuthEmailChallenge` 请求体只包含 email，不包含 `returnTo`、`purpose`、`displayName`、password、OAuth 或注册/登录选择；safe pendingAction 只随 `auth_verify` route params round-trip；发码成功进入 `auth_verify`，错误提示不泄露账号是否存在
  <!-- verified: 2026-05-28 commands="pnpm --filter @easyinterview/frontend test src/app/auth/AuthScreens.test.tsx src/app/auth/AuthVisual.test.tsx src/app/AppAuthDispatch.test.tsx; pnpm --filter @easyinterview/frontend typecheck" evidence="42 focused tests and typecheck PASS; AuthLoginScreen red/green covers email-only body and strips legacy returnTo from verify params while preserving safe pendingAction params" -->
- [x] 9.3 Profile completion routing；验证: App runtime/route tests 覆盖 verify 后 `/me.profileCompletionRequired=true`、刷新、已登录直开业务 URL、logout 后重新登录、不同 browser context 重新登录均进入 `auth_profile_setup`，未登录直开 `auth_profile_setup` 回到 `auth_login`，并且资料补全前不恢复 pendingAction
  <!-- verified: 2026-05-28 commands="pnpm --filter @easyinterview/frontend test; P0.101 real scenario" evidence="frontend route/runtime tests and real Playwright scenario cover verify, refresh, deep link, cross-browser relogin, and logout/relogin profile setup guard" -->
- [x] 9.4 Profile completion submit；验证: AuthProfileSetup tests 断言 displayName trim 后非空、`acceptedTerms=true` 才能调用 generated `completeMyProfile`；成功后刷新 `/me` 并只在 `profileCompletionRequired=false` 时恢复 pendingAction 或 Home；若后端仍返回 `profileCompletionRequired=true`，页面保持资料补全错误状态且不恢复业务 route；displayName 不进入 URL / pendingAction / 业务 route params
  <!-- verified: 2026-05-28 commands="pnpm --filter @easyinterview/frontend test src/app/auth/AuthScreens.test.tsx src/app/auth/AuthVisual.test.tsx src/app/AppAuthDispatch.test.tsx; pnpm --filter @easyinterview/frontend typecheck; pnpm --filter @easyinterview/frontend build" evidence="42 focused tests, typecheck and production build PASS; AuthProfileSetup red/green blocks pendingAction restore until completeMyProfile returns profileCompletionRequired=false" -->
- [x] 9.5 OpenAPI / fixture / dev mock handoff；验证: generated frontend client 含 `completeMyProfile` 与 `UserContext.profileCompletionRequired`；fixtures 覆盖 `profileIncomplete` 与 completion success；`createDevMockClient` 支持 unauthenticated -> verify profileIncomplete -> complete profile -> authenticated -> logout -> relogin completed 的连续状态流
  <!-- verified: 2026-05-28 commands="python3 scripts/lint/openapi_inventory.py openapi/openapi.yaml; python3 scripts/lint/validate_fixtures.py --repo-root .; pnpm --filter @easyinterview/frontend test" evidence="60-operation OpenAPI/fixtures PASS; devMockClient tests cover profileIncomplete and completion flow" -->
- [x] 9.6 BDD-Gate: 验证 E2E.P0.101 通过；验证: real frontend/backend/Mailpit 场景覆盖单入口新邮箱首次登录进入资料补全、资料补全页刷新仍停留、关闭/换浏览器后同邮箱重新登录仍停留、完成资料后 `/me.profileCompletionRequired=false` 并显示 displayName、退出后同邮箱再次登录不再进入资料补全；负向断言注册按钮、`auth_register` live page、`purpose=signup/login` request body、displayName-before-verify、旧 magic-link URL 不出现
  <!-- verified: 2026-05-28 command="bash test/scenarios/e2e/p0-101-auth-email-code-login-register/scripts/cleanup.sh && bash test/scenarios/e2e/p0-101-auth-email-code-login-register/scripts/setup.sh && bash test/scenarios/e2e/p0-101-auth-email-code-login-register/scripts/trigger.sh && bash test/scenarios/e2e/p0-101-auth-email-code-login-register/scripts/verify.sh && bash test/scenarios/e2e/p0-101-auth-email-code-login-register/scripts/cleanup.sh" evidence="P0.101 PASS: profile-required gates PASS refresh=profile-setup deepLink=profile-setup crossBrowser=profile-setup logoutRelogin=profile-setup authStartBodyKeys=email authRegisterLivePage=absent topbarRegister=absent" -->

## Phase 10: Unauthenticated interview route guard remediation

- [x] 10.1 UI truth source for signed-out Home；验证: `node --test ui-design/ui-design-contract.test.mjs` 断言 `ui-design/src/app.jsx` 向 Home 传 `signedIn`，`ui-design/src/screen-home.jsx` 仅在 signed-in 状态渲染 Recent mock interviews；`docs/ui-design/auth-and-entry.md` 锁定未登录不展示 Recent 模块
  <!-- verified: 2026-05-28 command="node --test ui-design/ui-design-contract.test.mjs" evidence="16 UI contract tests passed; Home recent mock interviews are signed-in only" -->
- [x] 10.2 Frontend runtime protected route guard；验证: `pnpm --filter @easyinterview/frontend test src/app/AppAuthDispatch.test.tsx` 覆盖未登录直开 `workspace` / `practice` / `report` / `jd_match` / `profile` / `settings` 等业务 route 时进入 `auth_login(pendingAction)`，auth loading 期间不挂载业务 screen 或发起受保护 API
  <!-- verified: 2026-05-28 command="pnpm --filter @easyinterview/frontend test src/app/AppAuthDispatch.test.tsx" evidence="AppAuthDispatch route guard tests passed inside focused auth-gate suite; protected direct URL rewrites to auth_login and auth loading renders auth-route-gate without business API calls" -->
- [x] 10.3 Home recent data fetch guard；验证: `pnpm --filter @easyinterview/frontend test src/app/screens/home/HomeRecentMocks.test.tsx src/app/screens/home/HomeAuthGate.test.tsx` 覆盖 unauthenticated / loading / auth error 下不调用 `listTargetJobs`、不渲染 `home-recent-mocks`、不显示 raw `AUTH_UNAUTHORIZED`，authenticated 下仍按 fixture 排序和限制 12 张卡
  <!-- verified: 2026-05-28 command="pnpm --filter @easyinterview/frontend test src/app/screens/home/HomeRecentMocks.test.tsx src/app/screens/home/HomeAuthGate.test.tsx src/app/AppAuthDispatch.test.tsx" evidence="3 files / 24 tests passed; signed-out Home hides home-recent-mocks and Home CTAs redirect with pendingAction" -->
- [x] 10.4 Backend protected API proof；验证: `cd backend && go test ./internal/auth -run TestSessionPolicyClassifiesPublicOptionalAndProtectedOperations -count=1 && go test ./cmd/api -run 'TestBuildAPIHandlerMounts(TargetJobRoutes|UploadPresign|ResumeRoutes|PracticeAndProfileRoutes|ReportRoutes|JobRoute)BehindSessionMiddleware|TestJDMatchRoutesRequireSessionOnAllRoutes' -count=1` 证明面试相关 API 保持 session middleware 保护
  <!-- verified: 2026-05-28 command="cd backend && go test ./internal/auth -run TestSessionPolicyClassifiesPublicOptionalAndProtectedOperations -count=1; cd backend && go test ./cmd/api -run 'TestBuildAPIHandlerMounts(TargetJobRoutes|UploadPresign|ResumeRoutes|PracticeAndProfileRoutes|ReportRoutes|JobRoute)BehindSessionMiddleware|TestJDMatchRoutesRequireSessionOnAllRoutes' -count=1" evidence="backend auth policy and cmd/api route middleware gates passed; added practice/profile route middleware regression coverage" -->
- [x] 10.5 BDD-Gate: 验证 E2E.P0.102 通过；验证: 新场景脚本执行 focused frontend + backend gates，确认未登录 Home 无 Recent 模块、受保护业务 route 先进入登录、受保护 API 返回 B1 `AUTH_UNAUTHORIZED` envelope；wrapper 必须从 `trigger.log` 校验 scenario runner marker、目标 Vitest 文件、Go package `ok` 行和每个后端 focused test 的 `--- PASS` 名称，并阻止 no-test / skip / fail 证据误判
  <!-- verified: 2026-05-28 command="bash test/scenarios/e2e/p0-102-auth-gated-interview-routes/scripts/setup.sh; bash test/scenarios/e2e/p0-102-auth-gated-interview-routes/scripts/trigger.sh; bash test/scenarios/e2e/p0-102-auth-gated-interview-routes/scripts/verify.sh; bash test/scenarios/e2e/p0-102-auth-gated-interview-routes/scripts/cleanup.sh" evidence="P0.102 PASS with strict wrapper evidence: UI contract, 3 Vitest files / 24 tests, backend session policy PASS, cmd/api route middleware named PASS markers, result.json result=PASS" -->

## Phase 11: Auth verify recovery and skipped probe consumption

- [x] 11.1 Consume public-auth initial probe skip once；验证: `pnpm --filter @easyinterview/frontend test src/app/runtime/AppRuntimeProvider.test.tsx` 覆盖 `skipInitialAuthProbe` 只消费首次 probe，`refreshAuth(user)` 直接提交 authenticated user 后，requestOptions / language 切换触发真实 `/me` refresh 而不是把 auth state 重置为 unauthenticated
  <!-- verified: 2026-05-28 command="pnpm --filter @easyinterview/frontend test src/app/runtime/AppRuntimeProvider.test.tsx src/app/AppAuthDispatch.test.tsx" evidence="AppRuntimeProvider 6 tests passed; Red reproduced missing /me refresh after requestOptions change, Green probes /me and remains authenticated" -->
- [x] 11.2 Separate verify success from post-verify `/me` failure；验证: `pnpm --filter @easyinterview/frontend test src/app/AppAuthDispatch.test.tsx` 覆盖 `verifyAuthEmailChallenge` 成功但后续 `/me` 失败时不显示验证码失败、不停留在 `auth_verify`，并通过 runtime auth refresh / route gate 表达 auth/profile loading 或 error
  <!-- verified: 2026-05-28 command="pnpm --filter @easyinterview/frontend test src/app/runtime/AppRuntimeProvider.test.tsx src/app/AppAuthDispatch.test.tsx" evidence="AppAuthDispatch 13 tests passed; Red reproduced verify page stuck with code failure, Green navigates to auth-route-gate for pending workspace" -->
- [x] 11.3 Phase 11 regression gates；验证: focused frontend auth gates 通过，`pnpm --filter @easyinterview/frontend typecheck` 通过，`python3 .agent-skills/implement/shared/scripts/validate_context.py --context docs/spec/frontend-shell/plans/001-app-shell-auth-settings/context.yaml --docs-root docs --target frontend` 通过
  <!-- verified: 2026-05-28 command="pnpm --filter @easyinterview/frontend test src/app/runtime/AppRuntimeProvider.test.tsx src/app/AppAuthDispatch.test.tsx; pnpm --filter @easyinterview/frontend typecheck; python3 .agent-skills/implement/shared/scripts/validate_context.py --context docs/spec/frontend-shell/plans/001-app-shell-auth-settings/context.yaml --docs-root docs --target frontend" evidence="19 focused auth tests passed; frontend typecheck passed; context validator resolved plan/checklist/spec/bdd docs with specVersion.to=1.21" -->

## Phase 12: UX funnel simplification alignment (D-16 / D-21)

- [ ] 12.1 删除 `auth_reset` route 与 `AuthResetScreen`，登录页改为静态帮助说明；验证: focused Vitest 断言 `auth_reset` route key 与 `/auth/reset` path 归一回 `auth_login` 且不 materialize 独立页面；`AuthResetScreen.tsx` 文件、`auth/index.ts` 导出、`App.tsx` 分支、`AuthShell` routeName、zh/en `auth.reset.*` / `auth.forgotPassword` 词条全部删除；`AuthLoginScreen` 无"忘记密码"导航按钮，渲染与 `ui-design/src/screen-auth.jsx` 一致的静态帮助说明（一个邮箱一个账号 + 收不到验证码下一步可重发/换邮箱），zh/en 双语断言
- [ ] 12.2 设置页收敛为 `个人资料` / `隐私与数据` 双 tab 并对齐登录与安全口径；验证: SettingsScreen 测试断言只有 profile / privacy 两个 tab，`settings-notifications-placeholder` / `settings-subscription-placeholder` 与对应 i18n 词条删除；个人资料 tab 按原型含账号基础信息、`登录与安全` 仅一行 `邮箱验证码 · 无密码`、字体预设、产品信息；"密码 / 两步验证"旧口径词条删除
- [ ] 12.3 默认主题与 fallback 改为 `ocean`；验证: DisplayPreferencesProvider 测试断言默认 `theme === "ocean"`、无效持久化值 fallback `ocean`；TopBar custom accent seed fallback 为 `CUSTOM_ACCENT_SEEDS.ocean`；主题菜单仍含四预设 + customAccent；p0-005 visual smoke 与相关 pixel-parity 默认主题断言同步更新
- [ ] 12.4 Phase 12 operation matrix 固化；验证: plan.md Phase 12.4 矩阵存在（UI-only N/A 行 + auth operations 维持 Phase 9 matrix），context validator 通过
- [ ] 12.5 Phase 12 回归与零残留 gate；验证: focused Vitest 全部更新后通过；`frontend/src` 负向搜索 `auth_reset` / `AuthResetScreen` / `forgotPassword` / 忘记密码 / 两步验证 / `settings-notifications-placeholder` / `settings-subscription-placeholder` 零残留（负向断言测试除外）；`pnpm --filter @easyinterview/frontend typecheck`、`pnpm --filter @easyinterview/frontend test`、`pnpm --filter @easyinterview/frontend build` 通过
