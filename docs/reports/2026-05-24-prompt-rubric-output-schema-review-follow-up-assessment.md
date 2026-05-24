# Prompt Rubric Output Schema Review Follow-up 交付复盘报告

> **日期**: 2026-05-24
> **审查人**: Codex

## 1 复盘范围与成功证据

- 范围：修复 `feat/prompt-rubric-registry-output-schema-0523` review 暴露的两个 L2 findings：`jd_match.recommendation.posted` required drift 与 `prompt_lint.py` invalid schema traceback。
- 成功证据：
  - `python3 -m pytest scripts/lint/prompt_lint_test.py -q` → `13 passed`
  - `python3 scripts/lint/prompt_lint.py` → `prompt_lint: 26 files clean`
  - `python3 scripts/lint/rubric_lint.py` → `rubric_lint: 26 files clean`
  - `python3 scripts/lint/migrations_lint.py` → `migration lint: ok`
  - `DATABASE_URL=postgres://easyinterview:dev@localhost:5432/easyinterview?sslmode=disable make migrate-check` → `migration lint: ok`
  - `go test ./backend/internal/ai/registry -count=1` → pass
  - `go test ./backend/cmd/api -run TestJDMatchA3F3AdapterUsesRegistryProfilesForSearchAndRecommendation -count=1` → pass
  - `go test ./backend/internal/ai/aiclient/observability -run TestDecorator_OutputSchema -count=1` → pass
  - `validate_context.py --context docs/spec/prompt-rubric-registry/plans/002-output-schema-contract/context.yaml --docs-root docs --target backend` → pass
  - `sync-doc-index.py --check` → zero drift
  - `git diff --check` → pass

## 2 会话中的主要阻点/痛点

- `jd_match.recommendation` schema 自洽但没有反查生产 caller 输入。
  - **证据**：schema required 曾包含 `posted`，但 `compactJDMatchJobsPool` 未传入该字段，`llmRecommendation.Posted` 也是 optional。
  - **影响**：真实 A3/F3 path 可能要求模型编造 freshness，或被 output schema fail-close 拒绝。
- lint negative path 缺少主流程级 regression。
  - **证据**：直接函数级测试能发现缺 `description`，但主流程仍在 renderer 中触发 `KeyError` traceback。
  - **影响**：invalid schema 的错误反馈不稳定，降低 lint gate 的可诊断性。

## 3 根因归类

- output schema required 字段缺少 runtime input / consumer optionality 反查。
  - **类别**：spec-plan
- `prompt_lint.py` 没有把 schema subset / contract validation 和 renderer 输入边界隔离。
  - **类别**：README / tooling
- L2 review 发现问题后需要把该模式沉淀到 Bug pattern。
  - **类别**：no repo change needed（本次已更新 `PATTERNS.md`）

## 4 对流程资产的改进建议

- 在后续 prompt schema plan/review gate 中明确：`required` 字段必须同时满足 schema、生产 caller 输入、consumer struct/parser 与 persistence contract。
  - **落点**：spec-plan
  - **优先级**：medium
- 对 lint 工具新增 invalid-input integration fixture 作为默认测试模式，避免只测 helper 函数。
  - **落点**：README / tooling
  - **优先级**：medium

## 5 建议优先级与后续动作

- 下一轮最值得做：继续用 `plan-code-review prompt-rubric-registry/002-output-schema-contract --base-rev ab258261b026eaf99f488bc3f2f8175a61641c9d` 做最终确认，重点扫所有 schema required 字段是否都有 runtime input 或 consumer 必要性支撑。
- 可延后：将 lint invalid-input fixture 模式推广到 rubric / migration lint，避免其他 lint 工具出现 traceback 式失败。
