# App Shell, Auth Gate, and Settings Entrypoints Checklist

> **版本**: 1.7
> **状态**: active
> **更新日期**: 2026-05-07

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
- [x] 2.7 I18n remediation: 拆分独立 locale 文件并固化语言下拉框契约；验证: focused structural test 断言 `zh` / `en` 各自位于独立 locale 文件且 `messages.ts` 不糅合多语言 map，component / scenario test 断言 TopBar 语言切换是 `select` 下拉框
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
