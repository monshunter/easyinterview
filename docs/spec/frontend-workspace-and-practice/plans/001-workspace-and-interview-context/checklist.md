# 001 Workspace + InterviewContext + Start Practice Contract Checklist

> **版本**: 1.37
> **状态**: active
> **更新日期**: 2026-07-12

**关联计划**: [plan](./plan.md)

## Phase 0: contract preflight

- [x] 0.1 `docs/development.md` §2 frontend/backend contract workflow is the execution boundary（验证：generated client + fixture-backed transport used; no ad hoc workspace fetch）
- [x] 0.2 UI truth source is current workspace prototype and docs（验证：`docs/ui-design/module-job-workspace.md`, `ui-design/src/screen-workspace.jsx`, `ui-design/src/app.jsx`, `ui-design/src/primitives.jsx`）
- [x] 0.3 Context manifest resolves current frontend target and spec version（验证：`validate_context.py frontend-workspace-and-practice/001 frontend` PASS）

## Phase 1: Workspace shell and InterviewContext

- [x] 1.1 `InterviewContextProvider` carries stable target/resume/round/plan/session IDs and `practiceGoal` across owner routes; mode/modality/hint fields are stripped（验证：`InterviewContext.test.tsx`, `App.test.tsx`）
- [x] 1.2 `workspace` route renders `WorkspaceScreen` instead of the route fallback shell; non-owner routes keep their own owners（验证：`App.test.tsx`）
- [x] 1.3 `workspace.*` zh/en messages and DOM anchors cover the pure plan-list eyebrow, title, cards, mini round rail, empty state and current actions（验证：`WorkspaceScreen.test.tsx`）
- [x] 1.4 BDD-Gate: `E2E.P0.018` covers workspace default render shell（验证：scenario trigger/verify）

## Phase 2: TargetJob, resume and workspace data

- [x] 2.1 Historical `useWorkspaceTargetJob` detail hook moved out of workspace owner; parse owner consumes generated `getTargetJob`（验证：source negative gate）
- [x] 2.2 Historical `useWorkspaceResume` detail hook moved out of workspace owner; parse owner consumes resume selection/list data（验证：`ParseResumeBinding.test.tsx`）
- [x] 2.3 Workspace list derives only declared `TargetJob` list fields and does not render detail header/launcher/JD breakdown（验证：`WorkspaceEmptyState.test.tsx`, source negative gate）
- [x] 2.4 BDD-Gate: `E2E.P0.018` covers workspace list and parse detail handoff; historical P0.019 detail loading belongs to external owner（验证：scenario trigger/verify）

## Phase 3: Plan and resume switching

- [x] 3.1 Historical `PlanSwitcherModal` runtime moved out of workspace owner; list cards now open parse detail（验证：source negative gate）
- [x] 3.2 Historical `ResumePickerModal` runtime moved out of workspace owner; parse detail owns resume selection（验证：`ParseResumeBinding.test.tsx`）
- [x] 3.3 Modal a11y is no longer a workspace gate; parse owner covers its picker behavior（验证：parse focused tests）
- [x] 3.4 BDD-Gate: `E2E.P0.018` covers workspace list and parse detail handoff（验证：scenario trigger/verify）

## Phase 4: Start practice and auth recovery

- [x] 4.1 Historical workspace practice-plan refresh hook moved out of this owner; parse/report handoff validates existing plan context（验证：`ParseResumeBinding.test.tsx`, `ReplayCta.test.tsx`）
- [x] 4.2 Shared `startPracticeFromParams` creates a plan when needed, starts a session, and uses stable idempotency keys for side effects（验证：`ParseResumeBinding.test.tsx`, `ReplayCta.test.tsx`）
- [x] 4.3 `workspace(autoStartPractice=1)` is removed from current runtime; pending actions no longer rely on workspace side effects（验证：source negative gate）
- [x] 4.4 BDD-Gate: start-practice behavior is covered by parse/report focused gates and external owner scenario（验证：focused tests）

## Phase 5: Workspace boundary and privacy

- [x] 5.1 Company insight component/API runtime stays outside workspace owner（验证：source negative gate）
- [x] 5.2 Records static affordance runtime stays outside workspace owner（验证：source negative gate）
- [x] 5.3 Workspace runtime does not import prototype data helpers or call report APIs for records static affordance（验证：`E2E.P0.021` verify grep）
- [x] 5.4 Sensitive fields are absent from URL, localStorage, console, telemetry and fixture transport logs（验证：privacy negative tests and scenario verify）
- [x] 5.5 BDD-Gate: `E2E.P0.021` covers embedded-only behavior, records static affordance and privacy/out-of-scope negative gates（验证：scenario trigger/verify）

