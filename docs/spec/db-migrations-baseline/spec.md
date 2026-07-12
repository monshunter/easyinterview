# DB Migrations Baseline Spec

> **版本**: 1.29
> **状态**: active
> **更新日期**: 2026-07-12

## 1 背景与目标

`db-migrations-baseline` 是当前数据库 schema owner。当前可执行真理源由本 spec、`migrations/`、`migrations/enum-sources.yaml`、`backend/cmd/migrate` 和 `make migrate-*` targets 共同组成。

本 spec 只描述当前 net-state：

- 22 张当前应用表；
- 3 张 auth / session 支撑表；
- 2 张迁移元数据表；
- 当前索引、check constraint、backfill ledger、privacy deletion matrix 与迁移执行 gate。

目标是：

1. **稳定 schema inventory**：干净 DB 迁移完成后，public schema 至少包含 27 张表，且当前应用 / auth / 元数据表清单与 §2.1 完全一致。
2. **统一迁移工具**：迁移入口使用 `golang-migrate` 包装器，不并行维护第二套迁移工具。
3. **可逆与可审计**：每个 migration 都有 `.down.sql`；行级 backfill 通过 Go registry 记录 dry-run / apply ledger。
4. **索引与约束可验证**：B-Tree、可选 GIN、enum/check source 与 migration lint 必须能被本地 gate 验证。
5. **隐私删除可执行**：用户关联表必须在 §3.1.2 中有明确 disposition，backend internal runner 按矩阵执行。

## 2 范围

### 2.1 In Scope

- **迁移目录与命名**：所有 SQL migration 位于 `migrations/`，文件名为 `NNNNNN_<verb>_<noun>.up.sql` / `NNNNNN_<verb>_<noun>.down.sql`，序号 6 位递增。
- **21 张当前应用表**：
  1. `users`
  2. `user_settings`
  3. `file_objects`
  4. `resumes`
  5. `target_jobs`
  6. `target_job_requirements`
  7. `target_job_sources`
  8. `practice_plans`
  9. `idempotency_records`
  10. `practice_sessions`
  11. `practice_session_events`
  12. `practice_messages`
  13. `feedback_reports`
  14. `source_records`
  15. `prompt_versions`
  16. `rubric_versions`
  17. `ai_task_runs`
  18. `async_jobs`
  19. `outbox_events`
  20. `privacy_requests`
  21. `audit_events`
- **Flat Resume schema**：`resumes` 承载 `original_text`、`parsed_text_snapshot`、`raw_text`、`file_object_id`、`structured_profile`、`display_name` 与 `source_type IN ('upload', 'paste')`；`practice_plans.resume_id` 是 practice 绑定简历的当前 FK。
- **3 张 auth / session 支撑表**：`auth_challenges`、`sessions`、`external_identities`，遵守 [ADR-Q1](../engineering-roadmap/decisions/ADR-Q1-auth.md) 与 [backend-auth](../backend-auth/spec.md) 的 token / session / identity 约束。
- **迁移元数据表**：`schema_migrations` 由迁移工具管理；`schema_backfills` 由 B4 Go registry 管理。
- **Outbox operational columns**：`outbox_events` 必须包含 retry、lock、last-error 与 due-query index 所需字段。
- **AI call meta typed columns**：`ai_task_runs` 必须包含 model profile、fallback、route、validation、output schema 与 prompt/rubric provenance typed columns，核心查询不得依赖 JSONB path scan。
- **索引 inventory**：当前 B-Tree index 与 `target_jobs` 可选 GIN 全文索引由 migration contract tests 和 lint gate 验证。
- **Make targets**：`make migrate-up`、`make migrate-down`、`make migrate-status`、`make migrate-create NAME=...`、`make migrate-check`。
- **本地迁移 gate**：`make migrate-check` 执行 up -> down -> up；`migrate-down` 在 prod 环境拒绝执行，除非显式 force。
- **Backfill registry**：真实行级 backfill 通过可选 `migrations/backfill/manifest.yaml` 与 `backend/internal/migrations/backfills/` 共同登记；当前没有已登记的行级 backfill，runner 在 manifest 缺失时直接跳过，仍支持 dry-run / apply / ledger。
- **Enum/check source lint**：所有 `text + check (col in (...))` 必须能由 §3.1.1 的 owner source 或 `migrations/enum-sources.yaml` 解释。
- **Privacy deletion matrix**：§3.1.2 是 `privacy_delete` backend internal runner 的表级真理源。

### 2.2 Out of Scope

- 业务表 CRUD / repository 行为：归各 backend subject。
- Admin / 运营后台表：不在当前 P0 schema 范围。
- 数据仓库、全量备份、跨区复制：归分析和发布运维 owner。
- 新 DB extension：当前 baseline 不启用向量或分区扩展；引入前必须先修订本 spec 与 migration gate。

