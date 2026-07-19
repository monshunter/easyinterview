# App Shell, Auth Gate, and Settings Entrypoints

> **版本**: 1.40
> **状态**: active
> **更新日期**: 2026-07-20

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)
**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)
**关联 BDD Checklist**: [bdd-checklist](./bdd-checklist.md)

## 1 目标

本计划交付当前正式前端 App shell：默认 Home、三入口 TopBar、email-code 认证页、资料补全 gate、`requestAuth(pendingAction)` 恢复、明确齿轮设置入口、含账号级 Appearance 的无 tab 设置页、runtime/generated client bootstrap，以及面试业务 route 的登录前置保护。

当前 Phase 16 原地承接设置/主题/Practice chrome 修订：现有 `PATCH /me` 按用户确认的方案 B 泛化为 `updateMe`，主题写入账号设置，并锁定 route 切换零重复 `/me`。Phase 1-15 的既有状态仅为历史证据。

Phase 1-13 的已勾选内容只保留历史交付证据；Phase 14 是当前设置合同 owner，并取代其中关于账号 chip/dropdown、Settings tab 与 font preset 的旧正向描述。实施与验收不得把历史文字当作现行 UI 要求。

## 2 当前合同

### 2.1 UI 与 route catalog

- UI 设计文档：`docs/ui-design/`、`frontend/src`。
- 一级 TopBar 入口：`home`、`workspace`、`resume_versions`。
- 上下文 route：`parse`、`practice`、`generating`、`report`。
- 账号入口 route：已登录 TopBar 用户名首字符设置按钮直达 `settings`；`auth_logout` 从设置页进入。
- 认证 route：`auth_login`、`auth_verify`、`auth_profile_setup`、`auth_logout`。
- Unsupported route / malformed URL 必须经同一 normalization 层折回当前 route catalog 或 `home`，不得 materialize 独立页面。

### 2.2 Auth / settings / display

- Auth UI 只通过 email-code flow 触发 `startAuthEmailChallenge`、`verifyAuthEmailChallenge`、`getMe`、`updateMe`、`logout` 和 first-party session cookie；`updateMe` 同时承接首次资料补全和 authenticated 主题更新。
- `profileCompletionRequired=true` 时，登录后必须先进入 `auth_profile_setup`；资料补全成功并刷新 `/me` 后，才恢复 pendingAction 或回 Home。
- `pendingAction` 只保存 route name、canonical URL 和 safe params，不保存 JD 原文、简历原文、验证码、AI prompt/response 或解析正文。
- Settings 统一命名为“设置 / Settings”且无 tab：Appearance 本地预览并保存账号级主题；Account 复用 runtime user 展示只读 `displayName/email`；Privacy 保留退出、导出不可用和账号删除。
- `getMe` 只在应用 bootstrap/auth recovery 读取完整 `UserContext`；Settings 挂载、普通 route 切换和 Practice 进入/离开均不得再次读取。主题保存只发送一次 `updateMe`，成功响应直接更新 runtime/display context，不追加 GET。
- 统一 auth route gate 的 loading/error eyebrow、title、body 必须消费 typed locale keys；中文模式不得回退或混入英文硬编码。
- TopBar 语言下拉使用清晰 SVG chevron 表达开合状态；设置入口字符从现有 authenticated runtime `displayName` 派生，名称为空显示 `?`，不新增 `getMe` 或账号菜单。

### 2.3 StrictMode-safe GET orchestration

