# Frontend Shell Auth State Parity 交付复盘报告

> **日期**: 2026-05-10
> **审查人**: Codex

## 1 复盘范围与成功证据

- 复盘范围：修复 `frontend-shell/001-app-shell-auth-settings` 已完成计划中的登录态漂移，包括已登录 TopBar 用户区与 `ui-design/src/app.jsx` 对齐、Vite dev fixture mock 默认非登录、登录成功、logout 后非登录闭环。
- 成功证据：
  - `E2E.P0.032` 场景目录、数据、setup/trigger/verify/cleanup 已创建并执行通过。
  - `pnpm --filter @easyinterview/frontend test` PASS（108 files / 681 tests）。
  - `pnpm --filter @easyinterview/frontend typecheck` PASS。
  - `pnpm --filter @easyinterview/frontend build` PASS。
  - `make docs-check` PASS。
  - `validate_context.py --context docs/spec/frontend-shell/plans/001-app-shell-auth-settings/context.yaml --docs-root docs --target frontend` PASS。
  - 与本次 UI 变更直接相关的 Playwright gate：`topbar.spec.ts`、`layout.spec.ts`、`screens.spec.ts` PASS（40 tests）。

## 2 会话中的主要阻点/痛点

- 已完成计划的历史 PASS 容易误导判断。
  - **证据**：`001-app-shell-auth-settings` 早已勾选用户菜单入口，`003-ui-design-pixel-parity-gate` 也为 completed，但用户截图显示正式 TopBar 仍是旧 inline 三按钮。
  - **影响**：必须先做 deep reconcile，重新读取 `ui-design`、docs、实现与 tests，不能直接相信 checklist。
- Mock runtime 没有建模 session transition。
  - **证据**：旧 `createDevMockClient()` 只是 fixture-backed fetch；`logout` 成功后 `/me` 仍由默认 authenticated fixture 响应。
  - **影响**：真实 dev 页面无法退出登录，也没有默认非登录态，导致用户看到的状态流和计划验收相反。
- 既有测试只断言菜单文案，没断言交互结构。
  - **证据**：旧 TopBar/i18n/locale tests 查找 `topbar-user-profile/settings/logout` 常驻存在；修复为 dropdown 后，这些旧断言需要改成“点击 chip 后出现”。
  - **影响**：历史测试会鼓励实现继续保留错误结构。
- 完整 pixel suite 暴露相邻测试债务。
  - **证据**：完整 `test:pixel-parity` 中 TopBar/layout 相关测试通过，但 workspace hydrated tests 因 `resume-unbound` 上下文进入 missing-resume 状态；Darwin 本地 screenshot baseline 也缺失。
  - **影响**：本次只能把直接相关 Playwright gate 作为交付证据，并把 workspace/pixel 前提漂移列为后续 owner 工作。

## 3 根因归类

- 用户菜单 parity 漂移：
  - **类别**：spec-plan
  - **说明**：原 plan/checklist 的 gate 粒度不足，只覆盖文本和 route 分流，未固化 `ui-design` 的头像 chip + dropdown DOM/交互不变量。
- Dev mock 登出无效：
  - **类别**：spec-plan
  - **说明**：mock-contract 与 frontend-shell plan 没有要求 fixture-backed dev mock 保持跨 operation session state。
- 旧测试断言常驻菜单：
  - **类别**：no repo change needed
  - **说明**：已在本次代码测试中修复，作为本次 implementation 的一部分闭环。
- 完整 pixel suite workspace 失败：
  - **类别**：spec-plan
  - **说明**：workspace pixel tests 的 hydrated path 依赖过期上下文假设，应由 workspace/pixel owner 修订。

## 4 对流程资产的改进建议

- 在 `frontend-shell/001-app-shell-auth-settings` 中保留 Phase 6 gate。
  - **落点**：spec-plan
  - **优先级**：high
  - **建议**：后续任何用户菜单/登录态改动必须复跑 `E2E.P0.032`，并保留旧 inline 三按钮负向断言。
- 为 mock-contract-suite 增加 stateful auth mock 口径。
  - **落点**：spec-plan
  - **优先级**：medium
  - **建议**：把 `getMe` 默认非登录、verify 后登录、logout 后非登录这类跨 operation 状态流纳入 mock runtime contract，而不只由 frontend dev mock 局部保证。
- 修订 workspace pixel hydrated path。
  - **落点**：spec-plan
  - **优先级**：medium
  - **建议**：workspace/pixel owner 应提供显式有效 `resumeVersionId` 的真实浏览器进入方式，或增加受控 test entry，不再依赖 `resume-unbound` recent card。

## 5 建议优先级与后续动作

- 最高优先级：执行一次 `/plan-code-review frontend-shell/001-app-shell-auth-settings frontend`，专门复查 Phase 6 与 `ui-design/src/app.jsx` 的剩余 parity 空洞，避免还有未覆盖的已登录 dropdown 视觉细节。
- 次优先级：派生或修订 workspace/pixel owner plan，修复完整 `test:pixel-parity` 中 hydrated workspace 与本地 screenshot baseline 前提问题。
- 可延后：把 stateful auth mock 从 frontend dev mock 抽象进更通用的 mock-contract-suite 口径，供后续 backend/frontend 双端共享。