## 3 用户决策 / 待确认事项

### 3.1 已锁定决策

| ID | 决策 | 锁定值 | 影响 |
|----|------|--------|------|
| D-1 | 迁移工具 | `golang-migrate v4.18+` + `backend/cmd/migrate` wrapper | 不与 atlas / goose 共存 |
| D-2 | 文件命名 | `NNNNNN_<verb>_<noun>.{up,down}.sql`；编号发布后不复用 | 防止跨 PR 同号冲突 |
| D-3 | 数据回填 | DDL 走 SQL；行级 backfill 走 Go registry；manifest 记录 reversible / dry-run | 支持单测、ledger 与幂等执行 |
| D-4 | 可逆要求 | 每条 migration 必须有 `.down.sql`；不可逆数据操作至少恢复结构骨架并记录 ledger | 支撑 dev rollback 和 migration-check |
| D-5 | 阻塞操作约束 | 高频表避免阻塞式大表重写；必须重写时先登记 owner 决策 | 控制 schema 变更风险 |
| D-6 | Enum/check source | 按 §3.1.1 source matrix 或 `migrations/enum-sources.yaml` 解释 | 防止 SQL 裸写无 owner 的枚举 |
| D-7 | 索引覆盖 | 当前应用表所需 B-Tree + 可选 `target_jobs` GIN | 性能 baseline |
| D-8 | Seed 数据 | Baseline migration 不插入业务 seed | 防止 demo 数据成为默认数据 |
| D-9 | 全文索引 | `target_jobs` GIN 在 dev 默认启用；其他环境由 deployment owner 决定 | 多语言检索保持可控 |
| D-10 | `migrate-down` 防呆 | `APP_ENV=prod` 拒绝 down；CI/dev 需显式 force | 防止误回滚 |
| D-11 | Go module 模式 | 使用根 `go.work` + `backend/go.mod`；不为 `migrations/` 另起 module | 与仓库拓扑一致 |
| D-12 | DB extension 生命周期 | 当前 baseline 不创建向量扩展；down 不管理 extension | 避免无用基础设施负担 |
| D-13 | Backfill ledger | `schema_backfills(version, name, mode, status, checksum, started_at, completed_at, error_message)` | 可审计、幂等、可 dry-run |
| D-14 | Outbox retry 字段 | `publish_attempts`、`next_attempt_at`、`locked_at`、`last_error_code`、`last_error_message` 与 due-query index | 支撑 dispatcher retry 和排查 |
| D-15 | AI call meta columns | `ai_task_runs` 包含 model / route / validation / schema / provenance typed columns | 支撑 AI routing、report 与观测查询 |
| D-16 | Privacy deletion matrix | §3.1.2 是 table disposition 真理源；新增用户关联列必须先更新矩阵 | 防止漏删或误删 |
| D-17 | Flat Resume net-state | `resumes` 是当前简历表；`source_type` 为 `upload` / `paste`；`practice_plans.resume_id` 绑定简历 | 支撑 Resume Workshop、Practice 与 privacy delete |
| D-18 | Practice message replay | `practice_messages.client_message_id` 在 session 内唯一；assistant `reply_to_message_id` 唯一 | 同一用户消息重试不重复落库或生成 reply |
| D-19 | Report generation columns | `feedback_reports.retry_focus_competency_codes` 与 `ai_task_runs` 承载 conversation-level report language / retry / provenance | 支撑 async report generation，不保留 question assessment 表 |
| D-20 | Privacy request tombstone | `privacy_requests.user_id` 可置空，FK 为 `ON DELETE SET NULL` | 用户行 hard delete 后保留最小删除证据 |
| D-21 | Current public schema count | 当前 public schema gate 为 21 app + 3 auth + 2 metadata，count >= 26 | 作为 migration inventory drift gate |
| D-22 | Practice conversation schema | 删除 `practice_turns`、`question_assessments`、`practice_plans.question_budget/mode`、`practice_sessions.turn_count/hints_enabled`；新增 `practice_messages` | pre-launch baseline 原地修订，不保留旧表/列兼容层 |

#### 3.1.1 Field-Level Enum / Check 来源矩阵

