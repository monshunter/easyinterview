# 004 — Report-derived Practice Plans Test Checklist

> **版本**: 1.1
> **状态**: completed
> **更新日期**: 2026-07-06

**关联 Test Plan**: [test-plan](./test-plan.md)

## Phase 1: Report-derived Contract Tests

- [x] Current report-derived unit and scenario tests exist for create/read/replay.
  <!-- verified: 2026-07-06 method=current-state-search evidence="rg found backend/cmd/api/practice_http_scenario_test.go TestE2EP0070PracticeDerivedPlanCreateReadReplay plus service/store/API Derived/Source tests." -->
- [x] Current report source validation / isolation tests exist.
  <!-- verified: 2026-07-06 method=current-state-search evidence="rg found backend/cmd/api/practice_http_scenario_test.go TestE2EP0072PracticeDerivedSourceValidationIsolationPrivacy and store/practice derived source isolation tests." -->

## Phase 2: Retired Debrief Negative Tests

- [x] Historical debrief start tests and scenarios are no longer current owner gates.
  <!-- verified: 2026-07-06 method=current-state-search evidence="rg over backend/cmd/api and backend/internal test files found P0.070/P0.072 only for this plan; find test/scenarios/e2e found no p0-071 or p0-073 scenario directories." -->
- [x] Retired debrief source tokens are absent from positive runtime/generated/fixture contract.
  <!-- verified: 2026-07-06 method=retired-token-negative-grep evidence="rg over backend/internal/practice backend/internal/api/practice backend/internal/store/practice openapi/openapi.yaml backend/internal/api/generated frontend/src/api/generated openapi/fixtures for sourceDebriefId/source_debrief_id/PracticeGoalDebrief/goal='debrief'/goal=debrief/debrief-derived/raw_questions/__DEBRIEF returned no matches." -->

## Phase 3: Docs Gates

- [x] Context validation, docs-check, and whitespace gate pass after reconciliation.
  <!-- verified: 2026-07-06 method=context-and-doc-gates evidence="validate_context.py backend-practice/004 backend PASS; sync-doc-index --check PASS; make docs-check PASS; git diff --check PASS." -->
