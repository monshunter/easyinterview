# App Shell, Auth Gate, and Settings Entrypoints Checklist

> **版本**: 1.42
> **状态**: active
> **更新日期**: 2026-07-20

**关联计划**: [plan](./plan.md)

> Phase 1-13 的勾选项是历史证据；Phase 14 取代其中账号 chip/dropdown、Settings tab 与 font preset 的旧正向口径。

## Phase 1: App shell and route contract

- [x] 1.1 App 默认渲染 Home，并在 `practice` / `generating` 等当前上下文 route 上按 spec 控制 chrome；验证: route/App focused tests 覆盖默认 route、chrome behavior 和 params preservation。
- [x] 1.2 Route normalization、URL codec 和 hash adapter 统一进入当前 route catalog；验证: route-state tests 覆盖 unsupported input fallback，不 materialize 独立页面。
- [x] 1.3 Runtime config 与 generated client bootstrap 接入；验证: `getRuntimeConfig`、`getMe` authenticated / unauthenticated / profileIncomplete、fixture-backed mock transport 和 unknown scenario fail-loud tests 通过。

## Phase 2: TopBar, display and i18n

- [x] 2.2 显示偏好支持主题、暗色、语言和字体预设；验证: DisplayPreferencesProvider / TopBar focused tests 覆盖登录前后稳定性、`ocean` 默认主题、custom accent fallback 和 local preference priority。
- [x] 2.3 UI i18n 使用独立 locale 文件和 TopBar language dropdown；验证: i18n structure/runtime tests 覆盖 `zh` / `en` 文案切换、browser locale normalization、`Accept-Language` display hint 和登录态不覆盖前端语言设置。

## Phase 3: Auth and pendingAction

- [x] 3.1 Auth pages 只保留 `auth_login`、`auth_verify`、`auth_profile_setup`、`auth_logout`；验证: Auth focused tests 覆盖四个当前 route、email-code generated operations、error copy 和 route transitions。
- [x] 3.2 `requestAuth(pendingAction)` 保存并恢复 safe route params；验证: AppAuthDispatch / pendingAction tests 覆盖未登录触发业务动作、email-code 验证、资料补全 gate 和恢复目标 route。
- [x] 3.3 Auth API contract gate 通过；验证: frontend source and tests 只 wire generated auth operations + first-party session cookie，不引入自定义 session API。
- [x] 3.5 Auth verify recovery 通过；验证: focused tests 覆盖 verify operation success 后 `/me` refresh failure 的 route-gate 表达，以及 public-auth initial probe skip 只消费一次。

## Phase 4: User menu and settings

- [x] 4.1 用户菜单按 `frontend/src` 呈现头像 chip + dropdown；验证: TopBar component tests 和 browser parity owner 覆盖 menu open/close、settings/logout 分流、desktop right alignment 和 mobile viewport containment。
- [x] 4.2 Settings shell 只保留 `个人资料` / `隐私与数据` 双 tab；验证: Settings visual tests 覆盖账号基础信息、登录与安全 email-code 展示、字体预设、产品信息和隐私数据区。

## Phase 5: Protected route guard and Home auth visibility

- [x] 5.1 面试业务 route 受 runtime auth guard 保护；验证: AppAuthDispatch tests 覆盖 auth loading 不挂载业务 screen、unauthenticated 进入 `auth_login(pendingAction)`、authenticated 恢复业务 route。
- [x] 5.2 Home recent records 只在 authenticated 状态请求和渲染；验证: Home auth guard tests 覆盖 unauthenticated / loading / auth error 不调用 `listTargetJobs` 且不展示 raw unauthorized body。
- [x] 5.3 Backend protected API proof 通过；验证: backend auth policy / cmd API focused gates 证明业务 APIs behind session middleware。

## Phase 6: Single-entry login and profile setup

- [x] 6.1 Single-entry email-code login 通过；验证: AuthLogin / App dispatch tests 覆盖 email-only challenge body、safe pendingAction round-trip、account-existence privacy copy and route transitions。
- [x] 6.2 Profile setup guard 通过；验证: runtime / route tests 覆盖 verify 后 profileIncomplete、refresh、deep link、logout/relogin 和 cross-browser relogin 均先进入 `auth_profile_setup`。
- [x] 6.3 Profile setup submit 通过；验证: AuthProfileSetup tests 覆盖 trimmed displayName、acceptedTerms、`updateMe`、`/me.profileCompletionRequired=false` 后恢复 pendingAction。
- [x] 6.4 BDD-Gate: `BDD.SHELL.AUTH.001` 由 [BDD checklist](./bdd-checklist.md) 关联 auth guard、profile setup、pendingAction 与 settings owner behavior tests。
- [x] 6.4a E2E-HANDOFF: `E2E.P0.101` 仅覆盖 Mailpit email-code/profile setup；本轮未运行，current-run 状态仍为 `Ready`。
- [x] 6.5 阶段单测完成证据统一为仓库根 `make test`；focused shell/auth tests 只作开发反馈。

