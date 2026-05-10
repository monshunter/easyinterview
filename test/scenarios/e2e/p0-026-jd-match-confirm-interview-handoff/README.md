# E2E.P0.026 — Confirm Interview from JD Match → Parse Handoff Params

> **Scenario ID**: E2E.P0.026
> **Owner**: frontend-home-job-picks-and-parse/002-jd-match-recommendations
> **Status**: Ready
> **Execution**: automated

## Scope

Verifies the Confirm interview handoff landed in plan 002 Phase 3.5:

- The Recommended-tab Confirm interview button navigates to the `parse`
  route with **strictly** `{ source: "jd_match", sourceJobMatchId }` —
  no other internal jd_match state (query, saved, hidden, label, ...)
  leaks into the navigation params
- `source` and `sourceJobMatchId` never enter URL / localStorage /
  sessionStorage / telemetry
- The parse screen consumed downstream of the handoff continues to
  satisfy plan 001 invariants (E2E.P0.015 + E2E.P0.016 regression)

## Fixtures

- `openapi/fixtures/JobMatch/listJobRecommendations.json` (`many`
  variant) — exercises the multi-card path so we can assert the right
  `sourceJobMatchId` is selected when the user picks a card other than
  the first
- Plan 001 fixtures `importTargetJob`, `getTargetJob`, `updateTargetJob`
  remain reused; this scenario does not introduce parse-side fixtures

## Verification Points

- The Confirm-interview Vitest spec (`RecommendedConfirmInterview.test.tsx`)
  passes
- Source-level grep confirms no extra params are added to the
  `nav("parse", ...)` call site

## Scripts

- `scripts/setup.sh` — record setup timestamp and output dir
- `scripts/trigger.sh` — run the Confirm-interview Vitest spec
- `scripts/verify.sh` — assert pass markers and source-level params shape
- `scripts/cleanup.sh` — clear setup.env

## Offline Limitations

- All data flows through fixture-backed mock transport (no network)
