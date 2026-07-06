# Frontend Shell BDD Plan

> **版本**: 1.11
> **状态**: completed
> **更新日期**: 2026-07-06

## Phase 2: TopBar and display controls

| 场景 ID | 场景 | Given | When | Then | 验证入口 |
|---------|------|-------|------|------|----------|
| E2E.P0.001 | 默认首页与三入口 Shell | 用户没有登录且没有保存 route | 打开 App | 用户看到 Home、三个一级入口、单一登录入口、用户区和显示控制；不会看到 welcome、注册入口、`jd_match`、`debrief`、`profile`、独立 voice 或旧模块入口 | `test/scenarios/e2e/p0-001-default-home-shell/` |
| E2E.P0.004 | App Shell 中英语言切换 | 用户打开默认 App shell，浏览器 locale 可被归一为中文，未登录用户仍可见 TopBar | 用户通过 TopBar language dropdown 把语言从中文切到 English，并进入 auth / settings shell | TopBar、单一登录入口、用户菜单和 D1 shell 静态文案立即切换为英文；route/testid/业务 params 不变；后续 generated client 请求携带当前 `Accept-Language` display hint；runtime locale 与登录态不覆盖前端语言设置；控件结构与 `ui-design/src/app.jsx` 一致；`profile` 不作为 live shell 出现 | `test/scenarios/e2e/p0-004-app-shell-language-switch/` |

## Phase 3: Auth pages and pending action

| 场景 ID | 场景 | Given | When | Then | 验证入口 |
|---------|------|-------|------|------|----------|
| E2E.P0.002 | 登录打断后恢复原业务动作 | 用户未登录且在当前面试规划点击 `立即面试` | 通过 passwordless mock auth 登录成功 | App 恢复到 `practice`，并保留 planId / targetJobId / jdId / resumeVersionId / roundId | `test/scenarios/e2e/p0-002-auth-pending-action-resume/` |

## Phase 6: Auth state and user menu parity remediation

| 场景 ID | 场景 | Given | When | Then | 验证入口 |
|---------|------|-------|------|------|----------|
| E2E.P0.032 | Dev mock 登录态菜单与退出闭环 | 用户在 Vite dev 默认 fixture-backed mock App 中打开首页，初始没有 session | 用户完成 passwordless mock 登录，打开头像菜单，进入 settings，再执行退出登录 | 默认首屏是非登录态；登录后 TopBar 显示与 `ui-design/src/app.jsx` 一致的头像 chip + dropdown；settings/logout 可从 dropdown 分流；logout 后 `/me` 回到 unauthenticated，TopBar 回到单一登录入口，旧 inline 三按钮、注册按钮结构和 `profile` 菜单项不回流 | `test/scenarios/e2e/p0-032-dev-mock-auth-state-and-user-menu/` |

## Phase 7: Historical real passwordless mail-link remediation

| 场景 ID | 场景 | Given | When | Then | 验证入口 |
|---------|------|-------|------|------|----------|
| E2E.P0.002 | 202 空响应 + 手动 code fallback | 未登录用户从 workspace 触发 auth gate；真实 backend 的 `startAuthEmailChallenge` 返回 `202 Accepted` 且无 response body | 前端提交邮箱并进入 `auth_verify`，用户输入 6 位验证码后验证 | generated client 不抛 JSON parse error；App 保留 pendingAction 并在验证成功后恢复 practice route | `frontend/src/api/generatedClient.test.ts` + `frontend/src/app/AppAuthDispatch.test.tsx` + `test/scenarios/e2e/p0-002-auth-pending-action-resume/` |
| E2E.P0.100 | Mailpit magic-link callback | 已由 Phase 8 code-only 邮件取代 | 不作为当前完成证据 | 不作为当前完成证据；当前 real Mailpit auth 验收使用 E2E.P0.101 | `test/scenarios/e2e/p0-100-real-provider-full-funnel-hybrid/` + focused App auth dispatch tests |

## Phase 8: Email-code auth and display-name remediation

| 场景 ID | 场景 | Given | When | Then | 验证入口 |
|---------|------|-------|------|------|----------|
| E2E.P0.101 | Mailpit email-code single-entry login + profile setup | 本地 frontend real mode、backend 和 Mailpit 可用；邮箱是唯一账号标识；未登录 TopBar 只有登录入口；新邮箱尚未补全资料 | 用户从 `auth_login` 提交新邮箱，从 Mailpit 读取 6 位验证码，在 `auth_verify` 手动输入 code；首次 verify 后刷新资料补全页、关闭/换浏览器后用同一邮箱重新登录；随后提交 displayName + 条款确认；退出后同一邮箱再次登录 | 新邮箱首次登录签发 `ei_session` 但 `/me.profileCompletionRequired=true`，每次重新登录都先进入 `auth_profile_setup`，资料补全前不恢复 pendingAction；补全后 `/me.profileCompletionRequired=false` 且 TopBar 显示 displayName；后续同邮箱登录不再进入资料补全；注册按钮、`auth_register` live page、`purpose=signup/login` request body、displayName-before-verify、magic link URL 和 prototype fallback 均不出现 | `test/scenarios/e2e/p0-101-auth-email-code-login-register/` |

## Phase 10: Unauthenticated interview route guard remediation

| 场景 ID | 场景 | Given | When | Then | 验证入口 |
|---------|------|-------|------|------|----------|
| E2E.P0.102 | 未登录首页与面试业务路由登录前置 | 用户未登录，Home 可公开访问，业务 route 与业务 API 均需要账号 session | 用户打开 Home、直开 `workspace` / `practice` / `report` / `settings` 等业务 route，或打开 `jd_match` / `debrief` / `profile` retired route，或触发 Home 业务 CTA | Home 不展示 Recent mock interviews、不请求 `listTargetJobs`、不显示 raw `AUTH_UNAUTHORIZED`；业务 route 在 auth loading 期间不挂载业务 screen，确认未登录后进入 `auth_login(pendingAction)`；retired route 归一回当前保留 route 或 `home`，不产生旧 screen；后端 focused gate 证明业务 API 由 session middleware 返回 B1 auth envelope | `test/scenarios/e2e/p0-102-auth-gated-interview-routes/` |
