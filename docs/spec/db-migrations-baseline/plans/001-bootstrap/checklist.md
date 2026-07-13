# DB Migrations Baseline Bootstrap Checklist

> **版本**: 1.21
> **状态**: active
> **更新日期**: 2026-07-13

**关联计划**: [plan](./plan.md)

## Phase 1: migration skeleton 与工具 wrapper

- [x] 1.1 落地 `backend/cmd/migrate/main.go`，包装 `golang-migrate v4.18+` 与 B4 backfill registry。验证: 新增/更新 wrapper focused tests 覆盖 up/down/status/create/check 命令解析、database URL 缺失错误、prod down 防呆入口，并运行 `go test ./cmd/migrate ./internal/migrations/... -count=1`
- [x] 1.2 根 `Makefile` 提供 `make migrate-up` / `make migrate-down` / `make migrate-status` / `make migrate-create NAME=...` / `make migrate-check`，根 `make migrate` 指向真实入口或 help。验证: `make migrate` / `make migrate-status` / `make migrate-create NAME=test_migration` smoke 走 wrapper 或清晰 help，生成的临时 migration 文件可清理后 `git diff --check` 通过
- [x] 1.3 `migrations/` 使用 `NNNNNN_<verb>_<noun>.up.sql` / `.down.sql`，编号从 `000001` 起 6 位连续，并由 `make migrate-create` 生成。验证: migration file naming test / lint 覆盖非 6 位、断号、缺 up/down pair、手写异常名失败路径；`make migrate-check` 调用该检查
- [x] 1.4 baseline up 不启用未使用 DB extension；down 不管理 extension 生命周期。验证: 干净 Postgres 上 `make migrate-up && make migrate-down && make migrate-up` 成功，且 SQL contract test 断言 baseline 不创建未使用 extension
- [x] 1.5 落地 `schema_backfills` ledger 与 backfill manifest contract；当前登记 `v000017/practice_plan_round_identity`。验证: Go backfill registry tests 覆盖 manifest 解析、dry-run/apply 状态写入、同一 `version + mode + checksum` 重复成功记录不重复执行、`--force` 在 `APP_ENV=prod` 被拒绝<!-- verified: 2026-07-12 method=go-test package=./internal/migrations/... manifest=v000017 -->
- [x] 1.6 L2 remediation: prod down 防呆必须在连接 DB 前失败，避免误执行 destructive down。验证: focused CLI test 覆盖 prod guard；`go test ./internal/migrations -run 'TestCommandRunDownRequiresForceInProd' -count=1`

## Phase 2: baseline DDL 与索引

- [x] 2.1 final schema 保留当前 21 张应用表 + ADR-Q1 `auth_challenges` / `sessions` / `external_identities` 3 张支撑表；只允许这 24 张当前应用 / auth 支撑表。验证: SQL inventory probe 断言 24 张表全部存在，且关键 FK / soft-delete / sensitive hash 字段符合 spec §4.2 / §4.4
- [x] 2.2 `make migrate-up` 后 public schema table count ≥26（含 `schema_migrations` / `schema_backfills`）。验证: 干净 Postgres 上执行 current full migration chain 后查询 `information_schema.tables`，结果 ≥26 并记录在 handoff
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
- [x] 4.3 SQL probes 验证 current table count ≥26、outbox retry columns、AI typed columns、pending due / B-Tree 索引存在且 explain 命中。验证: probe 命令输出保存到工作日志，覆盖 spec C-1 / C-2 / C-8 / C-11 / C-12
- [x] 4.4 临时修改 B3 job manifest 或 B1 enum，确认 `make migrate-check` / lint 报 drift；revert 后恢复。验证: negative drift case 先失败并指向具体 table.column/source/checksum，恢复 manifest 后 `make migrate-check` 通过且 `git diff --check` 通过
- [x] 4.5 privacy deletion dry-run 输出覆盖 spec §3.1.2 所有表组，且 `prompt_versions` / `rubric_versions` / migration metadata 被 retain。验证: dry-run fixture/probe 输出覆盖所有 disposition，retain 表组无用户内容删除动作，结果记录到工作日志
- [x] 4.6 本 plan checklist 全部勾选后，将 Header 切 completed，运行 sync-doc-index check/fix，并在 work journal 记录 migrate-check、SQL probes、lint probes 与 downstream handoff。验证: `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check` zero drift，work journal 有本 plan commit 记录与下游 handoff
- [x] 4.7 L2 remediation: prod down 防呆必须在 `DATABASE_URL` 校验前触发，确保无 DB URL 时仍输出 `MIGRATE_DOWN_FORCE=1` 且不执行 down SQL。验证: focused CLI test 与 `APP_ENV=prod make migrate-down` smoke 都先命中 prod guard

