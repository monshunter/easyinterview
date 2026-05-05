# Frontend Shell BDD Checklist

> **版本**: 1.0
> **状态**: active
> **更新日期**: 2026-05-05

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## E2E.P0.001 默认首页与五入口 Shell

- [ ] 创建场景目录 `test/scenarios/e2e/p0-001-default-home-shell/`
- [ ] 准备测试数据：未登录状态、无保存 route、默认 runtime config
- [ ] 实现 setup / trigger / verify / cleanup
- [ ] 执行并通过场景验证
- [ ] 记录验证证据

## E2E.P0.002 登录打断后恢复原业务动作

- [ ] 创建场景目录 `test/scenarios/e2e/p0-002-auth-pending-action-resume/`
- [ ] 准备测试数据：未登录用户、workspace plan context、mock auth 成功响应
- [ ] 实现 setup / trigger / verify / cleanup
- [ ] 执行并通过场景验证
- [ ] 记录验证证据
