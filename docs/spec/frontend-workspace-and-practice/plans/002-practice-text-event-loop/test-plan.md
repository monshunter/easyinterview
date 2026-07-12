# 002 Practice Continuous Conversation Test Plan

> **版本**: 2.1
> **状态**: completed
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
