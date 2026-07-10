# DB Migrations Baseline Flat Resume Migration Checklist

> **版本**: 1.5
> **状态**: completed
> **更新日期**: 2026-07-10

**关联计划**: [plan](./plan.md)

## Phase 1: Flat Resume schema net-state

- [x] 1.1 `resumes` is the current resume table with `structured_profile`, `display_name`, source snapshots and `source_type` limited to `upload` / `paste`.
- [x] 1.2 `practice_plans.resume_id` is the current practice-plan resume binding.
- [x] 1.3 Migration-chain contract tests assert up/down structure for the current flat Resume schema.

## Phase 2: Enum / check / backfill contract

- [x] 2.1 `migrations/enum-sources.yaml` explains SQL check sources and passes migration lint.
- [x] 2.2 Flat Resume content copy is SQL-local inside the migration transaction and does not rely on Go backfill registry order.
- [x] 2.3 `schema_backfills` remains scoped to row-level Go registry work.

## Phase 3: Privacy / consumer handoff

- [x] 3.1 B4 spec §3.1.2 covers `file_objects` / `resumes` and related user rows in the privacy deletion matrix.
- [x] 3.2 Backend Resume, Backend Practice, OpenAPI and frontend owners consume `resumeId` / `resume_id`; B4 owns only schema support.

## Phase 4: Verification evidence

- [x] 4.1 `python3 scripts/lint/migrations_lint.py --repo-root .` PASS.
- [x] 4.2 `cd backend && go test ./internal/migrations -run 'TestResumeVersionsAdditiveMigrationContract|TestResumeFlattenMigrationContract|TestDropJDMatchMigrationDeletesOutOfScopeAsyncJobsBeforeNarrowingCheck|TestDropJDMatchMigrationDropsOutOfScopeTablesAndRegistryRows' -count=1` PASS.
- [x] 4.3 `cd backend && go test ./internal/migrations -count=1` PASS.
- [x] 4.4 `DATABASE_URL` is missing in the current shell; this pass does not claim live `make migrate-check` up/down execution.
