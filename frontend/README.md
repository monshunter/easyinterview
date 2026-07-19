# frontend

TypeScript / React 前端工程根目录：应用壳层、模拟面试规划、简历工坊、模拟面试、证据化报告与设置等当前 UI 主屏幕落点。

Current design and contract sources: [product-scope](../docs/spec/product-scope/spec.md)、[docs/ui-design](../docs/ui-design/INDEX.md) 与 [openapi-v1-contract](../docs/spec/openapi-v1-contract/spec.md)。正式 UI 只在 `frontend/` 实现与验证。前端 implementation workstream 进入实现时按 [engineering-roadmap S1/S2](../docs/spec/engineering-roadmap/spec.md#62-s1--contract-backed-mock-runway) 创建对应 child spec / plan；包管理与 workspace 由 [shared-conventions-codified](../docs/spec/shared-conventions-codified/spec.md) 与 [local-dev-stack](../docs/spec/local-dev-stack/spec.md) 协同锁定。

Frontend / backend split workflow is governed by [docs/development.md](../docs/development.md#2-frontend--backend-contract-workflow). Frontend work may proceed independently only when the plan records the `operationId` / fixture / frontend consumer / backend handler / persistence / AI dependency / scenario matrix, and any backend gap is explicitly marked `mock-only` or `not-yet-implemented`.

## 1 工具链

D1 frontend-shell 引入 React 18 + Vite 5 + Vitest 2 + jsdom + @testing-library/react；TypeScript 启用 `strict`、`noUncheckedIndexedAccess`、`noUnusedLocals` 与 `noUnusedParameters`。脚本入口：

| 命令 | 用途 |
|------|------|
| `pnpm --filter @easyinterview/frontend dev` | 启动 Vite dev server（默认端口 10900，可用 `FRONTEND_HOST_PORT` 覆盖） |
| `pnpm --filter @easyinterview/frontend build` | typecheck + Vite 构建 |
| `pnpm --filter @easyinterview/frontend lint` | 运行当前前端静态门禁（`tsc --noEmit`）；未引入 ESLint 依赖前不保留未接入配置 |
| `pnpm --filter @easyinterview/frontend typecheck` | 仅运行 `tsc --noEmit` |
| `pnpm --filter @easyinterview/frontend test` | 运行 Vitest 全量套件 |

Vitest 默认 `node` environment 保留 `frontend/src/lib/events/envelope.test.ts` 等 file IO 测试；React 组件测试通过 `// @vitest-environment jsdom` 头切换到 jsdom。`src/test/setup.ts` 加载 `@testing-library/jest-dom` matcher。

## 2 App Shell 接入点（Frontend owner 必读）

App shell 不扩展业务屏幕。后续前端 workstream 在以下契约下接入即可，不要改写或绕过这些边界。

### 2.1 路由表（[`src/app/routes.ts`](./src/app/routes.ts)）

| 类别 | route name | 备注 |
|------|------------|------|
| 一级导航 | `home` / `workspace` / `resume_versions` | TopBar 三入口；唯一可见的 primary nav |
| 上下文页面 | `parse` / `practice` / `reports` / `generating` / `report` | `practice` 与 `reports` 保留 App chrome；`generating` 隐藏 chrome（`isChromeHidden`） |
| 用户菜单 | `settings` / `auth_logout` | TopBar 已登录态展示 |
| 认证页面 | `auth_login` / `auth_verify` / `auth_profile_setup` / `auth_logout` | `auth_register` 与 `auth_reset`（product-scope D-16）均为 out-of-scope alias，normalize 到登录入口，不渲染独立页面 |

Out-of-scope alias 输入在 [`src/app/normalizeRoute.ts`](./src/app/normalizeRoute.ts) 集中映射到当前 route，**不允许在新代码中作为 live route name 出现**；`src/app/scope.test.ts` 自动负向 grep 阻止回流。独立 `voice` route 输入按未知 route fallback 到 `home`；practice/generating/report 的 `mode` / `modality` query 全部丢弃，电话入口仅保留不可点击的 disabled 状态。显式 out-of-scope path 由 `routeUrl.OUT_OF_SCOPE_PATH_TO_ROUTE` 与 `scripts/spaFallback.mjs` 的 `FRONTEND_OUT_OF_SCOPE_PATHS` 共同承接，加载 App 后归一到当前 route；其余范围外路径由 App 兜底到 `home`。

### 2.2 Navigation / pendingAction 契约

- 所有屏幕通过 [`src/app/navigation/NavigationProvider.tsx`](./src/app/navigation/NavigationProvider.tsx) 暴露的 `useNavigation()` 拿 `navigate(loose)`；仅对无可信上下文的自动安全回退使用同一 provider 的 `replaceRoute(loose)`，避免把坏深链写回 browser history；不要自己持有 route state。
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
- 需要在 dev 下打真实 backend 时显式运行 `VITE_EI_API_MODE=real VITE_EI_API_BASE_URL=<full-api-base> pnpm --filter @easyinterview/frontend dev`（例如 `VITE_EI_API_BASE_URL=http://127.0.0.1:10901/api/v1`）。前端 dev 端口默认 `10900`，可通过 `FRONTEND_HOST_PORT` 覆盖；不要依赖相对 `/api/v1`，否则浏览器会打到 Vite 前端 origin。
- App 的 runtime + auth 状态通过 [`src/app/runtime/AppRuntimeProvider.tsx`](./src/app/runtime/AppRuntimeProvider.tsx) 暴露：`useAppRuntime()` 拿到 `client / runtime / auth / refreshAuth`；provider 通过 generated client 的 `getRuntimeConfig` 读取唯一 runtime-config endpoint。
- D1 only wires `getRuntimeConfig` / `getMe` / `startAuthEmailChallenge` / `verifyAuthEmailChallenge` / `updateMe` / `logout`。`getMe` 只在 auth bootstrap/recovery 读取，页面切换不得重新获取账号主题；`updateMe` success 直接刷新内存 auth context。**新增 client 操作必须先修订 B2 + C1 spec**，然后通过 `src/app/auth/authContractGate.test.ts` 把允许集合扩到允许列表。

### 2.4 Mock 数据源边界

- 生产入口与测试均以 `openapi/fixtures/<tag>/<operationId>.json` 为唯一 mock 来源；不得复制第二套页面展示数据。
- Vite dev 默认也以 `openapi/fixtures/<tag>/<operationId>.json` 为 mock 来源，确保未启动真实 backend 时仍能看到已开发页面；真实 backend 联调必须显式切 `VITE_EI_API_MODE=real`。
- 缺失 scenario 必须先在 fixtures 仓库补，再消费；`createFixtureBackedFetch` 在未知 scenario 上 fail loudly。
- 新增或改动业务数据消费时，前端 owner 必须同步维护 [development operation matrix](../docs/development.md#21-operation-matrix-requirement)，不得把 fixture-backed UI 误标为真实 backend 闭环。
- 只有 `src/api/generated/client.ts` 暴露的 generated method 和 `src/api/mockTransport.ts` 的 fixture-backed fetch 可以作为 API 接入边界；不得在 screen 内手写 ad hoc fetch shape 或复制 fixture JSON。

### 2.5 I18n 接入边界

- D1 shell i18n helper 位于 [`src/app/i18n/messages.ts`](./src/app/i18n/messages.ts)，只负责导入 locale、BCP 47 tag 归一化、类型约束和 `useI18n()` helper。
- 每种 UI 语言必须有独立 locale 文件：[`src/app/i18n/locales/zh.ts`](./src/app/i18n/locales/zh.ts)、[`src/app/i18n/locales/en.ts`](./src/app/i18n/locales/en.ts)。不要把多语言 message map 糅合回 `messages.ts` 或组件文件。
- UI 语言优先级为用户显式选择 > 浏览器 locale > English fallback；显式选择写入 `localStorage["ei-lang"]`，未知、缺失或不支持时 fallback English。语言选择只关联前端显示偏好，不由 runtime config 或登录态覆盖。
- 新增语言时新增 locale 文件，并在 `src/app/i18n/localeCatalog.ts` 追加 `SUPPORTED_LOCALES` 元数据（`code` / `label` / `shortLabel` / `aliases`）；TypeScript 必须通过 `LocaleMessages` 校验 key 完整性，同时扩展 `localeFiles.test.ts` 和 i18n component test。只有真实运行环境中的完整用户流程需要覆盖新语言时才新增 E2E。
- TopBar 语言选择保持为可访问 icon dropdown：`button[data-testid="topbar-lang-toggle"]` 显示 globe icon + 当前语言标签（如 `中文` / `English`）并打开 `topbar-lang-menu`，语言项使用 `topbar-lang-option-{locale}`，方便后续新增 locale；不要改成按钮组或只切状态的静态控件。
- RouteName、testid、URL/hash 和业务语言字段不本地化；`Accept-Language` 只作为 generated client 的 UI display hint，不覆盖 `targetLanguage` / practice language 等业务字段。

### 2.6 Frontend owner 边界

| owner | 写入范围 |
|-------|----------|
| `frontend-home-job-picks-and-parse` | `src/app/screens/home/HomeScreen.tsx` / `src/app/screens/parse/ParseScreen.tsx`（out-of-scope job-recommendation inputs 归一回 `home`） |
| `frontend-workspace-and-practice` | `src/app/screens/workspace/WorkspaceScreen.tsx` / `src/app/screens/practice/PracticeScreen.tsx`；完成动作只负责 stable `reportId` handoff |
| `frontend-report-dashboard` | `src/app/screens/reports/ReportsScreen.tsx` / `src/app/screens/generating/GeneratingScreen.tsx` / `src/app/screens/report/ReportScreen.tsx`；ReportsScreen 组合 `getTargetJob + listTargetJobReports` 并只展示当前规划 current/latest，两页详情共同消费 `getFeedbackReport` 的服务端状态与冻结上下文投影 |
| `frontend-resume-workshop` | `src/app/screens/resume-workshop/ResumeWorkshopScreen.tsx` 及子模块 |

替换 `RouteShellScreen` 派发时只需在 [`src/app/App.tsx`](./src/app/App.tsx) `renderRouteScreen` switch 内增加分支；不要新增独立路由表或独立 navigation provider。

### 2.7 视觉骨架接入点

`frontend-shell/002-app-shell-visual-system` 建立了统一的视觉 token、字体、TopBar 节奏、auth 卡片骨架与通用 screen shell。后续 frontend owner 在以下接入点内扩展业务内容，不要绕过 token 体系或重写视觉骨架。

#### Design tokens 入口

- 语义 token：[`src/app/theme/tokens.ts`](./src/app/theme/tokens.ts) — 仅导出 CSS variable 名（`--ei-color-*` / `--ei-radius-*` / `--ei-shadow-*` / `--ei-space-*` / `--ei-text-*` / `--ei-font-*`），不导出 hex 字面量。
- 主题数据：[`src/app/theme/themes.data.ts`](./src/app/theme/themes.data.ts)（内部）— `ocean` / `plum` 2 主题 × 2 模式 × 21 个颜色角色；`THEME_METADATA` 供 Settings Appearance 使用，custom accent 与 Ocean / Plum 一起作为账号级偏好保存。全站固定使用 Noto Serif SC + Inter，并以 JetBrains Mono 承接标签和代码文本，不提供设置页字体预设。
- 主题 CSS：[`src/app/theme/themes.css`](./src/app/theme/themes.css) — `:root[data-theme=X][data-mode=Y]` 8 组合声明所有色板。
- Custom accent helper：[`src/app/theme/customAccent.ts`](./src/app/theme/customAccent.ts) — 维护 oklch 公式（light=58 / dark=68 / soft 92/28，chroma clamp [0,0.28]，hue normalize [0,360)），仅覆盖 `--ei-color-accent` / `--ei-color-accent-soft`。

新增 token 必须按 `tokens.test.ts` / `themes.css` / `themes.data.ts` 三处同步追加，并在测试中固化正式前端内部契约。

#### 主题 / 暗色 / customAccent 根级 wiring

[`src/app/display/DisplayPreferencesProvider.tsx`](./src/app/display/DisplayPreferencesProvider.tsx) 在 `theme` / `dark` / `customAccent` 任一切换时立即把 `<html>` 的 `data-theme` / `data-mode` / `data-custom-accent` 翻转，并把 customAccent overlay 写入根元素 inline style。**所有主题相关样式必须走 `:root[data-theme][data-mode]` selector + var() token，不在组件内 hardcode hex / rgb。**

- TopBar 只承接暗色 toggle、语言 dropdown 与设置齿轮；账号主题控件由 [`src/app/screens/SettingsScreen.tsx`](./src/app/screens/SettingsScreen.tsx) 的 Appearance 区承接。
- 账号主题 testid：`settings-theme-{ocean,plum,custom}` / `settings-custom-accent-{hue,chroma}` / `settings-theme-save`。Custom accent picker 只保留 hue / saturation，不保留 preview / value / reset anchor 或隐藏清除入口。

#### 字体加载

- 字体来源：[`src/app/theme/fonts.css`](./src/app/theme/fonts.css) 仅通过 `@fontsource/{noto-serif-sc,inter,jetbrains-mono}` 引入；fontsource 默认带 `font-display: swap`，首屏使用 system fallback 链。
- Typography scale：[`src/app/theme/typography.css`](./src/app/theme/typography.css) 提供 `--ei-text-{display,title,subtitle,body,caption,label}-*` 4 维度 24 个 token + `.ei-text-*` 6 类 className。组件内**禁止内联 px font-size / line-height**，改用 `ei-text-*` className。
- 不引入私有品牌字体（`copernicus` / `styreneb` 等）；新增字体必须以 fontsource 或可仓库自托管为前提。

#### 视觉骨架与卡片节奏

| 区域 | className 入口 | CSS 文件 |
|------|---------------|---------|
| TopBar | `ei-shell-topbar` / `ei-topbar-{nav,nav-button,controls,user,dark,lang,auth-{login,register},user-button}` | [`src/app/topbar/topbar.css`](./src/app/topbar/topbar.css) |
| 认证页 | `ei-auth-shell` / `ei-auth-{side,side-panel,card,form,field,cta,status,row}` | [`src/app/auth/auth.css`](./src/app/auth/auth.css) |
| Settings / fallback shell / 业务屏幕 | `ei-screen-shell` / `ei-screen-card` / `ei-skeleton-{stripe,line}` | [`src/app/screens/screens.css`](./src/app/screens/screens.css) |

Frontend owner 替换 fallback shell 时应保留 `ei-screen-shell` 外壳与 `ei-screen-card` 节奏；新分区只在 card 内部展开内容，不在 shell 外加自定义 wrapper。

#### 代码层 visual smoke 与浏览器验证

D2 视觉系统由正式前端自身的可执行 gate 守住：

1. **jsdom unit smoke（毫秒级）**：覆盖 DOM 锚点 / className / `:root[data-theme][data-mode]` selector resolution / customAccent inline overlay / out-of-scope module 负向。开发中可 focused 运行，阶段完成统一执行根 `make test` 全量回归。

   ```bash
   pnpm --filter @easyinterview/frontend test src/app/__tests__/app-shell-visual-system.test.tsx
   make test
   ```

2. **真实浏览器验证**：只有需要验证运行态交互、响应式布局或前后端业务链路时才使用 repository-defined browser scenario；场景必须访问真实 frontend，且业务请求落到真实 backend。截图是针对正式 UI 的辅助证据，不建立第二套参考页面或快照基线。

## 3 UI 设计文档与正式实现

- `docs/ui-design/` 定义信息架构、页面职责、流程、状态、响应式约束和视觉原则；正式 `frontend/` 是唯一可运行实现。
- 新页面或大幅交互修订先更新对应设计文档，再实施组件、路由、数据与样式。
- DOM、token、className 和具体样式值由正式前端代码维护，并通过组件测试、响应式断言、可访问性检查、构建以及必要的真实浏览器场景验证。
- 不为展示目的建立平行运行时、重复组件、重复 fixture 或专用预览路由；设计文档与实现偏离时，修订错误的一方。
- 不引入外部品牌设计系统作为替代参考；新的视觉方向先在 `docs/ui-design/` 收敛，再由正式前端实现。
