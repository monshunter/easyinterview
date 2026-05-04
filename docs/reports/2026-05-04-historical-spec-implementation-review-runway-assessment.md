# Historical Spec Implementation Review Runway 交付复盘报告

> **日期**: 2026-05-04
> **审查人**: Codex

## 1 复盘范围与成功证据

本次交付覆盖 `historical-spec-implementation-review/001-implement-review-runway` 的 docs target 执行：完成 historical spec / plan inventory、L1 文档审查、9 个历史 plan 的质量门禁分类补齐、`repo-scaffold/001-bootstrap` lifecycle 收口，以及本轮无可执行 implementation / L2 target 的证据化记录。

已通过的验证：

- 15/15 `context.yaml` default target validation PASS。
- `sync-doc-index --fix-index` 后 Post-fix Verification zero drift；最终 `sync-doc-index --check` zero drift。
- `python3 scripts/lint/check_md_links.py docs` OK。
- docs/spec heading anchor audit `TOTAL 0`。
- `make docs-check` PASS。
- `make codegen-check` PASS。
- `make test` PASS：后端 Go tests + 前端 Vitest 10 files / 49 tests。
- `make build` PASS（当前 frontend build target 仍为 D1 前占位，返回 0）。
- `git diff --check` PASS。

## 2 会话中的主要阻点/痛点

- 历史 plan 与当前 `/implement` 质量门禁存在结构性漂移。
  - **证据**：15 个 plan 中有 9 个缺少 `## 3 质量门禁分类`；这些 plan 在当前 `/implement` Step 4.2 下会先被拦回 `/plan-review --fix`。
  - **影响**：即使历史实现已经 completed，后续重进 review / implement 时仍会被文档门禁阻塞。

- Runway checklist 把“可能执行 historical implementation”和“本轮是否存在可执行目标”写成同一条直线。
  - **证据**：L1 后没有安全可启动目标：A3 002 仍 draft-gated，roadmap 剩余项是 future child 创建规则，completed truth-source plans 没有新 code delta。
  - **影响**：需要用 skip evidence 标注 Phase 3 / 4，否则容易为了勾选而误启动 draft 或 completed plan。

- docs/spec heading anchor audit 仍是临时命令，不是稳定 repo gate。
  - **证据**：`make docs-check` 覆盖 Header/INDEX 和文件链接，但不验证 fragment anchor；本轮仍需手写 audit 命令确认 `TOTAL 0`。
  - **影响**：后续章节重命名可能再次造成深链静默失效。

## 3 根因归类

- **spec-plan**：历史 completed/draft plan 创建时早于当前强制质量门禁规则，缺少批量 backfill。
- **spec-plan**：runway checklist 没有显式的“无 eligible implementation target 时如何完成/跳过 Phase 3-4”分支。
- **skill / tooling**：heading anchor audit 尚未沉淀为 `scripts/lint` 或 `make docs-check` 子步骤。

## 4 对流程资产的改进建议

- 为 historical plan migration 增加一次性 gate/backfill 检查。
  - **落点**：`/plan-review` 或 `/sync-doc-index` 辅助脚本。
  - **优先级**：high。

- 在 runway plan 模板中显式加入 no-op branch。
  - **落点**：`docs/spec/historical-spec-implementation-review/plans/001-implement-review-runway/plan.md` 的后续版本，或未来同类编排 plan 模板。
  - **优先级**：medium。

- 将 docs/spec fragment anchor audit 固化为 repo 脚本并接入 `make docs-check`。
  - **落点**：`scripts/lint/` + root `Makefile`。
  - **优先级**：high。

## 5 建议优先级与后续动作

最高价值后续动作是把 heading anchor audit 纳入正式 docs gate；它已经连续两轮成为额外手工验证项。其次是为 historical/completed plan 增加质量门禁 backfill 检查，避免下一次 `/implement` 又在 Step 4.2 被相同问题打断。
