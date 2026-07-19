# App Shell, Auth Gate, and Settings Entrypoints Checklist

> **版本**: 1.34
> **状态**: active
> **更新日期**: 2026-07-19

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
