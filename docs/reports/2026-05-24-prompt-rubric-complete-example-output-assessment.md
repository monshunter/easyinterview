# Prompt Rubric Complete Example Output 交付复盘报告

> **日期**: 2026-05-24
> **审查人**: Codex

## 1 复盘范围与成功证据

- 范围：修复 `prompt-rubric-registry/002-output-schema-contract` 中 prompt example 过于简陋的问题，将 13 个 chat feature_key × 2 language coordinates 的 prompt example 从 required-only 占位 JSON 升级为完整代表性 JSON output。
- 代码证据：`scripts/lint/prompt_lint.py` 的 renderer 现在输出 schema-declared required + optional properties，并显式要求模型返回 JSON value 而不是 JSON Schema / OpenAPI schema。
- Prompt 证据：26 个 `config/prompts/**/v0.1.0*.md` 已重渲染，26 个 YAML `template_hash` 与 `migrations/000002_seed_baseline_prompt_rubric_versions.up.sql` 里已有 prompt seed row 的 body/hash 已同步。
- 验证证据：
  - `python3 -m pytest scripts/lint/prompt_lint_test.py -q` pass
  - `make lint-prompts` pass
  - `rg -n '"string"|: 1,|Example JSON:' config/prompts -g '*.md'` 0 matches
  - `go test ./backend/internal/ai/registry -run TestF3ReportGenerateAndAssessmentPreflight -count=1` pass
  - `validate_context.py` + `sync-doc-index.py --check` pass
  - `make docs-check` pass
  - `make lint` pass
  - `git diff --check` pass
- 文档证据：`prompt-rubric-registry` spec 升至 v2.6；`002-output-schema-contract` plan/checklist 升至 v1.3；新增 [BUG-0096](../bugs/BUG-0096.md)。

## 2 会话中的主要阻点/痛点

- 原 plan/README 把 example 定义为 minimal JSON，导致 lint gate 稳定地重渲染低信号示例。
  - **证据**：`config/prompts/README.md` 原文写 “minimal valid JSON”；`example_for_schema` 原实现只遍历 `required` 字段。
  - **影响**：26 个 prompt 都有合法但教学价值不足的 example，用户指出后需要同时修 renderer、prompt、hash、migration 和 owner docs。

- 初始 focused Go test filter 写错，返回 `[no tests to run]`。
  - **证据**：`go test ./backend/internal/ai/registry -run TestBackendReviewPreflight -count=1` 返回 no tests；随后通过 `go test ... -list` 找到真实测试名 `TestF3ReportGenerateAndAssessmentPreflight` 并重跑通过。
  - **影响**：如果不做 no-op 反查，会把无效测试当成证据。

## 3 根因归类

- 根因 1：prompt renderer 的验收只覆盖 schema-valid，不覆盖 example 的完整性与业务有效性。
  - **类别**：spec-plan

- 根因 2：测试命令证据需要保持 no-op guard，尤其是手写 focused `-run` filter。
  - **类别**：skill / no repo change needed（当前会话已用 `-list` 反查纠正，既有 PATTERNS 模式 4 已覆盖 no-op 风险）

## 4 对流程资产的改进建议

- 对后续 prompt / evaluator plan，将 example quality 明确拆成单独验收项：required + optional 完整性、业务形态值、not schema output 负向提示。
  - **落点**：spec-plan
  - **优先级**：high

- 保留 focused test no-op guard；当 `go test -run` 输出 `[no tests to run]` 时，必须用 `go test -list` 或源码反查纠正后再计入证据。
  - **落点**：已有 BUG pattern / skill 习惯
  - **优先级**：medium

## 5 建议优先级与后续动作

- 推荐下一步：进入 `prompt-rubric-registry/003-real-model-profile-and-evals` 前，先审一遍真实模型评估 plan 的 prompt/eval example gate，确保 eval case 也使用完整 output 样例而不是 schema 或占位 JSON。
- 备选：若继续做 F3 L2 hardening，可对所有 prompt/rubric README 中的 “example” 语义做一次负向搜索，确认没有残留 minimal placeholder 口径。
