# Frontend Shell Spec

> **版本**: 1.34
> **状态**: active
> **更新日期**: 2026-07-15

## 1 背景与目标

`frontend-shell` 是正式前端的 App 壳 owner。它负责依据当前产品 spec 与 `docs/ui-design/` 中的 UI 架构、流程和交互约束，把 shell、TopBar、display preferences、auth pages、settings、route normalization、runtime bootstrap 和 protected route guard 落到 `frontend/`。

目标是让业务页面 owner 复用同一个 route/auth/display 基座，而不是各自实现导航、登录恢复或显示偏好状态。

## 2 范围

### 2.1 In Scope

- 默认入口：`home`。
- 一级 TopBar 入口：`home`、`workspace`、`resume_versions`。
- 上下文 route：`parse`、`practice`、`reports`、`generating`、`report`。
- 账号入口 route：已登录 TopBar 设置齿轮直达 `settings`；`auth_logout` 只由设置页发起。
- Auth route：`auth_login`、`auth_verify`、`auth_profile_setup`、`auth_logout`。
- Settings：无 tab 的账号与隐私单页；复用 runtime `/me` 展示真实 `displayName/email`（完整账号邮箱），提供既有退出登录入口、导出暂不可用状态和账号删除确认流程。
- `requestAuth(pendingAction)`：未登录用户触发受保护动作时进入登录页，登录和资料补全完成后恢复 safe route params。
- Email-code auth：`auth_verify` 承接 6 位验证码输入，通过 generated `verifyAuthEmailChallenge` 完成验证。
- Runtime bootstrap：`getRuntimeConfig`、`getMe`、generated client、fixture-backed mock transport and dev mock session state。
- GET request orchestration：React StrictMode 保持开启；同一逻辑 GET 的同时在途调用共享一个底层 request，并在 settle 后立即驱逐，不成为数据缓存。
- URL-addressable routing：Browser History canonical path + query，支持直开、刷新、复制链接和 back/forward。
- Protected route guard：业务 route 只在 runtime auth 明确 authenticated 后挂载 screen 和调用受保护 API。
- Display preferences：主题、暗色和语言下拉；默认主题和无效值 fallback 为 `ocean`；当前 TopBar 只提供 `ocean` / `plum` 两个预定义主题和只含色相、饱和度控件的 custom accent。字体使用固定产品栈，不再是 preference。

### 2.2 Out of Scope

- JD 导入、模拟面试规划、练习 session、报告正文和简历工坊业务内容。
- Backend auth implementation；后端能力由 `backend-auth` owning。
- 扩大 route catalog 或新增当前范围外的可见入口。
- 创建与正式前端重复的可运行 UI Demo、prototype fixture 或双源 parity 流程。
- 在 URL、`pendingAction`、storage 或 browser history 中保存 JD 原文、简历原文、答案正文、解析结果、AI prompt/response、验证码或 session secret。

## 3 用户决策

| ID | 决策 | 锁定值 | 影响 |
|----|------|--------|------|
| D-1 | 默认入口 | `home` | 未登录用户也能看到首页和输入入口 |
| D-2 | 一级导航 | `home` / `workspace` / `resume_versions` | TopBar 一级导航保持三入口 |
| D-3 | Auth gate | 操作级 `requestAuth(pendingAction)` | 登录成功后恢复原业务动作和 safe params |
| D-4 | Auth API | `startAuthEmailChallenge`、`verifyAuthEmailChallenge`、`getMe`、`completeMyProfile`、`logout` + first-party session cookie | 前端不创建自定义 session contract |
| D-5 | Profile setup gate | `/me.profileCompletionRequired=true` 强制进入 `auth_profile_setup` | 资料补全前不恢复业务 route |
| D-6 | Display preferences | 前端持有语言、主题和暗色；字体固定为产品默认栈 | 登录态和 runtime config 不覆盖用户显式选择；不保留 font preset 状态或兼容键 |
| D-7 | Canonical URL | Browser History path + safe query | URL 表达页面和稳定上下文，不表达后端 action 或敏感正文 |
| D-8 | UI 设计 owner | `docs/ui-design/` | UI 架构、页面流程、交互约束和设计决策先在文档中收敛；正式 `frontend/` 直接实施，不维护可运行 Demo 或双源 parity |
| D-9 | 主题色范围 | 预定义主题只保留 `ocean` / `plum`，另保留 custom accent | 移除 `warm` / `forest` active palette、TopBar option 和 locale 文案 |
| D-10 | 规划范围报告路由 | `/reports?targetJobId=<uuid>` 是受保护的上下文 route，chrome visible，但不属于 TopBar 一级导航 | safe params 只允许 `targetJobId`；全局一级导航仍严格保持三入口 |
| D-11 | Safe-read single-flight | 保留 React StrictMode；只合并同一 client、method/path/query、规范化相关 header、normalized `okStatuses`、read/auth epoch 与 auth scope 下、且没有 caller `AbortSignal` 的语义只读在途 GET | resolve/reject 后驱逐；不同 client/query/header/okStatuses/epoch/auth 不合并；带 signal、非 GET 与语义写入 GET（含 `/auth/email/verify`）永不合并；所有语义写请求在 dispatch 前与 settle 后都推进 read epoch |
| D-12 | 规划 route 分工 | `/parse?targetJobId` 只承载刚导入规划的 queued/processing 命令进度；`/workspace` 展示列表，`/workspace?targetJobId` 展示只读详情 | ready 初读或轮询转 ready 必须 replace 到 workspace 详情；已解析卡片不再经过 Parse 动画 |
| D-13 | Custom accent 最小控件 | `CustomAccentPicker` 只保留色相与饱和度；选择 Ocean / Plum 是退出自定义色的唯一清晰路径 | 删除 preview/value 区、恢复默认色按钮与 `onClear` / `active` 冗余 props，不增加第二套 reset 语义 |
| D-14 | 设置简化方案 A | 已登录 TopBar 只保留设置齿轮；Settings 为无 tab 的真实账号/隐私单页，退出迁入页面；删除账号 chip/dropdown、静态登录安全、字体预设、产品信息和无后端事实字段 | `getMe` 复用 runtime context；导出按 501 诚实禁用；`deleteMe` 覆盖确认/pending/failure/202 success；无兼容 UI 或重复请求 |

