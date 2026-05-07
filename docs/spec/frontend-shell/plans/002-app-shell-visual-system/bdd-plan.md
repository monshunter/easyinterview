# App Shell Visual System BDD Plan

> **版本**: 1.0
> **状态**: active
> **更新日期**: 2026-05-07

## Phase 6: Visual smoke and regression

| 场景 ID | 场景 | Given | When | Then | 验证入口 |
|---------|------|-------|------|------|----------|
| E2E.P0.005 | App Shell 视觉系统 smoke | D1 App shell 已可渲染，用户未登录，显示偏好默认 warm/light，`ui-design/` 是视觉验收真理源头 | 用户打开默认 App shell，切换 dark，激活 custom accent，并在 desktop / mobile viewport 查看 TopBar、auth、profile、settings 和 placeholder shell | 页面非空渲染；TopBar、五入口、显示控制和用户区不重叠；warm/light、dark 与 custom accent 在 computed-style 或截图中产生可见差异；D1 testid / route / i18n 行为不变；welcome、growth、mistakes、drill、独立 voice 等旧入口不回流 | `test/scenarios/e2e/p0-005-app-shell-visual-system-smoke/` |

## Regression References

| 场景 ID | 场景 | 复用目的 | 验证入口 |
|---------|------|----------|----------|
| E2E.P0.001 | 默认首页与五入口 Shell | 证明视觉接入未破坏默认 App shell 与旧入口负向约束 | `test/scenarios/e2e/p0-001-default-home-shell/` |
| E2E.P0.002 | 登录打断后恢复原业务动作 | 证明视觉接入未破坏 auth pendingAction 恢复 | `test/scenarios/e2e/p0-002-auth-pending-action-resume/` |
| E2E.P0.004 | App Shell 中英语言切换 | 证明视觉接入未破坏 TopBar 语言下拉框、i18n 与 `Accept-Language` display hint | `test/scenarios/e2e/p0-004-app-shell-language-switch/` |
