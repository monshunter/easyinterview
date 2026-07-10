# P0.037 Ready DOM Synchronization 交付复盘报告

> **日期**: 2026-07-10
> **审查人**: Codex

**关联计划**: [Frontend Resume Workshop Listing Routing and Detail Readonly](../spec/frontend-resume-workshop/plans/001-listing-routing-and-detail-readonly/plan.md)
**关联 Bug**: [BUG-0153](../bugs/BUG-0153.md)

## 1 复盘范围与成功证据

- 本次修复仅调整 P0.037 pending-PDF 场景及其 `ResumeDetailView` owner mirror 的异步等待条件，不修改生产代码或用户行为。
- 4 个并发 focused 进程均通过，每个进程为 2 files / 14 tests。
- P0.037 四段 wrapper 通过 6 / 6；Resume Workshop owner 回归通过 20 files / 113 tests；frontend typecheck 通过。
- 完整 frontend 回归通过 137 files / 841 tests；owner/product context、docs/index/link/diff/pruning gates 在收口阶段执行。

## 2 会话中的主要阻点/痛点

- 请求次数被误当作 React 可见状态提交屏障。
  - **证据**：并行 build/full-test 时第二次 `getResume` 已记录，但 DOM 仍显示 `Loading resume...`，完整套件为 840 / 841；focused 重跑为 6 / 6。
  - **影响**：回归门禁在高负载下偶发误失败，单次 focused PASS 无法证明稳定性。
- `change-intake` matcher 被通用 API 关键词误导。
  - **证据**：查询包含 P0.037、Resume detail 和 pending PDF，matcher 仍因 `getResume` 把 report dashboard owner 排第一；场景 README 和测试路径明确指向 Resume Workshop 001。
  - **影响**：若不做 artifact-level owner 反查，会把 completed plan 原地修订到错误 subject。
- 最终证据需要区分压力暴露与稳定收口。
  - **证据**：并行 build/test 暴露竞态；修复后通过 4 个并发 focused 进程，并再串行运行完整 frontend 套件。
  - **影响**：只保留任一单一路径都不足以同时证明“竞态已被触发过”和“标准交付门禁稳定通过”。

## 3 根因归类

- P0.037 与 owner mirror 的同步条件错误。
  - **类别**：spec/plan
  - Phase 15 已把“等待用户可见 ready DOM，再断言调用次数”固化到原 owner。
- change-intake 候选排序对场景 ID 和场景 README owner 边权不足。
  - **类别**：skill
  - 通用 operationId/API 关键词分值压过了更直接的场景归属证据。
- 并发压力与串行最终门禁承担不同目的。
  - **类别**：无需仓库改动
  - 本次通过执行顺序即可表达，不需要把所有完整 build/test 默认并行化。

## 4 对流程资产的改进建议

- `change-intake` matcher 应优先解析精确场景 ID，并读取目标场景 README 的 Owner 链接；该证据应高于通用 API keyword。
  - **落点**：`.agent-skills/change-intake` matcher
  - **优先级**：high
- 异步 UI 场景指南应明确：网络调用次数、Promise resolve 或 loading flag 只能作为辅助证据，主等待条件必须是用户可见 DOM 或明确终态。
  - **落点**：`test/scenarios/README.md` 或 frontend test owner plan
  - **优先级**：medium
- 对 timing-sensitive 修复保留“并发 focused 压力 + 串行 full suite”双证据，但不强制所有常规交付都并行跑完整 build/test。
  - **落点**：无需仓库改动；按风险在 checklist 中声明
  - **优先级**：low

## 5 建议优先级与后续动作

- 下一轮最高价值改进是修正 `change-intake` matcher 对 `E2E.P*.NNN` 与场景 README Owner 的权重，避免跨 subject 误路由。
- Resume owner 已通过 Phase 15 和 BUG-0153 承接本次测试规则，无需新增 sibling plan 或生产补丁。
- 技术债长时任务可继续处理 AST 重扫剩余的零读取 prototype props；它们与本 Bug 无依赖，应保持独立批次。
