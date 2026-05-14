# 003 — Mode Policies and Provenance Test Checklist

> **版本**: 1.0
> **状态**: active
> **更新日期**: 2026-05-14

**关联 Test Plan**: [test-plan](./test-plan.md)

## Phase 0: 跨 spec 前置修订 + Preflight

- [ ] Phase 0 本计划定义的单元测试项全部通过：
  - `TestBaselineMigrationAcceptsHintGenerateTaskType` / `TestBaselineMigrationRejectsUnknownTaskType`（migration apply test，位于 `backend/internal/migrations/`；命令入口 `cd backend && go test ./internal/migrations/... -count=1`）
  - `migrations/lint.sh`（`migrations/enum-sources.yaml` 与 SQL CHECK 同步；漏改 enum manifest 时必须失败）
  - `TestValidTaskTypesIncludesHintGenerate`（writers.go enum 与 `allowedAITaskRunCapabilities` 集合一致性 + `Validate()` 接受 `hint_generate`）
  - `TestF3LightweightObservePreflight`（F3 baseline + ResolveActive 可用性，扩展 `backend/internal/ai/registry/backend_practice_preflight_test.go`）
  - `make validate-fixtures` + `make codegen-check` + `python3 scripts/lint/conventions_drift.py --repo-root .` 通过

## Phase 1: handleHintRequested mode-binding & strict 边界覆盖

- [ ] Phase 1 本计划定义的单元测试项全部通过：
  - `TestHandleHintRequestedModeMatrix`（assisted × 4 goal + strict × 4 goal + unknown × 2 fallback；strict 路径 fake registry 调用次数为 0）
  - `TestHandleHintRequestedTurnLifecycle`（OutboxRecord nil / NextTurn 未推进 / session 状态不变 / AuditMetadata 仅含 event_kind+mode）
  - `TestHandleHintRequestedStrictDoesNotInvokeAI`（fake registry / fake AIClient 调用次数为 0）
  - `TestAppendSessionEventHintStrictDoesNotLeavePendingReservation`（strict / unknown 409 后同 `(session_id, client_event_id)` 存在 finalized `kind='hint_requested'` 行，payload 仅含 sanitized conflict envelope，不存在 `payload.pending=true` stuck row；同 clientEventId 重试 replay 同一 409）

## Phase 2: assisted hint AI 接入 + practice_turns.hint_text 写入

- [ ] Phase 2 本计划定义的单元测试项全部通过：
  - `TestApplyHintAISuccess`（fake F3 + fake AIClient 成功路径 → outcome.Hint 非空 + Provenance 6 wire 字段）
  - `TestApplyHintAIBuildsPromptWithoutLeaks`（prompt user message 不拼接 question_text / answer_text 明文）
  - `TestApplyHintAIUsesHintGenerateCapability`（fake AITaskRunWriter 捕获 capability=`hint_generate`、resource_type=`target_job`、resource_id=plan.target_job_id）
  - `TestServiceAppliesHintAIForAssisted` / `TestServiceSkipsHintAIForStrict` / `TestServiceHintAndFollowUpAreMutuallyExclusive`（service layer 接入）
  - `TestAppendSessionEventWritesHintTextForAssistedSuccess`（repository UPDATE practice_turns.hint_text）
  - `TestAppendSessionEventDoesNotEmitTurnCompletedOutboxForHint`（hint 不发 outbox）
  - `TestAppendSessionEventDoesNotAdvanceTurnCountForHint`（hint 不递增 turn_count / 不改 turn.status）
  - `TestAppendSessionEventDoesNotWriteAuditForHint`（hint 不写 audit_events）
  - `TestAssistantActionProvenanceFieldsAreWireOnly`（reflect + struct tag）
  - `TestAssistantActionProvenanceJSONShape`（json.Marshal → 6 keys 严格等于 B2 GenerationProvenance）
  - `TestAssistantActionProvenanceCrossActionParity`（5 种 action type 的 provenance 字段集严格一致）

## Phase 3: AI failure graceful degrade & provenance wire boundary

- [ ] Phase 3 本计划定义的单元测试项全部通过：
  - `TestApplyHintAIGracefulDegradeMatrix`（表驱动 7 路径：F3 `registry.ErrPromptUnsupported` / `registry.ErrLanguageUnsupported` / A3 secret_missing / A3 timeout / A3 invalid output / A3 capability mismatch / parsed hint empty）
  - `TestApplyHintAIRecordsAITaskRunFromA3Decorator`（A3 路径：成功 → decorator 写 succeeded；A3 失败 → decorator 写 failed + B1 error_code）
  - `TestApplyHintAIWritesAITaskRunOnF3Failure`（F3 路径：ResolveActive 返回 `registry.ErrPromptUnsupported` / `registry.ErrLanguageUnsupported` 时 applyHintAI 显式调用 `aiclient.AITaskRunWriter.WriteAITaskRun` 写 failed + `error_code='AI_PROVIDER_CONFIG_INVALID'`）
  - `TestApplyHintAIWritesAITaskRunOnParsedHintEmpty`（Complete 成功但 parsed hint 为空时 applyHintAI 额外显式写 failed + `error_code='AI_OUTPUT_INVALID'`）
  - `TestApplyHintAIPrivacyRedaction`（log fields + service-local AuditMetadata 不含 hint_text / prompt body / response body / provider secret；AuditMetadata 不进入 `audit_events` 或 wire）
  - `TestApplyHintAIRejectsCapabilityMismatch`（A3 capability 错配走 degrade 路径）
  - `TestAppendSessionEventHintAssistedSuccess`（handler 200 + show_hint）
  - `TestAppendSessionEventHintAssistedDegrade`（handler 200 + session_wait + provenance non-AI default；200 SessionEventResult envelope **不**携带 `details.policy`；degrade 原因仅在 service-local AuditMetadata / ai_task_runs.error_code；不进入 `audit_events` 或 wire；**不是** 502/503）
  - `TestAppendSessionEventHintStrictConflict`（handler 409 ApiError + details.policy='hint_disabled_in_mode'）
  - `TestAppendSessionEventHintCrossUserNotFound`（cross-user 404 PRACTICE_SESSION_NOT_FOUND）
  - `TestObservabilityDecoratorWritesHintGenerateTaskRun`（A3 observability decorator typed columns）

## Phase 4: 隐私 / 观测 / Legacy-Negative / Handoff

- [ ] Phase 4 本计划定义的单元测试项全部通过：
  - `TestObservabilityHintRedaction`（log / metric label / ai_task_runs typed columns / practice_session_events.payload 红线）
  - `test_backend_practice_legacy_includes_003_terms` / `test_backend_practice_legacy_allows_negative_docs`（pytest）+ `python3 scripts/lint/backend_practice_legacy.py --repo-root . --phase all`（scoped legacy grep）
  - `make codegen-check` / `make validate-fixtures` / `make lint-events` / `make codegen-events-check` / `python3 scripts/lint/conventions_drift.py --repo-root .` 全部通过
  - `migrations/lint.sh` 通过
  - `cd backend && go test ./... -count=1` 全绿

## 全局收口

- [ ] `cd backend && go test ./...`
- [ ] `make codegen-check`
- [ ] `make validate-fixtures`
- [ ] `migrations/lint.sh`
- [ ] `make lint-events`
- [ ] `make codegen-events-check`
- [ ] `python3 scripts/lint/conventions_drift.py --repo-root .`
- [ ] `python3 scripts/lint/backend_practice_legacy.py --repo-root . --phase all`
- [ ] `python3 -m pytest scripts/lint/backend_practice_legacy_test.py -q`
- [ ] `make docs-check`
- [ ] `git diff --check`
