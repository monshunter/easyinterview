# App Shell Visual System BDD Checklist

> **版本**: 1.9
> **状态**: completed
> **更新日期**: 2026-07-14

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## `BDD.SHELL.VISUAL.001` Shell 与显示偏好

- [x] Owner behavior tests 覆盖 shell 渲染、语言、显示偏好恢复与业务状态隔离。
- [x] 根 `make test` 已执行对应 Vitest；该结果是代码层行为证据，不是 E2E PASS。
- [x] 当前无 visual-system 真实 E2E owner；source/parity gate 不包装成场景。
