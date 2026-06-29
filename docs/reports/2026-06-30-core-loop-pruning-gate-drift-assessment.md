# Core Loop Pruning Gate Drift 交付复盘报告

> **日期**: 2026-06-30
> **审查人**: Codex

## 1 复盘范围与成功证据

- 本次交付覆盖 `product-scope/001-core-loop-module-pruning` review 后第二轮修复：同步 active A3/F3 AI profile truth source、OpenAPI 35-operation test contract、B3 14 event-schema codegen contract、frontend envelope fixture count，并把证据写回 owner plan/checklist v1.2。
- 关联 Bug： [BUG-0129](../bugs/BUG-0129.md)。
- 成功证据：
  - `python3 scripts/lint/ai_profile_coverage.py --repo-root .` PASS。
  - `cd backend && go test ./cmd/codegen/openapi` PASS。
  - `PYTHONPATH=. python3 -m pytest -q scripts/lint/openapi_diff_test.py` PASS (25 tests)。
  - `cd backend && go test ./cmd/codegen/events` PASS。
  - `pnpm --filter @easyinterview/frontend test src/lib/events/envelope.test.ts` PASS (2 tests)。
  - `make lint` PASS。
  - `make test` PASS（backend Go + frontend 154 files / 941 tests；现有 React `act(...)` warning 不影响结果）。
  - `make docs-check` PASS。
  - `make codegen-check` PASS。
  - `git diff --check` PASS。

## 2 会话中的主要阻点/痛点

- Active spec-derived lint 和配置删除被混淆。
  - **证据**：`config/ai-profiles.yaml` 已删除 debrief/profile profiles，但 `ai_profile_coverage.py` 仍从 A3 §4.5 和 F3 §3.1.1 读到 `debrief.*` / `profile.update`。
  - **影响**：`make lint` 红灯，且 active spec 会误导后续 AI owner 重新补回退役 profile。

- Codegen / fixture consumer 使用固定历史数量，删除合同时未同步。
  - **证据**：OpenAPI Go test 仍断言 43-row comment，Python diff test 仍期望 43 operations，event generator test 仍期望 16 schema files，frontend envelope test 仍期望 3 fixtures。
  - **影响**：`make test` 在 backend codegen 和 frontend focused test 上失败，即使 generated artifacts 本体已是当前 35-operation / 14-schema / 2-fixture 状态。

- Owner checklist 的 final gate 证据没有把 direct consumer focused gates 单独列为删除合同后的必须项。
  - **证据**：原 checklist 5.4 记录了 `make lint` / `make test` 等聚合 gate，但 review 仍能复现这些命令失败。
  - **影响**：历史 PASS 文案被当作完成证据，未形成对 stale test consumer 的当前态反查。

## 3 根因归类

- Active spec-derived lint 漏同步：
  - **类别**：spec-plan
  - **根因**：删除 `config/ai-profiles.yaml` 时没有把 lint 输入源 A3 Product/UI catalog 与 F3 feature_key table 列成同等 truth source。

- Fixed-count tests 漏同步：
  - **类别**：spec-plan / README
  - **根因**：module pruning plan 对 generated artifacts 和 fixture 删除有明确要求，但没有逐一列出 OpenAPI diff test、event schema-count test、frontend fixture-count test 这些 direct consumers。

- 聚合 gate 证据弱化：
  - **类别**：skill / spec-plan
  - **根因**：删除型 review 需要先跑 exact focused gates，再跑 `make lint` / `make test`；只写聚合命令名无法证明每个 direct consumer 当前真的覆盖了删除合同。

## 4 对流程资产的改进建议

- 删除型 plan 的 direct-consumer gate 明确列出 active spec-derived lint、fixed-count tests、generated comment tests、fixture-count tests。
  - **落点**：spec-plan
  - **优先级**：high

- `/plan-code-review` 删除型 review 输出中增加一行：对每个删除的 operation/event/profile/fixture，找出至少一个失败会暴露 stale expectation 的 focused test。
  - **落点**：skill
  - **优先级**：medium

- README 或 pattern 库保留 cleanup 漏同步 consumer 的检查项，覆盖“active spec-derived lint”而不只是代码测试。
  - **落点**：README / Bug PATTERNS
  - **优先级**：medium

## 5 建议优先级与后续动作

- 最高价值：后续 hard-delete 计划在 Phase 5 前新增 direct-consumer matrix，逐项列出 `profile/spec lint`、`operation-count`、`event schema-count`、`fixture-count` 等 focused gates。
- 本次已完成的流程资产沉淀：`product-scope/001-core-loop-module-pruning` checklist 5.5 和 [BUG-0129](../bugs/BUG-0129.md) 已记录这组 focused + aggregate gates；`PATTERNS.md` Pattern 1 已补充 active spec-derived lint consumer。