## Phase 7: UX simplification and closeout gates

- [x] 7.1 登录页静态帮助说明、settings 双 tab 和 `ocean` 默认主题对齐 `frontend/`；验证: focused Vitest、visual tests、typecheck and build gates 通过。
- [x] 7.2 Operation matrix 与 context manifest 对齐当前 generated-client and route catalog；验证: `validate_context.py frontend-shell/001 frontend` 通过。
- [x] 7.3 当前清理回归 gate 通过；验证: owner residual grep、frontend focused tests、product-scope context validation、`sync-doc-index --check`、`make docs-check`、`git diff --check`、`make lint-core-loop-pruning-surface`。

## Phase 8: auth alias test lifecycle isolation

- [x] 8.1 `auth_reset` / `auth_register` 同步 normalization tests 在断言后显式 unmount，清除无关 runtime-provider state update（验证：AppAuthDispatch 14 tests 无 act warning、frontend-shell/full frontend test/typecheck/build、owner context/docs gates）

## Phase 9: i18n catalog reachability cleanup

- [x] 9.1 新增 TypeScript AST locale reachability test，先红并精确报告 production 无字面量 consumer 的 key。
- [x] 9.2 通过 domain owner 将 Report 动态 key 类型化、Practice 原型文案接回正式 TopBar，并删除其余 zh/en orphan keys 与 Home 原型孤儿属性。
  <!-- verified: 2026-07-10 method=typed-locale-reachability-green evidence="Classified the 46-key red inventory as 13 Report dynamic keys, 3 current Practice prototype keys and 30 true orphans. Report now uses typed MessageKey maps, Practice renders Question/Pause/Resume through typed messages, both locale catalogs shrink from 397 to 367 keys, and the unrendered Home uploadSourceSub prototype property is deleted. The AST reachability gate reports zero keys." -->
- [x] 9.3 仓库根 `make test` 完成前后端全量单测回归；typecheck/build、UI contract/parity、owner contexts 与 docs/diff/pruning 作为独立 gates。

## Phase 10: auth prototype call-surface pruning

- [x] 10.1 新增 auth 原型参数消费 contract，并先红证明登录页仍接收未读取的 `onSignIn`、资料补全页仍接收未读取的 `nav`。
  <!-- verified: 2026-07-10 method=auth-prototype-call-surface-red evidence="UI contract ran 40 tests: the new consumed-callback contract failed on the existing AuthLoginScreen onSignIn parameter while the prior 39 tests passed; the same source inventory also pins the unread AuthProfileSetupScreen nav parameter and caller arguments." -->
- [x] 10.2 删除两个零读取参数及 `app.jsx` 对应调用方传参；验证：AST auth 参数消费 inventory 归零，验证码登录、资料补全与 pendingAction 回跳代码路径保持原样。
  <!-- verified: 2026-07-10 method=auth-prototype-call-surface-green evidence="Removed only AuthLoginScreen.onSignIn, AuthProfileSetupScreen.nav and the two matching app.jsx arguments. UI contract passes 40/40; Babel binding inventory reports authUnread=[] while preserving AuthVerifyScreen.onSignIn, AuthLoginScreen.nav and AuthProfileSetupScreen.onCompleteProfile." -->

## Phase 11: settings prototype call-surface pruning

- [x] 11.1 新增 Settings 原型参数消费 contract，并先红证明 `SettingsScreen` 与 `app.jsx` 仍保留未读取的 `nav`。
  <!-- verified: 2026-07-10 method=settings-prototype-call-surface-red evidence="UI contract ran 41 tests: the new Settings consumed-dependency contract failed on the existing nav parameter while the prior 40 tests passed; the same contract pins the caller argument and retained font preset dependencies." -->
- [x] 11.2 删除 Settings 的零读取 `nav` 形参与调用方传参；验证：AST Settings 参数消费 inventory 归零，字体预设读写链保持原样。
  <!-- verified: 2026-07-10 method=settings-prototype-call-surface-green evidence="Removed only SettingsScreen.nav and the matching app.jsx argument. UI contract passes 41/41; Babel binding inventory reports settingsUnread=[] while the contract retains fontPreset and setFontPreset at both callee and caller." -->

## Phase 12: zero-consumer Auth CSS pruning

- [x] 12.1 Add an AuthVisual source RED gate for the link-row wrapper with no formal DOM or prototype consumer.
- [x] 12.2 Delete the CSS rule without an alias, placeholder or removal marker; retain current secondary-link/help selectors.
  <!-- verified: 2026-07-10 method=auth-css-source-green evidence="AuthVisual passes 17/17; ei-auth-link-row is absent outside its negative assertion, while secondary-link, help/help-line and auth-row retain current component consumers." -->

## Phase 13: StrictMode-safe GET single-flight