- React StrictMode 保持开启；不得通过关闭 StrictMode、延迟 mount 或 screen-local boolean 掩盖重复 GET。
- generated/runtime client 的共享 in-flight registry 只合并同时在途的语义只读 GET。逻辑 key 至少包含 client identity、method、path、canonical query、规范化相关 headers（含 `Accept-Language`）、normalized `okStatuses`、read/auth epoch 与 auth/session scope。
- resolve/reject 都必须立即驱逐 registry entry；这是 single-flight，不是跨 settle cache。不同 client/query/header/okStatuses/epoch/auth、带 caller `AbortSignal` 的 GET、所有非 GET 与语义写入 GET 必须绕过合并。每个语义写请求在 dispatch 前与 resolve/reject settle 后都推进 read epoch，切断 mutation 前、期间与之后的错误复用。`/auth/email/verify` 虽是 GET wire method，但会消费 challenge/更新 session，必须按语义写请求 bypass；成功后另推进 auth/session epoch。
- `AppRuntimeProvider` 以及 Home / `useRecentTargetJobs`、Parse、`useWorkspaceTargetJobs`、Reports、Practice 等 screen loader 的 bootstrap/refresh effect 依赖稳定 client、auth、request-option 与 route identity 输入，不把每次 render 都变化的整体 runtime object 作为依赖；auth/locale/epoch 真实变化仍触发独立请求。

## 3 质量门禁分类

- **Plan 类型**: `feature-behavior` + `frontend` + `contract`。
- **TDD 策略**: 本计划按 `/implement frontend-shell/001-app-shell-auth-settings frontend` -> `/tdd` 完成。focused Vitest / component / route-state test 只用于开发反馈；阶段完成由仓库根 `make test` 承接前后端全量单测。
- **BDD 策略**: 既有 auth/settings behaviors 保持；`BDD.SHELL.SETTINGS.THEME.001` 覆盖主题预览、一级选项与 Save 固定主行、自定义二级展开和请求预算；`BDD.SHELL.TOPBAR.IDENTITY.005` 覆盖语言 chevron 与用户名首字符设置入口。`E2E.P0.101` 保持真实主题保存、logout/relogin 恢复；账号删除与本次纯 shell 显示修订不新增 E2E。

## 4 Operation Matrix

| operationId | fixture | frontend consumer | backend handler | persistence | AI dependency | scenario coverage |
|-------------|---------|-------------------|-----------------|-------------|---------------|-------------------|
| `getRuntimeConfig` | `openapi/fixtures/Auth/getRuntimeConfig.json#default` | `AppRuntimeProvider` | backend-auth runtime config handler | 无 | 无 | 当前无真实 E2E owner；root `make test` |
| `getMe` | `openapi/fixtures/Auth/getMe.json#authenticated\|unauthenticated\|profileIncomplete` | `AppRuntimeProvider` bootstrap/auth recovery；Settings/route 仅消费内存 context | backend-auth current-user handler | `users` + `user_settings` account theme + session lookup | 无 | `E2E.P0.101` session/profile/theme restore；component request-count gate |
| `startAuthEmailChallenge` | `openapi/fixtures/Auth/startAuthEmailChallenge.json#default` | `AuthLoginScreen` | backend-auth challenge handler | auth challenge storage | 无 | `E2E.P0.101` |
| `verifyAuthEmailChallenge` | `openapi/fixtures/Auth/verifyAuthEmailChallenge.json#default` | `AuthVerifyScreen` | backend-auth verify handler | challenge consumption + session | 无 | `E2E.P0.101` |
| `updateMe` | `openapi/fixtures/Auth/updateMe.json#profileCompletion\|themeOcean\|themePlum\|customAccent` | `AuthProfileSetupScreen` + `SettingsScreen` | backend-auth generic `PATCH /me` handler | one transaction over user profile/terms + `user_settings` theme/custom accent | 无 | root `make test`；`E2E.P0.101` profile completion + theme save/relogin |
| `logout` | `openapi/fixtures/Auth/logout.json#default` | `AuthLogoutScreen`，由 Settings 进入 | backend-auth logout handler | session revocation | 无 | `E2E.P0.101` logout/relogin |
| `deleteMe` | `openapi/fixtures/Auth/deleteMe.json#default` | `SettingsScreen` destructive confirmation | backend-auth delete handoff | users soft delete + all-session revoke + privacy delete job | 无 | `BDD.SHELL.SETTINGS.DELETE.001` + backend contract；不加入共享 E2E |
| `requestPrivacyExport` | `openapi/fixtures/Privacy/requestPrivacyExport.json#default` | Settings disabled/unavailable presentation；P0 不发伪请求 | Privacy P0 typed 501 handler | 无 | 无 | contract/component gate；不新增 E2E |
| `listTargetJobs` / `importTargetJob` | current fixtures | Home authenticated read/import guard | targetjob handlers | targetjob owner | import may enqueue AI parse | 当前无 auth-shell E2E owner；root `make test` |
| safe-read orchestration | existing generated operations | runtime and screen loaders | owner handlers | owner stores | unchanged | 当前无真实 E2E owner；root `make test` |
| UI shell/settings/display | N/A | routes、语言 SVG chevron、runtime username initial 设置入口、Settings Appearance、display/runtime provider、Practice chrome | 无 | theme only via `updateMe`; language/dark remain client display state | 无 | `E2E.P0.101` 设置/主题/logout/relogin；其余 component/responsive/Chrome gates |

