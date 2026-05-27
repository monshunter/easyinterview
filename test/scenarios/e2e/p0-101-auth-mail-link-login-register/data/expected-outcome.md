# Expected Outcome

- `POST /auth/email/start` accepts the login email and returns an empty `202 Accepted` response without frontend JSON parse errors.
- `POST /auth/email/start` accepts the register email and returns an empty `202 Accepted` response without frontend JSON parse errors.
- Mailpit receives one sign-in message for each scenario email.
- Each message contains a frontend callback URL under `/auth/verify`; no message exposes `http://127.0.0.1:8080/api/v1/auth/email/verify` as the user-clickable link.
- Opening each frontend callback consumes the one-shot token, mints a first-party `ei_session`, removes `token=` from the browser URL, and lands on `/`.
- `GET /me` returns 200 after each flow.
- Browser console/page errors and HTTP >=400 responses are zero for both flows.
- Scenario logs redact magic-link tokens and session cookie values.
