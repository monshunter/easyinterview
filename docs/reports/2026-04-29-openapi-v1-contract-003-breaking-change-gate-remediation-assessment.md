# OpenAPI v1 Contract 003 Breaking Change Gate Remediation 交付复盘报告

> **日期**: 2026-04-29
> **审查人**: Codex

## 1 复盘范围与成功证据

- 本次交付覆盖 `/plan-code-review --fix` 对 [openapi-v1-contract/003-breaking-change-gate](../spec/openapi-v1-contract/plans/003-breaking-change-gate/plan.md) 的两项 L2 finding：composition schema diff 漏检、privacy export 白名单 history gate 默认 ref 语义错误。
- 已完成原地修订：003 plan v1.2 / checklist v1.1 追加并完成 Phase 2.5，状态恢复 `completed`；[history.md](../spec/openapi-v1-contract/history.md) 追加 1.7 行；[BUG-0001](../bugs/BUG-0001.md) 已建档。
- 验证证据：
  - `python3 -m unittest scripts.lint.openapi_diff_test -v`：24/24 PASS。
  - `make codegen-check && make validate-fixtures && make openapi-diff`：PASS；`openapi-diff` summary 为 `breaking=0, additive=0, informational=0`。
  - `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check`：Zero drift detected。
  - `git diff --check`：PASS。

## 2 会话中的主要阻点/痛点

- L2 review 才发现 wrapper 没覆盖当前 contract 实际使用的 `oneOf` / `allOf`。
  - **证据**：新增 red tests 证明旧实现对 nullable `oneOf` type 变更和 paginated `allOf` `$ref` 变更返回 zero findings。
  - **影响**：v1.0.0 freeze gate 曾可能放过 schema-level breaking change。
- `history-ref=HEAD` 的问题已在 003 原交付复盘中作为痛点出现，但当时只作为后续建议，没有进入同 plan 的立即 remediation。
  - **证据**：本次新增 committed feature-branch 测试复现了已提交 history 行仍被误判 `history-not-incremented`。
  - **影响**：后续 privacy export P1 切换如果按正常 feature branch commit 流程推进，会被本地 gate 错误阻塞。

## 3 根因归类

- wrapper 测试矩阵缺少真实 OpenAPI schema 形态样本。
  - **类别**：spec/plan。
- “同 PR 增量”检查在设计时使用了 self-check 工作树模型，未明确 base branch / merge-base 默认语义。
  - **类别**：spec/plan。
- L2 remediation 需要同时修改代码、Makefile、README、history、plan/checklist、Bug 记录和 retrospective，收尾面较宽。
  - **类别**：无须仓库改动；本次已经按现有 skill 流程完成。

## 4 对流程资产的改进建议

- 在后续 contract gate 类 plan 的 Phase 2 自检中，强制列出“当前契约实际使用的复杂 schema 形态”作为回归样本，例如 `oneOf` nullable、`allOf` envelope、`anyOf` 扩展。
  - **落点**：spec-plan
  - **优先级**：medium
- 若未来同类问题再次出现，可把 [BUG-0001](../bugs/BUG-0001.md) 提炼进 [PATTERNS.md](../bugs/PATTERNS.md)：自研 diff/lint gate 需要覆盖真实 AST 结构与 base-ref 语义。
  - **落点**：README / Bug pattern
  - **优先级**：low

## 5 建议优先级与后续动作

- 下一轮最值得做的是把“真实契约形态驱动测试矩阵”纳入后续 contract/generator plan 的默认自检口径。
- PATTERNS.md 更新可延后；本次已有 BUG-0001 可检索，等同类 diff gate 漏检复现后再归纳模式更稳。
