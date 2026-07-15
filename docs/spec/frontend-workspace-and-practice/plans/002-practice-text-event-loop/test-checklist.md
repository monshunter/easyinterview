# 002 Practice Continuous Conversation Test Checklist

> **版本**: 2.8
> **状态**: completed
> **更新日期**: 2026-07-14

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
- [x] Formal/prototype four-state desktop/mobile DOM/style/bbox/viewport/responsive browser verification passes. (`practice.spec.ts`: 16/16 desktop 1440 + mobile 390；`ui-design-contract.test.mjs`: 60/60)


## Phase 12

- [x] Required-field small-limit ASCII/multibyte and zero-send/draft-preservation tests pass；no per-field fallback/default-sized input.

- [x] Shared user/assistant `react-markdown + remark-gfm` renderer tests pass with `skipHtml` and no `rehypeRaw`.
- [x] Raw HTML/event handlers, remote image and unsafe-URI negative tests pass；safe external links are hardened.
- [x] Exact raw text/clientMessageId send and same-ID retry tests pass without DOM/normalized-Markdown payload drift or next-draft mutation.
- [x] Desktop/mobile GFM typography, local pre/code/table overflow and zero document-overflow parity tests pass.
- [x] Terminal CTA exact Workspace-detail route passes；query-free workspace, planId and current-scope `parse(targetJobId)` recovery remain negative.
