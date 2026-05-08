# App Shell Visual System

> **版本**: 1.4
> **状态**: completed
> **更新日期**: 2026-05-08

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)
**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)
**关联 BDD Checklist**: [bdd-checklist](./bdd-checklist.md)

## 1 目标

承接 [001-app-shell-auth-settings](../001-app-shell-auth-settings/plan.md) 已交付的 routing / i18n / auth / fixture-backed mock 行为骨架，把 `ui-design/` 静态原型 100% 源级复刻进正式前端 App 壳：建立设计 token 与 typography scale、接入主题 / 暗色 / `customAccent` 根级 wiring、为 TopBar / 五入口 / 显示控制 / 认证页 / 用户菜单 / settings & profile placeholder 提供与静态原型一致的视觉表达。本 plan 不引入业务页面细节、不引入新构建框架、不引入外部品牌设计系统，不允许 AI 自由重设计、重新解释或重新组合视觉，并在交付时通过 D1 前端全量测试、`E2E.P0.001` / `E2E.P0.002` / `E2E.P0.004` 行为 regression、`E2E.P0.005` visual smoke + `ui-design` parity gate 证明视觉接入未破坏既有行为契约且忠实复刻当前 UI 真理源。

## 2 背景

D1 plan 完成后，正式前端已具备 `home` 默认壳、五入口 TopBar、`requestAuth(pendingAction)`、用户菜单、`profile` / `settings` placeholder、display preferences (`warm` / `forest` / `ocean` / `plum` × `light` / `dark`)、独立 locale 文件以及 fixture-backed mock transport。`docs/ui-design/` 与 `ui-design/` 是本 plan 的唯一 UI 验收真理源头；外部品牌设计系统不再作为参考。spec.md §4 / §6 C-8 已收紧为：正式前端必须 100% 源级复刻 `ui-design/` 的 DOM 构图、布局、间距、字号、字体层级、控件密度、颜色、阴影、边框、圆角、状态、响应式行为和交互节奏，`customAccent` 必须进入正式前端主题系统，visual smoke 工具必须作为用户可见视觉渲染 gate。

`ui-design/src/primitives.jsx` 持有 `EI_THEMES` 真理源（warm / forest / ocean / plum × light / dark）和 `EI_FONT_PRESETS`；`ui-design/src/app.jsx` 持有 `customAccent` 运行时模型与 TopBar 自定义 accent 控件。`ui-design/src/screen-*.jsx` 与 `docs/ui-design/auth-and-entry.md` / `user-profile-and-settings.md` 等文档定义具体页面节奏、卡片层级与字体层级。实现前必须直接读取这些文件并建立 source-to-target 映射；所有样式值必须来自原型源码或由原型源码抽取的 token，不得只按文字描述、组件库默认值或 AI 直觉重建 UI。

## 3 质量门禁分类

- **Plan 类型**: `feature-behavior` + `frontend`（视觉接入是用户可见的前端交付，但用户行为流不变）。
- **TDD 策略**: 通过 `/implement frontend-shell/002-app-shell-visual-system frontend` → `/tdd` 执行；每个 checklist item 先写 focused Vitest / component test / structural test / visual smoke / parity fixture，再实现最小前端代码；测试断言写在 checklist 的 `验证:` 后。Token wiring、主题 / 暗色 / `customAccent` data-attribute 或 CSS variable 切换、字体加载 helper、TopBar / Auth / Profile / Settings 视觉表达、visual smoke 工具必须各自有断言；关键视觉项必须与 `ui-design` golden preview 比对 DOM 锚点、computed style、bounding box 和必要截图差异，并附 source-to-target 映射证据；D1 当前前端全量测试 + `pnpm --filter @easyinterview/frontend build` + `make build` 必须在 Phase 6 通过。任何可见偏差不得以“风格接近”收口，必须修到与原型一致或先修改 `ui-design/` 真理源。
- **BDD 策略**: 需要 BDD。本 plan 引入用户可见视觉系统、主题 / 暗色 / custom accent 交互和 visual smoke 验证工具，必须维护 [bdd-plan](./bdd-plan.md)、[bdd-checklist](./bdd-checklist.md)，并在主 checklist 使用 `BDD-Gate:` 引用 `E2E.P0.005`。D1 已交付的 `E2E.P0.001` / `E2E.P0.002` / `E2E.P0.004` 作为行为 regression gate 继续重跑。
- **替代验证 gate**: 不适用；BDD gate 是本 plan 的用户可见视觉验证入口。补充 gate 包括 D1 frontend 全量测试、visual contract tests、`ui-design` parity tests、visual smoke 工具、build smoke、active-scope 负向搜索和 `make docs-check`。

