# DB Migrations Baseline Bootstrap Checklist

> **版本**: 1.7
> **状态**: completed
> **更新日期**: 2026-07-06

**关联计划**: [plan](./plan.md)

## Phase 1: migration skeleton 与工具 wrapper

- [x] 1.1 落地 `backend/cmd/migrate/main.go`，包装 `golang-migrate v4.18+` 与 B4 backfill registry。验证: 新增/更新 wrapper focused tests 覆盖 up/down/status/create/check 命令解析、database URL 缺失错误、prod down 防呆入口，并运行 `go test ./cmd/migrate ./internal/migrations/... -count=1`
- [x] 1.2 根 `Makefile` 提供 `make migrate-up` / `make migrate-down` / `make migrate-status` / `make migrate-create NAME=...` / `make migrate-check`，A1 占位 `make migrate` 指向真实入口或 help。验证: `make migrate` / `make migrate-status` / `make migrate-create NAME=test_migration` smoke 走 wrapper 或清晰 help，生成的临时 migration 文件可清理后 `git diff --check` 通过
- [x] 1.3 `migrations/` 使用 `NNNNNN_<verb>_<noun>.up.sql` / `.down.sql`，编号从 `000001` 起 6 位连续，并由 `make migrate-create` 生成。验证: migration file naming test / lint 覆盖非 6 位、断号、缺 up/down pair、手写异常名失败路径；`make migrate-check` 调用该检查
- [x] 1.4 baseline up 不启用未使用 DB extension；down 不管理 extension 生命周期。验证: 干净 Postgres 上 `make migrate-up && make migrate-down && make migrate-up` 成功，且 SQL contract test 断言 baseline 不创建未使用 extension
- [x] 1.5 落地 `schema_backfills` ledger、`migrations/backfill/manifest.yaml` 与 1 个 dry-run/apply 示例 registry。验证: Go backfill registry tests 覆盖 manifest 解析、dry-run/apply 状态写入、同一 `version + mode + checksum` 重复成功记录不重复执行、`--force` 在 `APP_ENV=prod` 被拒绝
- [x] 1.6 L2 remediation: prod down 防呆必须在连接 DB 前失败，避免误执行 destructive down。验证: focused CLI test 覆盖 prod guard；`go test ./internal/migrations -run 'TestCommandRunDownRequiresForceInProd' -count=1`

## Phase 2: baseline DDL 与索引

- [x] 2.1 落地当前产品范围内 22 张应用表 + ADR-Q1 `auth_challenges` / `sessions` / `external_identities` 3 张支撑表；旧 `mistake_entries`、JD Match、简历版本树、候选人画像与真实复盘表不再作为 current baseline 创建。验证: SQL inventory probe 断言 25 张当前应用 / auth 支撑表全部存在，且关键 FK / soft-delete / sensitive hash 字段符合 spec §4.2 / §4.4
- [x] 2.2 `make migrate-up` 后 public schema table count ≥27（含 `schema_migrations` / `schema_backfills`）。验证: 干净 Postgres 上 `make migrate-up` 后执行 `select count(*) from information_schema.tables where table_schema='public'`，结果 ≥27 并记录在 handoff
- [x] 2.3 `outbox_events` 包含 `publish_attempts` / `next_attempt_at` / `locked_at` / `last_error_code` / `last_error_message`，并有 `(publish_status, next_attempt_at, created_at)` pending due 查询索引。验证: information_schema column probe + `pg_indexes` probe + pending due `EXPLAIN` 命中对应索引
- [x] 2.4 `async_jobs.job_type` check 包含 B3 当前 8 个 canonical jobType（含 internal-only `source_refresh` / `email_dispatch` 与 contract-only `privacy_export`），且 B2 API-facing subset 仍为 6 项。验证: migration lint 读取 B3/B2 manifests 后断言 DB check 值等于 B3 canonical 8 项，且 B2 API-facing subset 未被 internal-only `source_refresh` / `email_dispatch` 扩大
- [x] 2.5 `ai_task_runs` 包含 `model_family` / `model_profile_name` / `model_profile_version` / `fallback_chain` / `route` / `validation_status` / `output_schema_version` typed columns。验证: information_schema probe 断言 typed columns、`fallback_chain jsonb not null default '[]'::jsonb`，并有 dashboard 查询不依赖 JSONB path scan 的 SQL/explain probe
- [x] 2.6 覆盖 B4 B-Tree 索引与可选 `target_jobs` GIN 全文索引。验证: `pg_indexes` inventory 与关键 query `EXPLAIN` probes 覆盖 B-Tree 与 dev 默认 `target_jobs` GIN 全文索引

## Phase 3: enum/check 来源、backfill 与 privacy lint

