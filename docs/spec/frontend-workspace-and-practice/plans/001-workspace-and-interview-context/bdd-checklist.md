# 001 BDD Checklist

> **版本**: 1.4
> **状态**: completed
> **更新日期**: 2026-07-07

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## E2E.P0.018 Workspace 默认渲染 + Plan Switcher + active Resume Picker

- [x] Scenario assets exist under `test/scenarios/e2e/p0-018-workspace-default-render/`
- [x] Given fixtures cover `getTargetJob=with-rounds`, `getResume=default`, `getPracticePlan=default(ready)`, and `listTargetJobs` candidates
- [x] Trigger runs App, Workspace, Header, modal integration, Plan Switcher, Resume Picker and modal a11y tests
- [x] Verify covers workspace DOM anchors, Plan Switcher `listTargetJobs`, Resume Picker flat `listResumes` active-list, modal a11y and non-current testid negative grep
- [x] Scenario `setup -> trigger -> verify -> cleanup` passes

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
