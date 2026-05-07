# Frontend Shell BDD Checklist

> **版本**: 1.4
> **状态**: active
> **更新日期**: 2026-05-07

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## E2E.P0.001 默认首页与五入口 Shell

- [x] 创建场景目录 `test/scenarios/e2e/p0-001-default-home-shell/`
- [x] 准备测试数据：未登录状态、无保存 route、默认 runtime config
- [x] 实现 setup / trigger / verify / cleanup；verify 必须断言 Home、五个一级入口、登录/注册、显示控制可见，welcome、独立 `voice`、Growth / Mistakes / Drill 旧入口不可见
- [x] 执行并通过场景验证
- [x] 记录验证证据
<!-- evidence: .test-output/e2e/p0-001-default-home-shell/trigger.log (1 vitest test passed; verify.sh: no legacy entry leaked) -->

## E2E.P0.004 App Shell 中英语言切换

- [x] 创建场景目录 `test/scenarios/e2e/p0-004-app-shell-language-switch/`
- [x] 准备测试数据：可归一为中文的浏览器 locale、未登录 `/me`、可触发语言切换的 TopBar 与 D1 shell route 集
- [x] 实现 setup / trigger / verify / cleanup；verify 必须断言语言切换控件是 TopBar 下拉框，切换到 English 后 TopBar、登录/注册、用户菜单、auth/profile/settings/placeholder shell 文案为英文，route/testid/params 未被 locale 改写，generated client 请求包含 `Accept-Language`，runtime locale 与登录态不覆盖前端语言设置
- [x] 执行并通过场景验证
- [x] 记录验证证据
<!-- evidence: .test-output/e2e/p0-004-app-shell-language-switch/trigger.log (1 vitest test passed; verify.sh: English copy + Accept-Language evidence present; legacy/prototype leak gates clean) -->


## E2E.P0.002 登录打断后恢复原业务动作

- [x] 创建场景目录 `test/scenarios/e2e/p0-002-auth-pending-action-resume/`
- [x] 准备测试数据：未登录用户、workspace plan context、`verifyAuthEmailChallenge` / `getMe(authenticated)` mock auth 成功响应
- [x] 实现 setup / trigger / verify / cleanup；verify 必须断言登录后恢复 `practice` 且 planId / targetJobId / jdId / resumeVersionId / roundId 未丢失
- [x] 执行并通过场景验证
- [x] 记录验证证据
<!-- evidence: .test-output/e2e/p0-002-auth-pending-action-resume/trigger.log (1 vitest test passed; verify.sh: legacy testid + ui-design/src/data leak gates clean) -->
