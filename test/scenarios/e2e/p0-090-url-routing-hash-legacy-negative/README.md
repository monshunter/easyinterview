# E2E.P0.090 URL Routing — Hash Compatibility + Legacy Route Negative

> **场景 ID**: E2E.P0.090
> **执行方式**: automated
> **隔离级别**: in-process (vitest jsdom)
> **parallel-safe**: No
> **状态**: Ready

## 1 Given

正式前端保留 `parseInitialRouteHash`（`#route=...` adapter）用于 static
preview / Playwright pixel parity / 场景 harness；同时 Plan 004 已删除独立
`voice` 与历史 alias 入口。SPA 主机回退（`vite preview` + 自定义
`serve-pixel-parity.mjs`）只对 canonical 前端路径返回 `index.html`，对
`/api/*`、`/openapi/*`、`/ui-design/*`、`/health`、`/assets/*` 不发生
swallow。

测试在 vitest + jsdom 中运行；SPA fallback 行为通过对
`scripts/spaFallback.mjs` 的纯函数断言验证（实际 mjs 已被
`serve-pixel-parity.mjs` 加载）。

## 2 When

- 启动 App 时浏览器 URL 为 `/#route=home` / `/#route=workspace&...` /
  `/#route=practice&mode=voice&...` / `/#route=voice` /
  `/#route=welcome` 等历史 / 退役入口。
- 启动 App 时浏览器 URL 为 `/totally-unknown?foo=bar` 或 `/voice?mode=voice`。
- 对 `ROUTE_TO_PATH` 与 `FRONTEND_CANONICAL_PATHS` 做静态 negative grep。
- 调用 `isCanonicalFrontendPath` 验证 `/api/*` / `/openapi/*` / `/health` /
  `/assets/*` / 文件请求被拒绝。

## 3 Then

- 每个 hash 启动后 URL 立即被 `replaceState` 重写为 canonical path，
  `location.hash` 为空。
- `#route=voice` / `/voice` 都规范化为 `home`（独立 voice route 永远不
  materialize）。
- 退役 alias (`welcome` / `growth` / `plan` / `mistakes` / `drill` /
  `followup` / `experiences` / `star` / `onboarding` / `debrief` /
  `debrief_full` / `profile`)
  全部映射到当前保留 canonical path，并且 `normalizeRouteName` 返回的
  不再是退役 alias。
- `ROUTE_TO_PATH` 与 `FRONTEND_CANONICAL_PATHS` 不包含任何退役 path。
- SPA fallback 对每个 canonical path 返回 index.html，对 `/api/*` /
  `/openapi/*` / `/health` / `/assets/*` / 任意带扩展名文件请求返回
  null（不 swallow）。

## 4 执行

```bash
./test/scenarios/e2e/p0-090-url-routing-hash-legacy-negative/scripts/setup.sh
./test/scenarios/e2e/p0-090-url-routing-hash-legacy-negative/scripts/trigger.sh
./test/scenarios/e2e/p0-090-url-routing-hash-legacy-negative/scripts/verify.sh
./test/scenarios/e2e/p0-090-url-routing-hash-legacy-negative/scripts/cleanup.sh
```

## 5 污染控制

场景在 vitest + jsdom 中运行，不写共享数据库，不启动 Kind cluster；trigger.sh
仅产生 `.test-output/e2e/p0-090-url-routing-hash-legacy-negative/trigger.log`
作为验证证据，cleanup.sh 移除 setup marker，保留日志。
