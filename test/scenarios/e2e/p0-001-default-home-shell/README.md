# E2E.P0.001 Default Home Shell

> **场景 ID**: E2E.P0.001
> **执行方式**: automated
> **隔离级别**: in-process (vitest jsdom)
> **parallel-safe**: No
> **状态**: Ready

## 1 Given

用户没有任何保存的 session 与 route。OpenAPI fixture transport 提供
`getRuntimeConfig` `default` 与 `getMe` `unauthenticated` 场景；前端不会读取
`ui-design/src/data.jsx` 等 prototype 数据源。

## 2 When

场景在 jsdom 中渲染 `<App client={fixtureClient} requestOptions={Prefer:
example=unauthenticated} />`，模拟用户首次打开 App。

## 3 Then

- 默认渲染 Home（`route-home` 出现）。
- TopBar 一级导航严格包含 `home` / `workspace` / `resume_versions` 三项，
  且只包含这三项。
- 全局显示控制（主题色、暗色、语言）可见。
- 未登录 user-area 渲染 `登录` 入口；`topbar-user-area` data-signed-in 为
  `false`，不渲染独立注册入口。
- 非当前入口 `welcome`、`jd_match`、独立 `voice`、`growth` / `mistakes` / `drill`、
  `debrief` / `profile` 在 TopBar 与路由层均不可达。

## 4 执行

```bash
./test/scenarios/e2e/p0-001-default-home-shell/scripts/setup.sh
./test/scenarios/e2e/p0-001-default-home-shell/scripts/trigger.sh
./test/scenarios/e2e/p0-001-default-home-shell/scripts/verify.sh
./test/scenarios/e2e/p0-001-default-home-shell/scripts/cleanup.sh
```

## 5 污染控制

场景在 vitest + jsdom 中运行，不写共享数据库，不启动 Kind cluster；trigger.sh
仅产生 `.test-output/e2e/p0-001-default-home-shell/trigger.log` 作为验证证据，
cleanup.sh 删除 setup marker，保留日志。
