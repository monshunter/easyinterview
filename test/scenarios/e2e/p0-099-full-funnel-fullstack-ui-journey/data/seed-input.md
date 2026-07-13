# Synthetic Acceptance Inputs

Use synthetic `.example.test` account data and synthetic interview material.
Run scenario setup first, then create all three isolated practice-session/report
resources after the `setup_at` recorded in `setup.env`:

1. Chinese long-content report whose real provider result is `needs_practice`.
2. English long-content report whose real provider result is `well_prepared`.
3. A distinct real completion handoff captured while its report is honestly
   `generating`.

The tracked scenario does not prescribe report prose. If a generated ready
report lands in the wrong preparedness state, create another synthetic run; do
not edit DB/API status, inject a fixture header, or override the route.

Before trigger, provide the current synthetic browser session cookie value
only through the temporary `P0_099_SESSION_COOKIE` environment variable. Do
not include the `ei_session=` name and do not write the value to a file. Trigger
captures the three live HTTP and read-only PostgreSQL projections before running
focused Vitest, then removes the cookie and database URL from its general
child-process environment. It reads the default `DATABASE_URL` from the local
mode-`0600` dev-stack env; `P0_099_DATABASE_URL` is only a temporary override.
The browser-authored manifest is provisional: trigger atomically replaces its
machine DB/API/digest/count fields from that same live capture.

Do not copy input prose or raw provider output into tracked or `.test-output`
evidence. Persist only opaque references, digests, state summaries, verdict
counts, current-run timestamps, the redacted live-capture projection, canonical
screenshot paths, and the exact no-OCR boolean audit.
