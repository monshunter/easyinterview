# E2E.P0.030 — JD Match Profile Chip + AGENT Scan Status + Auth Pending Action Cross-Tab

> **Scenario ID**: E2E.P0.030
> **Owner**: frontend-home-job-picks-and-parse/002-jd-match-recommendations
> **Status**: Ready
> **Execution**: automated

## Scope

Verifies the cross-tab data-driven shell landed in plan 002 Phase 2 + 3:

- Profile chip avatar / SEARCHING AS / skills tags / PROFILE SOURCES are
  data-driven from `getJobMatchProfile`; partial-profile + unauthenticated
  fixtures degrade gracefully
- AGENT scan status badge tone reflects `idle` / `scanning` / `error`
  variants with `lastScanAt` / `nextScanAt` relative formatting
- `getJobMatchProfile` and `getAgentScanStatus` each fire once on tab entry;
  switching back to Recommended re-fires `getAgentScanStatus`; no
  setInterval / SSE / WebSocket is ever registered
- Unauthenticated side-effect actions (Save / Unsave / Mark not relevant /
  Confirm interview / Run search / Save current) dispatch
  `requestAuth({ type: "jd_match_action", route: "jd_match", ... })`;
  Source button and Watchlist chevron stay outside the pending-action
  surface

## Fixtures

- `openapi/fixtures/JobMatch/getJobMatchProfile.json`
  (default/unauthenticated/partial-profile)
- `openapi/fixtures/JobMatch/getAgentScanStatus.json`
  (idle/scanning/error/next-scan-soon)
- `openapi/fixtures/Auth/getMe.json` (authenticated/unauthenticated)
- `openapi/fixtures/Auth/getRuntimeConfig.json`

## Verification Points

- All Profile / AGENT / Auth-gate Vitest spec files pass
- AGENT badge tone toggles (idle / scanning / error) renders correctly
- No setInterval / SSE / WebSocket registered (asserted by Vitest spies)

## Scripts

- `scripts/setup.sh` — record setup timestamp and output dir
- `scripts/trigger.sh` — run the data-driven Profile / AGENT Vitest specs
- `scripts/verify.sh` — assert pass markers and required spec presence
- `scripts/cleanup.sh` — clear setup.env

## Offline Limitations

- All data flows through fixture-backed mock transport (no network)