## Phase 6: closeout

- [x] 6.1 Frontend focused tests passed for App, Workspace, Header, modals, start practice, auth and handoff（验证：owner focused Vitest suites）
- [x] 6.2 Pixel parity passed for workspace desktop/mobile and theme states（验证：`pnpm --filter @easyinterview/frontend test:pixel-parity`）
- [x] 6.3 Fixtures remain valid for TargetJobs, Resumes, PracticePlans and PracticeSessions（验证：`make validate-fixtures`）
- [x] 6.4 Owner docs/index/context are current and completed（验证：`validate_context.py frontend-workspace-and-practice/001 frontend`; `sync-doc-index --check`; `make docs-check`）

## Phase 7: interview nav and plan-list landing revision

- [x] 7.1 Product/UI truth sources and static prototype use TopBar `面试` / `Interview` and define `workspace` no-context plan-list landing（验证：`ui-design/src/app.jsx`, `ui-design/src/screen-workspace.jsx`, `docs/ui-design/module-job-workspace.md`, `docs/ui-design/ui-architecture.md`）
- [x] 7.2 Formal TopBar labels and i18n use `面试` / `Interview` while route/testid remains `workspace`（验证：`TopBar.test.tsx`, `TopBarVisual.test.tsx`, `p0-004-app-shell-language-switch.test.tsx`）
- [x] 7.3 `WorkspacePlanList` consumes generated `listTargetJobs`, renders loading/empty/error/list states, and plan cards navigate to `parse` detail without fabricating resume/report data（验证：`WorkspaceScreen.test.tsx`, `WorkspaceEmptyState.test.tsx`）
- [x] 7.4 Workspace parity and route regression gates distinguish no-context list landing from hydrated current-plan detail（验证：`frontend/tests/pixel-parity/workspace.spec.ts`, `p0-088-url-addressable-routing-canonical.test.tsx`, `p0-090-url-routing-hash-out-of-scope-negative.test.tsx`）
- [x] 7.5 BDD-Gate: `E2E.P0.018` covers TopBar `面试` landing, plan-list card selection, and existing current-plan detail anchors（验证：scenario trigger/verify）

## Phase 8: plan-list card visual hardening

- [x] 8.1 UI truth sources define the no-context plan list as visible list cards with card background, border, subtle elevation, internal body/footer sections, and responsive desktop/mobile grid（验证：`docs/ui-design/module-job-workspace.md`, `ui-design/src/screen-workspace.jsx`）
- [x] 8.2 `WorkspacePlanList` mirrors the card treatment and keeps generated `listTargetJobs` + safe navigation semantics unchanged（验证：`WorkspaceEmptyState.test.tsx` red/green assertions）
- [x] 8.3 Pixel parity catches loose text-column regression through computed style and bounding-box assertions for card, body and footer sections（验证：`frontend/tests/pixel-parity/workspace.spec.ts`）
- [x] 8.4 BDD-Gate: `E2E.P0.018` remains green after card visual hardening and continues to cover TopBar `面试` landing + plan-card selection（验证：scenario trigger/verify）

## Phase 9: plan-list card simplification and theme consistency

- [x] 9.1 UI truth sources define concise no-context plan cards with no source/language metadata and theme accent CTA（验证：`docs/ui-design/module-job-workspace.md`, `ui-design/src/screen-workspace.jsx`）
- [x] 9.2 `WorkspacePlanList` removes `workspace.planList.cardMeta`, `sourceType` and `targetLanguage` display from cards while preserving generated `listTargetJobs` navigation（验证：`WorkspaceEmptyState.test.tsx` red/green assertions）
- [x] 9.3 Pixel parity catches metadata/secondary-button regression and verifies card/page separation via existing theme tokens（验证：`frontend/tests/pixel-parity/workspace.spec.ts`）
- [x] 9.4 BDD-Gate: `E2E.P0.018` remains green after simplification and rejects source/language metadata returning to the no-context plan cards（验证：scenario trigger/verify）

## Phase 10: plan-list bound resume navigation remediation