- [x] 13.1 Add RED focused tests that mount representative runtime/screen consumers under React StrictMode and prove the current same-key GET path issues a duplicate underlying request.
- [x] 13.2 Implement one shared in-flight registry for semantic safe-read GET only; key by client identity + method/path/canonical query + normalized relevant headers + normalized `okStatuses` + read/auth epoch + auth/session scope, and evict on both resolve and reject.
- [x] 13.3 Prove the separation/bypass matrix: different client/query/header/okStatuses/epoch/auth never merge；caller `AbortSignal`、every non-GET and semantic-write GET bypass coalescing；every semantic mutation advances read epoch before dispatch and after resolve/reject settle；`/auth/email/verify` also advances auth/session epoch on success；settle then retry creates a new underlying request.
- [x] 13.4 Narrow `AppRuntimeProvider`、Home/`useRecentTargetJobs`、Parse、`useWorkspaceTargetJobs`、Reports and Practice loader dependencies to stable client/auth/request-option/route identity inputs without suppressing real locale/auth/epoch refreshes; keep React StrictMode enabled.
- [x] 13.6 Run focused single-flight/runtime/each-loader tests, auth verify epoch tests, full frontend typecheck/build, owner contexts, `make docs-check`, `git diff --check` and broad-runtime-object dependency negative searches before restoring `completed`.

## Phase 14: settings simplification and real account actions

- [x] 14.1 RED-GATE: focused TopBar/Settings/source tests fail while the signed-in avatar/name chip, dropdown/backdrop/TopBar logout, settings tab rail, sign-in/security, font preset, product info, static privacy list, mobile/interface-language/time-zone fields or their locale/CSS/testid residues remain positive surfaces. Evidence (2026-07-15): new TopBar and Settings component tests fail on the current user chip/menu, static five-field account list, tabs/security/font/product blocks, absent export status and absent delete dialog.
- [x] 14.2 TOPBAR-GATE: replace the signed-in account chip/menu with one localized settings icon button that navigates directly to `settings`; retain signed-out login, desktop/mobile viewport containment, >=40px hit area, keyboard focus and accessible name; remove obsolete state/helpers/selectors without aliases.
  <!-- verified: 2026-07-15 method=focused-topbar-source-responsive evidence="TopBar/TopBarVisual PASS 27 tests; localized 40px gear, focus-visible and 720px containment rules locked; old menu DOM/CSS absent" -->
- [x] 14.3 ACCOUNT-GATE: render read-only `displayName/email` from authenticated `AppRuntimeContext`, with complete email and zero `emailMasked` compatibility field；route sign-out through existing `auth_logout`; tests prove Settings mount performs zero additional `getMe` calls, full email is not logged, and unauthenticated deep links preserve safe pendingAction.
  <!-- verified: 2026-07-15 method=focused-settings-runtime evidence="Settings 8/8, runtime 8/8 and TopBar 10/10 PASS; full alice@example.com is rendered from runtime with no page getMe or legacy account menu" -->
- [x] 14.4 PRIVACY-GATE: render export as typed P0 unavailable with a readable reason and no trigger；wire generated `deleteMe` behind an accessible destructive dialog (description, initial/trapped/returned focus, Escape/cancel), one confirmation-lifecycle idempotency key, pending close/submit lock, recoverable failure/retry and typed `401` auth re-probe；after `202`, reuse existing `refreshAuth()` to re-probe `/me` (expected 401), commit unauthenticated state and replace Home without a new session-clearing method；probe network/server errors remain honest auth errors.
  <!-- verified: 2026-07-15 method=settings-delete-state-machine evidence="Settings 8/8 and AppRuntimeProvider 8/8 PASS; focus trap/return, pending lock, retry same key, typed 401, 202 probe-before-replace and honest probe error covered" -->
- [x] 14.5 BDD-Gate: complete `BDD.SHELL.SETTINGS.001`, `BDD.SHELL.SETTINGS.002` and `BDD.SHELL.SETTINGS.DELETE.001`; extend `E2E.P0.101` only for real settings entry/account fields/logout and keep account deletion in domain/contract tests.
- [x] 14.6 REGRESSION-GATE: run focused component/domain tests, root `make test`, frontend typecheck/build, B2 fixture/codegen/migration negative gates, owner contexts, `sync-doc-index --check`, `make docs-check`, `git diff --check` and old-surface zero-reference scans before restoring `completed`.
- [x] 14.7 REVIEW-FIX-DEV-MOCK: add a RED default `createDevMockClient` regression for verify/profile/deleteMe/getMe, then make successful `DELETE /me` transition the fixture auth state to signed-out so the post-delete `getMe` returns 401 and the mounted Settings flow reaches Home.
  <!-- verified: 2026-07-15 method=focused-dev-mock-delete evidence="Red: client getMe resolved and mounted App remained on Settings after deleteMe. Green: devMockClient plus mounted App focused tests pass 10/10; deleteMe transitions signedIn=false, post-delete getMe returns 401 and Settings replaces Home." -->
