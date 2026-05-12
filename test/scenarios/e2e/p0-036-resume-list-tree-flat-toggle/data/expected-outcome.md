# Expected outcome — E2E.P0.036

trigger.log 含：

- `Tests  4 passed (4)`
- `Test Files  1 passed (1)`
- 测试文件路径 `src/app/scenarios/p0-036-resume-list-tree-flat-toggle.test.tsx`

verify.sh 在 trigger.log 中校验：

- 不含 retired testid 字面量：`route-welcome`、`route-mistakes`、`route-drill`、`route-followup`、`route-onboarding`、`route-experiences`、`route-star`、`route-voice`
- 不含 placeholder fallback 文本 `D2-D6`（resume_versions 已被实际 ResumeWorkshopScreen 接管）
