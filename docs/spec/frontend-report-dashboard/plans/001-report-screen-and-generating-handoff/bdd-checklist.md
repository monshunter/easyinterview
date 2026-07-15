# Honest Grounded Report Screen BDD Checklist

> **版本**: 3.8
> **状态**: active
> **更新日期**: 2026-07-15

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## Historical `BDD.REPORT.UI.001` through Phase 11

- [x] Owner behavior tests 覆盖 polling、typed failure、ready display、CTA、route recovery、identity/context 与 privacy。
- [x] 根 `make test` 已执行对应 Vitest；该结果是代码层行为证据，不是 `E2E.P0.099` PASS。

## E2E.P0.099 静态资产审计

- [x] The tracked scenario requires shared real frontend/backend/provider and current authenticated report API without request interception or mock transport.
- [x] The capture contract requires current-run en/zh ready reports、one honest generating resource and exactly six canonical `fullPage: true` desktop/mobile images.
- [x] The evidence contract binds current DB/API state and canonical report/session/context/screenshot digests, rejects stale/cross-run/manual-only state, and requires bounded redaction.
- [x] The manual contract requires a no-OCR review of ready/generating truthfulness、complete action region and clipping/ellipsis/hiding/overflow.

## E2E.P0.099 真实环境证据边界

本 checklist 只完成 owner 关联与静态资产审计；本轮未运行场景或记录 exact-six current-run PASS，当前结果以场景 INDEX 的 `Ready` 为准，后续只由显式 `/scenario-run` 产生。

## Independent code gates

- [x] Polling、typed failure、CTA、focus、route recovery、ReportsScreen isolation、privacy/a11y and deterministic parity remain code-level tests.
- [x] Exact 24/64 tests and provider/eval reliability do not become P0.099 steps or prerequisites.
- [x] Root `make test` remains the complete frontend/backend unit regression gate outside E2E；this classification does not claim a scenario run.

## Phase 12 `BDD.REPORT.UI.001` revision

- [x] RED/GREEN owner behavior tests cover ready desktop `3/2/2/2/1` and mobile same-order single-column layout.
- [x] Ready behavior proves top Summary Metrics contain only dimension/evidence counts；bottom full-width Overall Summary contains localized readiness and the unchanged server `summary` exactly once.
- [x] DOM/a11y/geometry negatives reject a top readiness+summary card、duplicate summary、wrong mobile order or a non-full-width desktop Overall Summary.
- [x] Root `make test` is rerun after implementation；the historical Phase 1-11 result does not satisfy this revision.

## Phase 12 real E2E evidence

- [ ] Align P0.099 README/manual-audit/capture-verification assertions, then explicitly run it against the current real environment after implementation；ready full-page images include actions and the following bottom interview summary.
- [ ] Record only current-run exact-six/manual audit evidence；the prior static asset audit and historical PASS state do not satisfy Phase 12.

## Phase 13 `BDD.REPORT.CONVERSATION.001` integration

- [x] Code-owner behavior tests prove Report and ReportsScreen entries open the same reportId-only readonly transcript.
- [x] Ready/non-ready Back, reportId switch stale fence, safe Markdown and strict role/sequence/closed-projection failures are covered.
- [x] Backend owner tests prove report-to-session authorization/binding, malformed locator no-read and hidden-404 semantics.
- [x] Root `make test` is rerun after conflict resolution; P0.099 remains a separate real API/UI handoff and is not claimed by these code tests.
  <!-- verified: 2026-07-15 method=domain-behavior+root-regression evidence="code-owner behavior PASS; no real E2E run claimed" -->
