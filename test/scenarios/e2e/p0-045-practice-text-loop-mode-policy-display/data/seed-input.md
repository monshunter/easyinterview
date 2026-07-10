# Seed Input — E2E.P0.045

| Source | Operation | Scenario | Notes |
|--------|-----------|----------|-------|
| `openapi/fixtures/PracticeSessions/getPracticeSession.json` | `getPracticeSession` | `default` | running session |
| `openapi/fixtures/PracticeSessions/appendSessionEvent.json` | `appendSessionEvent` | `show-hint` | assisted hint flow |
| `openapi/fixtures/PracticeSessions/appendSessionEvent.json` | `appendSessionEvent` | `hint-conflict` | backend hint conflict recovery path |
| `openapi/fixtures/PracticeSessions/appendSessionEvent.json` | `appendSessionEvent` | `pause-resume` | pause / resume disabling |

Route param matrix:

- assisted / out-of-scope strict input × baseline / retry_current_round / next_round
- `practiceGoal` enum values must NOT influence the visibility snapshot.
