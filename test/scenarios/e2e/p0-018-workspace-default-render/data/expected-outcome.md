# Expected Outcome

- TopBar `workspace` 中文显示 `面试`、英文显示 `Interview`
- 无上下文 workspace 命中 `workspace-plan-list-*`，且 `workspace-empty` 不命中
- plan-list card 选择导航到 `parse?targetJobId=...`，并在存在真实绑定时携带 `planId/resumeId`；不注入 `reportId`，也不伪造 `jobId` / `jdId` / `plan-${targetJobId}` / `resume-unbound`
- 带上下文 workspace 仍命中 `workspace-plan-list-*`，不命中 `parse-error` / `缺少目标岗位 ID`
- 统一详情页由 parse route 命中 `route-parse` / `unified-plan-detail` / `unified-plan-detail-title`
- 统一详情页 `parse-basics-*` / `parse-requirement-*` / `parse-hidden-signal-*` / `parse-round-*` / `parse-launch` / `parse-resume-binding` 命中，`立即面试` 直接进入 practice
- workspace 运行模块不包含 `WorkspaceHeader` / `PlanSwitcherModal` / `ResumePickerModal` / `useStartPractice` / `WorkspaceInsightCard` 等旧详情或启动上下文
- 非当前 prototype testid（`practice-mode-card-*` / `growth-*` / `drill-builder-*`）0 命中
- Vitest 输出 `Tests   all passed`