## 4 设计约束

- Route normalization 只能把 unsupported route input 映射到当前 route catalog 或 `home`。
- `practice` 和 `generating` 可以隐藏 TopBar；其他 route 默认保留 App chrome。
- `reports` 保留 App chrome 但不得加入 `PRIMARY_NAV_ROUTES` / TopBar；直开、刷新、back/forward 和 auth continuation 只保留合法 `targetJobId`。
- `parse` 不接受 `section=reports`；`report` / `generating` 的资源 locator 只接受 `reportId`。报告状态、target/round/resume 等业务事实必须由受保护 API response 提供，不能由 query/pendingAction 注入。
- `/workspace` 只允许可选 `targetJobId`；`planId`、`resumeId`、auto-start 和其他业务状态必须剔除。`/parse` 只允许 `targetJobId` 并作为 command/progress route；ready 后使用 replace 导航到 `/workspace?targetJobId`。
- `pendingAction` 只包含 route name、canonical URL 和 safe params，例如 `targetJobId`、`resumeId`、`planId`、`sessionId`、`reportId`、`roundId`、`flow`、`tab`、`mode`、`modality`、`next`。
- 登录成功恢复 route 前必须检查最新 `/me.profileCompletionRequired`；仍为 true 时进入 `auth_profile_setup` 并保留 safe pendingAction。
- `auth_verify` 只从受控 input 读取 6 位验证码；验证码不得进入 URL、pendingAction、storage 或 browser navigation chain。
- `auth_verify` 的错误语义必须区分 code verification 与 post-verify profile context refresh；verify 成功后的 `/me` failure 由 route gate 表达，不渲染为验证码错误。
- 公共 auth route 可以跳过首次 `/me` probe，但 skip 在 provider lifecycle 内只能消费一次；`refreshAuth(user)` 后的 request options 变化必须执行真实 `/me` refresh。
- Home 可未登录访问；账号记录数据只在 authenticated 状态请求和渲染。
- Safe-read single-flight key 必须包含 client identity、HTTP method、path、canonical query、规范化的相关 request headers、normalized `okStatuses`、read/auth epoch 和 auth/session scope。只在 Promise 未 settle 时复用；resolve/reject 都删除 registry entry。caller `AbortSignal`、非 GET 与语义写入 GET 绕过合并，避免共享取消所有权或改变写请求语义。每个语义写请求必须在 dispatch 前推进一次 read epoch，并在 resolve/reject settle 后再次推进，确保 mutation 期间与 mutation 后的读请求都不能复用 mutation 前的 in-flight。`/auth/email/verify` 虽使用 GET wire method，但会消费 challenge/更新 session，必须按语义写请求 bypass；成功后还要推进 auth/session epoch，使后续 `/me` 与业务读取不复用认证前 scope。
- `AppRuntimeProvider`、Home / `useRecentTargetJobs`、Parse、`useWorkspaceTargetJobs`、Reports 和 Practice 等 screen loader effect 只依赖稳定 client/auth/request-option/route identity 输入，不依赖每次 render 都变化的整体 runtime object；locale、auth scope 或 epoch 变化仍必须产生新的 request key 和真实 refresh。
- Demo-only `#route=...` adapter 不属于正式 route contract；真实开发和场景验证使用 canonical Browser History URL。
- TopBar language dropdown 从 locale catalog 渲染；locale priority 为用户显式选择 > browser locale > `en` fallback，并通过 `Accept-Language` 作为 display hint。
- TopBar 已登录账号区只渲染设置齿轮，必须具备本地化 accessible name、focus ring 与不小于 40×40px 的点击区域；头像、姓名、caret、backdrop、dropdown 与 TopBar logout 不属于当前 DOM/CSS/i18n 合同。
- Settings 只消费 `AppRuntimeContext.auth.user` 与 generated `deleteMe`；页面挂载不得再次调用 `getMe`。账号删除一次确认生命周期内复用同一 idempotency key，失败可重试；`202` 后调用现有 `refreshAuth()` 重探测 `/me`（预期 401），提交 unauthenticated 状态并 replace Home，不新增清 session 方法。默认 fixture-backed dev client 必须复现相同 auth transition：成功 `deleteMe` 后下一次无 `Prefer` 的 `getMe` 返回 401，不得继续投影 authenticated 用户。
- `E2E.P0.101` 的完整账号邮箱只用于真实页面/API 断言；PASS 与 FAIL reporter、`trigger.log` 和 result artifact 均不得包含原文或 URL percent-encoded 等价值。失败断言不得把完整邮箱作为 matcher expected/received 文本直接交给 reporter，场景落盘前还必须执行流式脱敏作为纵深防御。
- UI implementation 必须符合对应产品 spec 与 `docs/ui-design/` 的信息架构、流程、交互和响应式约束；具体实现由正式组件、token、可访问性、component/browser tests 与真实业务场景验证，不要求按设计合同实现或像素对照。

