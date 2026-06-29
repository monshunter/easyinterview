# Seed Input — E2E.P0.045

| Source | Operation | Scenario | Notes |
|--------|-----------|----------|-------|
| `openapi/fixtures/PracticeSessions/getPracticeSession.json` | `getPracticeSession` | `default` | running session |
| `openapi/fixtures/PracticeSessions/appendSessionEvent.json` | `appendSessionEvent` | `show-hint` | fixture-only assisted hint flow until backend-practice/003 |
| `openapi/fixtures/PracticeSessions/appendSessionEvent.json` | `appendSessionEvent` | `hint-strict-conflict` | strict-mode hint defensive 409 path |
| `openapi/fixtures/PracticeSessions/appendSessionEvent.json` | `appendSessionEvent` | `turn-skipped` | turn skip annotation |
| `openapi/fixtures/PracticeSessions/appendSessionEvent.json` | `appendSessionEvent` | `pause-resume` | pause / resume disabling |

Route param matrix:

- assisted / strict × baseline / retry_current_round / next_round
- `practiceGoal` enum values must NOT influence the visibility snapshot.