## 5 验收标准

- 默认打开 App 渲染 Home、三入口 TopBar、单一登录入口和显示控制；语言菜单有清晰开合 chevron；已登录时登录入口替换为从 runtime 用户名派生首字符的单一设置按钮。
- Browser History URL、hash adapter 输入和 in-memory route 均进入同一 normalization / route store 合同。
- 语言与暗色保持现有客户端显示合同；authenticated 主题由 bootstrap/auth recovery 的 `getMe` 恢复。普通 route 切换不得触发 `/me`；主题 slider 零请求，保存一次 `updateMe`，成功无 follow-up GET。
- Settings Appearance 的 desktop 主操作行必须同时持有 Ocean / Plum / Custom 一级选项与 Save，选项靠左、Save 靠右；Custom hue/chroma 只在该行下方展开，切换前后 Save 的纵向位置不变。窄屏按同一 DOM 顺序安全换行且不得横向溢出。
- 未登录用户触发受保护动作时进入 `auth_login(pendingAction)`；email-code 验证成功后先执行资料补全 gate，再恢复 safe route params。
- Settings 无 tab，复用 runtime user 展示真实姓名/完整邮箱，不重复调用 `getMe`；完整邮箱不写入日志/场景证据；退出进入既有确认页，导出诚实显示暂不可用，删除账号覆盖确认、pending、失败重试，以及 `202` 后复用 `refreshAuth()` 重探测 `/me` 并回到 Home。
- 面试业务 route 在 runtime auth loading / unauthenticated 状态下不挂载业务 screen，不调用受保护 API；Home 未登录态不请求账号记录。
- Vite dev mock 从 unauthenticated 开始，verify 后 `/me` 变为 authenticated 或 profileIncomplete，logout 后回到 unauthenticated。
- Auth verify 成功后的 `/me` refresh failure 不被渲染为验证码错误；App 离开 verify 页并在 route gate 中表达 auth/profile loading 或 error。
- Auth loading/error route gate 的四段可见文案全部跟随当前 `zh`/`en` 显示偏好；切换语言不改变 route、auth probe 或业务 API gate。
- StrictMode 下同 key safe-read GET 同时在途只发出一个底层 request；settle 后可重新读取。不同 client/query/header/epoch/auth、带 signal、非 GET 与 `/auth/email/verify` 保持独立；verify 成功推进 auth epoch，auth/locale 变化不会被旧 key 吞并。
- UI 结构、文案、密度、主题和交互节奏可追溯到 `frontend/` 与 `docs/ui-design/`；Practice 显示全局 TopBar 与独立 Practice Session Header。

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