- [x] 10.1 `WorkspacePlanList` card navigation uses declared `currentPracticePlanId` / `resumeId` projection fields and never fabricates `plan-${targetJobId}` or `resume-unbound`（验证：`pnpm --filter @easyinterview/frontend test src/app/screens/workspace/WorkspaceEmptyState.test.tsx ...` PASS）
- [x] 10.2 Generated OpenAPI/TS TargetJob contract exposes current practice-plan binding for plan-list consumers（验证：`make codegen-openapi`; `pnpm --filter @easyinterview/frontend typecheck` PASS）
- [x] 10.3 BDD-Gate: `E2E.P0.018` keeps plan-card selection on the parse bound-resume detail path（验证：focused equivalent `WorkspaceEmptyState.test.tsx`, `WorkspaceScreen.test.tsx` PASS）

## Phase 11: target job-level resume binding remediation

- [x] 11.1 `WorkspacePlanList` opens detail with target job-level `resumeId` even when `currentPracticePlanId` is absent and no `practice_plans` row exists（验证：`WorkspaceEmptyState.test.tsx` PASS）
- [x] 11.2 `TargetJob.resumeId` contract is documented as the target job-level binding used by plan-list re-entry, with practice-plan projection only contributing `currentPracticePlanId`（验证：OpenAPI/generated types + `make validate-fixtures` PASS）
- [x] 11.3 BDD-Gate: `E2E.P0.018` keeps plan-card selection on the bound-resume detail path for imported jobs without an existing practice plan（验证：focused equivalent workspace tests + local API smoke + `E2E.P0.018` scenario wrapper PASS）

## Phase 12: unified detail route remediation

- [x] 12.1 Historical `workspace?targetJobId=...` detail re-entry is superseded; parse route renders the `面试规划详情 / 面试上下文确认` mother page（验证：`ParseResumeBinding.test.tsx`, pixel parity parse detail PASS）
- [x] 12.2 Workspace `WorkspacePlanList` and plan-card navigation remain generated `listTargetJobs` backed, carrying declared `resumeId/currentPracticePlanId` only to `parse`（验证：`WorkspaceEmptyState.test.tsx`, `frontend/src/app/navigation/interviewContext.ts` tests PASS）
- [x] 12.3 `autoStartPractice=1` workspace ownership is superseded; parse/report handoff owns session start logic（验证：`ParseResumeBinding.test.tsx`, `ReplayCta.test.tsx` PASS）
- [x] 12.4 Pixel/source parity verifies workspace list + parse detail routing split across desktop/mobile and rejects out-of-scope independent workspace detail geometry（验证：`frontend/tests/pixel-parity/workspace.spec.ts` PASS）
- [x] 12.5 BDD-Gate: `E2E.P0.018` covers list re-entry to parse detail and parse/report focused gates cover direct practice start（验证：scenario trigger/verify PASS）

## Phase 13: plan-list admission and stale-context navigation remediation

- [x] 13.1 `useWorkspaceTargetJobs` requests `listTargetJobs` with `analysisStatus=ready` and ready page size, without ad hoc fetch（验证：`WorkspaceEmptyState.test.tsx` spy asserts query PASS）
- [x] 13.2 `WorkspacePlanList` defensively excludes failed / queued / processing / blank-title TargetJob records from visible cards（验证：`WorkspaceEmptyState.test.tsx` PASS）
- [x] 13.3 TopBar / out-of-scope-param `workspace` navigation clears or ignores stale detail `InterviewContext`, rendering `workspace-plan-list` instead of `parse-error` / “缺少目标岗位 ID”（验证：`WorkspaceScreen.test.tsx`, `App.test.tsx` PASS）
- [x] 13.4 BDD-Gate: `E2E.P0.018` covers failed/blank record exclusion and no-context TopBar landing after detail navigation（验证：focused equivalent PASS）

## Phase 14: workspace route purity remediation

- [x] 14.1 `WorkspaceScreen` is a pure list surface and no longer imports TargetJob detail/start/modal hooks（验证：source negative gate + `pnpm --filter @easyinterview/frontend test src/app/screens/workspace ...` PASS）
- [x] 14.2 `WorkspacePlanList` cards navigate to `parse`, while `App` clears InterviewContext whenever route name is `workspace`（验证：`WorkspaceEmptyState.test.tsx`, `App.test.tsx` PASS）
- [x] 14.3 Old workspace detail/start/modal runtime files and tests are removed from current owner（验证：`rg` negative + deleted files）
- [x] 14.4 Parse/report owners start practice directly through generated `getPracticePlan` / `createPracticePlan` / `startPracticeSession`, not through `workspace(autoStartPractice=1)`（验证：`ParseResumeBinding.test.tsx`, `ReplayCta.test.tsx` PASS）
- [x] 14.5 BDD-Gate: `E2E.P0.018` trigger/verify now covers workspace pure list + parse detail handoff and rejects out-of-scope workspace context files（验证：scenario assets updated）

