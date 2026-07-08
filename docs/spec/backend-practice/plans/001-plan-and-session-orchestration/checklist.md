# 001 — Plan and Session Orchestration Checklist

> **版本**: 1.6
> **状态**: completed
> **更新日期**: 2026-07-08

**关联计划**: [plan](./plan.md)

## Phase 1: Contract Foundation

- [x] 1.1 Shared conventions / OpenAPI / events / migrations aligned with current PracticeMode, PracticeGoal, practice error codes, idempotency table and practice schema.
  <!-- verified: 2026-05-09 method=contract-drift-gates evidence="shared/openapi/events/migrations gates passed during 001 delivery." -->
- [x] 1.2 F3 baseline practice feature_key preflight passed for `practice.session.first_question`, `practice.session.follow_up`, `practice.turn.lightweight_observe`.
  <!-- verified: 2026-05-09 method=preflight evidence="prompt-rubric 001 completed and ResolveActive tests passed." -->

## Phase 2: Plan And Session Success Path

- [x] 2.1 `createPracticePlan` / `getPracticePlan` real handler, service and store implemented with user isolation, audit metadata and idempotency.
  <!-- verified: 2026-07-07 method=focused-go evidence="cd backend && go test ./internal/practice ./internal/api/practice ./internal/store/practice -run 'Test.*CreatePracticePlan|TestSQLRepositoryCreatePlan|TestStartPracticeSessionRunsThreeStepFlowWithAIOutsideTransactions' -count=1 PASS." -->
- [x] 2.2 `startPracticeSession` three-step reservation / AI / commit flow implemented with first turn and `practice.session.started` outbox.
  <!-- verified: 2026-07-07 method=focused-go evidence="cd backend && go test ./internal/practice ./internal/api/practice ./internal/store/practice -count=1 PASS." -->
- [x] 2.3 `getPracticeSession` user-scoped read implemented.
  <!-- verified: 2026-07-07 method=focused-go evidence="cd backend && go test ./internal/api/practice -count=1 PASS." -->

## Phase 3: Failure, Idempotency, Isolation

- [x] 3.1 AI error mapping, failed reservation retry, success replay, mismatch conflict, cross-user isolation, TTL reset and single active session guard implemented.
  <!-- verified: 2026-07-07 method=focused-go evidence="cd backend && go test ./internal/practice ./internal/api/practice ./internal/store/practice -count=1 PASS." -->
- [x] 3.2 BDD-Gate: `E2E.P0.022`, `E2E.P0.023`, `E2E.P0.024`, `E2E.P0.025`, `E2E.P0.026` assets and execution checklist completed.
  <!-- verified: 2026-05-09 method=scenario bddChecklist=complete -->

## Phase 4: Privacy And Observability

- [x] 4.1 AI task metadata, audit metadata, outbox payload, metric label and redaction gates passed.
  <!-- verified: 2026-05-09 method=focused-go-and-scenario evidence="observability, audit, redaction and P0.026 gates passed during 001 delivery." -->

## Phase 5: Flat Resume Binding

- [x] 5.1 `createPracticePlan` handler + service + store use `resumeId` and `practice_plans.resume_id`.
  <!-- verified: 2026-07-07 method=focused-go-and-negative-grep evidence="practice CreatePlan focused tests PASS; flat-resume residual grep over backend practice packages returned no matches." -->
- [x] 5.2 First-question prompt context includes flat `resumes.structured_profile` through `{{resume_profile}}`.
  <!-- verified: 2026-07-07 method=red-green-focused-go evidence="Red: go test ./internal/practice -run TestStartPracticeSessionRunsThreeStepFlowWithAIOutsideTransactions -count=1 failed because SessionReservation lacked ResumeProfile. Green: go test ./internal/practice ./internal/store/practice -run 'TestStartPracticeSessionRunsThreeStepFlowWithAIOutsideTransactions|TestSQLRepositoryReserveSessionStartReusesFailedRetryableRecord' -count=1 PASS; cd backend && go test ./internal/practice ./internal/api/practice ./internal/store/practice -count=1 PASS." -->
- [x] 5.3 Phase 5 closeout gate passed.
  <!-- verified: 2026-07-07 method=focused-go-and-negative-grep evidence="cd backend && go test ./internal/practice/... ./cmd/api -count=1 PASS; flat-resume residual grep over backend practice packages returned no matches." -->

## Phase 6: PracticePlan resumeId response remediation

- [x] 6.1 `PracticePlan` OpenAPI/generated/fixtures include persisted `resumeId` on create/read responses（验证：`make codegen-openapi`; `make validate-fixtures` PASS）
- [x] 6.2 Practice handler/service/store return `practice_plans.resume_id` for create/read and preserve user isolation（验证：`cd backend && go test ./internal/practice ./internal/api/practice ./internal/store/practice -count=1` PASS）
