# Seed Input

| 输入 | 描述 |
|------|------|
| `/#route=home` | out-of-scope home hash entry |
| `/#route=workspace&targetJobId=tj-1` | out-of-scope workspace hash entry |
| `/#route=practice&mode=phone&modality=phone&sessionId=01918fa0-...` | legacy phone params that must be filtered |
| `/#route=practice&mode=voice&modality=voice&sessionId=01918fa0-...` | out-of-scope voice mode values that must be filtered |
| `/#route=voice` | 范围外 voice 独立路由 |
| `/#route=welcome` 等 12 个范围外 alias | welcome / growth / plan / mistakes / drill / followup / experiences / star / onboarding / debrief / debrief_full / profile |
| `/totally-unknown?foo=bar` | 未知 canonical path |
| `/voice?mode=voice` / `/debrief?...` / `/profile` | 范围外 canonical path |
| `ROUTE_TO_PATH` / `FRONTEND_CANONICAL_PATHS` 全量枚举 | 用于 negative grep |
| `/api/health` / `/openapi/openapi.yaml` / `/health` / `/assets/main.js` / `/index.html` / `/workspace.json` | SPA fallback 必须拒绝 |

无网络 fixture 依赖；测试不发送任何 HTTP 请求。
