# 003 — Mode Policies and Provenance Test Checklist

> **版本**: 1.0
> **状态**: completed
> **更新日期**: 2026-05-14

**关联 Test Plan**: [test-plan](./test-plan.md)

## Phase 0: 跨 spec 前置修订 + Preflight

- [x] Phase 0 本计划定义的单元测试项全部通过：
  - `TestBaselineAITaskRunsTaskTypeCheckIncludesHintGenerate`（migration contract test，位于 `backend/internal/migrations/`；命令入口 `cd backend && go test ./internal/migrations/... -count=1`，覆盖 `hint_generate` 存在且 unknown 未被加入 CHECK）
  - `migrations/lint.sh`（`migrations/enum-sources.yaml` 与 SQL CHECK 同步；漏改 enum manifest 时必须失败）
  - `TestAITaskRunCapabilityIncludesHintGenerate`（writers.go enum 与 `allowedAITaskRunCapabilities` 集合一致性 + `Validate()` 接受 `hint_generate`）
  - `TestBackendPracticeF3Preflight`（F3 baseline + ResolveActive 可用性，扩展 `backend/internal/ai/registry/backend_practice_preflight_test.go`）
  - `make validate-fixtures` + `make codegen-check` + `python3 scripts/lint/conventions_drift.py --repo-root .` 通过

## Phase 1: handleHintRequested mode-binding & strict 边界覆盖

- [x] Phase 1 本计划定义的单元测试项全部通过：
  - `TestHandleHintRequestedModeMatrix`（assisted × 4 goal + strict × 4 goal + unknown × 2 fallback；strict 路径 fake registry 调用次数为 0）
  - `TestHandleHintRequestedTurnLifecycle`（OutboxRecord nil / NextTurn 未推进 / session 状态不变 / AuditMetadata 仅含 event_kind+mode）
  - `TestServiceSkipsHintAIForStrict`（strict 409 路径不触发 fake registry / fake AIClient）
  - `TestAppendSessionEventHintStrictDoesNotLeavePendingReservation`（strict / unknown 409 后同 `(session_id, client_event_id)` 存在 finalized `kind='hint_requested'` 行，payload 仅含 sanitized conflict envelope，不存在 `payload.pending=true` stuck row；同 clientEventId 重试 replay 同一 409）

## Phase 2: assisted hint AI 接入 + practice_turns.hint_text 写入

- [x] Phase 2 本计划定义的单元测试项全部通过：
  - `TestApplyHintAISuccess`（fake F3 + fake AIClient 成功路径 → outcome.Hint 非空 + Provenance 6 wire 字段）
  - `TestApplyHintAIBuildsPromptWithoutLeaks`（prompt user message 不拼接 question_text / answer_text 明文）
  - `TestApplyHintAISuccess`（fake AIClient payload 捕获 capability=`hint_generate`、resource_type=`target_job`、resource_id=plan.target_job_id）
  - `TestServiceAppliesHintAIForAssisted` / `TestServiceSkipsHintAIForStrict`（service layer 接入；hint 与 follow_up 由 action type 分支互斥）
  - `TestSQLRepositoryAppendSessionEventWritesHintTextForAssistedSuccess`（repository UPDATE practice_turns.hint_text；SQL expectation 同时固定 hint 路径不发 outbox、不写 audit、不执行 turn status UPDATE，返回 session turn_count 不变）
  - `TestAssistantActionProvenanceFieldsAreWireOnly`（reflect + struct tag）
  - `TestAssistantActionProvenanceJSONShape`（json.Marshal → 6 keys 严格等于 B2 GenerationProvenance）
  - `TestAssistantActionProvenanceCrossActionParity`（5 种 action type 的 provenance 字段集严格一致）

## Phase 3: AI failure graceful degrade & provenance wire boundary

- [x] Phase 3 本计划定义的单元测试项全部通过：
  - `TestApplyHintAIGracefulDegradeMatrix`（表驱动 7 路径：F3 `registry.ErrPromptUnsupported` / `registry.ErrLanguageUnsupported` / A3 secret_missing / A3 timeout / A3 invalid output / A3 capability mismatch / parsed hint empty）
  - `TestApplyHintAIGracefulDegradeMatrix`（覆盖 F3 / A3 / parse 失败；F3 与 parsed hint empty 路径显式调用 `aiclient.AITaskRunWriter.WriteAITaskRun` 写 failed row + B1 error_code）
  - `TestApplyHintAIPrivacyRedaction`（log fields + service-local AuditMetadata 不含 hint_text / prompt body / response body / provider secret；AuditMetadata 不进入 `audit_events` 或 wire）
  - `TestApplyHintAIRejectsCapabilityMismatch`（A3 capability 错配走 degrade 路径）
  - `TestE2EP0048PracticeHintAssistedAcrossGoals`（HTTP scenario 200 + show_hint）
  - `TestE2EP0049PracticeHintStrictRefusalAcrossGoals`（HTTP scenario 409 ApiError + details.policy='hint_disabled_in_mode'）
  - `TestE2EP0050PracticeAssistantActionProvenanceAndTaskRuns`（HTTP scenario provenance + observed ai_task_runs）
  - `TestE2EP0051PracticeHintDegradeAndPrivacy`（HTTP scenario 200 + session_wait + privacy redlines；不是 502/503）
  - `TestStartPracticeSessionObservedAIClientWritesTaskRunColumns` / `TestPracticeObservedAIRedactsPromptResponseFromLogsMetricsAndAudit`（A3 observability decorator typed columns / privacy regression）
  - `TestTaskRunWriterInsertsTypedColumns`（SQL writer 持久化 `hint_generate` / provenance / token / latency / metadata typed columns）

## Phase 4: 隐私 / 观测 / Legacy-Negative / Handoff

- [x] Phase 4 本计划定义的单元测试项全部通过：
  - `TestApplyHintAIPrivacyRedaction` + `TestPracticeObservedAIRedactsPromptResponseFromLogsMetricsAndAudit` + `TestE2EP0051PracticeHintDegradeAndPrivacy`（log / metric label / ai_task_runs typed columns / practice_session_events.payload 红线）
  - `test_backend_practice_legacy_includes_003_terms` / `test_backend_practice_legacy_allows_negative_docs`（pytest）+ `python3 scripts/lint/backend_practice_legacy.py --repo-root . --phase all`（scoped legacy grep）
  - `make codegen-check` / `make validate-fixtures` / `make lint-events` / `make codegen-events-check` / `python3 scripts/lint/conventions_drift.py --repo-root .` 全部通过
  - `migrations/lint.sh` 通过
  - `cd backend && go test ./... -count=1` 全绿

## 全局收口

- [x] `cd backend && go test ./...`
- [x] `make codegen-check`
- [x] `make validate-fixtures`
- [x] `migrations/lint.sh`
- [x] `make lint-events`
- [x] `make codegen-events-check`
- [x] `python3 scripts/lint/conventions_drift.py --repo-root .`
- [x] `python3 scripts/lint/backend_practice_legacy.py --repo-root . --phase all`
- [x] `python3 -m pytest scripts/lint/backend_practice_legacy_test.py -q`
- [x] `make docs-check`
- [x] `git diff --check`