- [x] 14.8 REVIEW-FIX-EVIDENCE: add a RED code-level regression proving a failed P0.101 email assertion and reporter stream cannot persist the raw or URL-encoded current-run email; replace raw-value matcher output and redact the stream before `trigger.log`, while retaining the real UI/API equality check and existing PASS markers.
  <!-- verified: 2026-07-15 method=p0101-failure-output-redaction evidence="Red: scenario contract test failed because no pre-log redactor existed. Green: scenario contract tests pass 7/7; synthetic raw and percent-encoded email are replaced before tee, boolean Playwright assertions retain UI/API equality without value-bearing matcher output, and a pipefail probe preserves exit code 23." -->

## Phase 15: auth route gate locale drift remediation

- [x] 15.1 RED: 中文 locale 下 auth loading/error route gate 测试必须先证明当前硬编码英文仍可见，并锁定业务 screen/API 不提前挂载.<!-- verified: 2026-07-16 method=focused-locale-red evidence="app-shell-language-switch produced 2 expected failures: Chinese loading/error gates rendered AUTH plus hard-coded English title/body; protected Resume screen/API remained unmounted" -->
- [x] 15.2 GREEN: gate eyebrow、loading/error title 与 body 全部改为 typed locale keys；中文零英文残留，英文切换保持原文案语义。<!-- verified: 2026-07-16 method=focused-locale-green evidence="app-shell-language-switch passed 3/3; Chinese loading/error gates use typed locale keys with zero hard-coded English residue, and runtime switching preserves the English authentication copy" -->
- [ ] 15.3 BDD/REGRESSION: `BDD.SHELL.AUTH.LOCALE.001`、focused locale/auth tests、locale reachability、typecheck/build、根 `make test` 与 current-run Chrome skill 真实本地页面验收通过后恢复 completed；Chrome 证据不得冒充独立 E2E ID。

## Phase 16: account theme persistence and Practice global chrome

- [x] 16.1 RED-CONTRACT: OpenAPI/fixture/generated/backend tests 先失败并锁定 `updateMe` 完整 `UserContext`、profile-only/theme-only/combined/empty-invalid matrix，以及 `user_settings` theme/custom enum-range/all-or-none constraints 和 legacy ocean 默认。
- [x] 16.2 GREEN-CONTRACT: 非空 000021 migration、generated artifacts、fixture transport、backend service/store/handler 以同一事务完成 profile/theme 更新；`getMe` 返回同一持久投影，非法请求无部分写入。
- [x] 16.3 RED-FRONTEND: TopBar/Settings/runtime/route tests 先失败并锁定“设置 / Settings”、明确 gear、TopBar 零主题 menu、Appearance 草稿预览、拖动零请求、Save 单次 `updateMe`、成功零 follow-up GET、失败/离开恢复和 route 切换零 `/me`。
- [x] 16.4 GREEN-FRONTEND: runtime bootstrap/auth recovery hydrate server-confirmed theme；Settings 保存响应直接提交 runtime/display context；Practice 从 no-chrome allowlist 移除并同时渲染 global TopBar + Practice Session Header。
- [x] 16.5 BDD-Gate: 完成 `BDD.SHELL.SETTINGS.THEME.001`；`E2E.P0.101` 原地扩展真实主题保存、logout/relogin 恢复，不新增伪场景；Practice chrome 由 cross-owner `BDD.PRACTICE.GLOBAL_CHROME.005` 承接。
- [x] 16.6 REGRESSION: OpenAPI validate/fixture/codegen drift、migration/backend focused、frontend focused/typecheck/build、根 `make test`、owner contexts、sync/docs/diff/residual gates 全部通过。
- [x] 16.7 REAL-UI: 按 scenario-env/scenario-run 合同重建真实环境，运行 `E2E.P0.101`；使用 Chrome skill 验证 Settings 与 Practice desktop/mobile DOM、请求计数、无横向溢出并保存精选截图。
- [x] 16.8 POST-PASS: 同步 bug/retrospective/index/work-journal 与真实证据，恢复受影响 owner lifecycle 并在 feature branch 原子提交。

## Phase 17: account theme L2 review remediation

