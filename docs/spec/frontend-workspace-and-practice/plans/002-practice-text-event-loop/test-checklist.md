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
  <!-- verified: 2026-07-14 method=vitest+P0.044+P0.046+pixel-parity evidence="Immediate pending row, thinking/lock, draft preservation, same-ID retry icon, terminal no-retry, refresh recovery and reportId-only completion pass across focused tests and current desktop/mobile screenshots." -->

## Phase 10

- [x] UI source status-injection and terminal generic CTA contract tests pass.
- [x] Hook/loader signal forwarding, cleanup abort, bounded read and unresolved-data fail-lock tests pass.
- [x] Exact 95,000 ms timeout, same-ID reconciliation, all server-status adoption, uncertain-read fallback and stale-response guard tests pass.
- [x] Historical Phase 10 terminal `parse(targetJobId)` route/i18n/a11y matrix passed；Phase 11 supersedes only this route destination.
- [x] Persisted retryable row + failed online refresh keeps one row-local retry and editable next draft；typed HTTP `retryable=true` reuses the exact ID/text without duplicate or technical leakage. (`PracticeScreen.test.tsx` + hooks: 62/62；frontend typecheck)
- [x] Formal/prototype four-state desktop/mobile DOM/style/bbox/viewport/pixel parity passes. (`practice.spec.ts`: 16/16 desktop 1440 + mobile 390；`ui-design-contract.test.mjs`: 60/60)
- [x] P0.044/P0.046 source-fingerprint, screenshot-hash/geometry and exact-marker scenario gates pass with fresh artifacts. (fresh serial runs `13f3b898-4054-4949-8b85-4a15df35c712` / `e26ba887-5f71-4c25-834e-448b4595ede2`)

## Phase 11

## Phase 12

- [x] Runtime message/session ASCII/multibyte limit/+1 and override/fallback tests pass.
- [x] Overflow zero-send/draft preservation and P0.046 backend-authority tests pass.
  <!-- verified: 2026-07-14 evidence="Focused frontend byte-limit suites, full 126 files/1018 tests and fresh P0.046 backend-authority/zero-send evidence pass." -->

- [x] Shared user/assistant `react-markdown + remark-gfm` renderer tests pass with `skipHtml` and no `rehypeRaw`.
  <!-- verified: 2026-07-14 command="pnpm --dir frontend exec vitest run src/app/screens/practice/components/Transcript.test.tsx" result="2/2 PASS after recorded RED" -->
- [x] Raw HTML/event handlers, remote image and unsafe-URI negative tests pass；safe external links are hardened.
  <!-- verified: 2026-07-14 command="pnpm --dir frontend exec vitest run src/app/screens/practice/components/Transcript.test.tsx" result="3/3 PASS after recorded remote-image RED" -->
- [x] Exact raw text/clientMessageId send and same-ID retry tests pass without DOM/normalized-Markdown payload drift or next-draft mutation.
  <!-- verified: 2026-07-14 command="pnpm --dir frontend exec vitest run src/app/screens/practice/PracticeScreen.test.tsx -t 'rehydrates one retryable row|offers one same-ID retry|exact raw Markdown bytes'" result="3/3 PASS after recorded trim-drift RED" -->
- [x] Desktop/mobile GFM typography, local pre/code/table overflow and zero document-overflow parity tests pass.
  <!-- verified: 2026-07-14 commands="node --test ui-design/ui-design-contract.test.mjs; pnpm --dir frontend exec playwright test tests/pixel-parity/practice.spec.ts --grep 'user and assistant GFM'" result="64/64 UI contract plus 2/2 desktop/mobile parity PASS after recorded prototype-state RED" -->
- [x] Terminal CTA exact Workspace-detail route passes；query-free workspace, planId and current-scope `parse(targetJobId)` recovery remain negative.
  <!-- verified: 2026-07-14 commands="focused terminal Vitest; node --test ui-design/ui-design-contract.test.mjs; Playwright terminal grep" result="2/2 unit plus 64/64 UI contract plus 2/2 desktop/mobile parity PASS after recorded parse-route RED" -->
- [x] Existing P0.044/P0.046 pass with fresh source fingerprints, user+assistant GFM, malicious-content, raw-retry, terminal-route and 1440/390 evidence.
  <!-- verified: 2026-07-14 commands="P0.044 and P0.046 serial setup/trigger/verify/cleanup" result="Runs c71ceb11-300f-4b35-b68c-2ec726b8d4f7 / 04fd68b6-1ed7-49d7-b271-a37c236ea541 PASS on source SHA 85f46c92fd79aa61d6894b86d4b80bc8ac58d1acda9efac9cb8cc1419f761fdf; 12 current PNGs total, 8/8 browser checks per scenario, source fingerprint current, PostgreSQL residual=0." -->
