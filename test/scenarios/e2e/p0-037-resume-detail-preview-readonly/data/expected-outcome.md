# Expected outcome — E2E.P0.037

trigger.log 含：

- `Tests  5 passed (5)`
- `Test Files  1 passed (1)`
- 测试文件路径 `src/app/scenarios/p0-037-resume-detail-preview-readonly.test.tsx`

verify.sh 在 trigger.log 中校验：

- 不出现 non-current testid 字面量：`route-welcome`、`route-mistakes`、`route-drill`、`route-followup`、`route-onboarding`、`route-experiences`、`route-star`、`route-voice`
- 不直接回显 fixture error.code 字面量 `TARGET_JOB_NOT_FOUND`
- 不出现旧 fallback phase marker
