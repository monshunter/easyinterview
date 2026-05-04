# DB Migrations Baseline Bootstrap

> **版本**: 1.4
> **状态**: completed
> **更新日期**: 2026-05-03

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)

## 1 目标

把 [db-migrations-baseline spec](../../spec.md) v1.8 锁定的迁移工具、29 张应用 / auth 支撑表、2 张迁移元数据表、B3 outbox retry operational columns、A3/F1 `ai_task_runs` typed columns、enum/check 来源矩阵、backfill ledger 与 P0 privacy deletion matrix 落到 `migrations/` 与 `backend/cmd/migrate`。

本 plan 不实现业务 repository、不实现 C8 dispatcher、不实现 privacy worker；只提供 schema baseline、迁移可执行入口、lint/check gate 与下游 handoff。

## 2 背景

B4 是 Layer B contract 的 schema owner。A2 已提供 Postgres 16 + pgvector 本地实例；B1/B2/B3 分别提供 shared enum、API-facing async enum、event/job manifest；A3/F1 需要 `ai_task_runs` typed columns；C8 需要 `outbox_events` retry columns 与 privacy deletion matrix。B4 001 必须在后续 C/D 域 implementation 读取真实表之前完成。

## 3 质量门禁分类

- **Plan 类型**: `migration + tooling + code-internal + contract`。本 plan 修改 SQL migrations、Go migration wrapper、backfill registry、lint/probe 脚本与 Make target，属于内部 schema contract 和迁移工具链交付；不引入用户可感知 UI、HTTP API 行为、业务流程或端到端产品路径。
- **TDD 策略**: 必须通过 `/tdd --file docs/spec/db-migrations-baseline/plans/001-bootstrap/checklist.md --references docs/spec/db-migrations-baseline/plans/001-bootstrap/plan.md,docs/spec/db-migrations-baseline/spec.md --phase-commit db-migrations-baseline/001-bootstrap` 顺序执行。每个 checklist item 以本 checklist 内的 `验证:` 子句作为 Red-Green-Refactor 断言来源；涉及 migration wrapper、lint、backfill registry、privacy dry-run 或 SQL probe 的 item 必须先补 focused failing test / smoke / probe，再最小实现并复跑对应命令。
- **BDD 策略**: BDD 不适用。本 plan 只交付 DB schema baseline、migration wrapper、lint/probe 与内部 dry-run 工具，不产生浏览器 UI、外部 API、用户业务流程或 scenario-test 可观察行为，因此不创建 `bdd-plan.md` / `bdd-checklist.md`，主 checklist 也不设置 `BDD-Gate:`。
- **替代验证 gate**: 使用内部迁移与契约 gate 代替 BDD：Go wrapper/backfill 单元测试、`make migrate-check`、`APP_ENV=prod make migrate-down` fail-fast smoke、migration enum/check lint 与 drift negative case、SQL table/column/index/explain probes、privacy matrix dry-run 覆盖检查、`python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check`。

## 4 实施步骤

### Phase 1: migration skeleton 与工具 wrapper

#### 1.1 `golang-migrate` wrapper

落地 `backend/cmd/migrate/main.go`，包装 `golang-migrate v4.18+` 与 B4 backfill registry；根 `Makefile` 提供 `make migrate-up` / `make migrate-down` / `make migrate-status` / `make migrate-create NAME=...` / `make migrate-check`，并让 A1 占位 `make migrate` 指向 `migrate-up` 或清晰 help。

#### 1.2 migration 目录与命名

在 `migrations/` 使用 `NNNNNN_<verb>_<noun>.up.sql` / `.down.sql`，编号从 `000001` 起 6 位连续；所有新增 migration 必须由 `make migrate-create NAME=...` 生成，不手敲编号。

#### 1.3 pgvector 与 schema_backfills

