# Frontend Workspace L2 Remediation 交付复盘报告

> **日期**: 2026-05-09
> **审查人**: Codex

## 1 复盘范围与成功证据

- 范围：`frontend-workspace-and-practice/001-workspace-and-interview-context` 的 L2 code review remediation，覆盖 route → `InterviewContext` hydrate、Workspace modal 接线、`getPracticePlan` refresh/replacement、`autoStartPractice=1` auth 恢复、BDD 脚本加固、workspace full pixel parity 与 plan lifecycle 收口。
- 代码验证：`pnpm --filter @easyinterview/frontend test` 66 files / 423 tests PASS；`pnpm --filter @easyinterview/frontend build` PASS；`pnpm --filter @easyinterview/frontend test:pixel-parity` 96/96 PASS。
- 场景验证：`test/scenarios/e2e/p0-018-workspace-default-render`、`p0-019-workspace-context-loading`、`p0-020-workspace-start-practice`、`p0-021-workspace-handoff` 均 `setup -> trigger -> verify -> cleanup` PASS。
- 文档与契约验证：`make validate-fixtures` PASS；`make docs-check` zero drift；context validator PASS；`git diff --check` clean。
- 生命周期收口：001 plan / checklist / BDD plan / BDD checklist 更新到 `v1.1 completed`，plans INDEX 同步；新增 [BUG-0030](../bugs/BUG-0030.md) 记录本次 L2 漂移。

## 2 会话中的主要阻点/痛点

- 完成态证据没有覆盖真实 App runtime path。
  - **证据**：L2 发现 `App.tsx` 未自动 hydrate `InterviewContext`，旧测试依赖手动 context 前置；修复后新增 App route hydration 覆盖。
  - **影响**：workspace 从真实 route 进入时会丢失 `targetJobId/resumeVersionId/planId` 等关键上下文，后续 hook 与 CTA 行为无法闭环。
- 组件局部测试掩盖正式屏幕接线缺口。
  - **证据**：`PlanSwitcherModal` / `ResumePickerModal` 组件存在，但 `WorkspaceScreen` 未挂载；修复后新增 `WorkspaceModalIntegration.test.tsx` 与 pixel parity modal anchors。
  - **影响**：用户点击 `切换规划` / `更换简历` 时无法进入计划要求的 modal 流程。
- BDD checklist 与 scenario scripts 没有同步真实断言强度。
  - **证据**：`bdd-checklist.md` 仍未勾选，P0.018-P0.021 scripts 主要依赖弱 entry；修复后 trigger/verify 加入具体 Vitest entry、runtime negative grep 与 completion marker。
  - **影响**：历史 PASS 不能证明当前实现符合新版 spec/plan，L2 需要重新补 gate。
- full workspace pixel parity 晚于功能实现才暴露 shared shell mobile 风险。
  - **证据**：扩展 `workspace.spec.ts` 到 full workspace + modal + desktop/mobile 后，同步调整 TopBar mobile contract 与 snapshots。
  - **影响**：workspace 自身修复牵动 shared shell responsive 断言，增加了收口范围。

## 3 根因归类

- Route-level context hydrate 未被列为 completion 必跑 gate。
  - **类别**：spec-plan
- plan-code-review 需要更明确地区分“组件存在”与“正式 owner screen 已接线”。
  - **类别**：skill
- BDD checklist 没有强制以 trigger/verify 脚本的当前断言内容反向更新。
  - **类别**：spec-plan
- UI pixel parity 在 full route 场景前未覆盖 shared chrome + workspace 组合布局。
  - **类别**：spec-plan

## 4 对流程资产的改进建议

- 在后续 frontend feature plan 的 Phase 6 gate 中加入“App runtime route path”固定项：从 route params 进入 screen，验证 context hydrate、auth pendingAction restore、最终 navigation params。
  - **落点**：spec-plan
  - **优先级**：high
- 在 `/plan-code-review` 检查清单中增加“formal owner screen wiring”反查项：组件、hook、modal 已存在不等于用户可触达，必须从正式 screen button/action 反查到 DOM 或 side effect。
  - **落点**：skill
  - **优先级**：high
- 对 BDD checklist 增加 post-fix consistency gate：每个已勾选场景必须能指向当前 trigger/verify 中的具体断言，而不是只指向历史场景目录存在。
  - **落点**：spec-plan
  - **优先级**：medium
- 对 UI route plan 的 pixel parity 要求补充“shared shell + target screen combined viewport”覆盖，尤其是 TopBar 与 mobile wrapping。
  - **落点**：spec-plan
  - **优先级**：medium

## 5 建议优先级与后续动作

- 最高价值：下一轮进入 `frontend-workspace-and-practice` 的 `002-practice-text-event-loop` 前，先用 `/plan-review frontend-workspace-and-practice/002-practice-text-event-loop --fix` 或对应设计入口，把 App runtime route path、owner screen wiring、BDD trigger/verify 强断言写进 plan/checklist。
- 次优先：后续可以单独改进 `/plan-code-review` skill，把“组件存在 vs 正式用户路径可触达”固化为 L2 review 必查项；这能减少同类 frontend L2 返工。
- 可延后：TopBar/shared shell mobile 组合布局目前已有 pixel parity 覆盖，暂不需要额外 governance 变更，等下一次 shared chrome 改动时再提炼成通用规则。
