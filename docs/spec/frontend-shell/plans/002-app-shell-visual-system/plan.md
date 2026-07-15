# App Shell Visual System

> **版本**: 3.1
> **状态**: completed
> **更新日期**: 2026-07-14

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)
**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)
**关联 BDD Checklist**: [bdd-checklist](./bdd-checklist.md)

## 1 目标

本 owner 固化正式前端 App shell 的视觉系统合同：把 `frontend/` 当前原型中的 token、主题、暗色、`customAccent`、字体、TopBar、认证页、设置页与通用 screen shell 原生迁移到 `frontend/`，并保持路由、认证、pendingAction、i18n 和 mock runtime 行为不变。

正式前端的可见样式只能来自 `frontend/src`、对应 `frontend/src` 和 `docs/ui-design/`。本 owner 不创建业务页面细节、不扩大 route catalog、不引入外部设计系统、不复制 prototype data。

## 2 当前真理源

| 范围 | 当前来源 | 正式前端落点 |
|------|----------|--------------|
| 色板与字体预设 | `frontend/src` 的 `EI_THEMES`、`EI_THEME_LIST`、`EI_FONT_PRESETS` | `frontend/src/app/theme/themes.data.ts`、`themes.css`、`fonts.css`、`typography.css` |
| `customAccent` | `frontend/src` 的 hue / chroma 模型与 `TWEAK_DEFAULTS.theme = "ocean"` | `frontend/src/app/theme/customAccent.ts`、`DisplayPreferencesProvider` |
| TopBar | `frontend/src` | `frontend/src/app/topbar/TopBar.tsx`、`topbar.css` |
| Auth shell | `frontend/src` 与 `docs/ui-design/auth-and-entry.md` | `frontend/src/app/auth/*`、`auth.css` |
| Settings / screen shell | `docs/ui-design/module-map.md`、`ui-architecture.md` 和 current screen primitives | `frontend/src/app/screens/*`、`screens.css` |

## 3 质量门禁

- **Plan 类型**: `feature-behavior` + `frontend`
- **TDD 策略**: 通过 `/implement frontend-shell/002-app-shell-visual-system frontend` 进入 `/tdd`。每个可见视觉 surface 必须先有 focused token、component、structural 或 visual-smoke 断言，再写实现。
- **契约边界**: 正式 component、responsive 与 accessibility assertions 由各 frontend owner 承接；`frontend-shell/003-ui-design-responsive-browser-gate` 只负责防止已删除 Demo/parity 工具链回流。本 owner 保持 jsdom fast smoke，不用截图基线替代 formal implementation contract。

## 4 当前合同

### 4.1 Token / theme

`frontend/src/app/theme/` 以 CSS variables 表达语义 token。TS 侧只导出语义键名，不能导出 hex 字面量；`themes.css` 必须覆盖 ocean / plum × light / dark 4 个组合；默认与 fallback 均为 `ocean`。TopBar 主题菜单只暴露 `深海 / Ocean`、`梅子 / Plum` 和 `自定义 / Custom`，不得再展示 `暖陶 / Warm` 或 `苔林 / Forest`。

### 4.2 Display preferences

`DisplayPreferencesProvider` 负责把 `theme`、`dark`、`customAccent` 写入 `<html>` 的 `data-theme`、`data-mode`、`data-custom-accent` 和必要 inline variables。`customAccent` 只覆盖 accent / accent-soft，不改基础色板。

### 4.3 Fonts / typography

字体通过 fontsource 或仓库可交付静态资产进入 Vite bundle。Typography scale 暴露为 `ei-text-*` 语义 className；App shell、TopBar、auth 和 settings 不使用组件内 ad hoc px 排版值表达当前视觉层级。

### 4.4 TopBar

TopBar 保持三入口 nav、主题菜单、最小 custom accent row、暗色 icon toggle、语言 dropdown、登录入口和用户菜单。`CustomAccentPicker` 只保留色相与饱和度控件；不渲染 preview/value 区或“恢复主题默认色 / Reset to theme accent”按钮。选择 Ocean / Plum 是退出 custom accent 的唯一清晰路径。`data-testid`、`aria-current`、`aria-pressed`、i18n 文案与 route behavior 必须保持稳定；显示控件的 DOM 构图、间距、圆角、字号和状态来自 `frontend/src`。

### 4.5 Auth / Settings / Screen shell

`auth_login`、`auth_verify`、`auth_profile_setup`、`auth_logout` 使用统一 auth card shell；`settings` 使用账号、隐私、字体和产品信息分区；通用 screen shell 使用 `ei-screen-shell` / `ei-screen-card` 节奏，供业务 owner 在同一视觉骨架内替换内容。





### 4.8 Noto Serif SC bundle deduplication

