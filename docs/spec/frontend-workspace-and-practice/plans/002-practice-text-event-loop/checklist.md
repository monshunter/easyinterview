# 002 — Practice Continuous Text Conversation Checklist

> **版本**: 2.0
> **状态**: active
> **更新日期**: 2026-07-12

**关联计划**: [plan](./plan.md)

## Phase 1: UI truth source
- [ ] 1.1 RED-GREEN: update prototype tests/source to TopBar + full-width Conversation and delete question/hint/phone-positive source.
- [ ] 1.2 RED-GREEN: update desktop/mobile source geometry expectations and stale-contract negative checks.

## Phase 2: Formal screen structure
- [ ] 2.1 RED-GREEN: delete SessionMap/QuestionCard/PhoneSurface/hint/controller components and simplify PracticeScreen/TopBar.
- [ ] 2.2 RED-GREEN: disabled phone icon has native disabled/a11y/unavailable copy and no route/API action.
- [ ] 2.3 RED-GREEN: remove mode/modality/practiceMode/hint context/handoff/i18n/test contracts.
- [ ] 2.4 BDD-Gate: P0.045 simplified UI and phone-disabled scenario passes.

## Phase 3: Message hooks and states
- [ ] 3.1 RED-GREEN: loader renders ordered session.messages including refresh recovery.
- [ ] 3.2 RED-GREEN: send hook handles success/replay/failure/same-ID retry without duplicate messages.
- [ ] 3.3 RED-GREEN: loading/sending/error/local-paused/completing/session-lost states remain usable; pause has no backend event call and refresh resumes Running.
- [ ] 3.4 BDD-Gate: P0.044 and P0.046 pass.

## Phase 4: Completion/generating
- [ ] 4.1 RED-GREEN: finish handoff contains stable IDs only; generating copy is conversation-level.
- [ ] 4.2 BDD-Gate: P0.047 passes.

## Phase 5: Parity and real scenario
- [ ] 5.1 Run focused/full frontend, typecheck/build, UI contract and pixel parity desktop/mobile.
- [ ] 5.2 Run real backend/frontend P0.099 path and capture redacted conversation/report screenshots.
- [ ] 5.3 BDD-Gate: P0.099 real fullstack screenshot evidence passes.
