# E2E.P0.095 expected outcome

- Watchlist UNIQUE constraint: duplicate add returns the first item
  with the same id; row count stays at 1.
- tone derivation: 92 -> ok, 78 -> warn, 45 -> muted.
- IK replay on add / remove returns identical body / status.
- DELETE on absent linked_job_match_id (cross-user included) -> 404
  RESOURCE_NOT_FOUND.
- Saved searches: label / query / filters round-trip but never enter
  log / audit / outbox.
