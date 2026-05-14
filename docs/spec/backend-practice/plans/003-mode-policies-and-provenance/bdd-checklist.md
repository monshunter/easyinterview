# 003 — Mode Policies and Provenance BDD Checklist

> **版本**: 1.0
> **状态**: active
> **更新日期**: 2026-05-14

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## E2E.P0.048 assisted hint 主路径 × goal 矩阵

- [ ] 在 `backend/cmd/api/practice_http_scenario_test.go` 新增 `TestE2EP0048PracticeHintAssistedAcrossGoals`：4 个 user/plan/session 命名空间（goal 分别为 baseline / retry_current_round / next_round / debrief，mode 全部 assisted）
- [ ] 准备 fake F3 RegistryClient + fake AIClient + fake AITaskRunWriter；fake AIClient 为 `practice.turn.lightweight_observe` 返回合法 hint JSON Content
- [ ] 实现 setup：调用 `backend/internal/store/practice` 真实 SQL fake 或 mem store 写入 4 个 ready plan + 4 个 running session；记录 currentTurn.id
- [ ] 实现 trigger：对每个 session 发起 1 次 `POST /practice/sessions/{sessionId}/events kind='hint_requested'`；断言 HTTP status=200
- [ ] 实现 verify：
  - response.assistantAction.type='show_hint' + hint 非空 + provenance 6 字段
  - DB practice_turns.hint_text 非空（assisted 成功路径）
  - DB practice_turns.status / turn_index / practice_sessions.turn_count 不变
  - outbox_events 行数不增（`event_name='practice.turn.completed'` 在该 session 下零行）
  - audit_events 行数不增（append 路径不写 audit）
  - practice_session_events 写入 1 行 `kind='hint_requested'`，payload 不含 hint_text 明文
  - ai_task_runs 写入 1 行 `task_type='hint_generate', validation_status='succeeded'`
