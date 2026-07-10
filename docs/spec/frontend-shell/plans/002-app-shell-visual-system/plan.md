# App Shell Visual System

> **版本**: 3.0
> **状态**: completed
> **更新日期**: 2026-07-10

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

### 4.7 Visual-smoke test lifecycle isolation

P0.005 中只做同步 DOM/source negative 断言的 default shell、auth login 和 out-of-scope 用例在断言后显式 unmount。这样可在 fixture-backed runtime Promise 回写前完成 effect cleanup，删除无关 `AppRuntimeProvider` / `HomeScreen` act warnings；交互用例、生产 App、样式和视觉合同不变。

### 4.8 Noto Serif SC bundle deduplication

`fonts.css` 只导入 Noto Serif SC 400/500 默认 unicode-range 分片；这些分片已覆盖简体中文和 Latin glyph。不得同时导入无 unicode-range 的 `chinese-simplified-400.css` / `chinese-simplified-500.css` 完整字体，避免 Vite 将同权重中文 glyph 重复打包；字体 family、字重、fallback 和可见排版不变。

### 4.9 Western font subset pruning

产品 locale 只支持 `zh` / `en`。Inter、Source Serif Pro、Cormorant Garamond、IBM Plex Sans 与 JetBrains Mono 只导入各自 Latin 400/500/600 current weights，不打包默认 CSS 中额外的 Greek、Cyrillic、Vietnamese 等子集；Geist 默认 CSS 已等同 Latin bundle，保持原 import。所有 `EI_FONT_PRESETS` family、weight 和 fallback 合同不变。

### 4.10 P0.005 scenario contract reconciliation

`E2E.P0.005` 场景资产必须与 executable smoke 保持同一当前合同：默认 `ocean/light`，切换到 ocean/dark 与 plum/dark 时校验当前 token，Warm / Forest 仅作为菜单不存在的负向断言。jsdom fast smoke 只拥有 DOM/className/computed-variable/custom-accent/route-negative 覆盖；真实浏览器 viewport、geometry 与 screenshot buffer 由当前 `E2E.P0.006` 承接，P0.005 不维护浏览器安装或截图文件生命周期。

### 4.11 Theme data ownership wording

`themes.data.ts` contains four palette combinations, three serif/sans font preset pairs and two TopBar theme metadata entries. Palette/font data are consumed by source-traceability tests; `THEME_METADATA` is consumed by TopBar. There is no CSS generator in this owner: `themes.css` is the checked-in runtime source verified against the data and `ui-design` truth source.

### 4.12 Primitive export surface pruning

`ui-design/src/primitives.jsx` 只定义并导出当前静态原型实际消费的 primitive。跨文件 AST 引用图证明 `Sparkline` 与 `KV` 只有声明和 `window` export、没有 JSX/helper consumer，因此直接删除组件与导出项，不保留兼容别名或退役标记。BDD 不适用，因为两个 primitive 没有可执行用户路径；替代 gate 为 UI contract RED/GREEN、AST/export inventory、P0.005、visual-system/full frontend tests、owner contexts 与 docs/diff/pruning gates。

### 4.13 Design canvas consumer-surface pruning

`ui-design/design-canvas.jsx` 的组件参数面只承接仓库内唯一消费者 `ui-design/canvas.html` 当前传入的属性；删除未被任何画板实例传入的缩放、样式、间距和定位扩展参数。`canvas.html` 的 iframe helper 只拼装会改变当前画板输出的 route、language、session、mode 和 display 参数，默认面试上下文由 `src/app.jsx::createInterviewContext` 单点提供，不重复声明相同常量。BDD 不适用，因为本批不改变画板数量、路由、尺寸、主题、字体或交互；替代 gate 为 UI contract RED/GREEN、AST consumer inventory、静态浏览器画板 smoke、P0.005、owner/full frontend 与 docs/diff/pruning gates。

### 4.14 Design canvas unavailable sidecar removal

设计画板只保留当前页面会话内的 React state，用于重排、重命名和聚焦；删除仓库内没有 sidecar 实体、生成入口或 host 实现的 `.design-canvas.state.json` fetch/write bridge。仓库当前 `ui-design/run.sh` 静态服务器不得为不可达的跨刷新持久化请求制造 404，也不通过延迟 gate 阻塞首次画板渲染。BDD 不适用，因为当前可执行环境从未提供该持久化能力；替代 gate 为 source RED/GREEN、零引用 inventory、静态服务器请求日志、浏览器重排/重命名/聚焦 smoke、UI/P0.005/full frontend 与 docs/diff/pruning gates。

