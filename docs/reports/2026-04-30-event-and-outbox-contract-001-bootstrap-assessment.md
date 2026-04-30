# event-and-outbox-contract/001-bootstrap 交付复盘报告

> **日期**: 2026-04-30
> **审查人**: Codex

## 1 复盘范围与成功证据

- 本次交付覆盖 `event-and-outbox-contract/001-bootstrap`：`shared/events.yaml` / `shared/jobs.yaml` 双真理源、B3-owned event/job generator、Go / TS contract outputs、18 个 JSON Schema、baseline manifests、`make lint-events` / `codegen-events-check` drift gates、envelope / trace / breaking-change / `email_dispatch` 红线测试，以及 plan lifecycle close-out。
- 关联计划已收口：[plan](../spec/event-and-outbox-contract/plans/001-bootstrap/plan.md) / [checklist](../spec/event-and-outbox-contract/plans/001-bootstrap/checklist.md) Header 均为 `completed`，27/27 checklist items 完成，`plans/INDEX.md` 已移入 Completed。
- 通过验证：
  - `python3 -m pytest scripts/lint/events_inventory_test.py -q`
  - `python3 -m pytest scripts/lint/lint_events_test.py -q`
  - `go test ./cmd/codegen/events ./internal/shared/events ./internal/shared/jobs -count=1`
  - `pnpm --dir frontend test src/lib/events src/lib/jobs`
  - `pnpm --dir frontend typecheck`
  - `make codegen-events-check`
  - 临时删除 generated files 后 `make codegen-events` 可还原且 diff 为 0
  - 临时把 `report.generated.mistakeCount` 从 `int` 改 `string` 后 `make lint-events` 失败并包含 `breaking change requires eventVersion + 1`
  - C-10 grep 审计确认 implementation 侧没有 analytics snake_case eventName 混入
  - `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check`

## 2 会话中的主要阻点/痛点

- 旧 active plan 缺少当前质量门禁分类与逐项验证断言。
  - **证据**：`/implement` 入口发现 `plan.md` 缺 `## 3 质量门禁分类`，checklist 缺可执行 `验证:` 子句；用户确认方案 A 后先补 plan/checklist v1.2 再开始实现。
  - **影响**：实现前增加文档修复和确认往返；修复是必要的，但说明旧 active plans 仍会拖慢实施入口。
- 并行 plan 推进导致 `dev` 和当前分支多次变化。
  - **证据**：db-migrations plan 在同一仓库中快进 `dev` 并新增 retrospective / close-out commits；本 plan 需要多次把 event 分支基线快进到当前 `dev`，并显式排除 ci-pipeline / link-check 相关未提交改动。
  - **影响**：提高误提交风险；所有提交都必须精确 path staging，且收尾前需要反复确认当前 branch / HEAD。
- `codegen-events-check` 在未提交 generated/test diff 时天然失败。
  - **证据**：Phase 5 生成新 helper 与测试后，`make codegen-events-check` 的 `git diff --exit-code` 指出未提交 generated outputs；相关 diff 提交后同一 gate 通过。
  - **影响**：gate 正确，但作为“提交前验证命令”容易误读；执行者需要知道它适合作为 clean-tree / post-commit gate。
- 裸字面量扫描的实际允许路径比 checklist 原文更复杂。
  - **证据**：仓库中已有 `backend/internal/api/generated`、`frontend/src/api/generated`、generator tests 与 fixture 中的 job/event 字面量；最终 linter 需要允许 generated / fixture / test 路径，同时仍拒绝 backend/frontend runtime 裸字面量。
  - **影响**：计划原文“白名单仅 generated contract dirs”过窄；若严格照字面实现，会误伤已有 generated API 与 codegen test assets。
- 负例 drift 仍依赖临时修改真实仓库文件。
  - **证据**：Phase 4/6 分别临时修改 generated TS 文件和 `shared/events.yaml` 验证失败路径，再用 generator 或 patch 恢复。
  - **影响**：验证有效，但人工恢复成本高；若会话中断，容易留下临时 dirty diff。

## 3 根因归类

- 治理规则升级后，旧 plan 未被批量补齐。
  - **类别**：spec-plan
  - **说明**：本 plan 创建时没有当前 TDD/BDD quality-gate 模板；实施前修复是正确路径，但应尽量前置到 plan-review 阶段。
- 多 plan 并行时，branch / parent 基线没有显式自动收敛步骤。
  - **类别**：skill
  - **说明**：当前 `/implement` / `/tdd` 流程能人工处理，但没有把“parent dev 已快进、当前分支需 fast-forward 到 dev、只 staging owner paths”做成明显步骤。
- clean-tree drift gate 与 pre-commit verification 的语义混在一起。
  - **类别**：README / tooling
  - **说明**：`codegen-events-check` 的 `git diff --exit-code` 适合在 generated outputs 已提交后运行；在开发中途运行失败并不代表 generator 错误。
- source-scan lint 的 fixture/generated 例外没有在 plan 模板中标准化。
  - **类别**：spec-plan / README
  - **说明**：此类 contract lint 应统一说明 runtime source、generated source、tests、fixtures、docs 的扫描边界，避免每个 plan 临场解释。
- 临时 drift negative 是当前 contract plans 的常见验证方式。
  - **类别**：tooling
  - **说明**：本轮已恢复现场并通过 diff gates；问题不在功能缺陷，而在验证方式可自动化。

## 4 对流程资产的改进建议

- 对 active plans 做一次 quality-gate sweep。
  - **落点**：spec-plan / `/plan-review`
  - **优先级**：medium
  - **建议**：批量检查 active plan 是否缺 `## 3 质量门禁分类`、BDD 不适用说明、checklist `验证:` 子句，避免 `/implement` 执行时才发现。
- 在 `/implement` 或 `/tdd` phase-commit 规则中补并行分支收敛步骤。
  - **落点**：skill
  - **优先级**：medium
  - **建议**：当 `dev` 已包含 sibling plan commits 或当前 branch 与 plan branch 不一致时，要求先确认 branch ancestry、必要时快进 plan branch 到 `dev`、再精确 path staging。
- 为 event/job contract drift 提供临时目录 negative helper。
  - **落点**：tooling
  - **优先级**：medium
  - **建议**：把 breaking-change、generated drift、redacted-field drift 放到 temp repo / temp generated tree 中验证，减少直接 patch 真实源文件的次数。
- 在 lint/checklist 模板中明确 source scan 例外。
  - **落点**：spec-plan / README
  - **优先级**：low
  - **建议**：对 contract lint 明确允许 generated、fixtures、testdata、unit tests、docs；禁止范围聚焦 runtime implementation source。
- 标注 clean-tree drift gate 的推荐运行时机。
  - **落点**：README / Makefile help
  - **优先级**：low
  - **建议**：在 `codegen-events-check` 或通用 codegen gate 文档中说明：开发中途可跑 codegen + lint；`git diff --exit-code` clean-tree gate 适合提交后或准备提交前没有待提交 generated diff 时运行。

## 5 建议优先级与后续动作

- **Medium**：优先做 active plan quality-gate sweep，直接减少后续 `/implement` 入口阻塞。
- **Medium**：补 `/implement` / `/tdd` 分支收敛规则，降低并行 plan 快进 `dev` 时的误提交风险。
- **Medium**：为 event/job contract negative drift 增加 temp helper，减少真实工作树临时改写。
- **Low**：source scan 例外与 clean-tree gate 文档可在下一次 contract plan 模板修订时统一补齐。
