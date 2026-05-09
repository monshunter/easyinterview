# Expected Outcome

- First `startPracticeSession` returns `502 + AI_PROVIDER_TIMEOUT` with `retryable=true`.
- First failure leaves idempotency status `failed_retryable`, marks one session failed, and emits no outbox.
- Retrying with the same key and same body returns `201 + PracticeSession(status=running)`.
- Retry success marks the idempotency record `succeeded`, emits exactly one `practice.session.started` outbox row, and calls AI twice total.
- Error evidence excludes prompt body, response body, and provider secrets.
