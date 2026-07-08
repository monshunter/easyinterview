# Seed Input

- 用户登录态：已登录（fixture getMe=authenticated）
- API 数据源：OpenAPI fixture transport
  - `getTargetJob` scenario=default (含 fitSummary/requirements)
  - `listTargetJobs` scenario=default (面试规划列表)
  - `getResume` scenario=default
  - `getPracticePlan` scenario=default (status=ready)
- No-context route: workspace params={}
- Detail route params: targetJobId=01918fa0-0000-7000-8000-000000002000, jdId=jd-1, resumeId=01918fa0-0000-7000-8000-000000001000, roundId=round-hr
