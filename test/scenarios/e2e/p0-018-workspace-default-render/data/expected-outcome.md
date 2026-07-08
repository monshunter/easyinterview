# Expected Outcome

- TopBar `workspace` 中文显示 `面试`、英文显示 `Interview`
- 无上下文 workspace 命中 `workspace-plan-list-*`，且 `workspace-empty` 不命中
- plan-list card 选择导航到 `workspace?targetJobId=...&jobId=...&jdId=...&planId=...`，不注入 `resumeId/reportId`
- 带上下文详情页 testid `workspace-crumbs` / `workspace-plan-eyebrow-*` / `workspace-plan-action-*` 命中
- `workspace-header-*` / `workspace-cta-start` / `workspace-binding-{jd,resume}` 命中
- `workspace-insight-{summary,open}` / `workspace-jd-block-{must,nice,hidden}` 命中
- 当前规划记录 placeholder 命中
- Plan Switcher Modal `workspace-plan-modal-*` testid 命中
- Resume Picker Modal `workspace-resume-modal-*` testid 命中，并通过 flat `listResumes` active list 渲染
- 非当前 prototype testid（`practice-mode-card-*` / `growth-*` / `drill-builder-*`）0 命中
- Vitest 输出 `Tests   all passed`
