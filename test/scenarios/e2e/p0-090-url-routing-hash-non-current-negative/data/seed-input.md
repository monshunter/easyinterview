# Seed Input

| 输入 | 描述 |
|------|------|
| `/#route=home` | non-current home hash entry |
| `/#route=workspace&targetJobId=tj-1` | non-current workspace hash entry |
| `/#route=practice&mode=voice&modality=voice&sessionId=01918fa0-...` | non-current voice mode hash entry |
| `/#route=voice` | 非当前 voice 独立路由 |
| `/#route=welcome` 等 12 个非当前 alias | welcome / growth / plan / mistakes / drill / followup / experiences / star / onboarding / debrief / debrief_full / profile |
| `/totally-unknown?foo=bar` | 未知 canonical path |
| `/voice?mode=voice` / `/debrief?...` / `/profile` | 非当前 canonical path |
| `ROUTE_TO_PATH` / `FRONTEND_CANONICAL_PATHS` 全量枚举 | 用于 negative grep |
| `/api/health` / `/openapi/openapi.yaml` / `/health` / `/assets/main.js` / `/index.html` / `/workspace.json` | SPA fallback 必须拒绝 |

无网络 fixture 依赖；测试不发送任何 HTTP 请求。
