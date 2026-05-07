# UI-Design Pixel Parity Gate BDD Plan

> **版本**: 1.0
> **状态**: active
> **更新日期**: 2026-05-08

## Phase 5: Pixel parity smoke

| 场景 ID | 场景 | Given | When | Then | 验证入口 |
|---------|------|-------|------|------|----------|
| E2E.P0.006 | 真实浏览器 ui-design pixel parity gate | D2 视觉系统已落地、frontend dist 已构建、`ui-design/index.html` 静态原型可加载、Playwright + chromium 已安装、`pnpm --filter @easyinterview/frontend test:pixel-parity` 可执行 | 在 desktop (1440×900) 与 mobile (390×844) 两个 project 下并行加载 `frontend/dist/index.html` 与 `ui-design/index.html`，跑 topbar / screens / layout / screenshot 四个 spec | 两个 project 下 8 项 spec 全部 PASS：DOM 锚点 + computed style 一致；TopBar / auth / profile / settings 卡片 bounding box 不重叠不溢出；warm/light 截图差异 ≤ 阈值；dark + customAccent 与 light baseline 出现可见像素差异（避免静默失效）；retired entry 不回流；trigger.log 记录全部 PASS / 0 failed | `test/scenarios/e2e/p0-006-ui-design-pixel-parity-gate/` |

## Regression References

| 场景 ID | 场景 | 复用目的 | 验证入口 |
|---------|------|----------|----------|
| E2E.P0.005 | App Shell 视觉系统 smoke + ui-design 100% parity（jsdom） | 证明 fast smoke 不退化，jsdom 范围内 DOM / className / CSS variable / customAccent overlay / legacy 负向 / ui-design 源追溯六类断言继续通过 | `test/scenarios/e2e/p0-005-app-shell-visual-system-smoke/` |
| E2E.P0.001 | 默认首页与五入口 Shell | 证明 Playwright gate 接入未破坏默认 App shell 与旧入口负向约束 | `test/scenarios/e2e/p0-001-default-home-shell/` |
| E2E.P0.002 | 登录打断后恢复原业务动作 | 证明 Playwright gate 接入未破坏 auth pendingAction 恢复 | `test/scenarios/e2e/p0-002-auth-pending-action-resume/` |
| E2E.P0.004 | App Shell 中英语言切换 | 证明 Playwright gate 接入未破坏 i18n 与 `Accept-Language` display hint | `test/scenarios/e2e/p0-004-app-shell-language-switch/` |
