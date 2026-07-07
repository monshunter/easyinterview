# App Shell Visual System BDD Plan

> **版本**: 1.5
> **状态**: completed
> **更新日期**: 2026-07-07

## Visual Smoke

| 场景 ID | 场景 | Given | When | Then | 验证入口 |
|---------|------|-------|------|------|----------|
| E2E.P0.005 | App Shell visual smoke | App shell 使用默认 `ocean/light` 显示偏好，未登录用户打开正式前端 mock runtime | 用户打开默认页、切换 dark、激活 custom accent，并进入 auth / settings / screen shell 路由 | TopBar、三入口、显示控制、登录入口、auth shell、settings shell 和通用 screen shell 非空渲染；根级 CSS variable 随 theme/mode/accent 切换；current route anchors 和 className 存在；unsupported route aliases 不产生 standalone screen；D1 route / i18n / testid 行为保持稳定 | `test/scenarios/e2e/p0-005-app-shell-visual-system-smoke/` |

## Regression References

| 场景 ID | 场景 | 复用目的 | 验证入口 |
|---------|------|----------|----------|
| E2E.P0.001 | 默认首页与三入口 Shell | 证明视觉系统不破坏默认 App shell 与 route alias negative gate | `test/scenarios/e2e/p0-001-default-home-shell/` |
| E2E.P0.002 | 登录打断后恢复原业务动作 | 证明视觉系统不破坏 auth pendingAction 恢复 | `test/scenarios/e2e/p0-002-auth-pending-action-resume/` |
| E2E.P0.004 | App Shell 中英语言切换 | 证明视觉系统不破坏 language dropdown、i18n 和 `Accept-Language` display hint | `test/scenarios/e2e/p0-004-app-shell-language-switch/` |
