# E2E.P0.078 Expected Outcome

## Failure / Retry Outcomes

- Timeout attempts return retryable `AI_PROVIDER_TIMEOUT`, mark the async job failed, and do not emit `resume.tailor.completed`.
- Invalid output returns terminal `AI_OUTPUT_INVALID`, marks the async job failed, and does not emit `resume.tailor.completed`.
- Retry can move a failed async job back to generating and then ready on a later successful AI response.
- The successful retry writes ready run provenance plus ephemeral suggestions into async job result, and exactly one completed outbox event.
- `ai_task_runs` contains one row per AI attempt, including failed and successful attempts.

## Evidence Outcomes

- `method=cmd-api-http` appears in verify output.
- Trigger log contains no skipped or no-op focused gates.
- Retired vocabulary and privacy negative greps return zero hits for backend resume implementation and scenario evidence.
