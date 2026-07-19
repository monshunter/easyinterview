# Frontend Shell Auth and Settings BDD Checklist

> **版本**: 1.23
> **状态**: active
> **更新日期**: 2026-07-19

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

> 既有已勾选静态 handoff 是历史证据；下方 Settings 行为与 `E2E.P0.101` 原地扩展记录当前完成合同。

## `BDD.SHELL.AUTH.001` Shell auth 与安全恢复

- [x] Owner behavior tests 覆盖 auth loading、profile-incomplete、pending action、protected route、settings 与 logout recovery。
- [x] 根 `make test` 已执行对应 Vitest；该结果是代码层行为证据，不是 E2E PASS。

## `E2E.P0.101` 静态 handoff

- [x] 仅关联真实 email-code、session 与 profile-completion 行为，不承接通用 shell/pending/settings。
- [x] 不引用已删除场景目录、wrapper 或历史 PASS。

以上旧条目仅为静态关联审计；真实运行状态只由 `e2e-scenarios-p0/001` 维护。本轮扩展后的当前 run `e2e-p0-101-20260715114513-19516` 已 PASS，静态 INDEX 生命周期状态仍按场景契约保持 `Ready`。

## `BDD.SHELL.SETTINGS.001` 单一入口与真实账号字段

- [x] TopBar/Settings domain tests 证明设置齿轮直达、无账号 dropdown/tab、runtime `displayName/email` 完整回填、`emailMasked` 零引用、零额外 `getMe` 和既有 logout 路由。
- [x] `E2E.P0.101` 静态资产原地扩展设置入口、真实账号字段和 logout；当前真实环境 run `e2e-p0-101-20260715114513-19516` PASS。
- [x] P0.101 code-level evidence gate 证明 raw-value equality 失败与 reporter 输出落盘都不会包含当前 run 邮箱原文或 URL-encoded 等价值。

## `BDD.SHELL.SETTINGS.002` 未登录保护

- [x] Domain tests 覆盖未登录直开 Settings、业务 screen 不挂载、登录/profile setup 后 safe pendingAction 恢复。

## `BDD.SHELL.SETTINGS.DELETE.001` 账号删除状态机

- [x] Component/backend contract tests 覆盖 dialog focus/Escape/取消零副作用、pending 禁止关闭/去重、recoverable 失败恢复/同 key 重试、`401` auth probe、`202` 后 `refreshAuth()` 重探测 `/me` 并提交 unauthenticated，以及 probe 网络错误的 honest auth error；不把该破坏性路径加入共享登录 E2E。
- [x] Default `createDevMockClient` regression 覆盖 verify/profile/deleteMe 后的 `getMe` 401，证明 fixture preview 与真实 backend 的删除后 auth transition 一致。

## `BDD.SHELL.AUTH.LOCALE.001` Auth route gate 本地化

- [ ] 中文 loading/error gate 的 eyebrow/title/body 无英文硬编码；英文切换使用同一 typed keys。
- [ ] Protected business screen/API 仍不在 auth probe 完成前挂载；focused 与根回归通过。
- [ ] Chrome extension automation skill 在 current-run 真实本地页面验证中文 gate 与英文切换；该证据不标记新的 E2E ID。

## `BDD.SHELL.SETTINGS.THEME.001` 账号级主题与请求预算

- [x] Runtime/Settings tests 证明 bootstrap/auth recovery 的 `getMe` 提供确认主题，Settings/普通 route/Practice 切换均零额外 `/me`。（2026-07-19 frontend focused 45 tests PASS。）
- [x] Appearance tests 证明预定义主题/custom hue/chroma 本地预览零网络，Save 恰好一次 `updateMe`，成功直接更新 runtime/display 且无 follow-up GET。（2026-07-19 frontend focused 45 tests PASS。）
- [x] 失败、重试、离开未保存页面、迟到响应和 invalid server projection 均 fail closed，不污染最近一次确认主题或新的认证状态。（2026-07-19 新增 race/projection/atomicity tests 后 focused PASS。）
- [x] Backend/migration/OpenAPI tests 证明 profile-only/theme-only/combined 原子更新、约束和 `getMe` round-trip。
- [x] `E2E.P0.101` 原地覆盖真实主题保存、logout/relogin 恢复；current run `e2e-p0-101-20260719082610-75505` PASS（themeSavePatch=1、settingsMountedGetMe=0、settingsRouteSwitchGetMe=0、themeRelogin=plum、cleanup PASS）。
