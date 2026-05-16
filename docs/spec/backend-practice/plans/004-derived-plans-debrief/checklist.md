# 004 — Derived Plans and Debrief Seeding Checklist

> **版本**: 1.0
> **状态**: completed
> **更新日期**: 2026-05-16

**关联计划**: [plan](./plan.md)

## Phase 0: Contract, Migration, And Plan Preflight

- [x] 0.1 B2 source fields：`openapi/openapi.yaml` 的 `CreatePracticePlanRequest` / `PracticePlan` 增加 `sourceReportId` / `sourceDebriefId`；fixtures 增加 derived scenarios；运行 `make codegen-openapi` / `make lint-openapi` / `make validate-fixtures`；验证：generated Go/TS 字段存在 + fixture parity tests
  <!-- verified: 2026-05-16 red=`cd backend && go test ./internal/api/practice -run TestCreatePracticePlanMapsDerivedSourceIds -count=1` failed on missing generated/domain fields; green=`make codegen-openapi`, `make lint-openapi`, `make validate-fixtures`, `cd backend && go test ./internal/api/practice -run 'TestCreatePracticePlanMapsDerivedSourceIds|TestCreatePracticePlanFixtureParityDefault' -count=1` -->
- [x] 0.2 B4 source_debrief_id：`practice_plans` baseline migration 增加 `source_debrief_id`、FK 与 goal/source CHECK；验证：`cd backend && go test ./internal/migrations -run 'TestBaselineMigration.*PracticePlanSource|TestBaselinePracticePlansSourceCheck' -count=1`、`./migrations/lint.sh`、`set -a; . deploy/dev-stack/.env; set +a; make migrate-check`
  <!-- verified: 2026-05-16 red=`cd backend && go test ./internal/migrations -run TestBaselinePracticePlansDerivedSourceContract -count=1` failed on missing source_debrief_id; green=`cd backend && go test ./internal/migrations -run TestBaselinePracticePlansDerivedSourceContract -count=1`, `python3 -m pytest scripts/lint/migrations_lint_test.py -q`, `./migrations/lint.sh`, `set -a; . deploy/dev-stack/.env; set +a; make migrate-check` -->
- [x] 0.3 owner docs sync：backend-practice spec §7 / history / plans INDEX 指向本 plan；修复 active docs 中 stale `mode='debrief'` 口径；验证：`make docs-check` + scoped grep
  <!-- verified: 2026-05-16 `make docs-check`, `python3 .agent-skills/implement/shared/scripts/validate_context.py --context docs/spec/backend-practice/plans/004-derived-plans-debrief/context.yaml --docs-root docs --target backend`, `rg -n "startPracticeSession\\(mode='debrief'\\)|mode='debrief' session start|goal='debrief' \\+ mode='debrief'" docs/spec/frontend-debrief docs/spec/backend-debrief docs/spec/backend-practice -g '*.md'` no matches -->

## Phase 1: createPracticePlan Derived Source Validation

- [x] 1.1 Red: service 表驱动测试证明当前非 baseline goal 仍返回 owner=004，且 source id 规则缺失；验证：`cd backend && go test ./internal/practice -run 'TestServiceCreatePracticePlan.*Derived|TestServiceCreatePracticePlan.*Source' -count=1` 先失败
  <!-- verified: 2026-05-16 red=`cd backend && go test ./internal/practice -run 'TestServiceCreatePracticePlan.*Derived|TestServiceCreatePracticePlan.*Source' -count=1` failed: valid retry/next_round/debrief still returned owner=004 and baseline source ids were accepted -->
- [x] 1.2 Green: domain DTO / service validation 支持 baseline/report/debrief source 互斥规则；验证：同 1.1 命令通过
  <!-- verified: 2026-05-16 `cd backend && go test ./internal/practice -run 'TestServiceCreatePracticePlan.*Derived|TestServiceCreatePracticePlan.*Source|TestServiceCreatePracticePlanRejectsInvalidGoal' -count=1` -->
