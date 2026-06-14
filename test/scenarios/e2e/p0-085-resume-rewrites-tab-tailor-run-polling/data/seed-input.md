# P0.085 Seed Input

- `ResumeTailor/requestResumeTailor.json` `default` / `idempotency-replay` (both carry IK header).
- `ResumeTailor/getResumeTailorRun.json` four status variants: `default(ready) / queued / generating / failed`.
- Authenticated user; deterministic fake-timer harness (no real wall clock).