先以 focused component/source RED 固化当前可移除面：头像姓名 chip、用户 dropdown、TopBar logout、settings tab rail、静态账号/隐私列表、登录与安全、字体预设、产品信息及无后端事实字段。随后把已登录 TopBar 收敛为直接导航到 `settings` 的设置齿轮，并让 `SettingsScreen` 从 `AppRuntimeContext.auth.user` 读取 `displayName/email`，不得新增第二次 `getMe`；`emailMasked` 必须作为旧合同零引用删除。

Account 区用只读语义行展示真实值并进入既有 `auth_logout`；Privacy 区把 P0 export 呈现为带原因的禁用/暂不可用状态，不渲染可触发动作，也不发送会被误读为成功的请求。账号删除使用 generated `deleteMe`：打开 destructive confirm 时生成一次 idempotency key，同一确认生命周期内失败重试复用该 key；对话框具备 destructive description、初始/约束/归还焦点与 Escape/取消；pending 禁止关闭和重复提交；网络/服务失败留在对话框并显示可恢复错误，`401` 转入统一认证重探测。`202` 后关闭对话框，调用现有 `refreshAuth()` 重探测 `/me`（预期 401）并提交 unauthenticated runtime，再 replace 到 Home；重探测网络/服务错误保留 honest auth error。未登录直开 `/settings` 仍由统一 protected-route guard 转入登录和 safe pendingAction；不得新增 `clearAuth` 或第二套 session mutation。

本 Phase 与 `frontend-shell/002` Phase 20、`backend-auth/001` Phase 10、B2 001/002/003 settings correction 和 B4 001 Phase 13 同批交接；不保留旧 testid、CSS selector、locale key、font package 或兼容状态。

Review remediation 继续由本 Phase owning：默认 fixture-backed client 在 `deleteMe` 202 后必须把 auth state 切到 signed-out，使随后的 `refreshAuth()` / `getMe` 得到 401；`E2E.P0.101` 的邮箱断言不得把完整值写入 Playwright failure reporter，并在 reporter 输出进入 `trigger.log` 前同时过滤原文与 URL percent-encoded 表示。两项均先补当前可失败的 focused/code-level gate，再做最小实现；真实 E2E 主路径仍只覆盖 settings account/logout，不执行账号删除。

### Phase 15: Auth route gate locale drift remediation

先以 App shell locale behavior RED 复现：当前语言为中文且 `/me` probe loading/error 时，统一 `auth-route-gate` 仍渲染硬编码 `AUTH`、英文标题和英文说明。GREEN 仅把 eyebrow、loading/error title 与 body 接入现有 typed locale catalog；不改变 protected-route 判定、pendingAction、请求时序或 auth 状态机。中英文切换、业务 screen/API 不提前挂载、locale key reachability、typecheck/build 与根回归共同验收；current-run 还必须使用 Chrome extension automation skill 在真实本地前后端页面核对中文 gate 与英文切换，不新增或冒充 E2E ID。

### Phase 16: Account theme persistence and Practice global chrome

先以 OpenAPI/backend/frontend RED 锁定方案 B：`PATCH /me` operationId 改为 `updateMe`，request 支持 profile-completion 字段与可选 `displayPreferences.theme/customAccent`，response 为完整 `UserContext`。`user_settings` 新增 typed theme/custom columns 与 enum/range/all-or-none constraints；GET/PATCH 在 owner store 中读写同一投影，非法值 fail closed，legacy row 默认 ocean。profile-only、theme-only 和组合更新必须单事务，空 patch、字段组合错误与非法范围拒绝且不产生部分写入。

前端把主题菜单从 TopBar 移入 Settings Appearance；设置入口文案统一为“设置 / Settings”，图标使用明确齿轮。Appearance 以 runtime server-confirmed 值初始化草稿；选择/拖动即时本地预览且零请求，Save 只发一次 `updateMe`。成功以响应 `UserContext` 同步 runtime 与 display provider，不发 follow-up `getMe`；失败保留草稿/预览和错误，离开未保存页面恢复确认值。App bootstrap/auth recovery 读取一次 `/me`，Home/Settings/Practice 等 route 切换请求计数保持为零。