`fonts.css` 只导入 Noto Serif SC 400/500 默认 unicode-range 分片；这些分片已覆盖简体中文和 Latin glyph。不得同时导入无 unicode-range 的 `chinese-simplified-400.css` / `chinese-simplified-500.css` 完整字体，避免 Vite 将同权重中文 glyph 重复打包；字体 family、字重、fallback 和可见排版不变。

### 4.9 Western font subset pruning

产品 locale 只支持 `zh` / `en`。Inter、Source Serif Pro、Cormorant Garamond、IBM Plex Sans 与 JetBrains Mono 只导入各自 Latin 400/500/600 current weights，不打包默认 CSS 中额外的 Greek、Cyrillic、Vietnamese 等子集；Geist 默认 CSS 已等同 Latin bundle，保持原 import。所有 `EI_FONT_PRESETS` family、weight 和 fallback 合同不变。



### 4.11 Theme data ownership wording

`themes.data.ts` contains four palette combinations, three serif/sans font preset pairs and two TopBar theme metadata entries. Palette/font data are consumed by source-traceability tests; `THEME_METADATA` is consumed by TopBar. There is no CSS generator in this owner: `themes.css` is the checked-in runtime source verified against the data and `ui-design` truth source.













### 4.18 TopBar login rule consolidation

正式 TopBar stylesheet 中相邻的两段 `ei-topbar-auth-login` 同 selector 声明合并为一个最终计算值等价的规则，完整保留尺寸、间距、字体、交互、背景、边框、颜色及独立 hover state，不改变按钮 DOM。BDD 不适用，因为本批不改变用户可见行为；替代 gate 为 TopBar source RED/GREEN、declaration inventory、visual-system/full frontend、typecheck/build、owner contexts 与 docs/diff/pruning gates。

### 4.19 Minimal custom accent picker

`CustomAccentPicker` 只接收并消费当前色相、饱和度及其更新回调。删除 preview/value 展示区、“恢复主题默认色 / Reset to theme accent”按钮，以及只为这些已删 UI 服务的 `onClear` / `active` props 和调用方传参；不得保留空 wrapper、兼容 props 或隐藏 reset action。


## 5 验收标准

- Token/theme tests 证明 CSS variable 与 `ui-design` 源值可追溯，且源码不引入 Tailwind、CSS-in-JS 或私有字体。
- Display wiring tests 证明 theme / dark / `customAccent` 切换即时更新根级属性和 computed variable。
- Custom accent picker 只显示色相/饱和度；旧 preview/value/reset UI 与 `onClear` / `active` 冗余 props 零引用；选择 Ocean / Plum 能清晰退出 custom accent。
- TopBar/Auth/Settings visual tests 证明 DOM 锚点、className、testid、i18n 和可访问性行为与当前 App shell 合同一致。
- `pnpm --filter @easyinterview/frontend build` 通过。
- `frontend/README.md` 保留当前视觉系统接入点和业务 owner 扩展规则。

## 6 验证入口

```bash
python3 .agent-skills/implement/shared/scripts/validate_context.py --context docs/spec/frontend-shell/plans/002-app-shell-visual-system/context.yaml --target frontend
pnpm --filter @easyinterview/frontend build
```

## 7 修订记录

| 日期 | 版本 | 说明 |
|------|------|------|
| 2026-07-14 | 3.1 | Reopen to reduce CustomAccentPicker to hue/saturation and make Ocean/Plum the only custom-accent exit. |
| 2026-07-10 | 3.0 | Consolidate the TopBar login declarations into one equivalent rule. |
| 2026-07-10 | 2.9 | Delete three zero-consumer formal CSS selectors and the stale screen-grid handoff entry. |
| 2026-07-10 | 2.8 | Remove the zero-read canvas iframe mode binding while preserving no-chrome TopBar behavior. |
| 2026-07-10 | 2.7 | Remove the unavailable prototype edit-mode bridge, exclusive panel and dead role tweak channel. |
| 2026-07-10 | 2.6 | Remove the unavailable design-canvas sidecar bridge while retaining in-memory editing. |
| 2026-07-10 | 2.5 | Prune unused design-canvas component parameters and duplicate iframe context defaults. |
| 2026-07-10 | 2.4 | Remove two zero-consumer prototype primitive globals and reconcile plan/checklist/INDEX metadata. |
| 2026-07-10 | 2.2 | Correct theme data consumer and font preset documentation. |
| 2026-07-10 | 2.0 | Restrict Western font presets to the product's Latin locale subset while retaining every current family and weight. |
| 2026-07-10 | 1.9 | Remove duplicate full Noto Serif SC imports while retaining the unicode-range 400/500 bundles and visible typography contract. |
| 2026-07-09 | 1.7 | 收敛可选主题为 deep ocean / plum / custom accent，移除 warm / forest active palette、TopBar option、locale key 和 theme matrix 口径。 |
| 2026-07-07 | 1.6 | Compress owner docs to the current ui-design-native visual system contract and executable gates. |
