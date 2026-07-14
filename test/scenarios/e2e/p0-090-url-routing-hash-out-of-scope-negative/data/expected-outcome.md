# Expected Outcome

| 维度 | 期望 |
|------|------|
| `#route=reports&targetJobId=...` | URL 重写为 `/reports?targetJobId=01918fa0-0000-7000-8000-000000002000`；`section=reports` / `reportId` / `status` / `roundId` 被过滤；`reports-screen` 渲染 |
| Reports TopBar negative | `app-shell-topbar` 可见，但 `topbar-nav-reports` 不存在 |
| legacy Parse report params | URL 重写为 `/parse?targetJobId=tj-1`；旧报告参数全部被过滤且不恢复嵌入式报告区 |
| `#route=home` 启动 | `home-hero-label` 渲染；URL 重写为 `/`；`location.hash` 为空 |
| `#route=workspace&targetJobId=tj-1&...` | URL 重写为 `/workspace?targetJobId=tj-1`；`resumeId` / `planId` / `autoStartPractice` / unknown / raw / sensitive params 被过滤；Workspace 只读规划详情渲染 |
| Workspace detail runtime | 只调用一次 `getTargetJob`；不出现 Parse loading animation；不触发 `importTargetJob` 或 route-side polling |
| `#route=practice&mode=phone&modality=phone&sessionId=...` | URL 重写为仅保留 `sessionId` 的 `/practice?...`；显示连续聊天且电话按钮 disabled |
| `#route=practice&mode=voice&modality=voice&sessionId=...` | URL 重写为仅保留 `sessionId` 的 `/practice?...`；voice 参数被过滤，不渲染电话 surface |
| `#route=voice` | URL 重写为 `/`；`home-hero-label` 渲染（独立 voice route 不 materialize） |
| 12 + 1 个范围外 alias | 全部 hash 启动后 URL 落到对应保留 canonical path（见 plan §4.2）；`debrief` / `debrief_full` / `profile` 均落到 `/` |
| `/totally-unknown?foo=bar` | 渲染 `home-hero-label`，不崩溃 |
| `/voice?mode=voice` | 渲染 `home-hero-label`；canonical path 不暴露 voice |
| `ROUTE_TO_PATH` | 不包含 `/voice` / `/welcome` / `/growth` / `/plan` / `/mistakes` / `/drill` / `/followup` / `/experiences` / `/star` / `/onboarding` / `/debrief` / `/profile` |
| `normalizeRouteName(alias)` | 对每个范围外 alias 返回的不再是范围外 alias 本身 |
| `isCanonicalFrontendPath` | 对 known `/reports?targetJobId=<uuid>` 与每个保留 canonical path 返回 true；对 `/api/*` / `/openapi/*` / `/health` / `/assets/*` / 任意 `*.json` / `*.html` 返回 false |

| 反向断言 | 含义 |
|----------|------|
| `routeUrl.ROUTE_TO_PATH` 文件中不出现 `"/voice"` / `"/welcome"` / `"/debrief"` / `"/profile"` 等 out-of-scope path 字面值 | 禁止 alias 通过 typed table 复活 |
| `frontend/src/app/screens/` 下不存在 welcome / growth / mistakes / drill / followup / experiences / star / onboarding / debrief / profile 目录 | out-of-scope 模块零 materialize |
| Workspace hash / direct URL 只保留 `targetJobId` | ready 规划详情不会被 `resumeId` / `planId` / `autoStartPractice` 或 raw/secret 参数驱动，也不会混入 Parse command flow |

证据：`.test-output/e2e/p0-090-url-routing-hash-out-of-scope-negative/trigger.log`
必须出现 source contract `Ran 2 tests` / `OK`、Reports hash、Workspace
target-scoped read-only detail、Parse legacy strip、
known `/reports` fallback 测试标题、`Tests ... passed` 与 `Test Files ... passed` marker。
