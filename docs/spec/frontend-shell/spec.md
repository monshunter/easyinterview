# Frontend Shell Spec

> **版本**: 1.6
> **状态**: active
> **更新日期**: 2026-05-07

## 1 背景与目标

`frontend-shell` 是 `engineering-roadmap` S1 的第一个用户可见前端 workstream。它负责把当前静态 UI 中已经收敛的 App 壳、TopBar、五个一级入口、全局显示控制、用户菜单、认证页面和 pending action 恢复模型落到正式前端工程中。

本 subject 的目标是让后续 D2-D6 前端模块在同一个 App shell 内开发，而不是各自创建路由、认证跳转和显示偏好状态。

## 2 范围

### 2.1 In Scope

- App 默认进入 `home`，不展示独立 welcome。
- TopBar 五个一级入口：`home`、`jd_match`、`workspace`、`resume_versions`、`debrief`。
- 上下文页面路由：`parse`、`practice`、`generating`、`report`、`company_intel`。其中 `parse` 只由本 subject 承接 route shell / chrome / params 边界，JD 解析业务内容归后续 `frontend-home-job-picks-and-parse`。
- 用户菜单入口：`profile`、`settings`、`auth_logout`。
- 认证页面：`auth_login`、`auth_register`、`auth_verify`、`auth_reset`、`auth_logout`。
- `requestAuth(pendingAction)` 与登录成功后的 route / params 恢复。
- 全局显示控制：主题色、暗色、语言切换；设置页维护字体预设。
- Runtime config、generated API client 与 fixture-backed mock transport bootstrap 的前端接入边界。

### 2.2 Out of Scope

- 不实现 D2-D6 业务页面细节：JD 导入、岗位推荐、模拟面试规划、练习 session、报告、简历工坊、复盘业务内容由后续 subject 承接。
- 不实现真实 passwordless 认证后端；后端能力归 `backend-auth`。
- 不新增旧 `welcome`、`growth`、`plan`、`mistakes`、`drill`、`followup`、`experiences`、`star`、独立 `voice` route。
- 不把 `ui-design/src/data.jsx` 作为运行时数据源。

## 3 用户决策 / 待确认事项

| ID | 决策 | 锁定值 | 影响 |
|----|------|--------|------|
| D-1 | 默认入口 | `home` | 未登录也能看到首页并开始输入 JD 草稿 |
| D-2 | 一级导航 | `home` / `jd_match` / `workspace` / `resume_versions` / `debrief` | 报告、语音、公司情报和认证不进入一级导航 |
| D-3 | Auth gate | 操作级 `requestAuth(pendingAction)` | 登录不是默认落地页；成功后恢复原动作 |
| D-4 | 显示偏好 | TopBar 持有主题、暗色、语言；settings 持有字体预设 | 登录状态不能重置显示偏好 |
| D-5 | 数据源 | 前端 shell 通过 generated client + fixture-backed mock transport / runtime config 取数 | 不直接 import prototype data |
| D-6 | Auth API 边界 | D1 前端只消费 `startAuthEmailChallenge`、`verifyAuthEmailChallenge`、`getMe`、`logout` 和 first-party session cookie | 密码、OAuth、reset 只能作为当前 UI 壳或 stub 展示；新增真实 API 前必须修订 C1 / B2 |
| D-7 | UI i18n | 正式前端至少支持 `zh` / `en` 两种 UI locale；每种语言必须有独立 locale 文件，聚合层只负责导入、类型约束和 helper；初始 UI 语言跟随浏览器 locale，未知或缺失时 fallback `en`；语言切换只关联前端显示偏好，不依赖 runtime config 或登录态 | 语言切换必须通过 TopBar 下拉框改变用户可见文案，并通过 `Accept-Language` 影响后续 API display hint |

## 4 设计约束

