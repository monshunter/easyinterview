# Frontend Shell BDD Plan

> **版本**: 1.1
> **状态**: active
> **更新日期**: 2026-05-06

## Phase 2: TopBar and display controls

| 场景 ID | 场景 | Given | When | Then | 验证入口 |
|---------|------|-------|------|------|----------|
| E2E.P0.001 | 默认首页与五入口 Shell | 用户没有登录且没有保存 route | 打开 App | 用户看到 Home、五个一级入口、登录/注册、用户区和显示控制；不会看到 welcome、独立 voice 或旧模块入口 | `test/scenarios/e2e/p0-001-default-home-shell/` |

## Phase 3: Auth pages and pending action

| 场景 ID | 场景 | Given | When | Then | 验证入口 |
|---------|------|-------|------|------|----------|
| E2E.P0.002 | 登录打断后恢复原业务动作 | 用户未登录且在当前面试规划点击 `立即面试` | 通过 passwordless mock auth 登录成功 | App 恢复到 `practice`，并保留 planId / targetJobId / jdId / resumeVersionId / roundId | `test/scenarios/e2e/p0-002-auth-pending-action-resume/` |
