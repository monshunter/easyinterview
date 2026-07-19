# Frontend Shell Auth and Settings BDD Plan

> **版本**: 1.26
> **状态**: active
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
| `BDD.SHELL.SETTINGS.THEME.001` | authenticated runtime 已从首次 `getMe` 取得服务端确认主题 | 用户打开 Settings、切换预定义色或拖动 custom、保存、切换 route、刷新/重登；服务端也可能拒绝保存、返回非法投影或在页面离开后迟到返回 | Ocean / Plum / Custom 一级选择器始终可见；hue/chroma 二级编辑器仅在 Custom 激活时于一级下方展示且不遮挡一级，hue 轨道展示完整光谱，chroma 轨道跟随当前 hue 从低彩到高彩，选择预定义主题后隐藏；页面挂载/route 切换零额外 `getMe`；拖动零请求；Save 恰好一次 `updateMe` 且成功零 follow-up GET并立即同步全局；失败保留草稿并可重试，离开恢复确认值；迟到响应不覆盖新 auth/theme；非法投影 fail closed 为 ocean；刷新/其他平台首次 `getMe` 恢复账号主题 | Settings/DisplayPreferences/AppRuntime request-count/race/visual tests + dev-mock/backend update contract；根 `make test`，真实主路径原地关联 `E2E.P0.101` |
| `BDD.SHELL.PAGES.VISUAL.002` | 用户处于登录、验证码、退出或已登录设置页面 | 在 desktop/mobile 查看、键盘操作或触发既有动作 | Auth 三页共享宽幅双栏和主操作卡，Settings 展示 Header 与三张横向功能卡；页面无横向溢出，装饰不污染语义，业务状态机和请求预算不变且没有伪计时/伪成功 | Auth/Settings component + visual + accessibility tests、Phase 15 locale tests 与 current-run Chrome manual acceptance；根 `make test` 承接代码层回归 |
| `BDD.SHELL.TRANSITION.VISUAL.003` | authenticated 用户进入 Practice、Parse、Reports、Generating 或报告上下文 route，业务请求仍 pending | 查看等待反馈、切换显示偏好或使用返回动作 | 共享 TopBar 始终可见且“面试”高亮；等待态复用蓝白画布和对应 SVG variant，只有真实文案/步骤/indeterminate 语义，reduced-motion 与 mobile containment 有效 | Shared transition/TopBar/route component tests + current-run Chrome desktop/mobile manual acceptance；根 `make test` 承接代码层回归 |

`E2E.P0.101` 只原地增加真实设置齿轮、账号字段、主题保存与 logout/relogin 恢复主路径；请求次数、失败/离开恢复、pendingAction、导出 501 和账号删除仍由 domain/contract tests 承接，不另建并行场景。

Phase 15 current-run 另以 Chrome extension automation skill 在真实本地前后端页面验证中文 auth gate 与英文切换；该手工 Chrome 证据是本次交付门禁，但不新增或标记 E2E ID。
