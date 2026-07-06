# Frontend Shell Spec

> **版本**: 1.23
> **状态**: active
> **更新日期**: 2026-07-06

## 1 背景与目标

`frontend-shell` 是 `engineering-roadmap` S1 的第一个用户可见前端 workstream。它负责把当前静态 UI 中已经收敛的 App 壳、TopBar、三个一级入口、全局显示控制、用户菜单、认证页面和 pending action 恢复模型落到正式前端工程中。

本 subject 的目标是让后续 D2-D6 前端模块在同一个 App shell 内开发，而不是各自创建路由、认证跳转和显示偏好状态。

## 2 范围

### 2.1 In Scope

- App 默认进入 `home`，不展示独立 welcome。
- TopBar 三个一级入口：`home`、`workspace`、`resume_versions`（product-scope D-17 / D-22 收敛；旧 `jd_match`、`debrief` 一级入口删除，导航收敛实现证据由 `frontend-home-job-picks-and-parse/002` 与 product-scope/001 承接）。
- 上下文页面路由：`parse`、`practice`、`generating`、`report`。旧 `company_intel` 独立 route 已随 product-scope D-18 删除，归一为 `workspace` alias，实现证据由 `frontend-workspace-and-practice/001` 承接。其中 `parse` 只由本 subject 承接 route shell / chrome / params 边界，JD 解析业务内容归 `frontend-home-job-picks-and-parse`。
- 用户菜单入口：`settings`、`auth_logout`；旧 `profile` route 已随 product-scope D-22 删除，归一回 `home`。
- 认证页面：`auth_login`、`auth_verify`、`auth_profile_setup`、`auth_logout`。旧 `auth_reset` 页面与 route 已随 product-scope D-16 删除，归一回 `auth_login`；验证码重发与更换邮箱由 `auth_verify` 承担，登录页保留静态帮助说明。
- 设置页只保留 `个人资料` 与 `隐私与数据` 两个 tab（product-scope D-21）；个人资料 tab 内含账号基础信息、`登录与安全`（仅展示 `邮箱验证码 · 无密码`）、界面偏好（字体预设）与产品信息。
- `requestAuth(pendingAction)` 与登录成功后的 route / params 恢复。
- Mailpit / email-code 验证：`auth_verify` 承接用户手动输入的 6 位验证码，调用 generated `verifyAuthEmailChallenge` 后恢复目标 route；正式邮件不包含 magic link 或 URL token。
- 全局显示控制：主题色、暗色、语言下拉；设置页维护字体预设。
- Runtime config、generated API client 与 fixture-backed mock transport bootstrap 的前端接入边界。
- URL-addressable routing：正式前端使用 Browser History canonical path 表达当前 App location，并把稳定服务端资源 ID / 任务上下文保存在 query/path 参数中，支持直开、刷新、复制链接和浏览器前进/后退。

### 2.2 Out of Scope

- 不实现 D2-D6 业务页面细节：JD 导入、岗位推荐、模拟面试规划、练习 session、报告、简历工坊、复盘业务内容由后续 subject 承接。
- 不实现真实 passwordless 认证后端；后端能力归 `backend-auth`。
- 不新增旧 `welcome`、`growth`、`plan`、`mistakes`、`drill`、`followup`、`experiences`、`star`、独立 `voice` route；不恢复 `auth_reset`、`auth_register`、`jd_match`、`company_intel`、`debrief`、`profile` live route 或可见入口。
- 设置页不以空 tab 形式预留 `通知`、`订阅` 等 P1 占位能力（product-scope D-21）。
- 不把 `ui-design/src/data.jsx` 作为运行时数据源。
- 不把前端 URL 设计做成 OpenAPI / REST operation 的 1:1 镜像；URL 只表达用户所在页面与稳定上下文，不表达后端 action。
- 不在 URL、`pendingAction`、`localStorage`、session storage 或 browser history 中保存 JD 原文、简历原文、guided answers、解析结果、suggestion 文本、AI prompt / response 或其他敏感正文。

## 3 用户决策 / 待确认事项

