# Backend Runtime Topology Structured Gate 交付复盘报告

> **日期**: 2026-05-07
> **审查人**: Codex

## 1 复盘范围与成功证据

- 本次交付范围：修复 `/plan-code-review backend-runtime-topology/001-worker-consolidation repo --fix` 确认的两个 L2 findings：跨行 YAML / JSON producer `worker` 字段 false-negative，以及 owner plan/checklist current handoff 旧 worker 口径 false-negative。
- 成功证据：`python3 -m unittest scripts.lint.runtime_topology_test` Red 后 Green；`make lint-runtime-topology` 通过并扫描 301 个 active 文件；`make lint`、`make test`、`make codegen-check`、`make docs-check`、context validator、`git diff --check`、`cd backend && go build ./cmd/...` 均通过。
- 关联记录：[BUG-0017](../bugs/BUG-0017.md)；`backend-runtime-topology/001-worker-consolidation` spec/plan/checklist/context 已升到 v1.4，Phase 4.6 / 4.7 已完成并恢复 `completed`。

## 2 会话中的主要阻点/痛点

- `runtime_topology.py` 声称覆盖 generated contract / event truth source，但仍用逐行 grep 证明结构化字段。
  - **证据**：临时 fixture 中 `producer` 与 `worker` 分布在 YAML / JSON 不同行时，旧 lint 返回 `runtime_topology: OK`。
  - **影响**：未来 event schema 或 fixture 以真实格式回流 `worker` producer 时，gate 会误判通过。
- owner docs 的例外策略过宽。
  - **证据**：临时 owner plan 写入 `Current handoff: build backend/cmd/worker as the runtime entry.` 时，旧 lint 返回 `OK (0 active files scanned)`。
  - **影响**：原计划作为 owner 反而无法自证 current handoff 没有旧 worker 入口，违反 deep reconcile 的 owner 承接要求。
- 这是同一 runtime topology gate 的连续第三轮 false-negative hardening。
  - **证据**：本次关联 [BUG-0017](../bugs/BUG-0017.md)，前序已有 [BUG-0015](../bugs/BUG-0015.md) 和 [BUG-0016](../bugs/BUG-0016.md)。
  - **影响**：单次补 exact literal 的方法成本偏高，后续相似 gate 需要先抽象扫描面、结构化字段和例外语义。

## 3 根因归类

- 负向 gate 的 Red fixture 没按载体类型拆分：自然语言、raw field、结构化 block、generated schema、owner handoff 应分别有 fixture。
  - **类别**：spec-plan
- `/plan-code-review --fix` 的修复预案虽然要求 deep evidence，但没有强制把“例外规则双向证明”作为每个 false-negative 修复的默认 checklist 形态。
  - **类别**：skill
- Bug 模式库尚未沉淀“零残留 lint gate false-negative”通用模式。
  - **类别**：README

## 4 对流程资产的改进建议

- 对所有“旧口径零残留” owner plan，新增 gate 时应显式列出四类 Red fixture：path-scope、raw field、structured block、exception boundary。
  - **落点**：spec-plan
  - **优先级**：high
- 后续若继续修订 `/plan-code-review` 或 `/tdd`，把 false-negative 修复模板调整为“扫描面 + 结构化载体 + 例外双向证明”。
  - **落点**：skill
  - **优先级**：medium
- 将 BUG-0015 / BUG-0016 / BUG-0017 归纳为 `docs/bugs/PATTERNS.md` 中的测试治理模式，避免同类 lint gate 只补单个 literal。
  - **落点**：README
  - **优先级**：medium

## 5 建议优先级与后续动作

- 最高优先级：下一轮涉及 route / runtime / contract 删除的 plan，应先在 checklist 里写出扫描面、结构化载体和例外边界，再进入实现。
- 中优先级：在用户确认后，把 BUG-0015/0016/0017 归纳到 `docs/bugs/PATTERNS.md`，形成“零残留 gate false-negative”排查清单。
- 可延后：若后续 `/plan-code-review --fix` 再遇到同类问题，再把上述模板要求沉淀进 skill 本体。
