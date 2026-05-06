# Expected Outcome

- 渲染 Home 占位（`route-home`）。
- 一级导航严格为五项：`home` / `jd_match` / `workspace` / `resume_versions` /
  `debrief`，且 `topbar-primary-nav` 子项数量为 5。
- 全局显示控制 `topbar-theme-select` / `topbar-dark-toggle` /
  `topbar-lang-select` 可见。
- 未登录 user-area `topbar-user-area` 的 `data-signed-in` 为 `false`，渲染
  `topbar-login` 与 `topbar-register` 入口。
- 旧入口（`welcome` route、独立 `voice`、`growth` / `mistakes` / `drill`）
  在 DOM 中不可见。
- 场景日志中不得出现以上旧入口的 `data-testid` 字符串，否则视为回流污染。
