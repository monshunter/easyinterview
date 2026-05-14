# 003 — Mode Policies and Provenance Test Plan

> **版本**: 1.0
> **状态**: completed
> **更新日期**: 2026-05-14

**关联计划**: [plan](./plan.md) / [checklist](./checklist.md)
**关联 Test Checklist**: [test-checklist](./test-checklist.md)

## 0 范围与覆盖矩阵映射

本 test-plan 把 plan §3.5 Coverage Matrix 的 R1–R13 + plan §4 Phase 0–4 的实施步骤映射到具体测试包、文件、测试函数与命令。覆盖矩阵中的 BDD 行（场景级）由 `bdd-plan.md` 承担，本文件聚焦单元 / 集成 / contract / drift / privacy 单元测试。

| 覆盖矩阵行 | 测试包 / 文件 | 测试函数 | 测试命令 |
|------------|---------------|----------|----------|
| R1 (primary, assisted hint 写 hint_text) | `backend/internal/practice/append_session_event_service_test.go` + `backend/internal/store/practice/append_complete_test.go` | `TestApplyHintAISuccess` / `TestServiceAppliesHintAIForAssisted` / `TestSQLRepositoryAppendSessionEventWritesHintTextForAssistedSuccess` | `cd backend && go test ./internal/practice ./internal/store/practice -count=1` |
| R2 (provenance wire-only cross-action) | `backend/internal/practice/session_event_test.go` | `TestAssistantActionProvenanceFieldsAreWireOnly` / `TestAssistantActionProvenanceJSONShape` / `TestAssistantActionProvenanceCrossActionParity` | `cd backend && go test ./internal/practice -run 'TestAssistantActionProvenance' -count=1` |
| R3 (strict hint 拒绝) | `backend/internal/practice/session_event_test.go` + `backend/internal/practice/append_session_event_service_test.go` | `TestHandleHintRequestedModeMatrix`（strict subset）+ `TestServiceSkipsHintAIForStrict` | `cd backend && go test ./internal/practice -run 'TestHandleHintRequested|TestServiceSkipsHintAIForStrict' -count=1` |
| R4 (mode × goal 正交) | `backend/internal/practice/session_event_test.go` | `TestHandleHintRequestedModeMatrix`（assisted × 4 goal + strict × 4 goal + unknown × 2 fallback） | 同 R3 |
| R5 (hint AI 失败 graceful degrade 7 路径) | `backend/internal/practice/append_session_event_service_test.go` + `backend/cmd/api/practice_http_scenario_test.go` | `TestApplyHintAIGracefulDegradeMatrix`（表驱动 7 路径）+ `TestE2EP0051PracticeHintDegradeAndPrivacy` | `cd backend && go test ./internal/practice ./cmd/api -count=1` |
| R6 (hint turn-lifecycle 边界) | `backend/internal/practice/session_event_test.go` + `backend/internal/store/practice/append_complete_test.go` | `TestHandleHintRequestedTurnLifecycle` / `TestSQLRepositoryAppendSessionEventWritesHintTextForAssistedSuccess`（SQL expectation 固定不发 outbox / 不写 audit / 不改 turn status / turn_count 不变） | `cd backend && go test ./internal/practice ./internal/store/practice -count=1` |
| R7 (D-37 B4 CHECK + writers enum + spec history) | `backend/internal/migrations/baseline_aitaskruns_check_test.go` + `backend/internal/ai/aiclient/writers_test.go` + `scripts/lint/conventions_drift.py` + `migrations/enum-sources.yaml` | `TestBaselineAITaskRunsTaskTypeCheckIncludesHintGenerate` / `TestAITaskRunCapabilityIncludesHintGenerate` + drift lint + `migrations/lint.sh` | `cd backend && go test ./internal/migrations ./internal/ai/aiclient -count=1 && migrations/lint.sh && python3 scripts/lint/conventions_drift.py --repo-root .` |
| R8 (F3 lightweight_observe preflight + A3 capability) | `backend/internal/ai/registry/backend_practice_preflight_test.go` + `backend/internal/practice/append_session_event_service_test.go` | `TestBackendPracticeF3Preflight` / `TestApplyHintAIGracefulDegradeMatrix`（capability mismatch case） | `cd backend && go test ./internal/ai/registry ./internal/practice -count=1` |
| R9 (privacy: log / metric / audit / payload 不含 hint 明文) | `backend/internal/practice/append_session_event_service_test.go` + `backend/internal/practice/observability_test.go` + `backend/cmd/api/practice_http_scenario_test.go` | `TestApplyHintAIPrivacyRedaction` / `TestPracticeObservedAIRedactsPromptResponseFromLogsMetricsAndAudit` + `TestE2EP0051PracticeHintDegradeAndPrivacy` 子断言 | `cd backend && go test ./internal/practice ./cmd/api -count=1` |
| R10 (observability: ai_task_runs typed columns + metric allowlist) | `backend/internal/store/ai/task_runs_test.go` + `backend/internal/practice/append_session_event_service_test.go` + `backend/cmd/api/practice_http_scenario_test.go` + A3 decorator tests | `TestTaskRunWriterInsertsTypedColumns` / `TestApplyHintAIGracefulDegradeMatrix` / `TestE2EP0050PracticeAssistantActionProvenanceAndTaskRuns` / A3 decorator success/failure tests | `cd backend && go test ./internal/store/ai ./internal/practice ./cmd/api ./internal/ai/aiclient/observability -count=1` |
| R11 (UX/API status code 边界) | `backend/cmd/api/practice_http_scenario_test.go` | `TestE2EP0048PracticeHintAssistedAcrossGoals` / `TestE2EP0049PracticeHintStrictRefusalAcrossGoals` / `TestE2EP0051PracticeHintDegradeAndPrivacy` | `cd backend && go test ./cmd/api -run 'TestE2EP0048|TestE2EP0049|TestE2EP0051' -count=1` |
| R12 (scoped legacy-negative grep + spec text narrowing 残留) | `scripts/lint/backend_practice_legacy.py`（扩展 003 scoped scan/allowlist）+ `scripts/lint/backend_practice_legacy_test.py` | `test_backend_practice_legacy_includes_003_terms` + `test_backend_practice_legacy_allows_negative_docs` + scoped legacy grep | `python3 scripts/lint/backend_practice_legacy.py --repo-root . --phase all && python3 -m pytest scripts/lint/backend_practice_legacy_test.py -q` |
| R13 (out-of-scope boundary: 003 不动 004/005/006 owner 范围) | `backend/internal/practice/append_session_event_service_test.go` + `backend/internal/store/practice/append_complete_test.go` + scoped legacy grep | `TestApplyHintAISuccess`（metadata 只使用 lightweight_observe）/ `TestSQLRepositoryAppendSessionEventWritesHintTextForAssistedSuccess`（不触碰 plan source columns）+ `backend_practice_legacy.py --phase all` | 同 R6 + grep gate |