## Phase 15: plan-list card size stability

- [x] 15.1 UI truth source defines fixed desktop plan-card column sizing and rejects single-card full-row stretching（验证：`docs/ui-design/module-job-workspace.md`, `ui-design/src/screen-workspace.jsx`）
- [x] 15.2 Formal `WorkspacePlanList` uses `auto-fill` with fixed max column width and `justifyContent:start` on desktop, while compact layout remains single-column（验证：`WorkspaceScreen.test.tsx`）
- [x] 15.3 Browser screenshot acceptance captures the corrected single-card plan-list layout（验证：agent-browser screenshot）

## Phase 16: home recent / workspace list card fusion

- [x] 16.1 UI truth source defines workspace plan-list card as Home recent card body plus workspace footer CTA（验证：`docs/ui-design/module-job-workspace.md`, `ui-design/src/screen-workspace.jsx`, `node --test ui-design/ui-design-contract.test.mjs` PASS）
- [x] 16.2 Formal `WorkspacePlanList` reuses the Home recent card body/mini round rail and appends `进入规划` / `Open plan` CTA without losing fixed-width grid behavior（验证：`pnpm --filter @easyinterview/frontend test src/app/screens/home/MockInterviewCard.test.tsx src/app/screens/home/HomeRecentMocks.test.tsx src/app/screens/workspace/WorkspaceScreen.test.tsx src/app/screens/workspace/WorkspaceEmptyState.test.tsx` PASS）
- [x] 16.3 Browser screenshot acceptance captures the fused workspace card and theme menu after the optimization（验证：agent-browser screenshot + `pnpm --filter @easyinterview/frontend test:pixel-parity tests/pixel-parity/workspace.spec.ts` PASS）

## Phase 17: plan-list action row and card-click planning

- [x] 17.1 UI truth source defines workspace card body click as the planning-detail navigation, footer `立即面试`, and top-right resume-list trash icon delete（验证：`docs/ui-design/module-job-workspace.md`, `ui-design/src/screen-workspace.jsx`, `node --test ui-design/ui-design-contract.test.mjs`）
- [x] 17.2 Formal `MockInterviewCard` supports quick-start and top-right delete actions, stops action propagation, and uses the resume-list trash icon for delete（验证：`MockInterviewCard.test.tsx`）
- [x] 17.3 `WorkspacePlanList` removes visible `进入规划` footer button, starts practice through shared generated practice handoff with structured `roundId/roundName`, and keeps delete isolated from card navigation; Phase 18 owns backend-persistent archive（验证：`WorkspaceScreen.test.tsx`, `WorkspaceEmptyState.test.tsx` PASS）
- [x] 17.4 Home recent cards reuse the same quick-start action card and omit delete controls（验证：`HomeRecentMocks.test.tsx`）
- [x] 17.5 Browser screenshot acceptance captures workspace card actions and Home recent card actions after the optimization（验证：`.test-output/screenshots/workspace-plan-list-action-card.png`, `.test-output/screenshots/home-recent-action-card.png`, pixel parity workspace spec）

## Phase 18: persistent TargetJob archive integration

- [x] 18.1 UI truth source updates delete semantics from local-only hiding to persistent `archiveTargetJob`, with the delete icon fixed at the card top-right and footer kept for `立即面试` only; 验证: `docs/ui-design/module-job-workspace.md`, `ui-design/src/screen-workspace.jsx`, `node --test ui-design/ui-design-contract.test.mjs` PASS
- [x] 18.2 Generated client / mock transport expose and call `archiveTargetJob`; 验证: `make lint-openapi`, `make validate-fixtures`, `make lint-mock-contract`, generated `client.archiveTargetJob`
- [x] 18.3 `WorkspacePlanList` calls `archiveTargetJob` with `Idempotency-Key`, removes the card only on success, keeps the card on failure, and prevents top-right delete/quick-start events from bubbling to card navigation; 验证: `pnpm --filter @easyinterview/frontend test src/app/screens/home/MockInterviewCard.test.tsx src/app/screens/home/HomeRecentMocks.test.tsx src/app/screens/workspace/WorkspaceScreen.test.tsx src/app/screens/workspace/WorkspaceEmptyState.test.tsx` PASS, `pnpm --filter @easyinterview/frontend typecheck` PASS
- [x] 18.4 Home recent cards reuse the same card body and quick-start action but still omit delete controls; 验证: `HomeRecentMocks.test.tsx` PASS
- [x] 18.5 BDD-Gate: real-backend browser smoke proves archived TargetJob disappears after refresh; 验证: `test/scenarios/e2e/p0-018-workspace-default-render/scripts/setup.sh && .../trigger.sh && .../verify.sh && .../cleanup.sh` PASS；local real-backend browser smoke captured `.test-output/e2e/workspace-archive-real-browser/workspace-card-before-delete.png` and `.test-output/e2e/workspace-archive-real-browser/workspace-after-delete.png`; DB readback `archive-db-state.txt=archived|t`, refresh text excludes the target title/id, cleanup `cleanup-db-state.txt=0`

