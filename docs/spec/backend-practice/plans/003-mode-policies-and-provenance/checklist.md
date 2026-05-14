# 003 — Mode Policies and Provenance Checklist

> **版本**: 1.0
> **状态**: completed
> **更新日期**: 2026-05-14

**关联计划**: [plan](./plan.md)

## Phase 0: 跨 spec 前置修订 + Preflight

- [x] 0.1 修订 `docs/spec/backend-practice/spec.md` Header `1.7 → 1.8`；§3.1 锁定决策表追加 D-36 行（hint AI graceful degrade narrowing）+ D-38 行（hint turn-lifecycle 边界）；§6 C-17 行 inline narrowing 拆 session-survival vs 辅助 AI；§3.1 D-19 行追加 "（hint / lightweight_observe AI 失败例外，按 D-36 graceful degrade）" + 首句改为 "session-survival AI（first_question / follow_up）：..."；§4.3 line 140 整行修订为按调用类别区分；§2.1 line 45 失败语义整行修订；§7 row 3 描述追加 D-36 / D-37 / D-38 三标签 + plan 路径链接化（验证：`make docs-check` 通过 + 在 §6 C-17 / §3.1 D-19 / §4.3 / §2.1 四个来源点 grep "整个 operation.*session.*failed" / "所有 AI 调用必须 fail-closed" / "全 AI 一律 session=failed" 零残留 + §3.1 grep "D-36" / "D-38" 各 1 命中）
- [x] 0.2 修订 `docs/spec/backend-practice/history.md` Header `1.7 → 1.8` 同步 + 追加 1.8 row：覆盖 §6 C-17 / §3.1 D-19 / §4.3 / §2.1 四处 narrowing + D-36 / D-38 新增决策行 + §7 row 3 链接与 D-36 / D-37 / D-38 引用 + 与 002 D-19 follow_up fallback 一致性说明（验证：`docs/spec/backend-practice/plans/INDEX.md` 与 `docs/spec/INDEX.md` 同步无 drift）
- [x] 0.3 修订 `migrations/000001_create_baseline.up.sql:386` `ai_task_runs.task_type` CHECK 扩值 `hint_generate`（pre-launch baseline rebase），并同步 `migrations/enum-sources.yaml` values / checksum；验证：migration apply test 用 `cd backend && go test ./internal/migrations/... -count=1` 入口（既有 baseline 契约测试位于 `backend/internal/migrations/sql_contract_test.go`，新增 `backend/internal/migrations/baseline_aitaskruns_check_test.go` 或在 `sql_contract_test.go` 中扩展）断言 CHECK 接受 `hint_generate` + 拒绝 `unknown_task`，`migrations/lint.sh` 通过并能在漏改 enum manifest 时失败
- [x] 0.4 修订 `backend/internal/ai/aiclient/writers.go` 新增 `AITaskRunTaskHintGenerate AITaskRunCapability = "hint_generate"` 常量；把 `AITaskRunTaskHintGenerate` 加入 `allowedAITaskRunCapabilities` 集合（验证：`cd backend && go test ./internal/ai/aiclient -count=1` 全通过 + 新增 `TestAITaskRunCapabilityIncludesHintGenerate` 断言常量与 `allowedAITaskRunCapabilities` 集合一致 + `Validate()` 接受 `hint_generate`）
- [x] 0.5 修订 `docs/spec/db-migrations-baseline/spec.md` Header bump（next minor version）+ `history.md` 追加 row："授权 backend-practice/003 Phase 0 `ai_task_runs.task_type` CHECK 扩值 `hint_generate`（pre-launch baseline rebase）"（验证：spec history append 文件 grep 断言）
- [x] 0.6 扩展 `openapi/fixtures/PracticeSessions/appendSessionEvent.json` 命名场景：`hint-assisted-show` / `hint-assisted-ai-failed-degrade` / `hint-strict-debrief-conflict`（验证：`make validate-fixtures` 通过 + contract test 通过）
- [x] 0.7 F3 preflight assert：读取 `docs/spec/prompt-rubric-registry/spec.md` 当前版本 + `plans/001-baseline/checklist.md` `状态: completed` + work-journal `close 001-baseline lifecycle` commit；断言 `practice.turn.lightweight_observe` baseline 行存在且 `RegistryClient.ResolveActive(ctx, "practice.turn.lightweight_observe", "en")` 返回非空 ResolvedPrompt（验证：复用 `backend/internal/ai/registry/backend_practice_preflight_test.go::TestBackendPracticeF3Preflight`）
- [x] 0.8 Phase 0 收口：`make codegen-check` / `make validate-fixtures` / `migrations/lint.sh` / `python3 scripts/lint/conventions_drift.py --repo-root .` / `cd backend && go build ./...` 全部通过；本 plan 状态保持 `active`