| ID | 决策 | 锁定值 | 影响 |
|----|------|--------|------|
| D-1 | 默认入口 | `home` | 未登录也能看到首页并开始输入 JD 草稿 |
| D-2 | 一级导航 | `home` / `workspace` / `resume_versions`（product-scope D-17 / D-22 后三项；历史 `jd_match` / `debrief` 已裁剪） | 报告、语音、公司情报、真实复盘、画像和认证不进入一级导航；`jd_match` / `debrief` 不再是 live route，JD 获取唯一入口是首页导入 |
| D-3 | Auth gate | 操作级 `requestAuth(pendingAction)` | 登录不是默认落地页；成功后恢复原动作 |
| D-4 | 显示偏好 | TopBar 持有主题、暗色、语言；settings 持有字体预设；默认主题为 `ocean`（深海，product-scope D-21 v2.1 回调），主题菜单保留 warm / forest / ocean / plum 四预设 + `customAccent` | 登录状态不能重置显示偏好；无效或缺失主题 fallback `ocean` |
| D-5 | 数据源 | 前端 shell 通过 generated client + fixture-backed mock transport / runtime config 取数 | 不直接 import prototype data |
| D-6 | Auth API 边界 | D1 前端只消费 `startAuthEmailChallenge`、`verifyAuthEmailChallenge`、`getMe`、`completeMyProfile`、`logout` 和 first-party session cookie | 密码、OAuth、reset 不得以 live UI、stub 页面或 API 形式存在（product-scope D-16 无密码唯一流）；新增真实 API 前必须修订 C1 / B2 |
| D-7 | UI i18n | 正式前端至少支持 `zh` / `en` 两种 UI locale；每种语言必须有独立 locale 文件，语言元数据统一从 locale catalog 暴露，聚合层只负责导入、类型约束和 helper；UI 语言优先级为用户显式选择 > 浏览器 locale > `en` fallback，显式选择写入 `localStorage["ei-lang"]`；语言下拉只关联前端显示偏好，不依赖 runtime config 或登录态 | 语言选择必须按 `ui-design/src/app.jsx` 源码复刻为 TopBar icon dropdown：`topbar-lang-toggle` 显示 globe icon + 当前语言标签并打开 `topbar-lang-menu`，选项使用 `topbar-lang-option-{locale}`，改变用户可见文案，并通过 `Accept-Language` 影响后续 API display hint；后续新增语言只扩展 locale 文件、catalog 元数据和测试，不改变 TopBar 控件结构 |
| D-8 | Brand / version placement | TopBar 品牌区只展示 `E` mark + `EasyInterview`，品牌名不本地化，不放解释性副标题或版本号 | App 版本属于产品元数据，通过设置页 `产品信息 / Product info` 低频展示；版本值 `v1.0` 不进入翻译，周边标签必须走 i18n |
| D-9 | Dev mock session state | Vite dev 默认 fixture-backed mock 必须能表达未登录、登录成功和退出登录后的 session 状态变化 | operation-level fixture 存在不等于用户流程闭环；mock runtime 必须让 `verifyAuthEmailChallenge` 后的 `/me` 变为 authenticated，让 `logout` 后的 `/me` 变为 unauthenticated，且默认打开 App 必须可见非登录态 |
| D-10 | Canonical URL routing | 保持 SPA，但把正式导航从仅内存 route / `#route=` bootstrap 升级为 Browser History path + query。URL path 是 canonical 用户地址，内部仍使用现有 `Route { name, params }` / `LooseRoute` 合约；`#route=` 只作为 static preview / pixel parity / 迁移期 adapter 输入，不作为正式 canonical URL | 用户可复制、刷新、直开当前页面；back/forward 与 App route 同步；服务端部署必须把已知 frontend path fallback 到 `index.html`。URL 不等于 REST API，不新增旧 route alias，不恢复独立 `voice` |
| D-11 | Email code verification | `auth_verify` 只展示邮箱提示和 6 位验证码输入；generated client 仍按 B2 operation 的 `token` query 名调用 verify API，但 UI / 邮件 / 场景材料统一称为验证码，不再依赖 magic-link callback | Mailpit 邮件只给 code；code 不进入 `pendingAction`、storage、业务 route 或浏览器历史长链路 |
| D-12 | 单入口邮箱验证码登录 | 顶部和认证页只保留一个 `auth_login` 邮箱验证码入口；`startAuthEmailChallenge` 不再让用户选择注册或登录，也不得在发码前暴露邮箱是否已存在；邮箱验证成功后，既有账号直接登录，新邮箱创建资料未补全账号并进入 `auth_profile_setup` | 邮箱仍是唯一账号标识；displayName 不唯一、不参与登录或去重；旧 `auth_register` 不再是 live route 或可见入口 |
| D-13 | 首次登录资料补全 | `/me.profileCompletionRequired=true` 是前端强制跳转依据；用户只要登录态存在但资料未补全，无论首次验证后、关闭浏览器后重开、换浏览器重新登录、退出后重新登录、刷新或直开业务 URL，都必须先进入 `auth_profile_setup`，完成 displayName + 条款确认后再恢复 pendingAction 或回 Home | pendingAction 必须在资料补全前保持 safe params；资料补全页不得保存 raw JD / 简历 / prompt；补全成功后刷新 `/me` 并只在 `profileCompletionRequired=false` 时恢复业务动作 |
| D-14 | 面试业务路由登录前置 | `home` 和 auth 页面可未登录访问；`parse`、`workspace`、`resume_versions`、`practice`、`generating`、`report`、`settings` 等读取或写入用户面试上下文或账号设置的 route 必须在 runtime auth 明确为 authenticated 后才渲染业务 screen（历史口径中的 `jd_match` 已随 product-scope D-17 删除，`debrief` / `profile` 已随 D-22 删除并归一回 `home`） | 未登录直开或点击这些入口时统一进入 `auth_login(pendingAction)`；auth loading 期间不能提前挂载业务 screen 或发起受保护 API；Home 的 Recent mock interviews 模块仅已登录展示，未登录不显示 raw backend unauthorized error |
| D-15 | Auth verify recovery | `verifyAuthEmailChallenge` 成功即表示一次性 code 已被消费；后续 `/me` refresh 失败、超时或 profile context 加载错误不得被 UI 折叠成“验证码错误” | App 应调度 runtime auth refresh 并离开 `auth_verify`，在目标 route / auth route gate 上暴露 auth/profile loading 或 error；公共 auth route 的 initial `/me` skip 只能消费一次，直接提交 verified user 后的语言切换等 requestOptions 变化不得把已登录态重置为未登录 |

