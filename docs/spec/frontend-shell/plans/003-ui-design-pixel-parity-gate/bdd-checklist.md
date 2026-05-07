# UI-Design Pixel Parity Gate BDD Checklist

> **版本**: 1.0
> **状态**: active
> **更新日期**: 2026-05-08

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## E2E.P0.006 真实浏览器 ui-design pixel parity gate

- [ ] 创建场景目录 `test/scenarios/e2e/p0-006-ui-design-pixel-parity-gate/`
- [ ] 准备测试数据：未登录用户、默认 runtime config、warm/light 初始显示偏好、可激活的 dark 与 customAccent 状态、需要 frontend dist 与 ui-design golden preview 同时可加载、chromium 已通过 `pnpm exec playwright install --with-deps chromium` 安装
- [ ] 引入 visual smoke + parity 工具：`@playwright/test` + chromium；`frontend/playwright.config.ts` 声明 desktop + mobile project 与 `webServer` 指向 `serve-pixel-parity.mjs`；4 个 spec（topbar / screens / layout / screenshot）覆盖 DOM 锚点、computed style、bounding box、screenshot diff
- [ ] 实现 setup / trigger / verify / cleanup；setup 预检 chromium + dist；trigger 跑 `pnpm --filter @easyinterview/frontend test:pixel-parity`；verify 断言 trigger.log 包含 8 项 spec（4 spec × 2 project）PASS、0 failed、retired entry 不回流；cleanup 清理 setup marker
- [ ] 实现旧入口负向验证：welcome、growth、mistakes、drill、独立 voice、prototype data runtime import、retired 文案在 spec 与 verify.sh 中均阻止回流
- [ ] 执行并通过场景验证
- [ ] 记录验证证据

## Regression 场景重跑

- [ ] 重跑 `E2E.P0.005` jsdom App shell 视觉系统 smoke，并记录证据
- [ ] 重跑 `E2E.P0.001` 默认首页与五入口 Shell，并记录证据
- [ ] 重跑 `E2E.P0.002` 登录打断后恢复原业务动作，并记录证据
- [ ] 重跑 `E2E.P0.004` App Shell 中英语言切换，并记录证据
