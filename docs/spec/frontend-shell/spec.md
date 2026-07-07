# Frontend Shell Spec

> **版本**: 1.26
> **状态**: active
> **更新日期**: 2026-07-07

## 1 背景与目标

`frontend-shell` 是正式前端的 App 壳 owner。它负责把当前 UI truth source 中的 shell、TopBar、display preferences、auth pages、settings、route normalization、runtime bootstrap 和 protected route guard 落到 `frontend/`。

目标是让业务页面 owner 复用同一个 route/auth/display 基座，而不是各自实现导航、登录恢复或显示偏好状态。

## 2 范围

### 2.1 In Scope

- 默认入口：`home`。
- 一级 TopBar 入口：`home`、`workspace`、`resume_versions`。
- 上下文 route：`parse`、`practice`、`generating`、`report`。
- 用户菜单 route：`settings`、`auth_logout`。
- Auth route：`auth_login`、`auth_verify`、`auth_profile_setup`、`auth_logout`。
- Settings：`个人资料` / `隐私与数据` 双 tab；个人资料 tab 承接账号基础信息、登录与安全展示、字体预设和产品信息。
- `requestAuth(pendingAction)`：未登录用户触发受保护动作时进入登录页，登录和资料补全完成后恢复 safe route params。
- Email-code auth：`auth_verify` 承接 6 位验证码输入，通过 generated `verifyAuthEmailChallenge` 完成验证。
- Runtime bootstrap：`getRuntimeConfig`、`getMe`、generated client、fixture-backed mock transport and dev mock session state。
- URL-addressable routing：Browser History canonical path + query，支持直开、刷新、复制链接和 back/forward。
- Protected route guard：业务 route 只在 runtime auth 明确 authenticated 后挂载 screen 和调用受保护 API。
- Display preferences：主题、暗色、语言下拉、字体预设；默认主题和无效值 fallback 为 `ocean`。

### 2.2 Out of Scope

- JD 导入、模拟面试规划、练习 session、报告正文和简历工坊业务内容。
- Backend auth implementation；后端能力由 `backend-auth` owning。
- 扩大 route catalog 或新增当前范围外的可见入口。
- 把 `ui-design/src/data.jsx` 作为正式运行时数据源。
- 在 URL、`pendingAction`、storage 或 browser history 中保存 JD 原文、简历原文、答案正文、解析结果、AI prompt/response、验证码或 session secret。

## 3 用户决策

| ID | 决策 | 锁定值 | 影响 |
|----|------|--------|------|
| D-1 | 默认入口 | `home` | 未登录用户也能看到首页和输入入口 |
| D-2 | 一级导航 | `home` / `workspace` / `resume_versions` | TopBar 一级导航保持三入口 |
| D-3 | Auth gate | 操作级 `requestAuth(pendingAction)` | 登录成功后恢复原业务动作和 safe params |
| D-4 | Auth API | `startAuthEmailChallenge`、`verifyAuthEmailChallenge`、`getMe`、`completeMyProfile`、`logout` + first-party session cookie | 前端不创建自定义 session contract |
| D-5 | Profile setup gate | `/me.profileCompletionRequired=true` 强制进入 `auth_profile_setup` | 资料补全前不恢复业务 route |
| D-6 | Display preferences | 前端持有语言、主题、暗色和字体预设 | 登录态和 runtime config 不覆盖用户显式选择 |
| D-7 | Canonical URL | Browser History path + safe query | URL 表达页面和稳定上下文，不表达后端 action 或敏感正文 |
| D-8 | UI truth source | `docs/ui-design/` + `ui-design/` | 可见 UI 变更先更新静态原型，再迁移正式前端 |

## 4 设计约束

- Route normalization 只能把 unsupported route input 映射到当前 route catalog 或 `home`。
- `practice` 和 `generating` 可以隐藏 TopBar；其他 route 默认保留 App chrome。
- `pendingAction` 只包含 route name、canonical URL 和 safe params，例如 `targetJobId`、`resumeId`、`planId`、`sessionId`、`reportId`、`roundId`、`flow`、`tab`、`mode`、`modality`、`next`。
- 登录成功恢复 route 前必须检查最新 `/me.profileCompletionRequired`；仍为 true 时进入 `auth_profile_setup` 并保留 safe pendingAction。
- `auth_verify` 只从受控 input 读取 6 位验证码；验证码不得进入 URL、pendingAction、storage 或 browser navigation chain。
- `auth_verify` 的错误语义必须区分 code verification 与 post-verify profile context refresh；verify 成功后的 `/me` failure 由 route gate 表达，不渲染为验证码错误。
- 公共 auth route 可以跳过首次 `/me` probe，但 skip 在 provider lifecycle 内只能消费一次；`refreshAuth(user)` 后的 request options 变化必须执行真实 `/me` refresh。
- Home 可未登录访问；账号记录数据只在 authenticated 状态请求和渲染。
- `#route=...` adapter 仅服务 static preview / pixel parity / scenario harness 输入，解析后仍进入同一 route normalization and canonical URL layer。
- TopBar language dropdown 从 locale catalog 渲染；locale priority 为用户显式选择 > browser locale > `en` fallback，并通过 `Accept-Language` 作为 display hint。
- UI implementation 必须源级追溯到 `docs/ui-design/` 与 `ui-design/`：DOM 构图、布局、间距、字号、控件密度、颜色、阴影、边框、圆角、状态、响应式行为和交互节奏都以原型为准。