## 5 模块边界

| 边界 | Owner | 说明 |
|------|-------|------|
| frontend shell | `frontend-shell` | App bootstrap、router、TopBar、auth pages、pendingAction、settings、display controls |
| auth/runtime client | `frontend-shell` + `backend-auth` + A4/B2 | generated client、runtime config、auth operations、session-aware `/me` |
| mock data | `mock-contract-suite` | generated client mock transport、fixture-backed responses、dev mock session state |
| auth backend | `backend-auth` | email-code challenge、session cookie、/me、logout |
| UI design docs | `docs/ui-design/` | 信息架构、页面流程、交互约束、响应式与设计决策 |

## 6 验收标准

| ID | 场景 | Given | When | Then | 对应 Plan |
|----|------|-------|------|------|-----------|
| C-1 | 默认壳可用 | 用户未登录 | 打开 App | 渲染 Home、三入口 TopBar、单一登录入口和显示控制 | 001-app-shell-auth-settings |
| C-2 | Pending action 恢复 | 未登录用户触发受保护动作 | 完成 email-code 登录和资料补全 gate | 恢复目标 route，并保留 safe params | 001-app-shell-auth-settings |
| C-3 | Settings 单入口 | 已登录用户点击 TopBar 设置齿轮 | 进入 settings，查看账号或选择退出 | TopBar 无账号菜单；Settings 复用真实 runtime 用户并由页面进入 `auth_logout` | 001-app-shell-auth-settings |
| C-4 | Unsupported route fallback | URL / hash / localStorage 带 unsupported route input | App normalize route | 映射到当前 route catalog 或 Home，不产生独立页面 | 001-app-shell-auth-settings / 004-url-addressable-routing |
| C-5 | Runtime bootstrap | App 启动且 mock transport 可用 | 读取 runtime config 与 `/me` | 公开配置按 allowlist 生效，auth state 驱动用户区和 route guard | 001-app-shell-auth-settings |
| C-6 | Display preferences | 用户切换语言、主题或暗色 | 刷新、登录、退出或 `/me` refresh | 前端显式选择优先，generated client 请求带当前 UI locale display hint；字体保持固定产品栈 | 001-app-shell-auth-settings / 002-app-shell-visual-system |
| C-7 | Protected route guard | 用户未登录并打开业务 route | runtime auth loading / unauthenticated | 不挂载业务 screen，不调用受保护 API，进入 `auth_login(pendingAction)` | 001-app-shell-auth-settings |
| C-8 | Email-code profile setup | 新邮箱完成验证码验证 | `/me.profileCompletionRequired=true` | 先进入 `auth_profile_setup`，资料补全成功后再恢复 pendingAction 或 Home | 001-app-shell-auth-settings |
| C-9 | Canonical URL | 用户打开、刷新或复制 frontend URL | Browser History parse / back / forward | Route、safe params、chrome behavior 和 auth gate 保持一致 | 004-url-addressable-routing |
| C-10 | UI 设计一致性 | Shell / TopBar / Auth / Settings 可见 UI 变更 | 对照 `docs/ui-design/` 实施并运行 component、responsive、a11y 与必要 browser smoke | 正式前端满足当前架构、流程和交互约束，不依赖第二套 Demo | 002-app-shell-visual-system / 003-ui-design-responsive-browser-gate |
| C-11 | Reports deep link | 用户直开/刷新 `/reports?targetJobId=<uuid>`，或未登录后完成鉴权 | route normalize / history / pendingAction restore | 仅合法 targetJobId 被保留并进入受保护 ReportsScreen；缺失/非法 target 以 replace-only 回 workspace 且不形成 Back 循环；chrome visible、TopBar 无报告入口；旧 `section` 与 report/status/round 等 query 被剔除 | 004-url-addressable-routing |
| C-12 | StrictMode safe-read 去重 | AppRuntimeProvider 或 Home/Parse/Workspace/Reports/Practice loader 在 StrictMode mount cycle 内发出同 key safe-read GET | 两个 caller 同时等待、settle 后重试、使用不同 `okStatuses`，或在任一语义写请求前/期间/settle 后读取 | 同时在途只产生一次底层 GET；settle 后重试产生新 GET；不同 client/query/header/okStatuses/epoch/auth、带 signal、非 GET 与 verify GET 均不合并；所有语义写请求 dispatch 前和 settle 后推进 read epoch，verify 成功另推进 auth/session epoch并真实刷新 | 001-app-shell-auth-settings |
| C-13 | Parse/workspace route 分工 | TargetJob 为 queued/processing 或 ready | 打开 `/parse?targetJobId`、轮询转 ready、或打开 ready 卡片 | Parse 只在处理中展示进度；ready 使用 replace 进入 `/workspace?targetJobId`；无 target 的 workspace 仍为列表，详情不显示 Parse 动画 | 004-url-addressable-routing |
| C-14 | Custom accent 最小选择器 | TopBar 主题菜单打开 | 用户调整自定义色或选择 Ocean / Plum | 只显示色相、饱和度；不显示 preview/value、恢复主题默认色按钮；选择预定义主题清晰退出 custom accent | 002-app-shell-visual-system |
| C-15 | Settings 真实数据与隐私动作 | authenticated runtime 已取得 `/me` | 打开设置、查看导出状态、退出或删除账号 | 只显示真实 `displayName/email`，其中 email 完整显示但不进入 PASS/FAIL 日志或证据；不重复 `getMe`；导出显示暂不可用；删除流程具备确认/pending/failure/202 success；默认 fixture client 在删除后也返回 unauthenticated，且旧 tab/block/字段零引用 | 001-app-shell-auth-settings / 002-app-shell-visual-system |

