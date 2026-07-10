# Expected Outcome

| 维度 | 期望 |
|------|------|
| Workspace direct-open | `workspace-plan-list` 渲染；URL 规范化为无 query 的 `/workspace`；TopBar `workspace` `aria-current="page"` |
| Practice phone direct-open | `practice-phone-waveform` 出现；`app-shell-topbar` 不在 DOM；URL `/practice` 携带 `mode=phone&modality=phone` |
| Generating direct-open | `generating-screen` 出现，chrome 隐藏 |
| Report failed direct-open | `report-failure-state` 出现，chrome 保留；search 含 `reportStatus=failed&errorCode=AI_PROVIDER_TIMEOUT` |
| Resume workshop direct-open | `resume-workshop-screen` 出现；URL 过滤 `tab=rewrites&tailorRunId=...` 并保持 `/resume-versions` |
| Out-of-scope debrief/profile direct-open | `/debrief?...` 与 `/profile` 折回 `/`；不保留 `debriefId` / `debriefJobId` / `targetJobId`；不渲染 `debrief-screen` 或 `route-profile` |
| App-driven navigation | 三次 `pushState` 后 back 两次 → practice → workspace；forward 两次 → practice → report；chrome 状态正确切换 |
| Malformed query | URL canonical 化为 `/workspace`；`targetJobId`、`bogusKey`、`another` 全部被过滤 |
| Hash bootstrap | `#route=workspace` 启动后 URL 重写为 `/workspace`，`location.hash` 与 search 均为空 |

| 反向断言 | 含义 |
|----------|------|
| 不能出现 `route-welcome` / `route-voice` | 范围外 route 名禁止 materialize |
| 不能出现 `topbar-nav-mistakes` / `growth` / `drill` / `voice` | TopBar 不应承接范围外入口 |
| URL 不能出现 `/voice?` / `/welcome?` / `/growth?` / `/mistakes?` / `/drill?` / `/debrief?` / `/profile?` | canonical path 不可接入范围外入口 |
| URL 不能出现 `ui-design/src/data` | prototype data 不进入 runtime |

证据：`.test-output/e2e/p0-088-url-addressable-routing-canonical/trigger.log`
必须出现 `outOfScopeRouteNegative.test.ts`、`Tests 16 passed (16)` 与 `Test Files 2 passed (2)` marker。
