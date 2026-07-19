# App Shell Visual System BDD Plan

> **版本**: 2.1
> **状态**: completed
> **更新日期**: 2026-07-19

**关联 Plan**: [plan](./plan.md)

## 行为合同

| Behavior ID | Given | When | Then | 验证入口 |
|-------------|-------|------|------|----------|
| `BDD.SHELL.VISUAL.001` | 用户在受支持 viewport、语言和显示偏好下打开 shell/Settings/Practice | 点击明确设置齿轮，在 Appearance 预览/保存主题，或进入 Practice | TopBar 无主题菜单；Settings 保持 Appearance/Account/Privacy 层级；Practice 显示 global TopBar + Session Header；字体固定且业务事实隔离 | App shell/TopBar/Settings/Practice visual and responsive tests，由根 `make test` 承接 |

真实设置主路径只复用 `frontend-shell/001` 对 `E2E.P0.101` 的原地扩展。本 owner 的 source/component/responsive/font gate 属于代码层验证，不能作为 E2E 证据，也不创建并行场景。
