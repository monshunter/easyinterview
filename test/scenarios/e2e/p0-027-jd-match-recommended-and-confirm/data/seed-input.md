# Seed Input

- Authenticated runtime using fixture-backed API responses for deterministic UI
  variants, plus a real-mode generated-client gate with
  `VITE_EI_API_MODE=real`.
- JobMatch fixtures:
  - `getJobMatchProfile.default`
  - `getAgentScanStatus.idle`
  - `listJobRecommendations.default`
  - `getJobRecommendation.default`
  - `addToWatchlist.default` and 4xx failure variants
  - `removeFromWatchlist.default` and 4xx failure variant
  - `markJobNotRelevant.default` and 4xx failure variant
- Unauthenticated runtime using `getMe.unauthenticated` for pending-action coverage.
- User opens `jd_match`, stays on the Recommended tab, selects visible recommendation cards, and exercises Confirm, Save, Source, and Not relevant actions.
- Backend E2E.P0.094-P0.097 provide the paired live route/persistence/AI
  provenance proof for the same JobMatch operation family.