## 4 实施步骤

### Phase 1: 设计 token 与主题/暗色 wiring

#### 1.1 建立设计 token 模块

在 `frontend/src/app/theme/` 下建立 design token 入口，覆盖 warm / forest / ocean / plum × light / dark 色板、`customAccent` 覆盖层、阴影、圆角、间距和断点。Token 以 CSS variables 形式注入到根元素，TS 侧只导出语义键名常量供 className 与 testid 使用，**不导出 hex 字面量**。颜色、阴影、圆角、间距、字体预设等值必须从 `ui-design/src/primitives.jsx` 的 `EI_THEMES` / `EI_FONT_PRESETS` 与 `ui-design/src/app.jsx` 的 `customAccent` 计算模型抽取，不允许新增未能追溯到 UI 原型的视觉值。

#### 1.2 主题 / 暗色根级 wiring

由 `DisplayPreferencesProvider` 读取当前 `theme`、`dark` 与 `customAccent`，写入根元素 `data-theme`、`data-mode` 与 `data-custom-accent` 或等价 CSS variable；CSS variables 通过 `[data-theme=warm][data-mode=light]` 等选择器命中正确色板，并在 custom accent 激活时只覆盖 accent / accentSoft 语义变量。切换 TopBar 主题、暗色或自定义 accent 控件后，DOM 根级属性或 CSS variable 即时翻转，颜色 / 阴影 / 边框随之更新。

#### 1.3 全局基础样式入口

新增 `frontend/src/app/theme/global.css`（或等价 entry），由 `main.tsx` 一次性引入；包含 reset、根级排版、focus ring、滚动条等基础样式。**不**引入 Tailwind / PostCSS-Tailwind / CSS-in-JS 框架，仅使用 vite 默认 CSS 管线。

### Phase 2: 字体与 typography scale

#### 2.1 接入开源字体

通过 `fontsource`（或等价自托管方案）引入 `ui-design` 当前字体预设所需的开源 serif / sans 字体；字体文件由 vite 静态资源管线打包，运行时 `font-display: swap`。字体选择必须跟随 `EI_FONT_PRESETS` 或后续 `ui-design` 原型变更，不得引入无法随仓库交付的私有字体资产。

#### 2.2 Typography scale token

定义字号 / 行高 / 字重 / 字间距 token（标题、副标题、正文、辅助、代码），逐项复刻 `ui-design/` 静态原型节奏；token 通过 CSS variables + 语义 className（如 `ei-text-display` / `ei-text-body` / `ei-text-caption`）暴露，并保留来源注释或测试映射证明每个 token 来自原型。

### Phase 3: TopBar 视觉接入

#### 3.1 TopBar shell 节奏与卡片化

把 `frontend/src/app/topbar/TopBar.tsx` 的 `<header>` / `<nav>` / `<div>` 接入设计 token，并按 `ui-design/src/app.jsx` 的 TopBar DOM 构图、尺寸、gap、padding、背景、阴影和对齐方式逐项复刻。**保留**所有 D1 `data-testid` 不变；视觉只通过 `className` / 设计 token 表达。

#### 3.2 五入口与显示控制视觉

