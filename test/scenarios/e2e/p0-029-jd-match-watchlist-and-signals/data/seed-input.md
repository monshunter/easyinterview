# Seed Input

- Authenticated runtime using fixture-backed API responses.
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
