# 001 Workspace + InterviewContext + Start Practice Contract Checklist

> **版本**: 1.43
> **状态**: completed
> **更新日期**: 2026-07-14

**关联计划**: [plan](./plan.md)

## Phase 0: contract preflight

- [x] 0.1 `docs/development.md` §2 frontend/backend contract workflow is the execution boundary（验证：generated client + fixture-backed transport used; no ad hoc workspace fetch）
- [x] 0.2 UI design document is current workspace prototype and docs（验证：`docs/ui-design/module-job-workspace.md`, `frontend/src`）
- [x] 0.3 Context manifest resolves current frontend target and spec version（验证：`validate_context.py frontend-workspace-and-practice/001 frontend` PASS）

## Phase 1: Workspace shell and InterviewContext

- [x] 1.1 `InterviewContextProvider` carries stable target/resume/round/plan/session IDs and `practiceGoal` across owner routes; mode/modality/hint fields are stripped（验证：`InterviewContext.test.tsx`, `App.test.tsx`）
- [x] 1.2 `workspace` route renders `WorkspaceScreen` instead of the route fallback shell; non-owner routes keep their own owners（验证：`App.test.tsx`）
- [x] 1.3 `workspace.*` zh/en messages and DOM anchors cover the pure plan-list eyebrow, title, cards, mini round rail, empty state and current actions（验证：`WorkspaceScreen.test.tsx`）

## Phase 2: TargetJob, resume and workspace data

- [x] 2.1 Historical `useWorkspaceTargetJob` detail hook moved out of workspace owner; parse owner consumes generated `getTargetJob`（验证：source negative gate）
- [x] 2.2 Historical `useWorkspaceResume` detail hook moved out of workspace owner; parse owner consumes resume selection/list data（验证：`ParseResumeBinding.test.tsx`）
- [x] 2.3 Workspace list derives only declared `TargetJob` list fields and does not render detail header/launcher/JD breakdown（验证：`WorkspaceEmptyState.test.tsx`, source negative gate）

## Phase 3: Plan and resume switching

- [x] 3.1 Historical `PlanSwitcherModal` runtime moved out of workspace owner; list cards now open parse detail（验证：source negative gate）
- [x] 3.2 Historical `ResumePickerModal` runtime moved out of workspace owner; parse detail owns resume selection（验证：`ParseResumeBinding.test.tsx`）
- [x] 3.3 Modal a11y is no longer a workspace gate; parse owner covers its picker behavior（验证：parse focused tests）

## Phase 4: Start practice and auth recovery

- [x] 4.1 Historical workspace practice-plan refresh hook moved out of this owner; parse/report handoff validates existing plan context（验证：`ParseResumeBinding.test.tsx`, `ReplayCta.test.tsx`）
- [x] 4.2 Shared `startPracticeFromParams` creates a plan when needed, starts a session, and uses stable idempotency keys for side effects（验证：`ParseResumeBinding.test.tsx`, `ReplayCta.test.tsx`）
- [x] 4.3 `workspace(autoStartPractice=1)` is removed from current runtime; pending actions no longer rely on workspace side effects（验证：source negative gate）
- [x] 4.4 BDD-Gate: `BDD.WORKSPACE.CONTEXT.001` 由 [BDD checklist](./bdd-checklist.md) 关联 Workspace/start-practice owner behavior tests。

## Phase 5: Workspace boundary and privacy

- [x] 5.1 Company insight component/API runtime stays outside workspace owner（验证：source negative gate）
- [x] 5.2 Records static affordance runtime stays outside workspace owner（验证：source negative gate）
- [x] 5.4 Sensitive fields are absent from URL, localStorage, console, telemetry and fixture transport logs（验证：privacy negative tests and scenario verify）

## Phase 6: closeout

