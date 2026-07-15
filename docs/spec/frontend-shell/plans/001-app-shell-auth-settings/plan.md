# App Shell, Auth Gate, and Settings Entrypoints

> **版本**: 1.29
> **状态**: active
> **更新日期**: 2026-07-15

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)
**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)
**关联 BDD Checklist**: [bdd-checklist](./bdd-checklist.md)

## 1 目标

本计划交付当前正式前端 App shell：默认 Home、三入口 TopBar、全局显示控制、email-code 认证页、资料补全 gate、`requestAuth(pendingAction)` 恢复、已登录设置齿轮、无 tab 的真实账号/隐私设置页、runtime / generated client bootstrap，以及面试业务 route 的登录前置保护。

当前完成态文档只描述现行合同。任何新增可见页面、route、auth flow 或设置页能力，必须先更新 `frontend/` 静态原型、`docs/ui-design/` 和 `frontend-shell` spec，再修订本 owner 或派生明确边界的新 plan。

Phase 1-13 的已勾选内容只保留历史交付证据；Phase 14 是当前设置合同 owner，并取代其中关于账号 chip/dropdown、Settings tab 与 font preset 的旧正向描述。实施与验收不得把历史文字当作现行 UI 要求。

## 2 当前合同

### 2.1 UI 与 route catalog

- UI 设计文档：`docs/ui-design/`、`frontend/src`。
- 一级 TopBar 入口：`home`、`workspace`、`resume_versions`。
- 上下文 route：`parse`、`practice`、`generating`、`report`。
- 账号入口 route：已登录 TopBar 设置齿轮直达 `settings`；`auth_logout` 从设置页进入。
- 认证 route：`auth_login`、`auth_verify`、`auth_profile_setup`、`auth_logout`。
- Unsupported route / malformed URL 必须经同一 normalization 层折回当前 route catalog 或 `home`，不得 materialize 独立页面。

### 2.2 Auth / settings / display

- Auth UI 只通过 email-code flow 触发 `startAuthEmailChallenge`、`verifyAuthEmailChallenge`、`getMe`、`completeMyProfile`、`logout` 和 first-party session cookie。
- `profileCompletionRequired=true` 时，登录后必须先进入 `auth_profile_setup`；资料补全成功并刷新 `/me` 后，才恢复 pendingAction 或回 Home。
- `pendingAction` 只保存 route name、canonical URL 和 safe params，不保存 JD 原文、简历原文、验证码、AI prompt/response 或解析正文。
- Settings 为无 tab 单页：Account 复用 runtime user 展示只读 `displayName/emailMasked` 并提供退出入口；Privacy 展示导出暂不可用状态和账号删除确认流程。
- 显示偏好由前端持有：主题、暗色和语言下拉在登录前后保持稳定；默认主题与无效值 fallback 为 `ocean`。字体固定，不保留 preset 状态。

### 2.3 StrictMode-safe GET orchestration

- React StrictMode 保持开启；不得通过关闭 StrictMode、延迟 mount 或 screen-local boolean 掩盖重复 GET。
- generated/runtime client 的共享 in-flight registry 只合并同时在途的语义只读 GET。逻辑 key 至少包含 client identity、method、path、canonical query、规范化相关 headers（含 `Accept-Language`）、normalized `okStatuses`、read/auth epoch 与 auth/session scope。
- resolve/reject 都必须立即驱逐 registry entry；这是 single-flight，不是跨 settle cache。不同 client/query/header/okStatuses/epoch/auth、带 caller `AbortSignal` 的 GET、所有非 GET 与语义写入 GET 必须绕过合并。每个语义写请求在 dispatch 前与 resolve/reject settle 后都推进 read epoch，切断 mutation 前、期间与之后的错误复用。`/auth/email/verify` 虽是 GET wire method，但会消费 challenge/更新 session，必须按语义写请求 bypass；成功后另推进 auth/session epoch。
- `AppRuntimeProvider` 以及 Home / `useRecentTargetJobs`、Parse、`useWorkspaceTargetJobs`、Reports、Practice 等 screen loader 的 bootstrap/refresh effect 依赖稳定 client、auth、request-option 与 route identity 输入，不把每次 render 都变化的整体 runtime object 作为依赖；auth/locale/epoch 真实变化仍触发独立请求。

## 3 质量门禁分类

