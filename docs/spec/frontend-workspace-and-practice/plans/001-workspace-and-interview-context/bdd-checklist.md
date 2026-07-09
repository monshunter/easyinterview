# 001 BDD Checklist

> **版本**: 1.13
> **状态**: active
> **更新日期**: 2026-07-09

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## E2E.P0.018 面试入口规划列表 + parse 统一面试规划详情

- [x] Scenario assets exist under `test/scenarios/e2e/p0-018-workspace-default-render/`
- [x] Given fixtures cover `listTargetJobs` candidates plus parse detail resume/start fixtures
- [x] Trigger runs App, Workspace list and Parse resume/start handoff tests
- [x] Verify covers workspace DOM anchors, ready-list query, parse navigation, retired detail/start/modal negative grep and non-current testid negative grep
- [x] Scenario `setup -> trigger -> verify -> cleanup` passes
- [x] Revision 2026-07-08 trigger covers TopBar `面试` / `Interview`, no-context `WorkspacePlanList`, plan-card selection, and hydrated current-plan detail
- [x] Revision 2026-07-08 verify covers `workspace-plan-list-*` anchors, absence of `workspace-empty` on no-context landing, and updated scenario evidence
- [x] Revision 2026-07-08 card visual hardening covers plan-list card background, border, elevation, body/footer sections and responsive geometry
- [x] Revision 2026-07-08 card simplification covers removal of source/language metadata and theme accent `进入规划` CTA on no-context plan cards
- [x] Revision 2026-07-09 trigger covers plan-card selection into the parse unified `面试规划详情 / 面试上下文确认` mother page instead of independent workspace detail.
- [x] Revision 2026-07-09 verify covers parse detail marker, absence of `workspace-header` / `workspace-launcher` / `workspace-jd-card` independent detail anchors, active resume binding and privacy.
- [x] Revision 2026-07-09 trigger covers `analysisStatus=ready` list query, failed/blank TargetJob exclusion, and TopBar / legacy-param `workspace` landing after a detail page.
- [x] Revision 2026-07-09 verify covers absence of dirty failed JD cards, absence of `parse-error` / “缺少目标岗位 ID” on workspace, updated scenario evidence, and workspace retired context negative grep.
- [x] Revision 2026-07-09 trigger covers generated `archiveTargetJob` delete from the workspace card top-right icon and verifies delete does not bubble to card navigation.
- [x] Revision 2026-07-09 verify covers archived TargetJob absent after workspace refresh, delete failure preserving the card, and screenshot evidence for the post-delete list.

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
- [x] Revision 2026-07-09 trigger covers unified detail Start and verifies parse/report owners call `createPracticePlan` / `startPracticeSession` directly.
- [x] Revision 2026-07-09 verify covers no `workspace(autoStartPractice=1)` session-start dependency and no sensitive param leakage.

## E2E.P0.021 Embedded insight + records placeholder + privacy/non-current negative

- [x] Scenario assets exist under `test/scenarios/e2e/p0-021-workspace-handoff/`
- [x] Given fixtures cover ready workspace data without untyped records extension
- [x] Trigger runs workspace source negative and external report handoff regression tests
- [x] Verify covers embedded-only insight, records placeholder, no report API call, no prototype helper import, privacy field negative grep and non-current testid negative grep
- [x] Scenario `setup -> trigger -> verify -> cleanup` passes

## Closeout

- [ ] `validate_context.py --context docs/spec/frontend-workspace-and-practice/plans/001-workspace-and-interview-context/context.yaml --target frontend` passes
- [ ] `sync-doc-index --check` passes
- [ ] `make docs-check` passes
- [ ] `git diff --check` passes
