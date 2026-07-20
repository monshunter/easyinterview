# Frontend List Pages Review Remediation 交付复盘报告

> **日期**: 2026-07-20
> **审查人**: Codex

## 1 复盘范围与成功证据

- 本次按代码审查结论修复三项：Practice 启动过渡对完整 App shell 的交互阻断、ReportsScreen 日期事实标签、Workspace/Practice spec 合同 ID 重复；明确排除未被用户接受的卡片按钮嵌套修订。
- 原 owner 已在 [Workspace/Practice 001](../spec/frontend-workspace-and-practice/plans/001-workspace-and-interview-context/plan.md)、[Workspace/Practice 002](../spec/frontend-workspace-and-practice/plans/002-practice-text-event-loop/plan.md)、[Report 001](../spec/frontend-report-dashboard/plans/001-report-screen-and-generating-handoff/plan.md) 与 [A5 001](../spec/ci-pipeline-baseline/plans/001-local-quality-gates/plan.md) 原地修订并恢复 `completed`。
- RED 分别证明 App root 未 inert、TargetJob `createdAt` 被错误标注为 Interview date、唯一性脚本/Makefile gate 缺失；GREEN focused tests 为 40/40，duplicate-ID Python contracts 为 5/5。
- `pnpm --filter @easyinterview/frontend typecheck`、production `make build`、`make docs-check` 与四个 owner context 全部通过；根 `make test` 通过 Python 620 tests / 4615 subtests、全部 Go packages 与 frontend 136 files / 1115 tests。

## 2 会话中的主要阻点/痛点

- Practice 旧测试只在 RTL 外层 container 上断言 inert，没有复刻 `app-root > TopBar + main` 的真实层级。
  - **证据**：新增真实层级 RED 后，完整 App root 没有 `inert`，而旧实现只命中 `<main>`。
  - **影响**：视觉上保留 TopBar 的同时留下可交互穿透，最终态/孤立组件测试无法发现。
- ReportsScreen 直接把 TargetJob `createdAt` 写成“面试日期”，绕过了 UI 设计中“不伪造面试日期”的事实边界。
  - **证据**：英文 RED 收到 `Interview date: Jul 14`，修复后 zh/en 均只显示规划创建日期。
  - **影响**：页面显示了格式正确但业务含义错误的事实，容易被一般快照测试误判为正常。
- 新增全仓合同 ID 扫描后，发现三个不属于本次修复的既有重复项。
  - **证据**：首次扫描命中 `ai-provider-and-model-routing` 一项与 `frontend-home-job-picks-and-parse` 两项。
  - **影响**：若直接要求全量零重复，会迫使本次修复越权改动其他 owner；若不增加 gate，同类问题又会复发。

## 3 根因归类

- 启动过渡的阻断合同没有把“TopBar 可见”和“完整 App root inert”同时写成真实 DOM 不变量。
  - **类别**：spec/plan
- 日期显示只验证格式与存在性，没有验证 producer 字段与用户标签之间的语义归属。
  - **类别**：spec/plan
- Spec 合同 ID 依赖人工编号，`make docs-check` 之前只覆盖 Header/INDEX、链接与 fragment。
  - **类别**：README / spec-plan
- 首次 focused 命令通过 package script 追加参数时实际触发了全量 Vitest；结果仍有效，但属于一次性命令选择问题。
  - **类别**：无需仓库改动

## 4 对流程资产的改进建议

- 对 portal overlay / transition 测试统一使用真实 shell 层级，至少包含 app root、共享 chrome、业务 main 与 body portal，并同时断言挂载和卸载恢复。
  - **落点**：相关 UI spec / plan / component test
  - **优先级**：high
- 对展示日期、状态、计数等事实字段，测试必须断言“字段来源 → 标签语义”，不能只断言格式化输出存在。
  - **落点**：相关 UI spec / plan / domain behavior test
  - **优先级**：high
- 保留 A5 的 duplicate-ID ratchet，并由各历史 owner 独立清理三组 legacy baseline；清理完成后删除对应例外，不把新重复加入 baseline。
  - **落点**：A5 lint owner + 对应 subject spec/plan
  - **优先级**：medium

## 5 建议优先级与后续动作

- 下一步优先由 `ai-provider-and-model-routing` 与 `frontend-home-job-picks-and-parse` 各自 owner 原地修复三组 legacy contract ID，并同步删除 `check_spec_contract_ids.py` 中的例外。
- 后续 overlay 类 UI 变更直接复用本次 App-root 层级测试模式；日期类 UI 变更先写字段 provenance 断言，再写格式化断言。
- 本次问题已由现有 [BUG-0192](../bugs/BUG-0192.md) 的核心工作流 UI 漂移主题覆盖，按 Bug 建档阈值不新增重复 Bug 记录。