| 字段类别 | 真理源 owner | B4 迁移行为 | 验证要求 |
|----------|-------------|-------------|----------|
| Shared enum | [shared-conventions-codified](../shared-conventions-codified/spec.md) | 由 shared generator 或 checked source 输出 SQL check | 修改后跑 `make codegen-conventions` 与 migration lint |
| API-facing async enum | [openapi-v1-contract](../openapi-v1-contract/spec.md) | `async_jobs.resource_type` 与 API-facing job subset 保持兼容 | DB enum 不得扩大 API response surface |
| DB / event / outbox job | [event-and-outbox-contract](../event-and-outbox-contract/spec.md) | `async_jobs.job_type`、`outbox_events.publish_status` 与 retry 字段跟随 B3 | B4 lint 校验 job manifest 与 DB check |
| Auth / session 状态 | [ADR-Q1](../engineering-roadmap/decisions/ADR-Q1-auth.md) + [backend-auth](../backend-auth/spec.md) | auth tables 不私造业务状态；只存 hash / metadata | 禁止原始 token、cookie secret、provider token 入库 |
| B4 DB-local enum | `migrations/enum-sources.yaml` | 逐列登记 `table.column -> source -> values checksum` | SQL check 无来源时 lint 失败 |
| Migration metadata enum | 本 spec | `schema_backfills.status` 只服务迁移运行时 | 单测覆盖 success / failed / skipped / dry-run |

#### 3.1.2 P0 Privacy Deletion Table Matrix

| 表 / 表组 | P0 删除策略 | 顺序 / 说明 |
|-----------|-------------|-------------|
| `users` | sync soft delete + final hard delete | 请求受理时先置 `deleted_at` 并吊销 session；子表处理完成后硬删用户行 |
| `user_settings` | hard delete | 删除账号设置 |
| `file_objects` / `resumes` | hard delete + object storage delete | 先删对象存储，再删 DB 行；简历原文、解析快照、结构化内容一并删除 |
| `target_jobs` / `target_job_requirements` / `target_job_sources` | cascade / hard delete | 先删 requirements / sources，再删 target job；不得保留 raw JD 或 source URL |
| `practice_plans` / `practice_sessions` / `practice_session_events` / `practice_messages` | cascade / hard delete | messages/events 随 session 级联删除；raw conversation content 必须覆盖 |
| `idempotency_records` | hard delete | 按 `user_id` 删除幂等记录；不得保留可反查用户请求的 fingerprint、response 或错误 payload |
| `feedback_reports` | hard delete | 证据摘要、报告正文、能力重点和复练建议均视为用户内容 |
| `source_records` | hard delete | 外部 source 摘要与 owner 关联一并删除 |
| `ai_task_runs` | hard delete after audit summary | 删除前只允许聚合 token / cost / SLA 计数进入非用户维度指标 |
| `async_jobs` / `outbox_events` | hard delete or redacted terminal tombstone | 与用户资源关联的 payload / result 必须删除；隐私删除执行 job 只保留 redacted terminal 状态 |
| `privacy_requests` | audit tombstone | 完成后保留 request id、type、status、completed_at、duration bucket；`user_id` 置空 |
| `audit_events` | audit tombstone + user event hard delete | 写入 `privacy.delete_completed` 后删除或脱敏该用户 audit 记录 |
| `auth_challenges` / `sessions` / `external_identities` | hard delete | session 立即撤销；challenge hash 与 external provider subject 一并删除 |
| `prompt_versions` / `rubric_versions` | retain | 全局配置表，不按用户删除，不得包含用户内容 |
| `schema_migrations` / `schema_backfills` | retain | 迁移元数据，不含用户内容 |

### 3.2 待确认事项

- `audit_events` / `practice_session_events` 单表行数超过 10M/month 时，是否引入时间分区。
- `target_jobs` 全文索引语言默认 `simple`；若目标市场单一，可由 target-job owner 发起语言配置变更。

## 4 设计约束

### 4.1 命名与目录

- 迁移文件只落在 `migrations/` 根目录；不使用子目录。
- 新迁移通过 `make migrate-create NAME=...` 创建。
- 表名 / 列名使用 `snake_case`；外键、索引、唯一约束分别使用 `fk_`、`idx_`、`uq_` 前缀。

### 4.2 Schema

- 主键统一 `uuid`，由应用层生成 UUIDv7；migration 不设 `default gen_random_uuid()`。
- 时间字段统一 `timestamptz`；`updated_at` 由应用层维护。
- 用户核心资源表使用 `deleted_at timestamptz`；默认过滤由 repository 层负责。
- JSONB 字段必须有明确语义后缀，如 `_payload`、`_metadata`、`_summary`、`_results`。
- Auth 表只存 hash、subject、metadata，不存原始 token 或 provider secret。
- `outbox_events.last_error_message` 只保存 redacted summary，不落 prompt、answer、JD、resume text 或 provider raw response。

### 4.3 性能

- 干净 DB `migrate-up` 应在 60s 内完成。
- 单条 migration 执行目标不超过 5s；需要长耗时 index 时必须拆分 staged path 并标明 ops 窗口。
- 高频表避免长时间阻塞锁。

### 4.4 安全

- Migration 文件不得包含明文 secret / API key。
- `email`、`auth_provider_user_id`、`ip_hash`、`user_agent_hash` 等字段的索引必须考虑 PII 暴露。
- `migrations/lint.sh` 允许 `token_hash`、`session_hash`、`secret_ref`，拒绝 `raw_token`、`session_cookie`、`api_key` 等明文语义字段。

