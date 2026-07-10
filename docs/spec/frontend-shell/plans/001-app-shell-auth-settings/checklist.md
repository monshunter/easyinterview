# App Shell, Auth Gate, and Settings Entrypoints Checklist

> **版本**: 1.25
> **状态**: completed
> **更新日期**: 2026-07-10

**关联计划**: [plan](./plan.md)

## Phase 1: App shell and route contract

- [x] 1.1 App 默认渲染 Home，并在 `practice` / `generating` 等当前上下文 route 上按 spec 控制 chrome；验证: route/App focused tests 覆盖默认 route、chrome behavior 和 params preservation。
- [x] 1.2 Route normalization、URL codec 和 hash adapter 统一进入当前 route catalog；验证: route-state tests 覆盖 unsupported input fallback，不 materialize 独立页面。
- [x] 1.3 Runtime config 与 generated client bootstrap 接入；验证: `getRuntimeConfig`、`getMe` authenticated / unauthenticated / profileIncomplete、fixture-backed mock transport 和 unknown scenario fail-loud tests 通过。

## Phase 2: TopBar, display and i18n

- [x] 2.1 TopBar 只展示 `home`、`workspace`、`resume_versions` 三个一级入口；验证: component tests 与 BDD.P0.001 断言当前导航和用户区结构。
- [x] 2.2 显示偏好支持主题、暗色、语言和字体预设；验证: DisplayPreferencesProvider / TopBar focused tests 覆盖登录前后稳定性、`ocean` 默认主题、custom accent fallback 和 local preference priority。
- [x] 2.3 UI i18n 使用独立 locale 文件和 TopBar language dropdown；验证: i18n structure/runtime tests 覆盖 `zh` / `en` 文案切换、browser locale normalization、`Accept-Language` display hint 和登录态不覆盖前端语言设置。
- [x] 2.4 BDD-Gate: `E2E.P0.001` 默认首页与三入口 Shell 通过。
- [x] 2.5 BDD-Gate: `E2E.P0.004` App Shell 中英语言切换通过。

## Phase 3: Auth and pendingAction

- [x] 3.1 Auth pages 只保留 `auth_login`、`auth_verify`、`auth_profile_setup`、`auth_logout`；验证: Auth focused tests 覆盖四个当前 route、email-code generated operations、error copy 和 route transitions。
- [x] 3.2 `requestAuth(pendingAction)` 保存并恢复 safe route params；验证: AppAuthDispatch / pendingAction tests 覆盖未登录触发业务动作、email-code 验证、资料补全 gate 和恢复目标 route。
- [x] 3.3 Auth API contract gate 通过；验证: frontend source and tests 只 wire generated auth operations + first-party session cookie，不引入自定义 session API。
- [x] 3.4 BDD-Gate: `E2E.P0.002` 登录打断后恢复原业务动作通过。
- [x] 3.5 Auth verify recovery 通过；验证: focused tests 覆盖 verify operation success 后 `/me` refresh failure 的 route-gate 表达，以及 public-auth initial probe skip 只消费一次。

## Phase 4: User menu and settings

- [x] 4.1 用户菜单按 `ui-design/src/app.jsx` 呈现头像 chip + dropdown；验证: TopBar component tests 和 browser parity owner 覆盖 menu open/close、settings/logout 分流、desktop right alignment 和 mobile viewport containment。
- [x] 4.2 Settings shell 只保留 `个人资料` / `隐私与数据` 双 tab；验证: Settings visual tests 覆盖账号基础信息、登录与安全 email-code 展示、字体预设、产品信息和隐私数据区。
- [x] 4.3 BDD-Gate: `E2E.P0.032` Dev mock 登录态菜单与退出闭环通过。

## Phase 5: Protected route guard and Home auth visibility

- [x] 5.1 面试业务 route 受 runtime auth guard 保护；验证: AppAuthDispatch tests 覆盖 auth loading 不挂载业务 screen、unauthenticated 进入 `auth_login(pendingAction)`、authenticated 恢复业务 route。
- [x] 5.2 Home recent records 只在 authenticated 状态请求和渲染；验证: Home auth guard tests 覆盖 unauthenticated / loading / auth error 不调用 `listTargetJobs` 且不展示 raw unauthorized body。
- [x] 5.3 Backend protected API proof 通过；验证: backend auth policy / cmd API focused gates 证明业务 APIs behind session middleware。
- [x] 5.4 BDD-Gate: `E2E.P0.102` 未登录首页与面试业务路由登录前置通过。

## Phase 6: Single-entry login and profile setup