- [x] 3.1 落地 `migrations/enum-sources.yaml`，逐列登记 `table.column -> source -> checksum`。验证: lint/unit test 覆盖缺登记、checksum 漂移、B1/B2/B3 source 读取失败、DB-local enum 合法路径；`make migrate-check` 调用该 lint
- [x] 3.2 落地 `scripts/lint/migrations_lint.py`（或等价 Go 工具），SQL 中出现未登记 `check (col in (...))` 必须 fail。验证: pytest/Go tests 使用临时 SQL fixture 覆盖 registered check 通过、unregistered check 失败、`token_hash` allowlist 通过、`raw_token` / `session_cookie` / `api_key` 明文语义字段失败
- [x] 3.3 `backend/internal/migrations/backfills/<version>/` 注册 dry-run/apply；同一 `version + mode + checksum` apply 成功后不得重复执行。验证: backfill registry tests 覆盖 dry-run 不改数据、apply 写 ledger、重复 apply skip/fail、`--force` 非 prod 可重跑且 prod 被拒绝
- [x] 3.4 提供 privacy deletion matrix dry-run 入口，覆盖 spec §3.1.2 所有表组 disposition。验证: dry-run fixture 输出包含 spec §3.1.2 全部表组，且 `prompt_versions` / `rubric_versions` / `schema_migrations` / `schema_backfills` 为 retain
- [x] 3.5 L2 remediation: migration lint 必须读取 B1 `shared/conventions.yaml`、B3 `shared/jobs.yaml` 与 B2 OpenAPI `JobType`，并覆盖 `ALTER TABLE ... CHECK (...)` 约束发现。验证: pytest 覆盖 shared source drift、B3/B2 job subset drift、missing source file 与 ALTER TABLE unregistered check；`python3 -m pytest scripts/lint/migrations_lint_test.py -q`

## Phase 4: Verification + handoff

- [x] 4.1 在干净 Postgres 18 上执行 `make migrate-check`：`migrate-up -> migrate-down -> migrate-up` 全部成功。验证: 使用 A2 dev stack 或本地等价 Postgres 18 运行 `make migrate-check`，记录 migrate-up/down/up、backfill ledger 去重、exit 0 输出
- [x] 4.2 `APP_ENV=prod make migrate-down` fail-fast；stderr 提示需显式 force / 操作窗口。验证: `APP_ENV=prod make migrate-down` exit 非 0，stderr 包含 `MIGRATE_DOWN_FORCE=1` 或等价操作窗口提示；不连接 DB 或不执行 down SQL
- [x] 4.3 SQL probes 验证 table count ≥27、outbox retry columns、AI typed columns、pending due / B-Tree 索引存在且 explain 命中。验证: probe 命令输出保存到工作日志，覆盖 spec C-1 / C-2 / C-8 / C-11 / C-12
- [x] 4.4 临时修改 B3 job manifest 或 B1 enum，确认 `make migrate-check` / lint 报 drift；revert 后恢复。验证: negative drift case 先失败并指向具体 table.column/source/checksum，恢复 manifest 后 `make migrate-check` 通过且 `git diff --check` 通过
- [x] 4.5 privacy deletion dry-run 输出覆盖 spec §3.1.2 所有表组，且 `prompt_versions` / `rubric_versions` / migration metadata 被 retain。验证: dry-run fixture/probe 输出覆盖所有 disposition，retain 表组无用户内容删除动作，结果记录到工作日志
- [x] 4.6 本 plan checklist 全部勾选后，将 Header 切 completed，运行 sync-doc-index check/fix，并在 work journal 记录 migrate-check、SQL probes、lint probes 与 downstream handoff。验证: `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check` zero drift，work journal 有本 plan commit 记录与下游 handoff
- [x] 4.7 L2 remediation: prod down 防呆必须在 `DATABASE_URL` 校验前触发，确保无 DB URL 时仍输出 `MIGRATE_DOWN_FORCE=1` 且不执行 down SQL。验证: focused CLI test 与 `APP_ENV=prod make migrate-down` smoke 都先命中 prod guard

## Phase 5: product-scope v1.2 schema remediation

- [x] 5.1 Red: migration inventory / contract tests 期望排除 `mistake_entries`、旧字段和旧 practice enum 后，当前 SQL / probes 必须失败
  - 2026-05-03: `python3 scripts/lint/migrations_lint.py --repo-root .` exit 1，报 `mistake_entries.status` 的 `MistakeStatus` source 缺失，以及 `practice_plans.goal` / `mode` 与 B1 新枚举漂移；更新 SQL contract expectation 后，`cd backend && go test ./internal/migrations -run TestBaselineMigrationDefinesAllOwnedTables -count=1` exit 1，失败于仍创建 removed `mistake_entries` table。
- [x] 5.2 Green: 修订 baseline migration、enum source、privacy matrix 和 SQL contract tests，删除独立 `mistake_entries`，字段改为 `open_question_issue_count` / `included_in_retry_plan` / `review_status`
  - 2026-05-03: 删除 baseline migration 中 `mistake_entries` DDL / indexes / FK / down drop；`target_jobs.open_mistake_count` 改 `open_question_issue_count`；`question_assessments.written_to_mistake_book` 改 `review_status` + `included_in_retry_plan`；`practice_plans` enum 更新为 B1 当前 `PracticeGoal` / `PracticeMode`；`enum-sources.yaml` 和 privacy matrix 同步。
- [x] 5.3 Verify: `make migrate-check` 或 migration lint / SQL contract tests 通过；repo 搜索确认实现侧无 `mistake_entries`、`open_mistake_count`、`written_to_mistake_book`、旧 practice mode / goal check 值
  - 2026-05-03: `python3 scripts/lint/migrations_lint.py --repo-root .` exit 0；`python3 -m pytest scripts/lint/migrations_lint_test.py -q` 8 passed；`cd backend && go test ./internal/migrations ./cmd/migrate -count=1` pass；`make privacy-delete-dry-run` 输出不含 removed `mistake_entries`；实现侧搜索仅剩 SQL contract 负向断言，未在 migration/runtime 实现中命中旧字段或旧 enum 值。完整 `make migrate-check` 需真实 DB wrapper，本阶段按 checklist 采用 migration lint / SQL contract tests 验证。
