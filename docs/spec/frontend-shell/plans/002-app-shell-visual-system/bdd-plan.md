# App Shell Visual System BDD Plan

> **版本**: 2.0
> **状态**: active
> **更新日期**: 2026-07-15

**关联 Plan**: [plan](./plan.md)

## 行为合同

| Behavior ID | Given | When | Then | 验证入口 |
|-------------|-------|------|------|----------|
| `BDD.SHELL.VISUAL.001` | 用户在受支持 viewport、语言和显示偏好下打开 shell/Settings | 渲染、切换偏好、点击设置齿轮或查看账号/隐私状态 | 一级导航、单一设置入口、Account/Privacy 层级和可访问交互保持 UI 契约；字体固定且纯显示偏好不写业务事实 | App shell/TopBar/Settings visual and responsive tests，由根 `make test` 承接 |

真实设置主路径只复用 `frontend-shell/001` 对 `E2E.P0.101` 的原地扩展。本 owner 的 source/component/responsive/font gate 属于代码层验证，不能作为 E2E 证据，也不创建并行场景。
