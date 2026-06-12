# Frontend Shell BDD Checklist

> **版本**: 1.10
> **状态**: active
> **更新日期**: 2026-06-12

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## E2E.P0.001 默认首页与五入口 Shell

- [x] 创建场景目录 `test/scenarios/e2e/p0-001-default-home-shell/`
- [x] 准备测试数据：未登录状态、无保存 route、默认 runtime config
- [x] 实现 setup / trigger / verify / cleanup；verify 必须断言 Home、五个一级入口、单一登录入口、显示控制可见，welcome、注册入口、独立 `voice`、Growth / Mistakes / Drill 旧入口不可见
- [x] 执行并通过场景验证
- [x] 记录验证证据
<!-- evidence: .test-output/e2e/p0-001-default-home-shell/trigger.log (1 vitest test passed; verify.sh: no legacy entry leaked) -->

## E2E.P0.004 App Shell 中英语言切换

- [x] 创建场景目录 `test/scenarios/e2e/p0-004-app-shell-language-switch/`
- [x] 准备测试数据：可归一为中文的浏览器 locale、未登录 `/me`、可触发语言切换的 TopBar 与 D1 shell route 集
- [x] 实现 setup / trigger / verify / cleanup；verify 必须断言语言切换控件是 TopBar language dropdown，切换到 English 后 TopBar、单一登录入口、用户菜单、auth/profile/settings/placeholder shell 文案为英文，route/testid/params 未被 locale 改写，generated client 请求包含 `Accept-Language`，runtime locale 与登录态不覆盖前端语言设置
- [x] 执行并通过场景验证
- [x] 记录验证证据
<!-- evidence: .test-output/e2e/p0-004-app-shell-language-switch/trigger.log (1 vitest test passed; verify.sh: language dropdown + English copy + Accept-Language evidence present; legacy/prototype leak gates clean) -->


## E2E.P0.002 登录打断后恢复原业务动作

- [x] 创建场景目录 `test/scenarios/e2e/p0-002-auth-pending-action-resume/`
- [x] 准备测试数据：未登录用户、workspace plan context、`verifyAuthEmailChallenge` / `getMe(authenticated)` mock auth 成功响应
- [x] 实现 setup / trigger / verify / cleanup；verify 必须断言登录后恢复 `practice` 且 planId / targetJobId / jdId / resumeVersionId / roundId 未丢失
- [x] 执行并通过场景验证
- [x] 记录验证证据
<!-- evidence: .test-output/e2e/p0-002-auth-pending-action-resume/trigger.log (1 vitest test passed; verify.sh: legacy testid + ui-design/src/data leak gates clean) -->

## E2E.P0.032 Dev mock 登录态菜单与退出闭环

- [x] 创建场景目录 `test/scenarios/e2e/p0-032-dev-mock-auth-state-and-user-menu/`
- [x] 准备测试数据：默认 dev mock 非登录态、passwordless verify 成功、logout 成功、`getMe` authenticated / unauthenticated 状态切换
- [x] 实现 setup / trigger / verify / cleanup；verify 必须断言默认非登录态、登录后头像 chip + dropdown、profile/settings/logout 分流、logout 后非登录态、旧 inline 用户菜单、注册按钮和静态 authenticated default 负向约束
- [x] 执行并通过场景验证
- [x] 记录验证证据
<!-- evidence: .test-output/e2e/p0-032-dev-mock-auth-state-and-user-menu/trigger.log (1 vitest test passed; verify.sh: dev mock auth state + avatar dropdown + logout evidence present; legacy/prototype leak gates clean) -->

## Phase 7: Historical real passwordless mail-link remediation

- [x] 覆盖 `startAuthEmailChallenge` 真实 `202 Accepted` 空响应：generated client 不抛 JSON parse error，登录/注册提交后进入 `auth_verify`
  <!-- verified: 2026-05-27 method=focused-tests evidence="generatedClient.test.ts + AppAuthDispatch.test.tsx PASS; real 202 empty body no longer triggers JSON parse failure" -->
- [x] 覆盖 `auth_verify?token=...` 邮件 callback：自动调用 `verifyAuthEmailChallenge`，成功后 replace 到 pending route 或 Home，URL 不保留 token
  <!-- verified: 2026-05-27 method=focused-tests evidence="routeUrl.test.ts + AuthScreens.test.tsx + AppAuthDispatch.test.tsx PASS; token query is auth_verify-only and scrubbed after verify" -->
