# Frontend Workspace Follow-up Review 交付复盘报告

> **日期**: 2026-05-09
> **审查人**: Codex

## 1 复盘范围与成功证据

- 本次交付范围：修复 `frontend-workspace-and-practice/001-workspace-and-interview-context` L2 follow-up review 中的四类问题：server-bound id 归一化、target-job stale/error recovery、target 切换竞态、workspace 派生标签本地化。
- 同步修订 owner plan/checklist 到 v1.2 completed，并新增 [BUG-0032](../bugs/BUG-0032.md) 记录本次缺陷模式。
- 成功证据：
  - Focused Vitest：`buildCreatePlanRequest`、workspace hooks、`WorkspaceStartPractice`、`WorkspaceEmptyState`、`WorkspaceScreen`、`WorkspaceHeader` 共 58 tests PASS。
  - `pnpm --filter @easyinterview/frontend typecheck`：PASS。
  - `pnpm --filter @easyinterview/frontend test`：67 files / 432 tests PASS。
  - `pnpm --filter @easyinterview/frontend build`：PASS。
  - `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check`：zero drift。

## 2 会话中的主要阻点/痛点

- Full test 暴露了 review comment 未直接点名的 runtime flicker。
  - **证据**：focused tests 通过后，full Vitest 中 `WorkspaceHeader` 偶发回到 placeholder；进一步检查发现 hooks 依赖整个 runtime context。
  - **影响**：需要额外修复 `useWorkspaceTargetJob` / `useWorkspaceResume` / `useWorkspacePracticePlan` 的 effect dependency，否则生产路径也会因 `/me` 或 runtime config 状态变化重复清空数据。
- 旧测试仍把 synthetic 或非 UUID route params 当成可渲染 full workspace 的前置条件。
  - **证据**：`App.test.tsx` 使用 `targetJobId: "tj-1"`，修复后按 server-bound ID 规则正确进入 empty state，导致旧断言失败。
  - **影响**：需要同步测试 fixture，使测试表达 OpenAPI UUID contract，而不是依赖历史 placeholder 行为。
- 本地化修复需要覆盖派生 view-model，而不只是静态字典。
  - **证据**：JD block labels、round fallback、status/source/prep derived labels 分散在 `WorkspaceScreen.tsx` helper 和 render config 中。
  - **影响**：单纯补 locale key 不够，必须把 render-time config 改为 key-based label 并加 English negative assertions。

## 3 根因归类

- UI context 与 generated client boundary 的职责没有在 plan gate 中显式分开。
  - **类别**：spec-plan
- workspace hooks 的 async keying 只覆盖目标 ID 变化，未覆盖 runtime context object identity 变化引发的二次 fetch。
  - **类别**：spec-plan
- 旧测试 fixture 允许非 UUID route params 混入 full workspace path，削弱了 OpenAPI generated client 合约反馈。
  - **类别**：spec-plan
- 本地化 gate 只覆盖 namespace 存在性，没有覆盖 derived helper 输出。
  - **类别**：spec-plan

## 4 对流程资产的改进建议

- 在 workspace/practice 后续 plan 的 checklist 中加入 “UI-only context id vs server-bound id boundary” gate：所有 generated client path/body 输入必须先过 server-bound ID 归一化或显式校验。
  - **落点**：spec-plan
  - **优先级**：high
- 对所有 data-fetch hooks 增加 “effect dependency must be stable key + client, not mutable runtime context object” 的 review checklist。
  - **落点**：spec-plan
  - **优先级**：medium
- 对 i18n gate 增加 “derived labels / helper maps / render config must be locale-key based” 的 negative English assertions，不能只检查 locale files key parity。
  - **落点**：spec-plan
  - **优先级**：medium

## 5 建议优先级与后续动作

- 最高优先级：在 plan `002-practice-text-event-loop` 启动前，把 server-bound ID boundary 和 async hook keying gate 写入该 plan/checklist，避免 practice/generating 继续继承 workspace 的 synthetic id 漂移。
- 可延后：把上述 hook dependency pattern 提炼进通用前端 review checklist 或 Bug pattern library；当前已由 BUG-0032 和本复盘沉淀，后续若再次出现同类问题再升级为通用 PATTERNS 条目。
