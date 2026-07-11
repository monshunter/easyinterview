# 002 — Practice Text Event Loop BDD Checklist

> **版本**: 1.15
> **状态**: active
> **更新日期**: 2026-07-11

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## E2E.P0.044

- [x] Scenario directory exists with README, seed input, expected outcome and setup/trigger/verify/cleanup scripts.
- [x] Trigger runs real-mode owner marker plus practice screen/event/action/body tests updated for no independent side-panel controls, no dictation, no skip, no role switch and no visible strict switch.
- [x] Verify checks passing Vitest markers, current runtime testids, no prototype data helpers, no `getFeedbackReport` in practice runtime, no out-of-scope anchors and phone turn confined to owner hook.
  <!-- verified: 2026-07-09 command="./test/scenarios/e2e/p0-044-practice-text-loop-assisted-happy-path/scripts/setup.sh && ./test/scenarios/e2e/p0-044-practice-text-loop-assisted-happy-path/scripts/trigger.sh && ./test/scenarios/e2e/p0-044-practice-text-loop-assisted-happy-path/scripts/verify.sh" result="PASS" -->

## E2E.P0.045

- [x] Scenario directory exists with README, seed input, expected outcome and scripts.
- [x] Historical Phase 6 trigger covered the then-current goal parity, hint, pause/resume and phone controls; its restart expectation is superseded and is not current acceptance evidence.
- [x] Verify checks current goal visibility parity, append idempotency boundary, phone-mode user-visible labels, out-of-scope control absence and forbidden input markers.
  <!-- verified: 2026-07-09 command="./test/scenarios/e2e/p0-045-practice-text-loop-mode-policy-display/scripts/setup.sh && ./test/scenarios/e2e/p0-045-practice-text-loop-mode-policy-display/scripts/trigger.sh && ./test/scenarios/e2e/p0-045-practice-text-loop-mode-policy-display/scripts/verify.sh" result="PASS" -->
- [x] Trigger/verify use rendered goal/hint/mode tests only and do not depend on a constant-only assistance hook test.
  <!-- verified: 2026-07-10 method=constant-assistance-hook-removal evidence="P0.045 setup/trigger/verify/cleanup passes the real-mode gate plus 6 rendered test files/18 tests; trigger log and verify no longer reference the deleted hook test." -->
- [x] Phase 10 revision verifies one handset icon, shared same-session hang-up, immediate microphone/TTS stop, no later phone TTS, and zero segmented/live/cut-off/restart/callEnded controls.
- [x] Phase 10 revision verifies generated `getTargetJob` company/title and zero fixture transcript / raw `questionIntent` in real mode.
  <!-- verified: 2026-07-11 evidence="P0.045 PASS; 16 focused frontend files/84 tests cover shared exit and lifecycle. Real API browser loaded GET session + GET targetJob, rendered 真实联调测试公司/资深后端平台工程师 and the Chinese persisted question, with no raw questionIntent or fixture transcript." -->

## E2E.P0.046

- [x] Scenario directory exists with README, seed input, expected outcome and scripts.
- [x] Trigger runs error, lost-state, conflict, retry and completion hook tests.
- [x] Verify checks timeout recovery, 404 lost state, 409 mismatch refresh, hint retry/recovery, absence of strict-mode conflict UI and privacy redlines.
  <!-- verified: 2026-07-09 command="./test/scenarios/e2e/p0-046-practice-text-loop-failure-and-recovery/scripts/setup.sh && ./test/scenarios/e2e/p0-046-practice-text-loop-failure-and-recovery/scripts/trigger.sh && ./test/scenarios/e2e/p0-046-practice-text-loop-failure-and-recovery/scripts/verify.sh" result="PASS" -->
- [x] Phase 10 revision verifies text repair failure renders `session_wait`, retains the answer, avoids duplicate transcript and retries with a new `clientEventId`; voice repair failure renders the existing top-level typed error and permits text-mode exit; both remain in the same session and no canned question appears.
  <!-- verified: 2026-07-11 evidence="P0.046 PASS; focused frontend recovery tests and backend state-preserving failure tests cover retained input/new clientEventId, no duplicate transcript/canned question, typed AI_OUTPUT_INVALID and same-session text exit." -->

## E2E.P0.047

- [x] Scenario directory exists with README, seed input, expected outcome and scripts.
- [x] Trigger runs complete hook, handoff params, body parity, completion UI and privacy tests.
- [x] Verify checks `completePracticeSession` body/header contract, `resumeId` handoff, no `getFeedbackReport` practice call, and voice hook boundary.
  <!-- verified: 2026-07-09 command="./test/scenarios/e2e/p0-047-practice-text-loop-complete-and-generating-handoff/scripts/setup.sh && ./test/scenarios/e2e/p0-047-practice-text-loop-complete-and-generating-handoff/scripts/trigger.sh && ./test/scenarios/e2e/p0-047-practice-text-loop-complete-and-generating-handoff/scripts/verify.sh" result="PASS" -->

## Regression

- [x] Workspace owner P0.018/P0.021 remains compatible with practice route handoff.
- [x] Backend-practice and OpenAPI fixture gates remain compatible with the current generated-client contract.
- [x] Docs, index, diff whitespace and product-scope pruning surface gates pass after owner compression.

## REAL.ENV.SCREENSHOT

- [x] Verify local dev dependencies/backend/frontend are running from the current branch.
- [x] Open real text practice flow in browser and capture screenshot evidence under `.test-output/`.
- [x] Open real phone practice flow in browser and capture screenshot evidence under `.test-output/`.
- [x] Verify screenshots show no independent side-panel controls, speech-to-text, skip, role switch, visible strict switch or voice-analysis panels.
  <!-- verified: 2026-07-09 artifacts=".test-output/real-practice-session/practice-text-real-closed-loop.png .test-output/real-practice-session/practice-phone-real-default.png .test-output/real-practice-session/practice-phone-real-captions.png .test-output/real-practice-session/practice-generating-real-stable.png" result="PASS" -->
- [x] Refresh real-mode screenshots/runtime evidence for real company/title, one handset icon, red circular hang-up and same-session return to text; removed controls and mock/internal content remain absent.
  <!-- verified: 2026-07-11 artifacts=".test-output/practice-phone-session-flow-0711/practice-text-real.png .test-output/practice-phone-session-flow-0711/practice-phone-real.png .test-output/practice-phone-session-flow-0711/practice-after-hangup-real.png" evidence="All screenshots are 1440x900; URL sessionId is unchanged across phone entry and center hang-up, and the database session remains running with the same asked turn." -->
