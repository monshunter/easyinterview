# Expected Outcome

- `POST /auth/email/start` accepts the login email and returns an empty `202 Accepted` response without frontend JSON parse errors. Request bodies only contain `email`.
- Mailpit receives sign-in code messages for first login, profile-incomplete relogin, logout/relogin, and completed-account relogin.
- Each message contains a 6-digit code only; it does not expose frontend `/auth/verify?token=...` or `http://127.0.0.1:8080/api/v1/auth/email/verify` as a user-clickable link.
- Entering the first code consumes the one-shot challenge, mints a first-party `ei_session`, returns `GET /me.profileCompletionRequired=true`, and lands on `auth_profile_setup`.
- Before profile completion, refresh, business-route deep link, new browser context relogin, and logout/relogin all stay on `auth_profile_setup`.
- Completing displayName `Runtime Verify` with accepted terms calls `PATCH /me`, returns `profileCompletionRequired=false`, lands on `/`, and displays `Runtime Verify`.
- Later login with the same email signs in the same completed account and does not show `auth_profile_setup`.
- Browser console/page errors and unexpected HTTP >=400 responses are zero.
- Scenario logs redact raw verification codes and session cookie values.
