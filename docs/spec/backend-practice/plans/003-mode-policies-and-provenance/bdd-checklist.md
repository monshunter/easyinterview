# 003 BDD Checklist

> **版本**: 1.7
> **状态**: completed
> **更新日期**: 2026-07-11

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## E2E.P0.048 assisted hint 主路径 × goal 矩阵

- [x] Scenario uses three current goals with `mode='assisted'`
- [x] Trigger covers first hint, second hint on same turn and replay of first `clientEventId`
- [x] Verify covers 200 `show_hint`, six-field provenance, hint_text write, replay snapshot, no turn/outbox/audit side effect and success `ai_task_runs(hint_generate)`
- [x] Go HTTP scenario `TestE2EP0048PracticeHintAssistedAcrossGoals` passes

## E2E.P0.049 strict-mode optional hint × goal 矩阵

- [x] Scenario uses three current goals with `mode='strict'`
- [x] Trigger covers `hint_requested` plus replay
- [x] Verify covers 200 `show_hint`, replay snapshot, no pending reservation, hint_text write, no turn-completed outbox and no AI call inside the reservation transaction
- [x] Go HTTP scenario `TestE2EP0049PracticeHintOptionalAcrossStrictModeGoals` passes

## E2E.P0.050 AssistantAction provenance wire 边界

- [x] Scenario triggers AI-backed and non-AI AssistantAction types
- [x] Verify proves provenance JSON key set is exactly the six B2 fields
- [x] Verify proves runtime fields stay out of wire response and typed task-run rows reflect actual AI calls
- [x] Go HTTP scenario `TestE2EP0050PracticeAssistantActionProvenanceAndTaskRuns` passes

## E2E.P0.051 graceful degrade + privacy

- [x] Scenario injects F3, A3 and parser failure branches
- [x] Verify proves 200 `session_wait`, session running, no failure_code and failed `ai_task_runs(hint_generate)` rows
- [x] Verify proves privacy redlines over log, metric, audit, event and task-run payloads
- [x] Verify runs backend-practice runtime boundary lint
- [x] Go HTTP scenario `TestE2EP0051PracticeHintDegradeAndPrivacy` passes
- [x] Phase 7 revision injects zh-CN English cue and en Han cue, then proves `session_wait`, no `hint_text`, unchanged session/turn and no raw cue in response/replay/task-run evidence.
  <!-- verified: 2026-07-11 evidence="P0.051 wrong-language-zh/wrong-language-en variants PASS and assert session_wait, unchanged session/turn, no hint_text and no raw cue in response/replay/task-run evidence." -->

## Closeout

- [x] `cd backend && go test ./cmd/api -run 'TestE2EP0048|TestE2EP0049|TestE2EP0050|TestE2EP0051' -count=1` passes
- [x] `python3 scripts/lint/backend_practice_out_of_scope.py --repo-root . --phase all` passes
- [x] `python3 -m pytest scripts/lint/backend_practice_out_of_scope_test.py -q` passes
- [x] Current P0.051 and owner closeout gates pass after the language-integrity revision.
  <!-- verified: 2026-07-11 evidence="P0.051 setup-trigger-verify-cleanup PASS; P0.048/P0.050 wrappers and direct P0.049 PASS; full backend, privacy/runtime-boundary, fixture/codegen, context/docs/index/diff gates PASS." -->
