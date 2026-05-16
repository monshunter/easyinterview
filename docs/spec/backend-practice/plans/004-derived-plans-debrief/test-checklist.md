# 004 — Derived Plans and Debrief Seeding Test Checklist

> **版本**: 1.0
> **状态**: completed
> **更新日期**: 2026-05-16

**关联 Test Plan**: [test-plan](./test-plan.md)

## Phase 0: Contract / Migration Tests

- [x] Phase 0 OpenAPI/generated/fixture tests pass
  <!-- verified: 2026-05-16 `make codegen-openapi`, `make lint-openapi`, `make validate-fixtures`, `cd backend && go test ./internal/api/practice -run 'TestCreatePracticePlanMapsDerivedSourceIds|TestCreatePracticePlanFixtureParityDefault' -count=1` -->
- [x] Phase 0 migration SQL contract, migration lint, and migrate-check pass
  <!-- verified: 2026-05-16 `cd backend && go test ./internal/migrations -run TestBaselinePracticePlansDerivedSourceContract -count=1`, `python3 -m pytest scripts/lint/migrations_lint_test.py -q`, `./migrations/lint.sh`, `set -a; . deploy/dev-stack/.env; set +a; make migrate-check` -->
- [x] Phase 0 docs/index sync passes
  <!-- verified: 2026-05-16 `make docs-check`, `python3 .agent-skills/implement/shared/scripts/validate_context.py --context docs/spec/backend-practice/plans/004-derived-plans-debrief/context.yaml --docs-root docs --target backend`, scoped stale `mode='debrief'` grep no matches -->

## Phase 1: Derived Plan Creation Tests

- [x] Phase 1 service source-rule tests pass
  <!-- verified: 2026-05-16 `cd backend && go test ./internal/practice -run 'TestServiceCreatePracticePlan.*Derived|TestServiceCreatePracticePlan.*Source|TestServiceCreatePracticePlanRejectsInvalidGoal' -count=1` -->
- [x] Phase 1 repository source ownership / target / status tests pass
  <!-- verified: 2026-05-16 `cd backend && go test ./internal/store/practice -run 'TestSQLRepositoryCreatePlan.*Derived|TestSQLRepositoryCreatePlan.*Source|TestSQLRepositoryCreatePlanWritesPlanAndAudit|TestSQLRepositoryCreatePlanReturnsPrerequisite|TestSQLRepositoryGetPlanScopesByUser' -count=1` -->
- [x] Phase 1 handler/idempotency source tests pass
  <!-- verified: 2026-05-16 `cd backend && go test ./internal/api/practice -run 'TestCreatePracticePlan.*Derived|TestCreatePracticePlan.*Idempotency|TestCreatePracticePlanReturns422ForValidationError|TestCreatePracticePlanFixtureParityDefault' -count=1` -->

## Phase 2: Debrief Session Start Tests

- [x] Phase 2 service no-AI debrief start tests pass
  <!-- verified: 2026-05-16 `cd backend && go test ./internal/practice -run 'TestStartPracticeSession.*Debrief' -count=1` -->
- [x] Phase 2 repository first-turn seeding tests pass
  <!-- verified: 2026-05-16 `cd backend && go test ./internal/store/practice -run 'TestSQLRepositoryReserveSessionStart.*Debrief' -count=1` -->
- [x] Phase 2 handler/idempotency replay tests pass
  <!-- verified: 2026-05-16 `cd backend && go test ./internal/api/practice ./internal/store/practice -run 'TestStartPracticeSession.*Debrief|TestSQLRepositoryReserveSessionStartReplaysStoredResponseBody' -count=1` -->

## Phase 3: Isolation / Privacy / Legacy Tests

- [x] Phase 3 source isolation matrix tests pass
  <!-- verified: 2026-05-16 `cd backend && go test ./internal/practice ./internal/store/practice -run 'TestServiceCreatePracticePlan.*Unavailable|TestSQLRepositoryCreatePlanReturnsPrerequisiteErrorWhenDerivedSourceUnavailable' -count=1` -->
- [x] Phase 3 privacy redline tests pass
  <!-- verified: 2026-05-16 `cd backend && go test ./internal/store/practice -run TestSQLRepositoryCommitSessionStartWritesAuditMetadataWithoutQuestionText -count=1`; raw debrief marker grep no matches outside tests -->
- [x] Phase 3 legacy-negative grep passes
  <!-- verified: 2026-05-16 scoped runtime/generated/fixture grep for legacy debrief mode literals returned no matches outside negative tests -->

## Phase 4: Scenario And Handoff Tests

- [x] `E2E.P0.070-073` focused HTTP scenario tests pass
  <!-- verified: 2026-05-16 `cd backend && go test ./cmd/api -run 'TestE2EP0070|TestE2EP0071|TestE2EP0072|TestE2EP0073' -count=1` -->
- [x] backend-debrief 0.6 focused verification passes after this plan lands
  <!-- verified: 2026-05-16 backend-debrief/001 checklist 0.6 and test-checklist 0.F updated with backend-practice/004 evidence; `make docs-check` passed -->
