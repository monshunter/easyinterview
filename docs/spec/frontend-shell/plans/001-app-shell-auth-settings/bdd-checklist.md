# Frontend Shell Auth and Settings BDD Checklist

> **版本**: 1.18
> **状态**: active
> **更新日期**: 2026-07-15

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

> 既有已勾选静态 handoff 是历史证据；下方 Settings 行为与 `E2E.P0.101` 原地扩展是当前未完成合同。

## `BDD.SHELL.AUTH.001` Shell auth 与安全恢复

- [x] Owner behavior tests 覆盖 auth loading、profile-incomplete、pending action、protected route、settings 与 logout recovery。
- [x] 根 `make test` 已执行对应 Vitest；该结果是代码层行为证据，不是 E2E PASS。

## `E2E.P0.101` 静态 handoff

- [x] 仅关联真实 email-code、session 与 profile-completion 行为，不承接通用 shell/pending/settings。
- [x] 不引用已删除场景目录、wrapper 或历史 PASS。

以上仅为静态关联审计；真实运行状态只由 `e2e-scenarios-p0/001` 维护。本轮未运行 `E2E.P0.101`，当前结果仍为 `Ready`。

## `BDD.SHELL.SETTINGS.001` 单一入口与真实账号字段

- [ ] TopBar/Settings domain tests 证明设置齿轮直达、无账号 dropdown/tab、runtime `displayName/emailMasked` 回填、零额外 `getMe` 和既有 logout 路由。
- [ ] `E2E.P0.101` 静态资产原地扩展设置入口、真实账号字段和 logout；只有显式运行真实环境后才可更新其结果。

## `BDD.SHELL.SETTINGS.002` 未登录保护

- [ ] Domain tests 覆盖未登录直开 Settings、业务 screen 不挂载、登录/profile setup 后 safe pendingAction 恢复。

## `BDD.SHELL.SETTINGS.DELETE.001` 账号删除状态机

- [ ] Component/backend contract tests 覆盖 dialog focus/Escape/取消零副作用、pending 禁止关闭/去重、recoverable 失败恢复/同 key 重试、`401` auth probe、`202` 后 `refreshAuth()` 重探测 `/me` 并提交 unauthenticated，以及 probe 网络错误的 honest auth error；不把该破坏性路径加入共享登录 E2E。
