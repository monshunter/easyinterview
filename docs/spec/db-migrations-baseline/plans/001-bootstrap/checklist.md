# DB Migrations Baseline Bootstrap Checklist

> **版本**: 1.1
> **状态**: active
> **更新日期**: 2026-04-30

**关联计划**: [plan](./plan.md)

## Phase 1: migration skeleton 与工具 wrapper

- [x] 1.1 落地 `backend/cmd/migrate/main.go`，包装 `golang-migrate v4.18+` 与 B4 backfill registry。验证: 新增/更新 wrapper focused tests 覆盖 up/down/status/create/check 命令解析、database URL 缺失错误、prod down 防呆入口，并运行 `go test ./cmd/migrate ./internal/migrations/... -count=1`
- [x] 1.2 根 `Makefile` 提供 `make migrate-up` / `make migrate-down` / `make migrate-status` / `make migrate-create NAME=...` / `make migrate-check`，A1 占位 `make migrate` 指向真实入口或 help。验证: `make migrate` / `make migrate-status` / `make migrate-create NAME=test_migration` smoke 走 wrapper 或清晰 help，生成的临时 migration 文件可清理后 `git diff --check` 通过
- [x] 1.3 `migrations/` 使用 `NNNNNN_<verb>_<noun>.up.sql` / `.down.sql`，编号从 `000001` 起 6 位连续，并由 `make migrate-create` 生成。验证: migration file naming test / lint 覆盖非 6 位、断号、缺 up/down pair、手写异常名失败路径；`make migrate-check` 调用该检查
- [x] 1.4 baseline up 幂等启用 `vector` extension；down 默认保留 extension，仅 dev + `MIGRATE_DROP_EXTENSIONS=1` 允许 drop。验证: 干净 Postgres 上 `make migrate-up && make migrate-down && make migrate-up` 后 `select extname from pg_extension where extname='vector'` 存在；设置 `MIGRATE_DROP_EXTENSIONS=1 APP_ENV=dev make migrate-down` 时 extension drop 行为有 focused smoke 或 SQL probe
- [x] 1.5 落地 `schema_backfills` ledger、`migrations/backfill/manifest.yaml` 与 1 个 dry-run/apply 示例 registry。验证: Go backfill registry tests 覆盖 manifest 解析、dry-run/apply 状态写入、同一 `version + mode + checksum` 重复成功记录不重复执行、`--force` 在 `APP_ENV=prod` 被拒绝

## Phase 2: baseline DDL 与索引

- [x] 2.1 落地 03-db-definition 27 张 P0 应用表 + ADR-Q1 `auth_challenges` / `sessions` / `external_identities` 3 张支撑表。验证: SQL inventory probe 断言 30 张应用 / auth 支撑表全部存在，且关键 FK / soft-delete / sensitive hash 字段符合 spec §4.2 / §4.4
- [x] 2.2 `make migrate-up` 后 public schema table count ≥32（含 `schema_migrations` / `schema_backfills`）。验证: 干净 Postgres 上 `make migrate-up` 后执行 `select count(*) from information_schema.tables where table_schema='public'`，结果 ≥32 并记录在 handoff
- [x] 2.3 `outbox_events` 包含 `publish_attempts` / `next_attempt_at` / `locked_at` / `last_error_code` / `last_error_message`，并有 `(publish_status, next_attempt_at, created_at)` pending due 查询索引。验证: information_schema column probe + `pg_indexes` probe + pending due `EXPLAIN` 命中对应索引
- [x] 2.4 `async_jobs.job_type` check 包含 B3 10 个 canonical jobType（含 internal-only `email_dispatch`），且 B2 API-facing subset 仍为 7 项。验证: migration lint 读取 B3/B2 manifests 后断言 DB check 值等于 B3 canonical 10 项，且 B2 API-facing subset 未被 internal-only `email_dispatch` 扩大
- [x] 2.5 `ai_task_runs` 包含 `model_family` / `model_profile_name` / `model_profile_version` / `fallback_chain` / `route` / `validation_status` / `output_schema_version` typed columns。验证: information_schema probe 断言 typed columns、`fallback_chain jsonb not null default '[]'::jsonb`，并有 dashboard 查询不依赖 JSONB path scan 的 SQL/explain probe
- [x] 2.6 覆盖 03 §7 B-Tree 索引、`retrieval_chunks.embedding` ivfflat 与可选 `target_jobs` GIN 全文索引。验证: `pg_indexes` inventory 与关键 query `EXPLAIN` probes 覆盖 B-Tree、`idx_retrieval_chunks_embedding` ivfflat、dev 默认 `target_jobs` GIN 全文索引

