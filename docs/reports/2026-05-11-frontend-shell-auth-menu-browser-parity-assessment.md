# Frontend Shell Auth Menu Browser Parity 交付复盘报告

> **日期**: 2026-05-11
> **审查人**: Codex

## 1 复盘范围与成功证据

- 范围：修复 `frontend-shell/001-app-shell-auth-settings` Phase 6 登录态与 `ui-design` 剩余 parity 空洞，重点是 authenticated avatar dropdown 的真实浏览器 desktop / mobile geometry、logout flow，以及 P0.006 pixel gate 对应契约。
- 成功证据：
  - `pnpm --filter @easyinterview/frontend exec playwright test tests/pixel-parity/topbar.spec.ts` PASS（22 tests）。
  - `pnpm --filter @easyinterview/frontend test:pixel-parity` PASS（112 passed）。
  - `E2E.P0.006` setup→trigger→verify→cleanup PASS，verify 断言 `112 passed`。
  - `E2E.P0.032` setup→trigger→verify→cleanup PASS。
  - `pnpm --filter @easyinterview/frontend test` PASS（108 files / 681 tests，存在既有 React `act(...)` warning）。
  - `pnpm --filter @easyinterview/frontend build`、两个 owner context validator、`make docs-check`、`git diff --check` 均 PASS。
- 关联 Bug：[`BUG-0041`](../bugs/BUG-0041.md)。

## 2 会话中的主要阻点/痛点

- 登录态 parity 的原先 gate 只到 jsdom / P0.032，无法证明真实浏览器浮层几何。
  - **证据**：新增 authenticated user menu Playwright 用例后，mobile 项首次复现 `left=-64.984375` overflow；此前 `P0.032` 和 focused Vitest 都能通过。
  - **影响**：已完成 Phase 6 仍存在 mobile 端可见偏差，且 `test:pixel-parity` 的 110 项旧口径无法捕捉。
- P0.006 场景契约和 README 的固定计数需要随测试扩展同步修订。
  - **证据**：新增 1 个 topbar 用例后，由 desktop/mobile 两个 project 扩展为 112 tests；`verify.sh` 需要从 `110 passed` 更新为 `112 passed`。
  - **影响**：若只加测试不改 scenario verify，会出现场景 gate 与真实 suite 不一致。
- `ui-design` golden preview 受 CDN/字体加载时序影响，旧等待条件在完整 suite 中容易表现为 timeout。
  - **证据**：历史失败点是 mobile ui-design nav wait timeout，但失败 trace 中 nav 实际已渲染；本次把等待锚定到可见 `nav button` 并提高 focused timeout。
  - **影响**：真实实现问题和对照页加载时序问题容易混在一起，增加诊断成本。

## 3 根因归类

- Auth menu browser geometry 缺口。
  - **类别**：spec-plan
  - **说明**：001 Phase 6 已要求源级复刻头像菜单，但 checklist 证据主要在 component/jsdom 与 P0.032；003 pixel gate 没有把 authenticated dropdown 打开态纳入真实浏览器断言。
- Scenario 固定计数漂移。
  - **类别**：README / spec-plan
  - **说明**：P0.006 README、expected outcome、verify.sh 和 owner checklist 都需要在新增 Playwright 用例时同步更新，否则 BDD gate 与真实 suite 不一致。
- Golden preview wait 过于脆弱。
  - **类别**：spec-plan
  - **说明**：外部 CDN 参与的对照页不应只等待 load 或节点存在；真实可见 DOM 才是 gate 可用的信号。

## 4 对流程资产的改进建议

- 已在本次 owner 资产中落地：001 plan/checklist 增加 Phase 6 browser-level authenticated user menu parity 和 operation matrix。
  - **落点**：spec-plan
  - **优先级**：high
- 已在本次 owner 资产中落地：003 plan/checklist/bdd-plan/bdd-checklist 与 P0.006 scenario 从 110 tests 更新到 112 tests，并要求 authenticated user menu geometry + logout flow。
  - **落点**：spec-plan / README
  - **优先级**：high
- 可选后续：把“浮层类 UI 必须有真实浏览器 opened-state geometry gate”抽象进 `docs/bugs/PATTERNS.md`。
  - **落点**：Bug pattern library
  - **优先级**：medium
  - **说明**：本次未直接修改 PATTERNS，因为 bug-report workflow 要求写入模式库前先确认。

## 5 建议优先级与后续动作

- 最高优先级：提交当前修复时使用同一个工作日志条目关联 `BUG-0041`，并保持 commit title 与 BUG 记录一致：`fix(frontend-shell): close auth menu browser parity gap`。
- 下一轮建议：执行 `/plan-code-review frontend-shell/003-ui-design-pixel-parity-gate frontend`，专门复查 P0.006 是否还有其他登录态 / 浮层 / mobile opened-state 的 browser parity 空洞。
- 可延后：如果后续再出现类似 dropdown / modal / popover 漏测，将本次经验抽成 `PATTERNS.md` 模式条目。
