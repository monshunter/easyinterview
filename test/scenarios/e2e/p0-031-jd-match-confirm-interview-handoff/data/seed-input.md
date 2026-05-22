# Seed Input

- Authenticated runtime using fixture-backed API responses for deterministic UI
  variants, plus a real-mode generated-client gate with
  `VITE_EI_API_MODE=real`.
- JobMatch fixtures:
  - `listJobRecommendations.many`
  - `getJobRecommendation.default`
  - profile and agent status defaults needed by the JD Match shell
- Plan 001 parse fixtures remain available for downstream regression scripts:
  - `importTargetJob`
  - `getTargetJob`
  - `updateTargetJob`
- User opens `jd_match`, selects a recommendation that is not the first card, and clicks Confirm interview.
- Backend E2E.P0.094-P0.097 provide the paired live route/persistence/AI
  provenance proof for the same JobMatch operation family.
