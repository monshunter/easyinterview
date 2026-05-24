# Frontend Parse Confirm Context Gate 交付复盘报告

> **日期**: 2026-05-24
> **审查人**: Codex

## 1 复盘范围与成功证据

本次复盘范围是 `frontend-home-job-picks-and-parse/001-home-jd-import-and-parse` 的 P0.016 Parse Confirm → Workspace handoff 修复：`ParseScreen` Confirm 必须在已登录 navigate 和未登录 `requestAuth(pendingAction)` 两条路径上携带完整 7 字段 interview context，并由浏览器级场景验证 `/workspace` route query 与目标 DOM state。

已通过的成功证据：

- Red gate：扩展 `ParseEdit.test.tsx` / `ParseAuthGate.test.tsx` 后先复现旧实现缺失 `jobId` / `roundName`。
- `pnpm --filter @easyinterview/frontend test src/app/screens/parse/ParseEdit.test.tsx src/app/screens/parse/ParseAuthGate.test.tsx`：2 files / 13 tests PASS。
- `pnpm --filter @easyinterview/frontend build`：PASS（仅保留既有 Vite chunk size warning）。
- `pnpm --filter @easyinterview/frontend exec playwright test tests/pixel-parity/parse.spec.ts --grep "confirm navigates to workspace missing-resume with complete interview context"`：desktop/mobile 2 tests PASS，并输出完整 `contextKeys=targetJobId,jobId,jdId,planId,resumeVersionId,roundId,roundName` 与 screenshotBytes marker。
- `E2E.P0.016` setup → trigger → verify → cleanup：PASS；trigger 覆盖 real-mode generated client gate、Parse Confirm component tests、frontend build 与 focused Playwright browser gate。
- `pnpm --filter @easyinterview/frontend test src/app/screens/parse`：5 files / 27 tests PASS。
- `validate_context.py`、`sync-doc-index --check`、`make docs-check`、`git diff --check`：PASS。
- 交付 commit：`880c0867 fix(frontend-home): preserve parse confirm context (BUG-0100)`。

## 2 会话中的主要阻点/痛点

### 2.1 Confirm handoff 手写字段集合，偏离 owner helper

- **证据**：`ParseScreen.handleConfirm` 旧实现手写 `targetJobId` / `jdId` / `planId` / `resumeVersionId` / `roundId`，没有复用 `interviewContextFromTargetJob(targetJob)`。
- **影响**：Home recent card 与 Parse Confirm 对 workspace context 的字段集合不一致，导致 `jobId` / `roundName` 在 Confirm 路径丢失。

### 2.2 E2E.P0.016 旧 gate 只证明组件 spy，没有浏览器 route evidence

- **证据**：旧 `trigger.sh` 只跑 real-mode client gate 和 `ParseEdit` / `ParseAuthGate` Vitest；旧 `verify.sh` 检查宽泛 pass marker，没有 Playwright spec、test title、contextKeys 或 screenshotBytes。
- **影响**：scenario wrapper 通过但无法证明真实 URL query、routeStore hydration、Workspace DOM state 与 screenshot 证据，形成 false-green。

### 2.3 Workspace default handoff 状态需要在 gate 中显式命名

- **证据**：浏览器 gate 初版期待直接看到 hydrated workspace header，但默认 `resumeVersionId=resume-unbound` 按当前产品语义进入 `workspace-missing-resume` 下一步状态。
- **影响**：如果 gate 只检查 `/workspace` URL 或错误的 header，容易把正确的缺简历下一步状态误判为失败，或者忽略目标 route 真实 DOM。

## 3 根因归类

- **spec-plan**：owner plan 已在 §3.7 定义 7 字段 mapping，但 P0.016 checklist / BDD checklist 曾残留 5 字段口径，导致实现和测试有空间继续手写子集。
- **scenario scripts / README**：`verify.sh` 旧版缺少 context key exact marker 和 browser DOM state marker；这类 route handoff 必须由 browser gate 覆盖。
- **no repo change needed**：Vite chunk size warning、`NO_COLOR` / `FORCE_COLOR` warning 和 `LC_ALL=C.UTF-8` locale warning 在本次验证中出现，但不影响 P0.016 context gate，暂不纳入本修复范围。

## 4 对流程资产的改进建议

- 对用户可见 route/context handoff，BDD scenario verify 必须检查 browser URL query、目标 route DOM state、exact context key marker 和 screenshot bytes。
  - **落点**：scenario scripts / Bug pattern
  - **优先级**：high

- 对已有 owner helper 的 handoff contract，生产代码和 tests 应反查 helper exact output，避免每个 screen 手写字段集合。
  - **落点**：spec-plan
  - **优先级**：high

- `frontend/tests/pixel-parity/parse.spec.ts` 现在同时承载 loading browser gate 与 confirm handoff browser gate；如果后续继续增加 parse route-flow gate，可以提取 fixture/mock helper，降低重复 mock route 成本。
  - **落点**：frontend test helper
  - **优先级**：medium

## 5 建议优先级与后续动作

最高优先级是按同一 owner plan 做一次 P0.014-P0.017 小范围 scenario regression：P0.015/P0.016 已补 browser gate，P0.014/P0.017 仍应确认没有被本轮 parse context 改动破坏。

第二优先级是后续若继续审计 `frontend-workspace-and-practice` 的 route hydration，优先检查它是否稳定消费这 7 个 query params，尤其是 `resumeVersionId=resume-unbound` 时的 `workspace-missing-resume` 分支。

可以延后处理的是 Playwright mock helper 抽象；当前重复范围仍小，过早抽象会增加阅读成本。