`practice` 从 no-chrome allowlist 移除；全局 App TopBar 与 Practice Session Header 同时渲染并在 desktop/mobile 无溢出。Focused request-count/component/responsive/a11y tests、OpenAPI fixture/codegen drift、migration/backend contract、根 `make test`、`E2E.P0.101` 真实保存/重登、Chrome desktop/mobile 截图共同收口。

### Phase 17: Account theme L2 review remediation

本 Phase 原地修复 account-theme code review 发现的证据与竞态缺口：Settings 的迟到保存响应不得在页面卸载、退出或账号身份变化后回写旧 auth/theme；runtime 对未知、越界或非 closed 的服务端主题投影统一 fail closed 为 `ocean + null`；dev mock 对出现但非法的 `displayPreferences` 必须在任何 profile/theme state mutation 前整体拒绝，不得产生真实 backend 不会接受的部分成功。

同时同步 product-scope D-21、OpenAPI current schema inventory / handoff、frontend README 与 UI architecture 的当前合同，旧 `completeMyProfile` / TopBar theme menu / “设置与隐私”只能留在明确历史证据中。Focused failure/retry/leave/race/parity tests、根 `make test`、build/docs/codegen/diff gates 与重新部署后的 `E2E.P0.101` 共同作为本轮验收；Phase 15 的既有 locale Chrome 项仍独立保持未完成，主 plan 不因本 Phase 通过而虚假关闭。

### Phase 18: Screenshot-aligned Auth and Settings composition

先以 `AuthVisual`、Auth screen 与 Settings visual/component tests 固化参考稿可执行结构：desktop 宽幅双栏认证 shell、左侧标题/原则卡、右侧主操作卡、登录/验证码装饰插画、退出页堆叠确认，以及设置页 Header 插画和三张横向功能卡。旧 `1160px` 窄 auth shell、Settings `980px` 纵向块和退出双按钮行必须形成 RED。

GREEN 只调整正式 React DOM、page-scoped CSS、typed locale 与仓库内 SVG/CSS 装饰；保留 email-code、pendingAction、verify、logout、runtime user、主题 `updateMe`、导出不可用与 `deleteMe` 的现有状态机和请求预算。desktop 参考 viewport 与 mobile 单列均须无横向溢出，装饰不进入阅读顺序，也不得伪造验证码倒计时、重发成功或任何账号能力。

`BDD.SHELL.PAGES.VISUAL.002` 由 component/responsive/a11y tests 与 current-run Chrome desktop/mobile 真实页面验收承接，不建立第二套 Demo 或像素基线。完成 focused、locale reachability、typecheck/build、根 `make test`、owner context、Header/INDEX/docs/diff gate 后收口本 Phase；Phase 15.3 的 auth probe 中间 loading/error gate 仍是独立 Chrome 证据缺口，在真实运行时未捕获到该瞬态前不得冒充完成，主 plan 因此继续保持 `active`。

### Phase 19: Settings custom theme disclosure hierarchy remediation

先以 Settings behavior/visual RED 固化两层主题结构：Ocean / Plum / Custom 一级选择器在预定义与自定义状态下都必须存在；自定义 hue/chroma 二级编辑器只在 Custom 激活时挂载，并且与一级选择器处于同一内容列的后续正常文档流。现有把一级选择器和二级编辑器分配到同一 grid area 的实现必须形成 RED，避免仅用 jsdom “节点仍存在”掩盖真实遮挡。

GREEN 只在 `SettingsScreen` 增加明确的 theme editor 分组，并调整 `settings-appearance` page-scoped CSS，使一级选择器与条件二级编辑器纵向堆叠；保存按钮继续独立位于操作列。hue range 使用完整光谱轨道，chroma range 使用当前 hue 的低彩到高彩渐变，并保留原生 range 的键盘、focus 与值域语义。选择 Ocean / Plum 必须继续清除 custom accent 并隐藏二级编辑器；主题预览、单次 `updateMe`、失败恢复、runtime 更新与零 follow-up `getMe` 合同不变。

