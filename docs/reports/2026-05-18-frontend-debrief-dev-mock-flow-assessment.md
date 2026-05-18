# Frontend Debrief Dev Mock Flow 交付复盘报告

> **日期**: 2026-05-18
> **审查人**: Codex

## 1 复盘范围与成功证据

- 范围：修复 `frontend-debrief/001-debrief-screen-and-handoff` 在默认 Vite dev fixture-backed mock 下无法从 Step 1 `复盘分析` 进入 Step 2 `复盘面试` 的问题，见 [BUG-0078](../bugs/BUG-0078.md)。
- 代码修复：`frontend/src/api/devMockClient.ts` 增加 stateful debrief job / debrief-derived practice plan-session scenario selection；`openapi/fixtures/Jobs/getJob.json` 增加 `debrief-succeeded` fixture。
- 文档修复：`frontend-debrief` spec/history/plan/checklist/test-checklist 原地升至 v1.5；`docs/bugs/PATTERNS.md` 模式 3 补充跨 operation dev mock 状态推进检查。
- 成功证据：
  - `pnpm --filter @easyinterview/frontend exec vitest run src/api/devMockClient.test.ts`：9 tests pass。
  - `pnpm --filter @easyinterview/frontend exec vitest run src/app/screens/debrief/DebriefScreen.test.tsx`：7 tests pass。
  - `pnpm --filter @easyinterview/frontend exec vitest run src/app/screens/debrief`：6 files / 22 tests pass。
  - `pnpm --filter @easyinterview/frontend typecheck`：pass。
  - `make validate-fixtures`：59 fixtures pass。
  - `make docs-check`：Header / INDEX / links pass。

## 2 会话中的主要阻点/痛点

- completed 状态掩盖了默认 dev mock 的用户可见断点。
  - **证据**：plan 001 已标记 completed，历史测试也通过，但用户在实际前端看到 Step 1 永久 `AI 分析中...`，Step 2 不可达。
  - **影响**：需要重新打开原计划，补充 fixture-backed dev mock full-flow 测试，而不是继续相信 stubbed component happy path。
- mock transport 的单 operation fixture 覆盖率不足以证明业务流程可跑通。
  - **证据**：`createDebrief` fixture 和 `getJob` fixture 都存在，`getDevMockFixtureOperationIds()` 也覆盖全部 generated operation，但 default `getJob` 返回的是 generic `report_generate/running`，与 `createDebrief` 返回的 debrief job 无状态关联。
  - **影响**：用户无需真实后端也应能预览的前端路径被卡住。
- replay handoff 的 derived fixture 已存在但默认 dev mock 没有自动选择。
  - **证据**：`createPracticePlan.json` 有 `debrief-derived`，`startPracticeSession.json` 有 `debrief-derived-first-question`，但无显式 `Prefer` 时仍走 baseline default。
  - **影响**：即使 Step 1 修复，Step 2 也可能落到不符合 debrief 语义的 baseline practice fixture。

## 3 根因归类

- 根因：默认 dev mock 缺少跨 operation 状态机，而计划收口只覆盖了 generated fixture 存在性和 stubbed client happy path。
  - **类别**：spec-plan。
- 根因：dev mock 模式库只记录了 API base URL 打错端口这一类 bootstrap 问题，没有覆盖 fixture-backed mock 已启用但流程状态不推进的变体。
  - **类别**：README / bug pattern。
- 根因：组件层测试没有至少一条使用真实 `createDevMockClient()` 的主路径回归。
  - **类别**：spec-plan。

## 4 对流程资产的改进建议

- 已完成：在 `frontend-debrief/001` 的 spec/plan/checklist/test-checklist 中加入默认 dev mock full-flow gate。
  - **落点**：spec-plan。
  - **优先级**：high。
- 已完成：在 `docs/bugs/PATTERNS.md` 模式 3 中加入跨 operation async job / derived handoff 状态推进检查。
  - **落点**：bug pattern。
  - **优先级**：high。
- 建议后续：对其它含 `create* -> getJob -> get*` 的前端主流程补一条 `createDevMockClient()` full-flow smoke，避免只靠 fixture inventory 和 stubbed client。
  - **落点**：各 owner plan test-checklist。
  - **优先级**：medium。

## 5 建议优先级与后续动作

- 最高优先级：把 `frontend-debrief/001` 本次新增的 dev mock full-flow regression 作为后续 debrief 前端变更的必跑 focused gate。
- 次优先级：盘点 report / resume / target import 等同样依赖 async job 的前端默认 mock 流程，确认是否存在类似只覆盖单 operation fixture 的假绿。
- 可延后：将 `createDevMockClient` 的状态映射抽象成通用 helper，等第二个 owner 出现相同模式后再抽取，避免当前阶段过度抽象。
