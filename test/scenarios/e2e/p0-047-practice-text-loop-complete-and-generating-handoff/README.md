# E2E.P0.047 complete conversation handoff

Completing the conversation uses an idempotent completion request, creates one report job and one persisted `session_completed` ledger fact, and navigates to generating with stable owner IDs only. Replaying completion must not append a second fact that could advance practice progress twice. No display mode, hint, question, or message content is placed in the URL.
