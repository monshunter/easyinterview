# Expected Outcome

- `POST /auth/email/start` returns `202` and does not expose account existence.
- Dev mail sink retrieval provides the transient challenge link while persisted delivery metadata remains redacted.
- `GET /auth/email/verify` mints `ei_session`; invalid and repeated token verification return B1 auth error envelope.
- `GET /me` returns masked email and configured user context for an active session; missing, invalid, revoked, or logged-out sessions return B1 auth error envelope.
- `GET /runtime-config` uses the C1 session resolver only for A4 allowlisted user preference projection.
- `POST /auth/logout` is idempotent and clears the session cookie.
- Repeated `DELETE /me` with the same idempotency key returns the same privacy delete handoff and does not create duplicate jobs.
- Scenario evidence contains no token, full URL, plaintext email, session cookie, session id, or secret.