- [x] 6.1 App、Workspace、Header、modals、start practice、auth 与 handoff 的 focused Vitest 仅作开发反馈；阶段单测完成由仓库根 `make test` 承接。
- [x] 6.2 Formal component layout tests passed for workspace desktop/mobile and theme states（验证：`pnpm --filter @easyinterview/frontend test`）
- [x] 6.3 Fixtures remain valid for TargetJobs, Resumes, PracticePlans and PracticeSessions（验证：`make validate-fixtures`）
- [x] 6.4 Owner docs/index/context are current and completed（验证：`validate_context.py frontend-workspace-and-practice/001 frontend`; `sync-doc-index --check`; `make docs-check`）

## Phase 7: interview nav and plan-list landing revision

- [x] 7.1 Product/UI design documents and static prototype use TopBar `面试` / `Interview` and define `workspace` no-context plan-list landing（验证：`frontend/src`, `docs/ui-design/module-job-workspace.md`, `docs/ui-design/ui-architecture.md`）
- [x] 7.3 `WorkspacePlanList` consumes generated `listTargetJobs`, renders loading/empty/error/list states, and plan cards navigate to `parse` detail without fabricating resume/report data（验证：`WorkspaceScreen.test.tsx`, `WorkspaceEmptyState.test.tsx`）

## Phase 8: plan-list card visual hardening

- [x] 8.1 UI design documents define the no-context plan list as visible list cards with card background, border, subtle elevation, internal body/footer sections, and responsive desktop/mobile grid（验证：`docs/ui-design/module-job-workspace.md`, `frontend/src`）
- [x] 8.2 `WorkspacePlanList` mirrors the card treatment and keeps generated `listTargetJobs` + safe navigation semantics unchanged（验证：`WorkspaceEmptyState.test.tsx` red/green assertions）
- [x] 8.3 Formal frontend component tests catch loose text-column regression through layout/style assertions for card, body and footer sections。

## Phase 9: plan-list card simplification and theme consistency

- [x] 9.1 UI design documents define concise no-context plan cards with no source/language metadata and theme accent CTA（验证：`docs/ui-design/module-job-workspace.md`, `frontend/src`）
- [x] 9.2 `WorkspacePlanList` removes `workspace.planList.cardMeta`, `sourceType` and `targetLanguage` display from cards while preserving generated `listTargetJobs` navigation（验证：`WorkspaceEmptyState.test.tsx` red/green assertions）
- [x] 9.3 Formal frontend component tests catch metadata/secondary-button regression and verify card/page separation via existing theme tokens。

## Phase 10: plan-list bound resume navigation remediation

- [x] 10.1 `WorkspacePlanList` card navigation uses declared `currentPracticePlanId` / `resumeId` projection fields and never fabricates `plan-${targetJobId}` or `resume-unbound`（验证：`pnpm --filter @easyinterview/frontend test src/app/screens/workspace/WorkspaceEmptyState.test.tsx ...` PASS）
- [x] 10.2 Generated OpenAPI/TS TargetJob contract exposes current practice-plan binding for plan-list consumers（验证：`make codegen-openapi`; `pnpm --filter @easyinterview/frontend typecheck` PASS）

## Phase 11: target job-level resume binding remediation

- [x] 11.1 `WorkspacePlanList` opens detail with target job-level `resumeId` even when `currentPracticePlanId` is absent and no `practice_plans` row exists（验证：`WorkspaceEmptyState.test.tsx` PASS）
- [x] 11.2 `TargetJob.resumeId` contract is documented as the target job-level binding used by plan-list re-entry, with practice-plan projection only contributing `currentPracticePlanId`（验证：OpenAPI/generated types + `make validate-fixtures` PASS）

## Phase 12: unified detail route remediation

- [x] 12.1 Historical at the time: `workspace?targetJobId=...` detail re-entry was superseded and Parse rendered the `面试规划详情 / 面试上下文确认` mother page（当时验证：`ParseResumeBinding.test.tsx`；后续 Phase 16/17 再次 supersede route destination）
- [x] 12.2 Workspace `WorkspacePlanList` and plan-card navigation remain generated `listTargetJobs` backed, carrying declared `resumeId/currentPracticePlanId` only to `parse`（验证：`WorkspaceEmptyState.test.tsx`, `frontend/src/app/navigation/interviewContext.ts` tests PASS）
- [x] 12.3 `autoStartPractice=1` workspace ownership is superseded; parse/report handoff owns session start logic（验证：`ParseResumeBinding.test.tsx`, `ReplayCta.test.tsx` PASS）
- [x] 12.4 Pixel/formal implementation contract verifies workspace list + parse detail routing split across desktop/mobile and rejects out-of-scope independent workspace detail geometry（验证：`formal frontend component tests` PASS）