- [x] 17.1 REVIEW-RED: focused tests 先失败并分别锁定 Settings 卸载/退出后的迟到 `updateMe` 不得回写旧 auth/theme、invalid server projection 回退 `ocean + null`、保存失败保留草稿并可重试、离开未保存页面恢复确认值，以及 dev mock combined-invalid 请求零部分写入。（证据：2026-07-19 frontend focused `4 failed, 41 passed`；doc focused `2 failed`。失败/重试既有行为已通过并保留为回归门禁。）
- [x] 17.2 REVIEW-GREEN: 以 component/request generation guard 丢弃不再属于当前 mounted user 的保存响应；集中校验账号主题投影；dev mock 在任何 state mutation 前拒绝出现但非法的 `displayPreferences`。（证据：frontend focused 4 files / 45 tests PASS。）
- [x] 17.3 CONTRACT-DRIFT-GATE: product-scope D-21/route、OpenAPI current schema inventory/handoff、frontend README、UI architecture 与 engineering roadmap 同步方案 B；focused doc contract test 与 scoped negative search 阻止当前章节回流 `CompleteProfileRequest` / `completeMyProfile` / TopBar theme menu / “设置与隐私”。（证据：focused owner-doc 2 tests PASS。）
- [x] 17.4 BDD/REGRESSION: `BDD.SHELL.SETTINGS.THEME.001` 的 failure/retry/leave/invalid/race domain tests、frontend focused、根 `make test`、build、docs/codegen/diff/context gates与重新部署后的真实 `E2E.P0.101` 通过；Phase 15.3 仍未完成时 plan/checklist/BDD 保持 `active`。（证据：frontend focused 45 PASS；root Python 615 / Go all / frontend 1042 PASS；build、OpenAPI 38/38、fixture、codegen、diff、migration、docs/index/context/diff PASS；`e2e-p0-101-20260719082610-75505` PASS + cleanup PASS。）

## Phase 18: screenshot-aligned Auth and Settings composition

- [x] 18.1 RED: `AuthVisual`、Auth screen 与 Settings visual/component tests 锁定宽幅双栏、原则/主操作卡、页面装饰、退出堆叠操作、设置 Header 和三张横向功能卡；旧 `1160px`/`980px` 构图先失败。<!-- verified: 2026-07-19 method=focused-vitest-red evidence="AuthVisual/ScreensVisual/ParsePlanVisual ran 35 tests: 11 expected failures, including six Auth composition failures and two Settings composition failures while 24 existing behavior/negative assertions stayed green." -->
- [x] 18.2 GREEN: 仅重组正式 DOM、page-scoped CSS、typed locale 与仓库内 SVG/CSS 装饰；保留认证、pendingAction、主题保存、退出和删除账号状态机、请求次数与错误恢复。<!-- verified: 2026-07-19 method=focused-vitest-green evidence="Auth visual/screens and Settings visual/component suites pass 54/54; frontend typecheck passes. The implementation adds only shared DOM/CSS/typed copy/SVG decoration and keeps existing handler/client paths." -->
- [x] 18.3 BDD-Gate: `BDD.SHELL.PAGES.VISUAL.002` 由 desktop/mobile component/responsive/a11y tests 与 current-run Chrome 真实页面验收承接；不得伪造验证码计时/成功或新增 E2E ID。<!-- verified: 2026-07-19 method=chrome-extension-manual evidence="Chrome at 1916x821 and 390x844 verified login, verify, logout and authenticated Settings compositions with zero horizontal overflow; real locale switching restored Chinese after English, real logout returned Home, and a protected Settings route redirected to the pending auth login." -->
- [x] 18.4 REGRESSION: focused/locale reachability、typecheck/build、根 `make test`、owner context、Header/INDEX/docs/diff gates 通过后收口当前视觉 Phase；Phase 15.3 仍保持独立未完成，plan/checklist/BDD 不恢复 `completed`。<!-- verified: 2026-07-19 method=full-regression evidence="Auth/Settings focused suites pass 54/54; shared visual/detail suites pass 95/95; frontend typecheck/build pass; root make test passes Python 615 with 4615 subtests, Go all packages, and frontend 133 files / 1066 tests. Owner context, doc/index and diff gates pass; Phase 15.3 remains open because the transient auth probe gate was not observed in current-run Chrome." -->

## Phase 19: Settings custom theme disclosure hierarchy remediation

