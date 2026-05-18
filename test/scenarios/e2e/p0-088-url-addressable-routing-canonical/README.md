# E2E.P0.088 URL-Addressable Routing — Canonical Path Deep-Link / Reload / Browser History

> **场景 ID**: E2E.P0.088
> **执行方式**: automated
> **隔离级别**: in-process (vitest jsdom)
> **parallel-safe**: No
> **状态**: Ready

## 1 Given

正式前端已经接入 Plan 004 的 Browser History 路由：`useBrowserRoute` 通过
`window.__EASYINTERVIEW_INITIAL_ROUTE__` > 浏览器 canonical path > `#route=...`
hash adapter > `DEFAULT_ROUTE` 的优先级决定初始 route，并按 canonical 安全参数
allowlist 写入 `pushState` / `replaceState`。

测试在 vitest jsdom 中运行，使用 `window.history.replaceState` 模拟用户直
接打开 canonical URL，并通过 NavigationProvider 触发应用内导航。

## 2 When

- 直接打开 `/workspace?targetJobId=...&resumeVersionId=...&planId=...&autoStartPractice=1`、
  `/practice?mode=voice&modality=voice&sessionId=...`、`/generating?sessionId=...&reportId=...`、
  `/report?sessionId=...&reportId=...&reportStatus=failed&errorCode=AI_PROVIDER_TIMEOUT`、
  `/resume-versions?tab=rewrites&tailorRunId=...`、`/debrief?targetJobId=...&debriefId=...&debriefJobId=...`。
- 通过 App 内导航连续进入 workspace → practice → report，再 back / forward。
- 直接打开带有未知 query (`/workspace?bogusKey=42&targetJobId=tj-ok&another=zz`)。
- 直接打开旧 hash 入口 `/#route=workspace&targetJobId=...`，校验 canonical 重写。

## 3 Then

- 每个 canonical URL 解析到对应 `Route` + safe params；TopBar `aria-current`
  与 `app-shell-topbar` 隐藏行为符合 route catalog。
- `report?reportStatus=failed` 渲染 `report-failure-state`，并保留 errorCode。
- 当前 cross-owner handoff key (`autoStartPractice` / `reportId` /
  `reportStatus` / `tailorRunId` / `debriefId` / `debriefJobId`) 经
  allowlist 过滤后仍保留。
- back / forward 在 workspace / practice / report 之间正常切换且不丢失
  params；practice / generating 保持 chrome 隐藏。
- 未知 query 被 allowlist 过滤；canonical URL 不残留 `bogusKey`、`another`。
- `#route=workspace` hash 启动后 URL 被 `replaceState` 重写为
  `/workspace?targetJobId=...`，`location.hash` 为空。

## 4 执行

```bash
./test/scenarios/e2e/p0-088-url-addressable-routing-canonical/scripts/setup.sh
./test/scenarios/e2e/p0-088-url-addressable-routing-canonical/scripts/trigger.sh
./test/scenarios/e2e/p0-088-url-addressable-routing-canonical/scripts/verify.sh
./test/scenarios/e2e/p0-088-url-addressable-routing-canonical/scripts/cleanup.sh
```

## 5 污染控制

场景在 vitest + jsdom 中运行，不写共享数据库，不启动 Kind cluster；trigger.sh
仅产生 `.test-output/e2e/p0-088-url-addressable-routing-canonical/trigger.log`
作为验证证据，cleanup.sh 移除 setup marker，保留日志。