- **Plan 类型**: `feature-behavior` + `frontend` + `contract`。
- **TDD 策略**: 本计划按 `/implement frontend-shell/001-app-shell-auth-settings frontend` -> `/tdd` 完成。focused Vitest / component / route-state test 只用于开发反馈；阶段完成由仓库根 `make test` 承接前后端全量单测。
- **BDD 策略**: `BDD.SHELL.AUTH.001` 继续验证 auth guard、profile setup 与 pendingAction；`BDD.SHELL.SETTINGS.001/.002` 和 `BDD.SHELL.SETTINGS.DELETE.001` 覆盖单一设置入口、真实账号数据、登录保护与删除状态机。`E2E.P0.101` 原地扩展为真实 email-code/profile setup 后进入设置、核对真实账号字段并退出；删除账号只由 domain/contract tests 验证，不在共享登录场景中执行破坏性操作。

## 4 Operation Matrix

| operationId | fixture | frontend consumer | backend handler | persistence | AI dependency | scenario coverage |
|-------------|---------|-------------------|-----------------|-------------|---------------|-------------------|
| `getRuntimeConfig` | `openapi/fixtures/Auth/getRuntimeConfig.json#default` | `AppRuntimeProvider` | backend-auth runtime config handler | 无 | 无 | 当前无真实 E2E owner；root `make test` |
| `getMe` | `openapi/fixtures/Auth/getMe.json#authenticated\|unauthenticated\|profileIncomplete` | `AppRuntimeProvider`、Settings、profile/protected-route guards | backend-auth current-user handler | `users` + session lookup；`user_settings.analytics_opt_in` 仅供 runtime config | 无 | `E2E.P0.101` session/profile + Settings 真实字段 |
| `startAuthEmailChallenge` | `openapi/fixtures/Auth/startAuthEmailChallenge.json#default` | `AuthLoginScreen` | backend-auth challenge handler | auth challenge storage | 无 | `E2E.P0.101` |
| `verifyAuthEmailChallenge` | `openapi/fixtures/Auth/verifyAuthEmailChallenge.json#default` | `AuthVerifyScreen` | backend-auth verify handler | challenge consumption + session | 无 | `E2E.P0.101` |
| `completeMyProfile` | `openapi/fixtures/Auth/completeMyProfile.json#default` | `AuthProfileSetupScreen` | backend-auth `PATCH /me` handler | backend user display name, terms acceptance and profile completion fields | 无 | root `make test`；`E2E.P0.101` profile completion |
| `logout` | `openapi/fixtures/Auth/logout.json#default` | `AuthLogoutScreen`，由 Settings 进入 | backend-auth logout handler | session revocation | 无 | `E2E.P0.101` logout/relogin |
| `deleteMe` | `openapi/fixtures/Auth/deleteMe.json#default` | `SettingsScreen` destructive confirmation | backend-auth delete handoff | users soft delete + all-session revoke + privacy delete job | 无 | `BDD.SHELL.SETTINGS.DELETE.001` + backend contract；不加入共享 E2E |
| `requestPrivacyExport` | `openapi/fixtures/Privacy/requestPrivacyExport.json#default` | Settings disabled/unavailable presentation；P0 不发伪请求 | Privacy P0 typed 501 handler | 无 | 无 | contract/component gate；不新增 E2E |
| `listTargetJobs` / `importTargetJob` | current fixtures | Home authenticated read/import guard | targetjob handlers | targetjob owner | import may enqueue AI parse | 当前无 auth-shell E2E owner；root `make test` |
| safe-read orchestration | existing generated operations | runtime and screen loaders | owner handlers | owner stores | unchanged | 当前无真实 E2E owner；root `make test` |
| UI shell/settings/display | N/A | routes、TopBar 设置齿轮、Settings、display provider | 无 | display preference only | 无 | `E2E.P0.101` 只承接设置入口/真实字段/logout；其余 UI/typecheck/build gates |

## 5 验收标准