## Phase 13: plan-list admission and stale-context navigation remediation

- [x] 13.1 `useWorkspaceTargetJobs` requests `listTargetJobs` with `analysisStatus=ready` and ready page size, without ad hoc fetch（验证：`WorkspaceEmptyState.test.tsx` spy asserts query PASS）
- [x] 13.2 `WorkspacePlanList` defensively excludes failed / queued / processing / blank-title TargetJob records from visible cards（验证：`WorkspaceEmptyState.test.tsx` PASS）
- [x] 13.3 TopBar / out-of-scope-param `workspace` navigation clears or ignores stale detail `InterviewContext`, rendering `workspace-plan-list` instead of `parse-error` / “缺少目标岗位 ID”（验证：`WorkspaceScreen.test.tsx`, `App.test.tsx` PASS）

## Phase 14: workspace route purity remediation

- [x] 14.1 `WorkspaceScreen` is a pure list surface and no longer imports TargetJob detail/start/modal hooks（验证：source negative gate + `pnpm --filter @easyinterview/frontend test src/app/screens/workspace ...` PASS）
- [x] 14.2 `WorkspacePlanList` cards navigate to `parse`, while `App` clears InterviewContext whenever route name is `workspace`（验证：`WorkspaceEmptyState.test.tsx`, `App.test.tsx` PASS）
- [x] 14.3 Old workspace detail/start/modal runtime files and tests are removed from current owner（验证：`rg` negative + deleted files）
- [x] 14.4 Parse/report owners start practice directly through generated `getPracticePlan` / `createPracticePlan` / `startPracticeSession`, not through `workspace(autoStartPractice=1)`（验证：`ParseResumeBinding.test.tsx`, `ReplayCta.test.tsx` PASS）

## Phase 15: plan-list card size stability

- [x] 15.1 UI design document defines fixed desktop plan-card column sizing and rejects single-card full-row stretching（验证：`docs/ui-design/module-job-workspace.md`, `frontend/src`）
- [x] 15.2 Formal `WorkspacePlanList` uses `auto-fill` with fixed max column width and `justifyContent:start` on desktop, while compact layout remains single-column（验证：`WorkspaceScreen.test.tsx`）
- [x] 15.3 Browser screenshot acceptance captures the corrected single-card plan-list layout（验证：agent-browser screenshot）

## Phase 16: home recent / workspace list card fusion

- [x] 16.1 UI design document defines workspace plan-list card as Home recent card body plus workspace footer CTA（验证：`docs/ui-design/module-job-workspace.md`, `frontend/src`, `python3 scripts/lint/ui_demo_pruning.py` PASS）
- [x] 16.2 Formal `WorkspacePlanList` reuses the Home recent card body/mini round rail and appends `进入规划` / `Open plan` CTA without losing fixed-width grid behavior（验证：`pnpm --filter @easyinterview/frontend test src/app/screens/home/MockInterviewCard.test.tsx src/app/screens/home/HomeRecentMocks.test.tsx src/app/screens/workspace/WorkspaceScreen.test.tsx src/app/screens/workspace/WorkspaceEmptyState.test.tsx` PASS）
- [x] 16.3 Browser screenshot acceptance captures the fused workspace card and theme menu after the optimization（验证：agent-browser screenshot + `pnpm --filter @easyinterview/frontend test` PASS）

## Phase 17: plan-list action row and card-click planning

