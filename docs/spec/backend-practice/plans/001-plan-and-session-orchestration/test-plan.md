# 001 — Plan and Session Orchestration Test Plan

> **版本**: 1.3
> **状态**: completed
> **更新日期**: 2026-07-07

**关联计划**: [plan](./plan.md) / [checklist](./checklist.md)

## 1 测试策略

本计划测试覆盖 backend-practice baseline plan/session foundation。测试层级包括 Go unit/integration、OpenAPI/shared/events/migration drift、privacy grep、BDD scenario scripts 和 docs/index gate；不使用代码覆盖率百分比作为完成条件。

## 2 Coverage Matrix

| 测试源 | 覆盖合同 | 验证入口 |
|--------|----------|----------|
| shared/openapi/events/migration gates | Practice enums, errors, idempotency table, practice schema, generated artifacts | `make codegen-check`, `make validate-fixtures`, shared/events/migration focused gates |
| create/get plan tests | `createPracticePlan`, `getPracticePlan`, user isolation, audit metadata | `cd backend && go test ./internal/api/practice ./internal/practice ./internal/store/practice -run 'Test.*CreatePracticePlan|TestSQLRepositoryCreatePlan' -count=1` |
| start/get session tests | reservation, AI outside transaction, first turn, outbox, read model | `cd backend && go test ./internal/practice ./internal/api/practice ./internal/store/practice -count=1` |
| idempotency tests | success replay, mismatch, TTL reset, cross-user isolation, single active session | practice service/store/API focused tests |
| privacy / observability tests | AI task metadata, audit metadata, event payload redaction | practice focused tests + `E2E.P0.026` |
| flat resume prompt test | `resumeId`, `practice_plans.resume_id`, `resumes.structured_profile` prompt context | `TestStartPracticeSessionRunsThreeStepFlowWithAIOutsideTransactions`, `TestSQLRepositoryReserveSessionStartReusesFailedRetryableRecord` |
| BDD scenarios | user-visible API behavior and state transitions | `E2E.P0.022`-`E2E.P0.026` |

## 3 Focused Commands

```bash
cd backend && go test ./internal/practice ./internal/api/practice ./internal/store/practice -count=1
cd backend && go test ./internal/practice/... ./cmd/api -count=1
rg -n "<flat-resume residual tokens>" backend/internal/practice backend/internal/api/practice backend/internal/store/practice
```

## 4 Completion Evidence

- `2026-07-07`: flat Resume prompt context red/green test passed.
- `2026-07-07`: `go test ./internal/practice/... ./cmd/api -count=1` passed.
- `2026-07-07`: practice package flat-resume residual grep returned no matches.
