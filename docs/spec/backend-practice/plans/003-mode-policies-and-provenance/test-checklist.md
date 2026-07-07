# 003 Test Checklist

> **版本**: 1.4
> **状态**: completed
> **更新日期**: 2026-07-07

**关联 Test Plan**: [test-plan](./test-plan.md)

## Current Test Gates

- [x] Mode dispatch tests cover assisted, strict, unknown mode and goal orthogonality（验证：`TestHandleHintRequestedModeMatrix` / `TestHandleHintRequestedTurnLifecycle`）
- [x] Strict reservation finalization tests prove sanitized 409 replay and no pending event row（验证：`TestAppendSessionEventHintStrictDoesNotLeavePendingReservation`）
- [x] Assisted hint tests prove F3/A3 success, hint_text write, lifecycle invariants and replay snapshot（验证：`TestApplyHintAISuccess`, `TestSQLRepositoryAppendSessionEventWritesHintTextForAssistedSuccess`, `TestSQLRepositoryReserveSessionEventReplaysOriginalHintSnapshot`）
- [x] Task-run contract tests prove `hint_generate` migration/A3 support and typed columns（验证：migration/A3 writer focused tests）
- [x] Provenance tests prove six-field wire JSON and `show_hint` `rubricVersion='not_applicable'`（验证：`TestAssistantActionProvenanceJSONShape`）
- [x] Graceful degrade tests cover F3, A3 and parse failure branches（验证：`TestApplyHintAIGracefulDegradeMatrix`）
- [x] Privacy/redaction tests cover logs, metrics, audit, events and task-run payloads（验证：redaction focused tests + P0.051）
- [x] BDD HTTP scenario tests P0.048-P0.051 pass（验证：`cd backend && go test ./cmd/api -run 'TestE2EP0048|TestE2EP0049|TestE2EP0050|TestE2EP0051' -count=1` PASS）
- [x] Backend-practice non-current lint passes（验证：`python3 scripts/lint/backend_practice_non_current.py --repo-root . --phase all` PASS）
