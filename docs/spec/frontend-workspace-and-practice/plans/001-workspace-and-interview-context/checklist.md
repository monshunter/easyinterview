# 001 Workspace + InterviewContext + Start Practice Contract Checklist

> **版本**: 1.16
> **状态**: completed
> **更新日期**: 2026-07-09

**关联计划**: [plan](./plan.md)

## Phase 0: contract preflight

- [x] 0.1 `docs/development.md` §2 frontend/backend contract workflow is the execution boundary（验证：generated client + fixture-backed transport used; no ad hoc workspace fetch）
- [x] 0.2 UI truth source is current workspace prototype and docs（验证：`docs/ui-design/module-job-workspace.md`, `ui-design/src/screen-workspace.jsx`, `ui-design/src/app.jsx`, `ui-design/src/primitives.jsx`）
- [x] 0.3 Context manifest resolves current frontend target and spec version（验证：`validate_context.py frontend-workspace-and-practice/001 frontend` PASS）

## Phase 1: Workspace shell and InterviewContext

- [x] 1.1 `InterviewContextProvider` carries `targetJobId / jdId / resumeId / roundId / planId / practiceMode / practiceGoal / hintUsed / hintCount` across owner routes（验证：`InterviewContext.test.tsx`, `App.test.tsx`）
- [x] 1.2 `workspace` route renders `WorkspaceScreen` instead of placeholder; non-owner routes keep their own owners（验证：`App.test.tsx`）
- [x] 1.3 `workspace.*` zh/en messages and DOM anchors cover plan eyebrow, header, launcher, bindings, insight, requirements, preparation and records area（验证：`WorkspaceScreen.test.tsx`）
- [x] 1.4 BDD-Gate: `E2E.P0.018` covers workspace default render shell（验证：scenario trigger/verify）

## Phase 2: TargetJob, resume and workspace data

- [x] 2.1 `useWorkspaceTargetJob` consumes generated `getTargetJob` and handles loading, ready, not-found and retry states（验证：`WorkspaceHeader.test.tsx`, hook tests）
- [x] 2.2 `useWorkspaceResume` consumes generated `getResume`, binds `resumeId`, and renders missing-resume state when needed（验证：`WorkspaceEmptyState.test.tsx`, `useWorkspaceResume.test.tsx`）
- [x] 2.3 Header, launcher, JD breakdown and preparation signals derive only from declared `TargetJob` fields（验证：field-mapping and unsupported-field negative tests）
- [x] 2.4 BDD-Gate: `E2E.P0.019` covers context loading, empty state, missing resume and plan refresh（验证：scenario trigger/verify）

## Phase 3: Plan and resume switching

- [x] 3.1 `PlanSwitcherModal` consumes generated `listTargetJobs`, switches current plan and supports new-plan handoff to home（验证：`PlanSwitcherModal.test.tsx`, `WorkspaceModalIntegration.test.tsx`）
- [x] 3.2 `ResumePickerModal` consumes current flat `listResumes` active-list and dispatches selected `resumeId` back into `InterviewContext`（验证：`ResumePickerModal.test.tsx`）
- [x] 3.3 Modal a11y supports ESC, backdrop, close button, focus trap and focus return（验证：`useModalA11y.test.tsx`）
- [x] 3.4 BDD-Gate: `E2E.P0.018` covers plan switcher and active resume picker（验证：scenario trigger/verify）

## Phase 4: Start practice and auth recovery

- [x] 4.1 `useWorkspacePracticePlan` refreshes ready plans and clears absent plans before launch（验证：hook tests and `WorkspaceStartPractice.test.tsx`）
- [x] 4.2 `useStartPractice` creates a baseline plan when needed, starts a session, and uses stable idempotency keys for side effects and retry（验证：`WorkspaceStartPractice.test.tsx`）
- [x] 4.3 Unauthenticated start uses `requestAuth({ type: "start_practice" })`, returns to workspace, clears `autoStartPractice`, and resumes launch（验证：`WorkspaceAuthGate.test.tsx`）
- [x] 4.4 BDD-Gate: `E2E.P0.020` covers happy path, retry and pendingAction recovery（验证：scenario trigger/verify）

