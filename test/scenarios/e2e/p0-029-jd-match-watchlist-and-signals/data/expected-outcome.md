# Expected Outcome

- Watchlist rows render tone, left-border state, added-at label, optional change text, and chevron controls from fixtures.
- The trigger log includes `VITE_EI_API_MODE=real` and
  `jdMatch.realApiMode.test.ts`, proving all 12 JobMatch generated-client
  operations use the real backend base URL with `credentials: "include"`.
- Market signals render four cards with key, value, delta, tone, and dash fallback for missing delta.
- Chevron handoff selects the linked recommendation when it exists.
- Missing linked recommendation falls back to the first visible Recommended card and emits a warning toast.
- Empty, 4xx, and partial-data variants render their documented inline states.
- Watchlist item labels, linkedJobMatchId, and sourceJobUrl do not leak to URL, storage, console, or telemetry surfaces.
- Watchlist-tab Vitest specs all pass.
- Backend E2E.P0.094-P0.097 remain the live API proof for auth, IK,
  persistence, privacy, and AI provenance semantics.