`000001` baseline 必须 `CREATE EXTENSION IF NOT EXISTS vector`；down 默认不 drop extension，只有 dev + `MIGRATE_DROP_EXTENSIONS=1` 允许 drop。落地 `schema_backfills` ledger 表与 `migrations/backfill/manifest.yaml`，并提供 1 个 dry-run/apply 示例 registry。

### Phase 2: baseline DDL 与索引

#### 2.1 29 张当前应用 / auth 支撑表

落地当前产品范围内的 26 张应用表，加 ADR-Q1 的 `auth_challenges` / `sessions` / `external_identities` 3 张支撑表；旧 `mistake_entries` 不再创建。`make migrate-up` 后 public schema 至少 31 张表（含 `schema_migrations` / `schema_backfills`）。

#### 2.2 B3 outbox / async columns

`outbox_events` 必须包含 `publish_attempts` / `next_attempt_at` / `locked_at` / `last_error_code` / `last_error_message`；pending due 查询索引至少覆盖 `(publish_status, next_attempt_at, created_at)`。`async_jobs.job_type` check 必须包含 B3 10 个 canonical jobType（含 internal-only `email_dispatch`），但 API-facing subset 仍只由 B2 7 项暴露。

#### 2.3 A3/F1 AI typed columns

`ai_task_runs` 必须包含 `model_family` / `model_profile_name` / `model_profile_version` / `fallback_chain jsonb not null default '[]'::jsonb` / `route` / `validation_status` / `output_schema_version`；dashboard 核心查询不得依赖 JSONB path scan。

#### 2.4 索引覆盖

覆盖 03 §7 B-Tree 索引、`retrieval_chunks.embedding` ivfflat（默认 lists=100）与可选 `target_jobs` GIN 全文索引；用 SQL probe / explain 验证关键查询走索引。

### Phase 3: enum/check 来源、backfill 与 privacy lint

#### 3.1 enum/check 来源 lint

落地 `migrations/enum-sources.yaml` 与 `scripts/lint/migrations_lint.py`（或等价 Go 工具），逐列登记 `table.column -> source -> checksum`。SQL 中出现未登记 `check (col in (...))` 必须 fail；B1/B2/B3 修改 manifest 后 `make migrate-check` 必须发现漂移。

#### 3.2 backfill registry

`backend/internal/migrations/backfills/<version>/` 注册 dry-run/apply 函数；同一 `version + mode + checksum` apply 成功后不得重复执行，除非 `--force` 且 `APP_ENV!=prod`。

#### 3.3 privacy deletion dry-run

提供 `make privacy-delete-dry-run` 或 `backend/cmd/migrate privacy-matrix --dry-run` 等入口，读取 spec §3.1.2 table matrix，对测试用户输出每表 disposition（hard delete / cascade / retain / audit tombstone），不执行真实删除；C8 后续 plan 消费该矩阵。

### Phase 4: Verification + handoff

#### 4.1 migrate-check

在干净 Postgres 16 上执行 `make migrate-check`：`migrate-up -> migrate-down -> migrate-up` 全部成功；`APP_ENV=prod make migrate-down` 必须 fail-fast；`schema_backfills` ledger 无重复成功记录。

#### 4.2 table / column / index probes

运行 SQL probes：public table count ≥31；`outbox_events` retry columns 存在；`ai_task_runs` typed columns 存在；ivfflat / pending due / B-Tree 索引存在且 explain 命中。

#### 4.3 enum / privacy probes

临时修改 B3 job manifest 或 B1 enum，确认 `make migrate-check` / lint 报 drift；privacy deletion dry-run 输出覆盖 spec §3.1.2 所有表组，且 `prompt_versions` / `rubric_versions` / migration metadata 被 retain。

#### 4.4 文档与 INDEX

本 plan checklist 全部勾选后，将 Header 切 completed，运行 sync-doc-index check/fix，并在 work journal 记录 migrate-check、SQL probes、lint probes 与 downstream handoff。

### Phase 5: product-scope v1.2 schema remediation