- [ ] 实现 cleanup：按 [bdd-plan §6](./bdd-plan.md#6-数据隔离与污染恢复) 顺序删除自身资源
- [ ] 执行 `cd backend && go test ./cmd/api -run TestE2EP0048PracticeHintAssistedAcrossGoals -count=1`
- [ ] 记录验证证据到 plan §3.6 L2 修订说明（如经过 L2 review）或本 checklist 收口段

## E2E.P0.049 strict hint 拒绝 × goal 矩阵

- [ ] 在 `backend/cmd/api/practice_http_scenario_test.go` 新增 `TestE2EP0049PracticeHintStrictRefusalAcrossGoals`：4 个 user/plan/session 命名空间（goal 分别为 baseline / retry_current_round / next_round / debrief，mode 全部 strict）
- [ ] 准备 fake F3 RegistryClient + fake AIClient：fake AIClient 设置调用计数器，断言**不**被调用
- [ ] 实现 setup：写入 4 个 ready plan + 4 个 running session
- [ ] 实现 trigger：对每个 session 发起 1 次 `POST /practice/sessions/{sessionId}/events kind='hint_requested'`
- [ ] 实现 verify：
  - response HTTP status=409
  - response.error.code='PRACTICE_SESSION_CONFLICT'
  - response.error.details.policy='hint_disabled_in_mode'
  - response.error.details.mode='strict'
  - DB practice_turns.hint_text 保持 NULL
  - session 状态不变
  - practice_session_events 存在 finalized `kind='hint_requested'` 行，payload 仅含 sanitized conflict envelope，且不存在 `payload.pending=true` stuck row
  - 同 `clientEventId` 重试 replay 同一 409，不重复 reserve，不调 AI；mismatch 仍返回 conflict 且不泄露首次 payload
  - ai_task_runs 不写新行
  - outbox_events / audit_events 行数不增
  - fake AIClient.Complete 调用次数为 0
- [ ] 实现 cleanup：按隔离顺序删除资源
- [ ] 执行 `cd backend && go test ./cmd/api -run TestE2EP0049PracticeHintStrictRefusalAcrossGoals -count=1`
- [ ] 记录验证证据

## E2E.P0.050 AssistantAction provenance wire 边界 + ai_task_runs runtime 字段

- [ ] 在 `backend/cmd/api/practice_http_scenario_test.go` 新增 `TestE2EP0050PracticeAssistantActionProvenanceAndTaskRuns`
- [ ] 准备 fake F3 + fake AIClient（同时配置 `practice.session.follow_up` 与 `practice.turn.lightweight_observe`）+ fake AITaskRunWriter
- [ ] 实现 setup：写入 1 个 ready plan (mode=assisted, goal=baseline) + 1 个 running session；turn_count / question_budget 配置为 2，方便触发 session_completed
- [ ] 实现 trigger 序列：① answer_submitted（→ ask_follow_up，AI 调用）；② hint_requested（→ show_hint，AI 调用）；③ turn_skipped（→ ask_question，non-AI）；④ session_paused（→ session_wait，non-AI）；⑤ answer_submitted 达 question_budget（→ session_completed，non-AI）
- [ ] 实现 verify：
  - 每次 response.assistantAction.provenance JSON keys 集合严格等于 `{promptVersion, rubricVersion, modelId, language, featureFlag, dataSourceVersion}`
  - 任何 runtime 字段（`featureKey` / `feature_key` / `modelProfileName` / `provider` / `cost` / `latency` / `capability`）在 wire JSON 中零出现
  - fake AITaskRunWriter 捕获 2 行 ai_task_runs：① task_type='followup_generate'；② task_type='hint_generate'；两行都包含 prompt_token_count / completion_token_count / latency_ms / model_profile_name / validation_status='succeeded'
  - 后 3 个 non-AI action 不触发新 ai_task_runs 行
- [ ] 实现 cleanup
- [ ] 执行 `cd backend && go test ./cmd/api -run TestE2EP0050PracticeAssistantActionProvenanceAndTaskRuns -count=1`
- [ ] 记录验证证据

## E2E.P0.051 hint AI 失败 graceful degrade + 隐私 + legacy-negative

- [ ] 在 `backend/cmd/api/practice_http_scenario_test.go` 新增 `TestE2EP0051PracticeHintDegradeAndPrivacy`
- [ ] 准备 fake F3 RegistryClient + fake AIClient + fake AITaskRunWriter；按场景分支配置：① F3 `registry.ErrPromptUnsupported`；② A3 secret missing；③ A3 timeout；④ A3 invalid output；外加 1 个 strict session 用于隐私 baseline
- [ ] 实现 setup：写入 4 个 assisted plan/session 分别绑定 4 种失败注入 + 1 个 strict plan/session
- [ ] 实现 trigger：对每个 assisted session 发起 1 次 `hint_requested`；对 strict session 发起 1 次 `hint_requested`
- [ ] 实现 verify：
  - 4 个 assisted response 全部 HTTP status=200（**不是** 502/503）
  - response.assistantAction.type='session_wait' + hint=null + provenance 6 wire 字段（non-AI default）
  - response.session.status='running'
  - DB practice_sessions.status='running'（不进入 failed）
  - DB practice_sessions.failure_code IS NULL
  - DB practice_turns.hint_text IS NULL
  - ai_task_runs 写入 4 行 task_type='hint_generate', validation_status='failed', error_code IN ('AI_PROVIDER_CONFIG_INVALID','AI_PROVIDER_SECRET_MISSING','AI_PROVIDER_TIMEOUT','AI_OUTPUT_INVALID')；error_code 来自 B1 enum，不含 raw provider message
  - 隐私集合（log / metric label / audit_events / outbox_events / practice_session_events.payload / ai_task_runs typed columns / service log fields）扫描 `question_text` / `answer_text` / `hint_text` / `prompt body` / `response body` / `provider secret` 零出现
  - A3 `ai_task_*` metric label 命中 F1 allowlist 且不含 `feature_key` / prompt-rubric version / provider raw model id
  - hint 路径不写 audit_events 行（append 路径红线，已由 002 落地，本 scenario 补反查）
  - scoped legacy grep（in-test 执行）断言：`hint_disabled_globally` / `legacy_hint_policy` / `legacy_mode_assisted_value` / `legacy debrief replay value` / `warmup` / `single_drill` / `drill_builder` / `mistake_queue` / `growth_center` / 独立 `voice` route / `practiceModeCard` 在实现 / runtime 输出范围零出现；本 plan / BDD / test docs、negative tests 与 backend-practice spec D-20/D-21 prohibition rows 可枚举这些字面量作为禁止性断言；backend-practice `spec.md` / `history.md` 中 "全 AI 一律 session=failed" / "所有 AI 调用必须 fail-closed"通用文字零残留
- [ ] 实现 cleanup
- [ ] 执行 `cd backend && go test ./cmd/api -run TestE2EP0051PracticeHintDegradeAndPrivacy -count=1`
- [ ] 记录验证证据

## 收口

- [ ] `cd backend && go test ./cmd/api -run 'TestE2EP0048|TestE2EP0049|TestE2EP0050|TestE2EP0051' -count=1` 全绿
- [ ] `python3 scripts/lint/backend_practice_legacy.py --repo-root . --phase all` 通过（含 003 scoped legacy 反查项）
- [ ] `python3 -m pytest scripts/lint/backend_practice_legacy_test.py -q` 通过
