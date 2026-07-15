# App Shell, Auth Gate, and Settings Entrypoints

> **版本**: 1.28
> **状态**: completed
> **更新日期**: 2026-07-14

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)
**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)
**关联 BDD Checklist**: [bdd-checklist](./bdd-checklist.md)

## 1 目标

本计划交付当前正式前端 App shell：默认 Home、三入口 TopBar、全局显示控制、email-code 认证页、资料补全 gate、`requestAuth(pendingAction)` 恢复、用户菜单、settings 双 tab、runtime / generated client bootstrap，以及面试业务 route 的登录前置保护。

当前完成态文档只描述现行合同。任何新增可见页面、route、auth flow 或设置页能力，必须先更新 `frontend/` 静态原型、`docs/ui-design/` 和 `frontend-shell` spec，再修订本 owner 或派生明确边界的新 plan。

## 2 当前合同

### 2.1 UI 与 route catalog

- UI 设计文档：`docs/ui-design/`、`frontend/src`。
- 一级 TopBar 入口：`home`、`workspace`、`resume_versions`。
- 上下文 route：`parse`、`practice`、`generating`、`report`。
- 用户菜单 route：`settings`、`auth_logout`。
- 认证 route：`auth_login`、`auth_verify`、`auth_profile_setup`、`auth_logout`。
- Unsupported route / malformed URL 必须经同一 normalization 层折回当前 route catalog 或 `home`，不得 materialize 独立页面。

### 2.2 Auth / settings / display

- Auth UI 只通过 email-code flow 触发 `startAuthEmailChallenge`、`verifyAuthEmailChallenge`、`getMe`、`completeMyProfile`、`logout` 和 first-party session cookie。
- `profileCompletionRequired=true` 时，登录后必须先进入 `auth_profile_setup`；资料补全成功并刷新 `/me` 后，才恢复 pendingAction 或回 Home。
- `pendingAction` 只保存 route name、canonical URL 和 safe params，不保存 JD 原文、简历原文、验证码、AI prompt/response 或解析正文。
- Settings 只保留 `个人资料` 与 `隐私与数据` 两个 tab；`个人资料` tab 承接账号基础信息、登录与安全展示、字体预设和产品信息。
- 显示偏好由前端持有：主题、暗色、语言下拉和字体预设在登录前后保持稳定；默认主题与无效值 fallback 为 `ocean`。

### 2.3 StrictMode-safe GET orchestration

- React StrictMode 保持开启；不得通过关闭 StrictMode、延迟 mount 或 screen-local boolean 掩盖重复 GET。
- generated/runtime client 的共享 in-flight registry 只合并同时在途的语义只读 GET。逻辑 key 至少包含 client identity、method、path、canonical query、规范化相关 headers（含 `Accept-Language`）、normalized `okStatuses`、read/auth epoch 与 auth/session scope。
- resolve/reject 都必须立即驱逐 registry entry；这是 single-flight，不是跨 settle cache。不同 client/query/header/okStatuses/epoch/auth、带 caller `AbortSignal` 的 GET、所有非 GET 与语义写入 GET 必须绕过合并。每个语义写请求在 dispatch 前与 resolve/reject settle 后都推进 read epoch，切断 mutation 前、期间与之后的错误复用。`/auth/email/verify` 虽是 GET wire method，但会消费 challenge/更新 session，必须按语义写请求 bypass；成功后另推进 auth/session epoch。
- `AppRuntimeProvider` 以及 Home / `useRecentTargetJobs`、Parse、`useWorkspaceTargetJobs`、Reports、Practice 等 screen loader 的 bootstrap/refresh effect 依赖稳定 client、auth、request-option 与 route identity 输入，不把每次 render 都变化的整体 runtime object 作为依赖；auth/locale/epoch 真实变化仍触发独立请求。

## 3 质量门禁分类

- **Plan 类型**: `feature-behavior` + `frontend` + `contract`。
- **TDD 策略**: 本计划按 `/implement frontend-shell/001-app-shell-auth-settings frontend` -> `/tdd` 完成。focused Vitest / component / route-state test 只用于开发反馈；阶段完成由仓库根 `make test` 承接前后端全量单测。
- **BDD 策略**: `BDD.SHELL.AUTH.001` 由代码层 owner tests 验证 auth guard、profile setup、pendingAction 与 settings 行为，并由仓库根 `make test` 统一回归；`E2E.P0.101` 仅作为真实 email-code、session 与 profile completion 的独立 handoff，只有显式真实运行后才产生 PASS。safe-read 与通用 shell 合同继续由独立 typecheck/build/UI/docs gates 维护。

## 4 Operation Matrix

