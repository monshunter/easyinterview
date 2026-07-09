# Expected Outcome

- TopBar `workspace` 中文显示 `面试`、英文显示 `Interview`
- 无上下文 workspace 命中 `workspace-plan-list-*`，且 `workspace-empty` 不命中
- plan-list card 选择导航到 `workspace?targetJobId=...&jobId=...&jdId=...&planId=...&resumeId=...`，不注入 `reportId`，也不伪造 `plan-${targetJobId}` / `resume-unbound`
- 带上下文详情页 testid `route-workspace` / `unified-plan-detail` / `unified-plan-detail-title` 命中
- 统一详情页 `parse-basics-*` / `parse-requirement-*` / `parse-hidden-signal-*` / `parse-round-*` / `parse-launch` / `parse-resume-binding` 命中
- 普通详情页 `workspace-header` / `workspace-launcher` / `workspace-jd-card` / `workspace-prep-card` / `workspace-history-card` 0 命中
- 统一详情页简历选择器 `parse-resume-picker*` 通过 flat `listResumes` active list 渲染
- 非当前 prototype testid（`practice-mode-card-*` / `growth-*` / `drill-builder-*`）0 命中
- Vitest 输出 `Tests   all passed`