- [x] 17.1 UI design document defines workspace card body click as the planning-detail navigation, footer `立即面试`, and top-right resume-list trash icon delete（验证：`docs/ui-design/module-job-workspace.md`, `frontend/src`, `python3 scripts/lint/ui_demo_pruning.py`）
- [x] 17.2 Formal `MockInterviewCard` supports quick-start and top-right delete actions, stops action propagation, and uses the resume-list trash icon for delete（验证：`MockInterviewCard.test.tsx`）
- [x] 17.3 `WorkspacePlanList` removes visible `进入规划` footer button, starts practice through shared generated practice handoff with structured `roundId/roundName`, and keeps delete isolated from card navigation; Phase 18 owns backend-persistent archive（验证：`WorkspaceScreen.test.tsx`, `WorkspaceEmptyState.test.tsx` PASS）
- [x] 17.4 Home recent cards reuse the same quick-start action card and omit delete controls（验证：`HomeRecentMocks.test.tsx`）
- [x] 17.5 Formal component tests cover workspace card actions and Home recent card actions after the optimization；真实 UI screenshot 仅由明确场景或 acceptance run 产生。

## Phase 18: persistent TargetJob archive integration

- [x] 18.1 UI design document updates delete semantics from local-only hiding to persistent `archiveTargetJob`, with the delete icon fixed at the card top-right and footer kept for `立即面试` only; 验证: `docs/ui-design/module-job-workspace.md`, `frontend/src`, `python3 scripts/lint/ui_demo_pruning.py` PASS
- [x] 18.2 Generated client / mock transport expose and call `archiveTargetJob`; 验证: `make lint-openapi`, `make validate-fixtures`, `make lint-mock-contract`, generated `client.archiveTargetJob`
- [x] 18.3 `WorkspacePlanList` calls `archiveTargetJob` with `Idempotency-Key`, removes the card only on success, keeps the card on failure, and prevents top-right delete/quick-start events from bubbling to card navigation; 验证: `pnpm --filter @easyinterview/frontend test src/app/screens/home/MockInterviewCard.test.tsx src/app/screens/home/HomeRecentMocks.test.tsx src/app/screens/workspace/WorkspaceScreen.test.tsx src/app/screens/workspace/WorkspaceEmptyState.test.tsx` PASS, `pnpm --filter @easyinterview/frontend typecheck` PASS
- [x] 18.4 Home recent cards reuse the same card body and quick-start action but still omit delete controls; 验证: `HomeRecentMocks.test.tsx` PASS

## Phase 19: non-executable workspace scenario removal

  <!-- verified: 2026-07-10 method=workspace-dead-scenario-removal evidence="Deleted all 14 scenario files and empty directories with no placeholders; removed BDD matrix/details/commands, context discovery and scenario INDEX rows. Current-reference search retains only the Phase 19 removal record, while generic trigger-path pytest passes 3/3." -->

## Phase 20: auto-start context and implicit route-param removal

- [x] 20.1 Add focused red assertions that default InterviewContext has no auto-start field and start-practice output drops arbitrary/obsolete input keys.
- [x] 20.2 Delete the field/action/helper and build practice route params from an explicit current allowlist.
- [x] 20.3 仓库根 `make test` 完成前后端全量单测回归；typecheck、direct-start code regressions、owner/product contexts、docs、diff 与 pruning 作为独立 gates。

## Phase 21: test-only reducer action removal

- [x] 21.1 Add a focused source-surface red assertion for the five reducer actions with no production dispatch sites.
- [x] 21.2 Delete the unused action variants, reducer branches and self-only behavior tests; preserve all four runtime-used actions.
  <!-- verified: 2026-07-10 method=test-only-reducer-action-removal evidence="Deleted five action variants, five reducer branches and three self-only behavior tests. Non-test frontend source has zero references to the removed names; HYDRATE_FROM_ROUTE, MERGE_SESSION, INCREMENT_HINT_COUNT and CLEAR retain runtime dispatch sites." -->
