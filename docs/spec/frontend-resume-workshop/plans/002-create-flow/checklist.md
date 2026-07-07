# Frontend Resume Workshop Create Flow Checklist

> **版本**: 1.4
> **状态**: completed
> **更新日期**: 2026-07-07

**关联计划**: [plan](./plan.md)

## Phase 1: Create Flow shell

- [x] 1.1 `flow=create` renders `ResumeCreateFlow`.
- [x] 1.2 CreateFlow keeps upload / paste tabs only.
- [x] 1.3 Auth pending action carries route state only and excludes raw resume content.

## Phase 2: Upload / paste registration

- [x] 2.1 Upload path validates file shape, obtains presign data, performs browser PUT, then calls `registerResume`.
- [x] 2.2 Paste path calls `registerResume` with paste payload after non-empty validation and sends a content-derived title instead of generic “粘贴的简历 / Pasted resume”.
- [x] 2.3 Side-effect requests include `Idempotency-Key`; polling requests do not.

## Phase 3: Direct-to-detail navigation

- [x] 3.1 Upload registration success navigates directly to `resume_versions?resumeId=<id>` and does not render `resume-parse-flow` / `resume-preview-confirm`.
- [x] 3.2 Paste registration success navigates directly to `resume_versions?resumeId=<id>` and does not render parser animation, preview confirm, or call create-flow `updateResume`.
- [x] 3.3 Direct navigation does not persist raw resume content into URL, pending action, storage or logs.

## Phase 4: Retired parsing / preview-confirm negative

- [x] 4.1 `ResumeCreateFlow` no longer imports or renders `ParsingStage`, `ResumeParseFlow`, `PreviewStage`, or `ResumePreviewConfirm`.
- [x] 4.2 Create-flow tests and scenario scripts no longer execute parser/preview-confirm positive tests.
- [x] 4.3 Source negative scan fails on user-visible copy for “正在阅读你的原始内容 / 结构化草稿如下 / Confirm and save resume” inside current create-flow runtime.

## Phase 5: BDD / integration gates

- [x] 5.1 BDD-Gate: E2E.P0.081 create-flow upload/paste direct-to-detail happy path is maintained.
- [x] 5.2 Retired gate: E2E.P0.082 no longer validates parser failure UI; verify script records parser flow as non-current/absent.
- [x] 5.3 BDD-Gate: E2E.P0.083 Home / Workspace CTA direct-create handoff is maintained without preview confirm.
- [x] 5.4 `corepack pnpm --filter @easyinterview/frontend test src/app/screens/resume-workshop/create` PASS.
- [x] 5.5 `corepack pnpm --filter @easyinterview/frontend test src/app/screens/resume-workshop/ResumeWorkshopScreen.test.tsx src/app/screens/resume-workshop/fixture-parity.test.ts` PASS.
