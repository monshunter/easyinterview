# App Shell, Auth Gate, and Settings Entrypoints

> **版本**: 1.2
> **状态**: active
> **更新日期**: 2026-05-07

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)
**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)
**关联 BDD Checklist**: [bdd-checklist](./bdd-checklist.md)

## 1 目标

落地正式前端 App 壳：默认 Home、五入口 TopBar、全局显示控制、认证页面、用户菜单、`requestAuth(pendingAction)`、登录后恢复动作、`parse` route shell 与 runtime / API bootstrap。完成后，后续 D2-D6 前端 workstream 可以在同一壳内继续实现业务页面。

## 2 背景

当前静态原型已经在 `ui-design/src/app.jsx` 和 `docs/ui-design/` 中锁定了目标 route、TopBar、认证页面和 pending action 模型。`engineering-roadmap` S1 要求先创建 `frontend-shell`，再推进 D2-D6 前端 workstream。本 plan 是第一个正式前端代码 plan。前端新增 shell / auth / settings 组件时参考根目录 `DESIGN.md` 的语义命名和节奏，但验收仍以 `docs/ui-design/`、`ui-design/` 和本 spec 为准。

## 3 质量门禁分类

- **Plan 类型**: `feature-behavior` + `frontend`。
- **TDD 策略**: 通过 `/implement frontend-shell/001-app-shell-auth-settings frontend` -> `/tdd` 执行；每个 checklist item 先写 focused Vitest / component test / route-state test，再实现最小前端代码；测试断言写在 checklist 的 `验证:` 后。Runtime / API bootstrap 测试必须覆盖 `getRuntimeConfig`、`getMe` authenticated / unauthenticated、auth generated operations 与 mock scenario fail-loud。
- **BDD 策略**: 需要 BDD。本 plan 引入用户可见 App shell、TopBar、认证页面和 pending action 行为，必须维护 `bdd-plan.md`、`bdd-checklist.md`，并在主 checklist 中使用 `BDD-Gate:` 引用 `E2E.P0.001`、`E2E.P0.002`。
- **替代验证 gate**: 不适用；BDD gate 是本 plan 的用户行为验证入口。补充 gate 包括 frontend unit tests、typecheck、mock-contract-suite handoff、route negative search、`make docs-check`。

## 4 实施步骤

### Phase 1: App bootstrap and route normalization

#### 1.1 建立正式前端 App shell

创建或整理正式前端入口，使 App 默认进入 `home`，保留 `practice` / `generating` 的无 TopBar chrome 行为，并支持 route params。

#### 1.2 实现 route normalization 与旧 route 拦截

把旧 route key 映射到当前保留 route 或 Home；不得创建旧独立页面。

#### 1.3 接入 runtime config 与 typed API bootstrap

使用 generated client、fixture-backed mock transport 和 `getRuntimeConfig` 建立 App 启动边界；`/me` 只驱动用户区登录态，不得阻塞默认 Home。

#### 1.4 L2 remediation: 删除 `voice` route alias

正式前端不得保留 `voice` route alias。旧 `voice` 输入必须作为未知 route fallback `home`，语音面试只能由 `practice` route 显式携带 `mode=voice` / `modality=voice` 表达。

### Phase 2: TopBar and display controls

#### 2.1 实现五入口 TopBar

TopBar 只展示 `home`、`jd_match`、`workspace`、`resume_versions`、`debrief` 五个一级入口。

#### 2.2 实现全局显示控制

主题色、暗色和语言切换由 TopBar 持有；显示偏好在登录前后保持稳定。

### Phase 3: Auth pages and pending action

#### 3.1 实现认证页面壳

实现 `auth_login`、`auth_register`、`auth_verify`、`auth_reset`、`auth_logout` 页面流，先接 passwordless mock auth / fixture-backed client。D1 只允许真实 wire `startAuthEmailChallenge`、`verifyAuthEmailChallenge`、`getMe`、`logout`；密码登录、OAuth 和 reset 只能作为 UI shell 或 stub，不得私造 API。

#### 3.2 实现 `requestAuth(pendingAction)`

未登录用户触发需要身份的动作时进入登录页；登录成功后恢复 route 和 params。

#### 3.3 固化 Auth API contract gate

为 auth shell 增加负向断言：正式前端不得新增 password / OAuth / Bearer token / 自定义 session API；真实网络边界只通过 B2 generated auth operations 和 first-party session cookie。

#### 3.5 L2 remediation: verify token 与 auth-only params 隔离

`auth_verify` 必须把用户输入的验证 token 传给 generated `verifyAuthEmailChallenge` query；登录 / 注册页临时字段只允许停留在认证页，不得随 pendingAction 恢复到业务 route params。

### Phase 4: User menu, profile, settings

#### 4.1 实现用户菜单入口

未登录展示登录 / 注册；已登录展示头像菜单，并分流到 `profile`、`settings`、`auth_logout`。

#### 4.2 实现 settings/profile placeholder shell

为 `profile` 和 `settings` 提供当前 UI 边界内的页面壳；业务内容可用 mock data，但不得恢复旧 Growth / Experiences / Mistakes 模块。

### Phase 5: BDD and handoff

#### 5.1 执行 App shell BDD gate

按 `bdd-plan.md` 和 `bdd-checklist.md` 验证默认 Home + TopBar 场景。

#### 5.2 执行 auth pendingAction BDD gate

按 `bdd-plan.md` 和 `bdd-checklist.md` 验证登录打断后恢复业务动作。

#### 5.3 Handoff 给后续 D2-D6

记录稳定 route、shell API、mock runtime 入口和后续页面 owner 的接入点。

#### 5.4 UI 设计体系 handoff

记录 `DESIGN.md` 在 D1 后续组件中的只读参考边界：可复用语义组件命名和页面节奏，不机械同步 token，不引入私有授权字体。

## 5 验收标准

- 默认打开 App 渲染 Home、五入口 TopBar、登录 / 注册、显示控制，不出现 welcome。
- `requestAuth(pendingAction)` 能在登录成功后恢复 `practice` 或 `report` 上下文。
- 用户菜单的 `用户画像` 与 `设置与隐私` 分别进入 `profile` 和 `settings`。
- `parse` route 作为 shell route 可达，但 JD 解析业务细节留给后续 owner。
- Runtime config、`/me` 和 auth generated operations 均通过 fixture-backed client 测试，不直接读取 prototype data。
- 旧 route negative search 确认正式前端不保留独立 old route screen。
- `DESIGN.md` 参考边界写入 handoff，不作为验收真理源或 token 同步目标。
- BDD-Gate `E2E.P0.001`、`E2E.P0.002` 通过。

## 6 风险与应对

| 风险 | 应对措施 |
|------|----------|
| 前端 shell 把业务页面一次做大 | 本 plan 只做 route / shell / auth / menu；D2-D6 单独派生 |
| 登录成功丢失业务上下文 | Phase 3.2 unit test 和 E2E.P0.002 BDD gate 强制覆盖 |
| 旧 route 被当兼容入口保留 | Phase 1.2 和 negative search 要求旧 route 只能 normalize，不建独立 screen |
| mock 数据源漂移 | 依赖 `mock-contract-suite`，禁止 import prototype data |
| Auth UI 超出 C1/B2 契约 | Phase 3.1 / 3.3 只允许 generated passwordless session operations；密码 / OAuth / reset 不 wire 真实 API |
| 品牌参考被误当验收真理源 | Phase 5.4 明确 `DESIGN.md` 只读参考，验收仍回到 UI truth source |