### 4.5 Backfill 执行

- `backend/cmd/migrate` 是迁移唯一可执行入口；Make targets 不绕过 wrapper。
- Up 顺序：SQL up 成功后，若 manifest 存在同号 backfill，则执行 dry-run / apply 并写入 `schema_backfills`。
- Down 顺序：先检查 manifest reversible；不可逆 backfill 在 dev down 时记录 skipped。
- `make migrate-check` 使用干净 DB 执行 up -> down -> up，并断言 ledger 无重复成功记录。
- Backfill 默认 batch size <= 500，支持 `--limit`，失败必须返回非 0。

## 5 模块边界

| 边界 | Owner | 说明 |
|------|-------|------|
| `migrations/` | B4 | 当前应用 / auth / 元数据表、索引、check constraints |
| `backend/cmd/migrate` | B4 | 包装 migration tool 与 Go backfill registry |
| `make migrate-*` | B4 | 本地迁移执行、状态、创建、检查入口 |
| Postgres dev instance | local-dev-stack + B4 | 本地 DB 由 dev stack 提供，schema 由 B4 migration 管理 |
| Business CRUD / repository | 各 backend owner | 不在 B4 baseline 内实现 |
| Enum/check owner | B1 / B2 / B3 / backend-auth / B4 | 按 §3.1.1 分工 |
| Outbox retry semantics | event-and-outbox-contract + B4 | B3 owns 语义，B4 owns schema/index |
| Schema change review | 对应 owner + B4 | 业务 owner 提交 migration，B4 复查 schema gate |

## 6 验收标准

| ID | 场景 | Given | When | Then | 对应 Plan |
|----|------|-------|------|------|-----------|
| C-1 | 干净 DB baseline | 干净 Postgres 18 | `make migrate-up` | §2.1 的 22 app + 3 auth + 2 metadata table 全部存在；public schema count >= 27；app inventory 不多不少 | [001](./plans/001-bootstrap/plan.md) + [002](./plans/002-flat-resume-migration/plan.md) |
| C-2 | 索引覆盖 | C-1 完成 | 查询 `pg_indexes` | 当前 B-Tree inventory 与可选 `idx_target_jobs_fts` 存在 | [001](./plans/001-bootstrap/plan.md) |
| C-3 | 迁移可逆 | C-1 完成 | `make migrate-down` in dev | 应用 / auth / backfill metadata 按 down 语义回滚；exit 0 | [001](./plans/001-bootstrap/plan.md) |
| C-4 | 幂等执行 | 已迁移一次 | 再次 `make migrate-up` | exit 0；`schema_migrations` 无重复 | [001](./plans/001-bootstrap/plan.md) |
| C-5 | Enum/check drift | Source enum 或 job manifest 变化 | codegen + migration lint | DB check 与 owner source 一致，API-facing subset 不被扩大 | [001](./plans/001-bootstrap/plan.md) |
| C-6 | Prod rollback guard | `APP_ENV=prod` | `make migrate-down` | 非 0 退出，并提示需要显式 force | [001](./plans/001-bootstrap/plan.md) |
| C-7 | Backfill ledger | Migration 需要行级回填 | up SQL + Go backfill registry | dry-run / apply 写入 `schema_backfills`，重复执行不重复 apply | [001](./plans/001-bootstrap/plan.md) |
| C-8 | Outbox 查询性能 | Pending rows 积压 | 查询 due events | 使用 due-query index，P95 < 5ms | [001](./plans/001-bootstrap/plan.md) |
| C-9 | AI call meta 查询 | 干净 DB baseline | 查询 `ai_task_runs` columns | model / route / validation / schema / provenance typed columns 存在 | [001](./plans/001-bootstrap/plan.md) |
| C-10 | Privacy matrix | 测试用户产生覆盖样本 | `privacy_delete` dry-run / apply | §3.1.2 每表 disposition 输出并执行；用户可识别内容删除或脱敏 | [001](./plans/001-bootstrap/plan.md) |
| C-11 | Live test rerun-safe | `DATABASE_URL` 指向可用 DB | 固定 UUID migration tests 连续运行 | 重复运行不因样本残留失败；无 DB 时明确 skip | [002](./plans/002-flat-resume-migration/plan.md) |

## 7 关联计划

- [001-bootstrap](./plans/001-bootstrap/plan.md)：当前 baseline migration、wrapper、Make targets、backfill registry、privacy matrix dry-run 与 migration-check owner。
- [002-flat-resume-migration](./plans/002-flat-resume-migration/plan.md)：flat Resume net-state migration、`resumes` schema、`practice_plans.resume_id` binding、enum/check sync 与 live migration test owner。
