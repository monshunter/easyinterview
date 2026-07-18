# Practice Launch Transition 交付复盘报告

> **日期**: 2026-07-18
> **审查人**: Codex

**关联计划**: [Workspace and Interview Context](../spec/frontend-workspace-and-practice/plans/001-workspace-and-interview-context/plan.md)
**关联 Bug**: [BUG-0188](../bugs/BUG-0188.md)

## 1 复盘范围与成功证据

- 本次修复覆盖 Home recent、Workspace list/detail、Report replay/next-round 四类正式入口在 opening LLM 请求期间的共享全屏过渡态，以及成功导航、失败恢复、可访问性和 reduced-motion 契约。
- TDD RED 通过冻结 start Promise 在四类 caller 上复现无反馈状态；GREEN 后共享组件与 caller focused tests 为 5 files / 45 tests PASS，frontend typecheck PASS。
- Chrome skill 连接真实本地 frontend/backend，分别在 1440×900 与 390×844 视口捕获 LLM pending 状态：全屏 fixed overlay、交互阻断、滚动锁定、无横向溢出均符合契约；请求完成后正常进入 `/practice` 并显示 opening prompt。
- 根 `make test` 通过 Python 584 tests/4583 subtests、Go 全包与 frontend 127 files/1035 tests；`make lint`、frontend production build、`make docs-check`、index/context/diff gates 全部通过。

## 2 会话中的主要阻点/痛点

- 原 spec 的 `Loading: conversation skeleton` 描述的是会话页面，无法覆盖导航前的 opening LLM 等待。
  - **证据**：四类 caller 都在 `await startPracticeFromParams(...)` 返回后才导航；请求 pending 时 `/practice` 尚未挂载。
  - **影响**：如果只在 Practice 页面增加 skeleton，用户仍会在原入口感知页面卡死，形成看似完成但生命周期错位的修复。
- 四类入口各自拥有 in-flight 状态，视觉反馈却没有共享 owner。
  - **证据**：Home 使用 `startingRecentJobId`、Workspace 使用 `startingJobId`、详情使用 `confirming`、Report 使用 replay `starting`；旧实现都只用于禁用按钮。
  - **影响**：需要用一个视觉组件统一契约，同时保留各 caller 的错误恢复与导航边界，不能把业务 command 状态搬进新的平行 store。
- Chrome 响应式视口覆盖与可见截图尺寸不同步。
  - **证据**：390×844 的 DOM/computed-style 证据完整且无横向溢出，但普通截图按 Chrome 当前可见区域输出为 390×185；full-page capture 会拼接 fixed overlay 与底层长页。
  - **影响**：移动端验收需要以 viewport/DOM metrics 为主、截图为辅；这是本次工具表面限制，不是产品布局缺陷。

## 3 根因归类

- 缺少 pre-session 生命周期定义及 caller matrix。
  - **类别**：spec-plan。已在原 owner spec、Phase 30 plan/checklist、BDD plan/checklist 和 UI design 中原地补齐。
- 入口状态分散但等待视觉契约未抽取。
  - **类别**：spec-plan。实现已将 transition 抽为共享 UI owner，业务 command 状态继续由 caller 持有。
- Chrome 截图裁切与 viewport override 不一致。
  - **类别**：无需仓库改动。DOM、computed style、真实导航和桌面截图已形成充分且互补的验收证据。

## 4 对流程资产的改进建议

- 保留 Phase 30 的 caller coverage matrix；未来新增面试入口时，必须同时接入共享 transition，并用 pending Promise 覆盖 pending/success/failure。
  - **落点**：spec-plan
  - **优先级**：high
- 长耗时 UI command 的 plan 应明确反馈发生在 `before navigate` 还是目标页面内，避免用下游 skeleton 覆盖错误生命周期。
  - **落点**：spec-plan
  - **优先级**：medium
- Chrome 响应式验收继续同时记录 viewport/computed-style 与可见截图，不把 full-page stitched capture 当作 fixed overlay 的单一视觉证据。
  - **落点**：无需仓库改动
  - **优先级**：low

## 5 建议优先级与后续动作

- 最高优先级：以后新增面试启动入口时，直接复用 `PracticeLaunchTransition` 与 Phase 30 caller contract，避免再次出现只有按钮 disabled 的等待窗口。
- 下一步推荐：为其他“导航前等待 LLM”的正式流程做一次只读盘点；只有发现相同缺口时再回到对应 owner plan 修复，不在本次交付中扩大重构。
- 可延后：若过渡动画视觉方向需要调整，先修订 `docs/ui-design/module-practice-review.md`，再同步正式 frontend 与 reduced-motion/source-contract tests。
