# 001 — Plan and Session Orchestration Checklist

> **版本**: 1.9
> **状态**: completed
> **更新日期**: 2026-07-10

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

## Phase 7: first-question invalid-output test table consolidation

- [x] 7.1 Record scoped `dupl` RED and verify both top-level test names have no external consumers.
  <!-- red: 2026-07-10 method=practice-first-question-invalid-test-dupl evidence="Scoped dupl -t 100 reports missing-text and non-JSON first-question tests as session_starter_test.go's only clone group; repo search finds no external exact-name consumers." -->
- [x] 7.2 Replace both tests with one table-driven test while retaining both AI outputs, named cases, step-order and terminal failure assertions.
  <!-- verified: 2026-07-10 method=practice-first-question-invalid-table evidence="One TestStartPracticeSessionRejectsInvalidFirstQuestionOutput runs named missing-text and non-JSON cases with isolated stores/services, exact reserve-ai-fail ordering and terminal non-retryable AI_OUTPUT_INVALID assertions. Both pass; scoped dupl is zero and staticcheck passes." -->
- [x] 7.3 Run focused Practice tests, scoped dupl, full backend/vet/staticcheck and owner/product/docs/pruning gates; then restore the owner group to `completed`.
  <!-- verified: 2026-07-10 method=practice-first-question-invalid-table-consolidation evidence="Both named missing-text/non-JSON cases pass with reserve-ai-fail ordering and terminal non-retryable AI_OUTPUT_INVALID assertions; session_starter_test.go scoped dupl is zero. Practice owner packages, full backend, go vet/staticcheck, both contexts and final docs/index/link/diff/pruning gates pass with real_residuals=0. No behavior/BDD change, Bug/retrospective report or environment operation was needed." -->

## Phase 8: create-plan SQL expectation consolidation

- [x] 8.1 Record scoped `dupl` RED for the repeated successful create-plan insert row setup and confirm all owning top-level tests remain required.
  <!-- verified: 2026-07-10 method=practice-create-plan-sql-test-dupl evidence="Scoped dupl -t 100 reports the baseline and report-derived success insert expectations as internal/store/practice's only clone group. Existing exact test names remain owner/regression evidence and will not be removed." -->
- [x] 8.2 Extract one test-only successful insert expectation helper and reuse it from baseline, empty-focus and report-derived tests without changing query or row contracts.
  <!-- verified: 2026-07-10 method=practice-create-plan-success-expectation-helper evidence="The three existing top-level tests pass through one helper with their original query patterns, focus matcher and source-report row. Failure expectations remain explicit; scoped internal/store/practice dupl reports zero clone groups and staticcheck passes." -->
- [x] 8.3 Run focused 001/004 tests, scoped dupl, full backend/vet/staticcheck and owner/product/docs/pruning closeout gates.
  <!-- verified: 2026-07-10 method=practice-create-plan-sql-test-consolidation-closeout evidence="001 create-plan, 004 Derived/Source/P0.070/P0.072, store and Practice package gates PASS; full backend, go vet and staticcheck PASS; 001/004/product contexts, docs/index/diff and pruning gates PASS with real_residuals=0." -->

## Phase 9: plan/session GET fixture harness consolidation

- [x] 9.1 RED: scoped `dupl` reports the plan/session GET test pair; exact top-level fixture parity names have no external consumers but both operation behaviors remain required.
  <!-- red: 2026-07-10 method=practice-get-resource-test-dupl evidence="dupl -t 100 reports get_practice_plan_test.go and get_practice_session_test.go as one 49-line clone group covering scoped 404 and fixture parity setup; repo search finds no exact-name consumer." -->
- [x] 9.2 Introduce a test-only generic fixture runner with typed body decoding and a scoped-not-found assertion; delete per-operation fixture structs/loaders and require scoped `dupl` zero.
  <!-- verified: 2026-07-10 method=practice-get-resource-fixture-harness-green evidence="One TestGetPracticeResourcesFixtureParity preserves plan/session and all six named fixture cases through a generic runner with api.PracticePlan/api.PracticeSession decoding. Old top-level names and loaders have zero references; the three-file scoped dupl reports zero groups and focused GET tests pass." -->
- [x] 9.3 BDD-Gate: run plan/session focused tests, P0.022/P0.023 serial lifecycles, Practice/full backend/vet/staticcheck, contexts and docs/pruning gates.
  <!-- verified: 2026-07-10 method=practice-get-resource-fixture-harness-closeout evidence="Focused GET and full Practice/cmd-api packages pass; full backend, vet and staticcheck pass. P0.022/P0.023 complete serial lifecycles and cleanup; owner/product contexts and docs/index/diff/pruning gates pass with real_residuals=0." -->