## 4 设计约束

- Route normalization 只能把旧 route 映射到当前保留 route，不允许旧 route 作为独立页面目标留存。
- `practice` 和 `generating` 可隐藏 TopBar；其他页面默认保留 App chrome。
- `pendingAction` 至少包含 `type`、`label`、`route`、`params`，登录成功后必须恢复 route context。
- `report` 必须携带 `sessionId` 或等价上下文；无上下文只能显示缺 session 状态，不能展示假报告。
- `practice?mode=voice&modality=voice` 是语音面试的显式入口；不得恢复独立 `voice` route alias。
- Browser History canonical URL 必须与现有 route catalog 对齐：primary nav、context routes、user-menu routes、auth routes 都从 `Route.name` 映射而来；未知 path / malformed params / retired aliases 只能归一到当前保留 route 或 `home`，不能 materialize 旧页面。
- Canonical URL 只允许携带稳定、可重取的上下文标识和 display hint，例如 `targetJobId`、`resumeVersionId`、`planId`、`sessionId`、`reportId`、`roundId`、`flow`、`tab`、`mode`、`modality`、`next`；不得携带原始正文、AI 输出、表单草稿或 auth secret。
- `auth_verify` 不得把验证码放入 URL、`pendingAction` 或 storage；组件只从受控 input 读取 6 位数字，并通过 generated `verifyAuthEmailChallenge` 的 query 参数提交给后端。
- `auth_profile_setup` 只能出现在已认证但 `profileCompletionRequired=true` 的账号路径；未登录访问该路由应回到 `auth_login`，资料已补全访问该路由应按 pendingAction / Home 恢复，不得让用户重复填写。
- 登录成功恢复业务 route 前必须先检查最新 `/me.profileCompletionRequired`；如果仍为 true，则替换到 `auth_profile_setup` 并携带原 pendingAction safe params。该规则不能依赖当前浏览器 session state，必须由后端 `/me` 返回值驱动。
- 公共 auth route 为避免预期 401 noise 可跳过首次 `/me` probe，但该 skip 必须在 provider 生命周期内只消费一次；`refreshAuth(user)` 直接提交 verified user 后，后续语言切换、requestOptions 变化或 client 变化必须执行真实 `/me` refresh，而不是重新把 auth state 置为 unauthenticated。
- `auth_verify` 的错误语义必须区分 code verification 与 post-verify profile context refresh：只有 `verifyAuthEmailChallenge` 本身失败才能显示 code verification failure；verify 成功后的 `/me` 失败必须作为可恢复的 auth/profile loading error 处理，不能让用户用已消费的一次性 code 反复重试。
- `#route=...` adapter 必须保留到 static preview / Playwright pixel parity / 场景 harness 全部迁移完成；adapter 解析出的 loose route 仍经过 `normalizeRoute`，并在正式 App 中替换为 canonical path，不新增 hash-only 分支。
- Auth `pendingAction` 与 URL 恢复必须共用同一组 safe route params：登录前可保存目标 route、稳定 ID 与轻量 display hint；登录成功后恢复原 canonical path；任何 raw payload 必须留在受控 runtime state 或重新从 API 获取。
- App route guard 必须先判断 runtime auth 状态再渲染面试业务 screen：未挂载 runtime 的 isolated UI tests / static prototype harness 可继续渲染 route shell；正式 runtime 中 `loading` 只显示 auth gate loading 占位，不得调用业务 API，`unauthenticated` 统一跳转到 `auth_login` 并携带 safe pendingAction。
- Home 是可未登录访问的默认入口，但 Recent mock interviews 是账号历史数据，只能在 `runtime.auth.status === "authenticated"` 时渲染和请求 `listTargetJobs`；未登录或 auth error 不得显示模块标题、空态、skeleton 或后端 `AUTH_UNAUTHORIZED` 原始错误。
- History `pushState` / `replaceState` / `popstate` 必须由 router adapter 统一管理，避免组件直接拼接 `window.location`；同一路由不同参数变化必须保持 back/forward 可预期。
- 全局显示控制对未登录用户可见，并保持在登录前后稳定。
- TopBar brand 只显示 `EasyInterview`，不得恢复 `面试训练器`、`Interview trainer` 或 `v1.0` 等解释性副标题；版本信息保留在 settings 的产品信息区，通过 locale 文案显示标签。
- 语言选择不是纯状态占位：TopBar、auth shell、settings shell 与当前 placeholder shell 等 D1 可见静态文案必须通过 typed i18n helper 渲染，选择 `zh` / `en` 后立即重绘；旧 `profile` route 只作为 retired alias 负向对象。
- i18n 资源必须按 locale 文件拆分：`zh`、`en` 等每种语言各自维护独立 locale 文件；helper / index 文件不得把多个语言的 message map 糅合在同一对象字面量中。新增语言时必须新增 locale 文件、类型校验和 focused test。
- TopBar 语言选择 UI 必须按 `ui-design/src/app.jsx` 源码复刻为带 globe icon 的 dropdown button；按钮只显示当前语言标签（如 `中文` / `English`），菜单项使用稳定 locale testid，方便后续新增语言；不得退化为 native `select`、分离按钮组或只切换状态的占位控件，也不得把多个候选语言拼在按钮上。
- 正式前端语言菜单不得在 TopBar 内维护私有二语言数组；必须从 `src/app/i18n/localeCatalog.ts` 的 `SUPPORTED_LOCALES` 渲染选项。`ui-design/src/app.jsx` 使用同等的 `LANGUAGE_OPTIONS` 原型元数据，二者在新增语言时同步。
- Locale 优先级为：用户显式选择 > 浏览器 locale > `en` fallback。用户显式选择必须保存到 `localStorage["ei-lang"]` 并优先于下一次打开时的浏览器 locale；`zh-CN` / `zh` 归一为 `zh`，`en-US` / `en` 归一为 `en`；未知、缺失或不支持的 BCP 47 tag fallback `en`。语言选择不依赖登录态，`/me.uiLanguage` 与 runtime `defaultUiLanguage` 不得覆盖前端显示偏好。
- Auth bootstrap 必须使用 B2 generated client 和 C1 passwordless session cookie 契约；不得在前端私造 Bearer token、密码登录 API、OAuth API 或自定义 session storage contract。
- Fixture-backed dev mock 不得只返回静态 `getMe.default` authenticated 响应。默认 dev preview 必须从非登录态开始；登录验证成功后仅在同一 mock session 中切换为 authenticated；退出登录成功后必须清除 mock session 并在 TopBar 回到单一登录入口。该状态只用于 dev / test mock runtime，不写入 production session storage，不替代真实 first-party cookie。
- `getRuntimeConfig` 必须先经过 A4 allowlist / generated type 解析，再影响 feature flag 或公开配置；缺失或错误响应必须有可测试 fallback。UI 语言不由 runtime config 决定。
- Generated API client 默认请求头或 App runtime request options 必须带当前 UI locale 的 `Accept-Language` display hint；该 header 不得覆盖业务字段（如 `targetLanguage` / practice language）。
- 新增 App shell / TopBar / auth / settings 组件时必须以 `docs/ui-design/` 与 `ui-design/` 源码为唯一 UI 真理源；正式前端目标是 100% 源级复刻静态原型中的 DOM 构图、布局、间距、字号、字体层级、控件密度、颜色、阴影、边框、圆角、状态、响应式行为和交互节奏。不得引入外部品牌设计系统作为参考替代，不得由 AI 自由重设计、重新解释或重新组合视觉。主题系统必须承接 `ui-design/` 的 warm / forest / ocean / plum 与 `customAccent`，自定义 accent 不是降级项；默认主题与无效值 fallback 均为 `ocean`，与 `ui-design/src/app.jsx` 的 `TWEAK_DEFAULTS.theme` 保持一致。

