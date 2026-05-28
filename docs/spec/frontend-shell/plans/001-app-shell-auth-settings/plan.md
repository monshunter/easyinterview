# App Shell, Auth Gate, and Settings Entrypoints

> **版本**: 1.15
> **状态**: completed
> **更新日期**: 2026-05-28

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)
**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)
**关联 BDD Checklist**: [bdd-checklist](./bdd-checklist.md)

## 1 目标

落地正式前端 App 壳：默认 Home、五入口 TopBar、全局显示控制、认证页面、用户菜单、`requestAuth(pendingAction)`、登录后恢复动作、`parse` route shell 与 runtime / API bootstrap。修订 v1.4 补齐静态原型已具备但正式前端遗漏的 `zh` / `en` UI i18n 与 `Accept-Language` display hint；修订 v1.5 收紧 i18n 资源组织，要求每种语言使用独立 locale 文件；修订 v1.6 明确 UI 语言默认跟随浏览器 locale，未知时 fallback English，且语言切换只关联前端显示偏好、不依赖登录态；修订 v1.8 按当前 `ui-design/src/app.jsx` 将 TopBar 语言切换口径更新为 icon dropdown，旧 native select/dropdown 口径不再作为正式前端契约；修订 v1.10 明确按钮显示当前语言标签且用户显式选择持久化到 `localStorage["ei-lang"]`，并补齐已实施计划的登录态漂移修复：已登录用户区必须源级复刻头像 chip + dropdown，Vite dev fixture mock 必须覆盖默认非登录、登录成功和退出后非登录态全流程；browser-level parity 还必须覆盖 desktop / mobile dropdown geometry 与 logout flow。修订 v1.11 修复真实联调 passwordless 链路：登录和注册提交 `startAuthEmailChallenge` 时必须兼容后端 `202 Accepted` 空响应并导航到 verify 页；Mailpit magic link 必须落到前端 `auth_verify`，由前端自动消费 token、刷新 session，并用 replace 导航清理 URL token。修订 v1.12 将真实联调入口从 Mailpit magic link 改为 Mailpit 6 位 email code，并锁定邮箱是唯一账号标识：注册页传 `purpose=signup` + displayName，后续登录同一邮箱传 `purpose=login`，displayName 不唯一且不参与账号去重；TopBar 不再使用 `刘哲` / `Liu Zhe` / `liuzhe@example.com` 样例 fallback。修订 v1.13 将注册和登录合并为单一邮箱验证码入口：新邮箱 verify 后进入 `auth_profile_setup` 完成 displayName + 条款确认，`/me.profileCompletionRequired` 是强制跳转依据；旧 `auth_register` 不再是 live route 或可见入口。修订 v1.14 修复未登录 Home 展示 Recent mock interviews 与 raw backend unauthorized error 的回归，并把面试相关业务 route 统一前置到 `auth_login(pendingAction)`。修订 v1.15 收紧 L2 验收缺口：`AuthLoginScreen` 发码请求体只提交 email、`auth_profile_setup` 只在 `/me.profileCompletionRequired=false` 后恢复 pendingAction，且 P0.102 wrapper 直接校验 runner / Go `--- PASS` 证据。完成后，后续 D2-D6 前端 workstream 可以在同一壳内继续实现业务页面。

## 2 背景

当前静态原型已经在 `ui-design/src/app.jsx` 和 `docs/ui-design/` 中锁定了目标 route、TopBar、认证页面、pending action 模型和中英语言切换。`engineering-roadmap` S1 要求先创建 `frontend-shell`，再推进 D2-D6 前端 workstream。本 plan 是第一个正式前端代码 plan。前端新增 shell / auth / settings 组件时只以 `docs/ui-design/`、`ui-design/` 和本 spec 为准；外部品牌设计系统不再作为实现参考。

## 3 质量门禁分类

