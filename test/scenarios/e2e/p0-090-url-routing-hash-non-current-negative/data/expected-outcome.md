# Expected Outcome

| 维度 | 期望 |
|------|------|
| `#route=home` 启动 | `home-hero-label` 渲染；URL 重写为 `/`；`location.hash` 为空 |
| `#route=workspace&targetJobId=tj-1` | URL 重写为 `/workspace?targetJobId=tj-1`；`workspace-empty` 渲染 |
| `#route=practice&mode=voice&modality=voice&sessionId=...` | URL 重写为 `/practice?modality=voice&mode=voice&sessionId=...`；chrome 隐藏；当前 phone waveform 渲染 |
| `#route=voice` | URL 重写为 `/`；`home-hero-label` 渲染（独立 voice route 不 materialize） |
| 12 + 1 个非当前 alias | 全部 hash 启动后 URL 落到对应保留 canonical path（见 plan §4.2）；`debrief` / `debrief_full` / `profile` 均落到 `/` |
| `/totally-unknown?foo=bar` | 渲染 `home-hero-label`，不崩溃 |
| `/voice?mode=voice` | 渲染 `home-hero-label`；canonical path 不暴露 voice |
| `ROUTE_TO_PATH` | 不包含 `/voice` / `/welcome` / `/growth` / `/plan` / `/mistakes` / `/drill` / `/followup` / `/experiences` / `/star` / `/onboarding` / `/debrief` / `/profile` |
| `normalizeRouteName(alias)` | 对每个非当前 alias 返回的不再是非当前 alias 本身 |
| `isCanonicalFrontendPath` | 对每个保留 canonical path 返回 true；对 `/api/*` / `/openapi/*` / `/health` / `/assets/*` / 任意 `*.json` / `*.html` 返回 false |

| 反向断言 | 含义 |
|----------|------|
| `routeUrl.ROUTE_TO_PATH` 文件中不出现 `"/voice"` / `"/welcome"` / `"/debrief"` / `"/profile"` 等 non-current path 字面值 | 禁止 alias 通过 typed table 复活 |
| `frontend/src/app/screens/` 下不存在 welcome / growth / mistakes / drill / followup / experiences / star / onboarding / debrief / profile 目录 | non-current 模块零 materialize |

证据：`.test-output/e2e/p0-090-url-routing-hash-non-current-negative/trigger.log`
必须出现 `Tests 10 passed (10)` 与 `Test Files 1 passed (1)` marker。
