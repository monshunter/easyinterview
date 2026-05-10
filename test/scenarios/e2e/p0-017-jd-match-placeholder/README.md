# E2E.P0.017 — JD Match Three-Tab Smoke + Legacy Negative Grep

> **Scenario ID**: E2E.P0.017
> **Owner**: frontend-home-job-picks-and-parse/002-jd-match-recommendations
> **Status**: Ready
> **Execution**: automated

## Scope

After plan 002 lands, jd_match upgrades from a P1 placeholder shell to the
data-driven three-tab screen. This scenario remains a smoke test for the
core route reachability, asserts the new data-driven shell renders, and keeps
the negative grep on the prototype business testid pattern (numeric-index
and `jdmatch-search-bar` style) that must never be reintroduced.

Verifies:

- jd_match route reachable from TopBar nav and Home aux card
- Hero (label/title/sub), profile chip and three tabs present
- Recommended tab list + JDDetail sticky present (data-driven)
- Search tab natural-language search bar + four chip filters present
- Watchlist tab list + market signals 4-card grid + refresh footer present
- Legacy plan-001 placeholder testid (`jdmatch-placeholder*`) absent
- Prototype numeric-index testid (`jdmatch-card-0`, `jdmatch-saved-search-0`,
  `jdmatch-watchlist-0`, `jdmatch-search-bar`) absent

## Fixtures

Consumes JobMatch fixtures landed in plan 002 Phase 1:

- `openapi/fixtures/JobMatch/getJobMatchProfile.json`
- `openapi/fixtures/JobMatch/getAgentScanStatus.json`
- `openapi/fixtures/JobMatch/listJobRecommendations.json`
- `openapi/fixtures/JobMatch/listSavedSearches.json`
- `openapi/fixtures/JobMatch/listWatchlist.json`
- `openapi/fixtures/JobMatch/getMarketSignals.json`

## Verification Points

- All Vitest spec files under `frontend/src/app/screens/jd_match/` pass
- New three-tab DOM anchors appear in trigger.log
- Legacy placeholder testid 0 hits (run-time + source-level)
- Prototype numeric-index testid pattern 0 hits

## Scripts

- `scripts/setup.sh` — record setup timestamp and output dir
- `scripts/trigger.sh` — run the jd_match Vitest suite, tee to trigger.log
- `scripts/verify.sh` — assert pass markers + DOM anchors + negative grep
- `scripts/cleanup.sh` — clear setup.env

## Offline Limitations

- All data flows through fixture-backed mock transport (no network)
- Real backend handlers / agent scan pipeline are out of scope for plan 002
  and tracked separately by `backend-jobs-recommendations`
