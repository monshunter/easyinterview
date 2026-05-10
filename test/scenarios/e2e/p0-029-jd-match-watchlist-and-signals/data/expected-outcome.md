# Expected Outcome

- Watchlist rows render tone, left-border state, added-at label, optional change text, and chevron controls from fixtures.
- Market signals render four cards with key, value, delta, tone, and dash fallback for missing delta.
- Chevron handoff selects the linked recommendation when it exists.
- Missing linked recommendation falls back to the first visible Recommended card and emits a warning toast.
- Empty, 4xx, and partial-data variants render their documented inline states.
- Watchlist item labels, linkedJobMatchId, and sourceJobUrl do not leak to URL, storage, console, or telemetry surfaces.
- Watchlist-tab Vitest specs all pass.
