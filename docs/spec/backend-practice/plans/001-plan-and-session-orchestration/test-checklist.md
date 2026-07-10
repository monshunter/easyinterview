# 001 — Plan and Session Orchestration Test Checklist

> **版本**: 1.4
> **状态**: completed
> **更新日期**: 2026-07-07

**关联 Test Plan**: [test-plan](./test-plan.md)

## Phase 1: Contract Foundation

- [x] Shared / OpenAPI / events / migrations drift gates for plan 001 passed.
- [x] F3 practice feature_key preflight passed.

## Phase 2: Plan And Session Success Path

- [x] `createPracticePlan`, `getPracticePlan`, `startPracticeSession`, `getPracticeSession` focused tests passed.
- [x] `practice.session.started` outbox and first-turn response tests passed.

## Phase 3: Failure, Idempotency, Isolation

- [x] AI error mapping, failed reservation retry, idempotency replay/mismatch, TTL reset and cross-user isolation tests passed.

## Phase 4: Privacy And Observability

- [x] AI task metadata, audit metadata, redaction and payload privacy tests passed.

## Phase 5: Flat Resume Binding

- [x] `resumeId` / `practice_plans.resume_id` tests passed.
- [x] First-question prompt includes flat `resumes.structured_profile` context.
- [x] Practice package flat-resume residual grep returned no matches.

## Phase 8: Create-plan SQL expectation consolidation

- [x] Baseline, empty-focus and report-derived create-plan store tests pass through one shared successful insert expectation.
- [x] Scoped store `dupl` reports zero clone groups; plan 004 derived-source gates are included in closeout.

## Phase 9: Plan/session GET fixture harness consolidation

- [x] Plan/session fixture parity and scoped 404 tests pass through shared test helpers with typed fixture body decoding.
  <!-- verified: 2026-07-10 method=practice-get-resource-focused evidence="Focused TestGetPractice suite passes plan/session user-scoped success, cross-user 404 and all fixture parity subtests." -->
- [x] Scoped `dupl` over the plan/session GET tests and shared helper reports zero clone groups; full Practice and backend static gates pass.
  <!-- verified: 2026-07-10 method=practice-get-resource-static-closeout evidence="The three GET test files report zero clone groups; full Practice/backend Go tests, vet and staticcheck pass." -->
