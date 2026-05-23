# Prompt Rubric Output Schema L2 Remediation 交付复盘报告

> **日期**: 2026-05-24
> **审查人**: Codex

## 1 复盘范围与成功证据

- 范围：`plan-code-review prompt-rubric-registry/002-output-schema-contract --fix`，针对 A3 `validateOutputSchema` 在 schema-valid JSON 后追加 trailing prose 时未 fail-close 的 L2 缺口进行修复。
- 代码证据：`backend/internal/ai/aiclient/observability/decorator.go` 在第一次 JSON decode 后继续读取并只接受 EOF。
- 测试证据：新增 `TestDecorator_OutputSchemaRejectsTrailingTokens`；Red 阶段先失败，Green 后通过。
- 验证证据：
  - `go test ./backend/internal/ai/aiclient/observability -run 'TestDecorator_OutputSchemaRejectsTrailingTokens' -count=1` pass
  - `go test ./backend/internal/ai/aiclient/observability -run 'TestDecorator_OutputSchema' -count=1` pass
  - `go test ./backend/internal/ai/aiclient/observability -list 'TestDecorator_OutputSchema'` 确认 focused gate 非 no-op
  - `go test ./backend/internal/ai/registry/... ./backend/internal/ai/aiclient/... ./backend/internal/targetjob/... ./backend/internal/resume/jobs/... ./backend/internal/review/... ./backend/internal/practice/... ./backend/internal/debrief/... ./backend/internal/jdmatch/... ./backend/cmd/api/... -race` pass
  - `python3 -m pytest scripts/lint/prompt_lint_test.py -q` pass
  - `make lint-prompts` pass
  - `make lint` pass
- 文档证据：`002-output-schema-contract` plan/checklist 升至 v1.2 并记录 Phase 6.3 L2 remediation；新增 [BUG-0095](../bugs/BUG-0095.md) 与 Bug pattern 8。

## 2 会话中的主要阻点/痛点

- Phase 6 原证据覆盖了 invalid JSON、required、enum 与 array valid，但没有覆盖「首个 JSON value 合法、后续追加非空内容」这一 strict JSON 负向路径。
  - **证据**：`validateOutputSchema` 原实现只调用一次 `json.Decoder.Decode`；新增 focused test 在 Red 阶段失败。
  - **影响**：A3 validation status 与 metric/log 信号可能误判为通过，错误延后到业务 parser。

- `plan-code-review` 必须把旧 PASS 证据拆成具体 coverage row 审查，才能发现 validator 语义缺口。
  - **证据**：计划已是 completed 且已有 Phase 6 绿测，但 artifact-level 反查发现测试矩阵缺 trailing-token 行。
  - **影响**：若只信 checklist 与历史 PASS，会漏掉 runtime fail-close 语义。

## 3 根因归类

- 根因 1：Phase 6 test matrix 缺少 strict JSON trailing token negative case。
  - **类别**：spec-plan

- 根因 2：运行时 JSON validator 使用 streaming decoder 时没有执行 EOF 检查。
  - **类别**：no repo change needed（已由代码修复，并通过 BUG/PATTERNS 沉淀）

## 4 对流程资产的改进建议

- 对后续 AI output schema / provider response validator 计划，测试矩阵显式列出「single JSON document」负向项。
  - **落点**：spec-plan
  - **优先级**：high

- 在 L2 review 中对所有 `json.Decoder.Decode` validator 路径加入 trailing-token 反查。
  - **落点**：skill
  - **优先级**：medium

- 保留 BUG-0095 的 Pattern 8 作为后续 bugfix 的快速检查入口。
  - **落点**：README / bug pattern
  - **优先级**：medium

## 5 建议优先级与后续动作

- 推荐下一步：在后续 `prompt-rubric-registry/003-real-model-profile-and-evals` 或 A3 provider/eval 相关计划中，把 strict JSON single-document negative case 写入 test matrix，避免真实模型切换后把 provider prose 误判为 schema-valid。
- 备选：若短期继续做 L2 hardening，可 grep 仓库内其他 `json.NewDecoder(...).Decode` validator 路径，补同类 EOF 检查。
