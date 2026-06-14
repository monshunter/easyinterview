# Seed Input

| Route | Canonical URL (fixture) |
|-------|--------------------------|
| workspace | `/workspace?targetJobId=tj-canonical&resumeId=01918fa0-0000-7000-8000-000000001000&planId=01918fa0-0000-7000-8000-000000004000&autoStartPractice=1` |
| practice (voice) | `/practice?sessionId=01918fa0-0000-7000-8000-000000005000&mode=voice&modality=voice&planId=01918fa0-0000-7000-8000-000000004000` |
| generating | `/generating?sessionId=01918fa0-0000-7000-8000-000000005000&reportId=01918fa0-0000-7000-8000-00000000a000` |
| report (failed) | `/report?sessionId=01918fa0-0000-7000-8000-000000005000&reportId=01918fa0-0000-7000-8000-00000000a000&reportStatus=failed&errorCode=AI_PROVIDER_TIMEOUT` |
| resume_versions | `/resume-versions?tab=rewrites&tailorRunId=01918fa0-0000-7000-8000-00000000b000` |
| debrief | `/debrief?targetJobId=tj-canonical&debriefId=01918fa0-0000-7000-8000-00000000c000&debriefJobId=01918fa0-0000-7000-8000-00000000d000` |
| hash bootstrap | `/#route=workspace&targetJobId=tj-canonical` |
| malformed query | `/workspace?bogusKey=42&targetJobId=tj-ok&another=zz` |

Auth / session 状态：未登录或登录态均不影响本场景（trigger 不需要 fixture
client；URL 路由仅依赖 routeStore + routeUrl）。
