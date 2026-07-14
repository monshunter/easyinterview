# 004 — Report-derived Practice Plans Checklist

> **版本**: 1.9
> **状态**: completed
> **更新日期**: 2026-07-14

**关联计划**: [plan](./plan.md)

## Phase 1: Report-derived Plan Contract

- [x] 1.1 `sourceReportId` remains the only current derived-plan source field for `CreatePracticePlanRequest` / `PracticePlan`; verification: OpenAPI and generated Go/TS expose `sourceReportId` for retry / next-round and do not expose `sourceDebriefId`.
  <!-- verified: 2026-07-06 method=current-state-search evidence="rg over openapi/openapi.yaml, backend/internal/api/generated, frontend/src/api/generated, backend/internal/practice, backend/internal/store/practice shows sourceReportId/source_report_id current hits and no sourceDebriefId/source_debrief_id positive hits outside negative scenario text." -->

## Phase 2: Source Boundary Reconciliation

- [x] 2.1 Remove out-of-scope source contract from plan/checklist/test/BDD/context; verification: this plan no longer lists `goal='debrief'`, `sourceDebriefId`, `source_debrief_id`, `PracticeGoalDebrief`, or raw source-question seeding as current implementation items.
  <!-- verified: 2026-07-06 method=doc-rewrite evidence="plan/test/BDD docs rewritten to report-derived only; context keywords use current sourceReportId wording only." -->
- [x] 2.3 Run docs/context gates after reconciliation; verification: `validate_context.py`, `make docs-check`, `git diff --check`, and negative grep pass.

## Phase 3: Server-owned dimension focus

- [x] 3.1 RED-GREEN: generated/API tests enforce closed derived request `{goal,sourceReportId}` vs baseline request shape; no focus/settings/identity fields are accepted on derived requests.
  <!-- verified: 2026-07-12 method=red-green evidence="generated Go/TS expose the conditional request union; TestCreatePracticePlanDerivedRequestIsClosed rejects copied target/settings and unknown focus before service dispatch; full internal/api/practice suite PASS" -->
- [x] 3.2 RED-GREEN: retry atomically projects retryFocusDimensionCodes to plan focus_dimension_codes; empty means generic same-round retry, non-empty is unique and wholly issue-backed; next persists empty focus; read/start replay exact values.
  <!-- verified: 2026-07-12 method=red-green+sqlmock evidence="TestCreateDerivedPracticePlanUsesOnlySourceAuthority, TestSQLRepositoryCreateDerivedPlanProjectsDimensionFocus, TestSQLRepositoryGetPlanReplaysExactDimensionFocus and start reservation current-column assertion PASS; full internal/practice + store/practice packages PASS" -->
- [x] 3.3 RED-GREEN: retry start/send with non-empty focus joins source report code→label+issues into JSON-encoded `semanticFocus/{{semantic_focus_json}}`; empty focus injects no fabricated guidance; code-only, unsupported/missing cross-ref, raw transcript/anchor and report focus on next/baseline fail negative gates. Runtime tests may use the F3-owned immutable v0.2 exact coordinate before release, but final runtime must resolve the active v0.2 pair while v0.1 remains exactly retrievable for rollback; completion requires `REPORT_DERIVED_SEMANTIC_PROMPT_PASS` plus F3 `PRACTICE_SEMANTIC_FOCUS_PROMPT_V020_PASS`. (`go test ./backend/internal/practice ./backend/internal/store/practice -run 'DimensionFocus|Derived|SemanticFocus' -count=1`)
- [x] 3.4 RED-GREEN: missing/cross-user/non-ready/missing-context/unsupported-or-duplicate-non-empty-focus and IK mismatch fail without insert/leak; empty focus stays valid, and server derives settings/identity/duration rather than validating client copies.
- [x] 3.6 REGRESSION-GATE: final active runtime/generated/OpenAPI/fixtures/scenarios contain zero positive `focusCompetencyCodes|focus_competency_codes|retryFocusCompetencyCodes|retry_focus_competency_codes|retryFocusTurnIds|retry_focus_turn_ids|retry_round` or code-only prompt focus; immutable v0.1/000002 rollback, history and explicit negative fixtures are allowlisted but never active consumers after F3 activation. Emit owner-only `REPORT_DERIVED_LEGACY_IDENTIFIER_NEGATIVE_PASS` after the exact-set gate passes.

## BDD Gate

- [x] BDD-Gate: `BDD.PRACTICE.DERIVED.001` 由 [BDD checklist](./bdd-checklist.md) 关联报告派生 plan 的 owner behavior tests；不创建或声明真实 E2E PASS。
