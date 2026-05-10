# E2E.P0.022 — JD Match Recommended Tab + 4-Button Loop + Auth Pending Action

> **Scenario ID**: E2E.P0.022
> **Owner**: frontend-home-job-picks-and-parse/002-jd-match-recommendations
> **Status**: Ready
> **Execution**: automated

## Scope

Verifies the Recommended tab end-to-end loop landed in plan 002 Phase 3:

- Recommended tab list renders fixture-backed JobMatchCard items
- JDDetail header / why / risk / snapshot / intel section + 4 action buttons
- Save / Unsave dispatches `addToWatchlist` / `removeFromWatchlist` with
  unique Idempotency-Key and reverts optimistic state on 4xx
- Mark not relevant dispatches `markJobNotRelevant` with default
  `not_relevant` reason, no `freeNote`, and reverts on 4xx
- Confirm interview navigates to `parse` with strictly
  `{ source: "jd_match", sourceJobMatchId }`
- Source button calls `window.open` with `_blank` and
  `noopener,noreferrer`; bypasses the auth pending-action surface
- Unauthenticated Save / Unsave / Mark not relevant / Confirm interview
  dispatch `requestAuth({ type: "jd_match_action", route: "jd_match",
  params: { tab: "recommended", selectedJobMatchId, action } })`
- `query` / `freeNote` / `sourceUrl` / `label` never enter pendingAction
  params, console, URL or localStorage

## Fixtures

- `openapi/fixtures/JobMatch/getJobMatchProfile.json` (default)
- `openapi/fixtures/JobMatch/getAgentScanStatus.json` (idle/scanning/error)
- `openapi/fixtures/JobMatch/listJobRecommendations.json`
  (default/empty/one/many/failed)
- `openapi/fixtures/JobMatch/getJobRecommendation.json`
  (default/network-intel-empty/failed)
- `openapi/fixtures/JobMatch/addToWatchlist.json`
  (default/4xx-validation/5xx-server-error)
- `openapi/fixtures/JobMatch/removeFromWatchlist.json` (default/4xx-not-found)
- `openapi/fixtures/JobMatch/markJobNotRelevant.json` (default/4xx)

## Verification Points

- The Recommended Vitest spec files all PASS in the trigger log
- 4 Action button testids present (`jdmatch-detail-action-{confirm,save,
  source,dismiss}`)
- Idempotency-Key headers attached on the 3 side-effect calls
- No leak of `jobMatchId` / `freeNote` / `sourceUrl` in console / URL /
  localStorage in the spec runs (encoded by RecommendedPrivacy.test.tsx)

## Scripts

- `scripts/setup.sh` — record setup timestamp and output dir
- `scripts/trigger.sh` — run the Recommended-tab Vitest specs, tee log
- `scripts/verify.sh` — assert pass markers and required spec presence
- `scripts/cleanup.sh` — clear setup.env

## Offline Limitations

- All data flows through fixture-backed mock transport (no network)
- Real backend handlers (`addToWatchlist`, `markJobNotRelevant`, ...)
  remain `not-yet-implemented`; tracked by `backend-jobs-recommendations`