## Phase 3: enum/check 来源、backfill 与 privacy lint

- [ ] 3.1 落地 `migrations/enum-sources.yaml`，逐列登记 `table.column -> source -> checksum`。验证: lint/unit test 覆盖缺登记、checksum 漂移、B1/B2/B3 source 读取失败、DB-local enum 合法路径；`make migrate-check` 调用该 lint
- [ ] 3.2 落地 `scripts/lint/migrations_lint.py`（或等价 Go 工具），SQL 中出现未登记 `check (col in (...))` 必须 fail。验证: pytest/Go tests 使用临时 SQL fixture 覆盖 registered check 通过、unregistered check 失败、`token_hash` allowlist 通过、`raw_token` / `session_cookie` / `api_key` 明文语义字段失败
- [ ] 3.3 `backend/internal/migrations/backfills/<version>/` 注册 dry-run/apply；同一 `version + mode + checksum` apply 成功后不得重复执行。验证: backfill registry tests 覆盖 dry-run 不改数据、apply 写 ledger、重复 apply skip/fail、`--force` 非 prod 可重跑且 prod 被拒绝
- [ ] 3.4 提供 privacy deletion matrix dry-run 入口，覆盖 spec §3.1.2 所有表组 disposition。验证: dry-run fixture 输出包含 spec §3.1.2 全部表组，且 `prompt_versions` / `rubric_versions` / `schema_migrations` / `schema_backfills` 为 retain

## Phase 4: Verification + handoff

- [ ] 4.1 在干净 Postgres 16 上执行 `make migrate-check`：`migrate-up -> migrate-down -> migrate-up` 全部成功。验证: 使用 A2 dev stack 或本地等价 Postgres 16 运行 `make migrate-check`，记录 migrate-up/down/up、backfill ledger 去重、exit 0 输出
- [ ] 4.2 `APP_ENV=prod make migrate-down` fail-fast；stderr 提示需显式 force / 操作窗口。验证: `APP_ENV=prod make migrate-down` exit 非 0，stderr 包含 `MIGRATE_DOWN_FORCE=1` 或等价操作窗口提示；不连接 DB 或不执行 down SQL
- [ ] 4.3 SQL probes 验证 table count ≥32、outbox retry columns、AI typed columns、ivfflat / pending due / B-Tree 索引存在且 explain 命中。验证: probe 命令输出保存到工作日志，覆盖 spec C-1 / C-2 / C-8 / C-9 / C-11 / C-12
- [ ] 4.4 临时修改 B3 job manifest 或 B1 enum，确认 `make migrate-check` / lint 报 drift；revert 后恢复。验证: negative drift case 先失败并指向具体 table.column/source/checksum，恢复 manifest 后 `make migrate-check` 通过且 `git diff --check` 通过
- [ ] 4.5 privacy deletion dry-run 输出覆盖 spec §3.1.2 所有表组，且 `prompt_versions` / `rubric_versions` / migration metadata 被 retain。验证: dry-run fixture/probe 输出覆盖所有 disposition，retain 表组无用户内容删除动作，结果记录到工作日志
- [ ] 4.6 本 plan checklist 全部勾选后，将 Header 切 completed，运行 sync-doc-index check/fix，并在 work journal 记录 migrate-check、SQL probes、lint probes 与 downstream handoff。验证: `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check` zero drift，work journal 有本 plan commit 记录与下游 handoff
