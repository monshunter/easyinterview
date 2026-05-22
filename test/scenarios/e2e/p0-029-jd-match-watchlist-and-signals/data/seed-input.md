# Seed Input

- Authenticated runtime using fixture-backed API responses for deterministic UI
  variants, plus a real-mode generated-client gate with
  `VITE_EI_API_MODE=real`.
- JobMatch fixtures:
  - `listWatchlist.default`
  - `listWatchlist.empty`
  - `listWatchlist.few`
  - `listWatchlist.4xx`
  - `getMarketSignals.default`
  - `getMarketSignals.partial-data`
  - `getMarketSignals.failed`
  - `listJobRecommendations.default`
  - `listJobRecommendations.empty`
- User opens `jd_match`, switches to Watchlist, inspects watched rows and market signals, then uses chevrons for linked and missing recommendation handoff paths.
- Backend E2E.P0.094-P0.097 provide the paired live route/persistence/AI
  provenance proof for the same JobMatch operation family.