为五入口按钮、主题 menu、custom accent 控件、暗色 icon toggle、语言 icon dropdown、登录 / 注册 / 用户菜单按钮接入语义 className，并原生复刻 `ui-design/src/app.jsx` TopBar 与 `ui-design/src/screen-home.jsx` 的视觉节奏。`aria-current` / `aria-pressed` / `aria-label` 行为不变；字号、padding、圆角、gap、active 状态、控件密度必须来自原型事实，不得凭组件库默认值自由生成。

### Phase 4: 认证页视觉接入

#### 4.1 Auth 页 shell 视觉

为 `auth_login` / `auth_register` / `auth_verify` / `auth_reset` / `auth_logout` 五个 screen 接入卡片化视觉（标题、表单容器、CTA 按钮、错误 / 状态提示），节奏对齐 `ui-design/src/screen-auth.jsx` 与 `docs/ui-design/auth-and-entry.md`。**保留** D1 表单字段、`data-testid`、pendingAction wiring、auth-only params 隔离不变。

### Phase 5: Profile / Settings placeholder 视觉接入

#### 5.1 Profile / Settings shell 视觉

为 `profile` 与 `settings` placeholder shell 接入卡片节奏与分区标题（账号 / 隐私 / 字体预设），参考 `ui-design/src/screen-profile.jsx` 与 `docs/ui-design/user-profile-and-settings.md`。**禁止**恢复旧 Growth / Experiences / Mistakes / Drill / 独立 Voice 模块视觉；**保留** D1 `data-testid` 不变。

#### 5.2 PlaceholderScreen 视觉占位

为 `PlaceholderScreen.tsx` 提供与 D2-D6 业务页面相同的卡片骨架占位（标题 + 描述 + skeleton 区），让后续 D2-D6 owner 在同一视觉骨架内展开业务内容。占位仅展示 route name 与 i18n 文案，**不**承担 D2-D6 业务内容。

### Phase 6: Regression / handoff

#### 6.1 D1 行为 regression

重跑 `pnpm --filter @easyinterview/frontend test`（当前 D1 前端全量测试必须全通过）+ `E2E.P0.001` / `E2E.P0.002` / `E2E.P0.004`（场景验证必须全通过）。任一退化必须在本 plan 内修复，不得带入 D2-D6。

#### 6.2 真实 build smoke

执行 `pnpm --filter @easyinterview/frontend build` 与根 `make build`，确认视觉接入未破坏 vite bundler / HTML entry / runtime entry。

#### 6.3 Active-scope 负向搜索

`grep` active 前端代码与 `frontend/package.json`，确认未新增 `tailwindcss` / `postcss-tailwind` / `styled-components` / `@emotion/*` 依赖；确认源码未引入无法随仓库交付的私有字体资产；确认 `frontend/`、owner spec/plan/checklist、`AGENTS.md` / `CLAUDE.md` / `GEMINI.md` 不再引用已删除的设计参考文件名、历史设计导入标识或私有品牌字体名称。执行证据必须记录具体 grep 模式，并排除本 gate 文本自身与 work-journal 历史记录；允许保留“不得引入外部品牌设计系统”的治理性禁用说明。

#### 6.4 Visual smoke 工具与 BDD gate

引入 visual smoke + parity 工具（优先 Playwright 或等价浏览器渲染工具），在本 plan 的关键 shell 场景中同时打开正式 `frontend` 与 `ui-design` golden preview，检查默认 App shell、TopBar、auth/profile/settings/placeholder shell 在 desktop 与 mobile viewport 下非空渲染、核心控件不重叠、warm/light 与 dark/custom accent 产生可见 computed-style 或截图差异、旧入口未回流，并对源/目标 DOM 锚点、computed style、bounding box 与必要截图差异执行 100% 复刻 gate。该工具作为 `E2E.P0.005` 的验证入口，并由 `bdd-checklist.md` 记录场景资产与执行证据。

#### 6.5 Handoff