#### 5.1 Red: migration inventory 期望排除 `mistake_entries`

先调整 migration contract test / lint 的期望：public schema count gate 32→31，table inventory 不再允许 `mistake_entries`，B1 enum source 不再登记 `MistakeStatus`。旧 SQL / privacy matrix / generated probes 仍引用 `mistake_entries` 时必须失败。

#### 5.2 Green: 修订 SQL、enum source 与 privacy matrix

修订 baseline migration：删除 `mistake_entries` 表及其索引 / FK / privacy disposition；`target_jobs.open_mistake_count` 改为 `open_question_issue_count`；`question_assessments.written_to_mistake_book` 改为 `included_in_retry_plan`；`practice_plans` 的 mode / goal check 跟随 B1 v1.7；如需要记录报告内题目状态，使用 `question_assessments.review_status` 引用 `QuestionReviewStatus`。

#### 5.3 Verify

运行 `make migrate-check` 或可用的 migration lint / SQL contract tests；repo 搜索确认实现侧不再出现 `mistake_entries`、`open_mistake_count`、`written_to_mistake_book`、旧 practice mode / goal check 值。

## 5 验收标准

- spec §6 C-1..C-13 全部具备本 plan 或下游 handoff 证据；C8/F1/C11 等运行时验证由各自 owner 后续关闭。
- `make migrate-check` 可在干净本地 DB 重复执行；prod down 防呆有效。
- enum/check 来源、B3 jobType manifest、A3/F1 AI typed columns、P0 privacy deletion matrix 都有可执行 lint/probe。

## 6 风险与应对

| 风险 | 应对措施 |
|------|----------|
| baseline DDL 过大难以 review | 分 phase 编排，保持 migration 文件编号稳定；用 table/column/index inventory probe 辅助 review |
| internal-only jobType 误暴露到 B2 API | Phase 3 enum/source lint 对比 B2 API-facing 7 项与 B3 canonical 10 项，`email_dispatch` 只能进入 DB/C8 check |
| migration down 误在 prod 执行 | wrapper 在 `APP_ENV=prod` 时拒绝 down，除非显式 `MIGRATE_DOWN_FORCE=1` 且执行环境允许 |
| backfill 重复 apply | `schema_backfills` 以 version/mode/checksum 做幂等 ledger；重复执行必须 fail 或 skip |
| privacy deletion matrix 漏表 | Phase 3.3 dry-run 输出必须覆盖所有 baseline 表组；新增表时 lint 要求同步 matrix |

## 7 修订记录

| 日期 | 版本 | 变更 | 关联 |
|------|------|------|------|
| 2026-05-03 | 1.4 | 修正 Phase 2 / Phase 4 中遗留的旧表数量口径：当前 baseline 为 26 应用表 + 3 auth 支撑表 + 2 迁移元数据表，public schema count gate ≥31。 | readiness reconcile |
| 2026-05-03 | 1.3 | 原地 reopen，新增 Phase 5 remediation：按 product-scope v1.2 删除独立 `mistake_entries` 表，迁移字段改为报告题目回顾 / 本轮复练语义。 | db-migrations-baseline v1.6 |
| 2026-04-30 | 1.2 | 原地 reopen 001-bootstrap，修复 L2 code-review 发现的 prod down fail-fast 顺序、dev-only vector extension drop 限制、B1/B2/B3 enum source drift gate 与 ALTER TABLE check 发现能力。 | plan-code-review remediation |
| 2026-04-30 | 1.1 | 补齐 TDD/BDD 质量门禁分类与 checklist 可执行验证断言；确认 BDD 不适用并以 migration / lint / probe / smoke gate 替代。 | implement gate remediation |
| 2026-04-29 | 1.0 | 初始物化 B4 `001-bootstrap`：migration wrapper、baseline DDL、enum/backfill/privacy lint 与 verification handoff。 | plan-review remediation |
