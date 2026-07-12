# 001 BDD Checklist

> **版本**: 1.20
> **状态**: completed
> **更新日期**: 2026-07-12

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## E2E.P0.018 面试入口规划列表 + parse 统一面试规划详情

- [x] Scenario assets exist under `test/scenarios/e2e/p0-018-workspace-default-render/`
- [x] Given fixtures cover `listTargetJobs` candidates plus parse detail resume/start fixtures
- [x] Trigger runs App, Workspace list and Parse resume/start handoff tests
- [x] Verify covers workspace DOM anchors, ready-list query, parse navigation, out-of-scope detail/start/modal negative grep and out-of-scope testid negative grep
- [x] Scenario `setup -> trigger -> verify -> cleanup` passes
- [x] Revision 2026-07-08 trigger covers TopBar `面试` / `Interview`, no-context `WorkspacePlanList`, plan-card selection, and hydrated current-plan detail
- [x] Revision 2026-07-08 verify covers `workspace-plan-list-*` anchors, absence of `workspace-empty` on no-context landing, and updated scenario evidence
- [x] Revision 2026-07-08 card visual hardening covers plan-list card background, border, elevation, body/footer sections and responsive geometry
- [x] Revision 2026-07-08 card simplification covers removal of source/language metadata and theme accent `进入规划` CTA on no-context plan cards
- [x] Revision 2026-07-09 trigger covers plan-card selection into the parse unified `面试规划详情 / 面试上下文确认` mother page instead of independent workspace detail.
- [x] Revision 2026-07-09 verify covers parse detail marker, absence of `workspace-header` / `workspace-launcher` / `workspace-jd-card` independent detail anchors, active resume binding and privacy.
- [x] Revision 2026-07-09 trigger covers `analysisStatus=ready` list query, failed/blank TargetJob exclusion, and TopBar / out-of-scope-param `workspace` landing after a detail page.
- [x] Revision 2026-07-09 verify covers absence of dirty failed JD cards, absence of `parse-error` / “缺少目标岗位 ID” on workspace, updated scenario evidence, and out-of-scope workspace context negative grep.
- [x] Revision 2026-07-09 trigger covers generated `archiveTargetJob` delete from the workspace card top-right icon and verifies delete does not bubble to card navigation.
- [x] Revision 2026-07-09 verify covers archived TargetJob absent after workspace refresh, delete failure preserving the card, and screenshot evidence for the post-delete list.

## E2E.P0.021 Workspace boundary + privacy/out-of-scope negative

- [x] Scenario assets exist under `test/scenarios/e2e/p0-021-workspace-handoff/`
- [x] Given fixtures cover ready workspace data without untyped records extension
- [x] Trigger runs workspace source negative and report replay handoff regression tests
- [x] Verify covers no standalone insight API, no workspace report API call, no prototype helper import, privacy field negative grep and out-of-scope testid negative grep
- [x] Scenario `setup -> trigger -> verify -> cleanup` passes
- [x] Revision 2026-07-12 trigger covers shared start-practice structured-round duration, stale-budget plan replacement and unknown-round fail-closed behavior
- [x] Revision 2026-07-12 scenario `setup -> trigger -> verify -> cleanup` passes（24/24 Vitest PASS）

## E2E.P0.045 Practice structured-round budget display

- [x] Update scenario README/data so selected round duration -> PracticePlan budget -> Top Bar display is the asserted user outcome
- [x] Trigger runs current UI contract, shared start budget tests and PracticeScreen budget tests
- [x] Verify rejects fixed `25:00`, requires plan-derived budget markers, and rejects no-test/failed output
- [x] Scenario `setup -> trigger -> verify -> cleanup` passes（UI contract 45/45、Vitest 65/65、Go voice gate PASS）

## E2E.P0.057 Report retry / next-round handoff boundaries

- [x] Replace the fixed canonical ladder expectation with TargetJob structured-round ordering
- [x] Cover middle -> immediate next, final/single/empty/unknown/loading/failure fail-closed and in-flight duplicate-click guard
- [x] Verify rejects fixed `ROUND_ORDER`, `DEFAULT_NEXT_ROUND` and fallback-to-current/fallback-to-first behavior
- [x] Scenario `setup -> trigger -> verify -> cleanup` passes（34/34 Vitest PASS）

## Closeout

- [x] `validate_context.py --context docs/spec/frontend-workspace-and-practice/plans/001-workspace-and-interview-context/context.yaml --target frontend` passes
- [x] `sync-doc-index --check` passes
- [x] `make docs-check` passes
- [x] `git diff --check` passes
