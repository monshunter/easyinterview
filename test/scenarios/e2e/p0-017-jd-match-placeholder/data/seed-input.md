# Seed Input

- Route: `jd_match` through TopBar nav and Home aux entry.
- Fixture-backed JobMatch data:
  - `openapi/fixtures/JobMatch/getJobMatchProfile.json`
  - `openapi/fixtures/JobMatch/getAgentScanStatus.json`
  - `openapi/fixtures/JobMatch/listJobRecommendations.json`
  - `openapi/fixtures/JobMatch/listSavedSearches.json`
  - `openapi/fixtures/JobMatch/listWatchlist.json`
  - `openapi/fixtures/JobMatch/getMarketSignals.json`
- Frontend suite: all Vitest specs under `frontend/src/app/screens/jd_match/`.
- Negative source scan: production `jd_match` screen files excluding test/spec files.
