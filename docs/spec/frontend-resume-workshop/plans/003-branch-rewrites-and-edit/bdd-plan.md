# 003 BDD Plan

> **版本**: 1.3
> **状态**: active
> **更新日期**: 2026-06-14

**关联 Plan**: [plan](./plan.md)

## 1 场景矩阵

> D-20 Phase 8 收口说明：P0.084-P0.087 保留原 scenario ID 和目录名以维持历史可追溯性；当前验收语义已从 BranchFlow / version-tree 迁移为 flat resume detail / rewrites / edit gate。目录名中的 `branch-flow` / `accept-reject` 属于历史标签，不代表当前产品入口。

| 场景 ID | 类别 | 关联 Phase | 关联 Spec C-* | 关联 BDD-Gate（主 checklist） |
|---------|------|-----------|--------------|----------------------------|
| E2E.P0.084 | regression + route + legacy-negative · retired BranchFlow + flat Resume Workshop route dispatch | Phase 8.1 | C-11, C-8, C-9 | Phase 8.6 |
| E2E.P0.085 | primary + failure · flat Rewrites tailor polling + rerun + ready/failed/timeout | Phase 8.4 + 8.6 | C-11, C-8 | Phase 8.6 |
| E2E.P0.086 | primary + boundary · accept-only save modal + `updateResume` / `duplicateResume` + Edit Tab `updateResume` | Phase 8.2 + 8.3 + 8.5 | C-11, C-8 | Phase 8.6 |
| E2E.P0.087 | regression + UX · export/copy non-regression + flat detail/Rewrites/Edit parity + retired negative | Phase 8.5 + 8.6 | C-11, C-8, C-9 | Phase 8.6 |

---

## Phase 8.1: Retired BranchFlow + Flat Route Regression

| 场景 ID | 场景 | Given | When | Then | 验证入口 |
|---------|------|-------|------|------|----------|
| E2E.P0.084 | Retired BranchFlow + flat route / auth / create / detail smoke | fixture-backed frontend client；authenticated 与 unauthenticated paths；D-20 flat `listResumes` / `getResume` / `updateResume` / `duplicateResume` / tailor fixtures 可用 | 运行 `ResumeWorkshopScreen.test.tsx`、`ResumeDetailView.test.tsx`、`ResumeRewritesTab.test.tsx`、`PreviewStage.test.tsx`、`ResumeWorkshopAuthGate.test.tsx`；执行 retired grep | `flow=branch` 不渲染 `resume-branch-flow`，未知 flow 回落 flat list；auth gate 不触发 protected APIs；flat detail/create/rewrites surfaces 可渲染；runtime source 中 `ResumeBranchFlow` / `branchResumeVersion` / `seedStrategy` / `acceptResumeTailorSuggestion` / `rejectResumeTailorSuggestion` / `updateResumeVersion` 0 命中；retired tailor mode 与 prototype import 0 命中 | `test/scenarios/e2e/p0-084-resume-branch-flow-three-seed-strategies/` |

## Phase 8.4: Flat Rewrites Tailor Polling + Rerun

| 场景 ID | 场景 | Given | When | Then | 验证入口 |
|---------|------|-------|------|------|----------|
| E2E.P0.085 | Rewrites Tab polling + rerun + ready/failed/timeout + no-IK read path | fixture-backed `requestResumeTailor` default/idempotency-replay and `getResumeTailorRun` queued/generating/default/failed；flat resume + optional route `targetJobId` | 运行 `ResumeRewritesTab.test.tsx`、`useResumeTailorRunPolling.test.tsx`、`useRequestResumeTailor.test.tsx` | polling banner / failed / timeout / retry CTA render；`getResumeTailorRun` read path has no IK; `requestResumeTailor` has IK replay/rotation; ready callback fires once; unmount cancels timers; retired tailor mode grep 0 | `test/scenarios/e2e/p0-085-resume-rewrites-tab-tailor-run-polling/` |

## Phase 8.2 + 8.3 + 8.5: Accept-Only Save + Flat Profile Merge + Edit Save

| 场景 ID | 场景 | Given | When | Then | 验证入口 |
|---------|------|-------|------|------|----------|
| E2E.P0.086 | Accept-only rewrite save + flat structuredProfile merge + Edit Tab `updateResume` | fixture-backed flat `getResume`, `updateResume`, `duplicateResume`, `requestResumeTailor`, `getResumeTailorRun`; D-20 suggestion decisions are ephemeral | 运行 `PreviewStage.test.tsx`、`ResumeRewritesTab.test.tsx`、`ResumeDetailView.test.tsx`、`ResumeEditTab.test.tsx` | accept is local only; overwrite calls `updateResume`, save-as-new calls `duplicateResume`; accepted rewrites merge into `sections` / `experience` / `experiences` / `projects`; omitted `structuredProfile` fallback does not crash; Edit Tab saves flat headline/summary via `updateResume`; retired accept/reject/updateVersion grep 0; privacy red lines hold | `test/scenarios/e2e/p0-086-resume-suggestion-accept-reject-edit-and-update-version/` |

## Phase 8.5 + 8.6: Export / Copy + Flat UI Parity

| 场景 ID | 场景 | Given | When | Then | 验证入口 |
|---------|------|-------|------|------|----------|
| E2E.P0.087 | Export PDF / copyText non-regression + flat detail/Rewrites/Edit parity + retired negative | flat `getResume` ready; `exportResume` P0 501 fixture; Playwright static server can render frontend dist and ui-design reference | Run focused Vitest, `pnpm --filter @easyinterview/frontend build`, and Playwright `tests/pixel-parity/resume-workshop-branch-rewrites-edit.spec.ts` | export uses `exportResume` IK + friendly 501 toast; copyText writes `buildResumePlainText`; flat detail/Rewrites/Edit DOM/style/bounding/screenshot smoke/axe PASS desktop + mobile; runtime retired operation grep, retired tailor mode grep, and prototype import grep are 0 | `test/scenarios/e2e/p0-087-resume-detail-export-copy-consistency-and-parity/` |