- Route normalization 只能把旧 route 映射到当前保留 route，不允许旧 route 作为独立页面目标留存。
- `practice` 和 `generating` 可隐藏 TopBar；其他页面默认保留 App chrome。
- `pendingAction` 至少包含 `type`、`label`、`route`、`params`，登录成功后必须恢复 route context。
- `report` 必须携带 `sessionId` 或等价上下文；无上下文只能显示缺 session 状态，不能展示假报告。
- `practice?mode=voice&modality=voice` 是语音面试的显式入口；不得恢复独立 `voice` route alias。
- 全局显示控制对未登录用户可见，并保持在登录前后稳定。
- 语言切换不是纯状态占位：TopBar、auth shell、profile/settings shell 等 D1 可见静态文案必须通过 typed i18n helper 渲染，切换 `zh` / `en` 后立即重绘。
- i18n 资源必须按 locale 文件拆分：`zh`、`en` 等每种语言各自维护独立 locale 文件；helper / index 文件不得把多个语言的 message map 糅合在同一对象字面量中。新增语言时必须新增 locale 文件、类型校验和 focused test。
- TopBar 语言切换 UI 必须是下拉框（native `select` 或等价可访问 combobox），不得退化为两个按钮或只切换状态的占位控件。
- Locale 优先级为：用户显式切换 > 浏览器 locale > `en` fallback。`zh-CN` / `zh` 归一为 `zh`，`en-US` / `en` 归一为 `en`；未知、缺失或不支持的 BCP 47 tag fallback `en`。语言切换不依赖登录态，`/me.uiLanguage` 与 runtime `defaultUiLanguage` 不得覆盖前端显示偏好。
- Auth bootstrap 必须使用 B2 generated client 和 C1 passwordless session cookie 契约；不得在前端私造 Bearer token、密码登录 API、OAuth API 或自定义 session storage contract。
- `getRuntimeConfig` 必须先经过 A4 allowlist / generated type 解析，再影响 feature flag 或公开配置；缺失或错误响应必须有可测试 fallback。UI 语言不由 runtime config 决定。
- Generated API client 默认请求头或 App runtime request options 必须带当前 UI locale 的 `Accept-Language` display hint；该 header 不得覆盖业务字段（如 `targetLanguage` / practice language）。
- 新增 App shell / TopBar / auth / settings 组件时必须参考根目录 `DESIGN.md` 的语义命名和节奏，但 `docs/ui-design/` 与 `ui-design/` 是验收真理源头；不得机械同步 token 或引入私有授权字体。主题系统必须承接 `ui-design/` 的 warm / forest / ocean / plum 与 `customAccent`，自定义 accent 不是降级项。

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
| C-1 | 默认壳可用 | 用户未登录 | 打开 App | 渲染 Home、五入口 TopBar、登录/注册和显示控制，不出现 welcome | 001-app-shell-auth-settings |
| C-2 | Pending action 恢复 | 未登录用户在 workspace 点击立即面试 | 完成登录 | 跳回 practice 并保留 planId / targetJobId / jdId / resumeVersionId / roundId | 001-app-shell-auth-settings |
| C-3 | 用户菜单分流 | 用户已登录 | 打开用户菜单 | `用户画像` 与 `设置与隐私` 分别进入 profile / settings | 001-app-shell-auth-settings |
| C-4 | 旧 route 不复活 | URL 或 localStorage 带旧 route | App normalize route | 映射到当前 route 或 Home，不产生独立旧页面 | 001-app-shell-auth-settings |
| C-5 | Runtime / session bootstrap | App 启动且 mock transport 可用 | 读取 runtime config 与 `/me` | 公开配置按 allowlist 生效，未登录返回认证态，已登录渲染用户区，不读取 prototype data | 001-app-shell-auth-settings |
| C-6 | Parse shell 可达 | 用户从 Home 或 Job Picks 进入 JD 解析确认 | App route 到 `parse` | 保留 App shell / route params，不把 JD 解析业务细节并入 D1 | 001-app-shell-auth-settings |
| C-7 | 中英 UI 切换 | 用户打开默认 App shell | 在 TopBar 下拉框切换语言到 English / 中文 | TopBar、auth 入口和 D1 shell 静态文案即时切换；初始语言跟随浏览器 locale，不支持时 fallback English；登录态和 runtime locale 不覆盖前端语言设置；后续 API 请求带 `Accept-Language`；`zh` / `en` 文案分别来自独立 locale 文件 | 001-app-shell-auth-settings |
| C-8 | 视觉接入对齐 ui-design 真理源 | D1 已交付的 App 壳 / TopBar / 五入口 / 显示控制 / 认证页 / 用户菜单 / settings & profile placeholder | D2 视觉系统接入 | 渲染语义对齐 `ui-design/` 静态原型与根目录 `DESIGN.md` 节奏；4 基础主题（warm 完整对齐，其余主题至少色板正确）+ `customAccent` 在 light / dark 下均通过根级 `data-theme` / `data-mode` / `data-custom-accent` 或等价 CSS variable 切换生效；字体使用开源 serif (Cormorant Garamond / EB Garamond) 与 sans (Inter)，不引入私有授权字体；`EI_THEMES` 不被 `DESIGN.md` hex 机械覆盖；`E2E.P0.005` visual smoke 工具对关键 viewport 完成非空渲染、无核心控件重叠、主题/暗色/custom accent 可见变化检查；D1 testid 与 `E2E.P0.001` / `E2E.P0.002` / `E2E.P0.004` regression 全部通过 | 002-app-shell-visual-system |

## 7 关联计划

- [001-app-shell-auth-settings](./plans/001-app-shell-auth-settings/plan.md)
- [002-app-shell-visual-system](./plans/002-app-shell-visual-system/plan.md)
