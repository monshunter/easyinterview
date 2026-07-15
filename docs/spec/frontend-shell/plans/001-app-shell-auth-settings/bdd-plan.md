# Frontend Shell Auth and Settings BDD Plan

> **版本**: 1.18
> **状态**: active
> **更新日期**: 2026-07-15

**关联 Plan**: [plan](./plan.md)

## 有真实 E2E owner 的行为

| 场景 | Given | When | Then | 真实 E2E |
|---|---|---|---|---|
| Email-code、资料补全与设置入口 | real frontend、backend 与 Mailpit 已运行 | 用户获取验证码、登录、补全资料，点击设置齿轮核对账号字段并退出 | shell 读取真实 session/profile，Settings 显示同一账号的姓名与脱敏邮箱，logout 清除 session | `E2E.P0.101`（原地扩展） |

## Domain behavior

| Behavior ID | Given | When | Then | 验证入口 |
|-------------|-------|------|------|----------|
| `BDD.SHELL.AUTH.001` | 用户处于 anonymous、profile-incomplete 或 authenticated 状态 | 访问受保护页面、完成登录/profile setup、恢复 pending action 或切换 shell 设置 | shell 按 route/session 状态渲染并安全恢复；显示设置不持久化业务事实 | `frontend/src/app/AppAuthDispatch.test.tsx` + `frontend/src/app/__tests__/auth-pending-action-resume.test.tsx`，由根 `make test` 承接 |
| `BDD.SHELL.SETTINGS.001` | authenticated runtime 已有 `displayName/emailMasked` | 用户点击 TopBar 设置齿轮并选择退出 | 直接进入无 tab Settings，复用 runtime 账号字段且不重复 `getMe`；退出进入既有确认页 | TopBar + Settings + App dispatch component/domain tests，由根 `make test` 承接 |
| `BDD.SHELL.SETTINGS.002` | 用户未登录并直开 `settings` | route guard 判定 unauthenticated，用户完成登录/资料补全 | Settings 不提前挂载或调用账号 API；登录后通过 safe pendingAction 回到同一路由 | App auth dispatch/pendingAction domain tests，由根 `make test` 承接 |
| `BDD.SHELL.SETTINGS.DELETE.001` | authenticated 用户以鼠标或键盘在 Settings 打开删除确认 | 取消/Escape、确认、重复点击、请求失败/401 后处理或收到 `202` | dialog focus 正确且取消零副作用；pending 只发一次并不可关闭；recoverable 失败保留 UI/同 key，401 交给 auth probe；成功复用 `refreshAuth()` 重探测 `/me`、提交 unauthenticated 并回 Home | Settings component tests + backend `deleteMe` contract tests，由根 `make test` 承接 |

`E2E.P0.101` 只原地增加真实设置齿轮、账号字段与 logout 主路径；pendingAction 通用矩阵、导出 501 和破坏性的账号删除仍由 domain/contract tests 承接，不为设置另建并行场景。
