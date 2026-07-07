# frontend

TypeScript / React 前端工程根目录：应用壳层、模拟面试规划、简历工坊、模拟面试、证据化报告与设置等当前 UI 主屏幕落点。

Current truth sources: [product-scope](../docs/spec/product-scope/spec.md)、[docs/ui-design](../docs/ui-design/INDEX.md)、[ui-design](../ui-design/) 与 [openapi-v1-contract](../docs/spec/openapi-v1-contract/spec.md)。前端 implementation workstream 进入实现时按 [engineering-roadmap S1/S2](../docs/spec/engineering-roadmap/spec.md#62-s1--contract-backed-mock-runway) 创建对应 child spec / plan；包管理与 workspace 由 [shared-conventions-codified](../docs/spec/shared-conventions-codified/spec.md) 与 [local-dev-stack](../docs/spec/local-dev-stack/spec.md) 协同锁定。

Frontend / backend split workflow is governed by [docs/development.md](../docs/development.md#2-frontend--backend-contract-workflow). Frontend work may proceed independently only when the plan records the `operationId` / fixture / frontend consumer / backend handler / persistence / AI dependency / scenario matrix, and any backend gap is explicitly marked `mock-only` or `not-yet-implemented`.

## 1 工具链

D1 frontend-shell 引入 React 18 + Vite 5 + Vitest 2 + jsdom + @testing-library/react；TypeScript `strict` + `noUncheckedIndexedAccess`。脚本入口：

| 命令 | 用途 |
|------|------|
| `pnpm --filter @easyinterview/frontend dev` | 启动 Vite dev server（默认端口 5173，可用 `FRONTEND_HOST_PORT` 覆盖） |
| `pnpm --filter @easyinterview/frontend build` | typecheck + Vite 构建 |
| `pnpm --filter @easyinterview/frontend typecheck` | 仅运行 `tsc --noEmit` |
| `pnpm --filter @easyinterview/frontend test` | 运行 Vitest 全量套件 |

Vitest 默认 `node` environment 保留 `frontend/src/lib/events/envelope.test.ts` 等 file IO 测试；React 组件测试通过 `// @vitest-environment jsdom` 头切换到 jsdom。`src/test/setup.ts` 加载 `@testing-library/jest-dom` matcher。

## 2 App Shell 接入点（Frontend owner 必读）

App shell 不扩展业务屏幕。后续前端 workstream 在以下契约下接入即可，不要改写或绕过这些边界。

### 2.1 路由表（[`src/app/routes.ts`](./src/app/routes.ts)）

| 类别 | route name | 备注 |
|------|------------|------|
| 一级导航 | `home` / `workspace` / `resume_versions` | TopBar 三入口；唯一可见的 primary nav |
| 上下文页面 | `parse` / `practice` / `generating` / `report` | `practice` / `generating` 隐藏 chrome（`isChromeHidden`） |
| 用户菜单 | `settings` / `auth_logout` | TopBar 已登录态展示 |
| 认证页面 | `auth_login` / `auth_verify` / `auth_profile_setup` / `auth_logout` | `auth_register` 与 `auth_reset`（product-scope D-16）均为 non-current alias，normalize 到登录入口，不渲染独立页面 |

保留 alias（`welcome` / `growth` / `plan` / `mistakes` / `drill` / `followup` / `experiences` / `star` / `resume` / `onboarding` / `jd_match` / `debrief` / `debrief_full` / `profile` / `auth_register` / `auth_reset`）在 [`src/app/normalizeRoute.ts`](./src/app/normalizeRoute.ts) 集中映射到当前 route，**不允许在新代码中作为 live route name 出现**；`src/app/scope.test.ts` 自动负向 grep 阻止回流。独立 `voice` route 输入按未知 route fallback 到 `home`；`practice?mode=voice&modality=voice` 是语音面试唯一显式入口。`/auth/reset`、`/jd-match`、`/debrief`、`/profile` path 由 `routeUrl.NON_CURRENT_PATH_TO_ROUTE` 与 `scripts/spaFallback.mjs` 的 `FRONTEND_NON_CURRENT_PATHS` 共同承接，加载 App 后归一到当前 route；其余非当前路径由 App 兜底到 `home`。

### 2.2 Navigation / pendingAction 契约

- 所有屏幕通过 [`src/app/navigation/NavigationProvider.tsx`](./src/app/navigation/NavigationProvider.tsx) 暴露的 `useNavigation()` 拿 `navigate(loose)`；不要自己持有 route state。
- 业务动作用 [`src/app/auth/useRequestAuth.ts`](./src/app/auth/useRequestAuth.ts) 钩子，传入 `PendingAction`：

```ts
const requestAuth = useRequestAuth();
requestAuth({
  type: "start_practice",
  label: "立即面试",
  route: "practice",
  params: { planId, targetJobId, jdId, resumeId, roundId },
});
```

未登录时跳 `auth_login` 并把 pending action 编码到路由参数；登录成功后 `AuthVerifyScreen` 自动 decode 并接续目标 route + 5 个 interview-context key。pendingAction 编码规则见 [`src/app/auth/pendingAction.ts`](./src/app/auth/pendingAction.ts) 与 [`docs/ui-design/auth-and-entry.md` §6 / §8](../docs/ui-design/auth-and-entry.md)。

### 2.3 Runtime / Mock transport 入口

- [`src/api/generated/client.ts`](./src/api/generated/client.ts) 是 B2 OpenAPI 生成的强类型客户端，禁止手改。
- [`src/api/mockTransport.ts`](./src/api/mockTransport.ts) 提供 `createFixtureBackedFetch + createFixtureRegistry`，scenario 通过请求头 `Prefer: example=<scenario>` 选择 fixture。
- [`src/api/clientFactory.ts`](./src/api/clientFactory.ts) 是正式 bootstrap 入口：Vite dev (`import.meta.env.DEV`) 默认使用 fixture-backed client，production 默认使用 same-origin `/api/v1`。
- 需要在 dev 下打真实 backend 时显式运行 `VITE_EI_API_MODE=real VITE_EI_API_BASE_URL=<full-api-base> pnpm --filter @easyinterview/frontend dev`（例如 `VITE_EI_API_BASE_URL=http://localhost:8080/api/v1`）。前端 dev 端口可通过 `FRONTEND_HOST_PORT` 覆盖；不要依赖相对 `/api/v1`，否则浏览器会打到 Vite 前端 origin。
- App 的 runtime + auth 状态通过 [`src/app/runtime/AppRuntimeProvider.tsx`](./src/app/runtime/AppRuntimeProvider.tsx) 暴露：`useAppRuntime()` 拿到 `client / runtime / auth / refreshAuth`；非 React 路径用 [`src/lib/runtime-config`](./src/lib/runtime-config) 直接读取 runtime config。
- D1 only wires `getRuntimeConfig` / `getMe` / `startAuthEmailChallenge` / `verifyAuthEmailChallenge` / `completeMyProfile` / `logout`。**新增 client 操作必须先修订 B2 + C1 spec**，然后通过 `src/app/auth/authContractGate.test.ts` 把允许集合扩到允许列表。

### 2.4 Mock 数据源边界

- 生产入口与测试均以 `openapi/fixtures/<tag>/<operationId>.json` 为唯一 mock 来源；`src/app/scope.test.ts` 阻止 `frontend/src` 直接 import `ui-design/src/data*`。
- Vite dev 默认也以 `openapi/fixtures/<tag>/<operationId>.json` 为 mock 来源，确保未启动真实 backend 时仍能看到已开发页面；真实 backend 联调必须显式切 `VITE_EI_API_MODE=real`。
- 缺失 scenario 必须先在 fixtures 仓库补，再消费；`createFixtureBackedFetch` 在未知 scenario 上 fail loudly。
- 新增或改动业务数据消费时，前端 owner 必须同步维护 [development operation matrix](../docs/development.md#21-operation-matrix-requirement)，不得把 fixture-backed UI 误标为真实 backend 闭环。
- 只有 `src/api/generated/client.ts` 暴露的 generated method 和 `src/api/mockTransport.ts` 的 fixture-backed fetch 可以作为 API 接入边界；不得在 screen 内手写 ad hoc fetch shape 或复制 fixture JSON。

### 2.5 I18n 接入边界

- D1 shell i18n helper 位于 [`src/app/i18n/messages.ts`](./src/app/i18n/messages.ts)，只负责导入 locale、BCP 47 tag 归一化、类型约束和 `useI18n()` helper。
- 每种 UI 语言必须有独立 locale 文件：[`src/app/i18n/locales/zh.ts`](./src/app/i18n/locales/zh.ts)、[`src/app/i18n/locales/en.ts`](./src/app/i18n/locales/en.ts)。不要把多语言 message map 糅合回 `messages.ts` 或组件文件。
- UI 语言优先级为用户显式选择 > 浏览器 locale > English fallback；显式选择写入 `localStorage["ei-lang"]`，未知、缺失或不支持时 fallback English。语言选择只关联前端显示偏好，不由 runtime config 或登录态覆盖。
- 新增语言时新增 locale 文件，并在 `src/app/i18n/localeCatalog.ts` 追加 `SUPPORTED_LOCALES` 元数据（`code` / `label` / `shortLabel` / `aliases`）；TypeScript 必须通过 `LocaleMessages` 校验 key 完整性，同时扩展 `localeFiles.test.ts`、i18n component test 和 E2E.P0.004 类场景。
- TopBar 语言选择必须保持为与 `ui-design/src/app.jsx` 一致的可访问 icon dropdown：`button[data-testid="topbar-lang-toggle"]` 显示 globe icon + 当前语言标签（如 `中文` / `English`）并打开 `topbar-lang-menu`，语言项使用 `topbar-lang-option-{locale}`，方便后续新增 locale；不要改成 native select、按钮组或只切状态的占位控件。
- RouteName、testid、URL/hash 和业务语言字段不本地化；`Accept-Language` 只作为 generated client 的 UI display hint，不覆盖 `targetLanguage` / practice language 等业务字段。

### 2.6 Frontend owner 边界

| owner | 写入范围 |
|-------|----------|
| `frontend-home-job-picks-and-parse` | `src/app/screens/home/HomeScreen.tsx` / `src/app/screens/parse/ParseScreen.tsx`（`jd_match` 非当前输入归一回 `home`） |
| `frontend-workspace-and-practice` | `src/app/screens/workspace/WorkspaceScreen.tsx` / `src/app/screens/practice/PracticeScreen.tsx` / `src/app/screens/generating/GeneratingScreen.tsx` |
| `frontend-report-dashboard` | `src/app/screens/report/ReportScreen.tsx` |
| `frontend-resume-workshop` | `src/app/screens/resume-workshop/ResumeWorkshopScreen.tsx` 及子模块 |

替换 `PlaceholderScreen` 派发时只需在 [`src/app/App.tsx`](./src/app/App.tsx) `renderRouteScreen` switch 内增加分支；不要新增独立路由表或独立 navigation provider。

### 2.7 视觉骨架接入点

`frontend-shell/002-app-shell-visual-system` 把 ui-design 静态原型整体迁移到正式前端，建立了统一的视觉 token、字体、TopBar 节奏、auth 卡片骨架与通用 screen shell。后续 frontend owner 在以下接入点内扩展业务内容，不要绕过 token 体系或重写视觉骨架。

#### Design tokens 入口

- 语义 token：[`src/app/theme/tokens.ts`](./src/app/theme/tokens.ts) — 仅导出 CSS variable 名（`--ei-color-*` / `--ei-radius-*` / `--ei-shadow-*` / `--ei-space-*` / `--ei-text-*` / `--ei-font-*`），不导出 hex 字面量。
- 主题数据：[`src/app/theme/themes.data.ts`](./src/app/theme/themes.data.ts)（私有）— 4 主题 × 2 模式 21 色板与 7 字体预设，逐项转写自 `ui-design/src/primitives.jsx::EI_THEMES` / `EI_FONT_PRESETS` / `EI_THEME_LIST`。
- 主题 CSS：[`src/app/theme/themes.css`](./src/app/theme/themes.css) — `:root[data-theme=X][data-mode=Y]` 8 组合声明所有色板。
- Custom accent helper：[`src/app/theme/customAccent.ts`](./src/app/theme/customAccent.ts) — 镜像 `app.jsx` oklch 公式（light=58 / dark=68 / soft 92/28，chroma clamp [0,0.28]，hue normalize [0,360)），仅覆盖 `--ei-color-accent` / `--ei-color-accent-soft`。

新增 token 必须按 `tokens.test.ts` / `themes.css` / `themes.data.ts` 三处同步追加，并在测试中固化追溯到 `ui-design` 源。

#### 主题 / 暗色 / customAccent 根级 wiring

[`src/app/display/DisplayPreferencesProvider.tsx`](./src/app/display/DisplayPreferencesProvider.tsx) 在 `theme` / `dark` / `customAccent` 任一切换时立即把 `<html>` 的 `data-theme` / `data-mode` / `data-custom-accent` 翻转，并把 customAccent overlay 写入根元素 inline style。**所有主题相关样式必须走 `:root[data-theme][data-mode]` selector + var() token，不在组件内 hardcode hex / rgb。**

- TopBar 主题 menu、暗色 toggle、custom accent 控件、语言 dropdown 的 testid / aria 契约见 §2.5 与 [`src/app/topbar/TopBar.tsx`](./src/app/topbar/TopBar.tsx)。
- D2 testid 新增：`topbar-theme-button` / `topbar-theme-menu` / `topbar-theme-option-{warm,forest,ocean,plum}` / `topbar-theme-custom-option` / `topbar-custom-accent-{swatch,picker,hue,chroma,clear}`。

#### 字体加载

- 字体来源：[`src/app/theme/fonts.css`](./src/app/theme/fonts.css) 通过 `@fontsource/{noto-serif-sc,inter,source-serif-pro,cormorant-garamond,ibm-plex-sans,geist-sans,jetbrains-mono}` 引入；fontsource 默认带 `font-display: swap`，首屏使用 system fallback 链。
- Typography scale：[`src/app/theme/typography.css`](./src/app/theme/typography.css) 提供 `--ei-text-{display,title,subtitle,body,caption,label}-*` 4 维度 24 个 token + `.ei-text-*` 6 类 className。组件内**禁止内联 px font-size / line-height**，改用 `ei-text-*` className。
- 不引入私有品牌字体（`copernicus` / `styreneb` 等）；新增字体必须以 fontsource 或可仓库自托管为前提。

#### 视觉骨架与卡片节奏

| 区域 | className 入口 | CSS 文件 |
|------|---------------|---------|
| TopBar | `ei-shell-topbar` / `ei-topbar-{nav,nav-button,controls,user,theme,dark,lang,custom-accent,auth-{login,register},user-button}` | [`src/app/topbar/topbar.css`](./src/app/topbar/topbar.css) |
| 认证页 | `ei-auth-shell` / `ei-auth-{side,side-panel,card,form,field,cta,stub,status,row}` | [`src/app/auth/auth.css`](./src/app/auth/auth.css) |
| Settings / fallback shell / 业务屏幕 | `ei-screen-shell` / `ei-screen-card` / `ei-skeleton-{stripe,line}` / `ei-screen-card-grid` | [`src/app/screens/screens.css`](./src/app/screens/screens.css) |

Frontend owner 替换 fallback shell 时应保留 `ei-screen-shell` 外壳与 `ei-screen-card` 节奏；新分区只在 card 内部展开内容，不在 shell 外加自定义 wrapper。

#### Visual smoke 工具与 parity gate 重跑

D2 视觉系统由 **两层 gate** 共同守住，分工互不替代：

1. **jsdom fast smoke（E2E.P0.005，毫秒级）**：覆盖 DOM 锚点 / className / `:root[data-theme][data-mode]` selector resolution / customAccent inline overlay / non-current module 负向 / `ui-design` 源字面量追溯。日常开发循环跑这个就够。

   ```bash
   pnpm --filter @easyinterview/frontend test src/app/scenarios/p0-005-app-shell-visual-system-smoke.test.tsx
   # 端到端 scenario
   ./test/scenarios/e2e/p0-005-app-shell-visual-system-smoke/scripts/{setup,trigger,verify,cleanup}.sh
   ```

2. **Playwright + chromium pixel parity gate（E2E.P0.006，秒级）**：在 desktop (1440×900) 与 mobile (390×844) 两个 chromium project 下加载 `frontend/dist/index.html` 与 `ui-design/index.html` golden preview，断言 DOM 锚点 + computed style + bounding box + screenshot smoke / 必要截图差异。CI / 主线合并前必跑。

   ```bash
   # 0. 一次性预装 chromium 二进制（首次或新机器）
   pnpm --filter @easyinterview/frontend test:pixel-parity:install

   # 1. 构建 frontend dist（serve-pixel-parity.mjs 依赖）
   pnpm --filter @easyinterview/frontend build

   # 2. 跑当前 pixel parity spec 集合（desktop / mobile viewport 项按 spec 声明）
   pnpm --filter @easyinterview/frontend test:pixel-parity

   # 3. 完整 scenario 入口（包含 pre-check / verify / cleanup）
   ./test/scenarios/e2e/p0-006-ui-design-pixel-parity-gate/scripts/{setup,trigger,verify,cleanup}.sh
   ```

   规约入口：

   - Playwright config：[`frontend/playwright.config.ts`](./playwright.config.ts) 声明 desktop / mobile 两个 chromium project + `webServer` 指向 `serve-pixel-parity.mjs`。
   - 静态 server fixture：[`frontend/scripts/serve-pixel-parity.mjs`](./scripts/serve-pixel-parity.mjs) 同时挂载 `frontend/dist`（`/`）与 `ui-design/`（`/ui-design/`），并暴露 `/health` 探活。
   - 13 个 spec：[`tests/pixel-parity/topbar.spec.ts`](./tests/pixel-parity/topbar.spec.ts)（含 authenticated 头像菜单 dropdown + logout flow）、[`screens.spec.ts`](./tests/pixel-parity/screens.spec.ts)、[`layout.spec.ts`](./tests/pixel-parity/layout.spec.ts)、[`screenshot.spec.ts`](./tests/pixel-parity/screenshot.spec.ts)、[`home.spec.ts`](./tests/pixel-parity/home.spec.ts)、[`parse.spec.ts`](./tests/pixel-parity/parse.spec.ts)、[`workspace.spec.ts`](./tests/pixel-parity/workspace.spec.ts)、[`practice.spec.ts`](./tests/pixel-parity/practice.spec.ts)、[`generating.spec.ts`](./tests/pixel-parity/generating.spec.ts)、[`report.spec.ts`](./tests/pixel-parity/report.spec.ts)、[`resume-workshop.spec.ts`](./tests/pixel-parity/resume-workshop.spec.ts)、[`resume-workshop-create.spec.ts`](./tests/pixel-parity/resume-workshop-create.spec.ts)、[`resume-workshop-branch-rewrites-edit.spec.ts`](./tests/pixel-parity/resume-workshop-branch-rewrites-edit.spec.ts)。

   截图基线维护：

   ```bash
   pnpm exec playwright test tests/pixel-parity/screenshot.spec.ts --update-snapshots
   ```

   baseline 文件位于 `frontend/tests/pixel-parity/*-snapshots/`，默认通过 `frontend/.gitignore` 排除入 git；CI / 本地各自维护。

   Clean checkout gate 不能依赖被 `.gitignore` 排除的本地 snapshot baseline：常规 PASS 证据必须来自 DOM anchor、computed style、bounding box、responsive geometry 或 screenshot smoke（例如非空截图 buffer）。只有在 baseline 可由 CI / checkout 稳定取得或本次显式 `--update-snapshots` 维护时，才能把 `toHaveScreenshot` diff 作为完成 gate。

   `workspace.spec.ts` 的 full-state pixel path 通过测试注入的 initial route 使用 server-bound `targetJobId` / `resumeId` / `planId` 进入完整规划态；不要把它改回 Home recent card 路径，后者按产品语义会携带 `resume-unbound` 并触发 missing-resume 状态。

   修改 frontend bundle 后重跑 Playwright parity 时，先确认 4173 端口没有复用非当前 `dist` 的 server；若存在 stale server，停止后重新运行 gate，避免 `reuseExistingServer` 读取非当前构建产物。

   离线 / 无外网时的局限：`ui-design/index.html` 通过 unpkg.com 加载 React + Babel，并通过 Google Fonts 加载字体；离线运行 ui-design 对照断言会失败。需要离线运行时按 [`test/scenarios/e2e/p0-006-ui-design-pixel-parity-gate/README.md`](../test/scenarios/e2e/p0-006-ui-design-pixel-parity-gate/README.md) §7 vendor CDN 资源。

#### `ui-design` 原生迁移规则

- 新组件 / 新视觉先在 `ui-design/src/*.jsx` 落原型，再在正式前端原生迁移；不允许 AI 自由发挥或外部品牌设计系统补全。
- 每条样式 / token / className 必须能追溯到 `ui-design/src/*.jsx`、`ui-design/src/primitives.jsx` 或 `ui-design/src/app.jsx`；test 文件已固化 hex / fontSize / 布局值 → ui-design 源的 lint 关系。
- 任何视觉偏差不得以「风格接近」收口；要么修到与原型一致，要么先修改 `ui-design/` 真理源（更新 `docs/ui-design/` + 静态原型 + scenario test），再回到正式前端做迁移。

## 3 UI 真理源与原生迁移

- `docs/ui-design/` 与 `ui-design/` 源码是前端 UI 验收的唯一真理源。新页面或大幅视觉修订必须先在 `ui-design/` 完成静态原型，并同步 `docs/ui-design/` 说明。
- 正式 `frontend/` 只做 100% 源级复刻：DOM 构图、布局、间距、字号、字体层级、控件密度、颜色、阴影、边框、圆角、状态、响应式行为和交互节奏必须来自对应 `ui-design/src/*.jsx` 与文档。真实路由、鉴权、数据、可访问性和工程约束可以适配，但不得重新设计、重新解释或重新组合视觉。
- 每个正式组件的样式、token、className 和布局规则必须能追溯到对应 `ui-design/src/*.jsx`、[`ui-design/src/primitives.jsx`](../ui-design/src/primitives.jsx)、[`ui-design/src/app.jsx`](../ui-design/src/app.jsx) 或 `docs/ui-design/`；不得凭 AI 判断补齐未在原型中出现的视觉值。
- 视觉 plan / checklist 必须带 parity gate：至少验证 DOM 锚点、关键 computed style、bounding box、viewport 布局和必要截图差异。只断言“组件存在”“不重叠”不足以证明符合 UI 原型；任何可见偏差必须修正或先回到 `ui-design/` 更新真理源。
- [`ui-design/src/primitives.jsx`](../ui-design/src/primitives.jsx) 的 `EI_THEMES` / `EI_FONT_PRESETS` 和 [`ui-design/src/app.jsx`](../ui-design/src/app.jsx) 的 runtime 交互模型是正式 token / theme / display controls 的抽取来源。
- 不引入外部品牌设计系统作为替代参考；后续如果需要新的视觉方向，先改 `ui-design/`，再迁移到正式前端。
