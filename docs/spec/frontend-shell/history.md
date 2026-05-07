# Frontend Shell History

> **版本**: 1.4
> **状态**: active
> **更新日期**: 2026-05-07

## 1 修订记录

| 日期 | 版本 | 变更 | 关联计划 |
|------|------|------|----------|
| 2026-05-07 | 1.4 | 修订 i18n 初始语言规则：默认跟随浏览器 locale，未知时 fallback English；语言切换只关联前端显示偏好，不再由 runtime config 或登录态覆盖。 | 001-app-shell-auth-settings |
| 2026-05-07 | 1.3 | 收紧 i18n 架构：每种语言必须独立 locale 文件，TopBar 语言切换必须是下拉框，聚合 helper 不得糅合多语言 message map。 | 001-app-shell-auth-settings |
| 2026-05-07 | 1.2 | 原地补齐 D1 i18n 决策：正式前端语言切换必须驱动 `zh` / `en` 静态文案、runtime locale 初始化和 `Accept-Language` display hint。 | 001-app-shell-auth-settings |
| 2026-05-05 | 1.0 | 从 engineering-roadmap S1 派生 frontend shell subject，锁定 App 壳、TopBar、auth pendingAction、用户菜单与设置入口。 | 001-app-shell-auth-settings |