- [x] 覆盖本地 Mailpit handoff：`EMAIL_VERIFY_BASE_URL` 指向 frontend callback，backend dev CORS origin 从该 URL 派生，frontend real mode 显式配置 `VITE_EI_API_BASE_URL`
  <!-- verified: 2026-05-27 method=backend-config-doc-gates evidence="backend SMTP writer/cmd API tests, make lint-config, make lint-mock-contract, make docs-check, P0.100 script bash -n, and git diff --check PASS" -->

## Phase 8: Email-code auth and display-name remediation

- [x] 覆盖旧 register/login purpose + displayName pass-through（历史完成项，已由 Phase 9 单入口语义取代）
- [x] 覆盖 6 位 code verify UI：用户只能输入最多 6 位数字 code，提交后 generated client 仍调用 `verifyAuthEmailChallenge?token=<code>`
- [x] 覆盖 TopBar fallback：缺 displayName / emailMasked 时展示中性 fallback，不展示 prototype 样例身份

## Phase 9: Unified email login and first-login profile setup

- [x] 更新 `E2E.P0.101` 场景目录与索引说明，使场景名称表达 single-entry login + profile setup，而不是 register-then-login
- [x] 准备测试数据：唯一新邮箱、资料未补全账号状态、第二 browser context / 无 cookie context、资料补全 displayName、Mailpit code-only 邮件、real frontend/backend API base、session cookie jar
- [x] 实现 setup / trigger / verify / cleanup：从单一登录入口提交新邮箱 -> Mailpit code -> `auth_verify` 手动输入 code -> 进入 `auth_profile_setup` -> 刷新仍停留 -> 关闭/换浏览器后同邮箱重新登录仍停留 -> 提交 displayName + acceptedTerms -> `/me.profileCompletionRequired=false` + TopBar displayName -> logout -> 同邮箱再次登录不再进入资料补全
- [x] 断言 pendingAction 路径：从业务 URL 或操作级 auth gate 登录时，资料补全前不恢复业务动作；资料补全成功后才恢复原 route 和 safe params
- [x] 断言错误/隐私/旧口径负向路径：TopBar 注册按钮、`auth_register` live page、`purpose=signup/login` request body、displayName-before-verify、magic link URL、`/auth/verify?token=`、raw session cookie、`刘哲` / `Liu Zhe` / `liuzhe@example.com` 不出现在 UI、URL、console 或 scenario evidence
- [x] 执行并通过场景验证，记录验证证据
  <!-- verified: 2026-05-28 command="bash test/scenarios/e2e/p0-101-auth-email-code-login-register/scripts/cleanup.sh && bash test/scenarios/e2e/p0-101-auth-email-code-login-register/scripts/setup.sh && bash test/scenarios/e2e/p0-101-auth-email-code-login-register/scripts/trigger.sh && bash test/scenarios/e2e/p0-101-auth-email-code-login-register/scripts/verify.sh && bash test/scenarios/e2e/p0-101-auth-email-code-login-register/scripts/cleanup.sh" evidence="profile-required gates PASS refresh=profile-setup deepLink=profile-setup crossBrowser=profile-setup logoutRelogin=profile-setup authStartBodyKeys=email authRegisterLivePage=absent topbarRegister=absent" -->

## E2E.P0.102 未登录首页与面试业务路由登录前置

- [x] 创建场景目录 `test/scenarios/e2e/p0-102-auth-gated-interview-routes/`
- [x] 准备测试数据：未登录 `/me`、auth loading probe、Home target job fixture spy、业务 route safe params、后端无 cookie request set
- [x] 实现 setup / trigger / verify / cleanup；verify 必须断言 Home 不展示 Recent mock interviews、不调用 `listTargetJobs`、不显示 raw `AUTH_UNAUTHORIZED`；业务 route 未登录时进入 `auth_login(pendingAction)`；后端 focused tests 证明业务 API 保持 session middleware 保护
- [x] 执行并通过场景验证
- [x] 记录验证证据
  <!-- verified: 2026-05-28 command="bash test/scenarios/e2e/p0-102-auth-gated-interview-routes/scripts/setup.sh && bash test/scenarios/e2e/p0-102-auth-gated-interview-routes/scripts/trigger.sh && bash test/scenarios/e2e/p0-102-auth-gated-interview-routes/scripts/verify.sh && bash test/scenarios/e2e/p0-102-auth-gated-interview-routes/scripts/cleanup.sh" evidence=".test-output/e2e/p0-102-auth-gated-interview-routes/trigger.log; verify.sh PASS; result.json status=passed" -->