- **Plan 类型**: `feature-behavior` + `frontend`。
- **TDD 策略**: 通过 `/implement frontend-shell/001-app-shell-auth-settings frontend` -> `/tdd` 执行；每个 checklist item 先写 focused Vitest / component test / route-state test，再实现最小前端代码；测试断言写在 checklist 的 `验证:` 后。Runtime / API bootstrap 测试必须覆盖 `getRuntimeConfig`、`getMe` authenticated / unauthenticated、auth generated operations、mock scenario fail-loud，以及 dev mock session 状态从默认 unauthenticated -> verify authenticated -> logout unauthenticated 的连续变化。当前 plan 一旦把 frontend package `build` script 从占位切换为真实 bundler gate，必须在同一验证面运行 `pnpm --filter @easyinterview/frontend build` 与根 `make build`。
- **BDD 策略**: 需要 BDD。本 plan 引入用户可见 App shell、TopBar、认证页面、pending action 行为和受保护业务 route guard，必须维护 `bdd-plan.md`、`bdd-checklist.md`，并在主 checklist 中使用 `BDD-Gate:` 引用 `E2E.P0.001`、`E2E.P0.002`、`E2E.P0.032`、`E2E.P0.101`、`E2E.P0.102`。
- **替代验证 gate**: 不适用；BDD gate 是本 plan 的用户行为验证入口。补充 gate 包括 frontend unit tests、typecheck、mock-contract-suite handoff、route negative search、`make docs-check`。

## 4 实施步骤

### Phase 1: App bootstrap and route normalization

#### 1.1 建立正式前端 App shell

创建或整理正式前端入口，使 App 默认进入 `home`，保留 `practice` / `generating` 的无 TopBar chrome 行为，并支持 route params。

#### 1.2 实现 route normalization 与旧 route 拦截

把旧 route key 映射到当前保留 route 或 Home；不得创建旧独立页面。

#### 1.3 接入 runtime config 与 typed API bootstrap

使用 generated client、fixture-backed mock transport 和 `getRuntimeConfig` 建立 App 启动边界；`/me` 只驱动用户区登录态，不得阻塞默认 Home。

#### 1.4 L2 remediation: 删除 `voice` route alias

正式前端不得保留 `voice` route alias。旧 `voice` 输入必须作为未知 route fallback `home`，语音面试只能由 `practice` route 显式携带 `mode=voice` / `modality=voice` 表达。

### Phase 2: TopBar and display controls

#### 2.1 实现五入口 TopBar

TopBar 只展示 `home`、`jd_match`、`workspace`、`resume_versions`、`debrief` 五个一级入口。

#### 2.2 实现全局显示控制

主题色、暗色和语言切换由 TopBar 持有；显示偏好在登录前后保持稳定。

#### 2.4 I18n remediation: 建立 D1 shell message catalog

为 TopBar、auth shell、profile/settings shell 和 placeholder route shell 建立 typed `zh` / `en` message catalog 或等价 helper。切换语言必须立即重绘 D1 可见静态文案；RouteName、testid、URL/hash 与业务字段仍使用稳定英文 key，不受 UI locale 影响。

#### 2.5 I18n remediation: browser locale bootstrap 与 request header

把浏览器 locale 归一为 `zh` / `en` 后作为初始默认；不支持、未知或缺失时 fallback `en`。用户显式切换优先级最高，登录态刷新、`/me.uiLanguage` 与 runtime `defaultUiLanguage` 不得覆盖。App runtime 通过 generated client request options 或默认 header 把当前 UI locale 作为 `Accept-Language` display hint 传给 `getRuntimeConfig`、`getMe` 和 D1 auth operations。

#### 2.6 I18n remediation: BDD language switch gate

新增 BDD 场景验证默认 App shell 从中文切到 English 后，TopBar 导航、登录/注册、用户菜单和 D1 auth/settings/profile shell 静态文案同步切换，并保留旧 route / prototype data 负向约束。

#### 2.7 I18n remediation: 独立 locale 文件与语言切换契约

把 `zh` / `en` message map 拆到独立 locale 文件，`messages.ts` 仅保留导入、类型约束、locale 归一化和 helper；新增结构测试阻止多语言字面量回流到同一文件。TopBar 语言切换必须按 `ui-design/src/app.jsx` 复刻为可访问 icon dropdown，并由 component / scenario test 直接断言。

#### 2.8 I18n remediation: 前端偏好独立于登录态

