# App Shell, Auth Gate, and Settings Entrypoints Checklist

> **版本**: 1.0
> **状态**: active
> **更新日期**: 2026-05-05

**关联计划**: [plan](./plan.md)

## Phase 1: App bootstrap and route normalization

- [ ] 1.1 建立正式前端 App shell；验证: frontend focused test 断言默认 route 为 `home`，`practice` / `generating` 隐藏 TopBar，其余 route 保留 App chrome
- [ ] 1.2 实现 route normalization 与旧 route 拦截；验证: route-state test 覆盖 `welcome`、`growth`、`plan`、`mistakes`、`drill`、`followup`、`experiences`、`star`、`resume`、`onboarding`、`voice` 映射到当前 route 或 Home，不创建独立 screen

## Phase 2: TopBar and display controls

- [ ] 2.1 实现五入口 TopBar；验证: component test 断言一级导航只包含 `home`、`jd_match`、`workspace`、`resume_versions`、`debrief`
- [ ] 2.2 实现全局显示控制；验证: state test 断言主题色、暗色、语言切换在未登录/已登录状态切换后保持稳定
- [ ] 2.3 BDD-Gate: 验证 E2E.P0.001 通过

## Phase 3: Auth pages and pending action

- [ ] 3.1 实现认证页面壳；验证: component/route test 覆盖 `auth_login`、`auth_register`、`auth_verify`、`auth_reset`、`auth_logout` 渲染和基本跳转
- [ ] 3.2 实现 `requestAuth(pendingAction)`；验证: route-state test 断言未登录点击 `立即面试` 后进入 login，登录成功恢复 `practice` 并保留 planId / targetJobId / jdId / resumeVersionId / roundId
- [ ] 3.3 BDD-Gate: 验证 E2E.P0.002 通过

## Phase 4: User menu, profile, settings

- [ ] 4.1 实现用户菜单入口；验证: component test 断言未登录显示登录/注册，已登录显示 `用户画像`、`设置与隐私`、`退出登录`
- [ ] 4.2 实现 settings/profile placeholder shell；验证: route/component test 断言 `profile` 和 `settings` 分离，settings 只维护账号/隐私/字体预设，不恢复旧 Growth / Experiences / Mistakes

## Phase 5: Handoff

- [ ] 5.1 记录后续 D2-D6 shell 接入点；验证: frontend README 或 package docs 说明 route table、pendingAction contract、mock runtime 入口和后续 owner 边界
- [ ] 5.2 active-scope 负向搜索通过；验证: frontend active code 不含独立 `voice` route、独立 `growth` / `mistakes` / `drill` 页面、prototype data runtime import
