# E2E.P0.004 App Shell Language Switch

> **场景 ID**: E2E.P0.004
> **执行方式**: automated
> **隔离级别**: in-process (vitest jsdom)
> **parallel-safe**: No
> **状态**: Ready

## 1 Given

用户打开默认 App shell。OpenAPI fixture transport 提供 `getRuntimeConfig`
`default` 与 `getMe` `unauthenticated` 场景；runtime config 暴露
`defaultUiLanguage`，但用户仍可在 TopBar 显式切换 UI 语言。

## 2 When

场景在 jsdom 中渲染 `<App client={fixtureClient} requestOptions={Prefer:
example=unauthenticated} />`，模拟用户通过 TopBar 语言下拉框把语言从中文切换到
English，并进入 auth / profile / settings / placeholder shell。

## 3 Then

- TopBar 五个一级导航从中文切换为英文。
- 未登录用户区从 `登录` / `注册` 切换为 `Sign in` / `Register`。
- 已登录用户菜单、auth shell、profile/settings shell 与 placeholder shell 的
  D1 静态文案随 locale 切换。
- RouteName、testid、route params 与业务字段不被 locale 改写。
- D1 generated client 请求携带当前 UI locale 的 `Accept-Language` display hint。
- 语言切换控件是下拉框，不是按钮组或纯状态占位。
- 旧入口 `welcome`、独立 `voice`、`growth` / `mistakes` / `drill`
  仍不可达，且场景不读取 `ui-design/src/data.jsx`。

## 4 执行

```bash
./test/scenarios/e2e/p0-004-app-shell-language-switch/scripts/setup.sh
./test/scenarios/e2e/p0-004-app-shell-language-switch/scripts/trigger.sh
./test/scenarios/e2e/p0-004-app-shell-language-switch/scripts/verify.sh
./test/scenarios/e2e/p0-004-app-shell-language-switch/scripts/cleanup.sh
```

## 5 污染控制

场景在 vitest + jsdom 中运行，不写共享数据库，不启动 Kind cluster；trigger.sh
仅产生 `.test-output/e2e/p0-004-app-shell-language-switch/trigger.log` 作为验证证据，
cleanup.sh 移除 setup marker，保留日志。
