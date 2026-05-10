# Expected Outcome

- Profile chip renders avatar initials, SEARCHING AS, skill tags, and profile sources from fixture data.
- Partial and unauthenticated profile fixtures degrade gracefully without layout breakage.
- AGENT scan badge renders idle, scanning, error, and next-scan-soon variants with relative time labels.
- `getJobMatchProfile` and `getAgentScanStatus` request cadence matches the README contract.
- The screen does not register setInterval, SSE, or WebSocket for scan status.
- Unauthenticated Recommended and Search side-effect actions create `jd_match_action` pending actions; Source and Watchlist chevrons do not.
- Profile, AGENT, and auth-gate Vitest specs all pass.
