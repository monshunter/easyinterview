# 002 — Practice Text Event Loop BDD Checklist

> **版本**: 1.12
> **状态**: active
> **更新日期**: 2026-07-10

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## E2E.P0.044

- [x] Scenario directory exists with README, seed input, expected outcome and setup/trigger/verify/cleanup scripts.
- [x] Trigger runs real-mode owner marker plus practice screen/event/action/body tests updated for no independent side-panel controls, no dictation, no skip, no role switch and no visible strict switch.
- [x] Verify checks passing Vitest markers, current runtime testids, no prototype data helpers, no `getFeedbackReport` in practice runtime, no out-of-scope anchors and phone turn confined to owner hook.
  <!-- verified: 2026-07-09 command="./test/scenarios/e2e/p0-044-practice-text-loop-assisted-happy-path/scripts/setup.sh && ./test/scenarios/e2e/p0-044-practice-text-loop-assisted-happy-path/scripts/trigger.sh && ./test/scenarios/e2e/p0-044-practice-text-loop-assisted-happy-path/scripts/verify.sh" result="PASS" -->

## E2E.P0.045

- [x] Scenario directory exists with README, seed input, expected outcome and scripts.
- [x] Trigger runs goal parity, optional hint, pause/resume, phone captions, hang-up/restart and mode switch tests.
- [x] Verify checks current goal visibility parity, append idempotency boundary, phone-mode user-visible labels, out-of-scope control absence and forbidden input markers.
  <!-- verified: 2026-07-09 command="./test/scenarios/e2e/p0-045-practice-text-loop-mode-policy-display/scripts/setup.sh && ./test/scenarios/e2e/p0-045-practice-text-loop-mode-policy-display/scripts/trigger.sh && ./test/scenarios/e2e/p0-045-practice-text-loop-mode-policy-display/scripts/verify.sh" result="PASS" -->

## E2E.P0.046

- [x] Scenario directory exists with README, seed input, expected outcome and scripts.
- [x] Trigger runs error, lost-state, conflict, retry and completion hook tests.
- [x] Verify checks timeout recovery, 404 lost state, 409 mismatch refresh, hint retry/recovery, absence of strict-mode conflict UI and privacy redlines.
  <!-- verified: 2026-07-09 command="./test/scenarios/e2e/p0-046-practice-text-loop-failure-and-recovery/scripts/setup.sh && ./test/scenarios/e2e/p0-046-practice-text-loop-failure-and-recovery/scripts/trigger.sh && ./test/scenarios/e2e/p0-046-practice-text-loop-failure-and-recovery/scripts/verify.sh" result="PASS" -->

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