删除 App shell 中从 runtime config 或 `/me` 回写 UI 语言的 bootstrap 逻辑；DisplayPreferencesProvider 负责根据浏览器语言初始化，后续只响应 TopBar 前端设置。Focused regression test 必须覆盖 `/me.uiLanguage` 与 runtime `defaultUiLanguage` 跟浏览器语言不一致时，已登录刷新不会改写当前 UI 语言或造成 locale 循环请求。

### Phase 3: Auth pages and pending action

#### 3.1 实现认证页面壳

历史 D1 初始实现曾包含 `auth_login`、`auth_register`、`auth_verify`、`auth_reset`、`auth_logout` 页面流。当前 Phase 9 已将 `auth_register` 收敛为 legacy alias，不得 materialize 独立页面；正式 wire 只允许 generated `startAuthEmailChallenge`、`verifyAuthEmailChallenge`、`getMe`、`completeMyProfile`、`logout` 与 first-party session cookie。密码登录、OAuth 和 reset 只能作为 UI shell 或 stub，不得私造 API。

#### 3.2 实现 `requestAuth(pendingAction)`

未登录用户触发需要身份的动作时进入登录页；登录成功后恢复 route 和 params。

#### 3.3 固化 Auth API contract gate

为 auth shell 增加负向断言：正式前端不得新增 password / OAuth / Bearer token / 自定义 session API；真实网络边界只通过 B2 generated auth operations 和 first-party session cookie。

#### 3.5 L2 remediation: verify token 与 auth-only params 隔离

`auth_verify` 必须把用户输入的验证 token 传给 generated `verifyAuthEmailChallenge` query；登录 / 注册页临时字段只允许停留在认证页，不得随 pendingAction 恢复到业务 route params。

### Phase 4: User menu, profile, settings

#### 4.1 实现用户菜单入口

未登录展示登录 / 注册；已登录展示头像菜单，并分流到 `profile`、`settings`、`auth_logout`。

#### 4.2 实现 settings/profile placeholder shell

为 `profile` 和 `settings` 提供当前 UI 边界内的页面壳；业务内容可用 mock data，但不得恢复旧 Growth / Experiences / Mistakes 模块。

### Phase 5: BDD and handoff

#### 5.1 执行 App shell BDD gate

按 `bdd-plan.md` 和 `bdd-checklist.md` 验证默认 Home + TopBar 场景。

#### 5.2 执行 auth pendingAction BDD gate

按 `bdd-plan.md` 和 `bdd-checklist.md` 验证登录打断后恢复业务动作。

#### 5.3 Handoff 给后续 D2-D6

记录稳定 route、shell API、mock runtime 入口和后续页面 owner 的接入点。

#### 5.4 UI 真理源 handoff

记录 D1 后续组件的 UI 真理源边界：正式前端视觉只从 `ui-design/` 与 `docs/ui-design/` 原生迁移，不允许 AI 自由重设计，不引入外部品牌设计系统作为替代参考。

#### 5.5 Review hardening: 真实 build smoke gate

把 [Frontend Shell Review Remediation Hardening 交付复盘](../../../../reports/2026-05-07-frontend-shell-review-remediation-hardening-assessment.md) 的最高优先级建议固化为 owner gate：当 D1 / 后续 frontend owner 将 package `build` 从占位切换为真实 Vite bundler 时，必须同时具备 HTML / runtime entry，并通过 `pnpm --filter @easyinterview/frontend build` 与根 `make build`。

### Phase 6: Auth state and user menu parity remediation

#### 6.1 源级复刻已登录用户菜单

按 `ui-design/src/app.jsx::TopBar` 把已登录用户区从 inline 三按钮修正为头像 chip + dropdown menu。按钮必须显示头像 initials、display name 和 caret；菜单打开后必须显示用户姓名 / masked email header、`用户画像`、`设置与隐私` 和 `退出登录` 三项，带图标、分隔线、关闭 backdrop、Escape 关闭，并在点击 profile/settings/logout 后关闭菜单。

#### 6.2 修复 fixture-backed dev mock session 状态

Vite dev 默认 `createDevMockClient()` 必须从非登录态开始；`verifyAuthEmailChallenge` 成功后同一 mock client 的后续 `/me` 返回 authenticated；`logout` 成功后后续 `/me` 返回 unauthenticated。该状态只存在于 dev mock client 实例中，不影响通用 `createFixtureBackedFetch` 的显式 scenario 选择语义。

