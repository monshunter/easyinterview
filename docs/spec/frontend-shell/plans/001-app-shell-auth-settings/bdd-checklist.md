# Frontend Shell BDD Checklist

> **版本**: 1.13
> **状态**: completed
> **更新日期**: 2026-07-07

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## E2E.P0.001 默认首页与三入口 Shell

- [x] 创建场景目录 `test/scenarios/e2e/p0-001-default-home-shell/`
- [x] 准备测试数据：未登录状态、无保存 route、默认 runtime config
- [x] 实现 setup / trigger / verify / cleanup；verify 断言 Home、三个一级入口、单一登录入口、用户区和显示控制可见，unsupported route input 不产生独立页面
- [x] 执行并通过场景验证
- [x] 记录验证证据

## E2E.P0.004 App Shell 中英语言切换

- [x] 创建场景目录 `test/scenarios/e2e/p0-004-app-shell-language-switch/`
- [x] 准备测试数据：可归一的浏览器 locale、未登录 `/me`、可触发语言切换的 TopBar 与 shell route 集
- [x] 实现 setup / trigger / verify / cleanup；verify 断言语言切换控件、英文/中文文案、route/testid/params 稳定性、`Accept-Language` display hint、frontend preference priority 和 `ui-design` 控件结构
- [x] 执行并通过场景验证
- [x] 记录验证证据

## E2E.P0.002 登录打断后恢复原业务动作

- [x] 创建场景目录 `test/scenarios/e2e/p0-002-auth-pending-action-resume/`
- [x] 准备测试数据：未登录用户、workspace plan context、auth verify success、authenticated `/me`
- [x] 实现 setup / trigger / verify / cleanup；verify 断言登录后恢复 `practice` 且 safe params 未丢失
- [x] 执行并通过场景验证
- [x] 记录验证证据

## E2E.P0.032 Dev mock 登录态菜单与退出闭环

- [x] 创建场景目录 `test/scenarios/e2e/p0-032-dev-mock-auth-state-and-user-menu/`
- [x] 准备测试数据：默认 dev mock 非登录态、email-code verify success、logout success、`getMe` 状态切换
- [x] 实现 setup / trigger / verify / cleanup；verify 断言默认非登录态、登录后头像 chip + dropdown、settings/logout 分流、logout 后非登录态和单一登录入口恢复
- [x] 执行并通过场景验证
- [x] 记录验证证据

## E2E.P0.101 Mailpit email-code single-entry login + profile setup

- [x] 更新场景目录与索引说明，使场景名称表达 single-entry email-code login + profile setup
- [x] 准备测试数据：唯一新邮箱、资料未补全账号状态、第二 browser context / no-cookie context、资料补全 displayName、Mailpit code-only 邮件、real frontend/backend API base、session cookie jar
- [x] 实现 setup / trigger / verify / cleanup：单一登录入口提交新邮箱 -> Mailpit code -> `auth_verify` 手动输入 code -> 进入 `auth_profile_setup` -> refresh / relogin 仍停留 -> 提交 displayName + acceptedTerms -> `/me.profileCompletionRequired=false` + TopBar displayName -> logout -> 同邮箱再次登录不再进入资料补全
- [x] 断言 pendingAction 路径：从业务 URL 或操作级 auth gate 登录时，资料补全前不恢复业务动作，资料补全成功后才恢复原 route 和 safe params
- [x] 断言错误、隐私和 URL 边界：验证码、raw session cookie、个人样本标识和敏感正文不出现在 UI、URL、console 或 scenario evidence
- [x] 执行并通过场景验证，记录验证证据

## E2E.P0.102 未登录首页与面试业务路由登录前置

- [x] 创建场景目录 `test/scenarios/e2e/p0-102-auth-gated-interview-routes/`
- [x] 准备测试数据：未登录 `/me`、auth loading probe、Home target job fixture spy、业务 route safe params、后端无 cookie request set
- [x] 实现 setup / trigger / verify / cleanup；verify 断言 Home 不展示账号记录、不调用 `listTargetJobs`、不显示 raw `AUTH_UNAUTHORIZED`；业务 route 未登录时进入 `auth_login(pendingAction)`；backend focused tests 证明业务 API 保持 session middleware 保护
- [x] 执行并通过场景验证
- [x] 记录验证证据
