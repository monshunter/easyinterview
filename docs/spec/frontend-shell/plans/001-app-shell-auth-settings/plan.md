# App Shell, Auth Gate, and Settings Entrypoints

> **版本**: 1.21
> **状态**: completed
> **更新日期**: 2026-07-10

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)
**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)
**关联 BDD Checklist**: [bdd-checklist](./bdd-checklist.md)

## 1 目标

本计划交付当前正式前端 App shell：默认 Home、三入口 TopBar、全局显示控制、email-code 认证页、资料补全 gate、`requestAuth(pendingAction)` 恢复、用户菜单、settings 双 tab、runtime / generated client bootstrap，以及面试业务 route 的登录前置保护。

当前完成态文档只描述现行合同。任何新增可见页面、route、auth flow 或设置页能力，必须先更新 `ui-design/` 静态原型、`docs/ui-design/` 和 `frontend-shell` spec，再修订本 owner 或派生明确边界的新 plan。

## 2 当前合同

### 2.1 UI 与 route catalog

- UI 真理源：`docs/ui-design/`、`ui-design/src/app.jsx`、`ui-design/src/screen-auth.jsx`、`ui-design/src/screens-p0-complete.jsx`。
- 一级 TopBar 入口：`home`、`workspace`、`resume_versions`。
- 上下文 route：`parse`、`practice`、`generating`、`report`。
- 用户菜单 route：`settings`、`auth_logout`。
- 认证 route：`auth_login`、`auth_verify`、`auth_profile_setup`、`auth_logout`。
- Unsupported route / malformed URL 必须经同一 normalization 层折回当前 route catalog 或 `home`，不得 materialize 独立页面。

### 2.2 Auth / settings / display

- Auth UI 只通过 email-code flow 触发 `startAuthEmailChallenge`、`verifyAuthEmailChallenge`、`getMe`、`completeMyProfile`、`logout` 和 first-party session cookie。
- `profileCompletionRequired=true` 时，登录后必须先进入 `auth_profile_setup`；资料补全成功并刷新 `/me` 后，才恢复 pendingAction 或回 Home。
- `pendingAction` 只保存 route name、canonical URL 和 safe params，不保存 JD 原文、简历原文、验证码、AI prompt/response 或解析正文。
- Settings 只保留 `个人资料` 与 `隐私与数据` 两个 tab；`个人资料` tab 承接账号基础信息、登录与安全展示、字体预设和产品信息。
- 显示偏好由前端持有：主题、暗色、语言下拉和字体预设在登录前后保持稳定；默认主题与无效值 fallback 为 `ocean`。

## 3 质量门禁分类

- **Plan 类型**: `feature-behavior` + `frontend` + `contract`。
- **TDD 策略**: 本计划按 `/implement frontend-shell/001-app-shell-auth-settings frontend` -> `/tdd` 完成。再次修改该合同必须先补 focused Vitest / component test / route-state test，再改实现；测试断言覆盖 route normalization、TopBar、display preferences、Auth screens、runtime auth provider、pendingAction、settings、dev mock session state 和 protected route guard。
- **BDD 策略**: 需要 BDD。本计划维护 [bdd-plan](./bdd-plan.md) 与 [bdd-checklist](./bdd-checklist.md)，主 checklist 以 `E2E.P0.001`、`E2E.P0.002`、`E2E.P0.004`、`E2E.P0.032`、`E2E.P0.101`、`E2E.P0.102` 作为当前行为 gate。
- **替代验证 gate**: UI-only 或 recovery-path 修订通过 focused Vitest、UI truth-source structure checks、source-level zero-residue grep、`pnpm --filter @easyinterview/frontend typecheck` / `build`、context validator、`make docs-check` 和 `git diff --check` 收口。

## 4 Operation Matrix

