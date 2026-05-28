# Frontend Shell History

> **版本**: 1.19
> **状态**: active
> **更新日期**: 2026-05-28

## 1 修订记录

| 日期 | 版本 | 变更 | 关联计划 |
|------|------|------|----------|
| 2026-05-28 | 1.19 | 修订为单入口邮箱验证码登录 + 首次资料补全：TopBar 和认证页只保留 `auth_login`，旧 `auth_register` 仅作为 legacy alias 折回登录；新邮箱 verify 后进入 `auth_profile_setup`，`/me.profileCompletionRequired` 未完成时在刷新、重开、换浏览器、退出重登和 deep link 前强制优先补全 displayName + 条款。 | 001-app-shell-auth-settings Phase 9 |
| 2026-05-27 | 1.18 | 修订 auth 为 6 位 email-code：`auth_verify` 只承接手动 code 输入，注册传 `purpose=signup` + `displayName`，登录传 `purpose=login` 且复用同一唯一邮箱，TopBar fallback 删除 `刘哲` / `Liu Zhe` / `liuzhe@example.com`。 | 001-app-shell-auth-settings Phase 8 |
| 2026-05-27 | 1.17 | 修订 auth magic-link callback：`auth_verify` 可短生命周期承接邮件 `token` query，前端自动调用 generated `verifyAuthEmailChallenge`，成功后 replace 到 pending route 或 Home 并清理 URL token；其他 route / pendingAction / storage 仍禁止保存 auth secret。 | 001-app-shell-auth-settings Phase 7 |
| 2026-05-18 | 1.16 | 关闭 URL-addressable routing 004 计划：落地 `frontend/src/app/routeUrl.ts` 类型化 codec + safe-param allowlist、`frontend/src/app/routeStore.ts` Browser History store（push/replace/popstate）、`pendingAction` canonical 过滤、`scripts/spaFallback.mjs` SPA host fallback 与 `vite.config.ts` `appType: "spa"`；新增 E2E.P0.088 / 089 / 090 三场景（in-process vitest jsdom）覆盖 C-11/C-12/C-13；retired aliases 与独立 voice 路由通过负向 grep 阻断。 | 004-url-addressable-routing |
| 2026-05-18 | 1.15 | 锁定 URL-addressable routing 方向：保持 SPA，但正式导航升级为 Browser History canonical path + query，内部继续使用 `Route` / `LooseRoute` 合约；`#route=` 保留为 static preview / pixel parity adapter；新增 D-10 与 C-11 / C-12 / C-13，固化 deep-link / reload / back-forward、auth pendingAction 隐私红线、hash compatibility 与旧 route / 独立 `voice` 负向回归。 | 004-url-addressable-routing |
| 2026-05-11 | 1.14 | 修订 Phase 6 登录态与真实浏览器 parity：authenticated user menu 进入 `test:pixel-parity`，完整 gate 扩展为 8 spec / 112 tests；desktop dropdown 与头像 chip 右对齐，mobile dropdown 必须保持在 viewport 内，logout flow 回到非登录态；001 计划补齐 Phase 6 operation matrix。 | 001-app-shell-auth-settings / 003-ui-design-pixel-parity-gate |
| 2026-05-10 | 1.13 | 修订真实浏览器 pixel parity gate：完整 `test:pixel-parity` 扩展为 8 spec / 110 tests，常规 clean checkout gate 不得依赖 ignored screenshot baseline；workspace full-state pixel tests 必须通过 server-bound route params 进入完整规划态，不依赖 Home recent card 的 `resume-unbound` synthetic path。 | 003-ui-design-pixel-parity-gate |
| 2026-05-10 | 1.12 | 修订 dev mock session 与用户菜单验收：默认 dev mock 必须从非登录态开始，mock verify/logout 必须驱动 `/me` authenticated/unauthenticated 状态切换；TopBar 已登录态必须源级复刻 `ui-design/src/app.jsx` 的头像 chip + dropdown 菜单，并新增 C-10 退出闭环验收。 | 001-app-shell-auth-settings |
| 2026-05-08 | 1.11 | 修订 TopBar brand 文案：品牌区只保留 `E` mark + `EasyInterview`，不再常驻 `面试训练器` 或版本号；`v1.0` 作为产品元数据迁入 settings 的 `产品信息 / Product info` 区，标签走 i18n。 | 002-app-shell-visual-system / 003-ui-design-pixel-parity-gate |
| 2026-05-08 | 1.10 | 按多语言扩展要求修订 TopBar 语言控件：从二选一 icon toggle 改为 icon dropdown，按钮显示当前语言标签，正式前端从 locale catalog 渲染 `topbar-lang-option-{locale}`；locale 优先级固化为用户显式选择（`localStorage["ei-lang"]`）> 浏览器 locale > English fallback，后续新增语言不改变 TopBar 控件结构。 | 002-app-shell-visual-system / 003-ui-design-pixel-parity-gate |
| 2026-05-08 | 1.9 | 修正 TopBar 显示控制契约：当时语言切换按 `ui-design/src/app.jsx` 复刻为 icon toggle（已由 1.10 替换为 language dropdown），主题控制按源码 menu + Custom row 内嵌 AccentPicker 复刻；旧 native select/dropdown 口径仅作为负向回流断言保留。 | 002-app-shell-visual-system |
| 2026-05-08 | 1.8 | 派生 D2 follow-up 003 plan：新增 §6 C-9 真实浏览器 pixel parity gate，要求 Playwright + chromium 在 desktop / mobile viewport 下断言 `frontend/dist/index.html` 与 `ui-design/index.html` golden preview 的 DOM 锚点 / computed style / bounding box / 截图差异；E2E.P0.005（jsdom）保留作为 fast smoke。 | 003-ui-design-pixel-parity-gate |
| 2026-05-07 | 1.7 | 删除废弃外部设计来源；前端视觉实施只以 `ui-design/` 与 `docs/ui-design/` 为唯一 UI 真理源，要求正式前端 100% 源级复刻静态原型并通过 parity gate 验证。 | 002-app-shell-visual-system |
| 2026-05-07 | 1.6 | 修订 D2 视觉系统接入门禁：确认 `ui-design/` 是验收真理源头，`customAccent` 必须进入正式前端主题系统，并新增 visual smoke 工具作为用户可见视觉渲染 gate。 | 002-app-shell-visual-system |
| 2026-05-07 | 1.5 | 派生 D2 视觉系统接入计划；新增 §6 C-8 视觉接入验收，将 `ui-design/` 真理源、4 主题 × 2 模式 wiring、字体与 D1 regression 固化为视觉接入门禁。 | 002-app-shell-visual-system |
| 2026-05-07 | 1.4 | 修订 i18n 初始语言规则：默认跟随浏览器 locale，未知时 fallback English；语言切换只关联前端显示偏好，不再由 runtime config 或登录态覆盖。 | 001-app-shell-auth-settings |
| 2026-05-07 | 1.3 | 收紧 i18n 架构：每种语言必须独立 locale 文件，TopBar 语言切换必须有可访问控件契约，聚合 helper 不得糅合多语言 message map。该控件结构已在 1.10 按当前 UI 真理源更新为 language dropdown。 | 001-app-shell-auth-settings |
| 2026-05-07 | 1.2 | 原地补齐 D1 i18n 决策：正式前端语言切换必须驱动 `zh` / `en` 静态文案、runtime locale 初始化和 `Accept-Language` display hint。 | 001-app-shell-auth-settings |
| 2026-05-05 | 1.0 | 从 engineering-roadmap S1 派生 frontend shell subject，锁定 App 壳、TopBar、auth pendingAction、用户菜单与设置入口。 | 001-app-shell-auth-settings |
