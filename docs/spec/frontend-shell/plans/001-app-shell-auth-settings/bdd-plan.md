# Frontend Shell BDD Plan

> **版本**: 1.6
> **状态**: completed
> **更新日期**: 2026-05-10

## Phase 2: TopBar and display controls

| 场景 ID | 场景 | Given | When | Then | 验证入口 |
|---------|------|-------|------|------|----------|
| E2E.P0.001 | 默认首页与五入口 Shell | 用户没有登录且没有保存 route | 打开 App | 用户看到 Home、五个一级入口、登录/注册、用户区和显示控制；不会看到 welcome、独立 voice 或旧模块入口 | `test/scenarios/e2e/p0-001-default-home-shell/` |
| E2E.P0.004 | App Shell 中英语言切换 | 用户打开默认 App shell，浏览器 locale 可被归一为中文，未登录用户仍可见 TopBar | 用户通过 TopBar language dropdown 把语言从中文切到 English，并进入 auth / profile / settings shell | TopBar、登录注册、用户菜单和 D1 shell 静态文案立即切换为英文；route/testid/业务 params 不变；后续 generated client 请求携带当前 `Accept-Language` display hint；runtime locale 与登录态不覆盖前端语言设置；控件结构与 `ui-design/src/app.jsx` 一致 | `test/scenarios/e2e/p0-004-app-shell-language-switch/` |

## Phase 3: Auth pages and pending action

| 场景 ID | 场景 | Given | When | Then | 验证入口 |
|---------|------|-------|------|------|----------|
| E2E.P0.002 | 登录打断后恢复原业务动作 | 用户未登录且在当前面试规划点击 `立即面试` | 通过 passwordless mock auth 登录成功 | App 恢复到 `practice`，并保留 planId / targetJobId / jdId / resumeVersionId / roundId | `test/scenarios/e2e/p0-002-auth-pending-action-resume/` |

## Phase 6: Auth state and user menu parity remediation

| 场景 ID | 场景 | Given | When | Then | 验证入口 |
|---------|------|-------|------|------|----------|
| E2E.P0.032 | Dev mock 登录态菜单与退出闭环 | 用户在 Vite dev 默认 fixture-backed mock App 中打开首页，初始没有 session | 用户完成 passwordless mock 登录，打开头像菜单，进入 profile/settings，再执行退出登录 | 默认首屏是非登录态；登录后 TopBar 显示与 `ui-design/src/app.jsx` 一致的头像 chip + dropdown；profile/settings/logout 均可从 dropdown 分流；logout 后 `/me` 回到 unauthenticated，TopBar 回到登录 / 注册，旧 inline 三按钮结构不回流 | `test/scenarios/e2e/p0-032-dev-mock-auth-state-and-user-menu/` |
