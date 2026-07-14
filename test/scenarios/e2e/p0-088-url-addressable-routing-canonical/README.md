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

- 直接打开 `/reports?targetJobId=<uuid>&section=reports&reportId=...&status=ready&roundId=...`，
  再卸载并重新挂载 App 模拟 reload。
- 直接打开缺失或携带 malformed `targetJobId` 的 `/reports`，验证自动安全回退不会制造 browser Back loop。
- 直接打开携带 out-of-scope detail/start params 的 `/workspace?targetJobId=...&resumeId=...&planId=...&autoStartPractice=1`、
  `/practice?mode=phone&modality=phone&sessionId=...`、`/generating?sessionId=...&reportId=...`、
  `/report?sessionId=...&reportId=...&reportStatus=failed&errorCode=AI_PROVIDER_TIMEOUT`、
  `/resume-versions?tab=rewrites&tailorRunId=...`，其中旧报告状态字段只作为应被过滤的 hostile 输入。
- 直接打开范围外 `/debrief?...` 与 `/profile`，验证它们折回首页且不保留范围外参数。
- 通过 App 内导航连续进入 workspace → practice → reports → report，再 back / forward。
- 直接打开带有未知 query (`/workspace?bogusKey=42&targetJobId=tj-ok&another=zz`)。
- 直接打开范围外 hash 入口 `/#route=workspace&targetJobId=...`，校验 canonical 重写。

## 3 Then

- Reports 规范化为 `/reports?targetJobId=<uuid>`，只保留 `targetJobId`；reload、
  App 导航与 back / forward 都恢复同一当前规划上下文，且不出现 `topbar-nav-reports`。
- Reports 缺失或非法 target identity 时以 `replaceState` 进入 `/workspace`；不调用 `pushState`，Back 不会重新进入坏链接。
- 每个 canonical URL 解析到对应 `Route` + safe params；TopBar `aria-current`
  与 `app-shell-topbar` 隐藏行为符合 route catalog。
- workspace 丢弃 `targetJobId` / `resumeId` / `planId` / `autoStartPractice` 并规范化为 query-free list route；generating/report 只保留 `reportId`，丢弃 `sessionId` / `reportStatus` / `errorCode`；简历工作台过滤 `tab` / `tailorRunId`。
- back / forward 在 workspace / practice / reports / report 之间正常切换且不丢失
  params；practice / generating 保持 chrome 隐藏。
- workspace 未知与详情 query 全部被过滤；canonical URL 不保留 `bogusKey`、`another` 或 `targetJobId`。
- `#route=workspace` hash 启动后 URL 被 `replaceState` 重写为
  `/workspace`，`location.hash` 与 search 均为空。
- `/debrief` 与 `/profile` 不再是 canonical path，范围外 deep-link 折回 `/`。

## 4 执行

```bash
./test/scenarios/e2e/p0-088-url-addressable-routing-canonical/scripts/setup.sh
./test/scenarios/e2e/p0-088-url-addressable-routing-canonical/scripts/trigger.sh
./test/scenarios/e2e/p0-088-url-addressable-routing-canonical/scripts/verify.sh
./test/scenarios/e2e/p0-088-url-addressable-routing-canonical/scripts/cleanup.sh
```

## 5 污染控制

场景在 Python source contract + vitest + jsdom 中运行，不写共享数据库；trigger.sh
仅产生 `.test-output/e2e/p0-088-url-addressable-routing-canonical/trigger.log`
作为验证证据，cleanup.sh 删除 setup marker，保留日志。
