# Expected Outcome

| 维度 | 期望 |
|------|------|
| Workspace direct-open | `workspace-plan-list` 渲染；URL 规范化为无 query 的 `/workspace`；TopBar `workspace` `aria-current="page"` |
| Reports direct-open | `reports-screen` 渲染；URL 规范化为 `/reports?targetJobId=01918fa0-0000-7000-8000-000000002000`；`section=reports` / `reportId` / `status` / `roundId` 被过滤 |
| Reports reload | App 卸载并重新挂载后仍恢复同一 target-scoped Reports URL，且 chrome 可见 |
| Reports missing/invalid target | 当前 history entry 被替换为 `/workspace`；不 `pushState`，browser Back 不会反复回到坏 `/reports` 链接 |
| Practice phone direct-open | `practice-conversation` 出现；`app-shell-topbar` 不在 DOM；`mode=phone&modality=phone` 被过滤且电话按钮 disabled |
| Generating direct-open | `generating-screen` 出现，chrome 隐藏 |
| Report direct-open | `report-dashboard-loading` 出现，chrome 保留；search 只含 `reportId`，状态由 API 响应决定 |
| Resume workshop direct-open | `resume-workshop-screen` 出现；URL 过滤 `tab=rewrites&tailorRunId=...` 并保持 `/resume-versions` |
| Out-of-scope debrief/profile direct-open | `/debrief?...` 与 `/profile` 折回 `/`；不保留 `debriefId` / `debriefJobId` / `targetJobId`；不渲染 `debrief-screen` 或 `route-profile` |
| App-driven navigation | workspace → practice → reports → report 后，back / forward 可恢复 Reports；Reports 始终只保留 `targetJobId` |
| Malformed query | URL canonical 化为 `/workspace`；`targetJobId`、`bogusKey`、`another` 全部被过滤 |
| Hash bootstrap | `#route=workspace` 启动后 URL 重写为 `/workspace`，`location.hash` 与 search 均为空 |

| 反向断言 | 含义 |
|----------|------|
| 不能出现 `route-welcome` / `route-voice` | 范围外 route 名禁止 materialize |
| 不能出现 `topbar-nav-reports` / `mistakes` / `growth` / `drill` / `voice` | Reports 是规划详情页入口的上下文页，不进入 TopBar |
| URL 不能出现 `/voice?` / `/welcome?` / `/growth?` / `/mistakes?` / `/drill?` / `/debrief?` / `/profile?` | canonical path 不可接入范围外入口 |
| URL 不能出现 `ui-design/src/data` | prototype data 不进入 runtime |

证据：`.test-output/e2e/p0-088-url-addressable-routing-canonical/trigger.log`
必须出现 source contract `Ran 2 tests` / `OK`、Reports direct/reload/back-forward
测试标题、`Tests ... passed` 与 `Test Files ... passed` marker。
