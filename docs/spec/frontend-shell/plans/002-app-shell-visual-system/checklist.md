# App Shell Visual System Checklist

> **版本**: 1.7
> **状态**: completed
> **更新日期**: 2026-07-09

**关联计划**: [plan](./plan.md)

## Phase 1: Token / theme

- [x] 1.1 建立语义 token、4 个 theme/mode 色板组合和 source-to-target 追溯测试；验证: `tokens.test.ts` 断言 token 只暴露语义键、CSS variable 非空、色板值来自 `ui-design/src/primitives.jsx`，且 active theme key 只包含 `ocean` / `plum`。
- [x] 1.2 接入 `customAccent` helper 与根级 display wiring；验证: `DisplayPreferencesRootWiring.test.tsx` 断言 `data-theme`、`data-mode`、`data-custom-accent` 和 accent variables 在切换后即时更新。
- [x] 1.3 保持 Vite CSS 管线；验证: `globalCss.test.ts` 断言 `main.tsx` 单次导入 `global.css`，且 `frontend/package.json` 不引入 Tailwind / CSS-in-JS 依赖。

## Phase 2: Fonts / typography

- [x] 2.1 使用可随仓库交付的开源字体；验证: `fonts.test.ts` 断言 fontsource 依赖、fallback 链和 private-font negative gate。
- [x] 2.2 暴露 `ei-text-*` typography scale；验证: `typography.test.tsx` 断言语义 className、字体族、字号、行高和 TopBar inline-px contract。

## Phase 3: TopBar

- [x] 3.1 TopBar shell 源级迁移；验证: `TopBarVisual.test.tsx` 断言 `ei-shell-topbar`、header 高度、padding、gap、sticky 层级、brand mark 和 nav structure。
- [x] 3.2 三入口与显示控制迁移；验证: `TopBar.test.tsx` 与 `TopBarVisual.test.tsx` 断言 `topbar-nav-home` / `workspace` / `resume_versions`、theme menu、custom accent row、dark toggle、language dropdown、login/user menu、i18n 和 aria behavior。

## Phase 4: Auth / Settings / screen shell

- [x] 4.1 Auth shell 迁移；验证: `AuthVisual.test.tsx` 断言 `auth_login` / `auth_verify` / `auth_profile_setup` / `auth_logout` 的 card shell、form、CTA、status 和 D1 testid regression。
- [x] 4.2 Settings 与通用 screen shell 迁移；验证: visual smoke 断言 `route-settings`、`ei-screen-shell`、`ei-screen-card`、skeleton 和 placeholder shell anchor。

## Phase 5: BDD / handoff

- [x] 5.1 BDD-Gate: `E2E.P0.005` App shell visual smoke；验证: 场景脚本覆盖 default shell、TopBar controls、auth/settings/screen shell、computed variables、`customAccent` overlay 和 route alias negative checks。
  <!-- verified: 2026-07-07 method=focused-frontend-and-scenario evidence="validate_context.py frontend-shell/002 frontend PASS; targeted owner wording grep returned no matches; pnpm --filter @easyinterview/frontend test src/app/theme/tokens.test.ts src/app/display/DisplayPreferencesRootWiring.test.tsx src/app/theme/globalCss.test.ts src/app/theme/fonts.test.ts src/app/theme/typography.test.tsx src/app/topbar/TopBarVisual.test.tsx src/app/topbar/TopBar.test.tsx src/app/auth/AuthVisual.test.tsx src/app/scenarios/p0-005-app-shell-visual-system-smoke.test.tsx PASS (9 files, 91 tests); pnpm --filter @easyinterview/frontend build PASS with existing chunk-size warning; E2E.P0.005 setup/trigger/verify/cleanup PASS (7 tests, existing C.UTF-8 locale warning only)" -->
- [x] 5.2 Handoff docs；验证: `frontend/README.md` 记录 token、display wiring、fonts、className 接入点、visual smoke 与 browser pixel parity gate 分工。

## Phase 6: theme option pruning

- [x] 6.1 UI truth source and formal frontend expose only `ocean` / `plum` plus custom accent in the TopBar theme menu（验证：`ui-design/src/primitives.jsx`, `ui-design/src/app.jsx`, `frontend/src/app/topbar/TopBar.tsx`, `TopBar.test.tsx`）
- [x] 6.2 Active palette types and CSS remove `warm` / `forest` from the supported theme matrix（验证：`tokens.test.ts`, `DisplayPreferencesProvider.test.tsx`, `DisplayPreferencesRootWiring.test.tsx`）
- [x] 6.3 BDD-Gate: `E2E.P0.005` rejects `topbar-theme-option-warm` and `topbar-theme-option-forest` while preserving custom accent（验证：`p0-005-app-shell-visual-system-smoke.test.tsx`）
