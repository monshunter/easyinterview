# frontend

TypeScript / React 前端工程根目录：应用壳层、模拟面试规划、简历工坊、模拟面试、证据化报告、真实面试复盘与设置等当前 UI 主屏幕落点。

Current truth sources: [product-scope](../docs/spec/product-scope/spec.md)、[docs/ui-design](../docs/ui-design/INDEX.md)、[ui-design](../ui-design/) 与 [openapi-v1-contract](../docs/spec/openapi-v1-contract/spec.md)。前端 implementation workstream 进入实现时按 [engineering-roadmap S1/S2](../docs/spec/engineering-roadmap/spec.md#62-s1--contract-backed-mock-runway) 创建对应 child spec / plan；包管理与 workspace 由 [shared-conventions-codified](../docs/spec/shared-conventions-codified/spec.md) 与 [local-dev-stack](../docs/spec/local-dev-stack/spec.md) 协同锁定。

## 1 工具链

D1 frontend-shell 引入 React 18 + Vite 5 + Vitest 2 + jsdom + @testing-library/react；TypeScript `strict` + `noUncheckedIndexedAccess`。脚本入口：

| 命令 | 用途 |
|------|------|
| `pnpm --filter @easyinterview/frontend dev` | 启动 Vite dev server（端口 5173） |
| `pnpm --filter @easyinterview/frontend build` | typecheck + Vite 构建 |
| `pnpm --filter @easyinterview/frontend typecheck` | 仅运行 `tsc --noEmit` |
| `pnpm --filter @easyinterview/frontend test` | 运行 Vitest 全量套件 |

Vitest 默认 `node` environment 保留 `frontend/src/lib/events/envelope.test.ts` 等 file IO 测试；React 组件测试通过 `// @vitest-environment jsdom` 头切换到 jsdom。`src/test/setup.ts` 加载 `@testing-library/jest-dom` matcher。

## 2 D1 App shell 接入点（D2-D6 owner 必读）

D1 不会再扩展业务屏幕。后续前端 workstream 在以下契约下接入即可，不要改写或绕过这些边界。

### 2.1 路由表（[`src/app/routes.ts`](./src/app/routes.ts)）

| 类别 | route name | 备注 |
|------|------------|------|
| 一级导航 | `home` / `jd_match` / `workspace` / `resume_versions` / `debrief` | TopBar 五入口；唯一可见的 primary nav |
| 上下文页面 | `parse` / `practice` / `generating` / `report` / `company_intel` | `practice` / `generating` 隐藏 chrome（`isChromeHidden`） |
| 用户菜单 | `profile` / `settings` / `auth_logout` | TopBar 已登录态展示 |
| 认证页面 | `auth_login` / `auth_register` / `auth_verify` / `auth_reset` / `auth_logout` | `auth_reset` 仅 UI shell |

旧 alias（`welcome` / `growth` / `plan` / `mistakes` / `drill` / `followup` / `experiences` / `star` / `resume` / `onboarding` / 独立 `voice`）在 [`src/app/normalizeRoute.ts`](./src/app/normalizeRoute.ts) 集中映射到当前 route，**不允许在新代码中作为 live route name 出现**；`src/app/scope.test.ts` 自动负向 grep 阻止回流。`practice?mode=voice&modality=voice` 是语音面试唯一显式入口。

### 2.2 Navigation / pendingAction 契约

- 所有屏幕通过 [`src/app/navigation/NavigationProvider.tsx`](./src/app/navigation/NavigationProvider.tsx) 暴露的 `useNavigation()` 拿 `navigate(loose)`；不要自己持有 route state。
- 业务动作用 [`src/app/auth/useRequestAuth.ts`](./src/app/auth/useRequestAuth.ts) 钩子，传入 `PendingAction`：

```ts
const requestAuth = useRequestAuth();
requestAuth({
  type: "start_practice",
  label: "立即面试",
  route: "practice",
  params: { planId, targetJobId, jdId, resumeVersionId, roundId },
});
```

未登录时跳 `auth_login` 并把 pending action 编码到路由参数；登录成功后 `AuthVerifyScreen` 自动 decode 并恢复目标 route + 5 个 interview-context key。pendingAction 编码规则见 [`src/app/auth/pendingAction.ts`](./src/app/auth/pendingAction.ts) 与 [`docs/ui-design/auth-and-entry.md` §6 / §8](../docs/ui-design/auth-and-entry.md)。

### 2.3 Runtime / Mock transport 入口

- [`src/api/generated/client.ts`](./src/api/generated/client.ts) 是 B2 OpenAPI 生成的强类型客户端，禁止手改。
- [`src/api/mockTransport.ts`](./src/api/mockTransport.ts) 提供 `createFixtureBackedFetch + createFixtureRegistry`，scenario 通过请求头 `Prefer: example=<scenario>` 选择 fixture。
- App 的 runtime + auth 状态通过 [`src/app/runtime/AppRuntimeProvider.tsx`](./src/app/runtime/AppRuntimeProvider.tsx) 暴露：`useAppRuntime()` 拿到 `client / runtime / auth / refreshAuth`；非 React 路径用 [`src/lib/runtime-config`](./src/lib/runtime-config) 直接读取 runtime config。
- D1 only wires `getRuntimeConfig` / `getMe` / `startAuthEmailChallenge` / `verifyAuthEmailChallenge` / `logout`。**新增 client 操作必须先修订 B2 + C1 spec**，然后通过 `src/app/auth/authContractGate.test.ts` 把允许集合扩到允许列表。

### 2.4 Mock 数据源边界

- 生产入口与测试均以 `openapi/fixtures/<tag>/<operationId>.json` 为唯一 mock 来源；`src/app/scope.test.ts` 阻止 `frontend/src` 直接 import `ui-design/src/data*`。
- 缺失 scenario 必须先在 fixtures 仓库补，再消费；`createFixtureBackedFetch` 在未知 scenario 上 fail loudly。

### 2.5 D2-D6 owner 边界

| owner | 写入范围 |
|-------|----------|
| `frontend-home-job-picks-and-parse` | `src/app/screens/HomeScreen.tsx` / `JdMatchScreen.tsx` / `ParseScreen.tsx`（替换 `PlaceholderScreen` 派发） |
| `frontend-workspace-and-practice` | `src/app/screens/WorkspaceScreen.tsx` / `PracticeScreen.tsx` / `GeneratingScreen.tsx` |
| `frontend-report-dashboard` | `src/app/screens/ReportScreen.tsx` / `CompanyIntelScreen.tsx` |
| `frontend-resume-workshop` | `src/app/screens/ResumeVersionsScreen.tsx` 及子模块 |
| `frontend-debrief` | `src/app/screens/DebriefScreen.tsx` |

替换 `PlaceholderScreen` 派发时只需在 [`src/app/App.tsx`](./src/app/App.tsx) `renderRouteScreen` switch 内增加分支；不要新增独立路由表或独立 navigation provider。

## 3 UI 设计体系参考边界（DESIGN.md handoff）

- [`DESIGN.md`](../DESIGN.md) 是 Anthropic Claude.com 品牌设计体系快照，仅作为**只读参考**：复用语义命名（`feature-card` / `code-window-card` / `callout-card-coral` / `pricing-tier-card-featured` / `cta-band-coral` 等）与章节节奏（cream → cream-card → dark-mockup → coral-callout → dark-footer）。
- `docs/ui-design/` 与 `ui-design/` 仍是 UI 验收的唯一真理源。spec / plan / checklist 不得把 “DESIGN.md 第 X 节” 写为完成判据。
- **不要机械同步 token**：[`ui-design/src/primitives.jsx`](../ui-design/src/primitives.jsx) `EI_THEMES` 是当前真理（warm / forest / ocean / plum + light / dark），不要替换为 `DESIGN.md` 的精确 hex。
- **不要引入私有授权字体**：Copernicus / StyreneB 是 Anthropic 私有授权字体；前端继续使用 serif: Cormorant Garamond / EB Garamond，sans: Inter（与 `EI_FONT_PRESETS` 对齐）。
- 完整使用方式与边界见 [`DESIGN.md` 顶部「使用方式与边界」章节](../DESIGN.md#使用方式与边界easyinterview-项目内)。