- 默认打开 App 渲染 Home、三入口 TopBar、单一登录入口和显示控制；已登录时登录入口替换为单一设置齿轮。
- Browser History URL、hash adapter 输入和 in-memory route 均进入同一 normalization / route store 合同。
- 语言、主题与暗色在未登录、登录、退出登录和 `/me` refresh 中保持前端偏好优先级；generated client 请求携带当前 UI locale display hint；字体保持固定产品栈。
- 未登录用户触发受保护动作时进入 `auth_login(pendingAction)`；email-code 验证成功后先执行资料补全 gate，再恢复 safe route params。
- Settings 无 tab，复用 runtime user 展示真实姓名/脱敏邮箱，不重复调用 `getMe`；退出进入既有确认页，导出诚实显示暂不可用，删除账号覆盖确认、pending、失败重试，以及 `202` 后复用 `refreshAuth()` 重探测 `/me` 并回到 Home。
- 面试业务 route 在 runtime auth loading / unauthenticated 状态下不挂载业务 screen，不调用受保护 API；Home 未登录态不请求账号记录。
- Vite dev mock 从 unauthenticated 开始，verify 后 `/me` 变为 authenticated 或 profileIncomplete，logout 后回到 unauthenticated。
- Auth verify 成功后的 `/me` refresh failure 不被渲染为验证码错误；App 离开 verify 页并在 route gate 中表达 auth/profile loading 或 error。
- StrictMode 下同 key safe-read GET 同时在途只发出一个底层 request；settle 后可重新读取。不同 client/query/header/epoch/auth、带 signal、非 GET 与 `/auth/email/verify` 保持独立；verify 成功推进 auth epoch，auth/locale 变化不会被旧 key 吞并。
- UI 结构、文案、密度、主题和交互节奏可追溯到 `frontend/` 与 `docs/ui-design/`。

## 6 当前验证面

- Focused frontend gates: route normalization / URL codec、App auth dispatch、runtime provider、Auth screens、TopBar、DisplayPreferencesProvider、Settings visual、dev mock client、Home auth guard。
- Contract/doc gates: frontend-shell context validation、product-scope context validation、`sync-doc-index --check`、`make docs-check`、`git diff --check`、`make lint-core-loop-pruning-surface`。

### Phase 8: auth alias test lifecycle isolation

`AppAuthDispatch.test.tsx` 中的 `auth_reset` / `auth_register` 用例只验证同步 route normalization。两项在断言后显式 unmount，使不属于测试目标的 runtime-config/auth Promise 不再于测试尾部回写 provider state；生产 App、runtime provider、route behavior 与 BDD 合同不变。

门禁：AppAuthDispatch focused tests 仅作开发反馈并要求无 React act warning；阶段单测完成由仓库根 `make test` 承接，typecheck/build 与 owner docs 为独立 gates。BDD 不适用，因为本批只修正测试生命周期。

### Phase 9: i18n catalog reachability cleanup

以 TypeScript AST 建立通用 locale-key reachability gate：每个 zh/en message key 必须在 production TS/TSX 中有字面量 consumer。Report 的 tab/missing-state 动态 key 改为类型化字面量映射，Practice TopBar 接回当前原型已有的题号与暂停/继续文案；其余零 consumer 双语 key 直接删除，不保留兼容或退役清单。

门禁：locale gate 先报告 46 个 orphan/dynamic key，修正后归零；focused Home/Practice/Report/Shell tests、typecheck/build、UI truth-source tests、owner contexts 与 docs/diff/pruning gates 通过。

### Phase 10: auth prototype call-surface pruning

`frontend/src` 只向 auth 原型 screen 传递真实消费的参数：登录页保留 `nav` 以进入验证码页，验证码页保留 `onSignIn` 以完成登录，资料补全页保留 `onCompleteProfile` 以恢复 pendingAction。删除登录页从未读取的 `onSignIn` 与资料补全页从未读取的 `nav`，同时删除调用方对应传参，不增加兼容参数或空转 wrapper。

门禁：UI contract 先对当前冗余签名和调用方传参失败，删除后以 AST 证明 auth 原型参数全部有读取点；focused UI/auth gates、静态浏览器 auth route smoke、full frontend、typecheck/build、owner contexts 与 docs/diff/pruning gates 通过。BDD 不适用，因为本批不改变 email-code、profile setup 或 pendingAction 用户行为。

### Phase 11: settings prototype call-surface pruning

`SettingsScreen` 只消费主题 token、语言、字体预设与字体更新回调；设置页内部通过本地 tab 切换展示个人资料与隐私内容，不发起 route 跳转。删除从未读取的 `nav` 形参与 `app.jsx` 对应传参，保留 `fontPreset` / `setFontPreset` 显示设置链，不增加空转参数或 wrapper。




### Phase 13: StrictMode-safe GET single-flight

以 focused RED 先复现 StrictMode 双 mount 对同一 logical safe-read GET 发出两次底层 request，再在共享 generated/runtime client 边界引入仅在途 single-flight registry。key 必须精确覆盖 client identity、method/path/query、规范化相关 headers、normalized `okStatuses`、read/auth epoch 与 auth/session scope；Promise resolve/reject 后驱逐。caller `AbortSignal`、所有非 GET 与语义写入 GET 绕过 registry，避免共享取消权和写请求语义。每个语义写请求必须在 dispatch 前与 resolve/reject settle 后分别推进 read epoch；`/auth/email/verify` 必须显式按写语义 bypass，并在成功后另推进 auth/session epoch。

