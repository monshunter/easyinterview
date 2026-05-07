# App Shell Visual System BDD Checklist

> **版本**: 1.2
> **状态**: completed
> **更新日期**: 2026-05-07

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## E2E.P0.005 App Shell 视觉系统 smoke

- [x] 创建场景目录 `test/scenarios/e2e/p0-005-app-shell-visual-system-smoke/`
- [x] 准备测试数据：未登录用户、默认 runtime config、warm/light 初始显示偏好、可激活的 dark 与 custom accent 状态、auth/profile/settings/placeholder route 集
- [x] 引入 visual smoke + parity 工具：当前阶段使用 vitest + jsdom 作为「等价工具」，注入 `themes.css` / `typography.css` / `topbar.css` / `auth.css` / `screens.css` 后同时验证 DOM 锚点、computed CSS variable resolution（`:root[data-theme=X][data-mode=Y]`）、`customAccent` inline overlay 范围、legacy 负向；DOM 锚点 + className + ui-design 源字面量追溯三类 parity 由 `p0-005-app-shell-visual-system-smoke.test.tsx` 与 README §6 共同记录。bounding-box / 截图差异 / desktop+mobile viewport 真实浏览器 parity 视为后续 Playwright follow-up，按 README §6 列出的步骤接入
- [x] 实现 setup / trigger / verify / cleanup；verify 断言 trigger.log 包含 `Tests 7 passed (7)` 与 `Test Files 1 passed (1)`，且 retired-module testid 不出现
- [x] 实现旧入口负向验证：welcome、growth、mistakes、drill、独立 voice、retired-module 文案在 vitest scenario + verify.sh 中均阻止回流
- [x] 执行并通过场景验证
  <!-- verified: 2026-05-07 method=scenario evidence=".test-output/e2e/p0-005-app-shell-visual-system-smoke/trigger.log Tests 7 passed (7); verify.sh PASS; legacy testid grep clean" -->
- [x] 记录验证证据
  <!-- verified: 2026-05-07 method=scenario evidence="trigger.log + verify.sh; INDEX entry added to test/scenarios/e2e/INDEX.md" -->

## Regression 场景重跑

- [x] 重跑 `E2E.P0.001` 默认首页与五入口 Shell，并记录证据
  <!-- verified: 2026-05-07 method=scenario evidence="setup→trigger→verify→cleanup PASS" -->
- [x] 重跑 `E2E.P0.002` 登录打断后恢复原业务动作，并记录证据
  <!-- verified: 2026-05-07 method=scenario evidence="setup→trigger→verify→cleanup PASS" -->
- [x] 重跑 `E2E.P0.004` App Shell 中英语言切换，并记录证据
  <!-- verified: 2026-05-07 method=scenario evidence="setup→trigger→verify→cleanup PASS" -->
