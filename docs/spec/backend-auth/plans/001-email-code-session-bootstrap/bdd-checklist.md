# Backend Auth BDD Checklist

> **版本**: 1.8
> **状态**: completed
> **更新日期**: 2026-07-14

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## `BDD.AUTH.EMAIL.001` Email-code session 与资料补全

- [x] Owner behavior tests 覆盖 challenge/verify、session、profile completion、logout/relogin、重放与隐私。
- [x] 根 `make test` 已执行对应 Go tests；该结果是代码层行为证据，不是 E2E PASS。

## `E2E.P0.101` 静态 handoff

- [x] 仅关联真实 frontend、backend 与 Mailpit 的 email-code 登录链路。
- [x] 不使用 fixture transport、进程内 handler 或代码层测试冒充 E2E。

以上仅为静态 owner/证据合同；真实运行状态只由 `e2e-scenarios-p0/001` 维护。本轮未运行 `E2E.P0.101`，当前结果仍为 `Ready`。