## 7 关联计划

- [001-app-shell-auth-settings](./plans/001-app-shell-auth-settings/plan.md)
- [002-app-shell-visual-system](./plans/002-app-shell-visual-system/plan.md)
- [003-ui-design-responsive-browser-gate](./plans/003-ui-design-responsive-browser-gate/plan.md)
- [004-url-addressable-routing](./plans/004-url-addressable-routing/plan.md)

## 8 修订记录

| 版本 | 日期 | 说明 |
|------|------|------|
| 1.34 | 2026-07-15 | 补齐 settings review remediation：fixture-backed deleteMe 后必须转 unauthenticated，并要求 P0.101 的失败 reporter 与落盘证据同样脱敏。 |
| 1.32 | 2026-07-15 | 采用设置简化方案 A：TopBar 已登录态收敛为设置齿轮，Settings 改为真实账号/隐私单页，删除字体预设与静态冗余字段，并接入 logout/deleteMe 合同。 |
| 1.31 | 2026-07-15 | 删除 UI Demo 与可运行原型权威来源合同；保留 `docs/ui-design/` 作为 UI 架构、流程、交互约束和设计决策 owner，正式前端直接实施和验证。 |
| 1.30 | 2026-07-14 | Add normalized `okStatuses` to safe-read identity and require every semantic mutation to advance read epoch before dispatch and after settle. |
| 1.29 | 2026-07-14 | Add StrictMode-safe GET single-flight, command-only Parse versus query-addressed Workspace detail, and the minimal hue/saturation custom-accent picker. |
| 1.28 | 2026-07-14 | Add protected target-scoped `/reports` with targetJobId-only routing, no TopBar entry, no Parse section compatibility, and reportId-only detail/generating locators. |
| 1.27 | 2026-07-09 | 收敛 TopBar 主题色范围为 deep ocean / plum / custom accent，移除 warm / forest active UI 合同。 |
| 1.26 | 2026-07-07 | 压缩 active spec 为当前 App shell、email-code auth、settings、display、URL 和 route-guard 合同。 |