#### 6.3 BDD-Gate: 验证 E2E.P0.032 通过

新增并执行 `E2E.P0.032`，覆盖 dev mock 默认非登录态、mock 登录后头像 dropdown 菜单、profile/settings 分流、logout 后回到非登录态，以及旧 inline 用户菜单 / 静态 authenticated default 回流负向断言。

#### 6.4 L2 remediation: 浏览器级 authenticated user menu parity gate

`frontend/tests/pixel-parity/topbar.spec.ts` 必须在 desktop / mobile 两个 chromium project 下通过 mocked Auth API 完成 login → avatar chip → dropdown → logout flow。断言范围包括：头像 chip text / initials / caret、dropdown header、masked email、`用户画像` / `设置与隐私` / `退出登录` 三项、`ui-design/src/app.jsx` 中的 `minWidth: 220` / `top: calc(100% + 6px)` / padding / shadow 等源码字面量、desktop 菜单右边与 chip 右边对齐、mobile 菜单保持在 viewport 内、logout confirm 后回到登录 / 注册入口。

#### 6.5 Phase 6 operation matrix

| operationId | fixture | frontend consumer | backend handler | persistence | AI dependency | scenario coverage |
|-------------|---------|-------------------|-----------------|-------------|---------------|-------------------|
| `getRuntimeConfig` | `openapi/fixtures/Auth/getRuntimeConfig.json#default` | `AppRuntimeProvider`、`topbar.spec.ts` mocked bootstrap | backend-auth runtime config handler | 无 | 无 | focused runtime tests、E2E.P0.032、E2E.P0.006 topbar |
| `getMe` | `openapi/fixtures/Auth/getMe.json#authenticated|unauthenticated` | `AppRuntimeProvider`、`TopBar` user area、`createDevMockClient` | backend-auth current-user handler | backend session cookie lookup；frontend 不持久化 session | 无 | devMockClient tests、E2E.P0.032、E2E.P0.006 topbar |
| `startAuthEmailChallenge` | `openapi/fixtures/Auth/startAuthEmailChallenge.json#default` | Auth login screen；request body 只包含 email；`returnTo` 仅属于历史 auth route param，不得提交给发码 API；不传 `purpose` 或 `displayName` | backend-auth challenge issue handler | backend auth challenge/session storage；frontend 无持久化 | 无 | AppAuthDispatch tests、E2E.P0.032、E2E.P0.006 topbar、E2E.P0.101 |
| `verifyAuthEmailChallenge` | `openapi/fixtures/Auth/verifyAuthEmailChallenge.json#default` | Auth verify screen、`createDevMockClient` state transition；成功后读取 `/me.profileCompletionRequired` 决定是否进入资料补全 | backend-auth verify handler | backend mints first-party session cookie；frontend mock client 仅有实例内状态 | 无 | devMockClient tests、E2E.P0.032、E2E.P0.006 topbar、E2E.P0.101 |
| `completeMyProfile` | `openapi/fixtures/Auth/completeMyProfile.json#default` | Auth profile setup screen；提交 trimmed displayName + `acceptedTerms=true` | backend-auth `PATCH /me` handler | backend `users.display_name` / `profile_completed_at` / `terms_accepted_at` | 无 | AuthProfileSetup tests、E2E.P0.101 |
| `logout` | `openapi/fixtures/Auth/logout.json#default` | Auth logout screen、TopBar logout route、`createDevMockClient` reset | backend-auth logout handler | backend clears session cookie/session；frontend mock client 仅重置实例内状态 | 无 | devMockClient tests、E2E.P0.032、E2E.P0.006 topbar |

### Phase 7: Historical real passwordless mail-link remediation

#### 7.1 `startAuthEmailChallenge` empty-body success

真实后端 `POST /api/v1/auth/email/start` 返回 `202 Accepted` 且无 JSON body。Generated client 必须把无 body 的 2xx 响应视为成功，登录页和注册页必须在该成功后进入 `auth_verify`，不得因为 `response.json()` 解析空响应而中断。

#### 7.2 `auth_verify` magic-link callback

该历史 gate 曾允许 `auth_verify` 接收邮件链接携带的一次性 `token` query，并在进入页面后自动 `verifyAuthEmailChallenge`。v1.12 Phase 8 已将当前验收口径改为 code-only 邮件和手动 6 位验证码输入；不要把 magic-link callback 作为新的完成证据。

