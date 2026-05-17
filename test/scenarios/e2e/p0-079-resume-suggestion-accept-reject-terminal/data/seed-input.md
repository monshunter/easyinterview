# E2E.P0.079 Seed Input

- User A owns one ready targeted resume version with a stable `structured_profile`.
- User A owns pending resume-tailor suggestions created by a ready tailor run.
- User B owns a separate targeted version and suggestion to prove cross-user isolation.
- OpenAPI fixtures include:
  - `acceptResumeTailorSuggestion.json`: `default`, `idempotency-replay`, `already-decided-409`
  - `rejectResumeTailorSuggestion.json`: `default`, `idempotency-replay`, `already-decided-409`