## Phase 5: product-scope v1.2 schema remediation

- [x] 5.1 Red: migration inventory / contract tests 期望排除 `mistake_entries`、范围外字段和未登记 practice enum 后，当前 SQL / probes 必须失败
  - 2026-05-03: `python3 scripts/lint/migrations_lint.py --repo-root .` exit 1，报 `mistake_entries.status` 的 `MistakeStatus` source 缺失，以及 `practice_plans.goal` / `mode` 与 B1 新枚举漂移；更新 SQL contract expectation 后，`cd backend && go test ./internal/migrations -run TestBaselineMigrationDefinesAllOwnedTables -count=1` exit 1，失败于仍创建 removed `mistake_entries` table。
- [x] 5.2 Green: 修订 baseline migration、enum source、privacy matrix 和 SQL contract tests，删除独立 `mistake_entries`，字段改为 `open_question_issue_count` / `included_in_retry_plan` / `review_status`
  - 2026-05-03: 删除 baseline migration 中 `mistake_entries` DDL / indexes / FK / down drop；`target_jobs.open_mistake_count` 改 `open_question_issue_count`；`question_assessments.written_to_mistake_book` 改 `review_status` + `included_in_retry_plan`；`practice_plans` enum 更新为 B1 当前 `PracticeGoal` / `PracticeMode`；`enum-sources.yaml` 和 privacy matrix 同步。
- [x] 5.3 Verify: `make migrate-check` 或 migration lint / SQL contract tests 通过；repo 搜索确认实现侧无 `mistake_entries`、`open_mistake_count`、`written_to_mistake_book`、旧 practice mode / goal check 值
  - 2026-05-03: `python3 scripts/lint/migrations_lint.py --repo-root .` exit 0；`python3 -m pytest scripts/lint/migrations_lint_test.py -q` 8 passed；`cd backend && go test ./internal/migrations ./cmd/migrate -count=1` pass；`make privacy-delete-dry-run` 输出不含 `mistake_entries`；实现侧搜索仅剩 SQL contract 负向断言，migration/runtime 实现中没有范围外字段或 enum 值。完整 `make migrate-check` 需真实 DB wrapper，本阶段按 checklist 采用 migration lint / SQL contract tests 验证。

## Phase 6: Migration CLI test-double cleanup

- [x] 6.1 删除生产包导出的 `StaticEnv` test helper，将 map-backed `Env` double 下沉到 `cli_test.go`，并确认 `Run` 仍由 `cmd/migrate.osEnv` 驱动；验证 production deadcode、symbol inventory、migration focused tests/staticcheck/lint、owner contexts 与 docs/diff/pruning gates。
  <!-- verified: 2026-07-10 method=migration-cli-test-double-relocation evidence="Production deadcode RED listed StaticEnv.Getenv. Removed the exported production test helper, added test-local mapEnv and retained Env/osEnv. Migration focused/full backend tests, staticcheck, lint 18 tests, config lint, privacy dry-run, symbol inventory and owner contexts PASS." -->

## Phase 7: normalized PracticePlan round identity

- [x] 7.1 RED: migration SQL/live tests require nullable paired `round_id` / `round_sequence`, positive sequence, partial lookup index, reversible down, and no TargetJob progress column.<!-- verified: 2026-07-12 method=static-contract-red missing=000017 -->
- [x] 7.2 Create `000017` only through `make migrate-create NAME=practice_plan_round_identity`; implement up/down and update schema inventory probes.<!-- verified: 2026-07-12 method=make-migrate-create+sql-contract evidence="paired nullable columns, nonblank/positive checks, plan/session indexes and reversible down" -->
- [x] 7.3 RED-GREEN: backfill registry covers unique duration match, zero match, same-duration ambiguity, canonical `round-{sequence}-{type}`, rerun idempotency, and ledger evidence; ambiguous rows remain null.<!-- verified: 2026-07-12 method=go-test package=internal/migrations/backfills/v000017 evidence="dry-run/apply/rerun plus manifest and ledger" -->
- [x] 7.4 Run migration lint/contract tests and real Postgres `make migrate-check` up/down/up; record pair/index/backfill probes.<!-- verified: 2026-07-12 method=isolated-postgres evidence="two migrate-check up/down/up runs; round-ddl-probe PASS; backfill probe ledger=2; temporary-db-residual=0" -->

## Phase 8: HISTORICAL-SUPERSEDED grounded direct report storage

