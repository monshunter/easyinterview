# 003 Test Checklist

> **版本**: 1.7
> **状态**: completed
> **更新日期**: 2026-07-11

**关联 Test Plan**: [test-plan](./test-plan.md)

## Current Test Gates

- [x] Mode dispatch tests cover assisted, strict mode and goal orthogonality（验证：`TestServiceAppliesHintAIForAssisted` / `TestServiceAppliesHintAIForStrictMode`）
- [x] Strict-mode replay tests prove `show_hint` replay and no pending event row（验证：`TestE2EP0049PracticeHintOptionalAcrossStrictModeGoals`）
- [x] Assisted hint tests prove F3/A3 success, hint_text write, lifecycle invariants and replay snapshot（验证：`TestApplyHintAISuccess`, `TestSQLRepositoryAppendSessionEventWritesHintTextForAssistedSuccess`, `TestSQLRepositoryReserveSessionEventReplaysOriginalHintSnapshot`）
- [x] Task-run contract tests prove `hint_generate` migration/A3 support and typed columns（验证：migration/A3 writer focused tests）
- [x] Provenance tests prove six-field wire JSON and `show_hint` `rubricVersion='not_applicable'`（验证：`TestAssistantActionProvenanceJSONShape`）
- [x] Graceful degrade tests cover F3, A3 and parse failure branches（验证：`TestApplyHintAIGracefulDegradeMatrix`）
- [x] Privacy/redaction tests cover logs, metrics, audit, events and task-run payloads（验证：redaction focused tests + P0.051）
- [x] BDD HTTP scenario tests P0.048-P0.051 pass（验证：`cd backend && go test ./cmd/api -run 'TestE2EP0048|TestE2EP0049|TestE2EP0050|TestE2EP0051' -count=1` PASS）
- [x] Backend-practice out-of-scope lint passes（验证：`python3 scripts/lint/backend_practice_out_of_scope.py --repo-root . --phase all` PASS）
- [x] Current language gate proves zh-CN English cue and en Han cue degrade to `session_wait`, do not persist `hint_text`, and preserve session/turn state.
- [x] Current P0.051 plus focused/full backend, privacy, fixture/codegen, context/docs/index and diff gates pass.
  <!-- verified: 2026-07-11 evidence="P0.051 wrong-language-zh/wrong-language-en variants and focused service/store/privacy tests PASS; full backend, fixtures/codegen, four contexts, docs/index and diff gates PASS." -->
