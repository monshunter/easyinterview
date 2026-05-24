# Prompt Rubric JD-Match Seed Migration 交付复盘报告

> **日期**: 2026-05-24
> **审查人**: Codex

## 1 复盘范围与成功证据

- 范围：`prompt-rubric-registry/002-output-schema-contract` L2 remediation，修复 seed migration 静态门禁漏掉 `jd_match.recommendation` / `jd_match.search` prompt/rubric baseline rows 的 false-green。
- 成功证据：
  - `go test ./backend/internal/ai/registry -run TestSeedMigrationCoversBaselineFeatureKeys -count=1 -v` red → green。
  - `go test ./backend/internal/ai/registry -count=1` → pass。
  - `python3 scripts/lint/prompt_lint.py`、`python3 scripts/lint/rubric_lint.py`、`python3 scripts/lint/migrations_lint.py` → pass。
  - `DATABASE_URL=postgres://easyinterview:***@localhost:5432/easyinterview?sslmode=disable make migrate-check` → pass。
  - `validate_context.py`、`sync-doc-index.py --check`、`git diff --check` → pass。

## 2 会话中的主要阻点/痛点

- Static gate false-green。
  - **证据**：原 `TestSeedMigrationCoversBaselineFeatureKeys` 写死 11 个 feature_key / 22 行；13 个 active chat feature_key 中的 JD-Match 两项未被期望集合覆盖。
  - **影响**：completed checklist 看似收口，但 DB-backed baseline seed 会缺 prompt_versions / rubric_versions rows。

- `make migrate-check` 首次缺环境变量。
  - **证据**：无 `DATABASE_URL` 时命令只跑到 migration lint 后失败；dev-stack Postgres 已运行，补默认连接串后完整 gate 通过。
  - **影响**：如果只记录第一次失败，容易把可执行 gate 误判为环境不可用。

## 3 根因归类

- Seed coverage gate 不是 truth-source-driven。
  - **类别**：spec-plan / test

- prompt lint 的 seed hash check 只能验证已存在 row。
  - **类别**：test

## 4 对流程资产的改进建议

- 对 seed / fixture / generated artifact gate 建立 exact set compare 模式。
  - **落点**：`docs/bugs/PATTERNS.md`（已补 Pattern 9）
  - **优先级**：high

- plan-code-review 遇到 completed plan 的 seed/hash gate 时，必须确认“完整缺行也会失败”。
  - **落点**：`plan-code-review` skill 后续可进一步固化为检查项
  - **优先级**：medium

## 5 建议优先级与后续动作

- 下一轮优先动作：用 `/plan-code-review prompt-rubric-registry/002-output-schema-contract backend` 做一次完整复查，确认 v1.4 文档、BUG-0097、retrospective 与实际 diff 没有遗漏收口项。
- 可延后动作：把 seed exact-set compare 模式抽成共享 helper，供后续 fixture/generated artifact gate 复用。
