# 002 — Practice Text Event Loop Test Checklist

> **版本**: 1.14
> **状态**: active
> **更新日期**: 2026-07-10

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
