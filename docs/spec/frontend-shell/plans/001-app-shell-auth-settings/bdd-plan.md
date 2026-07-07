# Frontend Shell BDD Plan

> **版本**: 1.13
> **状态**: completed
> **更新日期**: 2026-07-07

## Phase 1: App shell and display behavior

| 场景 ID | 场景 | Given | When | Then | 验证入口 |
|---------|------|-------|------|------|----------|
| E2E.P0.001 | 默认首页与三入口 Shell | 用户未登录，没有保存 App route | 打开 App | 用户看到 Home、三个一级入口、单一登录入口、用户区和显示控制；unsupported route input 不产生独立页面 | `test/scenarios/e2e/p0-001-default-home-shell/` |
| E2E.P0.004 | App Shell 中英语言切换 | 用户打开默认 App shell，浏览器 locale 可归一，TopBar 可见 | 用户通过 TopBar language dropdown 切换语言，并进入 auth / settings shell | TopBar、单一登录入口、用户菜单和 shell 静态文案立即切换；route/testid/业务 params 不变；generated client 请求携带当前 UI locale display hint；runtime locale 与登录态不覆盖前端语言设置 | `test/scenarios/e2e/p0-004-app-shell-language-switch/` |

## Phase 2: Auth interruption and dev mock state

| 场景 ID | 场景 | Given | When | Then | 验证入口 |
|---------|------|-------|------|------|----------|
| E2E.P0.002 | 登录打断后恢复原业务动作 | 用户未登录，并从当前面试规划触发受保护动作 | 用户完成 email-code mock auth | App 恢复到 `practice`，并保留 planId / targetJobId / jdId / resumeId / roundId 等 safe params | `test/scenarios/e2e/p0-002-auth-pending-action-resume/` |
| E2E.P0.032 | Dev mock 登录态菜单与退出闭环 | 用户在 Vite dev mock App 中打开首页，初始没有 session | 用户完成 email-code mock 登录，打开头像菜单，进入 settings，再退出登录 | 默认首屏是非登录态；登录后 TopBar 显示头像 chip + dropdown；settings/logout 可从 dropdown 分流；logout 后 `/me` 回到 unauthenticated，TopBar 回到单一登录入口 | `test/scenarios/e2e/p0-032-dev-mock-auth-state-and-user-menu/` |

## Phase 3: Real auth and protected route guard

| 场景 ID | 场景 | Given | When | Then | 验证入口 |
|---------|------|-------|------|------|----------|
| E2E.P0.101 | Mailpit email-code single-entry login + profile setup | 本地 frontend real mode、backend 和 Mailpit 可用；邮箱是唯一账号标识；新邮箱尚未补全资料 | 用户从 `auth_login` 提交新邮箱，从 Mailpit 读取 6 位验证码，在 `auth_verify` 输入 code；随后提交 displayName + 条款确认；退出后同邮箱再次登录 | 新邮箱首次登录签发 session 但 `/me.profileCompletionRequired=true`，资料补全前不恢复 pendingAction；补全后 `/me.profileCompletionRequired=false` 且 TopBar 显示 displayName；同邮箱再次登录不再进入资料补全 | `E2E.P0.101` scenario assets |
| E2E.P0.102 | 未登录首页与面试业务路由登录前置 | 用户未登录，Home 可公开访问，业务 route 与业务 API 均需要账号 session | 用户打开 Home、直开业务 route，或触发 Home 业务 CTA | Home 不展示账号记录、不请求 `listTargetJobs`、不显示 raw `AUTH_UNAUTHORIZED`；业务 route 在 auth loading 期间不挂载业务 screen，确认未登录后进入 `auth_login(pendingAction)`；backend gate 证明业务 API 由 session middleware 返回 B1 auth envelope | `test/scenarios/e2e/p0-102-auth-gated-interview-routes/` |