`BDD.SHELL.SETTINGS.THEME.001` 增补一级常驻、二级按需展开和预定义主题可回退断言；desktop/mobile component tests、根 `make test` 与 current-run Chrome 的 Custom -> Ocean/Plum 切换、无遮挡和无横向溢出共同收口本 Phase。Phase 15.3 仍是独立未完成项，不能因本修复通过而关闭主 plan。


## 7 风险与应对

| 风险 | 应对措施 |
|------|----------|
| Shell owner 扩大到业务页面实现 | 本 plan 只 owning App shell、auth、settings、display 和 route/auth gate；业务页面内容由对应 subject owner 承接 |
| Auth flow 绕过 generated client/session cookie | Operation matrix 固化 generated auth operations；focused tests 阻止自定义 session wire |
| PendingAction 泄露敏感正文 | Safe-param allowlist 与 URL/privacy tests 只允许稳定 ID 和 display hint |
| Route catalog 漂移 | `normalizeRoute` / `routeUrl` focused tests 与 BDD gates 共同验证 unsupported input 不 materialize 独立页面 |
| UI 与原型偏离 | `frontend/` 源码和 `docs/ui-design/` 是唯一 UI 设计文档；可见变更必须先更新原型再迁移正式前端 |
| 删除账号被重复提交、键盘不可达或失败后失去恢复路径 | confirmation lifecycle 复用同一 idempotency key，dialog 管理 focus/Escape/取消，pending 锁定关闭与提交，recoverable 错误留在对话框并允许重试，`401` 交给统一 auth probe；仅 `202` 后复用 `refreshAuth()` 重探测 `/me` |

### Phase 20: Shared asynchronous transition visual system

先以 shared transition / TopBar tests 锁定统一蓝白画布、四种代码内 SVG variant、真实 indeterminate 语义、reduced-motion、mobile containment，以及 Practice/Parse/Reports/Generating 上下文 route 统一高亮“面试”。随后只新增 shell-owned `AsyncTransitionScene` 视觉骨架和 route-to-primary-nav 映射；业务状态、generated client、轮询、错误恢复与返回动作仍由原 owner 持有。

`BDD.SHELL.TRANSITION.VISUAL.003` 由 shared component、TopBar、route/chrome、responsive/a11y tests 与 current-run Chrome desktop/mobile 验收承接；不建立并行 Demo、不生成像素基线，也不把装饰性 indeterminate 轨道描述为后端百分比。Phase 15.3 的 auth probe 中间态仍是独立缺口，不能因本 Phase 完成而关闭主 plan。

### Phase 21: Settings Header security illustration remediation

先以 Settings visual/component RED 拒绝现有山形折线、人物轮廓和独立圆形对勾的稀疏 Header SVG，并固定目标分层：一个带顶部栏、头像与资料行的半透明账号窗口，窗口内右下柱状图，左下前景锁卡、右下前景盾牌对勾，以及两侧星芒。GREEN 只重画 `SettingsHeaderArt` 与其 page-scoped CSS；所有填充、描边、光晕与阴影从当前主题 accent/token 派生，整组继续 `aria-hidden`，在窄屏沿用既有隐藏规则，不引入图片请求、业务事实或新交互。

`BDD.SHELL.SETTINGS.ART.004` 由 Settings DOM layer contract、视觉 CSS/响应式断言与 current-run Chrome desktop 验收承接；Chrome 必须确认目标图形层级、Header/card 左边界与横向无溢出，且浏览器 error/warning 为零。完成 focused、typecheck/build、根 `make test`、owner context、Header/INDEX/docs/diff gate 后收口本 Phase；Phase 15.3 仍保持独立未完成，主 plan 继续为 `active`。

### Phase 22: TopBar language affordance and account initial

