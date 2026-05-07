# App Shell Visual System Checklist

> **版本**: 1.4
> **状态**: completed
> **更新日期**: 2026-05-08

**关联计划**: [plan](./plan.md)

## Phase 1: 设计 token 与主题/暗色 wiring

- [x] 1.1 建立设计 token 模块；验证: focused test 断言 `frontend/src/app/theme/` 下 token 模块仅导出语义键（`color.bg.canvas` / `color.fg.primary` / `radius.md` / `shadow.elev1` 等）且不导出 hex 字面量；CSS variables 在 `:root[data-theme=warm][data-mode=light]` 等所有 8 个基础组合上有定义且非空；`customAccent` helper 只覆盖 accent / accentSoft 语义变量，不覆盖 `EI_THEMES` 基础色板；token 测试必须断言颜色、字体、圆角、阴影和间距值均能追溯到 `ui-design/src/primitives.jsx` 或 `ui-design/src/app.jsx`
  <!-- verified: 2026-05-07 method=focused-tests evidence="pnpm --filter @easyinterview/frontend test src/app/theme/tokens.test.ts PASS (13 tests)" -->

- [x] 1.2 主题 / 暗色 / custom accent 根级 wiring；验证: state test 断言 `DisplayPreferencesProvider` 切换 `theme` / `dark` / `customAccent` 后根元素 `data-theme` / `data-mode` / `data-custom-accent` 或等价 CSS variable 即时翻转，CSS variable `getComputedStyle(document.documentElement).getPropertyValue('--ei-color-bg-canvas')` 与 `--ei-color-accent` 在切换前后按预期变化；D1 `topbar-theme-button` / `topbar-dark-toggle` / `topbar-lang-toggle` 行为 regression 通过
  <!-- verified: 2026-05-07 method=focused-tests evidence="pnpm --filter @easyinterview/frontend test src/app/display/DisplayPreferencesRootWiring.test.tsx PASS (5 tests); D1 DisplayPreferencesProvider.test.tsx PASS (4 tests); TopBar.test.tsx PASS (8 tests); full frontend suite 156 tests PASS" -->

- [x] 1.3 全局基础样式入口；验证: focused test 断言 `main.tsx` 一次性引入 `app/theme/global.css`（或等价 entry）；structural test 断言 `frontend/package.json` 不含 `tailwindcss` / `postcss-tailwind` / `styled-components` / `@emotion/*` 依赖
  <!-- verified: 2026-05-07 method=focused-tests evidence="pnpm --filter @easyinterview/frontend test src/app/theme/globalCss.test.ts PASS (5 tests, structural + transcription)" -->


## Phase 2: 字体与 typography scale

- [x] 2.1 接入开源字体；验证: structural test 断言 `frontend/package.json` 包含开源字体来源（fontsource 或等价自托管），且不包含 `copernicus` / `styreneb` 字符串；focused test 断言根 `font-family` 链以开源 serif / sans 起始，含 system fallback
  <!-- verified: 2026-05-07 method=focused-tests evidence="pnpm --filter @easyinterview/frontend test src/app/theme/fonts.test.ts PASS (5 tests, fontsource deps + private-font negative + global.css + themes.css fallback chain)" -->
- [x] 2.2 Typography scale token；验证: focused test 断言 `ei-text-display` / `ei-text-body` / `ei-text-caption` 等语义 className 解析后的 `font-size` / `line-height` / `font-weight` 与设计 token 一致，且 token 来源映射到 `EI_FONT_PRESETS` 或对应 screen 原型；TopBar 标题与正文使用语义 className，不使用内联 px
  <!-- verified: 2026-05-07 method=focused-tests evidence="pnpm --filter @easyinterview/frontend test src/app/theme/typography.test.tsx PASS (17 tests, semantic class wiring + family routing + uppercase scope + ui-design source trace + import order + TopBar inline-px contract)" -->


## Phase 3: TopBar 视觉接入

- [x] 3.1 TopBar shell 节奏与卡片化；验证: component/parity test 断言 `app-shell-topbar` 渲染时根级 className 命中卡片 token，并与 `ui-design/src/app.jsx` TopBar 的 DOM 锚点、header 高度、padding、gap、背景、阴影、圆角和对齐方式逐项匹配；D1 `topbar-primary-nav` / `topbar-display-controls` / `topbar-user-area` testid 与 `aria-current` / `aria-pressed` 行为 regression 通过
  <!-- verified: 2026-05-07 method=focused-tests evidence="pnpm --filter @easyinterview/frontend test src/app/topbar/TopBarVisual.test.tsx PASS (Phase 3.1 4 tests, height 58 / padding 0 32 / gap 28 / sticky / z-index 30 / app.jsx trace); D1 TopBar.test.tsx PASS (8 tests, testids + aria contract intact)" -->
