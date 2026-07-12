# 002 — Practice Continuous Text Conversation Checklist

> **版本**: 2.1
> **状态**: completed
> **更新日期**: 2026-07-12

**关联计划**: [plan](./plan.md)

## Phase 1: UI truth source
- [x] 1.1 RED-GREEN: update prototype tests/source to TopBar + full-width Conversation and delete question/hint/phone-positive source.
- [x] 1.2 RED-GREEN: update desktop/mobile source geometry expectations and stale-contract negative checks.

## Phase 2: Formal screen structure
- [x] 2.1 RED-GREEN: delete SessionMap/QuestionCard/PhoneSurface/hint/controller components and simplify PracticeScreen/TopBar.
- [x] 2.2 RED-GREEN: disabled phone icon has native disabled/a11y/unavailable copy and no route/API action.
- [x] 2.3 RED-GREEN: remove mode/modality/practiceMode/hint context/handoff/i18n/test contracts.
- [x] 2.4 BDD-Gate: P0.045 simplified UI and phone-disabled scenario passes.

## Phase 3: Message hooks and states
- [x] 3.1 RED-GREEN: loader renders ordered session.messages including refresh recovery.
- [x] 3.2 RED-GREEN: send hook handles success/replay/failure/same-ID retry without duplicate messages.
- [x] 3.3 RED-GREEN: loading/sending/error/local-paused/completing/session-lost states remain usable; pause has no backend event call and refresh resumes Running.
- [x] 3.4 BDD-Gate: P0.044 and P0.046 pass.

## Phase 4: Completion/generating
- [x] 4.1 RED-GREEN: finish handoff contains stable IDs only; generating copy is conversation-level.
- [x] 4.2 BDD-Gate: P0.047 passes.

## Phase 5: Parity and real scenario
- [x] 5.1 Run focused/full frontend, typecheck/build, UI contract and pixel parity desktop/mobile.
  <!-- verified: 2026-07-12 method=full-frontend-and-parity evidence="111 files/708 Vitest tests, typecheck, build, 45 UI contracts and 8 desktop/mobile practice Playwright cases pass" -->
- [x] 5.2 Run real backend/frontend P0.099 path and capture redacted conversation/report screenshots.
- [x] 5.3 BDD-Gate: P0.099 real fullstack screenshot evidence passes.

## Phase 6: Review remediation
- [x] 6.1 RED-GREEN: PracticeScreen retries loader, message and completion failures through the correct operation and preserves message/completion idempotency. (`pnpm --filter @easyinterview/frontend test src/app/screens/practice/PracticeScreen.test.tsx`)
- [x] 6.2 RED-GREEN: Finish CTA is disabled during send, load, completion and non-mutable session states. (`pnpm --filter @easyinterview/frontend test src/app/screens/practice/PracticeScreen.test.tsx src/app/screens/practice/hooks/useCompletePracticeSession.test.tsx`; frontend typecheck)
- [x] 6.3 BDD-Gate: P0.046 and P0.047 screen-level failure/recovery and completion scenarios pass. (serial `setup.sh` → `trigger.sh` → `verify.sh` → `cleanup.sh`, both PASS)