更新 `frontend/README.md`（或 package docs），记录设计 token 入口、主题 / 暗色 wiring 方式、字体加载方式、`ui-design` 原生迁移规则、parity gate 重跑方式，以及 D2-D6 业务页面在视觉骨架上扩展的接入点。

## 5 验收标准

- spec.md §6 C-1（默认壳可用）、C-2（pending action 恢复）、C-3（用户菜单分流）、C-7（中英 UI 切换）、C-8（视觉接入对齐 ui-design 真理源）均通过。
- 主题切换：在 4 基础主题（warm / forest / ocean / plum）× 2 模式（light / dark）下，根级 `data-theme` / `data-mode` 翻转后颜色 / 阴影 / 卡片视觉同步更新；warm 主题完整对齐 `ui-design/` 静态原型，其余 3 主题至少色板正确；`customAccent` 激活后只覆盖 accent / accentSoft 语义变量，并在 light / dark 下可见。
- 字体：浏览器加载的 serif / sans 与 `ui-design` 当前字体预设一致，fallback 链完整；源码与依赖中不引入无法随仓库交付的私有字体资产。
- D1 focused / component / structural test 全量通过；D1 全部 `data-testid` 与 i18n key 不变。
- BDD 场景 `E2E.P0.001` / `E2E.P0.002` / `E2E.P0.004` 重跑全部通过；`E2E.P0.005` visual smoke 场景通过，并写入 checklist / bdd-checklist 证据注释。
- `pnpm --filter @easyinterview/frontend build` 与根 `make build` 通过。
- Active-scope 负向搜索：无 `tailwindcss` / `postcss-tailwind` / `styled-components` / `@emotion/*` 依赖；源码无无法随仓库交付的私有字体资产；`frontend/` 与 active owner docs 不再引用已删除的设计参考文件名、历史设计导入标识或私有品牌字体名称；custom accent 未被删除或降级为不可达配置。
- Visual smoke：关键 App shell viewport 非空渲染、TopBar 与显示控制不重叠、主题/暗色/custom accent 可见变化、旧入口负向断言全部通过，并通过 `ui-design` golden preview 的 100% 源级复刻 parity gate。
- `frontend/README.md` 或等价 package docs 更新视觉骨架接入点、`ui-design` 原生迁移规则与 parity gate 重跑方式。

## 6 风险与应对

| 风险 | 应对措施 |
|------|----------|
| 视觉接入破坏 D1 testid 与行为 | Phase 3-5 每项 checklist 均包含"D1 testid / 行为 regression"断言；Phase 6.1 全量 regression 是 hard gate |
| 主题 / 模式 / custom accent 组合工作量失控 | warm 主题完整打磨；forest / ocean / plum 确保色板与对比度正确；custom accent 只覆盖 accent / accentSoft 语义变量，不扩散为整套自定义色板 |
| `customAccent` 被误判为后续降级项 | spec v1.6 与本 plan v1.1 明确 custom accent 是 UI 真理源的一部分；Phase 1 / 3 / 6 均有 implementation、component、visual smoke 与 BDD gate |
| 只断言 className 导致视觉空渲染或重叠漏检 | Phase 6.4 引入 visual smoke 工具，覆盖 desktop/mobile viewport、非空渲染、核心控件重叠和主题/custom accent 可见差异 |
| 误引入 Tailwind / CSS-in-JS / 私有字体 | Phase 6.3 active-scope 负向搜索作为 hard gate；CI / `make build` 不通过即阻断 |
| AI 按自由审美重建设计导致偏离原型 | spec.md §4 与本 plan 明确只允许 `ui-design` 原生迁移；Phase 6.4 visual smoke 必须加入 `ui-design` golden preview parity gate |
| 字体加载阻塞首屏渲染 | `font-display: swap` + 自托管 woff2；首屏使用 sans fallback 链 |
| 视觉与业务页面边界混淆 | PlaceholderScreen 仅承担骨架；业务内容由 D2-D6 owner 在同骨架内扩展；本 plan 不引入业务文案 |
