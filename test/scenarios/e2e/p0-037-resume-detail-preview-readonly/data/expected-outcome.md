# Expected outcome — E2E.P0.037

trigger.log 含：

- `Tests  6 passed (6)`
- `Test Files  1 passed (1)`
- 测试文件路径 `src/app/scenarios/p0-037-resume-detail-preview-readonly.test.tsx`
- pending PDF upload detail 会轮询 `getResume` 直到展示 source page stack 和 LLM-derived `displayName`
- failed-with-snapshot PDF upload detail 只请求一次 `getResume`，直接展示 source page stack 与 backend generated `displayName`
- stdout/stderr 中不出现 `not wrapped in act`

verify.sh 在 trigger.log 中校验：

- 不出现 out-of-scope testid 字面量：`route-welcome`、`route-mistakes`、`route-drill`、`route-followup`、`route-onboarding`、`route-experiences`、`route-star`、`route-voice`
- 不直接回显 fixture error.code 字面量 `TARGET_JOB_NOT_FOUND`
- 不出现 removed fallback phase marker
