# App Shell, Auth Gate, and Settings Entrypoints Checklist

> **版本**: 1.12
> **状态**: completed
> **更新日期**: 2026-05-27

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

- [x] 3.1 实现认证页面壳；验证: component/route test 覆盖 `auth_login`、`auth_register`、`auth_verify`、`auth_reset`、`auth_logout` 渲染和基本跳转；真实 network wire 只使用 `startAuthEmailChallenge` / `verifyAuthEmailChallenge` / `getMe` / `logout`
- [x] 3.2 实现 `requestAuth(pendingAction)`；验证: route-state test 断言未登录点击 `立即面试` 后进入 login，登录成功恢复 `practice` 并保留 planId / targetJobId / jdId / resumeVersionId / roundId
- [x] 3.3 Auth API contract gate；验证: negative search / focused test 断言 frontend shell 不新增 password auth API、OAuth API、Bearer token auth 或自定义 session storage contract；`auth_reset` 保持 UI shell / stub，真实 API 变更必须先修订 C1 / B2
- [x] 3.4 BDD-Gate: 验证 E2E.P0.002 通过
<!-- verified: 2026-05-07 method=scenario bddChecklist=complete -->
- [x] 3.5 L2 remediation: 修复 `auth_verify` token wire 与 pendingAction auth-only params 泄漏；验证: focused test 断言 verify 请求带 `token` query，恢复后的业务 route params 不含 `email` / `displayName`

## Phase 4: User menu, profile, settings

- [x] 4.1 实现用户菜单入口；验证: component test 断言未登录显示登录/注册，已登录显示 `用户画像`、`设置与隐私`、`退出登录`
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
- [x] 6.5 Phase 6 operation matrix；验证: plan.md 固化 `getRuntimeConfig` / `getMe` / `startAuthEmailChallenge` / `verifyAuthEmailChallenge` / `logout` 的 operationId、fixture、frontend consumer、backend handler、persistence、AI dependency、scenario coverage；context validator 与 docs-check 通过
  <!-- verified: 2026-05-11 method=docs evidence="plan.md Phase 6 operation matrix updated; validate_context.py frontend target PASS; make docs-check PASS" -->

## Phase 7: Historical real passwordless mail-link remediation

- [x] 7.1 `startAuthEmailChallenge` empty-body success；验证: generated client focused test 使用 `new Response(null, { status: 202 })` 断言 `startAuthEmailChallenge` resolve；App auth dispatch tests 断言登录和注册提交邮箱后不会抛 `Unexpected end of JSON input`，并导航到 `auth_verify`
  <!-- verified: 2026-05-27 command="pnpm --filter @easyinterview/frontend test src/api/generatedClient.test.ts src/api/clientFactory.test.ts src/app/routeUrl.test.ts src/app/auth/AuthScreens.test.tsx src/app/AppAuthDispatch.test.tsx src/app/auth/AuthVisual.test.tsx src/app/scenarios/p0-089-url-routing-auth-privacy.test.tsx && pnpm --filter @easyinterview/frontend build" evidence="75 focused tests passed; frontend build passed; generated client accepts 202 empty body and login/register submit routes to auth_verify" -->
- [x] 7.2 `auth_verify` magic-link callback；验证: routeUrl tests 断言 `auth_verify` 独占允许 `token` query，其他 route 仍丢弃 raw token；AuthVerify/App tests 断言进入 `/auth/verify?token=...` 会自动调用 `verifyAuthEmailChallenge`，成功后 `replace` 到恢复 route 或 Home 且 URL 不再含 token；手动输入 token fallback 继续通过
  <!-- verified: 2026-05-27 command="pnpm --filter @easyinterview/frontend test src/api/generatedClient.test.ts src/api/clientFactory.test.ts src/app/routeUrl.test.ts src/app/auth/AuthScreens.test.tsx src/app/AppAuthDispatch.test.tsx src/app/auth/AuthVisual.test.tsx src/app/scenarios/p0-089-url-routing-auth-privacy.test.tsx" evidence="routeUrl/auth/App tests passed; auth_verify consumes token query, calls verifyAuthEmailChallenge, replaces restored URL, and keeps manual token fallback" -->
