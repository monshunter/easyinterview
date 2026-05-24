# Frontend Parse Target Switch 交付复盘报告

> **日期**: 2026-05-24
> **审查人**: Codex

## 1 复盘范围与成功证据

本次复盘范围是 `frontend-home-job-picks-and-parse/001-home-jd-import-and-parse` 的 same-route `targetJobId` switch 修复：同一个 mounted `ParseScreen` 从已解析 preview 切到新的 `targetJobId` 时，必须清空旧 preview/edit state，重新进入 `ui-design` loading gate，并在 tick 完成后 hydrate 新 TargetJob。

已通过的成功证据：

- Red gate：新增 `ParseFlow.test.tsx` rerender regression 后，旧实现停留在旧 preview，`parse-loading-step-0` 不存在。
- `pnpm --filter @easyinterview/frontend test src/app/screens/parse/ParseFlow.test.tsx`：7 tests PASS。
- `pnpm --filter @easyinterview/frontend test src/app/screens/parse`：5 files / 28 tests PASS。
- `pnpm --filter @easyinterview/frontend build`：PASS（仅保留既有 Vite chunk size warning）。
- `E2E.P0.015` trigger → verify：PASS；包含 ready-response loading browser gate screenshotBytes marker。
- `E2E.P0.016` trigger → verify：PASS；包含 complete workspace contextKeys 和 screenshotBytes marker。

## 2 会话中的主要阻点/痛点

### 2.1 ready-response loading gate 没覆盖 route-param switch

- **证据**：BUG-0099 修复后，ready TargetJob 会先写入 `pendingReadyJob`，但旧 preview 下 `stage !== "loading"`，新的 ready job 无法 hydrate。
- **影响**：首次进入 parse 的 loading parity 被修好后，同一路由参数切换仍可能显示旧 TargetJob，属于用户可见 correctness drift。

### 2.2 组件测试缺少 same mounted rerender 路径

- **证据**：旧 `ParseFlow.test.tsx` 覆盖 initial ready、queued/processing polling、failed、re-parse 和 unmount cleanup，但没有 `rerender` 新 `targetJobId`。
- **影响**：App route 不加 key 时不会自动 remount；只测 initial mount 会漏掉 SPA 内导航路径。

### 2.3 本地 state reset 边界不明确

- **证据**：`targetJob`、editable fields、hit toggles、error state、pending ready state、loading completion 和 polling timeout 分散在多个 state/effect 中。
- **影响**：新 owner identity 到来时，如果不统一 reset，旧数据和新 pending state 容易交叉。

## 3 根因归类

- **frontend implementation**：`ParseScreen` 没有把 `targetJobId` 作为 screen state reset boundary。
- **test coverage**：缺少 same mounted route-param switch regression；initial mount 与 re-parse 测试无法覆盖该路径。
- **spec-plan**：owner spec/plan 只写了 ready response loading gate，没有把同一 mounted screen 切换 TargetJob 的 state reset 行为写入 D-2 和 checklist。

## 4 对流程资产的改进建议

- 对 route-param owner identity 的 screen，新增或审查时必须包含 same mounted rerender regression。
  - **落点**：Bug pattern / frontend tests
  - **优先级**：high

- 对 loading/pending state gate 的修复，要同时反查 stage 条件与 owner identity 切换路径。
  - **落点**：plan coverage matrix
  - **优先级**：high

- 后续如果更多 screen 复用 route params，可以考虑在 App route composition 层对高风险 screen 使用 explicit key，但需先评估是否会破坏保留草稿的交互。
  - **落点**：frontend routing design
  - **优先级**：medium

## 5 建议优先级与后续动作

最高优先级是对相邻的 `frontend-workspace-and-practice` route hydration 做一次同类 review：workspace 已依赖 `targetJobId` / `resumeVersionId` / `roundId` 等 params，应该确认同一 mounted workspace 切换上下文时不会保留旧会话 state。

可以延后处理的是 App route 层 key 策略；当前 `ParseScreen` 内部 reset 已闭合本次 bug，是否抽到 route composition 层需要等更多 screen 出现相同模式后再定。
