# Frontend Resume Workshop Create Flow BDD Plan

> **版本**: 1.4
> **状态**: completed
> **更新日期**: 2026-07-07

**关联 Plan**: [plan](./plan.md)

## 1 场景矩阵

| 场景 ID | 类别 | 覆盖范围 | 关联 Checklist |
|---------|------|----------|----------------|
| E2E.P0.081 | primary | CreateFlow upload / paste happy path, presign, register, direct detail navigation, privacy and UI parity | 5.1 |
| E2E.P0.082 | retired negative | Parser animation / parse failure UI are non-current and absent from create flow | 5.2 |
| E2E.P0.083 | primary + handoff | Home CTA, Workspace CTA and auth pending action direct-create handoff | 5.3 |

## 2 场景说明

### E2E.P0.081

Given an authenticated user opens the Resume Workshop create route.
When the user completes upload or paste creation and `registerResume` returns a `resumeId`.
Then the app navigates directly to `resume_versions?resumeId=<id>`, side-effect requests use IK, parser/preview-confirm DOM is absent, raw resume content is not stored in route state or browser storage, and pasted raw first line is not submitted or displayed as the resume name.

### E2E.P0.082

Given current CreateFlow no longer exposes parser progress.
When source/runtime scans and create-route tests run.
Then parser animation, parser failure retry UI and preview confirm surfaces are absent.

### E2E.P0.083

Given Home / Workspace CTA enters resume creation.
When upload or paste registration succeeds, or unauthenticated auth pending action is created.
Then direct detail navigation is used and auth pending action keeps only safe route params.
