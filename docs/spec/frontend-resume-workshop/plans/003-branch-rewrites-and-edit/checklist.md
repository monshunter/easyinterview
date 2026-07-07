# Frontend Resume Workshop Rewrites and Edit Checklist

> **版本**: 1.4
> **状态**: completed
> **更新日期**: 2026-07-07

**关联计划**: [plan](./plan.md)

## Phase 1: Flat Resume Workshop Rewrites / Edit

- [x] 1.1 上游 flat operation matrix 已核对：`listResumes` / `getResume` / `requestResumeTailor` / `getResumeTailorRun` / `updateResume` / `duplicateResume` / `exportResume` generated client、fixtures 和 backend handler 可用。
  <!-- verified: 2026-06-14 method=scenario evidence="P0.084-P0.087 wrappers all ran with frontend real-backend generated-client preflight; make codegen-check PASS during D-20 closeout." -->
- [x] 1.2 Flat route gate：`flow=branch` 不 materialize，runtime 不渲染 `resume-branch-flow`；Resume Workshop runtime source 中 version-tree operation identifiers 0 命中。
  <!-- verified: 2026-06-14 method=scenario evidence="P0.084 setup->trigger->verify->cleanup PASS; trigger log shows real-backend gate + 5 files / 46 tests passed; verify branch/version negative grep 0 hit in runtime source." -->
- [x] 1.3 Rewrites accept-only + save modal gate：`ResumeRewritesTab` 只提供本地采纳；`RewriteSaveConfirmModal` 覆盖 `updateResume` overwrite 与 `duplicateResume` save-as-new；不发送 server-side suggestion decision operations。
  <!-- verified: 2026-06-14 method=scenario evidence="P0.086 setup->trigger->verify->cleanup PASS; trigger log shows real-backend gate + 4 files / 41 tests passed; verify accept/reject/updateVersion negative grep 0 hit." -->
- [x] 1.4 Flat structuredProfile merge gate：accepted rewrites 写入 `sections[]` / `experience[]` / `experiences[]` / `projects[]` bullets，payload 不写 `acceptedRewrites`，omitted `structuredProfile` fallback 不崩溃。
  <!-- verified: 2026-06-14 method=vitest+scenario evidence="P0.084/P0.086 trigger logs include ResumeDetailView regressions for sections/experience/experiences/projects merge, omitted structuredProfile fallback, and no acceptedRewrites payload." -->
- [x] 1.5 Tailor rerun route context gate：route `targetJobId` 从 `ResumeWorkshopScreen` 透传到 Rewrites rerun body；有 JD context 时发送 `{ resumeId, targetJobId, mode }`，无 JD context 时才允许 generic body。
  <!-- verified: 2026-06-14 method=vitest+scenario evidence="P0.084 trigger log includes rerun body regression with {resumeId,targetJobId,mode}; P0.085 setup->trigger->verify->cleanup PASS with 3 files / 30 tests passed." -->
- [x] 1.6 Edit Tab + export/copy gate：`ResumeEditTab` 使用 `updateResume` 保存 flat `displayName` / `headline` / `summary`，Export PDF 使用 `exportResume` P0 501 toast，copyText 使用 `buildResumePlainText`。
  <!-- verified: 2026-06-14 method=scenario evidence="P0.087 setup->trigger->verify->cleanup PASS; trigger log shows focused Vitest 5 files / 39 tests passed, frontend build PASS, Playwright flat detail/Rewrites/Edit parity 4 passed." -->
- [x] 1.7 UI parity / privacy / BDD wrappers：P0.084-P0.087 `setup -> trigger -> verify -> cleanup` PASS；verify 拒绝 no-test/fail marker，检查 real-backend generated-client marker、Playwright flat detail/Rewrites/Edit parity、privacy 与 negative greps。
  <!-- verified: 2026-06-14 method=scenario evidence="P0.084/P0.085/P0.086/P0.087 wrappers all PASS; verify scripts check real-backend marker, runner summaries, no-test/fail marker rejection, negative greps, and P0.087 Playwright/build gates." -->
- [x] 1.8 BDD-Gate: E2E.P0.084 / E2E.P0.085 / E2E.P0.086 / E2E.P0.087 scenario assets and wrappers match the current flat Resume Workshop contract.
  <!-- verified: 2026-07-07 method=owner-doc-reconcile evidence="BDD plan/checklist v1.5 point at current Phase 1 gates; no scenario directory creation required." -->
- [x] 1.9 Owner docs and indexes are current: plan/checklist completed, BDD docs completed, plans INDEX / docs spec INDEX synced, and stale status wording is absent from this owner package.
  <!-- verified: 2026-07-07 method=owner-doc-reconcile evidence="Targeted owner stale-wording grep returned no matches; validate_context.py frontend-resume-workshop/003 frontend PASS; sync-doc-index --check PASS; make docs-check PASS; git diff --check PASS." -->
