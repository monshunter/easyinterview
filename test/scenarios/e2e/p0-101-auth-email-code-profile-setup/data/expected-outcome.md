# Expected Outcome

- `POST /auth/email/start` accepts the login email and returns an empty `202 Accepted` response without frontend JSON parse errors. Request bodies only contain `email`.
- Mailpit receives sign-in code messages for first login, profile-incomplete relogin, logout/relogin, and completed-account relogin.
- Each message contains a 6-digit code only; it does not expose frontend `/auth/verify?token=...` or `http://127.0.0.1:10901/api/v1/auth/email/verify` as a user-clickable link.
- Entering the first code consumes the one-shot challenge, mints a first-party `ei_session`, returns `GET /me.profileCompletionRequired=true`, and lands on `auth_profile_setup`.
- Before profile completion, refresh, business-route deep link, new browser context relogin, and logout/relogin all stay on `auth_profile_setup`.
- Completing displayName `Runtime Verify` with accepted terms calls `PATCH /me updateMe`, returns `profileCompletionRequired=false`, lands on `/`, and exposes one settings gear.
- Settings reuses the bootstrap account context. Ocean / Plum / Forest / Custom are visible；Forest preview sends no request and exposes the contracted light/dark accent matrix；Save sends exactly one PATCH and no follow-up GET；Settings→Home→Settings sends no additional GET；relogin restores Forest.
- Logout is entered from Settings, clears the session through the existing confirmation flow, and never calls `DELETE /me`.
- Later login with the same email signs in the same completed account and does not show `auth_profile_setup`.
- Browser console/page errors and unexpected HTTP >=400 responses are zero.
- Scenario logs redact the complete synthetic email, raw verification codes and session cookie values.
