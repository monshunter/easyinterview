# 004 — Report-derived Practice Plans Test Checklist

> **版本**: 1.8
> **状态**: completed
> **更新日期**: 2026-07-12

**关联 Test Plan**: [test-plan](./test-plan.md)

## Phase 1: Report-derived Contract Tests

- [x] Current report-derived unit and scenario tests exist for create/read/replay.
  <!-- verified: 2026-07-06 method=current-state-search evidence="rg found backend/cmd/api/practice_http_scenario_test.go TestE2EP0070PracticeDerivedPlanCreateReadReplay plus service/store/API Derived/Source tests." -->
- [x] Current report source validation / isolation tests exist.
  <!-- verified: 2026-07-06 method=current-state-search evidence="rg found backend/cmd/api/practice_http_scenario_test.go TestE2EP0072PracticeDerivedSourceValidationIsolationPrivacy and store/practice derived source isolation tests." -->

## Phase 2: Source Boundary Negative Tests

- [x] Out-of-scope source start tests and scenarios are no longer current owner gates.
  <!-- verified: 2026-07-06 method=current-state-search evidence="rg over backend/cmd/api and backend/internal test files found P0.070/P0.072 only for this plan; find test/scenarios/e2e found no p0-071 or p0-073 scenario directories." -->
- [x] Prohibited source tokens are absent from positive runtime/generated/fixture contract.
  <!-- verified: 2026-07-06 method=negative-grep evidence="rg over backend/internal/practice backend/internal/api/practice backend/internal/store/practice openapi/openapi.yaml backend/internal/api/generated frontend/src/api/generated openapi/fixtures for sourceDebriefId/source_debrief_id/PracticeGoalDebrief/goal='debrief'/goal=debrief/debrief-derived/raw_questions/__DEBRIEF returned no matches." -->

- [x] Context validation, docs-check, and whitespace gate pass after reconciliation.
  <!-- verified: 2026-07-06 method=context-and-doc-gates evidence="validate_context.py backend-practice/004 backend PASS; sync-doc-index --check PASS; make docs-check PASS; git diff --check PASS." -->

## Phase 3: Server-owned dimension focus tests

- [x] Generated/API closed derived-request and frozen-field-copy rejection tests pass.<!-- verified: 2026-07-12 method=generated-contract+go-test evidence="direct generated request plus strict handler branch validation; copied target/settings and unknown focus rejected with zero service calls" -->
- [x] Real PostgreSQL generic-empty-focus retry, issue-backed non-empty projection, next empty-focus, isolation and idempotency tests pass.
  <!-- verified: 2026-07-12 method=scenario+postgres-v19 evidence="P0.070 and P0.072 setup/trigger/verify/cleanup PASS on disposable schema v19; exact tagged tests emit REPORT_GENERIC_RETRY_PASS, REPORT_DERIVED_FOCUS_PASS, REPORT_NEXT_EMPTY_FOCUS_PASS, REPORT_DERIVED_ISOLATION_PASS, REPORT_DERIVED_PRIVACY_PASS, REPORT_DERIVED_POSTGRES_PASS; API IK gate emits REPORT_DERIVED_IDEMPOTENCY_PASS." -->
- [x] P0.070/P0.072 named marker set plus F3 `PRACTICE_SEMANTIC_FOCUS_PROMPT_V020_PASS` pass with no raw source evidence; backend exact-version test reads v0.2 before release and final runtime asserts active `ResolveActive` v0.2 after the F3 marker.
  <!-- verified: 2026-07-12 method=scenario+active-registry evidence="P0.070 requires exact active registry test RUN/PASS, v0.2/v0.2 registry.v1 payload with code+label+issues, semantic prompt marker, and verified F3 marker consumption; raw transcript/anchors are absent." -->
- [x] Legacy focus/action/turn identifiers and code-only prompt focus have zero positive final-active consumers; immutable v0.1/000002 rollback/history/negative fixtures are explicitly allowlisted.
  <!-- verified: 2026-07-12 method=exact-negative-scenario-gate evidence="P0.072 emitted REPORT_DERIVED_LEGACY_IDENTIFIER_NEGATIVE_PASS after zero positive final-active hits; explicit historical/test/negative documentation allowlists remain auditable." -->
