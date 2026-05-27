# Account Material

> Owner: `e2e-scenarios-p0/002-manual-uat-real-provider-full-funnel`

## UAT Account

| Field | Value |
|-------|-------|
| Email | `manual-uat-full-funnel@example.test` |
| Local mailbox | `http://127.0.0.1:8025` (Mailpit Web UI) |
| Cookie name | `ei_session` |

The account is created through the real passwordless auth flow. Do not create a session by writing the database directly for this hybrid UAT.

## Sign-In Steps

1. Start the dev stack and backend with `EMAIL_PROVIDER=mailpit`.
2. Open the frontend in real mode.
3. Submit `manual-uat-full-funnel@example.test` on the login screen or an auth gate.
4. Open Mailpit at `http://127.0.0.1:8025`.
5. Open the latest message to this UAT address.
6. Click the `http://127.0.0.1:8080/api/v1/auth/email/verify?token=...` magic link, then return to the frontend and refresh.

The local mailbox and synthetic address remove the need for a real external email service or a real mailbox account. The flow still exercises `startAuthEmailChallenge`, `email_dispatch`, Mailpit SMTP delivery, `verifyAuthEmailChallenge`, and first-party session cookie minting.

## Expiry And Rerun

- Magic links expire per backend-auth challenge TTL. If expired, submit the email again and use the newest Mailpit message.
- If auth rate limiting is hit, wait at least one auth rate-limit window before retrying.
- Reusing the same synthetic email is expected; do not change the UAT email unless this material and the checklist are updated together.
- Do not commit screenshots or terminal output that include the magic-link token or `ei_session` cookie value.

## Cookie Check

After clicking the magic link, DevTools -> Application -> Cookies -> `http://127.0.0.1` should show:

| Field | Value |
|-------|-------|
| Name | `ei_session` |
| Value | generated locally by backend; do not copy into tracked files |
| Domain | `127.0.0.1` |
| Path | `/` |

Refresh the frontend. The TopBar should show the UAT account.

## Cleanup

Default cleanup uses the product privacy-delete path for this exact UAT account instead of wiping the dev database. Run it from the same browser session or copy the cookie value only into an untracked local shell:

```bash
SESSION_COOKIE_VALUE='<paste local ei_session value; do not commit>'

curl -i -X DELETE http://127.0.0.1:8080/api/v1/me \
  -H "Cookie: ei_session=${SESSION_COOKIE_VALUE}" \
  -H "Idempotency-Key: manual-uat-full-funnel-cleanup-$(date +%Y%m%d%H%M%S)"
```

Keep the backend process running after the request so the in-process `privacy_delete` runner can finish. Record the returned privacy request / job ids in the evidence file if cleanup is executed.

If the user explicitly asks to preserve the scene, keep the dev DB rows and only log out locally after copying non-secret evidence. Mailpit messages are local dev artifacts and can be cleared from the Mailpit UI.

Never use `make dev-reset`, broad `DELETE FROM users`, direct session table writes, or a whole-database wipe as the default cleanup for this UAT account.
