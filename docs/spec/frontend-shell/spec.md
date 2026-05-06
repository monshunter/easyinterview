# Frontend Shell Spec

> **版本**: 1.1
> **状态**: active
> **更新日期**: 2026-05-06

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

## 4 设计约束

- Route normalization 只能把旧 route 映射到当前保留 route，不允许旧 route 作为独立页面目标留存。
- `practice` 和 `generating` 可隐藏 TopBar；其他页面默认保留 App chrome。
- `pendingAction` 至少包含 `type`、`label`、`route`、`params`，登录成功后必须恢复 route context。
- `report` 必须携带 `sessionId` 或等价上下文；无上下文只能显示缺 session 状态，不能展示假报告。
- `practice?mode=voice&modality=voice` 是语音面试的显式入口；不得恢复独立 `voice` route alias。
- 全局显示控制对未登录用户可见，并保持在登录前后稳定。
- Auth bootstrap 必须使用 B2 generated client 和 C1 passwordless session cookie 契约；不得在前端私造 Bearer token、密码登录 API、OAuth API 或自定义 session storage contract。
- `getRuntimeConfig` 必须先经过 A4 allowlist / generated type 解析，再影响 UI 语言、feature flag 或公开配置；缺失或错误响应必须有可测试 fallback。
- 新增 App shell / TopBar / auth / settings 组件时必须参考根目录 `DESIGN.md` 的语义命名和节奏，但 `docs/ui-design/` 与 `ui-design/` 仍是验收真理源；不得机械同步 token 或引入私有授权字体。

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

## 7 关联计划

- [001-app-shell-auth-settings](./plans/001-app-shell-auth-settings/plan.md)
