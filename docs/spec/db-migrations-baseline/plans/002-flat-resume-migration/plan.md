# DB Migrations Baseline Flat Resume Migration

> **版本**: 1.4
> **状态**: completed
> **更新日期**: 2026-07-07

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)

## 1 目标

本计划承接当前 flat Resume DB net-state：

- `resumes` 是当前简历表，承载 `original_text`、`parsed_text_snapshot`、`raw_text`、`file_object_id`、`structured_profile`、`display_name` 与 `source_type IN ('upload', 'paste')`。
- `practice_plans.resume_id` 是模拟面试规划绑定简历的当前 FK。
- Migration chain、enum source lint、Go migration contract tests 和 privacy deletion matrix 必须共同证明当前 DB baseline 与 [B4 spec](../../spec.md) 一致。

本计划不实现 Resume CRUD、Practice handler、前端 Resume Workshop UI 或 OpenAPI schema；这些由对应 owner 维护。

## 2 背景

当前产品已经收敛为 flat Resume + TargetJob + Practice + Report 的核心链路。B4 只负责 schema net-state 与迁移可执行性，不在 current owner 文档中保留阶段性 schema 叙事。

本计划的完成状态以当前文件事实为准：

- `migrations/000015_resume_flatten.up.sql` 将最终 app schema 收敛到 `resumes` 与 `practice_plans.resume_id`。
- `migrations/enum-sources.yaml` 与 `scripts/lint/migrations_lint.py` 负责 check source drift。
- `backend/internal/migrations/sql_contract_test.go` 负责 migration-chain contract assertions。
- `docs/spec/db-migrations-baseline/spec.md` §3.1.2 负责 privacy deletion table disposition。

## 3 质量门禁分类

- **Plan 类型**: `migration` + `contract`
- **TDD 策略**: 迁移实现已完成；任何修改 migration chain、enum source、backfill wrapper 或 privacy matrix 时，先补/改 Go migration contract test 或 migration lint fixture，再改 SQL / YAML / wrapper。
- **BDD 策略**: 不适用。本 plan 不产生用户可见 UI、API 行为或端到端业务流程。
- **替代验证 gate**:
  - `python3 scripts/lint/migrations_lint.py --repo-root .`
  - `cd backend && go test ./internal/migrations -count=1`
  - `python3 .agent-skills/implement/shared/scripts/validate_context.py --context docs/spec/db-migrations-baseline/plans/002-flat-resume-migration/context.yaml --target contract`
  - `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check`
  - `make docs-check`
  - `git diff --check`
  - `make migrate-check` when `DATABASE_URL` points to a disposable local DB.

## 4 当前交付内容

### Phase 1: Flat Resume Schema Net-State

#### 1.1 Current table shape

Current schema contract:

- `resumes.id` is the stable resume id used by API, backend stores, frontend generated clients and privacy delete.
- `resumes.structured_profile` stores parsed structured content.
- `resumes.display_name` stores user-facing resume display name.
- `resumes.source_type` accepts only `upload` and `paste`.
- `practice_plans.resume_id` references `resumes` and is nullable with current store semantics.

#### 1.2 Migration-chain contract

The migration chain must keep both directions executable:

- Up path lands the current schema net-state.
- Down files provide structural rollback for dev / CI checks.
- Go contract tests assert the current table, column, check and rollback structures directly from SQL files.

### Phase 2: Enum / Check / Backfill Contract

#### 2.1 Enum source lint

`migrations/enum-sources.yaml` remains the owner-readable source map for SQL `CHECK (...)` values. Migration lint must reject check lists without source attribution.

#### 2.2 In-migration copy before table replacement

The flat Resume migration copies structured content into `resumes.structured_profile` inside the SQL transaction before dependent rollback structures are touched. This avoids relying on Go backfill order for a source table that no longer exists after the SQL chain completes.

#### 2.3 Backfill ledger boundary

`schema_backfills` remains for row-level Go registry work. The flat Resume conversion is SQL-local because its source and destination live inside the same migration transaction.

### Phase 3: Privacy / Consumer Handoff

#### 3.1 Privacy matrix

`file_objects` / `resumes` are deleted with object storage cleanup. `practice_plans` and downstream practice/report rows cascade or hard-delete by the order in [B4 spec §3.1.2](../../spec.md#312-p0-privacy-deletion-table-matrix).

#### 3.2 Consumer boundary

Backend Resume, Backend Practice, OpenAPI and frontend consumers use `resumeId` / `resume_id` as the current binding. B4 verifies schema support only; each consumer owner verifies handler, API, fixture and UI behavior.

### Phase 4: Verification Evidence

#### 4.1 Current local evidence

Current green gates:

- `python3 scripts/lint/migrations_lint.py --repo-root .`
- `cd backend && go test ./internal/migrations -run 'TestResumeVersionsAdditiveMigrationContract|TestResumeFlattenMigrationContract|TestDropJDMatchMigrationDeletesNonCurrentAsyncJobsBeforeNarrowingCheck|TestDropJDMatchMigrationDropsNonCurrentTablesAndRegistryRows' -count=1`
- `cd backend && go test ./internal/migrations -count=1`

#### 4.2 Environment-dependent evidence

`make migrate-check` requires `DATABASE_URL`. In the current shell, `DATABASE_URL` is missing, so live up/down migration execution is not claimed by this document compression pass. When a disposable local DB is available, B4 owner should run `make migrate-check` before changing migration SQL or wrapper behavior.

## 5 验收标准

| ID | 场景 | Given | When | Then | 证据 |
|----|------|-------|------|------|------|
| C-1 | Flat Resume schema | Migration SQL chain | Static contract tests | Current `resumes` shape and `practice_plans.resume_id` are asserted | `go test ./internal/migrations -count=1` |
| C-2 | Enum/check source | SQL check lists | Migration lint | Check values have registered source ownership | `migrations_lint.py` |
| C-3 | Privacy matrix | Current B4 spec | Docs and runner consumers | `resumes` and related user tables have disposition | B4 spec §3.1.2 |
| C-4 | Context validity | Plan context | `validate_context.py` | Plan, checklist and spec paths resolve | context gate |
| C-5 | Live migration loop | Disposable local DB | `make migrate-check` | up -> down -> up succeeds | environment-dependent gate |

## 6 风险与应对

| 风险 | 应对 |
|------|------|
| Migration-chain lint scans SQL files that are not current net-state | Keep source ownership in `migrations/enum-sources.yaml`; use current app inventory from B4 spec for product scope decisions |
| Live migration loop unavailable in shell | Do not claim live up/down success; run `make migrate-check` once `DATABASE_URL` points to a disposable DB |
| Consumer docs drift from flat Resume schema | Consumer owner plans must link this plan and B4 spec, then verify generated clients / handlers / UI gates |