#### 7.3 Local dev Mailpit handoff

该历史 gate 曾要求本地 Mailpit 默认邮件链接指向前端 `/auth/verify`，而不是后端 API verify URL。v1.12 Phase 8 已将当前验收口径改为 Mailpit code-only 邮件；backend dev CORS allowlist 仍可从 `EMAIL_VERIFY_BASE_URL` 派生 frontend origin，frontend real mode 仍必须显式设置 `VITE_EI_API_BASE_URL`，避免 CORS 端口和 API base port 分裂。

### Phase 8: Historical email-code auth and display-name remediation

本阶段是 2026-05-27 的历史完成记录，已被 Phase 9 的单入口邮箱登录与首次资料补全取代。以下 register/login purpose、AuthRegisterScreen 与 duplicate-register 证据不得再作为当前验收口径。

#### 8.1 Historical register displayName pass-through

历史注册页提交 `startAuthEmailChallenge` 时曾传 `purpose=signup` 与 trimmed `displayName`；登录页曾传 `purpose=login` 且不传 displayName。当前 `AuthLoginScreen` 只提交 normalized email；pendingAction / legacy `returnTo` 只作为 auth route params 在登录、验证和资料补全页之间传递，不进入 `startAuthEmailChallenge` request body；displayName 只能在 `auth_profile_setup` 通过 `completeMyProfile` 提交。

#### 8.2 Six-digit code verify UI

`auth_verify` 输入框必须按 `docs/ui-design/auth-and-entry.md` 与 `ui-design/src/screen-auth.jsx` 更新为 6 位数字验证码：numeric input mode、最多 6 位、过滤非数字、文案不再出现 link/token 口径。generated client 调用仍使用 B2 `token` query 名。

#### 8.3 TopBar user fallback cleanup

已登录用户区只能显示 `/me.displayName` / `/me.emailMasked` 或中性 fallback（`候选人` / `Candidate`、邮箱不可用文案），不得把 prototype 样例 `刘哲` / `Liu Zhe` / `liuzhe@example.com` 当运行时 fallback。

#### 8.4 Local Mailpit email-code handoff

历史 `E2E.P0.101` 与 Playwright auth real-mode 配置从 mail-link 改为 email-code：先注册唯一邮箱并显示 displayName，退出后使用同一邮箱登录；从 Mailpit 邮件正文提取 6 位 code，前端手动填入 `auth_verify`，并断言邮件正文不包含 `/auth/verify?token=`。当前 P0.101 场景已更新为 single-entry login + profile setup，见 Phase 9.6。

### Phase 9: Unified email login and first-login profile setup

#### 9.1 UI truth source and route catalog

`docs/ui-design/auth-and-entry.md`、`ui-design/src/screen-auth.jsx` 与 `ui-design/src/app.jsx` 必须先改为单一登录入口：TopBar 未登录用户区只显示 `登录 -> auth_login`；认证页集合为 `auth_login`、`auth_verify`、`auth_profile_setup`、`auth_reset`、`auth_logout`；旧 `auth_register` 只能 normalize 到当前保留 route，不得 materialize 成独立页面、按钮或 URL 入口。

#### 9.2 Unified login start

`AuthLoginScreen` 调用 `startAuthEmailChallenge` 时只提交 normalized email 与 safe pendingAction params，不提交 `purpose`、`displayName` 或注册/登录选择。发码成功后进入 `auth_verify`；错误提示不得泄露邮箱是否已注册。

#### 9.3 Profile completion routing

`AuthVerifyScreen` 验证 code 后必须刷新 `/me` 或读取等价 generated response；只要 `profileCompletionRequired=true`，无论首次验证、刷新、关闭浏览器后重开、换浏览器重新登录、退出后重新登录，或直开业务 URL，都必须先进入 `auth_profile_setup`。资料补全前不得恢复 pendingAction 或业务 route。

#### 9.4 Profile completion submit

