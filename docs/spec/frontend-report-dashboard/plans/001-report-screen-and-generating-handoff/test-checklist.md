# Honest Grounded Report Screen Test Checklist

> **版本**: 3.6
> **状态**: active
> **更新日期**: 2026-07-15

**关联 Test Plan**: [test-plan](./test-plan.md)

## Generating and report behavior

- [x] Polling schedule、pause/resume fences、single-run cap and truthful typed state tests pass.
- [x] Direct report contract、empty/non-empty focus、CTA request and mixed UI/report language tests pass.
- [x] Invalid/over-limit payloads fail closed without raw label, truncation or rewrite.

## Layout and privacy

- [x] Exact 24/64 deterministic fixtures wrap completely at desktop/mobile; 25/65 fail closed.
- [x] Historical formal frontend DOM/style/bbox/viewport tests through Phase 11 pass as code-level visual regression.
- [x] Report/session UUID sentinel DOM/a11y negatives and target/round/resume positives pass.

## Phase 12 report summary hierarchy

- [x] RED proves the current third top readiness+summary metric violates the revised design before implementation.
- [x] DOM tests prove exact ready group order/count `3/2/2/2/1` and that Overall Summary follows Next Actions.
- [x] Semantic/i18n tests prove Summary Metrics contain only dimension/evidence counts；Overall Summary contains localized title/readiness plus the unchanged server `summary` exactly once.
- [x] Formal frontend 1440/390 style/bbox/viewport/a11y tests prove desktop full-width bottom summary、mobile same-order single column、complete wrapping and zero horizontal overflow.
- [x] Source negatives reject top readiness/summary、duplicate summary、parallel prototype ownership and any backend/API/persistence/prompt change.

## ReportsScreen and routing

- [x] Current-target isolation、canonical join、current/latest-only、loading/empty/error and stale-response fence tests pass.
- [x] Trusted/untrusted Back matrix、reportId-only route and direct Workspace detail without Parse detour tests pass.
- [x] ReportsScreen-only list consumer and no TopBar/global/history/compatibility entry negatives pass.

## Historical E2E separation and full regression through Phase 11

- [x] P0.099 independently binds current real report/generating API/DB/screenshot evidence; mock/component outputs cannot satisfy it.
- [x] Provider/eval and deterministic parity remain independent code gates, not E2E steps.
- [x] Root `make test` runs the complete frontend/backend unit regression for phase completion.

## Phase 12 current gates

- [x] Root `make test`、typecheck/build/lint and document gates pass on the revised implementation.
- [ ] P0.099 README/manual-audit/capture-verification assertions are aligned, then an explicitly run current scenario produces ready full-page evidence covering actions and the following bottom interview summary；historical exact-six PASS is not reused.

## Phase 13 report conversation integration

- [x] Focused ReportConversation/ReportsScreen/Report route and safe-Markdown tests pass.
- [x] Focused backend report conversation service/store/handler tests pass with malformed-ID no-read and hidden-404 boundaries.
- [x] OpenAPI fixture/codegen/inventory/diff gates pass and `listPracticeSessions` has no active fixture/generated/runtime consumer.
- [x] Root `make test` passes while deleted Demo/prototype-sync assets remain absent.
  <!-- verified: 2026-07-15 method=focused+root-regression evidence="frontend 989; Python 551+4493 subtests; Go all packages; build/docs/codegen/pruning PASS" -->
