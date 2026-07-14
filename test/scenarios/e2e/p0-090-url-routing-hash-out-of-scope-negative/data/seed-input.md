# Seed Input

| 输入 | 描述 |
|------|------|
| `/#route=reports&targetJobId=01918fa0-0000-7000-8000-000000002000&section=reports&reportId=rpt-hostile&status=ready&roundId=round-hostile` | Reports hash；期望 canonical `/reports?targetJobId=<uuid>` |
| `/#route=parse&targetJobId=tj-1&section=reports&reportId=rpt-hostile&status=ready&roundId=round-hostile` | 旧 Parse 报告入口参数；期望只保留 `targetJobId` |
| `/#route=home` | canonical home hash entry |
| `/#route=workspace&targetJobId=tj-1&resumeId=rv-private&planId=plan-private&autoStartPractice=1&unknownKey=private&rawText=private&token=secret` | Workspace 只读详情 hash；期望只保留 `targetJobId`，其余业务 authority / unknown / raw / sensitive params 全部过滤 |
| `/#route=practice&mode=phone&modality=phone&sessionId=01918fa0-...` | legacy phone params that must be filtered |
| `/#route=practice&mode=voice&modality=voice&sessionId=01918fa0-...` | out-of-scope voice mode values that must be filtered |
| `/#route=voice` | 范围外 voice 独立路由 |
| `/#route=welcome` 等 12 个范围外 alias | welcome / growth / plan / mistakes / drill / followup / experiences / star / onboarding / debrief / debrief_full / profile |
| `/totally-unknown?foo=bar` | 未知 canonical path |
| `/voice?mode=voice` / `/debrief?...` / `/profile` | 范围外 canonical path |
| `ROUTE_TO_PATH` / `FRONTEND_CANONICAL_PATHS` 全量枚举 | 用于 negative grep |
| `/reports?targetJobId=01918fa0-0000-7000-8000-000000002000` | known SPA fallback 必须接受，但 `topbar-nav-reports` 必须不存在 |
| `/api/health` / `/openapi/openapi.yaml` / `/health` / `/assets/main.js` / `/index.html` / `/workspace.json` | SPA fallback 必须拒绝 |

Reports / Parse / Workspace 使用 fixture-backed App runtime；测试不写共享数据库
或调用外部服务。Workspace detail 只允许 `getTargetJob` read，不复用 Parse 的
导入、动画或轮询流程。
