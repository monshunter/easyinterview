# 002 — Practice Text Event Loop Test Checklist

> **版本**: 1.10
> **状态**: completed
> **更新日期**: 2026-07-07

**关联 Test Plan**: [test-plan](./test-plan.md)

- [x] Screen / component / hook tests cover `PracticeScreen`, `usePracticeSessionLoader`, `usePracticeEvents`, `usePracticeAssistance`, `usePracticeSession`, `useCompletePracticeSession`, AssistantAction rendering, controls, completion and privacy.
- [x] Contract tests prove `appendSessionEvent` and `completePracticeSession` request bodies match generated OpenAPI types and fixture variants.
- [x] Handoff tests prove `generating` params use `resumeId` and stable IDs only.
- [x] Negative tests prove practice runtime does not call `getFeedbackReport`, does not scatter `createPracticeVoiceTurn`, and does not materialize non-current route/module UI.
- [x] P0.044-P0.047 scenario scripts execute focused Vitest files and verify runtime grep markers.
- [x] `make validate-fixtures`, frontend typecheck, docs/index gates and product-scope pruning surface lint pass with current evidence.
