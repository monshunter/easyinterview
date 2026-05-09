# Expected Outcome

- testid `workspace-crumbs` / `workspace-plan-eyebrow-*` / `workspace-plan-action-*` 命中
- `workspace-header-*` / `workspace-cta-start` / `workspace-binding-{jd,resume}` 命中
- `workspace-companyintel-{summary,open}` / `workspace-jd-block-{must,nice,hidden}` 命中
- `workspace-history-card` / `workspace-history-empty` 命中（EmptyHistory placeholder）
- Plan Switcher Modal `workspace-plan-modal-*` testid 命中
- Resume Picker Modal `workspace-resume-modal-*` testid 命中，generated client `listResumes` 调用次数 0
- 旧 prototype testid（`practice-mode-card-*` / `growth-*` / `drill-builder-*`）0 命中
- Vitest 输出 `Tests   all passed`
