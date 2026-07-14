# Frontend Shell Auth and Settings BDD Checklist

> **版本**: 1.17
> **状态**: completed
> **更新日期**: 2026-07-14

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## `BDD.SHELL.AUTH.001` Shell auth 与安全恢复

- [x] Owner behavior tests 覆盖 auth loading、profile-incomplete、pending action、protected route、settings 与 logout recovery。
- [x] 根 `make test` 已执行对应 Vitest；该结果是代码层行为证据，不是 E2E PASS。

## `E2E.P0.101` 静态 handoff

- [x] 仅关联真实 email-code、session 与 profile-completion 行为，不承接通用 shell/pending/settings。
- [x] 不引用已删除场景目录、wrapper 或历史 PASS。

以上仅为静态关联审计；真实运行状态只由 `e2e-scenarios-p0/001` 维护。本轮未运行 `E2E.P0.101`，当前结果仍为 `Ready`。
