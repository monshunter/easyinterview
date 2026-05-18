# Seed Input

| 输入 | 描述 |
|------|------|
| `/#route=home` | 历史 home 入口 |
| `/#route=workspace&targetJobId=tj-1` | 历史 workspace 入口 |
| `/#route=practice&mode=voice&modality=voice&sessionId=01918fa0-...` | 历史 voice mode 入口 |
| `/#route=voice` | 退役 voice 独立路由 |
| `/#route=welcome` 等 9 个退役 alias | welcome / growth / plan / mistakes / drill / followup / experiences / star / onboarding / debrief_full |
| `/totally-unknown?foo=bar` | 未知 canonical path |
| `/voice?mode=voice` | 退役 canonical path |
| `ROUTE_TO_PATH` / `FRONTEND_CANONICAL_PATHS` 全量枚举 | 用于 negative grep |
| `/api/health` / `/openapi/openapi.yaml` / `/health` / `/assets/main.js` / `/index.html` / `/workspace.json` | SPA fallback 必须拒绝 |

无网络 fixture 依赖；测试不发送任何 HTTP 请求。
