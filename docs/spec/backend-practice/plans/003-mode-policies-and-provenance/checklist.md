# Backend Practice Mode Policies and Provenance Checklist

> **版本**: 1.5
> **状态**: completed
> **更新日期**: 2026-07-09

**关联计划**: [plan](./plan.md)

## Phase 0: contract preflight

- [x] 0.1 backend-practice spec contains current mode, graceful degrade, hint lifecycle and provenance decisions（验证：`docs/spec/backend-practice/spec.md` v1.16 active；C-7/C-8/C-12/C-17 still map to this plan）
- [x] 0.2 B4 baseline and A3 task-run writer support `hint_generate`（验证：migration/A3 writer focused tests are listed in [test-plan](./test-plan.md) and pass in owner closeout）
- [x] 0.3 F3 `practice.turn.lightweight_observe` preflight is executable and model profile resolves（验证：`backend/internal/ai/registry/backend_practice_preflight_test.go` owner evidence）
- [x] 0.4 `appendSessionEvent` fixtures include assisted success, legacy strict success and assisted degrade variants（验证：`make validate-fixtures` PASS in owner closeout）

## Phase 1: optional hint dispatch and legacy strict compatibility

- [x] 1.1 `handleHintRequested` keeps hint optional across current goals and legacy modes（验证：`TestSessionEventServiceRouteCoversAllKinds` + legacy strict service tests）
- [x] 1.2 legacy strict returns `show_hint`, calls AI outside the reservation transaction, persists `hint_text`, and replays without pending rows（验证：`TestAppendSessionEventHintLegacyStrictRunsAIAndAppends`, `TestServiceAppliesHintAIForLegacyStrict`, `TestE2EP0049PracticeHintOptionalAcrossLegacyStrictGoals`）
- [x] 1.3 BDD-Gate: `E2E.P0.049` legacy strict optional hint across goals is covered（验证：`cd backend && go test ./cmd/api -run TestE2EP0049PracticeHintOptionalAcrossLegacyStrictGoals -count=1` PASS）

## Phase 2: assisted hint AI and persistence

- [x] 2.1 assisted hint uses F3 `practice.turn.lightweight_observe` and A3 observed AIClient with task capability `hint_generate`（验证：`TestApplyHintAISuccess`, `TestServiceAppliesHintAIForAssisted`）
- [x] 2.2 assisted success writes `practice_turns.hint_text` and keeps turn/session lifecycle unchanged（验证：`TestSQLRepositoryAppendSessionEventWritesHintTextForAssistedSuccess`, `TestE2EP0048PracticeHintAssistedAcrossGoals`）
- [x] 2.3 replay preserves the original hint response snapshot after later hints on the same turn（验证：`TestSQLRepositoryReserveSessionEventReplaysOriginalHintSnapshot`, `TestE2EP0048PracticeHintAssistedAcrossGoals`）
- [x] 2.4 BDD-Gate: `E2E.P0.048` assisted hint across goals is covered（验证：`cd backend && go test ./cmd/api -run TestE2EP0048PracticeHintAssistedAcrossGoals -count=1` PASS）

## Phase 3: provenance and graceful degrade

- [x] 3.1 AssistantAction provenance serializes exactly the B2 wire fields for show_hint / ask_question / ask_follow_up / session_wait / session_completed（验证：`TestAssistantActionProvenanceJSONShape`, `TestE2EP0050PracticeAssistantActionProvenanceAndTaskRuns`）
- [x] 3.2 F3/A3/parser hint failures degrade to `session_wait`, keep session running, and write failed `ai_task_runs(hint_generate)` where required（验证：`TestApplyHintAIGracefulDegradeMatrix`, `TestE2EP0051PracticeHintDegradeAndPrivacy`）
- [x] 3.3 BDD-Gate: `E2E.P0.050` provenance/task-run boundary is covered（验证：`cd backend && go test ./cmd/api -run TestE2EP0050PracticeAssistantActionProvenanceAndTaskRuns -count=1` PASS）
- [x] 3.4 BDD-Gate: `E2E.P0.051` graceful degrade main path is covered（验证：`cd backend && go test ./cmd/api -run TestE2EP0051PracticeHintDegradeAndPrivacy -count=1` PASS）

## Phase 4: privacy, observability and closeout

- [x] 4.1 hint path redaction covers logs, metrics, audit, event payload and typed task-run payloads（验证：`TestApplyHintAIPrivacyRedaction`, `TestPracticeObservedAIRedactsPromptResponseFromLogsMetricsAndAudit`, P0.051）
- [x] 4.2 backend-practice runtime boundary lint rejects removed mode/goal/route vocabulary outside negative gates（验证：`python3 scripts/lint/backend_practice_non_current.py --repo-root . --phase all` PASS）
- [x] 4.3 BDD-Gate: P0.048-P0.051 HTTP scenario suite is covered（验证：`cd backend && go test ./cmd/api -run 'TestE2EP0048|TestE2EP0049|TestE2EP0050|TestE2EP0051' -count=1` PASS）
- [x] 4.4 Owner docs/index/context are current and completed（验证：`validate_context.py backend-practice/003 backend` PASS；`sync-doc-index --check` PASS；`make docs-check` PASS）