`auth_profile_setup` 提交 trimmed displayName 与 `acceptedTerms=true`，调用 generated `completeMyProfile` / `PATCH /me`。成功后必须刷新 auth context，确认 `profileCompletionRequired=false` 后才恢复 pendingAction 或回 Home；displayName 不进入账号唯一性判断，也不得写入 pendingAction、URL 或业务 route params。

#### 9.5 Phase 9 operation matrix

| operationId | fixture | frontend consumer | backend handler | persistence | AI dependency | scenario coverage |
|-------------|---------|-------------------|-----------------|-------------|---------------|-------------------|
| `startAuthEmailChallenge` | `openapi/fixtures/Auth/startAuthEmailChallenge.json#default` | `AuthLoginScreen` 单入口发码；不再传 `purpose` / `displayName` | backend-auth challenge issue handler | backend auth challenge storage；frontend 不持久化 code | 无 | App auth focused tests、E2E.P0.101 |
| `verifyAuthEmailChallenge` | `openapi/fixtures/Auth/verifyAuthEmailChallenge.json#default` | `AuthVerifyScreen` 6 位 code 验证 | backend-auth verify handler | backend mints first-party session cookie | 无 | App auth focused tests、E2E.P0.101 |
| `getMe` | `openapi/fixtures/Auth/getMe.json#authenticated|unauthenticated|profileIncomplete` | `AppRuntimeProvider`、profile completion route guard、TopBar | backend-auth current-user handler | backend session lookup；`profileCompletionRequired` 后端驱动 | 无 | route guard tests、E2E.P0.101 |
| `completeMyProfile` | `openapi/fixtures/Auth/completeMyProfile.json#default` | `AuthProfileSetupScreen` | backend-auth `PATCH /me` handler | backend `users.display_name` / `profile_completed_at` / `terms_accepted_at` | 无 | focused component/API tests、E2E.P0.101 |
| `logout` | `openapi/fixtures/Auth/logout.json#default` | Auth logout screen、TopBar logout route | backend-auth logout handler | backend clears session cookie/session | 无 | E2E.P0.101 logout/relogin branch |

#### 9.6 BDD-Gate: 验证 E2E.P0.101 通过

更新 real frontend/backend/Mailpit 场景：新邮箱从单一登录入口完成 6 位验证码验证后进入 `auth_profile_setup`；刷新该页仍停留在资料补全；关闭浏览器或换浏览器后用同一邮箱重新登录仍先进入资料补全；完成 displayName + 条款确认后 `/me.profileCompletionRequired=false`，TopBar 显示该 displayName；退出后同一邮箱再次登录不再进入资料补全并恢复 pendingAction 或 Home。场景必须负向断言 TopBar 注册按钮、`auth_register` live page、`purpose=signup/login` 请求体、displayName-before-verify 和旧 magic-link URL 不出现。

### Phase 10: Unauthenticated interview route guard remediation

#### 10.1 UI truth source for signed-out Home

更新 `docs/ui-design/auth-and-entry.md`、`ui-design/src/screen-home.jsx` 与 `ui-design/src/app.jsx`：Home 仍可未登录访问和输入 JD 草稿，但 Recent mock interviews 只在 `signedIn=true` 时渲染；未登录状态不得展示该模块、空态、skeleton 或 raw backend unauthorized error。

#### 10.2 Frontend runtime auth guard

正式前端 runtime 中，`home` 和 auth pages 仍可未登录访问；`jd_match`、`parse`、`workspace`、`resume_versions`、`practice`、`generating`、`report`、`debrief`、`profile`、`settings` 等业务 route 必须在 `runtime.auth.status === "authenticated"` 后才挂载对应 screen。`loading` 状态只渲染 auth gate loading 占位，`unauthenticated` 状态导航到 `auth_login` 并携带 safe pendingAction；登录成功仍先经过 `profileCompletionRequired` gate 再恢复原 route。

#### 10.3 Home recent data fetch guard

`useRecentTargetJobs` / Home recent card 数据源必须只在 authenticated 状态调用 `listTargetJobs`。未登录、auth loading 或 auth error 下不发起受保护 API，不展示 Recent mock interviews 模块，也不把后端 `AUTH_UNAUTHORIZED` 错误正文渲染到 Home。

#### 10.4 Backend protected API proof

