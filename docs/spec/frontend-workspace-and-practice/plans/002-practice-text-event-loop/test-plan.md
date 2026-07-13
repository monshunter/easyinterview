# 002 Practice Continuous Conversation Test Plan

> **版本**: 2.4
> **状态**: active
> **更新日期**: 2026-07-12

## Phase 1: Prototype/source
- DOM shape, disabled phone control, no stale question/hint/phone-positive source, desktop/mobile geometry.
## Phase 2: Formal DOM
- TopBar props, component deletion, a11y, responsive layout, context/i18n negative checks.
## Phase 3: Hooks/state
- Ordered messages, refresh, send/replay/failure/retry/no duplicates and UI states.
## Phase 4: Completion
- Stable-ID handoff and generating conversation copy.
## Phase 5: Integration/parity
- Vitest/typecheck/build, source/pixel parity, real P0.099 screenshots.

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
