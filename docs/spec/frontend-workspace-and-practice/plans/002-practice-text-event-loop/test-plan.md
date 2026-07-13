# 002 Practice Continuous Conversation Test Plan

> **版本**: 2.5
> **状态**: active
> **更新日期**: 2026-07-13

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
- P0.047 composes frontend disabled-reason evidence with backend `VALIDATION_FAILED`/zero-side-effect evidence, then covers one-answer success and idempotent completion replay.

## Phase 8: reportId-only handoff

- Component/router tests require exact navigation query/state/context `{reportId}` and fail on any copied target/plan/session/resume/round/status/error field.
- P0.047 inspects browser URL/history plus downstream `getFeedbackReport` request, proving reportId is the sole locator and completion replay keeps the same locator.

## Phase 9: Immediate feedback and message-local recovery

- Prototype/formal component tests hold `sendPracticeMessage` deferred and assert synchronous user-row insertion, composer clear/lock, accessible thinking animation and absence of retry before failure.
- Generated-client tests require typed `ApiClientError.apiError.retryable` plus HTTP status, `code/requestId/retryable/details` and transport cause；Practice classification must not inspect `Error.message`, and intentional abort/unmount is not retryable.
- Server-rehydration tables cover `pending` (thinking + composer/Finish lock + single-flight re-read/no duplicate send), `retryable_failed` (one row-local icon), `terminal_failed` (no retry + fact recovery) and `complete` (one user/reply pair) from `getPracticeSession.clientMessageId + replyStatus`, with no URL/browser-storage retry identity.
- Failure/retry tables assert row-local icon placement, server original text + same-ID payload, next-draft preservation, retry lock, repeated-failure stability and server-adopt deduplication；transport failures without a response or typed `ApiClientError.apiError.retryable=true` are retryable, while validation/auth/not-found/conflict/mismatch have no retry icon and recover through server truth.
- The recovery integration path is explicit: AI failure → page reload → `getPracticeSession` returns the same user text/clientMessageId as `retryable_failed` → retry → exactly one assistant reply and no duplicate user row.
- Finish-state tests cover pending, retryable-failed, retrying and terminal-recovery in addition to existing loader/completion guards；no unresolved message state can enable completion.
- Pixel parity uses identical pending/failed fixtures at 1440 and 390, comparing DOM, computed style, key bounding boxes, viewport overflow and screenshot diff; P0.044/P0.046 record current runtime evidence.