先以 TopBar component/visual RED 拒绝 `9px` 文本 `▾` 与固定 `E`，并锁定语言按钮使用 code-native SVG chevron、独立可见底板和展开旋转状态。设置入口只从 `AppShell` 已持有的 authenticated `displayName` 取 trim 后首个 Unicode 字符，拉丁字母大写；空名称显示 `?`，不读取新字段、不发新请求，也不恢复账号 dropdown。

`BDD.SHELL.TOPBAR.IDENTITY.005` 由 TopBar/App component、CSS responsive/a11y tests 与 current-run Chrome desktop 验收承接；验证中文用户名、语言菜单开合、设置直达、无横向溢出和零 browser error/warning，不新增 E2E ID。完成 focused、typecheck/build、根 `make test`、owner context、Header/INDEX/docs/diff gate 后收口本 Phase；Phase 15.3 仍保持独立未完成，主 plan 继续为 `active`。

### Phase 23: Settings theme primary-row action anchoring

先以 Settings component/visual RED 固定 DOM owner 与 CSS layout：Ocean / Plum / Custom 选项组和 Save 必须同属 `.ei-settings-theme-primary-row`，条件 Custom editor 与错误提示位于其后的正常文档流；旧的独立第三列 action 及跨整张 Appearance 内容垂直居中必须失败。GREEN 只重组 `SettingsScreen` 的主题区 DOM 与 page-scoped CSS；desktop 使用同一行左右分布，Custom 展开不会改变 Save 的 `top`/`bottom`，mobile 允许同一 primary row 内安全换行并保持 editor 在后，不改主题草稿、禁用态、单次 `updateMe`、错误恢复、runtime 同步或请求预算。

`BDD.SHELL.SETTINGS.THEME.001` 原地扩展此布局行为。current-run Chrome 必须在 desktop 分别量测 preset 与 Custom 状态的 Save bbox 差值不超过 1px，并确认选项/Save 同一行、editor 严格位于下方；390px mobile 验证 DOM 顺序、可操作性和无横向溢出。完成 focused、typecheck/build、根 `make test`、owner context、Header/INDEX/docs/diff gate 后更新 BUG-0193 与既有复盘；Phase 15.3 仍保持独立未完成，主 plan 继续为 `active`。

## 8 修订记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-07-20 | 1.40 | Reopen Phase 23 to anchor Save beside the first-level theme choices while the conditional custom editor expands only beneath that stable primary row. |
| 2026-07-20 | 1.39 | Reopen Phase 22 to strengthen the language-menu chevron and derive the Settings entry mark from the authenticated display name without changing navigation or request budgets. |
| 2026-07-20 | 1.38 | Reopen Phase 21 to redraw the Settings Header as the approved layered profile, chart, lock, shield and sparkle security illustration without changing account behavior. |
| 2026-07-19 | 1.37 | Reopen Phase 20 for a shared screenshot-aligned asynchronous transition scene, persistent TopBar chrome and context-route navigation mapping. |
| 2026-07-19 | 1.36 | Reopen Phase 19 to keep first-level Settings theme choices visible, stack the conditional custom editor below them, and add full-spectrum hue plus hue-aware chroma tracks without changing persistence behavior. |
| 2026-07-19 | 1.35 | Reopen Phase 18 for screenshot-aligned login, verify, logout and settings compositions while preserving the existing generated-client and state-machine contracts. |
| 2026-07-19 | 1.34 | Reopen Phase 17 to remediate stale account-theme responses, fail-closed runtime projection, dev-mock atomic parity and active owner contract drift found by L2 review. |
| 2026-07-19 | 1.33 | Reopen Phase 16 for generic updateMe, account-persisted theme with zero route refetch, Settings naming/gear, and Practice global chrome. |
| 2026-07-16 | 1.32 | Reopen Phase 15 to localize the shared auth loading/error route gate without changing auth behavior. |
| 2026-07-15 | 1.31 | Complete Phase 14 review remediation for fixture delete auth state and failure-path evidence redaction. |
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
