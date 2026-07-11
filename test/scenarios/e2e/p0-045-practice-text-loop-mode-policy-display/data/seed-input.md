# Seed Input — E2E.P0.045

| Source | Operation | Scenario | Notes |
|--------|-----------|----------|-------|
| `openapi/fixtures/PracticeSessions/getPracticeSession.json` | `getPracticeSession` | `default` | running session |
| `openapi/fixtures/PracticeSessions/appendSessionEvent.json` | `appendSessionEvent` | `show-hint` | assisted hint flow |
| `openapi/fixtures/PracticeSessions/appendSessionEvent.json` | `appendSessionEvent` | `hint-conflict` | backend hint conflict recovery path |
| `openapi/fixtures/PracticeSessions/appendSessionEvent.json` | `appendSessionEvent` | `pause-resume` | pause / resume disabling |
| `openapi/fixtures/TargetJobs/getTargetJob.json` | `getTargetJob` | `default` | server-backed target display fallback |

Route param matrix:

- assisted / out-of-scope strict input × baseline / retry_current_round / next_round
- `practiceGoal` enum values must NOT influence the visibility snapshot.
- Text/phone mode uses one TopBar handset icon; phone mode also exposes one center hang-up control, with no restart/call-ended state.
- VAD, playback settlement, microphone lifecycle and server-backed target display are included in the runner set.
- Session loader adoption and session A-to-B continuity/reset coverage are included in the runner set.
