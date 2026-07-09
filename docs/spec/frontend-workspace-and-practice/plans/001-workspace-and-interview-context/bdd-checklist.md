# 001 BDD Checklist

> **版本**: 1.9
> **状态**: completed
> **更新日期**: 2026-07-09

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## E2E.P0.018 面试入口规划列表 + 统一面试规划详情 + active Resume Picker

- [x] Scenario assets exist under `test/scenarios/e2e/p0-018-workspace-default-render/`
- [x] Given fixtures cover `getTargetJob=with-rounds`, `getResume=default`, `getPracticePlan=default(ready)`, and `listTargetJobs` candidates
- [x] Trigger runs App, Workspace, Header, modal integration, Plan Switcher, Resume Picker and modal a11y tests
- [x] Verify covers workspace DOM anchors, Plan Switcher `listTargetJobs`, Resume Picker flat `listResumes` active-list, modal a11y and non-current testid negative grep
- [x] Scenario `setup -> trigger -> verify -> cleanup` passes
- [x] Revision 2026-07-08 trigger covers TopBar `面试` / `Interview`, no-context `WorkspacePlanList`, plan-card selection, and hydrated current-plan detail
- [x] Revision 2026-07-08 verify covers `workspace-plan-list-*` anchors, absence of `workspace-empty` on no-context landing, and updated scenario evidence
- [x] Revision 2026-07-08 card visual hardening covers plan-list card background, border, elevation, body/footer sections and responsive geometry
- [x] Revision 2026-07-08 card simplification covers removal of source/language metadata and theme accent `进入规划` CTA on no-context plan cards
- [x] Revision 2026-07-09 trigger covers plan-card selection into the unified `面试规划详情 / 面试上下文确认` mother page instead of the independent workspace detail.
- [x] Revision 2026-07-09 verify covers shared detail marker, absence of `workspace-header` / `workspace-launcher` / `workspace-jd-card` independent detail anchors, active resume binding and privacy.

## E2E.P0.019 Workspace context loading

- [x] Scenario assets exist under `test/scenarios/e2e/p0-019-workspace-context-loading/`
- [x] Given fixtures cover TargetJob ready/not-found/5xx, Resume ready/not-found and PracticePlan ready/not-found variants
- [x] Trigger runs route hydration, TargetJob, Resume and PracticePlan hook coverage
- [x] Verify covers empty state, missing-resume state, plan refresh recovery, retry UI and privacy negative grep
- [x] Scenario `setup -> trigger -> verify -> cleanup` passes

## E2E.P0.020 立即面试 + idempotency + auth recovery

- [x] Scenario assets exist under `test/scenarios/e2e/p0-020-workspace-start-practice/`
- [x] Given fixtures cover create/start success, validation failure and AI timeout failure variants
- [x] Trigger runs start-practice and auth-gate tests
- [x] Verify covers create-then-start, ready-plan direct start, idempotency retry, pendingAction auto-start and sensitive-param negative assertions
- [x] Scenario `setup -> trigger -> verify -> cleanup` passes
- [x] Revision 2026-07-09 trigger covers `autoStartPractice=1` after unified detail Start and verifies workspace still owns `createPracticePlan` / `startPracticeSession`.
- [x] Revision 2026-07-09 verify covers no duplicated session start logic in unified detail and no sensitive param leakage.

## E2E.P0.021 Embedded insight + records placeholder + privacy/non-current negative

- [x] Scenario assets exist under `test/scenarios/e2e/p0-021-workspace-handoff/`
- [x] Given fixtures cover ready workspace data without untyped records extension
- [x] Trigger runs WorkspaceHandoff and WorkspaceScreen regression tests
- [x] Verify covers embedded-only insight, records placeholder, no report API call, no prototype helper import, privacy field negative grep and non-current testid negative grep
- [x] Scenario `setup -> trigger -> verify -> cleanup` passes

## Closeout

- [x] `validate_context.py --context docs/spec/frontend-workspace-and-practice/plans/001-workspace-and-interview-context/context.yaml --target frontend` passes
- [x] `sync-doc-index --check` passes
- [x] `make docs-check` passes
- [x] `git diff --check` passes
