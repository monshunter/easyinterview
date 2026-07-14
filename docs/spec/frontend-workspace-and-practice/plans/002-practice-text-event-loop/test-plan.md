# 002 Practice Continuous Conversation Test Plan

> **版本**: 2.8
> **状态**: completed
> **更新日期**: 2026-07-14

## Phase 1: Prototype/source
- DOM shape, disabled phone control, no stale question/hint/phone-positive source, desktop/mobile geometry.
## Phase 2: Formal DOM
- TopBar props, component deletion, a11y, responsive layout, context/i18n negative checks.
## Phase 3: Hooks/state
- Ordered messages, refresh, send/replay/failure/retry/no duplicates and UI states.
## Phase 4: Completion
- Stable-ID handoff and generating conversation copy.
## Phase 5: Integration/parity
- Vitest/typecheck/build, source/pixel parity and Practice owner-scoped desktop/mobile screenshots.

## Phase 6: Review remediation
- PracticeScreen component tests simulate loader 5xx, send failure and completion failure, then assert the retry button invokes only the matching operation.
- CTA-state tests hold send/load/complete promises open and assert Finish remains disabled until the mutable state returns.

## Phase 7: Zero-answer finish

- Eligibility table covers opening-only, route/draft-only, first committed user message, pending assistant reply, loading, sending, completing and non-mutable session states; only the committed-user/no-pending/mutable idle combination enables Finish.
- Prototype/formal DOM and i18n tests require the same visible zh/en zero-answer reason and stable `aria-describedby`; no hardcoded bilingual copy lives in the component.

## Phase 8: reportId-only handoff

- Component/router tests require exact navigation query/state/context `{reportId}` and fail on any copied target/plan/session/resume/round/status/error field.

## Phase 9: Immediate feedback and message-local recovery

- Prototype/formal component tests hold `sendPracticeMessage` deferred and assert synchronous user-row insertion, composer clear/lock, accessible thinking animation and absence of retry before failure.
- Generated-client tests require typed `ApiClientError.apiError.retryable` plus HTTP status, `code/requestId/retryable/details` and transport cause；Practice classification must not inspect `Error.message`, and intentional abort/unmount is not retryable.
- Server-rehydration tables cover `pending` (thinking + composer/Finish lock + single-flight re-read/no duplicate send), `retryable_failed` (one row-local icon), `terminal_failed` (no retry + fact recovery) and `complete` (one user/reply pair) from `getPracticeSession.clientMessageId + replyStatus`, with no URL/browser-storage retry identity.
- Failure/retry tables assert row-local icon placement, server original text + same-ID payload, next-draft preservation, retry lock, repeated-failure stability and server-adopt deduplication；transport failures without a response or typed `ApiClientError.apiError.retryable=true` are retryable, while validation/auth/not-found/conflict/mismatch have no retry icon and recover through server truth.
- The recovery integration path is explicit: AI failure → page reload → `getPracticeSession` returns the same user text/clientMessageId as `retryable_failed` → retry → exactly one assistant reply and no duplicate user row.
- Finish-state tests cover pending, retryable-failed, retrying and terminal-recovery in addition to existing loader/completion guards；no unresolved message state can enable completion.

## Phase 10: Timeout reconciliation, terminal recovery and fresh parity

- UI source contract RED/GREEN covers injected initial `replyStatus`, immediate/persisted pending, retryable icon and terminal generic CTA. Promise success-only mocks are insufficient；terminal and retry branches must be reachable.
- `usePracticeMessages` tests assert the exact request body plus forwarded `AbortSignal`. Loader/reconcile tests assert cleanup abort, bounded reads and preservation/fail-locking of unresolved same-session data after refresh failure.
- PracticeScreen fake-timer tests assert no timeout at 94,999 ms, POST abort + same-ID GET at 95,000 ms, adoption of each authoritative status, missing-ID/read-failure unresolved fallback, no new-ID send, and stale late POST/reconcile responses ignored after a newer request sequence.
- Historical Phase 10 terminal tests asserted no retry/thinking, safe localized copy and the then-current `parse(targetJobId)` route. Phase 11 supersedes only the destination; no-retry, no-composer-send and no-raw-error behavior remains regression coverage.
- Pixel-parity Playwright compares formal and prototype surfaces for four states at 1440x900 and 390x844 using DOM snapshot, computed styles, key bounding boxes, overflow and screenshot ratio；scenario screenshots record exact pixel dimensions and SHA-256.

## Phase 11: Safe Markdown/GFM projection and Workspace-detail recovery

- Renderer unit/component tests require one `react-markdown + remark-gfm` projection for persisted user and assistant messages, with `skipHtml` enabled and no `rehypeRaw` dependency or configuration.
- Security cases inject raw HTML, event handlers, Markdown images, `javascript:`/unsafe links and safe links；they prove HTML is inert, remote images do not create a network-fetching `<img>`, unsafe URIs are rejected and safe external links are hardened.
- Payload tests distinguish source from projection: initial send and same-ID retry must receive the byte-identical raw `message.text` and original `clientMessageId`, never rendered DOM text or normalized Markdown, while preserving the next draft.

## Phase 12: Runtime byte guards

- ASCII/multibyte tests use `TextEncoder` with small injected message/session limits；no default-sized strings are constructed.
- Required-field and zero-send/draft-preservation tests compose with backend focused rejection；defaults/overrides remain A4-owned.
- Responsive parity covers headings, lists, blockquotes, inline/fenced code and GFM tables at 1440 and 390；pre/code/table may scroll locally but cannot create document horizontal overflow.
- Terminal route tests require exactly `{name:"workspace", params:{targetJobId}}` / `/workspace?targetJobId=...` read-only detail and reject query-free workspace, `planId`, current-scope `parse(targetJobId)`, row retry, composer send and technical error text.