复用 `backend-auth` 的 C1 session policy 和 `cmd/api` wiring gate，证明 OpenAPI document-level security 与 runtime handler 均只把 `startAuthEmailChallenge`、`verifyAuthEmailChallenge`、`getRuntimeConfig` 作为 public，`logout` 作为 optional，其他面试相关 API 由 session middleware 保护。

#### 10.5 Phase 10 operation matrix

| operationId | fixture | frontend consumer | backend handler | persistence | AI dependency | scenario coverage |
|-------------|---------|-------------------|-----------------|-------------|---------------|-------------------|
| `getMe` | `openapi/fixtures/Auth/getMe.json#unauthenticated|authenticated` | `AppRuntimeProvider`、App protected route guard、Home recent visibility | backend-auth current-user handler | backend session lookup；frontend 不持久化 session | 无 | App auth route guard tests、E2E.P0.102 |
| `listTargetJobs` | `openapi/fixtures/TargetJobs/listTargetJobs.json#default|empty|one-job|twelve-plus` | `useRecentTargetJobs`、Home Recent mock interviews | `cmd/api` targetjob list handler behind `auth.SessionMiddleware` | `target_jobs` filtered by session user | target job parse may use AI outside this read path | Home recent auth guard tests、backend `TestBuildAPIHandlerMountsTargetJobRoutesBehindSessionMiddleware`、E2E.P0.102 |
| `importTargetJob` | `openapi/fixtures/TargetJobs/importTargetJob.json#manual-text-primary` | Home paste/upload/url submit via `requestAuth` + pending import source | `cmd/api` targetjob import handler behind session middleware | `target_jobs` / async parse job | target job parse AI after authenticated import | Home auth gate tests、E2E.P0.102 |
| `N/A protected frontend routes` | N/A | App route guard for `jd_match` / `parse` / `workspace` / `resume_versions` / `practice` / `generating` / `report` / `debrief` / `profile` / `settings` | OpenAPI document security + per-route session policy for their API calls | route-specific domain stores | route-specific | App route guard tests、E2E.P0.102 |

#### 10.6 BDD-Gate: 验证 E2E.P0.102 通过

新增并执行 `E2E.P0.102`：未登录 Home 不显示 Recent mock interviews、不会调用 `listTargetJobs`、不会显示 raw `AUTH_UNAUTHORIZED`；未登录直开 `workspace` / `practice` / `report` / `jd_match` 等业务 route 时先进入 `auth_login(pendingAction)`，业务 screen 不挂载且受保护 API 不被调用；后端 focused gate 证明面试相关 API 仍由 session middleware 返回 B1 `AUTH_UNAUTHORIZED` envelope。

## 5 验收标准

- 默认打开 App 渲染 Home、五入口 TopBar、单一登录入口和显示控制，不出现 welcome 或注册入口。
- `requestAuth(pendingAction)` 能在登录成功后恢复 `practice` 或 `report` 上下文。
- 用户菜单的 `用户画像` 与 `设置与隐私` 分别进入 `profile` 和 `settings`。
- `parse` route 作为 shell route 可达，但 JD 解析业务细节留给后续 owner。
- Runtime config、`/me` 和 auth generated operations 均通过 fixture-backed client 测试，不直接读取 prototype data。
- Vite dev 默认 mock App 首屏展示非登录态；passwordless mock verify 后根据 `/me.profileCompletionRequired` 进入资料补全或展示源级复刻的头像 dropdown 用户菜单；logout 后 `/me` 回到 unauthenticated，TopBar 重新展示单一登录入口。
- 真实后端 `202 Accepted` 空响应不会让 generated client 抛出 JSON parse 错误；单一登录入口提交邮箱后进入 verify 页并显示邮件已发送/等待验证状态。
- Mailpit 邮件只展示 6 位验证码；用户在前端 `auth_verify` 手动输入 code 后调用 generated `verifyAuthEmailChallenge`，刷新 session，并恢复 pending route。
- 首次使用的新邮箱在 verify 后必须进入 `auth_profile_setup`；资料补全前刷新、换浏览器重新登录、退出后重新登录或直开业务 URL 都不能恢复业务动作；`completeMyProfile` 成功后才恢复 pendingAction 或 Home。
- Authenticated user menu 的 browser-level parity 覆盖 desktop / mobile 两个 viewport；mobile 下 dropdown 不得从 viewport 左右溢出。
- TopBar 语言切换通过 `ui-design/src/app.jsx` 一致的 icon dropdown 驱动 `zh` / `en` 静态文案；按钮显示当前语言标签，locale 优先级为用户显式选择（`localStorage["ei-lang"]`）> 浏览器 locale > English fallback；runtime `defaultUiLanguage` / `/me.uiLanguage` 不参与 UI 语言决策；D1 generated client 请求带当前 locale 的 `Accept-Language` display hint。
- `zh` / `en` message map 分别位于独立 locale 文件，i18n helper 只聚合导入并提供类型安全 API，不在单文件内糅合多语言文案。
- 旧 route negative search 确认正式前端不保留独立 old route screen。
- UI 真理源边界写入 handoff：正式前端视觉只以 `ui-design/` 与 `docs/ui-design/` 为准。
- BDD-Gate `E2E.P0.001`、`E2E.P0.002`、`E2E.P0.032`、`E2E.P0.101`、`E2E.P0.102` 通过。
- Frontend package 真实 build gate 与根 build 聚合 gate 通过，避免 `frontend/package.json` 脚本升级后缺 entry 破坏 `make build`。

