# Frontend Shell History

> **版本**: 1.31
> **状态**: active
> **更新日期**: 2026-07-15

## 1 Change Log

| 日期 | 版本 | 变更 | 关联计划 |
|------|------|------|----------|
| 2026-07-15 | 1.31 | 用户确认删除 `ui-design/` 可运行 Demo 和双源 parity 合同，保留 `docs/ui-design/` 作为 UI 架构、流程、交互约束与设计决策 owner；正式前端直接实施和验证。 | [003-ui-design-responsive-browser-gate](./plans/003-ui-design-responsive-browser-gate/plan.md) |
| 2026-07-14 | 1.30 | 固化 StrictMode safe-read single-flight 的完整 identity 与 mutation fence：`okStatuses` 参与 key，所有语义写请求在 dispatch 前与 settle 后推进 read epoch；同时登记 command-only Parse、Workspace 只读详情与最小 custom accent 合同。 | 001-app-shell-auth-settings / 002-app-shell-visual-system / 004-url-addressable-routing |
| 2026-07-14 | 1.28 | 注册受保护且仅接受 `targetJobId` 的 `/reports` 上下文 route；保留 chrome 但不加入三入口 TopBar，缺失/非法 target 以 replace-only 回 workspace 且无 Back 循环，并删除 Parse reports section 与 detail/generating 非 reportId 参数权威。 | [004-url-addressable-routing](./plans/004-url-addressable-routing/plan.md) |
| 2026-07-07 | 1.26 | Active spec / history 压缩为当前 App shell、三入口 TopBar、email-code auth、profile setup guard、settings 双 tab、display preferences、canonical URL and protected route guard contract。 | product-scope/001-core-loop-module-pruning Phase 6.124 |
| 2026-07-07 | 1.25 | 对齐验证码、Browser History、安全 pendingAction 和账号记录可见性措辞。 | product-scope/001-core-loop-module-pruning |
| 2026-07-07 | 1.24 | 对齐当前 route catalog、auth flow、safe params and unsupported route normalization。 | product-scope/001-core-loop-module-pruning |
| 2026-06-12 | 1.22 | 对齐 email-code single-entry auth、settings profile/privacy tabs and `ocean` default theme。 | 001-app-shell-auth-settings |
| 2026-05-28 | 1.20-1.21 | 固化 profile setup guard、protected route guard、auth verify recovery and request-options refresh behavior。 | 001-app-shell-auth-settings |
| 2026-05-18 | 1.15-1.16 | 建立 Browser History canonical URL、safe-param allowlist、route store and SPA fallback contract。 | 004-url-addressable-routing |
| 2026-05-11 | 1.12-1.14 | 固化 dev mock session state、authenticated user menu geometry and browser responsive browser verification gate。 | 001-app-shell-auth-settings / 003-ui-design-responsive-browser-gate |
| 2026-05-08 | 1.8-1.11 | 固化 language dropdown、brand placement、theme controls and visual parity gates。 | 002-app-shell-visual-system / 003-ui-design-responsive-browser-gate |
| 2026-05-07 | 1.2-1.7 | 建立 i18n、UI design document、custom accent and visual smoke requirements。 | 001-app-shell-auth-settings / 002-app-shell-visual-system |
| 2026-05-05 | 1.0 | Seeded frontend-shell subject for App shell, TopBar, auth pendingAction, user menu and settings entrypoints。 | 001-app-shell-auth-settings |