### Phase 14: Settings simplification and real account actions

先以 focused component/source RED 固化当前可移除面：头像姓名 chip、用户 dropdown、TopBar logout、settings tab rail、静态账号/隐私列表、登录与安全、字体预设、产品信息及无后端事实字段。随后把已登录 TopBar 收敛为直接导航到 `settings` 的设置齿轮，并让 `SettingsScreen` 从 `AppRuntimeContext.auth.user` 读取 `displayName/emailMasked`，不得新增第二次 `getMe`。

Account 区用只读语义行展示真实值并进入既有 `auth_logout`；Privacy 区把 P0 export 呈现为带原因的禁用/暂不可用状态，不渲染可触发动作，也不发送会被误读为成功的请求。账号删除使用 generated `deleteMe`：打开 destructive confirm 时生成一次 idempotency key，同一确认生命周期内失败重试复用该 key；对话框具备 destructive description、初始/约束/归还焦点与 Escape/取消；pending 禁止关闭和重复提交；网络/服务失败留在对话框并显示可恢复错误，`401` 转入统一认证重探测。`202` 后关闭对话框，调用现有 `refreshAuth()` 重探测 `/me`（预期 401）并提交 unauthenticated runtime，再 replace 到 Home；重探测网络/服务错误保留 honest auth error。未登录直开 `/settings` 仍由统一 protected-route guard 转入登录和 safe pendingAction；不得新增 `clearAuth` 或第二套 session mutation。

本 Phase 与 `frontend-shell/002` Phase 20、`backend-auth/001` Phase 10、B2 001/002/003 settings correction 和 B4 001 Phase 13 同批交接；不保留旧 testid、CSS selector、locale key、font package 或兼容状态。


## 7 风险与应对

| 风险 | 应对措施 |
|------|----------|
| Shell owner 扩大到业务页面实现 | 本 plan 只 owning App shell、auth、settings、display 和 route/auth gate；业务页面内容由对应 subject owner 承接 |
| Auth flow 绕过 generated client/session cookie | Operation matrix 固化 generated auth operations；focused tests 阻止自定义 session wire |
| PendingAction 泄露敏感正文 | Safe-param allowlist 与 URL/privacy tests 只允许稳定 ID 和 display hint |
| Route catalog 漂移 | `normalizeRoute` / `routeUrl` focused tests 与 BDD gates 共同验证 unsupported input 不 materialize 独立页面 |
| UI 与原型偏离 | `frontend/` 源码和 `docs/ui-design/` 是唯一 UI 设计文档；可见变更必须先更新原型再迁移正式前端 |
| 删除账号被重复提交、键盘不可达或失败后失去恢复路径 | confirmation lifecycle 复用同一 idempotency key，dialog 管理 focus/Escape/取消，pending 锁定关闭与提交，recoverable 错误留在对话框并允许重试，`401` 交给统一 auth probe；仅 `202` 后复用 `refreshAuth()` 重探测 `/me` |

## 8 修订记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-07-15 | 1.29 | Reopen Phase 14 for the approved single settings icon, real account/privacy page, logout relocation, deleteMe state machine and removal of static/font-preset surfaces. |
| 2026-07-14 | 1.28 | Separate code-owned shell/auth BDD from the Ready-only P0.101 real API/UI handoff. |
| 2026-07-14 | 1.27 | Complete the single-flight identity with normalized okStatuses and fence every semantic mutation before dispatch and after settle. |
| 2026-07-14 | 1.26 | Reopen for StrictMode-safe GET single-flight and stable AppRuntimeProvider dependencies. |
| 2026-07-10 | 1.25 | Delete the zero-consumer Auth link-row CSS wrapper. |
| 2026-07-10 | 1.24 | Remove the unread Settings prototype navigation prop and caller argument. |
| 2026-07-10 | 1.23 | Prune two unread auth prototype props and their caller arguments. |
| 2026-07-10 | 1.22 | Add production locale-key reachability and remove orphan catalog entries. |
| 2026-07-10 | 1.21 | Isolate synchronous auth alias tests from unrelated runtime-provider updates. |
| 2026-07-07 | 1.20 | Compress owner plan to the current App shell / auth / settings contract, operation matrix, and current gate surface. |
