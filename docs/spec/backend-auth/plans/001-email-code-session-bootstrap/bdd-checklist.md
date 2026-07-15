# Backend Auth BDD Checklist

> **版本**: 1.9
> **状态**: active
> **更新日期**: 2026-07-15

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## `BDD.AUTH.EMAIL.001` Email-code session 与资料补全

- [ ] Owner behavior tests 覆盖 challenge/verify、四字段 `/me`、session、profile completion、logout/relogin、重放与隐私。
- [ ] 根 `make test` 执行对应 Go tests；该结果是代码层行为证据，不是 E2E PASS。

## `E2E.P0.101` 静态 handoff

- [ ] 静态资产原地扩展真实设置齿轮、姓名/脱敏邮箱与 logout；不另建场景。
- [ ] 不使用 fixture transport、进程内 handler 或代码层测试冒充 E2E；账号删除不进入该共享场景。

以上仅为静态 owner/证据合同；真实运行状态只由 `e2e-scenarios-p0/001` 维护。本轮未运行 `E2E.P0.101`，当前结果仍为 `Ready`。