- [x] 6.1 Single-entry email-code login 通过；验证: AuthLogin / App dispatch tests 覆盖 email-only challenge body、safe pendingAction round-trip、account-existence privacy copy and route transitions。
- [x] 6.2 Profile setup guard 通过；验证: runtime / route tests 覆盖 verify 后 profileIncomplete、refresh、deep link、logout/relogin 和 cross-browser relogin 均先进入 `auth_profile_setup`。
- [x] 6.3 Profile setup submit 通过；验证: AuthProfileSetup tests 覆盖 trimmed displayName、acceptedTerms、`completeMyProfile`、`/me.profileCompletionRequired=false` 后恢复 pendingAction。
- [x] 6.4 BDD-Gate: `E2E.P0.101` Mailpit email-code single-entry login + profile setup 通过。

## Phase 7: UX simplification and closeout gates

- [x] 7.1 登录页静态帮助说明、settings 双 tab 和 `ocean` 默认主题对齐 `ui-design/`；验证: focused Vitest、visual tests、typecheck and build gates 通过。
- [x] 7.2 Operation matrix 与 context manifest 对齐当前 generated-client and route catalog；验证: `validate_context.py frontend-shell/001 frontend` 通过。
- [x] 7.3 当前清理回归 gate 通过；验证: owner residual grep、frontend focused tests、product-scope context validation、`sync-doc-index --check`、`make docs-check`、`git diff --check`、`make lint-core-loop-pruning-surface`。

## Phase 8: auth alias test lifecycle isolation

- [x] 8.1 `auth_reset` / `auth_register` 同步 normalization tests 在断言后显式 unmount，清除无关 runtime-provider state update（验证：AppAuthDispatch 14 tests 无 act warning、frontend-shell/full frontend test/typecheck/build、owner context/docs gates）
  <!-- verified: 2026-07-10 method=auth-alias-test-lifecycle-isolation evidence="Focused red reproduced one AppRuntimeProvider act warning in each synchronous alias test. Explicit unmount after assertions reuses the file's existing lifecycle pattern. AppAuthDispatch 14/14 and frontend-shell auth/runtime scenarios 72/72 pass warning-free; frontend build and owner/product contexts pass. Full frontend 137 files/829 tests pass and AppAuthDispatch is absent from the remaining warning list; diff/pruning gates pass real_residuals=0." -->

## Phase 9: i18n catalog reachability cleanup

- [x] 9.1 新增 TypeScript AST locale reachability test，先红并精确报告 production 无字面量 consumer 的 key。
  <!-- verified: 2026-07-10 method=locale-reachability-red evidence="Focused localeFiles test failed only the new AST gate and reported exactly 46 production-unreachable or dynamically constructed keys; the other five locale structure tests passed." -->
- [x] 9.2 通过 domain owner 将 Report 动态 key 类型化、Practice 原型文案接回正式 TopBar，并删除其余 zh/en orphan keys 与 Home 原型孤儿属性。
  <!-- verified: 2026-07-10 method=typed-locale-reachability-green evidence="Classified the 46-key red inventory as 13 Report dynamic keys, 3 current Practice prototype keys and 30 true orphans. Report now uses typed MessageKey maps, Practice renders Question/Pause/Resume through typed messages, both locale catalogs shrink from 397 to 367 keys, and the unrendered Home uploadSourceSub prototype property is deleted. The AST reachability gate reports zero keys." -->
- [x] 9.3 运行 focused/full frontend、typecheck/build、UI contract/parity、owner contexts 与 docs/diff/pruning gates。
  <!-- verified: 2026-07-10 method=frontend-locale-parity-regression evidence="Focused locale/Practice/Report tests pass 5 files/34 tests; owner directories pass 46 files/239 tests; full frontend passes 137 files/841 tests. Typecheck/build, 35 UI prototype contracts, Practice Playwright 11 pass plus 1 expected desktop skip, P0.045 real-mode 1 plus 18 tests, P0.059 real-mode 1 plus 18 Vitest plus 3 pytest plus 14 Playwright, five owner/product contexts and pruning/diff gates pass. No scenario environment restart or data cleanup occurred." -->

## Phase 10: auth prototype call-surface pruning

- [x] 10.1 新增 auth 原型参数消费 contract，并先红证明登录页仍接收未读取的 `onSignIn`、资料补全页仍接收未读取的 `nav`。
  <!-- verified: 2026-07-10 method=auth-prototype-call-surface-red evidence="UI contract ran 40 tests: the new consumed-callback contract failed on the existing AuthLoginScreen onSignIn parameter while the prior 39 tests passed; the same source inventory also pins the unread AuthProfileSetupScreen nav parameter and caller arguments." -->
