# Grounded Conversation Report BDD Checklist

> **版本**: 2.25
> **状态**: active
> **更新日期**: 2026-07-15

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## `BDD.REPORT.GENERATE.001` Grounded report generation

- [x] Owner behavior tests 覆盖 frozen context、validator/repair/retry、persistence、replay、input guard 与 privacy/fencing。
- [x] 根 `make test` 已执行对应 Go tests；该结果是代码层行为证据，不是 `E2E.P0.099` PASS。

## E2E.P0.099 静态资产审计

- [x] Tracked runbook requires isolated current-run en/zh ready reports and one honest generating resource in the shared real stack.
- [x] The browser contract requires exactly six canonical `fullPage: true` desktop/mobile images without request interception, fixture transport or mock backend.
- [x] The evidence contract binds authenticated live report API、read-only PostgreSQL state、canonical report/session/context digests and current screenshot SHA-256, with bounded redacted cleanup.
- [x] The manual contract requires a no-OCR review of ready/generating state、complete action region、clipping/ellipsis/hiding/overflow and false-ready claims.

## `BDD.REPORT.CONVERSATION.API.001` Report-owned transcript API

- [x] Owner tests cover four report statuses, owned empty `messages` 200, existing unique relation, strict ordered/non-blank projection, hidden 404, corruption fail-closed, zero internal IDs and zero AI/write/new-table behavior.
- [x] Scoped removal gate proves `listPracticeSessions` is absent from current positive OpenAPI/generated/router/handler/fixture/mock/frontend surfaces.
- [x] 根 `make test` 执行对应 Go tests；该结果是代码层行为证据，不是 E2E PASS。
- [x] P0.099 contract adds real conversation API/DB binding and browser click/load/back while preserving exact-six screenshots and bounded redaction.

## E2E.P0.099 真实环境证据

- [x] 在当前真实环境显式运行场景并记录 exact-six current-run PASS。
  <!-- verified: 2026-07-15 run_id="e2e-p0-099-20260715T021319Z-57232" result="PASS" evidence="live report/conversation API and PostgreSQL binding; exact six Chrome screenshots; manual visual audit" -->

## Independent gates

- [x] Validator、repair/retry、persistence、canonical-round overview and small injected input guard are covered by owner code/integration tests.
- [x] Provider/judge reliability is covered by independent eval; it is not an E2E scenario and is not a prerequisite for P0.099.
- [x] Root `make test` remains the complete backend/frontend unit regression gate outside E2E；this classification does not claim a scenario run.