- [x] 1.3 Green: SQL repository 插入并验证 report/debrief source 同 user、same target job、ready/completed、raw_questions 非空；验证：`cd backend && go test ./internal/store/practice -run 'TestSQLRepositoryCreatePlan.*Derived|TestSQLRepositoryCreatePlan.*Source' -count=1`
  <!-- verified: 2026-05-16 red=`cd backend && go test ./internal/store/practice -run 'TestSQLRepositoryCreatePlan.*Derived|TestSQLRepositoryCreatePlan.*Source' -count=1` failed on old SQL; green=`cd backend && go test ./internal/store/practice -run 'TestSQLRepositoryCreatePlan.*Derived|TestSQLRepositoryCreatePlan.*Source|TestSQLRepositoryCreatePlanWritesPlanAndAudit|TestSQLRepositoryCreatePlanReturnsPrerequisite|TestSQLRepositoryGetPlanScopesByUser' -count=1` -->
- [x] 1.4 Handler/idempotency：handler 映射 source ids，fingerprint/replay 包含 source ids，response body 返回 source ids；验证：`cd backend && go test ./internal/api/practice -run 'TestCreatePracticePlan.*Derived|TestCreatePracticePlan.*Idempotency' -count=1`
  <!-- verified: 2026-05-16 `cd backend && go test ./internal/api/practice -run 'TestCreatePracticePlan.*Derived|TestCreatePracticePlan.*Idempotency|TestCreatePracticePlanReturns422ForValidationError|TestCreatePracticePlanFixtureParityDefault' -count=1` -->
- [x] 1.5 BDD-Gate: `E2E.P0.070` 覆盖 report/debrief derived plan create/read + replay
  <!-- verified: 2026-05-16 `cd backend && go test ./cmd/api -run TestE2EP0070PracticeDerivedPlanCreateReadReplay -count=1` -->

## Phase 2: debrief startPracticeSession First Turn Seeding

- [x] 2.1 Red: start debrief session 当前仍调用 first_question AI 或无法读取 debrief source；验证：`cd backend && go test ./internal/practice -run 'TestStartPracticeSession.*Debrief' -count=1` 先失败
  <!-- verified: 2026-05-16 red=`cd backend && go test ./internal/practice -run 'TestStartPracticeSession.*Debrief' -count=1` failed because SessionReservation lacked DebriefFirstQuestionText/DebriefFirstQuestionIntent fields -->
- [x] 2.2 Green: `ReserveSessionStart` 返回 debrief source first question；service 对 `goal='debrief'` 跳过 F3/A3 first_question；验证：`cd backend && go test ./internal/practice ./internal/store/practice -run 'TestStartPracticeSession.*Debrief|TestSQLRepositoryReserveSessionStart.*Debrief' -count=1`
  <!-- verified: 2026-05-16 `cd backend && go test ./internal/practice ./internal/store/practice -run 'TestStartPracticeSession.*Debrief|TestSQLRepositoryReserveSessionStart.*Debrief|TestSQLRepositoryReserveSessionStartReusesFailedRetryableRecord|TestSQLRepositoryReserveSessionStartResetsExpired|TestSQLRepositoryReserveSessionStartScopesIdempotencyByUser' -count=1` -->
- [x] 2.3 Handler/start session envelope：`startPracticeSession` 返回 currentTurn = debrief first question，status running，idempotency replay 保持相同 currentTurn；验证：`cd backend && go test ./internal/api/practice -run 'TestStartPracticeSession.*Debrief' -count=1`
  <!-- verified: 2026-05-16 `cd backend && go test ./internal/api/practice ./internal/store/practice -run 'TestStartPracticeSession.*Debrief|TestSQLRepositoryReserveSessionStartReplaysStoredResponseBody' -count=1` -->
- [x] 2.4 BDD-Gate: `E2E.P0.071` 覆盖 debrief plan start + no first_question AI call + first turn from raw_questions
  <!-- verified: 2026-05-16 `cd backend && go test ./cmd/api -run TestE2EP0071PracticeDebriefStartUsesSourceQuestion -count=1` -->

## Phase 3: Privacy, Isolation, And Legacy-Negative

