# 003 BDD Checklist

> **版本**: 1.3
> **状态**: active
> **更新日期**: 2026-06-14

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## E2E.P0.084 retired branch flow + flat route regression

- [x] 场景目录 `test/scenarios/e2e/p0-084-resume-branch-flow-three-seed-strategies/` 保留历史目录名但 README / data / scripts 语义已改为 D-20 flat route regression。
- [x] `scripts/trigger.sh` 前置 `frontend-real-backend-gate.sh`，并运行 `ResumeWorkshopScreen.test.tsx`、`ResumeDetailView.test.tsx`、`ResumeRewritesTab.test.tsx`、`PreviewStage.test.tsx`、`ResumeWorkshopAuthGate.test.tsx`。
- [x] `scripts/verify.sh` 检查 real-backend marker、Vitest RUN marker、passing Test Files / Tests summary、目标 spec 文件名、no-test/fail marker rejection。
- [x] `scripts/verify.sh` 执行 retired grep：`ResumeBranchFlow` / `branchResumeVersion` / `seedStrategy` / `acceptResumeTailorSuggestion` / `rejectResumeTailorSuggestion` / `updateResumeVersion` runtime source 0 命中，retired tailor mode 与 prototype import 0 命中。
- [x] 执行 `setup → trigger → verify → cleanup` PASS，输出 `.test-output/e2e/p0-084-resume-branch-flow-three-seed-strategies/trigger.log`。
  <!-- verified: 2026-06-14 method=scenario evidence=P0.084 setup->trigger->verify->cleanup PASS; trigger log shows real-backend marker + Vitest 5 files / 46 tests passed; verify enforces no-test/fail rejection and retired grep 0 hit -->

## E2E.P0.085 flat rewrites tailor polling + rerun

- [x] 场景目录 `test/scenarios/e2e/p0-085-resume-rewrites-tab-tailor-run-polling/` README / data / scripts 描述 flat resume + optional `targetJobId` rerun path。
- [x] `scripts/trigger.sh` 前置 `frontend-real-backend-gate.sh`，并运行 `ResumeRewritesTab.test.tsx`、`useResumeTailorRunPolling.test.tsx`、`useRequestResumeTailor.test.tsx`。
- [x] `scripts/verify.sh` 检查 real-backend marker、Vitest RUN marker、passing summary、目标 spec 文件名、no-test/fail marker rejection。
- [x] gate 语义覆盖 `requestResumeTailor` IK replay/rotation、`getResumeTailorRun` read-only no-IK、ready/failed/timeout/retry/unmount cleanup、retired tailor mode grep。
- [x] 执行 `setup → trigger → verify → cleanup` PASS，输出 `.test-output/e2e/p0-085-resume-rewrites-tab-tailor-run-polling/trigger.log`。
  <!-- verified: 2026-06-14 method=scenario evidence=P0.085 setup->trigger->verify->cleanup PASS; trigger log shows real-backend marker + Vitest 3 files / 30 tests passed; verify enforces no-test/fail rejection and retired tailor grep -->

## E2E.P0.086 accept-only save + flat profile merge + Edit Tab updateResume

- [x] 场景目录 `test/scenarios/e2e/p0-086-resume-suggestion-accept-reject-edit-and-update-version/` 保留历史目录名但 README / data / scripts 语义已改为 D-20 accept-only + flat save。
- [x] `scripts/trigger.sh` 前置 `frontend-real-backend-gate.sh`，并运行 `PreviewStage.test.tsx`、`ResumeRewritesTab.test.tsx`、`ResumeDetailView.test.tsx`、`ResumeEditTab.test.tsx`。
- [x] `scripts/verify.sh` 检查 real-backend marker、Vitest RUN marker、passing summary、目标 spec 文件名、no-test/fail marker rejection。
- [x] gate 语义覆盖 overwrite `updateResume`、save-as-new `duplicateResume`、flat profile bullets (`sections` / `experience` / `experiences` / `projects`)、omitted `structuredProfile` fallback、route `targetJobId` rerun body、retired accept/reject/updateVersion grep。
- [x] 执行 `setup → trigger → verify → cleanup` PASS，输出 `.test-output/e2e/p0-086-resume-suggestion-accept-reject-edit-and-update-version/trigger.log`。
  <!-- verified: 2026-06-14 method=scenario evidence=P0.086 setup->trigger->verify->cleanup PASS; trigger log shows real-backend marker + Vitest 4 files / 41 tests passed; verify enforces flat profile gate, no-test/fail rejection, and retired accept/reject/updateVersion grep 0 hit -->

## E2E.P0.087 export/copy + flat UI parity + retired negative

- [x] 场景目录 `test/scenarios/e2e/p0-087-resume-detail-export-copy-consistency-and-parity/` README / data / scripts 描述 flat detail/Rewrites/Edit parity。
- [x] `scripts/trigger.sh` 前置 `frontend-real-backend-gate.sh`，运行 focused Vitest、`pnpm --filter @easyinterview/frontend build`、Playwright `tests/pixel-parity/resume-workshop-branch-rewrites-edit.spec.ts`。
- [x] `scripts/verify.sh` 检查 real-backend marker、Vitest/build/Playwright runner marker、passing summaries、目标 spec 文件名、no-test/fail marker rejection。
- [x] gate 语义覆盖 `exportResume` IK + 501 toast、copyText `buildResumePlainText`、flat detail/Rewrites/Edit DOM/style/bounding/screenshot smoke/axe、retired operation grep、retired tailor mode grep、prototype import grep。
- [x] 执行 `setup → trigger → verify → cleanup` PASS，输出 `.test-output/e2e/p0-087-resume-detail-export-copy-consistency-and-parity/trigger.log`。
  <!-- verified: 2026-06-14 method=scenario evidence=P0.087 setup->trigger->verify->cleanup PASS; trigger log shows real-backend marker, Vitest 5 files / 39 tests passed, frontend build PASS, Playwright 4 passed, and verify retired/prototype greps 0 hit -->
