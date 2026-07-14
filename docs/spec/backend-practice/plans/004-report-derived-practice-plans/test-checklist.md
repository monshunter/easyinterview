# 004 — Report-derived Practice Plans Test Checklist

> **版本**: 1.8
> **状态**: completed
> **更新日期**: 2026-07-12

**关联 Test Plan**: [test-plan](./test-plan.md)

## Phase 1: Report-derived Contract Tests

- [x] Current report-derived service/store/API code tests cover create/read/replay；focused runs only provide development feedback.
- [x] Current report source validation / isolation tests exist.
- [x] Phase completion is reported by repository-root `make test`；database integration and negative searches remain separate gates.

## Phase 2: Source Boundary Negative Tests

- [x] Out-of-scope source start tests and scenarios are no longer current owner gates.
- [x] Prohibited source tokens are absent from positive runtime/generated/fixture contract.
  <!-- verified: 2026-07-06 method=negative-grep evidence="rg over backend/internal/practice backend/internal/api/practice backend/internal/store/practice openapi/openapi.yaml backend/internal/api/generated frontend/src/api/generated openapi/fixtures for sourceDebriefId/source_debrief_id/PracticeGoalDebrief/goal='debrief'/goal=debrief/debrief-derived/raw_questions/__DEBRIEF returned no matches." -->

- [x] Context validation, docs-check, and whitespace gate pass after reconciliation.
  <!-- verified: 2026-07-06 method=context-and-doc-gates evidence="validate_context.py backend-practice/004 backend PASS; sync-doc-index --check PASS; make docs-check PASS; git diff --check PASS." -->

## Phase 3: Server-owned dimension focus tests

- [x] Generated/API closed derived-request and frozen-field-copy rejection tests pass.<!-- verified: 2026-07-12 method=generated-contract+go-test evidence="direct generated request plus strict handler branch validation; copied target/settings and unknown focus rejected with zero service calls" -->
- [x] Real PostgreSQL generic-empty-focus retry, issue-backed non-empty projection, next empty-focus, isolation and idempotency tests pass.
- [x] Legacy focus/action/turn identifiers and code-only prompt focus have zero positive final-active consumers; immutable v0.1/000002 rollback/history/negative fixtures are explicitly allowlisted.