## Phase 1: handleHintRequested mode-binding & strict 边界覆盖

- [x] 1.1 重构 `backend/internal/practice/session_event.go::handleHintRequested(session, plan)`：按 `plan.Mode` 拆分 assisted（RequiresAI=true pending outcome）与 strict / unknown（409 PRACTICE_SESSION_CONFLICT outcome）（验证：`session_event_test.go::TestHandleHintRequestedModeDispatch` 覆盖 strict / assisted / unknown / legacy 取值；新增 unit test）
- [x] 1.2 实现 mode × goal 4×2 + 兜底单元测试矩阵：`TestHandleHintRequestedModeMatrix` 覆盖 8 + 2 路径；fake registry 计数断言 strict 路径调用次数为 0（验证：`cd backend && go test ./internal/practice -run TestHandleHintRequested -count=1`）
- [x] 1.3 实现 D-38 turn lifecycle 边界单元测试：`TestHandleHintRequestedTurnLifecycle` 断言 OutboxRecord nil / NextTurn 未推进 / session 状态不变 / AuditMetadata 仅含 event_kind + mode（验证：同 1.2 命令）
- [x] 1.4 修复 strict / unknown 409 的 reserved event 收尾：`ReserveSessionEvent` 已写入 pending row 后，409 outcome 必须 finalized 为 sanitized conflict replay payload；禁止留下 `payload.pending=true` stuck row，且不得清理掉 D-38 要求的 `hint_requested` 事件留痕（验证：`append_session_event_service_test.go::TestAppendSessionEventHintStrictDoesNotLeavePendingReservation`，并覆盖同 clientEventId 重试语义）
- [x] 1.5 Phase 1 收口：`cd backend && go test ./internal/practice/... -count=1` 全部通过；handleHintRequested 单元测试覆盖率包含 8 + 2 组合；strict / unknown 409 no-pending gate 通过
- [x] 1.6 BDD-Gate: 验证 `E2E.P0.049` 通过（strict hint 拒绝 + goal × mode 矩阵 strict 子集 + no-pending reservation 子断言；Go HTTP scenario `TestE2EP0049PracticeHintStrictRefusalAcrossGoals`，跟踪在 [bdd-checklist](./bdd-checklist.md)）

## Phase 2: assisted hint AI 接入 + practice_turns.hint_text 写入

- [x] 2.1 新增 `backend/internal/practice/hint_ai.go`：const `hintFeatureKey = "practice.turn.lightweight_observe"`；`applyHintAI(ctx, reservation, payload, outcome)` 函数沿用 002 `applyFollowUpAI` 模式；TaskRun.Capability = `aiclient.AITaskRunTaskHintGenerate`；TaskRun.ResourceType = `AITaskRunResourceTargetJob`；TaskRun.ResourceID = `reservation.Plan.TargetJobID`（验证：`append_session_event_service_test.go::TestApplyHintAISuccess` 用 fake F3 + fake AIClient 断言 outcome.Hint 非空 + outcome.Provenance 6 wire 字段）
- [x] 2.2 修改 `backend/internal/practice/append_session_event_service.go::AppendSessionEvent`：在 router.Route 后、applyFollowUpAI 前，按 `outcome.AssistantAction.Type == "show_hint" && RequiresAI` 调用 `s.applyHintAI`；与 follow_up 分支互斥（验证：`append_session_event_service_test.go::TestServiceAppliesHintAIForAssisted` + `TestServiceSkipsHintAIForStrict`）
- [x] 2.3 修改 `backend/internal/store/practice/append_event.go::AppendSessionEvent`：事务内根据 `Outcome.AssistantAction.Type == "show_hint" && Hint != ""` UPDATE `practice_turns.hint_text`；hint 路径不改 turn.status / turn_index / session.turn_count，不写 `outbox_events practice.turn.completed`，不写 `audit_events`（验证：`append_complete_test.go::TestSQLRepositoryAppendSessionEventWritesHintTextForAssistedSuccess` 的 SQL expectation 固定只写 hint_text + session replay payload，不发 outbox / 不写 audit / 不改 turn status，并断言 session turn_count 不变）
- [x] 2.4 实现 AssistantAction wire provenance cross-action regression 单元测试：`provenance_test.go::TestAssistantActionProvenanceFieldsAreWireOnly` 用 reflect + json.Marshal 断言 5 种 action type 的 provenance 字段集严格等于 B2 GenerationProvenance 6 字段；不含 runtime 字段（验证：`cd backend && go test ./internal/practice -run TestAssistantActionProvenance -count=1`）
- [x] 2.5 BDD-Gate: 验证 `E2E.P0.048` 通过（assisted hint 主路径 + goal × mode 矩阵 assisted 子集 4 个子断言；Go HTTP scenario `TestE2EP0048PracticeHintAssistedAcrossGoals`）
- [x] 2.6 BDD-Gate: 验证 `E2E.P0.049` 通过（Phase 1 已覆盖 strict outcome 单元测试；本步在 HTTP scenario 层串联 strict 矩阵 4 个子断言）

