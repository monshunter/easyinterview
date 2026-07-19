# Frontend Shell Auth and Settings BDD Plan

> **版本**: 1.22
> **状态**: completed
> **更新日期**: 2026-07-19

**关联 Plan**: [plan](./plan.md)

## 有真实 E2E owner 的行为

| 场景 | Given | When | Then | 真实 E2E |
|---|---|---|---|---|
| Email-code、资料补全、主题与设置入口 | real frontend、backend 与 Mailpit 已运行 | 用户登录/补全资料，在设置保存主题，退出并重新登录 | shell 读取真实 session/profile/theme；Settings 显示账号字段，主题跨重登恢复；reporter/证据不保存完整邮箱 | `E2E.P0.101`（原地扩展） |

## Domain behavior

| Behavior ID | Given | When | Then | 验证入口 |
|-------------|-------|------|------|----------|
| `BDD.SHELL.AUTH.001` | 用户处于 anonymous、profile-incomplete 或 authenticated 状态 | 访问受保护页面、完成登录/profile setup、恢复 pending action 或切换 shell 设置 | shell 按 route/session 状态渲染并安全恢复；显示设置不持久化业务事实 | `frontend/src/app/AppAuthDispatch.test.tsx` + `frontend/src/app/__tests__/auth-pending-action-resume.test.tsx`，由根 `make test` 承接 |
| `BDD.SHELL.SETTINGS.001` | authenticated runtime 已有 `displayName/email` | 用户点击 TopBar 设置齿轮并选择退出 | 直接进入无 tab Settings，完整显示 runtime email 且不重复 `getMe`；无 `emailMasked` alias；退出进入既有确认页 | TopBar + Settings + App dispatch component/domain tests，由根 `make test` 承接 |
| `BDD.SHELL.SETTINGS.002` | 用户未登录并直开 `settings` | route guard 判定 unauthenticated，用户完成登录/资料补全 | Settings 不提前挂载或调用账号 API；登录后通过 safe pendingAction 回到同一路由 | App auth dispatch/pendingAction domain tests，由根 `make test` 承接 |
| `BDD.SHELL.SETTINGS.DELETE.001` | authenticated 用户以鼠标或键盘在 Settings 打开删除确认 | 取消/Escape、确认、重复点击、请求失败/401 后处理或收到 `202` | dialog focus 正确且取消零副作用；pending 只发一次并不可关闭；recoverable 失败保留 UI/同 key，401 交给 auth probe；成功复用 `refreshAuth()` 重探测 `/me`、提交 unauthenticated 并回 Home；默认 fixture client 复现相同 signed-out transition | Settings component tests + default `createDevMockClient` flow + backend `deleteMe` contract tests，由根 `make test` 承接 |
| `BDD.SHELL.AUTH.LOCALE.001` | 当前显示偏好为中文或英文，受保护 route 的 auth probe 为 loading/error | App 渲染统一 auth route gate 或用户切换语言 | eyebrow/title/body 全部跟随当前 locale；中文无英文硬编码，且业务 screen/API 仍不提前挂载 | App shell language-switch + AppAuthDispatch focused tests，由根 `make test` 承接 |
| `BDD.SHELL.SETTINGS.THEME.001` | authenticated runtime 已从首次 `getMe` 取得服务端确认主题 | 用户打开 Settings、切换预定义色或拖动 custom、保存、切换 route、刷新/重登；服务端也可能拒绝保存 | 页面挂载/route 切换零额外 `getMe`；拖动零请求；Save 恰好一次 `updateMe` 且成功零 follow-up GET并立即同步全局；失败保留草稿、离开恢复确认值；刷新/其他平台首次 `getMe` 恢复账号主题 | Settings/DisplayPreferences/AppRuntime request-count tests + backend update/get contract；根 `make test`，真实主路径原地关联 `E2E.P0.101` |

`E2E.P0.101` 只原地增加真实设置齿轮、账号字段、主题保存与 logout/relogin 恢复主路径；请求次数、失败/离开恢复、pendingAction、导出 501 和账号删除仍由 domain/contract tests 承接，不另建并行场景。

Phase 15 current-run 另以 Chrome extension automation skill 在真实本地前后端页面验证中文 auth gate 与英文切换；该手工 Chrome 证据是本次交付门禁，但不新增或标记 E2E ID。