## Phase 0: 跨 spec 前置修订 + Preflight

- **测试目标**：B4 `ai_task_runs.task_type` CHECK 接受 `hint_generate` + 拒绝 unknown；`migrations/enum-sources.yaml` 与 SQL CHECK 同步且 `migrations/lint.sh` 可抓 drift；writers.go enum 与 `allowedAITaskRunCapabilities` 集合一致；`AITaskRunContext.Validate()` 接受 `hint_generate`；fixtures 命名场景通过 validator；backend-practice / B4 spec history append 文件存在；F3 lightweight_observe baseline 可被 ResolveActive 找回。
- **测试文件**：
  - `backend/internal/migrations/baseline_aitaskruns_check_test.go`（新增）或在既有 `backend/internal/migrations/sql_contract_test.go` 中扩展 baseline migration CHECK assertion
  - `backend/internal/ai/aiclient/writers_test.go`（扩展 `TestAITaskRunCapabilityIncludesHintGenerate`）
  - `backend/internal/ai/registry/backend_practice_preflight_test.go`（扩展 `TestBackendPracticeF3Preflight`）
  - `openapi/fixtures/PracticeSessions/appendSessionEvent.json`（fixtures，3 个新命名场景）
- **测试命令**：
  - `cd backend && go test ./internal/migrations ./internal/ai/aiclient ./internal/ai/registry -count=1`
  - `make validate-fixtures`
  - `make codegen-check`
  - `migrations/lint.sh`
  - `python3 scripts/lint/conventions_drift.py --repo-root .`