## Phase 3: AI failure graceful degrade & provenance wire boundary

- [x] 3.1 在 `hint_ai.go::applyHintAI` 完善 7 路径 graceful degrade：F3 `registry.ErrPromptUnsupported` / `registry.ErrLanguageUnsupported` / A3 secret_missing / A3 timeout / A3 invalid output / A3 capability mismatch / parsed hint empty；每路径 outcome.AssistantAction.Type 置 `session_wait`、Hint 清空、Provenance 用 non-AI default、service-local AuditMetadata 写 sanitized `hint_degrade_reason`；不调 store hint_text UPDATE（验证：`append_session_event_service_test.go::TestApplyHintAIGracefulDegradeMatrix` 表驱动覆盖 7 路径）
- [x] 3.2 实现 fake AITaskRunWriter 断言 ai_task_runs 行（A3 路径 vs F3 路径分桶）：
  - **A3 路径**：`cmd/api` HTTP scenario harness 的 observed AIClient 覆盖 assisted 成功路径 `task_type='hint_generate'` + `validation_status='succeeded'`，并由 A3 decorator 既有失败测试覆盖 callErr → failed + B1 error_code 语义
  - **F3 / parse 路径**：`append_session_event_service_test.go::TestApplyHintAIGracefulDegradeMatrix`——F3 ResolveActive 返回 `registry.ErrPromptUnsupported` / `registry.ErrLanguageUnsupported` 时，`applyHintAI` 显式调用 `aiclient.AITaskRunWriter.WriteAITaskRun`（通过 `observability.AITaskRunRowFromMeta` 构造行）写一行 `task_type='hint_generate', validation_status='failed', error_code='AI_PROVIDER_CONFIG_INVALID'`；Complete 成功但 parsed hint 为空时额外显式写 failed row（`error_code='AI_OUTPUT_INVALID'`）。两类失败行都不得含 raw provider message / prompt body / response body / hint_text 明文
- [x] 3.2.1 补齐生产 wiring：新增 SQL `ai_task_runs` writer 并在 `cmd/api::buildPracticeRoutes` 中把 Practice AI client 包上 A3 observability decorator；同时把 writer 传入 `domainpractice.ServiceOptions.AITaskRuns`，确保 A3 success/failure 与 F3/parse explicit failure 都有同一持久化出口（验证：`internal/store/ai::TestTaskRunWriterInsertsTypedColumns` + `cmd/api` focused build tests + E2E.P0.050/P0.051 observed harness）
- [x] 3.3 实现 wire provenance boundary regression：`provenance_test.go::TestAssistantActionProvenanceJSONShape` 用 reflect 断言 struct 字段数 == 6 + 无 runtime 字段名；用 json.Marshal + Unmarshal 断言 keys 集合；增加 show_hint 子断言 `rubricVersion=='not_applicable'`（D-10 非评分动作硬编码，不沿用 resolution.RubricVersion）（验证：同 2.4 命令）
- [x] 3.4 实现 handler error mapping：通过 `cmd/api` HTTP scenario `TestE2EP0048PracticeHintAssistedAcrossGoals` / `TestE2EP0049PracticeHintStrictRefusalAcrossGoals` / `TestE2EP0051PracticeHintDegradeAndPrivacy` 断言 200 / 409 / 200 状态码与 envelope shape；assisted-degrade 路径断言 200 + `AssistantAction{type:'session_wait', hint:null, provenance(non-AI default), sessionStatus:'running'}`（200 envelope 不携带 `details.policy`，degrade 原因仅在 service-local AuditMetadata / ai_task_runs.error_code；不进入 `audit_events` 或 wire）；strict 路径断言 409 + `ApiError{code:'PRACTICE_SESSION_CONFLICT', details:{policy:'hint_disabled_in_mode', mode}}`（409 ApiError envelope 才有 details）（验证：`cd backend && go test ./cmd/api -run 'TestE2EP0048|TestE2EP0049|TestE2EP0051' -count=1`）
- [x] 3.5 BDD-Gate: 验证 `E2E.P0.050` 通过（AssistantAction provenance wire 边界 + ai_task_runs runtime 字段验证；Go HTTP scenario `TestE2EP0050PracticeAssistantActionProvenanceAndTaskRuns`）
- [x] 3.6 BDD-Gate: 验证 `E2E.P0.051` 主路径通过（graceful degrade + 隐私 + legacy negative；Go HTTP scenario `TestE2EP0051PracticeHintDegradeAndPrivacy`）

