# DB Migrations Baseline Spec

> **版本**: 1.2
> **状态**: active
> **更新日期**: 2026-04-27

## 1 背景与目标

[engineering-roadmap spec §5.2](../engineering-roadmap/spec.md#52-layer-b--contract4-份全部-p0) 把 B4 `db-migrations-baseline` 列为 Layer B · Contract 第四份 child（依赖 [B1 `shared-conventions-codified`](../shared-conventions-codified/spec.md) 与 [A2 `local-dev-stack`](./../local-dev-stack/spec.md)）。它把 [03-db-definition.md §5](../../../easyinterview-tech-docs/03-db-definition.md) 的 27 张 P0 应用表 + 索引 + pgvector 扩展落到迁移文件层，决定了：

- 后端任何 C 域 spec 在自己的 plan 里能够 import 真实表；
- 各 C 域不得在自己的 plan 中另起重复的 migration 文件夹或选型；
- 迁移文件命名 / 工具 / 回滚策略统一。

A2 已锁定本地 Postgres 16 + pgvector 实例的可用性；本 spec 在此基础上落地 schema。

目标是：

1. **27 张 P0 应用表 + schema_migrations 元数据表**：覆盖 03 §4 全集，外加最小迁移系统自身需要的元数据表，达到 [engineering-roadmap §5.2](../engineering-roadmap/spec.md#52-layer-b--contract4-份全部-p0) 估算的「29 表」量级（精确数以 03 §4 P0 范围 27 张 + schema_migrations 1 张为底线，第 29 张由 B4 自身 plan 在落地时决策是否新增 admin / view 表）。
2. **统一迁移工具**：`golang-migrate` 或 `atlas` 之一；本 spec §3.1 选型 `golang-migrate`（轻量、社区成熟），在 B4 自身 plan 落地时如发现限制可在本 spec 修订中切换。
3. **可逆 + 数据回填策略**：所有 migration 必须说明是否可逆 / 是否需数据回填 / 是否阻塞写入；与 [03 §9](../../../easyinterview-tech-docs/03-db-definition.md#9-迁移策略) 一致。
4. **索引策略锁定**：覆盖 03 §7 必要 B-Tree + ivfflat 向量索引 + 可选 GIN 全文索引；性能查询路径在 W2/W3 业务域加表时不得遗漏索引。

本 spec 不实现具体业务表的 CRUD（归各 C 域）、不部署 DB 服务（归 A2）、不实现 Admin 面板（不在 P0 范围）。

## 2 范围

### 2.1 In Scope

- **迁移工具与目录**：`migrations/` 目录（A1 已锁定根容器）；工具 `golang-migrate`（CLI）；文件命名 `NNNNNN_<verb>_<noun>.up.sql` / `NNNNNN_<verb>_<noun>.down.sql`，序号从 `000001` 起 6 位。
- **27 张 P0 应用表**：与 [03 §4](../../../easyinterview-tech-docs/03-db-definition.md#4-表清单) 完全一致：
  1. `users` / 2. `user_settings` / 3. `candidate_profiles` / 4. `experience_cards` / 5. `file_objects` / 6. `resume_assets` / 7. `target_jobs` / 8. `target_job_requirements` / 9. `target_job_sources` / 10. `practice_plans` / 11. `practice_sessions` / 12. `practice_session_events` / 13. `practice_turns` / 14. `question_assessments` / 15. `feedback_reports` / 16. `mistake_entries` / 17. `resume_tailor_runs` / 18. `debriefs` / 19. `source_records` / 20. `retrieval_chunks` / 21. `prompt_versions` / 22. `rubric_versions` / 23. `ai_task_runs` / 24. `async_jobs` / 25. `outbox_events` / 26. `privacy_requests` / 27. `audit_events`。
- **元数据表**：`schema_migrations`（由 `golang-migrate` 自动创建管理）。
- **扩展**：`pgvector`（与 [A2 D-5](../local-dev-stack/spec.md#31-已锁定决策) 协同；A2 已在本地 dev 启用，B4 在 staging / prod 也通过迁移启用）。
- **索引**：覆盖 03 §7 列出的全部 B-Tree 索引 + `retrieval_chunks.embedding` 上的 `ivfflat`（默认 lists=100）+ 可选 `target_jobs` GIN 全文索引（dev 默认启用，prod 由 C4 在自己 spec 决策是否启用）。
- **Make target**：A1 已占位 `make migrate`，本 spec 落地：`make migrate-up` / `make migrate-down` / `make migrate-status` / `make migrate-create NAME=...`。
- **本地迁移校验**：本地 gate 跑 `make migrate-up && make migrate-down && make migrate-up` 验证 idempotency 与可逆性（`migrate-down` 仅在 dev 模式下生效，prod 模式禁止）；远端 CI 仅在 A5 触发条件成立后再接入。
- **数据回填脚本**：`migrations/backfill/<NNNNNN>.go`（数据回填用 Go 脚本而非 SQL 单段），由 `golang-migrate` driver 调用；B4 提供 `cmd/migrate/main.go` 可执行入口。
- **field-level lint**：与 [B1 D-6 枚举](../shared-conventions-codified/spec.md#31-已锁定决策) 同源 —— 任何 `text + check (col in (...))` 的枚举值列表必须由 B1 generator 生成，不允许手写。

### 2.2 Out of Scope

- 业务表的 CRUD / repository 层：归各 C 域。
- Admin / 后台运营表：当前 P0 不在范围。
- workspace / multi-tenant 扩展：归 P1+；本 spec 仅在表注释中预留扩展空间。
- 数据 ETL / data warehouse：归 [F2 `analytics-funnel`](../engineering-roadmap/spec.md#56-layer-f--quality-横切4-份)（PostHog 自托管路径）。
- 全量备份 / DR：归 [E4 `release-gate-and-rollout`](../engineering-roadmap/spec.md#55-layer-e--integration4-份) + 运维。

## 3 用户决策 / 待确认事项

### 3.1 已锁定决策

| ID | 决策 | 锁定值 | 影响 |
|----|------|--------|------|
| D-1 | 迁移工具 | `golang-migrate v4.18+`（轻量、CLI + Go API、社区成熟） | 不与 atlas / goose 共存；后续切换需本 spec 修订 |
| D-2 | 文件命名 | `NNNNNN_<verb>_<noun>.{up,down}.sql`；序号 6 位连续；编号一旦发布不可回收 | 防止跨 PR 同号冲突 |
| D-3 | 数据回填策略 | DDL 走 SQL；行级数据回填走 Go 脚本（位于 `migrations/backfill/<NNNNNN>.go`）；同一编号的 SQL + Go 必须配对 | 复杂回填可写 unit test |
| D-4 | 可逆要求 | 每条 migration 必须有 `.down.sql`；不可逆操作（如删列）的 down 至少恢复表结构骨架（数据可记录在 backfill log） | – |
| D-5 | 阻塞操作约束 | 不在高频表（`practice_session_events` / `outbox_events` / `audit_events`）执行阻塞式大表重写；优先 additive；若必须重写需先在本 spec 中登记 | 与 [03 §9.1](../../../easyinterview-tech-docs/03-db-definition.md#91-原则) 一致 |
| D-6 | enum / check 约束源头 | 所有 `text + check (col in ('a','b'))` 的枚举值集合必须由 B1 generator 输出 SQL 片段；手写违规 lint 拦截 | 与 [B1 D-6](../shared-conventions-codified/spec.md#31-已锁定决策) 同源 |
| D-7 | 索引覆盖 | 03 §7 全部 B-Tree 索引 + `ivfflat` + 可选 GIN 全文（按 D-9 决策） | 性能 baseline |
| D-8 | seed 数据 | 不在 baseline migration 中插入业务数据；business seed 由各 C 域 mock-server plan 提供 | 防止把 demo 数据当默认 |
| D-9 | 全文索引开关 | `target_jobs` GIN 全文索引在 dev 默认启用，staging / prod 由 C4 spec 决策是否启用（迁移条件化） | – |
| D-10 | migrate-down 防呆 | `make migrate-down` 在 `APP_ENV=prod` 时拒绝执行；CI 中通过 `MIGRATE_DOWN_FORCE=1` 显式开启 | 防止误回滚 |
| D-11 | 多 module 模式 | 当前 P0 不引入 `go.work`（与 [A1 §3.2](../repo-scaffold/spec.md#32-待确认事项) 默认一致）；`cmd/migrate/main.go` 与 backend module 同 module | – |

### 3.2 待确认事项

- 是否在 W2/W3 业务域加表时引入 `pg_partman` / 时间分区（针对 `audit_events` / `practice_session_events` 高频表）：默认 P0 不分区；如 monthly 表行数 > 10M 再决策。
- pgvector 索引 `lists` 参数：默认 100；C11 `backend-retrieval` spec 在 W3 时按数据量调整。
- `target_jobs` 全文索引语言（`simple` vs `english`）：默认 `simple`（多语言兼容），由 C4 在自己 spec 中决策。

## 4 设计约束

### 4.1 命名与目录约束

- 所有迁移文件落 `migrations/`（A1 锁定的根目录）；不允许子目录（`migrations/<domain>/...`）—— `golang-migrate` 默认按文件名扁平排序。
- 序号必须严格递增；新增迁移由 `make migrate-create NAME=add_xxx_to_yyy` 自动生成，不允许手敲。
- 表名 / 列名严格 `snake_case`；与 [03 §2.1](../../../easyinterview-tech-docs/03-db-definition.md#21-基本约定) 一致；外键命名 `fk_<from_table>_<column>` / `idx_<table>_<purpose>` / `uq_<table>_<columns>`。

### 4.2 schema 约束

- 主键统一 `uuid` + 应用层生成 UUIDv7（由 [B1 idx 工具](../shared-conventions-codified/spec.md#21-in-scope) 提供）；migration 不设 `default gen_random_uuid()`。
- 时间字段统一 `timestamptz`；默认 `now()`；`updated_at` 由应用层维护（不引 trigger）。
- 软删字段 `deleted_at timestamptz` 在用户核心资源表上必须存在；查询默认 filter `deleted_at IS NULL` 由各 C 域 repo 层负责。
- jsonb 字段命名以 `_payload` / `_metadata` / `_summary` / `_results` 等明确语义；不允许 `data jsonb`。

### 4.3 性能约束

- `migrate-up` 在干净 DB 上必须 ≤ 60s 完成（dev 机器）；超出由 owner 拆分迁移文件。
- 单条 migration 执行时间 ≤ 5s（CI 模式）；如必须长时间运行（如 `CREATE INDEX CONCURRENTLY`）需在 PR 描述中标注「需要 ops 窗口」，并提供 staged migration 路径。
- 任何 migration 不得 `LOCK TABLE` 在生产高频表上 > 100ms。

### 4.4 安全约束

- migration 文件不得包含明文 secret / API key。
- 任何敏感字段（`email` / `auth_provider_user_id` / `ip_hash` / `user_agent_hash`）的索引必须考虑 PII 暴露：默认不索引明文 email（已通过 `email unique` 实现），其它 hash 字段索引可接受。
- B4 提供 `migrations/lint.sh`（可选）扫描 `password` / `secret` / `token` 关键字。

## 5 模块边界

| 边界 | Owner | 说明 |
|------|-------|------|
| `migrations/` 目录与 baseline 文件 | B4 | 27 P0 表 + 索引 + pgvector |
| `cmd/migrate/main.go` 与 `make migrate-*` target | B4 | A1 占位 `make migrate` 的真实实现 |
| Postgres 实例与 pgvector 启用（dev） | A2 | 本 spec 仅 `CREATE EXTENSION IF NOT EXISTS vector` |
| 业务表 CRUD / repository | 各 C 域 | 通过 generated DTO（B2）+ raw SQL / sqlc |
| Schema enum / check 值 | B1 + B4 | B1 generator 生成 SQL 片段；B4 在 migration 中 include |
| 后续表新增 / 修改 | 各 C 域 + B4 review | 各 C 域提交 PR 时由 B4 owner approve schema 变更 |
| Workspace / multi-tenant 扩展 | P1+ | 本 spec 仅预留扩展空间 |
| 备份 / DR / 跨区复制 | E4 + 运维 | 不在本 spec 范围 |

## 6 验收标准

| ID | 场景 | Given | When | Then | 对应 Plan |
|----|------|-------|------|------|-----------|
| C-1 | 干净 DB baseline 落地 | 干净 Postgres 16（A2 dev stack 启动） | `make migrate-up` | 27 张表 + `schema_migrations` 元数据表全部存在；`pgvector` 扩展已启用；`select count(*) from information_schema.tables where table_schema='public'` ≥ 28 | B4 后续 001 |
| C-2 | 索引覆盖 | C-1 完成 | `select indexname from pg_indexes where schemaname='public'` | 03 §7 列出的全部 B-Tree 索引 + `idx_retrieval_chunks_embedding`（ivfflat） + 可选 `idx_target_jobs_fts`（GIN）存在 | B4 后续 001 |
| C-3 | 迁移可逆 | C-1 完成 | `make migrate-down` （dev） | 全部表 + 扩展回滚到空；exit 0 | B4 后续 001 |
| C-4 | idempotent | 已 `migrate-up` 一次 | 再次 `make migrate-up` | exit 0；`schema_migrations` 元数据无重复 | B4 后续 001 |
| C-5 | enum 与 B1 同源 | B1 modify enum 增加新值 | `make codegen-conventions && make migrate-up`（在干净 DB）+ 本地 drift check | 新增枚举值出现在对应 `text + check (col in (...))` 中；本地 drift 通过 | B4 后续 001 + B1 |
| C-6 | prod 安全 | `APP_ENV=prod` | `make migrate-down` | exit 非 0；stderr 提示需 `MIGRATE_DOWN_FORCE=1` | B4 后续 001 |
| C-7 | 数据回填 | 某 migration 需要回填 `users.display_name` | up SQL + Go backfill 联动 | up SQL 加列；backfill Go 在本机 dry-run 通过；远端 CI 执行仅在 A5 触发条件成立后再接入 | B4 后续 001 + 实际场景 |
| C-8 | outbox + async_jobs 索引性能 | `outbox_events` 1 万行积压 | `select * from outbox_events where publish_status='pending' order by created_at limit 100` | 走索引；P95 < 5ms | B4 后续 001 + C8 |
| C-9 | retrieval_chunks 向量索引 | 1 万行 embedding 数据 | top-K 查询 | ivfflat 命中；P95 < 50ms（lists=100，ef=10） | B4 后续 001 + C11 |
| C-10 | 迁移完整闭环 | 本地修改 schema | `make migrate-check` 或等价本地 gate | `migrate-up && migrate-down && migrate-up` 全部成功；diff 检查通过；远端 CI 仅在 A5 触发条件成立后再接入 | B4 后续 001 |

## 7 关联计划

B4 在本次 W1 spec 阶段不创建 impl plan（参见 [001-decompose-subspecs §3.1](../engineering-roadmap/plans/001-decompose-subspecs/plan.md#3-实施步骤)）。后续由 B4 自身的 `001-bootstrap`（W1 末或 W2 初）承接：

- 落地 `migrations/000001_*.up.sql` ~ `000028_*.up.sql`（每张 P0 表 + 扩展 + 索引一份 migration），与 03 §5 DDL 对齐。
- 落地 `cmd/migrate/main.go`、`make migrate-*` target，替换 A1 占位。
- 落地 `migrations/backfill/` 框架与 1 个最小示例。
- 本地 `make migrate-check` 完整闭环（C-10）；远端 CI 仅在 A5 触发条件成立后再评估。

后续业务域加表 / 改 schema：各 C 域在自己 plan 中提交 migration 文件，schema 变更必须由 B4 owner 复查；本 spec 不需逐次修订（仅在跨域 schema 决策变更时修订，如分区 / multi-tenant 升格）。