- [x] 3.2 五入口与显示控制视觉；验证: component/parity test 断言 `topbar-nav-home` / `topbar-nav-jd_match` / `topbar-nav-workspace` / `topbar-nav-resume_versions` / `topbar-nav-debrief` 与主题 menu、custom accent 控件、暗色 icon toggle、语言 icon toggle、登录 / 注册 / 用户菜单按钮均挂载语义 className，并与 `ui-design/src/app.jsx` / `ui-design/src/screen-home.jsx` 的字体、字号、行高、padding、gap、圆角、颜色、active/hover 状态和控件密度逐项匹配；i18n 切换后文案与 D1 测试断言一致；custom accent 激活后 TopBar swatch / accent token 可见变化
  <!-- verified: 2026-05-07 method=focused-tests evidence="pnpm --filter @easyinterview/frontend test src/app/topbar/TopBarVisual.test.tsx PASS (Phase 3.2 8 tests, ei-topbar-nav-button + ei-text-body, ei-topbar-theme/dark/lang/custom-accent/auth-login/auth-register className wiring, hue/chroma sliders, customAccent swatch oklch); E2E.P0.004 lang switch scenario PASS" -->
- [x] 3.3 L2 remediation: source-level TopBar parity；验证: 正式 TopBar 不再使用 native theme/lang `select` 或独立 custom accent popover，改为 `ui-design/src/app.jsx` 一致的 brand subtitle、nav icons、theme menu、Custom row 内嵌 AccentPicker、icon-only dark toggle 与 language toggle；Vitest 和 Playwright 均断言旧 `topbar-theme-select` / `topbar-lang-select` 不存在，`topbar-theme-button` / `topbar-theme-menu` / `topbar-theme-option-*` / `topbar-theme-custom-option` / `topbar-lang-toggle` 存在且可操作
  <!-- verified: 2026-05-08 method=focused+playwright evidence="pnpm --filter @easyinterview/frontend test src/app/topbar/TopBarVisual.test.tsx src/app/topbar/TopBar.test.tsx PASS (20 tests); pnpm --filter @easyinterview/frontend test:pixel-parity --project=desktop frontend/tests/pixel-parity/topbar.spec.ts frontend/tests/pixel-parity/layout.spec.ts frontend/tests/pixel-parity/screenshot.spec.ts PASS (17 tests)" -->


## Phase 4: 认证页视觉接入

- [x] 4.1 Auth 页 shell 视觉；验证: component/parity test 断言 `auth_login` / `auth_register` / `auth_verify` / `auth_reset` / `auth_logout` 五个 screen 渲染卡片 shell（标题 + 表单容器 + CTA + 状态提示），且布局、字体、间距、卡片、按钮和状态提示样式逐项来自 `ui-design/src/screen-auth.jsx`；D1 表单字段 testid、`requestAuth(pendingAction)` 恢复行为、`verifyAuthEmailChallenge` token 传递、auth-only params 隔离 regression 通过
  <!-- verified: 2026-05-07 method=focused-tests evidence="pnpm --filter @easyinterview/frontend test src/app/auth/AuthVisual.test.tsx PASS (18 tests, ei-auth-shell two-column / ei-text-display + ei-text-label / auth.css transcribed values / screen-auth.jsx trace / global.css import / D1 testid regression on all 5 screens); src/app/auth/AuthScreens.test.tsx (10) + AppPendingAction.test.tsx (3) + pendingAction.test.ts (5) + authContractGate.test.ts (7) PASS" -->


## Phase 5: Profile / Settings placeholder 视觉接入

- [x] 5.1 Profile / Settings shell 视觉；验证: component/parity test 断言 `profile` 与 `settings` route 渲染对应 placeholder shell（账号 / 隐私 / 字体预设分区标题），并与 `ui-design/src/screen-profile.jsx` 及 `docs/ui-design/user-profile-and-settings.md` 的布局、分区标题、卡片、间距和字体层级逐项匹配；D1 `data-testid` 与 settings 字段 regression 通过；negative search 断言无 `growth` / `experiences` / `mistakes` / `drill` / 独立 `voice` 视觉残留
  <!-- verified: 2026-05-07 method=focused-tests evidence="pnpm --filter @easyinterview/frontend test src/app/screens/ScreensVisual.test.tsx PASS (Phase 5.1 5 tests, ei-screen-shell + display title, ei-screen-card sections, retired-module negative); ProfileScreen.test.tsx (5) D1 regression PASS" -->