- [x] 19.1 RED: Settings behavior/visual tests 证明 Custom 激活后一级 Ocean / Plum / Custom 仍可访问，二级 hue/chroma 必须位于同一 editor 分组的后续文档流；当前共享覆盖式 grid area 和无信息的默认 range 轨道先失败。<!-- verified: 2026-07-19 method=focused-vitest-red evidence="SettingsScreen/ScreensVisual ran 25 tests: the two new hierarchy/color-track assertions failed on the missing settings-theme-editor and existing overlapping grid/default range styling while the prior 23 tests passed." -->
- [x] 19.2 GREEN: `SettingsScreen` 与 page-scoped CSS 改为一级选择器常驻、二级编辑器按需纵向堆叠；hue 使用完整光谱轨道，chroma 使用当前 hue 的低彩到高彩渐变并保留 range 可访问语义；选择 Ocean / Plum 清除 custom accent 并隐藏二级编辑器，保存状态机和请求预算不变。<!-- verified: 2026-07-19 method=focused-vitest-green evidence="SettingsScreen/ScreensVisual pass 25/25; the editor wrapper owns persistent theme choices plus conditional custom controls, preset selection hides the panel with zero network calls, and CSS contracts pin full-spectrum hue plus hue-aware chroma tracks." -->
- [x] 19.3 BDD-Gate: `BDD.SHELL.SETTINGS.THEME.001` 覆盖一级常驻、Custom 二级展开、预定义主题回退；desktop/mobile component 与 current-run Chrome 验证无遮挡和无横向溢出。<!-- verified: 2026-07-19 method=chrome-extension-manual evidence="Real local Settings at 1440x900 and 390x844 kept all three preset/custom buttons visible, placed the custom panel strictly below options, supported Custom -> Ocean -> Custom reversal, rendered full-spectrum hue and live hue-aware chroma gradients, retained keyboard slider focus, matched document/viewport width, and logged no browser errors." -->
- [x] 19.4 REGRESSION: focused tests、typecheck/build、根 `make test`、owner context、Header/INDEX/docs/diff gates 通过并记录 current-run 证据；Phase 15.3 继续保持未完成，主 plan 保持 `active`。<!-- verified: 2026-07-19 method=full-regression evidence="Focused Settings/visual 25/25, frontend lint/typecheck/build, root make test (Python 615 / 4615 subtests, Go all packages, frontend full), local frontend redeploy, environment readiness 4/4, owner context, docs/index/link and git diff checks pass. BUG-0193 and the delivery retrospective are linked; Phase 15.3 remains open." -->

## Phase 20: Shared asynchronous transition visual system

- [x] 20.1 RED: shared transition、TopBar 与 route tests 先锁定四种 variant、统一画布、无 determinate percent、reduced-motion、mobile containment 和面试上下文导航高亮，当前重复等待态/Generating 隐藏 chrome 必须先失败。<!-- verified: 2026-07-19 method=focused-vitest-red evidence="8 files: 14 expected failures plus one missing shared-component suite; 52 existing assertions passed" -->
- [x] 20.2 GREEN: 新增单一 shell-owned `AsyncTransitionScene` 与 page-scoped CSS/code-native SVG；补齐 context route -> “面试”映射，不改变各业务 owner 的 API、轮询、错误恢复与动作合同。<!-- verified: 2026-07-19 method=focused-vitest-green evidence="12 files / 124 tests PASS; frontend typecheck PASS" -->
- [x] 20.3 BDD-Gate: `BDD.SHELL.TRANSITION.VISUAL.003` 由 component/responsive/a11y tests 与 current-run desktop Chrome 承接，不新增 E2E ID 或像素基线。<!-- verified: 2026-07-19 method=chrome-extension-manual evidence="Real local Practice/Resume/Generating/Parse transitions retained the shared TopBar and correct primary-nav context at 1920px; component/CSS contracts cover the 720px mobile boundary and reduced motion. Browser error/warning=0." -->
- [x] 20.4 REGRESSION: focused、typecheck/build、根 `make test`、owner contexts、Header/INDEX/docs/diff gate 通过；Phase 15.3 仍保持独立未完成。<!-- verified: 2026-07-19 method=full-regression evidence="Final transition-focused 9 files / 89 tests PASS; frontend production build and local redeploy PASS; root make test passes 615 tests / 4615 subtests; environment dependencies 4/4 OK. Phase 15.3 remains independently open." -->

## Phase 21: Settings Header security illustration remediation

- [x] 21.1 RED: Settings visual/component tests 拒绝山形折线、人物轮廓和独立圆形对勾，要求 Header SVG 暴露账号窗口、头像资料、柱状图、锁、盾牌对勾与星芒的分层锚点，并继续 `aria-hidden`。<!-- verified: 2026-07-20 method=focused-vitest-red evidence="ScreensVisual ran 13 tests: the new semantic layer assertion failed because the existing sparse SVG had no settings-header-art anchor, while the other 12 visual and negative assertions passed." -->
- [x] 21.2 GREEN: 仅重画 `SettingsHeaderArt` 与 page-scoped CSS，以当前主题 accent/token 形成半透明层次、柔和阴影和目标几何；保持 Header/card 构图、移动端隐藏、账号状态机和请求预算不变。<!-- verified: 2026-07-20 method=focused-vitest-green evidence="ScreensVisual and SettingsScreen pass 26/26; TypeScript typecheck passes. The new code-native SVG exposes window/profile/chart/lock/shield/two-sparkle layers and theme-token CSS without changing account handlers or request paths." -->
- [x] 21.3 BDD-Gate: `BDD.SHELL.SETTINGS.ART.004` 由 component/responsive/a11y tests 与 current-run Chrome desktop 验收承接；确认图形层级、对齐、无横向溢出和零 browser error/warning，不新增 E2E ID。<!-- verified: 2026-07-20 method=chrome-extension-manual bddChecklist=complete evidence="Real Settings at 1264x964 rendered a 360x200 decorative SVG with seven target layers, aligned the 1264px Header and appearance card, kept documentWidth=viewportWidth, adapted across Ocean/Plum/Custom previews, restored the confirmed Custom theme after reload, and surfaced no page/browser errors or warnings." -->
- [x] 21.4 REGRESSION: focused、typecheck/build、根 `make test`、owner context、Header/INDEX/docs/diff gate 通过并更新既有 BUG/复盘/日志；Phase 15.3 仍保持独立未完成。<!-- verified: 2026-07-20 method=full-regression evidence="Focused Settings 26/26, typecheck and production build pass; frontend redeploy and environment 4/4 pass; root make test passes Python 615/4615 subtests, Go all packages, frontend 134 files/1082 tests. Owner context, Header/INDEX, docs links, diff, BUG-0192, retrospective and work journal are current; Phase 15.3 remains independently open." -->

