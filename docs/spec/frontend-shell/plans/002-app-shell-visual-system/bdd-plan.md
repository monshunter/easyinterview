# App Shell Visual System BDD Plan

> **版本**: 2.3
> **状态**: completed
> **更新日期**: 2026-07-20

**关联 Plan**: [plan](./plan.md)

## 行为合同

| Behavior ID | Given | When | Then | 验证入口 |
|-------------|-------|------|------|----------|
| `BDD.SHELL.VISUAL.001` | 用户在受支持 viewport、语言和显示偏好下打开 shell/Home/Settings/Practice | 使用 dark/language、点击单一圆形用户名首字符设置入口、在 Appearance 保存主题，或进入 Practice | desktop chrome 对齐 76px 参考节奏；settings initial mark 保持单一 action 且不产生账号 menu；Settings/Practice/字体与业务事实隔离合同不变 | App shell/TopBar/Home/Settings/Practice visual and responsive tests，由根 `make test` 承接 |
| `BDD.SHELL.VISUAL.002` | 用户在 desktop/mobile 打开含主次、危险、失败恢复或小型图标 action 的正式页面 | 查看或操作 TopBar/Auth/Home/Workspace/Parse/Practice/Reports/Report/Generating/Resume/Settings 的有框按钮，并触发 focus/disabled/pending/error 状态 | inventory 中所有矩形/方形有框 action 使用统一 `8px` computed radius，Settings 保存/退出/注销/dialog action 一致；圆形 initial、pill toggle、无边框 link/back 与非按钮 surface 保持原语义，业务状态机、点击区域和 viewport containment 不变 | Token/framed-action source contract + affected component tests + current-run Chrome desktop/mobile manual acceptance；根 `make test` 承接代码层回归 |

真实设置主路径只复用 `frontend-shell/001` 对 `E2E.P0.101` 的原地扩展。本 owner 的 source/component/responsive/font gate 属于代码层验证，不能作为 E2E 证据，也不创建并行场景。
