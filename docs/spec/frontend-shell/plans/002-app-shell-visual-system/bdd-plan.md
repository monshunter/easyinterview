# App Shell Visual System BDD Plan

> **版本**: 1.9
> **状态**: completed
> **更新日期**: 2026-07-14

**关联 Plan**: [plan](./plan.md)

## 行为合同

| Behavior ID | Given | When | Then | 验证入口 |
|-------------|-------|------|------|----------|
| `BDD.SHELL.VISUAL.001` | 用户在受支持 viewport、语言和显示偏好下打开 shell | 渲染、切换偏好、刷新或继续导航 | 一级导航、用户区、内容层级和可访问交互保持 UI 契约；纯显示偏好可恢复但不写业务事实 | `frontend/src/app/__tests__/app-shell-visual-system.test.tsx` + `frontend/src/app/topbar/TopBarVisual.test.tsx`，由根 `make test` 承接 |

当前没有真实 API/UI E2E owner。视觉 source contract、组件测试和 parity gate 属于代码层验证，阶段回归统一由根 `make test` 承接，不能作为 E2E 证据。