- [x] 7.3 Local dev Mailpit handoff；验证: `EMAIL_VERIFY_BASE_URL` 默认值、dev-stack README、P0.100 场景材料、backend SMTP writer test、dev CORS origin 派生测试与 token redline 均指向前端 `/auth/verify` callback；后端 API verify endpoint 仍只由前端 generated client 调用；frontend real mode 必须显式配置 `VITE_EI_API_BASE_URL`
  <!-- verified: 2026-05-27 command="go test ./backend/internal/auth -run 'TestSMTPDeliveryWriter|TestSQLChallengeEmailLookup|TestPasswordlessSessionBDD|TestPasswordlessService|TestAuthObservabilityDoesNotLeak' -count=1 && go test ./backend/cmd/api -run 'TestLocalDevCORS|TestBuildAuthServiceUsesMailpitDeliveryWriterWhenConfigured|TestBuildAuthServiceRejectsEmptyAuthSecrets|TestLocalDevCORSAllowsFrontendRealModeOrigins' -count=1 && go test ./backend/cmd/codegen/openapi -count=1 && make lint-config && make lint-mock-contract && make docs-check && bash -n test/scenarios/e2e/p0-100-real-provider-full-funnel-hybrid/scripts/trigger.sh test/scenarios/e2e/p0-100-real-provider-full-funnel-hybrid/scripts/verify.sh && git diff --check" evidence="backend auth/cmd/api/codegen tests passed; CORS origin derives from EMAIL_VERIFY_BASE_URL; fixtures validate; docs/index/link gates and shell syntax passed" -->

## Phase 8: Email-code auth and display-name remediation

- [x] 8.1 Register/login purpose and displayName pass-through；验证: AuthRegisterScreen / AppAuthDispatch tests 断言注册提交 `purpose=signup` + trimmed `displayName` 给 `startAuthEmailChallenge`，登录页提交 `purpose=login` 且不传 displayName，verify 成功恢复业务 route 时不携带 displayName
  <!-- verified: 2026-05-27 command="pnpm --filter @easyinterview/frontend test src/app/auth/AuthScreens.test.tsx src/app/AppAuthDispatch.test.tsx src/app/routeUrl.test.ts src/app/topbar/TopBar.test.tsx" evidence="58 focused tests passed" -->
- [x] 8.2 Six-digit code verify UI；验证: AuthVerifyScreen tests 断言 input 为 numeric one-time-code、最多 6 位、过滤非数字、generated verify query 仍传 `token=<code>`，auth/i18n 文案不含 link/token 口径
  <!-- verified: 2026-05-27 command="pnpm --filter @easyinterview/frontend test src/app/auth/AuthScreens.test.tsx src/app/AppAuthDispatch.test.tsx src/app/routeUrl.test.ts src/app/topbar/TopBar.test.tsx && pnpm --filter @easyinterview/frontend build" evidence="manual code UI tests passed; production build passed" -->
- [x] 8.3 TopBar user fallback cleanup；验证: TopBar tests 断言缺 displayName / emailMasked 时展示中性 fallback，不出现 `刘哲` / `Liu Zhe` / `liuzhe@example.com`
  <!-- verified: 2026-05-27 command="pnpm --filter @easyinterview/frontend test src/app/topbar/TopBar.test.tsx" evidence="TopBar fallback tests passed" -->
- [x] 8.4 BDD-Gate: 验证 E2E.P0.101 通过；验证: Playwright real-mode auth email-code 使用同一邮箱完成 register -> logout -> login，注册后和再次登录后 TopBar 显示同一 displayName，重复注册同一 email 在发码前被拒绝且不覆盖 displayName，邮件和 evidence 不含 magic link 或 `/auth/verify?token=`
  <!-- verified: 2026-05-27 command="bash test/scenarios/e2e/p0-101-auth-email-code-login-register/scripts/setup.sh && bash test/scenarios/e2e/p0-101-auth-email-code-login-register/scripts/trigger.sh && bash test/scenarios/e2e/p0-101-auth-email-code-login-register/scripts/verify.sh && bash test/scenarios/e2e/p0-101-auth-email-code-login-register/scripts/cleanup.sh" evidence="P0.101 PASS: register/login same email, duplicate-register finalUrl=/auth/register mailSubject=not-sent, consoleErrors=0 pageErrors=0 httpFailures=0" -->
