# Expected Outcome

- Confirm interview navigates to `parse` with exactly `source=jd_match` and the selected `sourceJobMatchId`.
- The trigger log includes `VITE_EI_API_MODE=real` and
  `jdMatch.realApiMode.test.ts`, proving all 12 JobMatch generated-client
  operations use the real backend base URL with `credentials: "include"`.
- No query, saved-search label, hidden state, source URL, or other internal jd_match state is included in navigation params.
- Handoff identifiers do not leak to URL, localStorage, sessionStorage, console, or telemetry.
- Downstream parse regression scenarios E2E.P0.015 and E2E.P0.016 remain green.
- Confirm-interview Vitest spec and source-level params-shape grep pass.
- Backend E2E.P0.094-P0.097 remain the live API proof for auth, IK,
  persistence, privacy, and AI provenance semantics.
