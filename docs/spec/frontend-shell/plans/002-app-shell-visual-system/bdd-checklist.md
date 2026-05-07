# App Shell Visual System BDD Checklist

> **版本**: 1.1
> **状态**: active
> **更新日期**: 2026-05-07

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## E2E.P0.005 App Shell 视觉系统 smoke

- [ ] 创建场景目录 `test/scenarios/e2e/p0-005-app-shell-visual-system-smoke/`
- [ ] 准备测试数据：未登录用户、默认 runtime config、warm/light 初始显示偏好、可激活的 dark 与 custom accent 状态、auth/profile/settings/placeholder route 集
- [ ] 引入 visual smoke + parity 工具：优先 Playwright 或等价浏览器渲染工具；工具必须能同时打开正式 frontend 与 `ui-design` golden preview，在 desktop / mobile viewport 读取 DOM、computed style、bounding box 和必要截图证据，并输出 source-to-target 映射
- [ ] 实现 setup / trigger / verify / cleanup；verify 必须断言默认 App shell、TopBar、auth/profile/settings/placeholder shell 非空渲染，核心控件不重叠，warm/light、dark、custom accent 产生可见差异，正式 frontend 与 `ui-design` golden preview 的关键 DOM 锚点、computed style、bounding box 与必要截图差异满足 100% 源级复刻阈值，任何可见偏差必须修正或回到 `ui-design/` 更新真理源，D1 testid / route / i18n 行为不变
- [ ] 实现旧入口负向验证：welcome、growth、mistakes、drill、独立 voice、prototype data runtime import 不得回流
- [ ] 执行并通过场景验证
- [ ] 记录验证证据

## Regression 场景重跑

- [ ] 重跑 `E2E.P0.001` 默认首页与五入口 Shell，并记录证据
- [ ] 重跑 `E2E.P0.002` 登录打断后恢复原业务动作，并记录证据
- [ ] 重跑 `E2E.P0.004` App Shell 中英语言切换，并记录证据
