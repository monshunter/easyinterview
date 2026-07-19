# URL Addressable Routing BDD Checklist

> **版本**: 1.13
> **状态**: completed
> **更新日期**: 2026-07-19

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## `BDD.SHELL.ROUTING.001` URL 恢复与安全回退

- [x] Generating 直开、刷新与 Back/Forward 均保留共享 TopBar；不再存在业务 route 的 no-chrome 例外。
- [x] Practice、Parse、Reports、Generating 与报告上下文 route 的 TopBar 一级选中态统一属于“面试”，URL/query ownership 不变。

- [x] Owner behavior tests 覆盖 deep-link、refresh、Back/Forward、auth guard、unsupported URL，以及所有 canonical route 的 visible shared chrome。（TopBar 17、canonical 11、App history 15 PASS；desktop Chrome 四态 PASS。）
- [x] 根 `make test` 已执行对应 Vitest；该结果是代码层行为证据，不是 E2E PASS。
- [x] 当前无 routing 真实 E2E owner；已删除 wrapper 场景不再作为证据。
