# Honest Grounded Report Screen Test Checklist

> **版本**: 3.11
> **状态**: completed
> **更新日期**: 2026-07-16

**关联 Test Plan**: [test-plan](./test-plan.md)

## Generating and report behavior

- [x] Polling schedule、pause/resume fences、single-run cap and truthful typed state tests pass.
- [x] Direct report contract、empty/non-empty focus、CTA request and mixed UI/report language tests pass.
- [x] Invalid/over-limit payloads fail closed without raw label, truncation or rewrite.

## Layout and privacy

- [x] Exact 24/64 deterministic fixtures wrap completely at desktop/mobile; 25/65 fail closed.
- [x] Historical formal frontend DOM/style/bbox/viewport tests through Phase 11 pass as code-level visual regression.
- [x] Report/session UUID sentinel DOM/a11y negatives and target/round/resume/interview-record positives pass；frozen-resume URL and privacy-preserving conversation action remain usable.

## Phase 12 report summary hierarchy

- [x] RED proves the current third top readiness+summary metric violates the revised design before implementation.
- [x] DOM tests prove the previously delivered ready group order/count `3/2/2/2/1` and that Overall Summary follows Next Actions；the current four-item Context Strip revision is owned by the unchecked Phase 12 gates below.
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
- [x] `ConversationReport` and responsive contract tests pass for the four peer context children、canonical resume href、privacy-preserving conversation action、SPA clicks、desktop equal-height detail pairs and mobile single-column reset.
  <!-- verified: 2026-07-15 frontend full suite 126 files / 1003 tests PASS; Chrome desktop/mobile geometry and navigation PASS. -->
- [x] P0.099 README/manual-audit/capture-verification assertions are aligned with four context items、responsive detail-pair alignment、actions and the following bottom interview summary；the current ready behavior is separately accepted through real-backend Chrome desktop/mobile evidence, without reusing historical exact-six images or duplicating unchanged generating/language resources.
  <!-- verified: 2026-07-15 P0.099 evidence tests 8 PASS; current focused Chrome screenshots saved under .test-output/acceptance/report-context-grid/. -->

## Phase 13 report conversation integration

- [x] Focused ReportConversation/ReportsScreen/Report route and safe-Markdown tests pass.
- [x] Focused backend report conversation service/store/handler tests pass with malformed-ID no-read and hidden-404 boundaries.
- [x] OpenAPI fixture/codegen/inventory/diff gates pass and `listPracticeSessions` has no active fixture/generated/runtime consumer.
- [x] Root `make test` passes while deleted Demo/prototype-sync assets remain absent.
  <!-- verified: 2026-07-15 method=focused+root-regression evidence="frontend 989; Python 551+4493 subtests; Go all packages; build/docs/codegen/pruning PASS" -->

## Phase 14 failed report recovery

- [x] Generated-client bodyless POST/IK/typed response contract test passes.
- [x] ReportsScreen failed state, locator separation, oversize, pending/double-click/key/stale/malformed matrices pass.
- [x] Accessibility and raw-error/internal-ID negative tests pass.
- [x] Failed conversation hides Back while its report owner is resolving, then routes a trusted target to Reports or a resolved untrusted result to Workspace.
  <!-- verified: 2026-07-16 method=focused-vitest evidence="ReportConversationScreen 19/19 PASS; the RED first exposed an immediate workspace Back while getFeedbackReport was pending" -->
- [x] Focused frontend, typecheck/build and root `make test` pass after GREEN.
  <!-- verified: 2026-07-16 evidence="focused report recovery/conversation/client suites, typecheck and build PASS; root frontend 126/1026 PASS" -->

## Phase 15 completed-session conversation availability

- [x] RED captures missing latest-attempt conversation action beside queued/generating progress.
- [x] queued/generating/latest-ready/current/same-ID/empty action matrix passes with localized distinct a11y names.
- [x] Focused frontend, typecheck/build and Chrome real-environment acceptance pass after GREEN.
- [x] Root `make test` passes after GREEN.
  <!-- verified: 2026-07-16 evidence="Python 584/4583 subtests, Go all packages, frontend 126/1026 PASS" -->