## Phase 19: non-executable workspace scenario removal

- [x] 19.1 Add a generic scenario trigger-path contract test; red evidence must identify the five missing frontend test paths referenced by P0.019/P0.020.
  <!-- verified: 2026-07-10 method=workspace-dead-scenario-removal evidence="After correcting longest-suffix matching for .tsx, the generic pytest red reported exactly the three removed workspace hook tests in P0.019 and two removed start/auth tests in P0.020; all other explicit frontend trigger paths resolved." -->
- [x] 19.2 Delete P0.019/P0.020 scenario packages and remove their current BDD/context/index references; keep current behavior owned by P0.018, P0.021 and parse/report focused gates.
  <!-- verified: 2026-07-10 method=workspace-dead-scenario-removal evidence="Deleted all 14 scenario files and empty directories with no placeholders; removed BDD matrix/details/commands, context discovery and scenario INDEX rows. Current-reference search retains only the Phase 19 removal record, while generic trigger-path pytest passes 3/3." -->
- [x] 19.3 Run the generic contract suite, current start-practice tests, P0.018/P0.021 wrappers, owner/product contexts, docs, diff and pruning gates.
  <!-- verified: 2026-07-10 method=workspace-dead-scenario-removal evidence="Generic contract passes 3/3 and full scripts/lint passes 294 tests plus 4248 subtests. Parse/report direct-start passes 10/10; P0.018 passes real-mode 1/1 plus 57/57, P0.021 passes real-mode 1/1 plus 10/10. Workspace/shell/product contexts, docs/index/link/diff and pruning gates pass with real_residuals=0." -->

## Phase 20: auto-start context and implicit route-param removal

- [x] 20.1 Add focused red assertions that default InterviewContext has no auto-start field and start-practice output drops arbitrary/obsolete input keys.
  <!-- verified: 2026-07-10 method=auto-start-context-and-implicit-param-removal evidence="Focused red ran 19 tests with exactly two failures: default context still owned autoStartPractice, and startPractice output leaked sourceSessionId/replayItems/evidenceGaps/rawText from arbitrary input spreading." -->
- [x] 20.2 Delete the field/action/helper and build practice route params from an explicit current allowlist.
  <!-- verified: 2026-07-10 method=auto-start-context-and-implicit-param-removal evidence="Removed InterviewContext field/default/hydration/CLEAR action and withoutAutoStart helper. startPractice output now lists normalized target/job/JD/resume/source-report/round/mode/practice/hint/language plus new plan/session IDs explicitly; focused tests pass 19/19." -->
- [x] 20.3 Run focused/full frontend tests, typecheck, current direct-start/scenario regressions, owner/product contexts, docs, diff and pruning gates.
  <!-- verified: 2026-07-10 method=auto-start-context-and-implicit-param-removal evidence="Focused context/output tests pass 19/19; four direct-start callers pass 29/29; full frontend passes 138 files/841 tests with typecheck/build green and main bundle 656.63 kB. P0.018/P0.021/P0.057/P0.088 pass 57/10/20/16 tests plus real-mode gates and cleanup. Owner/product contexts, docs/index/link/diff and pruning gates pass with real_residuals=0." -->

## Phase 21: test-only reducer action removal

- [x] 21.1 Add a focused source-surface red assertion for the five reducer actions with no production dispatch sites.
  <!-- verified: 2026-07-10 method=test-only-reducer-action-source-red evidence="Focused InterviewContext test reached the source-surface assertion and failed on MERGE_TARGET_JOB, proving the five production-dispatch-free action names remain in InterviewContext.tsx before removal." -->
