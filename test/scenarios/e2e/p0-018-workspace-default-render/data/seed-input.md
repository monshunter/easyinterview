# Seed Input

- 用户登录态：已登录（fixture getMe=authenticated）
- API 数据源：OpenAPI fixture transport
  - `getTargetJob` scenario=default (含 fitSummary/requirements)
  - `getResume` scenario=default
  - `getPracticePlan` scenario=default (status=ready)
- Route params: targetJobId=01918fa0-0000-7000-8000-000000002000, jdId=jd-1, resumeVersionId=01918fa0-0000-7000-8000-000000001000, roundId=round-hr
