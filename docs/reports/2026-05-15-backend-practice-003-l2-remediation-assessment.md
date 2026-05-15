# Backend Practice 003 L2 Remediation 交付复盘报告

> **日期**: 2026-05-15
> **审查人**: Codex

## 1 复盘范围与成功证据

- 范围：`backend-practice/003-mode-policies-and-provenance` L2 code review remediation，聚焦 P0.048-P0.051 BDD evidence drift、hint replay 隐私红线、A3 callErr task-run metadata。
- 2026-05-15 follow-up：补充覆盖 strict / unknown 409 finalized replay payload 的 sanitized envelope、003 scenario runtime assets 的 legacy-negative 扫描范围、P0.048-P0.051 shell wrapper 的真实 exit code / PASS gate，以及 appendSessionEvent event-level replay snapshot。
- 成功证据：
  - `cd backend && go test ./cmd/api -run 'TestE2EP0048|TestE2EP0049|TestE2EP0050|TestE2EP0051' -count=1`
  - `cd backend && go test ./internal/store/practice -run 'TestSQLRepositoryAppendSessionEventWritesHintTextForAssistedSuccess|Test.*AppendSessionEvent' -count=1`
  - `cd backend && go test ./internal/ai/aiclient/observability -count=1`
  - `cd backend && go test ./internal/practice -run 'TestApplyHintAI|TestHandleHintRequested|TestServiceAppliesHintAI|TestServiceSkipsHintAI' -count=1`
  - `cd backend && go test ./internal/practice -run 'TestAppendSessionEventReplayReturnsStoredErrorBeforeResult' -count=1`
  - `cd backend && go test ./internal/store/practice -run 'TestMarshalAppendEventPayloadRedactsHintButReplayPayloadKeepsSnapshot|TestSQLRepositoryReserveSessionEventReplaysOriginalHintSnapshot|TestSQLRepositoryAppendSessionEventWritesHintTextForAssistedSuccess' -count=1`
  - `cd backend && go test ./internal/practice ./internal/store/practice ./cmd/api -count=1`
  - `cd backend && go test ./internal/migrations -count=1`
  - `migrations/lint.sh`
  - `cd backend && go test ./... -count=1`
  - `make codegen-check`
  - `make validate-fixtures`
  - `python3 scripts/lint/conventions_drift.py --repo-root .`
  - `make lint-events`
  - `make codegen-events-check`
  - `python3 scripts/lint/backend_practice_legacy.py --repo-root . --phase all`
  - `python3 -m pytest scripts/lint/backend_practice_legacy_test.py -q`
  - `git diff --check`
  - `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check`
  - `make docs-check`
  - P0.048 / P0.049 / P0.050 / P0.051 scenario `setup → trigger → verify → cleanup` 全 PASS。
- 环境阻断：`make migrate-check` 的 migration lint 通过后因当前 shell 未配置 `DATABASE_URL` 停止，需要按 `deploy/dev-stack/README.md` 提供数据库连接后重跑。
- 关联 Bug：[BUG-0058](../bugs/BUG-0058.md)、[BUG-0059](../bugs/BUG-0059.md)。

## 2 会话中的主要阻点/痛点

- BDD checklist 的完成状态与 executable assertions 不一致。
  - **证据**：P0.048-P0.051 文档声明 DB hint_text、strict no-pending、5 action provenance、4 degrade paths；审查前测试只覆盖主干 response。
  - **影响**：completed plan 隐藏隐私与观测 drift，直到 L2 反向补断言才暴露。
- replay payload 隐私边界没有单独建模。
  - **证据**：P0.048 补强后首次失败，`practice_session_events.payload` 中保存 `AssistantAction.Hint` 明文。
  - **影响**：违反 backend-practice D-11 / D-36 / BDD P0.048 隐私红线，且会污染后续回放与排障数据面。
