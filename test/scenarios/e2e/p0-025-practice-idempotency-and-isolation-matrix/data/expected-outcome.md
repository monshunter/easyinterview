# Expected Outcome

- Same user + same key + same fingerprint returns the same session and leaves outbox count unchanged.
- Same user + same key + different fingerprint returns `409 PRACTICE_SESSION_CONFLICT` without leaking the first session ID.
- User B using the same key creates an independent session.
- User A using a different key for an already active plan returns `409 PRACTICE_SESSION_CONFLICT`.
- User B cannot read User A's plan or session and receives not-found envelopes.