- [x] 10.2 删除两个零读取参数及 `app.jsx` 对应调用方传参；验证：AST auth 参数消费 inventory 归零，验证码登录、资料补全与 pendingAction 回跳代码路径保持原样。
  <!-- verified: 2026-07-10 method=auth-prototype-call-surface-green evidence="Removed only AuthLoginScreen.onSignIn, AuthProfileSetupScreen.nav and the two matching app.jsx arguments. UI contract passes 40/40; Babel binding inventory reports authUnread=[] while preserving AuthVerifyScreen.onSignIn, AuthLoginScreen.nav and AuthProfileSetupScreen.onCompleteProfile." -->
- [x] 10.3 运行 UI contract、focused auth/P0.005、静态浏览器 auth route smoke、full frontend、typecheck/build、owner contexts 与 docs/diff/pruning gates。
  <!-- verified: 2026-07-10 method=auth-prototype-regression-closeout evidence="UI contract passes 40/40 and focused auth/App/P0.005 passes 4 files/49 tests. P0.005 setup/trigger/verify/cleanup passes 8 tests; full frontend passes 137 files/841 tests, typecheck and build pass. Static browser traverses auth_login -> auth_verify -> auth_profile_setup -> home, persists both completion flags and reports no errors; server requests are 200/304. Both owner contexts, diff and pruning gates pass with real_residuals=0. No scenario environment restart or data cleanup occurred." -->

## Phase 11: settings prototype call-surface pruning

- [x] 11.1 新增 Settings 原型参数消费 contract，并先红证明 `SettingsScreen` 与 `app.jsx` 仍保留未读取的 `nav`。
  <!-- verified: 2026-07-10 method=settings-prototype-call-surface-red evidence="UI contract ran 41 tests: the new Settings consumed-dependency contract failed on the existing nav parameter while the prior 40 tests passed; the same contract pins the caller argument and retained font preset dependencies." -->
- [x] 11.2 删除 Settings 的零读取 `nav` 形参与调用方传参；验证：AST Settings 参数消费 inventory 归零，字体预设读写链保持原样。
  <!-- verified: 2026-07-10 method=settings-prototype-call-surface-green evidence="Removed only SettingsScreen.nav and the matching app.jsx argument. UI contract passes 41/41; Babel binding inventory reports settingsUnread=[] while the contract retains fontPreset and setFontPreset at both callee and caller." -->
- [x] 11.3 运行 UI contract、focused Settings/P0.005、静态浏览器 settings tab/font smoke、full frontend、typecheck/build、owner contexts 与 docs/diff/pruning gates。
  <!-- verified: 2026-07-10 method=settings-prototype-regression-closeout evidence="UI contract passes 41/41 and focused Settings/display/P0.005 passes 4 files/24 tests. P0.005 setup/trigger/verify/cleanup passes 8 tests; full frontend passes 137 files/841 tests, typecheck and build pass. Static browser switches Profile/Privacy tabs and applies Modern as Source Serif Pro/Geist with no errors; server requests are 200. Both owner contexts, diff and pruning gates pass with real_residuals=0. No scenario environment restart or data cleanup occurred." -->

## Phase 12: zero-consumer Auth CSS pruning

- [x] 12.1 Add an AuthVisual source RED gate for the link-row wrapper with no formal DOM or prototype consumer.
  <!-- verified: 2026-07-10 method=auth-css-source-red evidence="Focused AuthVisual ran 17 tests: all 16 existing Auth contracts passed and only the new zero-consumer gate failed on .ei-auth-link-row." -->
- [x] 12.2 Delete the CSS rule without an alias, placeholder or removal marker; retain current secondary-link/help selectors.
  <!-- verified: 2026-07-10 method=auth-css-source-green evidence="AuthVisual passes 17/17; ei-auth-link-row is absent outside its negative assertion, while secondary-link, help/help-line and auth-row retain current component consumers." -->
- [x] 12.3 Run focused Auth/P0.005, full frontend, typecheck/build, owner/product contexts and docs/index/diff/pruning gates; then restore the owner to `completed`.
  <!-- verified: 2026-07-10 method=auth-zero-consumer-css-pruning evidence="AuthVisual passes 17, Auth owner passes 6 files/49 tests, P0.005 passes 8, full frontend passes 136 files/841 tests, typecheck/build and both contexts pass. Target runtime inventory is zero and current secondary-link/help/row consumers remain; final docs/index/diff/pruning gates run during closeout. No Bug/retrospective report, environment restart or data cleanup was needed." -->
