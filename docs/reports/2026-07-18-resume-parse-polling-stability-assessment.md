# Resume Parse Polling Stability 交付复盘报告

> **日期**: 2026-07-18
> **审查人**: Codex

**关联计划**: [Resume Workshop Listing Routing and Detail Readonly](../spec/frontend-resume-workshop/plans/001-listing-routing-and-detail-readonly/plan.md)
**关联 Bug**: [BUG-0187](../bugs/BUG-0187.md)

## 1 复盘范围与成功证据

- 本次修复覆盖 Resume 详情处于 queued/processing 时的后台轮询稳定性：已有数据的 poll 保留解析等待 DOM，仅首次读取或资源身份变化时进入通用 loading。
- TDD RED 以首轮 processing 与第二轮 pending Promise 稳定复现闪现；GREEN 后 CSS/组件 focused 17 tests 与 frontend typecheck 通过。
- 真实 Chrome 连接正式 frontend/backend 连续采样 40 次：解析等待态 40/40，通用 loading 闪现 0 次；图标、标题与说明的几何采样稳定，computed transform 恒为 `none`。
- 根 `make test` 通过 Python 584 tests/4583 subtests、Go 全包与 frontend 126 files/1029 tests；owner context、docs/index/diff gates 均通过。

## 2 会话中的主要阻点/痛点

- 原有 polling 测试只验证最终 ready 和请求次数，没有固定第二轮请求的 pending 窗口。
  - **证据**：补充 pending Promise 后，旧实现精确失败于 `resume-detail-parse-waiting` 消失并显示通用 loading。
  - **影响**：历史测试得到假绿，未覆盖请求间隙的用户可见状态。
- 初始现象同时包含整页 loading 闪现与图标亚像素抖动。
  - **证据**：CSS keyframes 中的 scale 只能解释图标自身抖动；后台 poll 的 `setData(null)` 才是整块内容交替的原因。
  - **影响**：需要分别修复数据状态和动画几何，不能仅调整视觉样式。

## 3 根因归类

- `useResume` 未区分首次加载与已有 pending data 的后台刷新，且 plan 未锁定请求间隙的 DOM 不变式。
  - **类别**：spec-plan。已在原 owner Phase 21 补充 stable polling contract、pending-window test 和真实 Chrome 采样 gate。
- 循环 scale 将进行中反馈与元素几何变化绑定。
  - **类别**：spec-plan。UI design 与 CSS owner test 已锁定只允许 opacity/不参与布局的光晕变化。

## 4 对流程资产的改进建议

- 保留 Phase 21 的 pending Promise 断言，所有“有旧数据的后台刷新”都应验证请求间隙的 DOM，不只验证最终响应。
  - **落点**：spec-plan
  - **优先级**：high
- 解析等待动画继续由 CSS source contract 与 reduced-motion 断言约束，禁止恢复循环 scale/translate。
  - **落点**：spec-plan
  - **优先级**：medium

## 5 建议优先级与后续动作

- 最高优先级：未来修改 `useResume` 请求身份、retry 或 polling 时，继续运行 Phase 21 pending-window test，防止通用 loading 闪现回归。
- 下一步推荐：将同一“保留 stale data 直到刷新完成”原则用于后续发现的其他后台刷新页面，但不在本次修复中扩大重构范围。
- 可延后：若解析等待视觉方向调整，先修订 `docs/ui-design/resume-onboarding.md`，再更新正式 frontend 与 CSS contract test。