## Phase 22: TopBar language affordance and account initial

- [x] 22.1 RED: TopBar component/visual tests 拒绝 `9px` 文本 `▾` 与固定 `E`，要求 code-native SVG chevron 的可见底板/展开状态，以及从 authenticated `displayName` 派生 trim 后首个 Unicode 字符、拉丁字母大写和空名称 `?` fallback。<!-- verified: 2026-07-20 method=focused-vitest-red evidence="Four focused files ran 43 tests: 11 expected failures exposed the fixed E, missing runtime name wiring, text caret and weak CSS while 32 adjacent assertions passed." -->
- [x] 22.2 GREEN: `AppShell` 只把现有 runtime `displayName` 传给 `TopBar`；TopBar 渲染用户名首字符设置入口和可旋转 SVG chevron，保留本地化 accessible name、40px 点击区、settings 直达、语言切换、mobile containment 与零新增请求/菜单。<!-- verified: 2026-07-20 method=focused-vitest-green evidence="TopBar, visual, i18n and App language suites pass 43/43; typecheck passes. Chinese, Latin, whitespace and empty-name cases are covered, and App runtime wiring renders A for Alice Example." -->
- [x] 22.3 BDD-Gate: `BDD.SHELL.TOPBAR.IDENTITY.005` 由 TopBar/App component、responsive/a11y tests 与 current-run Chrome desktop 验收承接；验证语言菜单开合、中文用户名首字符、设置直达、无横向溢出和零 browser error/warning，不新增 E2E ID。<!-- verified: 2026-07-20 method=chrome-extension-automation evidence="Authenticated displayName 星期无 renders settings initial 星; language caret is a 14x14 SVG in a visible 20x20 panel and rotates 180 degrees when expanded; the 42x42 settings button navigates to /settings; documentWidth equals the 1920px viewport and browser warning/error logs are empty." -->
- [x] 22.4 REGRESSION: focused、typecheck/build、根 `make test`、owner context、Header/INDEX/docs/diff gate 通过并完成 bug/复盘/日志收尾；Phase 15.3 仍保持独立未完成。<!-- verified: 2026-07-20 method=root-and-document-gates evidence="Focused 43/43, typecheck, production build, frontend redeploy, environment 4/4, root make test (Python 615 with 4615 subtests, all Go packages, frontend 134 files with 1087 tests), owner context, zero-drift Header/INDEX, docs links and git diff checks pass. BUG-0192 and the existing assessment are updated; the phase journal is committed with this checklist." -->

## Phase 23: Settings theme primary-row action anchoring

- [x] 23.1 RED: Settings behavior/visual tests 要求一级主题 options 与 Save 同属固定 primary row，Custom editor/error 位于其后；旧独立第三列 action 和由展开高度驱动的垂直居中布局必须先失败。<!-- verified: 2026-07-20 method=focused-vitest-red evidence="SettingsScreen/ScreensVisual ran 26 tests: the two new primary-row assertions failed on the missing settings-theme-primary-row and the old grid-column 2 plus separate action-column CSS, while the adjacent 24 assertions passed." -->
- [x] 23.2 GREEN: 最小重组 `SettingsScreen` 与 page-scoped CSS；desktop 一级选项靠左、Save 靠右且 Custom 展开/收起不改变按钮纵向位置，mobile 安全换行并保持二级编辑器后置；主题状态机、保存禁用态和请求预算不变。<!-- verified: 2026-07-20 method=focused-vitest-green evidence="SettingsScreen/ScreensVisual pass 26/26. The new primary row owns options plus Save, the editor/error remain following siblings, and Appearance no longer has a separate third-column action rule; existing save/request tests stay green." -->
- [x] 23.3 BDD-Gate: `BDD.SHELL.SETTINGS.THEME.001` 原地覆盖固定主行；current-run Chrome 在 desktop 量测 preset/custom Save bbox 差值不超过 1px，并在 390px mobile 验证顺序、可操作性和无横向溢出，不新增 E2E ID。<!-- verified: 2026-07-20 method=chrome-extension-automation evidence="Real Settings at 1440x900 kept options and Save in the same 44px row at top=341/bottom=385 before and after Custom (delta=0px); the custom panel began below at top=399. At 390x844 the primary group and panel remained ordered and operable with documentWidth=viewportWidth=390; Custom -> Plum was reversible and browser warning/error logs were empty." -->
- [x] 23.4 REGRESSION: focused、typecheck/build、根 `make test`、frontend redeploy、环境 readiness、owner context、Header/INDEX/docs/diff gate 通过；更新 BUG-0193、既有复盘和工作日志，Phase 15.3 仍保持独立未完成。<!-- verified: 2026-07-20 method=root-and-document-gates evidence="Focused theme suites pass 34/34, typecheck and production build pass; root make test passes Python 615/4615 subtests, all Go packages and frontend 135 files/1091 tests. Frontend redeploy and dependency readiness 4/4 pass; owner context, zero-drift Header/INDEX, docs links, pruning and diff checks pass. BUG-0193, its existing assessment and the phase work journal are current; Phase 15.3 remains open." -->

