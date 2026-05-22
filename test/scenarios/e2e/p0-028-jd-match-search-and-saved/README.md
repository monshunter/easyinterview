# E2E.P0.028 — JD Match Search Tab + Saved Searches + Filters + Failure + Privacy

> **Scenario ID**: E2E.P0.028
> **Owner**: frontend-home-job-picks-and-parse/002-jd-match-recommendations
> **Status**: Ready
> **Execution**: automated

## Scope

Verifies the Search tab end-to-end loop landed in plan 002 Phase 4:

- Natural-language search bar + Run web search button + `SOURCES` label +
  five source chips, including Company sites / 公司官网
- 5-step AGENT scanning panel during in-flight search (`opacity 1/1/1/0.4/0.4`
  i18n-driven, no setInterval, no real step advancement)
- Saved searches list (`listSavedSearches` 1 call on tab entry) + create
  (`createSavedSearch` body+IK+toast)
- 4 chip filter pure client-side (all / strong / remote / unseen) — never
  triggers a new `searchJobs` request
- failed / empty / slow-response variants surface the right inline state
- Unauthenticated Run / Save current dispatches `requestAuth` with
  `tab: "search"` and the corresponding action enum, never carrying
  `query` or `label` in the pendingAction params
- `query` / saved-search `label` / `filter state` / `sourceJobUrl`
  never enter console / URL / localStorage; query resets on tab switch

## Fixtures

- `openapi/fixtures/JobMatch/searchJobs.json`
  (default/empty/failed/slow-response)
- `openapi/fixtures/JobMatch/listSavedSearches.json` (default/empty/4xx)
- `openapi/fixtures/JobMatch/createSavedSearch.json`
  (default/4xx-validation)

## Verification Points

- All Search-tab Vitest spec files pass
- Source-level parity asserts `NATURAL LANGUAGE SEARCH`, the search icon,
  and `jdmatch-search-source-company`
- 5-step scanning panel testid + opacity gradient asserted
- Filter switching does not enqueue an extra `searchJobs` call
- Negative grep on dynamic JD numbers and on `query` / `label` leaks

## Scripts

- `scripts/setup.sh` — record setup timestamp and output dir
- `scripts/trigger.sh` — run the real-mode JobMatch generated-client gate,
  then the Search-tab Vitest specs, tee log
- `scripts/verify.sh` — assert pass markers and required spec presence
- `scripts/cleanup.sh` — clear setup.env

## Real Backend Overlay

- UI behavior sub-cases still use fixture-backed component transports for
  deterministic DOM, auth, privacy, slow-response, empty, and failure variants.
- The trigger also runs `src/api/jdMatch.realApiMode.test.ts` with
  `VITE_EI_API_MODE=real` and
  `VITE_EI_API_BASE_URL=http://localhost:8080/api/v1`, proving the production
  bootstrap/generated client reaches the real backend base URL for all 12
  JobMatch operations, including `searchJobs` and `createSavedSearch`.
- Live search runtime, persistence, AI provenance, auth, IK, and privacy
  semantics are paired with backend E2E.P0.094-P0.097.
