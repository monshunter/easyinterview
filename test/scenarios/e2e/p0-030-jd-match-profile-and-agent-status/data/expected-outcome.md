# Expected Outcome

- Profile chip renders avatar initials, SEARCHING AS, skill tags, and profile sources from fixture data.
- The trigger log includes `VITE_EI_API_MODE=real` and
  `jdMatch.realApiMode.test.ts`, proving all 12 JobMatch generated-client
  operations use the real backend base URL with `credentials: "include"`.
- Partial and unauthenticated profile fixtures degrade gracefully without layout breakage.
- AGENT scan badge renders idle, scanning, error, and next-scan-soon variants with relative time labels.
- `getJobMatchProfile` and `getAgentScanStatus` request cadence matches the README contract.
- The screen does not register setInterval, SSE, or WebSocket for scan status.
- Unauthenticated Recommended and Search side-effect actions create `jd_match_action` pending actions; Source and Watchlist chevrons do not.
- Profile, AGENT, and auth-gate Vitest specs all pass.
- Backend E2E.P0.094-P0.097 remain the live API proof for auth, IK,
  persistence, privacy, and AI provenance semantics.
