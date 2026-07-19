# App Shell Visual System BDD Checklist

> **版本**: 2.1
> **状态**: completed
> **更新日期**: 2026-07-19

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## `BDD.SHELL.VISUAL.001` Shell 与显示偏好

- [x] Owner behavior tests 覆盖 shell 渲染、TopBar 零主题菜单、明确设置齿轮、Settings Appearance/Account/Privacy 状态、主题 draft/save/error、固定字体与业务状态隔离。
- [x] 根 `make test` 执行对应 Vitest；该结果是代码层行为证据，不是 E2E PASS。
- [x] Source/responsive/font gate 不包装成场景；真实设置路径仅引用 001 对 `E2E.P0.101` 的原地扩展。
