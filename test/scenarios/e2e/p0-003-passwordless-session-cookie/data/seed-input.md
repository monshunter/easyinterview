# Seed Input

- Email: `scenario.user+e2e-p0-003@example.test`
- Invalid token: controlled invalid value used only inside the focused scenario test.
- Challenge delivery: C1 backend-internal background mail dispatcher writes to dev mail sink.
- Session cookie jar: first login for `/me`, `/runtime-config`, logout; second login for repeated `DELETE /me`.
- Delete idempotency key: `delete-key-bdd`.