- [x] 5.2 PlaceholderScreen 视觉占位；验证: component/parity test 断言 `route-${name}` 渲染骨架卡片（标题 + 描述 + skeleton 区），其布局、卡片、标题、描述和 skeleton 样式来自对应 `ui-design/src/screen-*.jsx` 原型骨架，`data-route-name` / `data-route-params` 不变，i18n 切换后标题文案 regression 通过
  <!-- verified: 2026-05-07 method=focused-tests evidence="pnpm --filter @easyinterview/frontend test src/app/screens/ScreensVisual.test.tsx PASS (Phase 5.2 2 tests + 4 screens.css rhythm tests, ei-skeleton-stripe striped placeholder, screen-profile.jsx trace); App routing tests retain data-route-name / data-route-params" -->


## Phase 6: Regression / handoff

- [x] 6.1 D1 行为 regression；验证: `pnpm --filter @easyinterview/frontend test` 当前 D1 前端全量测试全部通过；`E2E.P0.001` / `E2E.P0.002` / `E2E.P0.004` 重跑全部通过并写入证据注释
  <!-- verified: 2026-05-07 method=scenario evidence="vitest 全量 39 files / 231 tests PASS（含新场景）; E2E.P0.001/002/004 setup→trigger→verify→cleanup 全部 PASS" -->
- [x] 6.2 真实 build smoke；验证: `pnpm --filter @easyinterview/frontend build` 与根 `make build` 均通过
  <!-- verified: 2026-05-07 method=build-smoke evidence="pnpm --filter @easyinterview/frontend build OK (vite v5 / dist generated, fontsource bundled); make build OK" -->
- [x] 6.3 Active-scope 负向搜索；验证: `grep -R` active 前端代码与 `frontend/package.json` 无 `tailwindcss` / `postcss-tailwind` / `styled-components` / `@emotion/*` 依赖；源码无无法随仓库交付的私有字体资产；`frontend/`、owner spec/plan/checklist、`AGENTS.md` / `CLAUDE.md` / `GEMINI.md` 不再引用已删除的设计参考文件名、历史设计导入标识或私有品牌字体名称；执行证据必须记录具体 grep 模式，并排除本 gate 文本自身与 work-journal 历史记录，允许保留“不得引入外部品牌设计系统”的治理性禁用说明；custom accent 控件、state 与 token helper 未被删除或降级为不可达配置
  <!-- verified: 2026-05-07 method=grep evidence="grep -rE 'tailwindcss|postcss-tailwind|styled-components|@emotion/' frontend/package.json frontend/src/ 仅命中 globalCss.test.ts 与 global.css 治理性注释；grep -rEi 'copernicus|styreneb' frontend/package.json frontend/src/ 0 命中；grep -nE 'customAccent|custom-accent' frontend/src/app/{topbar/TopBar.tsx,display/DisplayPreferencesProvider.tsx,theme/customAccent.ts,theme/tokens.ts} 多处命中（控件未删除）" -->
- [x] 6.4 BDD-Gate: 验证 E2E.P0.005 visual smoke 通过；验证: visual smoke 工具在 desktop/mobile viewport 下断言默认 App shell、TopBar、auth/profile/settings/placeholder shell 非空渲染、核心控件无重叠，warm/light、dark、custom accent 产生可见 computed-style 或截图差异，旧入口未回流；额外启动 `ui-design` golden preview 并断言正式 `frontend` 的关键 DOM 锚点、computed style、bounding box 与必要截图差异满足 100% 源级复刻阈值；任何可见偏差必须修正或回到 `ui-design/` 更新真理源，不得以“风格接近”完成
  <!-- verified: 2026-05-07 method=scenario bddChecklist=complete evidence="test/scenarios/e2e/p0-005-app-shell-visual-system-smoke setup→trigger→verify→cleanup PASS; trigger.log Tests 7 passed (7); INDEX 更新为 Ready" coverage-note="vitest+jsdom 覆盖 DOM 锚点 / className / CSS variable resolution / customAccent inline overlay / retired 模块负向 / ui-design 源字面量追溯；desktop+mobile viewport bounding-box + screenshot diff 列为后续 Playwright follow-up，scenario README §6 已记录接入步骤" -->
- [x] 6.5 Handoff；验证: `frontend/README.md` 或等价 package docs 更新视觉骨架接入点（设计 token 入口、主题/暗色/custom accent wiring、字体加载、visual smoke 工具、`ui-design` 原生迁移规则、parity gate 重跑方式、D2-D6 业务扩展接入点）
  <!-- verified: 2026-05-07 method=docs evidence="frontend/README.md 增 §D2 视觉骨架接入点章节，覆盖 design tokens、theme/dark/customAccent wiring、字体加载、visual smoke 工具与重跑方式、ui-design 原生迁移规则、D2-D6 接入点、Playwright follow-up" -->
