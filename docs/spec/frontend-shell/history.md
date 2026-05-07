# Frontend Shell History

> **版本**: 1.9
> **状态**: active
> **更新日期**: 2026-05-08

## 1 修订记录

| 日期 | 版本 | 变更 | 关联计划 |
|------|------|------|----------|
| 2026-05-08 | 1.9 | 修正 TopBar 显示控制契约：语言切换按当前 `ui-design/src/app.jsx` 复刻为 icon toggle，主题控制按源码 menu + Custom row 内嵌 AccentPicker 复刻；旧 native select/dropdown 口径仅作为负向回流断言保留。 | 002-app-shell-visual-system |
| 2026-05-08 | 1.8 | 派生 D2 follow-up 003 plan：新增 §6 C-9 真实浏览器 pixel parity gate，要求 Playwright + chromium 在 desktop / mobile viewport 下断言 `frontend/dist/index.html` 与 `ui-design/index.html` golden preview 的 DOM 锚点 / computed style / bounding box / 截图差异；E2E.P0.005（jsdom）保留作为 fast smoke。 | 003-ui-design-pixel-parity-gate |
| 2026-05-07 | 1.7 | 删除废弃外部设计来源；前端视觉实施只以 `ui-design/` 与 `docs/ui-design/` 为唯一 UI 真理源，要求正式前端 100% 源级复刻静态原型并通过 parity gate 验证。 | 002-app-shell-visual-system |
| 2026-05-07 | 1.6 | 修订 D2 视觉系统接入门禁：确认 `ui-design/` 是验收真理源头，`customAccent` 必须进入正式前端主题系统，并新增 visual smoke 工具作为用户可见视觉渲染 gate。 | 002-app-shell-visual-system |
| 2026-05-07 | 1.5 | 派生 D2 视觉系统接入计划；新增 §6 C-8 视觉接入验收，将 `ui-design/` 真理源、4 主题 × 2 模式 wiring、字体与 D1 regression 固化为视觉接入门禁。 | 002-app-shell-visual-system |
| 2026-05-07 | 1.4 | 修订 i18n 初始语言规则：默认跟随浏览器 locale，未知时 fallback English；语言切换只关联前端显示偏好，不再由 runtime config 或登录态覆盖。 | 001-app-shell-auth-settings |
| 2026-05-07 | 1.3 | 收紧 i18n 架构：每种语言必须独立 locale 文件，TopBar 语言切换必须有可访问控件契约，聚合 helper 不得糅合多语言 message map。该控件结构已在 1.9 按当前 UI 真理源更新为 icon toggle。 | 001-app-shell-auth-settings |
| 2026-05-07 | 1.2 | 原地补齐 D1 i18n 决策：正式前端语言切换必须驱动 `zh` / `en` 静态文案、runtime locale 初始化和 `Accept-Language` display hint。 | 001-app-shell-auth-settings |
| 2026-05-05 | 1.0 | 从 engineering-roadmap S1 派生 frontend shell subject，锁定 App 壳、TopBar、auth pendingAction、用户菜单与设置入口。 | 001-app-shell-auth-settings |
