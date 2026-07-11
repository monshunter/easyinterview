# 002 — Practice Text Event Loop Test Checklist

> **版本**: 1.15
> **状态**: active
> **更新日期**: 2026-07-11

**关联 Test Plan**: [test-plan](./test-plan.md)

- [x] Screen / component / hook tests cover `PracticeScreen`, `usePracticeSessionLoader`, `usePracticeEvents`, `usePracticeSession`, `useCompletePracticeSession`, AssistantAction rendering, rendered goal/hint policy, text controls, phone controls, completion and privacy under the current real-interview UI.
- [x] Contract tests prove `appendSessionEvent` and `completePracticeSession` request bodies match generated OpenAPI types and fixture variants, with no positive `turn_skipped` path.
- [x] Handoff tests prove `generating` params use `resumeId` and stable IDs only, and phone mode displays as `Phone / 电话模式`.
- [x] Negative tests prove practice runtime does not call `getFeedbackReport`, does not scatter `createPracticeVoiceTurn`, and does not materialize out-of-scope controls: independent side panel, dictation, skip, role switch, visible strict switch, voice analysis, manual transcript fallback UI.
- [x] P0.044-P0.047 scenario scripts execute focused Vitest files and verify updated runtime grep markers.
  <!-- verified: 2026-07-09 commands="P0.044/P0.045/P0.046/P0.047 setup -> trigger -> verify" result="PASS" -->
- [x] Real local environment browser screenshots for text and phone practice flows are captured under `.test-output/` and show current UI boundaries.
  <!-- verified: 2026-07-09 artifact=".test-output/real-practice-session/real-browser-closed-loop-result.json" result="PASS" -->
- [x] `make validate-fixtures`, frontend typecheck, docs/index gates and product-scope pruning surface lint pass with current evidence.
- [x] Phase 10 tests prove one handset icon and one center hang-up path, with segmented/live/cut-off/restart/callEnded controls absent across desktop/mobile parity.
- [x] Phase 10 tests prove shared `exitPhoneMode` stops microphone/TTS, settles non-empty capture without later phone TTS, emits no hang-up barge-in and returns to text for the same session.
- [x] Phase 10 generated-client/real-mode tests prove server-session-first `getTargetJob` supplies company/title, fixture dialogue and raw `questionIntent` do not render, text `session_wait` retains input without duplicate transcript and retries with a new `clientEventId`, and the existing top-level voice error preserves the same session with text-mode exit.
- [x] Updated P0.045/P0.046 wrappers, focused/full frontend tests, typecheck/build, pixel parity, browser screenshot, context/docs/index/diff gates pass.
  <!-- verified: 2026-07-11 evidence="P0.045/P0.046 PASS; frontend 144 files/889 tests, typecheck/build, Practice parity 11 pass/1 conditional skip, four owner contexts, docs/index/diff and three real API 1440x900 screenshots PASS." -->