### 4.15 Unavailable prototype edit-mode bridge removal

静态原型只保留当前页面可达的 TopBar 主题/暗色/语言/custom accent 与 Settings 字体预设控制。删除仓库无 host、无 listener、无文档入口的 `__edit_mode_*` message bridge，以及只能由该 bridge 打开的 `TweaksPanel`、专用 helper 和独占的 `role` tweak 通道；当前 display state 继续由 React state、hash 参数和可见控件直接驱动。BDD 不适用，因为 edit-mode 面板在仓库执行路径中不可达；替代 gate 为 source RED/GREEN、AST/message inventory、静态浏览器显示控制 smoke、UI/P0.005/full frontend 与 docs/diff/pruning gates。

### 4.16 Zero-read canvas mode binding removal

删除 edit-mode bridge 清理后留在 `ui-design/src/app.jsx`、没有任何读取点的 `isCanvasIframe` 局部变量；保留 `hideTopBar` 对 `data-nochrome` 的当前判断以及 canvas iframe 的无 TopBar 行为。跨文件 TypeScript reference inventory 中，位置上不可省略的 callback 首参只作为 positional placeholder 记录，不通过重写迭代结构制造替代代码。BDD 不适用，因为零读取 binding 不影响可执行页面；替代 gate 为 UI contract RED/GREEN、AST inventory、P0.005、visual-system/full frontend、owner contexts 与 docs/diff/pruning gates。

### 4.17 Zero-consumer formal CSS pruning

正式前端 CSS 只保留当前 DOM 消费的视觉规则。删除没有 TSX、原型或场景消费者的 `ei-screen-card-grid`、`visually-hidden` 与 `ei-topbar-theme-swatch--custom-active`，并删除 README 对 `ei-screen-card-grid` 的虚假接入说明；保留原型 Practice 正在使用且有 parity gate 的 `ei-scroll`。BDD 不适用，因为三个选择器均不可达；替代 gate 为三个 owner source RED/GREEN、class inventory、P0.005、visual-system/full frontend、build、owner contexts 与 docs/diff/pruning gates。

### 4.18 TopBar login rule consolidation

正式 TopBar stylesheet 中相邻的两段 `ei-topbar-auth-login` 同 selector 声明合并为一个最终计算值等价的规则，完整保留尺寸、间距、字体、交互、背景、边框、颜色及独立 hover state，不改变按钮 DOM。BDD 不适用，因为本批不改变用户可见行为；替代 gate 为 TopBar source RED/GREEN、declaration inventory、visual-system/full frontend、typecheck/build、owner contexts 与 docs/diff/pruning gates。

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
| 2026-07-10 | 3.0 | Consolidate the TopBar login declarations into one equivalent rule. |
| 2026-07-10 | 2.9 | Delete three zero-consumer formal CSS selectors and the stale screen-grid handoff entry. |
| 2026-07-10 | 2.8 | Remove the zero-read canvas iframe mode binding while preserving no-chrome TopBar behavior. |
| 2026-07-10 | 2.7 | Remove the unavailable prototype edit-mode bridge, exclusive panel and dead role tweak channel. |
| 2026-07-10 | 2.6 | Remove the unavailable design-canvas sidecar bridge while retaining in-memory editing. |
| 2026-07-10 | 2.5 | Prune unused design-canvas component parameters and duplicate iframe context defaults. |
| 2026-07-10 | 2.4 | Remove two zero-consumer prototype primitive globals and reconcile plan/checklist/INDEX metadata. |
| 2026-07-10 | 2.2 | Correct theme data consumer and font preset documentation. |
| 2026-07-10 | 2.1 | Align P0.005 scenario assets with the current ocean/plum fast-smoke contract and the existing P0.006 browser boundary. |
| 2026-07-10 | 2.0 | Restrict Western font presets to the product's Latin locale subset while retaining every current family and weight. |
| 2026-07-10 | 1.9 | Remove duplicate full Noto Serif SC imports while retaining the unicode-range 400/500 bundles and visible typography contract. |
| 2026-07-10 | 1.8 | Isolate synchronous P0.005 visual-smoke assertions from unrelated runtime effects. |
| 2026-07-09 | 1.7 | 收敛可选主题为 deep ocean / plum / custom accent，移除 warm / forest active palette、TopBar option、locale key 和 theme matrix 口径。 |
| 2026-07-07 | 1.6 | Compress owner docs to the current ui-design-native visual system contract and executable gates. |
