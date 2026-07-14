# E2E.P0.090 URL Routing — Hash Routing + Out-of-scope Route Negative

> **场景 ID**: E2E.P0.090
> **执行方式**: automated
> **隔离级别**: in-process (vitest jsdom)
> **parallel-safe**: No
> **状态**: Ready

## 1 Given

正式前端保留 `parseInitialRouteHash`（`#route=...` adapter）用于 static
preview / Playwright pixel parity / 场景 harness；同时 Plan 004 范围外独立
`voice` 与范围外 alias 入口。SPA 主机回退（`vite preview` + 自定义
`serve-pixel-parity.mjs`）只对 canonical 前端路径返回 `index.html`，对
`/api/*`、`/openapi/*`、`/ui-design/*`、`/health`、`/assets/*` 不发生
swallow。

测试在 vitest + jsdom 中运行；SPA fallback 行为通过对
`scripts/spaFallback.mjs` 的纯函数断言验证（实际 mjs 已被
`serve-pixel-parity.mjs` 加载）。

## 2 When

- 启动 App 时浏览器 URL 为 `/#route=reports&targetJobId=<uuid>&section=reports&reportId=...&status=ready&roundId=...`、
  `/#route=parse&targetJobId=...&section=reports&reportId=...&status=ready&roundId=...`、
  `/#route=home` / `/#route=workspace&...` /
  `/#route=practice&mode=phone&...` / `/#route=practice&mode=voice&...` /
  `/#route=voice` /
  `/#route=welcome` 等范围外入口。
- 启动 App 时浏览器 URL 为 `/totally-unknown?foo=bar` 或 `/voice?mode=voice`。
- 对 `ROUTE_TO_PATH` 与 `FRONTEND_CANONICAL_PATHS` 做静态 negative grep。
- 显式调用 `isCanonicalFrontendPath("/reports?targetJobId=<uuid>")` 验证 known
  `/reports` SPA fallback，同时验证 TopBar 不存在 `topbar-nav-reports`。
- 调用 `isCanonicalFrontendPath` 验证 `/api/*` / `/openapi/*` / `/health` /
  `/assets/*` / 文件请求被拒绝。

## 3 Then

- Reports hash 规范化为只含 `targetJobId` 的 `/reports?targetJobId=<uuid>`，
  chrome 可见但 Reports 不进入 TopBar。
- Parse hash 中旧 `section=reports` / `reportId` / `status` / `roundId` 全部被过滤，
  不恢复嵌入式报告区。
- 每个 hash 启动后 URL 立即被 `replaceState` 重写为 canonical path，
  `location.hash` 为空。
- Legacy `mode/modality` 参数（包括 `phone` 与 `voice`）全部被过滤，不形成任何电话模式入口。
- `#route=voice` / `/voice` 都规范化为 `home`（独立 voice route 永远不
  materialize）。
- 范围外 alias (`welcome` / `growth` / `plan` / `mistakes` / `drill` /
  `followup` / `experiences` / `star` / `onboarding` / `debrief` /
  `debrief_full` / `profile`)
  全部映射到当前保留 canonical path，并且 `normalizeRouteName` 返回的
  不再是范围外 alias。
- `ROUTE_TO_PATH` 与 `FRONTEND_CANONICAL_PATHS` 不包含任何范围外 path。
- SPA fallback 对 known `/reports` 与每个 canonical path 返回 index.html，对 `/api/*` /
  `/openapi/*` / `/health` / `/assets/*` / 任意带扩展名文件请求返回
  null（不 swallow）。

## 4 执行

```bash
./test/scenarios/e2e/p0-090-url-routing-hash-out-of-scope-negative/scripts/setup.sh
./test/scenarios/e2e/p0-090-url-routing-hash-out-of-scope-negative/scripts/trigger.sh
./test/scenarios/e2e/p0-090-url-routing-hash-out-of-scope-negative/scripts/verify.sh
./test/scenarios/e2e/p0-090-url-routing-hash-out-of-scope-negative/scripts/cleanup.sh
```

## 5 污染控制

场景在 Python source contract + vitest + jsdom 中运行，不写共享数据库；trigger.sh
仅产生 `.test-output/e2e/p0-090-url-routing-hash-out-of-scope-negative/trigger.log`
作为验证证据，cleanup.sh 删除 setup marker，保留日志。
