# Backend Auth BDD Checklist

> **版本**: 1.10
> **状态**: completed
> **更新日期**: 2026-07-15

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## `BDD.AUTH.EMAIL.001` Email-code session 与资料补全

- [x] Owner behavior tests 覆盖 challenge/verify、四字段 `/me`、session、profile completion、logout/relogin、重放与隐私。
- [x] 根 `make test` 执行对应 Go tests；该结果是代码层行为证据，不是 E2E PASS。

## `E2E.P0.101` 静态 handoff

- [x] 静态资产原地扩展真实设置齿轮、姓名/完整邮箱与 logout，且证据不保存完整邮箱；不另建场景。
- [x] 不使用 fixture transport、进程内 handler 或代码层测试冒充 E2E；账号删除不进入该共享场景。

以上 owner 合同由 `e2e-scenarios-p0/001` 维护真实运行状态；当前 run `e2e-p0-101-20260715114513-19516` 已在真实 frontend/backend/Mailpit 环境 PASS，静态 INDEX 状态仍为 `Ready`。
