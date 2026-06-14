# Seed Input — E2E.P0.044

| Source | Operation | Scenario | Notes |
|--------|-----------|----------|-------|
| `openapi/fixtures/PracticeSessions/getPracticeSession.json` | `getPracticeSession` | `default` | running session, turnIndex=1 |
| `openapi/fixtures/PracticeSessions/appendSessionEvent.json` | `appendSessionEvent` | `default` | `assistantAction.type=ask_follow_up` |

Route params: `sessionId / planId / targetJobId / jdId / resumeId / roundId / mode=text / modality=text / practiceMode=assisted / practiceGoal=baseline / hintUsed=false / hintCount=0`.

Auth state: signed-in (Vitest jsdom uses fixture-backed `getMe=authenticated`).
