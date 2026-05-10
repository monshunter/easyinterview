# Expected Outcome

- Recommended cards and JD detail render from JobMatch fixtures.
- Confirm interview navigates to `parse` with exactly `source=jd_match` and `sourceJobMatchId`.
- Save and Unsave send unique Idempotency-Key headers and revert optimistic state on fixture-backed 4xx responses.
- Not relevant sends the default reason without `freeNote` and reverts on fixture-backed 4xx responses.
- Source opens only safe source URLs in a new tab with `noopener,noreferrer` and does not require auth.
- Unauthenticated side-effect actions create `jd_match_action` pending actions without query, label, freeNote, or sourceUrl leakage.
- Recommended-tab Vitest specs all pass and privacy red-line assertions remain green.
