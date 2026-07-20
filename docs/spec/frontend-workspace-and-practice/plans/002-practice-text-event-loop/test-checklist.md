# 002 Practice Continuous Conversation Test Checklist

> **版本**: 3.2
> **状态**: completed
> **更新日期**: 2026-07-20

**关联 Test Plan**: [test-plan](./test-plan.md)

## Phase 1
- [x] Prototype source/geometry/negative tests pass.
## Phase 2
- [x] Formal DOM/a11y/context/i18n tests pass.
## Phase 3
- [x] Message hooks/state/retry tests pass.
## Phase 4
- [x] Completion/generating tests pass.
## Phase 5
- [x] Full frontend/parity/real screenshot gates pass.
## Phase 6
- [x] Source-aware retry and Finish CTA lifecycle tests pass. (`pnpm --filter @easyinterview/frontend test src/app/screens/practice/PracticeScreen.test.tsx src/app/screens/practice/hooks/useCompletePracticeSession.test.tsx`; frontend typecheck)
## Phase 7
- [x] Zero-answer eligibility, native-disabled/a11y/i18n, backend-authoritative rejection, one-answer completion and replay tests pass.
## Phase 8
- [x] reportId-only one-answer completion/replay scenario tests pass.
## Phase 9
- [x] Generated typed-error contract, server `clientMessageId + replyStatus` rehydration, pending single-flight/no-resend, AI failure → reload → same-ID retry → single reply, terminal no-retry, no browser persistence/string parsing, immediate row/thinking/lock/draft/dedupe and exact 1440/390 parity tests pass.

## Phase 10

- [x] UI source status-injection and terminal generic CTA contract tests pass.
- [x] Hook/loader signal forwarding, cleanup abort, bounded read and unresolved-data fail-lock tests pass.
- [x] Exact 95,000 ms timeout, same-ID reconciliation, all server-status adoption, uncertain-read fallback and stale-response guard tests pass.
- [x] Historical Phase 10 terminal `parse(targetJobId)` route/i18n/a11y matrix passed；Phase 11 supersedes only this route destination.
- [x] Persisted retryable row + failed online refresh keeps one row-local retry and editable next draft；typed HTTP `retryable=true` reuses the exact ID/text without duplicate or technical leakage. (`PracticeScreen.test.tsx` + hooks: 62/62；frontend typecheck)
- [x] Current formal Practice component tests cover four-state DOM, interaction, desktop/mobile responsive and accessibility contracts；the deleted Demo/parity suites are not current evidence.


## Phase 12

- [x] Required-field small-limit ASCII/multibyte and zero-send/draft-preservation tests pass；no per-field fallback/default-sized input.

- [x] Shared user/assistant `react-markdown + remark-gfm` renderer tests pass with `skipHtml` and no `rehypeRaw`.
- [x] Raw HTML/event handlers, remote image and unsafe-URI negative tests pass；safe external links are hardened.
- [x] Exact raw text/clientMessageId send and same-ID retry tests pass without DOM/normalized-Markdown payload drift or next-draft mutation.
- [x] Desktop/mobile GFM typography, local pre/code/table overflow and zero document-overflow parity tests pass.
- [x] Terminal CTA exact Workspace-detail route passes；query-free workspace, planId and current-scope `parse(targetJobId)` recovery remain negative.

## Phase 13

- [x] Reference visual source/component tests pass for available viewport height, shared desktop grid, Session Header, message surfaces, Composer and SVG controls.
- [x] Focused owner tests, root regression and formal-frontend fixture 1916×821 / 390×844 Chrome containment checks pass without changing message/completion truth.<!-- verified: 2026-07-19 evidence="practice Phase 13 focused 54 tests; final root frontend 1054 tests; typecheck/build; documentOverflow=0 at both viewports" -->

## Phase 14

- [x] Fixed Composer/helper source/component tests pass: `Transcript` is the only flexible scroll region; `InputBar` is fixed-size at the conversation bottom, has no ancestor relationship with Transcript, and owns the localized icon capsule above the input shell.
- [x] Short/long transcript Chrome bbox checks prove a stable helper/input gap at desktop/mobile before and after transcript scroll.<!-- verified: 2026-07-19 evidence="gap=8; desktop scrollTop 4→151; mobile scrollTop 0→175.5; helper/input coordinates stable; documentOverflow=0" -->

## Phase 15

- [x] Inner input-surface DOM/CSS RED-GREEN tests pass and reject the old outside-surface row plus overlay/large-right-padding variants.<!-- verified: 2026-07-20 evidence="PracticeVisual and InputBar owner tests PASS after expected RED failures" -->
- [x] Click, Ctrl/Meta+Enter and disabled owner behavior tests remain green.<!-- verified: 2026-07-20 evidence="PracticeScreen focused suite included 49 behavior tests; focused aggregate 55 PASS" -->
- [x] Chrome desktop/mobile containment, full-width content, fixed Composer and zero-overflow checks pass.<!-- verified: 2026-07-20 evidence="real session 1512x829 and 390x844; contained action, no overlap, width invariant, gap=8, overflow=0, no console errors" -->

## Phase 16

- [x] A5 duplicate-ID fixture tests、当前 workspace spec 唯一性扫描与 owner context/docs gates 通过；不新增 runtime test 或 E2E。<!-- verified: 2026-07-20 evidence="5 focused Python contracts, current docs/spec scan, make docs-check and 002 context PASS." -->
