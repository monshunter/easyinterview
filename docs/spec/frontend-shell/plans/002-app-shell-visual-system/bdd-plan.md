# App Shell Visual System BDD Plan

> **版本**: 1.7
> **状态**: completed
> **更新日期**: 2026-07-14

## Visual Smoke

| 场景 ID | 场景 | Given | When | Then | 验证入口 |
|---------|------|-------|------|------|----------|
| E2E.P0.005 | App Shell visual smoke | App shell 使用默认 `ocean/light` 显示偏好，未登录用户打开正式前端 mock runtime | 用户打开默认页、切换 dark、用色相/饱和度激活 custom accent，再选择 Ocean 或 Plum，并进入 auth / settings / screen shell 路由 | 根级 CSS variable 随 theme/mode/accent 切换；picker 只含 hue/saturation，preview/value/reset DOM 与旧双语文案、`onClear` / `active` props 零引用；选择预定义主题退出 custom accent；其余 shell/route/i18n/testid 行为稳定 | `test/scenarios/e2e/p0-005-app-shell-visual-system-smoke/` |

真实浏览器 viewport、geometry 与 screenshot buffer 由 `E2E.P0.006` 当前场景承接；P0.005 不维护浏览器安装或截图文件流程。

Phase 19 要求 `E2E.P0.006` 在 desktop/mobile 主题菜单上验证只含 hue/saturation 的 picker DOM、关键 computed style、bounding box、viewport containment 与 screenshot parity；不得出现旧 preview/value/reset 占位、空白区或横向溢出。

## Regression References

| 场景 ID | 场景 | 复用目的 | 验证入口 |
|---------|------|----------|----------|
| E2E.P0.001 | 默认首页与三入口 Shell | 证明视觉系统不破坏默认 App shell 与 route alias negative gate | `test/scenarios/e2e/p0-001-default-home-shell/` |
| E2E.P0.002 | 登录打断后恢复原业务动作 | 证明视觉系统不破坏 auth pendingAction 恢复 | `test/scenarios/e2e/p0-002-auth-pending-action-resume/` |
| E2E.P0.004 | App Shell 中英语言切换 | 证明视觉系统不破坏 language dropdown、i18n 和 `Accept-Language` display hint | `test/scenarios/e2e/p0-004-app-shell-language-switch/` |
| E2E.P0.006 | Real-browser UI parity | 证明最小 custom-accent picker 在 desktop/mobile 与当前 UI truth source 的 DOM、几何和截图一致 | `test/scenarios/e2e/p0-006-ui-design-pixel-parity-gate/` |