- [x] 21.2 Delete the unused action variants, reducer branches and self-only behavior tests; preserve all four runtime-used actions.
  <!-- verified: 2026-07-10 method=test-only-reducer-action-removal evidence="Deleted five action variants, five reducer branches and three self-only behavior tests. Non-test frontend source has zero references to the removed names; HYDRATE_FROM_ROUTE, MERGE_SESSION, INCREMENT_HINT_COUNT and CLEAR retain runtime dispatch sites." -->
- [x] 21.3 Run focused/full frontend tests, typecheck/build, dispatch inventory, owner/product contexts, docs, diff and pruning gates.
  <!-- verified: 2026-07-10 method=test-only-reducer-action-removal evidence="Frontend passes 138 files/839 tests, typecheck and build; main bundle is 656.01 kB. Removed-action production inventory is empty and all four retained actions have runtime dispatch sites. Owner/product contexts, docs/index/link/diff and pruning gates pass with real_residuals=0." -->

## Phase 22: unconsumed InterviewContext hook removal

- [x] 22.1 Add a focused source-surface RED assertion for the exported hook with zero repository consumers.
  <!-- verified: 2026-07-10 method=unconsumed-context-hook-source-red evidence="Focused InterviewContext failed 1/17 solely because InterviewContext.tsx still exported useStartPracticeContext; repository AST/text inventory found no production or test consumer." -->
- [x] 22.2 Delete `useStartPracticeContext` without adding a replacement or changing InterviewContext state behavior.
  <!-- verified: 2026-07-10 method=unconsumed-context-hook-removal evidence="Deleted the function and its false consumer comment with no replacement. Focused InterviewContext passes 17/17 and non-test frontend symbol search is empty; state/reducer/provider code is unchanged." -->
- [x] 22.3 Run focused/full frontend tests, typecheck, symbol inventory, owner/product contexts, docs, diff and pruning gates.
  <!-- verified: 2026-07-10 method=unconsumed-context-hook-removal evidence="Focused InterviewContext passes 17/17; full frontend passes 138 files/841 tests with zero React update warning; typecheck and non-test symbol inventory pass. Owner/product contexts and docs/index/link/diff/pruning gates pass with real_residuals=0." -->

## Phase 23: unreachable static Workspace detail removal

- [x] 23.1 Replace the old positive UI contract assertions with a focused RED gate requiring a pure plan-list prototype and zero old detail/modal/helper symbols.
  <!-- verified: 2026-07-10 method=workspace-static-detail-source-red evidence="UI contract ran 35 tests with exactly one failure: the new pure-list assertion rejected the old WorkspaceScreen params/requestAuth signature before reaching the old-symbol zero-residual loop; the other 34 prototype contracts passed." -->
- [x] 23.2 Delete the constant-false Workspace detail branch, all exclusive helpers and the unconsumed workspace-insight source/script; relocate the live Parse binding pill into its owner without changing the visible plan list/detail or shared resume-option provider.
  <!-- verified: 2026-07-10 method=unreachable-static-workspace-detail-removal evidence="screen-workspace.jsx shrank from 895 to 184 lines; deleted the 196-line workspace-insight source and its script entry, removed stale app props and both unused updated labels. Dependency inventory caught the live Parse window.BindingPill consumer, which was moved to a smaller local PlanBindingPill without the unused action branch. UI contract passes 35/35 and current UI source has zero old Workspace symbols/global binding coupling." -->
- [x] 23.3 Reconcile the active workspace/practice spec, history and spec INDEX to the current pure plan-list boundary with no positive company-insight contract.
  <!-- verified: 2026-07-10 method=workspace-pure-list-active-spec-reconcile evidence="Spec v1.32 now defines D-8 as list information density, removes the insight ownership row, and limits C-7 to parse/quick-start/report handoff; history v1.20 and spec INDEX are synchronized. Targeted old-positive wording search is zero, both owner contexts, docs/index/link and diff checks pass." -->
- [x] 23.4 Run UI contract, source inventory, formal workspace tests, P0.018, owner/product contexts and docs/diff/pruning gates.
  <!-- verified: 2026-07-10 method=workspace-static-pruning-regression evidence="UI contract passes 35/35; old UI-source symbols and active-spec positive insight wording are zero. Formal Workspace/Parse passes 24/24, typecheck/build, full frontend 137 files/841 tests, desktop/mobile Playwright 44/44, and direct static-browser Workspace/Parse checks pass with no page errors or horizontal overflow. P0.018 passes real-mode 1/1 plus 57/57 and cleanup. Both owner contexts, docs/index/link/diff and pruning gates pass with real_residuals=0; no environment restart or data cleanup occurred." -->
