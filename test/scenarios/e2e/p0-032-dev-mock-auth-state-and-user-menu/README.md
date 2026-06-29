# E2E.P0.032 Dev Mock Auth State And User Menu

> **场景 ID**: E2E.P0.032
> **执行方式**: automated
> **隔离级别**: in-process (vitest jsdom)
> **parallel-safe**: No
> **状态**: Ready

## 1 Given

用户在 Vite dev 默认 fixture-backed mock App 中打开首页，初始没有 session。
dev mock transport 使用 OpenAPI Auth fixtures，但必须维护本地 session 状态，
而不是把 `getMe` 的 `authenticated` fixture 当作固定默认登录态。

## 2 When

场景在 jsdom 中渲染 `<App client={createDevMockClient()} />`，模拟用户完成
passwordless mock 登录，打开 TopBar 头像菜单，进入 settings，
再从头像菜单执行退出登录确认。

## 3 Then

- 默认首屏是非登录态，`topbar-user-area` 的 `data-signed-in` 为 `false`，
  且显示 `登录` / `注册`。
- 登录成功后 TopBar 显示 `ui-design/src/app.jsx` 对齐的头像 chip；
  settings/logout 入口只在 dropdown 打开后出现。
- 头像 dropdown 展示用户姓名、脱敏邮箱、设置与隐私、退出登录，
  且 settings/logout 均能从 dropdown 正确分流。
- logout 确认后 `/me` 回到 unauthenticated，TopBar 回到登录 / 注册。
- 旧 inline 三按钮结构不回流，登录态菜单不会读取 `ui-design/src/data.jsx`。

## 4 执行

```bash
./test/scenarios/e2e/p0-032-dev-mock-auth-state-and-user-menu/scripts/setup.sh
./test/scenarios/e2e/p0-032-dev-mock-auth-state-and-user-menu/scripts/trigger.sh
./test/scenarios/e2e/p0-032-dev-mock-auth-state-and-user-menu/scripts/verify.sh
./test/scenarios/e2e/p0-032-dev-mock-auth-state-and-user-menu/scripts/cleanup.sh
```

## 5 污染控制

场景在 vitest + jsdom 中运行，不写共享数据库，不启动 Kind cluster；trigger.sh
仅产生 `.test-output/e2e/p0-032-dev-mock-auth-state-and-user-menu/trigger.log`
作为验证证据，cleanup.sh 移除 setup marker，保留日志。
