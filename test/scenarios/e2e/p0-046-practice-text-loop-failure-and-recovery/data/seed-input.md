# Seed Input — E2E.P0.046

| Source | Operation | Scenario | Notes |
|--------|-----------|----------|-------|
| `openapi/fixtures/PracticeSessions/getPracticeSession.json` | `getPracticeSession` | `default` | running |
| `openapi/fixtures/PracticeSessions/getPracticeSession.json` | `getPracticeSession` | `missing-session` | 404 PracticeSessionLost |
| `openapi/fixtures/PracticeSessions/appendSessionEvent.json` | `appendSessionEvent` | `default` | baseline answer ack |
| `openapi/fixtures/PracticeSessions/appendSessionEvent.json` | `appendSessionEvent` | `ai-timeout` | 502 AI_PROVIDER_TIMEOUT |
| `openapi/fixtures/PracticeSessions/appendSessionEvent.json` | `appendSessionEvent` | `mismatch` | 409 client_event_fingerprint_mismatch |
| `openapi/fixtures/PracticeSessions/appendSessionEvent.json` | `appendSessionEvent` | `hint-conflict` | 409 PRACTICE_SESSION_CONFLICT |
