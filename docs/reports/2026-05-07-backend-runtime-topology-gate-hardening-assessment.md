# Backend Runtime Topology Gate Hardening 交付复盘报告

> **日期**: 2026-05-07
> **审查人**: Codex

## 1 复盘范围与成功证据

- 本次交付范围：修复 code review 指出的 `runtime_topology.py` 二次 false-negative，覆盖 `scripts/` tooling 面旧 worker 口径回流，以及 `producer: worker` / `"producer": "worker"` raw contract 字段回流。
- 成功证据：`python3 -m unittest scripts.lint.runtime_topology_test` Red 后 Green；`make lint-runtime-topology` 通过并扫描 299 个 active 文件；`make lint` 通过。
- 关联记录：[BUG-0016](../bugs/BUG-0016.md)；`backend-runtime-topology/001-worker-consolidation` spec/plan/checklist/context 已升到 v1.3，Phase 4.5 已完成。

## 2 会话中的主要阻点/痛点

- 已有 false-negative gate 仍没有覆盖 tooling scripts。
  - **证据**：Red fixture 写入 `scripts/lint/env_dict.py` 的 `backend/cmd/worker` 和 `scripts/lint/getenv_boundary.go` 的 `WORKER_LISTEN_ADDR` 后，旧 lint 没有报告 scripts 路径。
  - **影响**：后续可以在 env dictionary 或 getenv boundary 这类 gate 脚本中重新引入 worker allowlist，而 runtime topology gate 不会拦截。
- Producer 旧口径只覆盖自然语言，没有覆盖字段形态。
  - **证据**：Red fixture 写入 `shared/events.yaml` 的 `producer: worker` 和 baseline JSON 的 `"producer":"worker"`；旧 lint 没有报告 raw producer 字段。
  - **影响**：虽然 `make lint-events` 覆盖当前 B3 truth source，但 runtime topology gate 自身不能证明它声称的旧 producer 零残留范围。
- 扩大扫描面需要明确 lint 自身例外。
  - **证据**：`runtime_topology.py` 必须保留 retired pattern 字面量用于定义规则；新增 allowlist fixture 验证该文件可被排除。
  - **影响**：如果没有最小例外，扫描 `scripts/` 会把 lint 本体误判为旧口径残留。

## 3 根因归类

- 负向 gate 的 Red fixture 仍偏向已经发生过的 exact literal，没有覆盖同一语义在 tooling / YAML / JSON 中的载体。
  - **类别**：spec-plan
- Lint 扫描范围和例外策略没有作为 checklist 验收内容显式列出。
  - **类别**：spec-plan

## 4 对流程资产的改进建议

- 后续任何“零残留” lint gate 都应在 checklist 中列出扫描范围、例外文件和至少一组 raw contract fixture。
  - **落点**：spec-plan
  - **优先级**：high
- `/plan-code-review --fix` 遇到 false-negative 修复时，应要求 Red fixture 同时覆盖 path-scope 漏扫和 pattern 漏扫，而不是只补当前字面量。
  - **落点**：skill
  - **优先级**：medium

## 5 建议优先级与后续动作

- 最高优先级：后续处理类似 runtime / route / contract 删除时，把 scripts/tooling 和 generated/raw fixture 一并纳入 owner plan gate。
- 可延后：若第三次出现同类 false-negative，再把“path-scope + raw-contract fixture”要求沉淀到 `plan-code-review` skill，而不是只保留在本次复盘中。
