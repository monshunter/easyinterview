# App Shell, Auth Gate, and Settings Entrypoints Checklist

> **版本**: 1.1
> **状态**: active
> **更新日期**: 2026-05-06

**关联计划**: [plan](./plan.md)

## Phase 1: App bootstrap and route normalization

- [x] 1.1 建立正式前端 App shell；验证: frontend focused test 断言默认 route 为 `home`，`parse` / `report` / `company_intel` 等上下文 route 保留 App chrome，`practice` / `generating` 隐藏 TopBar
- [x] 1.2 实现 route normalization 与旧 route 拦截；验证: route-state test 覆盖 `welcome`、`growth`、`plan`、`mistakes`、`drill`、`followup`、`experiences`、`star`、`resume`、`onboarding`、`voice` 映射到当前 route 或 Home，不创建独立 screen
- [x] 1.3 接入 runtime config 与 typed API bootstrap；验证: focused test 断言 `getRuntimeConfig` 经 generated client + fixture-backed mock transport 读取，`getMe` 覆盖 `unauthenticated` / `authenticated`，unknown scenario fail loudly，frontend shell 不 import `ui-design/src/data.jsx`

## Phase 2: TopBar and display controls

- [x] 2.1 实现五入口 TopBar；验证: component test 断言一级导航只包含 `home`、`jd_match`、`workspace`、`resume_versions`、`debrief`
- [x] 2.2 实现全局显示控制；验证: state test 断言主题色、暗色、语言切换在未登录/已登录状态切换后保持稳定
- [x] 2.3 BDD-Gate: 验证 E2E.P0.001 通过
<!-- verified: 2026-05-07 method=scenario bddChecklist=complete -->

## Phase 3: Auth pages and pending action

- [x] 3.1 实现认证页面壳；验证: component/route test 覆盖 `auth_login`、`auth_register`、`auth_verify`、`auth_reset`、`auth_logout` 渲染和基本跳转；真实 network wire 只使用 `startAuthEmailChallenge` / `verifyAuthEmailChallenge` / `getMe` / `logout`
- [x] 3.2 实现 `requestAuth(pendingAction)`；验证: route-state test 断言未登录点击 `立即面试` 后进入 login，登录成功恢复 `practice` 并保留 planId / targetJobId / jdId / resumeVersionId / roundId
- [x] 3.3 Auth API contract gate；验证: negative search / focused test 断言 frontend shell 不新增 password auth API、OAuth API、Bearer token auth 或自定义 session storage contract；`auth_reset` 保持 UI shell / stub，真实 API 变更必须先修订 C1 / B2
- [x] 3.4 BDD-Gate: 验证 E2E.P0.002 通过
<!-- verified: 2026-05-07 method=scenario bddChecklist=complete -->

## Phase 4: User menu, profile, settings

- [ ] 4.1 实现用户菜单入口；验证: component test 断言未登录显示登录/注册，已登录显示 `用户画像`、`设置与隐私`、`退出登录`
- [ ] 4.2 实现 settings/profile placeholder shell；验证: route/component test 断言 `profile` 和 `settings` 分离，settings 只维护账号/隐私/字体预设，不恢复旧 Growth / Experiences / Mistakes

## Phase 5: Handoff

- [ ] 5.1 记录后续 D2-D6 shell 接入点；验证: frontend README 或 package docs 说明 route table、pendingAction contract、mock runtime 入口和后续 owner 边界
- [ ] 5.2 active-scope 负向搜索通过；验证: frontend active code 不含独立 `voice` route、独立 `growth` / `mistakes` / `drill` 页面、prototype data runtime import
- [ ] 5.3 记录 UI 设计体系 handoff；验证: frontend README 或 package docs 说明 `DESIGN.md` 只读参考边界、可复用语义命名、不得机械同步 token、不得引入私有授权字体，验收仍以 `docs/ui-design/` 与 `ui-design/` 为准