## 5 模块边界

| 边界 | Owner | 说明 |
|------|-------|------|
| frontend shell | `frontend-shell` | App bootstrap、router、TopBar、auth pages、pendingAction、settings、display controls |
| auth/runtime client | `frontend-shell` + `backend-auth` + A4/B2 | generated client、runtime config、auth operations、session-aware `/me` |
| mock data | `mock-contract-suite` | generated client mock transport、fixture-backed responses、dev mock session state |
| auth backend | `backend-auth` | email-code challenge、session cookie、/me、logout |
| UI truth source | `docs/ui-design/` + `ui-design/` | route catalog、screen structure、visual parity source |

## 6 验收标准

| ID | 场景 | Given | When | Then | 对应 Plan |
|----|------|-------|------|------|-----------|
| C-1 | 默认壳可用 | 用户未登录 | 打开 App | 渲染 Home、三入口 TopBar、单一登录入口、用户区和显示控制 | 001-app-shell-auth-settings |
| C-2 | Pending action 恢复 | 未登录用户触发受保护动作 | 完成 email-code 登录和资料补全 gate | 恢复目标 route，并保留 safe params | 001-app-shell-auth-settings |
| C-3 | Settings 分流 | 已登录用户打开用户菜单 | 进入 settings 或 logout | 用户菜单只分流账号设置与退出登录 | 001-app-shell-auth-settings |
| C-4 | Unsupported route fallback | URL / hash / localStorage 带 unsupported route input | App normalize route | 映射到当前 route catalog 或 Home，不产生独立页面 | 001-app-shell-auth-settings / 004-url-addressable-routing |
| C-5 | Runtime bootstrap | App 启动且 mock transport 可用 | 读取 runtime config 与 `/me` | 公开配置按 allowlist 生效，auth state 驱动用户区和 route guard | 001-app-shell-auth-settings |
| C-6 | Display preferences | 用户切换语言、主题、暗色或字体预设 | 刷新、登录、退出或 `/me` refresh | 前端显式选择优先，generated client 请求带当前 UI locale display hint | 001-app-shell-auth-settings / 002-app-shell-visual-system |
| C-7 | Protected route guard | 用户未登录并打开业务 route | runtime auth loading / unauthenticated | 不挂载业务 screen，不调用受保护 API，进入 `auth_login(pendingAction)` | 001-app-shell-auth-settings |
| C-8 | Email-code profile setup | 新邮箱完成验证码验证 | `/me.profileCompletionRequired=true` | 先进入 `auth_profile_setup`，资料补全成功后再恢复 pendingAction 或 Home | 001-app-shell-auth-settings |
| C-9 | Canonical URL | 用户打开、刷新或复制 frontend URL | Browser History parse / back / forward | Route、safe params、chrome behavior 和 auth gate 保持一致 | 004-url-addressable-routing |
| C-10 | UI parity | Shell / TopBar / Auth / Settings 可见 UI 变更 | 运行 visual smoke / pixel parity owner gates | 正式前端与 `ui-design/` 源码结构和关键 computed style 对齐 | 002-app-shell-visual-system / 003-ui-design-pixel-parity-gate |

## 7 关联计划

- [001-app-shell-auth-settings](./plans/001-app-shell-auth-settings/plan.md)
- [002-app-shell-visual-system](./plans/002-app-shell-visual-system/plan.md)
- [003-ui-design-pixel-parity-gate](./plans/003-ui-design-pixel-parity-gate/plan.md)
- [004-url-addressable-routing](./plans/004-url-addressable-routing/plan.md)

## 8 修订记录

| 版本 | 日期 | 说明 |
|------|------|------|
| 1.26 | 2026-07-07 | 压缩 active spec 为当前 App shell、email-code auth、settings、display、URL 和 route-guard 合同。 |
