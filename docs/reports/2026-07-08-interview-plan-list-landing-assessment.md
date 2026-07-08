# Interview Plan List Landing 交付复盘报告

> **日期**: 2026-07-08
> **审查人**: Codex

## 1 复盘范围与成功证据

本次交付覆盖 `frontend-workspace-and-practice/plans/001-workspace-and-interview-context` Phase 7：一级导航 `模拟面试` 收敛为 `面试` / `Interview`，`workspace` 无上下文时展示面试规划列表，带 `targetJobId` / `planId` 等上下文时继续展示当前面试规划详情。

成功证据：

- `pnpm --filter @easyinterview/frontend typecheck` PASS。
- 聚焦 Vitest：TopBar、i18n、Workspace plan-list、handoff、AppNormalize 共 8 个测试文件 / 69 tests PASS。
- `E2E.P0.018` scenario `setup -> trigger -> verify -> cleanup` PASS，12 个测试文件 / 95 tests PASS。
- 路由回归：`p0-088`、`p0-089`、`p0-090`、`AppRoutingHistory`、`AppRoutingPrivacy` 共 5 个测试文件 / 38 tests PASS。
- `pnpm --filter @easyinterview/frontend build` PASS。
- Playwright parity：`topbar.spec.ts` + `workspace.spec.ts` 共 50 tests PASS。
- `make docs-check` PASS，Header / INDEX / markdown links zero drift。

## 2 会话中的主要阻点/痛点

- 文案断言最初使用子串匹配，`Mock Interview` 会误判通过 `Interview`。
  - **证据**：第一次聚焦红测中 TopBar 相关测试通过，但失败输出里仍可见旧 `Mock Interview` 文案。
  - **影响**：如果不改为正则精确匹配，导航文案回归会被测试漏掉。
- Playwright parity 首轮跑到旧 `frontend/dist`。
  - **证据**：首轮 Playwright 失败截图和 diff 显示 frontend dist 仍为 `Mock Interview`；执行 `pnpm --filter @easyinterview/frontend build` 后同一组 50 tests PASS。
  - **影响**：造成一次非产品缺陷的 parity 失败，增加定位成本。
- 场景脚本外层包装使用了 zsh 只读变量名 `status`。
  - **证据**：首次 P0.018 输出已出现 `E2E.P0.018 PASS`，但外层命令以 `read-only variable: status` 失败；改用 `scenario_exit` 后完整脚本成功退出 0。
  - **影响**：造成一次假失败，但未影响场景脚本本身。

## 3 根因归类

- 精确文案断言不足。
  - **类别**：spec-plan
  - **说明**：Phase 7 需要验证“更短文案”，测试应使用 exact text，而不是 substring。
- Pixel parity 依赖 dist 构建产物。
  - **类别**：README
  - **说明**：当前 parity runner 服务 `frontend/dist` + `/ui-design/`，在源码变更后需要先 build，否则测试会验证旧产物。
- shell 包装变量名冲突。
  - **类别**：no repo change needed
  - **说明**：这是本次执行命令的一次性失误，场景脚本契约无需修改。

## 4 对流程资产的改进建议

- 在涉及文案收敛的 plan checklist 中明确“导航文案必须精确断言，不能只做 substring match”。
  - **落点**：spec-plan
  - **优先级**：medium
- 在 pixel parity README 或相关场景说明中强调源码变更后先执行 `pnpm --filter @easyinterview/frontend build`，再跑 `test:pixel-parity`。
  - **落点**：README
  - **优先级**：medium
- 保持场景脚本自身的 `setup -> trigger -> verify -> cleanup` 契约不变。
  - **落点**：no repo change needed
  - **优先级**：low

## 5 建议优先级与后续动作

最高价值后续动作是把“pixel parity 前先 build”的前置条件固化到 `frontend/README.md` 或 pixel parity 运行文档，减少 stale dist 假失败。其次，在后续涉及 UI 文案简化的 checklist 中加入 exact text assertion gate，避免旧文案被子串测试放过。
