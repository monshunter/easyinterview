# Product Contract UI Scope Alignment 交付复盘报告

> **日期**: 2026-05-03
> **审查人**: Codex

## 1 复盘范围与成功证据

本次交付范围是把已创建的历史 spec、plan、checklist、context、契约生成物和已实施代码重新对齐到 product-scope v1.2 与当前 `ui-design` / `docs/ui-design` 静态设计范围。核心结论是：未被当前 UI 和 UI 文档保留、且未被 product-scope 明确列为规划例外的旧能力，均按已决策丢弃处理。

成功证据：

- 已提交 `0a7909e fix(product-contract): align specs and artifacts with ui scope`。
- B1 / B2 / B3 / B4 / A4 相关 plan/checklist 均完成本次 remediation 并切回 `completed`；`python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check` 显示 Header / INDEX 零漂移。
- `git diff --check`、`make lint-conventions`、`make lint-openapi`、`make validate-fixtures`、`make render-openapi-fixture-examples`、`make openapi-diff`、`make lint-events`、`make lint-config`、`python3 scripts/lint/migrations_lint.py --repo-root .` 均通过。
- 后端 Go 聚焦测试、前端 vitest 聚焦测试、Python unittest / pytest 聚焦测试均通过；`make codegen` 可重复执行。
- 实现侧排除搜索确认旧独立 Mistakes / Growth、旧 practice mode / goal、旧 feature flag、旧 DB 字段和旧 event 字段不再残留；剩余旧词仅用于历史说明、remediation 证据或负向测试。

## 2 会话中的主要阻点/痛点

- 旧产品范围分散在多层资产中，不只存在于旧 spec 文本。
  - **证据**：全局扫描发现 B2 fixture plan 主体仍写 37 operation / 14 tag / `listMistakes` / `getGrowthOverview`，B3 event plan 主体仍写 18 events / `MistakeStatus`，F1 observability 示例仍引用 `mistake_entries_created_total`。
  - **影响**：如果只改 product-scope 或 OpenAPI 主 spec，后续 `/implement` 或 `/plan-review` 仍可能按历史 plan 主体恢复已移除模块。

- A4 runtime-config 曾只按 `public` 标记透传 feature flag。
  - **证据**：更新 `TestBuildRuntimeConfigAllowlistAndOptOut` 后，旧 `mistake_book_export_enabled` / `growth_dashboard_v1_enabled` / `mock_session_dual_track_enabled` 仍会进入 public runtime config，测试先红后绿。
  - **影响**：即使配置文件删除旧 flag，只要上游 snapshot 返回旧 public flag，前端仍可能观察到已移除能力。

- 已完成 plan 原地 reopen 后容易留下 lifecycle 状态漂移。
  - **证据**：本次 B1 / B2 / B3 / B4 / A4 remediation 完成后，需要统一将 plan/checklist 从 `active` 切回 `completed`，并通过 `sync-doc-index --fix-index` 迁移 plans/INDEX 行。
  - **影响**：如果不做最后生命周期收口，后续任务入口会误判这些 plan 仍处于执行中。

- 本地 gitleaks 第二层扫描不可用。
  - **证据**：`make lint-config` 中 env 字典校验通过，但脚本提示本机未安装 gitleaks，按 A4 策略跳过第二层扫描。
  - **影响**：本次没有 secret 内容变更，风险可接受；但本地安全验证仍少一层。

## 3 根因归类

- 旧范围分散在 spec / plan / checklist / generated artifacts 中。
  - **类别**：spec-plan
  - **说明**：product-scope 的语义已经明确，但历史 plan 主体和验收口径没有天然随父级范围变化自动收敛。

- runtime-config 缺少当前 feature flag 正向 allowlist。
  - **类别**：spec-plan
  - **说明**：A4 spec 已锁当前 6 项 baseline flag，但实现层原本没有用正向清单把“已删除但上游仍返回”的场景挡住。

- lifecycle 收口依赖人工最后一步。
  - **类别**：skill
  - **说明**：`sync-doc-index` 能修复 INDEX 漂移，但入口流程需要更早提醒“reopen completed plan 后必须完成状态复原”。

- gitleaks 未安装。
  - **类别**：no repo change needed
  - **说明**：现有脚本和文档已说明安装方式；本次不是流程或代码缺陷。

## 4 对流程资产的改进建议

- 在 `/plan-review` 或 `/change-intake` 增加 product-scope removal 传播检查。
  - **落点**：skill
  - **优先级**：high
  - **建议**：当 product-scope 或 UI scope 明确“未提及即丢弃”时，自动要求扫描 spec / plan / checklist / OpenAPI / events / migrations / config / generated artifacts，并区分历史说明、负向测试与当前验收口径。

- 在 spec-centric plan 模板或 review checklist 中增加“完成态 plan 原地 reopen 收口”检查。
  - **落点**：spec-plan
  - **优先级**：medium
  - **建议**：明确 remediation 完成后必须恢复 plan/checklist `completed`、运行 `sync-doc-index --fix-index` / `--check`，并把该证据写入 checklist。

- 对 runtime-config feature flag 采用正向 allowlist 作为默认实现模式。
  - **落点**：spec-plan
  - **优先级**：medium
  - **建议**：A4 后续扩展 flag 时，先修 spec，再同步 allowlist、配置、测试；不要只依赖上游 provider 的 `public` 标记。

- 本地安装 gitleaks。
  - **落点**：no repo change needed
  - **优先级**：low
  - **建议**：开发机可按 `make lint-config` 提示安装；仓库侧已有跳过策略和远端 CI 承接说明，本次不需要修改流程资产。

## 5 建议优先级与后续动作

最高价值后续动作是增强 `/plan-review` 或 `/change-intake` 的产品范围删除传播检查，避免未来只修父级 spec 而漏掉历史 plan 主体、契约 fixture、event、migration、feature flag 等执行入口。

第二优先级是把 completed plan reopen 的 lifecycle 收口写入模板化检查，降低重复依赖人工记忆的风险。

runtime-config 正向 allowlist 已在本次代码中落地；后续只需在新增 flag 时按 A4 spec 修订流程执行。
