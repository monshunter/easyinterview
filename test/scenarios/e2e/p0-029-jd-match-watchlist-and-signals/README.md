# E2E.P0.029 — JD Match Watchlist Tab + Market Signals + Chevron Handoff

> **Scenario ID**: E2E.P0.029
> **Owner**: frontend-home-job-picks-and-parse/002-jd-match-recommendations
> **Status**: Ready
> **Execution**: automated

## Scope

Verifies the Watchlist tab end-to-end loop landed in plan 002 Phase 5:

- Watchlist tab renders fixture-backed WatchlistItem rows with the 3px
  left border colored per `tone` (ok / warn / muted), addedAt label,
  optional change signal, and chevron button
- Market signals 4-card grid renders k / v / d / tone columns with a
  fallback dash for missing `d`
- Chevron handoff: when `linkedJobMatchId` is in the current visible
  recommendations, the screen switches to the Recommended tab and selects
  that id; when missing, falls back to the first visible card and
  dispatches a warn `eiToast`
- Empty / 4xx / partial-data variants surface the right inline state
- `linkedJobMatchId` / watchlist `label` / `sourceJobUrl` never enter
  console / URL / localStorage

## Fixtures

- `openapi/fixtures/JobMatch/listWatchlist.json`
  (default/empty/few/4xx)
- `openapi/fixtures/JobMatch/getMarketSignals.json`
  (default/partial-data/failed)
- `openapi/fixtures/JobMatch/listJobRecommendations.json`
  (default/empty for the chevron handoff path)

## Verification Points

- All Watchlist-tab Vitest spec files pass
- 3 tone rendering + 3px left border data-tone testid asserted
- Chevron paths (linked + missing) both covered

## Scripts

- `scripts/setup.sh` — record setup timestamp and output dir
- `scripts/trigger.sh` — run the Watchlist-tab Vitest specs, tee log
- `scripts/verify.sh` — assert pass markers and required spec presence
- `scripts/cleanup.sh` — clear setup.env

## Offline Limitations

- All data flows through fixture-backed mock transport (no network)
- Real backend market signals computation is out of scope for plan 002