## 5 模块边界

| 边界 | Owner | 说明 |
|------|-------|------|
| frontend shell | `frontend-shell` | App bootstrap、router、TopBar、auth pages、pendingAction、display controls |
| auth/runtime client | `frontend-shell` + `backend-auth` + A4/B2 | generated client、runtime config bootstrap、passwordless auth operations、session-aware `/me` |
| mock data | `mock-contract-suite` | generated client mock transport 和 fixture-backed response |
| auth backend | `backend-auth` | passwordless challenge、session cookie、/me、logout |
| UI truth source | `docs/ui-design/` + `ui-design/` | 页面结构、目标路由和移除模块边界 |

## 6 验收标准

| ID | 场景 | Given | When | Then | 对应 Plan |
|----|------|-------|------|------|-----------|
| C-1 | 默认壳可用 | 用户未登录 | 打开 App | 渲染 Home、三入口 TopBar（`home` / `workspace` / `resume_versions`）、单一登录入口和显示控制，不出现 welcome、注册入口、`jd_match`、`debrief` 或 `profile` 可见入口；导航收敛实现证据由 `frontend-home-job-picks-and-parse/002` 与 product-scope/001 修订承接 | 001-app-shell-auth-settings |
| C-2 | Pending action 恢复 | 未登录用户在 workspace 点击立即面试 | 完成登录 | 跳回 practice 并保留 planId / targetJobId / jdId / resumeVersionId / roundId | 001-app-shell-auth-settings |
| C-3 | 用户菜单分流 | 用户已登录 | 打开用户菜单 | 只展示 settings / logout；`profile` 不作为 live route 或用户菜单项出现 | 001-app-shell-auth-settings |
| C-4 | 旧 route 不复活 | URL 或 localStorage 带旧 route | App normalize route | 映射到当前 route 或 Home，不产生独立旧页面 | 001-app-shell-auth-settings |
| C-5 | Runtime / session bootstrap | App 启动且 mock transport 可用 | 读取 runtime config 与 `/me` | 公开配置按 allowlist 生效，未登录返回认证态，已登录渲染用户区，不读取 prototype data | 001-app-shell-auth-settings |
| C-6 | Parse shell 可达 | 用户从 Home 或 Job Picks 进入 JD 解析确认 | App route 到 `parse` | 保留 App shell / route params，不把 JD 解析业务细节并入 D1 | 001-app-shell-auth-settings |
| C-7 | 中英 UI 切换 | 用户打开默认 App shell | 通过 TopBar language dropdown 选择 English / 中文 | TopBar、auth 入口和 D1 shell 静态文案即时切换；语言优先级为用户显式选择（`localStorage["ei-lang"]`）> 浏览器 locale > English fallback；登录态和 runtime locale 不覆盖前端语言设置；后续 API 请求带 `Accept-Language`；`zh` / `en` 文案分别来自独立 locale 文件；语言控件 DOM / 图标 / 当前语言标签 / 菜单项 / 文案节奏与 `ui-design/src/app.jsx` 一致 | 001-app-shell-auth-settings |
| C-8 | 视觉接入 100% 复刻 ui-design 真理源 | D1 已交付的 App 壳 / TopBar / 三入口 / 显示控制 / 认证页 / 用户菜单 / settings shell | D2 视觉系统接入 | 正式前端 100% 源级复刻 `ui-design/` 静态原型：DOM 构图、布局、间距、字号、字体层级、控件密度、颜色、阴影、边框、圆角、状态、响应式行为和交互节奏必须以对应 `ui-design/src/*.jsx` 与 `docs/ui-design/` 文档为准；4 基础主题（warm 完整对齐，其余主题至少色板正确）+ `customAccent` 在 light / dark 下均通过根级 `data-theme` / `data-mode` / `data-custom-accent` 或等价 CSS variable 切换生效；字体、token、className 与组件样式从 `ui-design/src/primitives.jsx`、`ui-design/src/app.jsx` 和对应 screen 原型抽取；`E2E.P0.005` visual smoke 工具必须对关键 viewport 完成非空渲染、无核心控件重叠、主题/暗色/custom accent 可见变化检查，并包含与 `ui-design` golden preview 的 DOM 锚点、computed style、bounding box 和必要截图差异 gate；任何可见偏差不得以”风格接近”收口，必须修到与原型一致或先修改 `ui-design/` 真理源；D1 testid 与 `E2E.P0.001` / `E2E.P0.002` / `E2E.P0.004` regression 全部通过 | 002-app-shell-visual-system |
| C-9 | 真实浏览器 pixel parity gate | D2 视觉系统已落地（`ei-shell-topbar` / `ei-screen-shell` / `ei-auth-shell` / fontsource / customAccent 全部接入），但 vitest+jsdom 不能验证 desktop / mobile viewport 下的 CSS 布局、bounding box 与截图差异 | 003 接入 Playwright + chromium 的 pixel parity gate | Playwright 在 desktop (1440×900) 与 mobile (390×844) 两个 viewport 下并行加载 `frontend/dist/index.html` 与 `ui-design/index.html` golden preview，断言：D2 testid / className / 文本内容在两边一致；默认主题 light 状态下 TopBar、auth、settings、placeholder、home、parse、workspace 的 `getBoundingClientRect()` 不重叠且 stays in viewport（历史断言中的 `jd_match` / `debrief` / `profile` 屏已随 product-scope D-17 / D-22 删除，相关 parity spec 由 `frontend-home-job-picks-and-parse/002` 与 product-scope/001 修订收口；默认主题自 v1.22 起为 `ocean`）；authenticated user menu 必须在 mocked login 后以头像 chip + dropdown 呈现，dropdown desktop 与 chip 右对齐、mobile 不溢出 viewport，且 logout 后回到非登录态；切换 dark / customAccent 后核心元素的 computed background / color 出现可见变化；workspace full-state pixel tests 必须通过 server-bound route params 进入完整规划态，不依赖 Home recent card 的 `resume-unbound` synthetic path；`E2E.P0.006` Playwright scenario `setup→trigger→verify→cleanup` 通过；`pnpm --filter @easyinterview/frontend test:pixel-parity` 默认在 CI / 本地都可运行（前提是 chromium 二进制已安装）且不得依赖 `.gitignore` 排除的本地 screenshot baseline；任何 pixel parity 失败必须修正到与 `ui-design/` 一致或先修订 `ui-design/` 真理源，不得以”差异在阈值内”收口；E2E.P0.005（jsdom 范围）保留作为 fast smoke gate | 003-ui-design-pixel-parity-gate |
| C-10 | 登录态菜单与退出闭环 | 用户打开 dev mock App | 默认看到单一登录入口；完成 passwordless mock 登录后打开头像菜单；点击退出登录并确认 | 用户菜单源级复刻 `ui-design/src/app.jsx` 的头像 chip、姓名/email header、settings/logout dropdown；browser-level parity gate 覆盖 desktop / mobile dropdown geometry 与 viewport-safe mobile placement；退出后 `/me` 为 unauthenticated，TopBar 回到单一登录入口，页面可继续浏览；`profile` 菜单项零残留 | 001-app-shell-auth-settings |
| C-11 | Canonical path deep link / reload / browser history | 用户拿到一个 canonical frontend URL，例如 `/workspace?targetJobId=...&resumeVersionId=...&planId=...`、`/practice?mode=voice&modality=voice&sessionId=...` 或 `/report?sessionId=...` | 直接打开、刷新页面、通过 App 导航进入下一页，再点击浏览器 back / forward | App 使用 Browser History 解析为同一个 `Route` + safe params；TopBar active / chrome hidden 行为与 route catalog 一致；InterviewContext 从 URL safe params hydrate；reload 不丢失稳定资源上下文；back/forward 不产生双重导航或旧页面 | 004-url-addressable-routing |
| C-12 | Auth pendingAction + URL privacy redline | 未登录用户从 URL-addressable 业务页触发登录，或打开带 safe params 的深链接并完成 Mailpit email-code 登录 | 登录前保存 pendingAction，或 `auth_verify` 手动提交 6 位验证码后登录成功 | pendingAction 只保存 route name、canonical URL 和 safe params；登录成功后回到原 canonical path；URL / pendingAction / localStorage / session storage / history / console 不含 JD 原文、简历原文、guided answers、parsed summary、suggestion、prompt、验证码或 auth secret query | 004-url-addressable-routing + 001-app-shell-auth-settings Phase 8 |
| C-13 | Hash compatibility + legacy route negative regression | static preview / pixel parity 仍使用 `#route=...`，或用户打开 unknown / malformed / retired route / 独立 `voice` URL | App bootstrap / normalize route | `#route=...` 被 adapter 解析后进入同一 `Route` 合约并替换/等价到 canonical path；unknown / malformed path fallback `home` 或当前保留 route；retired aliases 不 materialize standalone screens；`voice` 仍不是合法 route，只能用 `practice?mode=voice&modality=voice`；server fallback 对已知 frontend path 返回 `index.html`，API path 不被 frontend fallback 吞掉 | 004-url-addressable-routing |
| C-14 | 单入口登录与首次资料补全 | 用户使用新邮箱从单一登录入口完成 6 位验证码验证，但还没有 displayName / 条款确认 | 首次验证后关闭浏览器、换浏览器重新登录、退出后重新登录、刷新或直开带 pendingAction 的业务 URL | 每次登录后 `/me.profileCompletionRequired=true` 都先进入 `auth_profile_setup`；资料补全前不恢复业务动作；提交 displayName + 条款后 `/me.profileCompletionRequired=false`，再恢复 pendingAction 或 Home；TopBar 不显示注册入口，旧 `auth_register` 不能 materialize live page | 001-app-shell-auth-settings |
| C-15 | 未登录面试业务前置登录 | 用户未登录打开 Home 或直接访问任一面试业务 route | Home 首屏渲染、点击简历/工作台等业务入口，或直开 `workspace` / `practice` / `report` 等 URL | Home 不展示 Recent mock interviews 模块，不请求 `listTargetJobs`，不显示后端 raw unauthorized error；业务 route 在 auth loading 期间不挂载业务 screen，确认未登录后进入 `auth_login(pendingAction)`；登录成功后按 profile completion gate 恢复原 route；后端 session policy 仍证明除 auth start/verify/runtime-config 和 optional logout 外的业务 API 都要求 session | 001-app-shell-auth-settings |
| C-16 | Auth verify recovery and language switch stability | 用户从 `auth_verify` 成功提交 6 位 code；或在公共 auth route 首次 skip `/me` 后直接提交 verified user | 后续 `/me` refresh 失败，或用户切换 TopBar 语言导致 runtime requestOptions 变化 | 已消费 code 不被标记为验证码失败；App 离开 verify 页并进入 pending route / Home 的 auth-profile loading/error 表达；直接提交的 authenticated user 不会因语言切换重新落回 unauthenticated，后续变化会触发真实 `/me` refresh | 001-app-shell-auth-settings Phase 11 |
| C-17 | UX 漏斗收敛对齐（D-16 / D-21） | 2026-06-12 product-scope v2.1 与 `ui-design/` 原型已完成无密码唯一流与全局呈现收敛 | 正式前端对齐：直开 `/auth/reset` 或旧 `auth_reset` route key；打开登录页；打开设置页；首次打开 App 检查主题 | `auth_reset` 归一回 `auth_login`，`AuthResetScreen` 与"忘记密码"入口零残留，登录页展示静态帮助说明（一个邮箱一个账号 + 收不到验证码下一步可重发/换邮箱）；设置页只有 `个人资料` / `隐私与数据` 两个 tab，无通知/订阅占位 tab，个人资料 tab 含 `登录与安全` 仅展示 `邮箱验证码 · 无密码`；默认主题与无效值 fallback 为 `ocean`，主题菜单保留四预设 + `customAccent`；"密码 / 两步验证 / 忘记密码"口径在 frontend 源码与 i18n 零残留 | 001-app-shell-auth-settings Phase 12 |

