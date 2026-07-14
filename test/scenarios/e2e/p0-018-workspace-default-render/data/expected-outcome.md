# Expected Outcome

- TopBar `workspace` 中文显示 `面试`、英文显示 `Interview`
- 无上下文 workspace 命中 `workspace-plan-list-*`，且 `workspace-empty` 不命中
- plan-list ready card 选择导航到 `/workspace?targetJobId=...`，且 URL 只保留 `targetJobId`；不注入 `planId/resumeId/reportId/jobId/jdId`
- 带 `targetJobId` 的 Workspace 命中 `unified-plan-detail` / `unified-plan-detail-title`，不命中 `workspace-plan-list-*` 或 Parse loading 动画
- ready card 点击后 `getTargetJob` 底层 GET 恰好 1 次，import/update/poll 为 0；列表 StrictMode 下 `listTargetJobs` 底层 GET 恰好 1 次
- 统一详情页 `parse-basics-*` / `parse-requirement-*` / `parse-hidden-signal-*` / `parse-round-*` / `parse-launch` / `parse-resume-binding` 命中，`立即面试` 直接进入 practice
- 轮次卡依次命中 `done/current/pending`，对应“已进行/即将进行/未进行”，三种背景和边框色均不同
- workspace 运行模块不包含 `WorkspaceHeader` / `PlanSwitcherModal` / `ResumePickerModal` / `useStartPractice` / `WorkspaceInsightCard` 等范围外详情或启动上下文
- out-of-scope prototype testid（`practice-mode-card-*` / `growth-*` / `drill-builder-*`）0 命中
- Vitest 输出 `Tests   all passed`
