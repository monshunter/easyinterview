# 004 — Derived Plans and Debrief Seeding Test Plan

> **版本**: 1.0
> **状态**: completed
> **更新日期**: 2026-05-16

**关联计划**: [plan](./plan.md) / [checklist](./checklist.md)
**关联 Test Checklist**: [test-checklist](./test-checklist.md)

## 1 测试矩阵

| 行 | 风险 | 测试文件 / 入口 | 预期证据 |
|----|------|----------------|----------|
| T1 | OpenAPI source fields drift | `backend/internal/api/practice/*_test.go` + `make lint-openapi` + `make validate-fixtures` | request/response/generated/fixtures 均包含 source ids |
| T2 | migration CHECK 与 service 规则不一致 | `backend/internal/migrations/*source*_test.go` + `migrations/lint.sh` + `make migrate-check` | SQL 接受合法组合、拒绝非法组合 |
| T3 | non-baseline goal 仍被拒绝 | `backend/internal/practice/create_plan_test.go` | valid retry/next_round/debrief source 通过；缺失/互斥 source 422 |
| T4 | source ownership / target mismatch | `backend/internal/store/practice/create_plan_test.go` | cross-user / wrong target / draft / empty debrief source 失败 |
| T5 | debrief start 误调 first_question AI | `backend/internal/practice/session_starter_test.go` | fake registry / AIClient call count = 0 for debrief |
| T6 | debrief first turn 来源错误 | `backend/internal/store/practice/create_plan_test.go` / API tests | currentTurn.questionText 来自 `debriefs.raw_questions[0].questionText` |
| T7 | idempotency replay / mismatch | API tests + `backend/cmd/api/practice_http_scenario_test.go` | same key replay 原响应；same key different source conflict |
| T8 | privacy / legacy negative | HTTP scenario + scoped grep | no raw text in audit/outbox/log/metric/idempotency; no active `mode='debrief'` |

## 2 Red / Green 命令

### Phase 0

- Red: `rg -n "sourceDebriefId|source_debrief_id" openapi backend/internal/api/generated frontend/src/api/generated migrations backend/internal/practice backend/internal/store/practice` 现状缺口可见。
- Green:
  - `make codegen-openapi`
  - `make lint-openapi`
  - `make validate-fixtures`
  - `cd backend && go test ./internal/migrations -run 'PracticePlan.*Source|source_debrief|SourceCheck' -count=1`
  - `./migrations/lint.sh`
  - `set -a; . deploy/dev-stack/.env; set +a; make migrate-check`

### Phase 1

- Red: `cd backend && go test ./internal/practice -run 'TestServiceCreatePracticePlan.*Derived|TestServiceCreatePracticePlan.*Source' -count=1`
- Green:
  - `cd backend && go test ./internal/practice -run 'TestServiceCreatePracticePlan.*Derived|TestServiceCreatePracticePlan.*Source' -count=1`
  - `cd backend && go test ./internal/store/practice -run 'TestSQLRepositoryCreatePlan.*Derived|TestSQLRepositoryCreatePlan.*Source' -count=1`
  - `cd backend && go test ./internal/api/practice -run 'TestCreatePracticePlan.*Derived|TestCreatePracticePlan.*Idempotency' -count=1`

### Phase 2

- Red: `cd backend && go test ./internal/practice -run 'TestStartPracticeSession.*Debrief' -count=1`
- Green:
  - `cd backend && go test ./internal/practice ./internal/store/practice -run 'TestStartPracticeSession.*Debrief|TestSQLRepositoryReserveSessionStart.*Debrief' -count=1`
  - `cd backend && go test ./internal/api/practice -run 'TestStartPracticeSession.*Debrief' -count=1`

### Phase 3 / 4

- `cd backend && go test ./cmd/api -run 'TestE2EP0070|TestE2EP0071|TestE2EP0072|TestE2EP0073' -count=1`
- `cd backend && go test ./internal/practice ./internal/store/practice ./internal/api/practice ./cmd/api -run 'Derived|Debrief|Source' -count=1`
- `make docs-check`
- `git diff --check`

## 3 Legacy-Negative Gate

Active runtime / generated / fixture scope must not contain `mode='debrief'`, `PracticeModeDebrief`, `legacy debrief replay value`, `startDebriefInterview`, or `source_debrief` references outside the intended new source id fields:

```bash
rg -n "PracticeModeDebrief|mode='debrief'|mode=debrief|legacy debrief replay value|startDebriefInterview" \
  backend/internal/practice backend/internal/api/practice backend/internal/store/practice \
  openapi frontend/src/api/generated backend/internal/api/generated openapi/fixtures/PracticePlans openapi/fixtures/PracticeSessions
```

Expected: no active hits. Negative test/docs prose may enumerate forbidden literals only inside plan/test gate definitions.
