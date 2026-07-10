# Resume Create Prototype Handoff 交付复盘报告

> **日期**: 2026-07-10
> **审查人**: Codex

**关联计划**: [Frontend Resume Workshop Create Flow](../spec/frontend-resume-workshop/plans/002-create-flow/plan.md)
**关联 Bug**: [BUG-0154](../bugs/BUG-0154.md)

## 1 复盘范围与成功证据

- 本批 6.230 删除静态 `ResumeCreateFlow.nav` 未读参数及 caller argument，并修复创建后本地资产在 `flow=create` 到 `resumeId` handoff 中丢失的问题。
- UI contract 经两轮 focused RED/GREEN 后 44/44 通过；Babel inventory 扫描 11 个原型文件，`unusedProps=[]`、`unusedState=[]`。
- Create-flow focused tests 32/32；P0.081 real-mode generated-client gate 1/1、场景 focused tests 28/28，setup/trigger/verify/cleanup 完整执行。
- 浏览器覆盖 hash create、返回列表、本地 New resume、Paste、50ms waiting detail 和 1.2s ready detail；page errors 为空，静态资源为 200/304。
- Frontend 全量 137 files / 841 tests、typecheck、build、两个 owner contexts、docs check、diff check 与 pruning surface 均通过，`real_residuals=0`。

## 2 会话中的主要阻点/痛点

- 原 UI contract 只验证 `nav("resume_versions", { resumeId })` 调用存在，却没有验证导航后的 owner instance 和 DOM。
  - **证据**：第一次浏览器提交回到三条 fixture 列表，而原契约仍为 PASS。
  - **影响**：source-level 假绿掩盖了 `createdResumes` 在 remount 时丢失的真实行为。
- 首次修复只覆盖 hash `flow=create` 入口，没有覆盖列表内本地 `setFlow("create")` 入口。
  - **证据**：移除 flow key 后，hash 入口可进入详情，但从 “New resume” 再次提交仍停留 create；第二轮 RED 才要求 handler 显式退出 create。
  - **影响**：产生一次额外修复循环，也证明同一 UI 动作存在两种状态来源时不能只测一个入口。
- Change-intake matcher 将 detail/listing owner 排在当前 active create-flow owner 之前。
  - **证据**：相同问题查询以 high confidence 推荐 `001-listing-routing-and-detail-readonly`，而 `002-create-flow` 仅排第二；当前计划的 Phase 3 和 10.3 明确拥有创建后直达详情。
  - **影响**：本次通过人工 owner 语义核对避免了错误重开 sibling owner，但自动路由仍存在误导风险。

## 3 根因归类

- 导航调用与目标 DOM 之间缺少行为断言，根因属于 `spec-plan` gate 粒度不足；本批已在原 owner 10.3 和 UI contract 原地修复。
- CreateFlow 的 hash 参数入口与本地模式入口没有作为两条独立路径列出，根因属于 `spec-plan` 路径矩阵不完整；本批浏览器 gate 已覆盖两者。
- Matcher 更看重通用 route/detail 关键词，未充分利用 action vocabulary、active owner 和精确 Phase 语义，根因属于 `skill`。
- Product context target 首次传成 `product-scope` 而非声明的 `cross-layer`，属于一次性调用错误；validator 已给出可用 target，无需仓库改动。

## 4 对流程资产的改进建议

- 为 `.agent-skills/change-intake` matcher 增加 active owner tie-break，并提高 `create/save/register` 等 action 词与 plan discovery keyword 的精确匹配权重。
  - **落点**：skill
  - **优先级**：high
- UI owner checklist 遇到同一路由内的 transient mode + local asset handoff 时，显式列出“URL 参数入口”和“组件内本地入口”，并要求目标 DOM 而非 navigation spy。
  - **落点**：spec-plan
  - **优先级**：medium
- 保留当前 source contract 与 browser smoke 的分工：source contract 锁定 identity/state 不变量，browser smoke 验证等待/就绪页面结果；不引入新的状态持久化层。
  - **落点**：无需仓库改动
  - **优先级**：medium

## 5 建议优先级与后续动作

- 下一轮优先审查并 TDD 修复 change-intake matcher 的 owner 排序，使用本次查询作为回归 fixture，确保 active `002-create-flow` 胜过通用 listing/detail owner。
- 后续 UI 技术债扫描继续区分 source-level zero-reference 与 runtime state handoff；只有行为路径存在时才扩展 browser gate，避免为纯删除任务增加无关抽象。
