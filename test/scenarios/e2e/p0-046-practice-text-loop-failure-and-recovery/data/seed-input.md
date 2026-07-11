# Seed Input — E2E.P0.046

| Source | Operation | Scenario | Notes |
|--------|-----------|----------|-------|
| `openapi/fixtures/PracticeSessions/getPracticeSession.json` | `getPracticeSession` | `default` | running |
| `openapi/fixtures/PracticeSessions/getPracticeSession.json` | `getPracticeSession` | `missing-session` | 404 PracticeSessionLost |
| `openapi/fixtures/PracticeSessions/appendSessionEvent.json` | `appendSessionEvent` | `default` | baseline answer ack |
| `openapi/fixtures/PracticeSessions/appendSessionEvent.json` | `appendSessionEvent` | `ai-timeout` | provider timeout → 200 `session_wait` with the original asked turn |
| `openapi/fixtures/PracticeSessions/appendSessionEvent.json` | `appendSessionEvent` | `mismatch` | 409 client_event_fingerprint_mismatch |
| `openapi/fixtures/PracticeSessions/appendSessionEvent.json` | `appendSessionEvent` | `hint-conflict` | 409 PRACTICE_SESSION_CONFLICT |
| `openapi/fixtures/PracticeSessions/appendSessionEvent.json` | `appendSessionEvent` | `replay` | exact original successful `ask_follow_up` snapshot |
| `openapi/fixtures/PracticeSessions/createPracticeVoiceTurn.json` | `createPracticeVoiceTurn` | `chat-output-invalid` | localized voice AI_OUTPUT_INVALID recovery |

Backend service fixture returns invalid/wrong-language follow-up output twice so the runner can prove exactly-one repair, `session_wait`, unchanged turn control and no completion outbox.
