# Frontend Resume Workshop Create Flow BDD Plan

> **版本**: 1.11
> **状态**: completed
> **更新日期**: 2026-07-14

**关联 Plan**: [plan](./plan.md)

## 1 场景矩阵

| 场景 ID | 类别 | 覆盖范围 | 关联 Checklist |
|---------|------|----------|----------------|
| E2E.P0.081 | primary | CreateFlow upload / paste happy path, PDF / Markdown / TXT whitelist, DOCX rejection, 2MiB file limit, presign, register, waiting/detail navigation, privacy and UI parity | 5.1 / 6.1 / 6.3 / 7.1 |
| E2E.P0.082 | absence gate | Parser animation / parse failure UI are absent from create flow | 5.2 |
| E2E.P0.083 | primary + handoff | Home CTA and auth pending action direct-create handoff | 5.3 |
| E2E.P0.084 | historical regression | 2026-07-08 full-Resume picker evidence；current summary-only behavior is owned by active 001 Phase 19 | historical 8.1 / 8.2 / 8.3 |

## 2 场景说明

### E2E.P0.081

Given an authenticated user opens the Resume Workshop create route.
When the user completes upload or paste creation and `registerResume` returns a `resumeId`.
Then the app navigates to `resume_versions?resumeId=<id>` where detail owns waiting/terminal states and source-format rendering, side-effect requests use IK, files over 2MiB and DOCX files are rejected before presign, sidebar and preview-confirm DOM are absent, raw resume content is not stored in route state or browser storage, and pasted raw first line is not submitted or displayed as the resume name.

### E2E.P0.082

Given current CreateFlow no longer exposes parser progress.
When source/runtime scans and create-route tests run.
Then parser animation, parser failure retry UI and preview confirm surfaces are absent.

### E2E.P0.083

Given Home CTA enters resume creation.
When upload or paste registration succeeds, or unauthenticated auth pending action is created.
Then direct detail navigation is used and auth pending action keeps only safe route params.

### E2E.P0.084

This completed-plan scenario records the 2026-07-08 full-Resume contract only. Active 001 Phase 19 supersedes it: Home receives closed `ResumeSummary`, treats `parseStatus === ready || hasReadableContent` as selectable, and preserves `resumeId` in the import body；Parse/Workspace detail does not call `listResumes`.

## 3 Internal cleanup substitute gate

Phase 9 changes no user-visible behavior and adds no BDD scenario. Its completion gate is a source-level zero-reference assertion plus the existing create-flow regressions and frontend typecheck; E2E.P0.081-P0.084 remain unchanged.
