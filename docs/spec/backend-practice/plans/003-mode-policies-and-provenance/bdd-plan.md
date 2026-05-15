# 003 — Mode Policies and Provenance BDD Plan

> **版本**: 1.1
> **状态**: completed
> **更新日期**: 2026-05-15

**关联计划**: [plan](./plan.md) / [checklist](./checklist.md)
**关联 BDD Checklist**: [bdd-checklist](./bdd-checklist.md)

## 0 BDD 框架与编号

本 plan 的 4 个 BDD 场景保留 `E2E.P0.xxx` 编号与 Given / When / Then 语义；003 本次代码交付的可执行入口落在 `backend/cmd/api/practice_http_scenario_test.go` 的 HTTP scenario tests，用 `cmd/api` 真实路由、middleware、service/store fake 与 response / side-effect 断言覆盖用户可见 API 行为（与 002 同模式）。

- 套件: `e2e`
- 阶段: `P0`
- 已占用编号现状（[`test/scenarios/e2e/INDEX.md`](../../../../../test/scenarios/e2e/INDEX.md)）：`001-006`, `010-037`；002 已通过 Go HTTP scenario 占用 `038-043`（未挂 e2e INDEX 目录，与 002 bdd-plan §0 描述一致）。本 plan 在空闲号段 `048-051` 中分配 4 个场景
- 编号分配: `E2E.P0.048` / `E2E.P0.049` / `E2E.P0.050` / `E2E.P0.051`
- 执行入口: `cd backend && go test ./cmd/api -run 'TestE2EP0048|TestE2EP0049|TestE2EP0050|TestE2EP0051' -count=1`
- 外部 Kind / shell 场景资产: 003 不新增 `test/scenarios/e2e/p0-NNN-*` 目录；若未来需要 Kind live BDD，可由 scenarios owner 按当前编号把同一 Given / When / Then 提升为 `test/scenarios` 资产，不改变本 plan 的 API 行为语义

每个场景的执行证据在 [bdd-checklist](./bdd-checklist.md) 跟踪；本文件只记录场景的 Given / When / Then 与覆盖范围，不出现执行 checkbox。

## 1 场景矩阵

| 场景 ID | 名称 | 类别 | 关联 Plan Phase | 关联 spec AC / D |
|---------|------|------|----------------|-------------------|
| `E2E.P0.048` | assisted hint 主路径 × goal 矩阵（baseline / retry_current_round / next_round / debrief × assisted） | primary + alternate | Phase 2 | C-7, C-8b, D-5, spec §3.1 D-38（plan-level：hint turn-lifecycle 边界） |
| `E2E.P0.049` | strict hint 拒绝 × goal 矩阵（baseline / retry_current_round / next_round / debrief × strict） | alternate + failure | Phase 1 + Phase 2 | C-8, C-8b, D-5, D-16, D-34（002） |
| `E2E.P0.050` | AssistantAction provenance wire 边界 + ai_task_runs runtime 字段（show_hint / ask_question / ask_follow_up / session_wait / session_completed 五种 action） | cross-layer contract | Phase 2 + Phase 3 | C-12, D-10, D-37（plan-level） |
| `E2E.P0.051` | hint AI 失败 graceful degrade + 隐私红线 + legacy-negative grep + ai_task_runs typed columns 反查 | failure/recovery + privacy + regression | Phase 3 + Phase 4 | C-17（narrowed）, C-16（子集）, D-11, D-20, D-36（plan-level）, D-37（plan-level） |

## 2 Phase 2 — assisted hint 主路径

| 场景 ID | 场景 | Given | When | Then | 验证入口 |
|---------|------|-------|------|------|----------|
| `E2E.P0.048` | assisted hint 主路径 × goal 矩阵 | 用户 A 已通过 002 setup 拥有 4 个 ready plan，goal 分别为 `baseline` / `retry_current_round` / `next_round` / `debrief`，全部 `mode='assisted'`；对应 4 个 `practice_sessions(status='running', currentTurn.status='asked')`；F3 active；A3 fake AIClient 配置为 `practice.turn.lightweight_observe` 返回 hint 文本 | 用户 A 对每个 session `POST /practice/sessions/{sessionId}/events` 携带 `kind='hint_requested'` + `payload.turnId=currentTurn.id` + 各自独立 `clientEventId`；随后同一 turn 使用第二个 `clientEventId` 请求不同 hint，再重放第一个 `clientEventId` | 每次返回 `200 + SessionEventResult{acknowledged: true, session: {status: 'running', ...}, assistantAction: {type: 'show_hint', hint: <non-empty>, sessionStatus: 'running', provenance: {promptVersion, rubricVersion, modelId: 'model-profile:practice.turn_observe.default', language, featureFlag, dataSourceVersion}}}`；第一个 `clientEventId` replay 返回原始 hint，不受第二次 hint 覆盖影响；DB `practice_turns.hint_text` 被 UPDATE 非空；DB `practice_turns.status` / `turn_index` / `practice_sessions.turn_count` 不变；`outbox_events WHERE event_name='practice.turn.completed' AND aggregate_id=sessionId` 行数不增（hint 不发 outbox）；`audit_events` 行不增（hint 不写 audit）；`practice_session_events.payload` 不含 hint_text 明文；ai_task_runs 写入 `task_type='hint_generate', validation_status='succeeded'` | `backend/cmd/api/practice_http_scenario_test.go::TestE2EP0048PracticeHintAssistedAcrossGoals` |

