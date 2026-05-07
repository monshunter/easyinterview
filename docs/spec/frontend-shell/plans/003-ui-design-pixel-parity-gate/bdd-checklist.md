# UI-Design Pixel Parity Gate BDD Checklist

> **版本**: 1.2
> **状态**: completed
> **更新日期**: 2026-05-08

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## E2E.P0.006 真实浏览器 ui-design pixel parity gate

- [x] 创建场景目录 `test/scenarios/e2e/p0-006-ui-design-pixel-parity-gate/`
- [x] 准备测试数据：未登录用户、默认 runtime config、warm/light 初始显示偏好、可激活的 dark 与 customAccent 状态、需要 frontend dist 与 ui-design golden preview 同时可加载、chromium 已通过 `pnpm exec playwright install --with-deps chromium` 安装
- [x] 引入 visual smoke + parity 工具：`@playwright/test ^1.59.1` + chromium-headless-shell 1217；`frontend/playwright.config.ts` 声明 desktop (1440×900) + mobile (390×844) project 与 `webServer` 指向 `node ./scripts/serve-pixel-parity.mjs`；4 个 spec（topbar / screens / layout / screenshot）覆盖 DOM 锚点、computed style、bounding box、screenshot regression
- [x] 实现 setup / trigger / verify / cleanup；setup 预检 chromium 缓存 + frontend dist；trigger 跑 `pnpm --filter @easyinterview/frontend test:pixel-parity`；verify 断言 trigger.log 包含 `46 passed` + 0 failed + 4 spec 路径 + desktop/mobile project markers、retired-entry grep 模式 0 命中；cleanup 清理 setup marker
- [x] 实现旧入口负向验证：welcome、mistakes、growth、drill、独立 voice、prototype data runtime import、retired 文案在 spec（`screens.spec.ts` + `screenshot.spec.ts`）与 verify.sh 中均阻止回流
- [x] 执行并通过场景验证
  <!-- verified: 2026-05-08 method=scenario evidence=".test-output/e2e/p0-006-ui-design-pixel-parity-gate/trigger.log: 46 passed (4 spec × desktop+mobile chromium project)；setup→trigger→verify→cleanup 全 PASS" -->
- [x] 记录验证证据
  <!-- verified: 2026-05-08 method=scenario evidence="trigger.log + verify.sh；test/scenarios/e2e/INDEX.md 已添加 P0.006 行 (Ready)" -->

## Regression 场景重跑

- [x] 重跑 `E2E.P0.005` jsdom App shell 视觉系统 smoke，并记录证据
  <!-- verified: 2026-05-08 method=scenario evidence="setup→trigger→verify→cleanup PASS；trigger.log Tests 7 passed (7)" -->
- [x] 重跑 `E2E.P0.001` 默认首页与五入口 Shell，并记录证据
  <!-- verified: 2026-05-08 method=scenario evidence="setup→trigger→verify→cleanup PASS" -->
- [x] 重跑 `E2E.P0.002` 登录打断后恢复原业务动作，并记录证据
  <!-- verified: 2026-05-08 method=scenario evidence="setup→trigger→verify→cleanup PASS" -->
- [x] 重跑 `E2E.P0.004` App Shell 中英语言切换，并记录证据
  <!-- verified: 2026-05-08 method=scenario evidence="setup→trigger→verify→cleanup PASS" -->
