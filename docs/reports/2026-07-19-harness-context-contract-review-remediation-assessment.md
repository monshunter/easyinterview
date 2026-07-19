# Harness Context Contract Review Remediation 交付复盘报告

> **日期**: 2026-07-19
> **审查人**: Codex

**关联计划**: [Harness v1 渐进迁移 Plan](../spec/harness-simplification/plans/001-harness-v1-migration/plan.md)

## 1 复盘范围与成功证据

本次交付按用户裁决把简化后的 `context.yaml` 固定为当前 Harness 合同，暂不规划删除，并修复 code review 发现的四项漂移：无 manifest owner 旁路、generator 遗漏后增一等链接、Project Arch CLI/根 Make gate 缺失，以及 INDEX 测试硬编码旧标题。

成功证据：

- `make harness-test`：`181 passed`；
- `scripts/harness_arch_test.py`：覆盖 init/check/upgrade/repair、失败回滚、冲突 fail closed、fresh Project Arch 最小 `context.yaml` 与根 Make gate；
- 50 份 `context.yaml` batch validation：`50 passed, 0 failed`；
- generator 当前事实投影：`contexts=50, drifts=0`；
- `make docs-check`：Header、INDEX、孤儿文档、warning 与 Markdown link 均为零问题；
- `git diff --check` 与相关 Python `py_compile` 均通过；当前 owner 资产中旧“过渡期/Phase 2 删除 context”口径负向搜索为零命中。

## 2 会话中的主要阻点/痛点

### 2.1 上一轮退出方向与本轮用户裁决冲突

- **证据**：同日既有复盘仍建议 Phase 2 删除 `context.yaml`，而本轮用户明确要求以最小清单为标准并暂不考虑删除。
- **影响**：若只修代码、不修 active Spec、plan、AGENTS 与模板，后续 Skill 会继续按旧方向生成相反实现。

### 2.2 reader、writer 与 owner 发现没有共享同一合同

- **证据**：change-intake 能发现无 manifest plan，但 implement/review 入口要求 manifest；generator 保留旧 target 时不会吸收目录中新出现的 BDD/test 一等链接。
- **影响**：入口可以把工作交给下游无法执行的 owner，或把合法的新链接静默遗漏在生成结果之外。

### 2.3 计划勾选状态领先于可执行产物

- **证据**：Phase 1.2 所需 `scripts/harness_arch.py` 与根 `make harness-test` 在 review 时缺失；旧 INDEX 测试又把历史标题当作当前真理。
- **影响**：局部 Skill 测试为绿仍无法证明 Project Arch 主路径可运行，Spec 改名还会制造无业务意义的回归失败。

## 3 根因归类

- 最小 `context.yaml` 是否保留没有在所有 active owner 中形成一致决策。
  - **类别**：spec-plan / AGENTS.md / README
- matcher、generator、candidate 与下游 Skill 各自实现了部分计划发现语义，缺少同一 manifest 合同的负向断言。
  - **类别**：skill
- Project Arch 的阶段完成定义没有被根级可执行 gate 约束，标题测试也没有从当前 Spec 投影期望值。
  - **类别**：spec-plan / skill
- 本次发现来自合同审查而非独立运行时故障，根因和防复发测试已由当前 plan 直接承接。
  - **类别**：无需独立 Bug 记录

## 4 对流程资产的改进建议

- Phase 1.3 的 repository-fact discovery 必须以最小 `context.yaml` 为 owner 边界；允许提升检索与 handoff 质量，但不得恢复无 manifest 执行旁路。
  - **落点**：Harness spec-plan / change-intake / implement
  - **优先级**：high
- 后续任何 manifest 字段变化都同时维护 validator、generator、matcher、candidate、模板和 consumer dedupe 测试，并保留全库 batch validation 与 generator zero-drift gate。
  - **落点**：implement shared tooling / Harness plan
  - **优先级**：high
- Project Arch 后续阶段只在 `make harness-test` 可复现其 CLI 主路径、失败路径和安装后合同后勾选完成；文档投影断言继续从当前 owner 标题或结构生成。
  - **落点**：Harness plan / root Makefile / Project Arch tests
  - **优先级**：medium

## 5 建议优先级与后续动作

最高优先级是先审阅并提交本次 Phase 1.2 修订，使最小 `context.yaml` 合同、四项修复与 181 项聚合门禁处于同一原子变更。随后再进入 Phase 1.3，扩展仓库事实 discovery 与四态 handoff；该阶段应继续把缺失或非法 manifest 视为 `spec_required`/合同缺口，而不是恢复兼容旁路。