## 3 Phase 1 + Phase 2 — strict hint 拒绝

| 场景 ID | 场景 | Given | When | Then | 验证入口 |
|---------|------|-------|------|------|----------|
| `E2E.P0.049` | strict hint 拒绝 × goal 矩阵 | 用户 A 拥有 4 个 ready plan，goal 分别为 `baseline` / `retry_current_round` / `next_round` / `debrief`，全部 `mode='strict'`；对应 4 个 `practice_sessions(status='running')` | 用户 A 对每个 session `POST /practice/sessions/{sessionId}/events` 携带 `kind='hint_requested'` + `payload.turnId=currentTurn.id` + 各自独立 `clientEventId` | 每次返回 `409 + ApiError{code: 'PRACTICE_SESSION_CONFLICT', details: {policy: 'hint_disabled_in_mode', mode: 'strict'}}`；DB `practice_turns.hint_text` 保持 NULL；session 状态不变；`practice_session_events` 存在 finalized `kind='hint_requested'` 行，payload 仅含 sanitized conflict envelope，且不存在 `payload.pending=true` stuck row；同 `clientEventId` 重试 replay 同一 409；`ai_task_runs` 不写新行（strict 不调 AI）；`outbox_events` / `audit_events` 行不增；对 fake AIClient.Complete 的调用次数为 0 | `backend/cmd/api/practice_http_scenario_test.go::TestE2EP0049PracticeHintStrictRefusalAcrossGoals` |

## 4 Phase 2 + Phase 3 — provenance wire 边界

| 场景 ID | 场景 | Given | When | Then | 验证入口 |
|---------|------|-------|------|------|----------|
| `E2E.P0.050` | AssistantAction provenance wire 边界 + ai_task_runs runtime 字段 | 用户 A 拥有 `practice_sessions(status='running', mode='assisted', currentTurn.status='asked')`；F3 active；A3 fake AIClient 同时配置 `practice.session.follow_up`（answer_submitted → ask_follow_up 分支）与 `practice.turn.lightweight_observe`（hint_requested → show_hint 分支）；fake AITaskRunWriter 捕获所有 ai_task_runs typed columns 写入 | 用户 A 依次触发：① `answer_submitted` → AssistantAction.type=`ask_follow_up`（带 AI provenance）；② `hint_requested` → AssistantAction.type=`show_hint`（带 AI provenance）；③ `turn_skipped` → AssistantAction.type=`ask_question`（non-AI provenance）；④ `session_paused` → AssistantAction.type=`session_wait`（non-AI provenance）；⑤ 达到 `question_budget` 后 `answer_submitted` → AssistantAction.type=`session_completed`（non-AI provenance） | 每次 response 的 `assistantAction.provenance` 严格只含 6 个字段 `promptVersion` / `rubricVersion` / `modelId` / `language` / `featureFlag` / `dataSourceVersion`；任何 runtime 字段（`featureKey` / `feature_key` / `modelProfileName` / `provider` / `cost` / `latency` / `capability`）在 wire JSON 中零出现；fake AITaskRunWriter 捕获的 ai_task_runs 行包含 typed columns `task_type IN ('followup_generate', 'hint_generate')`、`validation_status='succeeded'`、`prompt_token_count` / `completion_token_count` / `latency_ms` / `model_profile_name`；non-AI action（`ask_question`、`session_wait`、`session_completed`）不触发 ai_task_runs 行写入 | `backend/cmd/api/practice_http_scenario_test.go::TestE2EP0050PracticeAssistantActionProvenanceAndTaskRuns` |

## 5 Phase 3 + Phase 4 — graceful degrade + 隐私 + regression

