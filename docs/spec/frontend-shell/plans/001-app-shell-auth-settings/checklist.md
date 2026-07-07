# App Shell, Auth Gate, and Settings Entrypoints Checklist

> **版本**: 1.20
> **状态**: completed
> **更新日期**: 2026-07-07

**关联计划**: [plan](./plan.md)

## Phase 1: App shell and route contract

- [x] 1.1 App 默认渲染 Home，并在 `practice` / `generating` 等当前上下文 route 上按 spec 控制 chrome；验证: route/App focused tests 覆盖默认 route、chrome behavior 和 params preservation。
- [x] 1.2 Route normalization、URL codec 和 hash adapter 统一进入当前 route catalog；验证: route-state tests 覆盖 unsupported input fallback，不 materialize 独立页面。
- [x] 1.3 Runtime config 与 generated client bootstrap 接入；验证: `getRuntimeConfig`、`getMe` authenticated / unauthenticated / profileIncomplete、fixture-backed mock transport 和 unknown scenario fail-loud tests 通过。

## Phase 2: TopBar, display and i18n

- [x] 2.1 TopBar 只展示 `home`、`workspace`、`resume_versions` 三个一级入口；验证: component tests 与 BDD.P0.001 断言当前导航和用户区结构。
- [x] 2.2 显示偏好支持主题、暗色、语言和字体预设；验证: DisplayPreferencesProvider / TopBar focused tests 覆盖登录前后稳定性、`ocean` 默认主题、custom accent fallback 和 local preference priority。
- [x] 2.3 UI i18n 使用独立 locale 文件和 TopBar language dropdown；验证: i18n structure/runtime tests 覆盖 `zh` / `en` 文案切换、browser locale normalization、`Accept-Language` display hint 和登录态不覆盖前端语言设置。
- [x] 2.4 BDD-Gate: `E2E.P0.001` 默认首页与三入口 Shell 通过。
- [x] 2.5 BDD-Gate: `E2E.P0.004` App Shell 中英语言切换通过。

## Phase 3: Auth and pendingAction

- [x] 3.1 Auth pages 只保留 `auth_login`、`auth_verify`、`auth_profile_setup`、`auth_logout`；验证: Auth focused tests 覆盖四个当前 route、email-code generated operations、error copy 和 route transitions。
- [x] 3.2 `requestAuth(pendingAction)` 保存并恢复 safe route params；验证: AppAuthDispatch / pendingAction tests 覆盖未登录触发业务动作、email-code 验证、资料补全 gate 和恢复目标 route。
- [x] 3.3 Auth API contract gate 通过；验证: frontend source and tests 只 wire generated auth operations + first-party session cookie，不引入自定义 session API。
- [x] 3.4 BDD-Gate: `E2E.P0.002` 登录打断后恢复原业务动作通过。
- [x] 3.5 Auth verify recovery 通过；验证: focused tests 覆盖 verify operation success 后 `/me` refresh failure 的 route-gate 表达，以及 public-auth initial probe skip 只消费一次。

## Phase 4: User menu and settings

- [x] 4.1 用户菜单按 `ui-design/src/app.jsx` 呈现头像 chip + dropdown；验证: TopBar component tests 和 browser parity owner 覆盖 menu open/close、settings/logout 分流、desktop right alignment 和 mobile viewport containment。
- [x] 4.2 Settings shell 只保留 `个人资料` / `隐私与数据` 双 tab；验证: Settings visual tests 覆盖账号基础信息、登录与安全 email-code 展示、字体预设、产品信息和隐私数据区。
- [x] 4.3 BDD-Gate: `E2E.P0.032` Dev mock 登录态菜单与退出闭环通过。

## Phase 5: Protected route guard and Home auth visibility

- [x] 5.1 面试业务 route 受 runtime auth guard 保护；验证: AppAuthDispatch tests 覆盖 auth loading 不挂载业务 screen、unauthenticated 进入 `auth_login(pendingAction)`、authenticated 恢复业务 route。
- [x] 5.2 Home recent records 只在 authenticated 状态请求和渲染；验证: Home auth guard tests 覆盖 unauthenticated / loading / auth error 不调用 `listTargetJobs` 且不展示 raw unauthorized body。
- [x] 5.3 Backend protected API proof 通过；验证: backend auth policy / cmd API focused gates 证明业务 APIs behind session middleware。
- [x] 5.4 BDD-Gate: `E2E.P0.102` 未登录首页与面试业务路由登录前置通过。

## Phase 6: Single-entry login and profile setup

- [x] 6.1 Single-entry email-code login 通过；验证: AuthLogin / App dispatch tests 覆盖 email-only challenge body、safe pendingAction round-trip、account-existence privacy copy and route transitions。
- [x] 6.2 Profile setup guard 通过；验证: runtime / route tests 覆盖 verify 后 profileIncomplete、refresh、deep link、logout/relogin 和 cross-browser relogin 均先进入 `auth_profile_setup`。
- [x] 6.3 Profile setup submit 通过；验证: AuthProfileSetup tests 覆盖 trimmed displayName、acceptedTerms、`completeMyProfile`、`/me.profileCompletionRequired=false` 后恢复 pendingAction。
- [x] 6.4 BDD-Gate: `E2E.P0.101` Mailpit email-code single-entry login + profile setup 通过。

## Phase 7: UX simplification and closeout gates

- [x] 7.1 登录页静态帮助说明、settings 双 tab 和 `ocean` 默认主题对齐 `ui-design/`；验证: focused Vitest、visual tests、typecheck and build gates 通过。
- [x] 7.2 Operation matrix 与 context manifest 对齐当前 generated-client and route catalog；验证: `validate_context.py frontend-shell/001 frontend` 通过。
- [x] 7.3 当前清理回归 gate 通过；验证: owner residual grep、frontend focused tests、product-scope context validation、`sync-doc-index --check`、`make docs-check`、`git diff --check`、`make lint-core-loop-pruning-surface`。