- **预期 Red / Green 证据**：
  - Red：修订 migration / writers.go 前，新测试 fail（CHECK 不含 `hint_generate` / 常量未定义 / fixture validator 缺少命名场景）
  - Green：Phase 0 完成后所有命令通过

## Phase 1: handleHintRequested mode-binding & strict 边界覆盖

- **测试目标**：handleHintRequested(session, plan) 按 `plan.Mode` 拆分；assisted 返回 RequiresAI=true pending outcome；strict / unknown / legacy 返回 409 outcome 且不调 AI；turn lifecycle 边界（不发 outbox / 不改 turn status / session 状态不变）；strict / unknown 409 finalize 为 sanitized conflict payload，不留下 `practice_session_events.payload.pending=true` stuck reservation，并保留 D-38 要求的 `hint_requested` 事件留痕。
- **测试文件**：
  - `backend/internal/practice/session_event_test.go`（扩展 `TestHandleHintRequestedModeMatrix` / `TestHandleHintRequestedTurnLifecycle`）
  - `backend/internal/practice/append_session_event_service_test.go`（扩展 `TestAppendSessionEventHintStrictDoesNotLeavePendingReservation`）
- **测试命令**：
  - `cd backend && go test ./internal/practice -run 'TestHandleHintRequested' -count=1`
  - `cd backend && go test ./internal/practice -run 'TestAppendSessionEventHintStrictDoesNotLeavePendingReservation' -count=1`
- **预期 Red / Green 证据**：
  - Red：重构前测试 fail（assisted 路径未返回 RequiresAI=true；unknown mode 未走 fail-safe strict）
  - Green：Phase 1 完成后 `cd backend && go test ./internal/practice -count=1` 通过

## Phase 2: assisted hint AI 接入 + practice_turns.hint_text 写入

- **测试目标**：applyHintAI 成功路径用 F3 lightweight_observe + A3 Complete 生成 hint；service layer 在 assisted + RequiresAI 时调用 applyHintAI 且与 follow_up 分支互斥；store layer 在 outcome.AssistantAction.Type='show_hint' 且 Hint 非空时 UPDATE `practice_turns.hint_text`；wire provenance 严格 6 字段。
- **测试文件**：
  - `backend/internal/practice/hint_ai.go`（新增；hint AI 接入逻辑）
  - `backend/internal/practice/append_session_event_service_test.go`（扩展）：`TestApplyHintAISuccess` / `TestApplyHintAIBuildsPromptWithoutLeaks` / `TestServiceAppliesHintAIForAssisted` / `TestServiceSkipsHintAIForStrict`
  - `backend/internal/practice/append_session_event_service.go`（修改）
  - `backend/internal/store/practice/append_event.go`（修改：增加 hint_text UPDATE 分支）
  - `backend/internal/store/practice/append_complete_test.go`（扩展）：`TestSQLRepositoryAppendSessionEventWritesHintTextForAssistedSuccess`
  - `backend/internal/practice/session_event_test.go`（扩展）：`TestAssistantActionProvenanceFieldsAreWireOnly` / `TestAssistantActionProvenanceJSONShape` / `TestAssistantActionProvenanceCrossActionParity`
- **测试命令**：
  - `cd backend && go test ./internal/practice ./internal/store/practice -count=1`
- **预期 Red / Green 证据**：
  - Red：Phase 1 完成后 `TestApplyHintAISuccess` 等 fail（函数未实现 / store 不写 hint_text / provenance 未填）
  - Green：Phase 2 完成后命令通过；handler 层 BDD `E2E.P0.048` / `E2E.P0.049` 通过

## Phase 3: AI failure graceful degrade & provenance wire boundary