## Phase 24: Shared secondary-page Back copy

- [x] 24.1 RED: locale/source/component tests 枚举正式返回控件并要求统一消费 `common.back`；目标特定 action key 和可见“返回首页/简历工坊/报告/面试规划/面试”必须先失败。<!-- verified: 2026-07-20 method=focused-vitest-red evidence="New shared Back source/catalog contract failed 16/16 against missing common.back, 14 legacy consumers and 20 target-specific action keys." -->
- [x] 24.2 GREEN: 新增中英文共享返回文案、迁移正式消费者并删除无消费者旧 key；保留每个页面的 route target、replace/push、trusted-context、fail-closed 与请求行为。<!-- verified: 2026-07-20 method=focused-vitest-green evidence="Shared contract passes 16/16; affected Auth/Parse/Resume/Practice/Reports/Generating/Report/Conversation scope passes 65 files and 495 tests, including existing navigation matrices." -->
- [x] 24.3 BDD-Gate: `BDD.SHELL.BACK.COPY.006` 通过 locale/source/component navigation tests 与 current-run Chrome 抽样验证“返回 / Back”、点击目标和无横向溢出。<!-- verified: 2026-07-20 method=chrome-extension-automation evidence="Real Generating displayed 返回 at 1512x777 and 390x844 with overflowX=0; real Report and Reports pages switched to English and displayed Back, Report Back navigated to its target-scoped Reports route, then Chinese was restored; browser warning/error logs were empty." -->
- [x] 24.4 REGRESSION: focused、typecheck/build、根 `make test`、frontend redeploy、环境 readiness、owner context、docs/index/diff gate 通过；Phase 15.3 仍保持独立未完成。<!-- verified: 2026-07-20 method=full-regression-and-document-gates evidence="Shared Back contract 16/16 and affected scope 65 files/495 tests pass; typecheck/build and root make test pass with Python 615/4615, all Go packages and frontend 136 files/1107 tests. Frontend redeploy, readiness 4/4, both owner contexts, docs/index/diff and Chrome gates pass; Shell remains active solely because Phase 15.3 is still open." -->

## Phase 25: Owner-specific Back copy exception

- [x] 25.1 RED: shared locale/source contract 要求 `common.back` 仍为默认，同时显式登记 Generating trusted Reports 专用文案；当前“所有消费者只能使用 common.back”断言必须先失败。<!-- verified: 2026-07-20 method=focused-vitest-red evidence="The shared Back-copy suite failed only for the missing generating.backToReports locale pair and its two approved Generating consumers; common.back remained present and all retired-key checks passed." -->
- [x] 25.2 GREEN: shared test 只允许 Generating owner 消费专用 key；Workspace fallback 与其它二三级页面继续使用 `common.back`，旧目标特定 keys 不得回流。<!-- verified: 2026-07-20 method=focused-vitest-green evidence="The shared Back-copy suite passes 19/19: only GeneratingScreen and GeneratingErrorState consume generating.backToReports, every return control still retains common.back for the default path, and retired keys stay absent." -->
- [x] 25.3 BDD-Gate: `BDD.SHELL.BACK.COPY.006` 与 `BDD.REPORT.GENERATING.VISUAL.003` 共同证明标签/目标一致、导航矩阵不变并通过 real Chrome 抽样。<!-- verified: 2026-07-20 method=component-and-chrome evidence="Shared source contract preserves common.back and admits only the Generating exception; component tests prove the reports/workspace destination matrix, and real Chrome captured 返回面试报告 on trusted Generating." -->
- [x] 25.4 REGRESSION: focused、typecheck/build、根 `make test`、owner context、docs/index/diff gate 通过；Phase 15.3 仍保持独立未完成。<!-- verified: 2026-07-20 method=full-regression-and-document-gates evidence="Shared and Generating copy gates, typecheck, production build and root make test 615 / 4615 pass; report/shell/resume contexts, docs links, Header/INDEX and diff gates pass. Shell remains active solely because Phase 15.3 is independently open." -->