- [x] 3.1 Source isolation matrix：missing/cross-user/wrong-target/draft/empty debrief source 均返回不泄露资源存在性的错误；验证：service + store + handler focused tests
  <!-- verified: 2026-05-16 `cd backend && go test ./internal/practice ./internal/store/practice -run 'TestServiceCreatePracticePlan.*Unavailable|TestSQLRepositoryCreatePlanReturnsPrerequisiteErrorWhenDerivedSourceUnavailable' -count=1` -->
- [x] 3.2 mode × goal regression：`goal='debrief'` 在 `mode='assisted'` / `mode='strict'` 下均可 start；`mode='debrief'` 非法；验证：service/API tests + generated enum grep
  <!-- verified: 2026-05-16 `cd backend && go test ./internal/practice ./internal/api/practice -run 'TestServiceCreatePracticePlanCreatesDerivedPlans|TestServiceCreatePracticePlanRejectsLegacyDebriefMode|TestStartPracticeSession.*Debrief' -count=1`; `rg -n "PracticeModeDebrief|mode='debrief'|mode=debrief|legacy debrief replay value|startDebriefInterview" backend/internal/practice backend/internal/api/practice backend/internal/store/practice openapi frontend/src/api/generated backend/internal/api/generated openapi/fixtures/PracticePlans openapi/fixtures/PracticeSessions -g '!**/*_test.go'` no matches -->
- [x] 3.3 Privacy redline：audit/outbox/log/metric/idempotency response 不含 debrief question text、answer summary、interviewer reaction、notes、risk prose；验证：HTTP scenario + scoped grep
  <!-- verified: 2026-05-16 `cd backend && go test ./internal/store/practice -run TestSQLRepositoryCommitSessionStartWritesAuditMetadataWithoutQuestionText -count=1`; `rg -n "__PRIVATE_DEBRIEF_TEXT__|__DEBRIEF_FIRST_QUESTION__" backend/internal/practice backend/internal/api/practice backend/internal/store/practice openapi frontend/src/api/generated backend/internal/api/generated openapi/fixtures/PracticePlans openapi/fixtures/PracticeSessions -g '!**/*_test.go'` no matches -->
- [x] 3.4 BDD-Gate: `E2E.P0.072` 覆盖 source validation / isolation / privacy
  <!-- verified: 2026-05-16 `cd backend && go test ./cmd/api -run TestE2EP0072PracticeDerivedSourceValidationIsolationPrivacy -count=1` -->
- [x] 3.5 BDD-Gate: `E2E.P0.073` 覆盖 debrief assisted/strict regression + legacy-negative
  <!-- verified: 2026-05-16 `cd backend && go test ./cmd/api -run TestE2EP0073PracticeDebriefAssistedStrictAndLegacyNegative -count=1` plus scoped legacy grep no runtime/generated/fixture matches -->

## Phase 4: Gates And Handoff

- [x] 4.1 Focused backend gates：`cd backend && go test ./internal/practice ./internal/store/practice ./internal/api/practice ./cmd/api -run 'Derived|Debrief|Source' -count=1` 通过
  <!-- verified: 2026-05-16 `cd backend && go test ./internal/practice ./internal/store/practice ./internal/api/practice ./cmd/api -run 'Derived|Debrief|Source' -count=1` -->
- [x] 4.2 Contract/migration gates：`make lint-openapi` / `make validate-fixtures` / `./migrations/lint.sh` / `make migrate-check` 通过
  <!-- verified: 2026-05-16 `make lint-openapi`, `make validate-fixtures`, `./migrations/lint.sh`, `set -a; . deploy/dev-stack/.env; set +a; make migrate-check` -->
- [x] 4.3 Repo hygiene gates：`make docs-check` / `git diff --check` / scoped legacy grep 通过
  <!-- verified: 2026-05-16 `make docs-check`, `git diff --check`, scoped runtime/generated/fixture legacy grep no matches outside negative tests -->
- [x] 4.4 backend-debrief handoff：更新 `backend-debrief/001` checklist 0.6 / test-checklist 0.F，记录 `backend-practice/004` 验证证据与依赖 commit
  <!-- verified: 2026-05-16 updated backend-debrief/001 checklist 0.6 and test-checklist 0.F with backend-practice/004 dependency evidence -->