## 7 关联计划

- [001-app-shell-auth-settings](./plans/001-app-shell-auth-settings/plan.md)
- [002-app-shell-visual-system](./plans/002-app-shell-visual-system/plan.md)
- [003-ui-design-pixel-parity-gate](./plans/003-ui-design-pixel-parity-gate/plan.md)
- [004-url-addressable-routing](./plans/004-url-addressable-routing/plan.md)

## 8 修订记录

| 版本 | 日期 | 说明 |
|------|------|------|
| 1.23 | 2026-07-06 | 对齐 product-scope D-22 与当前 `frontend/README.md` / `ui-design/src/app.jsx` / `frontend/src/app/routes.ts`：TopBar 一级导航收敛为 `home` / `workspace` / `resume_versions` 三项；旧 `debrief` 与 `profile` 不再作为 live route、用户菜单项、protected route 或 parity 正向对象，只作为 retired alias 归一回 `home` 的负向 gate。 |
| 1.22 | 2026-06-12 | 对齐 product-scope v2.1 UX 漏斗收敛：D-16 删除 `auth_reset`（归一回 `auth_login`，登录页静态帮助说明）；D-21 设置页收敛为个人资料/隐私与数据双 tab、`登录与安全` 仅展示无密码口径、默认主题与 fallback 改为 `ocean`；D-17 一级导航收敛为四项（实现证据由 `frontend-home-job-picks-and-parse/002` 承接）；D-18 删除 `company_intel` 独立 route（实现证据由 `frontend-workspace-and-practice/001` 承接）；新增 C-17 验收行，spec D-2/D-4/D-6 同步改写 |
