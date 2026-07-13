# 002 — Practice Continuous Text Conversation Checklist

> **版本**: 2.4
> **状态**: active
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

> Ownership note (2026-07-12): the completed evidence above is historical. Current work stops at stable `reportId` handoff; GeneratingScreen is exclusively owned by `frontend-report-dashboard/001`.

## Phase 5: Parity and real scenario
- [x] 5.1 Run focused/full frontend, typecheck/build, UI contract and pixel parity desktop/mobile.
  <!-- verified: 2026-07-12 method=full-frontend-and-parity evidence="111 files/708 Vitest tests, typecheck, build, 45 UI contracts and 8 desktop/mobile practice Playwright cases pass" -->
- [x] 5.2 Run real backend/frontend P0.099 path and capture redacted conversation/report screenshots.
- [x] 5.3 BDD-Gate: P0.099 real fullstack screenshot evidence passes.

## Phase 6: Review remediation
- [x] 6.1 RED-GREEN: PracticeScreen retries loader, message and completion failures through the correct operation and preserves message/completion idempotency. (`pnpm --filter @easyinterview/frontend test src/app/screens/practice/PracticeScreen.test.tsx`)
- [x] 6.2 RED-GREEN: Finish CTA is disabled during send, load, completion and non-mutable session states. (`pnpm --filter @easyinterview/frontend test src/app/screens/practice/PracticeScreen.test.tsx src/app/screens/practice/hooks/useCompletePracticeSession.test.tsx`; frontend typecheck)
- [x] 6.3 BDD-Gate: P0.046 and P0.047 screen-level failure/recovery and completion scenarios pass. (serial `setup.sh` → `trigger.sh` → `verify.sh` → `cleanup.sh`, both PASS)

## Phase 7: Zero-answer finish eligibility and backend authority

- [x] 7.1 RED-GREEN: PracticeScreen derives Finish eligibility only from server-loaded committed candidate `user` messages plus existing mutable/no-pending-reply/no-load/no-send/no-complete guards; opening assistant, composer draft and route params do not count. (`PracticeScreen.test.tsx` + completion hook tests)
  <!-- verified: 2026-07-12 method=focused-and-full-vitest evidence="PracticeScreen 8/8 PASS; related practice regression 24/24 PASS; full frontend 111 files/762 tests PASS" -->
- [x] 7.2 RED-GREEN: prototype and formal Finish are native disabled at zero answers and expose the same nearby zh/en reason through stable `aria-describedby`; first committed user message removes only the zero-answer reason. (ui-design source contract + i18n exact-set + DOM/a11y tests)
  <!-- verified: 2026-07-12 method=source-contract-dom-a11y-i18n evidence="ui-design contract 50/50 PASS; PracticeScreen and locale exact-set tests included in full Vitest PASS" -->
- [ ] 7.3 RED-GREEN: direct zero-answer completion is still rejected by backend `VALIDATION_FAILED`, session remains mutable and no report/job/outbox/idempotency success is written; one-answer completion and replay remain green. (consume backend-practice/002 Phase 9 service/store/API/PostgreSQL markers; do not duplicate backend logic in frontend)
- [ ] 7.4 BDD-Gate: E2E.P0.047 composes `ZERO_ANSWER_FINISH_DISABLED_PASS` + `ZERO_ANSWER_COMPLETION_REJECTED_PASS`, then proves one-answer completion, stable reportId handoff and exact replay.

## Phase 8: reportId-only completion handoff

- [x] 8.1 RED-GREEN: PracticeScreen completion navigation has exact query/state/context shape `{reportId}`; tests first fail on and then reject `targetJobId|planId|sessionId|resumeId|roundId|roundName|status|error` copies while preserving same-reportId completion replay.
  <!-- verified: 2026-07-12 method=screen-router-and-privacy-tests evidence="PracticeScreen, App, routeUrl, pendingAction and routing privacy cases included in 111-file/762-test PASS" -->
- [x] 8.2 REGRESSION-GATE: active PracticeScreen/context/router code contains no positive write of those copied fields to generating/report navigation; frontend-report consumes `getFeedbackReport(reportId)` as the sole downstream authority.
  <!-- verified: 2026-07-12 method=active-negative-and-route-tamper evidence="report/generating out-of-scope tests PASS; Playwright canonicalizes hostile report/generating URLs to reportId only" -->
- [ ] 8.3 BDD-Gate: E2E.P0.047 one-answer completion asserts URL/history state contains only reportId, downstream request is keyed only by reportId, and idempotent replay returns the same locator.
