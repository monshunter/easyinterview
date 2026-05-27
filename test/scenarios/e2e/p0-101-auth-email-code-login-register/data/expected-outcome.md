# Expected Outcome

- `POST /auth/email/start` accepts the signup email + display name and returns an empty `202 Accepted` response without frontend JSON parse errors.
- Mailpit receives one sign-in code message for the signup attempt.
- The signup message contains a 6-digit code only; it does not expose frontend `/auth/verify?token=...` or `http://127.0.0.1:8080/api/v1/auth/email/verify` as a user-clickable link.
- Entering the signup code consumes the one-shot challenge, mints a first-party `ei_session`, lands on `/`, and displays `Runtime Verify`.
- `POST /auth/email/start` accepts the same email for login and sends a new 6-digit code.
- Entering the login code signs in the same user; `GET /me` returns 200 and the display name remains `Runtime Verify`.
- Repeating signup with the same email and `Runtime Duplicate` display name is rejected by `POST /auth/email/start`, stays on `/auth/register`, sends no new code, and does not create a new user or overwrite the original display name.
- Browser console/page errors and unexpected HTTP >=400 responses are zero.
- Scenario logs redact raw verification codes and session cookie values.
