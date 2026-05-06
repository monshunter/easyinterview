# Backend Runtime Topology L2 Remediation 交付复盘报告

> **日期**: 2026-05-06
> **审查人**: Codex

## 1 复盘范围与成功证据

- 范围：`backend-runtime-topology/001-worker-consolidation` Phase 4.3 L2 remediation，修复 retired standalone worker process 口径在 active runtime code comments、completed owner plan/checklist 和 quality gate 中的残留。
- 成功证据：
  - `python3 -m unittest scripts.lint.runtime_topology_test`
  - `make lint-runtime-topology`
  - `make lint`
  - `python3 .agent-skills/implement/shared/scripts/validate_context.py --context docs/spec/backend-runtime-topology/plans/001-worker-consolidation/context.yaml --docs-root docs --target repo`
  - `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check`
  - `make docs-check`
  - `make test`
  - `make codegen-check`
  - `git diff --check && git diff --cached --check`
- 关联 Bug：[BUG-0014](../bugs/BUG-0014.md)。

## 2 会话中的主要阻点/痛点

- 原 4.2 negative search gate 扫描面过窄。
  - **证据**：L2 review 在 `docs/spec/*/plans/**` 中发现 `cmd/worker`、worker producer enum、worker 类组件 probe、privacy worker、Asynq worker 等旧口径。
  - **影响**：worker consolidation 虽然已移除独立 runtime entrypoint，但后续 owner plan 仍可能按旧 process topology 继续实现或验证。
- Active code comments 未纳入 topology drift gate。
  - **证据**：`make lint-runtime-topology` 首轮失败，拦截到 backend migration 与 secrets/config 注释中的旧 worker 口径。
  - **影响**：配置和迁移代码的注释会误导 runtime binding 与 component scope 认知。
- 历史记录和当前 handoff surface 的边界需要显式化。
  - **证据**：`backend-runtime-topology/history.md` 可以作为历史说明保留旧术语，但其他 completed owner plan/checklist 中的可执行语句必须修订。
  - **影响**：没有 allowlist 语义时，人工搜索容易在“历史说明可保留”和“active handoff 必须清零”之间误判。

## 3 根因归类

- Topology removal gate 缺少 executable lint。
  - **类别**：spec-plan
  - 原 plan/checklist 用一次性 `rg` evidence 表示零残留，未固化成 `make lint` 可重复执行的 gate。
- Plan-level active handoff surface 未被纳入 artifact-level 反查。
  - **类别**：spec-plan
  - completed plan/checklist 被当作历史完成证据，而不是当前后续执行者会读取的 owner handoff。
- Allowlist 规则缺失。
  - **类别**：spec-plan
  - 历史 owner 文档和当前 owner 文档的处理规则没有落入脚本，导致人工 review 需要重复判断。

## 4 对流程资产的改进建议

- 后续涉及删除 runtime entrypoint、route、job producer、feature flag 或组件拓扑的 plan，应在完成前新增对应 executable negative lint，并接入聚合 `make lint`。
  - **落点**：spec-plan
  - **优先级**：high
- `/plan-code-review --fix` 对“completed owner plan 但仍有 downstream handoff 语义”的场景，应默认扫描 `docs/spec/*/plans/**`，并区分 history allowlist 与 active plan/checklist。
  - **落点**：skill
  - **优先级**：medium
- 对 topology 类 lint 保持 unit test 覆盖：至少包含 positive fixture、negative fixture 和 allowlist fixture。
  - **落点**：spec-plan
  - **优先级**：medium

## 5 建议优先级与后续动作

- 优先：在下一轮 L2 review 中复用 `runtime_topology` lint 的模式，检查是否还有其他已删除 runtime surface 只靠人工 `rg` gate 验收。
- 次优先：若同类问题再次出现，再把 “completed owner plan handoff surface 必扫” 固化进 `plan-code-review` skill。
- 可延后：扩展通用 docs active-surface drift checker；当前 BUG-0014 与 Phase 4.3 gate 已能覆盖本次 worker topology 范围。
