# App Shell Visual System

> **版本**: 1.7
> **状态**: completed
> **更新日期**: 2026-07-09

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)
**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)
**关联 BDD Checklist**: [bdd-checklist](./bdd-checklist.md)

## 1 目标

本 owner 固化正式前端 App shell 的视觉系统合同：把 `ui-design/` 当前原型中的 token、主题、暗色、`customAccent`、字体、TopBar、认证页、设置页与通用 screen shell 原生迁移到 `frontend/`，并保持路由、认证、pendingAction、i18n 和 mock runtime 行为不变。

正式前端的可见样式只能来自 `ui-design/src/primitives.jsx`、`ui-design/src/app.jsx`、对应 `ui-design/src/screen-*.jsx` 和 `docs/ui-design/`。本 owner 不创建业务页面细节、不扩大 route catalog、不引入外部设计系统、不复制 prototype data。

## 2 当前真理源

| 范围 | 当前来源 | 正式前端落点 |
|------|----------|--------------|
| 色板与字体预设 | `ui-design/src/primitives.jsx` 的 `EI_THEMES`、`EI_THEME_LIST`、`EI_FONT_PRESETS` | `frontend/src/app/theme/themes.data.ts`、`themes.css`、`fonts.css`、`typography.css` |
| `customAccent` | `ui-design/src/app.jsx` 的 hue / chroma 模型与 `TWEAK_DEFAULTS.theme = "ocean"` | `frontend/src/app/theme/customAccent.ts`、`DisplayPreferencesProvider` |
| TopBar | `ui-design/src/app.jsx::TopBar` | `frontend/src/app/topbar/TopBar.tsx`、`topbar.css` |
| Auth shell | `ui-design/src/screen-auth.jsx` 与 `docs/ui-design/auth-and-entry.md` | `frontend/src/app/auth/*`、`auth.css` |
| Settings / screen shell | `docs/ui-design/module-map.md`、`ui-architecture.md` 和 current screen primitives | `frontend/src/app/screens/*`、`screens.css` |
| UI contract | `docs/spec/frontend-shell/spec.md` v1.25、`frontend/README.md` | focused Vitest、E2E.P0.005、D2 handoff docs |

## 3 质量门禁

- **Plan 类型**: `feature-behavior` + `frontend`
- **TDD 策略**: 通过 `/implement frontend-shell/002-app-shell-visual-system frontend` 进入 `/tdd`。每个可见视觉 surface 必须先有 focused token、component、structural 或 visual-smoke 断言，再写实现。
- **BDD 策略**: 需要 BDD。`E2E.P0.005` 是本 owner 的用户可见 visual-smoke gate，覆盖 DOM 锚点、className、根级 CSS variable、`customAccent` inline overlay、auth/settings/screen shell 和 route alias negative checks。
- **契约边界**: browser-level pixel parity 由 `frontend-shell/003-ui-design-pixel-parity-gate` 承接；本 owner 保持 jsdom fast smoke 与 source-to-target 映射，不用截图基线替代 source-level parity。

## 4 当前合同

### 4.1 Token / theme

`frontend/src/app/theme/` 以 CSS variables 表达语义 token。TS 侧只导出语义键名，不能导出 hex 字面量；`themes.css` 必须覆盖 ocean / plum × light / dark 4 个组合；默认与 fallback 均为 `ocean`。TopBar 主题菜单只暴露 `深海 / Ocean`、`梅子 / Plum` 和 `自定义 / Custom`，不得再展示 `暖陶 / Warm` 或 `苔林 / Forest`。

### 4.2 Display preferences

`DisplayPreferencesProvider` 负责把 `theme`、`dark`、`customAccent` 写入 `<html>` 的 `data-theme`、`data-mode`、`data-custom-accent` 和必要 inline variables。`customAccent` 只覆盖 accent / accent-soft，不改基础色板。

### 4.3 Fonts / typography

字体通过 fontsource 或仓库可交付静态资产进入 Vite bundle。Typography scale 暴露为 `ei-text-*` 语义 className；App shell、TopBar、auth 和 settings 不使用组件内 ad hoc px 排版值表达当前视觉层级。

### 4.4 TopBar

TopBar 保持三入口 nav、主题菜单、custom accent row、暗色 icon toggle、语言 dropdown、登录入口和用户菜单。`data-testid`、`aria-current`、`aria-pressed`、i18n 文案与 route behavior 必须保持稳定；显示控件的 DOM 构图、间距、圆角、字号和状态来自 `ui-design/src/app.jsx`。

### 4.5 Auth / Settings / Screen shell

`auth_login`、`auth_verify`、`auth_profile_setup`、`auth_logout` 使用统一 auth card shell；`settings` 使用账号、隐私、字体和产品信息分区；通用 screen shell 使用 `ei-screen-shell` / `ei-screen-card` 节奏，供业务 owner 在同一视觉骨架内替换内容。

### 4.6 Visual smoke / handoff

`src/app/scenarios/p0-005-app-shell-visual-system-smoke.test.tsx` 与 `test/scenarios/e2e/p0-005-app-shell-visual-system-smoke/` 共同作为 fast smoke gate。`frontend/README.md` 记录 token、display wiring、font、shell className 和 parity gate 重跑方式，业务 owner 必须在这些接入点内扩展页面。

## 5 验收标准

- Token/theme tests 证明 CSS variable 与 `ui-design` 源值可追溯，且源码不引入 Tailwind、CSS-in-JS 或私有字体。
- Display wiring tests 证明 theme / dark / `customAccent` 切换即时更新根级属性和 computed variable。
- TopBar/Auth/Settings visual tests 证明 DOM 锚点、className、testid、i18n 和可访问性行为与当前 App shell 合同一致。
- `E2E.P0.005` visual-smoke 场景通过，且 unsupported route aliases 不 materialize standalone screens。
- `pnpm --filter @easyinterview/frontend build` 通过。
- `frontend/README.md` 保留当前视觉系统接入点和业务 owner 扩展规则。

## 6 验证入口

```bash
python3 .agent-skills/implement/shared/scripts/validate_context.py --context docs/spec/frontend-shell/plans/002-app-shell-visual-system/context.yaml --target frontend
pnpm --filter @easyinterview/frontend test src/app/theme/tokens.test.ts src/app/display/DisplayPreferencesRootWiring.test.tsx src/app/theme/globalCss.test.ts src/app/theme/fonts.test.ts src/app/theme/typography.test.tsx src/app/topbar/TopBarVisual.test.tsx src/app/topbar/TopBar.test.tsx src/app/auth/AuthVisual.test.tsx src/app/scenarios/p0-005-app-shell-visual-system-smoke.test.tsx
pnpm --filter @easyinterview/frontend build
./test/scenarios/e2e/p0-005-app-shell-visual-system-smoke/scripts/setup.sh
./test/scenarios/e2e/p0-005-app-shell-visual-system-smoke/scripts/trigger.sh
./test/scenarios/e2e/p0-005-app-shell-visual-system-smoke/scripts/verify.sh
./test/scenarios/e2e/p0-005-app-shell-visual-system-smoke/scripts/cleanup.sh
```

## 7 修订记录

| 日期 | 版本 | 说明 |
|------|------|------|
| 2026-07-09 | 1.7 | 收敛可选主题为 deep ocean / plum / custom accent，移除 warm / forest active palette、TopBar option、locale key 和 theme matrix 口径。 |
| 2026-07-07 | 1.6 | Compress owner docs to the current ui-design-native visual system contract and executable gates. |
