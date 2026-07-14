# Seed Input

| Route | Canonical URL (fixture) |
|-------|--------------------------|
| workspace hostile params | `/workspace?targetJobId=tj-canonical&resumeId=01918fa0-0000-7000-8000-000000001000&planId=01918fa0-0000-7000-8000-000000004000&autoStartPractice=1`（期望仅保留 `/workspace?targetJobId=tj-canonical`） |
| reports hostile legacy params | `/reports?targetJobId=01918fa0-0000-7000-8000-000000002000&section=reports&reportId=01918fa0-0000-7000-8000-00000000a000&status=ready&roundId=round-hostile`（期望只保留 `targetJobId`） |
| practice (phone) | `/practice?sessionId=01918fa0-0000-7000-8000-000000005000&mode=phone&modality=phone&planId=01918fa0-0000-7000-8000-000000004000` |
| generating | `/generating?sessionId=01918fa0-0000-7000-8000-000000005000&reportId=01918fa0-0000-7000-8000-00000000a000` |
| report hostile route-selected state | `/report?sessionId=01918fa0-0000-7000-8000-000000005000&reportId=01918fa0-0000-7000-8000-00000000a000&reportStatus=failed&errorCode=AI_PROVIDER_TIMEOUT`（期望只保留 `reportId`） |
| resume_versions | `/resume-versions?tab=rewrites&tailorRunId=01918fa0-0000-7000-8000-00000000b000`（期望过滤为 `/resume-versions`） |
| out-of-scope debrief | `/debrief?targetJobId=tj-canonical&debriefId=01918fa0-0000-7000-8000-00000000c000` |
| out-of-scope profile | `/profile` |
| hash bootstrap | `/#route=workspace&targetJobId=tj-canonical` |
| malformed query | `/workspace?bogusKey=42&targetJobId=tj-ok&another=zz` |
| untrusted Reports locator | `/reports`、`/reports?targetJobId=not-a-uuid` |

Reports 使用 fixture-backed 默认登录态；本场景不写共享数据库，URL 路由仅依赖
routeStore、routeUrl 与正式 App screen dispatch。