| operationId | fixture | frontend consumer | backend handler | persistence | AI dependency | scenario coverage |
|-------------|---------|-------------------|-----------------|-------------|---------------|-------------------|
| `getRuntimeConfig` | `openapi/fixtures/Auth/getRuntimeConfig.json#default` | `AppRuntimeProvider`、runtime bootstrap、display request options | backend-auth runtime config handler | 无 | 无 | focused runtime tests、E2E.P0.001、E2E.P0.004、E2E.P0.032 |
| `getMe` | `openapi/fixtures/Auth/getMe.json#authenticated\|unauthenticated\|profileIncomplete` | `AppRuntimeProvider`、TopBar user area、profile setup guard、protected route guard | backend-auth current-user handler | backend session cookie lookup | 无 | App runtime tests、E2E.P0.032、E2E.P0.101、E2E.P0.102 |
| `startAuthEmailChallenge` | `openapi/fixtures/Auth/startAuthEmailChallenge.json#default` | `AuthLoginScreen`、`requestAuth(pendingAction)` start path | backend-auth challenge issue handler | backend auth challenge storage | 无 | Auth focused tests、E2E.P0.002、E2E.P0.101 |
| `verifyAuthEmailChallenge` | `openapi/fixtures/Auth/verifyAuthEmailChallenge.json#default` | `AuthVerifyScreen`、post-verify auth refresh | backend-auth verify handler | backend consumes challenge and mints first-party session cookie | 无 | Auth focused tests、E2E.P0.002、E2E.P0.101 |
| `completeMyProfile` | `openapi/fixtures/Auth/completeMyProfile.json#default` | `AuthProfileSetupScreen` | backend-auth `PATCH /me` handler | backend user display name, terms acceptance and profile completion fields | 无 | Auth focused tests、E2E.P0.101 |
| `logout` | `openapi/fixtures/Auth/logout.json#default` | `AuthLogoutScreen`、TopBar logout route、dev mock client state clear | backend-auth logout handler | backend clears session cookie/session | 无 | dev mock tests、E2E.P0.032、E2E.P0.101 |
| `listTargetJobs` | `openapi/fixtures/TargetJobs/listTargetJobs.json#default\|empty\|one-job\|twelve-plus` | Home recent records visibility under authenticated runtime | targetjob list handler behind session middleware | `target_jobs` filtered by session user | 无 | Home auth guard tests、E2E.P0.102 |
| `importTargetJob` | `openapi/fixtures/TargetJobs/importTargetJob.json#manual-text-primary` | Home submit via auth gate + safe pendingAction | targetjob import handler behind session middleware | `target_jobs` / async parse job | target job parse may use AI after authenticated import | Home auth gate tests、E2E.P0.102 |
| `N/A` UI shell/settings/display | N/A | `routes.ts`、`normalizeRoute.ts`、`routeUrl.ts`、`TopBar`、`SettingsScreen`、`DisplayPreferencesProvider` | 无 API 变更 | frontend local display preference only | 无 | focused Vitest、pixel parity / visual smoke owners、E2E.P0.001、E2E.P0.004、E2E.P0.032 |

## 5 验收标准

- 默认打开 App 渲染 Home、三入口 TopBar、单一登录入口、用户菜单区和显示控制。
- Browser History URL、hash adapter 输入和 in-memory route 均进入同一 normalization / route store 合同。
- 语言、主题、暗色与字体预设在未登录、登录、退出登录和 `/me` refresh 中保持前端偏好优先级；generated client 请求携带当前 UI locale display hint。
- 未登录用户触发受保护动作时进入 `auth_login(pendingAction)`；email-code 验证成功后先执行资料补全 gate，再恢复 safe route params。
- Settings 只有 `个人资料` / `隐私与数据` 两个 tab；账号登录与安全展示保持 email-code 口径。
- 面试业务 route 在 runtime auth loading / unauthenticated 状态下不挂载业务 screen，不调用受保护 API；Home 未登录态不请求账号记录。
- Vite dev mock 从 unauthenticated 开始，verify 后 `/me` 变为 authenticated 或 profileIncomplete，logout 后回到 unauthenticated。
- Auth verify 成功后的 `/me` refresh failure 不被渲染为验证码错误；App 离开 verify 页并在 route gate 中表达 auth/profile loading 或 error。
- UI 结构、文案、密度、主题和交互节奏可追溯到 `ui-design/` 与 `docs/ui-design/`。

## 6 当前验证面

- Focused frontend gates: route normalization / URL codec、App auth dispatch、runtime provider、Auth screens、TopBar、DisplayPreferencesProvider、Settings visual、dev mock client、Home auth guard。
- BDD gates: `E2E.P0.001`、`E2E.P0.002`、`E2E.P0.004`、`E2E.P0.032`、`E2E.P0.101`、`E2E.P0.102`。
- Contract/doc gates: frontend-shell context validation、product-scope context validation、`sync-doc-index --check`、`make docs-check`、`git diff --check`、`make lint-core-loop-pruning-surface`。

### Phase 8: auth alias test lifecycle isolation

`AppAuthDispatch.test.tsx` 中的 `auth_reset` / `auth_register` 用例只验证同步 route normalization。两项在断言后显式 unmount，使不属于测试目标的 runtime-config/auth Promise 不再于测试尾部回写 provider state；生产 App、runtime provider、route behavior 与 BDD 合同不变。

门禁：focused AppAuthDispatch 14 tests 通过且无 React act warning，frontend-shell focused/full frontend tests、typecheck/build 与 owner docs gates 保持绿色。BDD 不适用，因为本批只修正测试生命周期。

## 7 风险与应对

| 风险 | 应对措施 |
|------|----------|
| Shell owner 扩大到业务页面实现 | 本 plan 只 owning App shell、auth、settings、display 和 route/auth gate；业务页面内容由对应 subject owner 承接 |
| Auth flow 绕过 generated client/session cookie | Operation matrix 固化 generated auth operations；focused tests 阻止自定义 session wire |
| PendingAction 泄露敏感正文 | Safe-param allowlist 与 URL/privacy tests 只允许稳定 ID 和 display hint |
| Route catalog 漂移 | `normalizeRoute` / `routeUrl` focused tests 与 BDD gates 共同验证 unsupported input 不 materialize 独立页面 |
| UI 与原型偏离 | `ui-design/` 源码和 `docs/ui-design/` 是唯一 UI 真理源；可见变更必须先更新原型再迁移正式前端 |

## 8 修订记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-07-10 | 1.21 | Isolate synchronous auth alias tests from unrelated runtime-provider updates. |
| 2026-07-07 | 1.20 | Compress owner plan to the current App shell / auth / settings contract, operation matrix, and current gate surface. |
