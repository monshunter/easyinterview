# Seed Input

- Authenticated runtime using fixture-backed API responses.
- JobMatch fixtures:
  - `listJobRecommendations.many`
  - `getJobRecommendation.default`
  - profile and agent status defaults needed by the JD Match shell
- Plan 001 parse fixtures remain available for downstream regression scripts:
  - `importTargetJob`
  - `getTargetJob`
  - `updateTargetJob`
- User opens `jd_match`, selects a recommendation that is not the first card, and clicks Confirm interview.
