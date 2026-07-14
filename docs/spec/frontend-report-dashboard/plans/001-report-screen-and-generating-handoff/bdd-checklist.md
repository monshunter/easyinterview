# Honest Grounded Report Screen BDD Checklist

> **版本**: 3.7
> **状态**: active
> **更新日期**: 2026-07-15

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## `BDD.REPORT.UI.001` Report/generating UI

- [x] Owner behavior tests 覆盖 polling、typed failure、ready display、CTA、route recovery、identity/context 与 privacy。
- [x] 根 `make test` 已执行对应 Vitest；该结果是代码层行为证据，不是 `E2E.P0.099` PASS。

## E2E.P0.099 静态资产审计

- [x] The tracked scenario requires shared real frontend/backend/provider and current authenticated report API without request interception or mock transport.
- [x] The capture contract requires current-run en/zh ready reports、one honest generating resource and exactly six canonical `fullPage: true` desktop/mobile images.
- [x] The evidence contract binds current DB/API state and canonical report/session/context/screenshot digests, rejects stale/cross-run/manual-only state, and requires bounded redaction.
- [x] The manual contract requires a no-OCR review of ready/generating truthfulness、complete action region and clipping/ellipsis/hiding/overflow.

## E2E.P0.099 真实环境证据边界

本 checklist 只完成 owner 关联与静态资产审计；本轮未运行场景或记录 exact-six current-run PASS，当前结果以场景 INDEX 的 `Ready` 为准，后续只由显式 `/scenario-run` 产生。

## `BDD.REPORT.CONVERSATION.001` Report-owned transcript

- [ ] Owner behavior tests 覆盖报告四状态、strict order/role、Markdown security、父报告 Back、reportId switch、missing/cross-user/error 与 no-ID/no-live-control negatives。
- [ ] Prototype/formal source structure 与 visual geometry parity 分层通过；正式页面没有先于 `ui-design/` 自由设计。
- [ ] 根 `make test` 执行对应 frontend/generated-client tests；该结果仍不是 E2E PASS。
- [ ] P0.099 静态合同增加 real click/load/back + API/DB binding，且 exact-six screenshot manifest/directory/manual audit 行数保持不变。

## Independent code gates

- [x] Polling、typed failure、CTA、focus、route recovery、ReportsScreen isolation、privacy/a11y and deterministic parity remain code-level tests.
- [x] Exact 24/64 tests and provider/eval reliability do not become P0.099 steps or prerequisites.
- [x] Root `make test` remains the complete frontend/backend unit regression gate outside E2E；this classification does not claim a scenario run.