## Phase 4: 隐私 / 观测 / Legacy-Negative / Handoff

- [x] 4.1 Redaction 单元测试：assisted hint 路径 `practice_session_events.payload` json 不含 hint_text 明文；structured log + A3 metric label + ai_task_runs typed columns 不含 hint_text / AI prompt body / response body / provider secret；hint 路径不写 domain `audit_events`（验证：`append_session_event_service_test.go::TestApplyHintAIPrivacyRedaction` + `observability_test.go::TestPracticeObservedAIRedactsPromptResponseFromLogsMetricsAndAudit` + `TestE2EP0051PracticeHintDegradeAndPrivacy` 在 cmd/api 层串联）
- [x] 4.2 Metric label allowlist 单元测试：A3 `ai_task_*` metric label 命中 F1 allowlist；不含 `feature_key` / prompt-rubric version / provider raw model id；ai_task_runs row `task_type='hint_generate'` 合法（验证：复用 A3 observed AIClient 既有 allowlist 测试 + 003 新增 hint-specific assertion）
- [x] 4.3 Scoped legacy grep gate（在测试或 lint 中执行；扩展 `scripts/lint/backend_practice_legacy.py --phase all` 或新增 003-specific phase）：
    - 实现 / runtime 输出范围（backend practice/API/store、openapi PracticePlans/PracticeSessions fixtures、scenario runtime assets、generated/runtime tests）中 `hint_disabled_globally` / `legacy_hint_policy` / `legacy_mode_assisted_value` / `legacy debrief replay value` 零出现；negative fixture/test input 与禁止性文档说明可枚举这些字面量，但不得作为 active `PracticeMode` / schema value / route 出现
    - backend-practice spec.md / history.md 中 "全 AI 一律 session=failed" / "所有 AI 调用必须 fail-closed" 等未区分 session-survival / 辅助 的通用文字零残留；本 plan / checklist / bdd/test docs 允许在 negative-gate context 中引用这些短语作为禁止性断言
    - 实现 / runtime 输出范围中 `warmup` / `single_drill` / `drill_builder` / `mistake_queue` / `growth_center` / 独立 `voice` route / `practiceModeCard` 零出现；本 plan、BDD、test-plan、negative test allowlist 可枚举这些字面量作为 grep 输入
- [x] 4.4 修订 `backend/internal/api/practice/README.md`：新增 `## 003 Mode Policies and Provenance` 段记录 handleHintRequested mode dispatch / `practice.turn.lightweight_observe` wiring / `AITaskRunTaskHintGenerate` / D-36 graceful degrade / `practice_turns.hint_text` 写入边界；修订 `## Handoff Boundaries` 003 行从 "owned by 003" 改为 "delivered in 003"（验证：手工 review + grep `applyHintAI` / `hint_generate` / `AITaskRunTaskHintGenerate` 在 README 中均被提及）
- [x] 4.5 BDD-Gate: 验证 `E2E.P0.051` 子断言全部通过（含 ai_task_runs typed columns 隐私反查 + log/metric/audit 红线 + scoped legacy grep）
- [x] 4.6 收口 gate：`cd backend && go test ./... -count=1` 全绿；`make codegen-check`、`make validate-fixtures`、`migrations/lint.sh`、`make lint-events`、`make codegen-events-check`、`python3 scripts/lint/conventions_drift.py --repo-root .` 全通过
- [x] 4.7 更新 `docs/spec/backend-practice/plans/INDEX.md`：将 003 加到 active 表；本 plan 状态保持 `active`（后续 retrospective / closure 由 plan-review / sync-doc-index 推进到 completed，本 plan 编写阶段不自行 close）

## 收口证据

- `cd backend && go test ./internal/api/practice ./internal/practice ./internal/store/practice ./internal/ai/aiclient ./internal/ai/registry ./cmd/api -count=1`
- `cd backend && go test ./...`
- `make codegen-check`
- `make validate-fixtures`
- `migrations/lint.sh`
- `make lint-events`
- `make codegen-events-check`
- `python3 scripts/lint/conventions_drift.py --repo-root .`
- `python3 scripts/lint/backend_practice_legacy.py --repo-root . --phase all`（含 003 scoped legacy 反查项）
- `make docs-check`
- `git diff --check`
