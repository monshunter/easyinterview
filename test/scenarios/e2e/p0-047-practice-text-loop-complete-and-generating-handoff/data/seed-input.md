# Seed Input — E2E.P0.047

| Source | Operation | Scenario | Notes |
|--------|-----------|----------|-------|
| `openapi/fixtures/PracticeSessions/getPracticeSession.json` | `getPracticeSession` | `default` | running |
| `openapi/fixtures/PracticeSessions/getPracticeSession.json` | `getPracticeSession` | `completing` | status=completing |
| `openapi/fixtures/PracticeSessions/appendSessionEvent.json` | `appendSessionEvent` | `completed` | `assistantAction.type=session_completed` |
| `openapi/fixtures/PracticeSessions/completePracticeSession.json` | `completePracticeSession` | `default` | 202 ReportWithJob |
| `openapi/fixtures/PracticeSessions/completePracticeSession.json` | `completePracticeSession` | `replay` | same Idempotency-Key returns first response |
| `openapi/fixtures/PracticeSessions/completePracticeSession.json` | `completePracticeSession` | `mismatch` | 409 fingerprint mismatch |
| `openapi/fixtures/PracticeSessions/completePracticeSession.json` | `completePracticeSession` | `session-already-completed` | reuse existing report |