- **测试目标**：applyHintAI 7 失败路径都返回 graceful degrade outcome；ai_task_runs 行 validation_status='failed' + error_code 来自 B1 enum；A3 callErr 由 decorator 写 failed row，F3 resolve 与 parse-after-success 失败由 `applyHintAI` 显式写 failed row；wire provenance regression 在 5 种 action type 一致；handler 错误映射断言 status code 与 envelope shape。
- **测试文件**：
  - `backend/internal/practice/hint_ai.go`（扩展 7 路径 degrade 实现；F3 失败路径显式调用 `aiclient.AITaskRunWriter.WriteAITaskRun`）
  - `backend/internal/practice/append_session_event_service_test.go`（扩展）：`TestApplyHintAIGracefulDegradeMatrix`（表驱动 7 路径，覆盖 F3 / A3 / parsed-empty）/ `TestApplyHintAIPrivacyRedaction`
  - `backend/internal/store/ai/task_runs.go` + `task_runs_test.go`（新增）：SQL `ai_task_runs` writer 与 typed column 持久化测试
  - `backend/cmd/api/practice_http_scenario_test.go`（扩展）：`TestE2EP0048...` / `TestE2EP0049...` / `TestE2EP0050...` / `TestE2EP0051...` 串联 handler status code / envelope / observed task-run 断言
  - `backend/internal/practice/session_event_test.go`（扩展）：5 种 action type cross-marshal regression；show_hint 子断言 `rubricVersion=='not_applicable'`
- **测试命令**：
  - `cd backend && go test ./internal/practice ./internal/api/practice ./internal/ai/aiclient/observability -count=1`
- **预期 Red / Green 证据**：
  - Red：未实现 7 路径 / handler 错误映射前测试 fail
  - Green：Phase 3 完成后命令通过；handler 返回 200（assisted degrade）/ 200（assisted success）/ 409（strict）/ 404（cross-user）正确

## Phase 4: 隐私 / 观测 / Legacy-Negative / Handoff

- **测试目标**：log / metric label / audit / ai_task_runs / practice_session_events.payload / structured log 不含 hint_text / AI prompt / response 明文；F1 allowlist 命中；hint 路径不写 audit；scoped legacy grep 反查实现 / runtime 输出中的 retired 术语 + spec 通用 fail-closed 文字残留，并允许 negative test / prohibition docs 枚举被禁字面量。
- **测试文件**：
  - `backend/internal/practice/observability_test.go`（复用）：`TestPracticeObservedAIRedactsPromptResponseFromLogsMetricsAndAudit`
  - `backend/internal/store/practice/outbox_emitter_test.go`（保持 002 行为，hint 路径不产生 outbox 由 R6 单测覆盖）
  - `backend/cmd/api/practice_http_scenario_test.go`（新增 4 个 003 scenario）：`TestE2EP0048...` / `TestE2EP0049...` / `TestE2EP0050...` / `TestE2EP0051...`（详见 bdd-plan）
  - `scripts/lint/backend_practice_legacy.py`（扩展 003 scoped scan/allowlist 反查项）
  - `scripts/lint/backend_practice_legacy_test.py`（扩展 `test_backend_practice_legacy_includes_003_terms` / `test_backend_practice_legacy_allows_negative_docs`）
- **测试命令**：
  - `cd backend && go test ./internal/practice ./internal/store/practice ./cmd/api -count=1`
  - `python3 scripts/lint/backend_practice_legacy.py --repo-root . --phase all`
  - `python3 -m pytest scripts/lint/backend_practice_legacy_test.py -q`
- **预期 Red / Green 证据**：
  - Red：未补 003 反查项时 lint test fail
  - Green：Phase 4 完成后 plan 收口命令全绿

## 全局收口测试命令

```bash
cd backend && go test ./...
make codegen-check
make validate-fixtures
migrations/lint.sh
make lint-events
make codegen-events-check
python3 scripts/lint/conventions_drift.py --repo-root .
python3 scripts/lint/backend_practice_legacy.py --repo-root . --phase all
python3 -m pytest scripts/lint/backend_practice_legacy_test.py -q
make docs-check
```

不引入硬编码覆盖率门槛（observational only）。