## 6 风险与应对

| 风险 | 应对措施 |
|------|----------|
| 前端 shell 把业务页面一次做大 | 本 plan 只做 route / shell / auth / menu；D2-D6 单独派生 |
| 登录成功丢失业务上下文 | Phase 3.2 unit test 和 E2E.P0.002 BDD gate 强制覆盖 |
| 旧 route 被当兼容入口保留 | Phase 1.2 和 negative search 要求旧 route 只能 normalize，不建独立 screen |
| mock 数据源漂移 | 依赖 `mock-contract-suite`，禁止 import prototype data |
| Auth UI 超出 C1/B2 契约 | Phase 3.1 / 3.3 只允许 generated passwordless session operations；密码 / OAuth / reset 不 wire 真实 API |
| 外部品牌参考或 AI 自由发挥被误当正式视觉依据 | Phase 5.4 明确 `ui-design/` 与 `docs/ui-design/` 是唯一 UI truth source |
| 语言切换退化为状态占位 | Phase 2.4 / 2.6 增加文案切换测试与 BDD gate，禁止只断言控件状态；控件结构必须继续对齐 `ui-design/src/app.jsx` icon dropdown |
| i18n 资源糅合导致后续语言扩展困难 | Phase 2.7 增加 locale 文件结构测试，要求每个语言独立文件，聚合层只做 helper |
| operation-level fixture 被误当用户状态流 | Phase 6.2 使用 stateful dev mock client 测试默认非登录、verify 后 authenticated、logout 后 unauthenticated，防止静态 `getMe.default` 掩盖真实流程缺口 |
| 用户菜单结构测试只断言文本存在 | Phase 6.1 / 6.3 反查 `ui-design/src/app.jsx` 的头像 chip、dropdown header、分隔线和关闭路径，禁止 inline 三按钮再次回流 |
| 浏览器 viewport 下菜单几何与 jsdom 断言脱节 | Phase 6.4 通过 Playwright desktop / mobile 直接断言 dropdown 与 chip 的几何关系和 viewport containment |
| Phase 6 前后端契约只停留在 fixture 名称 | Phase 6.5 固化 operation matrix，把 operationId、fixture、frontend consumer、backend handler、persistence、AI dependency 和 scenario coverage 放进同一 owner plan |
| 真实后端 2xx 空响应与 fixture body 不一致 | Phase 7.1 使用空 body Response 回归测试 generated client，禁止只用 fixture `{}` 掩盖真实联调错误 |
| Mailpit 邮件回流旧链接口径 | Phase 8.4 将本地邮件改为 code-only；邮件正文、场景和 evidence 不得再包含 `/auth/verify?token=` |
| 本地端口在代码中分散硬编码 | Phase 7.3 将 backend dev CORS origin 从 `EMAIL_VERIFY_BASE_URL` 派生，并要求 frontend real mode 显式配置 `VITE_EI_API_BASE_URL`；Vite dev/preview 端口通过 `FRONTEND_HOST_PORT` / `FRONTEND_PREVIEW_PORT` 覆盖 |
