# DB Migrations Baseline Spec

> **版本**: 1.16
> **状态**: active
> **更新日期**: 2026-05-12

## 1 背景与目标

[engineering-roadmap spec §5.1](../engineering-roadmap/spec.md#51-当前已存在的-active-spec) 将 B4 `db-migrations-baseline` 定义为当前 active Contract spec（依赖 [B1 `shared-conventions-codified`](../shared-conventions-codified/spec.md) 与 [A2 `local-dev-stack`](./../local-dev-stack/spec.md)）。当前迁移 baseline 由本 spec、`migrations/` 与 product-scope 当前范围决定：28 张当前应用表、[ADR-Q1](../engineering-roadmap/decisions/ADR-Q1-auth.md) 锁定的 3 张 auth / session 支撑表、迁移元数据表与索引契约落到迁移文件层，决定了：

- 后端任何 C 域 spec 在自己的 plan 里能够 import 真实表；
- 各 C 域不得在自己的 plan 中另起重复的 migration 文件夹或选型；
- 迁移文件命名 / 工具 / 回滚策略统一。

A2 已锁定本地 Postgres 18 实例的可用性；本 spec 在此基础上落地 schema。

目标是：

1. **31 张 baseline 应用 / 支撑表 + 2 张迁移元数据表**：覆盖当前产品范围内的 28 张应用表，外加 ADR-Q1 指派给 B4 的 `auth_challenges` / `sessions` / `external_identities` 3 张支撑表；迁移系统自身使用 `schema_migrations` 与 `schema_backfills`。`make migrate-up` 后 public schema 至少有 33 张表。旧 `mistake_entries` 与向量检索表已删除；报告内题目回顾由 `question_assessments` / `feedback_reports` 承载。
2. **统一迁移工具**：`golang-migrate` 或 `atlas` 之一；本 spec §3.1 选型 `golang-migrate`（轻量、社区成熟），在 B4 自身 plan 落地时如发现限制可在本 spec 修订中切换。
3. **可逆 + 数据回填策略**：所有 migration 必须说明是否可逆 / 是否需数据回填 / 是否阻塞写入；当前执行口径以本 spec 的 migration / backfill gate 为准。
4. **索引策略锁定**：当前 B-Tree 与可选 GIN 全文索引 inventory 由本 spec 直接承接；性能查询路径在后续业务域加表时不得遗漏索引。
5. **AI 与隐私承载完整**：`ai_task_runs` 必须显式承载 A3/F1 需要检索的 call meta；隐私删除必须按 per-table matrix 执行，避免 backend internal runner 漏删或误删全局配置表。

本 spec 不实现具体业务表的 CRUD（归各 C 域）、不部署 DB 服务（归 A2）、不实现 Admin 面板（不在 P0 范围）。

## 2 范围

### 2.1 In Scope

- **迁移工具与目录**：`migrations/` 目录（A1 已锁定根容器）；工具 `golang-migrate`（CLI）；文件命名 `NNNNNN_<verb>_<noun>.up.sql` / `NNNNNN_<verb>_<noun>.down.sql`，序号从 `000001` 起 6 位。
- **28 张 P0 应用表**：与 product-scope 当前范围一致：
  1. `users` / 2. `user_settings` / 3. `candidate_profiles` / 4. `experience_cards` / 5. `file_objects` / 6. `resume_assets` / 7. `resume_versions` / 8. `resume_version_suggestions` / 9. `target_jobs` / 10. `target_job_requirements` / 11. `target_job_sources` / 12. `practice_plans` / 13. `idempotency_records` / 14. `practice_sessions` / 15. `practice_session_events` / 16. `practice_turns` / 17. `question_assessments` / 18. `feedback_reports` / 19. `resume_tailor_runs` / 20. `debriefs` / 21. `source_records` / 22. `prompt_versions` / 23. `rubric_versions` / 24. `ai_task_runs` / 25. `async_jobs` / 26. `outbox_events` / 27. `privacy_requests` / 28. `audit_events`。D-17 Resume Workshop additive 升级已新增 `resume_versions`（结构化主版本 + 岗位定制版本）与 `resume_version_suggestions`（tailor run 改写建议状态）；`resume_version_edits`（手动编辑历史）归 P1 延后；`resume_assets` 字段 additive 扩展 `source_type` / `original_text` / `guided_answers` / `parsed_text_snapshot`。具体 migration up/down + check constraint + idx 覆盖由 [db-migrations-baseline/002-resume-versions-additive](./plans/002-resume-versions-additive/plan.md) 落地。旧 `mistake_entries` 与向量检索表已删除；不得作为 P0/P1/P2 后续表恢复。
- **B3 outbox operational columns**：`outbox_events` 必须纳入 [B3 current event-and-outbox contract](../event-and-outbox-contract/spec.md) 锁定的 retry / dead-letter 字段与 due-query 索引。
- **A3/F1 AI call meta typed columns**：`ai_task_runs` 必须显式增加 `model_family` / `model_profile_name` / `model_profile_version` / `fallback_chain` / `route` / `validation_status` / `output_schema_version` typed columns，避免核心观测字段只落 JSONB。
- **3 张 auth / session 支撑表**：`auth_challenges` / `sessions` / `external_identities`，由 [ADR-Q1 §4](../engineering-roadmap/decisions/ADR-Q1-auth.md#4-影响范围) 指派给 B4 baseline；`external_identities` 是 P1 SSO 扩展槽，P0 可保持空表但必须随 baseline 建立。
- **元数据表**：`schema_migrations`（由 `golang-migrate` 自动创建管理）与 `schema_backfills`（由 B4 `backend/cmd/migrate` 记录 Go backfill dry-run / apply 状态）。
- **扩展**：当前 baseline 不启用向量扩展；未来如重新引入向量检索，必须先修订本 spec、A2 dev-stack 与 migration gate。
- **索引**：覆盖本 spec 索引 inventory 中的全部 B-Tree 索引 + 可选 `target_jobs` GIN 全文索引（dev 默认启用，prod 由 C4 在自己 spec 决策是否启用）。
- **Make target**：A1 已占位 `make migrate`，本 spec 落地：`make migrate-up` / `make migrate-down` / `make migrate-status` / `make migrate-create NAME=...` / `make migrate-check`。
- **本地迁移校验**：本地 gate 跑 `make migrate-up && make migrate-down && make migrate-up` 验证 idempotency 与可逆性（`migrate-down` 仅在 dev 模式下生效，prod 模式禁止）；远端 CI 仅在 A5 触发条件成立后再接入。
- **数据回填框架**：DDL 仍由 SQL migration 执行；行级 backfill 由 B4 `backend/cmd/migrate` 包装器按 `migrations/backfill/manifest.yaml` 调用编译进 backend module 的 Go backfill registry，不由 `golang-migrate` 自动调用。每个 backfill 必须支持 `dry-run` / `apply`，并写入 `schema_backfills`。
- **field-level lint**：所有 `text + check (col in (...))` 必须在 §3.1.1 的来源矩阵中登记；B4 lint 根据来源调用 B1 / B2 / B3 generator 输出或 B4-owned manifest，不允许在 SQL 中裸手写未登记枚举列表。
- **privacy deletion matrix**：§3.1.2 锁定 P0 `privacy_delete` backend internal runner 的每表处理策略（hard delete / cascade / retain / audit tombstone），后续删除链路实现必须逐项验证。

### 2.2 Out of Scope

- 业务表的 CRUD / repository 层：归各 C 域。
- Admin / 后台运营表：当前 P0 不在范围。
- workspace / multi-tenant 扩展：归 P1+；本 spec 仅在表注释中预留扩展空间。
- 数据 ETL / data warehouse：归 [F2 `analytics-funnel`](../engineering-roadmap/spec.md#52-当前-p0-实施-workstream-候选)（PostHog 自托管路径）。
- 全量备份 / DR：归 [E4 `release-gate-and-rollout`](../engineering-roadmap/spec.md#52-当前-p0-实施-workstream-候选) + 运维。

## 3 用户决策 / 待确认事项

### 3.1 已锁定决策

| ID | 决策 | 锁定值 | 影响 |
|----|------|--------|------|
| D-1 | 迁移工具 | `golang-migrate v4.18+`（轻量、CLI + Go API、社区成熟） | 不与 atlas / goose 共存；后续切换需本 spec 修订 |
| D-2 | 文件命名 | `NNNNNN_<verb>_<noun>.{up,down}.sql`；序号 6 位连续；编号一旦发布不可回收 | 防止跨 PR 同号冲突 |
| D-3 | 数据回填策略 | DDL 走 SQL；行级数据回填走 Go registry（实现位于 `backend/internal/migrations/backfills/<NNNNNN>/`，入口由 `backend/cmd/migrate` 编译引用）；`migrations/backfill/manifest.yaml` 记录 version / name / reversible / dry-run 支持；同一编号的 SQL + Go 必须配对 | 复杂回填可写 unit test；生产镜像不依赖 `go run` |
| D-4 | 可逆要求 | 每条 migration 必须有 `.down.sql`；不可逆操作（如删列）的 down 至少恢复表结构骨架（数据可记录在 backfill log） | – |
| D-5 | 阻塞操作约束 | 不在高频表（`practice_session_events` / `outbox_events` / `audit_events`）执行阻塞式大表重写；优先 additive；若必须重写需先在本 spec 中登记 | 当前 gate 以本 spec 与 B4 tooling 为准 |
| D-6 | enum / check 约束源头 | 以 §3.1.1 来源矩阵为准：B1 只拥有 shared enum；B2 拥有 API-facing `ResourceType` / `JobType` 子集；B3 拥有 DB / event / outbox jobType 与 publish 状态；ADR-Q1 + C1 拥有 auth/session 状态；B4 只拥有 migration metadata enum | 防止把非 B1 枚举误生成为 shared enum |
| D-7 | 索引覆盖 | 当前 28 张应用表所需 B-Tree 索引 + 可选 GIN 全文（按 D-9 决策） | 性能 baseline |
| D-8 | seed 数据 | 不在 baseline migration 中插入业务数据；business seed 由各 C 域 mock-server plan 提供 | 防止把 demo 数据当默认 |
| D-9 | 全文索引开关 | `target_jobs` GIN 全文索引在 dev 默认启用，staging / prod 由 C4 spec 决策是否启用（迁移条件化） | – |
| D-10 | migrate-down 防呆 | `make migrate-down` 在 `APP_ENV=prod` 时拒绝执行；CI 中通过 `MIGRATE_DOWN_FORCE=1` 显式开启 | 防止误回滚 |
| D-11 | Go workspace / module 模式 | 复用当前根 `go.work` + `backend/go.mod`；迁移命令落 `backend/cmd/migrate/main.go`，不为 `migrations/` 另起 Go module | 与 B1 已落地 module 拓扑一致 |
| D-12 | DB extension 生命周期 | 当前 baseline 不创建向量扩展，down migration 不管理 DB extension；新增扩展必须先经本 spec 修订登记 | 避免当前阶段为未使用能力引入基础设施负担 |
| D-13 | backfill ledger | `schema_backfills(version, name, mode, status, checksum, started_at, completed_at, error_message)` 由 B4 管理；同一 `version + mode + checksum` apply 成功后不得重复执行，除非 `--force` 且 APP_ENV!=prod | backfill 可审计、幂等、可 dry-run |
| D-14 | outbox retry 字段 | `outbox_events` 必须追加 `publish_attempts integer not null default 0` / `next_attempt_at timestamptz not null default now()` / `locked_at timestamptz` / `last_error_code text` / `last_error_message text`；pending due 查询索引至少覆盖 `(publish_status, next_attempt_at, created_at)` | 支撑 B3/C8 dispatcher retry、failed 人工排查与 P2 告警 |
| D-15 | AI call meta typed columns | `ai_task_runs` 必须追加 `model_family text` / `model_profile_name text` / `model_profile_version text` / `fallback_chain jsonb not null default '[]'::jsonb` / `route text` / `validation_status text` / `output_schema_version text`，并由 [F3 `prompt-rubric-registry/001-baseline`](../prompt-rubric-registry/plans/001-baseline/plan.md) 阶段 4.2 追加 `feature_key text not null` / `feature_flag text not null default 'none'` / `data_source_version text not null default 'not_applicable'` 三个 prompt/rubric provenance typed columns；`metadata` 只承载 hash / 长度 / profile 摘要与后向兼容字段 | 支撑 A3 `AICallMeta`、F1 dashboard、B2 `GenerationProvenance` 可查询闭环；F3 prompt/rubric provenance 通过 typed columns 直接落盘，不再依赖 JSONB path scan |
| D-16 | P0 privacy deletion matrix | §3.1.2 是 `privacy_delete` backend internal runner 的表级真理源；新增表或新增用户关联列时必须先更新本矩阵，再改 migration / runner | 防止删除链路漏表、误删全局配置或丢失最小审计证据 |
| D-17 | Resume Workshop additive 表与字段（已落地） | 本次修订落地 Resume Workshop 阶段 0 contract additive 升级所需 DB 范围：（1）新增 `resume_versions` 表（id / user_id / resume_asset_id FK / parent_version_id FK self / version_type [structured_master, targeted] / target_job_id FK NULLABLE / display_name / seed_strategy NULLABLE [copy_master, blank, ai_select] / focus_angle NULLABLE / structured_profile jsonb / match_score numeric NULLABLE / prompt_version / rubric_version / model_id / provider / created_at / updated_at / deleted_at + idx `(user_id, updated_at DESC)` + idx `(resume_asset_id, version_type)` + idx `(parent_version_id) WHERE parent_version_id IS NOT NULL`）；（2）新增 `resume_version_suggestions` 表（id / resume_version_id FK ON DELETE CASCADE / tailor_run_id FK / original_bullet / suggested_bullet / reason / status [pending, accepted, rejected] / decided_at + idx `(resume_version_id, status)` + idx `(tailor_run_id)`）；（3）`resume_assets` additive 字段（`source_type` text NULL CHECK ∈ {`upload`, `paste`, `guided`} / `original_text` text NULL / `guided_answers` jsonb NULL / `parsed_text_snapshot` text NULL，保持向后兼容）；（4）`resume_version_edits` 表（手动编辑历史）归 P1 延后；`migrations/enum-sources.yaml` 同步登记 3 个 B1 enum check 来源（B1 D-10 `ResumeVersionType` / `ResumeSeedStrategy` / `ResumeTailorSuggestionStatus`）+ 1 个 B2 API-facing sourceType 来源（B2 D-18 `RegisterResumeRequest.sourceType`）；具体 SQL 与 idx 由 `migrations/000005_resume_versions.{up,down}.sql` 承接 | `migrations/000005_resume_versions.up.sql` + `.down.sql`、`migrations/enum-sources.yaml`、§2.1 表 inventory（28 张 P0 应用表）、§3.1.2 privacy deletion matrix、`docs/spec/openapi-v1-contract` D-18 同步、`docs/spec/event-and-outbox-contract` D-14 同步、§6 验收（baseline 表数 28）|

#### 3.1.1 field-level enum / check 来源矩阵

| 字段类别 | 真理源 owner | B4 迁移行为 | 验证要求 |
|----------|-------------|-------------|----------|
| B1 shared enum（如 target / practice / report / question review / privacy request status 等已进入 `shared/conventions.yaml` 的值） | [B1 `shared-conventions-codified`](../shared-conventions-codified/spec.md#31-已锁定决策) | 通过 B1 generator 输出 SQL check 片段或 B4 可消费的 enum manifest | 修改 B1 enum 后，`make codegen-conventions && make migrate-check` 必须产生或验证对应 check 漂移 |
| API-facing async `ResourceType` / `JobType` | [B2 §3.1.2](../openapi-v1-contract/spec.md#312-b2-专属-async-enum-字面量) | `async_jobs.resource_type` 必须兼容 B2 `ResourceType`；会经 `GET /api/v1/jobs/{jobId}` 暴露的 `job_type` 不得超出 B2 `JobType` | B4 lint 检查 DB schema 与 B2 API-facing enum 的兼容子集 |
| DB / event / outbox job、publish 状态与 retry 字段 | [B3 `event-and-outbox-contract`](../event-and-outbox-contract/spec.md) | `async_jobs.job_type` 可包含 B3 canonical 9 项（含 internal-only `email_dispatch`）；`outbox_events.publish_status`、retry operational columns、事件字段约束跟随 B3 | B4 lint 检查 B3 job manifest 与 DB check 一致，确认 B2 API-facing 子集不被扩大，并确认 outbox due-query 索引存在 |
| Auth / session 支撑表状态与类型 | [ADR-Q1](../engineering-roadmap/decisions/ADR-Q1-auth.md#3-决策) + 后续 C1 `backend-auth` spec | `auth_challenges` / `sessions` / `external_identities` 的 token / session / provider 字段不得在 B4 私造业务状态；若 C1 需要新增 check 值，先修订对应 spec 再改 migration | 原始 magic-link token、cookie secret、provider token 不入库；只存 hash / metadata |
| B4 DB-local enum（尚未提升到 B1/B2/B3） | B4 `migrations/enum-sources.yaml` | 001-bootstrap 必须逐列登记 `table.column -> source -> values checksum`；后续若该 enum 需要跨 API / event 复用，再提升到 B1/B2/B3 | SQL 中出现未登记 check list 时 `migrations/lint.sh` 失败 |
| B4 migration metadata enum | B4 本 spec | `schema_backfills.status` 等只服务迁移运行时，不进入 API / event / shared enum | B4 单测覆盖 success / failed / skipped / dry-run 状态 |

#### 3.1.2 P0 privacy deletion table matrix

本矩阵覆盖 B4 baseline 表与 ADR-Q1 支撑表，是 C8 `privacy_delete` 的执行真理源。`hard delete` 表示删除当前用户关联行；`cascade` 表示由 FK / repository 顺序删除子表；`retain` 表示全局配置或 migration 元数据不得按用户删除；`audit tombstone` 表示只保留无法反推用户身份的最小删除证据。

| 表 / 表组 | P0 删除策略 | 顺序 / 说明 |
|-----------|-------------|-------------|
| `users` | sync soft delete + final hard delete | 请求受理时先置 `deleted_at` 并吊销 session；所有子表处理完成后硬删用户行 |
| `user_settings` / `candidate_profiles` / `experience_cards` | hard delete | 先删画像与偏好，避免后续生成链路继续读取 |
| `file_objects` / `resume_assets` | hard delete + object storage delete | 先删除对象存储文件，再删 DB 行；失败时保留 retryable job 状态 |
| `target_jobs` / `target_job_requirements` / `target_job_sources` | cascade / hard delete | 先删 requirements / sources，再删 target job；不得保留 raw JD / source URL |
| `practice_plans` / `practice_sessions` / `practice_session_events` / `practice_turns` | cascade / hard delete | 先删事件流与 turns，再删 session / plan；raw answer text 必须覆盖 |
| `idempotency_records` | hard delete | 按 `user_id` 删除幂等记录；不得保留可反查用户请求的 fingerprint、response 或错误 payload |
| `question_assessments` / `feedback_reports` | hard delete | 证据片段、报告正文、题目回顾和本轮复练建议均视为用户内容；无独立 `mistake_entries` 表 |
| `resume_tailor_runs` / `debriefs` | hard delete | P0 debrief replay 与简历定制输出均随用户删除 |
| `resume_versions` / `resume_version_suggestions` | hard delete + cascade | 先删 suggestions（FK 子表）再删 versions；结构化主版本、岗位定制版本与 AI 改写建议均视为用户内容；`structured_profile` jsonb、`original_bullet` / `suggested_bullet` / `reason` 文本必须覆盖；与 `resume_assets` 删除同一事务 |
| `source_records` | hard delete | 外部 source 摘要与 owner 关联一并删除 |
| `ai_task_runs` | hard delete after audit summary | 删除前只允许聚合 token/cost/SLA 计数进入非用户维度指标；不保留 prompt/response 摘要行 |
| `async_jobs` / `outbox_events` | hard delete or redacted terminal tombstone | 与用户资源关联的 payload/result 必须删除；若为正在执行的 `privacy_delete` job，仅保留 redacted terminal 状态直到 backend internal runner 完成 |
| `privacy_requests` | audit tombstone | 完成后保留 request id、request type、status、completed_at、duration bucket；移除或 hash `user_id`，不保留可识别内容 |
| `audit_events` | audit tombstone + user event hard delete | 先写 `privacy.delete_completed`，再删除/脱敏该用户历史 audit；仅保留不可反推用户的删除完成 tombstone 供 SLA 证明 |
| `auth_challenges` / `sessions` / `external_identities` | hard delete | session 立即撤销；challenge token hash、external provider subject 一并删除 |
| `prompt_versions` / `rubric_versions` | retain | 全局 prompt/rubric 版本表，不按用户删除；不得包含用户内容 |
| `schema_migrations` / `schema_backfills` | retain | migration 元数据，不含用户内容 |

### 3.2 待确认事项

- 是否在后续业务域加表时引入 `pg_partman` / 时间分区（针对 `audit_events` / `practice_session_events` 高频表）：默认 P0 不分区；如 monthly 表行数 > 10M 再决策。
- `target_jobs` 全文索引语言（`simple` vs `english`）：默认 `simple`（多语言兼容），由 C4 在自己 spec 中决策。

## 4 设计约束

### 4.1 命名与目录约束

- 所有迁移文件落 `migrations/`（A1 锁定的根目录）；不允许子目录（`migrations/<domain>/...`）—— `golang-migrate` 默认按文件名扁平排序。
- 序号必须严格递增；新增迁移由 `make migrate-create NAME=add_xxx_to_yyy` 自动生成，不允许手敲。
- 表名 / 列名严格 `snake_case`；与 B1 当前命名约定和 B4 migration lint 一致；外键命名 `fk_<from_table>_<column>` / `idx_<table>_<purpose>` / `uq_<table>_<columns>`。

### 4.2 schema 约束

- 主键统一 `uuid` + 应用层生成 UUIDv7（由 [B1 idx 工具](../shared-conventions-codified/spec.md#21-in-scope) 提供）；migration 不设 `default gen_random_uuid()`。
- 时间字段统一 `timestamptz`；默认 `now()`；`updated_at` 由应用层维护（不引 trigger）。
- 软删字段 `deleted_at timestamptz` 在用户核心资源表上必须存在；查询默认 filter `deleted_at IS NULL` 由各 C 域 repo 层负责。
- jsonb 字段命名以 `_payload` / `_metadata` / `_summary` / `_results` 等明确语义；不允许 `data jsonb`。
- Auth 支撑表必须遵守 ADR-Q1：`auth_challenges` 只存 challenge token hash / pepper 后摘要与 IP / UA hash，不存原始 token；`sessions` 是 server-side session 真理源，cookie 值只作为不透明引用；`external_identities` 是 P1 SSO 空表槽，不在 P0 引入 OAuth token 存储。
- `outbox_events` 必须遵守 B3 D-7/D-8：`last_error_message` 只保存 redacted summary，不得落 raw provider response / prompt / answer / JD / resume text；`publish_status` check 仍只允许 `pending` / `published` / `failed`。

### 4.3 性能约束

- `migrate-up` 在干净 DB 上必须 ≤ 60s 完成（dev 机器）；超出由 owner 拆分迁移文件。
- 单条 migration 执行时间 ≤ 5s（CI 模式）；如必须长时间运行（如 `CREATE INDEX CONCURRENTLY`）需在 PR 描述中标注「需要 ops 窗口」，并提供 staged migration 路径。
- 任何 migration 不得 `LOCK TABLE` 在生产高频表上 > 100ms。

### 4.4 安全约束

- migration 文件不得包含明文 secret / API key。
- 任何敏感字段（`email` / `auth_provider_user_id` / `ip_hash` / `user_agent_hash`）的索引必须考虑 PII 暴露：默认不索引明文 email（已通过 `email unique` 实现），其它 hash 字段索引可接受。
- B4 提供 `migrations/lint.sh`（可选）扫描 `password` / `secret` / `token` 关键字。
- `migrations/lint.sh` 必须允许 `token_hash` / `session_hash` / `secret_ref` 等安全命名，但拒绝 `raw_token` / `session_cookie` / `api_key` 等明文语义字段，豁免需在本 spec 修订登记。

### 4.5 backfill 执行约束

- `backend/cmd/migrate` 是迁移唯一可执行入口；`make migrate-up` / `make migrate-down` / `make migrate-status` / `make migrate-check` 都调用该入口，不直接绕过到裸 `golang-migrate` CLI。
- up 顺序：按 SQL version 执行 `.up.sql`；每个 version 成功后，若 `migrations/backfill/manifest.yaml` 存在同号 backfill，则先执行 `dry-run`，再在非 dry-run 模式执行 `apply`；两步状态都写入 `schema_backfills`。
- down 顺序：先按 manifest 检查该 version 是否声明 `reversible=true`；不可逆 backfill 在 dev down 时允许跳过并记录 `skipped`，prod down 仍由 D-10 防呆拒绝。
- `make migrate-check` 在临时干净 DB 中执行 `migrate-up -> migrate-down -> migrate-up`，并额外断言 `schema_backfills` 对 dry-run / apply 状态无重复成功记录。
- backfill 必须按批次提交，默认 batch size ≤ 500，支持 `--limit`；失败时保留错误消息但不得吞掉非 0 退出码。

## 5 模块边界

| 边界 | Owner | 说明 |
|------|-------|------|
| `migrations/` 目录与 baseline 文件 | B4 | 31 张当前应用 / auth 支撑表 + 2 张迁移元数据表 + 索引 |
| `backend/cmd/migrate/main.go` 与 `make migrate-*` target | B4 | A1 占位 `make migrate` 的真实实现；包装 `golang-migrate` 与 Go backfill registry |
| Postgres 实例（dev） | A2 + B4 | A2 提供空 Postgres；B4 通过 migration 管理当前应用 schema |
| 业务表 CRUD / repository | 各 C 域 | 通过 generated DTO（B2）+ raw SQL / sqlc |
| Schema enum / check 值 | B1 + B2 + B3 + ADR-Q1/C1 + B4 | 按 §3.1.1 来源矩阵生成或校验；B4 不私造跨域 enum |
| `outbox_events` retry operational columns | B3 + B4 | B3 owns 字段语义与 dispatcher 查询；B4 owns migration / index / rollback |
| 后续表新增 / 修改 | 各 C 域 + B4 review | 各 C 域提交 PR 时由 B4 owner approve schema 变更 |
| Workspace / multi-tenant 扩展 | P1+ | 本 spec 仅预留扩展空间 |
| 备份 / DR / 跨区复制 | E4 + 运维 | 不在本 spec 范围 |

## 6 验收标准

| ID | 场景 | Given | When | Then | 对应 Plan |
|----|------|-------|------|------|-----------|
| C-1 | 干净 DB baseline 落地 | 干净 Postgres 18（A2 dev stack 启动） | `make migrate-up` | 28 张当前应用表 + `auth_challenges` / `sessions` / `external_identities` + `schema_migrations` + `schema_backfills` 全部存在；`select count(*) from information_schema.tables where table_schema='public'` ≥ 33；不存在 `mistake_entries` 或向量检索表 | B4 后续 001 + 002 |
| C-2 | 索引覆盖 | C-1 完成 | `select indexname from pg_indexes where schemaname='public'` | 本 spec 当前索引 inventory 中的 B-Tree 索引 + 可选 `idx_target_jobs_fts`（GIN）存在 | B4 后续 001 |
| C-3 | 迁移可逆 | C-1 完成 | `make migrate-down` （dev） | B4 应用 / auth 支撑表与 `schema_backfills` 按 down 语义回滚；`schema_migrations` 可作为工具元数据保留；down 不管理 DB extension；exit 0 | B4 后续 001 |
| C-4 | idempotent | 已 `migrate-up` 一次 | 再次 `make migrate-up` | exit 0；`schema_migrations` 元数据无重复 | B4 后续 001 |
| C-5 | enum/check 来源矩阵 | B1 修改 shared enum，或 B3 修改 `shared/jobs.yaml` jobType（如新增 internal-only `email_dispatch`） | `make codegen-conventions && make migrate-check`（在干净 DB）+ B4 lint | 对应 `text + check (col in (...))` 与 §3.1.1 来源一致；B2 API-facing `JobType` 子集未被 DB-only job 扩大；internal-only job 只进入 DB/C8 check；本地 drift 通过 | B4 后续 001 + B1/B2/B3 |
| C-6 | prod 安全 | `APP_ENV=prod` | `make migrate-down` | exit 非 0；stderr 提示需 `MIGRATE_DOWN_FORCE=1` | B4 后续 001 |
| C-7 | 数据回填 | 某 migration 需要回填 `users.display_name` | up SQL + `backend/cmd/migrate` Go backfill registry 联动 | up SQL 加列；backfill `dry-run` 与 `apply` 状态写入 `schema_backfills`；重复执行不重复 apply；远端 CI 执行仅在 A5 触发条件成立后再接入 | B4 后续 001 + 实际场景 |
| C-8 | outbox + async_jobs 索引性能 | `outbox_events` 1 万行积压，其中部分 `next_attempt_at <= now()` | `select * from outbox_events where publish_status='pending' and next_attempt_at <= now() order by next_attempt_at asc, created_at asc limit 100` | 走 `(publish_status, next_attempt_at, created_at)` 或等价索引；P95 < 5ms | B4 后续 001 + B3 + C8 |
| C-10 | 迁移完整闭环 | 本地修改 schema 或 backfill | `make migrate-check` | `migrate-up && migrate-down && migrate-up` 全部成功；`schema_backfills` dry-run / apply ledger 无重复成功记录；enum/check 来源 lint 与 diff 检查通过；远端 CI 仅在 A5 触发条件成立后再接入 | B4 后续 001 |
| C-11 | B3 outbox 字段承载 | 干净 DB baseline 落地 | `select column_name from information_schema.columns where table_name='outbox_events'` | `publish_attempts` / `next_attempt_at` / `locked_at` / `last_error_code` / `last_error_message` 存在；`publish_status` check 仍只允许 `pending` / `published` / `failed` | B4 后续 001 + B3 |
| C-12 | A3/F1 AI call meta 承载 | 干净 DB baseline 落地 | `select column_name from information_schema.columns where table_name='ai_task_runs'` | `model_family` / `model_profile_name` / `model_profile_version` / `fallback_chain` / `route` / `validation_status` / `output_schema_version` 以及 F3 prompt/rubric provenance 三字段（`feature_key` / `feature_flag` / `data_source_version`）typed columns 存在；核心 dashboard 查询不依赖 JSONB path scan | B4 后续 001 + A3 + F1 + F3 `prompt-rubric-registry/001-baseline` |
| C-13 | P0 privacy deletion matrix 可执行 | B4 baseline migration 已完成，测试用户产生覆盖 31 张当前应用 / auth 支撑表的样本数据 | 触发 C8 `privacy_delete` dry-run / apply | dry-run 输出 §3.1.2 每表 disposition；apply 后用户可识别内容已 hard delete / cascade / tombstone；`prompt_versions` / `rubric_versions` / migration metadata 保留 | B4 后续 001 + 002 + C8 |
| C-14 | live migration tests rerun-safe cleanup | `DATABASE_URL` 指向已 migrate-up 的 live Postgres，测试使用固定 UUID seed | 连续运行 `go test ./internal/migrations/... -run 'TestResumeVersions\|TestResumeAssetDeleteRequiresVersionCleanup' -count=2 -v` | 若 live DB 可用，重复运行不得因 `resume_versions` / `resume_version_suggestions` / `resume_assets` / `target_jobs` 固定 UUID 残留失败；清理错误必须使测试失败。若 `DATABASE_URL` 缺失，输出 skip 证据，不得把 skip 当 live cleanup PASS | db-migrations-baseline/002 Phase 5 |

## 7 关联计划

B4 由以下 plan 承接：

- [001-bootstrap](./plans/001-bootstrap/plan.md)（已完成）：
  - 落地一组严格递增的 `migrations/000001_*.up.sql` / `.down.sql`，覆盖 29 张应用 / auth 支撑表、B3 outbox retry operational columns、A3/F1 AI call meta typed columns、`schema_backfills`、索引与 `schema_migrations` 工具表协同；不强制一表一文件，但编号必须稳定、可 review。
  - 落地 `backend/cmd/migrate/main.go`、`make migrate-*` target，替换 A1 占位。
  - 落地 `migrations/backfill/manifest.yaml`、`backend/internal/migrations/backfills/` registry 框架与 1 个最小示例。
  - 本地 `make migrate-check` 完整闭环（C-10），并提供 privacy deletion matrix dry-run fixture（C-13）；远端 CI 仅在 A5 触发条件成立后再评估。

- [002-resume-versions-additive](./plans/002-resume-versions-additive/plan.md)：D-17 Resume Workshop additive 表与字段落地。新增 `migrations/000005_resume_versions.{up,down}.sql` 引入 `resume_versions` / `resume_version_suggestions` 表 + `resume_assets` 字段补充（`source_type` / `original_text` / `guided_answers` / `parsed_text_snapshot`）；同步 `migrations/enum-sources.yaml` 与 §3.1.2 privacy deletion matrix；baseline 表 inventory 已升至 28 张 P0 应用表；提供 migration up/down 幂等测试、check constraint 验证、idx 覆盖测试；Phase 5 已补 live test cleanup 顺序，固定 UUID 测试必须 rerun-safe。`resume_version_edits` 表归 P1 延后。

后续业务域加表 / 改 schema：各 C 域在自己 plan 中提交 migration 文件，schema 变更必须由 B4 owner 复查；本 spec 不需逐次修订（仅在跨域 schema 决策变更时修订，如分区 / multi-tenant 升格）。
