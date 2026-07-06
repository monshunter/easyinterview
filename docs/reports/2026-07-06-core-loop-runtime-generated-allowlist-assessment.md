# Core Loop Runtime / Generated Allowlist 交付复盘报告

> **日期**: 2026-07-06
> **审查人**: Codex

## 1 复盘范围与成功证据

本次交付继续以 `product-scope/001-core-loop-module-pruning` 为 owner，把上一份报告建议的 runtime / generated allowlist 变成可执行 gate：新增 `scripts/lint/core_loop_pruning_surface.py`，扫描 `backend/`、`frontend/`、`openapi/`、`shared/`、`config/`、`scripts/`、`migrations/`、`test/scenarios/`、`ui-design/`，并把旧 Debrief / Profile / JD Match 命中分桶为 `historical_migrations`、`legacy_normalization`、`negative_tests`、`real_residuals`。`real_residuals` 非空时脚本失败，允许项必须落入明确分桶。

成功证据：

- `product-scope/001-core-loop-module-pruning` plan / checklist 原地升到 v1.15，并新增 6.16 gate。
- `python3 -m pytest -q scripts/lint/core_loop_pruning_surface_test.py scripts/lint/makefile_dry_run_test.py -k 'core_loop_pruning_surface or lint_wires_core_loop_pruning_surface_gate'` PASS，6 selected。
- `python3 -m pytest -q scripts/lint/core_loop_pruning_surface_test.py scripts/lint/makefile_dry_run_test.py` PASS，10 tests。
- `make lint-core-loop-pruning-surface` PASS，输出 `historical_migrations=54`、`legacy_normalization=15`、`negative_tests=271`、`real_residuals=0`。
- `node --test ui-design/ui-design-contract.test.mjs` PASS，29 tests。
- `pnpm --filter @easyinterview/frontend test src/app/screens/practice/PracticeScreen.test.tsx src/app/screens/practice/hooks/usePracticeAssistance.test.ts src/app/i18n/localeFiles.test.ts` PASS，27 tests。
- `make lint` PASS，新 gate 已纳入 top-level lint。
- `make test` PASS；过程中修正 `backend/internal/ai/registry/backend_review_preflight_test.go` 中 stale 的 prompt-rubric spec 版本断言。
- `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check` PASS。
- `make docs-check` PASS。
- `git diff --check` PASS。

本轮真实残留处理：脚本暴露 user-visible English copy 中的 `experience cards` 残留，已在正式 frontend locale 与 `ui-design/` 原型中同步改为 `resume evidence`，保持前端实现与 UI 真理源一致。

## 2 会话中的主要阻点/痛点

- 手工 runtime sweep 难以区分“允许历史命中”和“真实残留”。
  - **证据**：同一组 retired keywords 在 migration 历史、legacy route normalization、负向测试、lint guard 和真实 UI copy 中都有命中。
  - **影响**：如果继续依赖人工 `rg`，容易误删历史迁移或放过正式资产残留。

- `profile` 相关匹配容易误伤当前 AI model profile / auth profile 语义。
  - **证据**：脚本实现中需要把 CandidateProfile / retired profile surface 与正常 `model profile`、`/auth/profile` 分开。
  - **影响**：过宽 allowlist 会掩盖真实残留，过宽 denylist 又会阻塞当前有效功能。

- 顶层 `make test` 暴露了邻近测试 drift。
  - **证据**：`TestF3ReportGenerateAndAssessmentPreflight` 仍断言 `prompt-rubric-registry/spec.md` v2.12，而当前文档为 v2.13。
  - **影响**：运行全量 gate 前看似与本任务无关，但不修会让交付无法完整闭环。

## 3 根因归类

- Runtime / generated pruning 之前缺少可执行、可复用的分桶脚本。
  - **类别**：tooling / spec-plan

- Retired module keywords 与当前通用术语有重叠，单纯字符串搜索不足以作为最终 gate。
  - **类别**：tooling

- 文档版本断言容易随 spec 升级漂移，需要在相关 plan 收口时同步维护。
  - **类别**：test

## 4 对流程资产的改进建议

- 保留 `lint-core-loop-pruning-surface` 作为 `make lint` 的固定 gate；后续删除 Debrief / Profile / JD Match 残留时先看 `real_residuals`，再决定删除或迁移语义。
  - **落点**：Makefile / lint scripts
  - **优先级**：high

- 若后续目标变成“历史迁移文件也零残留”，先做 pre-launch migration squash 或 migration lint final-state modeling 决策；不要在当前历史链语义下直接删除 `000009` / `000010`。
  - **落点**：migration / tooling
  - **优先级**：high

- 对精确 spec 版本断言，后续 plan 收口时把断言文件纳入对应 owner 的 docs/code review checklist，避免全量测试才暴露。
  - **落点**：spec-plan / test
  - **优先级**：medium

## 5 建议优先级与后续动作

下一步建议先不要进入 migration squash。更稳妥的路径是以 `product-scope/001-core-loop-module-pruning` 为 owner 跑一次独立 `/plan-code-review product-scope/001-core-loop-module-pruning cross-layer --fix`，重点审查新 gate 的 `real_residuals=0` 是否覆盖了当前所有 runtime / generated / config 真值源；确认无漏扫后，再把 migration squash 作为单独设计决策处理。
