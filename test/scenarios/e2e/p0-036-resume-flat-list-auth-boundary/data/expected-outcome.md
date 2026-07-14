# Expected outcome — E2E.P0.036

trigger.log 含：

- `Tests  5 passed (5)`
- `Test Files  1 passed (1)`
- 测试文件路径 `src/app/scenarios/p0-036-resume-flat-list-auth-boundary.test.tsx`
- 列表 fixture 每项精确为 9 个 ResumeSummary 字段，不含 detail-only 字段
- 列表稳定后 `listResumes=1` 且 `getResume=0`；点击一行进入详情后 `getResume=1`
- 首次列表 transport reject 后错误状态可见；retry 发起新的第 2 次 transport 并成功

verify.sh 在 trigger.log 中校验：

- 不含 out-of-scope testid 字面量：`route-welcome`、`route-mistakes`、`route-drill`、`route-followup`、`route-onboarding`、`route-experiences`、`route-star`、`route-voice`
- 不含 removed fallback phase marker（resume_versions 已由 ResumeWorkshopScreen 接管）
- 不含 `FAIL`、`SKIP`、`no tests to run` 或 `not wrapped in act`
