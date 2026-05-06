# Backend Runtime Topology False-negative Gate 交付复盘报告

> **日期**: 2026-05-06
> **审查人**: Codex

## 1 复盘范围与成功证据

- 本次交付范围：修复 `backend-runtime-topology/001-worker-consolidation` L2 复扫发现的 `make lint-runtime-topology` false negative，覆盖旧 producer、listen addr、worker binding 与 `backend-async-runtime` shorthand 回流。
- 成功证据：`python3 -m unittest scripts.lint.runtime_topology_test` 先 Red 后 Green；`make lint-runtime-topology`、`make lint`、`make test`、`make codegen-check`、`make docs-check`、context validator、`sync-doc-index --check` 与 `git diff --check` 均通过。
- 关联记录：[BUG-0015](../bugs/BUG-0015.md)；owner spec/plan/checklist/context 已升到 v1.2，Phase 4.4 已完成。

## 2 会话中的主要阻点/痛点

- 旧 lint gate 通过但 active handoff 仍有旧口径。
  - **证据**：review pass 发现 `docs/spec/event-and-outbox-contract/plans/001-bootstrap/plan.md` 仍写 `` `worker` producer``，`docs/spec/secrets-and-config/plans/001-bootstrap/plan.md` 仍写 `app/worker listen addr`，ADR-Q3 仍写 `backend-async-runtime`。
  - **影响**：上一轮 L2 remediation 的 PASS 容易被误读为语义零残留，后续 owner 仍可能按旧 subject 或旧 worker 口径接力。
- Red fixture 起初没有覆盖一行多 pattern 的报告行为。
  - **证据**：新增 fixture 初次 Green 前，`worker bindings` 与 `app/worker listen addr` 写在同一行时只报告第一类命中。
  - **影响**：测试断言需要贴近实际 lint 行为，否则会把报告格式差异误判成修复失败。

## 3 根因归类

- 负向 lint 只覆盖上一轮暴露的 exact literal，没有把同义语序和 owner shorthand 纳入 Red fixture。
  - **类别**：spec/plan
- `runtime_topology.py` 的扫描行为是“每行命中后停止报告后续 pattern”，测试数据未一开始体现这个契约。
  - **类别**：skill / no repo change needed

## 4 对流程资产的改进建议

- 负向 drift gate 的 checklist 应要求至少一个反向语序或同义 shorthand fixture，而不仅是 exact literal。
  - **落点**：spec/plan
  - **优先级**：high
- 后续若 lint 采用“一行只报第一类命中”的策略，测试 fixture 应将不同 bug class 放在不同行，或显式断言该报告策略。
  - **落点**：skill
  - **优先级**：medium

## 5 建议优先级与后续动作

- 最高优先级：后续 `plan-code-review --fix` 处理负向搜索 gate 时，把“同义语序 / 自然语言 shorthand / owner rename”作为 Red fixture 设计要求。
- 可延后：若类似 false negative 再出现，可把上述要求沉淀到 `plan-code-review` skill 的 Step 4/Step 6 remediation 指南，而不是只依赖各 plan 自行记忆。