| operationId | fixture | frontend consumer | backend handler | persistence | AI dependency | scenario coverage |
|-------------|---------|-------------------|-----------------|-------------|---------------|-------------------|
| `getRuntimeConfig` | `openapi/fixtures/Auth/getRuntimeConfig.json#default` | `AppRuntimeProvider` | backend-auth runtime config handler | 无 | 无 | 当前无真实 E2E owner；root `make test` |
| `getMe` | `openapi/fixtures/Auth/getMe.json#authenticated\|unauthenticated\|profileIncomplete` | `AppRuntimeProvider`、TopBar、profile/protected-route guards | backend-auth current-user handler | session cookie lookup | 无 | `E2E.P0.101` 仅 session/profile 状态 |
| `startAuthEmailChallenge` | `openapi/fixtures/Auth/startAuthEmailChallenge.json#default` | `AuthLoginScreen` | backend-auth challenge handler | auth challenge storage | 无 | `E2E.P0.101` |
| `verifyAuthEmailChallenge` | `openapi/fixtures/Auth/verifyAuthEmailChallenge.json#default` | `AuthVerifyScreen` | backend-auth verify handler | challenge consumption + session | 无 | `E2E.P0.101` |
| `completeMyProfile` | `openapi/fixtures/Auth/completeMyProfile.json#default` | `AuthProfileSetupScreen` | backend-auth `PATCH /me` handler | backend user display name, terms acceptance and profile completion fields | 无 | root `make test`；`E2E.P0.101` profile completion |
| `logout` | `openapi/fixtures/Auth/logout.json#default` | `AuthLogoutScreen`、TopBar | backend-auth logout handler | session revocation | 无 | `E2E.P0.101` 仅 logout/relogin |
| `listTargetJobs` / `importTargetJob` | current fixtures | Home authenticated read/import guard | targetjob handlers | targetjob owner | import may enqueue AI parse | 当前无 auth-shell E2E owner；root `make test` |
| safe-read orchestration | existing generated operations | runtime and screen loaders | owner handlers | owner stores | unchanged | 当前无真实 E2E owner；root `make test` |
| UI shell/settings/display | N/A | routes、TopBar、Settings、display provider | 无 | display preference only | 无 | 当前无真实 E2E owner；UI/typecheck/build gates |

## 5 验收标准

- 默认打开 App 渲染 Home、三入口 TopBar、单一登录入口、用户菜单区和显示控制。
- Browser History URL、hash adapter 输入和 in-memory route 均进入同一 normalization / route store 合同。
- 语言、主题、暗色与字体预设在未登录、登录、退出登录和 `/me` refresh 中保持前端偏好优先级；generated client 请求携带当前 UI locale display hint。
- 未登录用户触发受保护动作时进入 `auth_login(pendingAction)`；email-code 验证成功后先执行资料补全 gate，再恢复 safe route params。
- Settings 只有 `个人资料` / `隐私与数据` 两个 tab；账号登录与安全展示保持 email-code 口径。
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


## 7 风险与应对

| 风险 | 应对措施 |
|------|----------|
| Shell owner 扩大到业务页面实现 | 本 plan 只 owning App shell、auth、settings、display 和 route/auth gate；业务页面内容由对应 subject owner 承接 |
| Auth flow 绕过 generated client/session cookie | Operation matrix 固化 generated auth operations；focused tests 阻止自定义 session wire |
| PendingAction 泄露敏感正文 | Safe-param allowlist 与 URL/privacy tests 只允许稳定 ID 和 display hint |
| Route catalog 漂移 | `normalizeRoute` / `routeUrl` focused tests 与 BDD gates 共同验证 unsupported input 不 materialize 独立页面 |
| UI 与原型偏离 | `frontend/` 源码和 `docs/ui-design/` 是唯一 UI 设计文档；可见变更必须先更新原型再迁移正式前端 |

## 8 修订记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-07-14 | 1.28 | Separate code-owned shell/auth BDD from the Ready-only P0.101 real API/UI handoff. |
| 2026-07-14 | 1.27 | Complete the single-flight identity with normalized okStatuses and fence every semantic mutation before dispatch and after settle. |
| 2026-07-14 | 1.26 | Reopen for StrictMode-safe GET single-flight and stable AppRuntimeProvider dependencies. |
| 2026-07-10 | 1.25 | Delete the zero-consumer Auth link-row CSS wrapper. |
| 2026-07-10 | 1.24 | Remove the unread Settings prototype navigation prop and caller argument. |
| 2026-07-10 | 1.23 | Prune two unread auth prototype props and their caller arguments. |
| 2026-07-10 | 1.22 | Add production locale-key reachability and remove orphan catalog entries. |
| 2026-07-10 | 1.21 | Isolate synchronous auth alias tests from unrelated runtime-provider updates. |
| 2026-07-07 | 1.20 | Compress owner plan to the current App shell / auth / settings contract, operation matrix, and current gate surface. |
