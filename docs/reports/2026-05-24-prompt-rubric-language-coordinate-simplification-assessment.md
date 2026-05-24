# Prompt Rubric Language Coordinate Simplification 交付复盘报告

> **日期**: 2026-05-24
> **审查人**: Codex

## 1 复盘范围与成功证据

- 范围：`prompt-rubric-registry/003-language-coordinate-simplification`，将 F3 prompt/rubric baseline 从默认 `multi + en` 双坐标收敛为 canonical `multi` truth source，并保留 `ResolveActive(..., "en")` 到 `multi` 的 runtime fallback。
- 主要变更：删除 39 个 `v0.1.0.en.*` 配置文件；seed migration 仅保留 13 个 prompt `multi` rows 与 13 个 rubric `multi` rows；更新 prompt/rubric lint、registry loader/resolver/seed coverage tests、spec/README/plan/checklist。
- 成功证据：`python3 -m pytest scripts/lint/prompt_lint_test.py -q` 15 passed；`python3 -m pytest scripts/lint/rubric_lint_test.py -q` 6 passed；`go test ./backend/internal/ai/registry -count=1` passed；`make lint-prompts` / `make lint-rubrics` both clean；`python3 scripts/lint/migrations_lint.py` passed。
- 生命周期证据：`validate_context.py` passed；`sync-doc-index --check` zero drift；`make docs-check` passed；`git diff --check` passed；`make lint` passed；`DATABASE_URL=... make migrate-check` 当前输出为 `migration lint: ok`。

## 2 会话中的主要阻点/痛点

- 旧 lint/test 把重复 `en` 文件当成正确性。
  - **证据**：新增负向 fixture 后，旧 `prompt_lint` / `rubric_lint` 对重复 `en` override 返回 clean；Go red gate 显示 `SnapshotSize: want 13, got 26`。
  - **影响**：如果只改文档不改 lint，重复维护成本会继续回流。

- 邻近测试仍携带旧 spec/version 与 exact `en` 假设。
  - **证据**：`go test ./backend/internal/ai/registry -count=1` 首次失败于 `spec.md missing "> **版本**: 2.6"` 和 `GetPrompt(..., "en")` 的未知版本断言。
  - **影响**：focused tests 先绿后，全包测试仍可能 false-red，说明 current truth 反查不能只覆盖显式计划列出的测试名。

- stale grep 需要分类 residual，而不是简单追求字符串零出现。
  - **证据**：`v0.1.0.en` 仍需要出现在负向 fixture 与 003 plan 的删除证明中；这些不是 active storage requirement。
  - **影响**：若 gate 只按原始 grep 输出判定，会误伤必要的负向测试和交付记录。

- `migrate-check` 当前行为与命名存在轻微语义落差。
  - **证据**：带 `DATABASE_URL` 执行后输出为 `migration lint: ok`，没有显示实际连接 dev-stack Postgres。
  - **影响**：本次可作为 static migration gate 证明，但不能当作真实数据库迁移链运行证明。

## 3 根因归类

- 重复语言坐标缺少 allowlist 语义：**类别** `spec-plan`。旧 spec/README/test 把“有两个 language coordinates”当成 baseline completeness。
- 邻近测试漂移：**类别** `spec-plan`。当前 checklist 已覆盖 focused registry tests，但全包消费者里的 preflight/version/exact debug path 没在 Phase 2 明确列出。
- stale grep residual 误判风险：**类别** `spec-plan`。负向 fixture 与删除证明需要被允许并分类。
- migrate-check 语义落差：**类别** `README`。migration README 或 Make target 需要明确 static lint 与 live DB chain 的边界。

## 4 对流程资产的改进建议

- 在后续 prompt-rubric registry plan 模板中加入“full package adjacent test”gate。
  - **落点**：spec-plan
  - **优先级**：high

- 将 stale-contract negative search 写成“active positive requirement = fail；negative fixture / deletion proof / historical evidence = classify”的规则。
  - **落点**：spec-plan
  - **优先级**：medium

- 明确 `make migrate-check` 当前是否应连接数据库；如只是 static lint，新增或重命名 live migration gate。
  - **落点**：README / Make target
  - **优先级**：medium

## 5 建议优先级与后续动作

- 最高优先级：下一轮 F3 计划进入真实 Model Profile / evals 前，先把 full registry package test 和 stale-contract 分类 gate 写入计划 checklist，避免 focused green 后遗漏邻近消费者。
- 次优先级：为 migrations 补一个明确的 live Postgres migration-chain gate，或把当前 `migrate-check` 的 static-only 行为写进 `migrations/README.md` 与相关 plan gate。