- [x] 21.3 仓库根 `make test` 完成前后端全量单测回归；typecheck/build、dispatch inventory、owner/product contexts、docs、diff 与 pruning 作为独立 gates。
  <!-- verified: 2026-07-10 method=test-only-reducer-action-removal evidence="Frontend passes 138 files/839 tests, typecheck and build; main bundle is 656.01 kB. Removed-action production inventory is empty and all four retained actions have runtime dispatch sites. Owner/product contexts, docs/index/link/diff and pruning gates pass with real_residuals=0." -->

## Phase 22: unconsumed InterviewContext hook removal

- [x] 22.1 Add a focused source-surface RED assertion for the exported hook with zero repository consumers.
- [x] 22.2 Delete `useStartPracticeContext` without adding a replacement or changing InterviewContext state behavior.
- [x] 22.3 仓库根 `make test` 完成前后端全量单测回归；typecheck、symbol inventory、owner/product contexts、docs、diff 与 pruning 作为独立 gates。

## Phase 23: unreachable static Workspace detail removal

- [x] 23.1 Replace the old positive UI contract assertions with a focused RED gate requiring a pure plan-list prototype and zero old detail/modal/helper symbols.
  <!-- verified: 2026-07-10 method=workspace-static-detail-source-red evidence="UI contract ran 35 tests with exactly one failure: the new pure-list assertion rejected the old WorkspaceScreen params/requestAuth signature before reaching the old-symbol zero-residual loop; the other 34 prototype contracts passed." -->
- [x] 23.2 Delete the constant-false Workspace detail branch, all exclusive helpers and the unconsumed workspace-insight source/script; relocate the live Parse binding pill into its owner without changing the visible plan list/detail or shared resume-option provider.
  <!-- verified: 2026-07-10 method=unreachable-static-workspace-detail-removal evidence="screen-workspace.jsx shrank from 895 to 184 lines; deleted the 196-line workspace-insight source and its script entry, removed stale app props and both unused updated labels. Dependency inventory caught the live Parse window.BindingPill consumer, which was moved to a smaller local PlanBindingPill without the unused action branch. UI contract passes 35/35 and current UI source has zero old Workspace symbols/global binding coupling." -->
- [x] 23.3 Reconcile the active workspace/practice spec, history and spec INDEX to the current pure plan-list boundary with no positive company-insight contract.
  <!-- verified: 2026-07-10 method=workspace-pure-list-active-spec-reconcile evidence="Spec v1.32 now defines D-8 as list information density, removes the insight ownership row, and limits C-7 to parse/quick-start/report handoff; history v1.20 and spec INDEX are synchronized. Targeted old-positive wording search is zero, both owner contexts, docs/index/link and diff checks pass." -->

## Phase 24: structured round runtime consistency

- [x] 24.1 RED-GREEN: UI design document and focused contracts require the selected structured round duration instead of fixed `25:00`, and reject the fixed report `ROUND_ORDER` / default fallback（验证：`python3 scripts/lint/ui_demo_pruning.py`; `roundAssumptions.test.ts`; `ReplayCta.test.tsx`）
- [x] 24.2 RED-GREEN: shared start resolves `TargetJob.summary.interviewRounds[]`, sends the selected `durationMinutes` as `timeBudgetMinutes`, and reuses a baseline plan only when target/resume/time budget all match（验证：`buildCreatePlanRequest.test.ts`; `startPractice.test.ts`; Home/Workspace/Parse caller tests）
  <!-- verified: 2026-07-12 method=red-green evidence="RED: request stayed 30, stale plan reused, unknown round reached empty-plan crash. GREEN: shared request/start plus Home/Workspace/Parse/report callers pass 42/42; typecheck passes; stale 30-minute plans are recreated with the selected 50/60-minute round budget." -->