## Phase 5: Embedded insight, records placeholder and privacy

- [x] 5.1 `WorkspaceInsightCard` stays embedded and does not call a standalone company signal API（验证：`WorkspaceHandoff.test.tsx`）
- [x] 5.2 Records area renders placeholder only and does not synthesize report rows from untyped fixture data（验证：`WorkspaceHandoff.test.tsx`）
- [x] 5.3 Workspace runtime does not import prototype data helpers or call report APIs for records placeholder（验证：`E2E.P0.021` verify grep）
- [x] 5.4 Sensitive fields are absent from URL, localStorage, console, telemetry and fixture transport logs（验证：privacy negative tests and scenario verify）
- [x] 5.5 BDD-Gate: `E2E.P0.021` covers embedded-only behavior, records placeholder and privacy/non-current negative gates（验证：scenario trigger/verify）

## Phase 6: closeout

- [x] 6.1 Frontend focused tests passed for App, Workspace, Header, modals, start practice, auth and handoff（验证：owner focused Vitest suites）
- [x] 6.2 Pixel parity passed for workspace desktop/mobile and theme states（验证：`pnpm --filter @easyinterview/frontend test:pixel-parity`）
- [x] 6.3 Fixtures remain valid for TargetJobs, Resumes, PracticePlans and PracticeSessions（验证：`make validate-fixtures`）
- [x] 6.4 Owner docs/index/context are current and completed（验证：`validate_context.py frontend-workspace-and-practice/001 frontend`; `sync-doc-index --check`; `make docs-check`）

## Phase 7: interview nav and plan-list landing revision

- [x] 7.1 Product/UI truth sources and static prototype use TopBar `面试` / `Interview` and define `workspace` no-context plan-list landing（验证：`ui-design/src/app.jsx`, `ui-design/src/screen-workspace.jsx`, `docs/ui-design/module-job-workspace.md`, `docs/ui-design/ui-architecture.md`）
- [x] 7.2 Formal TopBar labels and i18n use `面试` / `Interview` while route/testid remains `workspace`（验证：`TopBar.test.tsx`, `TopBarVisual.test.tsx`, `p0-004-app-shell-language-switch.test.tsx`）
- [x] 7.3 `WorkspacePlanList` consumes generated `listTargetJobs`, renders loading/empty/error/list states, and plan cards navigate to current-plan detail without fabricating resume/report data（验证：`WorkspaceScreen.test.tsx`, `WorkspaceEmptyState.test.tsx`, `WorkspaceHandoff.test.tsx`）
- [x] 7.4 Workspace parity and route regression gates distinguish no-context list landing from hydrated current-plan detail（验证：`frontend/tests/pixel-parity/workspace.spec.ts`, `p0-088-url-addressable-routing-canonical.test.tsx`, `p0-090-url-routing-hash-non-current-negative.test.tsx`）
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
- [x] 10.3 BDD-Gate: `E2E.P0.018` keeps plan-card selection on the current bound-resume detail path（验证：focused equivalent `WorkspaceEmptyState.test.tsx`, `WorkspaceScreen.test.tsx`, `WorkspaceHandoff.test.tsx` PASS）

## Phase 11: target job-level resume binding remediation

- [x] 11.1 `WorkspacePlanList` opens detail with target job-level `resumeId` even when `currentPracticePlanId` is absent and no `practice_plans` row exists（验证：`WorkspaceEmptyState.test.tsx` PASS）
- [x] 11.2 `TargetJob.resumeId` contract is documented as the target job-level binding used by plan-list re-entry, with practice-plan projection only contributing `currentPracticePlanId`（验证：OpenAPI/generated types + `make validate-fixtures` PASS）
- [x] 11.3 BDD-Gate: `E2E.P0.018` keeps plan-card selection on the bound-resume detail path for imported jobs without an existing practice plan（验证：focused equivalent workspace tests + local API smoke + `E2E.P0.018` scenario wrapper PASS）