| 场景 ID | 场景 | Given | When | Then | 验证入口 |
|---------|------|-------|------|------|----------|
| `E2E.P0.051` | hint AI 失败 graceful degrade + 隐私红线 + legacy-negative grep | 用户 A 拥有 4 个 `practice_sessions(status='running', mode='assisted')`，分别配置 fake F3 + fake AIClient 模拟 4 种失败：① F3 `practice.turn.lightweight_observe` ResolveActive 返回 `registry.ErrPromptUnsupported`；② A3 fake AIClient 返回 `AI_PROVIDER_SECRET_MISSING`；③ A3 fake AIClient 返回 timeout；④ A3 fake AIClient 返回 invalid JSON content；log / metric / audit / ai_task_runs 收集器可读 | 用户 A 对每个 session `POST /practice/sessions/{sessionId}/events` 携带 `kind='hint_requested'`；并补一个 strict session 触发 `kind='hint_requested'` 用于隐私收集；运行 scoped legacy grep + lint | 每个 assisted session 返回 `200 + SessionEventResult{acknowledged: true, session: {status: 'running'}, assistantAction: {type: 'session_wait', hint: null, sessionStatus: 'running', provenance: <non-AI default 6 字段>}}`；状态码**不是** 502 / 503；DB `practice_sessions.status` 保持 `running`，`failure_code` 保持 NULL；`practice_turns.hint_text` 保持 NULL；ai_task_runs 写入 4 行 `task_type='hint_generate', validation_status='failed', error_code ∈ {AI_PROVIDER_CONFIG_INVALID, AI_PROVIDER_SECRET_MISSING, AI_PROVIDER_TIMEOUT, AI_OUTPUT_INVALID}`（error_code 来自 B1 enum，不含 raw provider message）；隐私集合中 `question_text` / `answer_text` / `hint_text` / AI prompt body / response body / provider secret 在 log / metric label / audit_events / outbox_events / practice_session_events.payload / ai_task_runs typed columns 中零出现；A3 `ai_task_*` metric label 命中 F1 allowlist 且不含 `feature_key` / prompt-rubric version / provider raw model id；scoped legacy grep 断言 `hint_disabled_globally` / `legacy_hint_policy` / `legacy_mode_assisted_value` / `legacy debrief replay value` / `warmup` / `single_drill` / `drill_builder` / `mistake_queue` / `growth_center` / 独立 `voice` route / `practiceModeCard` 在实现 / runtime 输出范围零出现；本 plan / BDD / test docs、negative tests 与 backend-practice spec D-20/D-21 prohibition rows 可枚举这些字面量作为禁止性断言；backend-practice spec.md / history.md 中"全 AI 一律 session=failed" / "所有 AI 调用必须 fail-closed"通用文字零残留 | `backend/cmd/api/practice_http_scenario_test.go::TestE2EP0051PracticeHintDegradeAndPrivacy` |

## 6 数据隔离与污染恢复

每个场景按 `test/scenarios/e2e/README.md` §5 / §3 / §6 / §8 约定：

- 数据隔离：每个场景使用独立的 `user_id` / `plan_id` / `session_id` / `clientEventId` 命名空间；不复用 `E2E.P0.022 ~ E2E.P0.043` 已占用的资源
- 清理顺序：cleanup 先删自身 `practice_session_events` → `practice_turns` → `feedback_reports`（如有）→ `async_jobs`（如有）→ `practice_sessions` → `practice_plans` → `idempotency_records` → `ai_task_runs` → `audit_events` → `outbox_events` → users
- 污染恢复：场景失败时按 README §8 顺序：① 清理场景自身资源；② 定位并恢复 shared 组件（F3 cache、idempotency_records pending；003 阶段无 runtime dispatcher，不需恢复 dispatcher 状态）；③ 仅在 ① ② 失败时 `test/scenarios/env-cleanup.sh && env-setup.sh` 全量重建
- 不预设 Helm chart / 外部 Git 平台名称；所有命令以本仓库脚本为真理源

## 7 与单元测试边界

本 BDD plan 验证用户可见行为切片（HTTP API + DB 状态 + ai_task_runs 写入 + log/metric/audit 红线 + provenance wire 边界 + scoped legacy grep）；不重复内部接口签名、序列化结构、错误映射、状态机内部分支等单元测试覆盖（详见 [test-plan](./test-plan.md)）。003 阶段同样不存在 runtime outbox→asynq dispatcher（与 002 一致），"dispatcher 集成测试" 不在本 BDD plan 范围内，由 future `backend-async-runner` plan 承接。

## 8 与 spec AC 映射

| spec AC | 覆盖场景 |
|---------|----------|
| C-7（hint assisted 生效） | `E2E.P0.048` |
| C-8（hint strict 被拒） | `E2E.P0.049` |
| C-8b（assisted+debrief 允许 hint） | `E2E.P0.048`（assisted+debrief 子断言）+ `E2E.P0.049`（strict+debrief 子断言） |
| C-12（AssistantAction provenance wire 边界） | `E2E.P0.050` |
| C-16（hint 路径隐私红线子集） | `E2E.P0.051` |
| C-17（F3/A3 narrowed fail-closed → 辅助 AI graceful degrade） | `E2E.P0.051` |
| D-11（隐私红线） | `E2E.P0.051` |
| D-20（retired 术语 + 旧 fail-closed 通用文字） | `E2E.P0.051` |
| D-36（spec §3.1 锁定，由 003 Phase 0 narrowing 引入：hint / lightweight_observe AI graceful degrade narrowing） | `E2E.P0.051` |
| D-37（plan-level：B4 `ai_task_runs.task_type` 扩值 `hint_generate`） | `E2E.P0.048` + `E2E.P0.050` + `E2E.P0.051` |
| D-38（spec §3.1 锁定，由 003 Phase 0 narrowing 引入：hint turn-lifecycle 边界） | `E2E.P0.048`（DB 不递增 turn_count / outbox 行数不增 / audit 行数不增 / hint_text 仅 assisted 成功写 子断言） |