- A3 decorator 对 callErr metadata 的 fallback 不完整。
  - **证据**：P0.051 注入 shared `AI_PROVIDER_SECRET_MISSING` / `AI_PROVIDER_TIMEOUT` 后，failed `hint_generate` row 初始缺少 `error_code` 与 invalid validation status。
  - **影响**：业务层能 degrade，但运维 typed columns 不能准确反映 B1 错误码。
- 2026-05-15 follow-up 暴露 scenario wrapper 本身会假绿。
  - **证据**：P0.048-P0.051 `trigger.sh` 使用 `go test ... | tee ...`，POSIX `sh` 下返回 `tee` 的状态；`verify.sh` 只检查 test 名称、包名和 `no tests to run`，不要求 `--- PASS` / `ok`，也不拒绝 `FAIL`。
  - **影响**：即使 Go scenario 失败，wrapper 也可能留下可通过的日志证据。
- 2026-05-15 reviewer follow-up 暴露 replay source 仍未绑定原事件。
  - **证据**：SQL-backed strict 409 replay 同时返回零值 `ReplayResult` 与 `ReplayError`，service 先返回 success；同一 turn 多个 assisted hint 覆盖 `practice_turns.hint_text` 后，首个 `clientEventId` replay 会拿到后续 hint。
  - **影响**：违反 appendSessionEvent per-event idempotency contract，且说明上一版隐私修复把 redacted payload 与 replay source 的边界混在一起。

## 3 根因归类

- BDD evidence drift。
  - **类别**：spec/plan
  - 003 checklist 把 scenario 名称和 wrapper 执行当成完整证明，没有要求每条 BDD assertion 映射到代码断言。
- Replay privacy model gap。
  - **类别**：spec/plan
  - plan 强调 `practice_session_events.payload` 不含 hint 明文，但实现没有区分 replay payload 与业务持久化字段。
- Observability fallback gap。
  - **类别**：no repo change needed
  - 修复已在 A3 decorator 中完成；后续只需保持 callErr metadata regression。
- Scenario wrapper evidence gap。
  - **类别**：spec/plan
  - BDD scenario asset 的 executable gate 没有把 test process exit code、PASS line 和 package ok 作为证据要求。
- Event-level replay snapshot gap。
  - **类别**：spec/plan
  - 003 plan 只要求 `payload` 不泄漏 hint 明文，没有显式要求 replay snapshot 必须绑定原事件且覆盖同一 turn 多 hint replay。

## 4 对流程资产的改进建议

- 在 `plan-code-review` 的 Deep Evidence 检查中，抽样读取 BDD test body，而不是只读取 scenario wrapper。
  - **落点**：skill
  - **优先级**：high
- backend-practice 后续 plan 的 BDD checklist 对隐私红线使用“payload 字段 + 明文值”双轨断言。
  - **落点**：spec-plan
  - **优先级**：high
- A3 observability 的 future plan 增加一条 generic callErr metadata fallback regression，覆盖 `Complete` / `Transcribe` / `Synthesize`。
  - **落点**：spec-plan
  - **优先级**：medium
- 后续 BDD wrapper 模板应默认避免 `go test | tee`，并要求 `verify.sh` 同时断言 `=== RUN`、`--- PASS`、package `ok`，显式拒绝 `FAIL` / `no tests to run`。
  - **落点**：spec-plan
  - **优先级**：high
- 后续所有 `clientEventId` replay 修复应同时写成功 replay、错误 replay、payload redaction、mutable business snapshot 覆盖四类 regression。
  - **落点**：spec-plan
  - **优先级**：high

## 5 建议优先级与后续动作

- 最高优先级：在下一次 backend-practice 或 AI-observability L2 review 中，把 BDD checklist 的每条 assertion 逐项映射到 test body，尤其是隐私、replay、typed columns 和 negative paths。
- 同等优先级：把 scenario wrapper 脚本也纳入 L2 evidence review，不只读取 Go test body；wrapper 必须证明真实测试进程成功。
- 中优先级：把 `show_hint` replay payload 脱敏 + event-level `replay_payload` snapshot 模式作为 backend-practice 后续 plan 的显式模式，避免未来 report / debrief 等派生文本重复踩同一类问题。
