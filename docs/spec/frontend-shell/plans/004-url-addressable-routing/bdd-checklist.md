# URL Addressable Routing BDD Checklist

> **版本**: 1.12
> **状态**: completed
> **更新日期**: 2026-07-19

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## `BDD.SHELL.ROUTING.001` URL 恢复与安全回退

- [x] Owner behavior tests 覆盖 deep-link、refresh、Back/Forward、auth guard、unsupported URL，以及 Practice visible / Generating hidden chrome。
- [x] 根 `make test` 已执行对应 Vitest；该结果是代码层行为证据，不是 E2E PASS。
- [x] 当前无 routing 真实 E2E owner；已删除 wrapper 场景不再作为证据。