- [x] 24.3 RED-GREEN: Practice Top Bar reads the current `PracticePlan.timeBudgetMinutes`, formats arbitrary positive minute budgets, and never hard-codes `25:00`; missing/failed plan load does not fabricate a budget（验证：`pnpm --filter @easyinterview/frontend exec vitest run src/app/screens/practice/PracticeScreen.test.tsx`，6/6 PASS；`pnpm --filter @easyinterview/frontend typecheck` PASS）
- [x] 24.4 RED-GREEN: report next-round uses the immediate ordered successor and disables start for duplicate derived IDs, final/single/empty/unknown/loading round state and while either CTA start is in flight; repeated clicks create at most one plan/session（验证：`ReplayCta.test.tsx`、`useReportContextData.test.tsx`、`roundAssumptions.test.ts`，20/20 PASS；`pnpm --filter @easyinterview/frontend typecheck` PASS）
- [x] 24.6 仓库根 `make test` 完成前后端全量单测回归；typecheck/build、UI contract/parity、owner context、docs/index/diff 与 fixed `25:00`、`ROUND_ORDER`、default next-round fallback 负向搜索作为独立 gates。

## Phase 25: backend-persisted round progress and exact plan reuse

- [x] 25.1 RED-GREEN: UI contract/data/helpers require `practiceProgress`, reject `nextRound` and lifecycle-status/text fallback, preserve mini rail DOM/computed style/bounds/screenshots, and render final/invalid states correctly.
- [x] 25.2 RED-GREEN: strict mapper/navigation tests cover positive int32 strictly increasing but non-contiguous sequences (`1,2,4`), first→next-existing, final, missing/mismatched/duplicate/non-prefix facts and lifecycle-status independence; no `sequence + 1` assumption.
- [x] 25.3 RED-GREEN: create/start sends `roundId` without sequence; exact non-null pair is required for reuse; equal-duration wrong round, legacy null, stale baseline, final state and mismatched create response cannot start a session.
- [x] 25.4 RED-GREEN: Home/Workspace/Parse quick-start and Report next-round consume backend current progress and the next existing canonical successor; final/invalid buttons disabled with zero plan/session calls; retry-current remains server-validated。focused tests only provide development feedback.
- [x] 25.5 E2E-HANDOFF: `E2E.P0.098` 仅承接真实 completion/progress refresh；本轮未运行，current-run 状态仍为 `Ready`。
- [x] 25.6 Repository-root `make test` provides the frontend/backend unit regression; typecheck/build, UI parity, generated contract, contexts/docs/index/diff and browser-storage negative search remain separate gates.

## Phase 26: Workspace list/detail route split

- [x] 26.1 RED-GREEN: supersede Phase 14 pure-list assertions; query-free `/workspace` loads ready cards, while valid `/workspace?targetJobId` mounts the unified read-only detail.
- [x] 26.2 RED-GREEN: card body carries only targetJobId and directly enters workspace detail; non-safe plan/resume/auto-start params are ignored/stripped by shell routing.
- [x] 26.3 RED-GREEN: bottom transport spies under StrictMode prove list `listTargetJobs` count=1 and detail `getTargetJob` count=1 for same-key initial loads via shell/001 Phase 13.
- [x] 26.4 NEGATIVE: detail makes zero `importTargetJob`, zero Parse scheduler/poll after ready load, zero Parse animation DOM, and zero route-side practice start; mismatch/not-found fails closed.

## Phase 27: Workspace detail round-state affordance

- [x] 27.1 RED: focused component and UI source-contract tests fail while round cards have no persisted state attributes, labels or distinct visual treatments.
- [x] 27.2 GREEN: prototype and formal detail derive `done/current/pending` only from strict persisted progress, render localized labels/attributes with ok/accent/neutral tokens, and keep invalid projection neutral/non-startable.
- [x] 27.3 PARITY-Gate: desktop/mobile DOM, computed background/border, bbox, viewport overflow and screenshots prove formal implementation contract across valid states; dark/custom themes retain semantic distinction.<!-- verified: 2026-07-14 method=parse-responsive-browser result="desktop+mobile 2/2; source styles equal; distinct backgrounds/borders; bbox/no-overflow; screenshots" -->
- [x] 27.5 POST-PASS: 仓库根 `make test` 完成前后端全量单测回归；typecheck/build、docs/context/index/diff 与 lifecycle/URL/storage 负向搜索作为独立 gates；随后恢复 completed lifecycle。
