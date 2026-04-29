# DB Migrations Baseline Bootstrap Checklist

> **版本**: 1.0
> **状态**: active
> **更新日期**: 2026-04-29

**关联计划**: [plan](./plan.md)

## Phase 1: migration skeleton 与工具 wrapper

- [ ] 1.1 落地 `backend/cmd/migrate/main.go`，包装 `golang-migrate v4.18+` 与 B4 backfill registry
- [ ] 1.2 根 `Makefile` 提供 `make migrate-up` / `make migrate-down` / `make migrate-status` / `make migrate-create NAME=...` / `make migrate-check`，A1 占位 `make migrate` 指向真实入口或 help
- [ ] 1.3 `migrations/` 使用 `NNNNNN_<verb>_<noun>.up.sql` / `.down.sql`，编号从 `000001` 起 6 位连续，并由 `make migrate-create` 生成
- [ ] 1.4 baseline up 幂等启用 `vector` extension；down 默认保留 extension，仅 dev + `MIGRATE_DROP_EXTENSIONS=1` 允许 drop
- [ ] 1.5 落地 `schema_backfills` ledger、`migrations/backfill/manifest.yaml` 与 1 个 dry-run/apply 示例 registry

## Phase 2: baseline DDL 与索引

- [ ] 2.1 落地 03-db-definition 27 张 P0 应用表 + ADR-Q1 `auth_challenges` / `sessions` / `external_identities` 3 张支撑表
- [ ] 2.2 `make migrate-up` 后 public schema table count ≥32（含 `schema_migrations` / `schema_backfills`）
- [ ] 2.3 `outbox_events` 包含 `publish_attempts` / `next_attempt_at` / `locked_at` / `last_error_code` / `last_error_message`，并有 `(publish_status, next_attempt_at, created_at)` pending due 查询索引
- [ ] 2.4 `async_jobs.job_type` check 包含 B3 10 个 canonical jobType（含 internal-only `email_dispatch`），且 B2 API-facing subset 仍为 7 项
- [ ] 2.5 `ai_task_runs` 包含 `model_family` / `model_profile_name` / `model_profile_version` / `fallback_chain` / `route` / `validation_status` / `output_schema_version` typed columns
- [ ] 2.6 覆盖 03 §7 B-Tree 索引、`retrieval_chunks.embedding` ivfflat 与可选 `target_jobs` GIN 全文索引

## Phase 3: enum/check 来源、backfill 与 privacy lint

- [ ] 3.1 落地 `migrations/enum-sources.yaml`，逐列登记 `table.column -> source -> checksum`
- [ ] 3.2 落地 `scripts/lint/migrations_lint.py`（或等价 Go 工具），SQL 中出现未登记 `check (col in (...))` 必须 fail
- [ ] 3.3 `backend/internal/migrations/backfills/<version>/` 注册 dry-run/apply；同一 `version + mode + checksum` apply 成功后不得重复执行
- [ ] 3.4 提供 privacy deletion matrix dry-run 入口，覆盖 spec §3.1.2 所有表组 disposition

## Phase 4: Verification + handoff

- [ ] 4.1 在干净 Postgres 16 上执行 `make migrate-check`：`migrate-up -> migrate-down -> migrate-up` 全部成功
- [ ] 4.2 `APP_ENV=prod make migrate-down` fail-fast；stderr 提示需显式 force / 操作窗口
- [ ] 4.3 SQL probes 验证 table count ≥32、outbox retry columns、AI typed columns、ivfflat / pending due / B-Tree 索引存在且 explain 命中
- [ ] 4.4 临时修改 B3 job manifest 或 B1 enum，确认 `make migrate-check` / lint 报 drift；revert 后恢复
- [ ] 4.5 privacy deletion dry-run 输出覆盖 spec §3.1.2 所有表组，且 `prompt_versions` / `rubric_versions` / migration metadata 被 retain
- [ ] 4.6 本 plan checklist 全部勾选后，将 Header 切 completed，运行 sync-doc-index check/fix，并在 work journal 记录 migrate-check、SQL probes、lint probes 与 downstream handoff