- [x] 8.1 HISTORICAL-SUPERSEDED: SQL contract tests required nullable-until-ready summary, generation_context, `llm_attempt_count integer NOT NULL DEFAULT 0 CHECK (0..4)`, report+plan dimension-focus columns, exact 21+3+2 current inventory, and rejected the superseded boolean repair flag plus old competency names. <!-- verified: 2026-07-13 method=pytest result="historical contract only; superseded by Phase 9" -->
- [x] 8.2 HISTORICAL-SUPERSEDED: reversible `000018_grounded_report_context` included durable attempts 1..4 and crash/replay cap semantics. <!-- verified: 2026-07-13 method=go-test evidence="historical contract only; superseded by Phase 9" -->
- [x] 8.3 Run clean/populated Postgres up/down/up, current-invalid-context, column rename/down restoration and privacy deletion/non-content leakage probes.
  <!-- verified: 2026-07-12 method=two-disposable-postgres-paths evidence="Clean disposable DB make migrate-check completes up/down/up at version=18 dirty=false. Separate populated DB integration emits REPORT_STORAGE_V18_POPULATED_MIGRATION_PASS, REPORT_STORAGE_V18_INVALID_CONTEXT_PASS, REPORT_STORAGE_V18_RENAME_ROLLBACK_PASS and REPORT_STORAGE_V18_PRIVACY_PROBE_PASS; focus values survive both renames, empty context remains invalid, user deletion removes report content, audit/job/outbox contain zero sentinel hits, final table count is exactly 26, and all temporary DBs are dropped." -->
- [x] 8.4 HISTORICAL-SUPERSEDED: migration/storage gates included durable attempt probes before emitting `REPORT_STORAGE_V18_PASS`. <!-- verified: 2026-07-13 evidence="historical marker contract only; Phase 9 must re-emit against current shape" -->

## Phase 9: remove durable product retry storage

- [x] 9.1 RED-GREEN: migration lint/SQL tests require summary, generation_context, report+plan dimension-focus and exact current inventory while rejecting `llm_attempt_count` and all synonymous product retry columns.
  <!-- verified: 2026-07-13 method=lint+sql-contract evidence="scenario+migration lint 31 PASS; SQL contract rejects llm_attempt_count" -->
- [x] 9.2 Reconcile reversible `000018_grounded_report_context` in place: remove attempt-column up/down DDL, preserve empty-invalid legacy context and focus rename rollback, and add no compatibility mirror.
  <!-- verified: 2026-07-13 method=postgres-integration evidence="disposable empty PostgreSQL v18 up/down/up PASS with no llm_attempt_count" -->
- [x] 9.3 Run clean/populated PostgreSQL up/down/up, current-invalid-context, rename/down restoration and privacy/non-content leakage probes against the no-retry-column shape.
  <!-- verified: 2026-07-13 method=postgres-integration evidence="disposable and dev PostgreSQL completion/storage/privacy probes PASS; REPORT_STORAGE_V18_PASS re-emitted" -->
- [x] 9.4 Run `make migrate-check`, migration lint, backend migration tests, C-13 schema/privacy probes and `git diff --check`; only then re-emit owner-only `REPORT_STORAGE_V18_PASS` and return plan to completed.
  <!-- verified: 2026-07-13 method=full-migration-gate evidence="disposable PostgreSQL make migrate-check PASS; migration tests/lint and git diff --check PASS" -->

## Phase 10: TargetJob paste-only schema net-state

- [ ] 10.1 RED: migration lint/SQL contracts/inventory tests 要求 20+3+2，并断言旧 TargetJob 来源列/表、JD attachment purpose 与 JD source refresh jobType 不存在；记录当前失败证据。
- [ ] 10.2 GREEN: 原地修订 baseline up/down、enum sources、privacy matrix 与 SQL contracts；删除旧结构，保留 `raw_jd_text`、独立 `source_records`、resume/privacy purpose，不创建兼容层。
- [ ] 10.3 BDD-Gate: 不适用；替代 gate 运行 migration contract、enum-source lint、privacy dry-run、focused migration tests 与 clean/populated PostgreSQL up/down/up。
- [ ] 10.4 Zero-ref: migrations/enum sources/backend migration probes 中旧结构精确零命中；正向 probe 证明 `raw_jd_text`、`source_records`、resume/privacy purpose 与 20+3+2 inventory。

## Phase 11: Practice reply-status net-state

- [ ] 11.1 RED: SQL/enum/store contracts fail until `practice_messages.reply_status` exists with the exact user-only four-state allowlist and populated-row expectations.
- [ ] 11.2 GREEN: revise baseline up/down, enum source and role CHECK so user rows carry `pending|retryable_failed|terminal_failed|complete`, assistant rows carry NULL, and original client/reply uniqueness remains intact.
- [ ] 11.3 REGRESSION-GATE: migration/store tests cover pending→retryable/terminal→pending retry→complete, completed replay, illegal role/status pairs, duplicate reply and cross-session client IDs.
- [ ] 11.4 BDD-Gate: not applicable; run migration lint, clean/populated PostgreSQL up/down/up, privacy cascade and backend-practice/002 composed persistence gates.
