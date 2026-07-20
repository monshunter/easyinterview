# Resume 与 Workspace 删除二次确认交付复盘报告

> **日期**: 2026-07-20
> **审查人**: Codex

## 1 复盘范围与成功证据

- 本次在 [Resume Workshop 001](../spec/frontend-resume-workshop/plans/001-listing-routing-and-detail-readonly/plan.md) 与 [Workspace/Practice 001](../spec/frontend-workspace-and-practice/plans/001-workspace-and-interview-context/plan.md) 原地补充删除二次确认，保留既有 `archiveResume` / `archiveTargetJob` 软归档 API 与 generated client 合同，不新增后端接口或浏览器 E2E 编号。
- RED 证明旧实现首次点击即调用 archive；GREEN focused suites 为 3 files / 23 tests，邻接前端回归为 9 files / 73 tests，frontend typecheck、production build 与根 `make test` 均通过。
- 共享确认框覆盖取消初始焦点、focus trap、Escape/遮罩关闭、trigger focus restore、pending 锁定、失败保留与 same-key retry；两个 consumer 分别锁定首次点击零 archive、确认单次提交与成功隐藏。
- host-run frontend 已重新部署，四个共享依赖 readiness 为 4/4 OK。真实 Chrome 在 1212×912 下分别验收 Resume 与 Workspace 弹窗：取消后卡片保留、焦点回到删除按钮、无横向溢出，Workspace 控制台 error 为 0；截图保存在 `.test-output/delete-confirmation-ui/`，并经 magic bytes 校验为真实 PNG。
- 两个 owner context、Header/INDEX、文档链接、合同 ID 与 `git diff --check` 均通过，临时重开的 spec/plan 已恢复 `completed`。

## 2 会话中的主要阻点/痛点

- 同一危险操作交互同时落在 Resume 与 Workspace 两个 owner，若各自复制弹窗，会出现 focus、pending 与失败重试语义漂移。
  - **证据**：两条旧删除路径都直接发 archive，但错误状态、删除中状态和卡片容器不同；本次必须由一个共享 dialog 组件承接交互不变量，再由两个 consumer 各自持有 API 与 idempotency key。
  - **影响**：实现与测试必须同时修改两个既有 plan，单 owner 局部修补无法完整关闭用户可见风险。
- Chrome 截图 API 返回 JPEG bytes，即使证据路径使用 `.png` 后缀；此外，首次复用用户已有设置页标签时页面状态会被现有浏览活动切回。
  - **证据**：`file` 与 magic bytes 首次识别为 JFIF；切换为同一登录会话中的临时验收标签后页面稳定，截图转换并复核为 PNG。
  - **影响**：若只检查扩展名会留下伪 PNG；复用活跃用户标签也会导致跨步骤 DOM 不一致与无效定位。

## 3 根因归类

- 删除确认的交互不变量此前未同时进入 Resume 与 Workspace 的 active behavior contract。
  - **类别**：spec/plan
- Settings 已有危险操作视觉，但其 DOM 与 CSS 由 Settings 私有实现持有，列表删除无法直接复用行为组件。
  - **类别**：spec/plan
- Chrome 截图编码和活跃标签切换属于本轮工具/浏览状态差异，现有 PNG magic-byte 与临时标签清理流程已足以发现并恢复。
  - **类别**：无需仓库改动

## 4 对流程资产的改进建议

- 后续新增危险操作时，优先复用 `DestructiveActionDialog`，consumer 只拥有目标对象、generated client、idempotency key 与成功后的本地投影；不得复制 focus trap、pending 锁定或失败重试状态机。
  - **落点**：相关 UI spec / plan / component tests
  - **优先级**：high
- 只有在 Settings 注销语义与共享组件合同完全一致并具备回归测试时，再评估把 Settings 私有确认框迁移到共享组件；本次不做顺带重构。
  - **落点**：frontend-shell Settings owner spec/plan
  - **优先级**：low
- 保留真实截图的 `file` 与 magic-byte 校验；Chrome 中已有用户标签不稳定时，使用同一已登录会话的新临时标签，验收完成后 finalize 清理。
  - **落点**：无需仓库改动
  - **优先级**：medium

## 5 建议优先级与后续动作

- 当前交付不需要新的功能修订；下一步优先对本次 diff 做一次 owner-scope L2 review，然后由用户决定是否连同现有同分支改动一起提交。
- 后续危险操作直接复用共享确认框及其 focused tests；Settings 迁移保持为低优先级候选，不阻塞本次交付。
- 本次属于用户体验增强，不是已确认的实现缺陷，按 Bug 建档阈值不新增 Bug 记录。
