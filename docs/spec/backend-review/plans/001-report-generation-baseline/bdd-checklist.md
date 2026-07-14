# Grounded Conversation Report BDD Checklist

> **版本**: 2.21
> **状态**: active
> **更新日期**: 2026-07-14

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## `BDD.REPORT.GENERATE.001` Grounded report generation

- [x] Owner behavior tests 覆盖 frozen context、validator/repair/retry、persistence、replay、input guard 与 privacy/fencing。
- [x] 根 `make test` 已执行对应 Go tests；该结果是代码层行为证据，不是 `E2E.P0.099` PASS。

## E2E.P0.099 静态资产审计

- [x] Tracked runbook requires isolated current-run en/zh ready reports and one honest generating resource in the shared real stack.
- [x] The browser contract requires exactly six canonical `fullPage: true` desktop/mobile images without request interception, fixture transport or mock backend.
- [x] The evidence contract binds authenticated live report API、read-only PostgreSQL state、canonical report/session/context digests and current screenshot SHA-256, with bounded redacted cleanup.
- [x] The manual contract requires a no-OCR review of ready/generating state、complete action region、clipping/ellipsis/hiding/overflow and false-ready claims.

## E2E.P0.099 真实环境证据

- [ ] 在当前真实环境显式运行场景并记录 exact-six current-run PASS；本轮仅完成静态资产审计，未执行该场景。

## Independent gates

- [x] Validator、repair/retry、persistence、canonical-round overview and small injected input guard are covered by owner code/integration tests.
- [x] Provider/judge reliability is covered by independent eval; it is not an E2E scenario and is not a prerequisite for P0.099.
- [x] Root `make test` remains the complete backend/frontend unit regression gate outside E2E；this classification does not claim a scenario run.
