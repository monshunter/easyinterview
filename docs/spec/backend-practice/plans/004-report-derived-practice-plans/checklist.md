# 004 — Report-derived Practice Plans Checklist

> **版本**: 1.2
> **状态**: completed
> **更新日期**: 2026-07-06

**关联计划**: [plan](./plan.md)

## Phase 1: Report-derived Plan Contract

- [x] 1.1 `sourceReportId` remains the only current derived-plan source field for `CreatePracticePlanRequest` / `PracticePlan`; verification: OpenAPI and generated Go/TS expose `sourceReportId` for retry / next-round and do not expose `sourceDebriefId`.
  <!-- verified: 2026-07-06 method=current-state-search evidence="rg over openapi/openapi.yaml, backend/internal/api/generated, frontend/src/api/generated, backend/internal/practice, backend/internal/store/practice shows sourceReportId/source_report_id current hits and no sourceDebriefId/source_debrief_id positive hits outside negative scenario text." -->
- [x] 1.2 Report-derived create/read/idempotency remains covered by executable tests and scenario docs; verification: `TestE2EP0070PracticeDerivedPlanCreateReadReplay` exists and `test/scenarios/e2e/p0-070-practice-derived-plan-create-read-replay` describes retry / next-round `sourceReportId`.
  <!-- verified: 2026-07-06 method=current-state-search evidence="rg found backend/cmd/api/practice_http_scenario_test.go TestE2EP0070PracticeDerivedPlanCreateReadReplay and P0.070 README states retry_current_round/next_round sourceReportId create/read/replay." -->
- [x] 1.3 Report source validation / isolation remains covered by executable tests and scenario docs; verification: `TestE2EP0072PracticeDerivedSourceValidationIsolationPrivacy` exists and `test/scenarios/e2e/p0-072-practice-derived-source-isolation-privacy` covers missing/cross-user/wrong-target source.
  <!-- verified: 2026-07-06 method=current-state-search evidence="rg found backend/cmd/api/practice_http_scenario_test.go TestE2EP0072PracticeDerivedSourceValidationIsolationPrivacy and P0.072 README describes missing/cross-user/wrong-target source validation." -->

## Phase 2: Source Boundary Reconciliation

- [x] 2.1 Remove out-of-scope source contract from plan/checklist/test/BDD/context; verification: this plan no longer lists `goal='debrief'`, `sourceDebriefId`, `source_debrief_id`, `PracticeGoalDebrief`, or raw source-question seeding as current implementation items.
  <!-- verified: 2026-07-06 method=doc-rewrite evidence="plan/test/BDD docs rewritten to report-derived only; context keywords use current sourceReportId wording only." -->
- [x] 2.2 Keep only current scenario gates for this owner; verification: active BDD docs retain only `E2E.P0.070` and `E2E.P0.072`, and no P0.071/P0.073 scenario directories exist under `test/scenarios/e2e`.
  <!-- verified: 2026-07-06 method=current-state-search evidence="find test/scenarios/e2e -maxdepth 1 -type d -name 'p0-07*' shows P0.070/P0.072 plus resume P0.074-079, no P0.071/P0.073." -->
- [x] 2.3 Run docs/context gates after reconciliation; verification: `validate_context.py`, `make docs-check`, `git diff --check`, and negative grep pass.
  <!-- verified: 2026-07-06 method=focused-go-doc-gates evidence="cd backend && go test ./internal/practice ./internal/store/practice ./internal/api/practice ./cmd/api -run 'Derived|Source|TestE2EP0070|TestE2EP0072' -count=1 PASS; runtime/generated/fixture negative grep for sourceDebriefId/source_debrief_id/PracticeGoalDebrief/goal='debrief'/debrief-derived/raw_questions returned no matches; validate_context.py backend-practice/004 backend PASS; sync-doc-index --check PASS; git diff --check PASS." -->
