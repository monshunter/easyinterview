# E2E.P0.096 jd-match-search-and-market-signals

> **Status**: scaffold (live DB + cmd/api drainer registration required)
> **Owner plan**: [backend-jobs-recommendations/001](../../../docs/spec/backend-jobs-recommendations/plans/001-jd-match-real-backend-baseline/plan.md)
> **Spec acceptance**: C-9, C-11, C-13, C-14, C-17

## Scope

- `POST /api/v1/jd-match/search` sync via A3/F3 `jd_match.search`
  feature_key; 30s timeout maps to 502 AI_PROVIDER_TIMEOUT.
- jd_match_search_runs persists provenance; `query` / `filters` never
  enter log / audit / outbox.
- `GET /api/v1/jd-match/market-signals?window=7d|14d|30d` returns 4
  signals with tone derivation; unknown window -> 422.

## Live-environment prerequisite

cmd/api + AIClient stub provider with happy / timeout / output_invalid
variants. Scripts exit non-zero when live env absent.

## Deferred

Script bodies land with Phase 5.7 / Phase 6.7.
